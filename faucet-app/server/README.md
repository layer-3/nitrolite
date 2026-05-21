# Nitrolite Faucet Server

A Go-based faucet server that distributes tokens through the Nitronode network using WebSocket connections.

## Features

- **Nitrolite SDK Integration**: Uses the local `github.com/layer-3/nitrolite` SDK for Nitronode communication
- **Ethereum Wallet Integration**: Uses ECDSA private key for signing channel states and transactions
- **RESTful API**: Simple HTTP endpoints for token requests
- **Structured Logging**: JSON-formatted logs with configurable levels
- **Graceful Shutdown**: Proper cleanup of connections and resources
- **Address Validation**: Validates Ethereum addresses before processing requests

## Architecture

The application is structured into several packages:

- `internal/config`: Configuration management with environment variables
- `internal/logger`: Structured logging with logrus
- `internal/nitronode`: Thin wrapper around the Nitrolite SDK client
- `internal/server`: HTTP server with Gin framework

### Nitronode Client

The `internal/nitronode` package wraps the Nitrolite SDK's `sdk.Client`. Connection and message signing are handled internally by the SDK — no manual WebSocket management is required.

## Quick Start

1. **Setup**:

   ```bash
   cd faucet-app/server
   go mod tidy
   ```

2. **Configure environment**:

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run the server**:

   ```bash
   go run main.go
   ```

## Configuration

The application uses [cleanenv](https://github.com/ilyakaznacheev/cleanenv) for configuration management. Configuration can be provided via:

1. **`.env` file** in the current directory
2. **Environment variables** (used when `.env` is absent)

Set the following environment variables:

| Variable | Required | Default | Description | Example |
|----------|----------|---------|-------------|---------|
| `SERVER_PORT` | No | `8080` | HTTP server port | `8080` |
| `OWNER_PRIVATE_KEY` | **Yes** | - | Owner private key (without 0x prefix) — signs channel states and transfers | `abcdef123...` |
| `NITRONODE_URL` | **Yes** | - | Nitronode WebSocket URL | `wss://nitronode.example.com/ws` |
| `TOKEN_SYMBOL` | **Yes** | - | Token symbol to distribute | `usdc` |
| `STANDARD_TIP_AMOUNT` | **Yes** | - | Amount to send per request (decimal format) | `10.0` |
| `MIN_TRANSFER_COUNT` | **Yes** | - | Minimum number of transfers the server should have balance for | `5` |
| `COOLDOWN_PERIOD` | **Yes** | - | Cooldown between requests per wallet/IP (Go duration format) | `24h` |
| `TRUSTED_PROXIES` | No | `""` | Comma-separated trusted proxy IPs; empty means direct exposure only | `10.0.0.1,10.0.0.2` |
| `LOG_LEVEL` | No | `info` | Logging level (debug/info/warn/error) | `info` |

> **Note on `TRUSTED_PROXIES`:** If the faucet is deployed behind an ingress or load balancer, set this to the proxy IP(s). Without it, `c.ClientIP()` returns the proxy address and all requests share one IP rate-limit bucket.

## API Endpoints

### `POST /requestTokens`

Request tokens for an Ethereum address.

**Request body:**
```json
{ "userAddress": "0x..." }
```

**Success response (200):**
```json
{
  "success": true,
  "message": "Tokens sent successfully",
  "txId": "...",
  "amount": "10",
  "asset": "usdc",
  "destination": "0x..."
}
```

**Error responses:**
- `400` — Invalid address or request format
- `429` — Rate limit exceeded
- `500` — Transfer failed
- `503` — Nitronode unavailable or balance insufficient

### `GET /info`

Returns server metadata.

## WebSocket Connection Management

The Nitrolite SDK maintains a persistent WebSocket connection with Nitronode:

- **Connection**: Established on startup inside `nitronode.NewClient()`; no separate connect/auth step is needed
- **Authentication**: Handled internally by the SDK
- **Reconnection**: On each request, `EnsureConnected()` detects a lost connection (via `WaitCh()`) and reconnects with exponential backoff (3 attempts, 300 ms → 600 ms → 2 s)
- **Post-reconnect ping**: Each reconnect attempt is validated with a `Ping` before the new client is accepted
- **Message Handling**: Fully managed by the SDK's internal RPC layer

## Startup Log Example

```
{"level":"info","msg":"Starting Nitrolite Faucet Server","time":"..."}
{"level":"info","msg":"Configuration loaded: Server port=8080, Nitronode URL=wss://nitronode.example.com","time":"..."}
{"level":"debug","msg":"Token 'usdc' is supported by Nitronode","time":"..."}
{"level":"info","msg":"✓ Sufficient usdc balance: 50000000","time":"..."}
{"level":"info","msg":"Successfully connected to Nitronode","time":"..."}
{"level":"info","msg":"Faucet server is ready to serve requests","time":"..."}
```

## Security Features

- **Address Validation**: Validates Ethereum address format before processing
- **Private Key Security**: Private key is only used for signing, never exposed
- **CORS Support**: Configurable CORS headers for web integration
- **Request Signing**: All Nitronode requests are cryptographically signed by the SDK
- **Balance Guard**: Refuses to operate below minimum balance threshold
- **URL Redaction**: `NITRONODE_URL` is never logged in full — only scheme and host are shown

## Building for Production

```bash
go build -o faucet-server main.go
./faucet-server
```

## Development

```bash
# Run tests (from repo root)
go test ./faucet-app/...

# Run with hot reload
go install github.com/cosmtrek/air@latest
air
```

## Error Handling

- **Connection Errors**: Returns 503 if Nitronode is unavailable or reconnection fails
- **Validation Errors**: Returns 400 for invalid addresses or request format
- **Transfer Errors**: Returns 500 for Nitronode transfer failures
- **Service Unavailable**: Returns 503 if token is unsupported or balance is insufficient

## Troubleshooting

**Connection Issues:**

- Verify `NITRONODE_URL` is correct and accessible
- Check firewall settings for WebSocket connections

**Token Not Supported:**

- Verify `TOKEN_SYMBOL` is supported by the Nitronode instance

**Insufficient Balance:**

- Top up the faucet wallet; the server requires at least `MIN_TRANSFER_COUNT × STANDARD_TIP_AMOUNT` available balance
