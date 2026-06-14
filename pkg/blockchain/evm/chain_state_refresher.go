package evm

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/layer-3/nitrolite/pkg/core"
)

// On-chain ChannelStatus enum values (contracts/src/interfaces/Types.sol).
const (
	onchainChannelStatusVoid        uint8 = 0
	onchainChannelStatusOperating   uint8 = 1
	onchainChannelStatusDisputed    uint8 = 2
	onchainChannelStatusClosed      uint8 = 3
	onchainChannelStatusMigratingIn uint8 = 4
	onchainChannelStatusMigratedOut uint8 = 5
)

// ChannelChainResolver returns the host chain ID for a given channelID.
// The EVM refresher uses it to dispatch the on-chain read to the right
// ChannelHubCaller when more than one chain hosts ChannelHub deployments.
type ChannelChainResolver func(channelID string) (uint64, error)

// EVMChainStateRefresher implements core.ChainStateRefresher by dispatching
// getChannelData reads to a per-chain ChannelHubCaller. The caller map is
// populated at startup with one entry per blockchain that has a ChannelHub
// contract deployed; see nitronode/main.go for wire-up.
//
// The refresher is read-only and stateless beyond the caller map; it is safe
// for concurrent use from multiple reactor handler goroutines.
type EVMChainStateRefresher struct {
	callers     map[uint64]*ChannelHubCaller
	resolveHome ChannelChainResolver
}

// NewEVMChainStateRefresher constructs a refresher backed by the supplied
// per-chain ChannelHubCaller map and chain resolver. The map is taken by
// reference; callers must not mutate it after construction. resolveHome must
// not be nil — the refresher relies on it to pick the right caller for the
// channel under refresh.
func NewEVMChainStateRefresher(callers map[uint64]*ChannelHubCaller, resolveHome ChannelChainResolver) *EVMChainStateRefresher {
	return &EVMChainStateRefresher{
		callers:     callers,
		resolveHome: resolveHome,
	}
}

// RefreshChannelFromChain reads the authoritative on-chain snapshot for
// channelID from the ChannelHub on the channel's host chain and returns it
// for the caller to overwrite the local row. See core.ChainStateRefresher
// for semantics.
func (r *EVMChainStateRefresher) RefreshChannelFromChain(ctx context.Context, channelID string) (*core.RefreshedChannel, error) {
	if r.resolveHome == nil {
		return nil, fmt.Errorf("no channel chain resolver configured")
	}

	chainID, err := r.resolveHome(channelID)
	if err != nil {
		return nil, fmt.Errorf("resolve home chain for channel %s: %w", channelID, err)
	}

	caller, ok := r.callers[chainID]
	if !ok || caller == nil {
		return nil, fmt.Errorf("no ChannelHub caller registered for chain %d (channel %s)", chainID, channelID)
	}

	channelIDBytes, err := hexToBytes32(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID %q: %w", channelID, err)
	}

	data, err := caller.GetChannelData(&bind.CallOpts{Context: ctx}, channelIDBytes)
	if err != nil {
		return nil, fmt.Errorf("getChannelData(%s) on chain %d: %w", channelID, chainID, err)
	}

	status, err := mapOnchainChannelStatus(data.Status)
	if err != nil {
		return nil, fmt.Errorf("map on-chain status for channel %s: %w", channelID, err)
	}

	var expiry *time.Time
	if data.ChallengeExpiry != nil && data.ChallengeExpiry.Sign() > 0 {
		t := time.Unix(data.ChallengeExpiry.Int64(), 0)
		expiry = &t
	}

	return &core.RefreshedChannel{
		Status:             status,
		StateVersion:       data.LastState.Version,
		ChallengeExpiresAt: expiry,
		LastStateUserSig:   encodeSig(data.LastState.UserSig),
	}, nil
}

// mapOnchainChannelStatus translates the contract's ChannelStatus enum to the
// off-chain core.ChannelStatus. MIGRATING_IN is treated as Open because the
// channel is live from this hub's perspective. MIGRATED_OUT is treated as
// Closed because no further transitions can land on this hub.
func mapOnchainChannelStatus(s uint8) (core.ChannelStatus, error) {
	switch s {
	case onchainChannelStatusVoid:
		return core.ChannelStatusVoid, nil
	case onchainChannelStatusOperating, onchainChannelStatusMigratingIn:
		return core.ChannelStatusOpen, nil
	case onchainChannelStatusDisputed:
		return core.ChannelStatusChallenged, nil
	case onchainChannelStatusClosed, onchainChannelStatusMigratedOut:
		return core.ChannelStatusClosed, nil
	default:
		return 0, fmt.Errorf("unknown on-chain ChannelStatus %d", s)
	}
}
