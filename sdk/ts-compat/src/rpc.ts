/**
 * RPC compatibility helpers for v0.5.3 imports.
 *
 * In v0.5.3, apps used a create-sign-send-parse pattern:
 *   const msg = await createGetChannelsMessage(signer.sign, addr);
 *   const raw = await sendRequest(msg);
 *   const parsed = parseGetChannelsResponse(raw);
 *
 * The compat layer keeps that import surface available. Helpers that map
 * directly onto a live v1 method emit real v1-compatible payloads inside the
 * legacy req/sig envelope. Helpers without an honest one-to-one mapping fail
 * fast with migration guidance instead of fabricating fake wire payloads.
 */

import { Decimal } from 'decimal.js';
import type { Address } from 'viem';

import {
    RPCAppStateIntent,
    type CloseAppSessionRequestParams,
    type CreateAppSessionRequestParams,
    type MessageSigner,
    type NitroliteRPCMessage,
    type RPCChannelStatus,
    type RPCResponse,
    type RPCAppDefinition,
    type RPCAppSessionAllocation,
    type SubmitAppStateRequestParams,
    type SubmitAppStateRequestParamsV04,
} from './types.js';

// ---------------------------------------------------------------------------
// parseAnyRPCResponse -- response normalization
// ---------------------------------------------------------------------------

type LegacyRPCEnvelope = {
    req?: [number, string, unknown, number];
    res?: [number, string, unknown];
    sig?: string;
};

function legacyJSONReplacer(key: string, value: unknown): unknown {
    if (typeof value === 'bigint') {
        return value.toString();
    }

    if (
        (key === 'blockchain_id' || key === 'epoch' || key === 'version' || key === 'nonce') &&
        (typeof value === 'number' || typeof value === 'bigint')
    ) {
        return value.toString();
    }

    return value;
}

function parseEnvelope(raw: string): LegacyRPCEnvelope | unknown {
    return JSON.parse(raw);
}

function extractResponsePayload(raw: string): { requestId: number; method: string; payload: any } {
    const data = parseEnvelope(raw);

    if (Array.isArray(data)) {
        return {
            requestId: data[1] ?? 0,
            method: data[2] ?? '',
            payload: data[3] ?? {},
        };
    }

    if (typeof data === 'object' && data !== null && 'res' in data && Array.isArray((data as LegacyRPCEnvelope).res)) {
        const response = (data as LegacyRPCEnvelope).res!;
        return {
            requestId: response[0] ?? 0,
            method: response[1] ?? '',
            payload: response[2] ?? {},
        };
    }

    return { requestId: 0, method: '', payload: data };
}

function normalizePayloadField<T>(payload: any, ...keys: string[]): T | undefined {
    for (const key of keys) {
        if (payload && Object.prototype.hasOwnProperty.call(payload, key)) {
            return payload[key] as T;
        }
    }
    return undefined;
}

function defaultRequestId(): number {
    return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
}

function serializeMessage(message: NitroliteRPCMessage): string {
    return JSON.stringify(message, legacyJSONReplacer);
}

function newUnsignedMessage(method: string, params: Record<string, unknown>, requestId = defaultRequestId(), timestamp = Date.now()): string {
    return serializeMessage(
        NitroliteRPC.createRequest({
            requestId,
            method,
            params,
            timestamp,
        }),
    );
}

async function newSignedMessage(
    signer: MessageSigner,
    method: string,
    params: Record<string, unknown>,
    requestId = defaultRequestId(),
    timestamp = Date.now(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        requestId,
        method,
        params,
        timestamp,
    });
    const signed = await NitroliteRPC.signRequestMessage(request, signer);
    return serializeMessage(signed);
}

function missingFieldError(helperName: string, fieldName: string, guidance: string): Error {
    return new Error(
        `[compat] ${helperName} requires ${fieldName} for the live v1 mapping. ${guidance}`,
    );
}

function unsupportedHelperError(helperName: string, guidance: string): Error {
    return new Error(`[compat] ${helperName} is not supported as a direct v1 RPC helper. ${guidance}`);
}

function toRPCAppDefinition(definition: RPCAppDefinition): Record<string, unknown> {
    return {
        application_id: definition.application,
        participants: definition.participants.map((walletAddress, index) => ({
            wallet_address: walletAddress,
            signature_weight: definition.weights[index] ?? 1,
        })),
        quorum: definition.quorum,
        nonce: BigInt(definition.nonce ?? Date.now()),
    };
}

function toRPCAppAllocations(allocations: RPCAppSessionAllocation[]): Array<Record<string, unknown>> {
    return allocations.map((allocation) => ({
        participant: allocation.participant,
        asset: allocation.asset,
        amount: new Decimal(allocation.amount),
    }));
}

function toRPCAppStateIntent(intent: RPCAppStateIntent): number {
    switch (intent) {
        case RPCAppStateIntent.Deposit:
            return 1;
        case RPCAppStateIntent.Withdraw:
            return 2;
        case RPCAppStateIntent.Close:
            return 3;
        case RPCAppStateIntent.Operate:
        default:
            return 0;
    }
}

function requireSubmitVersion(
    helperName: 'createSubmitAppStateMessage' | 'createCloseAppSessionMessage',
    version: number | undefined,
): number {
    if (version === undefined) {
        throw missingFieldError(
            helperName,
            'params.version',
            'Use NitroliteClient.submitAppState(...) / closeAppSession(...) or include the current app-session version explicitly.',
        );
    }
    return version;
}

function buildSubmitAppStateParams(
    params: SubmitAppStateRequestParams,
    intentOverride?: RPCAppStateIntent,
    quorumSigsOverride?: string[],
): Record<string, unknown> {
    const intent = intentOverride ?? ('intent' in params ? params.intent : RPCAppStateIntent.Operate);
    const version = requireSubmitVersion(
        intentOverride === RPCAppStateIntent.Close ? 'createCloseAppSessionMessage' : 'createSubmitAppStateMessage',
        'version' in params ? params.version : undefined,
    );

    return {
        app_state_update: {
            app_session_id: params.app_session_id,
            intent: toRPCAppStateIntent(intent),
            version,
            allocations: toRPCAppAllocations(params.allocations),
            session_data: params.session_data ?? '',
        },
        quorum_sigs: quorumSigsOverride ?? params.quorum_sigs ?? [],
    };
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

export function parseAnyRPCResponse(raw: string): RPCResponse {
    try {
        const { requestId, method, payload } = extractResponsePayload(raw);
        return { requestId, method, params: payload };
    } catch {
        return { requestId: 0, method: 'error', params: { error: 'parse failed' } };
    }
}

export async function createGetChannelsMessage(
    _signer: MessageSigner,
    participant?: Address,
    status?: RPCChannelStatus,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    if (!participant) {
        throw missingFieldError(
            'createGetChannelsMessage',
            'participant',
            'Pass the participant wallet or migrate to NitroliteClient.getChannels().',
        );
    }

    return newUnsignedMessage(
        'channels.v1.get_channels',
        {
            wallet: participant,
            ...(status ? { status } : {}),
        },
        requestId,
        timestamp,
    );
}

export const parseGetChannelsResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: { channels: normalizePayloadField<any[]>(payload, 'channels') ?? [] } };
};

export async function createGetLedgerBalancesMessage(
    signer: MessageSigner,
    accountId?: string,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    if (!accountId) {
        throw missingFieldError(
            'createGetLedgerBalancesMessage',
            'accountId',
            'Pass the wallet/account address or migrate to NitroliteClient.getBalances().',
        );
    }

    return newSignedMessage(
        signer,
        'user.v1.get_balances',
        { wallet: accountId },
        requestId,
        timestamp,
    );
}

export const parseGetLedgerBalancesResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return {
        params: {
            ledgerBalances: normalizePayloadField<any[]>(payload, 'ledgerBalances', 'balances') ?? [],
        },
    };
};

export const parseGetLedgerEntriesResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return {
        params: {
            ledgerEntries: normalizePayloadField<any[]>(payload, 'ledgerEntries', 'transactions') ?? [],
        },
    };
};

export const parseGetAppSessionsResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return {
        params: {
            appSessions: normalizePayloadField<any[]>(payload, 'appSessions', 'app_sessions') ?? [],
        },
    };
};

export async function createTransferMessage(
    _signer: MessageSigner,
    _params: unknown,
    _requestId?: number,
    _timestamp?: number,
): Promise<string> {
    throw unsupportedHelperError(
        'createTransferMessage',
        'Use NitroliteClient.transfer(destination, allocations) instead.',
    );
}

export async function createAppSessionMessage(
    signer: MessageSigner,
    params: CreateAppSessionRequestParams,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    return newSignedMessage(
        signer,
        'app_sessions.v1.create_app_session',
        {
            definition: toRPCAppDefinition(params.definition),
            session_data: params.session_data ?? JSON.stringify({ allocations: params.allocations }),
            quorum_sigs: params.quorum_sigs ?? [],
            ...(params.owner_sig ? { owner_sig: params.owner_sig } : {}),
        },
        requestId,
        timestamp,
    );
}

export const parseCreateAppSessionResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return {
        params: {
            appSessionId: normalizePayloadField<string>(payload, 'appSessionId', 'app_session_id') ?? '',
            version: normalizePayloadField<string | number>(payload, 'version') ?? '',
            status: normalizePayloadField<string>(payload, 'status') ?? '',
        },
    };
};

export async function createCloseAppSessionMessage(
    signer: MessageSigner,
    params: CloseAppSessionRequestParams,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    return newSignedMessage(
        signer,
        'app_sessions.v1.submit_app_state',
        buildSubmitAppStateParams(
            {
                app_session_id: params.app_session_id,
                intent: RPCAppStateIntent.Close,
                version: requireSubmitVersion('createCloseAppSessionMessage', params.version),
                allocations: params.allocations,
                session_data: params.session_data,
                quorum_sigs: params.quorum_sigs,
            } as SubmitAppStateRequestParamsV04,
            RPCAppStateIntent.Close,
            params.quorum_sigs,
        ),
        requestId,
        timestamp,
    );
}

export const parseCloseAppSessionResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: payload ?? {} };
};

export async function createSubmitAppStateMessage(
    signer: MessageSigner,
    params: SubmitAppStateRequestParams,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    return newSignedMessage(
        signer,
        'app_sessions.v1.submit_app_state',
        buildSubmitAppStateParams(params),
        requestId,
        timestamp,
    );
}

export const parseSubmitAppStateResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: payload ?? {} };
};

export async function createGetAppDefinitionMessage(
    _signer: MessageSigner,
    appSessionId: string,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    return newUnsignedMessage(
        'app_sessions.v1.get_app_definition',
        { app_session_id: appSessionId },
        requestId,
        timestamp,
    );
}

export const parseGetAppDefinitionResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: normalizePayloadField<Record<string, unknown>>(payload, 'definition') ?? {} };
};

export async function createGetAppSessionsMessage(
    _signer: MessageSigner,
    participant: Address,
    status?: RPCChannelStatus,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    if (!participant) {
        throw missingFieldError(
            'createGetAppSessionsMessage',
            'participant',
            'Pass the participant wallet or migrate to NitroliteClient.getAppSessionsList().',
        );
    }

    return newUnsignedMessage(
        'app_sessions.v1.get_app_sessions',
        {
            participant,
            ...(status ? { status } : {}),
        },
        requestId,
        timestamp,
    );
}

export async function createCreateChannelMessage(
    _signer: MessageSigner,
    _params: unknown,
    _requestId?: number,
    _timestamp?: number,
): Promise<string> {
    throw unsupportedHelperError(
        'createCreateChannelMessage',
        'Use NitroliteClient.deposit(tokenAddress, amount) or depositAndCreateChannel(...) instead.',
    );
}

export const parseCreateChannelResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: payload ?? {} };
};

export async function createCloseChannelMessage(
    _signer: MessageSigner,
    _channelId: string,
    _fundDestination: Address,
    _requestId?: number,
    _timestamp?: number,
): Promise<string> {
    throw unsupportedHelperError(
        'createCloseChannelMessage',
        'Use NitroliteClient.closeChannel(...) instead.',
    );
}

export const parseCloseChannelResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: payload ?? {} };
};

export async function createResizeChannelMessage(
    _signer: MessageSigner,
    _params: unknown,
    _requestId?: number,
    _timestamp?: number,
): Promise<string> {
    throw unsupportedHelperError(
        'createResizeChannelMessage',
        'Use NitroliteClient.resizeChannel(...) instead.',
    );
}

export const parseResizeChannelResponse = (raw: string) => {
    const { payload } = extractResponsePayload(raw);
    return { params: payload ?? {} };
};

export async function createPingMessage(
    _signer: MessageSigner,
    requestId?: number,
    timestamp?: number,
): Promise<string> {
    return newUnsignedMessage('node.v1.ping', {}, requestId, timestamp);
}

export const convertRPCToClientChannel = (ch: any) => ch;
export const convertRPCToClientState = (st: any, _sig?: string) => st;
