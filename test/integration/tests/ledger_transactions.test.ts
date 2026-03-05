import { createAuthSessionWithClearnode } from '@/auth';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { CONFIG } from '@/setup';
import { getGetLedgerTransactionsPredicate, TestWebSocket } from '@/ws';
import {
    createGetLedgerTransactionsMessageV2,
    GetLedgerTransactionsFilters,
    parseGetLedgerTransactionsResponse,
    RPCTxType,
} from '@layer-3/nitrolite';

describe('Ledger Transactions Integration', () => {
    let ws: TestWebSocket;
    let identity: Identity;
    let databaseUtils: DatabaseUtils;

    beforeAll(async () => {
        // Setup database utils for cleanup
        databaseUtils = new DatabaseUtils();

        // Create identity
        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);

        // Create WebSocket connection
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        // Create authenticated session
        await createAuthSessionWithClearnode(ws, identity);
    });

    afterAll(async () => {
        if (ws) {
            ws.close();
        }

        // Clean up database
        await databaseUtils.resetClearnodeState();
        await databaseUtils.close();
    });

    describe('createGetLedgerTransactionsMessage', () => {
        it('should successfully request ledger transactions with no filters', async () => {
            const accountId = identity.walletAddress;
            const msg = createGetLedgerTransactionsMessageV2(accountId);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
        });

        it('should successfully request ledger transactions with asset filter', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: 'usdc',
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);

            // If there are transactions, they should all be for usdc
            if (ledgerTransactions.length > 0) {
                ledgerTransactions.forEach((transaction) => {
                    expect(transaction.asset).toBe('usdc');
                });
            }
        });

        it('should successfully request ledger transactions with tx_type filter', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                tx_type: RPCTxType.Deposit,
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);

            // If there are transactions, they should all be of type 'deposit'
            if (ledgerTransactions.length > 0) {
                ledgerTransactions.forEach((transaction) => {
                    expect(transaction.txType).toBe(RPCTxType.Deposit);
                });
            }
        });

        it('should successfully request ledger transactions with pagination', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                limit: 5,
                offset: 0,
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);

            // Should not return more than the limit
            expect(ledgerTransactions.length).toBeLessThanOrEqual(5);
        });

        it('should successfully request ledger transactions with sort order', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                sort: 'desc',
                limit: 10,
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);

            // If there are multiple transactions, they should be sorted by createdAt in descending order
            if (ledgerTransactions.length > 1) {
                for (let i = 0; i < ledgerTransactions.length - 1; i++) {
                    const currentDate = ledgerTransactions[i].createdAt;
                    const nextDate = ledgerTransactions[i + 1].createdAt;
                    expect(currentDate.getTime()).toBeGreaterThanOrEqual(nextDate.getTime());
                }
            }
        });

        it('should successfully request ledger transactions with all filters', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: 'usdc',
                tx_type: RPCTxType.Deposit,
                offset: 0,
                limit: 3,
                sort: 'desc',
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);

            // Should not return more than the limit
            expect(ledgerTransactions.length).toBeLessThanOrEqual(3);

            // All transactions should match the filters
            ledgerTransactions.forEach((transaction) => {
                expect(transaction.asset).toBe('usdc');
                expect(transaction.txType).toBe(RPCTxType.Deposit);
            });
        });

        it('should handle empty results gracefully', async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: 'NONEXISTENT_ASSET',
            };

            const msg = createGetLedgerTransactionsMessageV2(accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);
            expect(ledgerTransactions.length).toBe(0);
        });

        it('should handle invalid account ID gracefully', async () => {
            const invalidAccountId = '0x0000000000000000000000000000000000000000';

            const msg = createGetLedgerTransactionsMessageV2(invalidAccountId);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetLedgerTransactionsResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();

            const ledgerTransactions = parsedResponse.params.ledgerTransactions;
            expect(Array.isArray(ledgerTransactions)).toBe(true);
            // Should return empty array for invalid/non-existent account
            expect(ledgerTransactions.length).toBe(0);
        });
    });
});
