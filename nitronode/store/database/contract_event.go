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
	BlockHash       string    `gorm:"column:block_hash"`
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
		BlockHash:       ev.BlockHash,
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

// IsContractEventProcessed reports whether an event identified by (txHash, logIndex, blockchainID)
// has already been committed, regardless of which block it appeared in.
func (s *DBStore) IsContractEventProcessed(txHash string, logIndex uint32, blockchainID uint64) (bool, error) {
	var ev ContractEvent
	err := s.db.Where("transaction_hash = ? AND log_index = ? AND blockchain_id = ?",
		strings.ToLower(txHash), logIndex, blockchainID).
		Take(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetLatestContractEventBlockHashAndNumber returns the block_number and block_hash of the
// highest stored event for the given contract. Returns (0, "", nil) when no rows exist.
func (s *DBStore) GetLatestContractEventBlockHashAndNumber(contractAddress string, blockchainID uint64) (uint64, string, error) {
	var ev ContractEvent
	err := s.db.Where("blockchain_id = ? AND contract_address = ?", blockchainID, strings.ToLower(contractAddress)).
		Order("block_number DESC").
		First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return ev.BlockNumber, ev.BlockHash, nil
}

// GetPreviousDistinctBlockHash returns the block_number and block_hash of the highest
// stored event whose block_number is strictly below belowBlockNumber. Returns (0, "", nil)
// when no such row exists (signals genesis fallback).
func (s *DBStore) GetPreviousDistinctBlockHash(contractAddress string, blockchainID uint64, belowBlockNumber uint64) (uint64, string, error) {
	var ev ContractEvent
	err := s.db.Where("blockchain_id = ? AND contract_address = ? AND block_number < ?",
		blockchainID, strings.ToLower(contractAddress), belowBlockNumber).
		Order("block_number DESC").
		First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return ev.BlockNumber, ev.BlockHash, nil
}
