# Security properties of on-chain Nitrolite protocol infrastructure

## Behavior

These are behavior rules of the Clearnode or the logic how a user (should) operate.

1. if challenged with an older state, then checkpoint with the latest one

This produces the following invariant:
> A channel can only be challenged with the latest known (even off-chain) state.

---

2. if Node is low on liquidity (below some threshold), User checkpoints latest off-chain state, and optionally closes the channel
(Or User requests to migrate channel to another chain where Node has liquidity)

Invariant:
> The Node always have funds to transfer to the User IN-BETWEEN OPERATIONS
(this is NOT TRUE for non-home chain deposit, -//- withdrawal or a home chain migration, please see below).

---

3. Node stops issuing new states when NON-HOME chain deposit, -//- withdrawal or a home chain migration has started and not yet finished

---

4. Both `cross-chain withdrawal` and `home-chain` migration end with a state pushed to a non-home chain, and
  `cross-chain deposit` results in either funds automatically unlocked for Node, or an already signed state that an unlock them.

Given the 3 and 4, an invariant:
> at any moment of time a CCTB state will certainly contain not more than 2 per-chain states.

---

5. A party never signs a state with a `version` that was already signed for this channel.

Invariant:
> No different states with the same `version` can exist for the same channel.

## Invariants

---

- (NOT TRUE) only less-or-equal amount of internally-accounted funds can be withdrawn (NOT TRUE for states that include "receive" off-chain ops)

The absence of the aforementioned invariant creates a huge risk of an attacker draining the Node.
To protect from this, the Node should keep CORRECT track of off-chain user funds.
CAUTION IS REQUIRED.

P.S. This invariant still can be enforced by updating `lockedFunds` per channel meta-variable during on-chain state processing,
e.g. when processing "receive X, withdraw Y", increase `lockedFunds` (and "lock" Node's funds in channel) by X, and then decrease by Y.

---

- User funds can be withdrawn only after channel is finalized (closed or challenged) or during WITHDRAW action
- any action is valid only with a Node's signature (for now, but this condition may be loosened to improve UX by making protocol more complex)
- a state with `version` <= `latestKnownVersion` per chain cannot be accepted as valid
- for challenge a state with `version` < `latestKnownVersion` per chain cannot be accepted as valid
- a channel with the same `channelId` cannot be created twice
- an escrow with the same `escrowId` cannot be created twice
- on-chain-stored state has already been processed

---

## Formal Invariants List

### Channel identity and authorization

1. **Channel uniqueness**: A channel identified by `channelId = hash(Definition)` can be created at most once.
2. **Cross-deployment replay protection**: Each ChannelHub deployment has a `VERSION` constant (currently 1). The version is encoded as the first byte of `channelId = setFirstByte(hash(Definition), VERSION)`, ensuring that the same channel definition produces different `channelId` values across different ChannelHub versions. This prevents signature replay attacks across different ChannelHub deployments on the same chain. Only one ChannelHub deployment per version per chain is intended. The `escrowId = hash(channelId, stateVersion)` inherits this protection.
3. **Signature authorization**: Every enforceable state must be signed by both User and Node (unless explicitly relaxed in future versions).
4. **Pluggable signature validation**: Signature validation is performed by validator contracts implementing the `ISignatureValidator` interface. The ChannelHub has a `defaultSigValidator` (0x00), and nodes maintain a registry of validators (0x01-0xFF). The first byte of each signature determines which validator is used: `0x00` for default, `0x01-0xFF` for node-registered validators.
5. **Validator security requirements**: Signature validators must be trustworthy, gas-efficient, and correctly implement validation logic. A compromised or buggy validator can break authorization for all channels using that validator. Validators should be immutable or have strict upgrade controls. Nodes are responsible for registering only trusted validators in their registry.
6. **Version monotonicity**: For a given channel, every valid state has a strictly increasing `version`.
7. **Version uniqueness**: No two different states with the same `version` may exist for the same channel.

---

### State validity

5. **Per-chain correctness**: For any per-chain state, allocations and net flows are internally consistent and non-negative where required by the chain role (home vs non-home).
6. **Single-chain enforcement (current scope)**: For single-chain operation, the home-state `chainId` must equal `block.chainid`.
7. **Allocation backing**: The sum of allocations in an enforced state must equal the amount of locked collateral implied by previous state plus net flow deltas.
8. **No retrogression**: A state with `version ≤ lastEnforcedVersion` cannot be enforced or checkpointed.

---

### Liquidity and accounting

9. **Locked funds safety**: Channel locked funds are never negative.
10. **Node liquidity constraint**: Whenever a state requires locking Node funds, the Node must have sufficient available on-chain liquidity at enforcement time.
11. **Controlled imbalance**: User or Node net flows may temporarily exceed allocations only during explicitly allowed escrow or migration phases.

---

### Operational semantics

12. **Deposit semantics**: A state with intent `DEPOSIT` must include a positive user net-flow delta.
13. **Withdrawal semantics**: A state with intent `WITHDRAW` must include a negative user net-flow delta and must not increase user allocation beyond previous allocation.
14. **Operate / checkpoint semantics**: A state with intent `OPERATE` must not change user net flow on the enforcing chain.
15. **Close semantics**: A state with intent `CLOSE` finalizes the channel and distributes allocations to both parties.

---

### Challenge mechanism

16. **Challenge admissibility**: A channel can only be challenged when in `OPERATING` state.
17. **Latest-state challenge rule**: A challenge must reference a state with `version ≥ lastEnforcedVersion`; if higher, that state is enforced first.
18. **Challenge resolution**: Any strictly newer valid state supersedes an active challenge and returns the channel to `OPERATING`.
19. **Challenge finality**: If no newer state is enforced before challenge expiry, the channel may be unilaterally closed using the last enforced state.

---

### Cross-chain and multi-state structure

20. **Bounded per-chain states**: At any moment, a cross-chain channel state contains at most two per-chain states (home and non-home).
21. **Flow suspension**: During escrow deposit, escrow withdrawal, or migration, the Node must not issue new states until completion or challenge.
22. **Recoverability**: Every escrow or migration phase must be completable or revertible via timeout and challenge on at least one chain.

---

## Cross-chain Operation Ordering

### Invariant 21 is an off-chain, not an on-chain, guarantee

Invariant 21 states that the Node must not issue new states during an in-progress escrow or migration.
This ordering constraint **cannot be fully enforced by the on-chain contract** for a fundamental
reason: a contract on chain B has no visibility into what is happening on chain A. It cannot query
whether an escrow is `INITIALIZED` on the other chain, whether a migration is halfway through, or
whether any other channel operation is pending.

### Why a flag-based on-chain guard is insufficient

An `actionStarted` flag per channel could block new operations while a cross-chain one is in
progress, but it would be asymmetric:

- **Escrow withdrawal**: flag can be raised on initiate and lowered on finalize — both happen on the
  non-home chain, so the full lifecycle is visible.
- **Escrow deposit**: flag raised on initiate (non-home chain), but the protocol-level finalization
  that matters (user allocation credited) happens on the home chain. The non-home chain only sees
  the node reclaiming locked funds, which is a different event.
- **Migration**: flag raised on initiate (non-home chain, `MIGRATING_IN`), but finalization happens
  on the old home chain, advancing the channel to `MIGRATED_OUT`.

Applying the flag consistently across all three would require per-channel storage tracking active
operations, cross-operation coordination logic, and would still not close the gap for deposit and
migration, where the terminal event is on a different chain.

### Design consequence: the on-chain contract must handle concurrent operations gracefully

Because on-chain enforcement of ordering is incomplete, the contract is designed so that any
reachable state sequence produces a correct outcome, even if the off-chain ordering invariant was
violated. Version monotonicity is enforced independently on each chain: each contract tracks only
the version of the last state it enforced locally, with no cross-chain synchronization of version
counters. The canonical example is:

1. Escrow withdrawal initiated on chain B (non-home) — `EscrowWithdrawalMeta` created, node funds
   locked.
2. Migration initiated on chain B — channel becomes `MIGRATING_IN`, chain B is now treated as home.
3. Escrow withdrawal finalized on chain B — routed via metadata presence
   (`_isEscrowWithdrawalHomeChain`), not via the mutable `_isChannelHomeChain` result, so the
   non-home path remains reachable and funds are correctly released to the user.

This flow is reachable on-chain only with submission order X → Y → X+1 (X = initiate escrow,
X+1 = finalize escrow, Y = initiate migration, Y > X+1); the signing order must still be
monotonically increasing — X+1 pre-signed as the execution phase before Y is signed — so only
rule 21 is broken. If the signing order were also X → Y → X+1, the on-chain contract would
behave identically; that would only add a violation of the version-monotonicity signing rule
with no additional on-chain effect.

Under correct Node behavior this sequence never occurs. But the contract handles it safely so that
no funds can be permanently locked if it does.

---

### Safety guarantees

23. **Enforcement determinism**: Enforcing the same `(prevState, candidateState)` pair always yields the same on-chain result.
24. **Invariant preservation**: Every state transition that can be enforced on-chain preserves all invariants listed above.
25. **Latest-state dominance**: The economically correct outcome is always determined by the latest valid signed state, regardless of enforcement order.

---

## Signature Validation Security

The Nitrolite protocol uses a pluggable signature validation system to support flexible authorization schemes. This section describes the security model and considerations for signature validators.

### Validator Architecture

The protocol uses two mechanisms for validator selection to prevent signature forgery attacks:

**Validator selection (via approved validators bitmask):**

- Agreed validators are specified in the `ChannelDefinition.approvedSignatureValidators` field (uint256 bitmask)
- The default ECDSA validator (0x00) is **always** available, regardless of the bitmask value
- The bitmask specifies additional validators from the node's registry that are agreed validators (e.g., if bit 42 is 1, validator ID 42 is approved)
- Since `approvedSignatureValidators` is part of `channelId` computation, agreed validators cannot be changed during cross-chain operations without invalidating signatures
- This prevents malicious nodes from forging user signatures by controlling validator selection

**Node validator registry:**

- Nodes register signature validators and assign them 1-byte identifiers (0x01-0xFF)
- Both users and nodes can only use agreed validators (from the bitmask) or the default validator
- The first byte of each signature determines which validator is used for verification

**Validator selection:**

- **Default validator** (0x00): The ChannelHub is initialized with a `defaultSigValidator` address that implements `ISignatureValidator`. This validator is used when the signature's first byte is `0x00`. **Always available**, regardless of `approvedSignatureValidators` bitmask.
- **Node-registered validators** (0x01-0xFF): Nodes register validators on-chain with unique IDs. Only available if the corresponding bit is set in `ChannelDefinition.approvedSignatureValidators` (e.g., bit 42 set = validator ID 42 approved).

**Registration security:**

- Nodes register validators by signing `abi.encode(validatorId, validatorAddress, block.chainid)` off-chain
- The signature includes `block.chainid` for cross-chain replay protection (chain-specific registrations)
- Anyone can relay the registration transaction (relayer-friendly)
- Registration uses ECDSA recovery (EIP-191 with raw ECDSA fallback)
- Registration is immutable (cannot change once set)
- Node's private key only signs, never sends transactions (supports cold storage/HSM usage)

**Cross-chain compatibility:**
The node registry design enables cross-chain operation without requiring validators to deploy to the same address on all chains. Validators are referenced by 1-byte IDs rather than addresses, ensuring channelId remains consistent across chains (derived from user, node, nonce, metadata - no validator addresses). The same validator ID can map to different addresses on different chains, and nodes register their validators independently on each chain.

**Domain separation:**
The protocol maintains clear separation between protocol concerns (ChannelHub) and cryptographic concerns (validators). ChannelHub defines protocol message structure (when and how channelId binds to states) and manages channel lifecycle. Validators verify cryptographic signatures using specific schemes and remain agnostic to protocol-level message structure. This separation is important for validator registration: it uses direct ECDSA recovery in ChannelHub (infrastructure concern, no channelId) rather than going through the validator abstraction (protocol state validation with channelId binding). This keeps `ISignatureValidator` focused on its primary purpose while allowing registration to be operational setup rather than protocol-critical security.

#### Available Validator Implementations

1. **ECDSAValidator** (`src/sigValidators/ECDSAValidator.sol`)
   - Standard ECDSA signature validation
   - Automatically tries EIP-191 (with Ethereum prefix) and raw ECDSA
   - 65-byte signatures: `[r: 32 bytes][s: 32 bytes][v: 1 byte]`
   - Recommended for all users and nodes

2. **SessionKeyValidator** (`src/sigValidators/SessionKeyValidator.sol`)
   - Session key delegation with metadata
   - Enables temporary signing authority (hot wallets, time-limited access)
   - Two-level validation: participant authorizes session key, session key signs state
   - **Safe for user usage** (with Clearnode validation)
   - **NOT safe for node usage** (no user-side validation) - see SessionKeyValidator Security Considerations below

See `signature-validators.md` for detailed documentation on each validator.

### Trust Model

- **Default validator trust**: All participants using the default validator (0x00) trust the ChannelHub deployer's choice of default validator.
- **User validator control**: Users control which additional validators (beyond the always-available default) can verify signatures via the `approvedSignatureValidators` bitmask in `ChannelDefinition`. This prevents nodes from forging user signatures by registering malicious validators. Users can approve specific validators from the node's registry by setting the corresponding bits.
- **Validator agreement**: Both users and nodes can only use agreed validators specified in the bitmask (plus the always-available default validator). This ensures that validators are mutually agreed upon and prevents unilateral changes to signature validation schemes.
- **Registration immutability**: Once a node registers a validator at a specific ID, it cannot be changed. This ensures that signatures created with a given validator ID remain valid for the lifetime of the ChannelHub deployment.
- **Cross-chain consistency**: The same validator ID may map to different validator addresses on different chains, but the security properties must remain equivalent. Nodes are responsible for registering compatible validators across chains.

---

### SessionKeyValidator Security Considerations

⚠️ **CRITICAL: SessionKeyValidator is designed primarily for USER usage, not NODE usage.**

#### Background

SessionKeyValidator enables delegation of signing authority to temporary session keys. The session key is authorized by a participant's signature, and metadata (expiration, scope, permissions) is hashed and included in the authorization.

**Key architectural decision**: Metadata validation is performed **off-chain** by the Clearnode, not on-chain. The smart contract only validates cryptographic signatures, not the semantic meaning of the metadata.

#### User Usage (Safe)

When a **user** employs SessionKeyValidator:

1. **Off-chain enforcement layer**: The Clearnode (node software) retrieves and validates session key metadata
   - Checks expiration timestamps
   - Enforces allowed channel IDs
   - Validates operation permissions
   - Refuses to countersign if metadata is invalid

2. **Countersignature protection**: Every state requires the Node to countersign
   - Node verifies session key authorization
   - Node rejects suspicious or invalid activity

3. **Limited blast radius**: If a user's session key is compromised:
   - Expired keys are rejected by Clearnode
   - Out-of-scope operations are rejected by Clearnode
   - Node refuses to countersign
   - Channel can be challenged and closed
   - User's main key remains secure

4. **Revocability**: User can stop using the session key at any time
   - Switch back to main key
   - Issue new authorization with different session key
   - No on-chain action required

#### Node Usage (Unsafe - Current Implementation)

When a **node** employs SessionKeyValidator (NOT RECOMMENDED):

1. **No off-chain enforcement**: The user has no equivalent to Clearnode
   - User cannot decode or validate node's session key metadata
   - No user-side software validates expiration or scope

2. **No countersignature protection**: The user's signature provides no protection in this scenario, as the user has no mechanism to validate the node's session key authorization. A compromised node session key has full, unchecked authority from the user's perspective.

3. **Unlimited and irrevocable authority**: If node's session key is compromised:
   - On-chain validation only checks cryptographic signatures
   - User cannot verify expiration (metadata is hashed)
   - User cannot verify scope limitations (metadata is hashed)
   - Session key has full node authority
   - User has no protection against misuse

4. **Asymmetric security**: User-side session keys are safe (Clearnode validates), node-side session keys are unsafe (no user-side validator)

---

## ERC20 Transfer Failure Attack Vectors

### Background

ERC20 transfers can fail for reasons beyond insufficient balance:

- **Token blacklists**: Centralized tokens (USDC, USDT) have admin-controlled blacklists
- **Token hooks**: ERC777/ERC1363 tokens execute recipient hooks that can revert
- **Token features**: Pausable, upgradeable, or custom token logic
- **Malicious control**: Users may programmatically trigger blacklisting or control hook behavior

The protocol cannot guarantee that ERC20 transfers to users will succeed, even when the ChannelHub is functioning correctly.

---

### Inbound Transfer Failures (User → ChannelHub)

**Impact**: Low - Protocol is protected

Inbound transfer failures occur during:

- Channel deposits (DEPOSIT intent)
- Escrow deposit initiation (INITIATE_ESCROW_DEPOSIT on non-home chain)

**Mitigation**: The Clearnode only processes operations after observing successful on-chain events. If a user signs a deposit state but the transfer fails on-chain, the state is never enforced, and the Node does not provide services based on unconfirmed deposits.

---

### Outbound Transfer Failures (ChannelHub → User)

**Impact**: CRITICAL - Multiple attack vectors

Outbound transfer failures create two categories of attacks:

#### 1. Channel Lifecycle Stuck

Any operation requiring payment to the user will revert if the transfer fails, blocking:

**Challenge response denial**:

- User challenges with old state
- Node attempts to respond with newer state requiring user payment
- Transfer to user reverts → Node cannot respond
- Node loses funds after challenge timeout

**State enforcement denial**:

- Node attempts to enforce withdrawal or closure
- Transfer to user reverts → Operation fails
- Channel state cannot advance

**Cooperative closure rug pull**:

- User signs CLOSE state
- User blacklists themselves before execution
- Closure transaction reverts → Node funds locked

#### 2. Node Funds Lock Attack (Most Critical)

**Attack scenario**: User forces Node to lock large funds with minimal capital

**Execution flow**:

0. User creates a channel with a small initial deposit (e.g., $0.000001)
1. User initiates escrow deposit with any amount (it can even be successfully retrieved later) on non-home chain
2. Node forced to lock equal liquidity on home chain
3. State V+1 checkpointed on-chain (preparation phase complete)
4. **User deliberately does NOT sign state V+2** (execution phase never completes)
5. User blacklists themselves (or triggers token blacklist)
6. User challenges escrow on non-home chain → Node cannot respond (no V+2 exists)
7. Node attempts operations on home chain (closure, withdrawal, challenge response)
8. **All operations requiring transfer to user REVERT**
9. **Node's funds locked forever in channel allocation**
10. User challenges escrow deposit on non-home chain after timeout → Node cannot respond (no V+2)

---

### Solution: Reclaim Pattern

**Design**: Never revert on outbound transfer failure. Instead, accumulate failed transfers and allow later claims.

---

### Gas Depletion Attacks

**Problem**: Gas depletion during outbound transfers creates the same attack vectors as transfer reverts, but through a different mechanism.

When a transfer consumes all available gas, the transaction reverts, enabling:

- Channel lifecycle stuck (preventing state enforcement, challenge responses, closures)
- Node funds lock attacks (forcing Node to lock large funds with minimal user capital)

**How it occurs:**

**Native token (ETH) transfers**: Current implementation forwards all available gas to recipient. Malicious recipient contract can consume arbitrary gas in `receive()` or `fallback()` function.

**ERC20 tokens with hooks**:

- **ERC777** (most dangerous): Executes `tokensReceived()` hook on recipient even for standard `transfer()` calls. This creates two distinct attack vectors:
  1. **Gas depletion**: hook consumes all forwarded gas, causing the transaction to revert
  2. **Donation-back double-spend**: hook sends tokens back to ChannelHub during the transfer, increasing ChannelHub's balance above `balanceBefore - amount`. A balance-delta success check would misidentify this as a failed transfer and incorrectly credit `_reclaims`, letting the recipient claim the same amount twice. The protocol therefore uses **return-value checking** (not balance-delta checking) to detect ERC20 transfer success, matching `SafeERC20.trySafeTransfer` semantics with a gas cap.
- **ERC1363/ERC677** (lower risk): Include `transferAndCall()` methods that trigger recipient hooks

**Why it matters**: Even if protocol primarily supports standard ERC20, human error can introduce vulnerable tokens:

- Token implements ERC20 interface but has hidden hooks (ERC777)
- Token upgrades to add hooks without interface change
- Wrapped/bridge tokens add hook functionality
- Future standards may introduce hooks

**Solution**: Limit gas forwarded to recipient contracts (100,000 gas for both native and ERC20 transfers).

**Why 100,000 gas is sufficient**:

- Native ETH: Simple transfers (~21k-23k), smart wallets (6k-9k)
- ERC20 standard: Base transfer (~50k), ERC777 hooks (~2.6k registry + <5k hook)
- Covers >99% of legitimate use cases

**Combined with reclaim pattern**: Gas limiting prevents depletion attacks; reclaim pattern handles all other failure modes (blacklists, paused tokens). Both protections are essential.

---
