import { Identity } from '@/identity';
import { TestWebSocket, getSubmitAppStatePredicate, getCloseAppSessionPredicate } from '@/ws';
import {
    RPCAppSessionAllocation,
    RPCProtocolVersion,
    RPCAppStateIntent,
    createSubmitAppStateMessage,
    createCloseAppSessionMessage,
    parseSubmitAppStateResponse,
    parseCloseAppSessionResponse,
    RPCChannelStatus,
} from '@layer-3/nitrolite';
import { Hex } from 'viem';

/**
 * Submits an app state update for NitroRPC v0.2 (without intent/version parameters).
 */
export async function submitAppStateUpdate_v02(
    participantAppWS: TestWebSocket,
    participantAppIdentity: Identity,
    appSessionId: string,
    allocations: RPCAppSessionAllocation[],
    sessionData: object,
    expectedVersion: number
) {
    const submitAppStateMsg = await createSubmitAppStateMessage<RPCProtocolVersion.NitroRPC_0_2>(
        participantAppIdentity.messageSKSigner,
        {
            app_session_id: appSessionId as Hex,
            allocations,
            session_data: JSON.stringify(sessionData),
        }
    );

    const submitAppStateResponse = await participantAppWS.sendAndWaitForResponse(
        submitAppStateMsg,
        getSubmitAppStatePredicate(),
        1000
    );

    const submitAppStateParsedResponse = parseSubmitAppStateResponse(submitAppStateResponse);
    expect(submitAppStateParsedResponse).toBeDefined();
    expect(submitAppStateParsedResponse.params.appSessionId).toBe(appSessionId);
    expect(submitAppStateParsedResponse.params.status).toBe(RPCChannelStatus.Open);
    expect(submitAppStateParsedResponse.params.version).toBe(expectedVersion);

    return submitAppStateParsedResponse;
}

/**
 * Submits an app state update for NitroRPC v0.4 (with intent/version parameters).
 */
export async function submitAppStateUpdate_v04(
    participantAppWS: TestWebSocket,
    participantAppIdentity: Identity,
    appSessionId: string,
    intent: RPCAppStateIntent,
    version: number,
    allocations: RPCAppSessionAllocation[],
    sessionData: object
) {
    const submitAppStateMsg = await createSubmitAppStateMessage<RPCProtocolVersion.NitroRPC_0_4>(
        participantAppIdentity.messageSKSigner,
        {
            app_session_id: appSessionId as Hex,
            intent,
            version,
            allocations,
            session_data: JSON.stringify(sessionData),
        }
    );

    const submitAppStateResponse = await participantAppWS.sendAndWaitForResponse(
        submitAppStateMsg,
        getSubmitAppStatePredicate(),
        1000
    );

    const submitAppStateParsedResponse = parseSubmitAppStateResponse(submitAppStateResponse);
    expect(submitAppStateParsedResponse).toBeDefined();
    expect(submitAppStateParsedResponse.params.appSessionId).toBe(appSessionId);
    expect(submitAppStateParsedResponse.params.status).toBe(RPCChannelStatus.Open);
    expect(submitAppStateParsedResponse.params.version).toBe(version);

    return submitAppStateParsedResponse;
}

/**
 * Closes an app session with final state.
 */
export async function closeAppSessionWithState(
    participantAppWS: TestWebSocket,
    participantAppIdentity: Identity,
    appSessionId: string,
    allocations: RPCAppSessionAllocation[],
    sessionData: object,
    expectedVersion: number
) {
    const closeAppSessionMsg = await createCloseAppSessionMessage(participantAppIdentity.messageSKSigner, {
        app_session_id: appSessionId as Hex,
        allocations,
        session_data: JSON.stringify(sessionData),
    });

    const closeAppSessionResponse = await participantAppWS.sendAndWaitForResponse(
        closeAppSessionMsg,
        getCloseAppSessionPredicate(),
        1000
    );

    expect(closeAppSessionResponse).toBeDefined();

    const closeAppSessionParsedResponse = parseCloseAppSessionResponse(closeAppSessionResponse);
    expect(closeAppSessionParsedResponse).toBeDefined();
    expect(closeAppSessionParsedResponse.params.appSessionId).toBe(appSessionId);
    expect(closeAppSessionParsedResponse.params.status).toBe(RPCChannelStatus.Closed);
    expect(closeAppSessionParsedResponse.params.version).toBe(expectedVersion);

    return closeAppSessionParsedResponse;
}
