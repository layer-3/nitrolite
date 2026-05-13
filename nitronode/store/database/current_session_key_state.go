package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrSessionKeyNotAllowed is returned by LockSessionKeyState when the session key for the
// requested kind is bound to a wallet other than the submitter. The message is intentionally
// generic so the API does not confirm whether a given session_key is registered elsewhere.
var ErrSessionKeyNotAllowed = errors.New("session key not allowed")

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
//
// The uniqueIndex on (session_key, kind) mirrors the postgres constraint added by
// 20260508000000_session_key_ownership_constraints.sql so AutoMigrate (sqlite) enforces the
// same one-owner-per-key invariant that LockSessionKeyState relies on. The index name
// matches the postgres constraint name so both paths converge on a single source of truth.
type CurrentSessionKeyStateV1 struct {
	UserAddress string         `gorm:"column:user_address;primaryKey;size:42"`
	SessionKey  string         `gorm:"column:session_key;primaryKey;size:42;uniqueIndex:current_session_key_states_v1_key_kind_uniq,priority:1"`
	Kind        SessionKeyKind `gorm:"column:kind;primaryKey;type:smallint;uniqueIndex:current_session_key_states_v1_key_kind_uniq,priority:2"`
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

// LockSessionKeyState seeds the pointer row for (userAddress, session_key, kind) if absent
// and locks the (session_key, kind) row for the surrounding transaction. Returns the latest
// stored version for the caller's row, or ErrSessionKeyNotAllowed if the key is bound to a
// different wallet for this kind.
//
// The (session_key, kind) unique constraint guarantees there is at most one pointer row per
// (session_key, kind), so the SELECT ... FOR UPDATE that follows the no-op-on-conflict insert
// always converges on the same physical row regardless of who tried to seed first. A foreign
// wallet that races a legitimate owner ends up reading the legitimate owner back from the
// locked row and is rejected here, without parsing constraint-violation errors at write time.
//
// SELECT ... FOR UPDATE is postgres-only; on sqlite the locking clause is skipped and the
// surrounding transaction provides the necessary ordering for the in-process test setup.
//
// Seed-row permanence: the version=0 row written below is intentionally never deleted on
// failure paths (sig validation, version mismatch, cap exceeded, mid-tx errors). Once a wallet
// has staked a claim on (session_key, kind), no other wallet can take it for that kind — the
// seed is the ownership reservation, not a transient placeholder. CountSessionKeysForUser
// excludes version=0 rows so the per-user cap is unaffected, but the (session_key, kind)
// ownership bind is permanent by design.
func (s *DBStore) LockSessionKeyState(userAddress, sessionKey string, kind SessionKeyKind) (uint64, error) {
	userAddress = strings.ToLower(userAddress)
	sessionKey = strings.ToLower(sessionKey)

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

	query := s.db.Where("session_key = ? AND kind = ?", sessionKey, kind)
	if s.db.Dialector.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var locked CurrentSessionKeyStateV1
	err := query.First(&locked).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Must not happen: the seed insert above either created our row or no-op'd on
			// an existing one, so a SELECT keyed on (session_key, kind) must hit a row.
			// Treat as a hard error rather than falling through as unowned — silently
			// returning version 0 here would let a submit bypass ownership enforcement.
			return 0, fmt.Errorf("session key pointer row missing after seed insert for (session_key=%s, kind=%d)", sessionKey, kind)
		}
		return 0, fmt.Errorf("failed to lock current session key state: %w", err)
	}

	if !strings.EqualFold(locked.UserAddress, userAddress) {
		return 0, ErrSessionKeyNotAllowed
	}
	return locked.Version, nil
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
