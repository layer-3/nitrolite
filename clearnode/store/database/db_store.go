package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBStore struct {
	inTx bool
	db   *gorm.DB
}

func NewDBStore(db *gorm.DB) DatabaseStore {
	return &DBStore{db: db}
}

func (s *DBStore) ExecuteInTransaction(txFunc StoreTxHandler) error {
	if s.inTx {
		return txFunc(s)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txStore := &DBStore{
			inTx: true,
			db:   tx,
		}
		return txFunc(txStore)
	})
}

// GetUserBalances retrieves the balances for a user's wallet.
func (s *DBStore) GetUserBalances(wallet string) ([]core.BalanceEntry, error) {
	wallet = strings.ToLower(wallet)

	var balances []UserBalance
	err := s.db.Where("user_wallet = ?", wallet).Find(&balances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
	}

	result := make([]core.BalanceEntry, 0, len(balances))
	for _, balance := range balances {
		result = append(result, core.BalanceEntry{
			Asset:   balance.Asset,
			Balance: balance.Balance,
		})
	}

	return result, nil
}

// LockUserState locks a user's balance row for update (postgres only, must be used within a transaction).
// Uses INSERT ... ON CONFLICT DO NOTHING to ensure the row exists, then SELECT ... FOR UPDATE to lock it.
// Returns the current balance or zero if the row was just inserted.
func (s *DBStore) LockUserState(wallet, asset string) (decimal.Decimal, error) {
	wallet = strings.ToLower(wallet)
	now := time.Now()

	// Check if this is postgres - only postgres supports FOR UPDATE in the way we need
	if s.db.Dialector.Name() == "postgres" {
		// First, ensure the row exists using INSERT ... ON CONFLICT DO NOTHING
		newBalance := UserBalance{
			UserWallet: wallet,
			Asset:      asset,
			Balance:    decimal.Zero,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newBalance).Error
		if err != nil {
			return decimal.Zero, fmt.Errorf("failed to ensure user balance row exists: %w", err)
		}

		// Now lock the row and retrieve the balance using FOR UPDATE
		var balance UserBalance
		err = s.db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_wallet = ? AND asset = ?", wallet, asset).
			First(&balance).Error
		if err != nil {
			return decimal.Zero, fmt.Errorf("failed to lock user balance: %w", err)
		}

		return balance.Balance, nil
	}

	// For non-postgres databases (like sqlite in tests), just return the balance without locking
	var balance UserBalance
	err := s.db.Where("user_wallet = ? AND asset = ?", wallet, asset).First(&balance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create the row if it doesn't exist
			balance = UserBalance{
				UserWallet: wallet,
				Asset:      asset,
				Balance:    decimal.Zero,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := s.db.Create(&balance).Error; err != nil {
				return decimal.Zero, fmt.Errorf("failed to create user balance: %w", err)
			}
			return decimal.Zero, nil
		}
		return decimal.Zero, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance.Balance, nil
}

// EnsureNoOngoingStateTransitions validates that no conflicting blockchain operations are pending.
// This method prevents race conditions by ensuring blockchain state versions
// match the user's last signed state version before accepting new transitions.
//
// Validation logic by transition type:
//   - home_deposit: Verify last_state.version == home_channel.state_version
//   - mutual_lock: Verify last_state.version == home_channel.state_version == escrow_channel.state_version
//     AND next transition must be escrow_deposit
//   - escrow_lock: Verify last_state.version == escrow_channel.state_version
//     AND next transition must be escrow_withdraw or migrate
//   - escrow_withdraw: Verify last_state.version == escrow_channel.state_version
//   - migrate: Verify last_state.version == home_channel.state_version
func (s *DBStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	wallet = strings.ToLower(wallet)

	type versionCheck struct {
		TransitionType       core.TransitionType
		StateVersion         uint64
		HomeChannelVersion   *uint64
		EscrowChannelVersion *uint64
	}

	var result versionCheck
	tx := s.db.Raw(`
		SELECT
			s.transition_type as transition_type,
			s.version as state_version,
			hc.state_version as home_channel_version,
			ec.state_version as escrow_channel_version
		FROM channel_states s
		LEFT JOIN channels hc ON hc.channel_id = s.home_channel_id
		LEFT JOIN channels ec ON ec.channel_id = s.escrow_channel_id
		WHERE s.user_wallet = ?
			AND s.asset = ?
			AND s.user_sig IS NOT NULL
			AND s.node_sig IS NOT NULL
		ORDER BY s.epoch DESC, s.version DESC
		LIMIT 1
	`, wallet, asset).Scan(&result)

	if tx.Error != nil {
		return fmt.Errorf("failed to check state transitions: %w", tx.Error)
	}

	// No previous state found - check RowsAffected instead of StateVersion == 0
	// (StateVersion == 0 could be a valid version)
	if tx.RowsAffected == 0 {
		return nil
	}

	// Validation logic by transition type
	switch result.TransitionType {
	case core.TransitionTypeHomeDeposit:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home deposit is still ongoing")
		}

	case core.TransitionTypeHomeWithdrawal:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home withdrawal is still ongoing")
		}

	case core.TransitionTypeMutualLock:
		// Verify last_state.version == home_channel.state_version == escrow_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion ||
			result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("mutual lock is still ongoing")
		}

	case core.TransitionTypeEscrowLock:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow lock is still ongoing")
		}

	case core.TransitionTypeEscrowWithdraw:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow withdrawal is still ongoing")
		}

	case core.TransitionTypeMigrate:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home chain migration is still ongoing")
		}
	}

	return nil
}
