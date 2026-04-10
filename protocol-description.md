# Nitrolite Protocol — On-Chain and Off-Chain Architecture

## High-level goal

Nitrolite is an **extended state-channel protocol** that enables:

* continuous off-chain transfers and application interactions,
* frequent on-chain settlement (deposit / withdrawal),
* cross-chain liquidity movement (bridging),
* without locking all funds until channel closure.

The protocol trades atomic cross-chain guarantees for **optimistic enforcement with challenge recovery**, relying on cryptographic authorization and game-theoretic incentives.

---

## Core abstraction: Cross-Chain Token Balance (CCTB) State

A **channel** between a **User** and a **Node** is represented by a monotonically increasing sequence of **Cross-Chain States**.

Each state:

* has a strictly increasing `version`,
* is signed by both User and Node,
* encodes the net result of:

  * on-chain operations (deposit / withdrawal / migration),
  * off-chain transfers,
  * off-chain application sessions,
  * escrow preparation and execution.

A state may refer to **multiple chains**, but at any time only **at most two per-chain sub-states** exist (home and non-home).

Each **per-chain sub-state** represents accounting on a specific chain and consists of:

* **absolute allocations**
  (`userAllocation`, `nodeAllocation`) that must be fully backed by collateral locked on that chain, and

* **cumulative net flows**
  (`userNetFlow`, `nodeNetFlow`) that encode the aggregate effect of deposits, withdrawals, off-chain transfers, and app-session lock/unlock events since channel creation.

The difference between successive states’ net flows determines how much value must be pulled from or pushed to each party during on-chain enforcement.

---

## Off-chain protocol (control plane)

### Participants

* **User** — owns funds and initiates actions.
* **Node (Broker)** — provides liquidity, routing, and coordination.

### Off-chain responsibilities

The off-chain protocol is responsible for:

1. **State construction**

   * The Node aggregates:

     * off-chain transfers,
     * app-session lock/unlock events,
     * pending on-chain actions.
   * These are netted into a new `State` by updating per-chain allocations and cumulative net flows.

2. **State authorization**

   * Both User and Node sign the full state:

     ```text
     (channelId, version, intent, homeLedger, nonHomeLedger)
     ```

   * A party **never signs two different states with the same version**.

3. **Liquidity enforcement (Node responsibility)**

   * The Node must ensure it has enough liquidity to back absolute allocations:

     * between normal operations,
     * except during explicitly allowed escrow or migration phases.
   * If liquidity drops below a threshold, the User may:

     * checkpoint the latest state on-chain,
     * withdraw,
     * or migrate the channel.

4. **Node liquidity monitoring**

   * The Node continuously monitors its on-chain vault balance (`_nodeBalances`) across all supported chains.
   * When the available vault balance on a given chain approaches or falls below the amount required to back pending off-chain allocations, the Node takes corrective action:

     * enforce states that release locked Node funds back to the vault (e.g. checkpoint or close idle channels),
     * rebalance liquidity across chains via cross-chain operations,
     * or fire alerts for operator intervention.
   * The Node must not co-sign states that would require locking more vault funds than are currently available, as such states would fail on-chain enforcement.

   > **NOTE:** Node liquidity monitoring is not yet enforced in the Clearnode implementation but will be introduced in the near future.

   If Node liquidity monitoring is absent or fails, **no user funds are at risk**. On-chain enforcement always relies on the latest mutually signed state, and the previous on-chain state remains valid and enforceable at all times. The practical consequence is operational: a co-signed state that requires Node vault funds may fail when submitted on-chain if the vault has been depleted in the interim. In such cases, the Node simply replenishes its vault and resubmits. Users retain full access to their funds through challenge and closure paths based on the last successfully enforced state.

5. **Flow control**

   * When a cross-chain escrow or migration is in progress:

     * the Node **stops issuing new states**,
     * until the process completes or is challenged.

   This is an **off-chain responsibility** enforced by the Node. The on-chain contract cannot enforce
   operation ordering because it has no visibility into pending operations on other chains. Instead,
   the contract is designed to handle concurrent cross-chain operations correctly — for example,
   escrow operations remain reachable even after a subsequent migration on the same chain. Such
   concurrent code paths should never be reached under correct Node behavior, but the on-chain
   contract handles them safely to guarantee fund recovery in all cases.

6. **Optimistic bridging**

   * Cross-chain actions are **not atomically verifiable** on-chain.
   * Correctness is ensured by:

     * signed states,
     * cumulative net-flow accounting,
     * timeouts,
     * challenge rights.

---

## Off-chain actions encoded in states

### Off-chain transfers

* When a User **sends** funds off-chain:

  * user allocation decreases,
  * node net flow increases.
* When a User **receives** funds off-chain:

  * user allocation increases,
  * node net flow decreases.

These changes are reflected only in cumulative net flows until enforced on-chain.

---

### Off-chain application sessions

* App sessions are off-chain sub-channels governed by an external server.
* Funds may be:

  * **locked** into a session (flow to Node),
  * **unlocked** from a session (flow to User).
* Only signatures are required for persistence.
* Session effects are netted into cumulative net flows of the next enforceable state.

---

## On-chain protocol (enforcement plane)

The on-chain contract is the **final arbiter** of correctness.

It does not reconstruct intent — it **verifies and enforces signed states** by:

* validating signatures and monotonic versioning,
* applying the delta between the last enforced state and the submitted state,
* pulling or pushing funds according to net-flow differences,
* updating locked collateral to match absolute allocations.

---

## Channel lifecycle (on-chain)

### 1. Channel creation

* A channel is created with an initial signed state:

  * version = 0,
  * intent = CREATE,
  * funds pulled from the User (home chain).
* Channel enters `OPERATING`.

---

### 2. Normal operation (OPERATING)

While operating:

* Any **newer signed state** may be enforced on-chain.
* Enforcement may:

  * pull funds from User,
  * push funds to User,
  * lock or unlock Node liquidity.
* Enforcement may occur for:

  * deposit,
  * withdrawal,
  * checkpoint,
  * escrow execution,
  * migration execution.

Off-chain activity can continue indefinitely between enforcements.

---

### 3. Deposit (single-chain)

* User signs a state with intent = DEPOSIT.
* User net flow becomes positive.
* On enforcement:

  * funds are pulled from User,
  * locked into the channel.

---

### 4. Withdrawal (single-chain)

* User signs a state with intent = WITHDRAW.
* User net flow becomes negative.
* On enforcement:

  * funds are pushed to User,
  * channel locked funds decrease.

---

### 5. Checkpoint

* A state with intent = OPERATE.
* User net flow delta must be zero.
* Used to:

  * acknowledge off-chain transfers,
  * clear challenges,
  * synchronize cumulative net-flow accounting.

---

## Challenge mechanism (optimistic safety)

### Purpose

Challenges protect against:

* submission of outdated states,
* malicious or crashed counterparties,
* incomplete cross-chain operations.

---

### Challenge rules

* Only channels in `OPERATING` or `MIGRATING_IN` can be challenged.
* A challenge references a signed state.
* If the challenged state is **older than the latest signed state**:

  * the newest valid signed state **must be enforced first**, regardless of its intent.
* The following intents **cannot** be submitted via `challengeChannel`:

  * `CLOSE` — channel closure is a terminal operation; enforcing it leaves no live channel to dispute. Parties holding a valid CLOSE state should call `closeChannel` directly instead.
  * `FINALIZE_MIGRATION` on the **old home chain** (channel status `OPERATING`/`DISPUTED`) — this would release the node's funds and move the channel to `MIGRATED_OUT`, which is incompatible with entering `DISPUTED` state.

Invariant:

> Dispute resolution always requires processing the newest valid signed state, even if that state represents escrow execution or migration rather than deposit or withdrawal.

---

### Resolving a challenge

* Any party may submit a **strictly newer signed state**.
* If valid:

  * it is enforced,
  * net-flow deltas are applied,
  * the challenge is cleared,
  * channel returns to `OPERATING`.

---

### Challenge timeout

* If no newer state is submitted before expiry:

  * channel may be closed unilaterally,
  * allocations are paid out according to the last enforced state.

---

## Channel closure

A channel can be closed:

1. **Cooperatively**

   * via a signed CLOSE state.

2. **Unilaterally**

   * after a challenge expires.

Closure:

* Both parties encode their final payouts as negative net flow deltas (equivalent to a full withdrawal), so the CLOSE state must have `userAllocation == 0` and `nodeAllocation == 0`.
* The net flow deltas push all remaining funds to User and Node.
* Sets channel status to CLOSED.

---

## Cross-chain operations (bridging)

Cross-chain actions are **two-phase** and **optimistic**.

### Why two-phase?

Because:

* one chain cannot directly observe or verify another chain’s state,
* atomic enforcement is impossible without foreign-chain verification.

The protocol deliberately does **not** rely on light clients (on-chain verification of foreign headers, proofs, and validator signatures), as they are complex, expensive, and chain-specific.

The two phases are:

1. **Preparation phase**

   * liquidity is locked on chains where needed,
   * an escrow object (possible with timeouts) is created,
   * Node stops issuing new states.

2. **Execution phase**

    * an execution state that updates allocations and net flows is issued and signed
    * this state may be enforced immediately or later, but is enforceable to resolve disputes.

---

## Escrow deposit (bridging in)

### Preparation phase

* User locks funds on the **non-home chain**.
* Node locks equal liquidity on the **home chain**.
* An escrow object with timeouts is created.

---

### Execution phase

* A signed execution state updates allocations and net flows:

  * User’s non-home allocation decreases,
  * Node’s home allocation decreases,
  * corresponding net flows encode the swap.

This execution state **may be enforced immediately or later**, but must be enforceable to resolve disputes.

---

## Escrow withdrawal (bridging out)

### Preparation phase

* Node locks withdrawal liquidity on the **non-home chain**.

---

### Execution phase

* Signed state updates allocations and net flows so that:

  * User receives funds on the non-home chain.

If enforcement stalls:

* challenges and timeouts guarantee completion or reversion.

---

## Escrow Challenge Resolution

If an escrow process is challenged (status becomes `DISPUTED`) and the challenge period expires (`challengeExpireAt` passed) without a resolution:

* The `finalize` function handles this case explicitly.
* If called when `DISPUTED` and expired:
    1. Do **not** invoke the channel engine.
    2. Manually **unlock the locked funds to the Node**.
    3. Zero out `lockedFunds` and `challengeExpireAt`.
    4. Set status to `FINALIZED`.
    5. Emit a finalization event.

This logic mirrors the channel closure mechanism: if a challenge is not substantiated by a newer state within the timeout, the system defaults to a finalized state that releases locked resources.

---

## Escrow deposit purge queue

After an escrow deposit is resolved, the node's locked funds must be returned to the node vault. The contract maintains a FIFO queue of escrow deposit IDs (`_escrowDepositIds`), sorted by `unlockAt` ascending, and a monotonically advancing head pointer (`escrowHead`).

A purge pass processes queue entries in order:

* **FINALIZED** entries are skipped — funds were already credited during finalization.
* **DISPUTED** entries (challenge still active) are skipped — the challenge outcome is pending.
* **INITIALIZED** entries past their `unlockAt` timestamp are purged — the node's locked amount is returned to the node vault and the entry is marked FINALIZED.
* **INITIALIZED** entries not yet at `unlockAt` stop the scan — because the queue is sorted by `unlockAt` ascending, no subsequent entry can be purgeable either.

The scan is bounded by a `maxSteps` budget that counts every inspected entry, whether skipped, purged, or halting. `_purgeEscrowDeposits` is called automatically on every protocol operation with a budget of `MAX_DEPOSIT_ESCROW_STEPS = 64`. The public `purgeEscrowDeposits(maxSteps)` function allows any caller to drain a backlog explicitly.

---

## Home chain migration

Migration enables moving the channel's "home" security chain from one blockchain to another, preserving allocations and cumulative accounting.

Like other cross-chain operations, migration is **two-phase** and **optimistic**.

---

### Preparation phase (INITIATE_MIGRATION)

The preparation phase establishes the channel on the target (non-home) chain:

* A preparation state is constructed with:
  * intent = INITIATE_MIGRATION,
  * non-home state where Node deposits liquidity equal to User's allocation on the home chain.

**On the non-home chain:**

* This state is submitted via `initiateMigration()`.
* Effect:
  * creates a channel on the non-home chain with status `MIGRATING_IN`,
  * locks Node's funds on the non-home chain.
* Implementation note: States are swapped before storing to maintain the invariant that `homeLedger` represents the current chain.
* A `MIGRATING_IN` channel **is treated as the home chain** for all subsequent on-chain operations: because the swap has already been applied to the stored state, `homeLedger` already describes the current chain and ChannelEngine processes operations (deposit, withdraw, checkpoint, escrow, close) using standard home chain logic. The only distinction from `OPERATING` is that `finalizeMigration()` has not yet been called to confirm the migration on the old home chain.

**On the home chain:**

* The preparation state **can be submitted** via `initiateMigration()`:
  * updates the channel's latest state,
  * keeps the channel in `OPERATING` status,
  * can clear a challenge (following standard challenge resolution flow).
* The preparation state **cannot be checkpointed** via `checkpoint()`:
  * `checkpoint()` explicitly rejects states with migration intents.
* The home-chain channel **can be challenged** with a preparation state:
  * this enables dispute resolution if something goes wrong,
  * a valid FINALIZE_MIGRATION execution state can move the channel from `DISPUTED` to `MIGRATED_OUT`,
  * otherwise, after `challengeExpireAt`, funds may be withdrawn according to standard challenge rules.

---

### Execution phase (FINALIZE_MIGRATION)

The execution phase completes the migration by swapping home and non-home roles:

* An execution state is constructed that:
  * swaps the `homeLedger` and `nonHomeLedger` from the preparation phase,
  * swaps allocations between User and Node in each state,
  * intent = FINALIZE_MIGRATION.

**On the old home chain:**

* This state is submitted via `finalizeMigration()`:
  * releases Node liquidity on the old home chain,
  * moves the channel to `MIGRATED_OUT` status,
  * can clear a challenge (moving from `DISPUTED` to `MIGRATED_OUT`).
* Implementation note: States are swapped before validation to maintain the invariant that `homeLedger` represents the current chain.

**On the new home chain** (old non-home chain):

* The execution state **may be submitted explicitly** via `finalizeMigration()`:
  * moves the channel from `MIGRATING_IN` to `OPERATING`.
* However, the **intended usage** is to combine the execution phase with a subsequent operation:
  * any on-chain call (deposit, withdrawal, checkpoint, escrow initiate/finalize, or close) can be applied **on top of** the execution phase state,
  * this implicitly completes the migration and transitions the channel to `OPERATING`.
* The new home chain can be challenged with the execution state (or any newer valid state), triggering normal challenge resolution.

---

### Migrating back

A channel on a chain with status `MIGRATED_OUT` can be migrated back:

* Submitting a new preparation phase state via `initiateMigration()` on that chain:
  * moves the channel from `MIGRATED_OUT` to `MIGRATING_IN`,
  * initiates a reverse migration flow.

This enables round-trip migration as needed.

---

### Implementation: State Representation and Delta Calculation

Migration presents a unique challenge for on-chain implementation: **which state represents the current chain changes during migration**.

#### The Problem

The protocol describes migration as swapping `homeLedger` and `nonHomeLedger` roles, but this creates semantic ambiguity for on-chain validation:

1. **Preparation phase on non-home chain**: Actions (node deposits liquidity) are encoded in `nonHomeLedger`, but after the channel is created, subsequent operations must calculate deltas from this state—even though validation logic assumes `homeLedger` represents the current chain.

2. **Execution phase on old home chain**: After the user swaps states in the execution phase state, `nonHomeLedger` represents the old home (current chain), but validation logic expects `homeLedger` to represent the current chain.

3. **Delta calculation inconsistency**: After `INITIATE_MIGRATION` on the non-home chain creates a `MIGRATING_IN` channel, the next operation (e.g., deposit) cannot correctly calculate deltas because the previous state's allocations are in `nonHomeLedger`, not `homeLedger`.

#### The Solution: Context-Based Validation + Selective State Swapping

To maintain the invariant that **homeLedger always represents the chain where execution happens**, the implementation uses:

**1. Two Migration Intents with Context-Based Behavior:**

* `INITIATE_MIGRATION`: Single intent used on both home and non-home chains
* `FINALIZE_MIGRATION`: Single intent used on both old home and new home chains

The same signed state can be submitted on both chains. The contract determines the correct behavior based on the channel status:

* INITIATE_MIGRATION + status VOID/MIGRATED_OUT → non-home chain behavior (create MIGRATING_IN)
* INITIATE_MIGRATION + status OPERATING/DISPUTED → home chain behavior (update state)
* FINALIZE_MIGRATION + status MIGRATING_IN → new home chain behavior (move to OPERATING)
* FINALIZE_MIGRATION + status OPERATING/DISPUTED → old home chain behavior (move to MIGRATED_OUT)

**2. Four ChannelHub Functions:**

* `initiateMigration()`: Called on non-home chain to create `MIGRATING_IN` channel
* `initiateMigration()`: Called on home chain to update state
* `finalizeMigration()`: Called on new home chain to move `MIGRATING_IN` → `OPERATING`
* `finalizeMigration()`: Called on old home chain to release funds and move to `MIGRATED_OUT`

All functions accept the same intents (INITIATE_MIGRATION or FINALIZE_MIGRATION), allowing the same signed state to be used on both chains.

**2. Selective State Swapping (only where needed):**

* **`INITIATE_MIGRATION` (on new home chain)**: Swap `homeLedger` ↔ `nonHomeLedger` before storing
  * Incoming state has actions in `nonHomeLedger` (new home = current chain)
  * After swap, stored state has actions in `homeLedger` (current chain)
  * Result: Next operation calculates deltas correctly from `homeLedger`

* **`FINALIZE_MIGRATION` (on old home chain)**: Swap `homeLedger` ↔ `nonHomeLedger` before validation
  * Incoming state (after user swaps) has old home actions in `nonHomeLedger` (current chain)
  * After swap, validation sees actions in `homeLedger` (current chain)
  * Result: Validation and fund release logic work correctly

* **No swap needed** for `INITIATE_MIGRATION` (on old home chain) and `FINALIZE_MIGRATION` (on new home chain) (homeLedger already represents current chain)

**3. Special Delta Calculation for `FINALIZE_MIGRATION` (on new home chain):**

When finalizing migration on the new home chain, the previous state (from `INITIATE_MIGRATION`) has allocations in `nonHomeLedger` (before swap) but was swapped when stored. Delta calculation must account for this:

```solidity
delta = candidate.homeLedger.netFlow - prevStoredState.homeLedger.netFlow
```

This works because `prevStoredState` was swapped during `INITIATE_MIGRATION`.

#### Implementation Notes

* Signatures are validated **before** swapping (using the original signed state)
* After swapping, signatures are invalidated (`userSig = ""`, `nodeSig = ""`) to prevent misuse
* The swapped state is only used internally for storage and validation
* Events emit the original signed state (before swap) for off-chain observability
* This approach maintains the critical invariant: **ChannelEngine always sees homeLedger as the current chain**

---

## Security model summary

* **Authorization**: all state changes require valid signatures.
* **Monotonicity**: `version` strictly increases.
* **Replay resistance**: no two states with the same version can coexist.
* **Cross-deployment replay protection**: Each ChannelHub deployment has a `VERSION` constant. The version is encoded as the first byte of `channelId`, ensuring that signatures are bound to a specific ChannelHub version. This prevents replay attacks across different ChannelHub deployments on the same chain. The `escrowId` inherits this protection as it is derived from `channelId`.
* **Liquidity safety**: absolute allocations must be collateral-backed.
* **Optimistic safety**:

  * challenges always resolve by enforcing the newest valid state,
  * stalled cross-chain operations can always be completed or reverted.

* **Token compatibility**:

  Only standard ERC20 tokens and native ETH are supported. The following token types are incompatible with the static ledger model:

  * **Rebasing tokens** (e.g. stETH, aTokens): their autonomous balance changes are invisible to the ledger and create unrecoverable accounting divergence. Use non-rebasing wrappers instead (e.g. wstETH).
  * **Fee-on-transfer tokens**: the amount received by the contract is less than the amount recorded, causing the ledger to overstate holdings from the very first deposit.

  There is no hard-coded guardrail preventing deposit of these tokens — the contract will accept them, but any discrepancy will produce undefined accounting behavior for all users of that token. Enforcement is off-chain: the Node will not sign states that reference unsupported token types.

* **Transfer failure resilience**: Outbound transfers (to users) never revert on failure:

  * Failed transfers (due to blacklists, hooks, or token features) accumulate in a reclaim balance,
  * Gas limiting (100k gas) prevents gas depletion attacks from malicious recipients or ERC777/ERC1363 hooks,
  * Users can later claim accumulated funds to alternative addresses,
  * This prevents two critical attack vectors:
    1. **Channel lifecycle denial**: User blacklists prevent state enforcement, blocking Node operations,
    2. **Node fund lock**: User forces Node to lock large funds via escrow deposit, then blocks all recovery operations with minimal capital.
  * Combined gas limiting + reclaim pattern ensures channel operations continue regardless of transfer success.

---

## Signature validation

The protocol supports flexible signature validation through two complementary systems: a validator registry and a bitmask for agreed validators. This design prevents signature forgery attacks while enabling custom signature schemes and maintaining cross-chain compatibility.

### Validator selection via approved validators bitmask

Agreed validators are specified in the `ChannelDefinition.approvedSignatureValidators` field (uint256 bitmask). The default ECDSA validator (0x00) is **always** available, regardless of the bitmask value. The bitmask specifies which additional validators from the node's registry are agreed validators. For example, if bit 42 is set to 1, then validator ID 42 from the node's registry is approved.

Since `approvedSignatureValidators` is part of the `channelId` computation, agreed validators cannot be changed during cross-chain operations without invalidating signatures. This prevents malicious nodes from forging user signatures by registering fake validators.

**Security properties:**

* Users control which validators (beyond the always-available default) can be used
* Cross-chain compatible (approvedSignatureValidators is in channelId, which is in all signatures)
* Zero transaction overhead (no separate validator registration needed)
* Prevents node-controlled validator forgery attacks
* Default ECDSA validator always available as fallback

### Node validator registry

The protocol uses a validator registry where NODE registers signature validators and assigns them 1-byte identifiers (0x01-0xFF).

**Design rationale:** This allows NODE to use flexible signature schemes (SessionKey, multi-sig, etc.) for its own signatures while preventing it from controlling user signature validation. Benefits:

* NODE can enforce its security requirements for node signatures
* NODE benefits from flexible validator implementations
* Cross-chain compatibility (validator addresses don't affect channelId or signature verification)
* User signatures remain protected via approved validators bitmask

### Validator registration

Nodes register validators by providing a signature over the validator configuration. This allows node operators to use cold storage or hardware wallets without exposing private keys to send transactions.

**Registration message:**

```solidity
bytes memory message = abi.encode(validatorId, validatorAddress, block.chainid);
```

The signature is verified using ECDSA recovery:

1. Try EIP-191 recovery first (standard for wallet software)
2. Fall back to raw ECDSA if needed
3. Verify recovered address matches the node address

The registration signature includes `block.chainid` for cross-chain replay protection, ensuring validator registrations are chain-specific and cannot be replayed across chains.

### Signature format

All signatures in the protocol follow this structure:

```text
[validator_id: 1 byte][signature_data: variable length]
```

**For user signatures:**

* `0x00` = Use ChannelHub's default ECDSA validator (always available)
* `0x01-0xFF` = Look up validator in node's registry, only if corresponding bit is set in `approvedSignatureValidators` (e.g., ID 42 allowed if bit 42 is 1)

**For node signatures:**

* `0x00` = Use ChannelHub's default ECDSA validator
* `0x01-0xFF` = Look up validator in node's registry (always available for nodes)

The first byte determines which validator verifies the signature. The remaining bytes are passed to the selected validator for verification.

### Cross-chain compatibility

The dual validator selection system solves critical cross-chain problems:

**User validators (approved validators bitmask):** Since the allowed validator bitmask is in `ChannelDefinition.approvedSignatureValidators`, which is part of `channelId`, it travels with every signature across all chains. No cross-chain synchronization is needed.

**Node validators (per-node registry):** Validator contracts may not be deployed to the same address on all chains. The registry uses 1-byte IDs instead of addresses, allowing the same validator ID to map to different addresses on different chains. Nodes register their validators independently on each chain.

---

### Bootstrap problem: validating the initial user signature

> Full analysis with all considered options and trade-offs: [`contracts/initial-user-sig-validation.md`](contracts/initial-user-sig-validation.md).

The `approvedSignatureValidators` bitmask protects user signatures on existing channels — the bitmask is part of `channelId`, so it is covered by every previously signed state. However, for `createChannel` there is no prior state: the bitmask comes from calldata supplied by the transaction sender, which may be the node itself.

This creates a circular dependency: the validator used to confirm the user's consent to the `ChannelDefinition` is selected from data the node controls. A node (basically, any address could be a node) could register a permissive validator and craft a `ChannelDefinition` that points to it, opening a channel without any real user signature.

The key distinction is between **bootstrap validation** (proving the user consented to the initial `ChannelDefinition`) and **ongoing validation** (proving the user authorized a specific state on an already-agreed channel). The `approvedSignatureValidators` bitmask is the right tool for ongoing validation but cannot self-validate at creation time.

#### Current deployment model: per-node ChannelHub

The current implementation binds each ChannelHub to a single node address at deploy time (`NODE` immutable). `_requireValidDefinition` rejects any `createChannel` call where `def.node != NODE`. This eliminates the any-address variant of the attack: a malicious actor cannot forge user signatures on a ChannelHub bound to a different node.

Within the trust boundary of the bound node the original vulnerability remains. Users who interact with a given deployment must already trust that node — they sign off-chain states with it and grant it ERC20 allowances — so the residual risk sits inside an existing trust relationship rather than being exploitable by an arbitrary third party.

A contract-enforced `VALIDATOR_ACTIVATION_DELAY` (1 day) provides a partial, targeted defence within this trust boundary: a newly registered validator cannot be used until the delay has elapsed, creating an observable window during which a key compromise can be detected and users can revoke ERC20 approvals before the attack on undeposited funds can execute.

A consequence of this model is that each node requires its own ChannelHub deployment; a single contract instance cannot serve multiple independent nodes.

#### Stronger alternatives

Two designs fully close the bootstrap gap without per-node deployments:

**Protocol-managed bootstrap registry (Option F).** A separate registry, controlled by a protocol multisig (`bootstrapAdmin`), lists the only validators that may be used for the `createChannel` user signature. Nodes have no influence over this registry. The initial set covers ECDSA and ERC-1271; additional schemes (e.g. an ERC-4337 freezer validator) can be added by governance without redeployment. The remaining trust assumption is that the multisig is not compromised.

**Two-registry system with tiered trusted validators (Option G).** The trusted validator set is split into a hardcoded tier (validator IDs 0–2, in contract bytecode) and a governance tier (IDs 3+, extensible by a multisig with a contract-enforced activation delay). `createChannel` accepts only hardcoded-tier IDs for the user signature; no governance action can influence it.

Subsequent operations accept both tiers, filtered by the bitmask stored at creation time. This makes `createChannel` fully admin-proof while preserving extensibility for later operations. ERC-4337 wallets with key-rotation needs are supported within the hardcoded tier via a FreezerProxy ERC-1271 wrapper, requiring no governance action.

---

## Mental model

* Off-chain protocol **decides what should happen**.
* On-chain contract **enforces the latest authorized accounting state**.
* Bridging is **non-atomic but recoverable**.
* The channel is **continuously enforceable**, not locked until closure.
