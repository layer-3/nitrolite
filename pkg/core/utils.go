package core

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

const (
	// ChannelHubVersion is the version of the ChannelHub contract that this code is compatible with.
	// This version is encoded as the first byte of the channelId to prevent replay attacks
	// across different ChannelHub deployments on the same chain.
	ChannelHubVersion uint8 = 1
)

var (
	uint8Type, _   = abi.NewType("uint8", "", nil)
	uint32Type, _  = abi.NewType("uint32", "", nil)
	uint64Type, _  = abi.NewType("uint64", "", nil)
	uint256Type, _ = abi.NewType("uint256", "", nil)
)

func NormalizeHexAddress(s string) (string, error) {
	s = strings.ToLower(s)

	if !strings.HasPrefix(s, "0x") || len(s) != 42 {
		return "", fmt.Errorf("invalid hex address: incorrect length, expected 42 characters including 0x prefix")
	}

	for i := 2; i < len(s); i++ {
		c := s[i]
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f')) {
			return "", fmt.Errorf("invalid hex address: character '%c' at position %d is not a valid hexadecimal character", c, i)
		}
	}

	return s, nil
}

func TransitionToIntent(transition Transition) uint8 {
	switch transition.Type {
	case TransitionTypeTransferSend,
		TransitionTypeTransferReceive,
		TransitionTypeCommit,
		TransitionTypeRelease:
		return INTENT_OPERATE
	case TransitionTypeFinalize:
		return INTENT_CLOSE
	case TransitionTypeHomeDeposit:
		return INTENT_DEPOSIT
	case TransitionTypeHomeWithdrawal:
		return INTENT_WITHDRAW
	case TransitionTypeMutualLock:
		return INTENT_INITIATE_ESCROW_DEPOSIT
	case TransitionTypeEscrowDeposit:
		return INTENT_FINALIZE_ESCROW_DEPOSIT
	case TransitionTypeEscrowLock:
		return INTENT_INITIATE_ESCROW_WITHDRAWAL
	case TransitionTypeEscrowWithdraw:
		return INTENT_FINALIZE_ESCROW_WITHDRAWAL
	case TransitionTypeMigrate:
		return INTENT_INITIATE_MIGRATION
	default:
		return INTENT_OPERATE
	}
	// TODO: Add:
	// FINALIZE_MIGRATION.
}

// ValidateDecimalPrecision validates that an amount doesn't exceed the maximum allowed decimal places.
func ValidateDecimalPrecision(amount decimal.Decimal, maxDecimals uint8) error {
	if amount.Exponent() < -int32(maxDecimals) {
		return fmt.Errorf("amount exceeds maximum decimal precision: max %d decimals allowed, got %d", maxDecimals, -amount.Exponent())
	}
	return nil
}

// DecimalToBigInt converts a decimal.Decimal amount to *big.Int scaled to the token's smallest unit.
// For example, 1.23 USDC (6 decimals) becomes 1230000.
// This is used when preparing amounts for smart contract calls.
func DecimalToBigInt(amount decimal.Decimal, decimals uint8) (*big.Int, error) {
	// 1. Calculate the multiplier (e.g., 10^6)
	multiplier := decimal.New(1, int32(decimals))

	// 2. Scale the amount
	scaled := amount.Mul(multiplier)

	if !scaled.IsInteger() {
		return nil, fmt.Errorf("amount %s exceeds maximum decimal precision: max %d decimals allowed", amount.String(), decimals)
	}

	// 4. Safe to convert
	return scaled.BigInt(), nil
}

// getHomeChannelID is the internal implementation that generates a unique identifier for a primary channel
// based on its definition and version. This matches the Solidity getChannelId function which computes
// keccak256(abi.encode(ChannelDefinition)) and then sets the first byte to the version.
// The metadata is derived from the asset: first 8 bytes of keccak256(asset) padded to 32 bytes.
func getHomeChannelID(node, user, asset string, nonce uint64, challengeDuration uint32, approvedSigValidators string, channelHubVersion uint8) (string, error) {
	// Generate metadata from asset
	userAddr := common.HexToAddress(user)
	nodeAddr := common.HexToAddress(node)
	metadata := GenerateChannelMetadata(asset)

	// Define the struct to match Solidity's ChannelDefinition
	type channelDefinition struct {
		ChallengeDuration           uint32
		User                        common.Address
		Node                        common.Address
		Nonce                       uint64
		ApprovedSignatureValidators *big.Int
		Metadata                    [32]byte
	}

	def := channelDefinition{
		ChallengeDuration:           challengeDuration,
		User:                        userAddr,
		Node:                        nodeAddr,
		Nonce:                       nonce,
		ApprovedSignatureValidators: hexToBigInt(approvedSigValidators),
		Metadata:                    metadata,
	}

	// Define the struct type for ABI encoding
	channelDefType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "challengeDuration", Type: "uint32"},
		{Name: "user", Type: "address"},
		{Name: "node", Type: "address"},
		{Name: "nonce", Type: "uint64"},
		{Name: "approvedSignatureValidators", Type: "uint256"},
		{Name: "metadata", Type: "bytes32"},
	})
	if err != nil {
		return "", err
	}

	args := abi.Arguments{
		{Type: channelDefType},
	}

	packed, err := args.Pack(def)
	if err != nil {
		return "", err
	}

	// Calculate base channelId
	baseId := crypto.Keccak256Hash(packed)

	// Set the first byte (most significant byte) to the version
	versionedId := baseId
	versionedId[0] = channelHubVersion

	return versionedId.Hex(), nil
}

// GetHomeChannelID generates a unique identifier for a primary channel based on its definition.
// It uses the configured ChannelHubVersion to ensure compatibility with the deployed ChannelHub contract.
// The channelId includes version information to prevent replay attacks across different ChannelHub deployments.
func GetHomeChannelID(node, user, asset string, nonce uint64, challengeDuration uint32, approvedSigValidators string) (string, error) {
	return getHomeChannelID(node, user, asset, nonce, challengeDuration, approvedSigValidators, ChannelHubVersion)
}

// hexToBigInt converts a hex string (with optional 0x prefix) to *big.Int.
func hexToBigInt(s string) *big.Int {
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(s, "0x"), 16)
	return n
}

// GetEscrowChannelID derives an escrow-specific channel ID based on a home channel and state version.
// This matches the Solidity getEscrowId function which computes keccak256(abi.encode(channelId, version)).
func GetEscrowChannelID(homeChannelID string, stateVersion uint64) (string, error) {
	rawHomeChannelID := common.HexToHash(homeChannelID)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelId
		{Type: uint64Type}, // version
	}

	packed, err := args.Pack(rawHomeChannelID, stateVersion)
	if err != nil {
		return "", err
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetStateID creates a unique hash representing a specific snapshot of a user's wallet and asset state.
func GetStateID(userWallet, asset string, epoch, version uint64) string {
	userAddr := common.HexToAddress(userWallet)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}}, // userWallet
		{Type: abi.Type{T: abi.StringTy}},  // asset symbol/string
		{Type: uint256Type},                // epoch
		{Type: uint256Type},                // version
	}

	packed, _ := args.Pack(
		userAddr,
		asset,
		new(big.Int).SetUint64(epoch),
		new(big.Int).SetUint64(version),
	)

	return crypto.Keccak256Hash(packed).Hex()
}

func GetStateTransitionHash(transition Transition) ([32]byte, error) {
	hash := [32]byte{}

	type contractTransition struct {
		Type      uint8
		TxId      [32]byte
		AccountId [32]byte
		Amount    string
	}

	transitionType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "type", Type: "uint8"},
		{Name: "txId", Type: "bytes32"},
		{Name: "accountId", Type: "bytes32"},
		{Name: "amount", Type: "string"},
	})
	if err != nil {
		return hash, fmt.Errorf("failed to create transition type: %w", err)
	}

	args := abi.Arguments{
		{Type: transitionType},
	}

	txIdBytes, err := hexToBytes32(transition.TxID)
	if err != nil {
		return hash, fmt.Errorf("invalid txId: %w", err)
	}

	accountIdBytes, err := parseAccountIdToBytes32(transition.AccountID)
	if err != nil {
		return hash, fmt.Errorf("invalid accountId: %w", err)
	}

	payload := contractTransition{
		Type:      uint8(transition.Type),
		TxId:      txIdBytes,
		AccountId: accountIdBytes,
		Amount:    transition.Amount.String(),
	}

	packed, err := args.Pack(payload)
	if err != nil {
		return hash, fmt.Errorf("failed to pack transition: %w", err)
	}

	hash = crypto.Keccak256Hash(packed)
	return hash, nil
}

// GetSenderTransactionID calculates and returns a unique transaction ID reference for actions initiated by user.
func GetSenderTransactionID(toAccount string, senderNewStateID string) (string, error) {
	return getTransactionID(toAccount, senderNewStateID)
}

// GetReceiverTransactionID calculates and returns a unique transaction ID reference for actions initiated by node.
func GetReceiverTransactionID(fromAccount, receiverNewStateID string) (string, error) {
	return getTransactionID(fromAccount, receiverNewStateID)
}

func getTransactionID(account, newStateID string) (string, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}},
	}

	receiverStateID := common.HexToHash(newStateID)
	packed, err := args.Pack(account, receiverStateID)
	if err != nil {
		return "", fmt.Errorf("failed to pack transaction ID arguments: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GenerateChannelMetadata creates metadata from an asset by taking the first 8 bytes of keccak256(asset)
// and padding the rest with zeros to make a 32-byte array.
func GenerateChannelMetadata(asset string) [32]byte {
	assetHash := crypto.Keccak256Hash([]byte(asset))
	var metadata [32]byte
	copy(metadata[:8], assetHash[:8])
	return metadata
}

// hexToBytes32 converts a hex string (with or without 0x prefix) to [32]byte
func hexToBytes32(hexStr string) ([32]byte, error) {
	var result [32]byte

	// Use common.HexToHash which handles 0x prefix and validates length
	hash := common.HexToHash(hexStr)
	copy(result[:], hash[:])

	return result, nil
}

// parseAccountIdToBytes32 converts an account ID (address or hash) to [32]byte
// - If the input is a 20-byte address (40 hex chars), it's left-padded with zeros
// - If the input is a 32-byte hash (64 hex chars), it's used as-is
// In Ethereum, when an address is stored in bytes32, it occupies the rightmost 20 bytes,
// with the leftmost 12 bytes being zeros.
func parseAccountIdToBytes32(accountId string) ([32]byte, error) {
	var result [32]byte

	// Check if it's an address (20 bytes) or hash (32 bytes)
	if common.IsHexAddress(accountId) {
		// It's an address - convert to address type and then to bytes32
		addr := common.HexToAddress(accountId)
		// Left-pad with zeros: [12 zeros][20 address bytes]
		copy(result[12:], addr[:])
	} else {
		// Try to parse as a 32-byte hash
		hash := common.HexToHash(accountId)
		copy(result[:], hash[:])
	}

	return result, nil
}
