---
name: yellow-session-keys
description: |
  Session keys — the hot-wallet delegation model that lets apps sign day-to-day Nitro RPC operations without prompting the main wallet. Covers keypair generation, allowances, scopes, expiry, rotation, revocation, secure storage, and how session keys interact with the auth_request / Policy EIP-712 flow. Use when: building a long-lived agent that should not pop a wallet prompt on every transfer, setting spending limits per app, rotating keys before expiry, or recovering from a compromised session key.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/authentication
---

# Yellow Session Keys

A **session key** is a disposable keypair generated locally by the client. The
main wallet signs a one-time EIP-712 Policy delegating bounded authority to
the session key; from that point until `expires_at`, the session key signs
every authenticated Nitro RPC request instead of the main wallet.

Benefits: no wallet prompts per action, tighter blast radius if compromised
(bounded by allowances + expiry), private key never leaves the app.

## Lifecycle

```
1. Generate session key (local)     ──▶  (sessionKey.address, sessionKey.privKey)
2. Send auth_request with           ──▶  session_key, allowances, scope, expires_at
3. Receive auth_challenge
4. Sign Policy EIP-712 with main wallet  ──▶ sig carries authority
5. Send auth_verify                 ──▶  receives jwt
6. Use session key to sign every subsequent req
7. Before expires_at: rotate (repeat)
8. Revoke if compromised (see below)
```

## Generation

```ts
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';

const privKey = generatePrivateKey();               // 32 random bytes
const sessionKey = privateKeyToAccount(privKey);

// sessionKey.address   → register with ClearNode as `session_key`
// sessionKey.signMessage({ message }) → sign Nitro RPC req arrays
```

Do this **once per application + device**. Don't regenerate per request.

## Allowances

Caps on what the session key is allowed to spend. Array of `{ asset, amount }`.

```json
"allowances": [
  { "asset": "usdc", "amount": "100.0" },
  { "asset": "eth",  "amount": "0.01" }
]
```

- Unit: human-readable decimal strings (e.g., `"100.0"` USDC).
- Enforcement: the ClearNode totals spending by the session key against these
  limits across transfers and app session deposits.
- Once exceeded, further spend calls are rejected until you rotate with a new
  allowance.
- Omit allowances entirely to grant unrestricted spend (not recommended).

## Scopes

Comma-separated operation whitelist:

```json
"scope": "transfer,app.create,app.submit"
```

Common scopes:

| Scope | Allows |
|---|---|
| `transfer` | Direct P2P transfers |
| `app.create` | `create_app_session` |
| `app.submit` | `submit_app_state` |
| `app.close` | `close_app_session` |
| `channel.create` / `channel.close` | Open/close on-chain state channels |
| `query` | Read-only methods (`get_balance`, `get_assets`, …) |

Scopes are an **additional** restriction on top of allowances. A session key
scoped to `query` cannot spend even if allowances are set.

> **v0.5 migration**: `allowance.amount` in the EIP-712 Policy payload
> must be a **decimal string** (`"1000.0"`), not a `number`. This avoids
> `Number` precision loss past `2^53`. Clients passing a raw number will
> have their signatures rejected with `"Invalid signature"`.

## Expiry

Every session key has a hard expiry — a Unix **milliseconds** timestamp at
which the ClearNode refuses further requests signed by that key.

Pick a duration that balances UX vs. blast radius:

| Duration | Good for |
|---|---|
| 1 h | High-value agents, trading bots with tight risk controls |
| 24 h | Default for most apps |
| 7 d | Low-stakes consumer flows, messaging, reads |
| 30 d+ | Long-lived daemons with strong local key storage |

## Storage

The session private key must be stored — losing it means re-running the full
auth flow and nagging the user for a wallet signature again.

- **Browser**: IndexedDB or `localStorage` (acceptable for short expiries) — NEVER `sessionStorage` if you want it to survive reloads. Prefer `crypto.subtle` + a derived key from a user passphrase for anything high-value.
- **Mobile / Electron**: OS keychain (Keychain Services / Credential Manager / libsecret).
- **Node.js**: encrypted at rest with a master key in env/secret manager. Never log.
- **Server-side (custodial)**: KMS-wrapped keys, only decrypt in memory.

## Rotation

Before `expires_at` hits, generate a new session key and re-auth:

```ts
async function rotateIfExpiring(current, bufferSec = 3600) {
  const msLeft = current.expiresAt - Date.now();
  if (msLeft > bufferSec * 1000) return current;

  const newKey = newSessionKey();
  await authenticate(client, {
    address: mainWallet,
    sessionKeyAddress: newKey.address,
    application: APP_ID,
    allowances: current.allowances,
    scope: current.scope,
    expireSeconds: 86_400,
  }, policySigner(mainWallet, newKey));
  await secureStore.set('sessionKey', newKey);
  return newKey;
}
```

Run on a timer or before any spend to ensure you always hold a fresh key.

## Listing active sessions

Use `get_session_keys` to audit live sessions and their remaining
allowances. Response shape (verbatim per spec):

```ts
const { session_keys } = await client.send('get_session_keys', {
  offset: 0, limit: 10, sort: 'desc',  // all optional
});
// session_keys[] → {
//   id:          number,
//   session_key: address,
//   application: string,
//   allowances:  Array<{ asset: string, allowance: string, used: string }>,
//   scope:       string,
//   expires_at:  string,
//   created_at:  string,
// }
```

The clearnode tracks spending per (session_key, asset) pair:

```
remaining = allowance - used
```

Compute it client-side by subtracting `used` from `allowance` per allowance
entry. Once remaining for an asset hits 0, further spends of that asset are
rejected with `"Session key allowance exceeded: <required>, <remaining>"`.

## Revocation

Per the v0.5.0 migration guide, `revoke_session_key` is a real RPC method
with params `{ session_key }`, signed by the **main wallet**:

```ts
await client.send(
  'revoke_session_key',
  { session_key: compromised.address },
  mainWalletSigner,   // EIP-712 sig from the owning wallet
);
```

The Nitrolite SDK (`@erc7824/nitrolite` 0.5.x) wraps this as
`revokeSessionKey({ session_key })` and lists active keys via
`getSessionKeys()`.

Recovery playbook when a key is compromised:

1. **Call `revoke_session_key`** immediately. The key stops working on the
   ClearNode even though local code still holds the bytes.
2. **Rotate** — `auth_request` a fresh keypair so your app can keep working.
3. **Let expiry catch stragglers** — session keys cannot be reactivated
   after `expires_at`; tight expiries cap damage if revocation fails.
4. **Drain allowance as defence-in-depth** — shrink allowances at
   `auth_request` time so a leaked key can only spend a bounded amount.

## Multi-device

Each device gets its own session key. Do NOT share private keys across
devices. The main wallet can delegate to multiple session keys simultaneously;
revoke only the compromised one if a single device is lost.

## Security checklist

- [ ] Session key generated with `crypto.getRandomValues` / viem's
      `generatePrivateKey` (never from a user-chosen string).
- [ ] Allowances set and bounded to expected spend.
- [ ] Scope is the narrowest that still lets the app function.
- [ ] `expires_at` is no longer than needed.
- [ ] Private key stored encrypted at rest.
- [ ] Rotation runs before expiry automatically.
- [ ] Revocation path documented and tested.
- [ ] UI shows "active session key" + expiry so users can audit.
- [ ] Logout clears the session key from storage and revokes server-side.

## Common pitfalls

- **Reusing a session key across applications.** Each app has its own
  `application` field and Policy. A key scoped to one app won't authenticate
  to another — create one per app.
- **Signing auth_verify with the session key.** `auth_verify`'s Policy must
  be signed by the **main wallet**, not the session key. The session key
  signs *everything after* verify.
- **Forgetting to bump expires_at on rotate.** Same epoch = same Policy hash
  = no new authority granted.
- **Backing up the session key to cloud sync.** Defeats the "local only"
  security promise — use OS-scoped storage that doesn't sync.

## Related

- `yellow-clearnode-auth` — full 3-step flow the session key is registered through
- `yellow-nitro-rpc` — signatures on req bytes
- `yellow-sdk-v1` / `yellow-sdk-api` — typed SDKs that manage session keys for you
