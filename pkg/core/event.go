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

// EscrowDepositsPurgedEvent represents the EscrowDepositsPurged event emitted when expired
// escrow deposits are finalized by the purge queue without a signed FINALIZE_ESCROW_DEPOSIT state.
type EscrowDepositsPurgedEvent struct {
	// EscrowIDs holds the hex-encoded escrow IDs (== channel_id in the channels table) that were purged.
	EscrowIDs []string `json:"escrow_ids"`
}

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

// ValidatorRegisteredEvent is emitted by ChannelHub when the node registers a new
// signature validator. Users should react to unexpected registrations by revoking
// ERC20 approvals granted to ChannelHub — see contracts/SECURITY.md for details.
type ValidatorRegisteredEvent struct {
	BlockchainID uint64 `json:"blockchain_id"`
	ValidatorID  uint8  `json:"validator_id"`
	// Validator is the EIP-55 checksummed hex address of the registered validator contract.
	// Always compare using strings.EqualFold or common.HexToAddress(ev.Validator).Hex()
	// to avoid silent mismatches against lowercase or non-checksummed config values.
	Validator   string `json:"validator"`
	BlockNumber uint64 `json:"block_number"` // block where the event was emitted; use as fromBlock on reconnect
}

type BlockchainEvent struct {
	ContractAddress string `json:"contract_address"`
	BlockchainID    uint64 `json:"blockchain_id"`
	Name            string `json:"name"`
	BlockNumber     uint64 `json:"block_number"`
	TransactionHash string `json:"transaction_hash"`
	LogIndex        uint32 `json:"log_index"`
}
