import { Decimal } from 'decimal.js';

import {
    AppStateUpdateIntent,
    generateAppSessionIDV1,
    packAppStateUpdateV1,
    packCreateAppSessionRequestV1,
    type AppDefinitionV1,
} from '../../src/app/index.js';

const user = '0x1111111111111111111111111111111111111111';
const app = '0x2222222222222222222222222222222222222222';

const definition: AppDefinitionV1 = {
    applicationId: 'store-v1',
    participants: [
        { walletAddress: user, signatureWeight: 1 },
        { walletAddress: app, signatureWeight: 1 },
    ],
    quorum: 2,
    nonce: 123456789n,
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

    it('matches Go PackAppStateUpdateV1 deposit and operate vectors', () => {
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
});
