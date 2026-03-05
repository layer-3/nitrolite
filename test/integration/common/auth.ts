import {
    createEIP712AuthMessageSigner,
    createAuthRequestMessage,
    createAuthVerifyMessage,
    AuthRequestParams,
    parseAuthChallengeResponse,
} from '@layer-3/nitrolite';
import { Identity } from './identity';
import { getAuthChallengePredicate, getAuthVerifyPredicate, TestWebSocket } from './ws';

export const createAuthSessionWithClearnode = async (
    ws: TestWebSocket,
    identity: Identity,
    authRequestParams?: AuthRequestParams
) => {
    authRequestParams = authRequestParams || {
        address: identity.walletAddress,
        session_key: identity.sessionKeyAddress,
        application: 'clearnode', // Use 'clearnode' app name to allow session key to be used as a custody signer
        expires_at: BigInt(Math.floor(Date.now() / 1000) + 3600), // 1 hour expiration
        scope: 'console',
        allowances: [],
    };

    const eip712MessageSigner = createEIP712AuthMessageSigner(
        identity.walletClient,
        {
            scope: authRequestParams.scope,
            session_key: authRequestParams.session_key,
            expires_at: authRequestParams.expires_at,
            allowances: authRequestParams.allowances,
        },
        {
            name: authRequestParams.application,
        }
    );

    const authRequestMsg = await createAuthRequestMessage(authRequestParams);
    const authRequestResponse = await ws.sendAndWaitForResponse(authRequestMsg, getAuthChallengePredicate(), 1000);

    const authRequestParsedResponse = parseAuthChallengeResponse(authRequestResponse);

    const authVerifyMsg = await createAuthVerifyMessage(eip712MessageSigner, authRequestParsedResponse);
    await ws.sendAndWaitForResponse(authVerifyMsg, getAuthVerifyPredicate(), 1000);
};
