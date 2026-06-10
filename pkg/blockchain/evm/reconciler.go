package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/layer-3/nitrolite/pkg/log"
)

// findCommonAncestor determines the last block in the canonical chain that the
// node has already processed. It walks stored block hashes backward until it
// finds a stored hash that matches the canonical chain's hash at that height,
// then returns that block number as the safe replay start point.
//
// Returns 0 when no stored events exist or when every stored block has been
// reorged out — in both cases the caller should replay from genesis/start-block.
func findCommonAncestor(
	ctx context.Context,
	client EVMClient,
	getter ContractEventGetter,
	contractAddress string,
	blockchainID uint64,
	logger log.Logger,
) (uint64, error) {
	blockNum, blockHash, err := getter.GetLatestContractEventBlockHashAndNumber(contractAddress, blockchainID)
	if err != nil {
		return 0, fmt.Errorf("get latest contract event block hash: %w", err)
	}
	if blockHash == "" {
		// No stored events (blockNum=0) or pre-migration row with no hash (blockNum>0).
		// Either way, treat blockNum as the safe canonical resume point.
		return blockNum, nil
	}

	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		canonical, err := isStoredBlockCanonical(ctx, client, blockNum, common.HexToHash(blockHash))
		if err != nil {
			return 0, fmt.Errorf("check canonicality of block %d (%s): %w", blockNum, blockHash, err)
		}

		if canonical {
			logger.Info("reconciliation: found common ancestor",
				"blockchainID", blockchainID,
				"blockNumber", blockNum,
				"blockHash", blockHash,
			)
			return blockNum, nil
		}

		// Block was reorged out — walk to the next-older stored block.
		logger.Info("reconciliation: block reorged, walking backward",
			"blockchainID", blockchainID,
			"blockNumber", blockNum,
			"blockHash", blockHash,
		)
		prevNum, prevHash, err := getter.GetPreviousDistinctBlockHash(contractAddress, blockchainID, blockNum)
		if err != nil {
			return 0, fmt.Errorf("get previous distinct block hash below %d: %w", blockNum, err)
		}
		if prevHash == "" {
			// No older stored block (prevNum=0) or pre-migration row (prevNum>0).
			// Use prevNum as the safe canonical resume point.
			logger.Info("reconciliation: reached pre-migration or genesis boundary",
				"blockchainID", blockchainID,
				"blockNumber", prevNum,
			)
			return prevNum, nil
		}

		blockNum = prevNum
		blockHash = prevHash
	}
}

// isStoredBlockCanonical reports whether the block currently occupying blockNum
// in the canonical chain has the given storedHash. It uses HeaderByNumber rather
// than HeaderByHash because the two answer different questions:
//
//   - HeaderByHash returns any header the node has indexed, including orphan
//     side-chain headers still cached locally. A successful return does NOT prove
//     the block is in the canonical chain. A reorged-out hash may also come back
//     as ethereum.NotFound depending on the backend's pruning policy —
//     conflating those two outcomes with a single boolean is unsafe.
//
//   - HeaderByNumber returns the block currently occupying that height in the
//     canonical chain. Comparing its hash to the stored hash is definitive: equal
//     means the stored block is canonical, different means it has been reorged
//     out.
//
// ethereum.NotFound from HeaderByNumber (e.g. the chain has pruned the height or
// has not yet produced a block at that height) is treated as "not canonical"
// rather than a fatal error, so the caller walks backward instead of crashing
// the listener on startup.
func isStoredBlockCanonical(ctx context.Context, client EVMClient, blockNum uint64, storedHash common.Hash) (bool, error) {
	header, err := client.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return false, nil
		}
		return false, err
	}
	if header == nil {
		return false, nil
	}
	return header.Hash() == storedHash, nil
}
