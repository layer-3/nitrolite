## Foundry

**Foundry is a blazing fast, portable and modular toolkit for Ethereum application development written in Rust.**

Foundry consists of:

- **Forge**: Ethereum testing framework (like Truffle, Hardhat and DappTools).
- **Cast**: Swiss army knife for interacting with EVM smart contracts, sending transactions and getting chain data.
- **Anvil**: Local Ethereum node, akin to Ganache, Hardhat Network.
- **Chisel**: Fast, utilitarian, and verbose solidity REPL.

## Documentation

https://book.getfoundry.sh/

## Usage

### Build

```shell
$ forge build
```

### Test

```shell
$ forge test
```

### Format

```shell
$ forge fmt
```

### Gas Snapshots

```shell
$ forge snapshot
```

### Anvil

```shell
$ anvil
```

### Deploy

```shell
$ forge script script/Counter.s.sol:CounterScript --rpc-url <your_rpc_url> --private-key <your_private_key>
```

### Cast

```shell
$ cast <subcommand>
```

### Help

```shell
$ forge --help
$ anvil --help
$ cast --help
```

## Parametric Tokens

Nitrolite supports tokens with additional parameters (e.g., mintTime) through the `IParametricToken` interface. These tokens maintain separate balances per sub-account to preserve parameter integrity. `ParametricToken` contract provides implementation of a token with both mutable and immutable parameters.

### How It Works

When a token is marked parametric, the ChannelHub contract:

1. Converts its own account to Super account on the token
2. Creates a new sub-account for each channel at channel creation time
3. Stores the sub-account ID in channel metadata

All subsequent deposits, withdrawals, and transfers for that channel automatically use the correct sub-account.

### Important: Channel Must Exist First

For parametric tokens, funds **cannot be deposited before channel creation**. The workflow is:

1. **Create channel** → ChannelHub creates a sub-account and returns channel ID
2. **Deposit** → Funds go to the channel's sub-account
3. **Transfer/Withdraw** → Funds move from/to the sub-account

Depositing a parametric token without an existing channel leads to token lock and requires reclaim.

### Enabling Parametric Token Support

The vault contract owner must perform two steps:

```solidity
// Step 1: Mark token as parametric
channelHub.setParametricToken(tokenAddress, true);

// Step 2: Convert ChannelHub to Super account on the token
IParametricToken(tokenAddress).convertToSuper(address(channelHub));
```

After this, channel creation and deposits work through the standard NitroliteClient API - no additional user action required.

### Low-Level Access

For advanced use cases, the `IParametricToken` interface exposes direct sub-account operations:

- `transferToSub()` - Transfer from normal account to a vault sub-account

- `transferFromSub()` - Transfer from a vault sub-account to normal account

- `transferBetweenSubs()` - Transfer between sub-accounts of the same super account (including vault)

These are intended for custom integrations and use `subId` for sub-account identification; standard channel operations handle sub-accounts automatically.

### Standard ERC20 Tokens

For non-parametric tokens (USDC, ETH, etc.), the `isParametricToken` flag is disabled by default and no sub-accounts are created.
