# TypeScript SDK Compat (`@yellow-org/sdk-compat`)

Curated migration layer that bridges the old `@layer-3/nitrolite` v0.5.3 API to the `@yellow-org/sdk` v1 runtime. Helps existing dApps migrate supported app-facing paths.

## Quick Reference

| Command | What it does |
|---------|-------------|
| `npm test` | Run unit tests (Jest) |
| `npm run build` | Compile with tsc |
| `npm run typecheck` | Type check only |

## Package Details

- **Name:** `@yellow-org/sdk-compat`
- **Version:** 1.2.0
- **Peer deps:** `@yellow-org/sdk >=1.2.0`, `viem ^2.0.0`
- **Dev dep:** `"@yellow-org/sdk": "file:../ts"` — **must build `sdk/ts` first**

## Critical Constraint: No Barrel Re-Export of SDK Classes

The main SDK (`@yellow-org/sdk`) has side effects on module evaluation that break SSR.

**DO NOT** add `export { Client } from '@yellow-org/sdk'` to `index.ts`.

**SAFE:** `export type { StateSigner } from '@yellow-org/sdk'` (type-only exports are erased at compile time).

This is documented in `src/index.ts`.

## Source Layout

| Path | Purpose |
|------|---------|
| `src/index.ts` | Barrel export — update when adding public API |
| `src/client.ts` | NitroliteClient facade (~1100 lines) |
| `src/auth.ts` | Auth helpers (createAuthRequestMessage, etc.) |
| `src/rpc.ts` | RPC compat stubs (create*Message / parse*Response) |
| `src/types.ts` | All types, enums, interfaces |
| `src/signers.ts` | WalletStateSigner, createECDSAMessageSigner |
| `src/app-signing.ts` | App session hash packing |
| `src/errors.ts` | CompatError class hierarchy |
| `src/events.ts` | EventPoller (polling bridge for v0.5.3 push events) |
| `src/config.ts` | Configuration builders |
| `docs/` | Migration guides (overview, on-chain, off-chain) |
| `test/unit/` | Unit tests |

## Test Setup

- Framework: Jest with ts-jest
- Config: `jest.config.cjs`
- ESM handling: `transformIgnorePatterns` whitelists `@yellow-org` for ts-jest
- Pattern: manual mock signers (`async () => '0xsig'`)

## When Adding Exports

1. Add the function/type to the appropriate source file
2. Export it from `src/index.ts`
3. If re-exporting from `@yellow-org/sdk`, use `export type` only (SSR safety)
4. Update `README.md` (Types Reference, RPC Stubs, or Auth Helpers section)
5. Add or update a focused unit test under `test/unit/` that matches the surface you changed
