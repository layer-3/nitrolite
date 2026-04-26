---
name: yellow-clearnode-auth
description: |
  The 3-step ClearNode authentication handshake — auth_request → auth_challenge → auth_verify — with the exact field names, EIP-712 Policy signing, session-key setup, JWT reuse, and common failure modes. Use when: a fresh ClearNode connection will not authenticate, you are porting an old client that used `wallet`/`participant`/`app_name`, you need to sign the Policy, you want to reuse a JWT across reconnects, or the server rejects your signature.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/authentication
  - https://docs.yellow.org/docs/0.5.x/guides/migration-guide
---

# Yellow ClearNode — Authentication

The authentication handshake proves the session key belongs to the wallet that
owns the funds. Without it, every subsequent RPC (transfer, channels, app
sessions) is rejected. Spec:
<https://docs.yellow.org/docs/protocol/off-chain/authentication>.

## The 3 steps

```text
Client                                ClearNode
  │  1. auth_request                    │
  │  { address, session_key,            │
  │    application, allowances,         │
  │    scope?, expires_at }             │
  │────────────────────────────────────▶│
  │                                     │
  │  2. auth_challenge                  │
  │  { challenge_message: <uuid> }      │
  │◀────────────────────────────────────│
  │                                     │
  │  3. auth_verify                     │
  │  { challenge }                      │
  │  sig: [EIP-712 Policy sig]          │
  │────────────────────────────────────▶│
  │                                     │
  │  response                           │
  │  { jwt }                            │
  │◀────────────────────────────────────│
```

After step 3 the client is authenticated for the connection and holds a JWT for
reconnects.

## Field names (commonly wrong)

| Correct | ❌ Do NOT use (old docs / outdated clients) |
|---|---|
| `address` | `wallet` |
| `session_key` | `participant` |
| `application` | `app_name` |
| `expires_at` (Unix **ms**) | `expire` (seconds) |
| response `challenge_message` | `challenge` |

If you see any of the left column in a codebase, it's using the current spec.
If you see the right column, it's stale — fix before anything else.

## Step 1 — `auth_request` params

```json
{
  "address":     "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2",
  "session_key": "0x9876543210fedcba9876543210fedcba98765432",
  "application": "your-app-id",
  "allowances": [
    { "asset": "usdc", "amount": "1000.0" },
    { "asset": "eth",  "amount": "0.5" }
  ],
  "scope":      "transfer,app.create",          // optional, comma-separated
  "expires_at": 1762417328123                    // required, ms
}
```

`auth_request` is unsigned (public). Server validates addresses, checks the
session key is not already registered, and replies `auth_challenge`.

> **SDK compatibility note.** `@erc7824/nitrolite@0.5.3`'s
> `createAuthRequestMessage` helper still emits both `application` **and**
> a redundant `app_name`, and uses `expire` (a seconds **string**) rather
> than `expires_at` (ms). The sandbox accepts either shape so existing SDK
> users don't break, but the canonical wire format documented here — `expires_at`
> in ms, `application` only — is what fresh integrations should emit.

### `application` — security footgun

If you **omit `application`** (or pass the literal string `"clearnode"`),
the session is scoped to the ClearNode itself — **root access**, allowance
enforcement disabled. Only use this for operator tooling you fully control.
For any app-level session, always set `application` to your app id and
pass `allowances` to cap spend.

### `expires_at` unit — docs contradiction

The authentication spec page says `expires_at` is "Unix timestamp
(milliseconds) when the session key expires". The migration-guide code
sample computes it in **seconds** (`Math.floor(Date.now()/1000) + 7*24*3600`).
A 13-digit value is ms, a 10-digit value is seconds — they're unambiguously
distinguishable.

Safe bet: **pass milliseconds** (match the spec page), and if your ClearNode
rejects with "Challenge expired" unexpectedly, try seconds. Test against
your target deployment.

## Step 2 — `auth_challenge` response

```json
{ "challenge_message": "550e8400-e29b-41d4-a716-446655440000" }
```

A UUID. Store it; you need it in step 3 and in the EIP-712 payload.

## Step 3 — `auth_verify` with EIP-712 Policy signature

Params object is minimal — just the challenge:

```json
{ "challenge": "550e8400-e29b-41d4-a716-446655440000" }
```

The **signature goes in the envelope `sig` array**, not in params. It is
EIP-712 typed data over the `Policy` type, signed by the **main wallet**
(`address`, not the session key):

```ts
const typedData = {
  types: {
    EIP712Domain: [{ name: 'name', type: 'string' }],
    Policy: [
      { name: 'challenge',   type: 'string'  },
      { name: 'scope',       type: 'string'  },
      { name: 'wallet',      type: 'address' },
      { name: 'session_key', type: 'address' },
      { name: 'expires_at',  type: 'uint64'  },
      { name: 'allowances',  type: 'Allowance[]' },
    ],
    Allowance: [
      { name: 'asset',  type: 'string' },
      { name: 'amount', type: 'string' },
    ],
  },
  primaryType: 'Policy',
  domain: { name: 'your-app-id' },            // MUST match `application`
  message: {
    challenge:   challengeMessage,              // from auth_challenge
    scope:       'transfer,app.create',         // MUST match auth_request
    wallet:      mainWalletAddress,
    session_key: sessionKeyAddress,
    expires_at:  1762417328123,                 // MUST match auth_request
    allowances:  [{ asset: 'usdc', amount: '1000.0' }],
  },
};

// viem
import { createWalletClient, custom } from 'viem';
const client = createWalletClient({ chain, transport: custom(provider) });
const sig = await client.signTypedData({
  account: mainWalletAddress,
  ...typedData,
});
```

Server response:

```json
{
  "address":     "0x742d35Cc...",          // main wallet (echoed)
  "session_key": "0x9876...",              // session key (echoed)
  "jwt_token":   "eyJhbGciOi...",          // reuse on reconnect
  "success":     true
}
```

Save the `jwt_token` — you can reuse it on reconnect. (Common mistake:
clients look for a `jwt` field. The actual field name is **`jwt_token`**.)

## JWT reuse on reconnect

After your first `auth_verify`, subsequent WebSocket reconnects can skip
signing by presenting the JWT:

```json
{ "req": [1, "auth_verify", { "jwt": "<existing_jwt>" }, <ms>], "sig": [] }
```

No sig needed. Server validates and restores the authenticated session.

## Session key setup

Session keys are locally generated keypairs; the private key **never leaves
the client**. They sign day-to-day operations so the main wallet's private key
stays cold.

```ts
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';

const sessionPriv = generatePrivateKey();
const sessionKey = privateKeyToAccount(sessionPriv);
// sessionKey.address → pass as `session_key` in auth_request
// sessionKey.signMessage  → sign subsequent req arrays
```

Persist `sessionPriv` securely (OS keychain / encrypted storage). Rotate
before `expires_at`.

## Signing day-to-day requests

After auth, the client signs the `req` array of every authenticated method
with the session key (keccak256 of req bytes → ECDSA secp256k1). Wire the
signer in once as a `signRequest(req)` callback on your client.

```ts
await client.send('transfer', {
  destination: '0xB...',
  allocations: [{ asset: 'usdc', amount: '50.0' }],
}, async (req) => sessionKey.signMessage({ message: JSON.stringify(req) }));
```

## Common failures

| Symptom | Cause |
|---|---|
| No response ever | Using old field names; server drops the message silently (some deployments), or timestamp in seconds causes skew rejection |
| `error: invalid signature` | Signing over wrong JSON (re-encoded differently between sign and send) or using session key instead of main wallet for `auth_verify` |
| `error: session key already registered` | You tried to register the same `session_key` twice without logging out. Generate a fresh keypair. |
| `error: expired` | `expires_at` in the past or the server clock is ahead of yours. Use Date.now() + buffer. |
| `error: application mismatch` | EIP-712 `domain.name` doesn't match the `application` you passed in auth_request. They must be identical. |
| Auth works, then every method rejects | Forgetting to sign the `req` array of authenticated methods. Pass a signer to `client.send(...)`. |

## Recommended helper shape

```ts
// The server returns the JWT in a field named `jwt_token`. Match that
// shape in your helpers — using just `jwt` here is the bug the docs above
// flag as a "common mistake".
authenticate(client, config, signRequest): Promise<{ jwt_token: string }>
sendAuthRequest(client, config): Promise<challengeMessage>
sendAuthVerify(client, challenge, signRequest): Promise<{ jwt_token: string }>
```

Keep the `signRequest` callback external — the EIP-712 Policy signer for
`auth_verify` and the keccak256/ECDSA signer for everything else. Decoupling
the signing mechanism lets the same helper serve hardware wallets, session
keys, server-side custodial signers, and browser-injected providers.

## Related

- `yellow-nitro-rpc` — the envelope this flow lives inside
- `yellow-session-keys` — allowances, scopes, rotation
- `yellow-sdk-v1` / `yellow-sdk-api` — typed SDKs that wrap this flow
