package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// Channel represents a state channel between participants
type Channel struct {
	ChannelID             string             `gorm:"column:channel_id;primaryKey;"`
	UserWallet            string             `gorm:"column:user_wallet;not null"`
	Asset                 string             `gorm:"column:asset;not null"`
	Type                  core.ChannelType   `gorm:"column:type;not null"`
	BlockchainID          uint64             `gorm:"column:blockchain_id;not null"`
	Token                 string             `gorm:"column:token;not null"`
	ChallengeDuration     uint32             `gorm:"column:challenge_duration;not null"`
	ChallengeExpiresAt    *time.Time         `gorm:"column:challenge_expires_at;default:null"`
	Nonce                 uint64             `gorm:"column:nonce;not null;"`
	ApprovedSigValidators string             `gorm:"column:approved_sig_validators;not null;"`
	Status                core.ChannelStatus `gorm:"column:status;not null;"`
	StateVersion          uint64             `gorm:"column:state_version;not null;"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel entity in the database.
func (s *DBStore) CreateChannel(channel core.Channel) error {
	dbChannel := Channel{
		ChannelID:             strings.ToLower(channel.ChannelID),
		UserWallet:            strings.ToLower(channel.UserWallet),
		Asset:                 strings.ToLower(channel.Asset),
		Type:                  channel.Type,
		BlockchainID:          channel.BlockchainID,
		Token:                 strings.ToLower(channel.TokenAddress),
		ChallengeDuration:     channel.ChallengeDuration,
		ChallengeExpiresAt:    channel.ChallengeExpiresAt,
		Nonce:                 channel.Nonce,
		ApprovedSigValidators: channel.ApprovedSigValidators,
		Status:                channel.Status,
		StateVersion:          channel.StateVersion,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.db.Create(&dbChannel).Error; err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	return nil
}

// GetChannelByID retrieves a channel by its unique identifier.
func (s *DBStore) GetChannelByID(channelID string) (*core.Channel, error) {
	channelID = strings.ToLower(channelID)

	// Ensure channelID has 0x prefix
	if !strings.HasPrefix(channelID, "0x") {
		channelID = "0x" + channelID
	}

	var dbChannel Channel
	if err := s.db.Where("channel_id = ?", channelID).First(&dbChannel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get channel by ID: %w", err)
	}

	return databaseChannelToCore(&dbChannel), nil
}

// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
// "Active" means the node has co-signed the channel definition (status Void or Open) — it
// does NOT guarantee the channel has been materialized onchain. Callers requiring onchain
// materialization (e.g., cross-chain escrow operations) must additionally check that
// Status == ChannelStatusOpen.
func (s *DBStore) GetActiveHomeChannel(wallet, asset string) (*core.Channel, error) {
	var dbChannel Channel
	err := s.db.
		Where("user_wallet = ? AND asset = ?", strings.ToLower(wallet), strings.ToLower(asset)).
		Where("status <= ? AND type = ?", core.ChannelStatusOpen, core.ChannelTypeHome).
		First(&dbChannel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active home channel: %w", err)
	}

	return databaseChannelToCore(&dbChannel), nil
}

// HasNonClosedHomeChannel returns true if any home channel for (wallet, asset) has a
// status other than Closed, meaning a channel lifecycle is still in progress.
func (s *DBStore) HasNonClosedHomeChannel(wallet, asset string) (bool, error) {
	var count int64
	err := s.db.Model(&Channel{}).
		Where("user_wallet = ? AND asset = ? AND type = ? AND status != ?",
			strings.ToLower(wallet), strings.ToLower(asset),
			core.ChannelTypeHome, core.ChannelStatusClosed).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check non-closed home channel: %w", err)
	}
	return count > 0, nil
}

// CheckActiveChannel verifies if a user has an active home channel for the given asset
// and returns its approved signature validators along with the channel status.
// "Active" includes Void (DB-only, awaiting onchain confirmation) and Open (materialized
// onchain). This is intentional: non-escrow offchain transitions (transfers, etc.) are
// permitted before onchain confirmation lands. Callers operating on cross-chain escrow
// flows that depend on onchain home-channel materialization must check that the returned
// status is ChannelStatusOpen.
//
// A nil status pointer means no active channel was found.
func (s *DBStore) CheckActiveChannel(wallet, asset string) (string, *core.ChannelStatus, error) {
	var row struct {
		ApprovedSigValidators string             `gorm:"column:approved_sig_validators"`
		Status                core.ChannelStatus `gorm:"column:status"`
	}
	result := s.db.Raw(`
		SELECT approved_sig_validators, status
		FROM channels
		WHERE user_wallet = ?
			AND asset = ?
			AND status <= ?
			AND type = ?
		LIMIT 1
	`, strings.ToLower(wallet), strings.ToLower(asset), core.ChannelStatusOpen, core.ChannelTypeHome).Scan(&row)
	if result.Error != nil {
		return "", nil, fmt.Errorf("failed to check active channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", nil, nil
	}

	return row.ApprovedSigValidators, &row.Status, nil
}

// GetUserChannels retrieves all channels for a user with optional status, asset, and type filters.
func (s *DBStore) GetUserChannels(wallet string, status *core.ChannelStatus, asset *string, channelType *core.ChannelType, limit, offset uint32) ([]core.Channel, uint32, error) {
	query := s.db.Model(&Channel{}).Where("user_wallet = ?", strings.ToLower(wallet))

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if asset != nil && *asset != "" {
		query = query.Where("asset = ?", strings.ToLower(*asset))
	}

	if channelType != nil {
		query = query.Where("type = ?", *channelType)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user channels: %w", err)
	}

	var dbChannels []Channel
	if err := query.Order("created_at DESC").Limit(int(limit)).Offset(int(offset)).Find(&dbChannels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user channels: %w", err)
	}

	channels := make([]core.Channel, len(dbChannels))
	for i := range dbChannels {
		channels[i] = *databaseChannelToCore(&dbChannels[i])
	}

	return channels, uint32(totalCount), nil
}

// ChannelCount holds the result of a COUNT() GROUP BY query on channels.
type ChannelCount struct {
	Asset  string             `gorm:"column:asset"`
	Status core.ChannelStatus `gorm:"column:status"`
	Count  uint64             `gorm:"column:count"`
}

// GetChannelsCountByLabels returns current channel counts grouped by asset and status.
func (s *DBStore) GetChannelsCountByLabels() ([]ChannelCount, error) {
	var results []ChannelCount
	err := s.db.Raw(`
		SELECT asset,
		       status,
		       COUNT(channel_id) AS count
		FROM channels
		GROUP BY asset, status
	`).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get channel counts: %w", err)
	}

	return results, nil
}

// UpdateChannel persists changes to a channel's metadata (status, version, etc).
func (s *DBStore) UpdateChannel(channel core.Channel) error {
	updates := map[string]interface{}{
		"status":               channel.Status,
		"state_version":        channel.StateVersion,
		"blockchain_id":        channel.BlockchainID,
		"token":                strings.ToLower(channel.TokenAddress),
		"nonce":                channel.Nonce,
		"challenge_expires_at": channel.ChallengeExpiresAt,
		"updated_at":           time.Now(),
	}

	if err := s.db.Model(&Channel{}).Where("channel_id = ?", strings.ToLower(channel.ChannelID)).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}
