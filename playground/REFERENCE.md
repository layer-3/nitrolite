# Playground Reference

Developer-facing playground for Nitrolite — wallet, channel, and state channel operations.

**Stack**: React 19, TypeScript, Vite, Tailwind CSS, viem, @yellow-org/sdk, sonner (toasts), lucide-react (icons)

---

## Page Layout

Single-page app. Sticky header, scrollable body.

- **Header**: Wallet connection bar (left: brand + node status; center: Main/History/Session Keys tabs when connected; right: SK chip + chain + address + disconnect)
- **Body — Main tab (2-col on desktop)**:
  - Left: Action panel (deposit / withdraw / transfer / faucet tabs)
  - Right: Channel list (includes incoming unacknowledged channels at the top)
- **Body — History tab (full-width)**:
  - Transaction history table with column-filter popovers, per-cell quick filter, expandable row detail, and pagination (25/page)
- **Body — Session Keys tab (full-width)**:
  - Table of all session keys (active, expiring, expired, revoked) fetched from the node
  - Register, update (renew), and revoke keys via modals

Tab selector only appears when a wallet is connected. Switching tabs preserves state in both panels.

**Gating overlays** (modals) appear on top when:
- Wallet is on an unsupported chain
- User sets a home chain for an asset
- User sets up or renews a session key

---

## Components

### WalletBar
**For**: Wallet identity, chain context, and node health at a glance.
- Connect / disconnect MetaMask
- Display current address and active chain
- Session key chip: shows time remaining, allows clearing the key
- Node status: shows last successful communication timestamp, or error if unreachable
- Tab selector: Main / History / Session Keys (shown when wallet is connected)

### ActionPanel
**For**: All channel money operations initiated by the user.
- **Deposit tab** — move on-chain funds into a channel (requires MetaMask approval + transaction)
- **Withdraw tab** — move channel funds back on-chain
- **Transfer tab** — send channel funds to another address (requires recipient address input)
- **Faucet tab** — visible only for assets in `FAUCET_ASSETS` (currently YUSD); shows a "Request drip" button that calls the faucet endpoint (mock until endpoint is configured)
- Token selector (custom dropdown via TokenSelector), amount input, Max button (auto-fills from relevant balance)
- Enforces amount limits: cannot exceed on-chain balance (deposit) or channel balance (withdraw/transfer)
- **Cross-chain guard**: if the selected asset has a home channel on a different chain than the wallet's current chain, the deposit/withdraw/transfer form is blurred and a "Select {chain}" button appears to switch chains; the Faucet tab remains accessible regardless of chain state

### TokenSelector
**For**: Custom asset/token picker used inside ActionPanel.
- Displays each asset as a row: token icon + symbol + supported chain icons
- Assets in `FAUCET_ASSETS` show a small drip icon (Droplets) with a "Faucet available" tooltip next to the symbol
- Token and chain icons are loaded from CDN (see `src/icons.ts`); unknown symbols fall back to a letter avatar
- If a non-closed home channel exists for an asset, all chain icons except the home chain are dimmed (greyscale + low opacity)

### ChannelList
**For**: Overview of all channels and incoming unacknowledged states for the connected address.
- Refresh button to re-fetch channels from node
- Renders `IncomingChannelRow` for assets that have a balance but no open channel (incoming, unacknowledged)
- Renders a `ChannelRow` per established channel

### ChannelRow
**For**: Per-channel identity, status, and drill-down into state.
- Displays asset symbol, home chain, truncated channel ID with copy button
- Status badges: "wrong chain" (wallet is on a different chain), "closed"
- Expand/collapse to see channel state detail (StateViewer) or a closed notice
- Inline prompt to switch wallet chain when it doesn't match the channel's home chain
- Close channel button
- Does **not** show a session-key selector — the active SK is wallet-global (not per-channel); manage it via WalletBar chip + Session Keys tab

### StateViewer
**For**: Inspecting and acting on the three layers of channel state.
- **Enforced** — last checkpoint committed on-chain; lowest trust, highest finality
- **Signed** — both parties have signed; ready to checkpoint on-chain
- **Issued** — node has proposed; needs acknowledgement before it becomes signed
- Each layer shows version number and balance
- **Acknowledge** button: accepts issued state (moves to signed)
- **Checkpoint** button: commits signed state on-chain

### IncomingChannelRow
**For**: An asset that has an issued (node-proposed) state but no acknowledged channel yet.
- Shows a "NO HOME CHAIN" pill (reddish, low contrast) with tooltip explaining that the wallet's current chain becomes the home chain on acknowledge
- Expands to show the issued state (version + amount) in the same format as StateViewer
- **Acknowledge** button: calls `setHomeBlockchain(asset, currentChainId)` then `acknowledge(asset)`; after success the full channel list refreshes and the row transitions into a regular ChannelRow

### SessionKeyBanner
**For**: Nudging the user to set up a session key when none is active.
- Appears when wallet is connected but no session key exists
- "Set up" button navigates to the Session Keys tab

### SessionKeysTab
**For**: Full management of session keys (list, register, update/renew, revoke, reactivate).
- Fetches all session keys (active + inactive) from the node via `useSessionKeyManagement`
- Displays a table with address (truncated + copy), assets, expiration (3-way format toggle: relative / UTC date / Unix), status badge, version, and per-row actions
- "Expiring Soon" banner at the top when any key has < 1 hour remaining
- Register New opens `SessionKeyRegisterForm` in register mode
- Per-row actions depend on status and whether the key's private key is still in localStorage (`hasLocalKey`):
  - **active** → Update, Revoke, Use (if not current)
  - **expiring** → Renew, Revoke, Use (if not current)
  - **expired/revoked + hasLocalKey** → Reactivate (re-registers via update flow at version+1 with future expiry)
  - **expired/revoked + !hasLocalKey** → no row actions (cannot update without the private key)
- Revoke opens `SessionKeyRevokeModal` (sets expiresAt to past)
- "IN USE" badge on the currently active key

### SessionKeySetupModal
**For**: Legacy modal for quick session key creation (still present but no longer wired from the banner).
- Clarifies what a session key is (24h authorization stored locally, avoids MetaMask popups per state op)
- Confirm triggers one MetaMask signature, then session key is stored in localStorage
- Cancel dismisses without changes

### SetHomechainModal
**For**: Choosing which chain an asset settles on when performing a transfer.
- Appears automatically when a transfer cannot proceed because no home chain is set
- Radio list of eligible chains (those that support the asset)
- Confirm sets the home chain, then the pending transfer is retried automatically

### UnsupportedChainModal
**For**: Guiding the user off an unsupported chain.
- Appears when the connected wallet's chain is not recognized by the node
- Lists supported chains; clicking one triggers a wallet chain-switch request

### HistoryTab
**For**: Full-width transaction history view, shown when the History tab is active.
- Fetches up to 200 recent transactions via `client.getTransactions(address, { pageSize: 200 })`; sorts newest-first; paginates at 25/page client-side
- Column header popovers (funnel icon) for Type (multi-select checkboxes), Asset (multi-select), From/To (text input); Apply/Clear buttons in each popover
- Per-cell quick filter: clicking a Type badge, Asset name, or From/To address immediately adds/removes an exact-value filter with tooltip feedback
- Expandable rows: clicking a row reveals Sender new state ID, Receiver new state ID, Timestamp, and a Confirmation timeline (Signed → Co-signed for off-chain; Signed → Broadcasted → Confirmed for on-chain)
- All filtering is client-side after the initial fetch; Refresh button re-fetches from the node

### CopyButton
**For**: One-click copy of addresses or hashes throughout the UI.
- Checkmark feedback for 1.5s after copy

---

## Hooks

### useWallet
**Owns**: MetaMask connection lifecycle.
- Wallet address, chain ID, viem WalletClient
- Connect / disconnect / switch chain actions
- Passively detects account and chain changes from MetaMask events

### useNitrolite
**Owns**: SDK client lifecycle and all node-level data.
- Client creation: probes node for supported chains/assets, then builds final authenticated client
- Session key vs wallet signing: uses session key for state ops if one is active; wallet otherwise
- Channel balances (from node) and on-chain balances (from each chain's RPC)
- Node connection state, error state, last-comms timestamp

### useChannels
**Owns**: List of channels for the connected address.
- Fetches from node on demand, surfaces loading and error states

### useChannelStates
**Owns**: The three-layer state (enforced / signed / issued) for one channel.
- Determines whether Acknowledge and Checkpoint actions are available
- Runs acknowledge (sign + upload) and checkpoint (on-chain transaction) operations
- Refreshes balances after each successful operation

### useChannelOps
**Owns**: High-level channel operations triggered from ActionPanel.
- Deposit: token allowance check → approve if needed → deposit → checkpoint → wait for receipt (`confirming`) → if `confirmationDelaySecs > 0`, enter `awaiting_node` phase and poll `getBalances` until enforced balance rises (bounded by `min(confirmationDelaySecs, 60) + 15s`), then emit success or soft-fallback toast
- Withdraw: withdraw → checkpoint → wait for receipt (`confirming`) → if `confirmationDelaySecs > 0`, enter `awaiting_node` and poll enforced balance down; same cap and fallback as deposit
- Transfer: transfer → if home chain missing, shows modal → retries after chain is set
- Close: close channel → checkpoint → wait for receipt → if `confirmationDelaySecs > 0`, poll enforced balance down; same bounded wait
- `DepositPhase` and `WithdrawPhase` include `'awaiting_node'` after `'confirming'`; gate-disabled (`delaySecs === 0`) flows skip `'awaiting_node'` and emit the immediate success toast
- Accepts `supportedChains: Blockchain[]` to look up per-chain `confirmationDelaySecs`
- Tracks operation loading states per action; cancels stale ops on address/wallet change
- Distinguishes user rejections (MetaMask code 4001) from real errors for toast messaging

### useSessionKey
**Owns**: Session key storage and registration.
- Loads session key from localStorage on address change; also loads `allKeys` (all stored keys)
- Registers a new key: generates keypair, finds next version, signs ownership with wallet, submits to node
- `selectKey(address)` — switches the active signing key among locally stored keys
- Clears key from storage
- Re-checks expiry every 60s so the banner re-appears in the same session if a key expires

### useSessionKeyManagement
**Owns**: Server-side session key fetching and register/update/revoke operations.
- `fetchKeys()` — fetches all keys (including inactive) from the node via `client.getLastChannelKeyStates`
- `register(client, walletAddress, assets, expiresAt)` — registers a new key, saves to localStorage, returns StoredSessionKey
- `update(client, walletAddress, currentKey, assets, expiresAt)` — re-registers with new assets/expiry at the next version
- `revoke(client, walletAddress, currentKey)` — re-registers with `expiresAt` in the past to invalidate the key
- Exposes `serverKeys`, `isLoading`, `isSubmitting`

---

## Utilities

### icons.ts
- Static mapping of token symbol → CDN icon URL (CoinGecko), and chain ID → CDN icon URL (llamao.fi)
- Testnets map to their parent mainnet icon
- `tokenIconUrl(symbol)` / `chainIconUrl(chainId)` return `null` for unknown entries; callers render a fallback

### utils.ts
- `FAUCET_ASSETS` — Set of lowercase asset symbols that have a test faucet (currently `yusd`)
- Address formatting (abbreviated display)
- Balance formatting (thousands separators, 2 decimal places)
- Relative time ("just now", "5m ago", "2h ago")
- Ethereum address validation

### networks.ts
- Public RPC URL registry per chain ID
- Lookup helper used when building the SDK client

### sessionKey.ts
- localStorage key scheme: `nitrolite_playground_sk_{nodeUrl}::{walletAddress}`
- Expiry: 24h, with 5-min renewal buffer
- Load / save / clear helpers
- `registerSessionKey` — full registration flow: keypair generation, node version discovery, wallet signature, node submission
- `buildSessionKeyStateSigner` — wraps stored key into SDK StateSigner interface

### walletSigners.ts
- `WalletStateSigner` — state signing backed by MetaMask (fallback when no session key)
- `WalletTransactionSigner` — on-chain transaction signing backed by MetaMask (always used for on-chain ops)

---

## Key User Flows

**First visit**
1. "Connect MetaMask" prompt in body → connect → node probes and client builds → channels load

**Deposit**
1. Select asset + amount → Deposit → MetaMask: approve token spend (once) → MetaMask: deposit transaction → checkpoint written on-chain → wait for tx receipt → if `confirmationDelaySecs > 0`, poll for enforced balance credit (up to `min(delaySecs, 60) + 15s`) → success toast (or soft-fallback if poll times out)

**Withdraw**
1. Select asset + amount → Withdraw → MetaMask: withdraw transaction → checkpoint written on-chain → wait for tx receipt → if `confirmationDelaySecs > 0`, poll for enforced balance decrease → success toast (or soft-fallback)

**Transfer**
1. Select asset + amount + recipient address → Transfer → if no home chain set, modal appears → confirm chain → transfer proceeds

**Close channel**
1. In ActionPanel or ChannelRow → Close → MetaMask: close transaction → checkpoint

**Session key setup**
1. Banner appears → "Set up" → navigates to Session Keys tab → "Register New" → `SessionKeyRegisterForm` → confirm → one MetaMask signature → key stored locally + submitted to node → future state ops (acknowledge, checkpoint) skip MetaMask

**Session key management**
1. Session Keys tab → table of all keys from node (active, expiring, expired, revoked)
2. Update/Renew key → `SessionKeyRegisterForm` in update mode (new version registered on node)
3. Revoke key → `SessionKeyRevokeModal` confirm → re-registers with expiresAt in the past
4. Switch active key → "Use" button → `selectKey` updates localStorage active flag + SDK switches signer

**Acknowledge issued state**
1. Expand channel → StateViewer → Acknowledge button (issued row) → signs with session key or wallet → issued becomes signed

**Checkpoint signed state**
1. Expand channel → StateViewer → Checkpoint button (signed row) → MetaMask: on-chain transaction → signed becomes enforced

**Accept incoming channel**
1. Channels section → expand IncomingChannelRow → Acknowledge → `setHomeBlockchain` sets wallet chain as home → `acknowledge` co-signs the issued state → channel appears as a regular ChannelRow

---

## Environment

| Variable | Purpose |
|---|---|
| `VITE_NITRONODE_URL` | WebSocket URL of the Nitronode backend (e.g. `wss://nitronode-sandbox.yellow.org/v1/ws`) |
