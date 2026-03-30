#!/usr/bin/env node
/**
 * Nitrolite SDK MCP Server
 *
 * Exposes the Nitrolite SDK API surface to AI agents and IDEs via the
 * Model Context Protocol. Reads SDK source at startup to build structured
 * knowledge of methods, types, enums, and examples.
 */

import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';
import { readFileSync, existsSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const SDK_ROOT = resolve(__dirname, '../../ts');
const COMPAT_ROOT = resolve(__dirname, '../../ts-compat');

// ---------------------------------------------------------------------------
// Helpers — read SDK sources at startup
// ---------------------------------------------------------------------------

function readFile(path: string): string {
    if (!existsSync(path)) return '';
    return readFileSync(path, 'utf-8');
}

/** Extract named exports from a barrel file */
function extractExports(content: string): string[] {
    const names: string[] = [];
    // Match: export { Foo, Bar } from '...'  and  export type { Baz } from '...'
    for (const m of content.matchAll(/export\s+(?:type\s+)?\{([^}]+)\}/g)) {
        for (const name of m[1].split(',')) {
            const clean = name.replace(/\s+as\s+\w+/, '').replace(/type\s+/, '').trim();
            if (clean && !clean.startsWith('//')) names.push(clean);
        }
    }
    // Match: export * from '...'
    for (const m of content.matchAll(/export\s+\*\s+from\s+'([^']+)'/g)) {
        names.push(`* from '${m[1]}'`);
    }
    return names;
}

// ---------------------------------------------------------------------------
// SDK Data — populated at startup
// ---------------------------------------------------------------------------

interface MethodInfo {
    name: string;
    signature: string;
    description: string;
    category: string;
}

interface TypeInfo {
    name: string;
    kind: 'interface' | 'type' | 'enum' | 'class';
    fields: string;
    source: string;
}

const methods: MethodInfo[] = [];
const types: TypeInfo[] = [];
const compatExports: string[] = [];

function loadClientMethods(): void {
    const content = readFile(resolve(SDK_ROOT, 'src/client.ts'));
    const re = /\/\*\*\s*([\s\S]*?)\*\/\s*(?:async\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*Promise<([^>]+)>|\s*:\s*(\S+))?/g;
    let match;
    while ((match = re.exec(content)) !== null) {
        const doc = match[1].replace(/\s*\*\s*/g, ' ').trim();
        const name = match[2];
        const params = match[3].trim();
        const returnType = match[4] || match[5] || 'void';

        // Skip private/internal
        if (name.startsWith('_') || content.substring(Math.max(0, match.index - 20), match.index).includes('private')) continue;

        const category = categorizeMethod(name);
        const isAsync = content.substring(Math.max(0, match.index - 10), match.index).includes('async') ||
            content.substring(match.index, match.index + match[0].length).includes('async');
        const returnStr = isAsync ? `Promise<${returnType}>` : returnType;

        methods.push({
            name,
            signature: `${name}(${params}): ${returnStr}`,
            description: doc || `SDK method: ${name}`,
            category,
        });
    }
}

function categorizeMethod(name: string): string {
    if (/channel/i.test(name)) return 'Channels';
    if (/deposit|withdraw|transfer|escrow|approve/i.test(name)) return 'Transactions';
    if (/app.*session|submitApp/i.test(name)) return 'App Sessions';
    if (/sign|signer|key/i.test(name)) return 'Signing';
    if (/ping|config|asset|balance|blockchain/i.test(name)) return 'Node & Queries';
    return 'Other';
}

function loadTypes(): void {
    const files = [
        { path: resolve(SDK_ROOT, 'src/core/types.ts'), source: 'sdk/ts (core)' },
        { path: resolve(SDK_ROOT, 'src/rpc/types.ts'), source: 'sdk/ts (rpc)' },
        { path: resolve(COMPAT_ROOT, 'src/types.ts'), source: 'sdk-compat' },
    ];

    for (const { path, source } of files) {
        const content = readFile(path);
        if (!content) continue;

        // Enums
        for (const m of content.matchAll(/export\s+enum\s+(\w+)\s*\{([^}]+)\}/g)) {
            types.push({ name: m[1], kind: 'enum', fields: m[2].trim(), source });
        }
        // Interfaces
        for (const m of content.matchAll(/export\s+interface\s+(\w+)(?:\s+extends\s+\w+)?\s*\{([^}]+)\}/g)) {
            types.push({ name: m[1], kind: 'interface', fields: m[2].trim(), source });
        }
        // Type aliases
        for (const m of content.matchAll(/export\s+type\s+(\w+)\s*=\s*([^;]+);/g)) {
            types.push({ name: m[1], kind: 'type', fields: m[2].trim(), source });
        }
    }
}

function loadCompatExports(): void {
    const content = readFile(resolve(COMPAT_ROOT, 'src/index.ts'));
    compatExports.push(...extractExports(content));
}

// ---------------------------------------------------------------------------
// Initialize
// ---------------------------------------------------------------------------

loadClientMethods();
loadTypes();
loadCompatExports();

// ---------------------------------------------------------------------------
// MCP Server
// ---------------------------------------------------------------------------

const server = new McpServer({
    name: 'nitrolite-sdk',
    version: '0.1.0',
});

// ========================== RESOURCES ======================================

server.resource('api-methods', 'nitrolite://api/methods', async () => {
    const grouped: Record<string, MethodInfo[]> = {};
    for (const m of methods) {
        (grouped[m.category] ??= []).push(m);
    }
    let text = '# Nitrolite SDK — Client Methods\n\n';
    for (const [cat, ms] of Object.entries(grouped).sort()) {
        text += `## ${cat}\n\n`;
        for (const m of ms) {
            text += `### \`${m.signature}\`\n${m.description}\n\n`;
        }
    }
    return { contents: [{ uri: 'nitrolite://api/methods', text, mimeType: 'text/markdown' }] };
});

server.resource('api-types', 'nitrolite://api/types', async () => {
    const interfaces = types.filter(t => t.kind === 'interface');
    const aliases = types.filter(t => t.kind === 'type');
    let text = '# Nitrolite SDK — Types & Interfaces\n\n';
    text += `## Interfaces (${interfaces.length})\n\n`;
    for (const t of interfaces) {
        text += `### \`${t.name}\` (${t.source})\n\`\`\`typescript\n${t.fields}\n\`\`\`\n\n`;
    }
    text += `## Type Aliases (${aliases.length})\n\n`;
    for (const t of aliases) {
        text += `- \`${t.name}\` = \`${t.fields}\` (${t.source})\n`;
    }
    return { contents: [{ uri: 'nitrolite://api/types', text, mimeType: 'text/markdown' }] };
});

server.resource('api-enums', 'nitrolite://api/enums', async () => {
    const enums = types.filter(t => t.kind === 'enum');
    let text = '# Nitrolite SDK — Enums\n\n';
    for (const e of enums) {
        text += `## \`${e.name}\` (${e.source})\n\`\`\`typescript\n${e.fields}\n\`\`\`\n\n`;
    }
    return { contents: [{ uri: 'nitrolite://api/enums', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-channels', 'nitrolite://examples/channels', async () => {
    const text = `# Nitrolite SDK — Channel Examples

## Creating a Channel & Depositing

\`\`\`typescript
import { Client, DefaultConfig, withBlockchainRPC } from '@yellow-org/sdk';

const client = await Client.create(
    walletClient,
    'wss://clearnode.example.com/ws',
    DefaultConfig,
    withBlockchainRPC(11155111n, 'https://rpc.sepolia.org'),
);

// Deposit creates channel if needed
const state = await client.deposit(11155111n, 'usdc', '10.0');
console.log('New state version:', state.version);
\`\`\`

## Querying Channels

\`\`\`typescript
const channels = await client.getChannels({ wallet: address });
for (const ch of channels) {
    console.log(ch.channelId, ch.status, ch.asset);
}
\`\`\`

## Closing a Channel

\`\`\`typescript
const finalState = await client.closeHomeChannel('usdc');
// On-chain settlement happens automatically
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/channels', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-transfers', 'nitrolite://examples/transfers', async () => {
    const text = `# Nitrolite SDK — Transfer Examples

## Simple Transfer

\`\`\`typescript
const state = await client.transfer(recipientAddress, 'usdc', '5.0');
\`\`\`

## Using the Compat Layer

\`\`\`typescript
import { NitroliteClient, WalletStateSigner } from '@yellow-org/sdk-compat';

const client = await NitroliteClient.create({
    wsURL: 'wss://clearnode.example.com/ws',
    walletClient,
    chainId: 11155111,
    blockchainRPCs: { 11155111: 'https://rpc.sepolia.org' },
});

await client.transfer(recipientAddress, [{ asset: 'usdc', amount: '5.0' }]);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/transfers', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-app-sessions', 'nitrolite://examples/app-sessions', async () => {
    const text = `# Nitrolite SDK — App Session Examples

## Creating an App Session

\`\`\`typescript
const session = await client.createAppSession({
    appId: 'my-app-id',
    participants: [address1, address2],
    allocations: [
        { participant: address1, asset: 'usdc', amount: '10.0' },
        { participant: address2, asset: 'usdc', amount: '10.0' },
    ],
});
\`\`\`

## Submitting App State

\`\`\`typescript
await client.submitAppState({
    sessionId: session.id,
    allocations: updatedAllocations,
    intent: 'operate',
});
\`\`\`

## Closing an App Session

\`\`\`typescript
await client.closeAppSession(session.id);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/app-sessions', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-auth', 'nitrolite://examples/auth', async () => {
    const text = `# Nitrolite SDK — Authentication Examples

## Compat Layer Auth Flow (Legacy v0.5.3 Pattern)

\`\`\`typescript
import {
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createEIP712AuthMessageSigner,
    generateRequestId,
    getCurrentTimestamp,
    parseAuthChallengeResponse,
    parseAuthVerifyResponse,
} from '@yellow-org/sdk-compat';

// 1. Create auth request
const signer = createEIP712AuthMessageSigner(walletClient);
const authMsg = await createAuthRequestMessage(signer, address);
ws.send(authMsg);

// 2. Parse challenge
const challengeRaw = await waitForResponse();
const challenge = parseAuthChallengeResponse(challengeRaw);

// 3. Verify
const verifyMsg = await createAuthVerifyMessage(signer, challenge.params.challengeMessage);
ws.send(verifyMsg);

// 4. Parse verification result
const verifyRaw = await waitForResponse();
const result = parseAuthVerifyResponse(verifyRaw);
console.log('Authenticated:', result.params.success);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/auth', text, mimeType: 'text/markdown' }] };
});

server.resource('migration-overview', 'nitrolite://migration/overview', async () => {
    const content = readFile(resolve(COMPAT_ROOT, 'docs/migration-overview.md'));
    const text = content || '# Migration Overview\n\nNo migration docs found. Check sdk/ts-compat/docs/.';
    return { contents: [{ uri: 'nitrolite://migration/overview', text, mimeType: 'text/markdown' }] };
});

// ========================== TOOLS ==========================================

server.tool(
    'lookup_method',
    'Look up a specific SDK Client method by name — returns signature, params, return type, usage context',
    { name: z.string().describe('Method name (e.g. "transfer", "deposit", "getChannels")') },
    async ({ name }) => {
        const query = name.toLowerCase();
        const matches = methods.filter(m => m.name.toLowerCase().includes(query));
        if (matches.length === 0) {
            return { content: [{ type: 'text' as const, text: `No method matching "${name}" found. Available categories: ${[...new Set(methods.map(m => m.category))].join(', ')}` }] };
        }
        const text = matches.map(m =>
            `## ${m.name}\n**Signature:** \`${m.signature}\`\n**Category:** ${m.category}\n**Description:** ${m.description}`
        ).join('\n\n---\n\n');
        return { content: [{ type: 'text' as const, text }] };
    },
);

server.tool(
    'lookup_type',
    'Look up a type, interface, or enum by name — returns fields and source location',
    { name: z.string().describe('Type name (e.g. "Channel", "State", "RPCMethod")') },
    async ({ name }) => {
        const query = name.toLowerCase();
        const matches = types.filter(t => t.name.toLowerCase().includes(query));
        if (matches.length === 0) {
            return { content: [{ type: 'text' as const, text: `No type matching "${name}" found. ${types.length} types indexed.` }] };
        }
        const text = matches.map(t =>
            `## ${t.name} (${t.kind})\n**Source:** ${t.source}\n\`\`\`typescript\n${t.fields}\n\`\`\``
        ).join('\n\n---\n\n');
        return { content: [{ type: 'text' as const, text }] };
    },
);

server.tool(
    'search_api',
    'Fuzzy search across all SDK methods and types',
    { query: z.string().describe('Search query (e.g. "session key", "balance", "transfer")') },
    async ({ query }) => {
        const q = query.toLowerCase();
        const methodHits = methods.filter(m =>
            m.name.toLowerCase().includes(q) || m.description.toLowerCase().includes(q) || m.category.toLowerCase().includes(q)
        );
        const typeHits = types.filter(t =>
            t.name.toLowerCase().includes(q) || t.fields.toLowerCase().includes(q)
        );

        let text = `# Search results for "${query}"\n\n`;
        if (methodHits.length > 0) {
            text += `## Methods (${methodHits.length} matches)\n`;
            for (const m of methodHits.slice(0, 10)) {
                text += `- \`${m.signature}\` — ${m.category}\n`;
            }
            text += '\n';
        }
        if (typeHits.length > 0) {
            text += `## Types (${typeHits.length} matches)\n`;
            for (const t of typeHits.slice(0, 10)) {
                text += `- \`${t.name}\` (${t.kind}) — ${t.source}\n`;
            }
            text += '\n';
        }
        if (methodHits.length === 0 && typeHits.length === 0) {
            text += 'No matches found. Try a broader term.\n';
        }
        return { content: [{ type: 'text' as const, text }] };
    },
);

server.tool(
    'get_rpc_method',
    'Get the RPC wire format for a given compat-layer method (useful for integration test authors)',
    { method: z.string().describe('RPC method name (e.g. "get_channels", "transfer", "create_app_session")') },
    async ({ method }) => {
        const rpcMethods: Record<string, { wireMethod: string; reqFormat: string; resFormat: string }> = {
            ping: { wireMethod: 'node.v1.ping', reqFormat: '[requestId, "ping", [], timestamp]', resFormat: '{ res: [requestId, "ping", { pong: true }] }' },
            get_channels: { wireMethod: 'channels.v1.get_channels', reqFormat: '[requestId, "get_channels", [address], timestamp]', resFormat: '{ res: [requestId, "get_channels", { channels: ChannelUpdate[] }] }' },
            get_ledger_balances: { wireMethod: 'user.v1.get_balances', reqFormat: '[requestId, "get_ledger_balances", [address], timestamp]', resFormat: '{ res: [requestId, "get_ledger_balances", { balances: RPCBalance[] }] }' },
            transfer: { wireMethod: 'channels.v1.submit_state', reqFormat: '[requestId, "transfer", [recipient, [{asset, amount}]], timestamp]', resFormat: '{ res: [requestId, "transfer", { state }] }' },
            create_channel: { wireMethod: 'channels.v1.request_creation', reqFormat: '[requestId, "create_channel", [token, chainId, amount], timestamp]', resFormat: '{ res: [requestId, "create_channel", { channelId, status }] }' },
            close_channel: { wireMethod: 'channels.v1.submit_state', reqFormat: '[requestId, "close_channel", [channelId], timestamp]', resFormat: '{ res: [requestId, "close_channel", { channelId, status }] }' },
            create_app_session: { wireMethod: 'app_sessions.v1.create_app_session', reqFormat: '[requestId, "create_app_session", [appId, participants, allocations], timestamp]', resFormat: '{ res: [requestId, "create_app_session", { sessionId, status }] }' },
            close_app_session: { wireMethod: 'app_sessions.v1.close_app_session', reqFormat: '[requestId, "close_app_session", [sessionId], timestamp]', resFormat: '{ res: [requestId, "close_app_session", { sessionId, status }] }' },
            submit_app_state: { wireMethod: 'app_sessions.v1.submit_app_state', reqFormat: '[requestId, "submit_app_state", [sessionId, allocations, intent], timestamp]', resFormat: '{ res: [requestId, "submit_app_state", { accepted: boolean }] }' },
            get_app_sessions: { wireMethod: 'app_sessions.v1.get_app_sessions', reqFormat: '[requestId, "get_app_sessions", [filters], timestamp]', resFormat: '{ res: [requestId, "get_app_sessions", { sessions: AppSession[] }] }' },
            get_app_definition: { wireMethod: 'app_sessions.v1.get_app_definition', reqFormat: '[requestId, "get_app_definition", [appSessionId], timestamp]', resFormat: '{ res: [requestId, "get_app_definition", { definition }] }' },
            auth_request: { wireMethod: 'auth_request', reqFormat: '[requestId, "auth_request", [address], timestamp]', resFormat: '{ res: [requestId, "auth_challenge", { challengeMessage }] }' },
            auth_verify: { wireMethod: 'auth_verify', reqFormat: '[requestId, "auth_verify", [signature], timestamp]', resFormat: '{ res: [requestId, "auth_verify", { success, sessionKey, address }] }' },
            get_ledger_transactions: { wireMethod: 'user.v1.get_transactions', reqFormat: '[requestId, "get_ledger_transactions", [filters], timestamp]', resFormat: '{ res: [requestId, "get_ledger_transactions", { transactions: RPCTransaction[] }] }' },
            resize_channel: { wireMethod: 'channels.v1.submit_state', reqFormat: '[requestId, "resize_channel", [channelId, amount], timestamp]', resFormat: '{ res: [requestId, "resize_channel", { channelId, status }] }' },
        };

        const key = method.toLowerCase();
        const info = rpcMethods[key];
        if (!info) {
            return { content: [{ type: 'text' as const, text: `Unknown RPC method "${method}". Available: ${Object.keys(rpcMethods).join(', ')}` }] };
        }
        const text = `## RPC: ${method}\n\n**V1 Wire Method:** \`${info.wireMethod}\`\n\n**Request format (v0.5.3 compat):**\n\`\`\`\n${info.reqFormat}\n\`\`\`\n\n**Response format:**\n\`\`\`\n${info.resFormat}\n\`\`\``;
        return { content: [{ type: 'text' as const, text }] };
    },
);

server.tool(
    'validate_import',
    'Check if a symbol is exported from sdk-compat barrel — returns yes/no + correct import path',
    { symbol: z.string().describe('Symbol name (e.g. "NitroliteClient", "RPCMethod", "createTransferMessage")') },
    async ({ symbol }) => {
        const found = compatExports.some(e => e === symbol || e.includes(symbol));
        if (found) {
            return { content: [{ type: 'text' as const, text: `**${symbol}** is exported from \`@yellow-org/sdk-compat\`.\n\n\`\`\`typescript\nimport { ${symbol} } from '@yellow-org/sdk-compat';\n\`\`\`` }] };
        }

        // Check if it's in the main SDK
        const sdkBarrel = readFile(resolve(SDK_ROOT, 'src/index.ts'));
        const inSdk = sdkBarrel.includes(symbol);
        if (inSdk) {
            return { content: [{ type: 'text' as const, text: `**${symbol}** is NOT in \`@yellow-org/sdk-compat\` but IS in \`@yellow-org/sdk\`.\n\n\`\`\`typescript\nimport { ${symbol} } from '@yellow-org/sdk';\n\`\`\`\n\n> Note: SDK classes should not be re-exported from compat (SSR risk). Import directly from \`@yellow-org/sdk\`.` }] };
        }

        return { content: [{ type: 'text' as const, text: `**${symbol}** was not found in either \`@yellow-org/sdk-compat\` or \`@yellow-org/sdk\` barrel exports. It may be a deep import or may not exist.` }] };
    },
);

// ========================== PROMPTS ========================================

server.prompt(
    'create-channel-app',
    'Step-by-step guide to build an app using Nitrolite state channels',
    async () => ({
        messages: [{
            role: 'user' as const,
            content: {
                type: 'text' as const,
                text: `Guide me through building a Nitrolite state channel application. Cover:

1. **Setup** — Install dependencies (@yellow-org/sdk, viem), create Client with config
2. **Authentication** — Connect wallet, establish WebSocket, authenticate with clearnode
3. **Channel Lifecycle** — Deposit (auto-creates channel), query channels, close channel
4. **Transfers** — Send tokens to another participant via state channels
5. **App Sessions** — Create sessions for multi-party apps, submit state, close
6. **Error Handling** — Common errors and how to handle them
7. **Testing** — How to write tests against the SDK

For each step, show complete TypeScript code examples using the latest SDK API.
Use \`@yellow-org/sdk\` for new projects. Only use \`@yellow-org/sdk-compat\` if migrating from v0.5.3.`,
            },
        }],
    }),
);

server.prompt(
    'migrate-from-v053',
    'Interactive migration assistant from @layer-3/nitrolite v0.5.3 to the compat layer',
    async () => {
        const migrationDocs = readFile(resolve(COMPAT_ROOT, 'docs/migration-overview.md'));
        return {
            messages: [{
                role: 'user' as const,
                content: {
                    type: 'text' as const,
                    text: `I need to migrate my app from \`@layer-3/nitrolite\` v0.5.3 to the new SDK. Help me step by step.

Here is the official migration guide:

${migrationDocs}

Walk me through:
1. Installing \`@yellow-org/sdk-compat\` and peer deps
2. Swapping imports (package name change)
3. Replacing create-sign-send-parse pattern with NitroliteClient methods
4. Updating type references if any changed
5. Testing the migrated code

Ask me to paste my current code so you can provide specific migration instructions.`,
                },
            }],
        };
    },
);

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

async function main() {
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error('Nitrolite SDK MCP server running on stdio');
}

main().catch((err) => {
    console.error('Fatal:', err);
    process.exit(1);
});
