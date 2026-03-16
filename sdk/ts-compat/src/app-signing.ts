import { Address, Hex, concatHex, encodeAbiParameters, keccak256 } from 'viem';

import { RPCAppStateIntent } from './types.js';

const WALLET_QUORUM_PREFIX = '0xa1' as Hex;
const SESSION_KEY_QUORUM_PREFIX = '0xa2' as Hex;
const RAW_WALLET_SIGNATURE_LENGTH = 132; // 0x + 65-byte signature
const WRAPPED_WALLET_SIGNATURE_LENGTH = 134; // 0x + 1-byte prefix + 65-byte signature

export interface CreateAppSessionHashParticipant {
    walletAddress: Address | Hex;
    signatureWeight: number;
}

export interface CreateAppSessionHashParams {
    application: string;
    participants: CreateAppSessionHashParticipant[];
    quorum: number;
    nonce: bigint | number;
    sessionData?: string;
}

export interface SubmitAppStateHashAllocation {
    participant: Address | Hex;
    asset: string;
    amount: string;
}

export interface SubmitAppStateHashParams {
    appSessionId: Hex | string;
    intent: RPCAppStateIntent | 'close' | number;
    version: bigint | number;
    allocations: SubmitAppStateHashAllocation[];
    sessionData?: string;
}

function normalizeIntent(intent: SubmitAppStateHashParams['intent']): number {
    if (typeof intent === 'number') return intent;

    switch (intent) {
        case RPCAppStateIntent.Operate:
            return 0;
        case RPCAppStateIntent.Deposit:
            return 1;
        case RPCAppStateIntent.Withdraw:
            return 2;
        case 'close':
            return 3;
        default:
            throw new Error(`Unsupported app state intent: ${intent}`);
    }
}

/**
 * Deterministic hash for app-session creation quorum signatures.
 */
export function packCreateAppSessionHash(params: CreateAppSessionHashParams): Hex {
    return keccak256(
        encodeAbiParameters(
            [
                { type: 'string' },
                {
                    type: 'tuple[]',
                    components: [
                        { name: 'walletAddress', type: 'address' },
                        { name: 'signatureWeight', type: 'uint8' },
                    ],
                },
                { type: 'uint8' },
                { type: 'uint64' },
                { type: 'string' },
            ],
            [
                params.application,
                params.participants.map((participant) => ({
                    walletAddress: participant.walletAddress,
                    signatureWeight: participant.signatureWeight,
                })),
                params.quorum,
                BigInt(params.nonce),
                params.sessionData ?? '',
            ],
        ),
    );
}

/**
 * Deterministic hash for app-state update quorum signatures.
 */
export function packSubmitAppStateHash(params: SubmitAppStateHashParams): Hex {
    return keccak256(
        encodeAbiParameters(
            [
                { type: 'bytes32' },
                { type: 'uint8' },
                { type: 'uint64' },
                {
                    type: 'tuple[]',
                    components: [
                        { name: 'participant', type: 'address' },
                        { name: 'asset', type: 'string' },
                        { name: 'amount', type: 'string' },
                    ],
                },
                { type: 'string' },
            ],
            [
                params.appSessionId as Hex,
                normalizeIntent(params.intent),
                BigInt(params.version),
                params.allocations.map((allocation) => ({
                    participant: allocation.participant,
                    asset: allocation.asset,
                    amount: allocation.amount,
                })),
                params.sessionData ?? '',
            ],
        ),
    );
}

/**
 * Prefixes a wallet EIP-191 signature for quorum_sigs consumption by app sessions.
 */
export function toWalletQuorumSignature(signature: Hex | string): Hex {
    const normalized = signature.toLowerCase();
    if (!normalized.startsWith('0x')) {
        throw new Error('Signature must be a hex string with 0x prefix');
    }

    if (
        normalized.startsWith(WALLET_QUORUM_PREFIX) &&
        normalized.length === WRAPPED_WALLET_SIGNATURE_LENGTH
    ) {
        return normalized as Hex;
    }

    if (normalized.length !== RAW_WALLET_SIGNATURE_LENGTH) {
        throw new Error('Expected a 65-byte wallet signature (0x + 130 hex chars)');
    }

    return concatHex([WALLET_QUORUM_PREFIX, normalized as Hex]);
}

/**
 * Prefixes an app-session key EIP-191 signature for quorum_sigs consumption by app sessions.
 */
export function toSessionKeyQuorumSignature(signature: Hex | string): Hex {
    const normalized = signature.toLowerCase();
    if (!normalized.startsWith('0x')) {
        throw new Error('Signature must be a hex string with 0x prefix');
    }

    if (
        normalized.startsWith(SESSION_KEY_QUORUM_PREFIX) &&
        normalized.length === WRAPPED_WALLET_SIGNATURE_LENGTH
    ) {
        return normalized as Hex;
    }

    if (normalized.length !== RAW_WALLET_SIGNATURE_LENGTH) {
        throw new Error('Expected a 65-byte session key signature (0x + 130 hex chars)');
    }

    return concatHex([SESSION_KEY_QUORUM_PREFIX, normalized as Hex]);
}
