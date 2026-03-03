package app_session_v1

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
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

	RecordTransaction(tx core.Transaction) error

	// Channel state operations

	// LockUserState locks a user's balance for update, must be used within a transaction.
	// Returns the current balance.
	LockUserState(wallet, asset string) (decimal.Decimal, error)

	// CheckOpenChannel verifies if a user has an active channel for the given asset
	// and returns the approved signature validators if such a channel exists.
	CheckOpenChannel(wallet, asset string) (string, bool, error)
	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)
	StoreUserState(state core.State) error
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// App Session key state operations
	StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error
	GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error)
	GetLastAppSessionKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error)
	GetAppSessionKeyOwner(sessionKey, appSessionId string) (string, error)

	// Channel Session key state operations
	ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error)
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
