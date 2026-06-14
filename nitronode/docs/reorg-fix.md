# Reorg Attack Fix — Confirmation Window Specification

## 1. Risk

The Nitronode event listener credits a user's off-chain balance the moment it observes a deposit event on-chain. If the block containing that deposit is subsequently removed from the canonical chain (a "reorganisation"), the off-chain credit persists while the on-chain deposit no longer exists. Because the credited balance can be transferred to a receiver before the node has any way to detect the reorg, the node ends up honoring an off-chain state transition that is permanently unbacked.

The worst-case outcome is a net loss of node liquidity equal to the sum of all deposit amounts that were credited during a reorg window and successfully drained to attacker-controlled receivers before the reorg was detected. There is no recovery path for the node once a signed receiver state exists.

This risk is meaningful on any chain where head-level reorgs occur naturally or can be induced. On modern fast-finality chains (BNB, Polygon post-Rio, Avalanche) the residual probability is very low. On Ethereum L1, depth-1 reorgs are routine and cryptoeconomic finality takes ~12.8 minutes.

---

## 2. Solution Overview

A **per-chain confirmation window** is introduced between raw event delivery and handler invocation. When the listener observes any event on chain C:

- It does **not** invoke the handler immediately.
- It waits for `confirmation_delay_secs` seconds (configured per chain in `blockchains.yaml`).
- If no reorg of the event's block occurs during that window, the handler is invoked normally.
- If the event's block is reorged out (`removed: true` log arrives), the pending invocation is cancelled with no side effects.
- If the reorged transaction is re-included (the same event appears again), the confirmation window restarts from zero.

The delay applies uniformly to **all** events, not only deposit-class ones. Selective gating would require the component to understand event semantics and introduce ordering hazards when events for different channels arrive interleaved — for example, a deposit event and a challenge event on separate channels could fire their handlers out of original arrival order if only the deposit is delayed. Uniform delay preserves the relative order of all events as they arrived from the chain while adding a single, predictable latency layer.

### 2.1 Residual risk and the finality trade-off

The confirmation window eliminates the reorg risk only when `confirmation_delay_secs` is set to or above the chain's cryptoeconomic finality time. For the representative values in §3:

- **Ethereum at 780s (~13 min):** matches Casper FFG hard finality. Reorging past this point requires ≥1/3 of total stake to be slashed. No residual risk.
- **Polygon at 10s, BNB at 5s:** exceeds the empirical reorg tail depth. Residual risk is negligible but not cryptoeconomically eliminated.
- **Ethereum at 36s (3 blocks, "quick" finality):** P(reorg depth ≥ 4) ≈ 10⁻⁵–10⁻⁶ per event. Residual risk is real.

When `confirmation_delay_secs` is set *below* the chain's finality time, **this specification acknowledges a residual risk**: it is possible — with low but non-zero probability — that an event passes the gate, the reactor commits it to the database, and the block containing that event is subsequently reorged out by a reorg deeper than the gate window.

When this occurs, the committed state (balance credit, channel open) has no corresponding on-chain event in the canonical chain. If the transaction is re-mined in the new canonical block, the reactor's idempotency guard (§6.6) handles the re-delivery cleanly. If it is not re-mined, the DB retains stale state that can only be partially corrected on the next node restart via the reconciliation walk (§4.4). There is no automated rollback; the exposure scales with the deposit value and is bounded by the probability of deep reorgs on the target chain.

Operators who cannot accept this residual exposure should set `confirmation_delay_secs` to the chain's hard-finality time (Ethereum: 780s; Polygon: `finalized` tag resolves to ~5s; L2s: `finalized` maps to L1 Casper FFG at ~13 min). The gate's detection mechanisms (§6.5, §6.6) provide observability when the residual-risk scenario occurs.

---

## 3. Configuration

A new `confirmation_delay_secs` field is added per chain in `blockchains.yaml`. Representative values:

```yaml
chains:
  - id: 1          # Ethereum mainnet
    confirmation_delay_secs: 780   # ~13 min — Casper FFG hard finality
  - id: 137        # Polygon PoS (post-Heimdall v2 / Rio)
    confirmation_delay_secs: 10    # 5 blocks × ~2s; empirical reorg tail is sub-10s
  - id: 56         # BNB Smart Chain
    confirmation_delay_secs: 5     # fast-finality, ~3-4 blocks
  - id: 42161      # Arbitrum One
    confirmation_delay_secs: 120   # L2 `safe` tag (L1-posted batch), ~1-2 min
  - id: 8453       # Base
    confirmation_delay_secs: 120   # same L2 `safe` semantics
```

`confirmation_delay_secs: 0` disables the gate — events are processed immediately. Appropriate for BFT single-slot chains where the node operator accepts the negligible residual risk, or for chains using a finality-tag subscription rather than a block-count gate.

---

## 4. Confirmation Window Behavior

### 4.1 Normal path

When a log `E` arrives (without `Removed: true`):

1. Record the event in the live-entry map under `(txHash, logIndex)` with its `blockHash` as the tombstone discriminator, and append it to the FIFO drain queue with its block timestamp as `arrivedAt`.
2. The gate's drain goroutine (single shared timer per gate; see §6.3) treats the entry as eligible once `arrivedAt + confirmation_delay_secs` has elapsed.
3. When the entry matures, invoke the event handler.

### 4.2 Reorg path

If a log with `Removed: true` arrives for the same `(txHash, blockHash, logIndex)` before the timer fires:

- Cancel the pending timer.
- Do not invoke the handler — no state change occurs.
- The listener remains active. When the same transaction is re-included, its event will be delivered again (without `Removed: true`) and the gate starts a fresh window under the new block's key.

### 4.3 Out-of-order delivery

The re-added event (no `Removed: true`, new block) may arrive at the listener before the corresponding `Removed: true` log for the old block. When this happens, the gate **replaces** the pending entry for `(txHash, logIndex)` with the new one and resets the confirmation timer under the new block's key:

- On the non-removed re-add, overwrite `pending[(txHash, logIndex)]` with the new `blockHash` and append the new event to the queue tail with a fresh `arrivedAt`. The earlier queue entry remains in place as a tombstone — its `blockHash` no longer matches `pending`, so the drain goroutine silently skips it when it reaches the head.
- The subsequent `Removed: true` log for the OLD block carries the old `blockHash` and therefore matches neither `pending` (whose value is now the new block's hash) nor any `forwardedSet` record. It performs a no-op.

The tombstone-map design replaces the prior slice-scan approach: every live operation is O(1), and exactly one event per `(txHash, logIndex)` is forwarded — the latest re-mining.

- On a `Removed: true` log for a key that **has no live `pending` entry and no `forwardedSet` record**: no-op. The event either belongs to a block that was already replaced by a later re-add (handled above), or it is a stale removal from a fork the gate has no record of.

> Repeated reorgs of the same transaction are theoretically possible but imply a chain-level consensus failure. The gate's replace/restart cycle handles each naturally; no special cap is needed.

### 4.4 Startup and reconciliation

#### Prerequisites

Before the reconciliation logic described below can function, `block_hash` must be added as a column to `contract_events` and to the `core.BlockchainEvent` struct. The value is available in `types.Log.BlockHash` at the time the gate calls the reactor. Without this column, reorg detection in steps 2–4 is not possible.

**Why `block_hash` is the minimal required addition — and why alternatives fail:**

The reconciliation walk needs to answer one question per stored block: "is this specific block still in the canonical chain?" The definitive answer combines the stored hash with an `eth_getBlockByNumber(storedBlockNumber)` lookup — the canonical chain has exactly one block at each height, and comparing its hash to the stored hash tells us whether the stored block is still canonical. Without the stored hash, two alternatives were evaluated and both fail:

- **`block_number` alone is insufficient.** After a reorg, a *different* block can occupy the same height. Calling `eth_getBlockByNumber(storedBlockNumber)` always returns a block — but it may be a new block from the reorged fork. Without the original hash there is no way to tell whether the block returned is the one the reactor processed.

- **`transaction_hash` via `eth_getTransactionReceipt` is insufficient.** A block can be reorged out even if every one of its transactions was re-mined in a new block at the same height. In that case all receipt lookups return `blockNumber` matching the stored value, but the original block is gone and the stored DB state no longer corresponds to the canonical chain. Additionally, the backward walk (step 3) must traverse every stored *block* in descending order; rows in `contract_events` only exist for blocks that contained a `ChannelHub` event. A reorg that diverged entirely within a gap — blocks with no relevant events — is invisible to a tx-receipt-based walk.

Note that `eth_getBlockByHash(storedHash)` alone is **not** suitable as the canonicality check: a node may still have the orphan side-chain header cached locally and return it successfully, so a non-null response does not prove the block is in the canonical chain. The check must use `eth_getBlockByNumber` so the response is by definition the current canonical block at that height.

`block_hash` is a single `CHAR(66)` column. Its addition enables exact, O(1)-per-step canonicality checks and is the only approach that handles all reorg scenarios correctly.

#### Definition: latest processed block

The **latest processed block** for a chain is the highest block number at which the reactor successfully committed at least one event to the database — identical to the listener's existing startup cursor (`MAX(block_number)` in `contract_events` for this `blockchain_id` and contract address, computed by `GetLatestContractEventBlockNumber`). This is distinct from the highest block the listener ever *saw*: the listener may have seen many blocks that contained no relevant events and therefore left no `contract_events` rows.

#### Reconciliation steps

On startup, for each chain, after the `block_hash` migration has been applied:

1. Query `contract_events` for the latest committed event: `latestBlockNum = MAX(block_number)`, `latestBlockHash = block_hash` at that row. If no rows exist, start the scan from the chain's configured genesis / start block and skip to step 5.
2. Call `eth_getBlockByNumber(latestBlockNum)` on the chain's RPC and compare the returned block's hash against `latestBlockHash`.
   - **Hash matches** → the stored block is the current canonical block at that height; no reorg above it. Proceed to step 4.
   - **Hash differs** → a different block now occupies that height; the stored block has been reorged out. Proceed to step 3.
   - **`ethereum.NotFound`** (RPC has no canonical block at that number, e.g. the height was pruned) → treat as reorged-out and proceed to step 3 rather than failing startup.
3. **Common-ancestor walk using stored block hashes:** query `contract_events` for the next-older distinct `block_hash` (the highest `block_number` strictly below the current candidate). Repeat step 2 with this (number, hash) pair. Continue until a stored block is confirmed canonical, or until no older stored hash exists. This height is the **common ancestor**.

   > **Why walk stored hashes, not block numbers?** In normal operation most blocks contain no `ChannelHub` events, so `contract_events` has no row for them. A block-number walk would find nothing to compare at event-gap heights and could miss a reorg that occurred entirely within such a gap. Walking by stored block hashes ensures every comparison is against a block the reactor actually processed.

   If the walk exhausts stored rows without finding a canonical one **and** no older row exists (`prevNum == 0` with `prevHash == ""`), the listener resumes from the *original* latest stored block number. The orphaned hash is discarded; `eth_getLogs` is a canonical-chain range query, so canonical-replacement logs between that height and the current tip are re-fetched normally. The empty-store case (`latestNum == 0`) continues to skip historical replay and tracks the chain from the live subscription.

4. Set the scan start to `commonAncestorBlockNum`. Events between `commonAncestorBlockNum` and `latestBlockNum` that came from the reorged fork are still present in the DB. The reactor has no rollback mechanism for those rows — the re-scan below will re-apply canonical events over them where the transaction was re-mined (idempotent), and leave the orphaned DB state in place where the transaction was not re-mined (residual risk; see §2.1). State-setting operations (`UpdateChannel`, `RefreshUserEnforcedBalance`) will overwrite with canonical values for re-mined events; rows from dropped transactions remain as stale data with no automated cleanup.
5. Start the event scan from `commonAncestorBlockNum` (or genesis if step 1 found no rows). Replayed events are routed **per-event by block age**:
   - Events whose block timestamp is **older than `confirmation_delay_secs`** are routed directly to the reactor, bypassing the gate. Their block is past the reorg window — `eth_getLogs` returned them as canonical, and any reorg that could displace them would exceed the configured finality bound. There is no incremental reorg risk to guard against, and routing them through the gate would only add latency.
   - Events whose block timestamp is **younger than `confirmation_delay_secs`** are routed through the gate, the same path live events take. The common-ancestor walk only confirms the *starting* block is canonical; replay can fetch logs from blocks all the way up to the current chain tip, some of which are still inside the reorg window. Forwarding those directly to the reactor would re-introduce the very double-spend window the gate was built to close.

   The `Listener` accepts two handlers (`eventHandler` for live events and recent historical events, `historicalEventHandler` for mature historical events) and makes the per-event routing decision from `eventLog.BlockTimestamp`. To guarantee that field is populated regardless of the RPC provider's behavior, the listener calls `ensureBlockTimestamp` once per event, which uses `eventLog.BlockTimestamp` when present and falls back to `HeaderByHash` otherwise (at most one fetch per block regardless of event count).
   When `confirmation_delay_secs` is `0` the gate is disabled and every historical event is routed to `historicalEventHandler`. On an `ensureBlockTimestamp` failure the Listener falls back to `eventHandler` (the gate) — the conservative choice that preserves the reorg-protection invariant at the cost of a small delay.
6. The reactor is idempotent for replayed events: `HandleHomeChannelCreated` has an explicit early-return guard when the channel is already open; `HandleHomeChannelCheckpointed` and `RefreshUserEnforcedBalance` use set-semantics (not accumulation) and recompute from the latest DB state. Before opening a transaction, `HandleEvent` calls `IsContractEventProcessed`; if the event is already committed, it returns `nil` immediately with no DB transaction opened. If `IsContractEventProcessed` returns an error, `HandleEvent` returns the wrapped error; the listener unsubscribes and the process restarts (per the lifecycle closure in §6.8), re-fetching the same range via the DB cursor so the pre-check retries. For events that pass the pre-check, `StoreContractEvent` is called last inside the DB transaction and enforces a unique constraint on `(transaction_hash, log_index, blockchain_id)` as a final backstop.
7. Historical log queries (`eth_getLogs`) return only canonical chain events — there are no `Removed: true` signals during replay, and replay does not flow through the gate (step 5). Removal signals from the live WebSocket subscription that arrive during the replay phase are buffered in the listener's `currentCh` and reach the gate only after the historical replay phase completes; if they cancel a re-mined event that has already been forwarded by the live path, the post-gate reorg detection in §6.5 logs them.
8. When `confirmation_delay_secs == 0`, the listener drops `Removed:true` live logs at the Phase 2 boundary because there is no downstream gate to consume them; the reactor never receives `Removed:true` logs in either mode.

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
var liveHandler evm.HandleEvent
if confirmationDelay > 0 {
    gate, err := evm.NewConfirmationGate(confirmationDelay, b.ID, reactor.HandleEvent, logger)
    if err != nil { /* fatal */ }
    gate.Start(ctx)
    liveHandler = gate.HandleEvent
} else {
    liveHandler = reactor.HandleEvent
}
l := evm.NewListener(..., liveHandler, reactor.HandleEvent, ...)
```

The constructor returns an error for `delay <= 0`; the wiring layer is responsible for skipping gate construction when the operator configured `confirmation_delay_secs: 0` and routing live events straight to the reactor.

The reactor itself does not change. All the listener's existing logic — subscription management, cursor tracking, reconnection, historical replay — is unaffected.

**Handling `Removed: true` logs:** currently `listener.go:289-294` skips removed logs before they reach the handler. This skip must be moved: the listener should forward removed logs to `gate.HandleEvent` (they still carry the `Removed` flag on `types.Log`), and the gate alone decides whether to cancel a pending timer or ignore the signal. The reactor never sees a `Removed: true` log.

### 6.2 Event identity for queue keying

The Listener delivers events in strict block order, so the FIFO queue is naturally ordered by arrival time. Two distinct keys identify events at different layers of the design:

- **`(txHash, logIndex)` — the live-entry key, used as the tombstone-map (`pending`) key.** On a non-removed arrival, the Pusher sets `pending[ek] = eventLog.BlockHash` (overwriting any prior value) and appends to the queue tail. On a `Removed: true` arrival, the Pusher checks `pending[ek]` and cancels (deletes from `pending`) if the stored `blockHash` matches the removed log's. A stale removal for an OLD block whose `pending` value has already been overwritten by a newer re-add will not match and falls through to the `forwardedSet` lookup (§6.5). Both operations are O(1) map lookups; the queue body is never scanned.
- **`(txHash, blockHash, logIndex)` — the post-gate detection key (`forwardedKey`), used to index `forwardedSet`.** When the drain goroutine forwards an event, it inserts this triple into `forwardedSet` so a later `Removed: true` for the same exact occurrence can be matched and the post-gate reorg WARN emitted. Including `blockHash` ensures a stale removal for an already-replaced fork cannot cause a spurious WARN against a different re-mining.

`blockHash` is excluded from the live-entry key so that a re-mining of the same tx overwrites the original `pending` value regardless of which block it landed in. `blockHash` is included in the post-gate detection key so that the WARN matches the specific occurrence that was forwarded.

A single transaction can emit multiple events for the same `txHash` (e.g., two `ChannelDeposited` logs in a batch open). `logIndex` disambiguates these; it is unique per log within a block and is present in both the live event and its corresponding `Removed: true` log.

`blockHash` is also used by:

- The post-gate reorg detection map (`forwardedSet`, §6.5) — keyed by `(txHash, blockHash, logIndex)` to identify which specific occurrence was forwarded, with the FIFO `forwardedQueue` driving O(1) eviction.
- `StoreContractEvent` in the reactor — stored in `contract_events` for the reconciliation walk (§4.4).

### 6.3 Timer-and-kick design

**Data structure:** a FIFO queue of `(types.Log, arrivedAt time.Time)` paired with a `pending` tombstone map that is the source of truth for which queue entries are live. The queue is append-tail and pop-head only; stale entries are skipped at the head by comparing `pending[ek]` to the popped entry's `BlockHash`. Removal scans of the queue body are eliminated.

```go
type queueEntry struct {
    log       types.Log
    arrivedAt time.Time
}

type eventKey struct {          // used as the tombstone-map key (re-add replaces prior entry)
    txHash   common.Hash
    logIndex uint
}

type forwardedKey struct {      // post-gate detection key (full triple, written by drain goroutine, read on Removed)
    txHash    common.Hash
    blockHash common.Hash
    logIndex  uint
}

type forwardedExpiry struct {
    key         forwardedKey
    forwardedAt time.Time
}

type ConfirmationGate struct {
    delay   time.Duration
    chainID uint64
    handler HandleEvent
    logger  log.Logger

    mu             sync.Mutex
    queue          []queueEntry                       // protected by mu
    pending        map[eventKey]common.Hash           // live (txHash, logIndex) -> blockHash; protected by mu
    forwardedSet   map[forwardedKey]time.Time         // protected by mu; entries are kept for a small multiple of `delay` (see §6.5)
    forwardedQueue []forwardedExpiry                  // FIFO of (key, forwardedAt) driving O(1) eviction; protected by mu

    kick  chan struct{}  // buffered 1, non-blocking sender
    timer *time.Timer    // created in Start(ctx); reset to the head entry's deadline
}
```

---

**Pusher path** (driven by the existing Listener; implements the `HandleEvent` signature)

Receives `types.Log` from the Listener. On each event:

- If `Removed: true` — under `mu`: if `pending[ek] == eventLog.BlockHash`, `delete(pending, ek)` (pre-gate cancel; the tombstoned queue entry is silently skipped when it reaches the head). Otherwise, if `forwardedSet[fk]` is set, emit the post-gate WARN (§6.5) and `delete(forwardedSet, fk)`; leave the corresponding `forwardedQueue` entry in place — it expires on its own and the eviction loop's value-check makes the early delete safe. Otherwise, emit a DEBUG "removal for unknown/stale event".
- Otherwise — under `mu`: set `pending[ek] = eventLog.BlockHash` (replacing any prior value for the same `(txHash, logIndex)`) and append `(log, arrivedAt)` to the queue tail. `arrivedAt` is the block timestamp (see §6.7). Release `mu` and send a non-blocking `kick` (`select { case g.kick <- struct{}{}: default: }`).

No expiration check, no forwarding. Push only.

---

**Drain goroutine** (single, started by `Start(ctx)`)

A single timer drives forwarding; no idle wakeups. The timer is reset to the head entry's deadline; a 1-buffered `kick` channel coalesces wakeups from the Pusher when a new head deadline is sooner than the currently-armed timer (or when the queue was empty).

```go
for {
    select {
    case <-ctx.Done():
        return
    case <-g.kick:
    case <-g.timer.C:
    }
    g.drainAndReschedule()
}
```

`drainAndReschedule`:

1. Under `mu`: `now := time.Now()`. While the head entry is mature (`queue[0].arrivedAt + delay <= now`):
   - Pop it.
   - **Tombstone check:** if `pending[ek] != entry.log.BlockHash`, the live entry for that `(txHash, logIndex)` has been replaced by a re-add. Drop silently. Do **not** touch `pending[ek]` — it refers to the *new* live entry still in the queue.
   - Otherwise: `delete(pending, ek)`; `forwardedSet[fk] = now`; `forwardedQueue = append(forwardedQueue, forwardedExpiry{fk, now})`. **These three writes happen before releasing `mu`** around the handler call, so a fast `Removed: true` arriving immediately after forwarding always sees the entry and emits the post-gate WARN.
   - Release `mu`, call `handler`, re-acquire.
2. Evict aged-out `forwardedSet` entries (see §6.5).
3. Reset the timer to the new head's deadline, or leave it stopped if the queue is empty (the next `kick` will recompute).

No event handling, no Listener awareness. Drain-and-forward only.

---

**Properties**

| Property | Detail |
| --- | --- |
| Chain-agnostic | `confirmationDelay` is the only chain-specific input |
| Forward latency after window | Bounded by timer scheduling jitter; no fixed polling tick |
| Idle cost | None — no ticker; the goroutine blocks on `ctx.Done()`/`kick`/`timer.C` |
| Reorg within window | Pusher's tombstone delete cancels the entry; Reactor never sees the event |
| Reorg deeper than window | Rare; Reactor-level idempotency (§6.6) handles re-delivered events |
| Concurrency | Pusher and drain goroutine share `mu`; Reactor is called outside the lock |
| Shutdown | Drain goroutine exits on `ctx.Done()`; `defer g.timer.Stop()` cleans up the timer; entries still in queue are discarded (safe — they were never forwarded). `kick` is **not** closed — the Pusher may still be invoked by an in-flight listener event during shutdown, and the non-blocking send is safe whether the receiver is alive or gone. |

### 6.4 Exposing `confirmation_delay_secs` via API

Clients need to know the confirmation delay for each chain so they can display the correct waiting time to users after submitting a deposit. The best existing candidate is **`node.v1.GetConfig`**, which already returns a per-chain `BlockchainInfoV1` object.

Files to update:

- `pkg/rpc/types.go` — add `ConfirmationDelaySecs uint64` to `BlockchainInfoV1`.
- `nitronode/api/node_v1/utils.go` — populate the new field in `mapBlockchainV1` from the chain's loaded config.
- `pkg/core/types.go` (or wherever `core.Blockchain` is defined) — add `ConfirmationDelaySecs uint64` so the value flows from `blockchains.yaml` through config loading into the API handler.

No new endpoint is needed. The field appears alongside existing per-chain fields (contract addresses, asset list, block time) and is read-only from the client's perspective.

### 6.5 Post-gate reorg detection in the gate

The `forwardedSet` membership map (paired with the `forwardedQueue` FIFO; both in the `ConfirmationGate` struct, §6.3) provides detection without any DB access. The **drain goroutine** writes to both each time it forwards an event; the **Pusher** reads `forwardedSet` when a `Removed: true` log arrives and finds no live entry in `pending`.

When `Removed: true` arrives in the Pusher:

- **`pending[ek] == eventLog.BlockHash`** → normal pre-gate removal; delete from `pending` and return. No log.
- **No pre-gate match, but `forwardedKey{txHash, blockHash, logIndex}` is in `forwardedSet`** → the event was already forwarded to the Reactor and its block has now been reorged out. Log at **`WARN`** with `txHash`, `blockHash`, `logIndex`, `chainID`. `delete(forwardedSet, fk)`. The corresponding `forwardedQueue` entry is left in place — it ages out on its own; the eviction loop's value-check (below) tolerates the early delete.
- **Match in neither** → log at `DEBUG` ("removal for unknown/stale event" — predates the current run or arrived after FIFO eviction).

`forwardedSet` entries are kept for a small multiple of `delay` — long enough that any `Removed: true` for a forwarded event arrives while the entry is still present, short enough that the map remains bounded. The exact multiplier is an implementation choice (current value: see `recentMultiplier` in `confirmation_gate.go`; e.g. 2 or 3 work in practice).

Eviction is performed in `drainAndReschedule` (the timer/kick goroutine), not in a separate sweep:

- Pop the front of `forwardedQueue` while `now − forwardedAt > recentMultiplier × delay`.
- For each popped `forwardedExpiry{key, forwardedAt}`, **delete from `forwardedSet` only if `forwardedSet[key] == forwardedAt`**. The value check guards the rare re-forward case (same key forwarded a second time after the chain un-reorgs back to the original block and a fresh delay elapses): the older FIFO entry must not evict the newer set membership. It also makes the §6.5 early delete (post-gate WARN path) a safe no-op when the eviction loop later visits its `forwardedQueue` sibling.

`forwardedAt` is the gate's wall-clock at forward time — not `BlockTimestamp` — so FIFO ordering is monotonic regardless of how `arrivedAt` was sourced. The map stays small because post-gate reorgs are rare and `Removed: true` arrives within one or two block-times of the reorg. No separate cleanup goroutine is required.

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

At the top of `HandleEvent`, before entering `useStoreInTx`, call this method. If the event is already committed, log at **`INFO`** ("skipping re-delivered event, already committed") and return `nil` immediately. No transaction is opened; no state is touched. If `IsContractEventProcessed` itself returns an error, `HandleEvent` returns the wrapped error immediately; the listener unsubscribes and the process restarts (per the lifecycle closure in §6.8). On restart, the DB cursor re-fetches the same range and the pre-check retries.

Reorged events that pass through this check are still neutralized by the reactor's **business-logic idempotency**:

- `HandleHomeChannelCreated` has an explicit early-return when the channel is already open.
- `HandleHomeChannelCheckpointed` and `RefreshUserEnforcedBalance` use set-semantics (overwrite, not accumulate).
- The `StoreContractEvent` unique constraint on `(transaction_hash, log_index, blockchain_id)` remains as the final backstop for the case where `log_index` happens to be unchanged.

The value of `IsContractEventProcessed` is therefore:

1. **Noise reduction for exact re-deliveries** — converts a constraint-violation rollback (logged as an error by the gate's drain goroutine) into a clean INFO exit with no DB transaction opened.
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

### 6.7 Source of `arrivedAt` and the listener's timestamp fallback

The gate uses the **block timestamp** of each event as its `arrivedAt` reference rather than wall-clock time. This ensures that events replayed from historical blocks (whose timestamps are minutes or hours in the past) are forwarded immediately on the first drain, without waiting for the full confirmation delay to elapse again.

#### Source of `arrivedAt`

The gate reads `eventLog.BlockTimestamp` directly from the `types.Log` it receives. It performs no RPC, holds no timestamp cache, and depends on nothing other than the in-memory value on the log struct. The listener guarantees `BlockTimestamp` is non-zero before forwarding a non-removed event to the gate. If the gate ever observes a zero value (defense-in-depth for tests and edge cases), it falls back to `time.Now()` for that single event; the listener owns any operational warning.

#### Reliability and fallback

`blockTimestamp` is part of the Ethereum execution JSON-RPC spec (execution-apis `receipt.yaml`, 2024) and is populated by current Geth (≥1.13.10), Erigon, Nethermind, Reth, Besu, recent `bnb-chain/bsc`, Bor, Arbitrum Nitro, and op-geth (Base, Optimism). It is **not** populated by Avalanche C-Chain (`ava-labs/libevm` does not define the field) and is unreliable on older `bsc-dataseed` nodes still in production rotation.

Therefore the **listener** — not the gate — owns the fallback. Before forwarding a non-removed event to the gate (or to the reactor on the historical bypass), the listener calls `ensureBlockTimestamp`, which uses `eventLog.BlockTimestamp` when present and falls back to one `HeaderByHash(blockHash)` RPC otherwise. A single-entry cache keyed on `lastBlockHash` elides repeat fetches for consecutive events from the same block, which — because the listener delivers events in block order — is the only relevant case. `Removed: true` logs skip `ensureBlockTimestamp` entirely; the gate's cancel path never reads `BlockTimestamp`.
On `HeaderByHash` failure the listener logs a WARN and forwards the event through the gate anyway, where the zero-defense fallback above degrades the entry to a wall-clock delay rather than dropping it silently.

---

### 6.8 Handler error semantics

When a downstream handler invoked after the confirmation delay returns an error, the gate's `run` goroutine returns the error and the gate's lifecycle closure (passed to `Start`) is invoked with it. In `nitronode/main.go`, that closure calls `logger.Fatal` → process exit. The supervisor restarts the process; the next `Listen` invocation re-fetches the unstored event via the DB cursor in `findCommonAncestor` + Phase 1 reconciliation, restoring the pre-PR crash-restart-replay invariant. The gate does **not** retry handler errors in-process; this is intentional and matches pre-PR behavior. Events queued behind the failed event are dropped on teardown and re-fetched after restart. The gate's lifecycle (`Start(ctx, handleClosure)`) is identical to `Listener.Listen` and `BlockchainWorker.Start`; the listener does not know that its downstream handler may fail asynchronously — error propagation is handled by the supervisor (`main.go`), where it already is for the other two components.
