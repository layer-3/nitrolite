package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// AppSessionKeyStateV1 represents the state of a session key.
type AppSessionKeyStateV1 struct {
	// ID Hash(user_address + session_key + version)
	// UserAddress is the user wallet address
	UserAddress string
	// SessionKey is the session key address for delegation
	SessionKey string
	// Version is the version of the session key format
	Version uint64
	// ApplicationID is the application IDs associated with this session key
	ApplicationIDs []string
	// AppSessionID is the application session IDs associated with this session key
	AppSessionIDs []string
	// ExpiresAt is Unix timestamp in seconds indicating when the session key expires
	ExpiresAt time.Time
	// UserSig is the user's signature over the session key metadata to authorize the registration/update of the session key
	UserSig string
}

// GenerateSessionKeyStateIDV1 generates a deterministic ID from user_address, session_key, and version.
func GenerateSessionKeyStateIDV1(userAddress, sessionKey string, version uint64) (string, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},        // user_address
		{Type: abi.Type{T: abi.AddressTy}},        // session_key
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // version
	}

	packed, err := args.Pack(
		common.HexToAddress(userAddress),
		common.HexToAddress(sessionKey),
		version,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack session key state ID: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// PackAppSessionKeyStateV1 packs the session key state for signing using ABI encoding.
// This is used to generate a deterministic hash that the user signs when registering/updating a session key.
// The user_sig field is excluded from packing since it is the signature itself.
func PackAppSessionKeyStateV1(state AppSessionKeyStateV1) ([]byte, error) {
	bytes32ArrayType, err := abi.NewType("bytes32[]", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes32 array type: %w", err)
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},        // user_address
		{Type: abi.Type{T: abi.AddressTy}},        // session_key
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // version
		{Type: bytes32ArrayType},                  // application_ids (bytes32[])
		{Type: bytes32ArrayType},                  // app_session_ids (bytes32[])
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // expires_at (unix timestamp)
	}

	applicationIDHashes := make([][32]byte, len(state.ApplicationIDs))
	for i, id := range state.ApplicationIDs {
		applicationIDHashes[i] = common.HexToHash(id)
	}

	appSessionIDHashes := make([][32]byte, len(state.AppSessionIDs))
	for i, id := range state.AppSessionIDs {
		appSessionIDHashes[i] = common.HexToHash(id)
	}

	packed, err := args.Pack(
		common.HexToAddress(state.UserAddress),
		common.HexToAddress(state.SessionKey),
		state.Version,
		applicationIDHashes,
		appSessionIDHashes,
		uint64(state.ExpiresAt.Unix()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack session key state: %w", err)
	}

	return crypto.Keccak256(packed), nil
}

type AppSessionSignerTypeV1 uint8

const (
	AppSessionSignerTypeV1_Wallet     AppSessionSignerTypeV1 = 0xA1
	AppSessionSignerTypeV1_SessionKey AppSessionSignerTypeV1 = 0xA2
)

func (t AppSessionSignerTypeV1) String() string {
	switch t {
	case AppSessionSignerTypeV1_Wallet:
		return "wallet"
	case AppSessionSignerTypeV1_SessionKey:
		return "session_key"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

type AppSessionSignerV1 struct {
	signerType AppSessionSignerTypeV1
	sign.Signer
}

func NewAppSessionWalletSignerV1(signer sign.Signer) (*AppSessionSignerV1, error) {
	return newAppSessionSignerV1(AppSessionSignerTypeV1_Wallet, signer)
}

func NewAppSessionKeySignerV1(signer sign.Signer) (*AppSessionSignerV1, error) {
	return newAppSessionSignerV1(AppSessionSignerTypeV1_SessionKey, signer)
}

func newAppSessionSignerV1(signerType AppSessionSignerTypeV1, signer sign.Signer) (*AppSessionSignerV1, error) {
	if signerType != AppSessionSignerTypeV1_Wallet && signerType != AppSessionSignerTypeV1_SessionKey {
		return nil, fmt.Errorf("invalid signer type: %d", signerType)
	}

	return &AppSessionSignerV1{
		signerType: signerType,
		Signer:     signer,
	}, nil
}

func (s *AppSessionSignerV1) Sign(data []byte) (sign.Signature, error) {
	sig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	return append([]byte{byte(s.signerType)}, sig...), nil
}

type AppSessionKeyValidatorV1 struct {
	recoverer          sign.AddressRecoverer
	getSessionKeyOwner GetAppSessionKeyOwnerFuncV1
}

type GetAppSessionKeyOwnerFuncV1 func(sessionKeyAddr string) (string, error)

func NewAppSessionKeySigValidatorV1(ownerGetter GetAppSessionKeyOwnerFuncV1) *AppSessionKeyValidatorV1 {
	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create address recoverer: %v", err))
	}

	return &AppSessionKeyValidatorV1{
		recoverer:          recoverer,
		getSessionKeyOwner: ownerGetter,
	}
}

func (s *AppSessionKeyValidatorV1) Recover(data, sig []byte) (string, error) {
	if len(sig) < 1 {
		return "", fmt.Errorf("invalid signature: too short")
	}

	signerType := AppSessionSignerTypeV1(sig[0])
	switch signerType {
	case AppSessionSignerTypeV1_Wallet:
		addr, err := s.recoverer.RecoverAddress(data, sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address: %w", err)
		}
		return addr.String(), nil
	case AppSessionSignerTypeV1_SessionKey:
		sessionKeyAddr, err := s.recoverer.RecoverAddress(data, sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to recover session key address: %w", err)
		}

		return s.getSessionKeyOwner(sessionKeyAddr.String())
	default:
		return "", fmt.Errorf("invalid signature: unknown signer type %d", signerType)
	}
}

func (s *AppSessionKeyValidatorV1) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if !strings.EqualFold(address, wallet) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
