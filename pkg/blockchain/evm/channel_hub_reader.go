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

// EVMChannelHubReader implements core.ReadOnlyChannelHub by reading from a
// single chain's bound ChannelHub contract. Each ChannelHubReactor binds its
// own reader for the chain it listens on, so no chain-resolution dispatcher
// is required.
//
// The reader is read-only and stateless beyond the caller; it is safe for
// concurrent use from multiple reactor handler goroutines.
type EVMChannelHubReader struct {
	caller *ChannelHubCaller
}

// NewChannelHubReader constructs a reader backed by the supplied bound
// ChannelHub caller. The caller must be non-nil; passing nil panics on the
// first FetchChannel invocation rather than at construction.
func NewChannelHubReader(caller *ChannelHubCaller) *EVMChannelHubReader {
	return &EVMChannelHubReader{caller: caller}
}

// FetchChannel reads the authoritative on-chain snapshot for channelID from
// the ChannelHub contract bound to this reader and returns it for the caller
// to overwrite the local row. See core.ReadOnlyChannelHub for semantics.
func (r *EVMChannelHubReader) FetchChannel(ctx context.Context, channelID string) (*core.OnChainChannelSnapshot, error) {
	channelIDBytes, err := hexToBytes32(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID %q: %w", channelID, err)
	}

	ctx, cancel := context.WithTimeout(ctx, RpcCallTimeout)
	defer cancel()

	data, err := r.caller.GetChannelData(&bind.CallOpts{Context: ctx}, channelIDBytes)
	if err != nil {
		return nil, fmt.Errorf("getChannelData(%s): %w", channelID, err)
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

	return &core.OnChainChannelSnapshot{
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
