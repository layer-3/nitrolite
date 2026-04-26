---
name: yellow-swap-design
description: |
  How to build a swap (off-chain asset exchange) on Yellow Network when the
  protocol does not provide a canonical swap primitive. Covers the three
  realistic design patterns — treasury-convert, market-maker agent, and P2P
  matching — the trade-offs of each, why direct routing through "Yellow
  brokers" or yellow.pro is not currently possible, and what would unlock a
  v2 with real counterparty matching. Use when: deciding swap architecture
  for a Yellow-Network app, scoping how much liquidity ops you have to run
  yourself, or answering "can I just plug into the Yellow swap API?"
version: 1.0.0
network: mainnet
last_verified: 2026-04-26
---

# Designing a swap on Yellow Network

There is no native Yellow Network protocol primitive for cross-asset swap.
This skill explains why, and what to build instead.

## What the protocol gives you

- **State channels** for bilateral balance accounting (ERC-7824 / Nitrolite).
- **App sessions** for multi-party off-chain accounts with allocation rules.
- **Transfers** over Nitro RPC (instant, gasless, single-asset move between
  participants in the same channel set).
- **Escrow tx types** (`EscrowLock`, `EscrowUnlock`) inside the ledger.
- A sandbox ClearNode at `wss://clearnet-sandbox.yellow.com/ws` (free, no
  partnership) and mainnet at `wss://clearnet.yellow.com/ws`.

## What the protocol does NOT give you

- **No order book.** No on- or off-chain matching engine.
- **No quote feed.** No `get_quote` / `submit_order` / `cancel_order` RPC.
- **No public broker registry.** "Brokers" in Yellow's vocabulary are
  **ClearNode operators** (lock 250,000 $YELLOW to run a node) — they
  validate ledger state, not match trades.
- **No canonical `swap-v1` `application` protocol.** Spec lists
  `payment-app-v1`, `gaming-app-v1`, `escrow-app-v1`, `tournament-v1`,
  `subscription-v1` as documented examples — swap is your job to design.
- **No "yellow.pro" public API.** yellow.pro is a Yellow-built consumer UI
  for trading; it runs market-makers on top of the same protocol you're on.
  It is not exposed as backend infrastructure to third-party platforms.

## Three patterns to ship

### 1. Treasury-convert (1-party, simplest)

Your platform holds a $YELLOW + ETH/USDC inventory. When a user wants to
swap, you debit one side from their unified balance via `transfer` and
credit the other side from your treasury. Price is whatever your oracle
says.

```text
user --transfer(YELLOW)--> platform-treasury
platform-treasury --transfer(ETH)--> user
```

- **Pros:** ~50 lines of code, instant, predictable UX.
- **Cons:** the platform takes price risk on every trade and must keep both
  sides liquid. Breaks the moment volume is one-directional.
- **When OK:** v1 / closed beta / capped per-user, per-day amounts.
  Label it honestly ("Treasury Convert" / "v1") so users understand the
  model and the caps.
- **Reference impl:** YellowHive `SwapService.executeSwap` is exactly this
  pattern. Capped at 1k $YELLOW per trade.

### 2. Market-maker agent (2-party app session)

Run a dedicated agent ("MM-bot") whose only job is to provide swap
liquidity. When a user requests a quote:

1. MM-bot opens a **2-party app session** with the user
   (`participants: [user, mm]`, `weights: [50, 50]`, `quorum: 100`).
2. MM-bot signs an initial state with `allocations: { user: -inputAmount,
   mm: -outputAmount }` — both sides locked.
3. User co-signs to atomically execute.
4. App session closes; both ledgers update.

- **Pros:** clean accounting boundary (the MM-bot's risk is isolated from
  the rest of the platform). Same UX as treasury-convert from the user's
  side.
- **Cons:** still needs you to fund the MM-bot. Still doesn't get you
  external liquidity.
- **When OK:** scaling beyond v1 caps. You can also run multiple MM-bots
  with different fee curves to A/B price elasticity.
- **`application` field**: free-form string. Pick a stable name per
  protocol version (e.g. `"swap-mm-v1"`) and version it. There is no
  registry — your name only needs to make sense to your own backend.

### 3. P2P matching (multi-party, async)

Users post quotes / orders ("I'll trade 100 YELLOW for 0.002 ETH, valid
60s") into your platform. Other users accept. Acceptance opens an
`escrow-app-v1` 2-party app session that swaps the assets atomically.

- **Pros:** no inventory risk for the platform.
- **Cons:** thin liquidity unless the platform has many users with
  matching needs. You still build + run the order book yourself.
- **When OK:** community is liquid enough that the average wait is < 30s.

## Why you can't "just route through yellow.pro"

`yellow.pro` is a Yellow-built consumer trading product. It is not a
service you can call from a backend. There is no documented `POST
/api/quote` endpoint on yellow.pro for partners. The same applies to
`api.yellow.org` mentioned in older Medium articles — that endpoint is
not a public swap router today.

If you want infrastructure-level swap routing in the future, the
realistic asks of the Yellow Network team are:

1. Publish a canonical `application` value (e.g. `"swap-v1"`) with a
   reference matching protocol so partner platforms agree on the shape.
2. Run a hosted "Yellow market-maker" bot (with a fixed counterparty
   address, same as how Uniswap V2 has a known router contract) that any
   platform can pair with.
3. Open the matching engine that powers yellow.pro to platform partners
   under a documented broker SLA.

Until any of those land, every swap on Yellow Network is one of the three
patterns above. Pick the one that matches your liquidity model.

## What to put in the app session for a swap

The v1 SDK uses `appSessionsV1CreateAppSession` to open the session and
`submitAppState` with `intent: 'close'` to close — there's no dedicated
`close_app_session` method in v1 (see `yellow-sdk-v1` reference). The
v0.5.3 SDK still exposes `sendCreateAppSession` / `sendCloseAppSession`
helpers as a separate code path.

```ts
// v1 — @yellow-org/sdk
import { appSessionsV1CreateAppSession, submitAppState } from '@yellow-org/sdk';

const swap = await appSessionsV1CreateAppSession(signer, requestSigner, {
  definition: {
    application: 'swap-mm-v1',         // your name, free-form
    protocol:    'NitroRPC/0.4',       // NitroRPC version (NOT app type)
    participants: [userAddress, mmAddress],
    weights:     [50, 50],
    quorum:      100,                   // unanimous — both must sign
    challenge:   86_400,                // 24h dispute window
    nonce:       Date.now(),
  },
  allocations: [
    { participant: userAddress, asset: 'yellow', amount: '100' },
    { participant: mmAddress,   asset: 'eth',    amount: '0.002' },
  ],
});

// MM-bot pre-signed the desired ending state where the assets cross.
// User co-signs → submit_app_state with intent="close" releases the funds.
await submitAppState(signer, requestSigner, {
  app_session_id: swap.app_session_id,
  intent:  'close',                     // <- the v1 way to close a session
  version: 1,                            // current_version + 1
  allocations: [
    { participant: userAddress, asset: 'eth',    amount: '0.002' },
    { participant: mmAddress,   asset: 'yellow', amount: '100' },
  ],
});
```

The fee (your platform's cut) goes into a third allocation pointing at
your treasury, OR you net it from the MM-bot's output before signing.

## Related skills

- `yellow-app-sessions` — full reference for the app session primitive.
- `yellow-transfers` — single-asset move; what treasury-convert uses.
- `yellow-state-channels` — channel lifecycle the app sessions live in.
- `yellow-network-builder` — start here if you're new to the protocol.
