---
name: yellow-sdk-api
description: |
  The `@erc7824/nitrolite` TypeScript SDK surface — **every** exported builder, parser, class, and signer, with real names verified against the v0.5.3 tarball. Covers the four groups (RPC message builders, response parsers, crypto/signer helpers, on-chain `NitroliteClient`), the v0.5.3 → `@yellow-org/sdk` v1.x migration status, and when to skip the SDK entirely. Use when: starting a new client from scratch with the upstream SDK, migrating from an older nitrolite integration, deciding SDK vs raw Nitro RPC, or picking the right helper for a specific task.
version: 3.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3 (frozen) / @yellow-org/sdk@^1.2.0 (active)"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://www.npmjs.com/package/@erc7824/nitrolite
  - https://github.com/erc7824/nitrolite
---

# Yellow Nitrolite SDK — TypeScript API (v0.5.3 authoritative)

> **Version status**: `@erc7824/nitrolite@0.5.3` is the last release on the
> `erc7824` org and is now **frozen**. Active development has moved to
> `@yellow-org/sdk` (v1.2.0+). The migration shim `@yellow-org/sdk-compat`
> preserves the v0.5.3 surface but several helpers are **noop stubs**, not
> full re-implementations. New integrations should target `@yellow-org/sdk`
> directly — see the `yellow-sdk-v1` skill.

This skill documents v0.5.3 exhaustively because many existing integrations
are still pinned to this version. For new work, prefer v1.

## Install

```bash
npm install @erc7824/nitrolite viem
```

## Shape — four export groups

v0.5.3 is **not a `Client` class**. You bring your own WebSocket; the SDK
provides:

1. **RPC message builders** — `createXxxMessage` functions producing signed
   wire bytes for each Nitro RPC method. 36 functions.
2. **Response parsers** — `parseXxxResponse` functions that type-narrow
   inbound JSON into discriminated unions. 33 functions. **Main entry
   point: `parseAnyRPCResponse`** (not `parseRPCResponse` — that name does
   NOT exist).
3. **Signers + crypto** — `createEIP712AuthMessageSigner`,
   `createECDSAMessageSigner`, `NitroliteRPC` class,
   `WalletStateSigner` / `SessionKeyStateSigner`.
4. **On-chain `NitroliteClient`** — a viem-backed typed wrapper around the
   Custody contract (submits transactions, not RPC messages).

## Minimal auth flow

```ts
import {
  createAuthRequestMessage,
  createAuthVerifyMessageFromChallenge,
  createEIP712AuthMessageSigner,
  parseAnyRPCResponse,
} from '@erc7824/nitrolite';

const ws = new WebSocket('wss://clearnet-sandbox.yellow.com/ws');

ws.send(await createAuthRequestMessage({
  address: ownerAddress,
  session_key: sessionKeyAddress,
  application: 'my-app',
  allowances: [{ asset: 'usdc', amount: '1000' }],
  expires_at: Date.now() + 86_400_000,
}));

ws.onmessage = async (ev) => {
  const msg = parseAnyRPCResponse(ev.data);
  if (msg.type === 'auth_challenge') {
    const signer = createEIP712AuthMessageSigner(
      walletClient, { /* policy */ }, eip712AuthDomain,
    );
    ws.send(await createAuthVerifyMessageFromChallenge(signer, msg.challenge_message));
  }
};
```

## Common mistakes

| Wrong | Right |
|---|---|
| `parseRPCResponse` | **`parseAnyRPCResponse`** |
| `getAuthDomain()` function | `EIP712AuthDomain` **type** only — build the domain literal yourself |
| bare `getSessionKeys` / `revokeSessionKey` | `createGetSessionKeysMessage` / `createRevokeSessionKeyMessage` |
| bare `createChannel` (expecting RPC) | RPC builder is `createCreateChannelMessage`; `NitroliteClient.createChannel` is the on-chain method |

See `reference.md` for the full export catalogue (all 36 builders, 33
parsers, and `NitroliteClient` methods).

## Skip the SDK when

- Non-TS runtime (Go, Rust, Python) — write raw Nitro RPC per `yellow-nitro-rpc`
- Need byte-exact deterministic output (tests, signature fixtures)
- Hardware-wallet or server-side custody with bespoke signing
- SDK version is older than your ClearNode — v0.5.3 is **frozen**, newer
  ClearNode features won't land here

Use the SDK when you're in TS with viem, want typed discriminated unions
from `parseAnyRPCResponse`, or need EIP-1271/6492 smart-wallet wrapping.

## Navigation Guide

### When to read supporting files

**reference.md** — read when you need:

- Full catalogue of all 36 `createXxxMessage` builders by category
- All ~30 `parseXxxResponse` parsers including notification parsers (`bu`/`cu`/`tr`/`asu`)
- `NitroliteRPC` class statics (`createRequest`, `signRequestMessage`, `verifySingleSignature`)
- `NitroliteClient` on-chain method catalogue (deposit/withdrawal/createChannel/checkpointChannel/challengeChannel)
- Signer helper details (`WalletStateSigner`, `SessionKeyStateSigner`)
- End-to-end auth→transfer example
- Full v0.5.3 → v1 migration notes

## Related

- `yellow-sdk-v1` — the ACTIVE SDK; prefer this for new work
- `yellow-nitro-rpc` — wire format the builders produce
- `yellow-clearnode-auth` — Policy signing details
- `yellow-custody-contract` — what `NitroliteClient` submits on-chain
