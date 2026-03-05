import { Address, createWalletClient, Hex, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { chain } from './setup';
import { createECDSAMessageSigner } from '@layer-3/nitrolite';
import { SessionKeyStateSigner } from '@layer-3/nitrolite/dist/client/signer';

export class Identity {
    public walletClient = null;
    public stateSKSigner = null;
    public stateWalletSigner = null;
    public walletAddress: Address;
    public sessionKeyAddress: Address;
    public messageSKSigner = null;
    public messageWalletSigner = null;

    constructor(walletPrivateKey: Hex, sessionPrivateKey: Hex) {
        const walletAccount = privateKeyToAccount(walletPrivateKey);
        this.walletAddress = walletAccount.address;

        this.walletClient = createWalletClient({
            account: walletAccount,
            chain,
            transport: http(),
        });

        this.stateSKSigner = new SessionKeyStateSigner(sessionPrivateKey);
        this.messageSKSigner = createECDSAMessageSigner(sessionPrivateKey);
        this.sessionKeyAddress = this.stateSKSigner.getAddress();

        this.stateWalletSigner = new SessionKeyStateSigner(walletPrivateKey);
        this.messageWalletSigner = createECDSAMessageSigner(walletPrivateKey);
    }
}
