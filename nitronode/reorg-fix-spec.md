# Reorg Attack Fix — Confirmation Window Specification

## 1. Risk

The Nitronode event listener credits a user's off-chain balance the moment it observes a deposit event on-chain. If the block containing that deposit is subsequently removed from the canonical chain (a "reorganisation"), the off-chain credit persists while the on-chain deposit no longer exists. Because the credited balance can be transferred to a receiver before the node has any way to detect the reorg, the node ends up honoring an off-chain state transition that is permanently unbacked.

The worst-case outcome is a net loss of node liquidity equal to the sum of all deposit amounts that were credited during a reorg window and successfully drained to attacker-controlled receivers before the reorg was detected. There is no recovery path for the node once a signed receiver state exists.

This risk is meaningful on any chain where head-level reorgs occur naturally or can be induced. On modern fast-finality chains (BNB, Polygon post-Rio, Avalanche) the residual probability is very low. On Ethereum L1, depth-1 reorgs are routine and cryptoeconomic finality takes ~12.8 minutes.

---

## 2. Solution Overview

A **per-chain confirmation window** is introduced between raw event delivery and handler invocation. When the listener observes any event on chain C:

- It does **not** invoke the handler immediately.
- It waits for `confirmation_delay_sec` seconds (configured per chain in `blockchains.yaml`).
- If no reorg of the event's block occurs during that window, the handler is invoked normally.
- If the event's block is reorged out (`removed: true` log arrives), the pending invocation is cancelled with no side effects.
- If the reorged transaction is re-included (the same event appears again), the confirmation window restarts from zero.

The delay applies uniformly to **all** events, not only deposit-class ones. Selective gating would require the component to understand event semantics and introduce ordering hazards when events for different channels arrive interleaved — for example, a deposit event and a challenge event on separate channels could fire their handlers out of original arrival order if only the deposit is delayed. Uniform delay preserves the relative order of all events as they arrived from the chain while adding a single, predictable latency layer.

### 2.1 Residual risk and the finality trade-off

The confirmation window eliminates the reorg risk only when `confirmation_delay_sec` is set to or above the chain's cryptoeconomic finality time. For the representative values in §3:

- **Ethereum at 780s (~13 min):** matches Casper FFG hard finality. Reorging past this point requires ≥1/3 of total stake to be slashed. No residual risk.
- **Polygon at 10s, BNB at 5s:** exceeds the empirical reorg tail depth. Residual risk is negligible but not cryptoeconomically eliminated.
- **Ethereum at 36s (3 blocks, "quick" finality):** P(reorg depth ≥ 4) ≈ 10⁻⁵–10⁻⁶ per event. Residual risk is real.

When `confirmation_delay_sec` is set *below* the chain's finality time, **this specification acknowledges a residual risk**: it is possible — with low but non-zero probability — that an event passes the gate, the reactor commits it to the database, and the block containing that event is subsequently reorged out by a reorg deeper than the gate window.

When this occurs, the committed state (balance credit, channel open) has no corresponding on-chain event in the canonical chain. If the transaction is re-mined in the new canonical block, the reactor's idempotency guard (§6.6) handles the re-delivery cleanly. If it is not re-mined, the DB retains stale state that can only be partially corrected on the next node restart via the reconciliation walk (§4.4). There is no automated rollback; the exposure scales with the deposit value and is bounded by the probability of deep reorgs on the target chain.

Operators who cannot accept this residual exposure should set `confirmation_delay_sec` to the chain's hard-finality time (Ethereum: 780s; Polygon: `finalized` tag resolves to ~5s; L2s: `finalized` maps to L1 Casper FFG at ~13 min). The gate's detection mechanisms (§6.5, §6.6) provide observability when the residual-risk scenario occurs.

---

## 3. Configuration

A new `confirmation_delay_sec` field is added per chain in `blockchains.yaml`. Representative values:

```yaml
chains:
  - id: 1          # Ethereum mainnet
    confirmation_delay_sec: 780   # ~13 min — Casper FFG hard finality
  - id: 137        # Polygon PoS (post-Heimdall v2 / Rio)
    confirmation_delay_sec: 10    # 5 blocks × ~2s; empirical reorg tail is sub-10s
  - id: 56         # BNB Smart Chain
    confirmation_delay_sec: 5     # fast-finality, ~3-4 blocks
  - id: 42161      # Arbitrum One
    confirmation_delay_sec: 120   # L2 `safe` tag (L1-posted batch), ~1-2 min
  - id: 8453       # Base
    confirmation_delay_sec: 120   # same L2 `safe` semantics
```

`confirmation_delay_sec: 0` disables the gate — events are processed immediately. Appropriate for BFT single-slot chains where the node operator accepts the negligible residual risk, or for chains using a finality-tag subscription rather than a block-count gate.

---

## 4. Confirmation Window Behavior

### 4.1 Normal path

When a log `E` arrives (without `Removed: true`):

1. Record the event under a key of `(txHash, blockHash, logIndex)`.
2. Start an in-memory timer for the chain's `confirmation_delay_sec`.
3. When the timer fires, invoke the event handler.

### 4.2 Reorg path

If a log with `Removed: true` arrives for the same `(txHash, blockHash, logIndex)` before the timer fires:

- Cancel the pending timer.
- Do not invoke the handler — no state change occurs.
- The listener remains active. When the same transaction is re-included, its event will be delivered again (without `Removed: true`) and the gate starts a fresh window under the new block's key.

### 4.3 Out-of-order delivery

The re-added event (no `Removed: true`, new block) may arrive at the listener before the corresponding `Removed: true` log for the old block. Because the re-added event is in a different block, it carries a different `blockHash` and therefore a different key. The two events are handled independently:

- The re-added event starts a fresh timer under its own key.
- The `Removed: true` log, when it arrives, looks up the OLD block's key — which has no pending timer (it was never created, or it already expired) — and performs a no-op.

This means out-of-order delivery requires no special case beyond the normal path: the block-scoped key prevents the remove from accidentally cancelling the re-added event's timer.

- On a `Removed: true` log for a key that **has no pending timer**: no-op. The event either confirmed and was already processed (reorg arrived after the window), or belongs to a different block whose timer was never started (re-add arrived first, has its own key).

> Repeated reorgs of the same transaction are theoretically possible but imply a chain-level consensus failure. The gate's cancel/restart cycle handles each naturally; no special cap is needed.

### 4.4 Startup and reconciliation

#### Prerequisites

Before the reconciliation logic described below can function, `block_hash` must be added as a column to `contract_events` and to the `core.BlockchainEvent` struct. The value is available in `types.Log.BlockHash` at the time the gate calls the reactor. Without this column, reorg detection in steps 2–4 is not possible.

#### Definition: latest processed block

The **latest processed block** for a chain is the highest block number at which the reactor successfully committed at least one event to the database — identical to the listener's existing startup cursor (`MAX(block_number)` in `contract_events` for this `blockchain_id` and contract address, computed by `GetLatestContractEventBlockNumber`). This is distinct from the highest block the listener ever *saw*: the listener may have seen many blocks that contained no relevant events and therefore left no `contract_events` rows.

#### Reconciliation steps

On startup, for each chain, after the `block_hash` migration has been applied:

1. Query `contract_events` for the latest committed event: `latestBlockNum = MAX(block_number)`, `latestBlockHash = block_hash` at that row. If no rows exist, start the scan from the chain's configured genesis / start block and skip to step 5.
2. Call `eth_getBlockByHash(latestBlockHash)` on the chain's RPC.
   - If the response is non-null: `latestBlockHash` is still in the canonical chain — no reorg above this block. Proceed to step 4.
   - If the response is null: the block has been reorged out. Proceed to step 3.
3. **Common-ancestor walk using stored block hashes:** query `contract_events` for the next-older distinct `block_hash` (the highest `block_number` strictly below the current candidate). Repeat step 2 with this hash. Continue until a block hash is found that is still in the canonical chain, or until no older stored hash exists (treat genesis as the fallback). This height is the **common ancestor**.

   > **Why walk stored hashes, not block numbers?** In normal operation most blocks contain no `ChannelHub` events, so `contract_events` has no row for them. A block-number walk would find nothing to compare at event-gap heights and could miss a reorg that occurred entirely within such a gap. Walking by stored block hashes ensures every comparison is against a block the reactor actually processed.

4. Set the scan start to `commonAncestorBlockNum`. Events between `commonAncestorBlockNum` and `latestBlockNum` that came from the reorged fork are still present in the DB. The reactor has no rollback mechanism for those rows — the re-scan below will re-apply canonical events over them where the transaction was re-mined (idempotent), and leave the orphaned DB state in place where the transaction was not re-mined (residual risk; see §2.1). State-setting operations (`UpdateChannel`, `RefreshUserEnforcedBalance`) will overwrite with canonical values for re-mined events; rows from dropped transactions remain as stale data with no automated cleanup.
5. Start the event scan from `commonAncestorBlockNum` (or genesis if step 1 found no rows). Feed all replayed events **directly to the reactor, bypassing the gate entirely**. Historical events come from `eth_getLogs` and are, by definition, already in the current canonical chain. The common-ancestor walk in steps 2–3 additionally confirms that the starting block is canonical. There is no incremental reorg risk to guard against for these events, and applying a full confirmation delay would only stall the node on restart for no safety benefit. The gate applies exclusively to live WebSocket events; any reorgs of very-recent blocks during replay are handled by the buffered live-subscription signals processed immediately after replay completes (step 7).
6. The reactor is idempotent for replayed events: `HandleHomeChannelCreated` has an explicit early-return guard when the channel is already open; `HandleHomeChannelCheckpointed` and `RefreshUserEnforcedBalance` use set-semantics (not accumulation) and recompute from the latest DB state. `StoreContractEvent` is called last inside the DB transaction and enforces a unique constraint on `(transaction_hash, log_index, blockchain_id)`. If a duplicate is inserted, Postgres returns a constraint-violation error, causing the entire transaction (including all state mutations in the same `useStoreInTx` call) to roll back. The reactor therefore cannot double-apply state changes for an event it has already committed.
7. Historical log queries (`eth_getLogs`) return only canonical chain events — there are no `Removed: true` signals during replay. The gate operates in timer-only mode during reconciliation. Removal signals from the live WebSocket subscription that arrive during the replay phase are buffered in the listener's `currentCh` and processed only after the historical replay phase completes.

---

## 5. Scope

The delay applies to **all** events emitted by the `ChannelHub` contract on a given chain. No filtering by event type is performed inside the gate.

> **Note:** `ChannelCreated` (`handleHomeChannelCreated`) calls `RefreshUserEnforcedBalance`. Verify whether the initial channel state carries a non-zero deposit; if it does, the uniform delay already protects it — no special casing is needed.

---

## 6. Implementation Notes

### 6.1 Component placement and wiring

The `ConfirmationGate` is a thin in-memory component that sits between the raw log stream (`listener.go`) and the `ChannelHubReactor`.

**Existing wiring** (`nitronode/main.go:127-129`):

```go
reactor := evm.NewChannelHubReactor(b.ID, ...)
l := evm.NewListener(..., reactor.HandleEvent, ...)
```

The listener accepts a handler of type `HandleEvent func(ctx context.Context, eventLog types.Log) error`. The gate exposes the same signature and is inserted between the two:

```go
reactor := evm.NewChannelHubReactor(b.ID, ...)
gate    := evm.NewConfirmationGate(confirmationDelay, reactor.HandleEvent)
l       := evm.NewListener(..., gate.HandleEvent, ...)
```

The reactor itself does not change. All the listener's existing logic — subscription management, cursor tracking, reconnection, historical replay — is unaffected.

**Handling `Removed: true` logs:** currently `listener.go:289-294` skips removed logs before they reach the handler. This skip must be moved: the listener should forward removed logs to `gate.HandleEvent` (they still carry the `Removed` flag on `types.Log`), and the gate alone decides whether to cancel a pending timer or ignore the signal. The reactor never sees a `Removed: true` log.

### 6.2 Event identity for removal scanning

The Listener delivers events in strict block order, so the queue is naturally ordered by arrival time. When a `Removed: true` log arrives in the Pusher, it scans the queue for the **first** entry matching `(txHash, logIndex)` and deletes it.

`blockHash` is deliberately excluded from the removal scan key. Because the queue is FIFO and reorgs produce the re-add event *after* the original event, the original always sits earlier in the queue than any re-add. Scanning for `(txHash, logIndex)` and deleting the first match therefore always targets the original entry and leaves any re-add untouched.

A single transaction can emit multiple events for the same `txHash` (e.g., two `ChannelDeposited` logs in a batch open). `logIndex` disambiguates these; it is unique per log within a block and is present in both the live event and its corresponding `Removed: true` log.

`blockHash` is still present in each `types.Log` stored in the queue and is used by:
- The `recentlyForwarded` detection map (§6.5) — keyed by `(txHash, blockHash, logIndex)` to identify which specific occurrence was forwarded.
- `StoreContractEvent` in the reactor — stored in `contract_events` for the reconciliation walk (§4.4).

### 6.3 Two-goroutine design

**Data structure:** a FIFO queue of `(types.Log, arrivedAt time.Time)`. Naturally ordered by arrival time because the Listener delivers events in strict block order.

```go
type queueEntry struct {
    log       types.Log
    arrivedAt time.Time
}

type eventKey struct {          // used for removal scan
    txHash   common.Hash
    logIndex uint
}

type forwardedKey struct {      // used for post-gate reorg detection
    txHash    common.Hash
    blockHash common.Hash
    logIndex  uint
}

type ConfirmationGate struct {
    delay             time.Duration
    chainID           uint64
    handler           HandleEvent
    queue             []queueEntry               // protected by mu
    recentlyForwarded map[forwardedKey]time.Time  // protected by mu; TTL = 2× delay
    mu                sync.Mutex
}
```

---

**Goroutine 1 — Pusher** (driven by the existing Listener; implements the `HandleEvent` signature)

Receives `types.Log` from the Listener. On each event:

- If `Removed: true` — scan the queue for the first entry matching `(txHash, logIndex)` and delete it. If no match found, check `recentlyForwarded` for a post-gate reorg signal (see §6.5).
- Otherwise — append `(log, time.Now())` to the queue tail.

No expiration check, no forwarding. Push only.

---

**Goroutine 2 — Poller**

Wakes every ~50 ms on a ticker. Each wake:

- Inspect the queue front.
- While `front.arrivedAt + delay ≤ now`: pop the entry, record `forwardedKey{txHash, blockHash, logIndex}` in `recentlyForwarded` with the current timestamp, then forward the log to the Reactor outside the lock.
- Stop as soon as the front is not yet ready — everything behind it is newer.
- Sleep until next tick.

No event handling, no Listener awareness. Drain-and-forward only.

---

**Properties**

| Property | Detail |
| --- | --- |
| Zero RPC calls in the gate | Delay is a pure `time.Duration`; no chain queries |
| Chain-agnostic | `confirmationDelay` is the only chain-specific input |
| Forward latency after window | At most one tick (~50 ms) |
| Reorg within window | Pusher's scan removes the entry; Reactor never sees the event |
| Reorg deeper than window | Rare; Reactor-level idempotency (§6.6) handles re-delivered events |
| Concurrency | Both goroutines share `mu`; Reactor is called outside the lock |
| Shutdown | Poller exits on `ctx.Done()`; entries still in queue are discarded (safe — they were never forwarded) |

### 6.4 Exposing `confirmation_delay_secs` via API

Clients need to know the confirmation delay for each chain so they can display the correct waiting time to users after submitting a deposit. The best existing candidate is **`node.v1.GetConfig`**, which already returns a per-chain `BlockchainInfoV1` object.

Files to update:

- `pkg/rpc/types.go` — add `ConfirmationDelaySecs uint64` to `BlockchainInfoV1`.
- `nitronode/api/node_v1/utils.go` — populate the new field in `mapBlockchainV1` from the chain's loaded config.
- `pkg/core/types.go` (or wherever `core.Blockchain` is defined) — add `ConfirmationDelaySec uint64` so the value flows from `blockchains.yaml` through config loading into the API handler.

No new endpoint is needed. The field appears alongside existing per-chain fields (contract addresses, asset list, block time) and is read-only from the client's perspective.

### 6.5 Post-gate reorg detection in the gate

The `recentlyForwarded` map (already in the `ConfirmationGate` struct, §6.3) provides detection without any DB access. The **Poller** writes to it each time it forwards an event; the **Pusher** reads from it when a `Removed: true` log arrives and the queue scan finds no matching entry.

When `Removed: true` arrives in the Pusher:

- **Match found in queue** → normal removal; no log.
- **No match in queue, but `forwardedKey{txHash, blockHash, logIndex}` found in `recentlyForwarded`** → the event was already forwarded to the Reactor and its block has now been reorged out. Log at **`WARN`** with `txHash`, `blockHash`, `logIndex`, `chainID`. Remove the entry.
- **Match in neither** → log at `DEBUG` ("removal for unknown/stale event" — predates the current run or arrived long after the TTL).

`recentlyForwarded` entries are evicted lazily: when the Pusher reads an entry, it checks `time.Since(forwardedAt) > 2 × delay` and discards stale entries on access. The map stays small because post-gate reorgs are rare and `Removed: true` arrives within one or two block-times of the reorg. No separate cleanup goroutine is needed.

### 6.6 Reactor defense-in-depth: skip re-delivered events

When the gate lets a re-added event through (same tx re-mined in a new block after a reorg, confirmed by a fresh timer), the reactor would attempt to process an event it has already committed. Currently this surfaces as a DB constraint-violation error and a full transaction rollback — noisy and potentially confusing.

Add a new method to `ChannelHubReactorStore`:

```go
// IsContractEventProcessed reports whether an event identified by
// (txHash, logIndex, blockchainID) has already been committed,
// regardless of which block it appeared in.
IsContractEventProcessed(txHash string, logIndex uint, blockchainID uint64) (bool, error)
```

At the top of `HandleEvent`, before entering `useStoreInTx`, call this method. If the event is already committed, log at **`INFO`** ("skipping re-delivered event, already committed") and return `nil` immediately. No transaction is opened; no state is touched.

The existing unique constraint on `(transaction_hash, log_index, blockchain_id)` in `contract_events` remains as the definitive safety net. This pre-check converts the constraint-violation rollback path into a clean, explicit, logged early exit that also serves as the idempotency guard for the reconciliation re-scan path.

Together, §6.5 and §6.6 produce two complementary log signals:

| Signal | Source | Level | Meaning |
| --- | --- | --- | --- |
| "post-gate reorg detected for event X" | Gate | WARN | Committed block was reorged; residual-risk scenario is active |
| "skipping re-delivered event X" | Reactor | INFO | Same tx re-mined; reactor correctly skips it |

If the operator sees the WARN but never the INFO, the transaction was not re-mined — the stale DB state from §2.1 is in effect.

### 6.7 Block timestamp cache

#### Purpose

The gate uses the **block timestamp** of each event as its `arrivedAt` reference rather than wall-clock time. This ensures that events replayed from historical blocks (whose timestamps are minutes or hours in the past) are forwarded immediately on the first Poller tick, without waiting for the full confirmation delay to elapse again.

Fetching the block timestamp requires one `eth_getBlockByHash` RPC call per block. A single block can produce multiple events (e.g. two `ChannelDeposited` logs in a batch open). The **block timestamp cache** avoids the redundant RPC calls: the first event from a block fetches and stores the timestamp; subsequent events from the same block read it from the cache.

#### Data structure

```go
blockTimestampCache map[common.Hash]time.Time // protected by mu; evicted by Poller
```

The cache is keyed by `blockHash`. Values are written once (on the first event from a block) and are never modified.

#### Eviction

The cache grows monotonically without eviction: every block that produces at least one relevant event adds a permanent entry. Over the lifetime of a long-running node, this is an unbounded memory leak.

Entries are evicted by the Poller in the same sweep pass that cleans `recentlyForwarded`. An entry is safe to remove once:

> `now − blockTimestamp > recentMultiplier × delay`

At that age, every event from the block has either been forwarded (within `delay` of its `arrivedAt`) or cancelled by a `Removed: true` signal. No new event from the same block can arrive after it (the listener delivers events in ascending block order). The cached timestamp therefore serves no further purpose.

**Eviction is performed in `poll()`, under the mutex, after the `recentlyForwarded` sweep:**

```go
for bh, ts := range g.blockTimestampCache {
    if now.Sub(ts) > recentMultiplier*g.delay {
        delete(g.blockTimestampCache, bh)
    }
}
```

#### Bound after eviction

With eviction, the cache holds at most one entry per block whose timestamp falls within the window `[now − recentMultiplier×delay, now]`. That is at most `recentMultiplier × delay × (blocks per second)` entries — a small constant for every supported chain.

Each entry is 56 bytes (`common.Hash` 32 B + `time.Time` 24 B). Even the worst case would be under 100 KB.
