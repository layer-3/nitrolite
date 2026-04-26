# sdk-v1 — Complete Reference

> Extended documentation. Read the `SKILL.md` quick-start first; this file contains the full details moved out for progressive disclosure.


# Yellow SDK v1 (`@yellow-org/sdk`)

**Status**: Active, current mainnet SDK
**Latest version**: `1.2.0` (verified against tarball, 2026-04-23)
**Supersedes**: `@erc7824/nitrolite@0.5.3` (frozen)
**Runtime**: Node.js >= 20
**Dependencies** (runtime): `viem ^2.46.1`, `decimal.js ^10.4.3`, `zod ^4.3.6`, `abitype ^1.2.3`


## Why v1 exists (vs v0.5.3)

v0.5.3 exposed a bring-your-own-WebSocket model: the app dialed the WS, signed every `createXxxMessage`, and parsed every response. v1 collapses this into a single `Client` that owns the socket, the state-signer, and the on-chain `txSigner`. The wire format, auth flow, unit system, and asset resolution were redesigned across roughly 14 dimensions — direct migration touches dozens of files, which is why `@yellow-org/sdk-compat` exists (see bottom).

The package description verbatim: "The Yellow SDK empowers developers to build high-performance, scalable web3 applications using state channels. It's designed to provide near-instant transactions and significantly improved user experiences by minimizing direct blockchain interactions."


## Module layout

Top-level `@yellow-org/sdk` re-exports from:

- `./client` — `Client`, `DEFAULT_CHALLENGE_PERIOD` (86400), `StateSigner`, `TransactionSigner` types
- `./signers` — signer classes + `createSigners`
- `./config` — `Config`, `Option`, `DefaultConfig`, `withHandshakeTimeout`, `withErrorHandler`, `withBlockchainRPC`
- `./asset_store` — `ClientAssetStore`
- `./utils` — type transformations
- `./core` — state, enums (ChannelStatus, TransitionType, TransactionType, INTENT_* constants)
- `./app` — app session types + packing helpers
- `./blockchain` — EVM blockchain interactions
- `./rpc` — low-level RPCClient, types, methods, message codec, dialer, errors

Wildcard exports mean every symbol in `core`, `app`, `blockchain`, `rpc`, and `utils` is available from the root package.


## The `Client` class

Unlike v0.5.3, you don't call `new Client(...)`. Use the async factory:

```typescript
import { Client, createSigners, withBlockchainRPC, withHandshakeTimeout } from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY as `0x${string}`);

const client = await Client.create(
  'wss://clearnode.yellow.com/ws',
  stateSigner,       // signs channel states (EIP-191 prefixed)
  txSigner,          // signs on-chain transactions
  withBlockchainRPC(1n, process.env.MAINNET_RPC!), // required for checkpoint()
  withHandshakeTimeout(10_000),
);
```

`Client.create` signature (verified from `dist/client.d.ts`):

```typescript
static create(
  wsURL: string,
  stateSigner: StateSigner,
  txSigner: TransactionSigner,
  ...opts: Option[]
): Promise<Client>
```

### Exported Option factories (verified)

- `withHandshakeTimeout(ms: number)`
- `withErrorHandler((error: Error) => void)`
- `withBlockchainRPC(chainId: bigint, rpcUrl: string)`

> The README also documents `withPingInterval(ms)`, but as of `1.2.0` it is **not** re-exported from the package root (`dist/index.d.ts` only lists the three above). Treat `withPingInterval` as README-only drift until a future release; keep-alive is handled internally with a default interval.

### State operations (off-chain, two-step pattern)

Each returns `Promise<core.State>`. Settle on-chain afterward with `checkpoint(asset)`.

| Method | Purpose |
|---|---|
| `deposit(blockchainId: bigint, asset: string, amount: Decimal)` | Build + co-sign a deposit state. Creates the channel if none exists. |
| `withdraw(blockchainId: bigint, asset: string, amount: Decimal)` | Co-sign a withdrawal state. |
| `transfer(recipientWallet: string, asset: string, amount: Decimal)` | Co-sign an off-chain transfer. No checkpoint needed on an existing channel. |
| `acknowledge(asset: string)` | Acknowledge a received state (e.g. after an incoming transfer). |
| `closeHomeChannel(asset: string)` | Co-sign a finalize state. |

### On-chain settlement

| Method | Purpose |
|---|---|
| `checkpoint(asset: string): Promise<string>` | Single entry point for all on-chain transactions. Routes to create / checkpoint / close based on channel status and transition type. Returns tx hash. |
| `challenge(state: core.State): Promise<string>` | Submit an on-chain dispute using a co-signed state. |
| `approveToken(chainId: bigint, asset: string, amount: Decimal)` | ERC-20 approve for the ChannelHub. |
| `checkTokenAllowance(chainId: bigint, tokenAddress: string, owner: string): Promise<bigint>` | Read ERC-20 allowance. |

### Node info / queries

| Method | Signature |
|---|---|
| `ping()` | `Promise<void>` — health check |
| `getConfig()` | `Promise<core.NodeConfig>` |
| `getBlockchains()` | `Promise<core.Blockchain[]>` |
| `getAssets(blockchainId?: bigint)` | `Promise<core.Asset[]>` |

### User queries

| Method | Notes |
|---|---|
| `getBalances(wallet: Address): Promise<core.BalanceEntry[]>` | — |
| `getTransactions(wallet, options?)` | Returns `{ transactions, metadata }`. Options: `asset`, `txType`, `fromTime`, `toTime`, `page`, `pageSize`. |
| `getActionAllowances(wallet: Address)` | Gated-action allowances per wallet. |

### Channel queries

| Method | Notes |
|---|---|
| `getChannels(wallet, options?)` | Filter by `status`, `asset`, `channelType`, `pagination`. |
| `getHomeChannel(wallet, asset)` | Single home channel. |
| `getEscrowChannel(escrowChannelId)` | Escrow channel by ID. |
| `getLatestState(wallet, asset, onlySigned)` | Latest state; set `onlySigned=true` for both-signed. |

### App registry

- `getApps({ appId?, ownerWallet?, page?, pageSize? })` → `{ apps: AppInfoV1[], metadata }`
- `registerApp(appID: string, metadata: string, creationApprovalNotRequired: boolean)`

### App sessions (low-level)

- `getAppSessions({ appSessionId?, wallet?, status?, page?, pageSize? })`
- `getAppDefinition(appSessionId): Promise<AppDefinitionV1>`
- `createAppSession(definition, sessionData, quorumSigs, opts?: { ownerSig?: string })`
- `submitAppSessionDeposit(update, quorumSigs, asset, depositAmount: Decimal)`
- `submitAppState(update, quorumSigs)`
- `rebalanceAppSessions(signedUpdates: SignedAppStateUpdateV1[])`

### App-session key operations

- `signSessionKeyState(state)`
- `submitSessionKeyState(state)`
- `getLastKeyStates(userAddress, sessionKey?)`

### Channel-session key operations

- `signChannelSessionKeyState(state)`
- `submitChannelSessionKeyState(state)`
- `getLastChannelKeyStates(userAddress, sessionKey?)`

### Security-token locking (Custody "security" track)

Verified in `client.d.ts`; not in the main cheat sheet:

- `escrowSecurityTokens(targetWallet, blockchainId, amount)`
- `initiateSecurityTokensWithdrawal(blockchainId)`
- `cancelSecurityTokensWithdrawal(blockchainId)`
- `withdrawSecurityTokens(blockchainId, destinationWallet)`
- `approveSecurityToken(chainId, amount)`
- `getLockedBalance(chainId, wallet): Promise<Decimal>`

### Lifecycle / utilities

- `close(): Promise<void>`
- `waitForClose(): Promise<void>` — monitor connection
- `signState(state): Promise<Hex>` — advanced
- `getUserAddress(): Address`
- `setHomeBlockchain(asset, blockchainId)` — required before `transfer()` on a new channel. Immutable per instance.


## Signers

Two interfaces, with concrete classes backed by `viem/accounts`:

```typescript
interface StateSigner {
  getAddress(): Address;
  signMessage(hash: Hex): Promise<Hex>;
}
interface TransactionSigner {
  getAddress(): Address;
  sendTransaction(tx: any): Promise<Hex>;
  signMessage(message: { raw: Hex }): Promise<Hex>;
  signPersonalMessage?(hash: Hex): Promise<Hex>;
  getAccount?(): ReturnType<typeof privateKeyToAccount>;
}
```

### Exported signer classes (verified)

| Class | Role | Prefix byte |
|---|---|---|
| `EthereumMsgSigner` | Channel state signatures (EIP-191 prefixed). Default for `stateSigner`. | — |
| `EthereumRawSigner` | Raw hash signing for on-chain txs. Default for `txSigner`. | — |
| `ChannelDefaultSigner` | Wraps a `StateSigner`; the standard wrapper used internally. | — |
| `ChannelSessionKeyStateSigner` | Session-key-based channel state signing with walletAddress + metadataHash + authSignature. | — |
| `AppSessionWalletSignerV1` | Wraps a `StateSigner` for app-session operations signed by the main wallet. | `0xA1` |
| `AppSessionKeySignerV1` | Wraps a `StateSigner` for app-session operations signed by a delegated session key. | `0xA2` |

### `createSigners(privateKey: Hex)`

```typescript
export declare function createSigners(privateKey: Hex): {
  stateSigner: StateSigner;
  txSigner: TransactionSigner;
};
```

Convenience helper used in every example. Internally instantiates `EthereumMsgSigner` (stateSigner) and `EthereumRawSigner` (txSigner) from the same key.


## v1 RPC method namespaces

Low-level method constants live in `rpc/methods.d.ts`. They are grouped into five namespaces. All method names are exported as `Method` (`string`) constants. The on-wire names follow `<group>.v1.<verb>` shape (`channels.v1.get_home_channel`, `app_sessions.v1.submit_app_state`, …) and the SDK's `RPCClient` has a corresponding camelCase method per constant.

### `channels.v1.*` (group: `ChannelV1Group`)

- `ChannelsV1GetHomeChannelMethod`
- `ChannelsV1GetEscrowChannelMethod`
- `ChannelsV1GetChannelsMethod`
- `ChannelsV1GetLatestStateMethod`
- `ChannelsV1GetStatesMethod`
- `ChannelsV1RequestCreationMethod`
- `ChannelsV1SubmitStateMethod`
- `ChannelsV1SubmitSessionKeyStateMethod`
- `ChannelsV1GetLastKeyStatesMethod`

### `app_sessions.v1.*` (group: `AppSessionsV1Group`)

- `AppSessionsV1SubmitDepositStateMethod`
- `AppSessionsV1SubmitAppStateMethod`
- `AppSessionsV1RebalanceAppSessionsMethod`
- `AppSessionsV1GetAppDefinitionMethod`
- `AppSessionsV1GetAppSessionsMethod`
- `AppSessionsV1CreateAppSessionMethod`
- `AppSessionsV1CloseAppSessionMethod`
- `AppSessionsV1SubmitSessionKeyStateMethod`
- `AppSessionsV1GetLastKeyStatesMethod`

### `apps.v1.*` (group: `AppsV1Group`)

- `AppsV1GetAppsMethod`
- `AppsV1SubmitAppVersionMethod`

### `user.v1.*` (group: `UserV1Group`)

- `UserV1GetBalancesMethod`
- `UserV1GetTransactionsMethod`
- `UserV1GetActionAllowancesMethod`

### `node.v1.*` (group: `NodeV1Group`)

- `NodeV1PingMethod`
- `NodeV1GetConfigMethod`
- `NodeV1GetAssetsMethod`

(Also exports a `Group`, `Method`, and `Event` type alias — all `string`.)


## Core enums (from `core/types.d.ts`)

### `ChannelStatus`
`Void = 0`, `Open = 1`, `Challenged = 2`, `Closed = 3`

> Note: this is the SDK's client-facing `ChannelStatus` enum. It differs from the on-chain Custody contract's richer `Status` lifecycle (`VOID`, `INITIAL`, `ACTIVE`, `DISPUTE`, `FINAL`) — the SDK collapses INITIAL/ACTIVE into `Open` for most dApp use.

### `TransitionType`
`Void = 0`, `Acknowledgement = 1`, `HomeDeposit = 10`, `HomeWithdrawal = 11`, `EscrowDeposit = 20`, `EscrowWithdraw = 21`, `TransferSend = 30`, `TransferReceive = 31`, `Commit = 40`, `Release = 41`, `Migrate = 100`, `EscrowLock = 110`, `MutualLock = 120`, `Finalize = 200`

### `TransactionType`
Same shape as `TransitionType` but with `Transfer = 30` (collapsed) and `Rebalance = 42`.

### Intent constants (these supersede old `StateIntent`)
`INTENT_OPERATE = 0`, `INTENT_CLOSE = 1`, `INTENT_DEPOSIT = 2`, `INTENT_WITHDRAW = 3`, plus v1-new escrow/migration intents `INITIATE_ESCROW_DEPOSIT = 4`, `FINALIZE_ESCROW_DEPOSIT = 5`, `INITIATE_ESCROW_WITHDRAWAL = 6`, `FINALIZE_ESCROW_WITHDRAWAL = 7`, `INITIATE_MIGRATION = 8`, `FINALIZE_MIGRATION = 9`.

> This is the SDK-level intent vocabulary. It differs from the older "OPERATE / INITIALIZE / RESIZE / FINALIZE" enum that appeared in some v0.5-era docs — RESIZE is replaced by explicit `INTENT_DEPOSIT` / `INTENT_WITHDRAW` on an existing channel.

### Other enums
- `ChannelType`: `Home = 1`, `Escrow = 2`
- `ChannelParticipant`: `User = 0`, `Node = 1`
- `ChannelSignerType`: `Default = 0`, `SessionKey = 1`


## App session signing

App-session quorum signatures require the type-byte prefix:

```typescript
import {
  EthereumMsgSigner, AppSessionWalletSignerV1, AppSessionKeySignerV1,
} from '@yellow-org/sdk';
import { packCreateAppSessionRequestV1, packAppStateUpdateV1 } from '@yellow-org/sdk/app/packing';

const msg = new EthereumMsgSigner(privateKey);
const wallet = new AppSessionWalletSignerV1(msg);        // prefixes 0xA1
const sessionKey = new AppSessionKeySignerV1(msgSessKey); // prefixes 0xA2

const hash = packCreateAppSessionRequestV1(definition, sessionData);
const sig = await wallet.signMessage(hash);

const { appSessionId } = await client.createAppSession(definition, sessionData, [sig]);
```

Owner approval (when `registerApp(..., creationApprovalNotRequired=false)`):

```typescript
const ownerSig = await ownerAppSessionSigner.signMessage(createRequestHash);
await client.createAppSession(def, data, quorumSigs, { ownerSig });
```


## Relationship to `@yellow-org/sdk-compat`

The companion package `@yellow-org/sdk-compat` (also `1.2.0`, `peerDependencies: @yellow-org/sdk >=1.2.0`) translates the v0.5.3 API surface to v1 calls. Not every v0.5.3 export survives intact — some are **noops or thin normalizers only**:

### Fully wired (call freely)

- `NitroliteClient.create({ wsURL, walletClient, chainId, blockchainRPCs })` — wraps `Client.create` with a viem `WalletClient`.
- `client.deposit(token, amount)` — bigint raw units; channel created implicitly.
- `client.withdrawal(token, amount)` — raw-unit bigint.
- `client.transfer(destination, allocations[])` — preserves v0.5.3 array shape; divides by token decimals internally before delegating.
- `client.closeChannel()`, `client.resizeChannel({ allocate_amount, token })`, `client.challengeChannel({ state })`, `client.acknowledge(tokenAddress)`.
- Query methods: `getChannels`, `getBalances`, `getLedgerEntries`, `getAppSessionsList`, `getAssetsList`, `getAccountInfo`, `getConfig`, `getBlockchains`, `getActionAllowances`, `getEscrowChannel`, `getChannelData`.
- App sessions: `createAppSession`, `closeAppSession`, `submitAppState`, `getAppDefinition`.
- App registry: `getApps`, `registerApp`.
- Security-token locking: `lockSecurityTokens`, `initiateSecurityTokensWithdrawal`, `cancelSecurityTokensWithdrawal`, `withdrawSecurityTokens`, `approveSecurityToken`, `getLockedBalance`.
- Asset resolution helpers: `resolveToken`, `resolveAsset`, `resolveAssetDisplay`, `getTokenDecimals`, `formatAmount`, `parseAmount`, `findOpenChannel`.
- Session-key helpers: `signChannelSessionKeyState`, `submitChannelSessionKeyState`, `getLastChannelKeyStates`, `signSessionKeyState`, `submitSessionKeyState`, `getLastKeyStates`.
- Auth helpers for apps still using v0.5.3-style handshake: `createAuthRequestMessage`, `createAuthVerifyMessage`, `createAuthVerifyMessageWithJWT`, `createEIP712AuthMessageSigner`.
- Escape hatch: `client.innerClient` exposes the raw v1 `Client`.

### Noop / lightweight-normalizer stubs (compile but do nothing meaningful)

From the compat README, verbatim:

```
createGetChannelsMessage, parseGetChannelsResponse,
createGetLedgerBalancesMessage, parseGetLedgerBalancesResponse,
parseGetLedgerEntriesResponse, parseGetAppSessionsResponse,
createTransferMessage, createAppSessionMessage, parseCreateAppSessionResponse,
createCloseAppSessionMessage, parseCloseAppSessionResponse,
createCreateChannelMessage, parseCreateChannelResponse,
createCloseChannelMessage, parseCloseChannelResponse,
createResizeChannelMessage, parseResizeChannelResponse,
createPingMessage,
convertRPCToClientChannel, convertRPCToClientState,
parseAnyRPCResponse, NitroliteRPC
```

`create*` helpers are placeholders so old imports compile. `parse*` helpers do light normalization on known response shapes. **For new integrations, call `NitroliteClient` methods directly** — don't reach for these stubs.

### Event polling

v0.5.3 relied on server-push notifications (`bu`, `cu`, `tr`, `asu`). v1 is polled. `@yellow-org/sdk-compat` ships an `EventPoller` that re-synthesizes the push model on top of polling:

```typescript
import { EventPoller } from '@yellow-org/sdk-compat';
const poller = new EventPoller(client, {
  onChannelUpdate, onBalanceUpdate, onAssetsUpdate, onError,
}, 5000);
poller.start();
```


## Installation

```bash
npm install @yellow-org/sdk
# or, if migrating from v0.5.3:
npm install @yellow-org/sdk @yellow-org/sdk-compat viem
```

Next.js with Turbopack: add both packages to `transpilePackages` in `next.config.ts`.


## Honest inferences / caveats

- **`withPingInterval` drift**: the README shows it as an Option factory, but `dist/index.d.ts` in `1.2.0` does not re-export it. Use `withHandshakeTimeout`, `withErrorHandler`, `withBlockchainRPC` only; keep-alive is internal.
- **`ChannelStatus` (SDK) vs on-chain `Status`**: the SDK's client-level enum is 4-valued (Void/Open/Challenged/Closed). The on-chain `Status` in the Custody contract is 5-valued (VOID/INITIAL/ACTIVE/DISPUTE/FINAL). Don't conflate them.
- **Intent vocabulary**: v1 uses numeric `INTENT_OPERATE=0`, `INTENT_CLOSE=1`, `INTENT_DEPOSIT=2`, `INTENT_WITHDRAW=3` plus escrow/migration intents 4–9. The older v0.5-era `StateIntent` enum name still shows up in some docs but is not exported by v1; map to these constants instead.
- **Raw vs human amounts**: the v1 `Client` takes `decimal.js` human amounts (`new Decimal(100)`). The compat layer takes raw `bigint` smallest-units; it converts before delegating. Don't mix.
- **Example file reference**: the README mentions `examples/app_sessions/lifecycle.ts`, but examples are not shipped in the `1.2.0` tarball — only `dist/`, `README.md`, `package.json`.


## Reference

- npm: https://www.npmjs.com/package/@yellow-org/sdk
- compat: https://www.npmjs.com/package/@yellow-org/sdk-compat
- Repo: https://github.com/layer-3/nitrolite
- Verified tarball: `@yellow-org/sdk@1.2.0`, sha512 `m7TBzOs1J9fWT55DwZjuiv6bo9gAPyMri1D+11Do3u462oFZqo/Vo5SM/hivpX5sfSDTD0TBbO2qAZ612Jaklg==`
