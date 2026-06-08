import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { ChannelHubAbi } from '../../src/blockchain/evm/channel_hub_abi.js';
import { Erc20Abi } from '../../src/blockchain/evm/erc20_abi.js';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, '../../../..');

type AbiEntry = {
    type: string;
    name?: string;
    inputs?: AbiParam[];
    outputs?: AbiParam[];
    stateMutability?: string;
};

type AbiParam = {
    type: string;
    components?: AbiParam[];
};

type FunctionDiff = {
    contract: string;
    name: string;
    artifact?: string;
    sdk?: string;
};

function canonicalType(param: AbiParam): string {
    if (!param.components?.length) return param.type;
    return `${param.type}<${param.components.map(canonicalType).join(',')}>`;
}

function signature(entry: AbiEntry): string {
    const inputs = (entry.inputs ?? []).map(canonicalType).join(',');
    const outputs = (entry.outputs ?? []).map(canonicalType).join(',');
    const mutability = entry.stateMutability ? ` ${entry.stateMutability}` : '';
    return `${entry.name}(${inputs}) -> (${outputs})${mutability}`;
}

function functionSignatures(abi: readonly AbiEntry[]): Map<string, string> {
    const signaturesByName = new Map<string, string[]>();

    for (const entry of abi) {
        if (entry.type !== 'function' || !entry.name) continue;

        const signatures = signaturesByName.get(entry.name) ?? [];
        signatures.push(signature(entry));
        signaturesByName.set(entry.name, signatures);
    }

    return new Map(
        [...signaturesByName].map(([name, signatures]) => [name, signatures.sort().join('\n')])
    );
}

function loadArtifact(relativePath: string): readonly AbiEntry[] {
    const artifactPath = path.join(repoRoot, relativePath);
    if (!fs.existsSync(artifactPath)) {
        throw new Error(`ABI artifact not found: ${relativePath}. Run: cd contracts && forge build`);
    }
    const artifact = JSON.parse(fs.readFileSync(artifactPath, 'utf8'));
    return artifact.abi;
}

function sortedSignatureEntries(signatures: ReadonlyMap<string, string>): [string, string][] {
    return [...signatures].sort(([left], [right]) => left.localeCompare(right));
}

function diffConsumedFunctions(
    contract: string,
    artifactAbi: readonly AbiEntry[],
    sdkAbi: readonly AbiEntry[],
    consumedFunctions: readonly string[]
): FunctionDiff[] {
    const artifactSigs = functionSignatures(artifactAbi);
    const sdkSigs = functionSignatures(sdkAbi);

    return consumedFunctions
        .map((name) => ({
            contract,
            name,
            artifact: artifactSigs.get(name),
            sdk: sdkSigs.get(name),
        }))
        .filter(
            ({ artifact: artifactSig, sdk: sdkSig }) =>
                artifactSig !== sdkSig || artifactSig === undefined || sdkSig === undefined
        );
}

describe('contract ABI drift guards', () => {
    it('keeps checked-in ChannelHub ABI aligned with Foundry artifact for every artifact function', () => {
        const artifactSigs = functionSignatures(
            loadArtifact('contracts/out/ChannelHub.sol/ChannelHub.json')
        );
        const sdkSigs = functionSignatures(ChannelHubAbi as readonly AbiEntry[]);

        expect(sortedSignatureEntries(sdkSigs)).toEqual(sortedSignatureEntries(artifactSigs));
    });

    it('keeps SDK-consumed ChannelHub functions aligned with Foundry artifact', () => {
        const consumedFunctions = [
            'VERSION',
            'createChannel',
            'depositToNode',
            'withdrawFromNode',
            'depositToChannel',
            'withdrawFromChannel',
            'checkpointChannel',
            'challengeChannel',
            'closeChannel',
            'getChannelData',
            'getNodeBalance',
            'getNodeValidator',
            'getOpenChannels',
        ];

        expect(
            diffConsumedFunctions(
                'ChannelHub',
                loadArtifact('contracts/out/ChannelHub.sol/ChannelHub.json'),
                ChannelHubAbi as readonly AbiEntry[],
                consumedFunctions
            )
        ).toEqual([]);
    });

    it('keeps checked-in ERC20 ABI aligned with the Foundry artifact for SDK-consumed functions', () => {
        const consumedFunctions = [
            'allowance',
            'approve',
            'balanceOf',
            'decimals',
            'name',
            'symbol',
            'totalSupply',
            'transfer',
            'transferFrom',
        ];

        expect(
            diffConsumedFunctions(
                'ERC20',
                loadArtifact('contracts/out/ERC20.sol/ERC20.default.json'),
                Erc20Abi as readonly AbiEntry[],
                consumedFunctions
            )
        ).toEqual([]);
    });

    it('reports adversarial ChannelHub function signature changes with contract and function names', () => {
        expect(
            diffConsumedFunctions(
                'ChannelHub',
                [
                    {
                        type: 'function',
                        name: 'getNodeValidator',
                        inputs: [{ type: 'address' }, { type: 'uint8' }],
                        outputs: [{ type: 'address' }],
                        stateMutability: 'view',
                    },
                ],
                [
                    {
                        type: 'function',
                        name: 'getNodeValidator',
                        inputs: [{ type: 'uint8' }],
                        outputs: [{ type: 'address' }],
                        stateMutability: 'view',
                    },
                ],
                ['getNodeValidator']
            )
        ).toEqual([
            {
                contract: 'ChannelHub',
                name: 'getNodeValidator',
                artifact: 'getNodeValidator(address,uint8) -> (address) view',
                sdk: 'getNodeValidator(uint8) -> (address) view',
            },
        ]);
    });

    it('reports adversarial ERC20 missing consumed functions', () => {
        expect(
            diffConsumedFunctions(
                'ERC20',
                [
                    {
                        type: 'function',
                        name: 'approve',
                        inputs: [{ type: 'address' }, { type: 'uint256' }],
                        outputs: [{ type: 'bool' }],
                        stateMutability: 'nonpayable',
                    },
                ],
                [],
                ['approve']
            )
        ).toEqual([
            {
                contract: 'ERC20',
                name: 'approve',
                artifact: 'approve(address,uint256) -> (bool) nonpayable',
                sdk: undefined,
            },
        ]);
    });

});
