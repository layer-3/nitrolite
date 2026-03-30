---
paths:
  - "sdk/**/*.ts"
  - "sdk/**/*.tsx"
---

- Use `const` by default, `let` only when reassignment is needed. No `var`.
- Prefer `viem` over `ethers.js` for Ethereum interactions in production code.
- All public API functions must be exported through the barrel `index.ts`.
- Use strict TypeScript — no `any` unless absolutely unavoidable (e.g., RPC wire types).
- Async functions preferred over raw Promise chains.
- Test files use `.test.ts` extension (not `.spec.ts`).
- When modifying sdk-compat: never barrel re-export SDK classes (SSR risk). Only `export type` is safe.
