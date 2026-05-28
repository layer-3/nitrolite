---
name: typecheck
description: Run TypeScript type checking across SDK packages
allowed-tools: Bash, Read
---

Run typecheck for: $ARGUMENTS

1. If $ARGUMENTS specifies a package ("ts", "ts-compat", or "both"), check that package
2. If no arguments, check both TypeScript packages:

```bash
cd sdk/ts && npm run typecheck
cd sdk/ts-compat && npm run typecheck
```

3. Report any type errors with file paths and line numbers
4. If the user asks about Go type checking, suggest `go vet ./...` instead
