package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// ChannelSessionKeyStateV1 represents a channel session key state in the database.
type ChannelSessionKeyStateV1 struct {
	ID           string                     `gorm:"column:id;primaryKey"`
	UserAddress  string                     `gorm:"column:user_address;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:1"`
	SessionKey   string                     `gorm:"column:session_key;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:2"`
	Version      uint64                     `gorm:"column:version;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:3"`
	Assets       []ChannelSessionKeyAssetV1 `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	MetadataHash string                     `gorm:"column:metadata_hash;type:char(66);not null"`
	ExpiresAt    time.Time                  `gorm:"column:expires_at;not null"`
	UserSig      string                     `gorm:"column:user_sig;not null"`
	CreatedAt    time.Time
}

func (ChannelSessionKeyStateV1) TableName() string {
	return "channel_session_key_states_v1"
}

// ChannelSessionKeyAssetV1 links a channel session key state to an asset.
type ChannelSessionKeyAssetV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	Asset             string `gorm:"column:asset;not null;primaryKey;priority:2;index"`
}

func (ChannelSessionKeyAssetV1) TableName() string {
	return "channel_session_key_assets_v1"
}

// StoreChannelSessionKeyState stores a new channel session key state version.
func (s *DBStore) StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error {
	userAddress := strings.ToLower(state.UserAddress)
	sessionKey := strings.ToLower(state.SessionKey)

	id, err := core.GenerateSessionKeyStateIDV1(userAddress, sessionKey, state.Version)
	if err != nil {
		return fmt.Errorf("failed to generate session key state ID: %w", err)
	}

	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to compute metadata hash: %w", err)
	}

	dbState := ChannelSessionKeyStateV1{
		ID:           id,
		UserAddress:  userAddress,
		SessionKey:   sessionKey,
		Version:      state.Version,
		MetadataHash: strings.ToLower(metadataHash.Hex()),
		ExpiresAt:    state.ExpiresAt.UTC(),
		UserSig:      state.UserSig,
	}

	if err := s.db.Create(&dbState).Error; err != nil {
		return fmt.Errorf("failed to store channel session key state: %w", err)
	}

	if len(state.Assets) > 0 {
		assets := make([]ChannelSessionKeyAssetV1, len(state.Assets))
		for i, asset := range state.Assets {
			assets[i] = ChannelSessionKeyAssetV1{
				SessionKeyStateID: id,
				Asset:             strings.ToLower(asset),
			}
		}
		if err := s.db.Create(&assets).Error; err != nil {
			return fmt.Errorf("failed to store channel session key assets: %w", err)
		}
	}

	return nil
}

// GetLastChannelSessionKeyStates retrieves the latest channel session key states for a user with optional filtering.
// Returns only the highest-version row per session key that has not expired.
func (s *DBStore) GetLastChannelSessionKeyStates(wallet string, sessionKey *string) ([]core.ChannelSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)

	subQuery := s.db.Model(&ChannelSessionKeyStateV1{}).
		Select("user_address, session_key, MAX(version) as max_version").
		Where("user_address = ?", wallet).
		Group("user_address, session_key")

	if sessionKey != nil && *sessionKey != "" {
		subQuery = subQuery.Where("session_key = ?", strings.ToLower(*sessionKey))
	}

	query := s.db.
		Joins("JOIN (?) AS latest ON channel_session_key_states_v1.user_address = latest.user_address AND channel_session_key_states_v1.session_key = latest.session_key AND channel_session_key_states_v1.version = latest.max_version", subQuery).
		Preload("Assets").
		Order("channel_session_key_states_v1.created_at DESC")

	var dbStates []ChannelSessionKeyStateV1
	if err := query.Find(&dbStates).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []core.ChannelSessionKeyStateV1{}, nil
		}
		return nil, fmt.Errorf("failed to get channel session key states: %w", err)
	}

	states := make([]core.ChannelSessionKeyStateV1, len(dbStates))
	for i, dbState := range dbStates {
		states[i] = dbChannelSessionKeyStateToCore(&dbState)
	}

	return states, nil
}

// GetLastChannelSessionKeyVersion returns the latest version of a channel session key state.
// Returns 0 if no state exists.
func (s *DBStore) GetLastChannelSessionKeyVersion(wallet, sessionKey string) (uint64, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var result struct {
		Version uint64
	}
	err := s.db.Model(&ChannelSessionKeyStateV1{}).
		Select("version").
		Where("user_address = ? AND session_key = ?", wallet, sessionKey).
		Order("version DESC").
		Take(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to check channel session key state: %w", err)
	}

	return result.Version, nil
}

// ValidateChannelSessionKeyForAsset checks in a single query that:
// - a session key state exists for the (wallet, sessionKey) pair,
// - it is the latest version,
// - it is not expired,
// - the asset is in the allowed list,
// - the metadata hash matches.
func (s *DBStore) ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)
	asset = strings.ToLower(asset)
	metadataHash = strings.ToLower(metadataHash)

	now := time.Now().UTC()

	maxVersionSubQ := s.db.Model(&ChannelSessionKeyStateV1{}).
		Select("MAX(version)").
		Where("user_address = ? AND session_key = ?", wallet, sessionKey)

	var count int64
	err := s.db.Model(&ChannelSessionKeyStateV1{}).
		Where("user_address = ? AND session_key = ? AND expires_at > ? AND metadata_hash = ? AND version = (?)",
			wallet, sessionKey, now, metadataHash, maxVersionSubQ).
		Joins("JOIN channel_session_key_assets_v1 ON channel_session_key_assets_v1.session_key_state_id = channel_session_key_states_v1.id AND channel_session_key_assets_v1.asset = ?", asset).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to validate session key for asset: %w", err)
	}

	return count > 0, nil
}

func dbChannelSessionKeyStateToCore(dbState *ChannelSessionKeyStateV1) core.ChannelSessionKeyStateV1 {
	assets := make([]string, len(dbState.Assets))
	for i, a := range dbState.Assets {
		assets[i] = a.Asset
	}

	return core.ChannelSessionKeyStateV1{
		UserAddress: dbState.UserAddress,
		SessionKey:  dbState.SessionKey,
		Version:     dbState.Version,
		Assets:      assets,
		ExpiresAt:   dbState.ExpiresAt,
		UserSig:     dbState.UserSig,
	}
}
