# Off-Chain Migration Guide

This guide covers off-chain operations when migrating from v0.5.3 to the compat layer: authentication, app sessions, transfers, ledger queries, event polling, and transitional RPC helper imports.

## 1. Authentication

v1 handles authentication internally when using `NitroliteClient`. For legacy WebSocket-auth code paths, the compat layer keeps `createAuthRequestMessage`, `createAuthVerifyMessage`, `createAuthVerifyMessageWithJWT`, and `createEIP712AuthMessageSigner` available.

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

`createAppSessionMessage` now emits a real `app_sessions.v1.create_app_session` payload inside the legacy `req` / `sig` envelope, but new integrations should still prefer `client.createAppSession(...)`.

### Close

**Before:** `createCloseAppSessionMessage` + send + parse  
**After:** `client.closeAppSession(appSessionId, allocations)`

`createCloseAppSessionMessage` maps to `app_sessions.v1.submit_app_state` with `intent = close`, so it now requires an explicit `version` if you keep the helper path.

### Submit State

**Before:** `createSubmitAppStateMessage` + send  
**After:** `client.submitAppState(params)`

`createSubmitAppStateMessage` also requires `params.version` for the live v1 mapping.

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

`createTransferMessage` remains exported only so old imports keep compiling, but it now fails fast with migration guidance because transfer is no longer a single direct v1 RPC helper. This is a deliberate runtime change from the old silent placeholder behavior.

## 4. Ledger Queries

**Before:** `createGetLedgerBalancesMessage` / `createGetLedgerEntriesMessage` + send + parse  
**After:** `client.getBalances()`, `client.getLedgerEntries()`

`createGetLedgerBalancesMessage` now emits a real `user.v1.get_balances` request and requires the wallet/account parameter. `createGetAppSessionsMessage` and `createGetAppDefinitionMessage` likewise emit live v1 request shapes inside the legacy envelope.

## 5. Event Polling

v0.5.3 used WebSocket push events (`ChannelUpdate`, `BalanceUpdate`). v1 uses polling. The compat layer provides `EventPoller`:

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

The `create*Message` and `parse*Response` exports still exist primarily so legacy imports can keep compiling while you migrate call sites.

- Direct query/app-session helpers such as `createGetChannelsMessage`, `createGetLedgerBalancesMessage`, `createGetAppSessionsMessage`, `createGetAppDefinitionMessage`, `createAppSessionMessage`, `createSubmitAppStateMessage`, `createCloseAppSessionMessage`, and `createPingMessage` now emit live v1 method names and payload shapes inside the legacy envelope.
- Workflow helpers such as `createTransferMessage` stay exported but fail fast with migration guidance instead of returning fake wire payloads.
- `parse*Response` helpers only normalize known response fields; they do not recreate old payloads that the live v1 server no longer returns.

For new code, prefer `NitroliteClient` methods directly.

### Amount conventions

- `TransferAllocation.amount` remains a raw smallest-unit string such as `'5000000'` for 5 USDC.
- App-session allocation amounts in `createAppSession`, `closeAppSession`, and `submitAppState` remain human-readable decimal strings such as `'0.01'`.
