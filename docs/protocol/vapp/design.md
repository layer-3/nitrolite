# Virtual App Sessions (vApp) - Design Specification

> Status: **Draft** — open design questions remain in [Section 6](#6-open-design-questions).

## 1. Motivation

App Sessions V1 model application state as a multi-party state-channel enforcement layer.
Each session carries an `AppDefinitionV1` (participants, quorum, nonce) and advances
through `AppStateUpdateV1` entries (version, intent, allocations, session data).
While functional, the design has hit hard limits that prevent real-world apps from
building on top of it:

**Deposits are coupled to state updates.** A single deposit requires a valid user
channel state (Commit transition), and the deposit amount must exactly match one
allocation diff inside the app state update. Only one deposit per update is possible.

**Withdrawals scale poorly.** Each withdrawal issues a Release receiver state per
participant-asset pair. A session with N participants and M assets may need up to
N*M channel state transitions on a single close, making the operation increasingly
expensive.

**Validation is rigid.** Every update validates allocation diffs against the declared
intent and checks quorum over a fixed participant list — nothing more. There is no
way to run custom logic, change the participant set mid-session, or use session data
for application-specific rules. Despite this simplicity, the validation code is
already complex.

Virtual App Sessions (vApp) address these limitations with two fundamental changes:

1. **Voucher-based fund flow** — decouple sending and receiving money through
   issuable/redeemable vouchers, replacing direct deposit and withdrawal intents.
2. **WASM validation modules** — delegate state transition validation to a
   developer-supplied WebAssembly module referenced in the app definition, enabling
   arbitrary application logic.

---

## 2. V1 Field Audit

### 2.1 Fields to Keep

| V1 Field | V2 Status | Rationale |
|---|---|---|
| `AppDefinitionV1.ApplicationID` | **Keep** | Identifies the end application on the platform. |
| `AppStateUpdateV1.AppSessionID` | **Keep** | Links every update to its session. |
| `AppStateUpdateV1.Version` | **Keep (evaluate)** | Critical for replay prevention and ordered history. May evolve — see [Section 6.1](#61-data-storage). |
| `AppStateUpdateV1.SessionData` | **Keep (redesign)** | Renamed to `Data`. Becomes the primary payload validated by the WASM module. Storage model must change — see [Section 6.1](#61-data-storage). |

### 2.2 Fields to Remove

| V1 Field | Rationale |
|---|---|
| `AppDefinitionV1.Nonce` | Removed — uniqueness is delegated to the application via `Metadata`. Apps provide unique metadata per session (e.g. game ID, match parameters). |
| `AppDefinitionV1.Participants` | Replaced by `Metadata` — participant management moves into the WASM module. |
| `AppDefinitionV1.Quorum` | Same as above; quorum logic becomes module-defined. |
| `AppStateUpdateV1.Intent` | Deposit and withdraw intents are replaced by vouchers; operate and close semantics move into the module. The node no longer needs to interpret intent. |
| `AppStateUpdateV1.Allocations` | The node no longer tracks per-participant asset distribution. It tracks only total locked funds per asset. |

### 2.3 Fields to Add or Redesign

| V2 Field | Location | Description |
|---|---|---|
| `AppDefinition.Metadata` | Definition | Opaque bytes set at creation time, analogous to smart contract constructor arguments. The WASM module interprets this freely (participant lists, config, access control rules, etc.). |
| `AppStateUpdate.DataHash` | Update (signed) | `bytes32` — hash of application data. The actual data travels as an unsigned sidecar, verified by the node against this hash. Keeps signed payload fixed-size. |
| `AppStateUpdate.VouchersUsed` | Update (signed) | `[]Voucher` — vouchers consumed by this update, moving funds **into** locked funds. Participant-authorized (signed). |
| `ValidationResult.VouchersIssued` | Result (module) | `[]Voucher` — vouchers emitted by the module, moving funds **out** of locked funds. Module-authorized (computed, not signed). |
| `ValidationResult.ModuleState` | Result (module) | `bytes` — module's persistent working memory, stored by node, passed back on next validation. |
| `ValidationResult.Events` | Result (module) | `[]Event` — structured logs emitted during validation. Indexed and queryable, not signed. |
| `AppStateUpdate.CrossRefs` | Update (signed) | `[]CrossRef` — references to specific versions of other sessions. Node resolves and verifies; module receives resolved data. Enables cross-session data access and the multi-session architecture pattern. |
| `AppSession.LockedFunds` | Node record | `[]AssetValue` — node-computed from voucher accounting: `prev + sum(used) - sum(issued)`. Not part of the signed update. |

---

## 3. Data Model

### 3.1 AppDefinition

```
AppDefinition {
    NodeAddress    address     // node hosting this session, replay protection across nodes
    ApplicationID  string      // registered application identifier
    Module         bytes32     // content-addressed reference to the WASM validation module
    Metadata       bytes       // constructor arguments, interpreted only by the module
}
```

`NodeAddress` serves two purposes: it binds the session to a specific node for
replay protection (a signed update for node X cannot be replayed on node Y), and
it enables efficient session lookup and indexing without relying solely on the hash.

Session uniqueness is delegated to the application via `Metadata`. The app is
expected to include whatever uniqueness data it needs (game ID, match nonce,
participant set, timestamp, etc.). This removes the protocol-level nonce and gives
apps full control over session identity.

`SessionID = keccak256(abi.encode(NodeAddress, ApplicationID, Module, Metadata))`

### 3.2 Voucher

```
Voucher {
    SourceID        bytes32     // session or channel that issues the voucher
    SourceVersion   uint64      // version of the source update that issued it
    DestinationID   bytes32     // session or channel that will consume the voucher
    Asset           string      // unified asset symbol
    Amount          decimal     // positive non-zero value
}
```

`VoucherID = keccak256(abi.encode(SourceID, SourceVersion, DestinationID, Asset))`

A voucher is uniquely identified by its `VoucherID`. The combination of
`SourceID + SourceVersion` ensures uniqueness per asset per destination — a given
source update can issue at most one voucher per `(DestinationID, Asset)` pair. This
removes the need for a separate nonce field.

The node maintains a **global set of active vouchers** indexed by `VoucherID`.
Sessions and channels reference vouchers by ID rather than embedding full content,
keeping cross-session coupling minimal and lookups efficient.

**Consumption** (funds enter the destination — participant-authorized):
- Participants include the voucher in `AppStateUpdate.VouchersUsed` (signed).
- The node verifies the `VoucherID` exists in the active set and is unconsumed.
- The node adds `amount` to the destination's `LockedFunds[asset]`.
- The voucher is removed from the active set (consumed exactly once).

**Issuance** (funds leave the source — module-authorized):
- The WASM module returns the voucher in `ValidationResult.VouchersIssued`.
- The node subtracts `amount` from the source's `LockedFunds[asset]`. Must remain >= 0.
- The voucher is added to the global active set.

This split means participants explicitly authorize incoming funds (signed), while
the module controls outgoing funds (computed). The module decides **what** to issue
based on the signed `Data` — participants express intent through data, the module
translates it into voucher operations.

Benefits:
- Multiple deposits (voucher uses) and withdrawals (voucher issues) per update.
- Fund flow is not tied to channel state transitions at update time.
- Multi-hop transfers work naturally — session A issues to B, B issues to C — each
  hop is a distinct voucher with a traceable `SourceID -> DestinationID` path.
- Participants don't need to construct voucher details — the module handles it.
- The node's balance invariant is simple: `LockedFunds` after applying vouchers must
  equal `LockedFunds_prev + sum(used) - sum(issued)`, with no negative balances.

### 3.3 AssetValue

```
AssetValue {
    Asset   string      // unified asset symbol
    Amount  decimal     // non-negative
}
```

### 3.4 AppStateUpdate

```
AppStateUpdate {
    AppSessionID    bytes32         // deterministic session identifier
    Version         uint64          // monotonically increasing, starts at 1
    VouchersUsed    []Voucher       // funds entering the session (participant-authorized)
    CrossRefs       []CrossRef      // references to other sessions (see Section 3.11)
    DataHash        bytes32         // keccak256 of Data
}
```

The signed payload is compact and fixed-structure. `LockedFunds` and `VouchersIssued`
are absent — `LockedFunds` is node-computed from voucher accounting, and
`VouchersIssued` is produced by the WASM module in its `ValidationResult`.
`CrossRefs` defaults to empty for sessions that don't reference other sessions.

### 3.5 SignedAppStateUpdate

```
SignedAppStateUpdate {
    AppStateUpdate  AppStateUpdate
    QuorumSigs      []bytes         // signatures over the packed update — see Section 6.2
    Data            bytes           // raw data sidecar, verified: keccak256(Data) == DataHash
}
```

`Data` is not covered by signatures directly — it is covered indirectly through
`DataHash`. The node verifies `keccak256(Data) == AppStateUpdate.DataHash` before
passing `Data` to the WASM module. This decouples signature payload size from data
size while maintaining integrity.

### 3.6 Event

```
Event {
    Topic   bytes4      // event type identifier (analogous to Solidity event selectors)
    Data    bytes       // event payload, module-defined encoding
}
```

Events are structured logs emitted by the module during validation. They are:
- **Append-only** — stored alongside the update, never modified.
- **Queryable** — the node indexes events by topic, session, and version.
- **Deterministic** — same inputs always produce the same events.
- **Not signed** — computed by the module, recomputable by re-running validation.
- **Capped** — max events per update and max total bytes per update, node-configured.

Use cases: debugging, off-chain indexing, webhook triggers, analytics, UI updates.

### 3.7 Signer

```
Signer {
    Address     address     // resolved wallet address
    Type        uint8       // 0xA1 = wallet (direct), 0xA2 = session key (delegated)
}
```

The node recovers signer identities from `QuorumSigs` and resolves session keys to
their authorizing wallet addresses. The module receives the final resolved identity
with the signer type, so it can distinguish direct signatures from delegated ones
when the security policy requires it.

### 3.8 ValidationContext

```
ValidationContext {
    Definition          AppDefinition           // immutable session definition
    PrevVersion         uint64                  // previous version (0 on first update)
    PrevData            bytes                   // resolved data from previous version (empty on first)
    PrevLockedFunds     []AssetValue            // previous locked funds (empty on first)
    PrevModuleState     bytes                   // module's working memory from last validation (empty on first)
    Update              AppStateUpdate          // the proposed update (contains DataHash, CrossRefs)
    Data                bytes                   // resolved data for this update (verified against DataHash)
    Signers             []Signer                // recovered and resolved signer identities
    ResolvedCrossRefs   []ResolvedCrossRef      // node-verified cross-session data (see Section 3.11)
}
```

The context is optimized for validation power, not signing efficiency. The module
receives fully resolved `Data` (not just hashes) for both previous and current
versions. `ResolvedCrossRefs` carries the actual application data from referenced
sessions, resolved and verified by the node. The signed `AppStateUpdate` is
included for reference but the module primarily works with the resolved fields.

### 3.9 ValidationResult

```
ValidationResult {
    Close           bool            // true = close session after this update
    VouchersIssued  []Voucher       // funds leaving the session (module-authorized)
    ModuleState     bytes           // working memory, persisted by node, passed back next time
    Events          []Event         // structured logs
}
```

**`Close`** — signals the node to close the session. The node enforces that
`LockedFunds` is all-zero after applying vouchers (see [Section 5.3](#53-session-close)).

**`VouchersIssued`** — the module controls fund outflows. Participants express
withdrawal intent through `Data`; the module translates it into concrete vouchers.
The node validates that issuance does not make `LockedFunds` negative.

**`ModuleState`** — opaque bytes the node persists and feeds back via
`ValidationContext.PrevModuleState` on the next validation. Enables accumulated
counters, cached computations, and internal bookkeeping without bloating the signed
update. Size-capped (e.g. 32KB). Deterministic — any node running the same module
with the same inputs produces the same state.

**`Events`** — structured logs for developer tooling (see [Section 3.6](#36-event)).

### 3.10 AppSession (Node-Side Record)

```
AppSession {
    SessionID       bytes32
    Definition      AppDefinition
    IsClosed        bool            // false on creation, set to true on close
    Version         uint64
    LockedFunds     []AssetValue    // node-computed: prev + sum(used) - sum(issued)
    DataHash        bytes32         // latest data hash
    ModuleState     bytes           // latest module working memory
    CreatedAt       timestamp
    UpdatedAt       timestamp
}
```

`LockedFunds` is never set by participants — the node computes it from voucher
accounting after each update. `Data` is stored content-addressed by `DataHash` and
retrievable for historical access (host functions). `ModuleState` is the latest
output from the module's `ValidationResult`.

### 3.11 CrossRef and ResolvedCrossRef

```
CrossRef {
    SessionID       bytes32         // referenced session
    Version         uint64          // specific version pinned
}
```

Included in the signed `AppStateUpdate.CrossRefs`. The signer commits to depending
on a specific version of another session. This is deterministic — every node that
has accepted the referenced session at that version holds identical data.

```
ResolvedCrossRef {
    SessionID       bytes32         // referenced session
    Version         uint64          // version that was referenced
    Data            bytes           // application data from that session at that version
}
```

The node resolves each `CrossRef` before calling the module: verifies the
referenced session exists, has the specified version, and retrieves the stored
application data for that version. The module receives `ResolvedCrossRef` with the
actual data — no hash verification needed, the node already did it.

`ResolvedCrossRef.Data` is the application data (the `Data` sidecar from the
referenced session's update at that version). It does not include ModuleState or
other internal state — Data is the public interface between sessions.

---

## 4. WASM Validation Module

### 4.1 Role

The WASM module is the sole authority for application-level validation. The node
handles only structural and financial invariants (version ordering, locked fund
accounting, voucher balancing). Everything else — participant authorization, game
rules, access control, data schema — is the module's responsibility.

### 4.2 Module Lifecycle

1. **Registration** — modules are content-addressed (e.g. `keccak256(wasm_bytecode)`)
   and uploaded to the node or an external registry before session creation.
2. **Binding** — `AppDefinition.Module` references the module hash. Once a session is
   created, the module binding is immutable.
3. **Invocation** — on every `AppStateUpdate`, the node invokes the module with the
   update context. If the module returns an error, the update is rejected.
4. **Determinism** — modules must be pure functions of their inputs. No I/O, no
   randomness, no host callbacks. The module receives all data it needs through
   `ValidationContext` and returns all outputs through `ValidationResult`.

### 4.3 Module Interface

The module exports a single validation function:

```
validate(context: ValidationContext) -> Result<ValidationResult, Error>
```

See [Section 3.8](#38-validationcontext) for `ValidationContext` and
[Section 3.9](#39-validationresult) for `ValidationResult` definitions.

The module can:
- Validate that `Signers` satisfy application-defined quorum/permission rules.
- Validate `Data` transitions (game moves, order book changes, etc.).
- Access verified data from other sessions via `ResolvedCrossRefs`.
- Issue vouchers to move funds out of the session.
- Enforce constraints on `VouchersUsed` (incoming funds).
- Persist working memory via `ModuleState` for use in subsequent validations.
- Emit structured `Events` for debugging, indexing, and developer tooling.
- Signal session close via `Close`.
- Reject updates that violate application invariants.

The module cannot:
- Modify the signed update (it produces results, not mutations).
- Access external state or perform I/O.
- Exceed resource limits (CPU, memory) set by the host — see [Section 6.3](#63-security-and-resource-limits).

### 4.4 Host-Module Boundary

The node enforces **before** calling the module:

| Check | Description |
|---|---|
| Version ordering | `update.Version == session.Version + 1` |
| Session status | `IsClosed == false` |
| Data integrity | `keccak256(Data) == update.DataHash` |
| Voucher consumption | Used vouchers exist in global active set with matching `DestinationID` |
| CrossRef resolution | Each referenced session exists and has the specified version; resolve data |
| Structural integrity | Required fields present, amounts positive |

The module enforces **application rules** and produces results:

| Responsibility | Description |
|---|---|
| Signer authorization | Which addresses are allowed to propose this transition |
| Data validity | Application-specific payload validation |
| Voucher issuance | Produces `VouchersIssued` — which funds leave, to whom |
| Voucher policy | Whether the incoming `VouchersUsed` are acceptable |
| Close signal | When the session should be closed |
| Module state | Computed working memory for next validation |
| Events | Structured logs for external consumption |

The node enforces **after** the module returns:

| Check | Description |
|---|---|
| Voucher issuance validity | Issued `VoucherIDs` are globally unique; `SourceVersion` matches update version |
| Locked fund accounting | `new_locked = prev_locked + sum(used) - sum(issued)`, all assets >= 0 |
| Close invariant | If `Close == true`, all `LockedFunds` must be zero |
| Size limits | `ModuleState` and `Events` within configured caps |

---

## 5. State Update Flow

### 5.1 Standard Update

```
Participant(s) --> SignedAppStateUpdate { AppStateUpdate, QuorumSigs, Data } --> Node

Node (pre-module):
  1. Structural validation (fields, types, session exists, IsClosed == false)
  2. Version check (version == current + 1)
  3. Data integrity (keccak256(Data) == update.DataHash)
  4. CrossRef resolution: for each CrossRef, verify session exists at that
     version, resolve data for that version
  5. Voucher consumption: for each VoucherUsed, verify exists in active set
  6. Signature recovery (recover addresses from QuorumSigs)

Node (module):
  7. Build ValidationContext (prev state, resolved Data, signers,
     prev ModuleState, ResolvedCrossRefs)
  7. Invoke WASM module: result = validate(context)
  8. If module returns Err: reject update

Node (post-module):
  9. Validate VouchersIssued from result (unique IDs, SourceVersion matches)
  10. Compute LockedFunds: prev + sum(used amounts) - sum(issued amounts)
  11. Verify all LockedFunds assets >= 0
  12. Verify ModuleState and Events within size limits
  13. If Close == true: verify all LockedFunds are zero
  14. Persist: update, Data (by hash), ModuleState, Events, active vouchers
  15. Advance session: version, LockedFunds, DataHash, ModuleState
  16. If Close == true: set IsClosed = true
```

### 5.2 Session Creation

Session creation is a standard update with the `AppDefinition` supplied alongside
the first `SignedAppStateUpdate` (version 1). This ensures the session starts with
a validated initial state rather than an empty placeholder.

```
Creator --> CreateAppSession(definition, signed_update) --> Node

Node:
  1. Validate ApplicationID is registered
  2. Validate Module hash references a known module
  3. Compute SessionID = keccak256(abi.encode(definition))
  4. Verify no session with this ID exists
  5. Verify signed_update.Version == 1 and signed_update.AppSessionID == SessionID
  6. Run standard update flow (steps 3-17) with:
     PrevVersion=0, PrevData=empty, PrevLockedFunds=empty, PrevModuleState=empty
  7. If successful: create session record (IsClosed=false)
  8. If module returns Err: reject creation
```

This unifies creation and update into one code path. The module's validation of
version 1 acts as constructor validation — it can inspect `Definition.Metadata`
and the initial `Data` to accept or reject the session parameters.

### 5.3 Session Close

Closing is an application-level concern. The WASM module defines close conditions
through its validation logic. When the module returns `ValidationResult.Close = true`:

1. The module's `VouchersIssued` should drain all remaining `LockedFunds`.
2. The node computes final `LockedFunds` (applying both `VouchersUsed` and
   `VouchersIssued`) and verifies all assets are zero.
3. The node persists the update and sets `IsClosed = true`.
4. No further updates are accepted for this session.

This keeps close logic in the module (it decides **when** to close and **where**
remaining funds go via issued vouchers) while the node enforces the financial
invariant (no funds left behind on close).

---

## 6. Open Design Questions

### 6.1 Data Storage

**Resolved:** The signed `AppStateUpdate` carries only `DataHash: bytes32`. Raw data
is an unsigned sidecar in `SignedAppStateUpdate.Data`. The node verifies integrity
via `keccak256(Data) == DataHash`.

**Protocol level:** `Data` is opaque bytes. The protocol imposes no internal structure.
The module defines its own encoding (CBOR, protobuf, custom).

**Node-level storage:** Data is stored content-addressed by `DataHash`. If two
versions produce identical data, it is stored once.

**Remaining tradeoffs:**
- **Granular dedup:** If an app's data is large and only partially changes between
  versions, the whole blob gets a new hash. A future optimization could introduce
  slot-based storage at the node level (split data into keyed chunks, dedup per
  chunk) — but this is an implementation detail, not a protocol concern. The WASM SDK
  can provide slot helpers that developers use within their opaque bytes.
- **Maximum data size:** Hard cap per update (e.g. 64KB default, node-configurable).
- **Garbage collection:** Data blobs referenced by no active session version can be
  pruned. Retention policy is node-configurable.
- **ModuleState storage:** Same content-addressed approach. Capped at 32KB.

### 6.2 Signature Recovery and WASM Module Interface

**Resolved:** The node recovers signer identities and passes `[]Signer` (address +
type) to the module (see [Section 3.7](#37-signer)). The signed message is always
the canonical ABI-encoded `AppStateUpdate`.

**Rationale:** Every module that needs access control would otherwise re-implement
the same ecrecover logic. Keeping recovery in the node eliminates that boilerplate,
reduces module code size and execution cost, and ensures consistent signature
handling across all modules.

**Session key support:** The node handles delegation transparently:
1. Recover the session key address from the signature.
2. Look up the session key authorization (as in V1's `AppSessionKeyStateV1`).
3. Resolve to the authorizing wallet address.
4. Pass the wallet address in `Signer.Address` with `Signer.Type = 0xA2`.

The module doesn't need to understand session key delegation — it receives resolved
wallet addresses. When the security policy requires it, the module can check
`Signer.Type` to distinguish direct signatures (`0xA1`) from delegated ones (`0xA2`).

**Non-ECDSA signature schemes** (BLS, threshold sigs, etc.) are out of scope for V2.
If needed, they would be a protocol-level extension — not something pushed into
individual modules.

### 6.3 Security and Resource Limits

WASM modules introduce untrusted code execution into the node's critical path.
Most security properties have clear answers; abuse response remains open.

#### 6.3.1 Compute and Memory Limits (Resolved)

The node uses wazero (pure Go, no CGo) with fuel metering and memory caps.
Limits are **per-module configurable at registration time** with node-enforced
maximums — a simple counter module needs less than a game engine.

| Limit | Default | Max (node-enforced) |
|---|---|---|
| Wall-clock timeout | 10ms | node-configurable |
| Fuel units | 1M | node-configurable |
| Linear memory | 4 MB | node-configurable |
| Data size per update | 64 KB | node-configurable |
| ModuleState size | 32 KB | node-configurable |
| Events per update | 64 count, 32 KB total | node-configurable |
| CrossRefs per update | 16 | node-configurable |
| Total resolved CrossRef data | 128 KB | node-configurable |
| Outstanding vouchers per session | 1000 | node-configurable |

Modules exceeding limits cause the update to be rejected (not the session to be
closed).

The node must reject non-deterministic WASM instructions (floats with NaN
canonicalization issues). Wazero handles this correctly by default.

#### 6.3.2 Data Growth (Resolved)

Without per-participant allocations, the node tracks only `LockedFunds` — a bounded
structure (one entry per asset). Growth vectors and mitigations:

- `Data`: capped per update (default 64KB), stored content-addressed by hash.
- `ModuleState`: capped (default 32KB), stored content-addressed.
- `VouchersIssued` (from module results): capped outstanding count per session.
- `Events`: capped per update (count and bytes).
- History: retained for auditability. Node-configurable retention depth.
  Historical data access via host functions (e.g. `host_get_data(version)`) is
  scoped out of the initial design but the storage model supports it — data is
  content-addressed and retrievable by version.

Consumed vouchers can be pruned (retain ID, discard payload).

#### 6.3.3 Replay and Ordering Guarantees (Resolved)

Version monotonicity (V1's approach) remains the primary replay prevention mechanism:
- `Version` must equal `current + 1` (no gaps, no repeats).
- The node rejects any update with version <= current.
- Combined with `SessionID` binding, this prevents cross-session replay.

For multi-node replication scenarios (future):
- The ordered version chain ensures all nodes converge on the same state.
- Each update's validity depends only on the previous state — no global ordering
  needed beyond the session scope.
- `LockedFunds` accounting is deterministic given the same voucher sequence.
- WASM module determinism ensures identical validation outcomes across nodes.

#### 6.3.4 Voucher Security (Resolved)

Vouchers must be tamper-proof and non-replayable:
- Each voucher has a deterministic `VoucherID = keccak256(abi.encode(SourceID,
  SourceVersion, DestinationID, Asset))`. The combination of source identity,
  version, destination, and asset guarantees global uniqueness.
- The node maintains a global set of active (issued but unconsumed) vouchers
  indexed by `VoucherID`.
- A voucher can be consumed exactly once — on consumption it is removed from the
  active set.
- Voucher issuance is produced by the WASM module (in `ValidationResult`), not
  directly signed by participants. However, issuance is anchored to a signed,
  versioned update — the module only runs on valid updates, and `SourceVersion`
  is enforced to match the update's version. A module cannot issue vouchers
  without a corresponding signed state transition.
- Cross-session voucher usage requires the destination to reference the `VoucherID`
  in its signed `VouchersUsed`. The node validates the voucher exists in the
  active set and that `DestinationID` matches the consuming session.
- The `SourceID -> DestinationID` path is traceable, enabling full audit of
  multi-hop fund flows across sessions and channels.
- The module is deterministic — any node processing the same signed update
  produces the same `VouchersIssued`. This ensures voucher state converges
  across replicas.

#### 6.3.5 Fund Integrity (Resolved)

The core invariant the node must always enforce:

```
For each asset A in session S:
  S.LockedFunds[A] >= 0
  S.LockedFunds[A] == initial_funds + sum(used_vouchers[A]) - sum(issued_vouchers[A])
```

No funds appear from nowhere; no funds go negative. This is checked **after** the
WASM module runs (since `VouchersIssued` comes from the module result), but the
node is the sole enforcer — even a malicious module cannot violate fund integrity
because the node rejects any result where `LockedFunds` goes negative.

#### 6.3.6 Abuse Response (Open)

What happens when a module repeatedly hits resource limits?

**Tier 1 — Per-update rejection** (automatic): Update fails, session continues.
Handles honest bugs and transient issues.

**Tier 2 — Rate limiting** (automatic): If a session hits N rejections in M seconds,
throttle submissions. Handles adversarial spam. Thresholds are node-configurable.

**Tier 3 — Session/module suspension** (manual): Node operator flags a module hash
as suspended. All sessions using that module stop accepting updates. Nuclear option
for discovered exploits.

Questions to resolve:
- Exact thresholds and backoff curves for tier 2.
- Whether tier 3 suspension should have any automatic triggers or remain purely
  operator-driven.
- How to communicate suspension status to clients (error codes, session metadata).
- Recovery path: can a suspended module be un-suspended, or must sessions migrate?

### 6.4 WASM Module Publishing (Open)

The lifecycle from developer's code to running module raises several questions:

**Storage:**
- Where are compiled WASM binaries stored? On the node directly, in a shared
  registry, or in external storage (IPFS, S3) with the node caching locally?
- Content addressing (`keccak256(wasm_bytecode)`) provides integrity verification,
  but who hosts the canonical copy?
- Version management: modules are immutable once published (content-addressed), but
  developers need a way to publish new versions and associate them with their
  application. Should there be an application-level "latest module" pointer, or is
  every module version fully independent?

**Execution environment:**
- Does the WASM runtime execute in the same process as the node, or in a dedicated
  sandboxed service?
- Same-process: lower latency, simpler deployment, but a crashing module could
  affect the node. Wazero's sandboxing mitigates this, but resource exhaustion
  (memory pressure) is still a concern.
- Dedicated service: stronger isolation, independent scaling, but adds network
  latency per validation call and operational complexity.
- Hybrid: same-process by default, with an option to offload to a sidecar for
  modules that need higher resource limits?

**SDK and developer experience:**
- What language(s) does the SDK target? Rust and Go are natural first choices
  (both compile to WASM). AssemblyScript (TypeScript-like) would lower the barrier
  for web developers.
- The SDK must provide: canonical encoding/decoding of `ValidationContext` and
  `ValidationResult`, helper types (`Signer`, `Voucher`, `AssetValue`, `Event`),
  and testing utilities (mock context builder, local validator runner).
- How do developers test modules locally before publishing? A CLI tool that runs
  the module against a sequence of mock updates would be essential.
- Documentation: example modules for common patterns (simple quorum, token swap,
  turn-based game) to bootstrap the ecosystem.

### 6.5 Upgradability and API Design (Open)

#### Structure Extensibility

The structures `AppStateUpdate`, `ValidationContext`, `ValidationResult`, and
`AppSession` will evolve. Adding fields must not break already-deployed modules.

Questions:
- **Encoding format:** ABI encoding (Solidity-style) is positional — adding fields
  breaks existing decoders. Alternative: length-prefixed or tagged encoding (protobuf,
  CBOR, custom TLV) where unknown fields are silently skipped. The SDK would
  abstract the encoding, but the wire format must be forwards-compatible.
- **Context versioning:** Should `ValidationContext` carry a version number so the
  module can detect which fields are available? Or should the SDK handle this
  transparently (new fields have zero-values if the node is older)?
- **Result extensibility:** If a future `ValidationResult` gains a new field, old
  modules that don't set it should still work. Default values must be safe
  (e.g. `Close` defaults to `false`, `VouchersIssued` defaults to empty).
- **Module compatibility matrix:** A module compiled against SDK v2 running on a node
  that only supports context v1 — how is this detected and handled? At registration
  time (reject incompatible modules) or at runtime (graceful degradation)?

#### Read API

Write endpoints are straightforward (create session, submit update). The read API
needs more thought — what do developers actually query?

**Session-level queries:**
- Get session by ID (definition, current state, LockedFunds, IsClosed).
- List sessions by application ID, by participant (requires module-level participant
  tracking — the node doesn't know participants natively in V2).
- Get session history (all updates for a session, paginated by version).

**Update-level queries:**
- Get update by session ID + version.
- Get Data for a specific version (resolved from content-addressed storage).
- Get ModuleState at a specific version.
- Get Events for a session, filterable by topic and version range.

**Voucher queries:**
- List active (unconsumed) vouchers by session ID, by destination ID.
- Get voucher by VoucherID.
- Voucher history: all vouchers ever issued/consumed by a session.

**CrossRef queries:**
- List sessions that reference a given session via CrossRefs (reverse lookup).
- Get updates that reference a specific session+version (which state updates
  consumed a given action).
- Action session polling: subscribe to new versions on a session (for operators
  watching player action sessions).

**Cross-session queries:**
- Fund flow graph: trace voucher paths across sessions.
- Application-level aggregation: total locked funds across all sessions for an app.

**Event subscriptions:**
- WebSocket/SSE stream of events by topic, session, or application.
- Essential for real-time UIs.

**The participant problem:** V1 exposes participants natively (they're in the
definition). V2 moves participants into module-interpreted `Metadata`. The node
can't natively answer "which sessions is address X participating in?" unless:
- The module emits standardized events on participant join/leave.
- The node indexes a well-known event topic for participant tracking.
- Or the API delegates this query to the module (but modules don't handle queries).

This is a significant UX regression from V1 if not addressed. Likely solutions:
- Define **standard event topics** (e.g. `0x00000001` for `ParticipantAdded`,
  `0x00000002` for `ParticipantRemoved`) that the node indexes natively. Modules
  that want participant discoverability emit these events. The SDK provides helpers.
- For the multi-session pattern, CrossRef relationships provide structural
  discoverability — the node can track which action sessions are referenced by
  a state session, and action sessions are per-participant.

---

## 7. Migration from V1

V1 and V2 app sessions will coexist. Key differences that affect migration:

| Aspect | V1 | V2 (vApp) |
|---|---|---|
| Fund model | Per-participant allocations | Node-computed locked funds + vouchers |
| Validation | Fixed (intent + quorum + allocation diffs) | WASM module |
| Participant management | In definition (immutable) | In metadata (module-interpreted) |
| Deposit | Commit transition + single allocation | Voucher used (participant-signed, decoupled from channel state) |
| Withdrawal | Release transitions per participant | Voucher issued (module-produced, batch-friendly) |
| Intents | operate, deposit, withdraw, close, rebalance | None (module-defined semantics) |
| Data | Opaque string, stored inline | DataHash in signed update, raw bytes as sidecar |
| Module output | N/A | VouchersIssued, ModuleState, Events, Close signal |
| Cross-session data | N/A (sessions fully isolated) | CrossRefs for verified data references between sessions |

Existing V1 sessions continue under V1 rules. New sessions choose V1 or V2 at
creation time (determined by whether `Module` is set in the definition).

---

## 8. Summary

Virtual App Sessions (vApp) replace the fixed validation and per-participant
allocation model of V1 with three primitives:

1. **Vouchers** — decouple fund movement from state updates. Participants authorize
   incoming funds (VouchersUsed, signed); the module controls outgoing funds
   (VouchersIssued, computed). The node enforces balance integrity.
2. **WASM modules** — move application logic out of the node. Modules validate
   transitions, issue vouchers, emit events, and maintain working memory
   (ModuleState). Developers define their own participant rules, data schemas,
   and transition logic.
3. **CrossRefs** — enable verified cross-session data access. An update can
   reference specific versions of other sessions; the node resolves and verifies
   the data, and the module receives it directly. This enables shared catalogs,
   oracle feeds, and the multi-session architecture pattern where per-participant
   action sessions feed into an operator-driven state session, eliminating version
   contention for concurrent apps.

The signed update is minimal: `AppSessionID`, `Version`, `VouchersUsed`,
`CrossRefs`, `DataHash`. Everything else is either a sidecar (Data), module output
(VouchersIssued, ModuleState, Events, Close), or node-computed (LockedFunds).

Four design questions remain open before full specification:
1. **Abuse response** — rate limiting thresholds, suspension triggers, recovery paths (Section 6.3.6).
2. **WASM module publishing** — storage, execution environment, SDK, developer tooling (Section 6.4).
3. **Upgradability** — forward-compatible encoding, context versioning, module compatibility (Section 6.5).
4. **Read API** — session/update/voucher/event queries, CrossRef reverse lookups, participant discoverability (Section 6.5).
