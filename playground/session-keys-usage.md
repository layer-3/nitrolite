# Channel Session Keys — Usage Guide

Source of truth: `pkg/core/session_key.go`, `nitronode/api/channel_v1/`, `sdk/ts/src/signers.ts`, `playground/src/sessionKey.ts`.

---

## Overview

A channel session key is a delegated signing key. Once registered, it can sign off-chain channel state transitions on behalf of a wallet — no MetaMask popup per state. On-chain operations (deposit, checkpoint, close) always require the wallet.

---

## 1. Issuance (Registration)

### RPC method
`channels.v1.submit_session_key_state`

### State object — all fields required at submission

| Field | Type | Notes |
|-------|------|-------|
| `user_address` | string | Wallet address authorizing the delegation |
| `session_key` | string | Address of the session key being registered |
| `version` | string (uint64) | Must be exactly `latestVersion + 1`; first registration = `"1"` |
| `assets` | string[] | Asset IDs this key may sign for (e.g. `["usdc", "eth"]`) |
| `expires_at` | string (unix secs) | Future → active; past/now → immediate revocation |
| `user_sig` | string | EIP-191 wallet signature (triggers MetaMask once) |
| `session_key_sig` | string | EIP-191 session key ownership proof (local, no popup) |

### What the signatures sign

Both signatures sign the same packed payload:

```
keccak256(SESSION_KEY_AUTH_TYPEHASH, session_key, metadataHash)
```

where:

```
metadataHash = keccak256(abi.encode(user_address, version, assets, expires_at))
```

This binds both signatures to the exact `(wallet, session_key, version, assets, expires_at)` tuple — no replay possible across different pairs.

### TypeScript issuance flow

```typescript
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import { EthereumMsgSigner, type ChannelSessionKeyStateV1 } from '@yellow-org/sdk';

// 1. Generate a fresh throwaway key
const skPrivateKey = generatePrivateKey();
const skAccount   = privateKeyToAccount(skPrivateKey);
const skMsgSigner = new EthereumMsgSigner(skPrivateKey);

// 2. Build state (version = latestVersion + 1; first time = 1)
const state: ChannelSessionKeyStateV1 = {
  user_address:    walletAddress,
  session_key:     skAccount.address,
  version:         nextVersion.toString(),
  assets:          ['usdc', 'eth'],
  expires_at:      (Math.floor(Date.now() / 1000) + 86400).toString(), // 24 h
  user_sig:        '',
  session_key_sig: '',
};

// 3. Wallet signs (one MetaMask popup)
state.user_sig = await client.signChannelSessionKeyState(state);

// 4. Session key signs locally (no popup)
state.session_key_sig = await client.signChannelSessionKeyOwnership(state, skMsgSigner);

// 5. Submit
await client.submitChannelSessionKeyState(state);
```

### What to persist after registration

You need all of the following to reconstruct the signer in future sessions:

```typescript
type StoredSessionKey = {
  privateKey:        Hex;     // session key private key
  sessionKeyAddress: Address;
  walletAddress:     Address;
  version:           string;
  assets:            string[];
  expiresAt:         string;  // unix secs
  userSig:           Hex;     // wallet's auth signature
};
```

### Operations: registration / update / revocation

All three use the same RPC method and the same version-increment rule:

| Operation | `expires_at` | Effect |
|-----------|-------------|--------|
| Registration | future | Activates the key |
| Update (rotate assets / extend lifetime) | future | Replaces the active state |
| Revocation | `<= now` | Retires the key immediately; slot is freed |

---

## 2. Version Rules

### The rule
Every submit must have `version == latestVersion + 1` — exactly, not "at least". Any other value returns:
```
invalid_session_key_state: expected version <N>, got <M>
```

### Version 0 is invalid
Version `0` is rejected before any other validation. Internally it is a seed row that reserves ownership of a `(session_key, kind)` slot — never a valid submitted state.

### Version scoping
Version is scoped per `(user_address, session_key, kind)` triplet. Channel keys (`kind=1`) and app-session keys (`kind=2`) for the same private key have completely independent version sequences.

### How to get the current version before submitting

There is no dedicated version endpoint. Call `getLastChannelKeyStates` and read `state.version`:

```typescript
const states = await client.getLastChannelKeyStates(walletAddress, sessionKeyAddress);

const nextVersion = states.length === 0
  ? 1n
  : BigInt(states[0].version) + 1n;
```

### Race conditions
Concurrent submits for the same key are serialized by a row-level lock (`LockSessionKeyState`). The losing transaction re-reads the updated `latestVersion` and receives the "expected version N, got M" error. The client must re-fetch and retry with `latestVersion + 1`.

---

## 3. Using Session Keys for Channel Operations

### Client initialization

Pass a `ChannelSessionKeyStateSigner` as `stateSigner` to `Client.create()`. The `txSigner` (on-chain operations) always uses the wallet — never the session key.

```typescript
import {
  Client,
  ChannelDefaultSigner,
  ChannelSessionKeyStateSigner,
  EthereumMsgSigner,
  getChannelSessionKeyAuthMetadataHashV1,
} from '@yellow-org/sdk';

function buildSessionKeyStateSigner(sk: StoredSessionKey): StateSigner {
  const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
    sk.walletAddress,
    BigInt(sk.version),
    sk.assets,
    BigInt(sk.expiresAt),
  );
  return new ChannelSessionKeyStateSigner(
    sk.privateKey,
    sk.walletAddress,
    metadataHash,
    sk.userSig,   // wallet's auth signature from registration
  );
}

// Choose signer based on whether a valid session key exists
const stateSigner = sessionKey
  ? buildSessionKeyStateSigner(sessionKey)
  : new ChannelDefaultSigner(new WalletStateSigner(walletClient));

const client = await Client.create(NODE_URL, stateSigner, txSigner, ...opts);
```

### Wire format

When the client signs with a session key, it does not produce a bare 65-byte ECDSA signature. It produces:

```
0x01 || abi.encode(
  SessionKeyAuthorization { sessionKey, metadataHash, authSignature },
  sessionKeySig
)
```

The `0x01` type byte tells the server to unwrap the authorization bundle. The server then:
1. Verifies `authSignature` (wallet) over `(SESSION_KEY_AUTH_TYPEHASH, sessionKey, metadataHash)`
2. Verifies `sessionKeySig` (session key) over the actual state hash
3. Checks asset permission and expiry

(`sdk/ts/src/signers.ts:224-255`, `pkg/core/channel_signer.go:172-212`)

### What can and cannot be session-key signed

| Operation | Session key OK? |
|-----------|----------------|
| Off-chain state transitions (`submitState`, transfers, acks) | Yes |
| Session key registration (`submitChannelSessionKeyState`) | Partial — `user_sig` must be wallet; `session_key_sig` is the session key |
| On-chain: `checkpoint`, `deposit`, `withdraw`, `closeChannel` | No — always wallet via `txSigner` |

---

## 4. Constraints

### Asset list
Enforced **server-side only**. The server joins against `channel_session_key_assets_v1` and checks that the asset being signed for is in the registered list. Submitting a state for an asset not in the list returns:
```
session key does not have permission to sign for this data
```

### Expiry
Enforced **server-side only** (`expires_at > now`). The same generic error is returned for both expired keys and wrong-asset cases — intentional, to avoid leaking whether a key exists but is expired.

The playground uses a client-side `isExpired()` helper with a renewal buffer as a UX optimization, but the SDK itself does not pre-check expiry.

### `kind` isolation
The same private key can be registered independently as:
- `kind = 1` — channel session key (`channels.v1.submit_session_key_state`)
- `kind = 2` — app-session key (different RPC)

They have separate version histories and asset lists. The SDK selects `kind` automatically based on which RPC method is called — there is no SDK-level `kind` parameter.

### One owner per key per kind
Once a wallet seeds the ownership row for a `(session_key, kind)` pair it is permanent. No other wallet can register that session key address for the same kind, even after revocation.

### Server-side caps (configurable)
- Maximum session keys per user: `maxSessionKeysPerUser`
- Maximum assets per session key: `maxSessionKeyIDs`

Exceeding either cap at registration returns an error before the version check.
