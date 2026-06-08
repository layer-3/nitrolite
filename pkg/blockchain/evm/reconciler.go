package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/layer-3/nitrolite/pkg/log"
)

// findCommonAncestor determines the last block in the canonical chain that the
// node has already processed. It walks stored block hashes backward until it
// finds one that eth_getBlockByHash confirms is canonical, then returns that
// block number as the safe replay start point.
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

		hash := common.HexToHash(blockHash)
		header, err := client.HeaderByHash(ctx, hash)
		if err != nil {
			return 0, fmt.Errorf("check canonicality of block %d (%s): %w", blockNum, blockHash, err)
		}

		if header != nil {
			// This block is still in the canonical chain.
			if blockNum != header.Number.Uint64() {
				// Sanity check: the block at this hash should have the number we stored.
				return 0, fmt.Errorf("block hash %s has unexpected number: stored %d, chain %d", blockHash, blockNum, header.Number.Uint64())
			}
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
