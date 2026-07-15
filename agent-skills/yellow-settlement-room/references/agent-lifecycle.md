# Multiparty Settlement Room — Runnable Lifecycle

Real code, real method names, grounded in the official `@yellow-org/sdk` example at `github.com/layer-3/docs`, `examples/nitrolite-v1-lifecycle`. That example runs a two-party lifecycle against the Yellow sandbox; the flow below is the same API generalised to three participants. Nothing here is invented - every method appears in the official example.

**Key separation.** In a real agent-to-agent deployment each agent runs its own client with its own key, in its own process, and signs only its own part. The three-signers-in-one-process code below co-locates keys purely for a readable local run or a custodial server; do not ship it that way. See the skill's `Roles and key separation` section.

## Setup

ESM project, Node 20+.

```jsonc
// package.json
{ "type": "module",
  "dependencies": { "@yellow-org/sdk": "^1", "decimal.js": "^10", "viem": "^2" } }
```

Environment per participant: its own key, an RPC URL, and the Nitronode WS URL (`wss://nitronode-sandbox.yellow.org/v1/ws` for sandbox). Only participants that DEPOSIT also need a funded wallet (Sepolia gas + the test asset); non-depositing participants need neither gas nor the asset.

## 0. Prerequisite: a funded account for each depositor

A session opens with zero allocations, so not every participant needs funds. Only a participant that makes a deposit needs a funded account balance at Yellow (an account is backed by an on-chain channel; treat it as the participant's balance). Funding is done once, on-chain, and is a hard prerequisite for a depositor - `submitAppSessionDeposit` fails on an unfunded account. The funding calls, from the official example:

```ts
await client.setHomeBlockchain(asset, chainId);
await client.approveToken(chainId, asset, amount);
const depositState = await client.deposit(chainId, asset, amount);
const txHash = await client.checkpoint(asset);        // finalizes on-chain
```

For the exact funding sequence and any Node-specific transaction setup, follow the official `nitrolite-v1-lifecycle` example and the quickstart.

**This skill assumes each depositor is already funded.** It never funds accounts. Check a depositor's balance before it deposits:

```ts
const balances = await client.getBalances(wallet);    // account balances at Yellow, per asset
// if the asset balance is below what this agent must deposit, stop and report it.
```

## 1. Connect (one client per participant)

```ts
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';

async function connect(pk: `0x${string}`, wsURL: string, chainId: bigint, rpcURL: string) {
  const s = createSigners(pk);
  return Client.create(wsURL, s.stateSigner, s.txSigner, withBlockchainRPC(chainId, rpcURL));
}
```

## 2. Open a three-party room

```ts
import {
  AppSessionWalletSignerV1, EthereumMsgSigner,
  packCreateAppSessionRequestV1, type AppDefinitionV1,
} from '@yellow-org/sdk';

// LOCAL TEST ONLY: co-locating pkA, pkB, pkC in one process is a shortcut for a
// readable local run or a custodial server. In a real agent-to-agent deployment
// each agent builds ONLY its own signer, in its own process, from its own key,
// and signs only its own part.
const sign = {
  A: new AppSessionWalletSignerV1(new EthereumMsgSigner(pkA)),
  B: new AppSessionWalletSignerV1(new EthereumMsgSigner(pkB)),
  C: new AppSessionWalletSignerV1(new EthereumMsgSigner(pkC)),
};

const definition: AppDefinitionV1 = {
  applicationId: `room-${Date.now().toString(36)}`,   // ^[a-z0-9_-]{1,66}$
  participants: [
    { walletAddress: addrA, signatureWeight: 1 },
    { walletAddress: addrB, signatureWeight: 1 },
    { walletAddress: addrC, signatureWeight: 1 },
  ],
  quorum: 3,                                            // unanimous: the safe default
  nonce: BigInt(Date.now()) * 1_000_000n + BigInt(Math.floor(Math.random() * 1_000_000)),
};

const sessionData = JSON.stringify({ intent: 'init' });
const createPayload = packCreateAppSessionRequestV1(definition, sessionData);
const created = await clientA.createAppSession(definition, sessionData, [
  await sign.A.signMessage(createPayload),
  await sign.B.signMessage(createPayload),
  await sign.C.signMessage(createPayload),             // creation must meet quorum
]);
const appSessionId = created.appSessionId;
```

## 3. Read live state (before every update)

```ts
async function getSession(client, appSessionId: string) {
  const { sessions } = await client.getAppSessions({ appSessionId });
  const s = sessions[0];
  if (!s) throw new Error('session not found');
  return s;   // s.version, s.isClosed, s.allocations
}
```

## 4. Deposit funds into the room

```ts
import { AppStateUpdateIntent, packAppStateUpdateV1, type AppStateUpdateV1 } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';   // named import: default is not constructable under NodeNext

let session = await getSession(clientA, appSessionId);
const clientBudget = new Decimal('10');

// Only the DEPOSITING participant (the client) appears in allocations; do not
// add zero-value entries for the others.
const deposit: AppStateUpdateV1 = {
  appSessionId, intent: AppStateUpdateIntent.Deposit, version: session.version + 1n,
  allocations: [
    { participant: addrA, asset, amount: clientBudget },
  ],
  sessionData: JSON.stringify({ intent: 'fund' }),
};
const dp = packAppStateUpdateV1(deposit);
// The deposit state still needs signatures meeting quorum; each agent signs the
// hash with its own key (shown together here only for a local run).
await clientA.submitAppSessionDeposit(
  deposit,
  [await sign.A.signMessage(dp), await sign.B.signMessage(dp), await sign.C.signMessage(dp)],
  asset, clientBudget,                                  // amount == deposit allocation total
);
```

## 5. Operate: reallocate among the three, off-chain

The client pays two providers out of the pooled budget. Per-asset totals stay constant; every non-zero allocation is restated.

```ts
session = await getSession(clientA, appSessionId);      // re-read for live version
const payout: AppStateUpdateV1 = {
  appSessionId, intent: AppStateUpdateIntent.Operate, version: session.version + 1n,
  allocations: [
    { participant: addrA, asset, amount: new Decimal('4') },   // client keeps 4
    { participant: addrB, asset, amount: new Decimal('4') },   // provider B earns 4
    { participant: addrC, asset, amount: new Decimal('2') },   // provider C earns 2
  ],
  sessionData: JSON.stringify({ round: 'final-split' }),
};
const op = packAppStateUpdateV1(payout);
await clientA.submitAppState(payout, [
  await sign.A.signMessage(op), await sign.B.signMessage(op), await sign.C.signMessage(op),
]);
```

Repeat `operate` as many times as the work has milestones - all off-chain, no gas per step. That repetition, among N parties, is the capability a bilateral rail cannot express.

## 6. Close: release the final split

```ts
session = await getSession(clientA, appSessionId);
const close: AppStateUpdateV1 = {
  appSessionId, intent: AppStateUpdateIntent.Close, version: session.version + 1n,
  allocations: session.allocations,                     // must restate current state exactly
  sessionData: JSON.stringify({ intent: 'close' }),
};
const cp = packAppStateUpdateV1(close);
await clientA.submitAppState(close, [
  await sign.A.signMessage(cp), await sign.B.signMessage(cp), await sign.C.signMessage(cp),
]);
// allocations release back to each participant's channel, withdrawable on-chain.
```

## Gotchas

- ESM only: `"type": "module"`.
- The funding prerequisite (Step 0) applies to depositors only; `submitAppSessionDeposit` fails on an unfunded account (sandbox error: `no channel state to advance`). Check `getBalances(wallet)` first. Non-depositing participants need no funds.
- Distributed keys: in a real deployment each agent runs its own client and signs only its own part with its own key. The single-process code here is a readable local run, not a deployment pattern.
- `intent` is a number on the wire (`Operate=0, Deposit=1, Withdraw=2, Close=3`). Docs that show strings are wrong.
- Creation must meet quorum: multi-party create needs multiple signatures, not one.
- `version` must be exactly `session.version + 1n`; re-read live before signing and retry on collision, re-collecting signatures.
- `Operate` restates every non-zero allocation with per-asset totals unchanged; `Close` restates the current allocation exactly.
- Read decimals from `client.getAssets(chainId)` rather than hardcoding; the same symbol differs per chain. `yusd` is sandbox; production uses `usdc`/`usdt`.
- Weights: equal + quorum = sum is unanimous and safe. A subset summing to quorum can settle without the rest - fine for a trusted operator, dangerous between adversaries.
