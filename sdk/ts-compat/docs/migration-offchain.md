# Off-Chain Migration Guide

This guide covers off-chain operations when migrating from v0.5.3 to the compat layer: authentication, app sessions, transfers, ledger queries, event polling, and transitional RPC helper imports.

## 1. Authentication

v1.0.0 handles authentication internally when using `NitroliteClient`. For legacy WebSocket-auth code paths, the compat layer keeps `createAuthRequestMessage`, `createAuthVerifyMessage`, `createAuthVerifyMessageWithJWT`, and `createEIP712AuthMessageSigner` available.

## 2. App Sessions

### List

**Before:** `NitroliteRPC.createRequest` + `createGetAppSessionsMessage` + `sendRequest` + `parseGetAppSessionsResponse`  
**After:** `client.getAppSessionsList()`

### Create

**Before:**
```typescript
const msg = await createAppSessionMessage(signer.sign, { definition, allocations });
const raw = await sendRequest(msg);
parseCreateAppSessionResponse(raw);
```
**After:** `client.createAppSession(definition, allocations)`

### Close

**Before:** `createCloseAppSessionMessage` + send + parse  
**After:** `client.closeAppSession(appSessionId, allocations)`

### Submit State

**Before:** `createSubmitAppStateMessage` + send  
**After:** `client.submitAppState(params)`

### Get Definition

**Before:** `createGetAppDefinitionMessage` + send + parse  
**After:** `client.getAppDefinition(appDefinitionId)`

## 3. Transfers

**Before:**
```typescript
const msg = await createTransferMessage(signer.sign, { destination, allocations });
await sendRequest(msg);
```
**After:** `client.transfer(destination, allocations)`

## 4. Ledger Queries

**Before:** `createGetLedgerBalancesMessage` / `createGetLedgerEntriesMessage` + send + parse  
**After:** `client.getBalances()`, `client.getLedgerEntries()`

## 5. Event Polling

v0.5.3 used WebSocket push events (`ChannelUpdate`, `BalanceUpdate`). v1.0.0 uses polling. The compat layer provides `EventPoller`:

```typescript
import { EventPoller } from '@yellow-org/sdk-compat';

const poller = new EventPoller(client, {
  onChannelUpdate: (channels) => updateUI(channels),
  onBalanceUpdate: (balances) => updateBalances(balances),
  onAssetsUpdate:  (assets)   => updateAssets(assets),
  onError:         (err)      => console.error(err),
}, 5000);
poller.start();
```

## 6. RPC Compatibility Helpers

The `create*Message` and `parse*Response` exports still exist primarily so legacy imports can keep compiling while you migrate call sites. Treat them as transitional compatibility shims, not as proof of full one-to-one v1 RPC coverage. For new code, prefer `NitroliteClient` methods directly. Examples: `createGetChannelsMessage`, `parseGetChannelsResponse`, `createTransferMessage`, `createAppSessionMessage`, `createCloseAppSessionMessage`, etc.
