import { jest } from '@jest/globals';

import {
    RPCAppStateIntent,
    createAppSessionMessage,
    createCloseAppSessionMessage,
    createCloseChannelMessage,
    createGetAppDefinitionMessage,
    createGetAppSessionsMessage,
    createGetChannelsMessage,
    createGetLedgerBalancesMessage,
    createPingMessage,
    createSubmitAppStateMessage,
    createTransferMessage,
    parseCreateAppSessionResponse,
    parseGetAppDefinitionResponse,
    parseGetAppSessionsResponse,
} from '../../src/index.js';

const signer = jest.fn(async () => '0xsigned');
const participant = '0x00000000000000000000000000000000000000a1';
const otherParticipant = '0x00000000000000000000000000000000000000b2';

type CompatRequest = {
    req: [number, string, Record<string, unknown>, number];
    sig: string;
};

function parseCompatRequest(raw: string): CompatRequest {
    return JSON.parse(raw) as CompatRequest;
}

describe('compat RPC helper wire shapes', () => {
    beforeEach(() => {
        signer.mockClear();
    });

    it('creates a real v1 ping payload inside the legacy req/sig envelope', async () => {
        const raw = await createPingMessage(signer, 7, 1234);
        const parsed = parseCompatRequest(raw);

        expect(parsed).toEqual({
            req: [7, 'node.v1.ping', {}, 1234],
            sig: '',
        });
        expect(signer).not.toHaveBeenCalled();
    });

    it('requires participant for getChannels and emits the v1 method name', async () => {
        const raw = await createGetChannelsMessage(signer, participant, 'open', 11, 999);
        const parsed = parseCompatRequest(raw);

        expect(parsed.req).toEqual([
            11,
            'channels.v1.get_channels',
            { wallet: participant, status: 'open' },
            999,
        ]);
        expect(parsed.sig).toBe('');

        await expect(createGetChannelsMessage(signer, undefined, 'open')).rejects.toThrow(
            'createGetChannelsMessage requires participant',
        );
    });

    it('requires wallet/account for getLedgerBalances and signs the request', async () => {
        const raw = await createGetLedgerBalancesMessage(signer, participant, 12, 555);
        const parsed = parseCompatRequest(raw);

        expect(parsed.req).toEqual([
            12,
            'user.v1.get_balances',
            { wallet: participant },
            555,
        ]);
        expect(parsed.sig).toBe('0xsigned');
        expect(signer).toHaveBeenCalledTimes(1);

        await expect(createGetLedgerBalancesMessage(signer)).rejects.toThrow(
            'createGetLedgerBalancesMessage requires accountId',
        );
    });

    it('maps app-session query helpers to live v1 methods', async () => {
        const sessionsRaw = await createGetAppSessionsMessage(signer, participant, 'open', 13, 777);
        const sessionsReq = parseCompatRequest(sessionsRaw);
        expect(sessionsReq).toEqual({
            req: [13, 'app_sessions.v1.get_app_sessions', { participant, status: 'open' }, 777],
            sig: '',
        });

        const definitionRaw = await createGetAppDefinitionMessage(signer, 'session-1', 14, 778);
        const definitionReq = parseCompatRequest(definitionRaw);
        expect(definitionReq).toEqual({
            req: [14, 'app_sessions.v1.get_app_definition', { app_session_id: 'session-1' }, 778],
            sig: '',
        });
    });

    it('creates a signed v1 create_app_session request and encodes legacy allocations into session_data', async () => {
        const raw = await createAppSessionMessage(
            signer,
            {
                definition: {
                    application: 'chess',
                    protocol: '' as never,
                    participants: [participant, otherParticipant] as [`0x${string}`, `0x${string}`],
                    weights: [1, 2],
                    quorum: 2,
                    nonce: 42,
                },
                allocations: [
                    { participant, asset: 'yusd', amount: '0.25' },
                    { participant: otherParticipant, asset: 'yusd', amount: '0.75' },
                ],
                quorum_sigs: ['0xsig1', '0xsig2'],
            },
            15,
            779,
        );

        const parsed = parseCompatRequest(raw);
        expect(parsed.sig).toBe('0xsigned');
        expect(parsed.req[1]).toBe('app_sessions.v1.create_app_session');
        expect(parsed.req[2]).toEqual({
            definition: {
                application_id: 'chess',
                participants: [
                    { wallet_address: participant, signature_weight: 1 },
                    { wallet_address: otherParticipant, signature_weight: 2 },
                ],
                quorum: 2,
                nonce: '42',
            },
            quorum_sigs: ['0xsig1', '0xsig2'],
            session_data: JSON.stringify({
                allocations: [
                    { participant, asset: 'yusd', amount: '0.25' },
                    { participant: otherParticipant, asset: 'yusd', amount: '0.75' },
                ],
            }),
        });
    });

    it('requires explicit version for submit/close app-state mappings and emits submit_app_state', async () => {
        const submitRaw = await createSubmitAppStateMessage(
            signer,
            {
                app_session_id: 'session-1' as `0x${string}`,
                intent: RPCAppStateIntent.Operate,
                version: 7,
                allocations: [{ participant, asset: 'yusd', amount: '0.01' }],
                session_data: '{"move":"e4"}',
                quorum_sigs: ['0xsig'],
            },
            16,
            780,
        );
        const submitReq = parseCompatRequest(submitRaw);
        expect(submitReq.req).toEqual([
            16,
            'app_sessions.v1.submit_app_state',
            {
                app_state_update: {
                    app_session_id: 'session-1',
                    intent: 0,
                    version: '7',
                    allocations: [{ participant, asset: 'yusd', amount: '0.01' }],
                    session_data: '{"move":"e4"}',
                },
                quorum_sigs: ['0xsig'],
            },
            780,
        ]);

        const closeRaw = await createCloseAppSessionMessage(
            signer,
            {
                app_session_id: 'session-2',
                allocations: [{ participant, asset: 'yusd', amount: '1.5' }],
                version: 9,
                quorum_sigs: ['0xclose'],
            },
            17,
            781,
        );
        const closeReq = parseCompatRequest(closeRaw);
        expect(closeReq.req).toEqual([
            17,
            'app_sessions.v1.submit_app_state',
            {
                app_state_update: {
                    app_session_id: 'session-2',
                    intent: 3,
                    version: '9',
                    allocations: [{ participant, asset: 'yusd', amount: '1.5' }],
                    session_data: '',
                },
                quorum_sigs: ['0xclose'],
            },
            781,
        ]);

        await expect(
            createSubmitAppStateMessage(signer, {
                app_session_id: 'session-3' as `0x${string}`,
                allocations: [{ participant, asset: 'yusd', amount: '0.01' }],
            }),
        ).rejects.toThrow('createSubmitAppStateMessage requires params.version');

        await expect(
            createCloseAppSessionMessage(signer, {
                app_session_id: 'session-4',
                allocations: [{ participant, asset: 'yusd', amount: '0.01' }],
                quorum_sigs: [],
            }),
        ).rejects.toThrow('createCloseAppSessionMessage requires params.version');
    });

    it('fails fast for legacy workflow helpers that do not map to a single v1 RPC', async () => {
        await expect(createTransferMessage(signer, { destination: participant })).rejects.toThrow(
            'NitroliteClient.transfer(destination, allocations)',
        );
        await expect(createCloseChannelMessage(signer, 'channel-1', participant)).rejects.toThrow(
            'NitroliteClient.closeChannel(...)',
        );
    });

    it('normalizes snake_case live responses into legacy parse helper shapes', () => {
        expect(
            parseGetAppSessionsResponse(
                JSON.stringify({ res: [1, 'app_sessions.v1.get_app_sessions', { app_sessions: [{ app_session_id: 's1' }] }] }),
            ),
        ).toEqual({ params: { appSessions: [{ app_session_id: 's1' }] } });

        expect(
            parseCreateAppSessionResponse(
                JSON.stringify({ res: [1, 'app_sessions.v1.create_app_session', { app_session_id: 's1', version: '1', status: 'open' }] }),
            ),
        ).toEqual({ params: { appSessionId: 's1', version: '1', status: 'open' } });

        expect(
            parseGetAppDefinitionResponse(
                JSON.stringify({ res: [1, 'app_sessions.v1.get_app_definition', { definition: { application_id: 'app-1' } }] }),
            ),
        ).toEqual({ params: { application_id: 'app-1' } });
    });
});
