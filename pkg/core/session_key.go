package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// ChannelSessionKeyStateV1 represents the state of a session key.
type ChannelSessionKeyStateV1 struct {
	// ID Hash(user_address + session_key + version)
	UserAddress string    `json:"user_address"` // UserAddress is the user wallet address
	SessionKey  string    `json:"session_key"`  // SessionKey is the session key address for delegation
	Version     uint64    `json:"version"`      // Version is the version of the session key format
	Assets      []string  `json:"assets"`       // Assets associated with this session key
	ExpiresAt   time.Time `json:"expires_at"`   // Expiration time as unix timestamp of this session key
	UserSig     string    `json:"user_sig"`     // UserSig is the user's signature over the session key metadata to authorize the registration/update of the session key
}

type VerifyChannelSessionKePermissionsV1 func(walletAddr, sessionKeyAddr, metadataHash string) (bool, error)

type ChannelSessionKeySignerV1 struct {
	sign.Signer

	metadataHash common.Hash
	authSig      []byte
}

func NewChannelSessionKeySignerV1(signer sign.Signer, metadataHash, authSig string) (*ChannelSessionKeySignerV1, error) {
	authSigBytes, err := hexutil.Decode(authSig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth signature: %w", err)
	}

	return &ChannelSessionKeySignerV1{
		Signer:       signer,
		metadataHash: common.HexToHash(metadataHash),
		authSig:      authSigBytes,
	}, nil
}

func (s *ChannelSessionKeySignerV1) Sign(data []byte) (sign.Signature, error) {
	sessionKeySig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	fullSig, err := encodeChannelSessionKeySignature(
		channelSessionKeyAuthorization{
			SessionKey:    common.HexToAddress(s.Signer.PublicKey().Address().String()),
			MetadataHash:  s.metadataHash,
			AuthSignature: s.authSig,
		},
		sessionKeySig,
	)
	if err != nil {
		return sign.Signature{}, fmt.Errorf("failed to encode session key signature: %w", err)
	}

	return append([]byte{byte(ChannelSignerType_SessionKey)}, fullSig...), nil
}

func (s *ChannelSessionKeySignerV1) Type() ChannelSignerType {
	return ChannelSignerType_SessionKey
}

// PackChannelKeyStateV1 packs the session key state for signing using ABI encoding.
// This is used to generate a deterministic hash that the user signs when registering/updating a session key.
// The user_sig field is excluded from packing since it is the signature itself.
func PackChannelKeyStateV1(sessionKey string, metadataHash common.Hash) ([]byte, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},              // session_key
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // hashed metadata
	}

	packed, err := args.Pack(
		common.HexToAddress(sessionKey),
		metadataHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack session key state: %w", err)
	}

	return packed, nil
}

func GetChannelSessionKeyAuthMetadataHashV1(version uint64, assets []string, expiresAt int64) (common.Hash, error) {
	stringArrayType, err := abi.NewType("string[]", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create string array type: %w", err)
	}

	metadtataArgs := abi.Arguments{
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // version
		{Type: stringArrayType},                   // assets
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // expires_at (unix timestamp)
	}

	packedMetadataArgs, err := metadtataArgs.Pack(
		version,
		assets,
		uint64(expiresAt),
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack metadata args: %w", err)
	}

	hashedMetadata := crypto.Keccak256Hash(packedMetadataArgs)
	return hashedMetadata, nil
}

func ValidateChannelSessionKeyAuthSigV1(state ChannelSessionKeyStateV1) error {
	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to get metadata hash: %w", err)
	}

	packed, err := PackChannelKeyStateV1(state.SessionKey, metadataHash)
	if err != nil {
		return fmt.Errorf("failed to pack session key state: %w", err)
	}

	authSigBytes, err := hexutil.Decode(state.UserSig)
	if err != nil {
		return fmt.Errorf("failed to decode user signature: %w", err)
	}

	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		return fmt.Errorf("failed to create address recoverer: %w", err)
	}

	recoveredAddr, err := recoverer.RecoverAddress(packed, authSigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover address from signature: %w", err)
	}

	if !strings.EqualFold(recoveredAddr.String(), state.UserAddress) {
		return fmt.Errorf("invalid signature: recovered address %s does not match wallet %s", recoveredAddr.String(), state.UserAddress)
	}

	return nil
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

// channelSessionKeyAuthorization matches the Solidity SessionKeyAuthorization struct.
type channelSessionKeyAuthorization struct {
	SessionKey    common.Address
	MetadataHash  [32]byte
	AuthSignature []byte
}

func encodeChannelSessionKeySignature(skAuth channelSessionKeyAuthorization, skSignature []byte) ([]byte, error) {
	args, err := getChannelSessionKeyArgsV1()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel session key args: %w", err)
	}

	packed, err := args.Pack(skAuth, skSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to pack session key signature: %w", err)
	}

	return packed, nil
}

func decodeChannelSessionKeySignature(data []byte) (skAuth channelSessionKeyAuthorization, skSignature []byte, err error) {
	args, err := getChannelSessionKeyArgsV1()
	if err != nil {
		return skAuth, nil, fmt.Errorf("failed to get channel session key args: %w", err)
	}

	values, err := args.Unpack(data)
	if err != nil {
		return skAuth, nil, fmt.Errorf("failed to unpack session key signature: %w", err)
	}

	if len(values) != 2 {
		return skAuth, nil, fmt.Errorf("expected 2 values from unpack, got %d", len(values))
	}

	// The tuple unpacks to an anonymous struct; use reflect to access fields
	authStruct := reflect.ValueOf(values[0])
	if authStruct.Kind() != reflect.Struct || authStruct.NumField() != 3 {
		return skAuth, nil, fmt.Errorf("unexpected skAuth structure: kind=%s, fields=%d", authStruct.Kind(), authStruct.NumField())
	}

	sessionKey, ok := authStruct.Field(0).Interface().(common.Address)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for sessionKey: %T", authStruct.Field(0).Interface())
	}
	metadataHash, ok := authStruct.Field(1).Interface().([32]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for metadataHash: %T", authStruct.Field(1).Interface())
	}
	authSignature, ok := authStruct.Field(2).Interface().([]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for authSignature: %T", authStruct.Field(2).Interface())
	}

	skAuth = channelSessionKeyAuthorization{
		SessionKey:    sessionKey,
		MetadataHash:  metadataHash,
		AuthSignature: authSignature,
	}

	skSignature, ok = values[1].([]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for skSignature: %T", values[1])
	}

	return skAuth, skSignature, nil
}

func getChannelSessionKeyArgsV1() (*abi.Arguments, error) {
	skAuthType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sessionKey", Type: "address"},
		{Name: "metadataHash", Type: "bytes32"},
		{Name: "authSignature", Type: "bytes"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create skAuth type: %w", err)
	}

	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes type: %w", err)
	}

	return &abi.Arguments{
		{Type: skAuthType},
		{Type: bytesType},
	}, nil
}
