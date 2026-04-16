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
const REPO_ROOT = resolve(__dirname, '../../..');
const PROTOCOL_DOCS = resolve(REPO_ROOT, 'docs/protocol');
const API_YAML = resolve(REPO_ROOT, 'docs/api.yaml');
const GO_SDK_ROOT = resolve(REPO_ROOT, 'sdk/go');
const PKG_ROOT = resolve(REPO_ROOT, 'pkg');
const GO_MODULE_PATH = 'github.com/layer-3/nitrolite';
const GO_MODULE_VERSION = 'v1.2.0';

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
    // Match: export { Foo, type Bar } from '...' and export type { Baz } from '...'
    for (const m of content.matchAll(/export\s+(type\s+)?\{([^}]+)\}/g)) {
        const groupIsTypeOnly = Boolean(m[1]);
        for (const item of m[2].split(',')) {
            const raw = item.trim();
            if (!raw || raw.startsWith('//')) continue;

            const itemIsTypeOnly = groupIsTypeOnly || raw.startsWith('type ');
            const spec = raw.replace(/^type\s+/, '').trim();
            const aliasMatch = spec.match(/^(\w+)(?:\s+as\s+(\w+))?$/);
            if (!aliasMatch) continue;

            const exportedName = aliasMatch[2] || aliasMatch[1];
            names.push(itemIsTypeOnly ? `type ${exportedName}` : exportedName);
        }
    }
    // Match: export * from '...'
    for (const m of content.matchAll(/export\s+\*\s+from\s+'([^']+)'/g)) {
        names.push(`* from '${m[1]}'`);
    }
    return names;
}

function findNamedExport(exports: string[], symbol: string): { found: boolean; isTypeOnly: boolean } {
    if (exports.includes(`type ${symbol}`)) return { found: true, isTypeOnly: true };
    if (exports.includes(symbol)) return { found: true, isTypeOnly: false };
    return { found: false, isTypeOnly: false };
}

function renderImportStatement(pkg: string, symbol: string, isTypeOnly: boolean): string {
    return isTypeOnly
        ? `import type { ${symbol} } from '${pkg}';`
        : `import { ${symbol} } from '${pkg}';`;
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

// Go SDK data — populated at startup
interface GoTypeInfo {
    name: string;
    kind: 'struct' | 'enum' | 'type';
    fields: string;
    source: string;
}

interface GoMethodInfo {
    name: string;
    signature: string;
    comment: string;
    category: string;
}

const goTypes: GoTypeInfo[] = [];
const goMethods: GoMethodInfo[] = [];

// Protocol docs loaded at startup
const protocolDocs: Record<string, string> = {};

// Terminology concept → definition map
const concepts: Map<string, string> = new Map();

// V1 RPC method → { description, request fields, response fields } map (from docs/api.yaml)
interface RPCMethodDoc {
    /** Fully qualified v1 method name, e.g. "channels.v1.get_home_channel" */
    method: string;
    /** API group, e.g. "channels" */
    group: string;
    description: string;
    requestFields: string;
    responseFields: string;
}
const rpcMethodDocs: Map<string, RPCMethodDoc> = new Map();

function loadClientMethods(): void {
    const content = readFile(resolve(SDK_ROOT, 'src/client.ts'));
    const re = /\/\*\*\s*([\s\S]*?)\*\/\s*(?:(public|protected|private)\s+)?(?:static\s+)?(?:async\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*Promise<([^>]+)>|\s*:\s*(\S+))?/g;
    let match;
    while ((match = re.exec(content)) !== null) {
        const doc = match[1].replace(/\s*\*\s*/g, ' ').trim();
        const visibility = match[2] ?? 'public';
        const name = match[3];
        const params = match[4].trim();
        const returnType = match[5] || match[6] || 'void';

        // Index only public client methods.
        if (visibility !== 'public' || name.startsWith('_')) continue;

        const category = categorizeMethod(name);
        const isAsync = /\basync\b/.test(match[0]);
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

function categorizeGoMethod(name: string): string {
    const lower = name.toLowerCase();
    if (/channel|deposit|withdraw|transfer|checkpoint|challenge|acknowledge|close/.test(lower)) return 'Channels & Transactions';
    if (/appsession|appstate|appdef|rebalance/.test(lower)) return 'App Sessions';
    if (/sessionkey|keystate/.test(lower)) return 'Session Keys';
    if (/escrow|security|locked/.test(lower)) return 'Security Tokens';
    if (/app/.test(lower) && /register/.test(lower)) return 'App Registry';
    if (/balance|transaction|allowance|user/.test(lower)) return 'User Queries';
    if (/config|blockchain|asset|ping/.test(lower)) return 'Node & Config';
    return 'Other';
}

function loadGoTypes(): void {
    const fileSets = [
        { path: resolve(PKG_ROOT, 'core/types.go'), source: 'pkg/core' },
        { path: resolve(PKG_ROOT, 'app/app_session_v1.go'), source: 'pkg/app' },
        { path: resolve(PKG_ROOT, 'rpc/types.go'), source: 'pkg/rpc' },
        { path: resolve(GO_SDK_ROOT, 'config.go'), source: 'sdk/go' },
        { path: resolve(GO_SDK_ROOT, 'app_session.go'), source: 'sdk/go' },
        { path: resolve(GO_SDK_ROOT, 'app_registry.go'), source: 'sdk/go' },
        { path: resolve(GO_SDK_ROOT, 'user.go'), source: 'sdk/go' },
        { path: resolve(GO_SDK_ROOT, 'channel.go'), source: 'sdk/go' },
    ];

    for (const { path, source } of fileSets) {
        const content = readFile(path);
        if (!content) continue;

        // Pass 1: collect enum-like type declarations (name → kind)
        const enumTypeMap = new Map<string, GoTypeInfo>();
        for (const m of content.matchAll(/type\s+([A-Z]\w+)\s+(uint\d+|int\d+|string)\s*\n/g)) {
            const info: GoTypeInfo = { name: m[1], kind: 'enum', fields: '', source };
            enumTypeMap.set(m[1], info);
        }

        // Pass 2: parse const(...) and var(...) blocks to collect enum values
        for (const blockMatch of content.matchAll(/(?:const|var)\s*\(([^)]+)\)/gs)) {
            const block = blockMatch[1];
            let currentType: string | undefined;

            for (const rawLine of block.split('\n')) {
                const line = rawLine.replace(/\/\/.*$/, '').trim();
                if (!line) continue; // blank lines don't reset currentType

                // Fully-annotated: ExportedName TypeName = ...  or  ExportedName TypeName
                const fullAnnotated = line.match(/^([A-Z]\w+)\s+([A-Z]\w+)\s*(?:=.*)?$/);
                if (fullAnnotated) {
                    const [, valueName, typeName] = fullAnnotated;
                    if (enumTypeMap.has(typeName)) {
                        currentType = typeName;
                        const info = enumTypeMap.get(typeName)!;
                        info.fields += (info.fields ? '\n' : '') + valueName;
                        continue;
                    }
                }

                // Untyped-assignment: ExportedName = value (no type annotation)
                const untypedAssign = line.match(/^([A-Z]\w+)\s*=\s*(.+)$/);
                if (untypedAssign) {
                    const [, valueName] = untypedAssign;
                    let resolvedType = currentType;
                    if (!resolvedType) {
                        // Prefix inference: find enum type whose name is a prefix of valueName
                        for (const typeName of enumTypeMap.keys()) {
                            if (valueName.startsWith(typeName) && valueName[typeName.length]?.match(/[A-Z_]/)) {
                                resolvedType = typeName;
                                break;
                            }
                        }
                    }
                    if (resolvedType && enumTypeMap.has(resolvedType)) {
                        currentType = resolvedType;
                        const info = enumTypeMap.get(resolvedType)!;
                        info.fields += (info.fields ? '\n' : '') + valueName;
                    }
                    continue;
                }

                // Bare identifier: ExportedName (iota follow-on)
                const bareIdent = line.match(/^([A-Z]\w+)\s*$/);
                if (bareIdent && currentType && enumTypeMap.has(currentType)) {
                    const info = enumTypeMap.get(currentType)!;
                    info.fields += (info.fields ? '\n' : '') + bareIdent[1];
                    continue;
                }

                // Non-exported or unrecognised → reset
                if (!line.match(/^[A-Z]/)) currentType = undefined;
            }
        }
        for (const info of enumTypeMap.values()) {
            if (info.fields) goTypes.push(info);
        }

        // Structs
        for (const m of content.matchAll(/(?:\/\/[^\n]*\n)*type\s+([A-Z]\w+)\s+struct\s*\{([^}]*)\}/gs)) {
            goTypes.push({ name: m[1], kind: 'struct', fields: m[2].trim(), source });
        }

        // Functional option types (e.g. type Option func(*Config))
        for (const m of content.matchAll(/type\s+([A-Z]\w+)\s+(func\([^)]*\)[^\n]*)/g)) {
            goTypes.push({ name: m[1], kind: 'type', fields: m[2].trim(), source });
        }
    }
}

function loadGoSdkMethods(): void {
    // Prepend NewClient constructor (no receiver)
    goMethods.push({
        name: 'NewClient',
        signature: 'func NewClient(wsURL string, stateSigner core.ChannelSigner, rawSigner sign.Signer, opts ...Option) (*Client, error)',
        comment: 'Creates a new Nitrolite SDK client connected to a clearnode',
        category: 'Connection',
    });

    const files = [
        { path: resolve(GO_SDK_ROOT, 'channel.go'), category: '' },
        { path: resolve(GO_SDK_ROOT, 'node.go'), category: 'Node & Config' },
        { path: resolve(GO_SDK_ROOT, 'user.go'), category: 'User Queries' },
        { path: resolve(GO_SDK_ROOT, 'app_session.go'), category: 'App Sessions' },
        { path: resolve(GO_SDK_ROOT, 'app_registry.go'), category: 'App Registry' },
        { path: resolve(GO_SDK_ROOT, 'client.go'), category: '' },
    ];

    const methodRe = /((?:\/\/[^\n]*\n)+)func \(c \*Client\) (\w+)\(([^)]*)\)\s*(.*)/g;

    for (const { path, category } of files) {
        const content = readFile(path);
        if (!content) continue;
        methodRe.lastIndex = 0;
        let match;
        while ((match = methodRe.exec(content)) !== null) {
            const rawComment = match[1].trim();
            const comment = rawComment.split('\n')
                .map(l => l.replace(/^\/\/ ?/, '').trim())
                .filter(Boolean)
                .join(' ');
            const name = match[2];
            const params = match[3];
            const returns = match[4].trim();
            if (name[0] >= 'a' && name[0] <= 'z') continue; // skip unexported
            const cat = category || categorizeGoMethod(name);
            goMethods.push({
                name,
                signature: `func (c *Client) ${name}(${params}) ${returns}`,
                comment,
                category: cat,
            });
        }
    }
}

function buildGoApiMethodsContent(): string {
    const ORDER = ['Connection', 'Channels & Transactions', 'App Sessions', 'Session Keys', 'Security Tokens', 'App Registry', 'User Queries', 'Node & Config', 'Other'];
    const grouped = new Map<string, GoMethodInfo[]>();
    for (const m of goMethods) {
        const arr = grouped.get(m.category) ?? [];
        arr.push(m);
        grouped.set(m.category, arr);
    }
    let text = '# Nitrolite Go SDK — Client Methods\n\nPackage: `github.com/layer-3/nitrolite/sdk/go`\n\n';
    for (const cat of ORDER) {
        const ms = grouped.get(cat);
        if (!ms?.length) continue;
        text += `## ${cat}\n\n`;
        for (const m of ms) {
            text += `### \`${m.name}\`\n\`\`\`go\n${m.signature}\n\`\`\`\n${m.comment}\n\n`;
        }
    }
    return text;
}

function buildGoTypesContent(): string {
    const bySource = new Map<string, GoTypeInfo[]>();
    for (const t of goTypes) {
        const arr = bySource.get(t.source) ?? [];
        arr.push(t);
        bySource.set(t.source, arr);
    }
    let text = '# Nitrolite Go SDK — Types\n\n';
    for (const [src, ts] of [...bySource.entries()].sort()) {
        text += `## ${src}\n\n`;
        for (const t of ts) {
            if (t.kind === 'struct') {
                text += `### \`${t.name}\` (struct)\n\`\`\`go\ntype ${t.name} struct {\n${t.fields}\n}\n\`\`\`\n\n`;
            } else if (t.kind === 'enum') {
                text += `### \`${t.name}\` (enum)\n**Values:**\n${t.fields.split('\n').map(v => `- \`${v}\``).join('\n')}\n\n`;
            } else {
                text += `### \`${t.name}\` (${t.kind})\n\`\`\`go\ntype ${t.name} ${t.fields}\n\`\`\`\n\n`;
            }
        }
    }
    return text;
}

function loadTypes(): void {
    const files = [
        { path: resolve(SDK_ROOT, 'src/core/types.ts'), source: 'sdk/ts (core)' },
        { path: resolve(SDK_ROOT, 'src/rpc/types.ts'), source: 'sdk/ts (rpc)' },
        { path: resolve(SDK_ROOT, 'src/app/types.ts'), source: 'sdk/ts (app)' },
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

function loadProtocolDocs(): void {
    const files = [
        'overview.md', 'terminology.md', 'cryptography.md', 'state-model.md',
        'channel-protocol.md', 'enforcement.md', 'cross-chain-and-assets.md',
        'interactions.md', 'security-and-limitations.md',
    ];
    for (const file of files) {
        const key = file.replace('.md', '');
        const content = readFile(resolve(PROTOCOL_DOCS, file));
        if (content) protocolDocs[key] = content;
    }
}

function loadTerminology(): void {
    const content = protocolDocs['terminology'] || '';
    // Parse ### headings and their following paragraphs
    const sections = content.split(/^### /m).slice(1);
    for (const section of sections) {
        const lines = section.trim().split('\n');
        const name = lines[0].trim();
        const body = lines.slice(1).join('\n').trim()
            .replace(/^[\s\n]+/, '')
            .split(/\n(?=### |## )/)[0]
            .trim();
        if (name && body) {
            concepts.set(name.toLowerCase(), `**${name}**\n\n${body}`);
        }
    }
}

function loadV1API(): void {
    const content = readFile(API_YAML);
    if (!content) return;

    // Simple line-based parser for the well-structured api.yaml
    // Extracts: group name, method name, description, request fields, response fields
    let currentGroup = '';
    let currentMethod = '';
    let currentDesc = '';
    let currentSection: 'none' | 'request' | 'response' = 'none';
    let requestFields: string[] = [];
    let responseFields: string[] = [];

    const flushMethod = () => {
        if (currentGroup && currentMethod) {
            const fqn = `${currentGroup}.v1.${currentMethod}`;
            rpcMethodDocs.set(fqn, {
                method: fqn,
                group: currentGroup,
                description: currentDesc,
                requestFields: requestFields.length > 0 ? requestFields.join(', ') : '(none)',
                responseFields: responseFields.length > 0 ? responseFields.join(', ') : '(none)',
            });
        }
        currentMethod = '';
        currentDesc = '';
        currentSection = 'none';
        requestFields = [];
        responseFields = [];
    };

    // Only parse the api: section
    const apiStart = content.indexOf('\napi:\n');
    if (apiStart === -1) return;
    const lines = content.slice(apiStart).split('\n');

    for (const line of lines) {
        // Group: "    - name: channels"
        const groupMatch = line.match(/^ {4}- name:\s+(.+)/);
        if (groupMatch) {
            flushMethod();
            currentGroup = groupMatch[1].trim();
            continue;
        }

        // Method: "            - name: get_home_channel"
        const methodMatch = line.match(/^ {12}- name:\s+(.+)/);
        if (methodMatch) {
            flushMethod();
            currentMethod = methodMatch[1].trim();
            continue;
        }

        // Method description: "              description: ..."
        if (currentMethod && !currentDesc) {
            const descMatch = line.match(/^ {14}description:\s+(.+)/);
            if (descMatch) {
                currentDesc = descMatch[1].trim();
                continue;
            }
        }

        // Request/response section markers
        if (currentMethod) {
            const sectionMatch = line.match(/^ {14}(request|response):/);
            if (sectionMatch) {
                currentSection = sectionMatch[1] as 'request' | 'response';
                continue;
            }

            // Field name within request/response: "                - field_name: wallet"
            const fieldMatch = line.match(/^ {16}- field_name:\s+(.+)/);
            if (fieldMatch) {
                if (currentSection === 'request') requestFields.push(fieldMatch[1].trim());
                else if (currentSection === 'response') responseFields.push(fieldMatch[1].trim());
                continue;
            }

            // errors: or events: sections end request/response
            if (/^ {14}(errors|events):/.test(line)) {
                currentSection = 'none';
            }
        }
    }
    flushMethod(); // flush last method
}

// ---------------------------------------------------------------------------
// Go SDK static content constants ported from the former Go MCP server.
// ---------------------------------------------------------------------------

const AUTH_FLOW_CONTENT = `# Request Signing & Authorization

In v1, every RPC request includes a \`sig\` field — the client's signature over the entire \`req\` tuple. This is the authorization mechanism. There is no separate authentication handshake; request signatures are the identity proof.

## Session Keys

Session keys enable delegated signing with scoped permissions. They are managed via:
- \`channels.v1.submit_session_key_state\` — register/update channel session keys
- \`app_sessions.v1.submit_session_key_state\` — register/update app session keys

Session keys have:
- Per-asset allowances with spending caps
- Expiration timestamps
- Scoping to specific applications and app sessions

## Wire Format

\`\`\`json
// Every request is signed
{ "req": [REQUEST_ID, "channels.v1.submit_state", { ... }, TIMESTAMP], "sig": ["0xClientSignature..."] }

// Server responds with its own signature
{ "res": [REQUEST_ID, "channels.v1.submit_state", { ... }, TIMESTAMP], "sig": ["0xServerSignature..."] }
\`\`\`

## Note on 0.5.x Compat

The \`@yellow-org/sdk-compat\` layer exposes legacy auth helpers (\`createAuthRequestMessage\`, \`createAuthVerifyMessage\`) that implement a challenge-response flow with JWT. This is the 0.5.x auth surface bridged to v1. New applications using \`@yellow-org/sdk\` directly do not use this flow.
`;

const GO_TRANSFER_EXAMPLE = `# Complete Go Transfer Script

\`\`\`go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    recipient := os.Getenv("RECIPIENT")
    var chainID uint64 = 11155111 // Sepolia

    stateSigner, err := sign.NewEthereumMsgSigner(privateKey)
    if err != nil { log.Fatal(err) }
    txSigner, err := sign.NewEthereumRawSigner(privateKey)
    if err != nil { log.Fatal(err) }

    client, err := sdk.NewClient(clearnodeURL, stateSigner, txSigner,
        sdk.WithBlockchainRPC(chainID, rpcURL),
    )
    if err != nil { log.Fatal(err) }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Approve token spending (one-time)
    _, err = client.ApproveToken(ctx, chainID, "usdc", decimal.NewFromInt(1000))
    if err != nil { log.Fatal(err) }

    // Deposit — creates channel if needed
    state, err := client.Deposit(ctx, chainID, "usdc", decimal.NewFromInt(10))
    if err != nil { log.Fatal(err) }
    log.Printf("Deposited 10 USDC, state version: %d", state.Version)

    // Checkpoint on-chain
    txHash, err := client.Checkpoint(ctx, "usdc")
    if err != nil { log.Fatal(err) }
    log.Printf("On-chain tx: %s", txHash)

    // Transfer
    _, err = client.Transfer(ctx, recipient, "usdc", decimal.NewFromInt(5))
    if err != nil { log.Fatal(err) }
    log.Println("Transferred 5 USDC")

    // Close channel — prepare finalize state, then checkpoint
    _, err = client.CloseHomeChannel(ctx, "usdc")
    if err != nil { log.Fatal(err) }
    closeTx, err := client.Checkpoint(ctx, "usdc")
    if err != nil { log.Fatal(err) }
    log.Printf("Channel closed, tx: %s", closeTx)
}
\`\`\`

## Environment Variables

- \`PRIVATE_KEY\` — Hex private key (without 0x prefix)
- \`CLEARNODE_URL\` — WebSocket URL
- \`RPC_URL\` — Ethereum RPC endpoint
- \`RECIPIENT\` — Recipient address
`;

const GO_APP_SESSION_EXAMPLE = `# Complete Go App Session Script

\`\`\`go
package main

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/layer-3/nitrolite/pkg/app"
    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    peerAddr := os.Getenv("PEER_ADDRESS")
    var chainID uint64 = 11155111

    stateSigner, err := sign.NewEthereumMsgSigner(privateKey)
    if err != nil { log.Fatal(err) }
    txSigner, err := sign.NewEthereumRawSigner(privateKey)
    if err != nil { log.Fatal(err) }

    client, err := sdk.NewClient(clearnodeURL, stateSigner, txSigner,
        sdk.WithBlockchainRPC(chainID, rpcURL),
    )
    if err != nil { log.Fatal(err) }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    myAddr := client.GetUserAddress()

    // Fund the home channel before moving funds into the app session
    if _, err = client.Deposit(ctx, chainID, "usdc", decimal.NewFromInt(20)); err != nil { log.Fatal(err) }
    txHash, err := client.Checkpoint(ctx, "usdc")
    if err != nil { log.Fatal(err) }
    log.Printf("Home channel funded with 20 USDC, tx: %s", txHash)

    // 1. Create app session
    definition := app.AppDefinitionV1{
        ApplicationID: "my-game",
        Participants: []app.AppParticipantV1{
            {WalletAddress: myAddr, SignatureWeight: 50},
            {WalletAddress: peerAddr, SignatureWeight: 50},
        },
        Quorum: 100,
        Nonce:  1,
    }

    quorumSigs := []string{"0xMySignature...", "0xPeerSignature..."}
    sessionID, versionStr, status, err := client.CreateAppSession(ctx, definition, "{}", quorumSigs)
    if err != nil { log.Fatal(err) }
    log.Printf("Session %s created (version: %s, status: %s)", sessionID, versionStr, status)

    initVersion, err := strconv.ParseUint(versionStr, 10, 64)
    if err != nil { log.Fatal(err) }

    // 2. Fund the app session before submitting non-zero allocations
    depositUpdate := app.AppStateUpdateV1{
        AppSessionID: sessionID,
        Intent:       app.AppStateUpdateIntentDeposit,
        Version:      initVersion + 1,
        Allocations: []app.AppAllocationV1{
            {Participant: myAddr, Asset: "usdc", Amount: decimal.NewFromInt(15)},
            {Participant: peerAddr, Asset: "usdc", Amount: decimal.NewFromInt(5)},
        },
        SessionData:  "{}",
    }
    if _, err = client.SubmitAppSessionDeposit(ctx, depositUpdate, quorumSigs, "usdc", decimal.NewFromInt(20)); err != nil { log.Fatal(err) }
    log.Println("Session funded with 20 USDC")

    // 3. Submit state update (version = initial + 2)
    update := app.AppStateUpdateV1{
        AppSessionID: sessionID,
        Intent:       app.AppStateUpdateIntentOperate,
        Version:      initVersion + 2,
        Allocations:  depositUpdate.Allocations,
        SessionData:  "{\"round\":1,\"winner\":\"me\"}",
    }
    operateSigs := []string{"0xMySig...", "0xPeerSig..."}
    if err = client.SubmitAppState(ctx, update, operateSigs); err != nil { log.Fatal(err) }
    log.Println("State updated")

    // 4. Close session — submit with Close intent (version = initial + 3)
    closeUpdate := update
    closeUpdate.Intent = app.AppStateUpdateIntentClose
    closeUpdate.Version = initVersion + 3
    closeSigs := []string{"0xMyCloseSig...", "0xPeerCloseSig..."}
    if err = client.SubmitAppState(ctx, closeUpdate, closeSigs); err != nil { log.Fatal(err) }
    log.Println("Session closed")
}
\`\`\`
`;

const GO_SCAFFOLD_TRANSFER = `package main

import (
\t"context"
\t"log"
\t"os"
\t"time"

\t"github.com/layer-3/nitrolite/pkg/sign"
\tsdk "github.com/layer-3/nitrolite/sdk/go"
\t"github.com/shopspring/decimal"
)

func main() {
\tstateSigner, err := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }
\ttxSigner, err := sign.NewEthereumRawSigner(os.Getenv("PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }

\tclient, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
\t\tsdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
\t)
\tif err != nil { log.Fatal(err) }
\tdefer client.Close()

\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
\tdefer cancel()

\t_, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(10))
\tif err != nil { log.Fatal(err) }
\ttxHash, err := client.Checkpoint(ctx, "usdc")
\tif err != nil { log.Fatal(err) }
\tlog.Printf("Deposited 10 USDC, tx: %s", txHash)

\t_, err = client.Transfer(ctx, os.Getenv("RECIPIENT"), "usdc", decimal.NewFromInt(5))
\tif err != nil { log.Fatal(err) }
\tlog.Println("Transferred 5 USDC")
}
`;

const GO_SCAFFOLD_APP_SESSION = `package main

import (
\t"context"
\t"log"
\t"os"
\t"strconv"
\t"time"

\t"github.com/layer-3/nitrolite/pkg/app"
\t"github.com/layer-3/nitrolite/pkg/sign"
\tsdk "github.com/layer-3/nitrolite/sdk/go"
\t"github.com/shopspring/decimal"
)

func main() {
\tstateSigner, err := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }
\ttxSigner, err := sign.NewEthereumRawSigner(os.Getenv("PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }

\tclient, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
\t\tsdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
\t)
\tif err != nil { log.Fatal(err) }
\tdefer client.Close()

\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
\tdefer cancel()

\tmyAddr := client.GetUserAddress()
\tpeer := os.Getenv("PEER_ADDRESS")

\t// Fund the home channel before moving funds into the app session
\tif _, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(20)); err != nil { log.Fatal(err) }
\ttxHash, err := client.Checkpoint(ctx, "usdc")
\tif err != nil { log.Fatal(err) }
\tlog.Printf("Home channel funded with 20 USDC, tx: %s", txHash)

\tdef := app.AppDefinitionV1{
\t\tApplicationID: "my-app",
\t\tParticipants: []app.AppParticipantV1{
\t\t\t{WalletAddress: myAddr, SignatureWeight: 50},
\t\t\t{WalletAddress: peer, SignatureWeight: 50},
\t\t},
\t\tQuorum: 100,
\t\tNonce:  1,
\t}

\tquorumSigs := []string{"0xMySig...", "0xPeerSig..."}
\tsessionID, versionStr, _, err := client.CreateAppSession(ctx, def, "{}", quorumSigs)
\tif err != nil { log.Fatal(err) }
\tlog.Printf("Session created: %s", sessionID)

\tinitVersion, err := strconv.ParseUint(versionStr, 10, 64)
\tif err != nil { log.Fatal(err) }

\tdepositUpdate := app.AppStateUpdateV1{
\t\tAppSessionID: sessionID,
\t\tIntent:       app.AppStateUpdateIntentDeposit,
\t\tVersion:      initVersion + 1,
\t\tAllocations: []app.AppAllocationV1{
\t\t\t{Participant: myAddr, Asset: "usdc", Amount: decimal.NewFromInt(12)},
\t\t\t{Participant: peer, Asset: "usdc", Amount: decimal.NewFromInt(8)},
\t\t},
\t\tSessionData: "{}",
\t}
\toperateSigs := []string{"0xMySig...", "0xPeerSig..."}
\tif _, err := client.SubmitAppSessionDeposit(ctx, depositUpdate, operateSigs, "usdc", decimal.NewFromInt(20)); err != nil { log.Fatal(err) }
\tlog.Println("Session funded with 20 USDC")

\tupdate := app.AppStateUpdateV1{
\t\tAppSessionID: sessionID,
\t\tIntent:       app.AppStateUpdateIntentOperate,
\t\tVersion:      initVersion + 2,
\t\tAllocations:  depositUpdate.Allocations,
\t\tSessionData:  "{\"round\":1,\"winner\":\"me\"}",
\t}
\tif err := client.SubmitAppState(ctx, update, operateSigs); err != nil { log.Fatal(err) }
\tlog.Println("State updated")

\tupdate.Intent = app.AppStateUpdateIntentClose
\tupdate.Version = initVersion + 3
\tcloseSigs := []string{"0xMyCloseSig...", "0xPeerCloseSig..."}
\tif err := client.SubmitAppState(ctx, update, closeSigs); err != nil { log.Fatal(err) }
\tlog.Println("Session closed")
}
`;

const GO_SCAFFOLD_AI_AGENT = `package main

import (
\t"context"
\t"log"
\t"os"
\t"os/signal"
\t"syscall"
\t"time"

\t"github.com/layer-3/nitrolite/pkg/sign"
\tsdk "github.com/layer-3/nitrolite/sdk/go"
\t"github.com/shopspring/decimal"
)

func main() {
\tstateSigner, err := sign.NewEthereumMsgSigner(os.Getenv("AGENT_PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }
\ttxSigner, err := sign.NewEthereumRawSigner(os.Getenv("AGENT_PRIVATE_KEY"))
\tif err != nil { log.Fatal(err) }

\tclient, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
\t\tsdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
\t)
\tif err != nil { log.Fatal(err) }

\tsigCh := make(chan os.Signal, 1)
\tsignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
\tgo func() {
\t\t<-sigCh
\t\tlog.Println("Shutting down agent...")
\t\tclient.Close()
\t}()

\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
\tdefer cancel()

\t_, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(50))
\tif err != nil { log.Fatal(err) }
\ttxHash, err := client.Checkpoint(ctx, "usdc")
\tif err != nil { log.Fatal(err) }
\tlog.Printf("Agent funded with 50 USDC, tx: %s", txHash)

\trecipients := []string{"0x1111...", "0x2222..."}
\tfor _, r := range recipients {
\t\t_, err := client.Transfer(ctx, r, "usdc", decimal.NewFromFloat(0.10))
\t\tif err != nil {
\t\t\tlog.Printf("Payment to %s failed: %v", r, err)
\t\t\tcontinue
\t\t}
\t\tlog.Printf("Paid 0.10 USDC to %s", r)
\t}

\tlog.Println("Agent payments complete")
\t<-client.WaitCh()
}
`;

// ---------------------------------------------------------------------------
// Initialize
// ---------------------------------------------------------------------------

loadClientMethods();
loadTypes();
loadCompatExports();
loadProtocolDocs();
loadTerminology();
loadV1API();
loadGoTypes();
loadGoSdkMethods();

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
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const { stateSigner, txSigner } = createSigners('0xYourPrivateKey...');
const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, 'https://rpc.amoy.polygon.technology'),
);

// Deposit creates channel if needed
const state = await client.deposit(80002n, 'usdc', new Decimal(10));
const txHash = await client.checkpoint('usdc');
console.log('Deposit on-chain tx:', txHash);
\`\`\`

## Querying Channels

\`\`\`typescript
const userAddress = client.getUserAddress();
const { channels } = await client.getChannels(userAddress);
for (const ch of channels) {
    console.log(ch.channelId, ch.status);
}
\`\`\`

## Closing a Channel

\`\`\`typescript
// closeHomeChannel prepares the finalize state — checkpoint submits it on-chain
const finalState = await client.closeHomeChannel('usdc');
const closeTx = await client.checkpoint('usdc');
console.log('Channel closed, tx:', closeTx);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/channels', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-transfers', 'nitrolite://examples/transfers', async () => {
    const text = `# Nitrolite SDK — Transfer Examples

## Simple Transfer

\`\`\`typescript
import Decimal from 'decimal.js';

const state = await client.transfer('0xRecipient...', 'usdc', new Decimal('5.0'));
console.log('Transfer tx ID:', state.transition.txId);
\`\`\`

## Using the Compat Layer

\`\`\`typescript
import { NitroliteClient } from '@yellow-org/sdk-compat';

const client = await NitroliteClient.create({
    wsURL: 'wss://clearnode.example.com/ws',
    walletClient,
    chainId: 11155111,
    blockchainRPCs: { 11155111: 'https://rpc.sepolia.org' },
});

await client.transfer('0xRecipient...', [{ asset: 'usdc', amount: '5.0' }]);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/transfers', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-app-sessions', 'nitrolite://examples/app-sessions', async () => {
    const text = `# Nitrolite SDK — App Session Examples

## Creating an App Session

\`\`\`typescript
import { app } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// Assumes client and CHAIN_ID are already defined.
await client.deposit(CHAIN_ID, 'usdc', new Decimal(20));
await client.checkpoint('usdc');

// 1. Define the app session
const definition: app.AppDefinitionV1 = {
    applicationId: 'my-game-app',
    participants: [
        { walletAddress: '0xAlice...', signatureWeight: 50 },
        { walletAddress: '0xBob...', signatureWeight: 50 },
    ],
    quorum: 100, // Both must agree
    nonce: BigInt(Date.now()),
};

// 2. Collect quorum signatures from participants (off-band)
const quorumSigs: string[] = ['0xAliceSig...', '0xBobSig...'];

// 3. Create the session
const result = await client.createAppSession(definition, '{}', quorumSigs);
console.log('Created session:', result.appSessionId);

// 4. Fund the session before submitting non-zero allocations
const depositUpdate: app.AppStateUpdateV1 = {
    appSessionId: result.appSessionId,
    intent: app.AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
        { participant: '0xAlice...', asset: 'usdc', amount: new Decimal(15) },
        { participant: '0xBob...', asset: 'usdc', amount: new Decimal(5) },
    ],
    sessionData: '{}',
};
await client.submitAppSessionDeposit(depositUpdate, ['0xAliceSig...', '0xBobSig...'], 'usdc', new Decimal(20));
\`\`\`

## Submitting App State

\`\`\`typescript
import Decimal from 'decimal.js';

const appUpdate: app.AppStateUpdateV1 = {
    appSessionId: result.appSessionId,
    intent: app.AppStateUpdateIntent.Operate,
    version: 3n,
    allocations: [
        { participant: '0xAlice...', asset: 'usdc', amount: new Decimal(15) },
        { participant: '0xBob...', asset: 'usdc', amount: new Decimal(5) },
    ],
    sessionData: '{"move": "e4"}',
};
await client.submitAppState(appUpdate, ['0xAliceSig...', '0xBobSig...']);
\`\`\`

> **Note:** There is no \`closeAppSession()\` on the SDK Client. To close a session,
> submit a state update with \`intent: app.AppStateUpdateIntent.Close\`.
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
    parseAnyRPCResponse,
    type AuthRequestParams,
} from '@yellow-org/sdk-compat';

// 1. Create auth request
const authParams: AuthRequestParams = {
    address: account.address,
    session_key: '0x0000000000000000000000000000000000000000',
    application: 'My App',
    expires_at: BigInt(Math.floor(Date.now() / 1000) + 3600),
    scope: 'app.create',
    allowances: [{ asset: 'usdc', amount: '100.0' }],
};
const authMsg = await createAuthRequestMessage(authParams);
ws.send(authMsg);

// 2. Parse challenge
const challengeRaw = await waitForResponse();
const challenge = parseAnyRPCResponse(challengeRaw);

// 3. Create EIP-712 signer and verify
const signer = createEIP712AuthMessageSigner(
    walletClient,
    { scope: authParams.scope, session_key: authParams.session_key as \\\`0x\${string}\\\`, expires_at: authParams.expires_at, allowances: authParams.allowances },
    { name: 'Yellow Network' },
);
const verifyMsg = await createAuthVerifyMessage(signer, challenge);
ws.send(verifyMsg);

// 4. Parse verification result
const verifyRaw = await waitForResponse();
const result = parseAnyRPCResponse(verifyRaw);
console.log('Authenticated:', result);
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/auth', text, mimeType: 'text/markdown' }] };
});

server.resource('migration-overview', 'nitrolite://migration/overview', async () => {
    const content = readFile(resolve(COMPAT_ROOT, 'docs/migration-overview.md'));
    const text = content || '# Migration Overview\n\nNo migration docs found. Check sdk/ts-compat/docs/.';
    return { contents: [{ uri: 'nitrolite://migration/overview', text, mimeType: 'text/markdown' }] };
});

// ========================== PROTOCOL RESOURCES ==============================

server.resource('protocol-overview', 'nitrolite://protocol/overview', async () => {
    const text = protocolDocs['overview'] || '# Protocol Overview\n\nProtocol docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/overview', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-terminology', 'nitrolite://protocol/terminology', async () => {
    const text = protocolDocs['terminology'] || '# Terminology\n\nTerminology docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/terminology', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-wire-format', 'nitrolite://protocol/wire-format', async () => {
    const text = protocolDocs['interactions'] || '# Wire Format\n\nInteraction docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/wire-format', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-rpc-methods', 'nitrolite://protocol/rpc-methods', async () => {
    let text = '# V1 RPC Methods\n\nAll v1 RPC methods defined in `docs/api.yaml`. Methods use grouped naming: `{group}.v1.{method}`.\n\n';

    // Group methods by their API group
    const grouped: Record<string, RPCMethodDoc[]> = {};
    for (const doc of rpcMethodDocs.values()) {
        (grouped[doc.group] ??= []).push(doc);
    }
    for (const [group, docs] of Object.entries(grouped)) {
        text += `## ${group}\n\n`;
        text += '| Method | Description | Request Fields | Response Fields |\n|---|---|---|---|\n';
        for (const doc of docs) {
            text += `| \`${doc.method}\` | ${doc.description} | ${doc.requestFields} | ${doc.responseFields} |\n`;
        }
        text += '\n';
    }

    text += '## Message Format\n\n';
    text += 'All messages use compact ordered arrays:\n\n';
    text += '**Request:** `{ "req": [REQUEST_ID, METHOD, PARAMS, TIMESTAMP], "sig": ["SIGNATURE"] }`\n\n';
    text += '**Response:** `{ "res": [REQUEST_ID, METHOD, DATA, TIMESTAMP], "sig": ["SIGNATURE"] }`\n\n';
    text += '**With App Session:** Add `"sid": "APP_SESSION_ID"` to route messages to app session participants.\n';
    return { contents: [{ uri: 'nitrolite://protocol/rpc-methods', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-cryptography', 'nitrolite://protocol/cryptography', async () => {
    const text = protocolDocs['cryptography'] || '# Cryptography\n\nCryptography docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/cryptography', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-channel-lifecycle', 'nitrolite://protocol/channel-lifecycle', async () => {
    const text = protocolDocs['channel-protocol'] || '# Channel Protocol\n\nChannel protocol docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/channel-lifecycle', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-state-model', 'nitrolite://protocol/state-model', async () => {
    const text = protocolDocs['state-model'] || '# State Model\n\nState model docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/state-model', text, mimeType: 'text/markdown' }] };
});

// ========================== SECURITY RESOURCES ==============================

server.resource('security-overview', 'nitrolite://security/overview', async () => {
    const text = protocolDocs['security-and-limitations'] || '# Security\n\nSecurity docs not found.';
    return { contents: [{ uri: 'nitrolite://security/overview', text, mimeType: 'text/markdown' }] };
});

server.resource('security-app-session-patterns', 'nitrolite://security/app-session-patterns', async () => {
    const text = `# App Session Security Patterns

Best practices for building secure, decentralization-ready app sessions on Nitrolite.

## Quorum Design

App sessions use a weight-based quorum system for governance:

\`\`\`typescript
interface AppDefinitionV1 {
  applicationId: string;
  participants: AppParticipantV1[];  // each has walletAddress + signatureWeight
  quorum: number;                     // minimum total weight to authorize actions (uint8)
  nonce: bigint;
}
\`\`\`

### Recommended Patterns

**Equal 2-of-2 (peer-to-peer):** Both participants must agree.
\`\`\`json
{ "participants": [{ "signatureWeight": 50 }, { "signatureWeight": 50 }], "quorum": 100 }
\`\`\`

**3-of-3 (multi-party unanimous):** All three must agree.
\`\`\`json
{ "participants": [{ "signatureWeight": 34 }, { "signatureWeight": 33 }, { "signatureWeight": 33 }], "quorum": 100 }
\`\`\`

**2-of-3 with arbitrator:** Any two can authorize (third party can break ties).
\`\`\`json
{ "participants": [{ "signatureWeight": 50 }, { "signatureWeight": 50 }, { "signatureWeight": 50 }], "quorum": 100 }
\`\`\`

**Weighted (operator-controlled):** One party has majority weight.
\`\`\`json
{ "participants": [{ "signatureWeight": 70 }, { "signatureWeight": 30 }], "quorum": 70 }
\`\`\`

## Challenge Periods

The challenge duration defines how long a dispute can be contested on-chain:
- **Short (1 hour):** For low-value, high-frequency operations
- **Medium (24 hours):** Recommended default for most applications
- **Long (7 days):** For high-value operations requiring more security

## State Invariants

Developers MUST ensure these invariants hold in every state update:
1. **Fund conservation:** Total allocations across participants MUST equal the committed amount
2. **Version ordering:** Each state version MUST be exactly previous + 1
3. **Signature requirements:** Updates require signatures meeting the quorum threshold
4. **Non-negative allocations:** No participant's allocation can go below zero

## Decentralization-Ready Patterns

Even if not fully decentralized today, build app sessions so they would work in a decentralized environment:

1. **Never trust a single party** — Use quorum >= total weight of any single participant
2. **Use challenge periods** — They exist to protect against malicious state submissions
3. **Keep state deterministic** — All participants should be able to independently verify state transitions
4. **Support unilateral enforcement** — Any participant should be able to enforce the latest agreed state on-chain
5. **Separate signing from logic** — Use session keys with spending caps rather than raw wallet keys

## On-Chain Enforcement

If off-chain cooperation fails, any participant can:
1. Submit the latest mutually signed state to the blockchain
2. The blockchain validates signatures, version ordering, and ledger invariants
3. After the challenge period, the state becomes final
4. Assets are distributed according to the final allocations
`;
    return { contents: [{ uri: 'nitrolite://security/app-session-patterns', text, mimeType: 'text/markdown' }] };
});

server.resource('security-state-invariants', 'nitrolite://security/state-invariants', async () => {
    const text = `# State Invariants

Critical invariants that MUST hold across all state transitions. Violating these will cause on-chain enforcement to fail.

## Ledger Invariant (Fund Conservation)

\`\`\`
UserAllocation + NodeAllocation == UserNetFlow + NodeNetFlow
\`\`\`

This ensures no assets can be created or destroyed through state transitions. The total distributable balance always equals the total cumulative flows.

## Allocation Non-Negativity

All allocation values (UserAllocation, NodeAllocation) MUST be non-negative. Net flow values MAY be negative (outbound transfers exceeding inbound).

## Version Ordering

- **Off-chain:** Each new version MUST equal previous version + 1
- **On-chain enforcement:** Submitted version MUST be strictly greater than the current on-chain version

## Signature Requirements

- **Mutually signed states** (both user + node signatures) are the only states enforceable on-chain
- **Node-issued pending states** (node signature only) are NOT enforceable — they become enforceable after user acknowledgement
- Signature validation modes MUST be among the channel's approved validators

## Channel Binding

The channel identifier in every state MUST match the channel definition. This is verified both off-chain and on-chain.

## Locked Funds

Unless the channel is being closed:
\`\`\`
UserAllocation + NodeAllocation == LockedFunds
\`\`\`

Locked funds MUST never be negative. The node MUST have sufficient vault funds when required to lock additional assets.

## Empty Non-Home Ledger

For non-cross-chain operations, the non-home ledger MUST be fully zeroed (all fields set to 0/zero-address).
`;
    return { contents: [{ uri: 'nitrolite://security/state-invariants', text, mimeType: 'text/markdown' }] };
});

// ========================== USE CASES RESOURCES =============================

server.resource('use-cases', 'nitrolite://use-cases', async () => {
    const text = `# Nitrolite Use Cases

What you can build with the Nitrolite SDK and state channels.

## Peer-to-Peer Payments
Instant, gas-free token transfers between users. Open a channel, transfer any amount instantly, settle on-chain only when needed.
**SDK methods:** \`client.deposit()\`, \`client.transfer()\`, \`client.closeHomeChannel()\`

## Gaming (Real-Time Wagering)
Turn-based or real-time games where players wager tokens. App sessions track game state; winners receive payouts automatically.
**SDK methods:** \`client.createAppSession()\`, \`client.submitAppSessionDeposit()\`, \`client.submitAppState()\` (close via \`submitAppState\` with Close intent)
**Example:** Yetris — a Tetris-style game with token wagering built on app sessions.

## Multi-Party Checkout / Escrow
Multiple parties contribute to a shared pool (e.g., group payment, crowdfunding). Funds release when quorum conditions are met.
**SDK methods:** \`client.createAppSession()\`, \`client.submitAppSessionDeposit()\`, custom quorum weights, close via \`client.submitAppState()\` with Close intent
**Example:** Cosign Demo — a multi-party co-signing checkout flow.

## AI Agent Payments
Autonomous AI agents making and receiving payments via state channels. Agents manage their own wallets, open channels, and transact programmatically.
**SDK methods:** \`Client.create()\`, \`client.deposit()\`, \`client.transfer()\`
**See also:** \`nitrolite://use-cases/ai-agents\`

## DeFi Escrow & Atomic Swaps
Trustless exchange of assets between parties using escrow transitions. Cross-chain support via the unified asset model.
**SDK methods:** Escrow transitions via \`client.submitAppSessionDeposit()\`

## Streaming Payments
Continuous micro-transfers over time (e.g., pay-per-second for compute, bandwidth, or content). State channels make sub-cent payments feasible.
**SDK methods:** \`client.transfer()\` in a loop with small amounts

## Cross-Chain Operations
Move assets between blockchains through the escrow mechanism. Deposit on chain A, use on chain B, withdraw on chain C.
**SDK methods:** Cross-chain escrow transitions, \`client.deposit()\` on any supported chain
`;
    return { contents: [{ uri: 'nitrolite://use-cases', text, mimeType: 'text/markdown' }] };
});

server.resource('use-cases-ai-agents', 'nitrolite://use-cases/ai-agents', async () => {
    const text = `# AI Agent Use Cases

How to use Nitrolite for AI agent payments and agent-to-agent interactions.

## Why State Channels for AI Agents?

AI agents need to make frequent, small payments — often thousands per session. On-chain transactions are too slow and expensive. State channels provide:
- **Instant finality** — no waiting for block confirmations
- **Near-zero cost** — gas only on channel open/close, not per-transfer
- **Programmable** — agents manage channels autonomously via the SDK

## Agent Wallet Setup

\`\`\`typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';

// Agent has its own private key
const { stateSigner, txSigner } = createSigners(AGENT_PRIVATE_KEY);

const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(chainId, RPC_URL),
);
\`\`\`

## Agent-to-Agent Payments

Two AI agents can transact directly through state channels:
1. Both agents open channels with the same clearnode
2. Agent A calls \`client.transfer(agentB_address, 'usdc', new Decimal('0.01'))\`
3. Agent B receives the transfer instantly
4. No on-chain transactions needed

## Session Key Delegation

For security, agents can use delegated session keys with spending caps:
- The agent's main wallet authorizes a session key during authentication
- The session key has a maximum spending allowance (e.g., 100 USDC)
- Once the cap is reached, the session key is revoked
- The main wallet funds are never at risk beyond the allowance

## Autonomous Escrow

AI agents can participate in app sessions for complex multi-step workflows:
1. Agent creates an app session with another agent or user
2. Both commit funds to the session
3. The application logic determines final allocations
4. The session closes and funds are distributed

## Integration with Agent Frameworks

The SDK works with any agent framework (LangChain, AutoGPT, CrewAI, etc.):
- Wrap SDK methods as agent tools
- Let the agent decide when to make payments
- Use session keys for safe autonomous operation

## yao.com Proxy Pattern

For agents that need a unified interface, yao.com provides a proxy layer:
- Agents connect to yao.com instead of directly to a clearnode
- yao.com handles channel management and routing
- Agents focus on their application logic
`;
    return { contents: [{ uri: 'nitrolite://use-cases/ai-agents', text, mimeType: 'text/markdown' }] };
});

// ========================== FULL EXAMPLE RESOURCES ===========================

server.resource('examples-full-transfer', 'nitrolite://examples/full-transfer-script', async () => {
    const text = `# Complete Transfer Script

A fully working TypeScript script that connects to a clearnode, opens a channel, deposits funds, transfers tokens, and closes the channel.

\`\`\`typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// --- Configuration ---
const PRIVATE_KEY = process.env.PRIVATE_KEY as \`0x\${string}\`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL || 'https://rpc.sepolia.org';
const CHAIN_ID = 80002n; // Polygon Amoy
const RECIPIENT = process.env.RECIPIENT as \`0x\${string}\`;

async function main() {
    // 1. Create signers from private key
    const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);

    // 2. Create SDK client — connects WebSocket + authenticates
    const client = await Client.create(
        CLEARNODE_URL,
        stateSigner,
        txSigner,
        withBlockchainRPC(CHAIN_ID, RPC_URL),
    );
    console.log('Connected and authenticated');

    // 3. Approve token spending (one-time per token, or when increasing allowance)
    await client.approveToken(CHAIN_ID, 'usdc', new Decimal(1000));
    console.log('Token approved');

    // 4. Deposit — creates channel if needed, then checkpoint on-chain
    const depositState = await client.deposit(CHAIN_ID, 'usdc', new Decimal(10));
    const depositTx = await client.checkpoint('usdc');
    console.log('Deposited 10 USDC, tx:', depositTx);

    // 5. Transfer to recipient
    const transferState = await client.transfer(RECIPIENT, 'usdc', new Decimal(5));
    console.log('Transferred 5 USDC, state version:', transferState.version);

    // 6. Check balances
    const userAddress = client.getUserAddress();
    const balances = await client.getBalances(userAddress);
    console.log('Current balances:', balances);

    // 7. Close channel — two steps: prepare finalize state, then checkpoint on-chain
    const finalState = await client.closeHomeChannel('usdc');
    const closeTx = await client.checkpoint('usdc');
    console.log('Channel closed, tx:', closeTx);
}

main().catch(console.error);
\`\`\`

## Environment Variables

- \`PRIVATE_KEY\` — Your wallet private key (hex with 0x prefix)
- \`CLEARNODE_URL\` — WebSocket URL of the clearnode
- \`RPC_URL\` — Ethereum RPC endpoint for the target chain
- \`RECIPIENT\` — Address to transfer tokens to

## Dependencies

\`\`\`json
{
  "@yellow-org/sdk": "^1.2.0",
  "decimal.js": "^10.4.0",
  "viem": "^2.46.0"
}
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/full-transfer-script', text, mimeType: 'text/markdown' }] };
});

server.resource('examples-full-app-session', 'nitrolite://examples/full-app-session-script', async () => {
    const text = `# Complete App Session Script

A fully working TypeScript script that creates a multi-party app session, submits state updates, and closes with final allocations.

\`\`\`typescript
import { Client, createSigners, withBlockchainRPC, app } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// --- Configuration ---
const PRIVATE_KEY = process.env.PRIVATE_KEY as \`0x\${string}\`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL || 'https://rpc.sepolia.org';
const CHAIN_ID = 80002n;
const PEER_ADDRESS = process.env.PEER_ADDRESS as \`0x\${string}\`;

async function main() {
    const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);
    const myAddress = stateSigner.address;

    const client = await Client.create(
        CLEARNODE_URL,
        stateSigner,
        txSigner,
        withBlockchainRPC(CHAIN_ID, RPC_URL),
    );
    console.log('Connected');

    // Ensure funds are available in the home channel before app-session funding
    await client.deposit(CHAIN_ID, 'usdc', new Decimal(20));
    await client.checkpoint('usdc');

    // 1. Define app session
    const definition: app.AppDefinitionV1 = {
        applicationId: 'my-game-app',
        participants: [
            { walletAddress: myAddress, signatureWeight: 50 },
            { walletAddress: PEER_ADDRESS, signatureWeight: 50 },
        ],
        quorum: 100, // Both must agree
        nonce: BigInt(Date.now()),
    };

    // 2. Collect quorum signatures from participants (off-band signing)
    const quorumSigs: string[] = ['0xMySignature...', '0xPeerSignature...'];

    // 3. Create app session
    const session = await client.createAppSession(definition, '{}', quorumSigs);
    console.log('App session created:', session.appSessionId);

    // 4. Fund the app session before submitting non-zero allocations
    const depositUpdate: app.AppStateUpdateV1 = {
        appSessionId: session.appSessionId,
        intent: app.AppStateUpdateIntent.Deposit,
        version: 2n,
        allocations: [
            { participant: myAddress, asset: 'usdc', amount: new Decimal(15) },
            { participant: PEER_ADDRESS, asset: 'usdc', amount: new Decimal(5) },
        ],
        sessionData: '{}',
    };
    await client.submitAppSessionDeposit(depositUpdate, ['0xMySig...', '0xPeerSig...'], 'usdc', new Decimal(20));
    console.log('Session funded with 20 USDC');

    // 5. Submit state update — e.g., after a game round
    const appUpdate: app.AppStateUpdateV1 = {
        appSessionId: session.appSessionId,
        intent: app.AppStateUpdateIntent.Operate,
        version: 3n,
        allocations: depositUpdate.allocations,
        sessionData: '{"round": 1, "winner": "me"}',
    };
    await client.submitAppState(appUpdate, ['0xMySig...', '0xPeerSig...']);
    console.log('State updated: I won 5 USDC');

    // 6. Close app session — submit final state with Close intent
    const closeUpdate: app.AppStateUpdateV1 = {
        ...appUpdate,
        intent: app.AppStateUpdateIntent.Close,
        version: 4n,
    };
    await client.submitAppState(closeUpdate, ['0xMyCloseSig...', '0xPeerCloseSig...']);
    console.log('Session closed, funds returned to channels');
}

main().catch(console.error);
\`\`\`

## Key Concepts

- **Quorum:** Set to 100 with equal weights (50/50) — both parties must sign every state update
- **Allocations:** Must always sum to the total committed amount (fund conservation invariant)
- **Intent:** Use \`Operate\` for normal updates, \`Close\` for final settlement (there is no separate \`closeAppSession()\` method)
- **Session data:** Optional string field for app-specific metadata (game state, etc.)
- **Quorum sigs:** Participants sign the app state off-band; signatures are collected and submitted together

## Dependencies

\`\`\`json
{
  "@yellow-org/sdk": "^1.2.0",
  "viem": "^2.46.0",
  "decimal.js": "^10.6.0"
}
\`\`\`
`;
    return { contents: [{ uri: 'nitrolite://examples/full-app-session-script', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-enforcement', 'nitrolite://protocol/enforcement', async () => {
    const text = protocolDocs['enforcement'] || '# Enforcement\n\nEnforcement docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/enforcement', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-cross-chain', 'nitrolite://protocol/cross-chain', async () => {
    const text = protocolDocs['cross-chain-and-assets'] || '# Cross-Chain & Assets\n\nCross-chain docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/cross-chain', text, mimeType: 'text/markdown' }] };
});

server.resource('protocol-interactions', 'nitrolite://protocol/interactions', async () => {
    const text = protocolDocs['interactions'] || '# Interactions\n\nInteractions docs not found.';
    return { contents: [{ uri: 'nitrolite://protocol/interactions', text, mimeType: 'text/markdown' }] };
});

// ========================== GO SDK RESOURCES ================================

server.resource('go-api-methods', 'nitrolite://go-api/methods', async () => ({
    contents: [{ uri: 'nitrolite://go-api/methods', text: buildGoApiMethodsContent(), mimeType: 'text/markdown' }],
}));

server.resource('go-api-types', 'nitrolite://go-api/types', async () => ({
    contents: [{ uri: 'nitrolite://go-api/types', text: buildGoTypesContent(), mimeType: 'text/markdown' }],
}));

server.resource('go-examples-full-transfer', 'nitrolite://go-examples/full-transfer-script', async () => ({
    contents: [{ uri: 'nitrolite://go-examples/full-transfer-script', text: GO_TRANSFER_EXAMPLE, mimeType: 'text/markdown' }],
}));

server.resource('go-examples-full-app-session', 'nitrolite://go-examples/full-app-session-script', async () => ({
    contents: [{ uri: 'nitrolite://go-examples/full-app-session-script', text: GO_APP_SESSION_EXAMPLE, mimeType: 'text/markdown' }],
}));

server.resource('protocol-auth-flow', 'nitrolite://protocol/auth-flow', async () => ({
    contents: [{ uri: 'nitrolite://protocol/auth-flow', text: AUTH_FLOW_CONTENT, mimeType: 'text/markdown' }],
}));


// ========================== TOOLS ==========================================

server.tool(
    'lookup_method',
    'Look up a specific SDK Client method by name — returns signature, params, return type, usage context',
    {
        name: z.string().describe('Method name (e.g. "transfer", "deposit", "getChannels", "Transfer", "Deposit")'),
        language: z.enum(['typescript', 'go', 'both']).optional().default('typescript').describe('SDK language to search: "typescript" (default), "go", or "both"'),
    },
    async ({ name, language }) => {
        const query = name.toLowerCase();
        const parts: string[] = [];

        if (language === 'typescript' || language === 'both') {
            const tsMatches = methods.filter(m => m.name.toLowerCase().includes(query));
            if (tsMatches.length > 0) {
                const header = language === 'both' ? '## TypeScript SDK\n\n' : '';
                parts.push(header + tsMatches.map(m =>
                    `### ${m.name}\n**Signature:** \`${m.signature}\`\n**Category:** ${m.category}\n**Description:** ${m.description}`
                ).join('\n\n---\n\n'));
            }
        }

        if (language === 'go' || language === 'both') {
            const goMatches = goMethods.filter(m => m.name.toLowerCase().includes(query));
            if (goMatches.length > 0) {
                const header = language === 'both' ? '## Go SDK\n\n' : '';
                parts.push(header + goMatches.map(m =>
                    `### ${m.name}\n**Signature:**\n\`\`\`go\n${m.signature}\n\`\`\`\n**Category:** ${m.category}\n**Description:** ${m.comment}`
                ).join('\n\n---\n\n'));
            }
        }

        if (parts.length === 0) {
            return { content: [{ type: 'text' as const, text: `No method matching "${name}" found. Available categories: ${[...new Set(methods.map(m => m.category))].join(', ')}` }] };
        }
        return { content: [{ type: 'text' as const, text: parts.join('\n\n') }] };
    },
);

server.tool(
    'lookup_type',
    'Look up a type, interface, or enum by name — returns fields and source location',
    {
        name: z.string().describe('Type name (e.g. "Channel", "State", "RPCMethod", "AppSessionV1", "ChannelStatus")'),
        language: z.enum(['typescript', 'go', 'both']).optional().default('typescript').describe('SDK language to search: "typescript" (default), "go", or "both"'),
    },
    async ({ name, language }) => {
        const query = name.toLowerCase();
        const parts: string[] = [];

        if (language === 'typescript' || language === 'both') {
            const tsMatches = types.filter(t => t.name.toLowerCase().includes(query));
            if (tsMatches.length > 0) {
                const header = language === 'both' ? '## TypeScript SDK\n\n' : '';
                parts.push(header + tsMatches.map(t =>
                    `### ${t.name} (${t.kind})\n**Source:** ${t.source}\n\`\`\`typescript\n${t.fields}\n\`\`\``
                ).join('\n\n---\n\n'));
            }
        }

        if (language === 'go' || language === 'both') {
            const goMatches = goTypes.filter(t => t.name.toLowerCase().includes(query));
            if (goMatches.length > 0) {
                const header = language === 'both' ? '## Go SDK\n\n' : '';
                parts.push(header + goMatches.map(t => {
                    if (t.kind === 'struct') {
                        return `### ${t.name} (struct)\n**Source:** ${t.source}\n\`\`\`go\ntype ${t.name} struct {\n${t.fields}\n}\n\`\`\``;
                    } else if (t.kind === 'enum') {
                        return `### ${t.name} (enum)\n**Source:** ${t.source}\n**Values:**\n${t.fields.split('\n').map(v => `- \`${v}\``).join('\n')}`;
                    }
                    return `### ${t.name} (${t.kind})\n**Source:** ${t.source}\n\`\`\`go\ntype ${t.name} ${t.fields}\n\`\`\``;
                }).join('\n\n---\n\n'));
            }
        }

        if (parts.length === 0) {
            return { content: [{ type: 'text' as const, text: `No type matching "${name}" found. ${types.length} TS types and ${goTypes.length} Go types indexed.` }] };
        }
        return { content: [{ type: 'text' as const, text: parts.join('\n\n') }] };
    },
);

server.tool(
    'search_api',
    'Fuzzy search across all SDK methods and types',
    {
        query: z.string().describe('Search query (e.g. "session key", "balance", "transfer", "AppSession")'),
        language: z.enum(['typescript', 'go', 'both']).optional().default('typescript').describe('SDK language to search: "typescript" (default), "go", or "both"'),
    },
    async ({ query, language }) => {
        const q = query.toLowerCase();
        let text = `# Search results for "${query}"\n\n`;
        let totalHits = 0;

        if (language === 'typescript' || language === 'both') {
            const methodHits = methods.filter(m =>
                m.name.toLowerCase().includes(q) || m.description.toLowerCase().includes(q) || m.category.toLowerCase().includes(q)
            );
            const typeHits = types.filter(t =>
                t.name.toLowerCase().includes(q) || t.fields.toLowerCase().includes(q)
            );
            const prefix = language === 'both' ? 'TypeScript SDK ' : '';
            if (methodHits.length > 0) {
                text += `## ${prefix}Methods (${methodHits.length} matches)\n`;
                for (const m of methodHits.slice(0, 10)) text += `- \`${m.signature}\` — ${m.category}\n`;
                text += '\n';
                totalHits += methodHits.length;
            }
            if (typeHits.length > 0) {
                text += `## ${prefix}Types (${typeHits.length} matches)\n`;
                for (const t of typeHits.slice(0, 10)) text += `- \`${t.name}\` (${t.kind}) — ${t.source}\n`;
                text += '\n';
                totalHits += typeHits.length;
            }
        }

        if (language === 'go' || language === 'both') {
            const goMethodHits = goMethods.filter(m =>
                m.name.toLowerCase().includes(q) || m.comment.toLowerCase().includes(q) || m.category.toLowerCase().includes(q)
            );
            const goTypeHits = goTypes.filter(t =>
                t.name.toLowerCase().includes(q) || t.fields.toLowerCase().includes(q)
            );
            const prefix = language === 'both' ? 'Go SDK ' : '';
            if (goMethodHits.length > 0) {
                text += `## ${prefix}Methods (${goMethodHits.length} matches)\n`;
                for (const m of goMethodHits.slice(0, 10)) text += `- \`${m.name}\` — ${m.category}\n`;
                text += '\n';
                totalHits += goMethodHits.length;
            }
            if (goTypeHits.length > 0) {
                text += `## ${prefix}Types (${goTypeHits.length} matches)\n`;
                for (const t of goTypeHits.slice(0, 10)) text += `- \`${t.name}\` (${t.kind}) — ${t.source}\n`;
                text += '\n';
                totalHits += goTypeHits.length;
            }
        }

        if (totalHits === 0) text += 'No matches found. Try a broader term.\n';
        return { content: [{ type: 'text' as const, text }] };
    },
);

server.tool(
    'get_rpc_method',
    'Get the RPC wire format for a 0.5.x compat-layer method and its v1 equivalent. For v1 method reference, see docs/api.yaml.',
    { method: z.string().describe('0.5.x compat method name (e.g. "get_channels", "transfer", "create_app_session")') },
    async ({ method }) => {
        // NOTE: These are 0.5.x compat-layer method names mapped to their v1 wire equivalents.
        // The v1 API uses grouped methods (e.g. channels.v1.submit_state). The canonical v1
        // reference is docs/api.yaml. This tool exists for sdk-compat integration test authors.
        const rpcMethods: Record<string, { wireMethod: string; reqFormat: string; resFormat: string }> = {
            ping: { wireMethod: 'node.v1.ping', reqFormat: '{ req: [requestId, "ping", {}, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "ping", { pong: true }] }' },
            get_channels: { wireMethod: 'channels.v1.get_channels', reqFormat: '{ req: [requestId, "get_channels", { wallet?, status?, asset?, channel_type?, pagination? }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "get_channels", { channels: [...], metadata: {...} }] }' },
            get_ledger_balances: { wireMethod: 'user.v1.get_balances', reqFormat: '{ req: [requestId, "get_ledger_balances", { wallet }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "get_ledger_balances", { balances: RPCBalance[] }] }' },
            transfer: { wireMethod: 'channels.v1.submit_state', reqFormat: '{ req: [requestId, "transfer", { destination, allocations: [{ asset, amount }] }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "transfer", { state }] }' },
            create_channel: { wireMethod: 'channels.v1.request_creation', reqFormat: '{ req: [requestId, "create_channel", [{ chain_id, token }], timestamp], sig: [...] }', resFormat: '{ res: [requestId, "create_channel", [{ channel_id, channel, state, server_signature }], timestamp], sig: [...] }' },
            close_channel: { wireMethod: 'channels.v1.submit_state', reqFormat: '{ req: [requestId, "close_channel", { channel_id, funds_destination }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "close_channel", { channel_id, state, server_signature }] }' },
            create_app_session: { wireMethod: 'app_sessions.v1.create_app_session', reqFormat: '{ req: [requestId, "create_app_session", { definition, session_data, quorum_sigs, owner_sig? }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "create_app_session", { app_session_id, version, status }] }' },
            submit_app_state: { wireMethod: 'app_sessions.v1.submit_app_state', reqFormat: '{ req: [requestId, "submit_app_state", { app_state_update, quorum_sigs }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "submit_app_state", { accepted: boolean }] }' },
            get_app_sessions: { wireMethod: 'app_sessions.v1.get_app_sessions', reqFormat: '{ req: [requestId, "get_app_sessions", { filters? }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "get_app_sessions", { sessions: AppSession[] }] }' },
            get_app_definition: { wireMethod: 'app_sessions.v1.get_app_definition', reqFormat: '{ req: [requestId, "get_app_definition", { app_session_id }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "get_app_definition", { definition }] }' },
            get_ledger_transactions: { wireMethod: 'user.v1.get_transactions', reqFormat: '{ req: [requestId, "get_ledger_transactions", { wallet, filters? }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "get_ledger_transactions", { transactions: RPCTransaction[] }] }' },
            resize_channel: { wireMethod: 'channels.v1.submit_state', reqFormat: '{ req: [requestId, "resize_channel", { channel_id, resize_amount, allocate_amount, funds_destination }, timestamp], sig: [...] }', resFormat: '{ res: [requestId, "resize_channel", { channel_id, state, server_signature }] }' },
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
        const compatMatch = findNamedExport(compatExports, symbol);
        if (compatMatch.found) {
            return { content: [{ type: 'text' as const, text: `**${symbol}** is exported from \`@yellow-org/sdk-compat\`.\n\n\`\`\`typescript\n${renderImportStatement('@yellow-org/sdk-compat', symbol, compatMatch.isTypeOnly)}\n\`\`\`` }] };
        }

        // Check if it's in the main SDK
        const sdkBarrelContent = readFile(resolve(SDK_ROOT, 'src/index.ts'));
        const sdkExports = extractExports(sdkBarrelContent);
        const sdkMatch = findNamedExport(sdkExports, symbol);
        if (sdkMatch.found) {
            return { content: [{ type: 'text' as const, text: `**${symbol}** is NOT in \`@yellow-org/sdk-compat\` but IS in \`@yellow-org/sdk\`.\n\n\`\`\`typescript\n${renderImportStatement('@yellow-org/sdk', symbol, sdkMatch.isTypeOnly)}\n\`\`\`\n\n> Note: SDK classes should not be re-exported from compat (SSR risk). Import directly from \`@yellow-org/sdk\`.` }] };
        }

        return { content: [{ type: 'text' as const, text: `**${symbol}** was not found in either \`@yellow-org/sdk-compat\` or \`@yellow-org/sdk\` barrel exports. It may be a deep import or may not exist.` }] };
    },
);

server.tool(
    'explain_concept',
    'Plain-English explanation of a Nitrolite protocol concept (e.g. "state channel", "app session", "challenge period")',
    { concept: z.string().describe('Concept name (e.g. "state channel", "app session", "challenge period", "clearnode", "vault")') },
    async ({ concept }) => {
        const query = concept.toLowerCase().trim();

        // Direct match
        const direct = concepts.get(query);
        if (direct) {
            return { content: [{ type: 'text' as const, text: direct }] };
        }

        // Fuzzy match — find concepts that contain the query or vice versa
        const matches: string[] = [];
        for (const [key, value] of concepts) {
            if (key.includes(query) || query.includes(key)) {
                matches.push(value);
            }
        }
        if (matches.length > 0) {
            return { content: [{ type: 'text' as const, text: matches.join('\n\n---\n\n') }] };
        }

        // Word-level fuzzy — match any word
        const words = query.split(/\s+/);
        for (const [key, value] of concepts) {
            if (words.some(w => key.includes(w))) {
                matches.push(value);
            }
        }
        if (matches.length > 0) {
            return { content: [{ type: 'text' as const, text: `No exact match for "${concept}". Related concepts:\n\n${matches.slice(0, 5).join('\n\n---\n\n')}` }] };
        }

        return { content: [{ type: 'text' as const, text: `No concept matching "${concept}" found. ${concepts.size} concepts indexed from protocol terminology. Try broader terms like "channel", "state", "session", "escrow", "transfer".` }] };
    },
);

server.tool(
    'lookup_rpc_method',
    'Look up a v1 RPC method from docs/api.yaml — returns description, request/response fields. Methods use grouped naming: {group}.v1.{method}',
    { method: z.string().describe('V1 RPC method name or search term (e.g. "channels.v1.get_home_channel", "submit_state", "get_balances")') },
    async ({ method }) => {
        const query = method.toLowerCase().trim();

        // Direct match
        const doc = rpcMethodDocs.get(query);
        if (doc) {
            let text = `## V1 RPC Method: \`${doc.method}\`\n\n`;
            text += `**Group:** ${doc.group}\n**Description:** ${doc.description}\n\n`;
            text += `**Request fields:** ${doc.requestFields}\n`;
            text += `**Response fields:** ${doc.responseFields}\n`;
            return { content: [{ type: 'text' as const, text }] };
        }

        // Fuzzy match — search in full method name and short name
        const matches: RPCMethodDoc[] = [];
        for (const [key, val] of rpcMethodDocs) {
            const shortName = key.split('.').pop() || '';
            if (key.includes(query) || query.includes(shortName) || shortName.includes(query)) {
                matches.push(val);
            }
        }
        if (matches.length > 0) {
            const text = matches.map(d =>
                `- \`${d.method}\` — ${d.description}`
            ).join('\n');
            return { content: [{ type: 'text' as const, text: `Matching v1 RPC methods:\n\n${text}` }] };
        }

        return { content: [{ type: 'text' as const, text: `No v1 RPC method matching "${method}". Available methods:\n${[...rpcMethodDocs.keys()].join(', ')}` }] };
    },
);

server.tool(
    'scaffold_project',
    'Generate a starter project structure for a new Nitrolite app — TypeScript or Go templates',
    { template: z.enum(['transfer-app', 'app-session', 'ai-agent', 'go-transfer-app', 'go-app-session', 'go-ai-agent']).describe('Project template: TypeScript (transfer-app, app-session, ai-agent) or Go (go-transfer-app, go-app-session, go-ai-agent)') },
    async ({ template }) => {
        // Go templates — different output shape
        if (template === 'go-transfer-app' || template === 'go-app-session' || template === 'go-ai-agent') {
            const goTemplateMap: Record<string, string> = {
                'go-transfer-app': GO_SCAFFOLD_TRANSFER,
                'go-app-session': GO_SCAFFOLD_APP_SESSION,
                'go-ai-agent': GO_SCAFFOLD_AI_AGENT,
            };
            const baseName = template.replace('go-', '');
            const goMod = `module my-nitrolite-${baseName}\n\ngo 1.25.0\n\nrequire (\n\t${GO_MODULE_PATH} ${GO_MODULE_VERSION}\n\tgithub.com/shopspring/decimal v1.4.0\n)`;
            const envKey = template === 'go-ai-agent' ? 'AGENT_PRIVATE_KEY' : 'PRIVATE_KEY';
            const envExtra = template === 'go-transfer-app' ? '\nRECIPIENT=your_recipient_address' : template === 'go-app-session' ? '\nPEER_ADDRESS=peer_wallet_address' : '';
            const text = `# Scaffold: ${template}\n\n## go.mod\n\`\`\`\n${goMod}\n\`\`\`\n\n## main.go\n\`\`\`go\n${goTemplateMap[template]}\`\`\`\n\n## .env.example\n\`\`\`\n${envKey}=your_hex_key\nCLEARNODE_URL=wss://clearnode.example.com/ws\nRPC_URL=https://rpc.sepolia.org${envExtra}\n\`\`\`\n\n## Setup\n\`\`\`bash\ngo mod tidy\ngo run .\n\`\`\``;
            return { content: [{ type: 'text' as const, text }] };
        }
        const packageJson = {
            name: `nitrolite-${template}`,
            version: '0.1.0',
            private: true,
            type: 'module',
            scripts: { start: 'npx tsx src/index.ts', build: 'tsc', typecheck: 'tsc --noEmit' },
            dependencies: {
                '@yellow-org/sdk': '^1.2.0',
                'decimal.js': '^10.4.0',
                viem: '^2.46.0',
            },
            devDependencies: { typescript: '^5.7.0', tsx: '^4.19.0', '@types/node': '^22.0.0' },
        };

        const tsconfig = {
            compilerOptions: {
                target: 'es2020', module: 'ESNext', moduleResolution: 'bundler',
                strict: true, esModuleInterop: true, outDir: 'dist', declaration: true,
            },
            include: ['src'],
        };

        const templates: Record<string, string> = {
            'transfer-app': `import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const PRIVATE_KEY = process.env.PRIVATE_KEY as \`0x\$\{string}\`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL || 'https://rpc.sepolia.org';
const CHAIN_ID = 80002n;

async function main() {
    const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);

    const client = await Client.create(CLEARNODE_URL, stateSigner, txSigner, withBlockchainRPC(CHAIN_ID, RPC_URL));
    console.log('Connected to clearnode');

    // Deposit funds
    await client.deposit(CHAIN_ID, 'usdc', new Decimal(10));
    await client.checkpoint('usdc');
    console.log('Deposited 10 USDC');

    // Transfer
    const recipient = process.env.RECIPIENT as \`0x\$\{string}\`;
    await client.transfer(recipient, 'usdc', new Decimal(5));
    console.log('Transferred 5 USDC to', recipient);
}

main().catch(console.error);
`,
            'app-session': `import { Client, createSigners, withBlockchainRPC, app } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const PRIVATE_KEY = process.env.PRIVATE_KEY as \`0x\$\{string}\`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL || 'https://rpc.sepolia.org';
const CHAIN_ID = 80002n;
const PEER = process.env.PEER_ADDRESS as \`0x\$\{string}\`;

async function main() {
    const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);
    const myAddress = stateSigner.address;

    const client = await Client.create(CLEARNODE_URL, stateSigner, txSigner, withBlockchainRPC(CHAIN_ID, RPC_URL));

    // Fund the home channel before moving funds into the app session
    await client.deposit(CHAIN_ID, 'usdc', new Decimal(20));
    await client.checkpoint('usdc');

    // Define app session
    const definition: app.AppDefinitionV1 = {
        applicationId: 'my-app',
        participants: [
            { walletAddress: myAddress, signatureWeight: 50 },
            { walletAddress: PEER, signatureWeight: 50 },
        ],
        quorum: 100,
        nonce: BigInt(Date.now()),
    };

    // Collect quorum signatures from participants (off-band)
    const quorumSigs: string[] = ['0xMySig...', '0xPeerSig...'];

    // Create app session
    const session = await client.createAppSession(definition, '{}', quorumSigs);
    console.log('Session created:', session.appSessionId);

    // Fund the app session before non-zero allocations
    const depositUpdate: app.AppStateUpdateV1 = {
        appSessionId: session.appSessionId,
        intent: app.AppStateUpdateIntent.Deposit,
        version: 2n,
        allocations: [
            { participant: myAddress, asset: 'usdc', amount: new Decimal(12) },
            { participant: PEER, asset: 'usdc', amount: new Decimal(8) },
        ],
        sessionData: '{}',
    };
    await client.submitAppSessionDeposit(depositUpdate, ['0xMySig...', '0xPeerSig...'], 'usdc', new Decimal(20));
    console.log('Session funded with 20 USDC');

    // Submit state update
    const update: app.AppStateUpdateV1 = {
        appSessionId: session.appSessionId,
        intent: app.AppStateUpdateIntent.Operate,
        version: 3n,
        allocations: depositUpdate.allocations,
        sessionData: '{}',
    };
    await client.submitAppState(update, ['0xMySig...', '0xPeerSig...']);

    // Close session — submit with Close intent
    const closeUpdate: app.AppStateUpdateV1 = { ...update, intent: app.AppStateUpdateIntent.Close, version: 4n };
    await client.submitAppState(closeUpdate, ['0xMyCloseSig...', '0xPeerCloseSig...']);
    console.log('Session closed');
}

main().catch(console.error);
`,
            'ai-agent': `import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const AGENT_KEY = process.env.AGENT_PRIVATE_KEY as \`0x\$\{string}\`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL || 'https://rpc.sepolia.org';
const CHAIN_ID = 80002n;

async function createAgentClient() {
    const { stateSigner, txSigner } = createSigners(AGENT_KEY);
    return Client.create(CLEARNODE_URL, stateSigner, txSigner, withBlockchainRPC(CHAIN_ID, RPC_URL));
}

async function payForService(client: Awaited<ReturnType<typeof createAgentClient>>, recipient: \`0x\$\{string}\`, amount: Decimal) {
    const state = await client.transfer(recipient, 'usdc', amount);
    console.log(\`Paid \$\{amount} USDC to \$\{recipient}, version: \$\{state.version}\`);
    return state;
}

async function main() {
    const client = await createAgentClient();
    console.log('Agent connected to clearnode');

    // Ensure the agent has funds
    await client.deposit(CHAIN_ID, 'usdc', new Decimal(50));
    await client.checkpoint('usdc');

    // Agent payment loop — pay for each task
    const tasks = [
        { recipient: '0x1111111111111111111111111111111111111111' as \`0x\$\{string}\`, amount: new Decimal('0.10') },
        { recipient: '0x2222222222222222222222222222222222222222' as \`0x\$\{string}\`, amount: new Decimal('0.25') },
    ];

    for (const task of tasks) {
        await payForService(client, task.recipient, task.amount);
    }

    console.log('All payments complete');
}

main().catch(console.error);
`,
        };

        const envExample = `${template === 'ai-agent' ? 'AGENT_PRIVATE_KEY' : 'PRIVATE_KEY'}=0x...
CLEARNODE_URL=wss://clearnode.example.com/ws
RPC_URL=https://rpc.sepolia.org
${template === 'transfer-app' ? 'RECIPIENT=0x...' : ''}${template === 'app-session' ? 'PEER_ADDRESS=0x...' : ''}`;

        const text = `# Scaffold: ${template}

## package.json
\`\`\`json
${JSON.stringify(packageJson, null, 2)}
\`\`\`

## tsconfig.json
\`\`\`json
${JSON.stringify(tsconfig, null, 2)}
\`\`\`

## src/index.ts
\`\`\`typescript
${templates[template]}
\`\`\`

## .env.example
\`\`\`
${envExample}
\`\`\`

## Setup
\`\`\`bash
npm install
cp .env.example .env  # Fill in your values
npx tsx src/index.ts
\`\`\``;

        return { content: [{ type: 'text' as const, text }] };
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
5. **App Sessions** — Fund the home channel with deposit + checkpoint, then create sessions, fund them with submitAppSessionDeposit, submit state, close
6. **Error Handling** — Common errors and how to handle them
7. **Testing** — How to write tests against the SDK

For each step, show complete TypeScript code examples using the latest SDK API.
Use \`@yellow-org/sdk\` for new projects. Only use \`@yellow-org/sdk-compat\` if migrating from v0.5.3.

## Go SDK

Guide me through building a Nitrolite state channel application in Go. Cover:

1. **Setup** — Install dependencies, create signers with sign.NewEthereumMsgSigner and sign.NewEthereumRawSigner
2. **Client Creation** — sdk.NewClient with functional options (WithBlockchainRPC, WithHandshakeTimeout)
3. **Channel Lifecycle** — Deposit (creates channel), Transfer, Checkpoint (on-chain), CloseHomeChannel + Checkpoint
4. **App Sessions** — Deposit + Checkpoint on the home channel, then CreateAppSession, SubmitAppSessionDeposit, SubmitAppState (Operate/Withdraw/Close intents)
5. **Error Handling** — Go error patterns, context.WithTimeout, defer client.Close()
6. **Testing** — Standard Go test patterns with *_test.go files

For each step, show complete Go code examples using the latest SDK from github.com/layer-3/nitrolite/sdk/go.
Use github.com/shopspring/decimal for amounts. Use context.Context for all async operations.`,
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

server.prompt(
    'build-ai-agent-app',
    'Guided conversation for building an AI agent that uses Nitrolite for payments',
    async () => ({
        messages: [{
            role: 'user' as const,
            content: {
                type: 'text' as const,
                text: `I want to build an AI agent that uses Nitrolite state channels for payments. Guide me through:

1. **Agent Wallet Setup** — Create a wallet for the agent, configure the SDK client
2. **Channel Management** — Open a channel, deposit funds for the agent to use
3. **Automated Payments** — Implement a payment function the agent can call autonomously
4. **Session Key Delegation** — Set up a session key with spending caps for security
5. **Agent-to-Agent Payments** — Transfer funds between two autonomous agents
6. **Integration** — Wrap SDK methods as tools for an agent framework (LangChain, CrewAI, etc.)
7. **Error Handling** — Handle reconnection, insufficient funds, expired sessions

For each step, show complete TypeScript code examples using the latest SDK API (\`@yellow-org/sdk\`).
Use \`viem\` for Ethereum interactions. Include proper error handling and logging.

## Go SDK

I want to build an AI agent in Go that uses Nitrolite state channels for payments. Guide me through:

1. **Agent Wallet Setup** — Create signers from a private key, configure the SDK client
2. **Channel Management** — Open a channel, deposit funds for the agent
3. **Automated Payments** — A goroutine-safe payment function using context and mutexes
4. **Session Key Delegation** — Set up session keys with spending caps
5. **Agent-to-Agent Payments** — Transfer between two autonomous Go agents
6. **Graceful Shutdown** — Handle OS signals, defer client.Close(), WaitCh()
7. **Error Handling** — Wrapped errors, retry patterns, connection recovery

For each step, show complete Go code using github.com/layer-3/nitrolite/sdk/go.
Use context.Context for timeouts. Use decimal.Decimal for amounts. Follow standard Go patterns.`,
            },
        }],
    }),
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
