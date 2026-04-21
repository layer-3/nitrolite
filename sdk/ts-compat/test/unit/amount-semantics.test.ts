import { NitroliteClient } from '../../src/index.js';

const CURRENT_CHAIN = 84532n;
const CURRENT_TOKEN = '0x0000000000000000000000000000000000000b01';

function makeCompatClient() {
    const client = Object.create(NitroliteClient.prototype) as NitroliteClient & Record<string, unknown>;
    Object.assign(client, {
        innerClient: {
            getAssets: async () => ([
                {
                    name: 'Yellow USD',
                    symbol: 'yusd',
                    decimals: 6,
                    suggestedBlockchainId: CURRENT_CHAIN,
                    tokens: [
                        {
                            name: 'Yellow USD',
                            symbol: 'YUSD',
                            address: CURRENT_TOKEN,
                            blockchainId: CURRENT_CHAIN,
                            decimals: 8,
                        },
                    ],
                },
            ]),
        },
        userAddress: '0x00000000000000000000000000000000000000a1',
        walletClient: {
            chain: {
                rpcUrls: {
                    public: { http: ['https://rpc.base-sepolia.example'] },
                    default: { http: ['https://rpc.base-sepolia.example'] },
                },
            },
        },
        assetsByChainAndToken: new Map(),
        assetsByToken: new Map(),
        assetsBySymbol: new Map(),
        _chainId: CURRENT_CHAIN,
        _lastChannels: [],
        _lastAppSessionsListError: null,
        _lastAppSessionsListErrorLogged: null,
        _blockchains: [],
        _lockingTokenDecimals: new Map(),
        _blockchainRPCs: { 84532: 'https://rpc.base-sepolia.example' },
        _publicClients: new Map(),
    });
    return client;
}

describe('compat amount semantics', () => {
    it('formats and parses raw token amounts using token decimals', async () => {
        const client = makeCompatClient();
        await client.refreshAssets();

        await expect(client.formatAmount(CURRENT_TOKEN, 123456789n)).resolves.toBe('1.23456789');
        await expect(client.parseAmount(CURRENT_TOKEN, '1.23456789')).resolves.toBe(123456789n);
    });
});
