import {
    // Utility functions
    generateRequestId,
    getCurrentTimestamp,

    // Auth parse helpers
    parseAuthChallengeResponse,
    parseAuthVerifyResponse,

    // RPC parse helpers
    parseChannelUpdateResponse,
    parseTransferResponse,
    parseGetLedgerTransactionsResponse,

    // RPC create helpers
    createGetLedgerTransactionsMessageV2,
    createGetAppSessionsMessageV2,
    createCleanupSessionKeyCacheMessage,
    createRevokeSessionKeyMessage,

    // Types (compile-time check — importing is enough to verify they exist)
    type RPCData,
    type TransferRequestParams,
    type ResizeChannelParams,
    type RPCChannelOperation,
    type GetLedgerTransactionsFilters,
    type MessageSignerPayload,
    type NitroliteRPCRequest,
    type StateSigner,

    // Enum values
    RPCMethod,
} from '../../src/index';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeResponse(method: string, params: any): string {
    return JSON.stringify({ res: [1, method, params] });
}

// ---------------------------------------------------------------------------
// Export existence
// ---------------------------------------------------------------------------

describe('Export existence', () => {
    test('generateRequestId is a function that returns a number', () => {
        expect(typeof generateRequestId).toBe('function');
        const id = generateRequestId();
        expect(typeof id).toBe('number');
        expect(id).toBeGreaterThan(0);
        expect(Number.isInteger(id)).toBe(true);
    });

    test('getCurrentTimestamp returns a value close to Date.now()', () => {
        expect(typeof getCurrentTimestamp).toBe('function');
        const before = Date.now();
        const ts = getCurrentTimestamp();
        const after = Date.now();
        expect(ts).toBeGreaterThanOrEqual(before);
        expect(ts).toBeLessThanOrEqual(after);
    });

    test.each([
        ['parseAuthChallengeResponse', parseAuthChallengeResponse],
        ['parseAuthVerifyResponse', parseAuthVerifyResponse],
        ['parseChannelUpdateResponse', parseChannelUpdateResponse],
        ['parseTransferResponse', parseTransferResponse],
        ['parseGetLedgerTransactionsResponse', parseGetLedgerTransactionsResponse],
        ['createGetLedgerTransactionsMessageV2', createGetLedgerTransactionsMessageV2],
        ['createGetAppSessionsMessageV2', createGetAppSessionsMessageV2],
        ['createCleanupSessionKeyCacheMessage', createCleanupSessionKeyCacheMessage],
        ['createRevokeSessionKeyMessage', createRevokeSessionKeyMessage],
    ])('%s is exported as a function', (name, fn) => {
        expect(typeof fn).toBe('function');
    });

    test('RPCMethod includes new enum values', () => {
        expect(RPCMethod.Pong).toBe('pong');
        expect(RPCMethod.RevokeSessionKey).toBe('revoke_session_key');
        expect(RPCMethod.CleanupSessionKeyCache).toBe('cleanup_session_key_cache');
        expect(RPCMethod.GetSessionKeys).toBe('get_session_keys');
    });

    test('type aliases compile correctly', () => {
        // These are compile-time checks — if the types don't exist, TS will fail
        const _data: RPCData = [1, 'method', {}, 123];
        const _transfer: TransferRequestParams = { destination: '0x', allocations: [] };
        const _filters: GetLedgerTransactionsFilters = { limit: 10 };
        const _payload: MessageSignerPayload = new Uint8Array();
        const _req: NitroliteRPCRequest = [1, 'ping', {}, 123];

        expect(_data).toBeDefined();
        expect(_transfer).toBeDefined();
        expect(_filters).toBeDefined();
        expect(_payload).toBeDefined();
        expect(_req).toBeDefined();
    });
});

// ---------------------------------------------------------------------------
// Parse functions
// ---------------------------------------------------------------------------

describe('Parse functions', () => {
    test('parseAuthChallengeResponse extracts challenge message', () => {
        const raw = makeResponse('auth_challenge', { challengeMessage: 'sign-this' });
        const result = parseAuthChallengeResponse(raw);
        expect(result.method).toBe(RPCMethod.AuthChallenge);
        expect(result.params.challengeMessage).toBe('sign-this');
    });

    test('parseAuthChallengeResponse returns empty string for missing challenge', () => {
        const raw = makeResponse('auth_challenge', {});
        const result = parseAuthChallengeResponse(raw);
        expect(result.params.challengeMessage).toBe('');
    });

    test('parseAuthVerifyResponse extracts verify params', () => {
        const raw = makeResponse('auth_verify', {
            success: true,
            sessionKey: '0xabc',
            address: '0xdef',
            jwtToken: 'jwt123',
        });
        const result = parseAuthVerifyResponse(raw);
        expect(result.params.success).toBe(true);
        expect(result.params.sessionKey).toBe('0xabc');
        expect(result.params.address).toBe('0xdef');
        expect(result.params.jwtToken).toBe('jwt123');
    });

    test('parseChannelUpdateResponse extracts method and params', () => {
        const raw = makeResponse('channel_update', { channelId: '0x123', status: 'open' });
        const result = parseChannelUpdateResponse(raw);
        expect(result.method).toBe('channel_update');
        expect(result.params.channelId).toBe('0x123');
        expect(result.params.status).toBe('open');
    });

    test('parseTransferResponse extracts transactions array', () => {
        const txs = [{ id: 1, amount: '100' }];
        const raw = makeResponse('transfer', { transactions: txs });
        const result = parseTransferResponse(raw);
        expect(result.params.transactions).toEqual(txs);
    });

    test('parseTransferResponse returns empty array when no transactions', () => {
        const raw = makeResponse('transfer', {});
        const result = parseTransferResponse(raw);
        expect(result.params.transactions).toEqual([]);
    });

    test('parseGetLedgerTransactionsResponse extracts ledgerTransactions', () => {
        const txs = [{ id: 1, txType: 'transfer' }];
        const raw = makeResponse('get_ledger_transactions', { ledgerTransactions: txs });
        const result = parseGetLedgerTransactionsResponse(raw);
        expect(result.params.ledgerTransactions).toEqual(txs);
    });

    test('parseGetLedgerTransactionsResponse returns empty array for missing data', () => {
        const raw = makeResponse('get_ledger_transactions', {});
        const result = parseGetLedgerTransactionsResponse(raw);
        expect(result.params.ledgerTransactions).toEqual([]);
    });
});

// ---------------------------------------------------------------------------
// Create message functions
// ---------------------------------------------------------------------------

describe('Create message functions', () => {
    test('createGetLedgerTransactionsMessageV2 returns valid JSON with correct method', () => {
        const msg = createGetLedgerTransactionsMessageV2('0xabc');
        const parsed = JSON.parse(msg);
        expect(parsed.req[1]).toBe('get_ledger_transactions');
        expect(parsed.req[2].account_id).toBe('0xabc');
        expect(parsed.sig).toBe('0x');
    });

    test('createGetLedgerTransactionsMessageV2 includes filters', () => {
        const msg = createGetLedgerTransactionsMessageV2('0xabc', { limit: 10, sort: 'desc' });
        const parsed = JSON.parse(msg);
        expect(parsed.req[2].limit).toBe(10);
        expect(parsed.req[2].sort).toBe('desc');
    });

    test('createGetAppSessionsMessageV2 returns valid JSON with correct method', () => {
        const msg = createGetAppSessionsMessageV2('0xdef');
        const parsed = JSON.parse(msg);
        expect(parsed.req[1]).toBe('get_app_sessions');
        expect(parsed.req[2].account_id).toBe('0xdef');
        expect(parsed.sig).toBe('0x');
    });

    test('synchronous create functions return string (not Promise)', () => {
        const msg1 = createGetLedgerTransactionsMessageV2('0x1');
        const msg2 = createGetAppSessionsMessageV2('0x2');
        // If these were Promises, typeof would be 'object'
        expect(typeof msg1).toBe('string');
        expect(typeof msg2).toBe('string');
    });

    test('createCleanupSessionKeyCacheMessage returns a Promise', () => {
        const mockSigner = async () => '0xsig';
        const result = createCleanupSessionKeyCacheMessage(mockSigner);
        expect(result).toBeInstanceOf(Promise);
    });

    test('createRevokeSessionKeyMessage returns a Promise', () => {
        const mockSigner = async () => '0xsig';
        const result = createRevokeSessionKeyMessage(mockSigner, '0xkey');
        expect(result).toBeInstanceOf(Promise);
    });

    test('createCleanupSessionKeyCacheMessage produces signed message', async () => {
        const mockSigner = async () => '0xmocksig';
        const msg = await createCleanupSessionKeyCacheMessage(mockSigner);
        const parsed = JSON.parse(msg);
        expect(parsed.req[1]).toBe('cleanup_session_key_cache');
        expect(parsed.sig).toBe('0xmocksig');
    });

    test('createRevokeSessionKeyMessage produces signed message with session key', async () => {
        const mockSigner = async () => '0xmocksig';
        const msg = await createRevokeSessionKeyMessage(mockSigner, '0xSessionKey');
        const parsed = JSON.parse(msg);
        expect(parsed.req[1]).toBe('revoke_session_key');
        expect(parsed.req[2].session_key).toBe('0xSessionKey');
        expect(parsed.sig).toBe('0xmocksig');
    });
});
