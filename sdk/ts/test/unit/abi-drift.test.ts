import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { ChannelHubAbi } from '../../src/blockchain/evm/channel_hub_abi.js';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, '../../../..');

type AbiEntry = {
    type: string;
    name?: string;
    inputs?: Array<{ type: string }>;
    outputs?: Array<{ type: string }>;
    stateMutability?: string;
};

function signature(entry: AbiEntry): string {
    const inputs = (entry.inputs ?? []).map((input) => input.type).join(',');
    const outputs = (entry.outputs ?? []).map((output) => output.type).join(',');
    return `${entry.name}(${inputs}) -> (${outputs}) ${entry.stateMutability ?? ''}`;
}

function functionSignatures(abi: readonly AbiEntry[]) {
    return new Map(
        abi
            .filter((entry) => entry.type === 'function' && entry.name)
            .map((entry) => [entry.name as string, signature(entry)])
    );
}

describe('contract ABI drift guards', () => {
    it('keeps checked-in ChannelHub ABI aligned with Foundry artifact for SDK-consumed functions', () => {
        const artifact = JSON.parse(
            fs.readFileSync(
                path.join(repoRoot, 'contracts/out/ChannelHub.sol/ChannelHub.json'),
                'utf8'
            )
        );
        const artifactSigs = functionSignatures(artifact.abi);
        const sdkSigs = functionSignatures(ChannelHubAbi as readonly AbiEntry[]);

        const consumedFunctions = [
            'VERSION',
            'createChannel',
            'depositToChannel',
            'withdrawFromChannel',
            'checkpointChannel',
            'closeChannel',
            'getChannelData',
            'getNodeValidator',
            'isNodeValidatorActive',
            'registerSessionKey',
            'unregisterSessionKey',
        ].filter((name) => artifactSigs.has(name) || sdkSigs.has(name));

        const diffs = consumedFunctions
            .map((name) => ({
                name,
                artifact: artifactSigs.get(name),
                sdk: sdkSigs.get(name),
            }))
            .filter(({ artifact: artifactSig, sdk: sdkSig }) => artifactSig !== sdkSig);

        expect(diffs).toEqual([]);
    });

    it('reports adversarial function signature changes with function names', () => {
        const artifactSigs = functionSignatures([
            {
                type: 'function',
                name: 'getNodeValidator',
                inputs: [{ type: 'address' }, { type: 'uint8' }],
                outputs: [{ type: 'address' }],
                stateMutability: 'view',
            },
        ]);
        const sdkSigs = functionSignatures([
            {
                type: 'function',
                name: 'getNodeValidator',
                inputs: [{ type: 'uint8' }],
                outputs: [{ type: 'address' }],
                stateMutability: 'view',
            },
        ]);

        expect({
            name: 'getNodeValidator',
            artifact: artifactSigs.get('getNodeValidator'),
            sdk: sdkSigs.get('getNodeValidator'),
        }).toEqual({
            name: 'getNodeValidator',
            artifact: 'getNodeValidator(address,uint8) -> (address) view',
            sdk: 'getNodeValidator(uint8) -> (address) view',
        });
    });
});
