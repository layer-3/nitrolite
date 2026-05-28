---
paths:
  - "**/*.go"
---

- Follow standard Go conventions: `gofmt`, exported names have doc comments.
- Error handling: always check and return errors, never ignore with `_`.
- Use the functional options pattern for configuration (see `sdk/go/config.go`).
- Shared packages live in `pkg/` — check there before creating new utilities.
- Test files go in the same package with `_test.go` suffix.
- Run tests from repo root: `go test ./sdk/go/...` (scoped) or `go test ./...` (all).
