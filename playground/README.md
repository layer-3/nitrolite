# Nitrolite Playground

Developer-facing UI for Nitrolite state channels. Connects MetaMask to a Nitronode and exposes the SDK's high-level operations: deposit, withdraw, transfer, close, plus a per-channel state inspector (Enforced / Signed / Issued) and optional session-key delegation that skips MetaMask popups for off-chain state co-signing.

## Quick start

```bash
# From repo root
cd sdk/ts && npm install && npm run build    # build SDK once (playground symlinks to sdk/ts/dist)
cd ../../playground
npm install
cp .env.example .env                          # edit VITE_NITRONODE_URL if needed
npm run dev                                   # http://localhost:3001
```

A MetaMask install is required. The playground requests an account and listens for `accountsChanged` / `chainChanged`.

**Important:** `playground/node_modules/@yellow-org/sdk` is a symlink to `sdk/ts`. If the SDK source changes, rerun `cd sdk/ts && npx tsc` (or `npm run build`) to refresh `sdk/ts/dist/`; the playground picks the new build up on the next dev-server reload.

## Configuration

Only one env var:

| Var | Purpose |
|-----|---------|
| `VITE_NITRONODE_URL` | WebSocket URL of the Nitronode (defaults to `wss://nitronode-sandbox.yellow.org/v1/ws`) |

Supported chains are discovered from the Nitronode's `getConfig()`. Public RPC URLs for each chain are looked up at runtime; override via `src/networks.ts` if a chain is missing or you want a private endpoint.

## On-chain confirmation delay

Each chain reported by the Nitronode's `getConfig()` includes `confirmationDelaySecs`. After an on-chain
operation (deposit, withdraw, close) the node waits that many seconds before crediting the result to your
off-chain balance — a reorg-safety gate. The transaction is mined first; the balance updates only after
the delay (a few seconds where the gate is enabled, up to ~13 min on Ethereum L1; `0` = disabled).
Off-chain transfers are not gated and reflect immediately.

## Layout

| File | Purpose |
|------|---------|
| `src/hooks/useWallet.ts` | MetaMask EIP-1193 lifecycle |
| `src/hooks/useNitrolite.ts` | Nitrolite Client lifecycle, balances, signer selection (wallet vs session key) |
| `src/hooks/useChannelOps.ts` | deposit / withdraw / transfer / close with cancellation on wallet switch |
| `src/hooks/useChannels.ts` | `getChannels()` results |
| `src/hooks/useChannelStates.ts` | per-channel Enforced / Signed / Issued + Acknowledge / Checkpoint actions |
| `src/hooks/useSessionKey.ts` | localStorage-backed session-key state + register/clear |
| `src/sessionKey.ts` | storage + register helper + signer construction |
| `src/components/WalletBar.tsx` | top nav: wallet, chain, node status, session-key chip |
| `src/components/ActionPanel.tsx` | balance + tabbed Deposit / Withdraw / Transfer |
| `src/components/ChannelList.tsx` + `ChannelRow.tsx` + `StateViewer.tsx` | channel display |
| `src/components/PendingReceipts.tsx` | assets received but not yet acknowledged |
| `src/components/UnsupportedChainModal.tsx` | blocking modal when wallet on unsupported chain |
| `src/components/SetHomechainModal.tsx` | first transfer for an asset needs a home chain |
| `src/components/SessionKeyBanner.tsx` + `SessionKeySetupModal.tsx` | session-key onboarding UI |

## Session keys

After connecting MetaMask, a banner offers to set up a session key — a temporary key (default 24 h) that signs state updates on your behalf so deposit / withdraw / transfer / acknowledge / checkpoint stop popping MetaMask **for the off-chain co-sign step**. On-chain transactions (the actual deposit / withdraw / checkpoint txs, plus the ERC-20 `approve`) still require MetaMask — a session key cannot sign on-chain calls.

Key facts:

- Scope: all currently supported assets (`getAssets()` on connect).
- Storage: `localStorage` under `nitrolite_playground_sk_<nodeUrl>::<walletAddrLower>`. Per-node, per-wallet.
- Expiry: re-checked on load and once a minute. Expired entries are cleared automatically; the banner returns as "Renew".
- Clear: the `✕` on the SK chip in the wallet bar drops the local key. The node-side delegation remains valid until natural expiry; it just stops being used locally.

ERC-20 approvals use the "infinite approve" pattern: the first deposit for a token approves `floor((2^256 − 1) / 10^decimals)` human units (≈ MaxUint256 in smallest units), so subsequent deposits skip the approval popup.

## Not in scope

- Programmatic SK revoke (submit version+1 with `assets=[]`). Local clear only.
- Per-asset SK scoping (always all assets).
- Transaction history panel.
- App sessions.
- Auto-polling. State is refreshed after each operation and on the refresh button. Note: a deposit's
  off-chain credit lags the on-chain tx by the chain's confirmation delay (see "On-chain confirmation
  delay" above), so a refresh immediately after a deposit may not show the new balance yet — refresh
  again once the delay has elapsed.

See `playground/TODO.md` (in the original design pack) for the full deferred list.

## Stack

- React 19, TypeScript, Vite 5
- Tailwind 3 with custom CSS variables (see `src/index.css`)
- `@yellow-org/sdk` (file: linked to `sdk/ts`)
- `viem` for wallet + on-chain reads
- `sonner` for toasts
- `lucide-react` for icons
