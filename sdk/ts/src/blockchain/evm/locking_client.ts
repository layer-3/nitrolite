/**
 * Locking Contract Client
 * Provides lock/relock/unlock/withdraw functionality for the NonSlashableAppRegistry contract
 */

import { Address } from 'viem';
import Decimal from 'decimal.js';
import { decimalToBigInt } from '../../core/utils';
import { EVMClient, WalletSigner } from './interface';
import { AppRegistryAbi } from './app_registry_abi';
import { Erc20Abi } from './erc20_abi';

/**
 * LockingClient provides methods to interact with the Locking (NonSlashableAppRegistry) contract.
 * Supports lock, relock, unlock, withdraw, and approve operations.
 */
export class LockingClient {
  private contractAddress: Address;
  private evmClient: EVMClient;
  private walletSigner?: WalletSigner;

  private tokenAddress?: Address;
  private tokenDecimals?: number;

  constructor(contractAddress: Address, evmClient: EVMClient, walletSigner?: WalletSigner) {
    this.contractAddress = contractAddress;
    this.evmClient = evmClient;
    this.walletSigner = walletSigner;
  }

  private requireWalletSigner(): WalletSigner {
    if (!this.walletSigner) {
      throw new Error('Write operations require a wallet signer. In Node.js, use a TransactionSigner that implements getAccount() (e.g., EthereumRawSigner)');
    }
    return this.walletSigner;
  }

  /**
   * Lazily fetch and cache the token address and decimals from the contract.
   */
  private async ensureTokenInfo(): Promise<{ address: Address; decimals: number }> {
    if (this.tokenAddress !== undefined && this.tokenDecimals !== undefined) {
      return { address: this.tokenAddress, decimals: this.tokenDecimals };
    }

    const tokenAddress = (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'asset',
    })) as Address;

    const decimals = (await this.evmClient.readContract({
      address: tokenAddress,
      abi: Erc20Abi,
      functionName: 'decimals',
    })) as number;

    this.tokenAddress = tokenAddress;
    this.tokenDecimals = decimals;
    return { address: tokenAddress, decimals };
  }

  /**
   * Lock tokens into the Locking contract for the specified target address.
   * The caller must have approved the contract to spend tokens beforehand.
   *
   * @param target - The address to lock tokens for
   * @param amount - The amount of tokens to lock (human-readable decimals)
   * @returns Transaction hash
   */
  async lock(target: Address, amount: Decimal): Promise<string> {
    const walletSigner = this.requireWalletSigner();
    const { decimals } = await this.ensureTokenInfo();
    const amountBig = decimalToBigInt(amount, decimals);

    const { request } = await this.evmClient.simulateContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'lock',
      args: [target, amountBig],
      account: walletSigner.account!.address,
    });

    const hash = await walletSigner.writeContract(request as any);
    await this.evmClient.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Re-lock tokens that are in the unlocking state back to the locked state.
   *
   * @returns Transaction hash
   */
  async relock(): Promise<string> {
    const walletSigner = this.requireWalletSigner();
    const { request } = await this.evmClient.simulateContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'relock',
      account: walletSigner.account!.address,
    });

    const hash = await walletSigner.writeContract(request as any);
    await this.evmClient.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Initiate the unlock process for the caller's locked tokens.
   * After the unlock period elapses, withdraw can be called.
   *
   * @returns Transaction hash
   */
  async unlock(): Promise<string> {
    const walletSigner = this.requireWalletSigner();
    const { request } = await this.evmClient.simulateContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'unlock',
      account: walletSigner.account!.address,
    });

    const hash = await walletSigner.writeContract(request as any);
    await this.evmClient.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Withdraw unlocked tokens to the specified destination address.
   * Can only be called after the unlock period has elapsed.
   *
   * @param destination - The address to receive the withdrawn tokens
   * @returns Transaction hash
   */
  async withdraw(destination: Address): Promise<string> {
    const walletSigner = this.requireWalletSigner();
    const { request } = await this.evmClient.simulateContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'withdraw',
      args: [destination],
      account: walletSigner.account!.address,
    });

    const hash = await walletSigner.writeContract(request as any);
    await this.evmClient.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Approve the Locking contract to spend the specified amount of tokens.
   * This must be called before lock.
   *
   * @param amount - The amount of tokens to approve (human-readable decimals)
   * @returns Transaction hash
   */
  async approveToken(amount: Decimal): Promise<string> {
    const walletSigner = this.requireWalletSigner();
    const { address: tokenAddress, decimals } = await this.ensureTokenInfo();
    const amountBig = decimalToBigInt(amount, decimals);

    const { request } = await this.evmClient.simulateContract({
      address: tokenAddress,
      abi: Erc20Abi,
      functionName: 'approve',
      args: [this.contractAddress, amountBig],
      account: walletSigner.account!.address,
    });

    const hash = await walletSigner.writeContract(request as any);
    await this.evmClient.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Get the locked balance of a user in the Locking contract.
   *
   * @param user - The address to check
   * @returns The locked balance as a Decimal (adjusted for token decimals)
   */
  async getBalance(user: Address): Promise<Decimal> {
    const { decimals } = await this.ensureTokenInfo();

    const balance = (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'balanceOf',
      args: [user],
    })) as bigint;

    return new Decimal(balance.toString()).div(Decimal.pow(10, decimals));
  }

  /**
   * Get the lock state of a user.
   * Returns 0 for None, 1 for Locked, 2 for Unlocking.
   *
   * @param user - The address to check
   * @returns Lock state (0=None, 1=Locked, 2=Unlocking)
   */
  async getLockState(user: Address): Promise<number> {
    return (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: AppRegistryAbi,
      functionName: 'lockStateOf',
      args: [user],
    })) as number;
  }

  /**
   * Get the token decimals used by the Locking contract's asset.
   *
   * @returns Number of decimals
   */
  async getTokenDecimals(): Promise<number> {
    const { decimals } = await this.ensureTokenInfo();
    return decimals;
  }
}
