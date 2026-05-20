# Nitrolite Faucet Server

A Go-based faucet server that distributes tokens through the Clearnode network using WebSocket connections.

## Features

- **Nitrolite SDK Integration**: Uses the official `github.com/layer-3/nitrolite` SDK for Clearnode communication
- **Ethereum Wallet Integration**: Uses ECDSA private key for signing channel states and transactions
- **RESTful API**: Simple HTTP endpoints for token requests
- **Structured Logging**: JSON-formatted logs with configurable levels
- **Graceful Shutdown**: Proper cleanup of connections and resources
- **Address Validation**: Validates Ethereum addresses before processing requests

## Architecture

The application is structured into several packages:

- `internal/config`: Configuration management with environment variables
- `internal/logger`: Structured logging with logrus
- `internal/clearnode`: Thin wrapper around the Nitrolite SDK client
- `internal/server`: HTTP server with Gin framework

### Clearnode Client

The `internal/clearnode` package wraps the Nitrolite SDK's `sdk.Client`. Connection and message signing are handled internally by the SDK — no manual WebSocket management is required.

## Quick Start

1. **Clone and setup**:

   ```bash
   cd server
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

1. **Environment variables** (highest priority)
2. **`.env` file** in the current directory
3. **Default values** for optional settings

Set the following environment variables (or create a `.env` file):

| Variable | Required | Default | Description | Example |
|----------|----------|---------|-------------|---------|
| `SERVER_PORT` | No | `8080` | HTTP server port | `8080` |
| `OWNER_PRIVATE_KEY` | **Yes** | - | Owner private key (without 0x prefix) — signs channel states and transfers | `abcdef123...` |
| `CLEARNODE_URL` | **Yes** | - | Clearnode WebSocket URL | `wss://testnet.clearnode.io/ws` |
| `TOKEN_SYMBOL` | **Yes** | - | Token symbol to distribute | `usdc` |
| `STANDARD_TIP_AMOUNT` | **Yes** | - | Amount to send per request (decimal format) | `10.0` |
| `MIN_TRANSFER_COUNT` | **Yes** | - | Minimum number of transfers the server should have balance for | `5` |
| `COOLDOWN_PERIOD` | **Yes** | - | Cooldown between requests per wallet/IP (Go duration format) | `24h` |
| `LOG_LEVEL` | No | `info` | Logging level (debug/info/warn/error) | `info` |

## API Endpoints

### POST /requestTokens

Request tokens from the faucet.

**Request Body:**

```json
{
  "userAddress": "0x1234567890abcdef1234567890abcdef12345678"
}
```

**Success Response:**

```json
{
  "success": true,
  "message": "Tokens sent successfully",
  "txId": "abc123",
  "amount": "10",
  "asset": "usdc",
  "destination": "0x1234567890abcdef1234567890abcdef12345678"
}
```

**Error Response:**

```json
{
  "error": "Invalid address format."
}
```

### GET /info

Service information endpoint.

**Response:**

```json
{
  "service": "Nitrolite Faucet Server",
  "version": "1.0.0",
  "faucet_address": "0xabcd...",
  "standard_tip_amount": "10",
  "token_symbol": "usdc",
  "endpoints": ["/requestTokens"]
}
```

## Startup Validation

The server performs validation during startup:

### Token Support Validation

- Queries Clearnode using `GetAssets()` to fetch all supported tokens
- Validates that the configured `TOKEN_SYMBOL` exists
- Server refuses to start if the token is not supported

### Balance Verification

- Queries the owner balance using `GetBalances()`
- Requires balance ≥ `STANDARD_TIP_AMOUNT × MIN_TRANSFER_COUNT`
- Server refuses to start with insufficient funds

Example startup output:

```text
INFO Starting Nitrolite Faucet Server
INFO Faucet owner address: 0xabc...
INFO Successfully connected to Clearnode
INFO Token 'usdc' is supported by Clearnode
INFO ✓ Sufficient usdc balance: 50000000
INFO Faucet server is ready to serve requests
```

## WebSocket Connection Management

The Nitrolite SDK maintains a persistent WebSocket connection with Clearnode:

- **Connection**: Established on startup inside `clearnode.NewClient()`; no separate connect/auth step is needed
- **Authentication**: Handled internally by the SDK
- **Reconnection**: On each request, `EnsureConnected()` detects a lost connection (via `WaitCh()`) and recreates the SDK client automatically
- **Message Handling**: Fully managed by the SDK's internal RPC layer

## Security Features

- **Address Validation**: Validates Ethereum address format before processing
- **Private Key Security**: Private key is only used for signing, never exposed
- **CORS Support**: Configurable CORS headers for web integration
- **Request Signing**: All Clearnode requests are cryptographically signed by the SDK
- **Balance Guard**: Refuses to operate below minimum balance threshold

## Building for Production

```bash
# Build binary
go build -o faucet-server main.go

# Run with environment file
./faucet-server
```

## Docker Support

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o faucet-server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/faucet-server .
CMD ["./faucet-server"]
```

## Development

```bash
# Install dependencies
go mod tidy

# Run with hot reload (using air)
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...
```

## Logging

The application uses structured JSON logging:

```json
{
  "level": "info",
  "msg": "Processing faucet request for address: 0x1234...",
  "time": "2023-12-01T10:00:00Z"
}
```

Log levels: `debug`, `info`, `warn`, `error`, `fatal`

## Error Handling

- **Connection Errors**: Returns 503 if Clearnode is unavailable or reconnection fails
- **Validation Errors**: Returns 400 for invalid addresses or request format
- **Transfer Errors**: Returns 500 for Clearnode transfer failures
- **Service Unavailable**: Returns 503 if token is unsupported or balance is insufficient

## Monitoring

Key metrics to monitor:

- Transfer success/failure rates
- Response times
- Server resource usage

## Troubleshooting

**Connection Issues:**

- Verify `CLEARNODE_URL` is correct and accessible
- Check firewall settings for WebSocket connections

**Authentication Issues:**

- Verify `OWNER_PRIVATE_KEY` format (no `0x` prefix)
- Review logs for SDK connection errors

**Transfer Issues:**

- Verify `TOKEN_SYMBOL` is supported by the Clearnode instance
- Check faucet account balance meets the minimum threshold
- Review Clearnode transfer logs
