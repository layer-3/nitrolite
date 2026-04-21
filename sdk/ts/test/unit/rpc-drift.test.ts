import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, '../../../..');

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

    return new Set(methodNames.map((name) => namedLiterals.get(name)).filter(Boolean) as string[]);
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

    it('keeps every TS raw RPC method registered by the live router', () => {
        const routerMethods = extractRouterHandlers(
            fs.readFileSync(path.join(repoRoot, 'clearnode/api/rpc_router.go'), 'utf8')
        );
        const tsMethods = extractTsMethodLiterals(
            fs.readFileSync(path.join(repoRoot, 'sdk/ts/src/rpc/methods.ts'), 'utf8')
        );

        const { missing, extra } = diff(tsMethods, routerMethods);

        expect({ missing, extra }).toEqual({ missing: [], extra: [] });
    });
});
