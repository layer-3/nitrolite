# Example: Simple Order Book

> Demonstrates the multi-session architecture with CrossRefs — per-trader
> action sessions feeding into an operator-driven exchange session, enabling
> concurrent order submission with zero version contention.

## Concept

A simple order book where traders submit limit orders to their own action
sessions. An exchange operator matches orders and settles trades in a central
exchange session, referencing trader actions via CrossRefs.

## Sessions

### Trader Action Session (one per trader)

```
Metadata: { trader: address, exchange_session: bytes32 }
Module:   action_log (single-signer access control)
```

Data per version:
```
{
    type: "limit_buy" | "limit_sell" | "cancel",
    pair: "ETH/USDC",
    price: decimal,
    amount: decimal,
    order_id: string        // client-generated, unique per trader
}
```

Traders write orders to their own session. No version contention — each trader
is the sole writer. Orders are signed and immutable once submitted.

### Exchange Session (operator-driven)

```
Metadata: {
    operator: address,
    pair: "ETH/USDC",
    base_asset: "ETH",
    quote_asset: "USDC",
    trader_sessions: [bytes32]      // registered trader action session IDs
}
Module: orderbook_v1
```

Data per version:
```
{
    matches: [{
        buy_order:  { trader: address, order_id: string, price, amount },
        sell_order: { trader: address, order_id: string, price, amount },
        fill_price: decimal,
        fill_amount: decimal
    }]
}
```

## ModuleState (Exchange Session)

```
{
    order_book: {
        bids: [{ trader, order_id, price, amount_remaining }],
        asks: [{ trader, order_id, price, amount_remaining }]
    },
    action_cursors: {
        <trader_session_id>: <last_processed_version>,
        ...
    },
    balances: {
        <trader_address>: { ETH: decimal, USDC: decimal }
    }
}
```

## Flow

```
Trader A action session:
  v1: { type: "limit_buy", pair: "ETH/USDC", price: 3000, amount: 1.0 }

Trader B action session:
  v1: { type: "limit_sell", pair: "ETH/USDC", price: 2999, amount: 0.5 }

Exchange session:
  v5: CrossRefs: [{TraderA_session, v1}, {TraderB_session, v1}]
      Data: { matches: [{ buy: A's order, sell: B's order,
                          fill_price: 2999, fill_amount: 0.5 }] }
      Signers: [operator]
```

## Module Validation (Exchange Session)

On each update, the module:

1. **Ingests new orders** — iterates ResolvedCrossRefs, extracts order data from
   each referenced trader action session, verifies each CrossRef version >
   corresponding action cursor. Adds new orders to the order book in ModuleState.

2. **Validates matches** — verifies each match in Data is legitimate: buy price
   >= sell price, fill amount <= both orders' remaining amount, fill price is
   within the crossed spread.

3. **Updates balances** — adjusts trader balances in ModuleState based on fills.
   Buyer gets base asset, seller gets quote asset.

4. **Advances cursors** — updates action_cursors to reflect processed versions.

5. **Emits events:**
   ```
   Topic 0x0001  OrderPlaced    { trader, side, price, amount }
   Topic 0x0002  OrderFilled    { trader, side, fill_price, fill_amount }
   Topic 0x0003  OrderCancelled { trader, order_id }
   ```

## Fund Flow

Traders deposit into the exchange session via vouchers before trading:

```
Trader A's channel → issues voucher → Exchange session uses voucher (VouchersUsed)
```

On withdrawal, the module issues vouchers back:

```
Trader submits withdraw action → operator includes in exchange update →
module issues voucher from exchange session to trader's channel (VouchersIssued)
```

Trader balances in ModuleState track what each trader owns within the exchange.
LockedFunds (node-level) tracks total assets held by the exchange session.

## Why This Requires CrossRefs

**Without CrossRefs:** all traders submit to the exchange session directly.
With 100 active traders, version contention makes the system unusable — almost
every order submission fails and must be retried.

**With CrossRefs:** each trader writes to their own session (zero contention).
The operator reads all action sessions, batches orders, matches them, and
submits one exchange update referencing multiple trader actions. The node
verifies each CrossRef, and the module gets the actual order data directly
from ResolvedCrossRefs.

**Result:** trader throughput scales with the number of action sessions, not
with the exchange session's version rate. The operator can batch dozens of
orders into a single exchange update.
