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
			Asset:    balance.Asset,
			Balance:  balance.Balance,
			Enforced: balance.Enforced,
		})
	}

	return result, nil
}

// RefreshUserEnforcedBalance recomputes the enforced balance from the user's open home channel on-chain state.
func (s *DBStore) RefreshUserEnforcedBalance(wallet, asset string) error {
	wallet = strings.ToLower(wallet)
	asset = strings.ToLower(asset)

	// The protocol enforces at most one open home channel per (user, asset), so
	// the subquery matches a single row in practice. ORDER BY is added as
	// defence-in-depth to keep the result deterministic if that invariant is
	// ever violated.
	err := s.db.Exec(`
		UPDATE user_balances
		SET enforced = COALESCE((
			SELECT s.home_user_balance
			FROM channels c
			JOIN channel_states s ON s.home_channel_id = c.channel_id AND s.version = c.state_version
			WHERE c.user_wallet = user_balances.user_wallet
			  AND c.asset = user_balances.asset
			  AND c.type = 1
			  AND c.status <= 1
			  AND c.state_version > 0
			ORDER BY c.updated_at DESC
			LIMIT 1
		), 0)
		WHERE user_wallet = ? AND asset = ?
	`, wallet, asset).Error
	if err != nil {
		return fmt.Errorf("failed to refresh user locked balance: %w", err)
	}
	return nil
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

// LockUserStateForHomeChannel acquires the balance-row lock of the user owning channelID and
// returns the channel read *after* the lock is held. The order matters: the lock is taken first,
// then the channel is read in a separate statement, so a concurrent transaction (e.g. submit_state
// co-signing a Finalize and flipping the channel to Closing) that commits while we wait on the lock
// is reflected in the returned status. Callers that previously did GetChannelByID followed by
// LockUserState must use this instead — the separate read is read-before-lock and races.
//
// Returns (nil, nil) if the channel does not exist. Channel-type checks remain the caller's
// responsibility.
func (s *DBStore) LockUserStateForHomeChannel(channelID string) (*core.Channel, error) {
	channelID = strings.ToLower(channelID)
	if !strings.HasPrefix(channelID, "0x") {
		channelID = "0x" + channelID
	}

	// Non-postgres (sqlite in tests) cannot SELECT ... FOR UPDATE and has no real concurrency
	// in those paths; resolve, ensure the balance row via LockUserState, and read directly.
	if s.db.Dialector.Name() != "postgres" {
		channel, err := s.GetChannelByID(channelID)
		if err != nil || channel == nil {
			return channel, err
		}
		if _, err := s.LockUserState(channel.UserWallet, channel.Asset); err != nil {
			return nil, err
		}
		return channel, nil
	}

	// Resolve the channel's (wallet, asset) lock key. These columns are immutable for a given
	// channel, so reading them at this statement's snapshot is safe even though status is not.
	var key struct {
		UserWallet string
		Asset      string
	}
	result := s.db.Raw(`SELECT user_wallet, asset FROM channels WHERE channel_id = ?`, channelID).Scan(&key)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to resolve lock key for channel %s: %w", channelID, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	// Acquire the (wallet, asset) balance-row lock first (ensures the row, then SELECT ... FOR
	// UPDATE). This blocks until any concurrent transaction holding the row commits.
	if _, err := s.LockUserState(key.UserWallet, key.Asset); err != nil {
		return nil, err
	}

	// Read the channel only after the lock is held. A single SELECT ... FOR UPDATE OF b that
	// joins channels would return c.* from the statement-start snapshot — a Finalize that flips
	// status to Closing while we wait on the balance lock would not be reflected. This separate
	// statement takes a fresh snapshot once the lock is acquired, so the returned status reflects
	// any such committed transition.
	//
	// NOTE: requires READ COMMITTED isolation (Postgres default, and nitronode never overrides it).
	// Under REPEATABLE READ or SERIALIZABLE this statement still sees the transaction-start
	// snapshot, returning the stale pre-lock status and negating the fix.
	return s.GetChannelByID(channelID)
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
		HomeChannelStatus    *core.ChannelStatus
		EscrowChannelVersion *uint64
	}

	var result versionCheck
	tx := s.db.Raw(`
		SELECT
			s.transition_type as transition_type,
			s.version as state_version,
			hc.state_version as home_channel_version,
			hc.status as home_channel_status,
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
		// Verify last_state.version == home_channel.state_version AND channel is Open.
		// Defence-in-depth: without the status check, a Void channel
		// (status=Void, state_version=0) trivially matches a state_version=0 signed
		// HomeDeposit and the gate would treat the deposit as settled on-chain before
		// any confirmation event lands.
		if result.HomeChannelVersion == nil ||
			result.HomeChannelStatus == nil ||
			result.StateVersion != *result.HomeChannelVersion ||
			*result.HomeChannelStatus != core.ChannelStatusOpen {
			return fmt.Errorf("home deposit is still ongoing")
		}

	case core.TransitionTypeHomeWithdrawal:
		// Verify last_state.version == home_channel.state_version AND channel is Open.
		if result.HomeChannelVersion == nil ||
			result.HomeChannelStatus == nil ||
			result.StateVersion != *result.HomeChannelVersion ||
			*result.HomeChannelStatus != core.ChannelStatusOpen {
			return fmt.Errorf("home withdrawal is still ongoing")
		}

	case core.TransitionTypeMutualLock:
		// Verify last_state.version == home_channel.state_version == escrow_channel.state_version
		if result.HomeChannelVersion == nil || result.StateVersion != *result.HomeChannelVersion ||
			result.EscrowChannelVersion == nil || result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("mutual lock is still ongoing")
		}

	case core.TransitionTypeEscrowLock:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion == nil || result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow lock is still ongoing")
		}

	case core.TransitionTypeEscrowWithdraw:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion == nil || result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow withdrawal is still ongoing")
		}

	case core.TransitionTypeMigrate:
		// NOTE: Migration flows are not yet active off-chain. This case is currently unreachable
		// because submit_state.go rejects TransitionTypeMigrate before EnsureNoOngoingStateTransitions
		// is called. When migration is implemented, this check must be extended: verifying that
		// last_state.version == home_channel.state_version confirms only that the preparation state
		// was enforced on the OLD home chain. A complete gate must also confirm that
		// initiateMigration() was submitted on the NEW home chain (i.e. a MIGRATING_IN channel
		// exists with a matching state version), matching the MigrationInInitiated lifecycle step
		// documented in protocol-description.md.
		if result.HomeChannelVersion == nil || result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home chain migration is still ongoing")
		}
	}

	return nil
}

// EnsureNoOngoingEscrowOperation validates that the user has no in-flight escrow
// operation that would prevent the node from issuing receiver-side states (transfer
// receive, app-session release).
//
// Validation logic by latest signed transition type:
//   - escrow_lock / mutual_lock: always considered ongoing (no finalization yet)
//   - escrow_deposit: considered settled when the on-chain escrow channel version
//     equals the signed state version (finalize landed). When the chain is exactly
//     one behind, the signed N+1 finalize state lives off-chain and is gated on
//     the escrow channel status: allowed for Open (protocol-intended steady state
//     before the purge queue fires) and Closed (post-purge or post-finalize);
//     blocked for Challenged (on-chain resolution still racing — finalize tx may
//     not land, escrow chain may settle at INITIATE, and replaying N+1 later
//     could violate engine invariants).
//   - escrow_withdraw: considered ongoing while the on-chain escrow channel state
//     version has not caught up with the signed state version
//   - any other transition: not an escrow operation, allow
func (s *DBStore) EnsureNoOngoingEscrowOperation(wallet, asset string) error {
	wallet = strings.ToLower(wallet)

	type escrowCheck struct {
		TransitionType       core.TransitionType
		StateVersion         uint64
		EscrowChannelVersion *uint64
		EscrowChannelStatus  *core.ChannelStatus
	}

	var result escrowCheck
	tx := s.db.Raw(`
		SELECT
			s.transition_type as transition_type,
			s.version as state_version,
			ec.state_version as escrow_channel_version,
			ec.status as escrow_channel_status
		FROM channel_states s
		LEFT JOIN channels ec ON ec.channel_id = s.escrow_channel_id
		WHERE s.user_wallet = ?
			AND s.asset = ?
			AND s.user_sig IS NOT NULL
			AND s.node_sig IS NOT NULL
		ORDER BY s.epoch DESC, s.version DESC
		LIMIT 1
	`, wallet, asset).Scan(&result)

	if tx.Error != nil {
		return fmt.Errorf("failed to check ongoing escrow operation: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil
	}

	switch result.TransitionType {
	case core.TransitionTypeEscrowLock:
		return fmt.Errorf("escrow lock is still ongoing")

	case core.TransitionTypeMutualLock:
		return fmt.Errorf("mutual lock is still ongoing")

	case core.TransitionTypeEscrowDeposit:
		// Accept escrow channel at signed state version (finalize landed) or one
		// behind (signed N+1 finalize sits off-chain; on-chain still at INITIATE
		// version N). The one-behind branch is status-gated: Open is the protocol-
		// intended steady state until the purge queue fires, Closed covers post-
		// purge and post-finalize. Challenged is blocked because the on-chain
		// resolution is still racing — finalize may not land in time, the escrow
		// chain may settle at INITIATE, and replaying N+1 later could violate
		// engine invariants.
		// Compare via *v + 1 == StateVersion to avoid uint underflow when version is 0.
		if result.EscrowChannelVersion == nil {
			return fmt.Errorf("escrow deposit finalization is still ongoing")
		}
		onChain := *result.EscrowChannelVersion
		signed := result.StateVersion
		switch {
		case onChain == signed:
			// finalize already landed or signed state is older — allow
		case onChain+1 == signed:
			if result.EscrowChannelStatus == nil {
				return fmt.Errorf("escrow deposit finalization is still ongoing")
			}
			switch *result.EscrowChannelStatus {
			case core.ChannelStatusOpen, core.ChannelStatusClosed:
				// allow
			default:
				return fmt.Errorf("escrow deposit finalization is still ongoing")
			}
		default:
			return fmt.Errorf("escrow deposit finalization is still ongoing")
		}

	case core.TransitionTypeEscrowWithdraw:
		if result.EscrowChannelVersion == nil || result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow withdrawal finalization is still ongoing")
		}
	}

	return nil
}
