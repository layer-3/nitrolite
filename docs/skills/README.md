# Yellow Network builder skills

A reference library of [Claude Code skills](https://claude.com/blog/claude-code-skills) for builders integrating with Yellow Network. Each skill is a focused markdown document with frontmatter that AI coding agents (Claude Code, Cursor, etc.) can load on demand.

## When to use these

Reach for these skills when you're building **on top of Yellow Network** — opening state channels, routing transfers over Nitro RPC, signing the ClearNode auth handshake, deciding between session keys and main wallet, debugging a `8011 stale` error, picking between the v0.5.x and v1.x SDKs, etc. They cover the protocol surface, not any specific application built on it.

## Catalog

### Getting started

- [`yellow-network-builder`](./yellow-network-builder/SKILL.md) — start here. What Yellow solves, three-layer architecture (on-chain Custody + off-chain Nitro RPC + P2P YNP), key concepts, and a pointer index to every other skill.

### Off-chain protocol (Nitro RPC)

- [`yellow-nitro-rpc`](./yellow-nitro-rpc/SKILL.md) — wire format reference. The `[req, sig]` envelope, request/response/notification shapes, signing rules, request-id correlation.
- [`yellow-clearnode-auth`](./yellow-clearnode-auth/SKILL.md) — the 3-step `auth_request` → `auth_challenge` → `auth_verify` handshake with EIP-712 Policy.
- [`yellow-session-keys`](./yellow-session-keys/SKILL.md) — hot-wallet delegation: keypair generation, allowances, scopes, expiry, rotation, revocation.
- [`yellow-queries`](./yellow-queries/SKILL.md) — every read-only Nitro RPC method on ClearNode (balances, channels, app sessions, history, ping).
- [`yellow-notifications`](./yellow-notifications/SKILL.md) — the five server-to-client pushes (`assets`, `bu`, `cu`, `tr`, `asu`) and how to route them client-side.
- [`yellow-errors`](./yellow-errors/SKILL.md) — Nitro RPC has descriptive string errors only (no numeric codes). Canonical list and recovery playbooks.

### Money movement

- [`yellow-deposits-withdrawals`](./yellow-deposits-withdrawals/SKILL.md) — on-chain funding: ERC-20 `approve` + `deposit`, native ETH, `withdraw` via Custody, and how channels open zero-balance in v0.5+.
- [`yellow-transfers`](./yellow-transfers/SKILL.md) — off-chain peer-to-peer transfers, allowances, multi-asset, idempotency, delivery receipts.
- [`yellow-state-channels`](./yellow-state-channels/SKILL.md) — channel lifecycle (open, operate, close), cooperative + unilateral close paths, challenge-response.
- [`yellow-app-sessions`](./yellow-app-sessions/SKILL.md) — multi-party off-chain accounts: 2-of-3 escrow, stake-with-slashing, turn-based games. Definition, allocations, quorum.
- [`yellow-custody-contract`](./yellow-custody-contract/SKILL.md) — Solidity reference for the on-chain anchor: `IChannel`, `IDeposit`, `IChannelReader`, EIP-712 `VirtualApp:Custody` domain, structs, events.
- [`yellow-swap-design`](./yellow-swap-design/SKILL.md) — three patterns for off-chain asset exchange (treasury-convert, market-maker agent, P2P matching) when the protocol does not provide a canonical swap primitive.

### SDKs

- [`yellow-sdk-v1`](./yellow-sdk-v1/SKILL.md) — `@yellow-org/sdk` v1.x (the active upstream SDK). Single unified `Client` class, `EthereumMsgSigner` / `EthereumRawSigner`, namespaced methods.
- [`yellow-sdk-api`](./yellow-sdk-api/SKILL.md) — `@erc7824/nitrolite` v0.5.x (frozen). Full surface of builders, parsers, signers, `NitroliteClient`. Useful for audits or migrations from older code.

## Frontmatter conventions

Each `SKILL.md` carries:

- `name` — the skill identifier (matches the directory).
- `description` — one paragraph explaining when to use the skill. AI agents match on this.
- `version` — semantic version of the skill itself.
- `sdk_version` — which SDK release the examples are pinned to.
- `network` — `mainnet` (most), `sandbox`, or `local`.
- `last_verified` — date the skill was last checked against the live SDK / docs.

## Provenance

These skills started as the public reference set used by the YellowHive agent platform (registration, marketplace, swap, social network for AI agents on Yellow Network). They are protocol-generic — nothing in this set assumes the YellowHive runtime — and are contributed back here so any builder on Yellow Network can use them.

Contributions welcome: open a PR with a new skill, an edit, or a bumped `last_verified`. Aim for skills that teach the protocol, not your specific stack.
