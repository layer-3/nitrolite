# On-Chain Migration Guide

This guide covers on-chain operations when migrating from v0.5.3 to the compat layer: deposits, withdrawals, channel operations, amount handling, and contract addresses.

## 1. Deposits

**Before (v0.5.3):** Manual approve → deposit → createChannel

```typescript
await approveToken(custody, tokenAddress, amount);
await sendRequest(createCreateChannelMessage(signer.sign, { token: tokenAddress, amount }));
```

**After (compat):** Single call — approval and channel creation are implicit

```typescript
await client.deposit(tokenAddress, amount);
```

The legacy `createChannel()` client method and `createCreateChannelMessage(...)` helper still exist for migration, but they now throw with migration guidance instead of warning or fabricating a fake wire payload. Use `deposit(...)` or `depositAndCreateChannel(...)` instead.

`createCreateChannelMessage` remains exported so old imports keep compiling, but it now fails fast with migration guidance because channel creation is no longer a standalone protocol-backed RPC in v1.

## 2. Withdrawals

**Before (v0.5.3):** Manual close → checkpoint → withdraw

```typescript
const closeMsg = await createCloseChannelMessage(signer.sign, { channel_id });
const raw = await sendRequest(closeMsg);
// ... checkpoint on-chain ...
```

**After (compat):** Single call

```typescript
await client.withdrawal(tokenAddress, amount);
```

`createCloseChannelMessage` now fails fast with migration guidance instead of pretending to be a real v1 wire helper.

## 3. Channel Operations

Legacy channel helper imports may still exist to keep migration moving, but the supported path is the compat client methods below. Do not treat every legacy helper name as a protocol-backed one-to-one v1 RPC.

| Operation | v0.5.3 | Compat |
|-----------|--------|--------|
| Create | Explicit `createChannel()` (now throwing migration shim) | Implicit on first `deposit()` |
| Close | `createCloseChannelMessage` (now fail-fast shim) | `client.closeChannel()` |
| Resize | `createResizeChannelMessage` (now fail-fast shim) | `client.resizeChannel({ allocate_amount, token })` |

**Example — close:**

```typescript
// After
await client.closeChannel();
```

## 4. Amount Handling

**Before (v0.5.3):** Raw `BigInt` everywhere; app must handle decimals

```typescript
const amount = 11_000_000n; // 11 USDC (6 decimals)
// Manual conversion for display: formatUnits(amount, 6)
```

**After (compat):** App-facing methods still accept raw token amounts, and the compat layer handles the conversion needed to call the v1 SDK correctly

```typescript
// Raw BigInt still works
await client.deposit(tokenAddress, 11_000_000n);

// Or use helpers
const formatted = client.formatAmount(tokenAddress, 11_000_000n); // "11.0"
const parsed = client.parseAmount(tokenAddress, "11.0");         // 11_000_000n
```

Transfer allocations still use raw smallest-unit strings, for example `{ asset: 'usdc', amount: '5000000' }` for 5 USDC.

App-session allocations are different: those remain human-readable decimal strings such as `{ asset: 'usdc', amount: '5.0' }`.

## 5. Contract Addresses

**Before (v0.5.3):** Manual config — custody, adjudicator, etc.

```typescript
const addresses = {
  custody: '0x...',
  adjudicator: '0x...',
};
```

**After (compat):** Fetched from clearnode `get_config` — no manual setup. The `addresses` field in config is deprecated and ignored.
