# Protocol Drift Guards

This repo has deterministic drift checks for protocol, `@yellow-org/sdk`, and `@yellow-org/sdk-compat` surfaces that demo apps depend on.

## Commands

Run all implemented static checks from the repo root:

```bash
./scripts/check-protocol-drift.sh --static
```

Run package checks directly:

```bash
cd sdk/ts && npm run drift:check
cd sdk/ts-compat && npm run drift:check
```

Run the lightweight runtime smoke from the repo root:

```bash
./scripts/check-protocol-drift.sh --runtime
```

The runtime smoke builds the TS SDK, builds TS compat, builds a temporary local Clearnode binary, starts it with isolated SQLite config, and connects to `ws://127.0.0.1:7824/ws`. It checks `ping`, `getConfig`, `getAssets`, `getAppSessions`, key-state reads, and compat mapping over the live SDK app-session result.

This is not a load test. It uses empty local `blockchains` and `assets` config so PR CI does not depend on external RPC endpoints, wallets, or shared Clearnode deployments.

## Guard Layers

- RPC method drift: compares Go RPC method literals, Clearnode router registrations, TS method constants, and public TS client wrappers.
- RPC DTO drift: compares Go JSON-tagged DTO structs against TS request/response interfaces for required fields, optional fields, and scalar/container shape.
- Public API drift: snapshots root runtime exports and compiler-derived TypeScript signatures for `@yellow-org/sdk` and `@yellow-org/sdk-compat`, including type-only exports, interfaces, functions, classes, public class methods, enums, constants, and type aliases.
- ABI drift: compares checked-in `ChannelHub` functions against the current Foundry artifact, checks SDK-consumed ERC20 functions against the ERC20 artifact, and guards the manually checked-in AppRegistry ABI against an explicit consumed-function manifest until that contract artifact exists in this repo.
- Signing drift: compares TS app-session and session-key packers against Go-generated canonical vectors for create, deposit, withdraw, operate, fractional decimal, and uint64 boundary cases.
- Transform drift: checks raw Clearnode response fixtures for app sessions, node config, assets, and strict failure on unsupported required shapes.
- Compat drift: checks current v1 app-session shape, legacy flat fallback shape, and asset decimal conversion in `NitroliteClient.getAppSessionsList()`.
- Runtime smoke drift: starts an isolated local Clearnode and verifies live SDK/compat calls against the current runtime response shape.

## Intentional Updates

For intentional public runtime API changes, update snapshots with:

```bash
cd sdk/ts && npm run drift:check -- -u
cd sdk/ts-compat && npm run drift:check -- -u
```

For intentional ABI changes, regenerate artifacts and SDK ABI files before running drift checks:

```bash
cd contracts && forge build
cd ../sdk/ts && npm run codegen-abi
```

For a new RPC method, update all applicable surfaces in the same PR: Go method constants, router registration, TS method constants, and the public TS client wrapper unless the method is intentionally raw-only.

For a new DTO field, update the Go JSON-tagged struct and TS request/response interface together. Optionality must match unless a small, named override is added to the drift test.

For a new response transform, add a raw fixture and expected behavior test. Unsupported wire shapes should fail clearly instead of silently producing partial data.

For intentional app/session-key signing vector changes, regenerate the Go source-of-truth hashes from the repo root:

```bash
go run ./scripts/drift/generate-app-signing-vectors.go
```

Then update `sdk/ts/test/unit/app-signing-drift.test.ts` with the changed hashes in the same PR as the Go packing/protocol change.

## Adversarial Proof

Each guard includes at least one negative test or mutation-style check that proves the guard would fail if the relevant surface drifted. These checks must use fixtures, temp copies, or local in-test mutations. They must not leave tracked files dirty.

## CI Policy

`Test (Protocol Drift Static)` runs on PRs and main pushes. It is deterministic and does not call shared Clearnode deployments.

`Test (Protocol Drift Runtime)` also runs on PRs and main pushes. It starts an isolated local Clearnode inside the GitHub Actions job and does not use shared stress or sandbox endpoints.

If runtime smoke fails in CI, inspect the `protocol-drift-runtime-smoke-logs` artifact. The smoke command categorizes failures as setup, startup, connection, timeout, transform, or compat mapping failures.

The runtime job uses read-only repository permissions and no secrets. It builds Clearnode locally instead of pulling or publishing an image, so ordinary PRs do not need package-write permissions. If organization policy restricts forked PR workflows, a maintainer can rerun the same command locally or through an allowed CI rerun.

Shared stress Clearnode checks are manual/nightly only. `wss://clearnode-stress.yellow.org/v1/ws` can be newer than sandbox while audit remediations roll out, so it is useful for release confidence but must not be a default PR blocker.
