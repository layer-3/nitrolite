/**
 * Signer implementations for Nitrolite SDK
 * Provides EthereumMsgSigner and EthereumRawSigner matching the Go SDK patterns
 */

import { Address, Hex, encodeAbiParameters } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

/**
 * StateSigner interface for signing channel states
 * Used for signing off-chain state updates
 */
export interface StateSigner {
  /** Get the address of the signer */
  getAddress(): Address;
  /** Sign a message hash (used for EIP-191 message signing) */
  signMessage(hash: Hex): Promise<Hex>;
}

/**
 * TransactionSigner interface for signing blockchain transactions
 * Used for on-chain operations (deposits, withdrawals, etc.)
 */
export interface TransactionSigner {
  /** Get the address of the signer */
  getAddress(): Address;
  /** Send a transaction to the blockchain */
  sendTransaction(tx: any): Promise<Hex>;
  /** Sign a message (raw bytes) */
  signMessage(message: { raw: Hex }): Promise<Hex>;
  /** Sign a message with EIP-191 prefix (personal_sign). Used for app registration. */
  signPersonalMessage?(hash: Hex): Promise<Hex>;
  /** Get the underlying viem local account for wallet client creation */
  getAccount?(): ReturnType<typeof privateKeyToAccount>;
}

/**
 * EthereumMsgSigner implements StateSigner using EIP-191 message signing
 * Corresponds to Go SDK's sign.NewEthereumMsgSigner
 *
 * This signer prepends "\x19Ethereum Signed Message:\n" before signing,
 * making it compatible with eth_sign and personal_sign RPC methods.
 *
 * @example
 * ```typescript
 * import { privateKeyToAccount } from 'viem/accounts';
 * import { EthereumMsgSigner } from '@nitrolite/sdk';
 *
 * const account = privateKeyToAccount('0x...');
 * const signer = new EthereumMsgSigner(account);
 * ```
 */
export class EthereumMsgSigner implements StateSigner {
  private account: ReturnType<typeof privateKeyToAccount>;

  constructor(privateKeyOrAccount: Hex | ReturnType<typeof privateKeyToAccount>) {
    if (typeof privateKeyOrAccount === 'string') {
      this.account = privateKeyToAccount(privateKeyOrAccount);
    } else {
      this.account = privateKeyOrAccount;
    }
  }

  getAddress(): Address {
    return this.account.address;
  }

  /**
   * Sign a message hash using EIP-191 (with Ethereum message prefix)
   * The message is automatically prefixed with "\x19Ethereum Signed Message:\n"
   */
  async signMessage(hash: Hex): Promise<Hex> {
    return await this.account.signMessage({
      message: { raw: hash },
    });
  }
}

/**
 * EthereumRawSigner implements TransactionSigner using raw ECDSA signing
 * Corresponds to Go SDK's sign.NewEthereumRawSigner
 *
 * This signer signs raw hashes directly without any prefix,
 * making it suitable for transaction signing and EIP-712 typed data.
 *
 * @example
 * ```typescript
 * import { privateKeyToAccount } from 'viem/accounts';
 * import { EthereumRawSigner } from '@nitrolite/sdk';
 *
 * const account = privateKeyToAccount('0x...');
 * const signer = new EthereumRawSigner(account);
 * ```
 */
export class EthereumRawSigner implements TransactionSigner {
  private account: ReturnType<typeof privateKeyToAccount>;

  constructor(privateKeyOrAccount: Hex | ReturnType<typeof privateKeyToAccount>) {
    if (typeof privateKeyOrAccount === 'string') {
      this.account = privateKeyToAccount(privateKeyOrAccount);
    } else {
      this.account = privateKeyOrAccount;
    }
  }

  getAddress(): Address {
    return this.account.address;
  }

  /**
   * Send a transaction to the blockchain
   */
  async sendTransaction(tx: any): Promise<Hex> {
    throw new Error('sendTransaction requires a wallet client - use the blockchain client instead');
  }

  /**
   * Sign a message (raw bytes without prefix)
   */
  async signMessage(message: { raw: Hex }): Promise<Hex> {
    return await this.account.sign({ hash: message.raw });
  }

  /**
   * Sign a message with EIP-191 prefix (personal_sign)
   */
  async signPersonalMessage(hash: Hex): Promise<Hex> {
    return await this.account.signMessage({ message: { raw: hash } });
  }

  /**
   * Get the underlying viem local account for wallet client creation.
   * Required for Node.js environments where HTTP transport cannot sign transactions.
   */
  getAccount(): ReturnType<typeof privateKeyToAccount> {
    return this.account;
  }
}

/**
 * ChannelDefaultSigner wraps a StateSigner and prepends the 0x00 type byte
 * to signatures.
 * Corresponds to Go SDK's core.ChannelDefaultSigner.
 *
 * This signer wraps any StateSigner implementation and prepends the
 * ChannelSignerType_Default (0x00) byte to the resulting signature,
 * which is required by the Nitrolite protocol for signature validation.
 *
 * @example
 * ```typescript
 * const msgSigner = new EthereumMsgSigner(privateKey);
 * const channelSigner = new ChannelDefaultSigner(msgSigner);
 * const client = await Client.create(wsURL, channelSigner, txSigner);
 * ```
 */
export class ChannelDefaultSigner implements StateSigner {
  private inner: StateSigner;

  constructor(inner: StateSigner) {
    this.inner = inner;
  }

  getAddress(): Address {
    return this.inner.getAddress();
  }

  async signMessage(hash: Hex): Promise<Hex> {
    const sig = await this.inner.signMessage(hash);
    // Prepend 0x00 type byte (ChannelSignerType_Default)
    return `0x00${sig.slice(2)}` as Hex;
  }
}

/**
 * ChannelSessionKeyStateSigner implements StateSigner for session key delegation.
 * Corresponds to Go SDK's core.ChannelSessionKeySignerV1.
 *
 * Signs state data with the session key's private key (EIP-191), then wraps
 * the signature with the authorization data (session key address, metadata hash,
 * auth signature from the main wallet) and prepends a 0x01 type byte.
 *
 * The resulting compound signature format matches the Solidity SessionKeyValidator:
 *   0x01 || abi.encode(SessionKeyAuthorization{sessionKey, metadataHash, authSignature}, sessionKeySig)
 *
 * @example
 * ```typescript
 * const signer = new ChannelSessionKeyStateSigner(
 *   '0x...',           // session key private key
 *   '0xWallet...',     // main wallet address
 *   '0xMeta...',       // metadata hash from registration
 *   '0xAuthSig...',    // wallet's auth signature
 * );
 * const client = await Client.create(wsURL, signer, txSigner);
 * ```
 */
export class ChannelSessionKeyStateSigner implements StateSigner {
  private account: ReturnType<typeof privateKeyToAccount>;
  private walletAddress: Address;
  private metadataHash: Hex;
  private authSignature: Hex;

  constructor(sessionKeyPrivateKey: Hex, walletAddress: Address, metadataHash: Hex, authSignature: Hex) {
    this.account = privateKeyToAccount(sessionKeyPrivateKey);
    this.walletAddress = walletAddress;
    this.metadataHash = metadataHash;
    this.authSignature = authSignature;
  }

  /** Returns the main wallet address (not the session key address) */
  getAddress(): Address {
    return this.walletAddress;
  }

  /** Returns the session key address */
  getSessionKeyAddress(): Address {
    return this.account.address;
  }

  async signMessage(hash: Hex): Promise<Hex> {
    // Sign with session key (EIP-191)
    const sessionKeySig = await this.account.signMessage({
      message: { raw: hash },
    });

    // ABI-encode (SessionKeyAuthorization, sessionKeySig) matching Solidity struct
    const encoded = encodeAbiParameters(
      [
        {
          type: 'tuple',
          components: [
            { name: 'sessionKey', type: 'address' },
            { name: 'metadataHash', type: 'bytes32' },
            { name: 'authSignature', type: 'bytes' },
          ],
        },
        { type: 'bytes' },
      ],
      [
        {
          sessionKey: this.account.address,
          metadataHash: this.metadataHash,
          authSignature: this.authSignature,
        },
        sessionKeySig,
      ],
    );

    // Prepend 0x01 type byte (ChannelSignerType_SessionKey)
    return `0x01${encoded.slice(2)}` as Hex;
  }
}

/**
 * AppSessionWalletSignerV1 wraps an EthereumMsgSigner and prepends the 0x00 type byte
 * to signatures for app session operations.
 * Corresponds to Go SDK's app.NewAppSessionWalletSignerV1.
 *
 * @example
 * ```typescript
 * const msgSigner = new EthereumMsgSigner(privateKey);
 * const appSessionSigner = new AppSessionWalletSignerV1(msgSigner);
 * const sig = await appSessionSigner.signMessage(hash);
 * ```
 */
export class AppSessionWalletSignerV1 implements StateSigner {
  private inner: StateSigner;

  constructor(inner: StateSigner) {
    this.inner = inner;
  }

  getAddress(): Address {
    return this.inner.getAddress();
  }

  async signMessage(hash: Hex): Promise<Hex> {
    const sig = await this.inner.signMessage(hash);
    // Prepend 0xa1 type byte (AppSessionSignerTypeV1_Wallet)
    return `0xa1${sig.slice(2)}` as Hex;
  }
}

/**
 * AppSessionKeySignerV1 wraps an EthereumMsgSigner and prepends the 0xa2 type byte
 * to signatures for app session operations using session keys.
 * Corresponds to Go SDK's app.NewAppSessionKeySignerV1.
 *
 * @example
 * ```typescript
 * const msgSigner = new EthereumMsgSigner(sessionKeyPrivateKey);
 * const appSessionSigner = new AppSessionKeySignerV1(msgSigner);
 * const sig = await appSessionSigner.signMessage(hash);
 * ```
 */
export class AppSessionKeySignerV1 implements StateSigner {
  private inner: StateSigner;

  constructor(inner: StateSigner) {
    this.inner = inner;
  }

  getAddress(): Address {
    return this.inner.getAddress();
  }

  async signMessage(hash: Hex): Promise<Hex> {
    const sig = await this.inner.signMessage(hash);
    // Prepend 0xa2 type byte (AppSessionSignerTypeV1_SessionKey)
    return `0xa2${sig.slice(2)}` as Hex;
  }
}

/**
 * Helper function to create signers from a private key
 *
 * @param privateKey - Hex-encoded private key
 * @returns Object containing both state and transaction signers
 *
 * @example
 * ```typescript
 * import { createSigners } from '@nitrolite/sdk';
 *
 * const { stateSigner, txSigner } = createSigners('0x...');
 * const client = await Client.create(wsURL, stateSigner, txSigner);
 * ```
 */
export function createSigners(privateKey: Hex): {
  stateSigner: StateSigner;
  txSigner: TransactionSigner;
} {
  const account = privateKeyToAccount(privateKey);
  return {
    stateSigner: new ChannelDefaultSigner(new EthereumMsgSigner(account)),
    txSigner: new EthereumRawSigner(account),
  };
}
