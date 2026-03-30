---
paths:
  - "contracts/**/*.sol"
---

- Foundry project — use `forge build`, `forge test`, `forge fmt`.
- NatSpec comments on all public/external functions.
- Security first: validate all inputs, check for reentrancy, use OpenZeppelin where applicable.
- Test files in `contracts/test/` with `.t.sol` extension.
- Gas optimization matters — avoid unnecessary storage writes.
