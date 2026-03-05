import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getTransferPredicate, TestWebSocket } from '@/ws';
import {
    createTransferMessage,
    parseTransferResponse,
    parseAnyRPCResponse,
    TransferRequestParams,
    RPCMethod,
} from '@layer-3/nitrolite';

describe('Transfer Integration', () => {
    let ws: TestWebSocket;
    let senderIdentity: Identity;
    let recipientIdentity: Identity;
    let client: TestNitroliteClient;

    let databaseUtils: DatabaseUtils;
    let blockUtils: BlockchainUtils;

    beforeAll(async () => {
        // Setup database utils for cleanup
        databaseUtils = new DatabaseUtils();
        blockUtils = new BlockchainUtils();

        // Create identities
        senderIdentity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        recipientIdentity = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].SESSION_PK);

        client = new TestNitroliteClient(senderIdentity);

        // Create WebSocket connection
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    });

    beforeEach(async () => {
        await ws.connect();
        await createAuthSessionWithClearnode(ws, senderIdentity);
        await databaseUtils.seedLedger(senderIdentity.walletAddress, senderIdentity.walletAddress, 0, 'usdc', 1000);
        await blockUtils.makeSnapshot();
    });

    afterEach(async () => {
        await ws.close();
        await databaseUtils.resetClearnodeState();
        await blockUtils.resetSnapshot();
    });

    afterAll(async () => {
        if (ws) {
            ws.close();
        }

        await databaseUtils.close();
    });

    it('successfully transfers funds to another wallet', async () => {
        const transferParams: TransferRequestParams = {
            destination: recipientIdentity.walletAddress,
            allocations: [
                {
                    asset: 'usdc',
                    amount: '100',
                },
            ]
        };

        const transferMsg = await createTransferMessage(senderIdentity.messageSKSigner, transferParams);
        const response = await ws.sendAndWaitForResponse(transferMsg, getTransferPredicate(), 5000);

        expect(response).toBeDefined();

        const parsedResponse = parseTransferResponse(response);
        expect(parsedResponse).toBeDefined();
        expect(parsedResponse.params).toBeDefined();
        expect(parsedResponse.params.transactions).toBeDefined();
        expect(Array.isArray(parsedResponse.params.transactions)).toBe(true);
        expect(parsedResponse.params.transactions.length).toBe(1);

        // Verify transaction details
        const transaction = parsedResponse.params.transactions[0];
        expect(transaction.fromAccount).toBe(senderIdentity.walletAddress);
        expect(transaction.toAccount).toBe(recipientIdentity.walletAddress);
        expect(transaction.asset).toBe('usdc');
        expect(transaction.amount).toBe('100');
        expect(transaction.txType).toBe('transfer');
    });

    it('rejects duplicate transfer request', async () => {
        const transferParams: TransferRequestParams = {
            destination: recipientIdentity.walletAddress,
            allocations: [
                {
                    asset: 'usdc',
                    amount: '50',
                },
            ]
        };

        // First transfer - should succeed
        const transferMsg1 = await createTransferMessage(
            senderIdentity.messageSKSigner,
            transferParams,
        );
        const response1 = await ws.sendAndWaitForResponse(transferMsg1, getTransferPredicate(), 5000);

        expect(response1).toBeDefined();

        // Check if it's an error or success
        const parsed1 = parseAnyRPCResponse(response1);
        if (parsed1.method === RPCMethod.Error) {
            const errorParams = parsed1.params as { error: string };
            throw new Error(`First transfer failed: ${errorParams.error}`);
        }

        const parsedResponse1 = parseTransferResponse(response1);
        expect(parsedResponse1).toBeDefined();
        expect(parsedResponse1.params).toBeDefined();
        expect(parsedResponse1.params.transactions).toBeDefined();
        expect(parsedResponse1.params.transactions.length).toBe(1);

        // Verify first transaction succeeded
        const transaction1 = parsedResponse1.params.transactions[0];
        expect(transaction1.amount).toBe('50');

        try {
            await ws.sendAndWaitForResponse(transferMsg1, getTransferPredicate(), 5000);
        } catch (error) {
            // Expecting an error for duplicate transfer
            const err = error as Error;
            expect(err.message).toMatch(/RPC Error.*operation denied.*the request has already been processed/i);
            return;
        }

        throw new Error('Duplicate transfer request was not rejected as expected.');
    });

    it('rejects transfer if sender has NON-empty on-chain channel', async () => {
        const depositAmount = BigInt(50);
        await client.deposit(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);

        const {params, state} = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount,
        });

        const resizeAmount = BigInt(1);
        await client.resizeChannelAndWait(
            ws,
            params.channelId,
            state,
            senderIdentity.walletAddress,
            resizeAmount,
            BigInt(0) // resize only to the user, so that a channel contains on-chain funds
        );

        const balance = await client.getChannelBalance(params.channelId, CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(balance).toBe(resizeAmount);

        // try to make transfer
        const transferParams: TransferRequestParams = {
            destination: recipientIdentity.walletAddress,
            allocations: [
                {
                    asset: 'usdc',
                    amount: '20',
                },
            ]
        };

        const transferMsg = await createTransferMessage(
            senderIdentity.messageSKSigner,
            transferParams,
        );

        try {
            await ws.sendAndWaitForResponse(transferMsg, getTransferPredicate(), 5000);
        } catch (error) {
            // Expecting an error due to non-empty on-chain channel
            const err = error as Error;
            expect(err.message).toMatch(/RPC Error.*operation denied.*non-zero allocation.*detected/i);
            return;
        }

        throw new Error('Transfer request was not rejected as expected.');
    });
});
