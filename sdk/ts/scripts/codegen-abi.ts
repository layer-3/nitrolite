import { readFileSync, writeFileSync } from 'fs';
import { join } from 'path';
import { fileURLToPath } from 'url';

const JSON_ARTIFACT_PATH = '../../../contracts/out/ChannelHub.sol/ChannelHub.json';
const OUTPUT_PATH = '../src/blockchain/evm/channel_hub_abi.ts';

const __filename = fileURLToPath(import.meta.url);
const __dirname = join(__filename, '..');

/**
 * Converts a JSON value to a TypeScript literal string with:
 * - Unquoted object keys (when valid identifiers)
 * - Single-quoted string values
 * - 2-space indentation
 */
function formatValue(value: unknown, indent: number): string {
    const pad = ' '.repeat(indent);
    const innerPad = ' '.repeat(indent + 2);

    if (value === null) return 'null';
    if (typeof value === 'boolean') return String(value);
    if (typeof value === 'number') return String(value);
    if (typeof value === 'string') return `'${value.replace(/'/g, "\\'")}'`;

    if (Array.isArray(value)) {
        if (value.length === 0) return '[]';
        const items = value.map((item) => `${innerPad}${formatValue(item, indent + 2)}`);
        return `[\n${items.join(',\n')}\n${pad}]`;
    }

    if (typeof value === 'object') {
        const entries = Object.entries(value as Record<string, unknown>);
        if (entries.length === 0) return '{}';
        const lines = entries.map(([k, v]) => {
            const key = /^[a-zA-Z_$][a-zA-Z0-9_$]*$/.test(k) ? k : `'${k}'`;
            return `${innerPad}${key}: ${formatValue(v, indent + 2)}`;
        });
        return `{\n${lines.join(',\n')}\n${pad}}`;
    }

    return String(value);
}

function main() {
    const contractJsonPath = join(__dirname, JSON_ARTIFACT_PATH);
    const outputPath = join(__dirname, OUTPUT_PATH);

    const contractJson = JSON.parse(readFileSync(contractJsonPath, 'utf-8'));
    const abi: unknown[] = contractJson.abi;

    const entries = abi.map((entry) => `  ${formatValue(entry, 2)}`);
    const abiBody = entries.join(',\n');

    const output = `/**
 * ChannelHub contract ABI
 * Generated from contracts/src/ChannelHub.sol
 */

import { Abi } from 'abitype';

export const ChannelHubAbi = [
${abiBody}
] as const satisfies Abi;
`;

    writeFileSync(outputPath, output, 'utf-8');
    console.log(`Generated ${outputPath}`);
}

main();
