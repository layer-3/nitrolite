package database

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BlockchainActionType uint8

const (
	ActionTypeCheckpoint BlockchainActionType = 1

	ActionTypeInitiateEscrowDeposit BlockchainActionType = 10
	ActionTypeFinalizeEscrowDeposit BlockchainActionType = 11

	ActionTypeInitiateEscrowWithdrawal BlockchainActionType = 20
	ActionTypeFinalizeEscrowWithdrawal BlockchainActionType = 21
)

func (t BlockchainActionType) String() string {
	switch t {
	case ActionTypeCheckpoint:
		return "checkpoint"
	case ActionTypeInitiateEscrowDeposit:
		return "initiate_escrow_deposit"
	case ActionTypeFinalizeEscrowDeposit:
		return "finalize_escrow_deposit"
	case ActionTypeInitiateEscrowWithdrawal:
		return "initiate_escrow_withdrawal"
	case ActionTypeFinalizeEscrowWithdrawal:
		return "finalize_escrow_withdrawal"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

type BlockchainActionStatus uint8

const (
	BlockchainActionStatusPending BlockchainActionStatus = iota
	BlockchainActionStatusCompleted
	BlockchainActionStatusFailed
)

type BlockchainAction struct {
	ID           int64                  `gorm:"primary_key"`
	Type         BlockchainActionType   `gorm:"column:action_type;not null"`
	StateID      string                 `gorm:"column:state_id;size:66"`
	BlockchainID uint64                 `gorm:"column:blockchain_id;not null"`
	Data         datatypes.JSON         `gorm:"column:action_data;type:text"`
	Status       BlockchainActionStatus `gorm:"column:status;not null"`
	Retries      uint8                  `gorm:"column:retry_count;default:0"`
	Error        string                 `gorm:"column:last_error;type:text"`
	TxHash       string                 `gorm:"column:transaction_hash;size:66"`
	CreatedAt    time.Time              `gorm:"column:created_at"`
	UpdatedAt    time.Time              `gorm:"column:updated_at"`
}

func (BlockchainAction) TableName() string {
	return "blockchain_actions"
}

// ScheduleCheckpoint queues a blockchain action to checkpoint a state on home blockchain.
func (s *DBStore) ScheduleCheckpoint(stateID string, blockchainID uint64) error {
	return s.scheduleStateEnforcement(stateID, blockchainID, ActionTypeCheckpoint)
}

// ScheduleInitiateEscrowDeposit queues a blockchain action to initiate escrow deposit on home blockchain.
func (s *DBStore) ScheduleInitiateEscrowDeposit(stateID string, blockchainID uint64) error {
	return s.scheduleStateEnforcement(stateID, blockchainID, ActionTypeInitiateEscrowDeposit)
}

// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate withdrawal on non-home blockchain.
func (s *DBStore) ScheduleInitiateEscrowWithdrawal(stateID string, blockchainID uint64) error {
	return s.scheduleStateEnforcement(stateID, blockchainID, ActionTypeInitiateEscrowWithdrawal)
}

// ScheduleFinalizeEscrowDeposit schedules a finalize for an escrow deposit operation on non-home blockchain.
func (s *DBStore) ScheduleFinalizeEscrowDeposit(stateID string, blockchainID uint64) error {
	return s.scheduleStateEnforcement(stateID, blockchainID, ActionTypeFinalizeEscrowDeposit)
}

// ScheduleFinalizeEscrowWithdrawal schedules a finalize for an escrow withdrawal operation on non-home blockchain.
func (s *DBStore) ScheduleFinalizeEscrowWithdrawal(stateID string, blockchainID uint64) error {
	return s.scheduleStateEnforcement(stateID, blockchainID, ActionTypeFinalizeEscrowWithdrawal)
}

// scheduleStateEnforcement is a helper to create a blockchain action for state enforcement.
func (s *DBStore) scheduleStateEnforcement(stateID string, blockchainID uint64, actionType BlockchainActionType) error {
	action := &BlockchainAction{
		Type:         actionType,
		StateID:      stateID,
		BlockchainID: blockchainID,
		Status:       BlockchainActionStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.db.Create(action).Error
}

func (s *DBStore) Fail(actionID int64, err string) error {
	return s.updateAction(actionID, BlockchainActionStatusFailed, "", err, true)
}

func (s *DBStore) FailNoRetry(actionID int64, err string) error {
	return s.updateAction(actionID, BlockchainActionStatusFailed, "", err, false)
}

func (s *DBStore) RecordAttempt(actionID int64, err string) error {
	return s.updateAction(actionID, BlockchainActionStatusPending, "", err, true)
}

func (s *DBStore) Complete(actionID int64, txHash string) error {
	return s.updateAction(actionID, BlockchainActionStatusCompleted, txHash, "", false)
}

func (s *DBStore) updateAction(actionID int64, status BlockchainActionStatus, txHash, err string, increaseRetryCounter bool) error {
	updates := map[string]any{
		"status":     status,
		"last_error": err,
		"updated_at": time.Now(),
	}

	if txHash != "" {
		updates["transaction_hash"] = txHash
	}
	if increaseRetryCounter {
		updates["retry_count"] = gorm.Expr("retry_count + ?", 1)
	}

	if err := s.db.Model(&BlockchainAction{}).Where("id = ?", actionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update blockchain action: %w", err)
	}

	return nil
}

func (s *DBStore) GetActions(limit uint8, blockchainID uint64) ([]BlockchainAction, error) {
	var actions []BlockchainAction
	query := s.db.Where("status = ? AND blockchain_id = ?", BlockchainActionStatusPending, blockchainID).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if err := query.Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to get blockchain actions: %w", err)
	}
	return actions, nil
}
