package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidLedgerTransactionType = "invalid ledger transaction type"
	ErrRecordTransaction            = "failed to record transaction"
)

// Transaction represents an immutable transaction in the system
// ID is deterministic based on transaction initiation:
// 1) Initiated by User: Hash(ToAccount, SenderNewStateID)
// 2) Initiated by Node: Hash(FromAccount, ReceiverNewStateID)
type Transaction struct {
	// ID is a 64-character deterministic hash
	ID                 string               `gorm:"column:id;primaryKey;size:64"`
	Type               core.TransactionType `gorm:"column:tx_type;not null;index:idx_type;index:idx_from_to_account"`
	AssetSymbol        string               `gorm:"column:asset_symbol;not null"`
	FromAccount        string               `gorm:"column:from_account;not null;index:idx_from_account;index:idx_from_to_account"`
	ToAccount          string               `gorm:"column:to_account;not null;index:idx_to_account;index:idx_from_to_account"`
	SenderNewStateID   *string              `gorm:"column:sender_new_state_id;size:64"`
	ReceiverNewStateID *string              `gorm:"column:receiver_new_state_id;size:64"`
	Amount             decimal.Decimal      `gorm:"column:amount;type:decimal(38,18);not null"`
	CreatedAt          time.Time

	// ApplicationID is the self-declared origin tag of the client that caused
	// this transaction (see rpc.ApplicationIDQueryParam). Advisory only.
	ApplicationID *string `gorm:"column:application_id;size:66;index:idx_transactions_app_id"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// RecordTransaction creates a transaction record linking state transitions.
// applicationID is the client-declared origin tag; empty string → NULL column.
func (s *DBStore) RecordTransaction(tx core.Transaction, applicationID string) error {
	dbTx := Transaction{
		ID:          strings.ToLower(tx.ID),
		Type:        tx.TxType,
		AssetSymbol: tx.Asset,
		FromAccount: strings.ToLower(tx.FromAccount),
		ToAccount:   strings.ToLower(tx.ToAccount),
		Amount:      tx.Amount,
		CreatedAt:   tx.CreatedAt,
	}
	if tx.SenderNewStateID != nil {
		lowered := strings.ToLower(*tx.SenderNewStateID)
		dbTx.SenderNewStateID = &lowered
	}
	if tx.ReceiverNewStateID != nil {
		lowered := strings.ToLower(*tx.ReceiverNewStateID)
		dbTx.ReceiverNewStateID = &lowered
	}
	if applicationID != "" {
		dbTx.ApplicationID = &applicationID
	}

	if err := s.db.Create(&dbTx).Error; err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	return nil
}

// GetUserTransactions retrieves transaction history for a user with optional filters.
func (s *DBStore) GetUserTransactions(accountID string, asset *string, txType *core.TransactionType, fromTime *uint64, toTime *uint64, paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error) {
	query := s.db.Model(&Transaction{})

	// Filter by wallet (from or to account)
	accountID = strings.ToLower(accountID)
	query = query.Where("from_account = ? OR to_account = ?", accountID, accountID)

	// Filter by asset if provided
	if asset != nil && *asset != "" {
		query = query.Where("asset_symbol = ?", *asset)
	}

	// Filter by transaction type if provided
	if txType != nil {
		dbTxType := *txType
		query = query.Where("tx_type = ?", dbTxType)
	}

	// Filter by time range if provided
	if fromTime != nil {
		t := time.Unix(int64(*fromTime), 0)
		query = query.Where("created_at >= ?", t)
	}
	if toTime != nil {
		t := time.Unix(int64(*toTime), 0)
		query = query.Where("created_at <= ?", t)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to count transactions: %w", err)
	}

	offset, limit := paginate.GetOffsetAndLimit(DefaultLimit, MaxLimit)

	query = query.Order("created_at DESC").Offset(int(offset)).Limit(int(limit))

	var dbTransactions []Transaction
	if err := query.Find(&dbTransactions).Error; err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get transactions: %w", err)
	}

	transactions := make([]core.Transaction, len(dbTransactions))
	for i, dbTx := range dbTransactions {
		transactions[i] = *toCoreTransaction(&dbTx)
	}

	metadata := calculatePaginationMetadata(totalCount, offset, limit)

	return transactions, metadata, nil
}
