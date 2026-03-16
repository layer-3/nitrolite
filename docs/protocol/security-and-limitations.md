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
- **State finality** — the latest mutually signed state can always be enforced on-chain
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

- Any party MAY submit the latest signed state at any time
- The blockchain layer accepts only states with valid signatures and a higher version than the current on-chain state
- After the challenge period, the enforced state becomes final
- Final state allocations determine asset distribution

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

Participants do not need to trust nodes for:

- **Single-chain asset custody** — assets on the home chain can always be recovered through on-chain enforcement
- **State validity** — invalid states are rejected by signature and validation rules

## Known Limitations

The following capabilities are not yet implemented:

- Trustless off-chain state operations (node liquidity enforcement)
- Validator network for monitoring node behaviour and enforcing correctness
- Watchtower services for automated enforcement
- Support for non-EVM blockchains
- Formal verification of protocol rules

## Future Improvements

The protocol roadmap includes the following planned improvements:

- **Validator network** — off-chain state advancement can be independently validated; a validator network would monitor on-chain actions and penalize node misbehaviour that harms the ecosystem
- **Extension layer on-chain enforcement** — removing the reliance on node liquidity trust for extension layer operations
- **Non-EVM blockchain support** — redesigning the protocol to support blockchains beyond the EVM ecosystem (planned for V2)
- **Watchtower integration** — automated monitoring and enforcement on behalf of users

---

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)
