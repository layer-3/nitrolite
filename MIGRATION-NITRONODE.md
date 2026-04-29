# Clearnode → Nitronode Rename

`clearnode` was renamed to `nitronode` starting with the release that ships
this document. Releases up to and including **v1.2.0** were published as
`clearnode`; **v1.3.0 and later** ship as `nitronode`.

The protocol wire format and on-chain contracts are unaffected — this is a
naming change only. The rename does not break compatibility for clients
already authenticated against an existing node, but operators must update
deployments, image references and environment variables.

## What changed

| Surface | Before (≤ v1.2.0) | After (≥ v1.3.0) |
|---------|-------------------|------------------|
| Repository directory | `clearnode/` | `nitronode/` |
| Go binary | `clearnode` | `nitronode` |
| Go import path | `github.com/layer-3/nitrolite/clearnode/...` | `github.com/layer-3/nitrolite/nitronode/...` |
| Docker image | `ghcr.io/layer-3/nitrolite/clearnode:*` | `ghcr.io/layer-3/nitrolite/nitronode:*` |
| Helm chart name | `clearnode` | `nitronode` |
| Helm release / namespace | `clearnode` / `clearnode-<env>` | `nitronode` / `nitronode-<env>` |
| Env var prefix | `CLEARNODE_*` | `NITRONODE_*` |
| Prometheus metric prefix | `clearnode_*` | `nitronode_*` |
| Default sandbox WebSocket URL | `wss://clearnode-sandbox.yellow.org/v1/ws` | `wss://nitronode-sandbox.yellow.org/v1/ws` |
| Connection error message | `failed to connect to clearnode: …` | `failed to connect to nitronode: …` |

## Backwards compatibility

* **Env vars** — `nitronode` accepts both prefixes. Any `CLEARNODE_*` variable
  is automatically mapped to its `NITRONODE_*` counterpart at startup with a
  deprecation warning. The legacy prefix will be removed in a future release.
* **Cerebro config dir** — if a `clearnode-cli` directory already exists in
  the user config path, it is reused with a warning.
* **Old Docker images** — pre-rename tags remain published under
  `ghcr.io/layer-3/nitrolite/clearnode`. New tags are published under
  `nitronode` only; there is no dual-publish window.
* **Prometheus metrics** — there is no automatic alias. Dashboards and alert
  rules that reference `clearnode_*` must be updated to `nitronode_*`.
* **Old Helm configs** — historical environment overrides (`sandbox-v1`,
  `stress-v1`, `v1-rc`) are kept under `nitronode/chart/config/old/` for
  reference and are no longer wired into any helmfile.

## Steps for operators

1. Pull the new image: `ghcr.io/layer-3/nitrolite/nitronode:<tag>`.
2. Update Helm release / namespace names if you manage them yourself.
3. Rename `CLEARNODE_*` env vars to `NITRONODE_*` (the legacy names still
   work for now but emit warnings).
4. Update Prometheus alert rules and Grafana dashboards: replace
   `clearnode_*` metric names with `nitronode_*`.
5. Update any pinned Docker image references in CI / deployment manifests.
6. Update DNS / SDK clients to use `nitronode-sandbox.yellow.org` (the
   legacy host remains live for v1.2.0 and earlier).

## Steps for SDK users

The Go SDK and TypeScript SDK keep the same package paths
(`github.com/layer-3/nitrolite/sdk/go`, `@yellow-org/sdk`) and the same
exported API. Required updates:

* Default WebSocket URL constants now point at `nitronode-sandbox.yellow.org`.
* Error message strings change from `failed to connect to clearnode` to
  `failed to connect to nitronode`. Callers asserting on the old text must
  update their tests.
