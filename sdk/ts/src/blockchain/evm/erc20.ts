/**
 * ERC20 token contract wrapper
 */

import { Address } from 'viem';
import { Erc20Abi } from './erc20_abi.js';
import { EVMClient, WalletSigner } from './interface.js';

/**
 * ERC20 contract wrapper for token interactions
 */
export class ERC20 {
  private tokenAddress: Address;
  private client: EVMClient;
  private walletSigner?: WalletSigner;

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
   * Approve spender to spend amount of tokens.
   *
   * Sends the approve transaction directly without a pre-flight
   * `simulateContract` call. The simulate path tries to decode `approve`'s
   * return value against the standard ERC-20 ABI, which says it returns
   * `bool`. Non-standard tokens such as Tether USDT on Ethereum L1 declare
   * `approve` without a return value, so simulate fails with
   * `ContractFunctionZeroDataError` even though the tx itself would
   * succeed. `writeContract` for a non-view function only submits the tx
   * and never decodes its return data, so it works for both standard and
   * non-standard ERC-20s.
   */
  async approve(spender: Address, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer is required for approve operation');
    }

    try {
      const hash = await this.walletSigner.writeContract({
        address: this.tokenAddress,
        abi: Erc20Abi,
        functionName: 'approve',
        args: [spender, amount],
        account: this.walletSigner.account!,
      } as any);

      // Wait several blocks past the receipt so load-balanced public RPCs
      // converge on the post-approve state. Without this, a downstream
      // allowance read can hit a node that hasn't yet indexed the approve
      // and the next deposit/checkpoint reverts with "allowance is not
      // sufficient".
      await this.client.waitForTransactionReceipt({ hash, confirmations: 3 });

      return hash;
    } catch (error: any) {
      console.error('❌ Approve execution failed!');
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
