import { Hex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import type { MessageSigner, MessageSignerPayload } from './types.js';

/**
 * v0.5.3-compatible WalletStateSigner.
 * In v0.5.3, this wraps a WalletClient and signs EIP-191 messages.
 * In the compat layer, it is only stored as a reference -- actual signing
 * is handled by the v1 SDK's ChannelDefaultSigner internally.
 * We keep the class so that NitroliteStore.state.walletStateSigner compiles.
 */
export class WalletStateSigner {
    public readonly address: Hex;

    constructor(private walletClient: any) {
        this.address = walletClient?.account?.address ?? ('0x' as Hex);
    }

    async sign(data: Uint8Array): Promise<string> {
        if (!this.walletClient?.signMessage) {
            throw new Error('WalletClient does not support signMessage');
        }
        return this.walletClient.signMessage({
            account: this.walletClient.account,
            message: { raw: data },
        });
    }
}

/**
 * v0.5.3-compatible createECDSAMessageSigner.
 * Returns a sign function compatible with the MessageSigner type.
 */
export function createECDSAMessageSigner(privateKey: Hex): MessageSigner {
    const account = privateKeyToAccount(privateKey);
    const encoder = new TextEncoder();

    const toSignableBytes = (payload: MessageSignerPayload): Uint8Array => {
        if (payload instanceof Uint8Array) return payload;

        const normalized = JSON.stringify(payload, (_key, value) =>
            typeof value === 'bigint' ? value.toString() : value,
        );
        return encoder.encode(normalized);
    };

    return async (payload: MessageSignerPayload): Promise<string> => {
        return account.signMessage({ message: { raw: toSignableBytes(payload) } });
    };
}
