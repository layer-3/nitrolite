package evm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
)

// ========= hexToBytes32 Tests =========

func TestHexToBytes32_Success(t *testing.T) {
	t.Parallel()
	// Create a valid 32-byte hex string
	expected := [32]byte{}
	for i := 0; i < 32; i++ {
		expected[i] = byte(i)
	}
	hexStr := hexutil.Encode(expected[:])

	result, err := hexToBytes32(hexStr)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestHexToBytes32_InvalidHex(t *testing.T) {
	t.Parallel()
	invalidHex := "zzzzz"
	_, err := hexToBytes32(invalidHex)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode hex string")
}

func TestHexToBytes32_InvalidLength(t *testing.T) {
	t.Parallel()
	// 16 bytes instead of 32
	shortHex := hexutil.Encode(make([]byte, 16))
	_, err := hexToBytes32(shortHex)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid length: expected 32 bytes, got 16")

	// 64 bytes instead of 32
	longHex := hexutil.Encode(make([]byte, 64))
	_, err = hexToBytes32(longHex)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid length: expected 32 bytes, got 64")
}

// ========= coreDefToContractDef Tests =========

func TestCoreDefToContractDef_Success(t *testing.T) {
	t.Parallel()
	asset := "ETH"
	userWallet := "0x1234567890123456789012345678901234567890"
	nodeAddress := common.HexToAddress("0x9876543210987654321098765432109876543210")

	coreDef := core.ChannelDefinition{
		Nonce:                 12345,
		Challenge:             3600,
		ApprovedSigValidators: "0x03",
	}

	result, err := coreDefToContractDef(coreDef, asset, userWallet, nodeAddress)
	require.NoError(t, err)

	// Verify the fields
	require.Equal(t, uint32(3600), result.ChallengeDuration)
	require.Equal(t, common.HexToAddress(userWallet), result.User)
	require.Equal(t, nodeAddress, result.Node)
	require.Equal(t, uint64(12345), result.Nonce)
	require.Equal(t, big.NewInt(3), result.ApprovedSignatureValidators)

	// Verify metadata contains asset hash
	assetHash := crypto.Keccak256Hash([]byte(asset))
	expectedMetadata := [32]byte{}
	copy(expectedMetadata[:8], assetHash[:8])
	require.Equal(t, expectedMetadata, result.Metadata)
}

func TestCoreDefToContractDef_EmptyValidators(t *testing.T) {
	t.Parallel()
	asset := "ETH"
	userWallet := "0x1234567890123456789012345678901234567890"
	nodeAddress := common.HexToAddress("0x9876543210987654321098765432109876543210")

	coreDef := core.ChannelDefinition{
		Nonce:     1,
		Challenge: 3600,
	}

	result, err := coreDefToContractDef(coreDef, asset, userWallet, nodeAddress)
	require.NoError(t, err)
	require.NotNil(t, result.ApprovedSignatureValidators)
	require.Equal(t, big.NewInt(0), result.ApprovedSignatureValidators)
}

// ========= coreLedgerToContractLedger Tests =========

func TestCoreLedgerToContractLedger_Success(t *testing.T) {
	t.Parallel()
	tokenAddr := "0x1234567890123456789012345678901234567890"
	coreLedger := core.Ledger{
		BlockchainID: 1,
		TokenAddress: tokenAddr,
		UserBalance:  decimal.NewFromFloat(100.5),
		UserNetFlow:  decimal.NewFromFloat(150.25),
		NodeBalance:  decimal.NewFromFloat(50.75),
		NodeNetFlow:  decimal.NewFromFloat(1.0),
	}

	decimals := uint8(18)
	result, err := coreLedgerToContractLedger(coreLedger, decimals)
	require.NoError(t, err)

	// Verify basic fields
	require.Equal(t, uint64(1), result.ChainId)
	require.Equal(t, common.HexToAddress(tokenAddr), result.Token)
	// Note: Decimals field is not set by coreLedgerToContractLedger

	// Verify amounts are converted correctly
	expectedUserAllocation := decimal.NewFromFloat(100.5).Shift(int32(decimals)).BigInt()
	expectedUserNetFlow := decimal.NewFromFloat(150.25).Shift(int32(decimals)).BigInt()
	expectedNodeAllocation := decimal.NewFromFloat(50.75).Shift(int32(decimals)).BigInt()
	expectedNodeNetFlow := decimal.NewFromFloat(1.0).Shift(int32(decimals)).BigInt()

	require.Equal(t, expectedUserAllocation, result.UserAllocation)
	require.Equal(t, expectedUserNetFlow, result.UserNetFlow)
	require.Equal(t, expectedNodeAllocation, result.NodeAllocation)
	require.Equal(t, expectedNodeNetFlow, result.NodeNetFlow)
}

// ========= contractLedgerToCoreLedger Tests =========

func TestContractLedgerToCoreLedger_Success(t *testing.T) {
	t.Parallel()
	tokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	decimals := uint8(18)

	userAllocation, _ := new(big.Int).SetString("100500000000000000000", 10) // 100.5 * 10^18
	userNetFlow, _ := new(big.Int).SetString("150250000000000000000", 10)    // 150.25 * 10^18
	nodeAllocation, _ := new(big.Int).SetString("50750000000000000000", 10)  // 50.75 * 10^18
	nodeNetFlow, _ := new(big.Int).SetString("1000000000000000000", 10)      // 1.0 * 10^18

	contractLedger := Ledger{
		ChainId:        1,
		Token:          tokenAddr,
		Decimals:       decimals,
		UserAllocation: userAllocation,
		UserNetFlow:    userNetFlow,
		NodeAllocation: nodeAllocation,
		NodeNetFlow:    nodeNetFlow,
	}

	result := contractLedgerToCoreLedger(contractLedger)

	// Verify basic fields
	require.Equal(t, uint64(1), result.BlockchainID)
	require.Equal(t, tokenAddr.Hex(), result.TokenAddress)

	// Verify amounts are converted back correctly
	require.Equal(t, "100.5", result.UserBalance.String())
	require.Equal(t, "150.25", result.UserNetFlow.String())
	require.Equal(t, "50.75", result.NodeBalance.String())
	require.Equal(t, "1", result.NodeNetFlow.String())
}

// ========= coreStateToContractState Tests =========

func TestCoreStateToContractState_BasicState(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	userSigHex := "0x1234567890abcdef"
	nodeSigHex := "0xfedcba0987654321"

	coreState := core.State{
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
			TokenAddress: "0x1234567890123456789012345678901234567890",
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(50),
			NodeNetFlow:  decimal.NewFromInt(50),
		},
		UserSig:    &userSigHex,
		NodeSig:    &nodeSigHex,
		Transition: core.Transition{Type: core.TransitionTypeTransferSend, Amount: decimal.NewFromInt(10)},
	}

	mockTokenGetter := func(blockchainID uint64, tokenAddress string) (uint8, error) {
		return 18, nil
	}

	result, err := coreStateToContractState(coreState, mockTokenGetter)
	require.NoError(t, err)

	// Verify basic fields
	require.Equal(t, uint64(1), result.Version)
	require.Equal(t, uint8(core.INTENT_OPERATE), result.Intent)

	// Verify signatures
	expectedUserSig, _ := hexutil.Decode(userSigHex)
	expectedNodeSig, _ := hexutil.Decode(nodeSigHex)
	require.Equal(t, expectedUserSig, result.UserSig)
	require.Equal(t, expectedNodeSig, result.NodeSig)

	// Verify home ledger
	require.Equal(t, uint64(1), result.HomeLedger.ChainId)
	// Note: Decimals field is not set by coreLedgerToContractLedger
}

func TestCoreStateToContractState_WithEscrowLedger(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	escrowChannelID := "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

	coreState := core.State{
		Version:         1,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
			TokenAddress: "0x1234567890123456789012345678901234567890",
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(50),
			NodeNetFlow:  decimal.NewFromInt(50),
		},
		EscrowLedger: &core.Ledger{
			BlockchainID: 2,
			TokenAddress: "0xfedcba0987654321fedcba0987654321fedcba09",
			UserBalance:  decimal.NewFromInt(200),
			UserNetFlow:  decimal.NewFromInt(200),
			NodeBalance:  decimal.NewFromInt(100),
			NodeNetFlow:  decimal.NewFromInt(100),
		},
		Transition: core.Transition{Type: core.TransitionTypeMutualLock, Amount: decimal.NewFromInt(10)},
	}

	mockTokenGetter := func(blockchainID uint64, tokenAddress string) (uint8, error) {
		if blockchainID == 1 {
			return 18, nil
		}
		return 6, nil
	}

	result, err := coreStateToContractState(coreState, mockTokenGetter)
	require.NoError(t, err)

	// Verify intent
	require.Equal(t, uint8(core.INTENT_INITIATE_ESCROW_DEPOSIT), result.Intent)

	// Verify escrow ledger is populated
	require.Equal(t, uint64(2), result.NonHomeLedger.ChainId)
	// Note: Decimals field is not set by coreLedgerToContractLedger
}

func TestCoreStateToContractState_WithoutEscrowLedger(t *testing.T) {
	// This test ensures that when there's no escrow ledger,
	// the NonHomeState is initialized with proper empty values
	// to prevent ABI encoding panics
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	coreState := core.State{
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
			TokenAddress: "0x1234567890123456789012345678901234567890",
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(50),
			NodeNetFlow:  decimal.NewFromInt(50),
		},
		EscrowLedger: nil, // Explicitly nil
		Transition:   core.Transition{Type: core.TransitionTypeHomeDeposit, Amount: decimal.NewFromInt(100)},
	}

	mockTokenGetter := func(blockchainID uint64, tokenAddress string) (uint8, error) {
		return 18, nil
	}

	result, err := coreStateToContractState(coreState, mockTokenGetter)
	require.NoError(t, err)

	// Verify that NonHomeLedger is initialized with non-nil big.Int pointers
	require.NotNil(t, result.NonHomeLedger.UserAllocation, "UserAllocation should not be nil")
	require.NotNil(t, result.NonHomeLedger.UserNetFlow, "UserNetFlow should not be nil")
	require.NotNil(t, result.NonHomeLedger.NodeAllocation, "NodeAllocation should not be nil")
	require.NotNil(t, result.NonHomeLedger.NodeNetFlow, "NodeNetFlow should not be nil")

	// Verify they are zero values
	require.Equal(t, big.NewInt(0), result.NonHomeLedger.UserAllocation)
	require.Equal(t, big.NewInt(0), result.NonHomeLedger.UserNetFlow)
	require.Equal(t, big.NewInt(0), result.NonHomeLedger.NodeAllocation)
	require.Equal(t, big.NewInt(0), result.NonHomeLedger.NodeNetFlow)

	// Verify the intent
	require.Equal(t, uint8(core.INTENT_DEPOSIT), result.Intent)
}

func TestCoreStateToContractState_InvalidUserSig(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	invalidSigHex := "zzzzz"

	coreState := core.State{
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
			TokenAddress: "0x1234567890123456789012345678901234567890",
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(50),
			NodeNetFlow:  decimal.NewFromInt(50),
		},
		UserSig:    &invalidSigHex,
		Transition: core.Transition{Type: core.TransitionTypeTransferSend},
	}

	mockTokenGetter := func(blockchainID uint64, tokenAddress string) (uint8, error) {
		return 18, nil
	}

	_, err := coreStateToContractState(coreState, mockTokenGetter)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode user signature")
}

// ========= contractStateToCoreState Tests =========

func TestContractStateToCoreState_HomeOnly(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	userAllocation, _ := new(big.Int).SetString("100000000000000000000", 10)
	nodeAllocation, _ := new(big.Int).SetString("50000000000000000000", 10)

	contractState := State{
		Version: 1,
		Intent:  core.INTENT_OPERATE,
		HomeLedger: Ledger{
			ChainId:        1,
			Token:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
			Decimals:       18,
			UserAllocation: userAllocation,
			UserNetFlow:    userAllocation,
			NodeAllocation: nodeAllocation,
			NodeNetFlow:    nodeAllocation,
		},
		UserSig: []byte{0x12, 0x34},
		NodeSig: []byte{0x56, 0x78},
	}

	result, err := contractStateToCoreState(contractState, homeChannelID, nil)
	require.NoError(t, err)

	// Verify basic fields
	require.Equal(t, uint64(1), result.Version)
	require.NotNil(t, result.HomeChannelID)
	require.Equal(t, homeChannelID, *result.HomeChannelID)
	require.Nil(t, result.EscrowChannelID)
	require.Nil(t, result.EscrowLedger)

	// Verify home ledger
	require.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
	require.Equal(t, "100", result.HomeLedger.UserBalance.String())
	require.Equal(t, "50", result.HomeLedger.NodeBalance.String())

	// Verify signatures
	require.NotNil(t, result.UserSig)
	require.Equal(t, "0x1234", *result.UserSig)
	require.NotNil(t, result.NodeSig)
	require.Equal(t, "0x5678", *result.NodeSig)
}

func TestContractStateToCoreState_WithEscrow(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	escrowChannelID := "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

	homeUserAllocation, _ := new(big.Int).SetString("100000000000000000000", 10)
	homeNodeAllocation, _ := new(big.Int).SetString("50000000000000000000", 10)

	contractState := State{
		Version: 2,
		Intent:  core.INTENT_INITIATE_ESCROW_DEPOSIT,
		HomeLedger: Ledger{
			ChainId:        1,
			Token:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
			Decimals:       18,
			UserAllocation: homeUserAllocation,
			UserNetFlow:    homeUserAllocation,
			NodeAllocation: homeNodeAllocation,
			NodeNetFlow:    homeNodeAllocation,
		},
		NonHomeLedger: Ledger{
			ChainId:        2,
			Token:          common.HexToAddress("0xfedcba0987654321fedcba0987654321fedcba09"),
			Decimals:       6,
			UserAllocation: big.NewInt(200000000),
			UserNetFlow:    big.NewInt(200000000),
			NodeAllocation: big.NewInt(100000000),
			NodeNetFlow:    big.NewInt(100000000),
		},
	}

	result, err := contractStateToCoreState(contractState, homeChannelID, &escrowChannelID)
	require.NoError(t, err)

	// Verify escrow fields
	require.NotNil(t, result.EscrowChannelID)
	require.Equal(t, escrowChannelID, *result.EscrowChannelID)
	require.NotNil(t, result.EscrowLedger)

	// Verify escrow ledger
	require.Equal(t, uint64(2), result.EscrowLedger.BlockchainID)
	require.Equal(t, "200", result.EscrowLedger.UserBalance.String())
	require.Equal(t, "100", result.EscrowLedger.NodeBalance.String())
}

func TestContractStateToCoreState_EmptyChannelID(t *testing.T) {
	t.Parallel()
	userAllocation, _ := new(big.Int).SetString("100000000000000000000", 10)
	nodeAllocation, _ := new(big.Int).SetString("50000000000000000000", 10)

	contractState := State{
		Version: 1,
		Intent:  core.INTENT_OPERATE,
		HomeLedger: Ledger{
			ChainId:        1,
			Token:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
			Decimals:       18,
			UserAllocation: userAllocation,
			UserNetFlow:    userAllocation,
			NodeAllocation: nodeAllocation,
			NodeNetFlow:    nodeAllocation,
		},
	}

	result, err := contractStateToCoreState(contractState, "", nil)
	require.NoError(t, err)
	require.Nil(t, result.HomeChannelID)
}

func TestContractStateToCoreState_EmptySignatures(t *testing.T) {
	t.Parallel()
	homeChannelID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	userAllocation, _ := new(big.Int).SetString("100000000000000000000", 10)
	nodeAllocation, _ := new(big.Int).SetString("50000000000000000000", 10)

	contractState := State{
		Version: 1,
		Intent:  core.INTENT_OPERATE,
		HomeLedger: Ledger{
			ChainId:        1,
			Token:          common.HexToAddress("0x1234567890123456789012345678901234567890"),
			Decimals:       18,
			UserAllocation: userAllocation,
			UserNetFlow:    userAllocation,
			NodeAllocation: nodeAllocation,
			NodeNetFlow:    nodeAllocation,
		},
		UserSig: []byte{},
		NodeSig: nil,
	}

	result, err := contractStateToCoreState(contractState, homeChannelID, nil)
	require.NoError(t, err)
	require.Nil(t, result.UserSig)
	require.Nil(t, result.NodeSig)
}

// ========= Round-trip Tests =========

func TestLedgerConversion_RoundTrip(t *testing.T) {
	t.Parallel()
	original := core.Ledger{
		BlockchainID: 1,
		TokenAddress: "0x1234567890123456789012345678901234567890",
		UserBalance:  decimal.NewFromFloat(123.456),
		UserNetFlow:  decimal.NewFromFloat(200.789),
		NodeBalance:  decimal.NewFromFloat(76.544),
		NodeNetFlow:  decimal.NewFromFloat(-0.789),
	}

	decimals := uint8(18)

	// Convert to contract ledger
	contractLedger, err := coreLedgerToContractLedger(original, decimals)
	require.NoError(t, err)

	// Add decimals field (since it's not part of the conversion input)
	contractLedger.Decimals = decimals

	// Convert back to core ledger
	result := contractLedgerToCoreLedger(contractLedger)

	// Verify round-trip preserves values
	require.Equal(t, original.BlockchainID, result.BlockchainID)
	require.Equal(t, original.TokenAddress, result.TokenAddress)
	require.True(t, original.UserBalance.Equal(result.UserBalance))
	require.True(t, original.UserNetFlow.Equal(result.UserNetFlow))
	require.True(t, original.NodeBalance.Equal(result.NodeBalance))
	require.True(t, original.NodeNetFlow.Equal(result.NodeNetFlow))
}
