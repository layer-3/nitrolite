# TypeScript SDK (`@yellow-org/sdk`)

Official TypeScript SDK for building web and mobile applications on Nitrolite state channels.

## Quick Reference

| Command | What it does |
|---------|-------------|
| `npm test` | Run unit tests (Jest) |
| `npm run build` | Run tests **then** compile (`npm run test && tsc`) |
| `npm run typecheck` | Type check only (no emit) |
| `npm run lint` | ESLint |
| `npm run test:integration` | Integration tests (separate Jest config) |
| `npm run test:all` | Unit + integration tests |

**Important:** `npm run build` runs the full test suite before compiling. If you just want to compile, run `npx tsc` directly.

## Package Details

- **Name:** `@yellow-org/sdk`
- **Version:** 1.2.0
- **Module:** ESM-only (`"type": "module"`)
- **Node:** >=20.0.0
- **Entry:** `dist/index.js` / `dist/index.d.ts`

## TypeScript Configuration

- Target: `es2020`
- Module: `ESNext`
- Module Resolution: `bundler`
- Strict mode: enabled
- Declaration files: generated

## Code Style

- Prettier: 120 char width, 4-space indent, single quotes, semicolons
- ESLint configured via `@typescript-eslint`

## Source Layout

| Path | Purpose |
|------|---------|
| `src/index.ts` | Barrel export — all public API goes here |
| `src/client.ts` | Main SDK client (~1100 lines) |
| `src/signers.ts` | Signer implementations (ECDSA, session key, channel) |
| `src/config.ts` | Client configuration with functional options |
| `src/core/` | State machine, types, events, utilities |
| `src/rpc/` | RPC client, message types, API methods |
| `src/blockchain/` | On-chain interactions (deposits, withdrawals) |
| `src/app/` | App session logic and packing |
| `src/utils.ts` | Shared utility functions |
| `test/unit/` | Unit tests |

## Test Setup

- Framework: Jest with ts-jest
- Config: `jest.config.cjs` (unit), `jest.integration.config.js` (integration)
- Path alias: `@/` maps to `src/` in test configs
- Pattern: data-driven tests with `.forEach()`, manual mocks (not jest.mock)
- Naming: `*.test.ts` (not `.spec.ts`)

## Key Dependencies

- `viem` — Ethereum interactions (NOT ethers.js in production code; ethers is dev-only)
- `decimal.js` — Precise decimal arithmetic for token amounts
- `zod` — Runtime validation
- `abitype` — ABI type utilities
