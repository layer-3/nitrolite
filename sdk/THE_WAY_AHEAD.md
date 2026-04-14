# The Way Ahead: Productionalising Nitrolite AI Tooling

This document is the roadmap for taking our MCP servers from "works if you clone the repo" to "any developer, anywhere, gets Nitrolite AI tooling with one command." It's written for someone who has never published an npm package or a Go binary before.

---

## Where We Are

| What we have | Status |
|---|---|
| Unified MCP server (`sdk/mcp/`) | Works locally, 30 resources, 8 tools, 3 prompts — covers TypeScript and Go SDKs |
| Protocol docs, clearnode API, SDK source | Indexed at startup by the MCP server |

**The problem:** The server reads files from the repo at runtime (`docs/protocol/`, `docs/api.yaml`, `sdk/ts/src/`, `sdk/go/`). If you don't have the full repo cloned, it doesn't work. No external developer can use it.

---

## IMPORTANT: Release Parity

**The MCP server currently reads from repo source on disk, not from the published SDK release.** This means:

- The MCP parses `sdk/ts/src/client.ts` directly — which may contain unreleased methods not yet in `@yellow-org/sdk` on npm
- The MCP parses `sdk/go/*.go` — same issue, may expose unreleased Go API surface
- Protocol docs in `docs/protocol/` may document unreleased behavior

**This is fine for internal development** (you want AI to know about in-progress code). But for the published MCP package, this creates a real problem: an AI agent could tell an external developer "use `client.someNewMethod()`" — but when they `npm install @yellow-org/sdk`, that method doesn't exist yet.

**When publishing (Phase 1 below), you MUST:**

1. **Embed content from the tagged release commit, not from main.** The CI workflow that publishes the MCP package should check out the corresponding SDK release tag (e.g., `sdk/ts/v1.2.0`) and embed content from that snapshot.
2. **Version-lock the MCP to its SDK.** The MCP package version should track the SDK version it documents. If `@yellow-org/sdk` is at `1.2.0`, the MCP should be `1.2.0` too (or `1.2.0-mcp.1` if iterating on MCP-only changes).
3. **Add a version check at startup.** When running locally from repo source (not published), the MCP should log a warning: "Reading from source — API surface may differ from published SDK."
4. **Coordinate release timing.** Publish the MCP package as part of the same release workflow that publishes the SDK, so they're always in sync.

This is the single most important detail to get right before publishing. An MCP that tells developers about methods they can't use is worse than no MCP at all.

---

## Phase 1: Publish TS MCP to npm

**Goal:** Any developer runs `npx @yellow-org/sdk-mcp` and gets the full Nitrolite MCP server.

### Step 1: Embed content at build time

The server currently reads files from disk. For npm, we need to bundle all content into the compiled JavaScript. Two approaches:

**Option A (simpler):** A build script that reads all protocol docs, examples, and SDK source, then generates a `src/embedded-content.ts` file with all content as exported string constants. The server imports from this file instead of using `fs.readFileSync`.

**Option B:** Use a bundler like `tsup` or `esbuild` with a plugin that inlines file reads at build time.

Option A is recommended — it's explicit, debuggable, and doesn't add build tool complexity.

### Step 2: Set up package.json for publishing

Current `package.json` needs these changes:

```jsonc
{
  "name": "@yellow-org/sdk-mcp",
  "version": "0.1.0",
  "description": "MCP server exposing the Nitrolite SDK to AI agents and IDEs",
  "main": "dist/index.js",           // compiled output, not .ts source
  "types": "dist/index.d.ts",
  "bin": {
    "nitrolite-sdk-mcp": "dist/index.js"
  },
  "files": [                          // what goes in the npm tarball
    "dist",
    "README.md"
  ],
  "publishConfig": {
    "access": "public"               // scoped packages default to private!
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/layer-3/nitrolite.git",
    "directory": "sdk/mcp"
  },
  "license": "MIT",
  "scripts": {
    "build": "tsc",
    "prepublishOnly": "npm run build"  // auto-builds before publish
  }
}
```

**Important:** The compiled `dist/index.js` needs a shebang (`#!/usr/bin/env node`) at the top for the `bin` entry to work. TypeScript's `tsc` strips shebangs. Fix with a postbuild step:

```json
"postbuild": "echo '#!/usr/bin/env node' | cat - dist/index.js > temp && mv temp dist/index.js"
```

Or use `tsup` as the build tool, which handles shebangs natively.

### Step 3: Create an npm account and get access

1. **Create account:** Go to https://www.npmjs.com/signup
2. **Enable 2FA:** Avatar > Account Settings > Two-Factor Authentication. Scan QR code with an authenticator app. This is mandatory for publishing.
3. **Get org access:** The `@yellow-org` org already exists (it publishes `@yellow-org/sdk`). An existing org admin must invite you at https://www.npmjs.com/settings/yellow-org/members with the "developer" role (can publish packages).

### Step 4: Publish manually (first time)

```bash
# Login to npm (opens browser for auth)
npm login

# Verify you're logged in
npm whoami

# Go to the package directory
cd sdk/mcp

# Build
npm run build

# Preview what will be in the tarball (sanity check)
npm pack --dry-run

# Publish! --access public is critical for scoped packages
npm publish --access public
```

After this, anyone in the world can run:
```bash
npx @yellow-org/sdk-mcp
```

### Step 5: Automate publishing with GitHub Actions

After the first manual publish, set up automation so future versions publish on git tag.

**Option A: Trusted Publishing (recommended, no secrets needed)**

1. On npmjs.com, go to the package > Access > Trusted Publishers > Add GitHub Actions
2. Enter: org=`layer-3`, repo=`nitrolite`, workflow=`publish-sdk-mcp.yml`
3. Create the workflow:

```yaml
# .github/workflows/publish-sdk-mcp.yml
name: Publish @yellow-org/sdk-mcp

on:
  push:
    tags:
      - 'sdk-mcp/v*'       # triggers on tags like sdk-mcp/v0.2.0

permissions:
  contents: read
  id-token: write            # required for OIDC auth with npm

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'
      - run: npm ci
        working-directory: ./sdk/mcp
      - run: npm run build
        working-directory: ./sdk/mcp
      - run: npm publish --access public --provenance
        working-directory: ./sdk/mcp
        # No NODE_AUTH_TOKEN needed! OIDC handles it.
```

4. To publish a new version:
```bash
# Bump version in package.json
cd sdk/mcp
npm version patch  # or minor, or major

# Tag and push
git add sdk/mcp/package.json
git commit -m "chore(sdk-mcp): bump to v0.2.0"
git tag sdk-mcp/v0.2.0
git push origin main
git push origin sdk-mcp/v0.2.0
# GitHub Actions takes over from here
```

**Option B: npm token (fallback)**

If trusted publishing doesn't work for your setup:
1. Create a granular access token at npmjs.com > Access Tokens
2. Scope it to `@yellow-org/sdk-mcp` only, with "Require 2FA" disabled (so CI can use it)
3. Add it as `NPM_TOKEN` in GitHub repo secrets
4. Add to the publish step: `env: NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}`

### Step 6: Submit to the official MCP Registry

The MCP Registry (registry.modelcontextprotocol.io) is a discovery directory. Getting listed means AI tools can find your server automatically.

```bash
# Install the publisher CLI
brew install modelcontextprotocol/tap/mcp-publisher

# Generate server.json in sdk/mcp/
cd sdk/mcp
mcp-publisher init
```

Edit `server.json`:
```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.layer-3/nitrolite-sdk-mcp",
  "description": "MCP server exposing the Nitrolite state channel SDK to AI agents",
  "version": "0.1.0",
  "repository": {
    "url": "https://github.com/layer-3/nitrolite",
    "source": "github",
    "id": "layer-3/nitrolite",
    "subfolder": "sdk/mcp"
  },
  "packages": [{
    "registry_type": "npm",
    "name": "@yellow-org/sdk-mcp",
    "version": "0.1.0",
    "runtime": "node",
    "package_arguments": [],
    "environment_variables": []
  }]
}
```

```bash
# Authenticate with GitHub
mcp-publisher login github    # opens browser

# Publish to registry
mcp-publisher publish
```

Verify: `curl "https://registry.modelcontextprotocol.io/v0.1/servers?search=nitrolite"`

**When updating:** Bump both `version` and `packages[].version` in `server.json`, then run `mcp-publisher publish` again.

### Common npm gotchas

- Scoped packages (`@yellow-org/*`) default to **private**. Always use `--access public` on first publish or set `publishConfig.access` in package.json
- `npm pack --dry-run` shows exactly what goes in the tarball — use it before every publish
- You **cannot unpublish** a version after 72 hours. Version numbers are permanent.
- npm requires 2FA for all publishing since late 2025

---

Go SDK context is now served by the unified TypeScript MCP server (`sdk/mcp`). A separate Go binary distribution is no longer planned — the single npm-published `@yellow-org/sdk-mcp` package covers both TypeScript and Go SDK surfaces.

---

## Phase 2: Cross-IDE Context Files

These are simple text files that make the repo AI-friendly for every coding tool — not just Claude Code.

| File | Tool | What it does |
|------|------|-------------|
| `llms.txt` | Any LLM / web crawler | Machine-readable summary of the project |
| `llms-full.txt` | Claude Projects, ChatGPT | Full protocol + SDK reference in one file |
| `.cursorrules` | Cursor | Project-specific AI coding rules |
| `.github/copilot-instructions.md` | GitHub Copilot | Org-wide coding conventions (4K char limit) |
| `.windsurfrules` | Windsurf | Same as cursorrules |

These are low-effort, high-reach. Content already exists in `.claude/rules/` and `CLAUDE.md` — it just needs to be reformatted.

---

## Phase 3: Agent Skills (SKILL.md)

The SKILL.md format (from Anthropic, adopted by OpenAI Codex and VS Code Copilot) teaches AI agents *how to do things* — complementary to MCP which provides *tools*.

Create a `skills/` directory or separate repo (`yellow-org/nitrolite-skills`) with:

| Skill | What it teaches |
|-------|----------------|
| `build-transfer-app` | How to build a token transfer app from scratch |
| `build-app-session` | How to build a multi-party app session |
| `migrate-sdk` | How to migrate from SDK v0.5.3 to v1.x |
| `build-ai-agent` | How to build an autonomous payment agent |

Published via: `npx skills add yellow-org/nitrolite-skills`

---

## Phase 4: Hosted Remote MCP (Future)

Deploy the MCP server as a hosted HTTP endpoint so developers don't need to install anything at all:

```json
{
  "mcpServers": {
    "nitrolite": {
      "url": "https://mcp.yellow.org/ts"
    }
  }
}
```

Uses MCP Streamable HTTP transport. Requires infrastructure (Cloudflare Workers or similar) and API key management. Only Thirdweb does this in Web3 today.

---

## Summary: Developer Experience by Phase

| Phase | Developer experience | Effort |
|-------|---------------------|--------|
| Today | Clone repo, run from source | Already done |
| Phase 1 (npm publish) | `npx @yellow-org/sdk-mcp` — covers both TS and Go SDKs | Medium |
| Phase 2 (context files) | IDE auto-loads project rules | Low |
| Phase 3 (skills) | `npx skills add yellow-org/nitrolite-skills` | Medium |
| Phase 4 (hosted) | Add a URL to IDE config | High |

**Priority order:** Phase 1 > Phase 2 > Phase 3 > Phase 4

Phase 1 (npm publish) is the single highest-impact step. It's the difference between "only our team can use this" and "any developer on the internet can use this with one command."
