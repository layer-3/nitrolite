# Nitrolite Copilot Instructions

Nitrolite is a state channel protocol for Ethereum/EVM blockchains — off-chain instant transactions with on-chain security. Repo contains: Solidity contracts (Foundry), Go clearnode broker, TypeScript + Go SDKs.

## Architecture

- Channels: state containers between User and Node, hold asset allocations
- States: versioned allocations, each version = previous + 1, mutually signed = enforceable on-chain
- Clearnode: off-chain broker, WebSocket JSON-RPC wire format
- App Sessions: multi-party extensions with quorum-based governance

## TypeScript SDK (`@yellow-org/sdk`)

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

## Key SDK Methods (both TS and Go)

Deposit, Transfer, Checkpoint, CloseHomeChannel, CreateAppSession, SubmitAppState, CloseAppSession, GetChannels, GetBalances, GetConfig

## Commits

`feat|fix|chore|test|docs(scope): description`

## Avoid

- ethers.js (use viem), ignoring Go errors, barrel re-exporting classes in sdk-compat
- Running `npm test && npm run build` in sdk/ts (build already runs tests)
- Editing .env files, committing secrets
