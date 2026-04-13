package database

import (
	"errors"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"gorm.io/gorm"
)

var ErrEventHasAlreadyBeenProcessed = errors.New("contract event has already been processed")

type ContractEvent struct {
	ID              int64     `gorm:"primary_key;column:id"`
	ContractAddress string    `gorm:"column:contract_address"`
	BlockchainID    uint64    `gorm:"column:blockchain_id"`
	Name            string    `gorm:"column:name"`
	BlockNumber     uint64    `gorm:"column:block_number"`
	TransactionHash string    `gorm:"column:transaction_hash"`
	LogIndex        uint32    `gorm:"column:log_index"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (ContractEvent) TableName() string {
	return "contract_events"
}

// StoreContractEvent stores a blockchain event to the database.
// This function matches the signature required by pkg/blockchain/evm.StoreContractEvent.
func (s *DBStore) StoreContractEvent(ev core.BlockchainEvent) error {
	contractEvent := &ContractEvent{
		ContractAddress: strings.ToLower(ev.ContractAddress),
		BlockchainID:    ev.BlockchainID,
		Name:            ev.Name,
		BlockNumber:     ev.BlockNumber,
		TransactionHash: strings.ToLower(ev.TransactionHash),
		LogIndex:        ev.LogIndex,
		CreatedAt:       time.Now(),
	}

	return s.db.Create(contractEvent).Error
}

// GetLatestContractEventBlockNumber returns the highest block number stored for a given contract.
func (s *DBStore) GetLatestContractEventBlockNumber(contractAddress string, blockchainID uint64) (uint64, error) {
	var blockNumber uint64
	err := s.db.Model(&ContractEvent{}).
		Where("blockchain_id = ? AND contract_address = ?", blockchainID, strings.ToLower(contractAddress)).
		Select("COALESCE(MAX(block_number), 0)").
		Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

// IsContractEventPresent checks whether a specific contract event has already been stored.
func (s *DBStore) IsContractEventPresent(blockchainID, blockNumber uint64, txHash string, logIndex uint32) (bool, error) {
	var ev ContractEvent
	err := s.db.Where("blockchain_id = ? AND block_number = ? AND transaction_hash = ? AND log_index = ?",
		blockchainID, blockNumber, strings.ToLower(txHash), logIndex).
		Take(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
