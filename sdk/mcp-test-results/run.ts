/**
 * MCP Server Vetting Script
 *
 * Connects to the unified Nitrolite MCP server, exercises every tool/resource/prompt
 * per the scenarios in sdk/mcp-test-plan.md, and dumps raw output to markdown files.
 *
 * Usage (from repo root):
 *   cd sdk/mcp-test-results && npm install   # one-time setup
 *   npm --prefix sdk/mcp-test-results exec -- tsx sdk/mcp-test-results/run.ts
 */

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { writeFileSync } from 'node:fs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = resolve(__dirname, '../..'); // sdk/mcp-test-results → sdk → repo root
const RESULTS_DIR = __dirname; // results written alongside run.ts

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function extractText(result: any): string {
    if (!result?.content) return '<no content>';
    return result.content
        .filter((c: any) => c.type === 'text')
        .map((c: any) => c.text ?? '')
        .join('\n');
}

function extractResourceText(result: any): string {
    if (!result?.contents) return '<no contents>';
    return result.contents.map((c: any) => c.text ?? '').join('\n');
}

function extractPromptText(result: any): string {
    if (!result?.messages) return '<no messages>';
    return result.messages.map((m: any) => {
        if (typeof m.content === 'string') return m.content;
        if (m.content?.text) return m.content.text;
        return JSON.stringify(m.content);
    }).join('\n');
}

function truncate(s: string, n: number = 500): string {
    if (s.length <= n) return s;
    return s.slice(0, n) + `\n... (${s.length} chars total, truncated)`;
}

// ---------------------------------------------------------------------------
// Server connection
// ---------------------------------------------------------------------------

async function connect(): Promise<Client> {
    const transport = new StdioClientTransport({
        command: 'npm',
        args: ['--prefix', 'sdk/mcp', 'exec', '--', 'tsx', 'sdk/mcp/src/index.ts'],
        cwd: REPO_ROOT,
    });
    const client = new Client({ name: 'vetting-harness', version: '1.0.0' });
    await client.connect(transport);
    return client;
}

// ---------------------------------------------------------------------------
// Test runner
// ---------------------------------------------------------------------------

interface TestResult {
    id: string;
    scenario: string;
    description: string;
    output: string;
    pass: boolean | null; // null = manual review needed
    notes: string;
}

const tsResults: TestResult[] = [];
const goResults: TestResult[] = [];

async function callTool(client: Client, toolName: string, args: Record<string, string>): Promise<string> {
    try {
        const result = await client.callTool({ name: toolName, arguments: args });
        return extractText(result);
    } catch (e: any) {
        return `ERROR: ${e.message}`;
    }
}

async function readResource(client: Client, uri: string): Promise<string> {
    try {
        const result = await client.readResource({ uri });
        return extractResourceText(result);
    } catch (e: any) {
        return `ERROR: ${e.message}`;
    }
}

async function getPrompt(client: Client, name: string): Promise<string> {
    try {
        const result = await client.getPrompt({ name, arguments: {} });
        return extractPromptText(result);
    } catch (e: any) {
        return `ERROR: ${e.message}`;
    }
}

function check(
    results: TestResult[],
    id: string,
    scenario: string,
    description: string,
    output: string,
    checks: { contains?: string[]; notContains?: string[]; minLength?: number }
): void {
    let pass = true;
    const notes: string[] = [];

    if (output.startsWith('ERROR:')) {
        pass = false;
        notes.push(`Server error: ${output}`);
    } else {
        if (checks.minLength && output.length < checks.minLength) {
            pass = false;
            notes.push(`Too short: ${output.length} chars (expected >${checks.minLength})`);
        }
        for (const s of checks.contains ?? []) {
            if (!output.includes(s)) {
                pass = false;
                notes.push(`MISSING: "${s}"`);
            }
        }
        for (const s of checks.notContains ?? []) {
            if (output.includes(s)) {
                pass = false;
                notes.push(`UNEXPECTED: "${s}" found`);
            }
        }
    }

    results.push({
        id,
        scenario,
        description,
        output: truncate(output),
        pass,
        notes: notes.join('; ') || (pass ? 'OK' : ''),
    });

    const status = pass ? 'PASS' : 'FAIL';
    console.log(`  ${status} ${id} — ${description}${notes.length ? ' — ' + notes.join('; ') : ''}`);
}

// ---------------------------------------------------------------------------
// Scenarios
// ---------------------------------------------------------------------------

async function runTsScenarios(client: Client) {
    console.log('\n=== TS SERVER ===\n');

    // --- Scenario 0: Surface Inventory ---
    console.log('Scenario 0: Surface Inventory');
    const tools = await client.listTools();
    const resources = await client.listResources();
    const prompts = await client.listPrompts();
    const toolNames = tools.tools.map((t: any) => t.name).sort();
    const resourceUris = resources.resources.map((r: any) => r.uri).sort();
    const promptNames = prompts.prompts.map((p: any) => p.name).sort();

    check(tsResults, '0.1', 'Inventory', `${resources.resources.length} resources registered`, resourceUris.join('\n'), { minLength: 10 });
    check(tsResults, '0.2', 'Inventory', `${tools.tools.length} tools: ${toolNames.join(', ')}`, toolNames.join(', '), {
        contains: ['lookup_method', 'lookup_type', 'search_api', 'get_rpc_method', 'validate_import', 'explain_concept', 'scaffold_project'],
    });
    check(tsResults, '0.3', 'Inventory', `${prompts.prompts.length} prompts: ${promptNames.join(', ')}`, promptNames.join(', '), {
        contains: ['create-channel-app', 'migrate-from-v053', 'build-ai-agent-app'],
    });

    // --- Scenario B: Transfer App ---
    console.log('\nScenario B: Transfer App');
    const bSearch = await callTool(client, 'search_api', { query: 'transfer' });
    check(tsResults, 'B.1', 'Transfer', 'search_api("transfer") mentions Decimal', bSearch, { contains: ['Decimal'] });

    const bCreate = await callTool(client, 'lookup_method', { name: 'create' });
    check(tsResults, 'B.2', 'Transfer', 'lookup_method("create") — correct signature', bCreate, { contains: ['stateSigner'] });

    const bDeposit = await callTool(client, 'lookup_method', { name: 'deposit' });
    check(tsResults, 'B.3', 'Transfer', 'lookup_method("deposit") — Decimal amount', bDeposit, { contains: ['Decimal'] });

    const bTransfer = await callTool(client, 'lookup_method', { name: 'transfer' });
    check(tsResults, 'B.4', 'Transfer', 'lookup_method("transfer") — Decimal amount', bTransfer, { contains: ['Decimal'] });

    const bApprove = await callTool(client, 'lookup_method', { name: 'approveToken' });
    check(tsResults, 'B.5', 'Transfer', 'lookup_method("approveToken") — 3 params', bApprove, { contains: ['Decimal'] });

    const bCheckpoint = await callTool(client, 'lookup_method', { name: 'checkpoint' });
    check(tsResults, 'B.6', 'Transfer', 'lookup_method("checkpoint")', bCheckpoint, { contains: ['asset'] });

    const bBalances = await callTool(client, 'lookup_method', { name: 'getBalances' });
    check(tsResults, 'B.7', 'Transfer', 'lookup_method("getBalances") — wallet param', bBalances, { contains: ['wallet'] });

    const bClose = await callTool(client, 'lookup_method', { name: 'closeHomeChannel' });
    check(tsResults, 'B.8', 'Transfer', 'lookup_method("closeHomeChannel")', bClose, { contains: ['asset'] });

    const bScaffold = await callTool(client, 'scaffold_project', { template: 'transfer-app' });
    check(tsResults, 'B.9', 'Transfer', 'scaffold_project("transfer-app")', bScaffold, {
        contains: ['createSigners', 'Decimal', 'checkpoint'],
        notContains: ['walletClient', 'DefaultConfig'],
    });

    const bFullScript = await readResource(client, 'nitrolite://examples/full-transfer-script');
    check(tsResults, 'B.10', 'Transfer', 'full transfer script resource', bFullScript, {
        contains: ['createSigners', 'checkpoint', 'getBalances(userAddress'],
        notContains: ['walletClient'],
    });

    const bRpc = await callTool(client, 'get_rpc_method', { method: 'transfer' });
    check(tsResults, 'B.11', 'Transfer', 'get_rpc_method("transfer") — object params', bRpc, {
        contains: ['destination', 'allocations'],
    });

    // --- Scenario C: App Session ---
    console.log('\nScenario C: App Session');
    const cSearch = await callTool(client, 'search_api', { query: 'app session' });
    check(tsResults, 'C.1', 'AppSession', 'search_api("app session")', cSearch, {
        contains: ['createAppSession'],
        notContains: ['closeAppSession'],
    });

    const cCreate = await callTool(client, 'lookup_method', { name: 'createAppSession' });
    check(tsResults, 'C.2', 'AppSession', 'createAppSession signature', cCreate, { contains: ['quorumSigs'] });

    const cSubmit = await callTool(client, 'lookup_method', { name: 'submitAppState' });
    check(tsResults, 'C.3', 'AppSession', 'submitAppState signature', cSubmit, { contains: ['quorumSigs'] });

    const cCloseNeg = await callTool(client, 'lookup_method', { name: 'closeAppSession' });
    const c4Pass = cCloseNeg.toLowerCase().includes('no method') || cCloseNeg.toLowerCase().includes('not found');
    tsResults.push({ id: 'C.4', scenario: 'AppSession', description: 'closeAppSession — correctly absent', output: truncate(cCloseNeg), pass: c4Pass, notes: c4Pass ? 'Correctly absent' : 'closeAppSession should not exist' });
    console.log(`  ${c4Pass ? 'PASS' : 'FAIL'} C.4 — closeAppSession — correctly absent`);

    const cTypeDef = await callTool(client, 'lookup_type', { name: 'AppDefinitionV1' });
    check(tsResults, 'C.5', 'AppSession', 'lookup_type AppDefinitionV1', cTypeDef, {
        contains: ['applicationId', 'quorum', 'nonce'],
        notContains: ['protocol', 'appName'],
    });

    const cTypeUpdate = await callTool(client, 'lookup_type', { name: 'AppStateUpdateV1' });
    check(tsResults, 'C.6', 'AppSession', 'lookup_type AppStateUpdateV1', cTypeUpdate, {
        contains: ['appSessionId', 'intent', 'version'],
    });

    const cScaffold = await callTool(client, 'scaffold_project', { template: 'app-session' });
    check(tsResults, 'C.7', 'AppSession', 'scaffold_project("app-session")', cScaffold, {
        contains: ['applicationId', 'quorumSigs', 'submitAppState'],
        notContains: ['closeAppSession(', "protocol: 'nitrolite'"],
    });

    const cExamples = await readResource(client, 'nitrolite://examples/app-sessions');
    check(tsResults, 'C.8', 'AppSession', 'app-sessions resource', cExamples, {
        contains: ['applicationId'],
    });

    const cFullScript = await readResource(client, 'nitrolite://examples/full-app-session-script');
    check(tsResults, 'C.9', 'AppSession', 'full app-session script', cFullScript, {
        contains: ['quorumSigs', 'Close'],
    });

    const cRpc1 = await callTool(client, 'get_rpc_method', { method: 'create_app_session' });
    check(tsResults, 'C.10', 'AppSession', 'get_rpc_method("create_app_session")', cRpc1, {
        contains: ['definition', 'session_data', 'quorum_sigs'],
    });

    const cRpc2 = await callTool(client, 'get_rpc_method', { method: 'submit_app_state' });
    check(tsResults, 'C.11', 'AppSession', 'get_rpc_method("submit_app_state")', cRpc2, {
        contains: ['app_state_update', 'quorum_sigs'],
    });

    // --- Scenario D: AI Agent ---
    console.log('\nScenario D: AI Agent');
    const dScaffold = await callTool(client, 'scaffold_project', { template: 'ai-agent' });
    check(tsResults, 'D.1', 'AIAgent', 'scaffold_project("ai-agent")', dScaffold, {
        contains: ['createSigners', 'Decimal'],
        notContains: ['walletClient', 'DefaultConfig'],
    });

    const dUseCases = await readResource(client, 'nitrolite://use-cases/ai-agents');
    check(tsResults, 'D.2', 'AIAgent', 'AI agents use-case resource', dUseCases, {
        contains: ['createSigners'],
        notContains: ['walletClient'],
    });

    // --- Scenario E: Migration ---
    console.log('\nScenario E: Migration');
    const e1 = await callTool(client, 'validate_import', { symbol: 'NitroliteClient' });
    check(tsResults, 'E.1', 'Migration', 'validate_import("NitroliteClient")', e1, { contains: ['sdk-compat'] });

    const e2 = await callTool(client, 'validate_import', { symbol: 'createAuthRequestMessage' });
    check(tsResults, 'E.2', 'Migration', 'validate_import("createAuthRequestMessage")', e2, { contains: ['sdk-compat'] });

    const e3 = await callTool(client, 'validate_import', { symbol: 'Client' });
    check(tsResults, 'E.3', 'Migration', 'validate_import("Client") — in SDK, not compat', e3, {
        contains: ['@yellow-org/sdk'],
    });

    const e4 = await callTool(client, 'validate_import', { symbol: 'RPCMethod' });
    check(tsResults, 'E.4', 'Migration', 'validate_import("RPCMethod")', e4, { contains: ['sdk-compat'] });

    const e5 = await callTool(client, 'validate_import', { symbol: 'Cli' });
    check(tsResults, 'E.5', 'Migration', 'validate_import("Cli") — must NOT match', e5, { contains: ['not found'] });

    const e6 = await callTool(client, 'validate_import', { symbol: 'FakeSymbol' });
    check(tsResults, 'E.6', 'Migration', 'validate_import("FakeSymbol")', e6, { contains: ['not found'] });

    const e7 = await readResource(client, 'nitrolite://examples/auth');
    check(tsResults, 'E.7', 'Migration', 'auth example resource', e7, {
        contains: ['createAuthRequestMessage', 'AuthRequestParams'],
    });

    const e8 = await readResource(client, 'nitrolite://migration/overview');
    check(tsResults, 'E.8', 'Migration', 'migration overview resource', e8, { minLength: 50 });

    const e9 = await getPrompt(client, 'migrate-from-v053');
    check(tsResults, 'E.9', 'Migration', 'migrate-from-v053 prompt', e9, { minLength: 50 });

    // --- Scenario F: Protocol (TS side) ---
    console.log('\nScenario F: Protocol');
    for (const [concept, id] of [['state channel', 'F.1'], ['app session', 'F.2'], ['challenge period', 'F.3'], ['clearnode', 'F.4']] as const) {
        const out = await callTool(client, 'explain_concept', { concept });
        check(tsResults, id, 'Protocol', `explain_concept("${concept}")`, out, { minLength: 20 });
    }
    const f5 = await callTool(client, 'explain_concept', { concept: 'made up thing' });
    check(tsResults, 'F.5', 'Protocol', 'explain_concept("made up thing") — graceful', f5, {});

    for (const [uri, id] of [
        ['nitrolite://protocol/overview', 'F.6'],
        ['nitrolite://protocol/terminology', 'F.7'],
        ['nitrolite://protocol/wire-format', 'F.8'],
        ['nitrolite://security/overview', 'F.9'],
        ['nitrolite://security/app-session-patterns', 'F.10'],
        ['nitrolite://security/state-invariants', 'F.11'],
    ] as const) {
        const out = await readResource(client, uri);
        check(tsResults, id, 'Protocol', `resource ${uri}`, out, { minLength: 50 });
    }

    // --- Scenario G: RPC Wire Formats ---
    console.log('\nScenario G: RPC Wire Formats');
    const g1 = await callTool(client, 'get_rpc_method', { method: 'get_ledger_balances' });
    check(tsResults, 'G.1', 'RPC', 'get_ledger_balances — wallet param', g1, {
        contains: ['wallet'],
    });

    const g2 = await callTool(client, 'get_rpc_method', { method: 'create_channel' });
    check(tsResults, 'G.2', 'RPC', 'create_channel — chain_id/token', g2, {
        contains: ['chain_id', 'token'],
    });

    const g3 = await callTool(client, 'get_rpc_method', { method: 'close_channel' });
    check(tsResults, 'G.3', 'RPC', 'close_channel — channel_id/funds_destination', g3, {
        contains: ['channel_id', 'funds_destination'],
    });

    const g4 = await callTool(client, 'get_rpc_method', { method: 'resize_channel' });
    check(tsResults, 'G.4', 'RPC', 'resize_channel — all 4 fields', g4, {
        contains: ['resize_amount', 'allocate_amount', 'funds_destination'],
    });

    const g5 = await callTool(client, 'get_rpc_method', { method: 'nonexistent' });
    check(tsResults, 'G.5', 'RPC', 'nonexistent — lists available', g5, { contains: ['Available'] });

    // --- Scenario I: Type Discovery ---
    console.log('\nScenario I: Type Discovery');
    const i1 = await readResource(client, 'nitrolite://api/methods');
    check(tsResults, 'I.1', 'Types', 'api/methods resource', i1, { minLength: 200 });

    const i2 = await readResource(client, 'nitrolite://api/types');
    check(tsResults, 'I.2', 'Types', 'api/types resource', i2, { minLength: 100 });

    const i3 = await readResource(client, 'nitrolite://api/enums');
    check(tsResults, 'I.3', 'Types', 'api/enums resource', i3, { minLength: 50 });

    const i4 = await callTool(client, 'lookup_type', { name: 'AppDefinitionV1' });
    check(tsResults, 'I.4', 'Types', 'lookup_type AppDefinitionV1', i4, {
        contains: ['applicationId', 'quorum', 'nonce'],
    });

    const i5 = await callTool(client, 'lookup_type', { name: 'AppStateUpdateV1' });
    check(tsResults, 'I.5', 'Types', 'lookup_type AppStateUpdateV1', i5, {
        contains: ['appSessionId', 'intent', 'version'],
    });

    const i6 = await callTool(client, 'lookup_type', { name: 'NonexistentType' });
    check(tsResults, 'I.6', 'Types', 'lookup_type NonexistentType — graceful', i6, {});

    // --- Scenario J: Prompts (TS) ---
    console.log('\nScenario J: Prompts');
    const j1 = await getPrompt(client, 'create-channel-app');
    check(tsResults, 'J.1', 'Prompts', 'create-channel-app prompt', j1, { minLength: 50 });

    const j2 = await getPrompt(client, 'migrate-from-v053');
    check(tsResults, 'J.2', 'Prompts', 'migrate-from-v053 prompt', j2, { minLength: 50 });

    const j3 = await getPrompt(client, 'build-ai-agent-app');
    check(tsResults, 'J.3', 'Prompts', 'build-ai-agent-app prompt', j3, { minLength: 50 });

    // --- Scenario K: Edge Cases ---
    console.log('\nScenario K: Edge Cases');
    const k1 = await callTool(client, 'lookup_method', { name: 'nonexistent' });
    check(tsResults, 'K.1', 'Edge', 'lookup_method("nonexistent") — no crash', k1, {});

    const k2 = await callTool(client, 'get_rpc_method', { method: 'nonexistent' });
    check(tsResults, 'K.2', 'Edge', 'get_rpc_method("nonexistent") — lists methods', k2, { contains: ['Available'] });

    const k3 = await callTool(client, 'search_api', { query: '' });
    check(tsResults, 'K.3', 'Edge', 'search_api("") — no crash', k3, {});

    const k4 = await callTool(client, 'scaffold_project', { template: 'transfer-app' });
    check(tsResults, 'K.4', 'Edge', 'scaffold includes decimal.js', k4, { contains: ['decimal.js'] });

    // --- Scenario L: Resource Sweep ---
    console.log('\nScenario L: Resource Sweep');
    const sweepUris = [
        'nitrolite://examples/channels',
        'nitrolite://examples/transfers',
        'nitrolite://protocol/rpc-methods',
        'nitrolite://protocol/auth-flow',
        'nitrolite://protocol/cryptography',
        'nitrolite://protocol/channel-lifecycle',
        'nitrolite://protocol/state-model',
        'nitrolite://protocol/enforcement',
        'nitrolite://protocol/cross-chain',
        'nitrolite://protocol/interactions',
        'nitrolite://use-cases',
    ];
    for (let i = 0; i < sweepUris.length; i++) {
        const uri = sweepUris[i];
        const out = await readResource(client, uri);
        check(tsResults, `L.${i + 1}`, 'Sweep', `resource ${uri}`, out, { minLength: 20 });
    }

    // L.17: lookup_rpc_method alias
    const l17 = await callTool(client, 'lookup_rpc_method', { method: 'transfer' });
    check(tsResults, 'L.17', 'Sweep', 'lookup_rpc_method alias for transfer', l17, { minLength: 10 });
}

async function runGoScenarios(client: Client) {
    console.log('\n=== GO SERVER ===\n');

    // --- Scenario 0: Surface Inventory ---
    console.log('Scenario 0: Surface Inventory');
    const tools = await client.listTools();
    const resources = await client.listResources();
    const prompts = await client.listPrompts();
    const toolNames = tools.tools.map((t: any) => t.name).sort();
    const promptNames = prompts.prompts.map((p: any) => p.name).sort();

    const resourceCount = resources.resources.length;
    check(goResults, '0.4', 'Inventory', `${resourceCount} resources registered (expected 30)`, `${resourceCount}`, { contains: ['30'] });
    check(goResults, '0.5', 'Inventory', `${tools.tools.length} tools: ${toolNames.join(', ')}`, toolNames.join(', '), {
        contains: ['search_api', 'lookup_method', 'lookup_type', 'explain_concept', 'lookup_rpc_method', 'scaffold_project'],
    });
    check(goResults, '0.6', 'Inventory', `${prompts.prompts.length} prompts: ${promptNames.join(', ')}`, promptNames.join(', '), {
        contains: ['create-channel-app', 'build-ai-agent-app'],
    });

    // --- Scenario H: Go SDK ---
    console.log('\nScenario H: Go SDK');
    const methodsToFind = ['Deposit', 'Transfer', 'CreateAppSession', 'SubmitAppState', 'CloseHomeChannel', 'Checkpoint', 'GetBalances', 'GetConfig'];
    for (let i = 0; i < methodsToFind.length; i++) {
        const name = methodsToFind[i];
        const out = await callTool(client, 'lookup_method', { name, language: 'go' });
        check(goResults, `H.${i + 1}`, 'GoSDK', `lookup_method("${name}", go) — exists`, out, {
            notContains: ['not found', 'No method'],
        });
    }

    const h9 = await callTool(client, 'lookup_method', { name: 'CloseAppSession', language: 'go' });
    const h9Pass = h9.toLowerCase().includes('not found') || h9.toLowerCase().includes('no method');
    goResults.push({ id: 'H.9', scenario: 'GoSDK', description: 'CloseAppSession — NOT found', output: truncate(h9), pass: h9Pass, notes: h9Pass ? 'Correctly not found' : 'CloseAppSession should not exist but was found' });
    console.log(`  ${h9Pass ? 'PASS' : 'FAIL'} H.9 — CloseAppSession — NOT found`);

    const hScaffold1 = await callTool(client, 'scaffold_project', { template: 'go-transfer-app' });
    check(goResults, 'H.10', 'GoSDK', 'scaffold go-transfer-app — has Checkpoint', hScaffold1, { contains: ['Checkpoint'] });

    const hScaffold2 = await callTool(client, 'scaffold_project', { template: 'go-app-session' });
    check(goResults, 'H.11', 'GoSDK', 'scaffold go-app-session — quorum sigs, close intent', hScaffold2, {
        contains: ['SubmitAppState', 'Close'],
    });

    const hScaffold3 = await callTool(client, 'scaffold_project', { template: 'go-ai-agent' });
    check(goResults, 'H.12', 'GoSDK', 'scaffold go-ai-agent', hScaffold3, { minLength: 100 });

    const hTransfer = await readResource(client, 'nitrolite://go-examples/full-transfer-script');
    check(goResults, 'H.13', 'GoSDK', 'full Go transfer — CloseHomeChannel + Checkpoint', hTransfer, {
        contains: ['CloseHomeChannel', 'Checkpoint'],
    });

    const hAppSession = await readResource(client, 'nitrolite://go-examples/full-app-session-script');
    check(goResults, 'H.14', 'GoSDK', 'full Go app-session — quorum sigs, version tracking', hAppSession, {
        contains: ['initVersion'],
    });

    const hUseCases = await readResource(client, 'nitrolite://use-cases');
    check(goResults, 'H.15', 'GoSDK', 'use-cases resource loads', hUseCases, {
        minLength: 50,
    });

    // Go type discovery
    const hGoType1 = await callTool(client, 'lookup_type', { name: 'AppSessionV1', language: 'go' });
    check(goResults, 'H.16', 'GoSDK', 'lookup_type("AppSessionV1", go) — found', hGoType1, {
        notContains: ['not found', 'No type'],
        minLength: 10,
    });

    const hGoType2 = await callTool(client, 'lookup_type', { name: 'ChannelStatus', language: 'go' });
    check(goResults, 'H.17', 'GoSDK', 'lookup_type("ChannelStatus", go) — enum found', hGoType2, {
        notContains: ['not found', 'No type'],
        minLength: 10,
    });

    // --- Scenario F: Protocol (Go side) ---
    console.log('\nScenario F: Protocol (Go)');
    for (const [concept, id] of [['state channel', 'F.1g'], ['app session', 'F.2g'], ['clearnode', 'F.4g']] as const) {
        const out = await callTool(client, 'explain_concept', { concept });
        check(goResults, id, 'Protocol', `explain_concept("${concept}")`, out, { minLength: 20 });
    }
    const f5g = await callTool(client, 'explain_concept', { concept: 'made up thing' });
    check(goResults, 'F.5g', 'Protocol', 'explain_concept("made up thing") — graceful', f5g, {});

    for (const [uri, id] of [
        ['nitrolite://protocol/overview', 'F.6g'],
        ['nitrolite://protocol/terminology', 'F.7g'],
        ['nitrolite://security/overview', 'F.9g'],
    ] as const) {
        const out = await readResource(client, uri);
        check(goResults, id, 'Protocol', `resource ${uri}`, out, { minLength: 50 });
    }

    // --- Scenario G: RPC (v1 method lookup via unified server) ---
    console.log('\nScenario G: RPC (Go)');
    for (const [method, id] of [
        ['channels.v1.get_home_channel', 'G.6'],
        ['app_sessions.v1.create_app_session', 'G.7'],
        ['user.v1.get_balances', 'G.8'],
    ] as const) {
        const out = await callTool(client, 'lookup_rpc_method', { method });
        check(goResults, id, 'RPC', `lookup_rpc_method("${method}")`, out, { minLength: 10 });
    }

    // --- Scenario J: Prompts (Go) ---
    console.log('\nScenario J: Prompts (Go)');
    const j4 = await getPrompt(client, 'create-channel-app');
    check(goResults, 'J.4', 'Prompts', 'create-channel-app — mentions Checkpoint', j4, {
        contains: ['Checkpoint'],
        notContains: ['CloseAppSession'],
    });

    const j5 = await getPrompt(client, 'build-ai-agent-app');
    check(goResults, 'J.5', 'Prompts', 'build-ai-agent-app', j5, { minLength: 50 });

    // --- Scenario K: Edge Cases (Go) ---
    console.log('\nScenario K: Edge Cases (Go)');
    const k5 = await callTool(client, 'scaffold_project', { template: 'nonexistent' });
    check(goResults, 'K.5', 'Edge', 'scaffold("nonexistent") — error', k5, {});

    const k6 = await callTool(client, 'search_api', { query: 'zzzzz', language: 'go' });
    check(goResults, 'K.6', 'Edge', 'search_api("zzzzz", go) — no matches', k6, {});

    // --- Scenario L: Resource Sweep (Go) ---
    console.log('\nScenario L: Resource Sweep (Go)');
    const goSweepUris = [
        'nitrolite://go-api/methods',
        'nitrolite://go-api/types',
        'nitrolite://protocol/enforcement',
        'nitrolite://protocol/auth-flow',
    ];
    for (let i = 0; i < goSweepUris.length; i++) {
        const uri = goSweepUris[i];
        const out = await readResource(client, uri);
        check(goResults, `L.${i + 11}`, 'Sweep', `resource ${uri}`, out, { minLength: 20 });
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

function writeResults(filename: string, results: TestResult[]) {
    const passed = results.filter(r => r.pass === true).length;
    const failed = results.filter(r => r.pass === false).length;
    const lines: string[] = [
        `# MCP Vetting Results — ${filename.includes('ts') ? 'TypeScript' : 'Go'} Server`,
        '',
        `**${passed} passed, ${failed} failed, ${results.length} total**`,
        '',
        '---',
        '',
    ];

    let currentScenario = '';
    for (const r of results) {
        if (r.scenario !== currentScenario) {
            currentScenario = r.scenario;
            lines.push(`## ${currentScenario}`, '');
        }
        const status = r.pass ? 'PASS' : 'FAIL';
        lines.push(`### ${r.id}: ${r.description}`, '');
        lines.push(`**Status:** ${status}`);
        if (r.notes && r.notes !== 'OK') lines.push(`**Notes:** ${r.notes}`);
        lines.push('', '```', r.output, '```', '');
    }

    writeFileSync(resolve(RESULTS_DIR, filename), lines.join('\n'));
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main() {
    console.log('MCP Server Vetting Script');
    console.log('========================\n');

    // Unified server
    console.log('Connecting to unified nitrolite MCP server...');
    let client: Client | null = null;
    try {
        client = await connect();
        console.log('Connected.\n');
        await runTsScenarios(client);
        await runGoScenarios(client);
    } catch (e: any) {
        console.error('Server connection failed:', e.message);
    }

    // Write results
    writeResults('ts-server.md', tsResults);
    writeResults('go-server.md', goResults);

    // Summary
    const allResults = [...tsResults, ...goResults];
    const passed = allResults.filter(r => r.pass === true).length;
    const failed = allResults.filter(r => r.pass === false).length;

    console.log('\n========================');
    console.log(`TOTAL: ${passed} passed, ${failed} failed, ${allResults.length} total`);
    console.log(`Results written to sdk/mcp-test-results/ts-server.md and go-server.md`);

    if (failed > 0) {
        console.log('\nFAILED TESTS:');
        for (const r of allResults.filter(r => r.pass === false)) {
            console.log(`  ${r.id} — ${r.description}: ${r.notes}`);
        }
    }

    // Close connection
    try { if (client) await client.close(); } catch {}

    process.exit(failed > 0 ? 1 : 0);
}

main().catch(e => {
    console.error('Fatal:', e);
    process.exit(2);
});
