# batchDepositToNode

Batch `approve` + `depositToNode` across multiple chains and tokens.

Chains run in parallel; tokens within each chain run sequentially with on-chain
confirmation waited between each token (nonce safety). Token amounts in the config
are human-readable (e.g. `1000` for 1000 USDC) — the script fetches `decimals()`
from each token contract and converts automatically.

## Prerequisites

- [Foundry](https://getfoundry.sh/) (`forge`, `cast`)
- `jq`
- A signing method: keystore account, interactive key entry, or Ledger

**Set up a keystore account (recommended):**
```bash
cast wallet import topup-signer --interactive
# Prompts for private key and encryption password.
# Key is stored encrypted in ~/.foundry/keystores/topup-signer
```

## Config format

Create a `batchDepositToNode.json` file (not committed — contains RPC URLs):

```json
[
  {
    "rpcUrl": "https://mainnet.infura.io/v3/YOUR_KEY",
    "hubAddress": "0xCe87FD88F4B5Fd5475d163e2642C5c2c7dD655Ec",
    "tokens": [
      { "address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", "amount": "1000" },
      { "address": "0xdAC17F958D2ee523a2206206994597C13D831ec7", "amount": "500"  }
    ]
  },
  {
    "rpcUrl": "https://arb-mainnet.g.alchemy.com/v2/YOUR_KEY",
    "hubAddress": "0x...",
    "tokens": [
      { "address": "0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8", "amount": "1000" }
    ]
  }
]
```

`amount` is in human-readable token units (e.g. `"1000"` = 1000 USDC). The script
fetches `decimals()` from the token contract at runtime and computes the raw on-chain
value (`1000 × 10^6 = 1000000000` for USDC). Fractional amounts are supported
(e.g. `"0.5"`), but values requiring more than 15 significant digits may lose
precision due to floating-point arithmetic — use whole numbers for large amounts.

## Usage

Run from the `contracts/` directory:

```bash
# Keystore account — password prompted once, reused across all chains and tokens
./script/batchDepositToNode.sh batchDepositToNode.json --account topup-signer

# Interactive private key — prompted once, temp keystore created automatically
./script/batchDepositToNode.sh batchDepositToNode.json -i

# Ledger hardware wallet (sequential only — see note below)
./script/batchDepositToNode.sh batchDepositToNode.json -- --ledger

# Dry run — simulate without broadcasting (no txs sent, no confirmation wait)
./script/batchDepositToNode.sh batchDepositToNode.json --dry-run --account topup-signer

# Extra forge flags after --  (e.g. verbosity)
./script/batchDepositToNode.sh batchDepositToNode.json --account topup-signer -- -vvvv
```

`--account` and `-i` are shell-level flags handled before any forge invocations —
the password or key is collected once, then all forge calls reuse it silently.
Anything after `--` is forwarded verbatim to every `forge script` call.

> **`-i` security note:** The private key is prompted interactively (echo off, not
> in shell history). Internally the script runs one `cast wallet import` call to
> create a temp keystore — the key appears in the process list for that one brief
> call only, then the keystore (random password, deleted on exit) is used for all
> subsequent forge invocations. Clear your clipboard after pasting: `pbcopy < /dev/null`.

> **Ledger note:** `--ledger` requires sequential chain execution. The Ledger USB
> device can only be opened by one process at a time — running chains in parallel
> will cause all but the first to fail immediately. For Ledger, run chains one at a
> time by passing a single-entry config, or iterate manually.

## Output

A timestamped log file is written to the directory where the script is invoked:

```text
batchDepositToNode-20260520-143000.log
```

All `forge` and `cast` output is captured. On completion:

```text
=== Summary ===
Deposited 4 tokens across 3 chains
```

On partial failure:

```text
=== Summary ===
Deposited 3 tokens across 3 chains

FAILED:
  chain=42161:  0xFF970A61...
```

Failed tokens are reported but do not stop other chains or tokens from running.
Exit code is `0` on full success, `1` if any token failed.

## How it works

```text
batchDepositToNode.sh
├── chain 1 (background) ──► token A: fetch decimals → approve + depositToNode → await confirm
│                        └──► token B: fetch decimals → approve + depositToNode → await confirm
├── chain 2 (background) ──► token A: fetch decimals → approve + depositToNode → await confirm
│                        └──► ...
└── wait for all → summary
```

For each token, the script:
1. Calls `cast call <token> "decimals()(uint8)"` to get the token's decimal precision
2. Multiplies the human-readable `amount` by `10^decimals` to get the raw on-chain value
3. Runs `forge script DepositToNode.s.sol` with the raw amount (`approve` + `depositToNode`)
4. Reads `broadcast/DepositToNode.s.sol/<chainId>/run-latest.json` and calls
   `cast receipt --confirmations 1` on both tx hashes before moving to the next token

## Single-chain manual run

`DepositToNode.s.sol` can also be called directly. Note that `<AMOUNT>` must be the
raw on-chain value (already multiplied by decimals):

```bash
# Example: 1000 USDC = 1000000000 (6 decimals)
forge script script/DepositToNode.s.sol \
  --sig "run(address,address,uint256)" <HUB> <TOKEN> <RAW_AMOUNT> \
  --rpc-url <RPC_URL> \
  --broadcast \
  --account topup-signer
```
