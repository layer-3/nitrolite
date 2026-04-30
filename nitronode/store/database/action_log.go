package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/layer-3/nitrolite/pkg/core"
)

type ActionLogEntryV1 struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey"`
	UserWallet  string    `gorm:"column:user_wallet;not null"`
	GatedAction uint8     `gorm:"column:gated_action;not null"`
	CreatedAt   time.Time
}

func (ActionLogEntryV1) TableName() string {
	return "action_log_v1"
}

// RecordAction inserts a new action log entry for a user.
func (s *DBStore) RecordAction(wallet string, gatedAction core.GatedAction) error {
	if gatedAction.ID() == 0 {
		return fmt.Errorf("invalid gated action ID")
	}

	wallet = strings.ToLower(wallet)

	entry := ActionLogEntryV1{
		ID:          uuid.New(),
		UserWallet:  wallet,
		GatedAction: gatedAction.ID(),
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return fmt.Errorf("failed to record action log entry: %w", err)
	}

	return nil
}

// GetUserActionCount returns the number of actions matching the given wallet and gated action
// within the specified time window (counting backwards from now).
func (s *DBStore) GetUserActionCount(wallet string, gatedAction core.GatedAction, window time.Duration) (uint64, error) {
	wallet = strings.ToLower(wallet)
	since := time.Now().Add(-window)

	query := s.db.Model(&ActionLogEntryV1{}).
		Where("user_wallet = ? AND gated_action = ? AND created_at >= ?", wallet, gatedAction.ID(), since)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get user action count: %w", err)
	}

	return uint64(count), nil
}

func (s *DBStore) GetUserActionCounts(userWallet string, window time.Duration) (map[core.GatedAction]uint64, error) {
	userWallet = strings.ToLower(userWallet)
	since := time.Now().Add(-window)

	query := s.db.Model(&ActionLogEntryV1{}).
		Select("gated_action, COUNT(id) as count").
		Where("user_wallet = ? AND created_at >= ?", userWallet, since).
		Group("gated_action")

	type Result struct {
		GatedAction uint8
		Count       int64
	}

	var results []Result
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get user action counts: %w", err)
	}

	counts := make(map[core.GatedAction]uint64)
	for _, r := range results {
		action, ok := core.GatedActionFromID(r.GatedAction)
		if !ok {
			continue
		}
		counts[action] = uint64(r.Count)
	}

	return counts, nil
}
