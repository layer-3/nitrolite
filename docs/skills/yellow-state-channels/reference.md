# Yellow Network State Channels — Complete Reference

Extended documentation for `yellow-state-channels`. Read the SKILL.md first
for the quick-start and lifecycle overview; this file covers the wire-level
on-chain details.

## EIP-712 Custody domain

```text
name:              "VirtualApp:Custody"
version:           "0.3.0"
chainId:           <chain_id of deployment>
verifyingContract: <Custody address from get_config.networks[].custody_address>
```

### Type hashes

- `STATE_TYPEHASH` — regular state-update signing (INITIALIZE / OPERATE /
  RESIZE / FINALIZE intents)
- `CHALLENGE_STATE_TYPEHASH` — challenge-path state signing

The two typehashes are distinct to prevent replay: a signature collected
for a cooperative update cannot be reused to force a dispute, and vice
versa.

## Data structures

### `Channel`

```solidity
struct Channel {
    address[] participants;  // [client, clearnode] — order matters
    address   adjudicator;   // validation-rules contract
    uint64    challenge;     // dispute window in seconds (min 3600)
    uint64    nonce;         // unique per channel (typically timestamp)
}
```

Channel ID derivation (deterministic — same off-chain and on-chain):

```text
channelId = keccak256(abi.encode(participants, adjudicator, challenge, nonce, chainId))
```

`chainId` IS part of the hash. The same `(participants, adjudicator, ...)`
on different chains yields different IDs.

### `State`

```solidity
struct State {
    StateIntent     intent;       // OPERATE(0) | INITIALIZE(1) | RESIZE(2) | FINALIZE(3)
    uint256         version;      // monotonic; higher always wins
    bytes           data;         // app-specific payload or magic numbers
    Allocation[]    allocations;  // fund distribution
    bytes[]         sigs;         // participant signatures
}
```

### `Allocation`

```solidity
struct Allocation {
    address destination;  // recipient
    address token;        // ERC-20 address (or zero address for native ETH)
    uint256 amount;       // smallest unit (wei for ETH, 6-decimals for USDC)
}
```

### Packed state (canonical signing payload)

```solidity
packedState = abi.encode(
    channelId,
    state.intent,
    state.version,
    state.data,
    state.allocations
);
// Sign keccak256(packedState) per EIP-191/712/1271/6492
```

### `state.data` conventions

| Magic | Hex | Meaning |
|---|---|---|
| `CHANOPEN`  | `0x7877` | Channel opening state |
| `CHANCLOSE` | `0x7879` | Channel closing state |

App-specific states put arbitrary ABI-encoded payloads here; the adjudicator
decides what's valid.

## EIP-712 amount encoding (v0.5+)

`allocations[].amount` in the **EIP-712 payload** is encoded as a **decimal
string** (`"1000.0"`), not a number. Motivation: avoid JS `Number` precision
loss past 2^53. Clients passing a raw number get rejected with
`"Invalid signature"`.

(On the on-chain side, `amount` remains `uint256` in smallest unit. The
string vs number distinction applies to the EIP-712 typed-data payload only.)

## On-chain contracts

### Custody — central contract

Implements multiple interfaces:

| Interface | Methods |
|---|---|
| `IChannel` | `create`, `join`, `checkpoint`, `challenge`, `close` |
| `IDeposit` | `deposit`, `withdraw`, `balanceOf` |
| `IChannelReader` | `getChannel`, `getState`, `getStatus` |

### Adjudicator — pluggable validation

Validates state transitions per app-specific rules. Chosen at channel
creation via `channel.adjudicator`.

Built-in adjudicators:

| Adjudicator | Rule | Use case |
|---|---|---|
| **SimpleConsensus** | All participants must sign every state | Default payment channels |
| **Remittance** | Only sender signs | One-way payments |

Custom adjudicators implement:

```solidity
interface IAdjudicator {
    function isValid(
        Channel calldata channel,
        State   calldata prev,
        State   calldata next,
        bytes[] calldata sigs
    ) external view returns (bool);
}
```

Return `true` iff the proposed state transition is acceptable under the
app's rules. The contract rejects any state where `isValid` returns `false`.

### NodeRegistry

Tracks registered ClearNode operators on-chain. Used for discovery and
reputation scoring.

### AppRegistry

Tracks registered application protocols (adjudicators) and their
parameters. Enables composable governance.

## Challenge-response protocol (full detail)

### Preconditions

- Channel in `ACTIVE` status
- Caller holds a validly signed state with higher `version` than any
  previously submitted

### Step-by-step

```text
1. Challenger has state S_n with version = n
   → calls challenge(channelId, S_n, sigs)
   → Custody verifies signatures via adjudicator.isValid
   → Custody stores S_n as "challenged state"
   → Channel: ACTIVE → DISPUTE
   → Timer starts: now + channel.challenge seconds
   → Event emitted: Challenged(channelId, S_n, expiration)

2. Window open (default 24h, minimum 1h per MIN_CHALLENGE_PERIOD)

3a. Honest counterparty has state S_m where m > n:
    → calls checkpoint(channelId, S_m, sigs)
    → Custody verifies S_m is valid and version > stored
    → Custody updates stored state to S_m
    → Channel: DISPUTE → ACTIVE
    → Dispute resolved; operation resumes

3b. Window elapses with no newer state:
    → Any party calls close(channelId, emptyCandidate, proofs)
    → Custody verifies expiration < now
    → Custody distributes funds per S_n.allocations
    → Channel: DISPUTE → FINAL
    → Metadata deleted; funds distributed
```

### When to checkpoint proactively

Long-running sessions should anchor state periodically to reduce worst-case
rollback:

```typescript
// Every 100 state updates, or 30 min since last checkpoint
if (stateVersion % 100 === 0 || minutesSinceLastCheckpoint > 30) {
  await Custody.checkpoint(channelId, currentState, sigs);
}
```

After a checkpoint, a dispute can only roll back to that state — not to
the channel opening.

## Supported signature formats

The Custody contract's signature verifier accepts all of:

- **Raw ECDSA** — `r || s || v`, 65 bytes, secp256k1 over `keccak256(packedState)`
- **EIP-191** — `\x19Ethereum Signed Message:\n...` prefix
- **EIP-712** — typed-data sig over the `VirtualApp:Custody` domain
- **EIP-1271** — smart-contract wallet `isValidSignature(hash, sig)` callback
- **EIP-6492** — counterfactual sigs for pre-deployed smart wallets

This means Safe multisigs, Argent, Rainbow, and counterfactual smart wallets
work natively — no EOA-only assumptions.

## Constants

| Constant | Value | Meaning |
|---|---|---|
| `MIN_CHALLENGE_PERIOD` | 3600 (1 h) | Floor for `channel.challenge` |
| `CLIENT_IDX` | 0 | `participants[0]` |
| `SERVER_IDX` | 1 | `participants[1]` (ClearNode) |
| `PART_NUM` | 2 | `participants.length` |

Recommended defaults:
- Payment channel `challenge`: 86400 (1 day)
- App-session `challenge`: 86400 (1 day)
- Staking session `challenge`: 604800 (7 days) — gives ample dispute time

## Security model

### Threat matrix

| Threat | Protection |
|---|---|
| Replay of old state | Monotonic `version` — older states rejected |
| State withholding | Challenge flow — honest party submits their signed copy |
| Unauthorized transitions | Signature verification per transition via adjudicator |
| Fund theft | Contract is sole custodian; only valid sigs release funds |
| Stale-state settlement | Higher `version` always wins |
| Operator malice | Users can exit unilaterally via `challenge` at any time |

### Assumptions

- At least one honest party per channel
- Blockchain is secure and censorship-resistant
- Participants can reach the chain within the challenge window (liveness)
- Cryptographic primitives (ECDSA, keccak256) are secure

### Liveness — the critical requirement

If a participant cannot access the chain during a challenge window, the
challenged state wins. This is why:

- ClearNode is always-on and monitors for challenges
- Default `channel.challenge` is at least 1 day
- Long-lived sessions (staking, governance) use 7-day windows

### Current protocol constraints

- Two participants per payment channel (client at index 0, ClearNode at
  index 1)
- Single allocation per participant per channel
- Same-token allocations within a channel
- `MIN_CHALLENGE_PERIOD = 3600` enforced at contract level

## Payment channels vs app sessions — clarifying the split

| Feature | Payment channel (this skill) | App session (`yellow-app-sessions`) |
|---|---|---|
| Participants | Exactly 2 | 2+ with weights |
| Governance | Fixed adjudicator | Weights + quorum threshold |
| Fund source | On-chain Custody deposit | Unified balance (off-chain) |
| Mid-session funding | `resize_channel` (on-chain tx) | `submit_app_state` DEPOSIT (off-chain) |
| State protocol | ECDSA over packedState | NitroRPC/0.4 with intent |
| Use case | P2P payments, channel funding | Escrow, games, prediction markets |
| Session ID | `channelId` (derived from Channel) | `app_session_id` (from definition hash) |

Payment channels are the on-chain foundation. App sessions layer on top of
unified balance and never touch the chain unless a dispute forces on-chain
settlement (and in v0.5 app-session disputes settle via associated payment
channels, not the app-session directly).

## Reference material

- Spec: <https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/channel-methods>
- Data structures: <https://docs.yellow.org/docs/0.5.x/protocol/app-layer/on-chain/data-structures>
- Migration guide: <https://docs.yellow.org/docs/0.5.x/guides/migration-guide>
- Source repo: <https://github.com/erc7824/nitrolite>
