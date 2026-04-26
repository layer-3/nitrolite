---
name: yellow-custody-contract
description: |
  Solidity-level reference for the Yellow Network Custody contract — the on-chain anchor for state channels. Covers the IChannel, IDeposit, IChannelReader interfaces; the `VirtualApp:Custody` EIP-712 domain (v0.3.0) with STATE_TYPEHASH and CHALLENGE_STATE_TYPEHASH; Channel/State/StateAllocation structs; StateIntent enum values; events (Created/Joined/Opened/Challenged/Closed/Resized); and key constants (MIN_CHALLENGE_PERIOD). Use when: building a non-SDK native on-chain integration, writing a custom adjudicator, auditing the contract surface, or debugging a revert from Custody.create / resize / close.
version: 1.0.0
contract_version: "VirtualApp:Custody v0.3.0"
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/on-chain/data-structures
  - https://github.com/erc7824/nitrolite
---

# Yellow Custody Contract — on-chain reference

The Custody contract is the security anchor of the Yellow Network state
channel protocol. Off-chain Nitro RPC is cheap and fast; when things go
wrong or settlement is needed, the Custody contract is the source of
truth. This skill is the Solidity-level reference — interfaces, domain,
structs, events, errors.

## Deployment

Addresses per chain come from the ClearNode's `get_config.networks[]`
response:

```json
{
  "broker_address": "0x...",
  "networks": [
    {
      "chain_id":           1,
      "name":               "Ethereum",
      "custody_address":    "0x...",
      "adjudicator_address": "0x..."
    }
  ]
}
```

Chains with mainnet deployments include Ethereum (1), Polygon (137),
Base (8453), Arbitrum One (42161), Optimism (10). Others (BNB, Linea,
XRPL EVM, World Chain) are listed in Nitrolite release notes — always
verify via `get_config`.

## EIP-712 domain

```solidity
name:              "VirtualApp:Custody"
version:           "0.3.0"
chainId:           <network>
verifyingContract: <Custody address>
```

**Type hashes**:

- `STATE_TYPEHASH` — used for all non-challenge state signatures
  (initialize, operate, resize, finalize).
- `CHALLENGE_STATE_TYPEHASH` — used for state submitted via the dispute
  path. Having a distinct typehash prevents replay: a signature collected
  for an ordinary update can't be reused to push a challenge state.

Both typehashes are defined in the contract source (`Custody.sol` /
`VirtualApp.sol` — check the `@erc7824/nitrolite` ABIs folder).

## Structs

### `Channel`

```solidity
struct Channel {
    address[] participants;   // [userWallet, clearnodeWallet]
    address   adjudicator;    // per-channel adjudicator contract
    uint64    challenge;      // dispute window in seconds
    uint64    nonce;          // unique per (participants, adjudicator) tuple
}
```

**Channel ID derivation** (off-chain — the contract computes the same):

```solidity
channelId = keccak256(abi.encode(participants, adjudicator, challenge, nonce, chainId));
```

### `State`

```solidity
struct State {
    StateIntent         intent;       // OPERATE(0) | INITIALIZE(1) | RESIZE(2) | FINALIZE(3)
    uint64              version;      // monotonically increasing per channel
    bytes               state_data;   // ABI-encoded app-specific state
    StateAllocation[]   allocations;  // who gets what on close
}

enum StateIntent {
    OPERATE,     // 0 — app-session state updates
    INITIALIZE,  // 1 — channel creation
    RESIZE,      // 2 — allocation adjustment
    FINALIZE     // 3 — closure
}
```

**Integer ordering matters** for ABI encoding — `OPERATE` is 0, not 1.

### `StateAllocation`

```solidity
struct StateAllocation {
    address participant;   // wallet address (post-v0.5.0 — was session key pre-v0.5)
    address token;         // ERC-20 address or zero for native ETH
    uint256 amount;        // smallest-unit amount
}
```

## Interfaces

### `IChannel` — lifecycle

```solidity
function create(
    Channel calldata channel,
    State calldata state,
    bytes calldata userSig,
    bytes calldata serverSig
) external;

function resize(
    bytes32 channelId,
    State calldata state,
    bytes[] calldata signatures
) external;

function close(
    bytes32 channelId,
    State calldata state,
    bytes calldata userSig,
    bytes calldata serverSig
) external;

function challenge(
    bytes32 channelId,
    State calldata state,
    bytes[] calldata signatures
) external;

function checkpoint(
    bytes32 channelId,
    State calldata state,
    bytes[] calldata signatures
) external;
```

Typical flow:

```
off-chain    create_channel (RPC) ─▶ ClearNode returns channel + state0 + serverSig
on-chain     Custody.create(channel, state0, userSig, serverSig)     (2 sigs)
off-chain    resize_channel (RPC) ─▶ ClearNode returns stateN + serverSig
on-chain     Custody.resize(channelId, stateN, [userSig, serverSig]) (2 sigs)
off-chain    close_channel (RPC)  ─▶ ClearNode returns FINALIZE state
on-chain     Custody.close(channelId, state, userSig, serverSig)     (2 sigs)
```

Dispute path: `challenge` replaces the cooperative close — only valid
states with `CHALLENGE_STATE_TYPEHASH` sigs are accepted, and a timer of
`channel.challenge` seconds runs down before finalization.

### `IDeposit` — unified-balance funding

```solidity
function deposit(address token, uint256 amount) external payable;
function withdraw(address token, uint256 amount) external;
function balanceOf(address account, address token) external view returns (uint256);
```

See `yellow-deposits-withdrawals` for the client-side flow.

### `IChannelReader` — on-chain introspection

```solidity
function getChannel(bytes32 channelId) external view returns (Channel memory);
function getState(bytes32 channelId) external view returns (State memory);
function getStatus(bytes32 channelId) external view returns (ChannelStatus);
```

Useful for verifying off-chain claims — if a ClearNode claims a channel
is closed, `getStatus` confirms independently.

## Events

```solidity
event Created(bytes32 indexed channelId, Channel channel, State state);
event Joined(bytes32 indexed channelId, address indexed participant);
event Opened(bytes32 indexed channelId);
event Challenged(bytes32 indexed channelId, State state, uint256 challengePeriodEnd);
event Closed(bytes32 indexed channelId, State finalState);
event Resized(bytes32 indexed channelId, State newState);
```

Subscribe for reconciliation: any off-chain state change that affects
on-chain truth emits exactly one of these.

## Constants

- `MIN_CHALLENGE_PERIOD = 3600` (1 hour) — minimum for `channel.challenge`.
- Recommended app-session challenge default: `86400` (1 day).
- `CLIENT_IDX = 0` — user/agent index in `Channel.participants`
- `SERVER_IDX = 1` — ClearNode index in `Channel.participants`
- `PART_NUM = 2` — fixed length of `Channel.participants` array

## `Status` enum (channel lifecycle — distinct from `StateIntent`)

| Value | Name | Meaning |
|---|---|---|
| 0 | `VOID` | Not yet created |
| 1 | `INITIAL` | Created, awaiting both participants' sigs on state 0 |
| 2 | `ACTIVE` | Open; accepts resize / operate updates |
| 3 | `DISPUTE` | Under challenge; timer counting down |
| 4 | `FINAL` | Closed; funds released to unified balances |

`StateIntent` (see Structs) answers "what kind of state is this?".
`Status` answers "what lifecycle phase is the channel in?". They are
orthogonal.

## EIP-712 amount encoding

**Post-v0.5**: `StateAllocation.amount` and `allowance.amount` in the
EIP-712 payload are **decimal strings**, not numbers. This is a migration-
guide-specified change to avoid JS `Number` precision loss past 2^53.
Clients passing `amount: 1000` (number) get `"Invalid signature"`;
passing `amount: "1000"` (string) works.

## Signature formats accepted

The contract verifies via an internal dispatcher that accepts:

- **Raw ECDSA** (65 bytes r || s || v) over keccak256 of the state hash
- **EIP-191** personal-sign format (`\x19Ethereum Signed Message:\n…`)
- **EIP-712** typed data (domain + STATE_TYPEHASH / CHALLENGE_STATE_TYPEHASH)
- **EIP-1271** — smart-contract wallet `isValidSignature(hash, sig)`
- **EIP-6492** — counterfactual sigs for pre-deployed smart wallets

This means Safe multisigs, Argent, Rainbow, and similar work natively at
the Custody layer — no EOA-only assumptions.

## Common reverts

| Revert | Cause |
|---|---|
| `InvalidSignature()` | Signer's recovered address not in `channel.participants` |
| `InvalidState()` | `state.version` not monotonic, or `intent` wrong for the operation |
| `ChannelNotOpen()` | Trying to `close` / `resize` a channel not in open state |
| `ChallengePeriodActive()` | `close` during a live challenge window |
| `InsufficientBalance()` | `deposit` / `withdraw` against insufficient on-chain balance |
| `InvalidAdjudicator()` | `channel.adjudicator` not whitelisted at the contract |

## Related

- `yellow-state-channels` — protocol-level channel lifecycle
- `yellow-deposits-withdrawals` — funding flows
- `yellow-sdk-api` — TypeScript helpers that build these calls
- `yellow-errors` — off-chain ClearNode errors (distinct from on-chain reverts)
