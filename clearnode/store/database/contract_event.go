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
		TransactionHash: ev.TransactionHash,
		LogIndex:        ev.LogIndex,
		CreatedAt:       time.Now(),
	}

	return s.db.Create(contractEvent).Error
}

// GetLatestEvent returns the latest block number and log index for a given contract.
// This function matches the signature required by pkg/blockchain/evm.GetLatestEvent.
func (s *DBStore) GetLatestEvent(contractAddress string, blockchainID uint64) (core.BlockchainEvent, error) {
	var ev ContractEvent
	err := s.db.Where("blockchain_id = ? AND contract_address = ?", blockchainID, strings.ToLower(contractAddress)).
		Order("block_number DESC, log_index DESC").
		First(&ev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No events found, return zeros (will start from beginning)
		return core.BlockchainEvent{}, nil
	}

	if err != nil {
		return core.BlockchainEvent{}, err
	}

	return core.BlockchainEvent{
		BlockNumber:     ev.BlockNumber,
		BlockchainID:    ev.BlockchainID,
		Name:            ev.Name,
		ContractAddress: ev.ContractAddress,
		TransactionHash: ev.TransactionHash,
		LogIndex:        ev.LogIndex,
	}, nil
}
