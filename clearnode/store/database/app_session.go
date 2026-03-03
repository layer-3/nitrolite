package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// AppSessionV1 represents a virtual payment application session between participants
type AppSessionV1 struct {
	ID            string               `gorm:"primaryKey"`
	ApplicationID string               `gorm:"column:application_id;not null"`
	Nonce         uint64               `gorm:"column:nonce;not null"`
	Participants  []AppParticipantV1   `gorm:"foreignKey:AppSessionID;references:ID"`
	SessionData   string               `gorm:"column:session_data;type:text;not null"`
	Quorum        uint8                `gorm:"column:quorum;default:100"`
	Version       uint64               `gorm:"column:version;default:1"`
	Status        app.AppSessionStatus `gorm:"column:status;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (AppSessionV1) TableName() string {
	return "app_sessions_v1"
}

// AppParticipantV1 represents the definition for an app participant.
type AppParticipantV1 struct {
	AppSessionID    string `gorm:"column:app_session_id;not null;primaryKey;priority:1"`
	WalletAddress   string `gorm:"column:wallet_address;not null;primaryKey;priority:2"`
	SignatureWeight uint8  `gorm:"column:signature_weight;not null"`
}

func (AppParticipantV1) TableName() string {
	return "app_session_participants_v1"
}

// CreateAppSession initializes a new application session.
func (s *DBStore) CreateAppSession(session app.AppSessionV1) error {
	participants := make([]AppParticipantV1, len(session.Participants))
	for i, p := range session.Participants {
		participants[i] = AppParticipantV1{
			AppSessionID:    strings.ToLower(session.SessionID),
			WalletAddress:   strings.ToLower(p.WalletAddress),
			SignatureWeight: p.SignatureWeight,
		}
	}

	dbSession := AppSessionV1{
		ID:            strings.ToLower(session.SessionID),
		ApplicationID: session.Application,
		Nonce:         session.Nonce,
		Participants:  participants,
		SessionData:   session.SessionData,
		Quorum:        session.Quorum,
		Version:       session.Version,
		Status:        session.Status,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
	}

	if err := s.db.Create(&dbSession).Error; err != nil {
		return fmt.Errorf("failed to create app session: %w", err)
	}

	return nil
}

// GetAppSession retrieves a specific session by ID.
func (s *DBStore) GetAppSession(sessionID string) (*app.AppSessionV1, error) {
	sessionID = strings.ToLower(sessionID)

	var dbSession AppSessionV1
	err := s.db.Where("id = ?", sessionID).
		Preload("Participants").
		Order("nonce DESC").
		First(&dbSession).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get app session: %w", err)
	}

	return databaseAppSessionToCore(&dbSession), nil
}

// GetAppSessions retrieves filtered sessions with pagination.
func (s *DBStore) GetAppSessions(appSessionID *string, participant *string, status app.AppSessionStatus, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error) {
	query := s.db.Model(&AppSessionV1{})

	if appSessionID != nil && *appSessionID != "" {
		query = query.Where("id = ?", strings.ToLower(*appSessionID))
	}

	if participant != nil && *participant != "" {
		subQuery := s.db.Model(&AppSessionV1{}).
			Select("app_sessions_v1.id").
			Joins("JOIN app_session_participants_v1 ON app_sessions_v1.id = app_session_participants_v1.app_session_id").
			Where("app_session_participants_v1.wallet_address = ?", strings.ToLower(*participant))
		query = query.Where("id IN (?)", subQuery)
	}

	if status != app.AppSessionStatusVoid {
		query = query.Where("status = ?", status)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to count app sessions: %w", err)
	}

	offset, limit := pagination.GetOffsetAndLimit(DefaultLimit, MaxLimit)

	query = query.Preload("Participants").Order("created_at DESC").Offset(int(offset)).Limit(int(limit))

	var dbSessions []AppSessionV1
	if err := query.Find(&dbSessions).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get app sessions: %w", err)
	}

	sessions := make([]app.AppSessionV1, len(dbSessions))
	for i, dbSession := range dbSessions {
		sessions[i] = *databaseAppSessionToCore(&dbSession)
	}

	metadata := calculatePaginationMetadata(totalCount, offset, limit)

	return sessions, metadata, nil
}

// AppSessionCount holds the result of a COUNT() GROUP BY query on app sessions.
type AppSessionCount struct {
	Application string               `gorm:"column:application_id"`
	Status      app.AppSessionStatus `gorm:"column:status"`
	Count       uint64               `gorm:"column:count"`
}

// CountAppSessionsByStatus returns app session counts grouped by (application, status).
func (s *DBStore) CountAppSessionsByStatus() ([]AppSessionCount, error) {
	var results []AppSessionCount
	err := s.db.Model(&AppSessionV1{}).
		Select("application_id, status, COUNT(id) as count").
		Group("application_id, status").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count app sessions: %w", err)
	}
	return results, nil
}

// UpdateAppSession updates existing session data with optimistic locking.
func (s *DBStore) UpdateAppSession(session app.AppSessionV1) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		sessionID := strings.ToLower(session.SessionID)
		expectedVersion := session.Version - 1

		updates := map[string]interface{}{
			"session_data": session.SessionData,
			"version":      session.Version,
			"status":       session.Status,
			"updated_at":   time.Now(),
		}

		// Use optimistic locking: only update if version matches expected
		result := tx.Model(&AppSessionV1{}).
			Where("id = ? AND version = ?", sessionID, expectedVersion).
			Updates(updates)

		if result.Error != nil {
			return fmt.Errorf("failed to update app session: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("concurrent modification detected for session %s", sessionID)
		}

		return nil
	})
}
