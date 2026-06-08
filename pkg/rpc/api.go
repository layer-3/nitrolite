// Package rpc provides the RPC API types for the Nitrolite Node service.
//
// This file implements the API request and response definitions specified in api.yaml
// with versioned types organized by functional groups. All types follow the naming
// convention: {Group}{Version}{Name}{Request|Response}.
package rpc

import (
	"github.com/layer-3/nitrolite/pkg/core"
)

// ============================================================================
// Channels Group - V1 API
// ============================================================================

// ChannelsV1GetHomeChannelRequest retrieves current on-chain home channel information.
type ChannelsV1GetHomeChannelRequest struct {
	// UserWallet is the user's wallet address
	Wallet string `json:"wallet"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
}

// ChannelsV1GetHomeChannelResponse returns the on-chain channel information.
// Channel is nil when no home channel exists for the given wallet/asset pair;
// this is a successful response, not an error.
type ChannelsV1GetHomeChannelResponse struct {
	// Channel is the on-chain channel information, or nil if absent.
	Channel *ChannelV1 `json:"channel,omitempty"`
}

// ChannelsV1GetEscrowChannelRequest retrieves current on-chain escrow channel information.
type ChannelsV1GetEscrowChannelRequest struct {
	// EscrowChannelID is the escrow channel ID
	EscrowChannelID string `json:"escrow_channel_id"`
}

// ChannelsV1GetEscrowChannelResponse returns the on-chain channel information.
// Channel is nil when no escrow channel exists for the given ID; this is a
// successful response, not an error.
type ChannelsV1GetEscrowChannelResponse struct {
	// Channel is the on-chain channel information, or nil if absent.
	Channel *ChannelV1 `json:"channel,omitempty"`
}

// ChannelsV1GetChannelsRequest retrieves all channels for a user with optional filtering.
type ChannelsV1GetChannelsRequest struct {
	// Wallet filters by user's wallet address
	Wallet string `json:"wallet"`
	// Status filters by status
	Status *string `json:"status,omitempty"`
	// Asset filters by asset
	Asset *string `json:"asset,omitempty"`
	// ChannelType filters by channel type ("home" or "escrow")
	ChannelType *string `json:"channel_type,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// ChannelsV1GetChannelsResponse returns the list of channels.
type ChannelsV1GetChannelsResponse struct {
	// Channels is the list of channels
	Channels []ChannelV1 `json:"channels"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ChannelsV1GetLatestStateRequest retrieves the current state of the user stored on the Node.
type ChannelsV1GetLatestStateRequest struct {
	// UserWallet is the user's wallet address
	Wallet string `json:"wallet"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
	// OnlySigned can be enabled to get the latest signed state to know what is the current pending transition
	OnlySigned bool `json:"only_signed"`
}

// ChannelsV1GetLatestStateResponse returns the current state of the user.
// State is nil when no state has been stored for the given wallet/asset pair;
// this is a successful response, not an error.
type ChannelsV1GetLatestStateResponse struct {
	// State is the current state of the user, or nil if absent.
	State *StateV1 `json:"state,omitempty"`
}

// ChannelsV1RequestCreationRequest requests the creation of a channel from Node.
type ChannelsV1RequestCreationRequest struct {
	// State is the state to be submitted
	State StateV1 `json:"state"`
	// ChannelDefinition is the definition of the channel to be created
	ChannelDefinition ChannelDefinitionV1 `json:"channel_definition"`
}

// ChannelsV1RequestCreationResponse returns the Node's signature for the state.
type ChannelsV1RequestCreationResponse struct {
	// Signature is the Node's signature for the state
	Signature string `json:"signature"`
}

// ChannelsV1SubmitStateRequest submits a cross-chain state.
type ChannelsV1SubmitStateRequest struct {
	// State is the state to be submitted
	State StateV1 `json:"state"`
}

// ChannelsV1SubmitStateResponse returns the Node's signature for the state.
type ChannelsV1SubmitStateResponse struct {
	// Signature is the Node's signature for the state
	Signature string `json:"signature"`
}

// ChannelsV1HomeChannelCreatedEvent is emitted when a home channel is created.
type ChannelsV1HomeChannelCreatedEvent struct {
	// Channel is the created home channel information
	Channel ChannelV1 `json:"channel"`
	// InitialState is the initial state of the home channel
	InitialState StateV1 `json:"initial_state"`
}

// ChannelsV1SubmitSessionKeyStateRequest submits a channel session key state for registration,
// rotation/update, or revocation. A submit whose ExpiresAt is in the past (<= server's now)
// is treated as a revocation: the auth path stops accepting state signed by the key and the
// slot is freed against the per-user cap.
type ChannelsV1SubmitSessionKeyStateRequest struct {
	// State contains the session key metadata and delegation information
	State ChannelSessionKeyStateV1 `json:"state"`
}

// ChannelsV1SubmitSessionKeyStateResponse returns the result of session key state submission.
type ChannelsV1SubmitSessionKeyStateResponse struct {
}

// ChannelsV1GetLastKeyStatesRequest retrieves the latest session key states for a user with optional filtering by session key.
type ChannelsV1GetLastKeyStatesRequest struct {
	// UserAddress is the user's wallet address
	UserAddress string  `json:"user_address"`
	SessionKey  *string `json:"session_key,omitempty"` // Optionally filter by SessionKey
	// IncludeInactive, when true, includes latest states whose expires_at is in the past
	// (expired or revoked). Defaults to false: only currently active states are returned.
	IncludeInactive *bool `json:"include_inactive,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// ChannelsV1GetLastKeyStatesResponse returns the latest session key states for the user.
type ChannelsV1GetLastKeyStatesResponse struct {
	// States is the list of latest session key states for the user, filtered by IncludeInactive.
	States []ChannelSessionKeyStateV1 `json:"states"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ============================================================================
// App Sessions Group - V1 API
// ============================================================================

// AppSessionsV1SubmitDepositStateRequest submits an application session state update.
type AppSessionsV1SubmitDepositStateRequest struct {
	// AppStateUpdate is the application session state update to be submitted
	AppStateUpdate AppStateUpdateV1 `json:"app_state_update"`
	// QuorumSigs is the list of participant signatures for the app state update
	QuorumSigs []string `json:"quorum_sigs"`
	// UserState is the signed channel state from the user, used to fund the application session deposit
	UserState StateV1 `json:"user_state"`
}

// AppSessionsV1SubmitDepositStateResponse returns the Node's signature for the deposit state.
type AppSessionsV1SubmitDepositStateResponse struct {
	// StateNodeSig is the Node's signature for the deposit state
	StateNodeSig string `json:"signature"`
}

// AppSessionsV1SubmitAppStateRequest submits an application session state update.
type AppSessionsV1SubmitAppStateRequest struct {
	// AppStateUpdate is the application session state update to be submitted
	AppStateUpdate AppStateUpdateV1 `json:"app_state_update"`
	// QuorumSigs is the signature quorum for the application session
	QuorumSigs []string `json:"quorum_sigs"`
}

// AppSessionsV1SubmitAppStateResponse returns the Node's signature for the new User state.
type AppSessionsV1SubmitAppStateResponse struct{}

// SignedAppStateUpdateV1 represents a signed application session state update.
type SignedAppStateUpdateV1 struct {
	// AppStateUpdate is the application session state update
	AppStateUpdate AppStateUpdateV1 `json:"app_state_update"`
	// QuorumSigs is the signature quorum for the application session
	QuorumSigs []string `json:"quorum_sigs"`
}

// AppSessionsV1GetAppDefinitionRequest retrieves the application definition for a specific app session.
type AppSessionsV1GetAppDefinitionRequest struct {
	// AppSessionID is the application session ID
	AppSessionID string `json:"app_session_id"`
}

// AppSessionsV1GetAppDefinitionResponse returns the application definition.
// Definition is nil when no app session exists for the given ID; this is a
// successful response, not an error.
type AppSessionsV1GetAppDefinitionResponse struct {
	// Definition is the application definition, or nil if absent.
	Definition *AppDefinitionV1 `json:"definition,omitempty"`
}

// AppSessionsV1GetAppSessionsRequest lists all application sessions for a participant with optional filtering.
type AppSessionsV1GetAppSessionsRequest struct {
	// AppSessionID filters by application session ID
	AppSessionID *string `json:"app_session_id,omitempty"`
	// Participant filters by participant wallet address
	Participant *string `json:"participant,omitempty"`
	// Status filters by status (open/closed)
	Status *string `json:"status,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// AppSessionsV1GetAppSessionsResponse returns the list of application sessions.
type AppSessionsV1GetAppSessionsResponse struct {
	// AppSessions is the list of application sessions
	AppSessions []AppSessionInfoV1 `json:"app_sessions"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// AppSessionsV1CreateAppSessionRequest creates a new application session between participants.
type AppSessionsV1CreateAppSessionRequest struct {
	// Definition is the application definition including participants and quorum
	Definition AppDefinitionV1 `json:"definition"`
	// SessionData is the optional JSON stringified session data
	SessionData string `json:"session_data"`
	// QuorumSigs is the list of participant signatures for the app session creation
	QuorumSigs []string `json:"quorum_sigs,omitempty"`
	// OwnerSig is the optional owner signature for app session creation if approval required by the app registry
	OwnerSig string `json:"owner_sig,omitempty"`
}

// AppSessionsV1CreateAppSessionResponse returns the created application session information.
type AppSessionsV1CreateAppSessionResponse struct {
	// AppSessionID is the created application session ID
	AppSessionID string `json:"app_session_id"`
	// Version is the initial version of the session
	Version string `json:"version"`
	// Status is the status of the session (closed)
	Status string `json:"status"`
}

// AppSessionsV1SubmitSessionKeyStateRequest submits an app session key state for registration,
// rotation/update, or revocation. A submit whose ExpiresAt is in the past (<= server's now)
// is treated as a revocation: the auth path stops accepting state signed by the key and the
// slot is freed against the per-user cap.
type AppSessionsV1SubmitSessionKeyStateRequest struct {
	// State contains the session key metadata and delegation information
	State AppSessionKeyStateV1 `json:"state"`
}

// AppSessionsV1SubmitSessionKeyStateResponse returns the result of session key state submission.
type AppSessionsV1SubmitSessionKeyStateResponse struct {
}

// AppSessionsV1GetLastKeyStatesRequest retrieves the latest session key states for a user with optional filtering by session key.
type AppSessionsV1GetLastKeyStatesRequest struct {
	// UserAddress is the user's wallet address
	UserAddress string  `json:"user_address"`
	SessionKey  *string `json:"session_key,omitempty"` // Optionally filter by SessionKey
	// IncludeInactive, when true, includes latest states whose expires_at is in the past
	// (expired or revoked). Defaults to false: only currently active states are returned.
	IncludeInactive *bool `json:"include_inactive,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// AppSessionsV1GetLastKeyStatesResponse returns the latest session key states for the user.
type AppSessionsV1GetLastKeyStatesResponse struct {
	// States is the list of latest session key states for the user, filtered by IncludeInactive.
	States []AppSessionKeyStateV1 `json:"states"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ============================================================================
// User Group - V1 API
// ============================================================================

// UserV1GetBalancesRequest retrieves the balances of the user in YN.
type UserV1GetBalancesRequest struct {
	// Wallet is the user's wallet address
	Wallet string `json:"wallet"`
}

// UserV1GetBalancesResponse returns the list of asset balances.
type UserV1GetBalancesResponse struct {
	// Balances is the list of asset balances
	Balances []BalanceEntryV1 `json:"balances"`
}

// UserV1GetTransactionsRequest retrieves ledger transaction history with optional filtering.
type UserV1GetTransactionsRequest struct {
	// Wallet filters by user's wallet address
	Wallet string `json:"wallet"`
	// Asset filters by asset symbol
	Asset *string `json:"asset,omitempty"`
	// TxType filters by transaction type
	TxType *core.TransactionType `json:"tx_type,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
	// FromTime is the start time filter (Unix timestamp)
	FromTime *uint64 `json:"from_time,omitempty"`
	// ToTime is the end time filter (Unix timestamp)
	ToTime *uint64 `json:"to_time,omitempty"`
}

// UserV1GetTransactionsResponse returns the list of transactions.
type UserV1GetTransactionsResponse struct {
	// Transactions is the list of transactions
	Transactions []TransactionV1 `json:"transactions"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ============================================================================
// Node Group - V1 API
// ============================================================================

// NodeV1PingRequest is a simple connectivity check.
type NodeV1PingRequest struct{}

// NodeV1PingResponse is the response to a ping request.
type NodeV1PingResponse struct{}

// NodeV1GetConfigRequest retrieves broker configuration and supported networks.
type NodeV1GetConfigRequest struct{}

// NodeV1GetConfigResponse returns the broker configuration.
type NodeV1GetConfigResponse struct {
	// NodeAddress is the node wallet address
	NodeAddress string `json:"node_address"`
	// NodeVersion is the node software version
	NodeVersion string `json:"node_version"`
	// SupportedSigValidators is the list of supported signature validators identifiers for state sig verification
	SupportedSigValidators []core.ChannelSignerType `json:"supported_sig_validators"`
	// Blockchains is the list of supported networks
	Blockchains []BlockchainInfoV1 `json:"blockchains"`
}

// NodeV1GetAssetsRequest retrieves all supported assets with optional chain filter.
type NodeV1GetAssetsRequest struct {
	// BlockchainID filters by blockchain network ID
	BlockchainID *string `json:"blockchain_id,omitempty"`
}

// NodeV1GetAssetsResponse returns the list of supported assets.
type NodeV1GetAssetsResponse struct {
	// Assets is the list of supported assets
	Assets []AssetV1 `json:"assets"`
}
