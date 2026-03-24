/**
 * ERC20 token contract wrapper
 */

import { Address } from 'viem';
import { Erc20Abi } from './erc20_abi';
import { EVMClient, WalletSigner } from './interface';

/**
 * ERC20 contract wrapper for token interactions
 */
export class ERC20 {
  protected tokenAddress: Address;
  protected client: EVMClient;
  protected walletSigner?: WalletSigner;

  constructor(tokenAddress: Address, client: EVMClient, walletSigner?: WalletSigner) {
    this.tokenAddress = tokenAddress;
    this.client = client;
    this.walletSigner = walletSigner;
  }

  /**
   * Get the token balance of an account
   */
  async balanceOf(account: Address): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: Erc20Abi,
      functionName: 'balanceOf',
      args: [account],
    }) as Promise<bigint>;
  }

  /**
   * Get the allowance granted by owner to spender
   */
  async allowance(owner: Address, spender: Address): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: Erc20Abi,
      functionName: 'allowance',
      args: [owner, spender],
    }) as Promise<bigint>;
  }

  /**
   * Approve spender to spend amount of tokens
   */
  async approve(spender: Address, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer is required for approve operation');
    }

    try {
      const { request } = (await this.client.simulateContract({
        address: this.tokenAddress,
        abi: Erc20Abi,
        functionName: 'approve',
        args: [spender, amount],
        account: this.walletSigner.account!.address,
      } as any)) as any;

      const hash = await this.walletSigner.writeContract(request as any);

      await this.client.waitForTransactionReceipt({ hash });

      return hash;
    } catch (error: any) {
      console.error('❌ Approve simulation/execution failed!');
      if (error.message) {
        console.error('   Reason:', error.message);
      }
      throw error;
    }
  }

  /**
   * Get the decimals of the token
   */
  async decimals(): Promise<number> {
    // decimals() is a standard ERC20 function but not in our minimal ABI
    // We'll need to call it directly if needed
    throw new Error('decimals() not available in minimal ERC20 ABI');
  }
}

/**
 * Create a new ERC20 contract instance
 */
export function newERC20(tokenAddress: Address, client: EVMClient, walletSigner?: WalletSigner): ERC20 {
  return new ERC20(tokenAddress, client, walletSigner);
}
