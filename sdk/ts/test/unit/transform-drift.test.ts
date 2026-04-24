import { Decimal } from 'decimal.js';

import { transformAppSessionInfo, transformAssets, transformNodeConfig } from '../../src/utils.js';

describe('Clearnode response transform drift guards', () => {
    it('maps current get_app_sessions app_definition shape to SDK appDefinition', () => {
        const session = transformAppSessionInfo({
            app_session_id: '0xsession',
            app_definition: {
                application_id: 'store-v1',
                participants: [
                    {
                        wallet_address: '0x1111111111111111111111111111111111111111',
                        signature_weight: 1,
                    },
                    {
                        wallet_address: '0x2222222222222222222222222222222222222222',
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
                    participant: '0x1111111111111111111111111111111111111111',
                    asset: 'YUSD',
                    amount: '1.25',
                },
            ],
        });

        expect(session).toEqual({
            appSessionId: '0xsession',
            appDefinition: {
                applicationId: 'store-v1',
                participants: [
                    {
                        walletAddress: '0x1111111111111111111111111111111111111111',
                        signatureWeight: 1,
                    },
                    {
                        walletAddress: '0x2222222222222222222222222222222222222222',
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
                    participant: '0x1111111111111111111111111111111111111111',
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
});
