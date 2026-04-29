package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ErrGetAccountBalance = "failed to get account balance"
	ErrRecordLedgerEntry = "failed to record a ledger entry"
)

// AppLedgerEntryV1 represents a ledger entry in the database
type AppLedgerEntryV1 struct {
	ID          uuid.UUID       `gorm:"type:char(36);primaryKey"`
	AccountID   string          `gorm:"column:account_id;not null;index:idx_account_asset_symbol;index:idx_account_wallet"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null;index:idx_account_asset_symbol"`
	Wallet      string          `gorm:"column:wallet;not null;index:idx_account_wallet"`
	Credit      decimal.Decimal `gorm:"column:credit;type:varchar(78);not null"`
	Debit       decimal.Decimal `gorm:"column:debit;type:varchar(78);not null"`
	CreatedAt   time.Time
}

func (AppLedgerEntryV1) TableName() string {
	return "app_ledger_v1"
}

// RecordLedgerEntry logs a movement of funds within the internal ledger.
func (s *DBStore) RecordLedgerEntry(userWallet, accountID, asset string, amount decimal.Decimal) error {
	userWallet = strings.ToLower(userWallet)
	accountID = strings.ToLower(accountID)

	entry := &AppLedgerEntryV1{
		ID:          uuid.New(),
		AccountID:   accountID,
		Wallet:      userWallet,
		AssetSymbol: asset,
		Credit:      decimal.Zero,
		Debit:       decimal.Zero,
		CreatedAt:   time.Now(),
	}

	if amount.IsPositive() {
		entry.Credit = amount
	} else if amount.IsNegative() {
		entry.Debit = amount.Abs()
	} else {
		return nil // Zero amount, nothing to record
	}

	if err := s.db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to record ledger entry: %w", err)
	}

	return nil
}

// GetAppSessionBalances retrieves the total balances associated with a session.
func (s *DBStore) GetAppSessionBalances(appSessionID string) (map[string]decimal.Decimal, error) {
	appSessionID = strings.ToLower(appSessionID)

	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	err := s.db.Model(&AppLedgerEntryV1{}).
		Where("account_id = ?", appSessionID).
		Select("asset_symbol", "COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Group("asset_symbol").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balances for account %s: %w", appSessionID, err)
	}

	result := make(map[string]decimal.Decimal, len(rows))
	for _, r := range rows {
		result[r.Asset] = r.Balance
	}
	return result, nil
}

// GetParticipantAllocations retrieves specific asset allocations per participant.
// This will only return participants who have non-zero balances.
func (s *DBStore) GetParticipantAllocations(appSessionID string) (map[string]map[string]decimal.Decimal, error) {
	appSessionID = strings.ToLower(appSessionID)

	type AllocationRow struct {
		Wallet      string          `gorm:"column:wallet"`
		AssetSymbol string          `gorm:"column:asset_symbol"`
		Balance     decimal.Decimal `gorm:"column:balance"`
	}

	var rows []AllocationRow
	err := s.db.Raw(`
		SELECT
			l.wallet,
			l.asset_symbol,
			COALESCE(SUM(l.credit), 0) - COALESCE(SUM(l.debit), 0) AS balance
		FROM app_ledger_v1 l
		INNER JOIN app_session_participants_v1 p
			ON p.wallet_address = l.wallet AND p.app_session_id = ?
		WHERE l.account_id = ?
		GROUP BY l.wallet, l.asset_symbol
	`, appSessionID, appSessionID).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get participant allocations: %w", err)
	}

	result := make(map[string]map[string]decimal.Decimal)
	for _, row := range rows {
		if _, exists := result[row.Wallet]; !exists {
			result[row.Wallet] = make(map[string]decimal.Decimal)
		}
		result[row.Wallet][row.AssetSymbol] = row.Balance
	}

	return result, nil
}
