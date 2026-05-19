# Nitrolite builder skills

A small reference library of [Claude Code skills](https://claude.com/blog/claude-code-skills) for builders using the current Nitrolite v1 SDK and protocol surface. These skills are intentionally scoped to this repository's v1 sources of truth: `docs/api.yaml`, `sdk/ts/src/rpc/methods.ts`, and `contracts/src/ChannelHub.sol`.

## Scope

Use these skills for new integrations against `@yellow-org/sdk` v1.x. They do not document the legacy v0.5.x compatibility API, flat legacy RPC method names, or contracts that are not present in this monorepo.

## Catalog

- [`yellow-network-builder`](./yellow-network-builder/SKILL.md) — orientation for new Nitrolite builders: repository map, v1 source-of-truth files, protocol layers, and setup path.
- [`yellow-sdk-v1`](./yellow-sdk-v1/SKILL.md) — active `@yellow-org/sdk` v1.x TypeScript SDK: `Client.create`, signers, option factories, high-level methods, and v1 RPC namespaces.

## Frontmatter Conventions

Each `SKILL.md` carries:

- `name` — the skill identifier, matching the directory.
- `description` — one paragraph explaining when to load the skill.
- `version` — semantic version of the skill document itself.
- `sdk_version` — SDK release family the examples target.
- `network` — intended network context.
- `last_verified` — date the skill was checked against repo-local sources.
- `source_urls` — canonical upstream or repo-local references used for verification.
