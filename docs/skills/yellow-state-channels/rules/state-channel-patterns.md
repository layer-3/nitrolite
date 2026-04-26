# State Channel Protocol Patterns — Rules & Corrections

## Channel Opening

Always prefer the implicit join (single transaction) approach. Collect all participant signatures before calling `create()`. Avoid the two-step create→join flow unless you have no choice.

```
// CORRECT: supply both sigs, transition directly to ACTIVE
create(channel, stateWithSigs[0+1]) → ACTIVE in one tx

// SLOWER: two transactions, intermediate INITIAL state
create(channel, stateWithSig[0]) → INITIAL
join(channelId, sig[1]) → ACTIVE
```

## State Version Monotonicity

State versions must be strictly monotonically increasing. Never skip or repeat.

```
// CORRECT
version 1 → 2 → 3 → 4 → 5

// WRONG — version skipped
version 1 → 3  // Contract rejects (not prev + 1)

// WRONG — version repeated
version 5 → 5  // Contract rejects
```

## Intent System

Each state must carry the correct `StateIntent`:

| State | Intent | When |
|-------|--------|------|
| Channel creation | `INITIALIZE` | First state in funding phase |
| Normal off-chain update | `OPERATE` | Regular state transitions |
| Resize operation | `RESIZE` | Adding/removing channel funds |
| Cooperative close | `FINALIZE` | Final state before `close()` |

Calling `close()` with a state that lacks `intent = FINALIZE` will revert.

## Allocation Amounts

On-chain `Allocation.amount` is in **smallest token unit** (raw ERC-20 decimals). This is different from the off-chain app session API (which uses human-readable amounts). Do not mix them.

```solidity
// CORRECT: on-chain Solidity struct
Allocation({ destination: alice, token: usdc, amount: 50_000_000 })  // 50 USDC (6 decimals)

// CORRECT: off-chain app session (ClearNode API)
{ participant: alice, asset: 'usdc', amount: '50.0' }  // human-readable
```

## Challenge Period Sizing

Set challenge periods appropriate to the use case. Too short = participants can't respond.

| Use Case | Minimum | Recommended |
|----------|---------|-------------|
| High-frequency trading | 1 hour | 4 hours |
| Escrow / marketplace | 24 hours | 48 hours (`86400` seconds) |
| Staking / long-term | 7 days | 7 days (`604800` seconds) |

Never set challenge below 3600 seconds (1 hour) in production.

## Checkpointing Strategy

Checkpoint long-running channels periodically to bound worst-case dispute rollback:

- After every N state updates (e.g., N=100)
- After significant value movements
- Before extended participant downtime
- Minimum: at channel open and before close

## Resize Formula

The resize state data must encode `int256[]` delta amounts. Fund conservation must hold:

```
sum(allocations_resize_state) = sum(allocations_prev_state) + sum(delta_amounts)
```

Example: prev = `[5, 10]`, delta = `[-7, 6]`, resize = `[0, 14]`
Check: `0 + 14 = 5 + 10 + (-7 + 6)` → `14 = 15 + (-1)` = `14` ✓

## packedState Signing

The canonical payload to sign is deterministic:

```solidity
packedState = abi.encode(channelId, state.intent, state.version, state.data, state.allocations)
```

Never sign arbitrary JSON for on-chain state submission. The contract verifies against this exact encoding.

## Dispute Response: Time is Critical

If a challenge is filed against you, you have until the challenge expiration to respond. Monitor on-chain events. If you miss the window, the challenged state wins regardless of who was right.

```typescript
// Listen for challenge events
custodyContract.on('Challenged', (channelId, expiration) => {
  // Check if our version is newer
  if (ourLatestVersion > challengedVersion) {
    await custodyContract.checkpoint(channelId, ourLatestState, proofs);
  }
});
```

## Security Invariants

1. **Fund conservation**: Total allocations must equal funds locked in the channel at all times (for OPERATE/INITIALIZE/FINALIZE). For RESIZE, sum changes by the delta.
2. **Monotonic versions**: No state is valid if its version is not greater than all previously accepted states.
3. **Quorum signatures**: States must carry enough valid signatures to meet the adjudicator's requirements.
4. **No fund creation**: A state cannot create funds from nothing. The on-chain contract tracks total deposits.
