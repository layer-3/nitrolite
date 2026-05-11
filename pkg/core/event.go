package core

import "github.com/shopspring/decimal"

// On-chain events

// NodeBalanceUpdatedEvent represents the NodeBalanceUpdated event
type NodeBalanceUpdatedEvent struct {
	BlockchainID uint64          `json:"blockchain_id"`
	Asset        string          `json:"asset"`
	Balance      decimal.Decimal `json:"balance"`
}

// HomeChannelCreatedEvent represents the ChannelCreated event
type HomeChannelCreatedEvent channelEvent

// HomeChannelMigratedEvent represents the ChannelMigrated event
type HomeChannelMigratedEvent channelEvent

// HomeChannelCheckpointedEvent represents the Checkpointed event
type HomeChannelCheckpointedEvent channelEvent

// HomeChannelChallengedEvent represents the Challenged event
type HomeChannelChallengedEvent channelChallengedEvent

// HomeChannelClosedEvent represents the Closed event
type HomeChannelClosedEvent channelEvent

// EscrowDepositInitiatedEvent represents the EscrowDepositInitiated event
type EscrowDepositInitiatedEvent channelEvent

// EscrowDepositChallengedEvent represents the EscrowDepositChallenged event
type EscrowDepositChallengedEvent channelChallengedEvent

// EscrowDepositFinalizedEvent represents the EscrowDepositFinalized event
type EscrowDepositFinalizedEvent channelEvent

// EscrowWithdrawalInitiatedEvent represents the EscrowWithdrawalInitiated event
type EscrowWithdrawalInitiatedEvent channelEvent

// EscrowWithdrawalChallengedEvent represents the EscrowWithdrawalChallenged event
type EscrowWithdrawalChallengedEvent channelChallengedEvent

// EscrowWithdrawalFinalizedEvent represents the EscrowWithdrawalFinalized event
type EscrowWithdrawalFinalizedEvent channelEvent

type channelEvent struct {
	ChannelID    string `json:"channel_id"`
	StateVersion uint64 `json:"state_version"`
	// UserSig is the hex-encoded user signature recovered from the on-chain state payload.
	// Empty when the parsed event carries no user signature (e.g. unilateral node-only state).
	UserSig string `json:"user_sig,omitempty"`
}

type channelChallengedEvent struct {
	ChannelID       string `json:"channel_id"`
	StateVersion    uint64 `json:"state_version"`
	ChallengeExpiry uint64 `json:"challenge_expiry"`
	// UserSig is the hex-encoded user signature recovered from the on-chain state payload.
	UserSig string `json:"user_sig,omitempty"`
}

type UserLockedBalanceUpdatedEvent struct {
	UserAddress  string          `json:"user_address"`
	BlockchainID uint64          `json:"blockchain_id"`
	Balance      decimal.Decimal `json:"balance"`
}

type BlockchainEvent struct {
	ContractAddress string `json:"contract_address"`
	BlockchainID    uint64 `json:"blockchain_id"`
	Name            string `json:"name"`
	BlockNumber     uint64 `json:"block_number"`
	TransactionHash string `json:"transaction_hash"`
	LogIndex        uint32 `json:"log_index"`
}
