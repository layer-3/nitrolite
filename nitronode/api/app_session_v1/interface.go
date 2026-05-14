package app_session_v1

import (
	"github.com/layer-3/nitrolite/nitronode/action_gateway"
	"github.com/layer-3/nitrolite/nitronode/store/database"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// Store defines the persistence layer interface for app session management.
type Store interface {
	// App registry operations
	GetApp(appID string) (*app.AppInfoV1, error)

	// App session operations
	CreateAppSession(session app.AppSessionV1) error
	GetAppSession(sessionID string) (*app.AppSessionV1, error)
	GetAppSessions(appSessionID *string, participant *string, status app.AppSessionStatus, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error)
	UpdateAppSession(session app.AppSessionV1) error
	GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error)
	GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error)

	// Ledger operations
	RecordLedgerEntry(userWallet, accountID, asset string, amount decimal.Decimal) error

	// RecordTransaction records a transaction. applicationID is the client-declared
	// origin tag (rpc.ApplicationIDQueryParam); empty string is persisted as NULL.
	RecordTransaction(tx core.Transaction, applicationID string) error

	// Channel state operations

	// LockUserState locks a user's balance for update, must be used within a transaction.
	// Returns the current balance.
	LockUserState(wallet, asset string) (decimal.Decimal, error)

	// CheckActiveChannel verifies if a user has an active home channel for the given asset
	// and returns its approved signature validators and current status. A nil status means
	// no active channel exists. "Active" includes Void (DB-only, awaiting onchain confirmation)
	// and Open (materialized onchain); callers needing onchain materialization must additionally
	// require Status == core.ChannelStatusOpen.
	CheckActiveChannel(wallet, asset string) (string, *core.ChannelStatus, error)

	// GetHomeChannelStatus returns the current status of a home channel by ID.
	// Returns nil if no home channel exists for the given ID.
	GetHomeChannelStatus(channelID string) (*core.ChannelStatus, error)

	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)
	// StoreUserState persists a user state. applicationID is the client-declared
	// origin tag (rpc.ApplicationIDQueryParam); empty string is persisted as NULL.
	StoreUserState(state core.State, applicationID string) error
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// EnsureNoOngoingEscrowOperation validates that the user has no in-flight escrow
	// operation (escrow_lock, mutual_lock, or unfinalized escrow_deposit/escrow_withdraw)
	// that would prevent issuing a receiver-side state.
	EnsureNoOngoingEscrowOperation(wallet, asset string) error

	// App Session key state operations
	LockSessionKeyState(userAddress, sessionKey string, kind database.SessionKeyKind) (uint64, error)
	CountSessionKeysForUser(userAddress string) (uint32, error)
	StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error
	GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error)
	// GetLastAppSessionKeyStates retrieves the latest app session key states for a user,
	// optionally filtered by session key. When includeInactive is false, only non-expired
	// latest states are returned; when true, all latest states are returned regardless of
	// expiry. Results are paginated.
	GetLastAppSessionKeyStates(wallet string, sessionKey *string, includeInactive bool, limit, offset uint32) ([]app.AppSessionKeyStateV1, uint32, error)
	GetAppSessionKeyOwner(sessionKey, appSessionId string) (string, error)

	// Channel Session key state operations
	ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error)

	action_gateway.Store
}

type ActionGateway interface {
	// AllowAction checks if a user is allowed to perform a specific gated action based on their past activity and allowances.
	AllowAction(tx action_gateway.Store, userAddress string, gatedAction core.GatedAction) error
}

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(Store) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

// SigValidator validates cryptographic signatures on state transitions.
type SigValidator interface {
	// Recover recovers the wallet address from the signature and data.
	// Returns the recovered address or an error if the signature is invalid.
	Recover(data, sig []byte) (string, error)
	Verify(wallet string, data, sig []byte) error
}

// SigType identifies the signature validation algorithm to use.
type SigType string

// EcdsaSigType represents the ECDSA (Elliptic Curve Digital Signature Algorithm)
// validator, used for Ethereum-style signature verification.
const EcdsaSigType SigType = "ecdsa"

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)
}
