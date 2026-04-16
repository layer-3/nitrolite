---
name: build-sdk
description: Build SDK packages in the correct dependency order
allowed-tools: Bash, Read
---

Build SDK packages for: $ARGUMENTS

**Important:** sdk/ts-compat depends on sdk/ts via `"@yellow-org/sdk": "file:../ts"`. Always respect build order.

1. If $ARGUMENTS is "all" or empty, build everything in order:
   - `cd sdk/ts && npm run build` (this runs tests then tsc)
   - `cd sdk/ts-compat && npm run build`
   - `go build ./sdk/go/...`

2. If $ARGUMENTS specifies a single package, build just that:
   - "ts" -> `cd sdk/ts && npm run build`
   - "ts-compat" or "compat" -> check sdk/ts/dist/ exists first; if not, build sdk/ts first, then `cd sdk/ts-compat && npm run build`
   - "go" -> `go build ./sdk/go/...`

3. Report build success/failure for each package with any error details.
