package sdk

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/layer-3/nitrolite/pkg/blockchain/evm"
	"github.com/layer-3/nitrolite/pkg/core"
)

// WatchValidatorRegistered subscribes to ValidatorRegistered events on the ChannelHub
// contract for the given chain and delivers them on the returned channel.
//
// Each event carries the newly registered validator ID, its contract address, and the
// block number at which it was emitted. App builders should alert users and prompt them
// to revoke ERC20 approvals granted to ChannelHub whenever an unexpected validator
// appears — see contracts/SECURITY.md.
//
// Gap-free monitoring: pass fromBlock = 0 on the first call. On reconnect, pass
// lastEvent.BlockNumber + 1 so any events emitted during the outage are replayed
// before live events resume. This ensures the 1-day VALIDATOR_ACTIVATION_DELAY
// window is not silently shortened by network interruptions.
//
// The channel is closed when ctx is cancelled or the subscription is lost. On a
// lost subscription (network drop), call WatchValidatorRegistered again with the
// last received BlockNumber + 1 to resubscribe without gaps.
//
// The RPC URL configured via WithBlockchainRPC for chainID must be a WebSocket
// endpoint (wss:// or ws://). HTTP endpoints do not support event subscriptions
// and will return an error.
func (c *Client) WatchValidatorRegistered(ctx context.Context, chainID uint64, fromBlock uint64) (<-chan *core.ValidatorRegisteredEvent, error) {
	rpcURL, exists := c.config.BlockchainRPCs[chainID]
	if !exists {
		return nil, fmt.Errorf("blockchain RPC not configured for chain %d (use WithBlockchainRPC)", chainID)
	}

	channelHubAddress, err := c.getChannelHubAddress(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ChannelHub address for chain %d: %w", chainID, err)
	}

	ethCl, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain RPC for chain %d: %w", chainID, err)
	}

	rawCh, err := evm.WatchValidatorRegistered(ctx, common.HexToAddress(channelHubAddress), ethCl, chainID, fromBlock)
	if err != nil {
		ethCl.Close()
		return nil, err
	}

	// Proxy events to the caller and close the ethClient once the subscription ends.
	outCh := make(chan *core.ValidatorRegisteredEvent, 16)
	go func() {
		defer ethCl.Close()
		defer close(outCh)
		for ev := range rawCh {
			select {
			case outCh <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, nil
}
