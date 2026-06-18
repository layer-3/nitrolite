package database

import (
	"errors"
	"fmt"
	"math"
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
// Seed-row permanence: the version=0 row written below is part of the caller's transaction,
// so it persists only when that transaction commits. Failure paths that abort the tx (version
// mismatch, cap exceeded, mid-tx errors) roll the seed back with everything else, and callers
// must guard against committing a seed for an unauthorized claim — e.g. the submit handlers
// reject a revoke at version 1, so a wallet cannot stake a claim on (session_key, kind) without
// a prior delegation it proved possession of. Once a submit does commit, the ownership bind is
// permanent: no other wallet can take that (session_key, kind). CountSessionKeysForUser excludes
// version=0 rows so the per-user cap is unaffected.
//
// When locked.Version > 0, the matching history row's expires_at is also returned so callers
// can distinguish a reactivation (prev inactive → submitted active) from a rotation/update
// (prev still active) and re-run the per-user cap on the reactivation path. When
// locked.Version == 0 there is no history yet, so a zero time.Time is returned.
func (s *DBStore) LockSessionKeyState(userAddress, sessionKey string, kind SessionKeyKind) (uint64, time.Time, error) {
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
		return 0, time.Time{}, fmt.Errorf("failed to ensure current session key state row exists: %w", err)
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
			return 0, time.Time{}, fmt.Errorf("session key pointer row missing after seed insert for (session_key=%s, kind=%d)", sessionKey, kind)
		}
		return 0, time.Time{}, fmt.Errorf("failed to lock current session key state: %w", err)
	}

	if !strings.EqualFold(locked.UserAddress, userAddress) {
		return 0, time.Time{}, ErrSessionKeyNotAllowed
	}

	if locked.Version == 0 {
		return 0, time.Time{}, nil
	}

	expiresAt, err := s.fetchLatestSessionKeyExpiresAt(userAddress, sessionKey, locked.Version, kind)
	if err != nil {
		return 0, time.Time{}, err
	}
	return locked.Version, expiresAt, nil
}

// fetchLatestSessionKeyExpiresAt returns the expires_at of the history row at
// (user_address, session_key, version) for the given kind. The pointer table guarantees
// at most one such row per (user, key, kind, version) so a missing history row is a hard
// inconsistency and surfaces as an error rather than a zero expiry.
func (s *DBStore) fetchLatestSessionKeyExpiresAt(userAddress, sessionKey string, version uint64, kind SessionKeyKind) (time.Time, error) {
	type expiryRow struct {
		ExpiresAt time.Time `gorm:"column:expires_at"`
	}

	var table string
	switch kind {
	case SessionKeyKindAppSession:
		table = "app_session_key_states_v1"
	case SessionKeyKindChannel:
		table = "channel_session_key_states_v1"
	default:
		return time.Time{}, fmt.Errorf("unknown session key kind: %d", kind)
	}

	var row expiryRow
	err := s.db.Table(table).
		Select("expires_at").
		Where("user_address = ? AND session_key = ? AND version = ?", userAddress, sessionKey, version).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return time.Time{}, fmt.Errorf("session key history row missing for (user=%s, session_key=%s, version=%d, kind=%d)", userAddress, sessionKey, version, kind)
		}
		return time.Time{}, fmt.Errorf("failed to load latest session key expires_at: %w", err)
	}
	return row.ExpiresAt, nil
}

// CountSessionKeysForUser returns the number of distinct active session keys recorded for the
// wallet in the pointer table, across both kinds. Drives the per-user cap at submit time.
// Rows seeded by LockSessionKeyState (version=0) are excluded so that a failed-cap rejection
// does not itself leave a phantom row counted toward the cap. Revoked or naturally expired
// keys (expires_at <= now in the underlying history row) are also excluded so that a revoke
// frees the slot. A single now is bound for both kind branches so the count is internally
// consistent.
func (s *DBStore) CountSessionKeysForUser(userAddress string) (uint32, error) {
	userAddress = strings.ToLower(userAddress)
	now := time.Now().UTC()

	var channelCount int64
	err := s.db.Table("current_session_key_states_v1 AS c").
		Joins("JOIN channel_session_key_states_v1 h ON h.user_address = c.user_address AND h.session_key = c.session_key AND h.version = c.version").
		Where("c.user_address = ? AND c.kind = ? AND c.version > 0 AND h.expires_at > ?", userAddress, SessionKeyKindChannel, now).
		Count(&channelCount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count channel session keys for user: %w", err)
	}

	var appCount int64
	err = s.db.Table("current_session_key_states_v1 AS c").
		Joins("JOIN app_session_key_states_v1 h ON h.user_address = c.user_address AND h.session_key = c.session_key AND h.version = c.version").
		Where("c.user_address = ? AND c.kind = ? AND c.version > 0 AND h.expires_at > ?", userAddress, SessionKeyKindAppSession, now).
		Count(&appCount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count app session keys for user: %w", err)
	}

	total := channelCount + appCount
	if total < 0 || total > math.MaxUint32 {
		return 0, fmt.Errorf("session key count %d out of uint32 range", total)
	}
	return uint32(total), nil
}
