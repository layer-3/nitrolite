# State Enforcement

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)

---

This document describes how channel states are enforced on the blockchain layer.

## Purpose

Enforcement is the mechanism by which off-chain state is reflected on-chain. It serves two complementary roles:

1. **Regular state synchronization** — participants submit signed states to the blockchain layer to keep the on-chain record up-to-date with the latest off-chain state, particularly for transitions that require on-chain effects (deposits, withdrawals, escrow operations, migrations)
2. **Dispute resolution** — any participant MAY independently submit the latest mutually signed state to the blockchain layer to protect their assets if off-chain cooperation fails

The blockchain layer acts as the ultimate arbiter of channel state, providing security guarantees that do not depend on participant cooperation.

## Enforceable State Requirements

A state is enforceable on-chain if and only if:

- It is **mutually signed** — it carries valid signatures from both the user and the node
- The signatures use validation modes that are among the channel's approved signature validators (including session-key signatures if the session key validation mode is approved)
- The state has passed off-chain state advancement validation
- The node has sufficient balance on the target chain to cover any required fund locking

Node-issued pending states (those carrying only the node's signature) are NOT enforceable. They become enforceable only after the user acknowledges them, producing a mutually signed state.

## Enforcement Model

Off-chain states and on-chain enforcement states are related as follows:

- Participants advance state off-chain through signed updates
- At any time, any party MAY submit the latest mutually signed state to the blockchain layer
- The blockchain layer validates the submitted state and updates its record
- On-chain state always reflects the latest successfully checkpointed state

The on-chain state MAY lag behind the off-chain state. This is expected during normal operation for transitions with the OPERATE intent.

## Locked Funds Model

The blockchain layer tracks **locked funds** for each channel. Locked funds represent the total assets held by the enforcement contract on behalf of the channel.

Rules:

- Locked funds increase when assets are pulled from the user or from the node's vault into the channel
- Locked funds decrease when assets are released to the user or to the node
- Unless the channel is being closed, the sum of UserAllocation and NodeAllocation in the enforced state MUST equal the locked funds
- Locked funds MUST never be negative

The node maintains a **vault balance** per token on each chain. The vault is a pool of available funds separate from any specific channel. When a transition requires the node to lock additional funds, the required amount is deducted from the node's vault balance and added to the channel's locked funds.

| Operation  | User Fund Effect                          | Node Fund Effect                           | Locked Funds Effect           |
| ---------- | ----------------------------------------- | ------------------------------------------ | ----------------------------- |
| DEPOSIT    | Pull from user (positive delta)           | Adjusted by node net flow delta            | Increases by total deltas     |
| WITHDRAW   | Release to user (negative delta)          | Adjusted by node net flow delta            | Decreases by total deltas     |
| OPERATE    | No user fund movement                     | Adjusted by node net flow delta            | Adjusted by node delta only   |
| CLOSE      | Release UserAllocation to user            | Release NodeAllocation to node             | Set to zero                   |
| Challenge  | No fund movement                          | No fund movement                           | Unchanged (status changes)    |

## Channel Creation

Channels are created through an enforcement operation. A channel does not need to be created on-chain with its initial off-chain-created state — any validly signed state MAY be used for on-chain creation, provided the channel does not yet exist on-chain. This allows participants to advance state off-chain before enforcing the channel on-chain, e.g. when the user's first action is to receive a transfer from another user, they can additionally perform several transfer send or receive operations before submitting the state on-chain with a "WITHDRAW" intent, receiving funds simultaneously with creating a channel, both on-chain.

The creation process:

1. Participants agree on a channel definition and exchange signed state updates off-chain
2. A participant submits the channel definition and a signed state to the blockchain layer
3. The blockchain layer validates signatures, creates the channel record, and applies fund effects according to the state's intent
4. The channel is now active on the on-chain layer

The state submitted for channel creation MAY carry a DEPOSIT, WITHDRAW, or OPERATE intent.

## State Submission

State submission covers checkpoint, deposit, and withdrawal operations. The general process is identical for all three:

1. A participant constructs the enforcement representation of a signed state
2. The participant submits the enforcement representation along with all required signatures to the blockchain layer
3. The blockchain layer validates the submission
4. If valid, the on-chain state is updated and fund movements are applied

The behaviour differs only in intent-specific validation rules:

- **OPERATE** — the blockchain layer validates that the user net flow has not changed and that the node allocation is zero. No user fund movement occurs.
- **DEPOSIT** — the blockchain layer validates that the user net flow delta is positive (assets are flowing in). The deposited amount is pulled from the user and added to the channel's locked funds.
- **WITHDRAW** — the blockchain layer validates that the user net flow delta is negative (assets are flowing out). The withdrawn amount is released from the channel's locked funds to the user.

In all cases, the node's fund delta is adjusted according to the node net flow change.

## Challenge Operation

A challenge allows a participant to dispute the current on-chain state by submitting a signed state along with a separate challenger signature.

### Challenger Signature

The challenger signature is distinct from the state signatures. It is produced by signing the enforcement representation of the candidate state with the string "challenge" appended to the signing data. This guarantees that only a User or a Node can start a challenge, and not the third-party. However, a channel participant MAY share a valid challenger signature with a third-party, who then can successfully initiate a challenge.

**Only** the user or the node MAY act as the challenger.

### Challenge Process

1. The challenger submits a candidate state, state signatures, the challenger signature, and the challenger's participant index
2. The channel MUST NOT be in DISPUTE, MIGRATED_OUT or CLOSED statuses
3. The candidate version MUST be greater than or equal to the current on-chain version
4. If the candidate version is strictly greater than the current on-chain version, the blockchain layer validates and applies the new state (including fund effects)
5. The channel status is set to **DISPUTED** and the challenge expiry is set to the current time plus the challenge duration

### Resolving a Challenge

During the challenge period, any participant MAY respond by submitting a new valid state whose version is strictly greater than the currently disputed state. This replaces the disputed state, changes channel's status (transitions out from DISPUTED) and clears the challenge timer.

It should be noted that it is NOT possible to file another challenge on a channel that is already disputed. The current challenge must be resolved first.

Additionally, it is possible to close the channel unilaterally by submitting a valid "CLOSE" state (if present) even after a channel was challenged. In such case, the channel will transition to CLOSED status immediately, transferring out all funds to the User and the Node according to amounts agreed about in the CLOSE state.

### Challenge Finality

After the challenge period expires without being resolved, the disputed state becomes **final**. However, a separate **close call** is still required to release the channel's locked funds. Such close call does not require any state to be submitted alongside, only the id of a channel, and can be invoked by anyone.

## Close Operation

A close releases the channel's locked funds and terminates the channel lifecycle.

Two paths exist:

**Cooperative close** — a participant submits a state with the CLOSE intent, signed by all participants. The blockchain layer validates that amounts from the allocations are moved to the respective net flows (basically, it is a withdrawal operation). It should be noted that it is not possible to close an already CLOSED or MIGRATED_OUT channel.

**Unilateral close** — after a challenge period has expired, any party MAY call close without additional signatures. The blockchain layer releases assets according to the last enforced state's allocations (UserAllocation to the user, NodeAllocation to the node).

In both cases, the channel's locked funds are set to zero and the channel lifecycle ends.

## Enforcement Validation

The blockchain layer applies the following common validation rules when processing any enforcement operation:

1. The submitted state MUST reference the correct channel identifier
2. The home ledger chain identifier MUST match the current blockchain
3. The state version MUST be strictly greater than the currently recorded version
4. All required signatures MUST be present and valid
5. The approved signature validation modes MUST be respected
6. The ledger invariant MUST hold: UserAllocation + NodeAllocation == UserNetFlow + NodeNetFlow
7. The resulting locked funds (previous locked funds plus user and node fund deltas) MUST be non-negative
8. Unless the channel is being closed, the sum of allocations MUST equal the resulting locked funds
9. The node MUST have sufficient available funds in its vault when required to lock additional assets

## Escrow and Migration Enforcement

Cross-chain transitions are enforced through dedicated operations on the blockchain layer. The detailed escrow model is described in [Cross-Chain and Assets](cross-chain-and-assets.md). The following summarizes the on-chain effects:

| Operation                  | On-Chain Effect                                                             |
| -------------------------- | --------------------------------------------------------------------------- |
| Escrow Deposit Initiate    | On home chain: state updated, node funds adjusted. On non-home chain: escrow record created, user funds locked. |
| Escrow Deposit Finalize    | On home chain: state updated, user allocation increased. On home chain: state updated, node funds adjusted. On non-home chain: escrow record created, user funds locked, automatic release to the Node timer started. |
| Escrow Withdrawal Initiate | On home chain: state updated. On non-home chain: escrow record created, node funds locked from vault. |
| Escrow Withdrawal Finalize | On home chain: state updated, user allocation decreased. On non-home chain: escrowed funds released to user. |
| Migration Initiate         | On old home chain: state updated. On new home chain: channel created with migrating-in status, node funds locked. |
| Migration Finalize         | On new home chain: channel transitions to operating. On old home chain: all locked funds released, channel marked as migrated out. |

## Failure Conditions

Enforcement MAY fail in the following situations:

- **Invalid signatures** — one or more signatures cannot be verified
- **Stale version** — the submitted state version is not greater than the current on-chain version
- **Inconsistent allocations** — the ledger invariant is violated or resulting locked funds would be negative
- **Allocation-locked-funds mismatch** — the sum of allocations does not equal the expected locked funds (except during close)
- **Unknown channel** — the channel identifier does not correspond to a registered channel (except for channel creation)
- **Insufficient node funds** — the node's vault does not have enough assets to cover required fund locking
- **Invalid intent** — the transition intent does not match the expected operation
- **Chain mismatch** — the home / non-home ledger chain identifier does not match the current blockchain during home-chain / escrow operations
- **Incorrect channel status** — the operation is not permitted in the channel's current status (e.g. challenging an already challenged channel)

---

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)
