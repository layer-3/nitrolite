# Receiver-credit issuance during channel challenge

Status: reviewed, severity downgraded from reported.

## Finding (as reported)

Receiver-credit issuance can create and node-sign a newer state for a receiver whose home channel is already `Challenged`.

- `issueTransferReceiverState()` derives the receiver's next state with `NextState()` and signs whenever the inherited `HomeChannelID` is non-`nil`.
- `issueReleaseReceiverState()` follows the same pattern for app-session release credits.
- Neither path loads the referenced channel or requires `ChannelStatusOpen` before signing.

Code locations confirmed:
- `nitronode/api/channel_v1/handler.go:65-136` — `issueTransferReceiverState`, no `CheckActiveChannel` call before signing.
- `nitronode/api/app_session_v1/handler.go:127-208` — `issueReleaseReceiverState`, same gap. Triggered from `handleWithdrawIntent:409` and `handleCloseIntent:521`.
- `event_handlers/service.go:104-106` — `HandleHomeChannelCheckpointed` flips `Challenged → Open` on checkpoint event.
- `contracts/src/ChannelEngine.sol` `_calculateOperateEffects` — accepts `ChannelStatus.DISPUTED` for `INTENT_OPERATE`, clears `challengeExpireAt`.
- `contracts/src/sigValidators/SessionKeyValidator.sol:70-97` — `validateSignature` checks only ECDSA over `SessionKeyAuthorization` + state. No `expiresAt`, `channelId`, `asset`, or `intent` checks. Off-chain `ValidateChannelSessionKeyForAsset` (`channel_session_key_state.go:172-193`) enforces all three.

Reported claim: open-channel sender can transfer dust to a `Challenged` receiver (or app-session release can credit them), node issues `TransitionTypeTransferReceive` / `TransitionTypeRelease` state, attacker checkpoints onchain (using expired session key signature accepted by onchain validator), challenge resets, channel status returns to `OPERATING`, dispute timer cleared. Repeatable → fund lock / loss.

## Attack vectors considered

### V1 — Expired-session-key replay + challenge reset

Setup: channel `Challenged`; attacker holds expired session-key material; WS feed unauthenticated (any user can read any channel state, per node operator confirmation).

Sequence:
1. Some sender (Carol or attacker via second channel) transfers dust to the challenged-channel user (Alice).
2. Node, without status guard, signs `V'' = NextState(node_head, transfer-receive)` for Alice's channel.
3. Attacker reads `V''` over WS, signs with expired session-key material, calls `checkpointChannel(V'')`. Onchain validator accepts (no expiry check).
4. `_calculateOperateEffects` accepts `DISPUTED`, sets `OPERATING`, clears `challengeExpireAt`. `HandleHomeChannelCheckpointed` mirrors offchain.

Collapse:
- Node head is `V'` = Alice's true latest. `NextState(V', …)` derives `V''` carrying Alice's good allocation + dust. Not derived from the challenged state.
- Onchain head moves `V → V''`. Stale challenge state permanently superseded (`V'' > V`).
- Reset destroys whichever challenge was on the table. Two cases:
  - Challenger was malicious (stale `V`): reset benefits Alice, kills attacker's leverage.
  - Challenger was Alice doing legit forced-exit: requires node down for forced exit to be justified. If node down, node cannot produce `V''`, so attack does not fire. If node alive, Alice's appropriate path is cooperative close, not challenge; her challenge being reset costs gas but no funds.
- Under honest-node trust model, no fund-loss path survives. Worst case = gas griefing on a forced-exit attempt that should not have been needed.

Verdict: collapsed. Low-severity griefing.

### V2 — App-session release variant of V1

Identical primitive, release transition instead of transfer. Bob opens app session with Alice; close/withdraw triggers `issueReleaseReceiverState`; same checkpoint reset.

Collapse: same as V1. Released allocations are bounded to deposited assets; node derives release state from current head. Reset advances onchain head to Alice-favorable state. No extraction.

Verdict: collapsed. Low-severity griefing.

### V3 — Late-credit deadline squeeze, no malicious session key

Setup: Bob challenges Alice's channel with stale `V` at `T`, `challengeExpireAt = T+24h`. Alice holds double-signed `V' > V`. Near expiry, dust transfer to Alice causes node to sign `V'' = V'+1`.

Reported loss: Alice forced to chase `V''`, miss deadline, `V` finalizes.

Collapse:
- `V'` is already double-signed. Alice submits `V'` directly to `checkpointChannel`. Contract accepts (`V' > V`, fully signed). No dependency on `V''`.
- Any tooling that pushes Alice toward `V''` instead of `V'` is a wallet-UX bug, fix in tooling.
- Bob can drip dust to keep producing `V''+k` but Alice ignores them all; her response is `V'`, unchanged.
- Residual: transfer-DoS via version-mismatch churn, a pre-existing acknowledged class — every user can dust every other user with no special access.

Verdict: collapsed. Pre-existing DoS class, no new loss surface.

### V4 — Onchain validator scope bypass (asset / intent)

Setup: Alice's channel uses `SessionKeyValidator` with metadata scoping to asset USDC. Attacker signs a state touching WETH or a non-USDC intent with the same key.

Reported loss: onchain validator does not check `assets[]` / `intents[]` against metadata, so out-of-scope state checkpoints succeed.

Collapse:
- `SessionKeyValidator` registered on channel = channel session key surface (asset-scoped on home channel). App session keys (per app-id / app-session-id) do not enter the `checkpointChannel` signature path.
- Channel-level transitions reachable via this primitive = `transfer-receive` and `release-receive`. Both credit the receiver. State produced by node carries node's head allocation + a credit; node does not produce extractive states under honest operation.
- Cross-asset release requires app session deposited in that asset. Attacker holding USDC-scoped channel session key cannot induce node to release a non-deposited asset.
- Result: scope bypass at onchain layer only enables checkpointing receiver-favorable states. No fund extraction. Collapses into V1/V2 griefing under honest-node model.

Verdict: collapsed. Defense-in-depth value only.

## Assumptions and acknowledgements

Trust model (as stated by node operator team):
- Node is trusted. Compromised node ⇒ all session-key restrictions (expiry, asset, intent) lapse by design. Users accept this. Honest node enforces expiry/asset via `ValidateChannelSessionKeyForAsset`.
- `SessionKeyValidator.validateSignature` is intentionally minimal: cryptographic-only. Off-chain layer is the policy guard. Comment at `SessionKeyValidator.sol:50-53` documents this.

Acknowledged pre-existing classes, not introduced by this finding:
- **Transfer-DoS.** Any user can dust any other user via offchain transfer, producing version-mismatch churn. Generic DoS class, separate hardening track.
- **Challenge-while-node-online is bad practice.** Cooperative close is the correct path when node is responsive. Forced challenge while node alive may be punishable; users are expected to know this.

What still has value:
- **Go status-check fix.** Loading the channel and rejecting `ChannelStatusChallenged` in `issueTransferReceiverState` / `issueReleaseReceiverState` is cheap. Benefits:
  - Removes tooling/UX confusion from node-head advancing past a challenged-channel state.
  - Eliminates the gas-griefing path in V1/V2 by denying the attacker any fresh node-signed bytes during dispute.
  - Strictly hygienic — no legitimate flow advances state while the home channel is in dispute.
- **Onchain scope/expiry enforcement in `SessionKeyValidator`.** Defense-in-depth only under stated trust model. Lower priority. Worth revisiting if partial-compromise or operator-key-rotation lag enters scope.
- **Open question on session-key revocation on live RPC/WS sessions.** Not load-bearing under WS-unauthenticated model (any user reads any state anyway), but tracking for completeness.

Severity revision: reported as critical (fund lock / dispute-bypass / repeated reset). Revised to low (gas-griefing + dispute-flow hygiene gap) under honest-node trust model. Recommendation Go fix retained as cheap hygienic improvement; onchain validator hardening retained as long-term defense-in-depth.
