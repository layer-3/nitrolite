import { Decimal } from 'decimal.js';
import { jest } from '@jest/globals';

import { NitroliteClient } from '../../src/client.js';

const wallet = '0x1111111111111111111111111111111111111111';
const friend = '0x2222222222222222222222222222222222222222';

function makeClient(sessions: any[]) {
    const client = Object.create(NitroliteClient.prototype) as any;
    client.userAddress = wallet;
    client.innerClient = {
        getAppSessions: jest.fn().mockResolvedValue({ sessions }),
        getConfig: jest.fn(),
    };
    client.assetsBySymbol = new Map([
        ['yusd', { decimals: 6 }],
        ['yellow', { decimals: 18 }],
    ]);
    client._lastAppSessionsListError = null;
    client._lastAppSessionsListErrorLogged = null;

    return client;
}

describe('NitroliteClient getAppSessionsList compat mapping', () => {
    let infoSpy: jest.SpyInstance;
    let warnSpy: jest.SpyInstance;

    beforeEach(() => {
        infoSpy = jest.spyOn(console, 'info').mockImplementation(() => {});
        warnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        infoSpy.mockRestore();
        warnSpy.mockRestore();
    });

    it('maps the current v1 appDefinition session shape', async () => {
        const client = makeClient([
            {
                appSessionId: '0xsession',
                appDefinition: {
                    participants: [
                        { walletAddress: wallet, signatureWeight: 1 },
                        { walletAddress: friend, signatureWeight: 2 },
                    ],
                    quorum: 3,
                    nonce: 42n,
                },
                isClosed: false,
                version: 7n,
                allocations: [
                    { participant: wallet, asset: 'YUSD', amount: new Decimal('1.25') },
                    { participant: friend, asset: 'YELLOW', amount: new Decimal('2') },
                ],
                sessionData: '{"intent":"purchase"}',
            },
        ]);

        const sessions = await client.getAppSessionsList();

        expect(client.innerClient.getAppSessions).toHaveBeenCalledWith({
            wallet: wallet.toLowerCase(),
        });
        expect(sessions).toEqual([
            {
                app_session_id: '0xsession',
                nonce: 42,
                participants: [wallet, friend],
                protocol: '',
                quorum: 3,
                status: 'open',
                version: 7,
                weights: [1, 2],
                allocations: [
                    { participant: wallet, asset: 'YUSD', amount: '1250000' },
                    { participant: friend, asset: 'YELLOW', amount: '2000000000000000000' },
                ],
                sessionData: '{"intent":"purchase"}',
            },
        ]);
    });

    it('keeps the legacy flat session shape fallback', async () => {
        const client = makeClient([
            {
                appSessionId: '0xlegacy',
                participants: [
                    { walletAddress: wallet, signatureWeight: 50 },
                    { walletAddress: friend, signatureWeight: 50 },
                ],
                quorum: 100,
                nonce: 99n,
                isClosed: true,
                version: 4n,
                allocations: [],
                sessionData: '{"legacy":true}',
            },
        ]);

        const sessions = await client.getAppSessionsList(wallet, 'any');

        expect(client.innerClient.getAppSessions).toHaveBeenCalledWith({
            wallet: wallet.toLowerCase(),
        });
        expect(sessions).toEqual([
            {
                app_session_id: '0xlegacy',
                nonce: 99,
                participants: [wallet, friend],
                protocol: '',
                quorum: 100,
                status: 'closed',
                version: 4,
                weights: [50, 50],
                allocations: [],
                sessionData: '{"legacy":true}',
            },
        ]);
    });

    it('maps an empty app session list without requiring legacy fields', async () => {
        const client = makeClient([]);

        await expect(client.getAppSessionsList()).resolves.toEqual([]);
        expect(client.innerClient.getAppSessions).toHaveBeenCalledWith({
            wallet: wallet.toLowerCase(),
        });
    });

    it('passes through current SDK camelCase getConfig shape', async () => {
        const currentConfig = {
            nodeAddress: wallet,
            nodeVersion: 'test-node',
            supportedSigValidators: [0, 1],
            blockchains: [
                {
                    name: 'Sepolia',
                    id: 11155111n,
                    channelHubAddress: '0x3333333333333333333333333333333333333333',
                    lockingContractAddress: '0x4444444444444444444444444444444444444444',
                    blockStep: 0n,
                },
            ],
        };
        const client = makeClient([]);
        client.innerClient.getConfig.mockResolvedValue(currentConfig);

        await expect(client.getConfig()).resolves.toBe(currentConfig);
    });

    it('documents snake_case getConfig as pass-through, not normalized compat mapping', async () => {
        const rawConfig = {
            node_address: wallet,
            node_version: 'raw-node',
            supported_sig_validators: [0, 1],
            blockchains: [
                {
                    name: 'Sepolia',
                    blockchain_id: '11155111',
                    channel_hub_address: '0x3333333333333333333333333333333333333333',
                    locking_contract_address: '0x4444444444444444444444444444444444444444',
                },
            ],
        };
        const client = makeClient([]);
        client.innerClient.getConfig.mockResolvedValue(rawConfig);

        await expect(client.getConfig()).resolves.toBe(rawConfig);
    });
});
