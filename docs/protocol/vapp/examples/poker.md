# Example: Poker Table

> Demonstrates vouchers (multi-player buy-in/payout), events (hand history),
> and ModuleState (deck, dealer rotation, pot tracking) — capabilities that
> are impossible with V1.

## Concept

A multi-player poker table where players buy in with real assets, play hands
with validated dealing and betting, and cash out at any time. The module
enforces game rules, tracks the deck and pot, and emits a full hand history
as events.

## Definition

```
Metadata: {
    table_config: { max_players: 6, small_blind: 10, big_blind: 20, asset: "USDC" },
    operator: address
}
Module: poker_v1
```

## Data Schema

```
{
    action: "join" | "leave" | "bet" | "fold" | "check" | "call" | "deal",
    player: address,
    amount: decimal       // for bet actions
}
```

Data is intentionally minimal — it represents the current action only. All
accumulated state (player stacks, community cards, pot) lives in ModuleState.

## ModuleState

```
{
    players: [{ address, stack: decimal, status: "active" | "folded" | "allin" }],
    deck_seed: bytes32,         // deterministic shuffle from previous hands
    community_cards: [],
    pot: decimal,
    dealer_position: uint,
    betting_round: "preflop" | "flop" | "turn" | "river" | "showdown",
    current_turn: uint
}
```

This is the key V2 advantage: ModuleState holds the full game state without
participants needing to sign it. Players sign only their action. The module
computes everything else deterministically.

## Events

```
Topic 0x0001  HandStarted   { hand_number, dealer, players, blinds_posted }
Topic 0x0002  CardDealt      { player, card }  (encrypted or committed)
Topic 0x0003  BetPlaced      { player, amount, pot_total }
Topic 0x0004  CommunityCard  { card, round }
Topic 0x0005  HandResult     { winner, amount, hand_rank }
Topic 0x0006  PlayerJoined   { player, buy_in }
Topic 0x0007  PlayerLeft     { player, cashout }
```

Events create a complete, queryable hand history. Frontends subscribe to these
for real-time UI updates. Analytics services aggregate them for player stats.

## Lifecycle Sketch

```
v1  operator creates table            Data: { action: "deal" }
    Signers: [operator]
    ModuleState initialized: empty table, fresh deck seed

v2  player A joins                    Data: { action: "join", player: A }
    VouchersUsed: [100 USDC from A's channel]
    Signers: [operator]
    → Module adds A to players with stack=100, emits PlayerJoined

v3  player B joins                    Data: { action: "join", player: B }
    VouchersUsed: [100 USDC from B's channel]
    Signers: [operator]

v4  operator deals hand               Data: { action: "deal" }
    Signers: [operator]
    → Module shuffles deck (deterministic from seed), assigns hole cards,
      posts blinds from stacks, emits HandStarted + CardDealt events

v5  player A bets                     Data: { action: "bet", player: A, amount: 20 }
    Signers: [operator]
    → Module validates it's A's turn, amount is valid, updates pot

...

vN  hand resolves at showdown
    → Module determines winner, updates stacks in ModuleState,
      emits HandResult event

vM  player B cashes out               Data: { action: "leave", player: B }
    Signers: [operator]
    → Module issues voucher for B's stack to B's channel,
      removes B from table, emits PlayerLeft
```

## Why This Requires V2

**V1 cannot do this:**
- Allocation model requires tracking per-player balances in signed state.
  Every bet would need all players to sign an allocation change.
- No ModuleState means deck state, pot, and betting rounds would need to
  be in the signed Data — players would sign the full game state every turn.
- No events means no structured hand history.
- Fixed quorum means the participant list can't change (no join/leave).
- Single-deposit model means players can't buy in independently.

**V2 makes it natural:**
- Players sign only their action (tiny Data, one signer).
- Module manages all game state in ModuleState (never signed, deterministic).
- Vouchers handle buy-in and cashout independently.
- Events produce a full hand history for free.
- Operator pattern (single session key signer) eliminates version contention
  even with 6 players acting in sequence.
