import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { TestWebSocket, getGetAppSessionsPredicate } from '@/ws';
import { createAuthSessionWithClearnode } from '@/auth';
import {
    createGetAppSessionsMessageV2,
    parseGetAppSessionsResponse,
} from '@layer-3/nitrolite';

export interface TestSetupResult {
    alice: Identity;
    aliceWS: TestWebSocket;
    aliceClient: TestNitroliteClient;
    aliceAppIdentity: Identity;
    aliceAppWS: TestWebSocket;
    bob: Identity;
    bobAppIdentity: Identity;
    bobWS: TestWebSocket;
    bobClient: TestNitroliteClient;
}

/**
 * Sets up test identities, websockets, and clients for integration tests.
 * Uses Alice and Bob convention for two-party scenarios.
 */
export async function setupTestIdentitiesAndConnections(): Promise<TestSetupResult> {
    const alice = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
    const aliceWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    const aliceClient = new TestNitroliteClient(alice);

    const aliceAppIdentity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].APP_SESSION_PK);
    const aliceAppWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);

    const bobWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    const bob = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].SESSION_PK);
    const bobAppIdentity = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].APP_SESSION_PK);
    const bobClient = new TestNitroliteClient(bob);

    // Connect websockets
    await aliceWS.connect();
    await aliceAppWS.connect();
    await bobWS.connect();

    // Authenticate main identities
    await createAuthSessionWithClearnode(aliceWS, alice);
    await createAuthSessionWithClearnode(bobWS, bob);

    return {
        alice,
        aliceWS,
        aliceClient,
        aliceAppIdentity,
        aliceAppWS,
        bob,
        bobAppIdentity,
        bobWS,
        bobClient,
    };
}

/**
 * Fetches and parses app sessions for a given participant.
 * Expects exactly one app session to exist.
 */
export async function fetchAndParseAppSessions(
    participantAppWS: TestWebSocket,
    participantAppIdentity: Identity,
    appSessionId: string
) {
    const getAppSessionsMsg = createGetAppSessionsMessageV2(
        participantAppIdentity.walletAddress
    );
    const getAppSessionsResponse = await participantAppWS.sendAndWaitForResponse(
        getAppSessionsMsg,
        getGetAppSessionsPredicate(),
        1000
    );

    const getAppSessionsParsedResponse = parseGetAppSessionsResponse(getAppSessionsResponse);
    expect(getAppSessionsParsedResponse).toBeDefined();
    expect(getAppSessionsParsedResponse.params.appSessions).toHaveLength(1);

    const appSession = getAppSessionsParsedResponse.params.appSessions[0];
    expect(appSession.appSessionId).toBe(appSessionId);
    expect(appSession.sessionData).toBeDefined();

    return {
        appSession,
        sessionData: JSON.parse(appSession.sessionData!),
    };
}
