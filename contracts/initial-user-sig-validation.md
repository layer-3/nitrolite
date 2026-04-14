# Initial User Signature Validation: Approaches Comparison

## Background: The Vulnerability

`createChannel()` accepts a `ChannelDefinition` from calldata and passes
`def.approvedSignatureValidators` into `_validateSignatures()`. That function calls
`_extractValidator()`, which selects the validator based on the first byte of `userSig`
and only checks that the chosen `validatorId` is set in `approvedSignatureValidators`
before invoking it.

A node can register a malicious `ISignatureValidator` via `registerNodeValidator()` that
always returns `VALIDATION_SUCCESS`, then set `approvedSignatureValidators` in calldata to
include that validator. `createChannel()` then succeeds with an arbitrary `userSig` —
the channel is created without user consent.

The same bypass applies to `closeChannel()`, enabling the node to close the channel and
push locked funds to itself.

**Root cause — circular dependency:**

```txt
approvedSignatureValidators  (supplied by attacker in calldata)
    → selects which validator to verify user's sig
        → verifies user's "approval" of approvedSignatureValidators
```

The protocol description notes that `approvedSignatureValidators` being part of `channelId`
prevents retroactive validator swapping on existing channels. It does nothing to prevent a
node from crafting a fresh `ChannelDefinition` with a malicious validator for a brand-new
channel, because there is no pre-existing user signature to protect.

---

## Constraints on the Fix

1. **No additional user interaction.** The existing interaction may change format, but an
   extra signing step or extra on-chain transaction is not acceptable.
2. **Session key compatibility.** Session keys allow users to pre-sign states off-chain so
   they are not required online per operation. Any fix must preserve this: the user should
   not be required to submit `createChannel` themselves.
3. **Relayer model.** Nodes or relayers may call `createChannel` on behalf of the user
   (e.g., to enforce a state when the user is offline).
4. **ERC-4337 (smart wallet) support — including revocability.** ERC-4337 wallets can
   rotate keys, which revokes historical signatures. A signature valid at signing time may
   fail `IERC1271(def.user).isValidSignature()` by the time `createChannel` is submitted.
   The fix must not silently break ERC-4337 users.
5. **Extensible without ChannelHub redeployment.** New wallet formats (ERC-4337, future
   schemes) must be addable without requiring users to migrate to a new deployment.
6. **No new trust on calldata-controlled data.** The validator used to verify user consent
   must not be derivable from attacker-controlled calldata.

---

## Cross-Cutting Concerns

### Session Keys and `createChannel`

Session keys are node-registered validators (e.g., id = 5 in the node's registry). Any fix
that restricts `createChannel` user-sig validation to a separate, node-independent set will
make session keys **unusable for `createChannel` specifically**.

This is acceptable: the user must sign the initial `createChannel` state with a
bootstrap-compatible method (ECDSA or a protocol-approved validator). Session keys continue
to work for every subsequent on-chain operation (`depositToChannel`, `withdrawFromChannel`,
`checkpointChannel`, etc.), where `approvedSignatureValidators` is used normally.

The off-chain protocol must produce both:

- a session-key-signed version of each state (for off-chain use and subsequent on-chain ops)
- a bootstrap-compatible signature for any state that may be submitted to `createChannel`

This is one additional signature format for the initial state, not an additional round trip.

### ERC-4337 + Signature Revocability + Freezer Contract

ERC-4337 wallets implement `IERC1271` but may rotate their signing keys, revoking historical
signatures. Calling `IERC1271(def.user).isValidSignature(hash, sig)` at `createChannel` time
will fail if the wallet has rotated its key since the state was signed.

The solution is a **signature freezer contract**:

1. While the signature is still valid, the user (or relayer) calls
   `freezer.freeze(channelId, signingData, userSig)`.
2. The freezer validates `IERC1271(def.user).isValidSignature(hash, userSig)` and records
   the result: `frozenSigs[keccak256(hash, userSig, def.user)] = true`.
3. `createChannel` is called; the bootstrap validator checks `freezer.isFrozen(...)`.
4. Even if the wallet later rotates keys, the frozen record is permanent.

**Critical constraint:** the freezer contract address must NOT come from calldata. If it did,
an attacker would embed a malicious always-true freezer — reintroducing the original
vulnerability. The freezer address must be a hardcoded or governance-approved reference
inside the bootstrap validator contract itself.

---

## Options

### Option A — Hardcoded Immutable Bootstrap Validator Set  *(NOT VIABLE)*

A fixed list of validators is hardcoded in the contract (or set at deploy time and never
changed). For `createChannel` user sigs, only these validators may be used.

Example initial set:

- id = 0: Default ECDSA / EIP-191 (EOAs)
- id = 1: ERC-1271 validator (calls `IERC1271(def.user).isValidSignature`)

**Pros:**

- Simple and fully auditable.
- No governance or admin trust required.

**Cons:**

- **Not extensible without redeployment.** If ERC-4337 support (freezer validator) is not
  included at launch, adding it requires migrating all users to a new ChannelHub.
- ERC-1271 branch has revocability issue for ERC-4337 (see above).
- Any future wallet format not covered at launch requires redeployment.

---

### Option B — Universal Bootstrap Validator (immutable, ECDSA + ERC-1271 + ERC-6492) *(NOT VIABLE)*

A single validator contract, referenced by an `immutable` in ChannelHub, that internally
handles multiple signature formats:

- Raw ECDSA / EIP-191 (for EOAs)
- ERC-1271: calls `IERC1271(def.user).isValidSignature(hash, sig)`
- ERC-6492: for counterfactual smart wallets — deploys the wallet, then calls ERC-1271

For `createChannel`, the user sig is always validated through this one validator;
`approvedSignatureValidators` plays no role.

**Why ERC-1271 does NOT reintroduce the original vulnerability:**

The original attack works because the validator is selected from a node-controlled registry.
For ERC-1271, the call target is `def.user` — the victim's own address. The attacker
controls the `sig` bytes but cannot influence the code at `def.user`. For a real victim:

- EOA (no code): the ERC-1271 call reverts → validation fails.
- Smart wallet (deployed code): the call goes to the wallet's own logic, which only accepts
  signatures that genuinely satisfy its security model.

An attacker cannot inject a malicious contract through the signature bytes, because the
target is fixed to `def.user`.

**ERC-6492 safety:** A counterfactual wallet address is `CREATE2(factory, salt, initCode)`.
An attacker cannot deploy malicious code to this address without knowing the exact
factory + salt + initCode, which would just deploy the real wallet.

**Pros:**

- No node influence; validator address is immutable.
- ERC-4337 wallets that have not rotated keys are supported via ERC-1271.
- ERC-6492 supports counterfactual wallets.
- No governance required.

**Cons:**

- **Does not solve ERC-4337 revocability.** If the user's wallet rotates keys before
  `createChannel` is submitted, `isValidSignature` fails even for a legitimate state.
- The freezer pattern cannot be added without redeployment, because the freezer address
  cannot come from calldata (vulnerability) and there is no extensibility point.
- Session keys cannot be used for `createChannel` (acceptable — see cross-cutting section).

---

### Option C — Upgradeable Bootstrap Validator (single address, governance-controlled)  *(UNDESIRABLE)*

ChannelHub holds a mutable reference to a bootstrap validator contract. Governance
(multisig/DAO) can replace the contract address to add new validation schemes.

**Pros:**

- Fully extensible. Any new format can be supported by deploying a new validator and
  updating the reference.

**Cons:**

- **High-severity governance risk.** Admin can swap the bootstrap validator to a malicious
  contract, instantly breaking user signature guarantees for all future `createChannel`
  calls. This concentrates enormous power in the admin key.
- Requires timelock + multisig design to be safe, adding significant complexity to the
  security-critical path.
- Users cannot verify which bootstrap validator will be used when they sign states.

---

### Option D — User-Controlled Bootstrap Validator Registry  *(UNDESIRABLE)*

Each user maintains their own bootstrap validator registry:
`mapping(address user => mapping(uint8 id => ISignatureValidator))`.

For `createChannel`, the user sig must use either the default ECDSA validator or a validator
the user has registered under their own address.

**Pros:**

- Completely user-controlled. No node or protocol team can influence which validators apply.
- Supports arbitrary wallet types.

**Cons:**

- **Requires an extra on-chain transaction.** Users must call `registerBootstrapValidator`
  before they can use a custom validator for `createChannel`. This violates constraint 1.
- Bootstrap problem for the registry itself: registration via `msg.sender` works for
  ERC-4337 wallets, but ERC-4337 users who want to register before deploying their wallet
  have a chicken-and-egg problem.

---

### Option E — `msg.sender == def.user` *(NOT VIABLE)*

Require the user themselves to submit `createChannel`.

**Why this is not viable:**

The protocol's session key model allows users to pre-sign states off-chain so they do not
need to be online per operation. Requiring `msg.sender == def.user` forces the user to be
present and submit the transaction themselves — eliminating the session key benefit. This
directly contradicts constraint 2.

Additionally, if the user is offline (the precise scenario where a node might need to
enforce a state), the channel cannot be created.

---

### Option F — Protocol-Managed Bootstrap Validator Registry

ChannelHub maintains a separate, protocol-controlled bootstrap validator registry:

```solidity
mapping(uint8 bootstrapValidatorId => ISignatureValidator) internal _bootstrapValidators;
address public bootstrapAdmin; // multisig / governance
```

`bootstrapAdmin` is the only entity that can add validators to this registry. It is
completely separate from the node validator registry. For `createChannel`, the user sig's
first byte selects a bootstrap validator ID; only IDs present in `_bootstrapValidators`
are accepted.

Initial registry:

- id = 0: Default ECDSA / EIP-191 (hardcoded, always available)
- id = 1: ERC-1271 validator (added at deploy time)

Future additions via governance:

- id = 2: ERC-4337 freezer bootstrap validator (when ERC-4337 support is needed)
- id = N: future formats

**ERC-4337 freezer validator (id = 2):**

A separate `ERC4337BootstrapValidator` contract is deployed with the `SignatureFreezer`
contract address hardcoded. When governance adds it to the bootstrap registry:

1. User (or relayer) calls `freezer.freeze(channelId, signingData, userSig)` while the
   ERC-4337 signature is valid.
2. `createChannel` is called with `userSig = [0x02, originalSig]`.
3. `ERC4337BootstrapValidator` calls `freezer.isFrozen(hash, originalSig, def.user)`.
4. Channel is created. Future key rotations on the ERC-4337 wallet do not affect this.

The freezer address is **hardcoded in the validator contract**, not in the calldata.
An attacker cannot inject a malicious freezer by manipulating `userSig` bytes.

**Security comparison with original vulnerability:**

| | Original attack | Option F attack |
| --- | --- | --- |
| Who adds the malicious validator? | Any node (permissionless) | Only `bootstrapAdmin` (governance) |
| Precondition | None | Admin key compromise |
| Scope | Per-channel, any node can exploit | Affects all channels globally |
| Mitigation | Requires protocol fix | Multisig + timelock on admin |

Moving from "any node can exploit immediately" to "requires compromising a multisig admin
with a timelock" is a significant security improvement, comparable to the trust model of
most DeFi protocols.

**Pros:**

- Extensible without redeployment: governance adds new validators as needed.
- ERC-4337 fully supported via freezer validator (addable by governance, no migration).
- Session keys: unaffected for all operations except `createChannel` (acceptable).
- No node influence on bootstrap registry.
- Future wallet formats can be added incrementally.

**Cons:**

- Governance trust assumption: bootstrapAdmin must be secured (multisig + timelock).
- Users cannot verify at signing time which bootstrap validators will be available when
  `createChannel` is submitted (mitigated: id=0 ECDSA is always available as fallback).

---

### Option G — Two-Registry System with Tiered Trusted Validators

Split validator management into two completely independent systems, each with its own
bitmask field in `ChannelDefinition`:

- **`approvedTrustedValidators`** — bitmask into a protocol-managed registry
  (`validatorAdmin`, a hardcoded multisig)
- **`approvedNodeValidators`** — bitmask into the per-node registry (current system)

Both bitmasks are part of `channelId` and therefore part of every signed state. Adding a
new validator to either registry has zero effect on existing channels — the new ID is not
in their stored bitmasks.

**Tiered trusted registry — the key structural property:**

The trusted registry is split into two tiers:

- **Hardcoded tier** (ids 0–2, immutable in contract bytecode):
  - id = 0: Default ECDSA / EIP-191
  - id = 1: ERC-1271 (calls `IERC1271(def.user).isValidSignature`)
  - id = 2: ERC-6492 (counterfactual wallet deployment + ERC-1271)
- **Governance tier** (ids 3+, `validatorAdmin`-extensible):
  - id = 3: ERC-4337 freezer validator (when needed)
  - id = N: future formats

`createChannel` accepts **only hardcoded-tier IDs** for user sig validation, ignoring
`approvedTrustedValidators` entirely. All subsequent operations accept either tier,
filtered by the channel's stored `approvedTrustedValidators` bitmask.

**Why this eliminates the admin-compromise-at-createChannel risk:**

The circular dependency at `createChannel` is broken by construction: no governance action
can influence which validators are usable at channel creation. A compromised `validatorAdmin`
can add malicious validators to the governance tier, but those IDs are rejected by
`createChannel`'s hardcoded-tier check regardless of what appears in calldata.

For subsequent operations, the bitmask stored at creation time (already signed by the user
via the bootstrap step) gates which governance-tier validators are accepted. A newly added
malicious validator cannot be used on any existing channel. It could only be used on a
future channel where the user explicitly includes that ID in their `approvedTrustedValidators`
— which a well-behaved off-chain client would never do for an unknown ID.

**ERC-4337 revocability within the hardcoded tier:**

ERC-4337 wallets are covered by id = 1 (ERC-1271) for the non-rotation case. For the
rotation case, a **FreezerProxy** pattern keeps everything within the hardcoded tier:

1. User deploys a `FreezerProxy` smart contract as their `def.user` address (or via ERC-6492
   counterfactual deployment, id = 2).
2. `FreezerProxy` implements ERC-1271 and wraps the underlying ERC-4337 wallet. It
   permanently records successful validations: `frozenSigs[hash] = true`.
3. Before key rotation, user calls `FreezerProxy.freeze(hash, sig)` — freezer validates
   against the current wallet and records the result.
4. `createChannel` calls `IERC1271(def.user).isValidSignature()` → hits FreezerProxy →
   returns true regardless of subsequent key rotation.

No governance action or redeployment is needed for ERC-4337 support.

**Security comparison across scenarios:**

| Scenario | Result |
| --- | --- |
| Node registers malicious validator, calls `createChannel` | Rejected — node registry not used for createChannel user sigs |
| `validatorAdmin` compromised, malicious governance-tier validator added | `createChannel` unaffected (hardcoded tier only). Existing channels unaffected (bitmask). New channels unaffected (off-chain client won't include malicious ID). |
| `validatorAdmin` compromised, malicious governance-tier validator added, user explicitly opts in | User's own choice; off-chain client is responsible for warning |

**Pros:**

- `createChannel` is fully admin-proof: immutable hardcoded tier eliminates the bootstrap
  circular dependency without any governance dependency.
- Existing channels are protected from newly added validators by bitmask isolation.
- Extensible for subsequent operations without redeployment: governance tier grows over time.
- ERC-4337 supported within the hardcoded tier via FreezerProxy + ERC-1271/ERC-6492.
- Clean separation of responsibility: hardcoded tier = bootstrap trust anchor; governance
  tier = operational extensibility; node registry = per-node flexibility.
- Session keys unaffected for all operations except `createChannel` (acceptable).

**Cons:**

- More structural complexity: two bitmask fields in `ChannelDefinition`, two registry
  lookups, tiered logic in `createChannel`.
- FreezerProxy adds deployment overhead for ERC-4337 users who need rotation protection
  at the hardcoded tier (avoidable if they accept the governance-tier freezer validator
  once added).
- `validatorAdmin` compromise still affects the governance tier for future channels
  where users opt in; mitigated by multisig + contract-enforced activation delay.

**Contract-enforced activation delay:**

To further limit governance-tier risk, each added validator stores an `activatedAt`
timestamp. `createChannel` (where it applies) and subsequent operations reject validators
where `block.timestamp < activatedAt`. This gives users a guaranteed observation window
even if the multisig is partially compromised.

```solidity
struct TrustedValidator {
    ISignatureValidator validator;
    uint64 activatedAt;
}

mapping(uint8 id => TrustedValidator) internal _trustedValidators; // governance tier
// ids 0-2 hardcoded as immutables, never in this mapping
```

### Option H — Per-Node ChannelHub Deployment

Store a single `node` address immutably in the ChannelHub at deploy time. `createChannel`
requires `def.node == storedNode`; any attempt to open a channel with a different node
address is rejected.

```solidity
address public immutable NODE;

constructor(address node, ...) {
    NODE = node;
}

function createChannel(ChannelDefinition calldata def, State calldata initState) external {
    require(def.node == NODE, IncorrectNode());
    ...
}
```

**Threat model shift, not a direct fix:**

This option does not eliminate the malicious validator attack — the stored node can still
register a malicious validator, set `approvedSignatureValidators` to include it in
`ChannelDefinition` calldata, and forge a user signature. What it does is change the scope
of who can perform the attack: only the one node bound to this deployment, not any arbitrary
node.

This matters because users who interact with a given ChannelHub deployment must already
trust the bound node (they sign off-chain states with that node, grant ERC20 allowances to
that contract, etc.). The validator forgery attack therefore sits within the existing trust
boundary — it is one more thing a malicious version of an already-trusted node could do,
rather than something any third-party node can do against users of a shared hub.

**Validator activation delay (`VALIDATOR_ACTIVATION_DELAY`):**

Within Option H the bound node still has the ability to register a malicious validator and
immediately weaponise it. A contract-enforced activation delay partially closes this window
for one specific attack surface:

- Exploiting a newly registered validator to drain **user ERC20 approvals** via a fake
  `createChannel(DEPOSIT)` requires the validator to be active first.
- With a 1-day delay, the registration is visible on-chain before it can be used, giving
  the node operator (or watchers) time to detect the compromise, broadcast an alert, and
  give users time to revoke ERC20 approvals.

**Pros:**

- Trivially simple to implement — two lines of code for the node-binding; small struct
  change + one timestamp check for the activation delay.
- No changes to the validator selection logic required.
- Cross-node attacks are structurally impossible: a rogue node cannot exploit another node's
  ChannelHub to drain users of that node.
- No admin, no governance, no multisig dependency.
- Activation delay creates a detectable, on-chain signal for monitoring before an attack
  on ERC20 approvals can execute.
- Compatible with any other option: H can be combined with G to get both per-node scoping
  and a proper validator fix.

**Cons:**

- **Does not fix the vulnerability within the trust boundary.** The bound node itself retains
  the ability to forge user signatures (after the activation delay). Users must accept that
  residual risk or combine H with a validator-level fix.
- **One deployment per node.** Nodes cannot share a ChannelHub; each requires its own
  deployment and its own separate ERC20 approval from users.
- **Operational fragmentation.** Users interacting with multiple nodes must track multiple
  contract addresses and manage approvals per deployment.
- **Activation delay adds operational overhead.** Nodes must pre-register validators 1 day
  before first use — a minor one-time cost per validator.

---

## Options Comparison Table

| | Node-independent | `createChannel` admin-proof | ERC-4337 revocability | Session key (createChannel) | Session key (subsequent ops) | Extensible (no redeploy) | Extra user interaction | Multi-node support |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **A** (immutable set) | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ | No | ✓ |
| **B** (universal immutable) | ✓ | ✓ | ✗ (key rotation) | ✗ | ✓ | ✗ | No | ✓ |
| **C** (upgradeable single) | ✓ | ✗ | ✓ (via new validator) | ✗ | ✓ | ✓ | No | ✓ |
| **D** (user registry) | ✓ | ✓ | ✓ | ✗ | ✓ | ✓ | **Yes** (registration tx) | ✓ |
| **E** (msg.sender check) | ✓ | ✓ | ✓ | **✗ (breaks model)** | ✓ | ✓ | **Yes** (must be online) | ✓ |
| **F** (protocol registry) | ✓ | ✗ | ✓ (via freezer) | ✗ | ✓ | ✓ | No | ✓ |
| **G** (two-registry + tiered) | ✓ | ✓ | ✓ (via FreezerProxy) | ✗ | ✓ | ✓ | No | ✓ |
| **H** (per-node hub) | Partial (cross-node only) | Partial (cross-node only) | — | ✓ | ✓ | ✓ | No | **✗** |
