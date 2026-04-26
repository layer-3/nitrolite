---
name: yellow-deposits-withdrawals
description: |
  On-chain funding flows for Yellow Network — how to move tokens from your wallet into the unified balance (deposit) and back out (withdraw). Covers ERC-20 `approve` + `deposit`, native ETH handling, `withdraw` via Custody contract, the relationship between deposits and channel creation in v0.5+ (channels now open zero-balance then resize), supported chains, and common gotchas. Use when: funding an agent for the first time, cashing out to an EOA, debugging why a balance doesn't appear, or writing a native (non-SDK) on-chain integration.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/on-chain/data-structures
  - https://docs.yellow.org/docs/0.5.x/guides/migration-guide
---

# Yellow Deposits & Withdrawals

The on-chain gateway between your EOA/smart wallet and the off-chain
**unified balance** that powers transfers, swaps, and app sessions.

## Model

```text
 EOA / Smart Wallet
      │
      │  (1) ERC-20 approve(Custody, amount)
      │  (2) Custody.deposit(token, amount)
      ▼
 Custody contract  ────────────────────▶  Unified balance  (off-chain)
      ▲                                        │
      │                                        │  transfers, app-sessions,
      │  Custody.withdraw(token, amount)       │  swaps, tipping, ...
      │                                        │
      └────────────────────────────────────────┘
```

**v0.5+ change**: `create_channel` opens with **zero balance**. Depositing
does NOT implicitly fund a channel — deposits credit your unified balance.
To move funds from unified balance into a channel, call `resize_channel`
with a positive `resize_amount`. To move funds back, use a negative
`resize_amount` or `close_channel`.

## Depositing ERC-20

Two transactions — approve, then deposit:

```ts
import { erc20Abi, parseUnits } from 'viem';

const custody = '0x...custody...';  // from get_config.networks[].custody_address
const token   = '0x2791Bca1...';    // from get_assets
const amount  = parseUnits('100', 6); // 100 USDC (decimals: 6)

// 1. Approve
await walletClient.writeContract({
  address: token,
  abi: erc20Abi,
  functionName: 'approve',
  args: [custody, amount],
});

// 2. Deposit
await walletClient.writeContract({
  address: custody,
  abi: custodyAbi,            // contains `deposit(address token, uint256 amount)`
  functionName: 'deposit',
  args: [token, amount],
});
```

The ClearNode observes the on-chain event and credits your unified balance.
Expect a 1-2 block delay; subscribe to `bu` notifications to see the credit
arrive.

## Depositing native ETH

Native deposits skip approval — send value directly:

```ts
await walletClient.writeContract({
  address: custody,
  abi: custodyAbi,
  functionName: 'deposit',
  args: [ETH_PLACEHOLDER, 0n],   // placeholder address for native ETH (chain-specific)
  value: parseEther('0.5'),
});
```

The ETH placeholder is conventionally the zero address `0x0000...0000` in
the Nitrolite stack. Check `get_assets` for the exact `token` value used
for native on your target chain.

## Withdrawing

Withdraw pulls from unified balance back to your EOA:

```ts
await walletClient.writeContract({
  address: custody,
  abi: custodyAbi,
  functionName: 'withdraw',
  args: [token, amount],
});
```

Requires a sufficient unified-balance for the asset. If funds are locked in
a channel or app session, they're not available for withdraw — close/resize
first.

## Channel deposits (distinct from unified-balance deposits)

Moving funds **into a channel** (after v0.5) is always `resize_channel` —
never a direct on-chain call. The ClearNode returns a signed resize state,
you submit `Custody.resize(...)` on-chain, and both sides update.

Moving funds **out of a channel** is either:
- `close_channel` (cooperative, instant, one tx) — closes and releases, or
- `resize_channel` with negative `resize_amount` (keeps channel open).

## Supported chains

**Always call `get_config` first** — the list below is a snapshot and the
authoritative answer for your ClearNode lives in `get_config.networks[]`.

Whitepaper enumeration (2026):

| Chain | `chain_id` |
|---|---|
| Ethereum | 1 |
| Polygon | 137 |
| Base | 8453 |
| Arbitrum One | 42161 |
| Linea | 59144 |
| BNB Chain | 56 |

Release-note additions on some mainnet nodes: Optimism (10), World Chain,
XRPL-EVM. Don't hard-code — let `get_config` drive your UI.

## Security token locking (v1 only)

`@yellow-org/sdk@1.x` exposes a distinct **security-token** track for
tokens held as collateral (e.g., margin, governance bonds):

- `client.lockSecurityTokens(token, amount)` — transfer into locked escrow
- `client.initiateSecurityTokensWithdrawal(token, amount)` — start the
  timelock
- `client.cancelSecurityTokensWithdrawal(...)` — abort before expiry
- `client.withdrawSecurityTokens(...)` — finalize after timelock elapses
- `client.approveSecurityToken(token, amount)` — ERC-20 allowance helper
- `client.getLockedBalance(token)` — read current locked amount

Locked balance is **separate** from unified balance and **cannot be
transferred or used in app sessions** until withdrawn. This is an
opt-in feature, not part of the base deposit/withdraw flow documented
above.

## Custody contract functions (IDeposit interface)

```solidity
function deposit(address token, uint256 amount) external payable;
function withdraw(address token, uint256 amount) external;
function balanceOf(address account, address token) external view returns (uint256);
```

Events to subscribe to for on-chain reconciliation:

```solidity
event Deposited(address indexed account, address indexed token, uint256 amount);
event Withdrawn(address indexed account, address indexed token, uint256 amount);
```

Exact ABI lives in the `@erc7824/nitrolite` package (`abi/` folder) and
the Yellow Network contract repo.

## Common gotchas

| Symptom | Cause |
|---|---|
| Balance doesn't show up after deposit | Didn't `approve` first (ERC-20) or the ClearNode hasn't observed the tx yet. Wait a few blocks; check `get_ledger_transactions` with `tx_type: "deposit"`. |
| `ERC20: insufficient allowance` | Approval was for a prior amount; re-approve with the new amount. |
| Can't withdraw — "insufficient balance" despite having funds | Funds are in a channel or app session, not unified balance. Close first. |
| Deposited wrong token | No auto-refund path; open a support ticket or withdraw through whichever asset symbol the ClearNode recognises for that token (check `get_assets`). |
| Approval griefing on upgradable tokens | Use the 0-first-then-N pattern: `approve(0)` then `approve(N)` on USDT and similar. |
| Native ETH deposit reverts | Passing `value` but `token` arg not the zero-address placeholder, or vice versa. |

## Related

- `yellow-state-channels` — channel lifecycle after you've funded
- `yellow-custody-contract` — full Solidity interface reference
- `yellow-queries` — `get_assets` (token list), `get_ledger_balances`, `get_ledger_transactions`
- `yellow-notifications` — `bu` tells you the moment a deposit credits
