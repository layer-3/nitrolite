# Cerebro

REPL for the Nitrolite Go SDK. Provides access to both high-level smart operations and low-level RPC methods.

## Installation

```bash
# From the repository root
go build -o cerebro ./cerebro
```

Or install directly:

```bash
go install github.com/layer-3/nitrolite/cerebro@latest
```

## Quick Start

```bash
# Connect to node (a wallet is auto-generated on first run)
./cerebro wss://clearnode.example.com/ws

# Import an existing wallet (or use the auto-generated one)
cerebro> config wallet import

# Configure RPC endpoint for on-chain operations
cerebro> config rpc import 80002 https://polygon-amoy.g.alchemy.com/v2/YOUR_KEY

# Verify configuration
cerebro> config

# Approve token spending, deposit, transfer
cerebro> approve 80002 usdc 1000
cerebro> deposit 80002 usdc 100
cerebro> transfer 0xRecipient... usdc 50

# Withdraw funds
cerebro> withdraw 80002 usdc 25
```

## Commands

### Configuration

```text
config                                                       Display current configuration
config wallet                                                Display wallet address
config wallet import                                         Import existing private key
config wallet generate                                       Generate new wallet
config wallet export <path>                                  Export private key to file
config rpc import <chain_id> <url>                           Configure blockchain RPC endpoint
config node                                                  Show node info
config node set-ws-url <url>                                 Set clearnode WebSocket URL
config node set-home-blockchain <asset> <chain_id>           Set home blockchain for channels
config session-key                                           Show current session key info
config session-key generate                                  Generate new session key
config session-key import                                    Import existing session key
config session-key clear                                     Clear session key, revert to default signer
config session-key register-channel-key <key> <hours> <assets>     Register channel session key
config session-key channel-keys                              List active channel session keys
config session-key register-app-key <key> <hours> [apps] [sessions]  Register app session key
config session-key app-keys                                  List active app session keys
```

### Operations

```text
token-balance <chain_id> <asset>             Check on-chain token balance
approve <chain_id> <asset> <amount>          Approve token spending for deposits
deposit <chain_id> <asset> <amount>          Deposit to channel (auto-create if needed)
withdraw <chain_id> <asset> <amount>         Withdraw from channel
transfer <recipient> <asset> <amount>        Transfer to another wallet
acknowledge <asset>                          Acknowledge transfer or channel creation
close-channel <asset>                        Close home channel on-chain
checkpoint <asset>                           Submit latest state on-chain
```

### Queries

```text
ping                          Test node connection
chains                        List supported blockchains
assets [chain_id]             List supported assets (optionally filter by chain)
balances [wallet]             Get user balances (defaults to configured wallet)
transactions [wallet]         Get transaction history
action-allowances [wallet]    Get action allowances
state [wallet] <asset>        Get latest state
home-channel [wallet] <asset> Get home channel
escrow-channel <channel_id>   Get escrow channel by ID
```

### App Registry

```text
app-info <app_id>                    Show application details
my-apps                              List your registered applications
register-app <app_id> [no-approval]  Register a new application
app-sessions                         List app sessions
```

### Security Token Operations

```text
security-token approve <chain_id> <amount>                  Approve security token spending
security-token balance <chain_id> [wallet]                  Check escrowed security token balance
security-token escrow <chain_id> [target_address] <amount>  Escrow security tokens
security-token initiate-withdrawal <chain_id>               Start unlock period
security-token cancel-withdrawal <chain_id>                 Cancel unlock and re-lock
security-token withdraw <chain_id> <destination>            Withdraw unlocked security tokens
```

### Other

```text
help                          Display help message
exit                          Exit the CLI
```

## Configuration Storage

Default configuration locations:

- Linux: `~/.config/cerebro/config.db`
- macOS: `~/Library/Application Support/cerebro/config.db`
- Windows: `%APPDATA%\cerebro\config.db`

If a legacy `clearnode-cli` directory exists, it will be used with a warning suggesting you rename it to `cerebro`.

Override with environment variable:

```bash
export CLEARNODE_CLI_CONFIG_DIR=/custom/path
```

## Wallet Setup

On first run, a new wallet is automatically generated. You can also:

1. **Import existing** - `config wallet import` to enter your private key
2. **Generate new** - `config wallet generate` to create a new wallet
3. **Export** - `config wallet export <path>` to save your private key to a file

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

### Comma-Separated Lists

Assets, application IDs, and session IDs are passed as comma-separated values without spaces (e.g., `usdc,weth` or `app1,app2`).

## Interactive Features

- **Tab completion** - Press Tab for command suggestions
- **Command history** - Use arrow keys to navigate history
- **Context-aware** - Chain IDs and assets autocomplete based on node data

## Examples

### Initial Setup

```bash
./cerebro wss://testnet.clearnode.example.com/ws
cerebro> config wallet import
cerebro> config rpc import 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
cerebro> config
```

### Deposit and Transfer

```bash
cerebro> approve 80002 usdc 1000
cerebro> deposit 80002 usdc 1000
cerebro> balances
cerebro> transfer 0xRecipient... usdc 100
cerebro> transactions
```

### Session Keys

```bash
# Generate a new session keypair
cerebro> config session-key generate

# Register a channel session key (valid for 24 hours, for usdc and weth)
cerebro> config session-key register-channel-key 0xSessionKeyAddr... 24 usdc,weth

# List active channel session keys
cerebro> config session-key channel-keys

# Register an app session key (valid for 48 hours, for specific app IDs)
cerebro> config session-key register-app-key 0xSessionKeyAddr... 48 app1,app2

# List active app session keys
cerebro> config session-key app-keys
```

### Query Network

```bash
cerebro> chains
cerebro> assets
cerebro> config node
```

### Inspect State

```bash
cerebro> state usdc
cerebro> home-channel usdc
cerebro> balances 0xSomeAddress...
```

## Security

- Private keys stored locally in SQLite (unencrypted)
- Database protected by OS file permissions
- Never commit config database to version control
- Backup private keys securely
- Use hardware wallets for production

## Architecture

```text
cerebro/
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

Part of the Nitrolite project. Licensed under the MIT License.
