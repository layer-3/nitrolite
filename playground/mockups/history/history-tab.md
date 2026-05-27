# History Tab — Design & Data

Mockup file: `c-expandable-rows.html`

---

## Layout

The History tab is a **full-width view** that replaces the two-column (ActionPanel + ChannelList) layout. A tab bar is added at the top of the main content area to switch between **Overview** and **History**. The same WalletBar and outer container are shared with the rest of the app.

---

## Table

### Columns

| Column | Content |
|--------|---------|
| *(expand)* | Chevron to expand row detail |
| Time | Relative timestamp (e.g. "2h ago"); absolute UTC on hover |
| Type | Colored badge — see types below |
| Asset | Token icon + symbol |
| Amount | Mono font; `+` green for deposits, `−` red for withdrawals, neutral for transfers/finalize |
| From | Shortened address; copy icon on hover; clickable for quick filter |
| To | Same as From |
| Chain | Compact chain-name pill; clickable for quick filter; `—` for off-chain txs |
| Tx Hash | Shortened hash with block-explorer link (↗); `—` for off-chain txs |

### Transaction type badges

| Badge | Color |
|-------|-------|
| Deposit | Green |
| Withdraw | Red |
| Transfer | Blue |
| Finalize | Purple |
| Commit / Release / Rebalance | Muted gray |

### Sort order

Newest first (`createdAt` descending). 25 rows per page. Pagination footer shows "Showing 1–25 of N transactions" with Prev / Next buttons.

### Confirming row (session enrichment)

Rows deposited or withdrawn in the current session may carry a local **Confirming** badge (accent spinner) while `waitForTransactionReceipt` is pending. This state is ephemeral and lost on reload — see Session Enrichment section below.

---

## Filtering

Two complementary mechanisms, both updating the same filter state.

### Column header popovers

Clicking the funnel icon (▾) on a filterable column opens a popover anchored below that header:

| Column | Popover content |
|--------|----------------|
| Type | Checkbox list with colored badges (multi-select) |
| Asset | Checkbox list of available asset symbols |
| Chain | Checkbox list of available chain names |
| From | Free-text address input |
| To | Free-text address input |

Each popover has **Clear** and **Apply** buttons. The column label + icon turn accent-colored when a filter is active on that column.

### Quick filter (per-cell click)

Clicking directly on a cell value — a Type badge, Asset name, shortened address, or Chain pill — immediately applies an exact-value filter for that field. If that filter value is already active, the same click removes it.

Visual feedback:
- Cursor pointer on hover over any filterable cell value
- Subtle highlight background on hover
- Accent-dim background when the filter is active (`.qf-on`)
- Tooltip on hover: **"Quick filter: X"** or **"Remove filter: X"**

Copy buttons on From / To addresses stop click propagation so they never trigger quick filter.

---

## Expanded Row Detail

Clicking anywhere on a row (except interactive elements) expands it in-place. A detail panel renders below the row with a 3-column grid:

| Section | Content |
|---------|---------|
| Sender new state ID | Full state UUID — copy button |
| Receiver new state ID | Full state UUID — copy button |
| Timestamp | Absolute UTC timestamp |
| Confirmation *(spans all 3 cols)* | Step timeline — see below |

### Confirmation timelines

**Off-chain (Transfer):**
`Signed ──● Co-signed`

**On-chain (Deposit / Withdraw / Finalize):**
`Signed ──● Broadcasted ──● Confirmed`

Completed steps shown in green; pending steps in border color.

---

## Data Fetching

### Primary source — `client.getTransactions()`

**Location:** `sdk/ts/src/client.ts`

```typescript
async getTransactions(
  wallet: Address,
  options?: {
    asset?:    string;
    txType?:   TransactionType;
    fromTime?: bigint;   // Unix timestamp, seconds
    toTime?:   bigint;
    page?:     number;
    pageSize?: number;
  }
): Promise<{
  transactions: Transaction[];
  metadata:     PaginationMetadata;
}>
```

Underlying RPC method: `user.v1.get_transactions`

### `Transaction` type

```typescript
interface Transaction {
  id:                  string;
  asset:               string;
  txType:              TransactionType;
  fromAccount:         Address;
  toAccount:           Address;
  senderNewStateId?:   string;
  receiverNewStateId?: string;
  amount:              Decimal;
  createdAt:           Date;
}
```

### `TransactionType` enum

```typescript
enum TransactionType {
  HomeDeposit    = 10,
  HomeWithdrawal = 11,
  EscrowDeposit  = 20,
  EscrowWithdraw = 21,
  Transfer       = 30,
  Commit         = 40,
  Release        = 41,
  Rebalance      = 42,
  Migrate        = 100,
  EscrowLock     = 110,
  MutualLock     = 120,
  Finalize       = 200,
}
```

`Migrate`, `EscrowLock`, and `MutualLock` are not exposed through the API's `tx_type` filter.

### `PaginationMetadata` type

```typescript
interface PaginationMetadata {
  page:       number;  // 1-indexed
  perPage:    number;
  totalCount: number;
  pageCount:  number;
}
```

### Which filters are server-side vs. client-side

| UI filter | Handling | SDK option |
|-----------|----------|-----------|
| Type | **Server-side** | `txType` — single value per call; multi-select needs client-side post-filter |
| Asset | **Server-side** | `asset` |
| Date range | **Server-side** | `fromTime` / `toTime` |
| From address | **Client-side** | No API param; filter on `tx.fromAccount` after fetch |
| To address | **Client-side** | No API param; filter on `tx.toAccount` after fetch |
| Chain | **Client-side** | No API param; chain inferred from channel records |

For From / To / Chain filtering, either increase `pageSize` to fetch a larger window and filter locally, or fetch all pages (feasible for typical wallet history sizes).

### Session enrichment

The server does not store blockchain tx hashes — those are returned by `client.checkpoint()` and confirmed via `waitForTransactionReceipt`. During the current session, `useChannelOps` can maintain an enrichment map:

```typescript
// Key: senderNewStateId  Value: { txHash, status }
type EnrichmentMap = Map<string, { txHash: string; status: 'confirming' | 'confirmed' | 'failed' }>;

// After client.checkpoint(asset) resolves:
enrichment.set(senderNewStateId, { txHash, status: 'confirming' });

// After waitForTransactionReceipt resolves:
enrichment.set(senderNewStateId, { txHash, status: 'confirmed' });
```

When rendering the history table, look up each row's `senderNewStateId` in the map to overlay tx hash and confirmation status. This enrichment is lost on page reload, but the tx hash itself can be used to re-query on-chain status if needed.

### Suggested hook shape

```
useHistory(client, address, enrichmentMap)
  state:   { transactions, metadata, filters, page, isLoading }
  fetch:   client.getTransactions(address, { asset, txType, fromTime, toTime, page, pageSize: 25 })
  refetch: on address change · filter change · page change · onAfterOp callback
  render:  merge enrichmentMap by senderNewStateId before passing rows to the table
```
