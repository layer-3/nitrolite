# ChannelHub Deployments — Hook-Token Support

The `ChannelHub` deployments listed in the matrix below **do not support
hook-enabled tokens**. The following token classes MUST NOT be onboarded to
any of these deployments:

- **ERC777** (e.g. `imBTC`, legacy `xDAI`)
- **ERC1363 / ERC677**
- **Non-standard ERC20 with re-entrant `transferFrom`** (some rebasing or
  fee-on-transfer tokens with callbacks)

Enforcement is the responsibility of the Node operator. This constraint may
be lifted on future deployments to new chains; each new deployment must be
added to the matrix below with its support status recorded explicitly.

Last reviewed: 2026-06-15.

## Matrix

| Chain ID | Chain | ChannelHub Address | Deploy Commit | Deploy Tag |
| ---: | --- | --- | --- | --- |
| 1 | Ethereum | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 14 | Flare | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 56 | BNB Smart Chain | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 137 | Polygon | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 480 | World Chain | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 8453 | Base | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 59144 | Linea | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 80002 | Polygon Amoy | `0x5dba8515af063db0c243c15ece7b99f91459c7c3` | `b88d511c` | sandbox v1.3.0 |
| 84532 | Base Sepolia | `0x5dba8515af063db0c243c15ece7b99f91459c7c3` | `b88d511c` | sandbox v1.3.0 |
| 59141 | Linea Sepolia | `0x5dba8515af063db0c243c15ece7b99f91459c7c3` | `b88d511c` | sandbox v1.3.0 |
| 1440000 | XRPL EVM | `0x1a2f750170474d4c54f8d318d9d4343588b4c4d1` | `e07ad9c2` | prod v1.3.0 |
| 1449000 | XRPL EVM Testnet | `0x5dba8515af063db0c243c15ece7b99f91459c7c3` | `b88d511c` | sandbox v1.3.0 |
| 11155111 | Sepolia | `0x5dba8515af063db0c243c15ece7b99f91459c7c3` | `b88d511c` | sandbox v1.3.0 |
