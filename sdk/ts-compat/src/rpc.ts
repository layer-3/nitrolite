/**
 * RPC compatibility helpers for v0.5.3 imports.
 *
 * In v0.5.3, hooks used a create-sign-send-parse pattern:
 *   const msg = await createGetChannelsMessage(signer.sign, addr);
 *   const raw = await sendRequest(msg);
 *   const parsed = parseGetChannelsResponse(raw);
 *
 * In the compat layer, most apps should call NitroliteClient methods directly.
 * The helpers below keep legacy import sites compiling; most create* helpers are
 * lightweight placeholders while parse* helpers normalize response shapes.
 */

import type { MessageSigner, RPCResponse, NitroliteRPCMessage, GetLedgerTransactionsFilters } from './types';

// ---------------------------------------------------------------------------
// parseAnyRPCResponse -- pass-through
// ---------------------------------------------------------------------------

export function parseAnyRPCResponse(raw: string): RPCResponse {
    try {
        const data = JSON.parse(raw);
        if (Array.isArray(data)) {
            return { requestId: data[1] ?? 0, method: data[2] ?? '', params: data[3] ?? {} };
        }
        if (data.res) {
            return { requestId: data.res[0] ?? 0, method: data.res[1] ?? '', params: data.res[2] ?? {} };
        }
        return { requestId: 0, method: '', params: data };
    } catch {
        return { requestId: 0, method: 'error', params: { error: 'parse failed' } };
    }
}

// ---------------------------------------------------------------------------
// NitroliteRPC namespace compat
// ---------------------------------------------------------------------------

export const NitroliteRPC = {
    createRequest(opts: { requestId: number; method: string; params: any; timestamp: number }): NitroliteRPCMessage {
        return {
            req: [opts.requestId, opts.method, opts.params, opts.timestamp],
            sig: '',
        };
    },

    async signRequestMessage(msg: NitroliteRPCMessage, signer: MessageSigner): Promise<NitroliteRPCMessage> {
        const signature = await signer(msg.req);
        return { ...msg, sig: signature };
    },
};

// ---------------------------------------------------------------------------
// create*Message / parse*Response compatibility helpers
// ---------------------------------------------------------------------------

const noop = async (..._args: any[]): Promise<string> =>
    JSON.stringify({ req: [0, 'noop', {}, Date.now()], sig: '0x' });

export const createGetChannelsMessage = noop;
export const parseGetChannelsResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { channels: d?.res?.[2]?.channels ?? [] } };
};

export const createGetLedgerBalancesMessage = noop;
export const parseGetLedgerBalancesResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { ledgerBalances: d?.res?.[2]?.ledgerBalances ?? [] } };
};

export const parseGetLedgerEntriesResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { ledgerEntries: d?.res?.[2]?.ledgerEntries ?? [] } };
};

export const parseGetAppSessionsResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { appSessions: d?.res?.[2]?.appSessions ?? [] } };
};

export const createTransferMessage = noop;
export const createAppSessionMessage = noop;
export const parseCreateAppSessionResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { appSessionId: d?.res?.[2]?.appSessionId ?? '' } };
};

export const createCloseAppSessionMessage = noop;
export const parseCloseAppSessionResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { appSessionId: d?.res?.[2]?.appSessionId ?? '' } };
};

export const createSubmitAppStateMessage = noop;
export const parseSubmitAppStateResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { appSessionId: d?.res?.[2]?.appSessionId ?? '', version: d?.res?.[2]?.version ?? 0, status: d?.res?.[2]?.status ?? '' } };
};

export const createGetAppDefinitionMessage = noop;
export const parseGetAppDefinitionResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: d?.res?.[2] ?? {} };
};

export const createGetAppSessionsMessage = noop;

export const createCreateChannelMessage = noop;
export const parseCreateChannelResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: d?.res?.[2] ?? {} };
};

export const createCloseChannelMessage = noop;
export const parseCloseChannelResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: d?.res?.[2] ?? {} };
};

export const createResizeChannelMessage = noop;
export const parseResizeChannelResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: d?.res?.[2] ?? {} };
};

export const createPingMessage = noop;

export const convertRPCToClientChannel = (ch: any) => ch;
export const convertRPCToClientState = (st: any, _sig?: string) => st;

// ---------------------------------------------------------------------------
// Additional parse*Response helpers (v0.5.3 compat)
// ---------------------------------------------------------------------------

export const parseChannelUpdateResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { method: d?.res?.[1] ?? '', params: d?.res?.[2] ?? {} };
};

export const parseTransferResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { transactions: d?.res?.[2]?.transactions ?? [] } };
};

export const parseGetLedgerTransactionsResponse = (raw: string) => {
    const d = JSON.parse(raw);
    return { params: { ledgerTransactions: d?.res?.[2]?.ledgerTransactions ?? [] } };
};

// ---------------------------------------------------------------------------
// Additional create*Message helpers (v0.5.3 compat)
// ---------------------------------------------------------------------------

function _generateRequestId(): number {
    return Math.floor(Date.now() + Math.random() * 10000);
}

/** Synchronous — integration tests call this without await. */
export function createGetLedgerTransactionsMessageV2(
    accountId: string,
    filters?: GetLedgerTransactionsFilters,
): string {
    return JSON.stringify({
        req: [_generateRequestId(), 'get_ledger_transactions', { account_id: accountId, ...filters }, Date.now()],
        sig: '0x',
    });
}

/** Synchronous — integration tests call this without await. */
export function createGetAppSessionsMessageV2(accountId: string): string {
    return JSON.stringify({
        req: [_generateRequestId(), 'get_app_sessions', { account_id: accountId }, Date.now()],
        sig: '0x',
    });
}

export async function createCleanupSessionKeyCacheMessage(
    signer: MessageSigner,
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        requestId: _generateRequestId(),
        method: 'cleanup_session_key_cache',
        params: {},
        timestamp: Date.now(),
    });
    const signed = await NitroliteRPC.signRequestMessage(request, signer);
    return JSON.stringify(signed);
}

export async function createRevokeSessionKeyMessage(
    signer: MessageSigner,
    sessionKeyAddress: string,
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        requestId: _generateRequestId(),
        method: 'revoke_session_key',
        params: { session_key: sessionKeyAddress },
        timestamp: Date.now(),
    });
    const signed = await NitroliteRPC.signRequestMessage(request, signer);
    return JSON.stringify(signed);
}
