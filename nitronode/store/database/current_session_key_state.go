package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SessionKeyKind discriminates the two session-key flavors stored in
// current_session_key_states_v1. Stored as SMALLINT in the DB.
type SessionKeyKind uint8

const (
	SessionKeyKindChannel    SessionKeyKind = 1
	SessionKeyKindAppSession SessionKeyKind = 2
)

// CurrentSessionKeyStateV1 is the latest-version pointer per (user_address, session_key, kind).
// Reads of get_last_key_states JOIN this table to the corresponding history table
// (channel_session_key_states_v1 or app_session_key_states_v1) on
// (user_address, session_key, version), bounding per-request DB work to O(distinct keys).
type CurrentSessionKeyStateV1 struct {
	UserAddress string         `gorm:"column:user_address;primaryKey;size:42"`
	SessionKey  string         `gorm:"column:session_key;primaryKey;size:42"`
	Kind        SessionKeyKind `gorm:"column:kind;primaryKey;type:smallint"`
	Version     uint64         `gorm:"column:version;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (CurrentSessionKeyStateV1) TableName() string {
	return "current_session_key_states_v1"
}

// upsertCurrentSessionKeyState writes the latest version for (user_address, session_key, kind).
// EXCLUDED.version > version guard prevents an out-of-order writer from regressing the pointer.
func upsertCurrentSessionKeyState(tx *gorm.DB, userAddress, sessionKey string, kind SessionKeyKind, version uint64) error {
	row := CurrentSessionKeyStateV1{
		UserAddress: strings.ToLower(userAddress),
		SessionKey:  strings.ToLower(sessionKey),
		Kind:        kind,
		Version:     version,
		UpdatedAt:   time.Now().UTC(),
	}

	res := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_address"},
			{Name: "session_key"},
			{Name: "kind"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"version":    gorm.Expr("EXCLUDED.version"),
			"updated_at": gorm.Expr("EXCLUDED.updated_at"),
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			gorm.Expr("EXCLUDED.version > current_session_key_states_v1.version"),
		}},
	}).Create(&row)

	if err := res.Error; err != nil {
		return fmt.Errorf("failed to upsert current session key state: %w", err)
	}
	return nil
}

// LockSessionKeyState ensures a pointer row exists for (user, session_key, kind) and locks it
// for the duration of the surrounding transaction. Returns the current version (0 if newly
// created). Mirrors LockUserState. On non-postgres dialects, falls back to read-without-lock.
func (s *DBStore) LockSessionKeyState(userAddress, sessionKey string, kind SessionKeyKind) (uint64, error) {
	userAddress = strings.ToLower(userAddress)
	sessionKey = strings.ToLower(sessionKey)

	if s.db.Dialector.Name() == "postgres" {
		seed := CurrentSessionKeyStateV1{
			UserAddress: userAddress,
			SessionKey:  sessionKey,
			Kind:        kind,
			Version:     0,
			UpdatedAt:   time.Now().UTC(),
		}
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
			return 0, fmt.Errorf("failed to ensure current session key state row exists: %w", err)
		}

		var locked CurrentSessionKeyStateV1
		err := s.db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_address = ? AND session_key = ? AND kind = ?", userAddress, sessionKey, kind).
			First(&locked).Error
		if err != nil {
			return 0, fmt.Errorf("failed to lock current session key state: %w", err)
		}
		return locked.Version, nil
	}

	var existing CurrentSessionKeyStateV1
	err := s.db.Where("user_address = ? AND session_key = ? AND kind = ?", userAddress, sessionKey, kind).
		First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			seed := CurrentSessionKeyStateV1{
				UserAddress: userAddress,
				SessionKey:  sessionKey,
				Kind:        kind,
				Version:     0,
				UpdatedAt:   time.Now().UTC(),
			}
			if err := s.db.Create(&seed).Error; err != nil {
				return 0, fmt.Errorf("failed to create current session key state: %w", err)
			}
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read current session key state: %w", err)
	}
	return existing.Version, nil
}

// CountSessionKeysForUser returns the number of distinct session keys recorded for the wallet
// in the pointer table, across both kinds. Drives the per-user cap at submit time.
// Rows seeded by LockSessionKeyState (version=0) are excluded so that a failed-cap rejection
// does not itself leave a phantom row counted toward the cap.
func (s *DBStore) CountSessionKeysForUser(userAddress string) (uint32, error) {
	userAddress = strings.ToLower(userAddress)

	var count int64
	err := s.db.Model(&CurrentSessionKeyStateV1{}).
		Where("user_address = ? AND version > 0", userAddress).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count session keys for user: %w", err)
	}
	return uint32(count), nil
}
