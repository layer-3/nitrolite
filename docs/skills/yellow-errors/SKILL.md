---
name: yellow-errors
description: |
  How errors work in Nitro RPC — **no numeric codes**, only descriptive strings. Canonical list of common error messages across auth, transfers, channels, app sessions, and queries, with the cause of each and a recovery playbook. Use when: writing error handlers, building retry logic, debugging a rejected request, surfacing user-facing messages, or auditing code that tries to switch on error codes (it won't work).
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/message-format
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/authentication
---

# Yellow Errors

Yellow Network's Nitro RPC **does not use numeric error codes**. Every error
is a descriptive string inside an `error` field. Design your error handling
around string matching or prefix dispatch — you won't find a `code: 4001`
anywhere.

Spec trail: search "Error Handling" sections in auth, channel-methods,
app-sessions, transfers pages on `docs.yellow.org`.

## Wire shape

```json
{ "res": [<requestId>, "error", { "error": "<descriptive message>" }, <ts>], "sig": ["0x..."] }
```

Rules:
- The response's **method is literally `"error"`**, not the request method.
- The result object has a single `error` string field.
- Same `requestId` as your request, so you can still correlate to the call
  that failed.

Client-side handler pattern:

```ts
if (method === 'error') {
  cb.reject(new Error(data.error ?? JSON.stringify(data)));
}
```

## Canonical error catalog

### Authentication

| Message | Cause | Recovery |
|---|---|---|
| `Invalid address format` | Main wallet address malformed | Verify 0x + 40 hex chars |
| `Invalid session key format` | Session key malformed | Same |
| `Invalid parameters` | Missing/invalid required field | Re-check `auth_request` schema |
| `Session key already registered` | The `session_key` address was used in a prior auth | Generate a fresh keypair |
| `invalid challenge or signature` | Policy signature doesn't recover to `address`, or challenge isn't in pending auths. Current sandbox collapses the old `Invalid signature` / `Invalid challenge` / `Challenge mismatch` rows into this one string. | Ensure sigs are by **main wallet** (not session key); verify JSON bytes match; retry with a fresh `auth_request` if the challenge has aged out |
| `Challenge expired` | More than 5 minutes between `auth_request` and `auth_verify` | Start over with a fresh `auth_request` |
| `Challenge already used` | Challenge previously verified (replay) | Fresh `auth_request` |
| `failed to generate JWT token` | Policy signature verified, but server-side JWT issuance failed — typically because the account has never been registered on this ClearNode | Register/fund the wallet before auth, or retry if the server was transiently unavailable |
| `session expired, please re-authenticate` | Session `expires_at` passed | Re-auth with fresh `auth_request` |
| `Session key allowance exceeded: <required>, <remaining>` | Aggregate spend exceeded the allowance for an asset | Rotate with a new allowance |
| `authentication required` | Calling authenticated methods before auth completed (the short, lowercase form the sandbox currently returns — older docs show `Authentication required: session not established`) | Run the 3-step auth flow |

### Transfers

| Message | Cause | Recovery |
|---|---|---|
| `Insufficient balance: required <X>, available <Y>` | Sender balance < amount for one of the assets | Top up unified balance or reduce amount |
| `Unsupported asset` / `Asset not found` | Unknown `asset` symbol for the chain | Call `get_assets` first; remember symbols are lowercase |
| `Invalid destination` | Neither `destination` nor `destination_user_tag` valid | Verify addresses / resolve tag via `get_user_tag` |
| `Self-transfer not allowed` | `destination` equals authenticated sender | Transfer to someone else |
| `Allowance exceeded` | Session key hit its allowance cap | Rotate session with higher allowance |

### Channels

| Message | Cause | Recovery |
|---|---|---|
| `Unsupported chain` | `chain_id` not in this ClearNode's config | Check `get_config.networks[]` |
| `Token not supported` | Asset config missing for that `(chain_id, token)` | Check `get_assets` |
| `Channel not found` | Unknown `channel_id` | List via `get_channels` |
| `Channel already exists` | Open channel with this broker already present | Reuse existing, or close first |
| `Channel challenged` | Channel in dispute window | Resolve challenge before new ops |
| `Channel not open/resizing` | Status isn't `open` or `resizing` (close only valid from those) | Wait for resize to complete |
| `Operation denied: resize already ongoing` | Concurrent resize attempt | Wait for the first to finalize |
| `Insufficient unified balance` | Positive delta > available unified balance | Deposit more or reduce delta |
| `New channel amount must be positive` | Resize would drive channel balance negative | Reduce withdraw magnitude |
| `Resize operation requires non-zero amounts` | Both `resize_amount` and `allocate_amount` are 0 | Pass a non-zero value |
| `Failed to pack/sign state` | Internal state encoding error | Retry; if persistent, contact ClearNode operator |

### App Sessions

| Message | Cause | Recovery |
|---|---|---|
| `Insufficient balance` (on create) | A participant's unified balance < their allocation | Top up |
| `Quorum not met` | Summed weights of signers < `quorum` | Collect more sigs |
| `Allocation mismatch` / `Close total != locked total` | Sum of close allocations doesn't equal current session total for some asset | Recompute, match exactly |
| `Version conflict` | Two concurrent `submit_app_state` with the same version | Refetch via `get_app_sessions`, bump |
| `Invalid intent` | `intent` not one of `operate`/`deposit`/`withdraw` (NitroRPC/0.4) | Fix enum value |
| `DEPOSIT requires depositing participant signature` | Missing sig from the one adding funds | Add their sig to envelope `sig[]` |
| `Session closed` | Trying to mutate after close | Create a new session |

### Queries & misc

| Message | Cause | Recovery |
|---|---|---|
| `Method not found: '<name>'` | Typo or unsupported method | Verify against `yellow-queries` list |
| `Pagination limit exceeded` | `limit > 100` | Cap at 100 per call, paginate with `offset` |
| `Rate limit exceeded` | Too many RPC calls per window | Backoff; consider a JWT-based reconnect instead of re-auth |

## Handler patterns

### Prefix-match dispatch

```ts
function handleError(err: Error) {
  const msg = err.message;
  if (msg.startsWith('session expired') || msg.startsWith('Authentication required')) {
    return reAuthenticate();
  }
  if (msg.includes('allowance exceeded') || msg.includes('Allowance exceeded')) {
    return rotateSessionKey();
  }
  if (msg.startsWith('Challenge expired') || msg.startsWith('Challenge already used')) {
    return restartAuthFlow();
  }
  if (msg.startsWith('Insufficient balance')) {
    return promptTopUp();
  }
  if (msg.startsWith('Channel challenged') || msg.startsWith('Channel not found')) {
    return refreshChannels();
  }
  // Fallback — surface to user with the raw message
  showToast(`ClearNode: ${msg}`);
}
```

### Retry with backoff — but not on everything

Safe to retry:
- `Failed to pack/sign state` — likely transient
- Network-level failures (disconnect)
- `Rate limit exceeded` (with backoff)

**Never retry blindly**:
- `Session key already registered` — you'll just fail again
- `invalid challenge or signature` — bug in your signer or wrong wallet; retry won't help, perform a fresh `auth_request` only if the challenge aged out
- `Challenge expired` — must go through a fresh auth first
- `Insufficient balance` — no amount of retrying will fund you

### Surface-safe user messages

Raw messages can leak internals. For end-user UI, map to friendly copy:

```ts
const USER_MSG: Array<[RegExp, string]> = [
  [/^Insufficient balance/, 'Not enough balance'],
  [/^Session key allowance exceeded/, 'Your session spending limit is reached. Reconnect to continue.'],
  [/^Challenge expired/, "Your sign-in window timed out. Let's try again."],
  [/^Channel challenged/, 'Your channel is in dispute. Please wait for resolution.'],
  [/^Rate limit exceeded/, "You're going too fast — slow down for a moment."],
];
function friendly(err: string): string {
  for (const [re, msg] of USER_MSG) if (re.test(err)) return msg;
  return 'Something went wrong. Please retry.';
}
```

## Tips

- **Log the raw `error` string plus `method` plus `requestId`.** Descriptive
  strings evolve over time — keeping the raw form lets you adapt without
  shipping a client update.
- **Never rely on wording for side-effect-free code paths.** If you branch on
  error content in a happy-path assertion, you'll break the next time the
  docs are rephrased. Catch, log, fall through.
- **Test your matchers against the live ClearNode.** Feed known-bad
  parameters and record the exact string returned — then build regex tests
  around those exact strings. The docs are close but not always byte-exact.

## Related

- `yellow-nitro-rpc` — envelope format, where the `"error"` method lives
- `yellow-clearnode-auth` — authentication error messages in context
- `yellow-app-sessions` — app session error scenarios
- `yellow-queries` — `get_rpc_history` to inspect errors server-side
