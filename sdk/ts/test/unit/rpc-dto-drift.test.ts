import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, '../../../..');

type FieldShape = {
    optional: boolean;
    container: 'array' | 'scalar';
};

type DTOShape = Record<string, Record<string, FieldShape>>;

function normalizeGoContainer(typeText: string): FieldShape['container'] {
    return typeText.replace(/^\*/, '').trim().startsWith('[]') ? 'array' : 'scalar';
}

function normalizeTsContainer(typeText: string): FieldShape['container'] {
    return typeText.trim().endsWith('[]') || typeText.includes('Array<') ? 'array' : 'scalar';
}

function extractGoDTOShapes(source: string): DTOShape {
    const shapes: DTOShape = {};
    const emptyStructs = source.matchAll(
        /^type\s+([A-Za-z0-9]+(?:Request|Response))\s+struct\s*\{\s*\}$/gm
    );
    for (const [, typeName] of emptyStructs) {
        shapes[typeName] = {};
    }

    const structs = source.matchAll(
        /^type\s+([A-Za-z0-9]+(?:Request|Response))\s+struct\s*\{\n([\s\S]*?)^\}/gm
    );

    for (const [, typeName, body] of structs) {
        const fields: Record<string, FieldShape> = {};
        for (const line of body.split('\n')) {
            const tag = line.match(/`json:"([^"]+)"`/);
            if (!tag) continue;

            const [wireName, ...tagOptions] = tag[1].split(',');
            if (wireName === '-') continue;

            const beforeTag = line.slice(0, line.indexOf('`')).trim();
            const parts = beforeTag.split(/\s+/);
            const typeText = parts.slice(1).join(' ');

            fields[wireName] = {
                optional: typeText.startsWith('*') || tagOptions.includes('omitempty'),
                container: normalizeGoContainer(typeText),
            };
        }
        shapes[typeName] = fields;
    }

    return shapes;
}

function extractTsDTOShapes(source: string): DTOShape {
    const shapes: DTOShape = {};
    const emptyInterfaces = source.matchAll(
        /^export\s+interface\s+([A-Za-z0-9]+(?:Request|Response))\s*\{\s*\}$/gm
    );
    for (const [, typeName] of emptyInterfaces) {
        shapes[typeName] = {};
    }

    const interfaces = source.matchAll(
        /^export\s+interface\s+([A-Za-z0-9]+(?:Request|Response))\s*\{\n([\s\S]*?)^\}/gm
    );

    for (const [, typeName, body] of interfaces) {
        const fields: Record<string, FieldShape> = {};
        for (const line of body.split('\n')) {
            const field = line.match(/^\s*([A-Za-z0-9_]+)(\?)?:\s*([^;]+);/);
            if (!field) continue;

            const [, wireName, optionalMarker, typeText] = field;
            fields[wireName] = {
                optional: optionalMarker === '?',
                container: normalizeTsContainer(typeText),
            };
        }
        shapes[typeName] = fields;
    }

    return shapes;
}

function diffDTOShapes(tsShapes: DTOShape, goShapes: DTOShape) {
    const missingTypes = Object.keys(goShapes).filter((typeName) => !(typeName in tsShapes)).sort();
    const extraTypes = Object.keys(tsShapes).filter((typeName) => !(typeName in goShapes)).sort();
    const fieldDiffs: string[] = [];

    for (const typeName of Object.keys(goShapes).filter((name) => name in tsShapes).sort()) {
        const goFields = goShapes[typeName];
        const tsFields = tsShapes[typeName];
        for (const fieldName of Object.keys(goFields).sort()) {
            if (!(fieldName in tsFields)) {
                fieldDiffs.push(`${typeName}.${fieldName}: missing in TS`);
                continue;
            }
            if (goFields[fieldName].optional !== tsFields[fieldName].optional) {
                fieldDiffs.push(
                    `${typeName}.${fieldName}: optionality Go=${goFields[fieldName].optional} TS=${tsFields[fieldName].optional}`
                );
            }
            if (goFields[fieldName].container !== tsFields[fieldName].container) {
                fieldDiffs.push(
                    `${typeName}.${fieldName}: container Go=${goFields[fieldName].container} TS=${tsFields[fieldName].container}`
                );
            }
        }
        for (const fieldName of Object.keys(tsFields).sort()) {
            if (!(fieldName in goFields)) {
                fieldDiffs.push(`${typeName}.${fieldName}: extra in TS`);
            }
        }
    }

    return { missingTypes, extraTypes, fieldDiffs };
}

describe('RPC DTO drift guards', () => {
    it('keeps Go RPC JSON DTO fields aligned with TS RPC interfaces', () => {
        const goShapes = extractGoDTOShapes(
            fs.readFileSync(path.join(repoRoot, 'pkg/rpc/api.go'), 'utf8')
        );
        const tsShapes = extractTsDTOShapes(
            fs.readFileSync(path.join(repoRoot, 'sdk/ts/src/rpc/api.ts'), 'utf8')
        );

        expect(diffDTOShapes(tsShapes, goShapes)).toEqual({
            missingTypes: [],
            extraTypes: [],
            fieldDiffs: [],
        });
    });

    it('reports adversarial Go-only required fields with field-level paths', () => {
        const goShapes = extractGoDTOShapes(`
type NodeV1PingRequest struct {
    Required string \`json:"required"\`
}
`);
        const tsShapes = extractTsDTOShapes(`
export interface NodeV1PingRequest {}
`);

        expect(diffDTOShapes(tsShapes, goShapes).fieldDiffs).toEqual([
            'NodeV1PingRequest.required: missing in TS',
        ]);
    });

    it('reports adversarial optionality and array/scalar drift', () => {
        const goShapes = extractGoDTOShapes(`
type NodeV1GetAssetsResponse struct {
    Assets []AssetV1 \`json:"assets,omitempty"\`
}
`);
        const tsShapes = extractTsDTOShapes(`
export interface NodeV1GetAssetsResponse {
  assets: AssetV1;
}
`);

        expect(diffDTOShapes(tsShapes, goShapes).fieldDiffs).toEqual([
            'NodeV1GetAssetsResponse.assets: optionality Go=true TS=false',
            'NodeV1GetAssetsResponse.assets: container Go=array TS=scalar',
        ]);
    });
});
