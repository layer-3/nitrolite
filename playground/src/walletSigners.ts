import type { WalletClient, Address, Hex } from 'viem';

export class WalletStateSigner {
  constructor(private walletClient: WalletClient) {}

  getAddress(): Address {
    if (!this.walletClient.account?.address) {
      throw new Error('Wallet client does not have an account address');
    }
    return this.walletClient.account.address;
  }

  async signMessage(hash: Hex): Promise<Hex> {
    if (!this.walletClient.account) {
      throw new Error('Wallet client does not have an account');
    }
    return await this.walletClient.signMessage({
      account: this.walletClient.account,
      message: { raw: hash },
    });
  }
}

export class WalletTransactionSigner {
  constructor(private walletClient: WalletClient) {}

  getAddress(): Address {
    if (!this.walletClient.account?.address) {
      throw new Error('Wallet client does not have an account address');
    }
    return this.walletClient.account.address;
  }

  async sendTransaction(_tx: unknown): Promise<Hex> {
    throw new Error('sendTransaction requires a wallet client - use the blockchain client instead');
  }

  async signMessage(message: { raw: Hex }): Promise<Hex> {
    return await this.signRaw(message.raw);
  }

  async signPersonalMessage(hash: Hex): Promise<Hex> {
    if (!this.walletClient.account) {
      throw new Error('Wallet client does not have an account');
    }
    return await this.walletClient.signMessage({
      account: this.walletClient.account,
      message: { raw: hash },
    });
  }

  async signRaw(hash: Hex): Promise<Hex> {
    if (!this.walletClient.account) {
      throw new Error('Wallet client does not have an account');
    }
    return await this.walletClient.signTypedData({
      account: this.walletClient.account,
      domain: { name: 'Nitrolite', version: '1', chainId: 1 },
      types: { Message: [{ name: 'data', type: 'bytes32' }] },
      primaryType: 'Message',
      message: { data: hash },
    });
  }
}
