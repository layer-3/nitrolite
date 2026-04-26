---
name: yellow-nitro-rpc
description: |
  Nitro RPC message format reference — the wire protocol every ClearNode WebSocket client speaks. Covers the compact `[req, sig]` envelope, request/response/notification shapes, signatures (keccak256 over the req array, ECDSA secp256k1), timestamps in Unix milliseconds, error messages, and request-ID correlation. Use when: implementing a ClearNode client from scratch, debugging why a request is rejected, wiring signatures into the envelope, or auditing a pull request that changes the wire format.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/message-format
---

# Yellow Nitro RPC — Wire Format

Every message between a client and a ClearNode is a JSON object that wraps a
compact 4-element array. Get this wrong and every other method (auth, transfer,
app sessions) fails silently. This skill is the single source of truth for the
envelope — spec: <https://docs.yellow.org/docs/protocol/off-chain/message-format>.

## Envelope

```ts
// Request
{
  "req": [requestId, method, params, timestamp],
  "sig": [signature1, signature2, ...]   // 0..N signatures
}

// Response
{
  "res": [requestId, method, result, timestamp],
  "sig": [serverSig]                     // usually 1
}
```

| Field | Type | Notes |
|---|---|---|
| `requestId` | uint64 | Client-generated, unique per connection. Correlates the response. |
| `method` | string | `snake_case` RPC name (`auth_request`, `create_app_session`, `transfer`…). Response echoes the same method — **except** `auth_request` → `auth_challenge`, `ping` → `pong`, and any error → `error`. |
| `params` | object | Method-specific. See Auth, Transfer, App Sessions, Channels specs. |
| `result` | object | On `res`, replaces `params`. |
| `timestamp` | uint64 | **Unix milliseconds** (not seconds). `Date.now()`. |

## Signatures

Every authenticated request carries 1..N signatures in the `sig` array.

- **Algorithm**: ECDSA over `secp256k1`.
- **Hash**: `keccak256` of the **exact JSON-encoded bytes of the `req` array** — no extra whitespace, stable key order.
- **Format**: `0x`-prefixed hex string, 65 bytes (r || s || v).
- **Multi-party ops** (co-signed app sessions) attach one signature per participant in the order required by the method.

Signature over `auth_verify` is special: it is EIP-712 typed data over the
`Policy` type (see `yellow-clearnode-auth`), not a keccak256 of the req bytes.

### Canonical JSON serialization

The exact bytes on the wire are what the signature covers. Re-encoding with
different formatting invalidates the signature. Five rules per spec:

1. **Deterministic key ordering** — every object uses the same key order on
   every run (alphabetical, or any stable order both sides agree on).
2. **No unnecessary whitespace** — compact output (no pretty-printing).
3. **Standard JSON** — no trailing commas, no comments.
4. **UTF-8 encoding** — the bytes you hash must be the UTF-8 serialization.
5. **Large integers as strings** — `"18446744073709551615"`, not the raw
   number literal (JS `Number` loses precision past 2^53).

Implementation: sign the **exact `JSON.stringify(...)` output** you're about
to `ws.send(...)`. Don't re-stringify inside your signer — hash the same
bytes you transmit.

## Notifications (server-initiated)

ClearNode pushes unsolicited notifications using `res` with `requestId` = 0.
The five notification methods are `assets` (hydration push on connect), `bu`,
`cu`, `tr`, `asu` — full catalog lives in `yellow-notifications`.

```json
{ "res": [0, "assets", { "assets": [...] }, 1776953834665] }
{ "res": [0, "bu", { "balance_updates": [{ "asset": "usdc", "amount": "123.45" }] }, ...] }
```

Client routers should:
1. If `requestId > 0` and we have a pending callback → resolve/reject it.
2. Else dispatch by `method` to notification handlers.

A typical router implementation is 20–30 lines of dispatch logic.

## Errors

Errors arrive as a response where the method is replaced by `"error"`. The
error object carries **a descriptive string, not a numeric code**:

```json
{ "res": [42, "error", { "error": "Insufficient balance: required 100 USDC, available 75 USDC" }, 1699...] }
```

This is a design choice of Yellow Network — there is no numeric error code
table. Dispatch on substrings (`"authentication required"`, `"channel not
found"`, `"session expired"`, `"insufficient balance"`). See the
`yellow-errors` skill for the canonical message list and recovery steps.

## Reference implementation

```ts
async send<T>(method, params, signer?) {
  const reqId = ++this.requestId;
  const timestamp = Date.now();           // ← ms, not s
  const req = [reqId, method, params, timestamp];
  const message: Record<string, unknown> = { req };
  if (signer) {
    const sig = await signer(req);        // signs the req array
    message['sig'] = [sig];
  }
  this.ws.send(JSON.stringify(message));
  // ...register reqId -> Promise, resolve on matching res
}
```

## Quick sanity test

Connect to sandbox and send any method — you should see your `auth_challenge`
response come back with the same `requestId` you sent:

```bash
node -e "
const WS = require('ws');
const ws = new WS('wss://clearnet-sandbox.yellow.com/ws');
ws.on('open', () => ws.send(JSON.stringify({
  req: [42, 'auth_request', {
    address: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2',
    session_key: '0x9876543210fedcba9876543210fedcba98765432',
    application: 'smoke-test',
    allowances: [],
    expires_at: Date.now() + 60000,
  }, Date.now()],
})));
ws.on('message', (d) => { console.log(d.toString()); process.exit(0); });
"
```

You should see:
```
{"res":[42,"auth_challenge",{"challenge_message":"<uuid>"},<ms>],"sig":["0x..."]}
```

## Common gotchas

- **Seconds vs milliseconds.** The older `@erc7824/nitrolite` docs and some tutorials use seconds. Current spec is **milliseconds**. Mixing them causes clock-skew rejections.
- **Key ordering.** If you prettify JSON or reorder keys between signing and sending, the signature won't verify. Sign the exact bytes you put on the wire.
- **Notification vs response.** Don't route a notification (`requestId = 0`) to a pending callback. Guard on `id > 0 && pending.has(id)`.
- **`auth_request` response method = `auth_challenge`.** If you key your callbacks on method name, you'll miss it.

## Related skills

- `yellow-clearnode-auth` — full auth handshake and EIP-712 Policy signing
- `yellow-sdk-v1` / `yellow-sdk-api` — typed SDKs that wrap this wire format
- `yellow-session-keys` — session key generation, scopes, allowances
- `yellow-app-sessions` — payload shapes for escrow / stake / multi-party
