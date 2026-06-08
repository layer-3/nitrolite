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

The re-added event (no `Removed: true`, new block) may arrive at the listener before the corresponding `Removed: true` log for the old block. When this happens, the gate **replaces** the pending entry for `(txHash, logIndex)` with the new one and resets the confirmation timer under the new block's key:

- On the non-removed re-add, scan the queue by `(txHash, logIndex)` — ignoring `blockHash` — and drop any existing entry. Append the new event with a fresh `arrivedAt`.
- The subsequent `Removed: true` log for the OLD block carries the old `blockHash` and therefore matches neither the queued (new-block) entry nor any `recentlyForwarded` record. It performs a no-op.

This collapses the two-entry coexistence model into a single live entry per `(txHash, logIndex)`. The behavior is observationally equivalent — exactly one event is forwarded, and it is the latest re-mining — and it removes the only state-divergence path between the queue and `recentlyForwarded`.

- On a `Removed: true` log for a key that **has no pending timer and no `recentlyForwarded` record**: no-op. The event either belongs to a block that was already replaced by a later re-add (handled above), or it is a stale removal from a fork the gate has no record of.

> Repeated reorgs of the same transaction are theoretically possible but imply a chain-level consensus failure. The gate's replace/restart cycle handles each naturally; no special cap is needed.

### 4.4 Startup and reconciliation

#### Prerequisites

Before the reconciliation logic described below can function, `block_hash` must be added as a column to `contract_events` and to the `core.BlockchainEvent` struct. The value is available in `types.Log.BlockHash` at the time the gate calls the reactor. Without this column, reorg detection in steps 2–4 is not possible.

**Why `block_hash` is the minimal required addition — and why alternatives fail:**

The reconciliation walk needs to answer one question per stored block: "is this specific block still in the canonical chain?" The only RPC call that answers it directly is `eth_getBlockByHash(hash)` — it returns `null` if the block is no longer canonical. Without the stored hash, two alternatives were evaluated and both fail:

- **`block_number` alone is insufficient.** After a reorg, a *different* block can occupy the same height. Calling `eth_getBlockByNumber(storedBlockNumber)` always returns a block — but it may be a new block from the reorged fork. Without the original hash there is no way to tell whether the block returned is the one the reactor processed.

- **`transaction_hash` via `eth_getTransactionReceipt` is insufficient.** A block can be reorged out even if every one of its transactions was re-mined in a new block at the same height. In that case all receipt lookups return `blockNumber` matching the stored value, but the original block is gone and the stored DB state no longer corresponds to the canonical chain. Additionally, the backward walk (step 3) must traverse every stored *block* in descending order; rows in `contract_events` only exist for blocks that contained a `ChannelHub` event. A reorg that diverged entirely within a gap — blocks with no relevant events — is invisible to a tx-receipt-based walk.

`block_hash` is a single `CHAR(66)` column. Its addition enables exact, O(1)-per-step canonicality checks and is the only approach that handles all reorg scenarios correctly.

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
5. Start the event scan from `commonAncestorBlockNum` (or genesis if step 1 found no rows). Replayed events may flow through the same path as live events (Listener → gate → reactor); the gate does not require a separate bypass. Historical events come from `eth_getLogs` and are, by definition, already in the current canonical chain — the common-ancestor walk in steps 2–3 additionally confirms that the starting block is canonical, so there is no incremental reorg risk to guard against. Provided that the gate uses each event's **block timestamp** as the `arrivedAt` reference (see §6.7), historical events are immediately mature on first poll and forward without per-event delay; the only added cost is one block-timestamp RPC per unique historical block and at most one poll-tick of latency.
6. The reactor is idempotent for replayed events: `HandleHomeChannelCreated` has an explicit early-return guard when the channel is already open; `HandleHomeChannelCheckpointed` and `RefreshUserEnforcedBalance` use set-semantics (not accumulation) and recompute from the latest DB state. `StoreContractEvent` is called last inside the DB transaction and enforces a unique constraint on `(transaction_hash, log_index, blockchain_id)`. If a duplicate is inserted, Postgres returns a constraint-violation error, causing the entire transaction (including all state mutations in the same `useStoreInTx` call) to roll back. The reactor therefore cannot double-apply state changes for an event it has already committed.
7. Historical log queries (`eth_getLogs`) return only canonical chain events — there are no `Removed: true` signals during replay. Removal signals from the live WebSocket subscription that arrive during the replay phase are buffered in the listener's `currentCh` and reach the gate only after the historical replay phase completes; if they cancel a re-mined event that has already been forwarded, the post-gate reorg detection in §6.5 logs them.

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

### 6.2 Event identity for queue keying

The Listener delivers events in strict block order, so the queue is naturally ordered by arrival time. Two distinct scan keys are used against the queue:

- **`(txHash, logIndex)` — used by both Pusher paths (non-removed re-add and removed cancellation).** On a non-removed arrival, any existing entry with the same `(txHash, logIndex)` is dropped and the new event appended with a fresh `arrivedAt`. Because re-adds always replace the prior entry, the queue holds at most one entry per `(txHash, logIndex)` at any time.
- **`(txHash, blockHash, logIndex)` — used by the Removed-cancel scan against the queue.** A `Removed: true` log only cancels a queued entry when the full key matches. A Removed for an OLD block whose entry has already been replaced by a newer re-add will not match the queued (new-block) entry and will fall through to the `recentlyForwarded` lookup (§6.5).

`blockHash` is excluded from the re-add scan key so that a re-mining of the same tx replaces the original regardless of which block it landed in. `blockHash` is included on the Removed scan so that a stale removal for an already-replaced fork cannot cancel a live entry.

A single transaction can emit multiple events for the same `txHash` (e.g., two `ChannelDeposited` logs in a batch open). `logIndex` disambiguates these; it is unique per log within a block and is present in both the live event and its corresponding `Removed: true` log.

`blockHash` is also used by:
- The `recentlyForwarded` detection map (§6.5) — keyed by `(txHash, blockHash, logIndex)` to identify which specific occurrence was forwarded.
- `StoreContractEvent` in the reactor — stored in `contract_events` for the reconciliation walk (§4.4).

### 6.3 Two-goroutine design

**Data structure:** a FIFO queue of `(types.Log, arrivedAt time.Time)`. Naturally ordered by arrival time because the Listener delivers events in strict block order.

```go
type queueEntry struct {
    log       types.Log
    arrivedAt time.Time
}

type eventKey struct {          // used for re-add scan (replace prior entry)
    txHash   common.Hash
    logIndex uint
}

type forwardedKey struct {      // used for Removed-cancel scan and post-gate reorg detection
    txHash    common.Hash
    blockHash common.Hash
    logIndex  uint
}

type ConfirmationGate struct {
    delay             time.Duration
    chainID           uint64
    handler           HandleEvent
    queue             []queueEntry                // protected by mu
    recentlyForwarded map[forwardedKey]time.Time  // protected by mu; entries are kept for a small multiple of `delay` (see §6.5)
    mu                sync.Mutex
}
```

---

**Goroutine 1 — Pusher** (driven by the existing Listener; implements the `HandleEvent` signature)

Receives `types.Log` from the Listener. On each event:

- If `Removed: true` — scan the queue for an entry matching the full `(txHash, blockHash, logIndex)` key and delete it. If no match is found, check `recentlyForwarded` for a post-gate reorg signal (see §6.5).
- Otherwise — drop any existing queue entry with the same `(txHash, logIndex)` (ignoring `blockHash`), then append `(log, arrivedAt)` to the queue tail. `arrivedAt` is the block timestamp (see §6.7), falling back to `time.Now()` only on fetch failure.

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

`recentlyForwarded` entries are evicted on a TTL that is a small multiple of `delay` — long enough that any `Removed: true` for a forwarded event arrives while the entry is still present, short enough that the map remains bounded. The exact multiplier is an implementation choice (current value: see `recentMultiplier` in `confirmation_gate.go`; e.g. 2 or 3 work in practice). Eviction may be performed lazily on Pusher access, in a periodic Poller sweep, or by any equivalent strategy; the post-gate detection contract above is what matters, not the eviction mechanism. The map stays small because post-gate reorgs are rare and `Removed: true` arrives within one or two block-times of the reorg. No separate cleanup goroutine is required.

### 6.6 Reactor defense-in-depth: skip re-delivered events

When a re-added event reaches the reactor (same tx re-mined in a new block after a reorg, confirmed by a fresh gate timer), the reactor attempts to process an event it has already committed. This guard converts what is currently a DB constraint-violation error and a full transaction rollback into a clean, explicit logged exit.

**Important limitation:** this guard identifies events by `(txHash, logIndex, blockchainID)`, where `log_index` is a **block-level** index in go-ethereum — the position of this log among all logs in the entire block, across all transactions. If a transaction is re-mined in a new block where different transactions precede it, its logs receive different block-level `log_index` values. The new `(txHash, newLogIndex, blockchainID)` tuple does not match any committed row, so `IsContractEventProcessed` returns `false` and **the reorged event passes through this check**. In that case the reactor's business-logic idempotency is the actual guard (see below). This guard therefore only catches exact re-deliveries — cases where `log_index` is unchanged.

Add a new method to `ChannelHubReactorStore`:

```go
// IsContractEventProcessed reports whether an event identified by
// (txHash, logIndex, blockchainID) has already been committed,
// regardless of which block it appeared in.
// NOTE: uses block-level logIndex — does not detect reorged events
// where the same tx re-mines with a different block-level log position.
IsContractEventProcessed(txHash string, logIndex uint, blockchainID uint64) (bool, error)
```

At the top of `HandleEvent`, before entering `useStoreInTx`, call this method. If the event is already committed, log at **`INFO`** ("skipping re-delivered event, already committed") and return `nil` immediately. No transaction is opened; no state is touched.

Reorged events that pass through this check are still neutralized by the reactor's **business-logic idempotency**:

- `HandleHomeChannelCreated` has an explicit early-return when the channel is already open.
- `HandleHomeChannelCheckpointed` and `RefreshUserEnforcedBalance` use set-semantics (overwrite, not accumulate).
- The `StoreContractEvent` unique constraint on `(transaction_hash, log_index, blockchain_id)` remains as the final backstop for the case where `log_index` happens to be unchanged.

The value of `IsContractEventProcessed` is therefore:

1. **Noise reduction for exact re-deliveries** — converts a constraint-violation rollback (logged as an error by the gate poller) into a clean INFO exit with no DB transaction opened.
2. **Correctness for the reconciliation walk (§4.4)** — when the node replays already-processed historical events on startup, every re-delivered event would otherwise produce a constraint-violation error and potentially stall the walk. This pre-check makes the reconciliation path viable.

Together, §6.5 and §6.6 produce two complementary log signals:

| Signal | Source | Level | Meaning |
| --- | --- | --- | --- |
| "post-gate reorg detected for event X" | Gate | WARN | Committed block was reorged; residual-risk scenario is active |
| "skipping re-delivered event X" | Reactor | INFO | Same tx re-mined at same block position; reactor correctly skips it |

If the operator sees the WARN but never the INFO, either the transaction was not re-mined, or it was re-mined at a different block position (this check did not fire; business-logic idempotency handled it silently).

#### Reorg-safe idempotency — separate task

To make the idempotency check itself robust to reorged events regardless of block position, the idempotency key must be stable across re-mining. The block-level `log_index` is not stable; a **tx-relative log index** is.

The tx-relative log index is the 0-based position of a log within its own transaction's emitted logs. It is invariant: the same transaction always emits the same logs in the same order, so its tx-relative indices never change across reorgs. The EVM guarantees that all logs of a transaction arrive consecutively in ascending block-level order, so the tx-relative index can be computed in-process as:

```
tx_log_index = l.Index - min(l.Index for all logs of l.TxHash in this block)
```

No RPC call is required — the minimum is established by the first log of each transaction seen in a block, which always arrives before subsequent logs of the same transaction.

Implementing this requires:

- **DB migration**: add `tx_log_index` column to `contract_events`; replace the unique index `(transaction_hash, log_index, blockchain_id)` with `(transaction_hash, tx_log_index, blockchain_id)`.
- **`BlockchainEvent` struct**: add `TxLogIndex uint32` field.
- **Reactor**: maintain a small in-memory map `(blockHash, txHash) → minBlockLogIndex` to compute `tx_log_index` for each incoming event; evict entries when a new block is first seen.
- **`IsContractEventProcessed` and `StoreContractEvent`**: operate on `tx_log_index` instead of `log_index`.

**This is a separate task.** It is not part of the current confirmation-gate scope. Until it is implemented, the reactor relies on business-logic idempotency for the reorged-different-position case, which is correct but not explicitly guarded at the storage layer.

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
