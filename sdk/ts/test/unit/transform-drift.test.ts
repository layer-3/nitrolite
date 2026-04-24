import { Decimal } from 'decimal.js';
import { jest } from '@jest/globals';

import { Client } from '../../src/client.js';
import {
    transformAppSessionInfo,
    transformAssets,
    transformNodeConfig,
} from '../../src/utils.js';
import {
    transformAppSessionKeyState,
    transformChannelSessionKeyState,
} from '../../src/session_key_state_transforms.js';

const userAddress = '0x1111111111111111111111111111111111111111';
const sessionKeyAddress = '0x2222222222222222222222222222222222222222';

const appSessionRaw = {
    app_session_id: '0xsession',
    app_definition: {
        application_id: 'store-v1',
        participants: [
            {
                wallet_address: userAddress,
                signature_weight: 1,
            },
            {
                wallet_address: sessionKeyAddress,
                signature_weight: 1,
            },
        ],
        quorum: 2,
        nonce: '123',
    },
    status: 'open',
    session_data: '{"intent":"purchase"}',
    version: '4',
    allocations: [
        {
            participant: userAddress,
            asset: 'YUSD',
            amount: '1.25',
        },
    ],
};

const channelKeyStateRaw = {
    user_address: userAddress,
    session_key: sessionKeyAddress,
    version: '7',
    assets: ['YUSD'],
    expires_at: '1739812234',
    user_sig: '0xabc123',
};

const appSessionKeyStateRaw = {
    user_address: userAddress,
    session_key: sessionKeyAddress,
    version: '8',
    application_ids: ['0x00000000000000000000000000000000000000000000000000000000000000a1'],
    app_session_ids: ['0x00000000000000000000000000000000000000000000000000000000000000b1'],
    expires_at: '1739812234',
    user_sig: '0xdef456',
};

describe('Clearnode response transform drift guards', () => {
    it('maps current get_app_sessions app_definition shape to SDK appDefinition', () => {
        const session = transformAppSessionInfo(appSessionRaw);

        expect(session).toEqual({
            appSessionId: '0xsession',
            appDefinition: {
                applicationId: 'store-v1',
                participants: [
                    {
                        walletAddress: userAddress,
                        signatureWeight: 1,
                    },
                    {
                        walletAddress: sessionKeyAddress,
                        signatureWeight: 1,
                    },
                ],
                quorum: 2,
                nonce: 123n,
            },
            isClosed: false,
            sessionData: '{"intent":"purchase"}',
            version: 4n,
            allocations: [
                {
                    participant: userAddress,
                    asset: 'YUSD',
                    amount: new Decimal('1.25'),
                },
            ],
        });
    });

    it('rejects app sessions missing the required app_definition payload', () => {
        expect(() =>
            transformAppSessionInfo({
                app_session_id: '0xsession',
                status: 'open',
                session_data: '',
                version: '1',
                allocations: [],
            })
        ).toThrow('Invalid app definition: missing required fields');
    });

    it('rejects app sessions missing required allocation fields', () => {
        expect(() =>
            transformAppSessionInfo({
                ...appSessionRaw,
                allocations: [
                    {
                        participant: userAddress,
                        asset: 'YUSD',
                    },
                ],
            })
        ).toThrow('Invalid app session allocation[0]: missing required string field amount');
    });

    it('maps get_config supported_sig_validators from array and base64 forms', () => {
        const base = {
            node_address: '0x1111111111111111111111111111111111111111' as const,
            node_version: 'test',
            blockchains: [
                {
                    name: 'Sepolia',
                    blockchain_id: '11155111',
                    channel_hub_address: '0x2222222222222222222222222222222222222222',
                    locking_contract_address: '0x3333333333333333333333333333333333333333',
                },
            ],
        };

        expect(
            transformNodeConfig({
                ...base,
                supported_sig_validators: [0, 1],
            }).supportedSigValidators
        ).toEqual([0, 1]);

        expect(
            transformNodeConfig({
                ...base,
                supported_sig_validators: 'AAE=',
            } as any).supportedSigValidators
        ).toEqual([0, 1]);

        expect(transformNodeConfig(base as any).supportedSigValidators).toEqual([]);
    });

    it('maps get_assets symbols, decimals, suggested chain, and token chains', () => {
        expect(
            transformAssets([
                {
                    name: 'Yellow USD',
                    symbol: 'YUSD',
                    decimals: 6,
                    suggested_blockchain_id: '11155111',
                    tokens: [
                        {
                            name: 'Yellow USD',
                            symbol: 'YUSD',
                            address: '0x4444444444444444444444444444444444444444',
                            blockchain_id: '11155111',
                            decimals: 6,
                        },
                    ],
                },
            ])
        ).toEqual([
            {
                name: 'Yellow USD',
                symbol: 'YUSD',
                decimals: 6,
                suggestedBlockchainId: 11155111n,
                tokens: [
                    {
                        name: 'Yellow USD',
                        symbol: 'YUSD',
                        address: '0x4444444444444444444444444444444444444444',
                        blockchainId: 11155111n,
                        decimals: 6,
                    },
                ],
            },
        ]);
    });

    it('validates channel key-state and app-session key-state fixtures', () => {
        expect(transformChannelSessionKeyState(channelKeyStateRaw)).toEqual(channelKeyStateRaw);
        expect(transformAppSessionKeyState(appSessionKeyStateRaw)).toEqual(appSessionKeyStateRaw);
    });

    it('rejects malformed key-state fixtures with clear errors', () => {
        expect(() =>
            transformChannelSessionKeyState(
                {
                    ...channelKeyStateRaw,
                    user_sig: undefined,
                },
                'channel session key state[0]'
            )
        ).toThrow('Invalid channel session key state[0]: missing required string field user_sig');

        expect(() =>
            transformAppSessionKeyState(
                {
                    ...appSessionKeyStateRaw,
                    app_session_ids: 'not-an-array',
                },
                'app session key state[0]'
            )
        ).toThrow('Invalid app session key state[0]: expected app_session_ids to be string[]');

        expect(() =>
            transformAppSessionKeyState(
                {
                    ...appSessionKeyStateRaw,
                    application_ids: undefined,
                    applicationIds: appSessionKeyStateRaw.application_ids,
                },
                'app session key state[0]'
            )
        ).toThrow('Invalid app session key state[0]: expected application_ids to be string[]');
    });

    it('maps high-level client app-session and key-state responses through transform paths', async () => {
        const rpcClient = {
            appSessionsV1GetAppSessions: jest.fn(async () => ({
                app_sessions: [appSessionRaw],
                metadata: {
                    page: 1,
                    per_page: 10,
                    total_count: 1,
                    page_count: 1,
                },
            })),
            channelsV1GetLastKeyStates: jest.fn(async () => ({
                states: [channelKeyStateRaw],
            })),
            appSessionsV1GetLastKeyStates: jest.fn(async () => ({
                states: [appSessionKeyStateRaw],
            })),
        };
        const clientLike = { rpcClient };

        const sessionsResult = await (Client.prototype.getAppSessions as any).call(clientLike, {
            wallet: userAddress,
            page: 1,
            pageSize: 10,
        });
        const channelKeyStates = await (Client.prototype.getLastChannelKeyStates as any).call(
            clientLike,
            userAddress
        );
        const appSessionKeyStates = await (Client.prototype.getLastKeyStates as any).call(
            clientLike,
            userAddress
        );

        expect(sessionsResult.sessions).toHaveLength(1);
        expect(sessionsResult.metadata).toEqual({
            page: 1,
            perPage: 10,
            totalCount: 1,
            pageCount: 1,
        });
        expect(channelKeyStates).toEqual([channelKeyStateRaw]);
        expect(appSessionKeyStates).toEqual([appSessionKeyStateRaw]);
        expect(rpcClient.appSessionsV1GetAppSessions).toHaveBeenCalledWith({
            app_session_id: undefined,
            participant: userAddress,
            status: undefined,
            pagination: {
                offset: 0,
                limit: 10,
            },
        });
    });

    it('maps high-level client empty app-session responses', async () => {
        const clientLike = {
            rpcClient: {
                appSessionsV1GetAppSessions: jest.fn(async () => ({
                    app_sessions: [],
                    metadata: {
                        page: 1,
                        per_page: 10,
                        total_count: 0,
                        page_count: 0,
                    },
                })),
            },
        };

        const result = await (Client.prototype.getAppSessions as any).call(clientLike, {
            wallet: userAddress,
            page: 1,
            pageSize: 10,
        });

        expect(result).toEqual({
            sessions: [],
            metadata: {
                page: 1,
                perPage: 10,
                totalCount: 0,
                pageCount: 0,
            },
        });
    });
});
