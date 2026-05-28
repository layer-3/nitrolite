/**
 * Auth functions -- real implementations matching v0.5.3 SDK behavior.
 *
 * These create properly formatted RPC messages for the nitronode's
 * auth_request / auth_verify flow over WebSocket.
 */

import { NitroliteRPC } from './rpc.js';
import { RPCMethod, EIP712AuthTypes } from './types.js';
import type { MessageSigner, MessageSignerPayload } from './types.js';

export interface AuthRequestParams {
    address: string;
    session_key: string;
    application: string;
    expires_at: bigint;
    scope: string;
    allowances: { asset: string; amount: string }[];
}

function generateRequestId(): number {
    return Math.floor(Date.now() + Math.random() * 10000);
}

export async function createAuthRequestMessage(
    params: AuthRequestParams,
    requestId: number = generateRequestId(),
    timestamp: number = Date.now(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthRequest,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request, (key, value) => {
        if (typeof value !== 'bigint') return value;
        const asNumber = Number(value);
        if (!Number.isSafeInteger(asNumber)) {
            const fieldName = key || '<root>';
            throw new Error(
                `Auth request bigint field "${fieldName}" exceeds Number.MAX_SAFE_INTEGER: ${value.toString()}`,
            );
        }
        return asNumber;
    });
}

export async function createAuthVerifyMessage(
    signer: MessageSigner,
    challenge: { params: { challengeMessage: string } },
    requestId: number = generateRequestId(),
    timestamp: number = Date.now(),
): Promise<string> {
    const params = { challenge: challenge.params.challengeMessage };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthVerify,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);
    return JSON.stringify(signedRequest);
}

export async function createAuthVerifyMessageWithJWT(
    jwtToken: string,
    requestId: number = generateRequestId(),
    timestamp: number = Date.now(),
): Promise<string> {
    const params = { jwt: jwtToken };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthVerify,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

export function createEIP712AuthMessageSigner(
    walletClient: any,
    partialMessage: {
        scope: string;
        session_key: `0x${string}`;
        expires_at: bigint;
        allowances: { asset: string; amount: string }[];
    },
    domain: { name: string },
): MessageSigner {
    return async (payload: MessageSignerPayload) => {
        const address = walletClient.account?.address;
        if (!address) {
            throw new Error('Wallet client is not connected or does not have an account.');
        }

        if (!Array.isArray(payload) || payload.length < 3) {
            throw new Error('Invalid payload for AuthVerify: Expected an RPC request tuple.');
        }

        const method = payload[1];
        if (method !== RPCMethod.AuthVerify) {
            throw new Error(
                `This EIP-712 signer is designed only for the '${RPCMethod.AuthVerify}' method, but received '${method}'.`,
            );
        }

        const params = payload[2];
        if (!('challenge' in params) || typeof params.challenge !== 'string') {
            throw new Error('Invalid payload for AuthVerify: The challenge string is missing or malformed.');
        }

        const message = {
            ...partialMessage,
            challenge: params.challenge,
            wallet: address,
        };

        try {
            const signature = await walletClient.signTypedData({
                account: walletClient.account,
                domain,
                types: EIP712AuthTypes,
                primaryType: 'Policy',
                message: { ...message },
            });
            return signature;
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err);
            console.error('EIP-712 signing failed:', errorMessage);
            throw new Error(`EIP-712 signing failed: ${errorMessage}`);
        }
    };
}
