import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { getRevokeSessionKeyPredicate, getTransferPredicate, TestWebSocket } from '@/ws';
import { CONFIG } from '@/setup';
import {
    RPCAppStateIntent,
    RPCProtocolVersion,
    createRevokeSessionKeyMessage,
    createTransferMessage,
    parseAnyRPCResponse,
} from '@layer-3/nitrolite';

import {
    createTestChannels,
    authenticateAppWithAllowances,
    authenticateAppWithMultiAssetAllowances,
    createTestAppSession,
    toRaw,
    getLedgerBalances,
} from '@/testHelpers';
import { submitAppStateUpdate_v04 } from '@/testAppSessionHelpers';
import { setupTestIdentitiesAndConnections } from '@/testSetup';
import { generatePrivateKey } from 'viem/accounts';

describe('Session Keys', () => {
    const ASSET_SYMBOL = CONFIG.TOKEN_SYMBOL;
    const onChainDepositAmount = BigInt(1000);
    const spendingCapAmount = BigInt(500); // Session key limited to 500 USDC
    const initialDepositAmount = BigInt(100);

    const ETH_ASSET_SYMBOL = 'yintegration.eth';
    const ETH_CAP = BigInt(2); // 2 ETH spending cap

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

    let appSessionId: string;

    let currentVersion = 1;

    const SESSION_DATA = { gameType: 'chess', gameState: 'waiting' };

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        ({ alice, aliceWS, aliceClient, aliceAppIdentity, aliceAppWS, bob, bobWS, bobClient, bobAppIdentity } =
            await setupTestIdentitiesAndConnections());
    });

    beforeEach(async () => {
        await blockUtils.makeSnapshot();

        // Create channels for both Alice and Bob
        await createTestChannels(
            [
                { client: aliceClient, ws: aliceWS },
                { client: bobClient, ws: bobWS },
            ],
            toRaw(onChainDepositAmount)
        );

        // Authenticate with spending cap of 500 USDC
        await authenticateAppWithMultiAssetAllowances(aliceAppWS, aliceAppIdentity, [
            { asset: ASSET_SYMBOL, amount: spendingCapAmount.toString()},
            { asset: ETH_ASSET_SYMBOL, amount: ETH_CAP.toString() },
        ]);

        currentVersion = 1;
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

    describe('Initial deposit within cap', () => {
        it('should allow deposit within spending cap', async () => {
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                initialDepositAmount.toString(),
                SESSION_DATA
            );

            expect(appSessionId).toBeDefined();

            // Verify ledger balance decreased
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - initialDepositAmount).toString());
        });

        it('should reject deposit exceeding spending cap', async () => {
            const excessiveAmount = spendingCapAmount + BigInt(100); // 600 USDC (exceeds 500 cap)

            await expect(
                createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_4,
                    ASSET_SYMBOL,
                    excessiveAmount.toString(),
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Cumulative spending tracking', () => {
        beforeEach(async () => {
            // Create initial app session with 100 USDC deposit
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                initialDepositAmount.toString(),
                SESSION_DATA
            );
        });

        it('should allow additional deposit within remaining cap', async () => {
            const additionalDeposit = BigInt(200); // Total: 100 + 200 = 300 (within 500 cap)

            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + additionalDeposit).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify ledger balance
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe(
                (onChainDepositAmount - initialDepositAmount - additionalDeposit).toString()
            );
        });

        it('should reject additional deposit exceeding remaining cap', async () => {
            const excessiveAdditionalDeposit = BigInt(450); // Total: 100 + 450 = 550 (exceeds 500 cap)

            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + excessiveAdditionalDeposit).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });

        it('should track cumulative spending across multiple deposits', async () => {
            // First additional deposit: 150 USDC (total: 250)
            let allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + BigInt(150)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Second additional deposit: 200 USDC (total: 450, within 500 cap)
            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + BigInt(150) + BigInt(200)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify total spent is 450 USDC
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - BigInt(450)).toString());

            // Third deposit attempting 100 more (total would be 550, exceeds cap)
            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + BigInt(150) + BigInt(200) + BigInt(100)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Withdrawals do not affect spending cap', () => {
        beforeEach(async () => {
            // Create initial app session with 300 USDC deposit
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                '300',
                SESSION_DATA
            );
        });

        it('should not restore spending cap after withdrawal', async () => {
            // Withdraw 100 USDC
            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '200', // Withdraw 100 from 300
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Withdraw,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify ledger balance increased by 100
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - BigInt(200)).toString());

            // Try to deposit 300 more (total spent would be 300 + 300 = 600, exceeds cap)
            const depositAllocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '500', // 200 + 300 = 500
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    depositAllocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Float amounts in allowances', () => {
        it('should support float amounts in spending caps', async () => {
            // Authenticate with float spending cap (500.5 USDC)
            await authenticateAppWithAllowances(
                aliceAppWS,
                aliceAppIdentity,
                ASSET_SYMBOL,
                '500.5'
            );

            // Create app session with float deposit (100.25 USDC)
            const testSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                '100.25',
                SESSION_DATA
            );

            expect(testSessionId).toBeDefined();

            // Verify we can deposit additional float amount
            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '250.75', // Total: 100.25 + 150.5 = 250.75 (within 500.5 cap)
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                testSessionId,
                RPCAppStateIntent.Deposit,
                2,
                allocations,
                SESSION_DATA
            );

            // Try to exceed cap with float amount (should fail)
            const excessiveAllocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '501.0', // 501.0 > 500.5 cap
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    testSessionId,
                    RPCAppStateIntent.Deposit,
                    3,
                    excessiveAllocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Multi-asset spending caps', () => {
        let appSessionId1: string;
        let appSessionId2: string;

        beforeEach(async () => {
            // Create WETH channel for Alice to have ETH in ledger
            await aliceClient.createAndWaitForChannel(aliceWS, {
                tokenAddress: CONFIG.ADDRESSES.WETH_TOKEN_ADDRESS,
                amount: toRaw(BigInt(10), 18), // 10 WETH
            });
        });

        it('should enforce spending cap per asset independently', async () => {
            // Create app session with 400 USDC deposit (within 500 USDC cap)
            appSessionId1 = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                '400',
                SESSION_DATA
            );
            expect(appSessionId1).toBeDefined();

            // Create second app session with 0 initial deposit
            appSessionId2 = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                '0',
                SESSION_DATA
            );
            expect(appSessionId2).toBeDefined();

            // Add 1 ETH to second session (within 2 ETH cap)
            let allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '1',
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId2,
                RPCAppStateIntent.Deposit,
                2,
                allocations,
                SESSION_DATA
            );

            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '500', // at the cap
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId1,
                RPCAppStateIntent.Deposit,
                2,
                allocations,
                SESSION_DATA
            );

            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '2', // at the cap
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId2,
                RPCAppStateIntent.Deposit,
                3,
                allocations,
                SESSION_DATA
            );

            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '501', // exceeds cap, should fail
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId1,
                    RPCAppStateIntent.Deposit,
                    3,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);

            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '2.1', // exceeds cap, should fail
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ETH_ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId2,
                    RPCAppStateIntent.Deposit,
                    4,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Session key revocation', () => {
        let aliceClearnodeWS: TestWebSocket;
        let aliceClearnodeIdentity: Identity;

        beforeEach(async () => {
            // Authenticate with a non-clearnode application and allowance
            await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, ASSET_SYMBOL, spendingCapAmount.toString());

            // Create a clearnode session key for testing privileged revocation
            const clearnodeSessionPK = generatePrivateKey();
            aliceClearnodeIdentity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, clearnodeSessionPK);
            aliceClearnodeWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
            await aliceClearnodeWS.connect();
            await authenticateAppWithAllowances(aliceClearnodeWS, aliceClearnodeIdentity, ASSET_SYMBOL, spendingCapAmount.toString(), 'clearnode');
        });

        afterEach(async () => {
            if (aliceClearnodeWS) {
                aliceClearnodeWS.close();
            }
        });

        it('should allow session key to revoke itself', async () => {
            // Session key revokes itself
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageSKSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceAppWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );

            const parsedRevoke = parseAnyRPCResponse(revokeResponse);
            expect((parsedRevoke.params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);

            // Verify session key can no longer be used
            const transferMsg = await createTransferMessage(aliceAppIdentity.messageSKSigner, {
                destination: bob.walletAddress,
                allocations: [{ asset: ASSET_SYMBOL, amount: '10' }],
            });

            await expect(
                aliceAppWS.sendAndWaitForResponse(
                    transferMsg,
                    getTransferPredicate(),
                    5000
                )
            ).rejects.toThrow(/invalid signature/i);
        });

        it('should allow wallet to revoke its session key', async () => {
            // Wallet revokes its session key
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageWalletSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceAppWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );

            const parsedRevoke = parseAnyRPCResponse(revokeResponse);
            expect((parsedRevoke.params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);
        });

        it('should allow clearnode session key to revoke another session key', async () => {
            // Clearnode session key revokes a non-clearnode session key
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceClearnodeIdentity.messageSKSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceClearnodeWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );

            const parsedRevoke = parseAnyRPCResponse(revokeResponse);
            expect((parsedRevoke.params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);
        });

        it('should reject non-clearnode session key revoking another session key', async () => {
            // Create a second non-clearnode session key
            const app2SessionPK = generatePrivateKey();
            const aliceApp2Identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, app2SessionPK);
            const aliceApp2WS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
            await aliceApp2WS.connect();
            await authenticateAppWithAllowances(aliceApp2WS, aliceApp2Identity, ASSET_SYMBOL, spendingCapAmount.toString(), 'another-app');

            // Try to revoke the first session key with the second non-clearnode session key
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceApp2Identity.messageSKSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            await expect(
                aliceApp2WS.sendAndWaitForResponse(
                    revokeMsg,
                    getRevokeSessionKeyPredicate(),
                    5000
                )
            ).rejects.toThrow(/insufficient permissions for the active session key/i);

            aliceApp2WS.close();
        });

        it('should reject revoking another user\'s session key', async () => {
            // Bob tries to revoke Alice's session key
            const bobAppSessionPK = generatePrivateKey();
            const bobAppIdentityLocal = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, bobAppSessionPK);
            const bobAppWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
            await bobAppWS.connect();
            await authenticateAppWithAllowances(bobAppWS, bobAppIdentityLocal, ASSET_SYMBOL, spendingCapAmount.toString(), 'test-app');

            const revokeMsg = await createRevokeSessionKeyMessage(
                bobAppIdentityLocal.messageWalletSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            await expect(
                bobAppWS.sendAndWaitForResponse(
                    revokeMsg,
                    getRevokeSessionKeyPredicate(),
                    5000
                )
            ).rejects.toThrow(/not an active session key of this user/i);

            bobAppWS.close();
        });

        it('should reject revoking non-existent session key', async () => {
            // Generate a random address that's not a session key
            const randomSessionPK = generatePrivateKey();
            const randomWalletPK = generatePrivateKey();
            const randomIdentity = new Identity(randomWalletPK, randomSessionPK);
            const randomAddress = randomIdentity.sessionKeyAddress;

            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageWalletSigner,
                randomAddress
            );

            await expect(
                aliceAppWS.sendAndWaitForResponse(
                    revokeMsg,
                    getRevokeSessionKeyPredicate(),
                    5000
                )
            ).rejects.toThrow(/not an active session key of this user/i);
        });

        it('should reject transfer after session key is revoked', async () => {
            // Verify session key works with a transfer
            const transferMsg = await createTransferMessage(aliceAppIdentity.messageSKSigner, {
                destination: bob.walletAddress,
                allocations: [{ asset: ASSET_SYMBOL, amount: '100' }],
            });

            await aliceAppWS.sendAndWaitForResponse(
                transferMsg,
                getTransferPredicate(),
                5000
            );

            // Revoke the session key using wallet signature
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageWalletSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceAppWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );
            expect((parseAnyRPCResponse(revokeResponse).params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);

            // Try transfer with revoked session key - should fail
            const transferMsg2 = await createTransferMessage(aliceAppIdentity.messageSKSigner, {
                destination: bob.walletAddress,
                allocations: [{ asset: ASSET_SYMBOL, amount: '100' }],
            });

            await expect(
                aliceAppWS.sendAndWaitForResponse(
                    transferMsg2,
                    getTransferPredicate(),
                    5000
                )
            ).rejects.toThrow(/invalid signature/i);
        });

        it('should reject app session creation after session key is revoked', async () => {
            // Revoke the session key using wallet signature
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageWalletSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceAppWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );

            const parsedRevoke = parseAnyRPCResponse(revokeResponse);
            expect((parsedRevoke.params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);

            // Try to create app session with revoked session key - should fail
            await expect(
                createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_4,
                    ASSET_SYMBOL,
                    initialDepositAmount.toString(),
                    SESSION_DATA
                )
            ).rejects.toThrow(/missing signature for participant/i);
        });

        it('should reject app state submission after session key is revoked', async () => {
            // First create an app session before revoking
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                ASSET_SYMBOL,
                initialDepositAmount.toString(),
                SESSION_DATA
            );

            expect(appSessionId).toBeDefined();

            // Revoke the session key using wallet signature
            const revokeMsg = await createRevokeSessionKeyMessage(
                aliceAppIdentity.messageWalletSigner,
                aliceAppIdentity.sessionKeyAddress
            );

            const revokeResponse = await aliceAppWS.sendAndWaitForResponse(
                revokeMsg,
                getRevokeSessionKeyPredicate(),
                5000
            );

            const parsedRevoke = parseAnyRPCResponse(revokeResponse);
            expect((parsedRevoke.params as any).sessionKey).toBe(aliceAppIdentity.sessionKeyAddress);

            // Try to submit app state update with revoked session key - should fail
            const additionalDeposit = BigInt(50);
            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: (initialDepositAmount + additionalDeposit).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: ASSET_SYMBOL,
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/signature from unknown participant wallet/i);
        });
    });
});
