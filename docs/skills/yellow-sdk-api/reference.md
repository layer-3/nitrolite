# sdk-api — Complete Reference

> Extended documentation. Read the `SKILL.md` quick-start first; this file contains the full details moved out for progressive disclosure.


# Yellow Nitrolite SDK — TypeScript API (v0.5.3 authoritative)

> **Version status**: `@erc7824/nitrolite@0.5.3` is the last release on the
> `erc7824` org and is now **frozen**. Active development has moved to
> `@yellow-org/sdk` (v1.2.0+). A migration shim `@yellow-org/sdk-compat`
> preserves the v0.5.3 surface but several helpers are **noop stubs**, not
> full re-implementations. New integrations should target `@yellow-org/sdk`
> directly if possible. This skill documents **v0.5.3** exhaustively because
> many existing integrations are still pinned to this line.

Install:

```bash
npm install @erc7824/nitrolite viem        # v0.5.3 line (legacy-stable)
# or
npm install @yellow-org/sdk viem           # v1.x line (active)
```

## Shape of the SDK

v0.5.3 is **four groups** of exports:

1. **RPC message builders** — `create…Message` functions that produce the
   signed wire bytes for a specific Nitro RPC method. You send the bytes on
   your own WebSocket.
2. **Response parsers** — `parse…Response` functions that type-narrow
   inbound JSON into discriminated unions.
3. **Signers + crypto** — EIP-712 Policy signer, ECDSA signer, the
   `NitroliteRPC` class with request signing/verification.
4. **On-chain `NitroliteClient`** — a fully typed wrapper around the
   Custody contract (submits transactions via viem).

There is **no single `Client` class in v0.5.3** that manages the WebSocket
for you. You bring your own socket; the SDK builds/parses messages.

## Full v0.5.3 export catalogue

### RPC message builders (37 functions)

```ts
import {
  // Auth
  createAuthRequestMessage,
  createAuthVerifyMessage,
  createAuthVerifyMessageFromChallenge,
  createAuthVerifyMessageWithJWT,

  // Liveness + discovery
  createPingMessage, createPingMessageV2,
  createGetConfigMessage, createGetConfigMessageV2,
  createGetAssetsMessage, createGetAssetsMessageV2,
  createGetUserTagMessage,
  createGetRPCHistoryMessage,

  // Session keys
  createGetSessionKeysMessage,
  createRevokeSessionKeyMessage,
  createCleanupSessionKeyCacheMessage,

  // Balances / ledger
  createGetLedgerBalancesMessage,
  createGetLedgerEntriesMessage, createGetLedgerEntriesMessageV2,
  createGetLedgerTransactionsMessage, createGetLedgerTransactionsMessageV2,

  // Transfers
  createTransferMessage,

  // Channels (RPC-level — produce bytes, do NOT send on-chain tx)
  createCreateChannelMessage,
  createCloseChannelMessage,
  createResizeChannelMessage,
  createGetChannelsMessage, createGetChannelsMessageV2,

  // App sessions
  createAppSessionMessage,           // creates a new session
  createSubmitAppStateMessage,
  createCloseAppSessionMessage,
  createGetAppDefinitionMessage, createGetAppDefinitionMessageV2,
  createGetAppSessionsMessage, createGetAppSessionsMessageV2,
  createApplicationMessage,          // arbitrary signed application payload
} from '@erc7824/nitrolite';
```

### Response parsers (~30 functions)

```ts
import {
  parseAnyRPCResponse,   // ← MAIN entrypoint; use this for discriminated-union typing

  // Auth
  parseAuthRequestResponse,
  parseAuthChallengeResponse,
  parseAuthVerifyResponse,
  parseErrorResponse,

  // Queries
  parseGetConfigResponse,
  parseGetAssetsResponse, parseAssetsResponse,
  parseGetLedgerBalancesResponse,
  parseGetLedgerEntriesResponse,
  parseGetLedgerTransactionsResponse,
  parseGetUserTagResponse,
  parseGetSessionKeysResponse,
  parseGetRPCHistoryResponse,
  parseGetAppDefinitionResponse,
  parseGetAppSessionsResponse,
  parseGetChannelsResponse,

  // Mutations
  parseTransferResponse,
  parseCreateChannelResponse,
  parseResizeChannelResponse,
  parseCloseChannelResponse,
  parseCreateAppSessionResponse,
  parseSubmitAppStateResponse,
  parseCloseAppSessionResponse,
  parseCleanupSessionKeyCacheResponse,

  // Notifications (bu / cu / tr / asu)
  parseBalanceUpdateResponse,           // bu
  parseChannelUpdateResponse,           // cu (single)
  parseChannelsUpdateResponse,          // cu (batch)
  parseTransferNotificationResponse,    // tr
  parseAppSessionUpdateResponse,        // asu

  // Misc
  parsePingResponse, parsePongResponse,
  parseMessageResponse,
} from '@erc7824/nitrolite';
```

**Common mistake — `parseRPCResponse` does NOT exist.** The main entry
point is **`parseAnyRPCResponse`** (note the "Any"). Several tutorials and
older READMEs reference the shorter name — it's either been renamed or was
never exported. Use `parseAnyRPCResponse`.

### Signers + crypto

```ts
import {
  // EIP-712 Policy signer for auth_verify
  createEIP712AuthMessageSigner,

  // ECDSA signer for day-to-day req signing (session key)
  createECDSAMessageSigner,

  // Stateful signers — pre-wired for Custody state hashing
  WalletStateSigner,
  SessionKeyStateSigner,

  // Low-level RPC class
  NitroliteRPC,
} from '@erc7824/nitrolite';

// NitroliteRPC statics:
// NitroliteRPC.createRequest(method, params, id)     → req array
// NitroliteRPC.createAppRequest(...)                 → app-session wrapped req
// NitroliteRPC.signRequestMessage(req, signer)       → { req, sig }
// NitroliteRPC.verifySingleSignature(...)            → boolean
// NitroliteRPC.verifyMultipleSignatures(...)         → boolean
```

**`getAuthDomain()` does NOT exist.** Tutorials sometimes reference it.
The SDK exports the **`EIP712AuthDomain` type** only; you construct the
domain object yourself (name/version/chainId/verifyingContract) and pass it
to `createEIP712AuthMessageSigner`.

### On-chain client (`NitroliteClient`)

Separate from the RPC builders — submits **transactions** via viem to the
Custody contract:

```ts
import { NitroliteClient } from '@erc7824/nitrolite';

const client = new NitroliteClient({
  walletClient, publicClient, chainId,
  custodyAddress, adjudicatorAddress,
});

// Deposit / withdraw (unified balance)
await client.deposit(token, amount);
await client.approveTokens(token, amount);    // ERC-20 pre-approve
await client.withdrawal(token, amount);

// Channels (on-chain) — each takes a single params object
const { channelId, initialState, txHash } = await client.createChannel(params);
await client.depositAndCreateChannel(tokenAddress, depositAmount, params);
await client.resizeChannel(resizeParams);
await client.closeChannel(closeParams);
await client.checkpointChannel(checkpointParams);
await client.challengeChannel(challengeParams);

// Read
await client.getOpenChannels();                       // no args — returns ChannelId[]
await client.getAccountBalance(tokenAddress);         // tokenAddress | tokenAddress[]
await client.getChannelBalance(channelId, tokenAddress);
await client.getChannelData(channelId);
await client.getTokenAllowance(tokenAddress);         // one arg: the ERC-20
await client.getTokenBalance(tokenAddress);
```

**These are chain-level transactions**, not RPC messages. Name collisions
with the RPC builders (`createChannel`, `resizeChannel`, `closeChannel`)
are a common source of confusion — always disambiguate in code comments
which layer you mean.

## End-to-end minimal flow

```ts
import {
  createAuthRequestMessage,
  createAuthVerifyMessageFromChallenge,
  createEIP712AuthMessageSigner,
  createTransferMessage,
  parseAnyRPCResponse,
} from '@erc7824/nitrolite';

const ws = new WebSocket('wss://clearnet-sandbox.yellow.com/ws');

// 1. Auth request
ws.send(await createAuthRequestMessage({
  address: ownerAddress,
  session_key: sessionKeyAddress,
  application: 'my-app',
  allowances: [{ asset: 'usdc', amount: '1000' }],
  expires_at: Date.now() + 86_400_000,   // ms (spec); see auth skill for seconds caveat
}));

// 2. Parse inbound
ws.onmessage = async (ev) => {
  const msg = parseAnyRPCResponse(ev.data);
  if (msg.type === 'auth_challenge') {
    const signer = createEIP712AuthMessageSigner(
      walletClient,
      { application: 'my-app', wallet: ownerAddress, session_key: sessionKeyAddress, /* … */ },
      eip712AuthDomain,
    );
    ws.send(await createAuthVerifyMessageFromChallenge(signer, msg.challenge_message));
  }
};

// 3. After auth, send a transfer
ws.send(await createTransferMessage(ecdsaSigner, {
  destination: recipientAddress,
  allocations: [{ asset: 'usdc', amount: '50.0' }],
}));
```

## `@yellow-org/sdk` v1 — brief preview

If you're starting fresh, v1 gives you a batteries-included `Client`
class that hides the socket:

```ts
import { Client, createSigners } from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(privateKey);
const client = await Client.create(
  'wss://clearnet-sandbox.yellow.com/ws',
  stateSigner, txSigner,
);

await client.deposit({ chainId: 137, token: '0x...', amount: '100' });
await client.transfer({ destination: '0x...', allocations: [...] });
```

v1 RPC methods use **dotted namespaces** (`channels.v1.get_home_channel`,
`user.v1.get_balances`, `node.v1.ping`, etc.) — not the flat snake-case
names used in v0.5.3. Don't mix v0/v1 method names in the same client.

## SDK vs raw RPC — when to skip the SDK

Skip when:
- Non-TS environment (Go, Rust, Python): build the envelope yourself per
  `yellow-nitro-rpc`.
- You need byte-exact deterministic output (tests, signature fixtures).
- You want hardware-wallet or server-side custody signing with full
  control.
- SDK version pinned to an older nitrolite than your ClearNode (v0.5.3
  is frozen — newer ClearNode features won't be in the builders).

Use when:
- Browser / Node app with viem already in the stack.
- You want `parseAnyRPCResponse` to give you typed discriminated unions.
- Smart-contract wallets (EIP-1271/6492) — the SDK handles chain-specific
  signature wrapping.

## Known docs gotchas

- **`parseRPCResponse` vs `parseAnyRPCResponse`** — the tutorials say the
  first, the SDK exports the second. Use the second.
- **No `getAuthDomain()` function** — only the `EIP712AuthDomain` type.
  Build the domain object manually.
- **No bare `getSessionKeys` / `revokeSessionKey`** — use the builders
  `createGetSessionKeysMessage` / `createRevokeSessionKeyMessage` and parse
  the response.
- **`createChannel` name collision** — the RPC builder is
  `createCreateChannelMessage`; the on-chain method is
  `NitroliteClient.createChannel`. They do different things.
- **v0.5.3 is frozen** — bugs in builders/parsers will not be patched.
  Plan a migration to `@yellow-org/sdk` for any new long-lived integration.

## Related

- `yellow-nitro-rpc` — the wire format the builders produce
- `yellow-clearnode-auth` — full 3-step flow including EIP-712 Policy
- `yellow-custody-contract` — what `NitroliteClient` submits on-chain
- `yellow-deposits-withdrawals` — on-chain flows `NitroliteClient` wraps
