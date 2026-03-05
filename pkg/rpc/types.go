// Package rpc provides the RPC API types for the Nitrolite Node service.
//
// This file contains common types and structs shared across V1 API groups.
package rpc

import (
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
)

// ============================================================================
// Common Enums
// ============================================================================

// ============================================================================
// Channel Types
// ============================================================================

// ChannelV1 represents an on-chain channel.
type ChannelV1 struct {
	// ChannelID is the unique identifier for the channel
	ChannelID string `json:"channel_id"`
	// UserWallet is the user wallet address
	UserWallet string `json:"user_wallet"`
	// Asset is the asset symbol (e.g. USDC, ETH)
	Asset string `json:"asset"`
	// Type is the type of the channel (home, escrow)
	Type string `json:"type"`
	// BlockchainID is the unique identifier for the blockchain
	BlockchainID string `json:"blockchain_id"`
	// TokenAddress is the address of the token used in the channel
	TokenAddress string `json:"token_address"`
	// ChallengeDuration is the challenge period for the channel in seconds
	ChallengeDuration uint32 `json:"challenge_duration"`
	// ChallegeExpiresAt
	ChallengeExpiresAt *time.Time `json:"challenge_expires_at"`
	// Nonce is the nonce for the channel
	Nonce string `json:"nonce"`
	// ApprovedSigValidators is a hex string bitmap of approved signature validators
	ApprovedSigValidators string `json:"approved_sig_validators"`
	// Status is the current status of the channel (void, open, challenged, closed)
	Status string `json:"status"`
	// StateVersion is the on-chain state version of the channel
	StateVersion string `json:"state_version"`
}

// ChannelDefinitionV1 represents the configuration for creating a channel.
type ChannelDefinitionV1 struct {
	// Nonce is a unique number to prevent replay attacks
	Nonce string `json:"nonce"`
	// Challenge is the challenge period for the channel in seconds
	Challenge uint32 `json:"challenge"`
	// ApprovedSigValidators is a hex string bitmap representing the approved signature validators for the channel
	ApprovedSigValidators string `json:"approved_sig_validators"`
}

// ChannelSessionKeyStateV1 represents the state of a session key.
type ChannelSessionKeyStateV1 struct {
	// ID Hash(user_address + session_key + version)
	// UserAddress is the user wallet address
	UserAddress string `json:"user_address"`
	// SessionKey is the session key address for delegation
	SessionKey string `json:"session_key"`
	// Version is the version of the session key format
	Version string `json:"version"`
	// Assets associated with this session key
	Assets []string `json:"assets"`
	// Expiration time as unix timestamp of this session key
	ExpiresAt string `json:"expires_at"`
	// UserSig is the user's signature over the session key metadata to authorize the registration/update of the session key
	UserSig string `json:"user_sig"`
}

// ============================================================================
// State Types
// ============================================================================

// TransitionV1 represents a state transition.
type TransitionV1 struct {
	// Type is the type of state transition
	Type core.TransitionType `json:"type"`
	// TxID is the transaction ID associated with the transition
	TxID string `json:"tx_id"`
	// AccountID is the account identifier (varies based on transition type)
	AccountID string `json:"account_id"`
	// Amount is the amount involved in the transition
	Amount string `json:"amount"`
}

// LedgerV1 represents ledger balances for a channel.
type LedgerV1 struct {
	// TokenAddress is the address of the token used in this channel
	TokenAddress string `json:"token_address"`
	// BlockchainID is the unique identifier for the blockchain
	BlockchainID string `json:"blockchain_id"`
	// UserBalance is the user balance in the channel
	UserBalance string `json:"user_balance"`
	// UserNetFlow is the user net flow in the channel
	UserNetFlow string `json:"user_net_flow"`
	// NodeBalance is the node balance in the channel
	NodeBalance string `json:"node_balance"`
	// NodeNetFlow is the node net flow in the channel
	NodeNetFlow string `json:"node_net_flow"`
}

// StateV1 represents the current state of the user stored on Node.
type StateV1 struct {
	// ID is the deterministic ID (hash) of the state
	ID string `json:"id"`
	// Transition is the state transition that led to this state
	Transition TransitionV1 `json:"transition"`
	// Asset is the asset type of the state
	Asset string `json:"asset"`
	// UserWallet is the user wallet address
	UserWallet string `json:"user_wallet"`
	// Epoch is the user Epoch Index
	Epoch string `json:"epoch"`
	// Version is the version of the state
	Version string `json:"version"`
	// HomeChannelID is the identifier for the home Channel blockchain network
	HomeChannelID *string `json:"home_channel_id,omitempty"`
	// EscrowChannelID is the identifier for the escrow Channel blockchain network
	EscrowChannelID *string `json:"escrow_channel_id,omitempty"`
	// HomeLedger contains user and node balances for the home channel
	HomeLedger LedgerV1 `json:"home_ledger"`
	// EscrowLedger contains user and node balances for the escrow channel
	EscrowLedger *LedgerV1 `json:"escrow_ledger,omitempty"`
	// UserSig is the user signature for the state
	UserSig *string `json:"user_sig,omitempty"`
	// NodeSig is the node signature for the state
	NodeSig *string `json:"node_sig,omitempty"`
}

// ============================================================================
// App Session Types
// ============================================================================

// AppParticipantV1 represents the definition for an app participant.
type AppParticipantV1 struct {
	// WalletAddress is the participant's wallet address
	WalletAddress string `json:"wallet_address"`
	// SignatureWeight is the signature weight for the participant
	SignatureWeight uint8 `json:"signature_weight"`
}

// AppDefinitionV1 represents the definition for an app session.
type AppDefinitionV1 struct {
	// Application is the application identifier from an app registry
	Application string `json:"application_id"`
	// Participants is the list of participants in the app session
	Participants []AppParticipantV1 `json:"participants"`
	// Quorum is the quorum required for the app session
	Quorum uint8 `json:"quorum"`
	// Nonce is a unique number to prevent replay attacks
	Nonce string `json:"nonce"`
}

// AppAllocationV1 represents the allocation of assets to a participant in an app session.
type AppAllocationV1 struct {
	// Participant is the participant's wallet address
	Participant string `json:"participant"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
	// Amount is the amount allocated to the participant
	Amount string `json:"amount"`
}

// AppStateUpdateV1 represents the current state of an application session.
type AppStateUpdateV1 struct {
	// AppSessionID is the unique application session identifier
	AppSessionID string `json:"app_session_id"`
	// Intent is the intent of the app session update (operate, deposit, withdraw, close)
	Intent app.AppStateUpdateIntent `json:"intent"`
	// Version is the version of the app state
	Version string `json:"version"`
	// Allocations is the list of allocations in the app state
	Allocations []AppAllocationV1 `json:"allocations"`
	// SessionData is the JSON stringified session data
	SessionData string `json:"session_data"`
}

// AppSessionInfoV1 represents information about an application session.
type AppSessionInfoV1 struct {
	// AppSessionID is the unique application session identifier
	AppSessionID string `json:"app_session_id"`
	// Status is the session status (open/closed)
	Status string `json:"status"`
	// AppDefinitionV1 contains immutable application fields
	AppDefinitionV1 AppDefinitionV1 `json:"app_definition"`
	// SessionData is the JSON stringified session data
	SessionData *string `json:"session_data,omitempty"`
	// Version is the current version of the session state
	Version string `json:"version"`
	// Nonce is the nonce for the session
	Allocations []AppAllocationV1 `json:"allocations"`
}

// AppSessionKeyStateV1 represents the state of a session key.
type AppSessionKeyStateV1 struct {
	// ID Hash(user_address + session_key + version)
	// UserAddress is the user wallet address
	UserAddress string `json:"user_address"`
	// SessionKey is the session key address for delegation
	SessionKey string `json:"session_key"`
	// Version is the version of the session key format
	Version string `json:"version"`
	// ApplicationID is the application IDs associated with this session key
	ApplicationIDs []string `json:"application_ids"`
	// AppSessionID is the application session IDs associated with this session key
	AppSessionIDs []string `json:"app_session_ids"`
	// ExpiresAt is Unix timestamp in seconds indicating when the session key expires
	ExpiresAt string `json:"expires_at"`
	// UserSig is the user's signature over the session key metadata to authorize the registration/update of the session key
	UserSig string `json:"user_sig"`
}

// ============================================================================
// App Registry Types
// ============================================================================

// AppV1 represents a registered application definition (without timestamps).
type AppV1 struct {
	// ID is the application identifier
	ID string `json:"id"`
	// OwnerWallet is the owner's wallet address
	OwnerWallet string `json:"owner_wallet"`
	// Metadata is the application metadata
	Metadata string `json:"metadata"`
	// Version is the current version
	Version string `json:"version"`
	// CreationApprovalNotRequired indicates if sessions can be created without approval
	CreationApprovalNotRequired bool `json:"creation_approval_not_required"`
}

// AppInfoV1 represents full application info including timestamps.
type AppInfoV1 struct {
	AppV1
	// CreatedAt is the creation timestamp (unix seconds)
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the last update timestamp (unix seconds)
	UpdatedAt string `json:"updated_at"`
}

// ============================================================================
// Asset and Blockchain Types
// ============================================================================

// AssetV1 represents information about a supported asset.
type AssetV1 struct {
	// Name is the asset name
	Name string `json:"name"`
	// Symbol is the asset symbol
	Symbol string `json:"symbol"`
	// Decimals is the number of decimal places for the asset
	Decimals uint8 `json:"decimals"`
	// SuggestedBlockchainID is the suggested blockchain network ID for this asset
	SuggestedBlockchainID string `json:"suggested_blockchain_id"`
	// Tokens is the list of supported tokens for the asset
	Tokens []TokenV1 `json:"tokens"`
}

// TokenV1 represents information about a supported token.
type TokenV1 struct {
	// Name is the token name
	Name string `json:"name"`
	// Symbol is the token symbol
	Symbol string `json:"symbol"`
	// Address is the token contract address
	Address string `json:"address"`
	// BlockchainID is the blockchain network ID
	BlockchainID string `json:"blockchain_id"`
	// Decimals is the number of decimal places
	Decimals uint8 `json:"decimals"`
}

// BlockchainInfoV1 represents information about a supported network.
type BlockchainInfoV1 struct {
	// Name is the blockchain name
	Name string `json:"name"`
	// BlockchainID is the blockchain network ID
	BlockchainID string `json:"blockchain_id"`
	// ChannelHubAddress is the contract address on this network
	ChannelHubAddress string `json:"channel_hub_address"`
	// DefaultValidatorAddress is the validator contract address set in a channel supported by clearnode
}

// ============================================================================
// Balance and Transaction Types
// ============================================================================

// BalanceEntryV1 represents a balance for a specific asset.
type BalanceEntryV1 struct {
	// Asset is the asset symbol
	Asset string `json:"asset"`
	// Amount is the balance amount
	Amount string `json:"amount"`
}

// TransactionV1 represents a transaction record.
type TransactionV1 struct {
	// ID is the unique transaction reference
	ID string `json:"id"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
	// TxType is the transaction type
	TxType core.TransactionType `json:"tx_type"`
	// FromAccount is the account that sent the funds
	FromAccount string `json:"from_account"`
	// ToAccount is the account that received the funds
	ToAccount string `json:"to_account"`
	// SenderNewStateID is the ID of the new sender's channel state
	SenderNewStateID *string `json:"sender_new_state_id,omitempty"`
	// ReceiverNewStateID is the ID of the new receiver's channel state
	ReceiverNewStateID *string `json:"receiver_new_state_id,omitempty"`
	// Amount is the transaction amount
	Amount string `json:"amount"`
	// CreatedAt is when the transaction was created
	CreatedAt string `json:"created_at"`
}

// ============================================================================
// Action Gateway Types
// ============================================================================

// ActionAllowanceV1 represents the allowance information for a specific gated action.
type ActionAllowanceV1 struct {
	// GatedAction is the specific action being gated (e.g. transfer, app_operation)
	GatedAction core.GatedAction `json:"gated_action"`
	// TimeWindow is the time window for which the allowance is valid (e.g. "1h", "24h")
	TimeWindow string `json:"time_window"`
	// Allowance is the total allowance for the gated action within the time window
	Allowance string `json:"allowance"`
	// Used is the amount of the allowance that has already been used within the time window
	Used string `json:"used"`
}

// ============================================================================
// Pagination Types
// ============================================================================

// PaginationParamsV1 represents pagination request parameters.
type PaginationParamsV1 struct {
	// Offset is the pagination offset (number of items to skip)
	Offset *uint32 `json:"offset,omitempty"`
	// Limit is the number of items to return
	Limit *uint32 `json:"limit,omitempty"`
	// Sort is the sort order (asc/desc)
	Sort *string `json:"sort,omitempty"`
}

// PaginationMetadataV1 represents pagination information.
type PaginationMetadataV1 struct {
	// Page is the current page number
	Page uint32 `json:"page"`
	// PerPage is the number of items per page
	PerPage uint32 `json:"per_page"`
	// TotalCount is the total number of items
	TotalCount uint32 `json:"total_count"`
	// PageCount is the total number of pages
	PageCount uint32 `json:"page_count"`
}
