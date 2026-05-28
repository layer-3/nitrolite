---
name: lint
description: Run linters and formatters for the appropriate language
allowed-tools: Bash, Read
---

Lint and check formatting for: $ARGUMENTS

Route to the correct linter based on context:

- **sdk/ts/** -> `cd sdk/ts && npm run lint && npx prettier --check .`
- **sdk/ts-compat/** -> `cd sdk/ts-compat && npm run typecheck` (no dedicated lint script; typecheck is the closest)
- **contracts/** -> `cd contracts && forge fmt --check`
- **Go code** -> `go vet ./...` from repo root (or scope to specific path like `./sdk/go/...`)

If no arguments provided, run all applicable linters and report a summary.

Report any issues found. Suggest fixes but do not auto-apply unless the user asks.
