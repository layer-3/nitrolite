package core

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/sign"
)

type ChannelSignerType uint8

const (
	ChannelSignerType_Default    ChannelSignerType = 0x00
	ChannelSignerType_SessionKey ChannelSignerType = 0x01
)

func (t ChannelSignerType) String() string {
	switch t {
	case ChannelSignerType_Default:
		return "default"
	case ChannelSignerType_SessionKey:
		return "session_key"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

var (
	ChannelSignerTypes = []ChannelSignerType{
		ChannelSignerType_Default,
		ChannelSignerType_SessionKey,
	}
)

type ChannelSigner interface {
	sign.Signer
	Type() ChannelSignerType
}

// hexToBitmask parses a hex string (with optional 0x prefix) into a [32]byte big-endian bitmask.
func hexToBitmask(s string) ([32]byte, bool) {
	var val [32]byte
	b, err := hexutil.Decode(s)
	if err != nil || len(b) == 0 || len(b) > 32 {
		return val, false
	}
	copy(val[32-len(b):], b)
	return val, true
}

// signerTypesToBitmask builds a [32]byte big-endian bitmask from a slice of ChannelSignerType.
func signerTypesToBitmask(types []ChannelSignerType) [32]byte {
	var mask [32]byte
	for _, t := range types {
		idx := uint8(t)
		mask[31-idx/8] |= 1 << (idx % 8)
	}
	return mask
}

// bitmaskToHex converts a [32]byte bitmask to a compact hex string with 0x prefix.
func bitmaskToHex(mask [32]byte) string {
	for i := 0; i < 32; i++ {
		if mask[i] != 0 {
			return hexutil.Encode(mask[i:])
		}
	}
	return "0x00"
}

func IsChannelSignerSupported(approvedSigValidators string, signerType ChannelSignerType) bool {
	// Default signer is always supported, matching the smart contract behavior.
	if signerType == ChannelSignerType_Default {
		return true
	}

	// Mirrors Solidity: (approvedSigValidators >> signerType) & 1 == 1
	val, ok := hexToBitmask(approvedSigValidators)
	if !ok {
		return false
	}
	bitIndex := uint8(signerType)
	byteIndex := 31 - bitIndex/8
	bitOffset := bitIndex % 8
	return (val[byteIndex]>>bitOffset)&1 == 1
}

// SignerValidatorsSupported checks that every bit in channelValidators is
// covered by the node's supported ChannelSignerTypes.
func SignerValidatorsSupported(channelValidators string) bool {
	inc, ok := hexToBitmask(channelValidators)
	if !ok {
		return false
	}
	node := signerTypesToBitmask(ChannelSignerTypes)
	for i := 0; i < 32; i++ {
		if node[i]&inc[i] != inc[i] {
			return false
		}
	}
	return true
}

// BuildSigValidatorsBitmap constructs a hex string bitmap from a slice of ChannelSignerType.
// Each signer type sets a bit at its corresponding position in a 256-bit value.
func BuildSigValidatorsBitmap(signerTypes []ChannelSignerType) string {
	return bitmaskToHex(signerTypesToBitmask(signerTypes))
}

type ChannelDefaultSigner struct {
	sign.Signer
}

func NewChannelDefaultSigner(signer sign.Signer) (*ChannelDefaultSigner, error) {
	return &ChannelDefaultSigner{
		Signer: signer,
	}, nil
}

func (s *ChannelDefaultSigner) Sign(data []byte) (sign.Signature, error) {
	sig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	return append([]byte{byte(ChannelSignerType_Default)}, sig...), nil
}

func (s *ChannelDefaultSigner) Type() ChannelSignerType {
	return ChannelSignerType_Default
}

type ChannelSigValidator struct {
	recoverer         sign.AddressRecoverer
	verifyPermissions VerifyChannelSessionKePermissionsV1
}

func NewChannelSigValidator(permissionsVerifier VerifyChannelSessionKePermissionsV1) *ChannelSigValidator {
	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create address recoverer: %v", err))
	}

	return &ChannelSigValidator{
		recoverer:         recoverer,
		verifyPermissions: permissionsVerifier,
	}
}

func GetSignerType(sig []byte) (ChannelSignerType, error) {
	if len(sig) < 1 {
		return 0, fmt.Errorf("invalid signature: too short")
	}
	return ChannelSignerType(sig[0]), nil
}

func (s *ChannelSigValidator) Recover(data, sig []byte) (string, error) {
	if len(sig) < 1 {
		return "", fmt.Errorf("invalid signature: too short")
	}

	signerType := ChannelSignerType(sig[0])
	switch signerType {
	case ChannelSignerType_Default:
		addr, err := s.recoverer.RecoverAddress(data, sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address: %w", err)
		}
		return addr.String(), nil
	case ChannelSignerType_SessionKey:
		// Decode: (SessionKeyAuthorization memory skAuth, bytes memory skSignature) =
		//     abi.decode(signature, (SessionKeyAuthorization, bytes));
		skAuth, skSignature, err := decodeChannelSessionKeySignature(sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to decode session key signature: %w", err)
		}

		// Step 1: Verify participant authorized this session key
		// authMessage = _toSigningData(skAuth) = abi.encode(skAuth.sessionKey, skAuth.metadataHash)
		packedAuth, err := PackChannelKeyStateV1(skAuth.SessionKey.Hex(), skAuth.MetadataHash)
		if err != nil {
			return "", fmt.Errorf("failed to pack auth data: %w", err)
		}

		walletAddr, err := s.recoverer.RecoverAddress(packedAuth, skAuth.AuthSignature)
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address from auth signature: %w", err)
		}

		// Step 2: Verify session key signed the state data
		sessionKeyAddr, err := s.recoverer.RecoverAddress(data, skSignature)
		if err != nil {
			return "", fmt.Errorf("failed to recover session key address: %w", err)
		}

		if !strings.EqualFold(sessionKeyAddr.String(), skAuth.SessionKey.Hex()) {
			return "", fmt.Errorf("session key mismatch: recovered %s, expected %s", sessionKeyAddr.String(), skAuth.SessionKey.Hex())
		}

		ok, err := s.verifyPermissions(walletAddr.String(), sessionKeyAddr.String(), common.Hash(skAuth.MetadataHash).String())
		if err != nil {
			return "", err
		}

		if !ok {
			return "", fmt.Errorf("session key does not have permission to sign for this data")
		}
		// VerifyChannelSessionKeyPermissions(walletAddr, sessionKey, asset, metadataHash) (bool, error)

		return walletAddr.String(), nil
	default:
		return "", fmt.Errorf("invalid signature: unknown signer type %d", signerType)
	}
}

func (s *ChannelSigValidator) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if !strings.EqualFold(address, wallet) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
