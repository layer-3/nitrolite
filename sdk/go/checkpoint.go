package sdk

import (
	"context"
	"fmt"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// DefaultCheckpointPollInterval is the default interval between balance polls in WaitForCheckpoint.
const DefaultCheckpointPollInterval = 3 * time.Second

// DefaultCheckpointTimeout is the default overall poll timeout in WaitForCheckpoint
// (applied after the confirmation-delay lower-bound wait).
const DefaultCheckpointTimeout = 2 * time.Minute

// WaitForCheckpointOptions configures WaitForCheckpoint.
type WaitForCheckpointOptions struct {
	// ChainID, when non-nil, makes WaitForCheckpoint sleep for that chain's
	// confirmation_delay_secs before the first poll (the credit cannot arrive
	// before the gate elapses).
	ChainID *uint64

	// ExpectedBalance, when non-nil, resolves the wait once the polled balance for
	// the asset is >= this value. When nil, the wait resolves on the first balance
	// change relative to the value observed at call time.
	ExpectedBalance *decimal.Decimal

	// PollInterval between balance polls. Zero uses DefaultCheckpointPollInterval.
	PollInterval time.Duration

	// Timeout for polling after the lower-bound wait. Zero uses DefaultCheckpointTimeout.
	Timeout time.Duration
}

// WaitForCheckpoint waits until the off-chain credit for asset lands after an on-chain
// checkpoint transaction. Because the node applies a per-chain confirmation gate
// (confirmation_delay_secs) before crediting an event, the off-chain balance does not
// update the instant the tx receipt is mined — it updates up to confirmation_delay_secs later.
//
// When opts.ChainID is set, the method first sleeps for that chain's confirmation delay
// (the lower bound), then polls GetBalances every PollInterval until the target condition
// is met or Timeout elapses. The provided ctx cancels the whole operation, including the
// lower-bound sleep.
//
// Target condition:
//   - opts.ExpectedBalance set → balance for asset is >= ExpectedBalance.
//   - otherwise               → balance for asset differs from the value at call time.
//
// txHash is informational and is included in the timeout error.
func (c *Client) WaitForCheckpoint(ctx context.Context, asset, txHash string, opts *WaitForCheckpointOptions) (core.BalanceEntry, error) {
	if opts == nil {
		opts = &WaitForCheckpointOptions{}
	}
	pollInterval := opts.PollInterval
	if pollInterval <= 0 {
		pollInterval = DefaultCheckpointPollInterval
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultCheckpointTimeout
	}

	wallet := c.GetUserAddress()

	balanceFor := func(entries []core.BalanceEntry) (core.BalanceEntry, bool) {
		for _, e := range entries {
			if e.Asset == asset {
				return e, true
			}
		}
		return core.BalanceEntry{Asset: asset, Balance: decimal.Zero, Enforced: decimal.Zero}, false
	}

	// Snapshot starting ENFORCED balance for "changed" mode. The gate credits on-chain
	// events to `enforced` (RefreshUserEnforcedBalance), not spendable `balance`.
	startEntries, err := c.GetBalances(ctx, wallet)
	if err != nil {
		return core.BalanceEntry{}, fmt.Errorf("failed to read starting balance: %w", err)
	}
	startEntry, _ := balanceFor(startEntries)
	startEnforced := startEntry.Enforced

	// Lower-bound wait: the credit cannot land before the gate elapses.
	if opts.ChainID != nil {
		delaySecs, err := c.GetConfirmationDelay(ctx, *opts.ChainID)
		if err != nil {
			return core.BalanceEntry{}, err
		}
		if delaySecs > 0 {
			select {
			case <-ctx.Done():
				return core.BalanceEntry{}, ctx.Err()
			case <-time.After(time.Duration(delaySecs) * time.Second):
			}
		}
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		entries, err := c.GetBalances(ctx, wallet)
		if err != nil {
			return core.BalanceEntry{}, fmt.Errorf("failed to poll balance: %w", err)
		}
		entry, found := balanceFor(entries)

		satisfied := false
		if opts.ExpectedBalance != nil {
			satisfied = found && entry.Enforced.GreaterThanOrEqual(*opts.ExpectedBalance)
		} else {
			satisfied = found && !entry.Enforced.Equal(startEnforced)
		}
		if satisfied {
			return entry, nil
		}

		if time.Now().After(deadline) {
			return core.BalanceEntry{}, fmt.Errorf("waitForCheckpoint timed out after %s waiting for %s credit (tx %s)", timeout, asset, txHash)
		}

		select {
		case <-ctx.Done():
			return core.BalanceEntry{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
