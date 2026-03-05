import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { TestWebSocket } from '@/ws';
import { RPCAppStateIntent, RPCProtocolVersion, State } from '@layer-3/nitrolite';
import { Hex } from 'viem';
import { fetchAndParseAppSessions, setupTestIdentitiesAndConnections } from '@/testSetup';
import {
    createTestChannels,
    authenticateAppWithAllowances,
    createTestAppSession,
    toRaw,
} from '@/testHelpers';
import { submitAppStateUpdate_v04 } from '@/testAppSessionHelpers';
import { createAuthSessionWithClearnode } from '@/auth';
import { CONFIG } from '@/setup';

describe('App session v0.4', () => {
    const ASSET_SYMBOL = CONFIG.TOKEN_SYMBOL;

    const onChainDepositAmount = BigInt(1000);
    const SKAllowanceAmount = BigInt(500);
    const appSessionDepositAmount = BigInt(100);

    let aliceWS: TestWebSocket;
    let alice: Identity;
    let aliceClient: TestNitroliteClient;

    let aliceAppWS: TestWebSocket;
    let aliceAppIdentity: Identity;

    let bobWS: TestWebSocket;
    let bob: Identity;
    let bobAppIdentity: Identity;
    let bobClient: TestNitroliteClient;

    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;

    let aliceChannelId: Hex;
    let bobChannelId: Hex;
    let appSessionId: string;

    let initialStates: State[];

    const START_SESSION_DATA = { gameType: 'chess', gameState: 'waiting' };

    let START_ALLOCATIONS;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        ({alice, aliceWS, aliceClient, aliceAppIdentity, aliceAppWS, bob, bobWS, bobClient, bobAppIdentity} = await setupTestIdentitiesAndConnections());

        START_ALLOCATIONS = [
            {
                participant: aliceAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: (appSessionDepositAmount).toString(),
            },
            {
                participant: bobAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: '0',
            },
        ];
    });

    beforeEach(async () => {
        await blockUtils.makeSnapshot();
        ({channelIds: [aliceChannelId, bobChannelId], states: initialStates} = await createTestChannels([{client: aliceClient, ws: aliceWS}, {client: bobClient, ws: bobWS}], toRaw(onChainDepositAmount)));

        await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, ASSET_SYMBOL, SKAllowanceAmount.toString());
    });

    afterEach(async () => {
        await blockUtils.resetSnapshot();
        await databaseUtils.resetClearnodeState();
    });

    afterAll(async () => {
        aliceWS.close();
        aliceAppWS.close();
        bobWS.close();

        await databaseUtils.close();
    });

    describe('session creation', () => {
        it('rejects to create app session if one of participants has non-empty channel', async () => {
            const resizeAmount = BigInt(1);
            await aliceClient.resizeChannelAndWait(
                aliceWS,
                aliceChannelId,
                initialStates[0],
                alice.walletAddress,
                BigInt(0), // resize only to the user, so that a channel contains on-chain funds
                resizeAmount
            );

            const balance = await aliceClient.getChannelBalance(aliceChannelId, CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(balance).toBe(resizeAmount);

            try {
                await createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_4,
                    ASSET_SYMBOL,
                    appSessionDepositAmount.toString(),
                    START_SESSION_DATA
                );
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*operation denied.*non-zero allocation.*detected/i);
                return;
            }

            throw new Error('App session creation was not rejected as expected.');
        });

        it('rejects to deposit to an app session if one of participants has non-empty channel', async () => {
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                appSessionDepositAmount.toString(),
                START_SESSION_DATA
            );

            const resizeAmount = BigInt(1);
            await aliceClient.resizeChannelAndWait(
                aliceWS,
                aliceChannelId,
                initialStates[0],
                alice.walletAddress,
                BigInt(0), // resize only to the user, so that a channel contains on-chain funds
                resizeAmount
            );

            const balance = await aliceClient.getChannelBalance(aliceChannelId, CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(balance).toBe(resizeAmount);


            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString();

            try {
                await submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    2,
                    allocations,
                    { state: 'test' }
                );
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*operation denied.*non-zero allocation.*detected/i);
                return;
            }

            throw new Error('App session deposit was not rejected as expected.');
        });
    });

    describe('state submission error cases', () => {
        let currentVersion = 1;

        beforeEach(async () => {
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                appSessionDepositAmount.toString(),
                START_SESSION_DATA
            );

            currentVersion = Number(initialStates[0].version);
        });

        describe('operate intent', () => {
            it('should fail on skipping version number', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 42, allocations, { state: 'blah'});
                } catch (e) {
                    expect((e as Error).message).toMatch(
                        `RPC Error: incorrect app state: incorrect version: expected ${
                            currentVersion + 1
                        }, got ${currentVersion + 42}`
                    );

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on operate intent and positive delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);

                allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 - more than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect operate request.*non-zero allocations sum delta/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on operate intent and negative delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 - less than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect operate request.*non-zero allocations sum delta/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });
        });

        describe('deposit intent', () => {
            it('should fail on zero delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS); // same as deposited, zero delta

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*non-positive allocations sum delta/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on negative delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 - less than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*decreased allocation for participant/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on positive and negative allocation deltas', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(20)).toString(); // 120 - more than deposited
                allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(10)).toString(); // 90 - less than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*decreased allocation for participant/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on insufficient unified balance', async () => {
                // Try to deposit more than Alice has in ledger (she has 1000, already deposited 100, so has 900 available)
                let allocations = structuredClone(START_ALLOCATIONS);
                const hugeAmount = onChainDepositAmount * BigInt(10); // 10,000
                allocations[0].amount = hugeAmount.toString(); // 10,000 - way more than available

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*insufficient unified balance/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on depositing to v0.2 app session', async () => {
                // Create a v0.2 app session (which doesn't support deposits/withdrawals)
                const v02AppSessionId = await createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_2,
                    ASSET_SYMBOL,
                    '0', // No initial deposit to avoid spending Alice's allowance
                    START_SESSION_DATA
                );

                let allocations = [
                    {
                        participant: aliceAppIdentity.walletAddress,
                        asset: 'usdc',
                        amount: '10', // Try to deposit 10 USDC
                    },
                    {
                        participant: bobAppIdentity.walletAddress,
                        asset: 'usdc',
                        amount: '0',
                    },
                ];

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, v02AppSessionId, RPCAppStateIntent.Deposit, 2, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect request.*specified parameters are not supported in this protocol/i);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on quorum reached but without depositor', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                // Bob is depositing
                allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(10)).toString(); // 10 - more than deposited

                try {
                    // Alice signs and constitutes 100% of quorum, but is not a depositor
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*depositor signature is required/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on depositor signature but no quorum', async () => {
                // authenticate Bob's app identity, so that it can sign
                await createAuthSessionWithClearnode(bobWS, bobAppIdentity);

                let allocations = structuredClone(START_ALLOCATIONS);
                // Bob is depositing
                allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(10)).toString(); // 10 - more than deposited

                try {
                    // Bob signs and constitutes 0% of quorum, but is a depositor
                    await submitAppStateUpdate_v04(bobWS, bobAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*quorum not reached/i);

                    const { sessionData } = await fetchAndParseAppSessions(bobWS, bobAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });
        });

        describe('withdraw intent', () => {
            it('should fail on zero delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS); // same as deposited, zero delta

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*non-negative allocations sum delta/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on positive delta', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 - more than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*increased allocation for participant/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on positive and negative allocation deltas', async () => {
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 - more than deposited
                allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(20)).toString(); // 80 - less than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*increased allocation for participant/i);

                    const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                    expect(sessionData).toEqual(START_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on withdrawing from v0.2 app session', async () => {
                await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, ASSET_SYMBOL, (appSessionDepositAmount * BigInt(2)).toString());

                // Create a v0.2 app session (which doesn't support deposits/withdrawals)
                const v02AppSessionId = await createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_2,
                    ASSET_SYMBOL,
                    appSessionDepositAmount.toString(),
                    START_SESSION_DATA
                );

                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 USDC - less than deposited

                try {
                    await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, v02AppSessionId, RPCAppStateIntent.Withdraw, 2, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect request.*specified parameters are not supported in this protocol/i);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });

            it('should fail on no quorum reached', async () => {
                // authenticate Bob's app identity, so that it can sign
                await createAuthSessionWithClearnode(bobWS, bobAppIdentity);

                // for Bob to withdraw, he needs to get a balance first
                let allocations = structuredClone(START_ALLOCATIONS);
                allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(50)).toString(); // 50
                allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(50)).toString(); // 50

                const INTERMEDIATE_SESSION_DATA = { state: 'intermediate' };
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, INTERMEDIATE_SESSION_DATA);

                // Bob is withdrawing
                allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(10)).toString(); // 40 - less than before

                try {
                    // Bob signs and constitutes 0% of quorum
                    await submitAppStateUpdate_v04(bobWS, bobAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 2, allocations, { state: 'test' });
                } catch (e) {
                    expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*quorum not reached/i);

                    const { sessionData } = await fetchAndParseAppSessions(bobWS, bobAppIdentity, appSessionId);
                    expect(sessionData).toEqual(INTERMEDIATE_SESSION_DATA);
                    return;
                }

                throw new Error('Expected error was not thrown');
            });
        });
    });
});
