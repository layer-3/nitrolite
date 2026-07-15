---
name: yellow-settlement-room
description: Yellow Network Protocol app sessions for AI agents - a shared room where several agents pool funds, reallocate off-chain at machine speed, and settle one final split. Use for multiparty settlement among agents. Covers connecting, the funded-account prerequisite, creating a session with participants + weights + quorum, deposit, operate, withdraw, close, and the trust boundary. Grounded in the official @yellow-org/sdk lifecycle example.
---

# Yellow Settlement Room

Use this skill when several AI agents need to hold funds together and settle one outcome: open a shared session, each agent's balance is tracked inside it, they reallocate between themselves off-chain with no gas per step, and they co-sign the final split.

**The session holds N agents, not two.** A payment rail moves value from one payer to one payee; a swarm of agents settling over a rail needs a separate escrow per pair. One session settles all of them at once: one object, one deposit per agent, one final allocation. That is the shape to reach for when more than two agents have a stake in the same outcome.

This skill covers the **app-session (virtual) layer** only. It operates on funds that are *already in an account balance at Yellow*. A session opens with zero allocations, so not every participant needs funds - only a participant that makes a deposit does. Getting funds into an account is a one-time on-chain step, out of scope here - see `## Prerequisite`. This skill never funds accounts; before a participant deposits, check its balance with `client.getBalances(wallet)`, and if it is short, stop and report it.

Package: `@yellow-org/sdk` (v1). A complete runnable reference is the official example at `github.com/layer-3/docs`, `examples/nitrolite-v1-lifecycle`; the flow below matches it exactly, generalised from two participants to N. For method lookups, the docs MCP: `npx -y @yellow-org/sdk-mcp@^1`.

## Core Model

```text
Each agent runs its own client with its own key (backend: private-key based)
  -> agents open one app session: N participants, signature weights, quorum
     (opens with ZERO allocations; nobody needs funds yet)
  -> a depositing agent commits its OWN funds into the session
     (only depositors need a funded account; others can hold zero)
  -> agents reallocate between themselves (operate), each update co-signed to quorum
  -> withdraw / close: the final split releases back to channels, withdrawable on-chain
```

The session is an off-chain ledger hosted by the Yellow node. Its guarantee is signature-based: **no agent's allocation changes without signatures meeting the quorum.** It is not a trustless escrow; read `## Trust Boundary` before sizing exposure.

## Design Rules

- **One agent, one key, one client.** Each agent signs only with its own key, from its own process. A backend service is private-key based; a frontend is wallet based. You cannot take several agents' private keys into one client. Holding multiple participants' keys in one process is valid only as a local test or an explicitly custodial server. See `## Roles and key separation`.
- **Never fund accounts from this skill.** Funding happens once, on-chain, before a session. Only a participant that deposits needs funds; check its balance first with `client.getBalances(wallet)` and if it is short, stop and report the shortfall. Do not attempt `deposit`, `approveToken`, or `transfer` to self-fund.
- **Never call an allocation "locked on-chain and enforceable."** Funds are committed out of a channel and governed by quorum, so no counterparty agent can take them - but releasing them still needs the node to co-sign. State both halves.
- **Default to equal weights and unanimous quorum.** Only give a subset of agents combined weight >= quorum when that subset is intentionally trusted. See `## Weights and Quorum`.

## Prerequisite: a funded account for each depositor

A session is created with zero allocations, so not every participant needs funds. Only a participant that makes a deposit needs a funded account balance at Yellow for the asset. (An account is backed by an on-chain state channel, but you can treat it as the participant's balance.) That balance is the ceiling on what a depositor can commit and the most it can lose. Check a depositor's balance before it deposits:

```ts
const balances = await client.getBalances(wallet);   // account balances at Yellow, per asset
```

Funding is a one-time on-chain operation, done once before any session and out of scope here. The funding calls, for reference, are:

```ts
await client.setHomeBlockchain(asset, chainId);
await client.approveToken(chainId, asset, amount);
await client.deposit(chainId, asset, amount);
await client.checkpoint(asset);              // finalizes the deposit on-chain
```

For the exact funding sequence and any Node-specific transaction setup, see the quickstart at `docs.yellow.org/nitrolite/build/getting-started/quickstart` and the official `nitrolite-v1-lifecycle` example.

## Connect

```ts
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';

const signers = createSigners(privateKey);       // 0x-prefixed 32-byte hex
const client = await Client.create(
  wsURL,                                          // sandbox: wss://nitronode-sandbox.yellow.org/v1/ws
  signers.stateSigner, signers.txSigner,
  withBlockchainRPC(chainId, rpcURL),
);
```

There is no login handshake; authorization is per-call, from the signatures inside each payload. **Each agent runs its own client with its own key, in its own process.** The examples below show several signers together for readability; in a real agent-to-agent deployment each agent constructs only its own signer and signs only its own part.

## Roles and key separation

Agent-to-agent means the keys are distributed. You cannot take several agents' private keys into one client. Model these roles:

- **Each agent** holds its own key and runs its own client. A backend agent is private-key based; a frontend agent is wallet based. It signs only its own part of any state, in its own process.
- **Proposer** (an agent, or a coordinating server): builds a state update, packs the hash, and asks each required signer to sign it.
- **Signers** (the agents whose combined weight must meet quorum): each signs the packed hash with its own key and returns the signature.
- **Submitter** (usually the proposer): collects the signatures and calls the Nitronode (`createAppSession` / `submitAppSessionDeposit` / `submitAppState`).

A common topology: a Nitronode, agents connecting to a coordinating server (an app), and agents transacting agent-to-agent through shared sessions. The single-process code below co-locates keys only for readability and local testing; do not ship it that way.

## Open the session

```ts
import {
  AppSessionWalletSignerV1, EthereumMsgSigner,
  packCreateAppSessionRequestV1, type AppDefinitionV1,
} from '@yellow-org/sdk';

// One session signer per participant, a plain wallet signer (type 0xa1).
// LOCAL TEST ONLY: holding pkA, pkB, pkC in one process is a shortcut for a
// smoke test or a custodial server. In a real deployment each agent builds ONLY
// its own signer, in its own process, from its own key (see Roles and key
// separation above). Session keys are an optional friction-reducer, omitted here.
const signerA = new AppSessionWalletSignerV1(new EthereumMsgSigner(pkA));
const signerB = new AppSessionWalletSignerV1(new EthereumMsgSigner(pkB));
const signerC = new AppSessionWalletSignerV1(new EthereumMsgSigner(pkC));

const definition: AppDefinitionV1 = {
  applicationId: appId,                           // ^[a-z0-9_-]{1,66}$
  participants: [
    { walletAddress: addrA, signatureWeight: 1 },
    { walletAddress: addrB, signatureWeight: 1 },
    { walletAddress: addrC, signatureWeight: 1 },
  ],
  quorum: 3,                                       // = sum of weights -> unanimous (the safe default)
  nonce: BigInt(Date.now()) * 1_000_000n + BigInt(Math.floor(Math.random() * 1_000_000)),
};

const createPayload = packCreateAppSessionRequestV1(definition, sessionData);
const created = await client.createAppSession(definition, sessionData, [
  await signerA.signMessage(createPayload),
  await signerB.signMessage(createPayload),
  await signerC.signMessage(createPayload),       // creation must itself meet quorum
]);
// created.appSessionId, created.version, created.status
```

The participant set is **immutable after creation**; no agent can be added later. Creation must meet quorum, so every participant that makes up the quorum co-signs the create request. Optionally `client.registerApp(appId, meta, true)` first; some nodes disable the registry (`apps.v1 group is disabled`) - continue without it.

## Read the live state

```ts
const { sessions } = await client.getAppSessions({ appSessionId });
const session = sessions[0];    // session.version, session.isClosed, session.allocations
```

Read this immediately before signing any update: `version` must be exactly `session.version + 1n`.

## Deposit, operate, withdraw, close

All updates share one shape; `intent` is a **number**, not a string.

```ts
import { AppStateUpdateIntent, packAppStateUpdateV1, type AppStateUpdateV1 } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// DEPOSIT (own endpoint): a depositor commits its OWN funds into the session.
// List ONLY the depositing participant in allocations; do not add zero-value
// entries for participants who are not depositing here.
const deposit: AppStateUpdateV1 = {
  appSessionId, intent: AppStateUpdateIntent.Deposit, version: session.version + 1n,
  allocations: [
    { participant: addrA, asset, amount: new Decimal('10') },
  ],
  sessionData: JSON.stringify({ intent: 'fund' }),
};
const dp = packAppStateUpdateV1(deposit);
// The deposit state still needs signatures meeting quorum. Each agent signs the
// hash with its OWN key in its OWN process; here they are shown together only for
// readability. The submitter gathers the signatures and calls the node.
await client.submitAppSessionDeposit(
  deposit, [await signerA.signMessage(dp), await signerB.signMessage(dp), await signerC.signMessage(dp)],
  asset, new Decimal('10'),                        // amount must equal the deposit allocation total
);
```

```ts
// OPERATE: reallocate between participants. Per-asset totals must stay CONSTANT,
// and every non-zero allocation must be restated (not a delta).
const operate: AppStateUpdateV1 = {
  appSessionId, intent: AppStateUpdateIntent.Operate, version: /* live */ session.version + 1n,
  allocations: [
    { participant: addrA, asset, amount: new Decimal('4') },
    { participant: addrB, asset, amount: new Decimal('4') },
    { participant: addrC, asset, amount: new Decimal('2') },
  ],
  sessionData: JSON.stringify({ round: 'payout' }),
};
const op = packAppStateUpdateV1(operate);
await client.submitAppState(operate, [ /* signatures summing to quorum */
  await signerA.signMessage(op), await signerB.signMessage(op), await signerC.signMessage(op),
]);
```

`Withdraw` (intent 2) may only decrease allocations and releases to channels. `Close` (intent 3) must restate the current allocation exactly, releases everything, and is terminal - never close while work or a review is outstanding. Both use `submitAppState` the same way.

**Signatures are collected across agents, off the wire.** The proposer builds the state and hash; each agent signs that hash with its own key in its own process; the submitter gathers the signatures until summed weight meets quorum and calls the node. The protocol provides no transport for this exchange - it is the caller's responsibility. Duplicate signers count once.

## Weights and Quorum

`signatureWeight` per participant, `quorum` = the weight threshold a state needs to be valid.

- **Equal weights, quorum = sum** -> unanimous. The safe default. Use it unless you have a specific reason not to.
- A subset of agents whose weights sum to >= quorum can settle **without** the others. That is a feature for a trusted operator (e.g. a client that funds and controls a swarm) and a hazard between adversaries: in a client/orchestrator/worker room at 1/1/1 quorum 2, the orchestrator and worker together can sign a split that pays the worker and zeroes the client. A single scalar quorum cannot protect every party's balance at N > 2. If the parties mutually distrust, give each depositor a blocking stake, or keep them in separate sessions.

State which regime a session is in when you design it.

## Trust Boundary

State this before any agent puts value at risk. Do not soften it.

- **No agent in the session can take another's allocation.** Every change needs signatures meeting quorum. This is the real, strong guarantee.
- **The Yellow node is trusted for liveness and honest relay.** Sessions are off-chain; funds become enforceable on-chain only once released back to a channel as a node-co-signed state. If the node will not co-sign, there is no on-chain path out of the session. Yellow's own docs state the protocol is not fully trust-minimized. This is weaker than an on-chain escrow and stronger than a spending-allowance delegation (where the payer can drain the account at any time). Say which you rely on.
- **A session has no dispute mechanism, challenge, or timeout of its own.** If quorum is never reached, funds stay in the session. Cooperative settlement is the supported path today.
- **The deposit is the exposure ceiling.** A participant can lose at most what it committed into the session.

## Composes with

- **GenLayer** for subjective disputes: a session cannot judge whether work was acceptable; a jury can. The intended composition is verdict-informed cooperative settlement (parties agree up front to sign the allocation a verdict dictates). Adjudicator-enforced settlement against a non-cooperative party is roadmap, not shipped - do not place already-contested funds in a session.
- **Alkahest** (`vendored/arkhai/alkahest-user`) for a bilateral, one-shot deal that needs trustless on-chain escrow with a reclaim timeout. Reach for a settlement room when the deal is *multiparty and many-update*, which a single bilateral escrow cannot express.

## Failure Cases

- Version race: `version` must be exactly `session.version + 1n`. Concurrent signers collide and one update fails cleanly (the node serializes). Re-read the live version, re-collect signatures, retry. Most common friction in a busy room.
- Quorum never reached: permanent deadlock, no timeout. Prevent by design (cooperative weights); for adversarial bilateral deals use Alkahest.
- Unfunded participant: `submitAppSessionDeposit` fails if the account is not funded (on the sandbox the node error reads `no channel state to advance`). This is the prerequisite, not a bug - check `getBalances(wallet)` first and report a shortfall.
- `Operate` that drops a non-zero allocation or whose per-asset total drifts: rejected.
- `Close` while work or a review is outstanding: terminal and unrecoverable.
- Node fails to co-sign a release: surface as a stuck session; do not retry silently or report success.

## Output Checklist

1. Which agents are in the session, and whether they cooperate or mutually distrust.
2. Participant set with weights and quorum, and the arithmetic showing no coalition can rob a party.
3. Each depositor's funded balance confirmed via `getBalances(wallet)`, and its deposit as the max it can lose. Non-depositing participants need no funds.
4. The happy path: open, deposit, operate, (withdraw), close - and the disagreement path, or an explicit statement that the room deadlocks.
5. The trust disclosure from `## Trust Boundary`, in plain language, before any value moves.

## References

- `references/agent-lifecycle.md` - the full multiparty flow as runnable code, matching the official `nitrolite-v1-lifecycle` example.
