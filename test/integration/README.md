# Integration Tests

This directory contains integration tests that verify the system components work together correctly.

## Prerequisites

- Node, Docker and NPM installed

## Setup environment

In order to run the integration tests, you need components to integrate with: DB, Clearnode (BE) and blockchain node. All of these three components can be run using Docker.

Go to the root of the project (`..` from this file) and run:

```bash
docker-compose up -d --build
```

It will build clearnode from the current version of the code (you could use docker image instead, but it may not include the latest changes). Compile smart contracts and deploy them on predetermined addresses (based on mnemonic and unique salt, that are specified further in parenthesis):

```
Custody(58808): 0x8658501c98C3738026c4e5c361c6C3fa95DfB255
DummyAdjudicator(35305): 0xcbbc03a873c11beeFA8D99477E830be48d8Ae6D7
USDC ERC20(77360): 0xbD24c53072b9693A35642412227043Ffa5fac382
BalanceChecker(53231): 0x730dB3A1D3Ca47e7BaEb260c24C74ED4378726Bc
```

Additionally it will run migrations on database and seed it with initial data.

## Setup SDK

Integration tests require the sdk to be built in respective directory (`sdk` or `../sdk` from this file)

To set up the SDK, navigate to the `sdk` directory and run:

```bash
npm ci
npm run build
```

It's possible to use specific version of the sdk by updating the `package.json` with a specific version of `@layer-3/nitrolite` and change import path in the `tsconfig.json`

## Running Integration Tests

To run the integration tests, ensure you have the necessary environment set up as described above. Then, navigate to the `integration` directory and execute the following command:

```bash
npm ci
npm run test
```

## Debugging Tests

It may be useful to add some debugging logs to the tests. You can do this by modifying the `CONFIG.DEBUG_MODE` in `common/setup.ts` file. Set it to `true` to enable debug logs.
