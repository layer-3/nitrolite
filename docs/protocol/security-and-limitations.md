# Security and Limitations

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)

---

This document describes the security guarantees of the Nitrolite protocol, its current trust assumptions, and the known limitations of the present version.

## Protocol Maturity

The core protocol functionality is implemented and operational. A user MAY operate over a unified asset, deposit and withdraw on any supported blockchain, and conduct the majority of interactions without direct blockchain involvement. The protocol provides protection against unauthorized state changes from the user side — no user can unilaterally alter the state without valid signatures from all required participants.

However, the protocol in its current form is not fully trust-minimized. The primary remaining trust assumption concerns node behaviour and liquidity, as described in the sections below. The protocol is under active development, with planned improvements to address these limitations.

## Security Goals

The protocol aims to guarantee:

- **Asset safety** — participants MUST NOT lose assets without signing a state that authorizes the change
- **State finality** — the latest mutually signed state can be enforced on-chain when a valid execution path exists for its intent; parties MUST retain any intermediate states required to establish that path
- **Non-repudiation** — a participant cannot deny having signed a state
- **Censorship resistance** — any party MAY independently enforce state on the blockchain layer

## Off-Chain Safety

The protocol protects against invalid or malicious state submissions through:

**Signature requirements**
Every state update requires valid signatures from all required participants. No participant can unilaterally change the state.

**Version ordering**
State versions are strictly increasing. Old states cannot replace newer states.

**Asset conservation**
State transitions MUST preserve total asset amounts within each ledger. No assets can be created or destroyed through state updates.

**Transition validation**
Each state update MUST satisfy transition-specific rules. Invalid transitions are rejected.

## Enforcement Guarantees

The blockchain layer provides the following guarantees:

- Any party MAY submit the latest mutually signed state to the blockchain layer; enforcement succeeds when a valid execution path exists for that state's intent in the current channel context
- Parties MUST retain and enforce intermediate states (such as a DEPOSIT state) before discarding them — a subsequent OPERATE state built on top of an unenforced DEPOSIT cannot be used to create or checkpoint a channel on-chain, because OPERATE requires zero change in user net flow relative to the last enforced state
- The blockchain layer accepts only states with valid signatures and a higher version than the current on-chain state
- After the challenge period, the enforced state becomes final
- Final state allocations determine asset distribution

## Off-Chain Convergence with On-Chain State

The Node maintains a local view of each channel's state by subscribing to on-chain events emitted by the `ChannelHub` contract. To defend against out-of-order event delivery — which can be caused by contract-level reentrancy, indexer mis-ordering, or reorg replay — every event handler that mutates `channels.state_version` or `channels.status` enforces a strict version-monotonicity guard: an event whose `StateVersion` is lower than the row's current `StateVersion` is dropped with a structured warn log.

For home-channel events (`ChannelChallenged`, `ChannelCheckpointed`, `ChannelClosed`), a dropped event additionally triggers a chain-state refresh: the Node calls `getChannelData` on the home-chain `ChannelHub` contract via the `ChainStateRefresher` and overwrites the local row with the authoritative on-chain status, version, and challenge expiry. This ensures convergence with chain even when single-event delivery is insufficient — for example, when an outer `ChannelChallenged` is delivered after a higher-version inner `ChannelCheckpointed` emitted from the same transaction.

The refresh runs synchronously inside the event-processing transaction. On RPC failure the transaction rolls back and the listener replays the event, so the convergence opportunity is never silently lost.

Escrow event handlers ship the version guard without the refresh hook; the cross-chain RPC plumbing required for escrow refresh is a deferred follow-up item. Pending its arrival, escrow rows can remain divergent from chain across an interim window until the next on-chain event arrives.

The on-chain side complements this guard: every `ChannelHub` lifecycle entrypoint is protected by OpenZeppelin's `nonReentrant` modifier. This prevents inbound token hooks (ERC777 `tokensReceived`, non-standard `transferFrom` callbacks) from interleaving lifecycle operations and producing the out-of-order events that would otherwise force the Node's defense-in-depth path to fire.

## Node Liquidity and Cross-Chain Trust

Each user channel is opened with a node. To maintain cross-chain functionality, the node MUST hold sufficient liquidity on each supported blockchain to satisfy off-chain state allocations.

When a user with home chain A transfers assets to a user with home chain B, the node receives the amount on chain A and allocates from its own balance to the recipient on chain B. This process occurs entirely off-chain. If the recipient subsequently wishes to enforce their state on chain B and the node does not hold sufficient liquidity on that chain, the on-chain enforcement will fail.

In the current protocol version, this constitutes a trust assumption: users rely on the node operator to maintain adequate liquidity across all supported chains. Node operators are expected to manage their liquidity to cover off-chain obligations, but users cannot independently verify that this condition holds at all times.

## Current Trust Assumptions

In the current protocol version, participants MUST trust nodes for:

- **Liveness** — nodes MUST be online to facilitate off-chain state advancement
- **Cross-chain liquidity** — nodes MUST maintain sufficient funds on each supported chain to honour off-chain allocations; insufficient liquidity may cause on-chain enforcement to fail
- **Cross-chain relay** — nodes relay cross-chain state updates; trustless cross-chain enforcement is not yet implemented
- **Timely enforcement** — nodes are expected to submit checkpoints when requested; delayed enforcement may affect user experience but does not compromise single-chain asset safety
- **Off-chain transfer routing** — when a user sends funds off-chain to another party, the node must countersign both the sender's state (decreasing their allocation) and the receiver's credit state (increasing theirs); the on-chain contract cannot enforce atomicity between two independent channel updates. A malicious node could apply the sender's state while withholding the receiver's credit, capturing the transferred funds. Users must trust the node to faithfully execute both legs of every off-chain transfer.
- **Asset-symbol equivalence** — the node operator controls which chain-specific tokens are configured under a single unified asset symbol. The protocol treats all tokens sharing a symbol as fully fungible 1:1 representations of the same asset, so off-chain credit denominated in that asset can be redeemed from any of those token inventories regardless of which one originally backed it (the validator binds unchanneled credit to the chain/token chosen at first channel creation, enforcing only that the asset symbol matches). This is intended behaviour that enables cross-chain redemption. Operators MUST therefore configure only economically equivalent (1:1 redeemable) tokens under one symbol; grouping non-equivalent tokens (e.g. a test token and production USDC) under the same symbol would let credit sourced from the cheap inventory be redeemed against the valuable one. Token equivalence cannot be verified programmatically and is an operator configuration responsibility.
- **Signature validator registry** — the node operator controls which additional signature validators are registered on the ChannelHub contract. A malicious or compromised node could register a validator that approves forged user signatures, then use it to create channels or close them without the user's knowledge. A 1-day activation delay (`VALIDATOR_ACTIVATION_DELAY`) creates an observable window before any newly registered validator can be used. Users MUST monitor the `ValidatorRegistered` event on the ChannelHub contract and SHOULD revoke all ERC20 approvals granted to ChannelHub immediately upon detecting an unexpected registration. Once registered, a validator cannot be deactivated — the 1-day window is the entire response budget. Users SHOULD avoid granting large standing ERC20 approvals to ChannelHub to cap worst-case exposure.

Participants do not need to trust nodes for:

- **Single-chain asset custody** — assets on the home chain can always be recovered through on-chain enforcement
- **State validity** — invalid states are rejected by signature and validation rules

## Known Limitations

The following capabilities are not yet implemented or have acknowledged design trade-offs:

- Trustless off-chain state operations (node liquidity enforcement)
- Validator network for monitoring node behaviour and enforcing correctness
- Watchtower services for automated enforcement
- Support for non-EVM blockchains
- Formal verification of protocol rules
- Hook-enabled tokens (ERC777, ERC1363, non-standard ERC20 with re-entrant `transferFrom`) are not supported on currently-deployed `ChannelHub` instances. The Node operator MUST NOT onboard such tokens. Per-deployment status (chain ID, ChannelHub address, deploy commit) is recorded in [`contracts/deployments/HOOK-TOKEN-COMPATIBILITY.md`](../../contracts/deployments/HOOK-TOKEN-COMPATIBILITY.md). This restriction may be lifted on future deployments to new chains; each new deployment records its support status in the same file.
- Stored signatures are validated but not canonicalized. The node persists accepted `user_sig` / `node_sig` values verbatim (unbounded `text`) rather than re-encoding them after verification. A signature can therefore carry non-canonical or unused trailing bytes and still be retained once its cryptographic checks pass — most notably session-key signatures, where `SessionKeyValidator.validateSignature()` ABI-decodes a `(SessionKeyAuthorization, bytes)` payload and verifies only the embedded ECDSA signatures, leaving any extra bytes around an otherwise valid payload intact. Signatures learned from on-chain events (hex-encoded from the candidate and written via `UpdateStateSigsIfMissing`) follow the same path and bypass the WebSocket message-size limit enforced on direct RPC submissions. This carries no asset-safety risk — verification still gates acceptance — but consumers MUST NOT assume stored signature bytes are minimal or canonical.
- Session key off-chain scope enforcement does not apply to direct receive-state acknowledgement. Session key expiration and asset-scope restrictions are enforced by the Nitronode off-chain only; the `SessionKeyValidator` contract validates cryptographic signatures alone. A party holding a session key — even one that has expired, been revoked, or been retired — can bypass the `acknowledge` endpoint, manually sign a pending node-issued receive state, and submit it directly to the contract. This is accepted: receive states exclusively increase the user's allocation and cannot redirect funds away from the user, so out-of-scope acknowledgement carries no financial risk and preserves a recovery path when the node is unavailable.

## Future Improvements

The protocol roadmap includes the following planned improvements:

- **Validator network** — off-chain state advancement can be independently validated; a validator network would monitor on-chain actions and penalize node misbehaviour that harms the ecosystem
- **Extension layer on-chain enforcement** — removing the reliance on node liquidity trust for extension layer operations
- **Non-EVM blockchain support** — redesigning the protocol to support blockchains beyond the EVM ecosystem (planned for V2)
- **Watchtower integration** — automated monitoring and enforcement on behalf of users

---

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)
