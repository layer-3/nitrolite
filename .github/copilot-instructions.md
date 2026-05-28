# Nitrolite Copilot Instructions

Nitrolite is a state channel protocol for Ethereum/EVM blockchains — off-chain instant transactions with on-chain security. Repo contains: Solidity contracts (Foundry), Go nitronode broker (renamed from clearnode), TypeScript + Go SDKs.

## Architecture

- Channels: state containers between User and Node, hold asset allocations
- States: versioned allocations, each version = previous + 1, mutually signed = enforceable on-chain
- Nitronode (formerly Clearnode): off-chain broker, WebSocket JSON-RPC wire format
- App Sessions: multi-party extensions with quorum-based governance

## TypeScript SDK (`@yellow-org/sdk`)

- `@yellow-org/sdk` = v1 protocol SDK. Use for all new code.
- `@yellow-org/sdk-compat` = bridges 0.5.x API to v1 runtime. Migration only. Wraps Client with NitroliteClient.
- V1 API source of truth: `docs/api.yaml`. `docs/legacy/API.md` has 0.5.x method names.
- Use `const` by default, `viem` over `ethers.js`, strict TypeScript (no `any`)
- All public API through barrel `index.ts`, async/await preferred
- Tests: `.test.ts`, Jest. Run: `cd sdk/ts && npm test`
- Build order: `sdk/ts` before `sdk/ts-compat`
- sdk-compat: NEVER barrel re-export classes (SSR risk), only `export type`

## Go SDK (`github.com/layer-3/nitrolite/sdk/go`)

- Root go.mod, Go 1.25. Standard Go: `gofmt`, doc comments on exports
- Always check errors, never ignore with `_`
- Functional options pattern (`sdk/go/config.go`), shared utils in `pkg/`
- Tests: `_test.go`, run from repo root: `go test ./sdk/go/...`
- Use `context.Context`, `github.com/shopspring/decimal` for amounts

## Solidity (`contracts/`)

- Foundry: `forge build`, `forge test`, `forge fmt`
- NatSpec on public/external functions, security-first, OpenZeppelin
- Tests: `.t.sol` in `contracts/test/`
- Style: https://github.com/layer-3/clearsync/blob/master/contracts/solidity-style-guide.md
- Practices: https://github.com/layer-3/clearsync/blob/master/contracts/solidity-development-practices.md

## Key SDK Methods (both TS and Go)

Deposit, Transfer, Checkpoint (required after deposit/withdraw/close), CloseHomeChannel (follow with Checkpoint), CreateAppSession (definition, sessionData, quorumSigs), SubmitAppState (appStateUpdate, quorumSigs — use Close intent to close sessions), GetChannels(wallet), GetBalances(wallet), GetConfig

> No CloseAppSession() on SDK Client. Close via SubmitAppState with Close intent.

## Key Reference Files

- Protocol description: `protocol-description.md`
- Smart contract invariants: `contracts/SECURITY.md`
- Main SC entrypoint: `contracts/src/ChannelHub.sol` (design: `contracts/suggested-contract-design.md`)
- Nitronode logic: `nitronode/README.md`, `docs/legacy/`
- Advanced protocol docs: `docs/protocol/`

## Commits

`feat|fix|chore|test|docs(scope): description`

## Avoid

- ethers.js (use viem), ignoring Go errors, barrel re-exporting classes in sdk-compat
- Running `npm test && npm run build` in sdk/ts (build already runs tests)
- Editing .env files, committing secrets
