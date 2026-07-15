# Yellow Agent Skills

> AI agent skills for Yellow Network Protocol.

Skills that let an AI agent use Yellow's clearing and settlement layer directly: open a shared session with other agents, move value between them off-chain at machine speed, and settle one final split on-chain. Each skill is a self-contained `SKILL.md` written for an agent to read and act on, grounded in the real `@yellow-org/sdk`.

## What is a skill?

A skill is a single Markdown file (`SKILL.md`) plus optional references that teach an AI agent how to do one thing well: the exact methods to call, the order to call them in, the prerequisites, the failure cases, and the trust boundaries to respect. Agents like Claude Code load skills on demand and follow them directly, so the agent writes correct integration code the first time instead of guessing at an API.

## Available skills

| Skill | What it does |
|---|---|
| [`yellow-settlement-room`](./yellow-settlement-room) | Open a shared room where N agents pool funds, reallocate between themselves off-chain with no gas per step, and co-sign one final settlement. Covers connecting, the funded-account prerequisite, creating a session with participant weights and quorum, deposit, operate, withdraw, close, and the trust boundary. |

More skills will be added as the protocol surface grows.

## Installation

### Claude Code

Copy a skill into your project's skills directory:

```bash
mkdir -p .claude/skills
cp -r agent-skills/yellow-settlement-room .claude/skills/
```

The agent picks it up automatically and invokes it when a task matches the skill's description.

### Any agent

Point your agent at the skill's `SKILL.md` and its `references/` folder. The files are plain Markdown and self-contained.

### Method lookups

For live method and type lookups against the SDK while building, run the docs MCP server:

```bash
npx -y @yellow-org/sdk-mcp@^1
```

## Skill structure

```
agent-skills/
  yellow-settlement-room/
    SKILL.md                     # the skill: methods, workflow, trust boundary
    references/
      agent-lifecycle.md         # full multiparty flow as runnable code
    LICENSE
```

## Prerequisites

- Node.js 20+
- `@yellow-org/sdk` (v1)
- A funded account balance at Yellow for the asset you settle in. Funding is a one-time on-chain step; each skill checks balance first and never funds accounts itself.

## Build against

- **Docs:** [docs.yellow.org/nitrolite/builder-toolkit](https://docs.yellow.org/nitrolite/builder-toolkit)
- **Quickstart:** [docs.yellow.org/nitrolite/build/getting-started/quickstart](https://docs.yellow.org/nitrolite/build/getting-started/quickstart)
- **Runnable example:** [`examples/nitrolite-v1-lifecycle`](https://github.com/layer-3/docs/tree/main/examples/nitrolite-v1-lifecycle)
- **SDK:** `@yellow-org/sdk`

## License

MIT
