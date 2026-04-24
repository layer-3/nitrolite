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

`./scripts/check-protocol-drift.sh --runtime` is reserved for the future ephemeral Clearnode smoke harness. It intentionally fails until that harness exists.

## Guard Layers

- RPC method drift: compares Go RPC method literals, Clearnode router registrations, TS method constants, and public TS client wrappers.
- RPC DTO drift: compares Go JSON-tagged DTO structs against TS request/response interfaces for required fields, optional fields, and scalar/container shape.
- Public runtime API drift: snapshots root runtime exports for `@yellow-org/sdk` and `@yellow-org/sdk-compat`.
- ABI drift: compares SDK-consumed `ChannelHub` ABI functions against the current Foundry artifact.
- Signing drift: compares TS app-session packers against Go-generated canonical vectors.
- Transform drift: checks raw Clearnode response fixtures for app sessions, node config, assets, and strict failure on unsupported required shapes.
- Compat drift: checks current v1 app-session shape, legacy flat fallback shape, and asset decimal conversion in `NitroliteClient.getAppSessionsList()`.

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

## Adversarial Proof

Each guard includes at least one negative test or mutation-style check that proves the guard would fail if the relevant surface drifted. These checks must use fixtures, temp copies, or local in-test mutations. They must not leave tracked files dirty.

## CI Policy

`Test (Protocol Drift Static)` runs on PRs and main pushes. It is deterministic and does not call shared Clearnode deployments.

Shared stress Clearnode checks are manual/nightly only. `wss://clearnode-stress.yellow.org/v1/ws` can be newer than sandbox while audit remediations roll out, so it is useful for release confidence but must not be a default PR blocker.
