package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

type AppStateUpdateIntent uint8

const (
	AppStateUpdateIntentOperate AppStateUpdateIntent = iota
	AppStateUpdateIntentDeposit
	AppStateUpdateIntentWithdraw
	AppStateUpdateIntentClose
	AppStateUpdateIntentRebalance
)

func (intent AppStateUpdateIntent) String() string {
	switch intent {
	case AppStateUpdateIntentOperate:
		return "operate"
	case AppStateUpdateIntentDeposit:
		return "deposit"
	case AppStateUpdateIntentWithdraw:
		return "withdraw"
	case AppStateUpdateIntentClose:
		return "close"
	case AppStateUpdateIntentRebalance:
		return "rebalance"
	default:
		return "unknown"
	}
}

func (intent AppStateUpdateIntent) GatedAction() core.GatedAction {
	switch intent {
	case AppStateUpdateIntentOperate:
		return core.GatedActionAppSessionOperation
	case AppStateUpdateIntentDeposit:
		return core.GatedActionAppSessionDeposit
	case AppStateUpdateIntentWithdraw:
		return core.GatedActionAppSessionWithdrawal
	default:
		return ""
	}
}

type AppSessionStatus uint8

const (
	AppSessionStatusVoid AppSessionStatus = iota
	AppSessionStatusOpen
	AppSessionStatusClosed
)

func (status AppSessionStatus) String() string {
	switch status {
	case AppSessionStatusVoid:
		return ""
	case AppSessionStatusOpen:
		return "open"
	case AppSessionStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func (s *AppSessionStatus) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*s = AppSessionStatus(uint8(v))
		return nil
	case int32:
		*s = AppSessionStatus(uint8(v))
		return nil
	case int:
		*s = AppSessionStatus(uint8(v))
		return nil
	case string:
		return s.scanString(v)
	default:
		return fmt.Errorf("unsupported AppSessionStatus scan type %T", src)
	}
}

func (s *AppSessionStatus) scanString(v string) error {
	v = strings.TrimSpace(v)
	// if numeric
	if n, err := strconv.Atoi(v); err == nil {
		*s = AppSessionStatus(uint8(n))
		return nil
	}
	// else map names
	switch strings.ToLower(v) {
	case AppSessionStatusVoid.String():
		*s = AppSessionStatusVoid
	case AppSessionStatusOpen.String():
		*s = AppSessionStatusOpen
	case AppSessionStatusClosed.String():
		*s = AppSessionStatusClosed
	default:
		return fmt.Errorf("unknown AppSessionStatus %q", v)
	}
	return nil
}

// AppSessionV1 represents an application session in the V1 API.
type AppSessionV1 struct {
	SessionID     string
	ApplicationID string
	Participants  []AppParticipantV1
	Quorum        uint8
	Nonce         uint64
	Status        AppSessionStatus
	Version       uint64
	SessionData   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AppParticipantV1 represents the definition for an app participant.
type AppParticipantV1 struct {
	WalletAddress   string
	SignatureWeight uint8
}

// AppDefinitionV1 represents the definition for an app session.
type AppDefinitionV1 struct {
	ApplicationID string
	Participants  []AppParticipantV1
	Quorum        uint8
	Nonce         uint64
}

// AppSessionVersionV1 represents a session ID and version pair for rebalancing operations.
type AppSessionVersionV1 struct {
	SessionID string
	Version   uint64
}

// AppAllocationV1 represents the allocation of assets to a participant in an app session.
type AppAllocationV1 struct {
	Participant string
	Asset       string
	Amount      decimal.Decimal
}

// AppStateUpdateV1 represents the current state of an application session.
type AppStateUpdateV1 struct {
	AppSessionID string
	Intent       AppStateUpdateIntent
	Version      uint64
	Allocations  []AppAllocationV1
	SessionData  string
}

// SignedAppStateUpdateV1 represents a signed application session state update.
type SignedAppStateUpdateV1 struct {
	AppStateUpdate AppStateUpdateV1
	QuorumSigs     []string
}

// AppSessionInfoV1 represents information about an application session.
type AppSessionInfoV1 struct {
	AppSessionID  string
	AppDefinition AppDefinitionV1
	IsClosed      bool
	SessionData   string
	Version       uint64
	Allocations   []AppAllocationV1
}

// SessionKeyV1 represents a session key with spending allowances.
type SessionKeyV1 struct {
	ID          uint
	SessionKey  string
	Application string
	Allowances  []AssetAllowanceV1
	Scope       *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// AssetAllowanceV1 represents an asset allowance with usage tracking.
type AssetAllowanceV1 struct {
	Asset     string
	Allowance decimal.Decimal
	Used      decimal.Decimal
}

// PackCreateAppSessionRequestV1 packs the Definition and SessionData for signing using ABI encoding.
// This is used to generate a deterministic hash that participants sign when creating an app session.
func PackCreateAppSessionRequestV1(definition AppDefinitionV1, sessionData string) ([]byte, error) {
	// Define the participant tuple type
	participantType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "walletAddress", Type: "address"},
		{Name: "signatureWeight", Type: "uint8"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create participant type: %w", err)
	}

	// Define the arguments structure
	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},                        // application
		{Type: abi.Type{T: abi.SliceTy, Elem: &participantType}}, // participants array
		{Type: abi.Type{T: abi.UintTy, Size: 8}},                 // quorum (uint8)
		{Type: abi.Type{T: abi.UintTy, Size: 64}},                // nonce (uint64)
		{Type: abi.Type{T: abi.StringTy}},                        // sessionData
	}

	// Convert participants to the format expected by ABI packing
	participants := make([]struct {
		WalletAddress   common.Address
		SignatureWeight uint8
	}, len(definition.Participants))

	for i, p := range definition.Participants {
		participants[i] = struct {
			WalletAddress   common.Address
			SignatureWeight uint8
		}{
			WalletAddress:   common.HexToAddress(p.WalletAddress),
			SignatureWeight: p.SignatureWeight,
		}
	}

	// Pack the data using ABI encoding
	packed, err := args.Pack(
		definition.ApplicationID,
		participants,
		definition.Quorum,
		definition.Nonce,
		sessionData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack app session request: %w", err)
	}

	// Return the Keccak256 hash of the packed data
	return crypto.Keccak256(packed), nil
}

// PackAppStateUpdateV1 packs the AppStateUpdate for signing using ABI encoding.
// This is used to generate a deterministic hash that participants sign when updating an app session state.
func PackAppStateUpdateV1(stateUpdate AppStateUpdateV1) ([]byte, error) {
	// Define the allocation tuple type
	allocationType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "participant", Type: "address"},
		{Name: "asset", Type: "string"},
		{Name: "amount", Type: "string"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create allocation type: %w", err)
	}

	// Define the arguments structure
	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}},         // appSessionID (bytes32)
		{Type: abi.Type{T: abi.UintTy, Size: 8}},                // intent (uint8)
		{Type: abi.Type{T: abi.UintTy, Size: 64}},               // version (uint64)
		{Type: abi.Type{T: abi.SliceTy, Elem: &allocationType}}, // allocations array
		{Type: abi.Type{T: abi.StringTy}},                       // sessionData
	}

	// Convert allocations to the format expected by ABI packing
	allocations := make([]struct {
		Participant common.Address
		Asset       string
		Amount      string
	}, len(stateUpdate.Allocations))

	for i, a := range stateUpdate.Allocations {
		allocations[i] = struct {
			Participant common.Address
			Asset       string
			Amount      string
		}{
			Participant: common.HexToAddress(a.Participant),
			Asset:       a.Asset,
			Amount:      a.Amount.String(),
		}
	}

	// Convert app session ID from hex string to bytes32
	appSessionIDHash := common.HexToHash(stateUpdate.AppSessionID)

	// Pack the data using ABI encoding
	packed, err := args.Pack(
		appSessionIDHash,
		stateUpdate.Intent,
		stateUpdate.Version,
		allocations,
		stateUpdate.SessionData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack app state update: %w", err)
	}

	// Return the Keccak256 hash of the packed data
	return crypto.Keccak256(packed), nil
}

// GenerateAppSessionIDV1 generates a deterministic app session ID from the definition using ABI encoding.
func GenerateAppSessionIDV1(definition AppDefinitionV1) (string, error) {
	// Define the participant tuple type
	participantType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "walletAddress", Type: "address"},
		{Name: "signatureWeight", Type: "uint8"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create participant type: %w", err)
	}

	// Define the arguments structure
	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},                        // application
		{Type: abi.Type{T: abi.SliceTy, Elem: &participantType}}, // participants array
		{Type: abi.Type{T: abi.UintTy, Size: 8}},                 // quorum (uint8)
		{Type: abi.Type{T: abi.UintTy, Size: 64}},                // nonce (uint64)
	}

	// Convert participants to the format expected by ABI packing
	participants := make([]struct {
		WalletAddress   common.Address
		SignatureWeight uint8
	}, len(definition.Participants))

	for i, p := range definition.Participants {
		participants[i] = struct {
			WalletAddress   common.Address
			SignatureWeight uint8
		}{
			WalletAddress:   common.HexToAddress(p.WalletAddress),
			SignatureWeight: p.SignatureWeight,
		}
	}

	// Pack the data using ABI encoding
	packed, err := args.Pack(
		definition.ApplicationID,
		participants,
		definition.Quorum,
		definition.Nonce,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack app definition: %w", err)
	}

	// Return the Keccak256 hash as hex string
	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GenerateRebalanceBatchIDV1 creates a deterministic batch ID from session versions using ABI encoding.
// The batch ID is generated by hashing the list of (sessionID, version) pairs.
func GenerateRebalanceBatchIDV1(sessionVersions []AppSessionVersionV1) (string, error) {
	// Define the session version tuple type
	sessionVersionType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sessionID", Type: "bytes32"},
		{Name: "version", Type: "uint64"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session version type: %w", err)
	}

	// Define the arguments structure
	args := abi.Arguments{
		{Type: abi.Type{T: abi.SliceTy, Elem: &sessionVersionType}}, // session versions array
	}

	// Convert session versions to the format expected by ABI packing
	sessionVersionsArray := make([]struct {
		SessionID common.Hash
		Version   uint64
	}, len(sessionVersions))

	for i, sv := range sessionVersions {
		sessionVersionsArray[i] = struct {
			SessionID common.Hash
			Version   uint64
		}{
			SessionID: common.HexToHash(sv.SessionID),
			Version:   sv.Version,
		}
	}

	// Pack the data using ABI encoding
	packed, err := args.Pack(sessionVersionsArray)
	if err != nil {
		return "", fmt.Errorf("failed to pack session versions: %w", err)
	}

	// Return the Keccak256 hash as hex string
	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GenerateRebalanceTransactionIDV1 creates a deterministic transaction ID for a rebalance transaction using ABI encoding.
func GenerateRebalanceTransactionIDV1(batchID, sessionID, asset string) (string, error) {
	// Define the arguments structure
	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // batchID (bytes32)
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // sessionID (bytes32)
		{Type: abi.Type{T: abi.StringTy}},               // asset (string)
	}

	// Pack the data using ABI encoding
	packed, err := args.Pack(
		common.HexToHash(batchID),
		common.HexToHash(sessionID),
		asset,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack rebalance transaction data: %w", err)
	}

	// Return the Keccak256 hash as hex string
	return crypto.Keccak256Hash(packed).Hex(), nil
}
