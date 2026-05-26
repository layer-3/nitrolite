# Playground Reference

Developer-facing playground for Nitrolite — wallet, channel, and state channel operations.

**Stack**: React 19, TypeScript, Vite, Tailwind CSS, viem, @yellow-org/sdk, sonner (toasts), lucide-react (icons)

---

## Page Layout

Single-page app. Sticky header, scrollable body.

- **Header**: Wallet connection bar + session key status + node health indicator
- **Body (2-col on desktop)**:
  - Left: Action panel (deposit / withdraw / transfer tabs)
  - Right: Channel list + pending receipts section

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

### ActionPanel
**For**: All channel money operations initiated by the user.
- **Deposit tab** — move on-chain funds into a channel (requires MetaMask approval + transaction)
- **Withdraw tab** — move channel funds back on-chain
- **Transfer tab** — send channel funds to another address (requires recipient address input)
- Token selector (custom dropdown via TokenSelector), amount input, Max button (auto-fills from relevant balance)
- Enforces amount limits: cannot exceed on-chain balance (deposit) or channel balance (withdraw/transfer)
- **Cross-chain guard**: if the selected asset has a home channel on a different chain than the wallet's current chain, the tabs and form are blurred and a "Select {chain}" button appears to switch chains (cross-chain operations are not supported)

### TokenSelector
**For**: Custom asset/token picker used inside ActionPanel.
- Displays each asset as a row: token icon + symbol + supported chain icons
- Token and chain icons are loaded from CDN (see `src/icons.ts`); unknown symbols fall back to a letter avatar
- If a non-closed home channel exists for an asset, all chain icons except the home chain are dimmed (greyscale + low opacity)

### ChannelList
**For**: Overview of all channels the connected address participates in.
- Refresh button to re-fetch channels from node
- Renders a ChannelRow per channel

### ChannelRow
**For**: Per-channel identity, status, and drill-down into state.
- Displays asset symbol, home chain, truncated channel ID with copy button
- Status badges: "wrong chain" (wallet is on a different chain), "closed"
- Expand/collapse to see channel state detail (StateViewer) or a closed notice
- Inline prompt to switch wallet chain when it doesn't match the channel's home chain
- Close channel button

### StateViewer
**For**: Inspecting and acting on the three layers of channel state.
- **Enforced** — last checkpoint committed on-chain; lowest trust, highest finality
- **Signed** — both parties have signed; ready to checkpoint on-chain
- **Issued** — node has proposed; needs acknowledgement before it becomes signed
- Each layer shows version number and balance
- **Acknowledge** button: accepts issued state (moves to signed)
- **Checkpoint** button: commits signed state on-chain

### PendingReceipts
**For**: Assets received off-chain that haven't been acknowledged yet (no open channel).
- Lists assets with a non-zero balance but no active channel
- **Acknowledge** button opens the channel by accepting the received balance

### SessionKeyBanner
**For**: Nudging the user to set up a session key when none is active.
- Appears when wallet is connected but no session key exists
- "Set up" button opens SessionKeySetupModal

### SessionKeySetupModal
**For**: Explaining and confirming session key creation or renewal.
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
- Deposit: token allowance check → approve if needed → deposit → checkpoint
- Withdraw: withdraw → checkpoint
- Transfer: transfer → if home chain missing, shows modal → retries after chain is set
- Close: close channel → checkpoint
- Tracks operation loading states per action; cancels stale ops on address/wallet change
- Distinguishes user rejections (MetaMask code 4001) from real errors for toast messaging

### useSessionKey
**Owns**: Session key storage and registration.
- Loads session key from localStorage on address change
- Registers a new key: generates keypair, finds next version, signs ownership with wallet, submits to node
- Clears key from storage
- Re-checks expiry every 60s so the banner re-appears in the same session if a key expires

---

## Utilities

### icons.ts
- Static mapping of token symbol → CDN icon URL (CoinGecko), and chain ID → CDN icon URL (llamao.fi)
- Testnets map to their parent mainnet icon
- `tokenIconUrl(symbol)` / `chainIconUrl(chainId)` return `null` for unknown entries; callers render a fallback

### utils.ts
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
1. Select asset + amount → Deposit → MetaMask: approve token spend (once) → MetaMask: deposit transaction → checkpoint written on-chain

**Withdraw**
1. Select asset + amount → Withdraw → MetaMask: withdraw transaction → checkpoint written on-chain

**Transfer**
1. Select asset + amount + recipient address → Transfer → if no home chain set, modal appears → confirm chain → transfer proceeds

**Close channel**
1. In ActionPanel or ChannelRow → Close → MetaMask: close transaction → checkpoint

**Session key setup**
1. Banner appears → "Set up" → modal → confirm → one MetaMask signature → key stored locally → future state ops (acknowledge, checkpoint) skip MetaMask

**Acknowledge issued state**
1. Expand channel → StateViewer → Acknowledge button (issued row) → signs with session key or wallet → issued becomes signed

**Checkpoint signed state**
1. Expand channel → StateViewer → Checkpoint button (signed row) → MetaMask: on-chain transaction → signed becomes enforced

**Accept pending receipt**
1. PendingReceipts section → Acknowledge → channel opens with received balance

---

## Environment

| Variable | Purpose |
|---|---|
| `VITE_NITRONODE_URL` | WebSocket URL of the Nitronode backend (e.g. `wss://nitronode-sandbox.yellow.org/v1/ws`) |
