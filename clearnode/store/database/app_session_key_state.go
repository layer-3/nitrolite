package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"gorm.io/gorm"
)

// AppSessionKeyStateV1 represents a session key state in the database.
// ID is Hash(user_address + session_key + version).
type AppSessionKeyStateV1 struct {
	ID             string                        `gorm:"column:id;primaryKey"`
	UserAddress    string                        `gorm:"column:user_address;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:1"`
	SessionKey     string                        `gorm:"column:session_key;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:2"`
	Version        uint64                        `gorm:"column:version;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:3"`
	ApplicationIDs []AppSessionKeyApplicationV1  `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	AppSessionIDs  []AppSessionKeyAppSessionIDV1 `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	ExpiresAt      time.Time                     `gorm:"column:expires_at;not null"`
	UserSig        string                        `gorm:"column:user_sig;not null"`
	CreatedAt      time.Time
}

func (AppSessionKeyStateV1) TableName() string {
	return "app_session_key_states_v1"
}

// SessionKeyApplicationV1 links a session key state to an application ID.
type AppSessionKeyApplicationV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	ApplicationID     string `gorm:"column:application_id;not null;primaryKey;priority:2;index"`
}

func (AppSessionKeyApplicationV1) TableName() string {
	return "app_session_key_applications_v1"
}

// AppSessionKeyAppSessionIDV1 links a session key state to an app session ID.
type AppSessionKeyAppSessionIDV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	AppSessionID      string `gorm:"column:app_session_id;not null;primaryKey;priority:2;index"`
}

func (AppSessionKeyAppSessionIDV1) TableName() string {
	return "app_session_key_app_sessions_v1"
}

// StoreAppSessionKeyState stores a new session key state version.
func (s *DBStore) StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error {
	userAddress := strings.ToLower(state.UserAddress)
	sessionKey := strings.ToLower(state.SessionKey)

	id, err := app.GenerateSessionKeyStateIDV1(userAddress, sessionKey, state.Version)
	if err != nil {
		return fmt.Errorf("failed to generate session key state ID: %w", err)
	}

	dbState := AppSessionKeyStateV1{
		ID:          id,
		UserAddress: userAddress,
		SessionKey:  sessionKey,
		Version:     state.Version,
		ExpiresAt:   state.ExpiresAt.UTC(),
		UserSig:     state.UserSig,
	}

	if err := s.db.Create(&dbState).Error; err != nil {
		return fmt.Errorf("failed to store session key state: %w", err)
	}

	if len(state.ApplicationIDs) > 0 {
		applicationIDs := make([]AppSessionKeyApplicationV1, len(state.ApplicationIDs))
		for i, appID := range state.ApplicationIDs {
			applicationIDs[i] = AppSessionKeyApplicationV1{
				SessionKeyStateID: id,
				ApplicationID:     strings.ToLower(appID),
			}
		}
		if err := s.db.Create(&applicationIDs).Error; err != nil {
			return fmt.Errorf("failed to store application IDs: %w", err)
		}
	}

	if len(state.AppSessionIDs) > 0 {
		appSessionIDs := make([]AppSessionKeyAppSessionIDV1, len(state.AppSessionIDs))
		for i, sessID := range state.AppSessionIDs {
			appSessionIDs[i] = AppSessionKeyAppSessionIDV1{
				SessionKeyStateID: id,
				AppSessionID:      strings.ToLower(sessID),
			}
		}
		if err := s.db.Create(&appSessionIDs).Error; err != nil {
			return fmt.Errorf("failed to store app session IDs: %w", err)
		}
	}

	return nil
}

// GetLastAppSessionKeyStates retrieves the latest session key states for a user with optional filtering.
// Returns only the highest-version row per session key that has not expired.
func (s *DBStore) GetLastAppSessionKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)

	subQuery := s.db.Model(&AppSessionKeyStateV1{}).
		Select("user_address, session_key, MAX(version) as max_version").
		Where("user_address = ?", wallet).
		Group("user_address, session_key")

	if sessionKey != nil && *sessionKey != "" {
		subQuery = subQuery.Where("session_key = ?", strings.ToLower(*sessionKey))
	}

	query := s.db.
		Joins("JOIN (?) AS latest ON app_session_key_states_v1.user_address = latest.user_address AND app_session_key_states_v1.session_key = latest.session_key AND app_session_key_states_v1.version = latest.max_version", subQuery).
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		Order("app_session_key_states_v1.created_at DESC")

	var dbStates []AppSessionKeyStateV1
	if err := query.Find(&dbStates).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []app.AppSessionKeyStateV1{}, nil
		}
		return nil, fmt.Errorf("failed to get session key states: %w", err)
	}

	states := make([]app.AppSessionKeyStateV1, len(dbStates))
	for i, dbState := range dbStates {
		states[i] = dbSessionKeyStateToCore(&dbState)
	}

	return states, nil
}

// GetLastAppSessionKeyVersion returns the latest version of a session key state for a user.
// Returns 0 if no state exists.
func (s *DBStore) GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var result struct {
		Version uint64
	}
	err := s.db.Model(&AppSessionKeyStateV1{}).
		Select("version").
		Where("user_address = ? AND session_key = ?", wallet, sessionKey).
		Order("version DESC").
		Take(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to check session key state: %w", err)
	}

	return result.Version, nil
}

// GetLastAppSessionKeyState retrieves the latest version of a specific session key for a user.
// A newer version always supersedes older ones, even if expired.
// Returns nil if no state exists.
func (s *DBStore) GetLastAppSessionKeyState(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var dbState AppSessionKeyStateV1
	err := s.db.
		Where("user_address = ? AND session_key = ?", wallet, sessionKey).
		Order("version DESC").
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		First(&dbState).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest session key state: %w", err)
	}

	result := dbSessionKeyStateToCore(&dbState)
	return &result, nil
}

// GetAppSessionKeyOwner returns the user_address that owns the given session key
// authorized for the specified app session ID. Only the latest-version, non-expired key
// with matching permissions is considered. A newer version always supersedes older ones.
func (s *DBStore) GetAppSessionKeyOwner(sessionKey, appSessionId string) (string, error) {
	sessionKey = strings.ToLower(sessionKey)
	appSessionId = strings.ToLower(appSessionId)

	// Subquery to get the application ID from the app session
	appSubQuery := s.db.Model(&AppSessionV1{}).Select("application_id").Where("id = ?", appSessionId)

	maxVersionSubQ := s.db.Model(&AppSessionKeyStateV1{}).
		Select("MAX(version)").
		Where("session_key = ?", sessionKey)

	var dbState AppSessionKeyStateV1
	err := s.db.
		Joins("LEFT JOIN app_session_key_app_sessions_v1 ON app_session_key_app_sessions_v1.session_key_state_id = app_session_key_states_v1.id").
		Joins("LEFT JOIN app_session_key_applications_v1 ON app_session_key_applications_v1.session_key_state_id = app_session_key_states_v1.id").
		Where("app_session_key_states_v1.session_key = ? AND app_session_key_states_v1.version = (?) AND app_session_key_states_v1.expires_at > ? AND (app_session_key_app_sessions_v1.app_session_id = ? OR app_session_key_applications_v1.application_id = (?))",
			sessionKey, maxVersionSubQ, time.Now().UTC(), appSessionId, appSubQuery).
		First(&dbState).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("no active session key found for key %s and app session %s", sessionKey, appSessionId)
		}
		return "", fmt.Errorf("failed to get session key owner: %w", err)
	}

	return dbState.UserAddress, nil
}

func dbSessionKeyStateToCore(dbState *AppSessionKeyStateV1) app.AppSessionKeyStateV1 {
	applicationIDs := make([]string, len(dbState.ApplicationIDs))
	for i, a := range dbState.ApplicationIDs {
		applicationIDs[i] = a.ApplicationID
	}

	appSessionIDs := make([]string, len(dbState.AppSessionIDs))
	for i, a := range dbState.AppSessionIDs {
		appSessionIDs[i] = a.AppSessionID
	}

	return app.AppSessionKeyStateV1{
		UserAddress:    dbState.UserAddress,
		SessionKey:     dbState.SessionKey,
		Version:        dbState.Version,
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      dbState.ExpiresAt,
		UserSig:        dbState.UserSig,
	}
}
