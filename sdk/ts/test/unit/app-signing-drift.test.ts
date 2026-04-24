import { Decimal } from 'decimal.js';

import {
    AppStateUpdateIntent,
    generateAppSessionIDV1,
    packAppStateUpdateV1,
    packAppSessionKeyStateV1,
    packCreateAppSessionRequestV1,
    type AppDefinitionV1,
    type AppSessionKeyStateV1,
} from '../../src/app/index.js';

// Regenerate expected hashes with:
// go run ./scripts/drift/generate-app-signing-vectors.go

const user = '0x1111111111111111111111111111111111111111';
const app = '0x2222222222222222222222222222222222222222';
const maxUint64 = 18446744073709551615n;

const definition: AppDefinitionV1 = {
    applicationId: 'store-v1',
    participants: [
        { walletAddress: user, signatureWeight: 1 },
        { walletAddress: app, signatureWeight: 1 },
    ],
    quorum: 2,
    nonce: 123456789n,
};

const sessionKeyState: AppSessionKeyStateV1 = {
    user_address: user,
    session_key: app,
    version: '1',
    application_ids: ['0x00000000000000000000000000000000000000000000000000000000000000a1'],
    app_session_ids: ['0x00000000000000000000000000000000000000000000000000000000000000b1'],
    expires_at: '1739812234',
    user_sig: '0xSig',
};

describe('Go/TS app signing drift vectors', () => {
    it('matches Go PackCreateAppSessionRequestV1 and GenerateAppSessionIDV1 vectors', () => {
        expect(packCreateAppSessionRequestV1(definition, '{"cart":"demo"}')).toBe(
            '0x405d15a85c16ac1e555b3319de58acf7b4b86ebe2ccaf6af802d61e450b88632'
        );
        expect(generateAppSessionIDV1(definition)).toBe(
            '0x9b88181fc2ee0bc03abad5c4c9ea421c6748919882d4053204d95fbc79a175eb'
        );
    });

    it('matches Go PackCreateAppSessionRequestV1 uint64 nonce boundary vector', () => {
        expect(
            packCreateAppSessionRequestV1(
                {
                    ...definition,
                    nonce: maxUint64,
                },
                '{"cart":"max-nonce"}'
            )
        ).toBe('0xf15b0c1bc732b62d840e3c026e125cb5dec7da2b658c36355835ae56802c781c');
    });

    it('matches Go PackAppStateUpdateV1 deposit, withdraw, operate, fractional, and uint64 vectors', () => {
        const appSessionId = generateAppSessionIDV1(definition);

        expect(
            packAppStateUpdateV1({
                appSessionId,
                intent: AppStateUpdateIntent.Deposit,
                version: 2n,
                allocations: [
                    { participant: user, asset: 'YUSD', amount: new Decimal('1.25') },
                    { participant: app, asset: 'YUSD', amount: new Decimal('0') },
                ],
                sessionData: '{"intent":"deposit"}',
            })
        ).toBe('0x65e0856b8de315f40db44b9cc4165fa7e590169b3325e500a03aa380954c393d');

        expect(
            packAppStateUpdateV1({
                appSessionId,
                intent: AppStateUpdateIntent.Operate,
                version: 3n,
                allocations: [
                    { participant: user, asset: 'YUSD', amount: new Decimal('0.35') },
                    { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
                ],
                sessionData: '{"intent":"purchase","item_id":1,"item_price":"0.90"}',
            })
        ).toBe('0xe44d77fa3eda431b1bc088e6f89e114b2191ef5ce03cc6851c702d01bdbf3457');

        expect(
            packAppStateUpdateV1({
                appSessionId,
                intent: AppStateUpdateIntent.Withdraw,
                version: 4n,
                allocations: [
                    { participant: user, asset: 'YUSD', amount: new Decimal('0.10') },
                    { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
                ],
                sessionData: '{"intent":"withdraw"}',
            })
        ).toBe('0x4290525a204a34e5fc4d37427f1b0b1e2d375ed09ed0ac3b23d14dbc481c7d71');

        expect(
            packAppStateUpdateV1({
                appSessionId,
                intent: AppStateUpdateIntent.Deposit,
                version: 5n,
                allocations: [
                    { participant: user, asset: 'YUSD', amount: new Decimal('1.23456789') },
                    { participant: app, asset: 'YUSD', amount: new Decimal('0') },
                ],
                sessionData: '{"intent":"deposit","note":"fractional"}',
            })
        ).toBe('0x626e03a0850b83f3bac66dc7bd27b1e2d882fd88b54a61fd76d9fbaa35703098');

        expect(
            packAppStateUpdateV1({
                appSessionId,
                intent: AppStateUpdateIntent.Withdraw,
                version: maxUint64,
                allocations: [
                    { participant: user, asset: 'YUSD', amount: new Decimal('0') },
                    { participant: app, asset: 'YUSD', amount: new Decimal('1.25') },
                ],
                sessionData: '{"intent":"withdraw","boundary":"max_uint64_version"}',
            })
        ).toBe('0x6460b0c93c88da7fa34bfbf3893be74362e35987b525c7b34b4749e66fef8862');
    });

    it('matches Go PackAppSessionKeyStateV1 vector', () => {
        expect(packAppSessionKeyStateV1(sessionKeyState)).toBe(
            '0x9fedfbcd577c5e677b95b1273e38f52ffdeee096e98f731c5455e4c73e0274aa'
        );
    });

    it('proves adversarial allocation ordering changes the signed hash', () => {
        const appSessionId = generateAppSessionIDV1(definition);
        const canonical = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Deposit,
            version: 2n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('1.25') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0') },
            ],
            sessionData: '{"intent":"deposit"}',
        });
        const mutated = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Deposit,
            version: 2n,
            allocations: [
                { participant: app, asset: 'YUSD', amount: new Decimal('0') },
                { participant: user, asset: 'YUSD', amount: new Decimal('1.25') },
            ],
            sessionData: '{"intent":"deposit"}',
        });

        expect(mutated).not.toBe(canonical);
    });

    it('proves adversarial amount rounding changes the signed hash', () => {
        const appSessionId = generateAppSessionIDV1(definition);
        const canonical = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Deposit,
            version: 5n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('1.23456789') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0') },
            ],
            sessionData: '{"intent":"deposit","note":"fractional"}',
        });
        const rounded = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Deposit,
            version: 5n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('1.234568') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0') },
            ],
            sessionData: '{"intent":"deposit","note":"fractional"}',
        });

        expect(rounded).not.toBe(canonical);
    });

    it('proves adversarial intent enum changes the signed hash', () => {
        const appSessionId = generateAppSessionIDV1(definition);
        const canonical = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Withdraw,
            version: 4n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('0.10') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
            ],
            sessionData: '{"intent":"withdraw"}',
        });
        const wrongIntent = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Deposit,
            version: 4n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('0.10') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
            ],
            sessionData: '{"intent":"withdraw"}',
        });

        expect(wrongIntent).not.toBe(canonical);
    });

    it('proves adversarial session data normalization changes the signed hash', () => {
        const appSessionId = generateAppSessionIDV1(definition);
        const canonical = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Operate,
            version: 3n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('0.35') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
            ],
            sessionData: '{"intent":"purchase","item_id":1,"item_price":"0.90"}',
        });
        const normalized = packAppStateUpdateV1({
            appSessionId,
            intent: AppStateUpdateIntent.Operate,
            version: 3n,
            allocations: [
                { participant: user, asset: 'YUSD', amount: new Decimal('0.35') },
                { participant: app, asset: 'YUSD', amount: new Decimal('0.90') },
            ],
            sessionData: '{"item_id":1,"item_price":"0.90","intent":"purchase"}',
        });

        expect(normalized).not.toBe(canonical);
    });

    it('proves adversarial session-key ID placement changes the signed hash', () => {
        const canonical = packAppSessionKeyStateV1(sessionKeyState);
        const swappedIds = packAppSessionKeyStateV1({
            ...sessionKeyState,
            application_ids: sessionKeyState.app_session_ids,
            app_session_ids: sessionKeyState.application_ids,
        });

        expect(swappedIds).not.toBe(canonical);
    });
});
