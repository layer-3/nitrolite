package evm

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testContract    = "0x1234567890abcdef1234567890abcdef12345678"
	testBlockchainID = uint64(1)
)

func newTestLogger() log.Logger {
	return log.NewNoopLogger()
}

// TestFindCommonAncestor_NoStoredEvents verifies that when no contract events exist,
// findCommonAncestor returns 0 (genesis fallback).
func TestFindCommonAncestor_NoStoredEvents(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(0), "", nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(0), result)
	client.AssertNotCalled(t, "HeaderByHash")
}

// TestFindCommonAncestor_LatestBlockCanonical verifies that when the latest stored block
// is still canonical, findCommonAncestor returns that block number with no backward walk.
func TestFindCommonAncestor_LatestBlockCanonical(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	blockHash := common.HexToHash("0xaabbccdd")
	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(500), blockHash.Hex(), nil)

	canonicalHeader := &types.Header{Number: big.NewInt(500)}
	client.On("HeaderByHash", mock.Anything, blockHash).Return(canonicalHeader, nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(500), result)
	getter.AssertNotCalled(t, "GetPreviousDistinctBlockHash")
}

// TestFindCommonAncestor_SingleReorgDepth verifies that when the latest block is reorged out
// but the previous one is canonical, findCommonAncestor returns the previous block number.
func TestFindCommonAncestor_SingleReorgDepth(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	reorgedHash := common.HexToHash("0xreorged0")
	canonicalHash := common.HexToHash("0xcanon000")

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(200), reorgedHash.Hex(), nil)

	// Latest block (200) reorged out.
	client.On("HeaderByHash", mock.Anything, reorgedHash).Return(nil, nil)

	// Walk to previous block (190) which is canonical.
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(200)).
		Return(uint64(190), canonicalHash.Hex(), nil)

	canonicalHeader := &types.Header{Number: big.NewInt(190)}
	client.On("HeaderByHash", mock.Anything, canonicalHash).Return(canonicalHeader, nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(190), result)
}

// TestFindCommonAncestor_WalkToGenesis verifies that when all stored blocks are reorged out,
// findCommonAncestor returns 0 (genesis fallback).
func TestFindCommonAncestor_WalkToGenesis(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	hash300 := common.HexToHash("0x0000300")
	hash200 := common.HexToHash("0x0000200")

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(300), hash300.Hex(), nil)

	// Block 300 reorged out.
	client.On("HeaderByHash", mock.Anything, hash300).Return(nil, nil)
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(300)).
		Return(uint64(200), hash200.Hex(), nil)

	// Block 200 reorged out.
	client.On("HeaderByHash", mock.Anything, common.HexToHash(hash200.Hex())).Return(nil, nil)
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(200)).
		Return(uint64(0), "", nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(0), result)
}

// TestFindCommonAncestor_PreMigrationLatestRow verifies that when the latest stored row has
// an empty block_hash (pre-migration row), findCommonAncestor returns that block number
// without making any RPC call, treating the row as canonical.
func TestFindCommonAncestor_PreMigrationLatestRow(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	// blockNum=450 but blockHash="" — pre-migration row.
	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(450), "", nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(450), result)
	client.AssertNotCalled(t, "HeaderByHash")
}

// TestFindCommonAncestor_PreMigrationMidWalk verifies that when a pre-migration row (empty
// block_hash) is encountered during the backward walk, the walk stops and returns that
// block number rather than making an RPC call with a zero hash.
func TestFindCommonAncestor_PreMigrationMidWalk(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	reorgedHash := common.HexToHash("0xreorgedX")

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(300), reorgedHash.Hex(), nil)

	// Block 300 reorged out.
	client.On("HeaderByHash", mock.Anything, reorgedHash).Return(nil, nil)

	// Walk backward hits a pre-migration row with empty hash at block 250.
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(300)).
		Return(uint64(250), "", nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(250), result)
	// HeaderByHash must NOT be called for the zero-hash pre-migration row.
	client.AssertNumberOfCalls(t, "HeaderByHash", 1)
}

// TestFindCommonAncestor_HeaderByHashError verifies that RPC errors are propagated.
func TestFindCommonAncestor_HeaderByHashError(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	blockHash := common.HexToHash("0xfailhash")
	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(100), blockHash.Hex(), nil)

	client.On("HeaderByHash", mock.Anything, blockHash).Return(nil, errors.New("rpc timeout"))

	_, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc timeout")
}
