# Yellow Network Development Patterns — Rules & Corrections

## Protocol Version

Always use `NitroRPC/0.4` for new app sessions. Never use `NitroRPC/0.2`.

```typescript
// CORRECT
definition: { protocol: 'NitroRPC/0.4', ... }

// WRONG — deprecated, lacks intent system
definition: { protocol: 'NitroRPC/0.2', ... }
```

Using `NitroRPC/0.2` with intent/version params causes `"incorrect request: specified parameters not supported"` errors.

## Fund Operation Rules

Use `allocate_amount` when moving funds from unified balance. Only use `resize_amount` for direct L1 Custody Contract deposits (rare).

```typescript
// CORRECT — resizing from unified balance (normal case)
await resizeChannel(client, channelId, '100.0', signer);
// Internally sends: { allocate_amount: '100.0' }

// WRONG — causes InsufficientBalance if no L1 deposit preceded this
// Don't pass resize_amount manually unless you know you have a pending L1 deposit
```

## Asset Handling

Asset symbols must be **lowercase**. Amounts must be **human-readable decimal strings**.

```typescript
// CORRECT
{ asset: 'usdc', amount: '50.0' }
{ asset: 'yellow', amount: '1000.0' }
{ asset: 'eth', amount: '0.01' }

// WRONG
{ asset: 'USDC', amount: '50000000' }  // uppercase + raw units both wrong
```

ClearNode handles unit conversion internally. Never convert to wei or smallest unit before sending.

## Signature Formats

- **Auth** (`auth_verify`): EIP-712 typed data signature from main wallet
- **Session operations** (all post-auth messages): ECDSA signature from session key
- **Channel state signing**: Raw ECDSA on `keccak256(abi.encode(channelId, intent, version, data, allocations))`

Do not swap these — signing auth_verify with a session key will fail.

## Connection Management

Re-authenticate after every reconnect. `isAuthenticated` resets to `false` on WebSocket close.

```typescript
// CORRECT — handle reconnect + re-auth
const reconnectAndAuth = async () => {
  await client.connect();
  await authenticate(client, authConfig, signerFn);
};

// Many client wrappers reconnect the WebSocket automatically,
// but do NOT re-authenticate. Always trigger re-auth after reconnect.
```

## Session Governance Rules

App session `definition` fields:

| Field | Rule |
|-------|------|
| `protocol` | Must be `'NitroRPC/0.4'` |
| `participants` | Order matters — index determines which weight applies |
| `weights` | Each participant's voting power |
| `quorum` | Minimum total weight required to update state |
| `challenge` | Seconds for dispute window (min 3600, recommend 86400+) |
| `nonce` | Must be unique per session (use `Date.now()`) |

The `app_session_id` is derived deterministically: `keccak256(JSON.stringify(definition))`. If the definition is the same, the ID is the same — always use unique nonces.

## State Update Rules

- Version increments exactly by 1 per update (`currentVersion + 1`)
- Allocations show **final state**, not deltas
- Fund conservation law:
  - `OPERATE`: sum unchanged
  - `DEPOSIT`: sum increases (depositor must sign)
  - `WITHDRAW`: sum decreases (quorum sufficient, depositor sign not required)
  - `close_app_session`: final sum == current total

## Error Handling

| Error | Root Cause | Fix |
|-------|-----------|-----|
| `InsufficientBalance` | Not enough unified balance | Check balance before creating sessions |
| `Quorum not met` | Signer weights < quorum | Verify quorum math: sum of signer weights >= quorum |
| `Version mismatch` | Wrong version number | Track version from `create_app_session` response |
| `non-zero allocation` | Stale open channels | Close all channels before opening new ones |
| `incorrect request: specified parameters not supported` | NitroRPC/0.2 session receiving v0.4 params | Use `NitroRPC/0.4` |
| `Session expired` | Past `expireSeconds` | Re-authenticate |
| `Allowance exceeded` | Session key spending cap hit | Create new session key with higher allowances |

## Multi-Chain Notes

The unified balance is chain-agnostic. Funds deposited on any supported chain contribute to the same unified balance. Design implications:
- A single ClearNode session per wallet is typically sufficient; multiple sessions are not required per chain
- Custodial platforms often hold one platform-level ClearNode session on behalf of many users; non-custodial apps dial ClearNode directly from each client
- Cross-chain deposits are transparent to the application layer — query `get_ledger_balances` without a chain filter to see the aggregate
