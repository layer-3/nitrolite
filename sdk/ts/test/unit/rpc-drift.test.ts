import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, '../../../..');

// This guard intentionally parses rpc_router.go with regex. If router registration
// is refactored away from direct Handle(rpc.XxxMethod.String(), ...) calls, update
// extractRouterHandlers rather than treating the resulting failure as protocol drift.
function extractGoMethodLiterals(source: string): Set<string> {
    const matches = source.matchAll(/^\s*[A-Za-z0-9]+Method\s+Method\s*=\s*"([^"]+)"$/gm);
    return new Set(Array.from(matches, ([, method]) => method));
}

function extractTsMethodLiterals(source: string): Set<string> {
    const matches = source.matchAll(
        /^\s*export const [A-Za-z0-9]+Method:\s*Method\s*=\s*'([^']+)';$/gm
    );
    return new Set(Array.from(matches, ([, method]) => method));
}

function extractRouterHandlers(source: string): Set<string> {
    const matches = source.matchAll(/Handle\(rpc\.[A-Za-z0-9]+Method\.String\(\),/g);
    const methodNames = Array.from(matches, ([match]) =>
        match.match(/rpc\.([A-Za-z0-9]+Method)\.String/)?.[1]
    ).filter((methodName): methodName is string => Boolean(methodName));

    const methodsSource = fs.readFileSync(path.join(repoRoot, 'pkg/rpc/methods.go'), 'utf8');
    const namedLiterals = new Map<string, string>(
        Array.from(
            methodsSource.matchAll(/^\s*([A-Za-z0-9]+Method)\s+Method\s*=\s*"([^"]+)"$/gm),
            ([, name, literal]) => [name, literal]
        )
    );

    const unresolvedMethodNames = methodNames.filter((name) => !namedLiterals.has(name));
    if (unresolvedMethodNames.length > 0) {
        throw new Error(
            `rpc_router.go references unresolved rpc method constants: ${unresolvedMethodNames.join(
                ', '
            )}`
        );
    }

    return new Set(methodNames.map((name) => namedLiterals.get(name) as string));
}

function extractTsClientMethods(source: string): Set<string> {
    const matches = source.matchAll(/^\s{2}(?:static\s+)?(?:async\s+)?([A-Za-z0-9_]+)\(/gm);
    return new Set(Array.from(matches, ([, method]) => method));
}

function sorted(values: Set<string>): string[] {
    return Array.from(values).sort();
}

function diff(left: Set<string>, right: Set<string>): { missing: string[]; extra: string[] } {
    return {
        missing: sorted(new Set(Array.from(right).filter((value) => !left.has(value)))),
        extra: sorted(new Set(Array.from(left).filter((value) => !right.has(value)))),
    };
}

describe('TS RPC drift guards', () => {
    const publicClientMethodsByRPCMethod = new Map<string, string>([
        ['node.v1.ping', 'ping'],
        ['node.v1.get_config', 'getConfig'],
        ['node.v1.get_assets', 'getAssets'],
        ['user.v1.get_balances', 'getBalances'],
        ['user.v1.get_transactions', 'getTransactions'],
        ['user.v1.get_action_allowances', 'getActionAllowances'],
        ['channels.v1.get_home_channel', 'getHomeChannel'],
        ['channels.v1.get_escrow_channel', 'getEscrowChannel'],
        ['channels.v1.get_channels', 'getChannels'],
        ['channels.v1.get_latest_state', 'getLatestState'],
        ['channels.v1.submit_session_key_state', 'submitChannelSessionKeyState'],
        ['channels.v1.get_last_key_states', 'getLastChannelKeyStates'],
        ['app_sessions.v1.submit_deposit_state', 'submitAppSessionDeposit'],
        ['app_sessions.v1.submit_app_state', 'submitAppState'],
        ['app_sessions.v1.rebalance_app_sessions', 'rebalanceAppSessions'],
        ['app_sessions.v1.get_app_definition', 'getAppDefinition'],
        ['app_sessions.v1.get_app_sessions', 'getAppSessions'],
        ['app_sessions.v1.create_app_session', 'createAppSession'],
        ['app_sessions.v1.submit_session_key_state', 'submitSessionKeyState'],
        ['app_sessions.v1.get_last_key_states', 'getLastKeyStates'],
        ['apps.v1.get_apps', 'getApps'],
        ['apps.v1.submit_app_version', 'registerApp'],
    ]);

    const intentionallyRawOnlyMethods = new Set([
        'channels.v1.request_creation',
        'channels.v1.submit_state',
    ]);

    it('keeps the TS raw RPC method surface aligned with pkg/rpc', () => {
        const goMethods = extractGoMethodLiterals(
            fs.readFileSync(path.join(repoRoot, 'pkg/rpc/methods.go'), 'utf8')
        );
        const tsMethods = extractTsMethodLiterals(
            fs.readFileSync(path.join(repoRoot, 'sdk/ts/src/rpc/methods.ts'), 'utf8')
        );

        const { missing, extra } = diff(tsMethods, goMethods);

        expect({ missing, extra }).toEqual({ missing: [], extra: [] });
    });

    it('keeps the TS raw RPC method surface aligned with live router registrations', () => {
        const routerMethods = extractRouterHandlers(
            fs.readFileSync(path.join(repoRoot, 'clearnode/api/rpc_router.go'), 'utf8')
        );
        const tsMethods = extractTsMethodLiterals(
            fs.readFileSync(path.join(repoRoot, 'sdk/ts/src/rpc/methods.ts'), 'utf8')
        );

        const { missing, extra } = diff(tsMethods, routerMethods);

        expect({ missing, extra }).toEqual({ missing: [], extra: [] });
    });

    it('keeps public Client wrappers aligned with public RPC methods', () => {
        const routerMethods = extractRouterHandlers(
            fs.readFileSync(path.join(repoRoot, 'clearnode/api/rpc_router.go'), 'utf8')
        );
        const clientMethods = extractTsClientMethods(
            fs.readFileSync(path.join(repoRoot, 'sdk/ts/src/client.ts'), 'utf8')
        );

        const coveredMethods = new Set([
            ...publicClientMethodsByRPCMethod.keys(),
            ...intentionallyRawOnlyMethods,
        ]);
        const uncoveredRouterMethods = sorted(
            new Set(Array.from(routerMethods).filter((method) => !coveredMethods.has(method)))
        );
        const missingClientMethods = Array.from(publicClientMethodsByRPCMethod)
            .filter(([method]) => routerMethods.has(method))
            .filter(([, clientMethod]) => !clientMethods.has(clientMethod))
            .map(([method, clientMethod]) => `${method} -> Client.${clientMethod}()`)
            .sort();

        expect({ uncoveredRouterMethods, missingClientMethods }).toEqual({
            uncoveredRouterMethods: [],
            missingClientMethods: [],
        });
    });

    it('reports adversarial method additions as missing TS methods', () => {
        const tsMethods = new Set(['node.v1.ping']);
        const goMethods = new Set(['node.v1.ping', 'node.v1.fake_method']);

        expect(diff(tsMethods, goMethods)).toEqual({
            missing: ['node.v1.fake_method'],
            extra: [],
        });
    });

    it('reports adversarial TS method removals as missing public wrappers', () => {
        const routerMethods = new Set(['node.v1.ping']);
        const clientMethods = new Set<string>();
        const mapping = new Map([['node.v1.ping', 'ping']]);

        const missingClientMethods = Array.from(mapping)
            .filter(([method]) => routerMethods.has(method))
            .filter(([, clientMethod]) => !clientMethods.has(clientMethod))
            .map(([method, clientMethod]) => `${method} -> Client.${clientMethod}()`);

        expect(missingClientMethods).toEqual(['node.v1.ping -> Client.ping()']);
    });
});
