package evm

import (
	"context"
	"errors"
	"math/big"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testContract     = "0x1234567890abcdef1234567890abcdef12345678"
	testBlockchainID = uint64(1)
)

func newTestLogger() log.Logger {
	return log.NewNoopLogger()
}

// makeHeader builds a Header with a deterministic (and unique-per-seed) hash for
// the given block number. Two calls with different seeds produce headers whose
// Hash() values differ, which lets canonicality tests distinguish "this stored
// block is canonical" (same seed) from "this stored block was reorged out"
// (different seed at the same number).
func makeHeader(blockNum int64, seed int64) *types.Header {
	return &types.Header{
		Number:     big.NewInt(blockNum),
		Difficulty: big.NewInt(seed),
	}
}

func bigEqual(want *big.Int) interface{} {
	return mock.MatchedBy(func(got *big.Int) bool { return got != nil && got.Cmp(want) == 0 })
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
	client.AssertNotCalled(t, "HeaderByNumber")
	client.AssertNotCalled(t, "HeaderByHash")
}

// TestFindCommonAncestor_LatestBlockCanonical verifies that when the latest stored block
// is still canonical, findCommonAncestor returns that block number with no backward walk.
func TestFindCommonAncestor_LatestBlockCanonical(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	header := makeHeader(500, 1)
	storedHash := header.Hash()

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(500), storedHash.Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(500))).Return(header, nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(500), result)
	getter.AssertNotCalled(t, "GetPreviousDistinctBlockHash")
}

// TestFindCommonAncestor_SingleReorgDepth verifies that when the latest stored block has
// been reorged out (canonical chain has a different block at that height), findCommonAncestor
// walks back one step and returns the previous canonical block.
func TestFindCommonAncestor_SingleReorgDepth(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	// Block 200 was reorged: stored hash came from a now-orphan block; canonical chain
	// has a different block at the same height.
	storedAt200 := makeHeader(200, 1)
	canonicalAt200 := makeHeader(200, 2)
	require.NotEqual(t, storedAt200.Hash(), canonicalAt200.Hash())

	// Block 190 is canonical.
	headerAt190 := makeHeader(190, 1)
	storedAt190 := headerAt190.Hash()

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(200), storedAt200.Hash().Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(200))).Return(canonicalAt200, nil)
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(200)).
		Return(uint64(190), storedAt190.Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(190))).Return(headerAt190, nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(190), result)
}

// TestFindCommonAncestor_NotFoundTreatedAsReorg verifies that when HeaderByNumber returns
// ethereum.NotFound (e.g. the RPC backend has pruned that height, or no canonical block
// exists at that number yet), the walk continues backward instead of crashing the listener.
// This is the regression the colleague flagged: the old HeaderByHash path treated NotFound
// as a fatal startup error.
func TestFindCommonAncestor_NotFoundTreatedAsReorg(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	storedAt200 := common.HexToHash("0xreorged200")
	headerAt190 := makeHeader(190, 1)

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(200), storedAt200.Hex(), nil)
	// HeaderByNumber(200) returns NotFound — must NOT be treated as fatal.
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(200))).Return(nil, ethereum.NotFound)

	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(200)).
		Return(uint64(190), headerAt190.Hash().Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(190))).Return(headerAt190, nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(190), result)
}

// TestFindCommonAncestor_WalkToGenesis verifies that when all stored blocks have been
// reorged out (canonical hashes differ at every stored height), findCommonAncestor returns
// 0 (genesis fallback).
func TestFindCommonAncestor_WalkToGenesis(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	storedAt300 := makeHeader(300, 1).Hash()
	storedAt200 := makeHeader(200, 1).Hash()
	canonicalAt300 := makeHeader(300, 2)
	canonicalAt200 := makeHeader(200, 2)

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(300), storedAt300.Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(300))).Return(canonicalAt300, nil)
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(300)).
		Return(uint64(200), storedAt200.Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(200))).Return(canonicalAt200, nil)
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
	client.AssertNotCalled(t, "HeaderByNumber")
}

// TestFindCommonAncestor_PreMigrationMidWalk verifies that when a pre-migration row (empty
// block_hash) is encountered during the backward walk, the walk stops and returns that
// block number rather than making an RPC call with a zero hash.
func TestFindCommonAncestor_PreMigrationMidWalk(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	storedAt300 := makeHeader(300, 1).Hash()
	canonicalAt300 := makeHeader(300, 2)

	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(300), storedAt300.Hex(), nil)
	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(300))).Return(canonicalAt300, nil)

	// Walk backward hits a pre-migration row with empty hash at block 250.
	getter.On("GetPreviousDistinctBlockHash", testContract, testBlockchainID, uint64(300)).
		Return(uint64(250), "", nil)

	result, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.NoError(t, err)
	assert.Equal(t, uint64(250), result)
	// HeaderByNumber must NOT be called for the zero-hash pre-migration row.
	client.AssertNumberOfCalls(t, "HeaderByNumber", 1)
}

// TestFindCommonAncestor_RPCError verifies that non-NotFound RPC errors are propagated.
func TestFindCommonAncestor_RPCError(t *testing.T) {
	t.Parallel()

	client := new(MockEVMClient)
	getter := new(MockContractEventGetter)

	blockHash := common.HexToHash("0xfailhash")
	getter.On("GetLatestContractEventBlockHashAndNumber", testContract, testBlockchainID).
		Return(uint64(100), blockHash.Hex(), nil)

	client.On("HeaderByNumber", mock.Anything, bigEqual(big.NewInt(100))).
		Return(nil, errors.New("rpc timeout"))

	_, err := findCommonAncestor(context.Background(), client, getter, testContract, testBlockchainID, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc timeout")
}
