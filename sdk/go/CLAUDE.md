# Go SDK

Official Go SDK for building backend integrations and CLI tools on Nitrolite state channels.

## Quick Reference

```bash
# All commands run from the REPO ROOT (not sdk/go/)
go test ./sdk/go/... -v      # Test SDK only
go build ./sdk/go/...         # Build SDK only
go vet ./sdk/go/...           # Lint SDK only

go test ./...                 # Test EVERYTHING (nitronode, pkg, sdk, cerebro)
```

**Important:** This is NOT a separate Go module. It shares the root `go.mod` (`github.com/layer-3/nitrolite`, Go 1.25).

## Source Layout

| File | Purpose |
|------|---------|
| `client.go` | Main SDK client — entry point for all operations |
| `config.go` | Functional options pattern for client configuration |
| `channel.go` | Channel operations (open, close, resize, challenge) |
| `utils.go` | Encoding/packing helpers |
| `doc.go` | Package-level documentation with usage examples |
| `*_test.go` | Tests (same directory, standard Go convention) |
| `mock_dialer_test.go` | Test doubles for WebSocket dialer |

## Architecture

- Uses **functional options pattern** for configuration (see `config.go`)
- Key dependency: `github.com/ethereum/go-ethereum` for Ethereum interactions
- Shared packages in `pkg/` (sign, core, rpc, app, blockchain, log) are used by both this SDK and `nitronode/`
- WebSocket-based communication with nitronode via JSON-RPC

## Conventions

- Standard Go: `gofmt`, exported names have doc comments
- Error handling: always check and return errors, never ignore
- Tests use standard `testing` package with `*_test.go` in same directory
- Check `pkg/` for shared utilities before creating new ones
