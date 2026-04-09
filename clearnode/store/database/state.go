package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// stateSelectColumns lists explicit columns for State queries to avoid collision with AutoMigrate columns.
// Used in raw SQL queries with UNION ALL pattern.
const stateSelectColumns = `s.id, s.asset, s.user_wallet, s.epoch, s.version,
	s.transition_type, s.transition_tx_id, s.transition_account_id, s.transition_amount,
	s.home_channel_id, s.escrow_channel_id,
	s.home_user_balance, s.home_user_net_flow, s.home_node_balance, s.home_node_net_flow,
	s.escrow_user_balance, s.escrow_user_net_flow, s.escrow_node_balance, s.escrow_node_net_flow,
	s.user_sig, s.node_sig, s.created_at`

// UserBalance represents aggregated user balance for an asset
type UserBalance struct {
	UserWallet       string          `gorm:"column:user_wallet;primaryKey;size:42"`
	Asset            string          `gorm:"column:asset;primaryKey;size:20"`
	Balance          decimal.Decimal `gorm:"column:balance;type:varchar(78);not null"`
	Enforced         decimal.Decimal `gorm:"column:enforced;type:varchar(78);not null;default:0"`
	HomeBlockchainID uint64          `gorm:"column:home_blockchain_id;not null;default:0"`
	CreatedAt        time.Time       `gorm:"column:created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at"`
}

// TableName specifies the table name for the UserBalance model
func (UserBalance) TableName() string {
	return "user_balances"
}

// State represents an immutable state in the system
// ID is deterministic: Hash(UserWallet, Asset, CycleIndex, Version)
type State struct {
	// ID is a 64-character deterministic hash
	ID         string `gorm:"column:id;primaryKey;size:64"`
	Asset      string `gorm:"column:asset;not null"`
	UserWallet string `gorm:"column:user_wallet;not null"`
	Epoch      uint64 `gorm:"column:epoch;not null"`
	Version    uint64 `gorm:"column:version;not null"`

	// Transition
	TransitionType      uint8           `gorm:"column:transition_type;not null"`
	TransitionTxID      string          `gorm:"column:transition_tx_id;size:66;not null"`
	TransitionAccountID string          `gorm:"column:transition_account_id;size:66;not null"`
	TransitionAmount    decimal.Decimal `gorm:"column:transition_amount;type:varchar(78);not null"`

	// Optional channel references
	HomeChannelID   *string `gorm:"column:home_channel_id"`
	EscrowChannelID *string `gorm:"column:escrow_channel_id"`

	// Home Channel balances and flows
	// Using decimal.Decimal for int256 values and int64 for flow values
	HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance;type:varchar(78)"`
	HomeUserNetFlow decimal.Decimal `gorm:"column:home_user_net_flow;default:0"`
	HomeNodeBalance decimal.Decimal `gorm:"column:home_node_balance;type:varchar(78)"`
	HomeNodeNetFlow decimal.Decimal `gorm:"column:home_node_net_flow;default:0"`

	// Escrow Channel balances and flows
	EscrowUserBalance decimal.Decimal `gorm:"column:escrow_user_balance;type:varchar(78)"`
	EscrowUserNetFlow decimal.Decimal `gorm:"column:escrow_user_net_flow;default:0"`
	EscrowNodeBalance decimal.Decimal `gorm:"column:escrow_node_balance;type:varchar(78)"`
	EscrowNodeNetFlow decimal.Decimal `gorm:"column:escrow_node_net_flow;default:0"`

	UserSig *string `gorm:"column:user_sig;type:text"`
	NodeSig *string `gorm:"column:node_sig;type:text"`

	// Read-only fields populated from JOINs with channels table
	HomeBlockchainID   *uint64 `gorm:"->;column:home_blockchain_id"`
	HomeTokenAddress   *string `gorm:"->;column:home_token_address"`
	EscrowBlockchainID *uint64 `gorm:"->;column:escrow_blockchain_id"`
	EscrowTokenAddress *string `gorm:"->;column:escrow_token_address"`

	CreatedAt time.Time
}

// TableName specifies the table name for the State model
func (State) TableName() string {
	return "channel_states"
}

// GetStateByID retrieves a state by its deterministic ID.
func (s *DBStore) GetStateByID(stateID string) (*core.State, error) {
	stateID = strings.ToLower(stateID)

	var dbState State
	err := s.db.Table("channel_states AS s").
		Select("s.*, hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address, ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address").
		Joins("LEFT JOIN channels AS hc ON s.home_channel_id = hc.channel_id").
		Joins("LEFT JOIN channels AS ec ON s.escrow_channel_id = ec.channel_id").
		Where("s.id = ?", stateID).First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get state by ID: %w", err)
	}

	return databaseStateToCore(&dbState)
}

// GetLastUserState retrieves the most recent state for a user's asset.
func (s *DBStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	var dbState State
	query := s.db.Table("channel_states AS s").
		Select("s.*, hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address, ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address").
		Joins("LEFT JOIN channels AS hc ON s.home_channel_id = hc.channel_id").
		Joins("LEFT JOIN channels AS ec ON s.escrow_channel_id = ec.channel_id").
		Where("s.user_wallet = ? AND s.asset = ?", strings.ToLower(wallet), asset)

	if signed {
		query = query.Where("s.user_sig IS NOT NULL AND s.node_sig IS NOT NULL")
	}

	err := query.Order("s.epoch DESC, s.version DESC").First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last user state: %w", err)
	}

	return databaseStateToCore(&dbState)
}

// StoreUserState persists a new user state to the database.
func (s *DBStore) StoreUserState(state core.State) error {
	dbState, err := coreStateToDB(&state)
	if err != nil {
		return fmt.Errorf("failed to encode transitions while creating a db state: %w", err)
	}

	if err := s.db.Create(dbState).Error; err != nil {
		return fmt.Errorf("failed to store user state: %w", err)
	}

	// Update user_balances table with the new balance (entry should already exist from LockUserState)
	wallet := strings.ToLower(state.UserWallet)
	balance := dbState.HomeUserBalance

	err = s.db.Model(&UserBalance{}).
		Where("user_wallet = ? AND asset = ?", wallet, state.Asset).
		Updates(map[string]interface{}{
			"balance":             balance,
			"home_blockchain_id":  state.HomeLedger.BlockchainID,
			"updated_at":          time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

// GetLastStateByChannelID retrieves the most recent state for a given channel.
// Uses UNION ALL of two indexed queries instead of OR for better performance.
func (s *DBStore) GetLastStateByChannelID(channelID string, signed bool) (*core.State, error) {
	channelID = strings.ToLower(channelID)

	signedFilter := ""
	if signed {
		signedFilter = "AND s.user_sig IS NOT NULL AND s.node_sig IS NOT NULL"
	}

	// Use UNION ALL to leverage separate indexes on home_channel_id and escrow_channel_id
	// Each branch returns its own best match, then we pick the overall best
	var state State
	err := s.db.Raw(fmt.Sprintf(`
		SELECT * FROM (
			SELECT %s,
			       hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address,
			       ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address
			FROM channel_states s
			LEFT JOIN channels hc ON s.home_channel_id = hc.channel_id
			LEFT JOIN channels ec ON s.escrow_channel_id = ec.channel_id
			WHERE s.home_channel_id = ? %s
			ORDER BY s.epoch DESC, s.version DESC
			LIMIT 1
		) home_result
		UNION ALL
		SELECT * FROM (
			SELECT %s,
			       hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address,
			       ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address
			FROM channel_states s
			LEFT JOIN channels hc ON s.home_channel_id = hc.channel_id
			LEFT JOIN channels ec ON s.escrow_channel_id = ec.channel_id
			WHERE s.escrow_channel_id = ? %s
			ORDER BY s.epoch DESC, s.version DESC
			LIMIT 1
		) escrow_result
		ORDER BY epoch DESC, version DESC
		LIMIT 1
	`, stateSelectColumns, signedFilter, stateSelectColumns, signedFilter), channelID, channelID).Scan(&state).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get last state by channel ID: %w", err)
	}

	// Check if no results were found (empty ID means no rows)
	if state.ID == "" {
		return nil, nil
	}

	return databaseStateToCore(&state)
}

// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
// Uses UNION ALL of two indexed queries instead of OR for better performance.
func (s *DBStore) GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error) {
	channelID = strings.ToLower(channelID)

	// Use UNION ALL to leverage separate indexes on home_channel_id and escrow_channel_id
	var state State
	err := s.db.Raw(fmt.Sprintf(`
		SELECT * FROM (
			SELECT %s,
			       hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address,
			       ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address
			FROM channel_states s
			LEFT JOIN channels hc ON s.home_channel_id = hc.channel_id
			LEFT JOIN channels ec ON s.escrow_channel_id = ec.channel_id
			WHERE s.home_channel_id = ? AND s.version = ?
			LIMIT 1
		) home_result
		UNION ALL
		SELECT * FROM (
			SELECT %s,
			       hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address,
			       ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address
			FROM channel_states s
			LEFT JOIN channels hc ON s.home_channel_id = hc.channel_id
			LEFT JOIN channels ec ON s.escrow_channel_id = ec.channel_id
			WHERE s.escrow_channel_id = ? AND s.version = ?
			LIMIT 1
		) escrow_result
		LIMIT 1
	`, stateSelectColumns, stateSelectColumns), channelID, version, channelID, version).Scan(&state).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get state by channel ID and version: %w", err)
	}

	// Check if no results were found (empty ID means no rows)
	if state.ID == "" {
		return nil, nil
	}

	return databaseStateToCore(&state)
}
