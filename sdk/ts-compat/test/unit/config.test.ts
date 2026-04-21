import * as compat from '../../src/index.js';
import { blockchainRPCsFromEnv, buildClientOptions } from '../../src/config.js';

type SDKConfig = {
    url: string;
    blockchainRPCs?: Map<bigint, string>;
};

describe('compat root barrel config helpers', () => {
    const originalRPCsEnv = process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS;

    afterEach(() => {
        if (originalRPCsEnv === undefined) {
            delete process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS;
            return;
        }
        process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS = originalRPCsEnv;
    });

    it('resolves config helpers from the root barrel', () => {
        expect(compat.buildClientOptions).toBe(buildClientOptions);
        expect(compat.blockchainRPCsFromEnv).toBe(blockchainRPCsFromEnv);
        expect(typeof compat.NitroliteClient).toBe('function');
    });

    it('returns an empty mapping when NEXT_PUBLIC_BLOCKCHAIN_RPCS is unset', () => {
        delete process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS;

        expect(blockchainRPCsFromEnv()).toEqual({});
    });

    it('parses multiple RPC mappings and applies them through SDK options', () => {
        process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS =
            '11155111:https://rpc.sepolia.example,84532:https://base-sepolia.example';

        const mappings = blockchainRPCsFromEnv();
        expect(mappings).toEqual({
            84532: 'https://base-sepolia.example',
            11155111: 'https://rpc.sepolia.example',
        });

        const config: SDKConfig = { url: 'wss://clearnode.example/ws' };
        const opts = buildClientOptions({
            wsURL: config.url,
            blockchainRPCs: mappings,
        });
        expect(opts).toHaveLength(2);
        for (const opt of opts) {
            opt(config);
        }

        expect(config.blockchainRPCs).toBeInstanceOf(Map);
        expect(config.blockchainRPCs?.get(11155111n)).toBe('https://rpc.sepolia.example');
        expect(config.blockchainRPCs?.get(84532n)).toBe('https://base-sepolia.example');
    });
});
