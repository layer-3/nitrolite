# Clearnode CLI

Command-line interface for the Clearnode Go SDK. Provides interactive access to both high-level smart operations and low-level RPC methods.

## Installation

```bash
cd examples/cli
go build -o clearnode-cli
```

Or install directly:

```bash
go install github.com/layer-3/nitrolite/sdk/go/examples/cli@latest
```

## Quick Start

```bash
# Connect to node
./clearnode-cli wss://clearnode.example.com/ws

# Configure wallet
clearnode> import wallet

# Configure RPC endpoint
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/YOUR_KEY

# Verify configuration
clearnode> config

# Deposit funds
clearnode> deposit 80002 usdc 100

# Transfer funds
clearnode> transfer 0xRecipient... usdc 50

# Withdraw funds
clearnode> withdraw 80002 usdc 25
```

## Commands

### Configuration

```
help                            Display command reference
config                          Display current configuration
wallet                          Display wallet address
import wallet                   Configure wallet (import or generate)
import rpc <chain_id> <url>     Configure RPC endpoint
exit                            Exit the CLI
```

### High-Level Operations

```
deposit <chain_id> <asset> <amount>           Deposit to channel
withdraw <chain_id> <asset> <amount>          Withdraw from channel
transfer <recipient> <asset> <amount>         Transfer to another wallet
```

### Node Information

```
ping                            Test node connection
node info                       Get node configuration
chains                          List supported blockchains
assets [chain_id]               List supported assets
```

### User Queries

```
balances [wallet]               Get user balances
transactions [wallet]           Get transaction history
```

### State Management

```
state [wallet] <asset>          Get latest state
home-channel [wallet] <asset>   Get home channel
escrow-channel <channel_id>     Get escrow channel by ID
```

### App Sessions

```
app-sessions                    List app sessions
```

### Session Key Management

```
generate-session-key                                                Generate a new session keypair
create-channel-session-key <session_key> <expires_hours> <assets>   Register channel session key
channel-session-keys                                                List active channel session keys
create-app-session-key <session_key> <expires_hours> [app_ids] [session_ids]  Register app session key
app-session-keys                                                    List active app session keys
```

## Configuration Storage

Default configuration locations:

- Linux: `~/.config/clearnode-cli/config.db`
- macOS: `~/Library/Application Support/clearnode-cli/config.db`
- Windows: `%APPDATA%\clearnode-cli\config.db`

Override with environment variable:

```bash
export CLEARNODE_CLI_CONFIG_DIR=/custom/path
```

## Wallet Setup

When running `import wallet`, choose:

1. **Import existing** - Enter your private key (with or without 0x prefix)
2. **Generate new** - Create new wallet with random key

WARNING: Save generated private keys immediately. They cannot be recovered.

## Command Parameters

### Chain IDs

Use blockchain chain IDs (e.g., 80002 for Polygon Amoy, 84532 for Base Sepolia).

### Asset Symbols

Use lowercase asset symbols (e.g., usdc, eth, dai).

### Wallet Addresses

Full Ethereum addresses starting with 0x. Commands default to configured wallet when address is omitted.

### Amounts

Decimal amounts (e.g., 100, 0.5, 1000.25).

### Session Key Addresses

Full Ethereum addresses (0x-prefixed) of the session key to delegate signing authority to. Generate one with `generate-session-key`.

### Expiration Hours

Integer number of hours until a session key expires (e.g., 24, 48, 168).

### Comma-Separated Lists

Assets, application IDs, and session IDs are passed as comma-separated values without spaces (e.g., `usdc,weth` or `app1,app2`).

## Interactive Features

- **Tab completion** - Press Tab for command suggestions
- **Command history** - Use arrow keys to navigate history
- **Context-aware** - Chain IDs and assets autocomplete based on node data

## Error Messages

Errors are prefixed by type:

- `ERROR:` - Command failed, check parameters
- `WARNING:` - Non-critical issue
- `INFO:` - Informational message

## Examples

### Initial Setup

```bash
./clearnode-cli wss://testnet.clearnode.example.com/ws
clearnode> import wallet
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
clearnode> config
```

### Deposit and Transfer

```bash
clearnode> deposit 80002 usdc 1000
clearnode> balances
clearnode> transfer 0xRecipient... usdc 100
clearnode> transactions
```

### Query Network

```bash
clearnode> chains
clearnode> assets
clearnode> node info
```

### Inspect State

```bash
clearnode> state usdc
clearnode> home-channel usdc
clearnode> balances 0xSomeAddress...
```

### Session Keys

```bash
# Generate a new session keypair
clearnode> generate-session-key

# Register a channel session key (valid for 24 hours, for usdc and weth)
clearnode> create-channel-session-key 0xSessionKeyAddr... 24 usdc,weth

# List active channel session keys
clearnode> channel-session-keys

# Register an app session key (valid for 48 hours, for specific app IDs)
clearnode> create-app-session-key 0xSessionKeyAddr... 48 app1,app2

# List active app session keys
clearnode> app-session-keys
```

## Security

- Private keys stored locally in SQLite (unencrypted)
- Database protected by OS file permissions
- Never commit config database to version control
- Backup private keys securely
- Use hardware wallets for production

## Troubleshooting

**ERROR: No wallet configured**
```bash
clearnode> import wallet
```

**ERROR: No RPC configured for chain X**
```bash
clearnode> import rpc <chain_id> <rpc_url>
```

**Connection issues**
- Verify WebSocket URL
- Check network connectivity
- Confirm node is accessible

**Insufficient balance**
- Verify wallet has sufficient tokens
- Check token approval for contract
- Ensure gas funds available

## Architecture

```
cli/
├── main.go         Entry point and terminal setup
├── operator.go     Command routing and completion
├── commands.go     Command implementations
└── storage.go      SQLite configuration storage
```

Uses layered architecture:
- Smart Client for high-level operations
- Base Client for low-level RPC access
- Local SQLite for secure configuration storage

## License

Part of the Nitrolite project.
