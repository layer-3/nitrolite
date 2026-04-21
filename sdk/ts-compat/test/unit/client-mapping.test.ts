import { jest } from '@jest/globals';
import { Decimal } from 'decimal.js';

import { NitroliteClient, RPCAppStateIntent } from '../../src/index.js';

const USER = '0x00000000000000000000000000000000000000a1';
const CURRENT_CHAIN = 84532n;
const CURRENT_TOKEN = '0x0000000000000000000000000000000000000b01';
const OTHER_CHAIN_TOKEN = '0x0000000000000000000000000000000000000b02';

function makeAssetsFixture() {
    return [
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
                {
                    name: 'Yellow USD',
                    symbol: 'YUSD',
                    address: OTHER_CHAIN_TOKEN,
                    blockchainId: 11155111n,
                    decimals: 6,
                },
            ],
        },
    ];
}

function makeInnerClient(overrides: Record<string, unknown> = {}) {
    return {
        getAssets: jest.fn().mockResolvedValue(makeAssetsFixture()),
        getBalances: jest.fn().mockResolvedValue([]),
        getChannels: jest.fn().mockResolvedValue({ channels: [] }),
        getLatestState: jest.fn(),
        getHomeChannel: jest.fn(),
        getAppSessions: jest.fn().mockResolvedValue({ sessions: [] }),
        submitAppState: jest.fn().mockResolvedValue(undefined),
        submitAppSessionDeposit: jest.fn().mockResolvedValue('0xdeposit'),
        transfer: jest.fn().mockResolvedValue(undefined),
        setHomeBlockchain: jest.fn().mockResolvedValue(undefined),
        deposit: jest.fn().mockResolvedValue(undefined),
        withdraw: jest.fn().mockResolvedValue(undefined),
        closeHomeChannel: jest.fn().mockResolvedValue(undefined),
        checkpoint: jest.fn().mockResolvedValue('0xcheckpoint'),
        approveToken: jest.fn().mockResolvedValue('0xapprove'),
        checkTokenAllowance: jest.fn().mockResolvedValue(77n),
        getOnChainBalance: jest.fn().mockResolvedValue(new Decimal('5')),
        ping: jest.fn().mockResolvedValue(undefined),
        close: jest.fn().mockResolvedValue(undefined),
        waitForClose: jest.fn().mockResolvedValue(undefined),
        acknowledge: jest.fn().mockResolvedValue(undefined),
        getBlockchains: jest.fn().mockResolvedValue([
            {
                id: CURRENT_CHAIN,
                name: 'Base Sepolia',
                channelHubAddress: '0x0000000000000000000000000000000000000c01',
                blockStep: 0n,
            },
        ]),
        ...overrides,
    };
}

function makeCompatClient(innerOverrides: Record<string, unknown> = {}) {
    const innerClient = makeInnerClient(innerOverrides);
    const client = Object.create(NitroliteClient.prototype) as NitroliteClient & Record<string, unknown>;
    Object.assign(client, {
        innerClient,
        userAddress: USER,
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
        _blockchains: [
            {
                id: CURRENT_CHAIN,
                name: 'Base Sepolia',
                channelHubAddress: '0x0000000000000000000000000000000000000c01',
                blockStep: 0n,
            },
        ],
        _lockingTokenDecimals: new Map(),
        _blockchainRPCs: { 84532: 'https://rpc.base-sepolia.example' },
        _publicClients: new Map(),
    });
    return { client, innerClient };
}

describe('NitroliteClient compat mappings', () => {
    it('stores token decimals for token-facing helpers and asset decimals for ledger balance conversions', async () => {
        const { client, innerClient } = makeCompatClient({
            getBalances: jest.fn().mockResolvedValue([{ asset: 'yusd', balance: new Decimal('1.23') }]),
            getChannels: jest.fn().mockResolvedValue({
                channels: [
                    {
                        channelId: 'channel-1',
                        userWallet: USER,
                        status: 1,
                        asset: 'yusd',
                        tokenAddress: CURRENT_TOKEN,
                        blockchainId: CURRENT_CHAIN,
                        challengeDuration: 86400,
                        nonce: 1n,
                        stateVersion: 2n,
                    },
                ],
            }),
            getLatestState: jest.fn().mockResolvedValue({
                homeLedger: { userBalance: new Decimal('1.23') },
            }),
        });

        await client.refreshAssets();

        expect(await client.getTokenDecimals(CURRENT_TOKEN)).toBe(8);
        expect(await client.getBalances()).toEqual([{ asset: 'yusd', amount: '1230000' }]);
        expect(await client.getChannels()).toEqual([
            expect.objectContaining({
                channel_id: 'channel-1',
                amount: 123000000n,
                chain_id: 84532,
            }),
        ]);
        expect(innerClient.getLatestState).toHaveBeenCalled();
    });

    it('keeps app-session allocation amounts in human decimals for list/read/update flows', async () => {
        const { client, innerClient } = makeCompatClient({
            getAppSessions: jest.fn()
                .mockResolvedValueOnce({
                    sessions: [
                        {
                            appSessionId: 'session-1',
                            nonce: 1n,
                            participants: [{ walletAddress: USER, signatureWeight: 1 }],
                            quorum: 1,
                            isClosed: false,
                            version: 4n,
                            allocations: [{ participant: USER, asset: 'yusd', amount: new Decimal('0.01') }],
                            sessionData: '{"turn":1}',
                        },
                    ],
                })
                .mockResolvedValueOnce({
                    sessions: [
                        {
                            appSessionId: 'session-2',
                            version: 6n,
                            allocations: [],
                        },
                    ],
                })
                .mockResolvedValueOnce({
                    sessions: [
                        {
                            appSessionId: 'session-3',
                            version: 8n,
                            allocations: [],
                        },
                    ],
                }),
        });

        await client.refreshAssets();

        await expect(client.getAppSessionsList()).resolves.toEqual([
            expect.objectContaining({
                app_session_id: 'session-1',
                allocations: [{ participant: USER, asset: 'yusd', amount: '0.01' }],
            }),
        ]);

        await client.closeAppSession('session-2', [{ participant: USER, asset: 'yusd', amount: '1.5' }], ['0xclose']);
        expect(innerClient.submitAppState).toHaveBeenCalledWith(
            expect.objectContaining({
                appSessionId: 'session-2',
                version: 7n,
                allocations: [
                    expect.objectContaining({
                        asset: 'yusd',
                        amount: expect.any(Decimal),
                    }),
                ],
            }),
            ['0xclose'],
        );
        expect(innerClient.submitAppState.mock.calls[0][0].allocations[0].amount.toString()).toBe('1.5');

        await client.submitAppState({
            app_session_id: 'session-3',
            intent: RPCAppStateIntent.Operate,
            version: 9,
            allocations: [{ participant: USER, asset: 'yusd', amount: '0.25' }],
            quorum_sigs: ['0xoperate'],
            session_data: '{"turn":2}',
        });
        expect(innerClient.submitAppState.mock.calls[1][0].allocations[0].amount.toString()).toBe('0.25');
    });

    it('uses token decimals for raw-unit transfer/approval/balance helpers', async () => {
        const { client, innerClient } = makeCompatClient();

        await client.refreshAssets();

        await client.transfer('0x00000000000000000000000000000000000000d1', [
            { asset: 'yusd', amount: '500000000' },
        ]);
        expect(innerClient.transfer).toHaveBeenCalledWith(
            '0x00000000000000000000000000000000000000d1',
            'yusd',
            expect.any(Decimal),
        );
        expect(innerClient.transfer.mock.calls[0][2].toString()).toBe('5');

        await expect(client.approveTokens(CURRENT_TOKEN, 250000000n)).resolves.toBe('0xapprove');
        expect(innerClient.approveToken).toHaveBeenCalledWith(CURRENT_CHAIN, 'yusd', expect.any(Decimal));
        expect(innerClient.approveToken.mock.calls[0][2].toString()).toBe('2.5');

        await expect(client.getTokenAllowance(CURRENT_TOKEN)).resolves.toBe(77n);
        expect(innerClient.checkTokenAllowance).toHaveBeenCalledWith(CURRENT_CHAIN, CURRENT_TOKEN, USER);

        await expect(client.getTokenBalance(CURRENT_TOKEN)).resolves.toBe(500000000n);
        expect(innerClient.getOnChainBalance).toHaveBeenCalledWith(CURRENT_CHAIN, 'yusd', USER);
    });

    it('exposes token-and-chain specific display data and assets list entries', async () => {
        const { client } = makeCompatClient();

        await client.refreshAssets();

        await expect(client.resolveAssetDisplay(CURRENT_TOKEN)).resolves.toEqual({ symbol: 'yusd', decimals: 8 });
        await expect(client.resolveAssetDisplay(OTHER_CHAIN_TOKEN, 11155111)).resolves.toEqual({ symbol: 'yusd', decimals: 6 });
        await expect(client.getAssetsList()).resolves.toEqual([
            { token: CURRENT_TOKEN, chainId: 84532, symbol: 'yusd', decimals: 8 },
            { token: OTHER_CHAIN_TOKEN, chainId: 11155111, symbol: 'yusd', decimals: 6 },
        ]);
    });

    it('keeps unsupported legacy methods honest and getOpenChannels delegates to the current chain hub', async () => {
        const { client } = makeCompatClient();
        const readContract = jest.fn().mockResolvedValue(['0xabc', '0xdef']);
        (client as unknown as Record<string, unknown>).getReadClientForChain = jest.fn().mockResolvedValue({ readContract });

        await expect(client.getOpenChannels()).resolves.toEqual(['0xabc', '0xdef']);

        await expect(client.createChannel()).rejects.toThrow('deposit(tokenAddress, amount)');
        await expect(client.checkpointChannel({})).rejects.toThrow('client.innerClient.checkpoint(asset)');
        await expect(client.getAccountBalance(CURRENT_TOKEN)).rejects.toThrow('Use getBalances()');
        await expect(client.getChannelBalance('channel-1', CURRENT_TOKEN)).rejects.toThrow('Use getChannelData(channelId)');
    });
});
