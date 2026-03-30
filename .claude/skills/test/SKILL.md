---
name: test
description: Run tests for a specific module or the whole project
allowed-tools: Bash, Read, Grep, Glob
---

Run tests for: $ARGUMENTS

Detect the target module and run the appropriate test command:

1. If $ARGUMENTS names a specific file or directory, use that to determine the module
2. If no arguments, detect from the current working directory or recently edited files

Route to the correct test runner:
- **sdk/ts/** -> `cd sdk/ts && npm test`
- **sdk/ts-compat/** -> `cd sdk/ts-compat && npm test`
- **sdk/go/** or Go files under sdk/go/ -> `go test ./sdk/go/... -v` (from repo root)
- **contracts/** or .sol files -> `cd contracts && forge test`
- **clearnode/** -> `go test ./clearnode/... -v` (from repo root)
- **test/integration/** -> `cd test/integration && npm test`
- **Repo root with no argument** -> run `go test ./...` (Go packages only — does NOT cover `sdk/ts`, `sdk/ts-compat`, `contracts`, or `test/integration`). Ask the user to specify a target for non-Go tests.

If a specific test file is given (e.g., `sdk/ts/test/unit/utils.test.ts`), run only that file:
- For Jest: `cd sdk/ts && npx jest test/unit/utils.test.ts`
- For Go: `go test -run TestName ./sdk/go/...`
- For Forge: `cd contracts && forge test --match-path test/MyTest.t.sol`

Report results: pass count, fail count, and any error details.
