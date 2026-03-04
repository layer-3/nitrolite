package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

type UserStakedV1 struct {
	UserWallet   string          `gorm:"column:user_wallet;primaryKey;not null"`
	BlockchainID uint64          `gorm:"column:blockchain_id;primaryKey;not null"`
	Amount       decimal.Decimal `gorm:"column:amount;type:varchar(78);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserStakedV1) TableName() string {
	return "user_staked_v1"
}

// UpdateUserStaked upserts the staked amount for a user on a specific blockchain.
func (s *DBStore) UpdateUserStaked(wallet string, blockchainID uint64, amount decimal.Decimal) error {
	wallet = strings.ToLower(wallet)

	if wallet == "" {
		return fmt.Errorf("wallet address must not be empty")
	}
	if blockchainID == 0 {
		return fmt.Errorf("blockchain ID must not be zero")
	}
	if amount.IsNegative() {
		return fmt.Errorf("staked amount must not be negative")
	}

	now := time.Now()

	record := UserStakedV1{
		UserWallet:   wallet,
		BlockchainID: blockchainID,
		Amount:       amount,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_wallet"}, {Name: "blockchain_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"}),
	}).Create(&record).Error
	if err != nil {
		return fmt.Errorf("failed to update user staked amount: %w", err)
	}

	return nil
}

// GetTotalUserStaked returns the total staked amount for a user across all blockchains.
func (s *DBStore) GetTotalUserStaked(wallet string) (decimal.Decimal, error) {
	wallet = strings.ToLower(wallet)

	var result struct {
		Total decimal.Decimal `gorm:"column:total"`
	}
	err := s.db.Model(&UserStakedV1{}).
		Where("user_wallet = ?", wallet).
		Select("COALESCE(SUM(amount), 0) AS total").
		Scan(&result).Error
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get user staked total: %w", err)
	}

	return result.Total, nil
}
