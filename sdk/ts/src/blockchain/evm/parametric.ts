/**
 * Parametric ERC20 token contract wrapper with sub-account support
 */

import { Address } from 'viem';
import { ERC20 } from './erc20';
import { ParametricTokenAbi } from './parametric_abi';
import { EVMClient, WalletSigner } from './interface';

/**
 * Account type enum matching Solidity
 */
export enum AccountType {
  Normal = 0,
  Super = 1,
}

/**
 * Parametric ERC20 contract wrapper with sub-account methods
 */
export class ParametricToken extends ERC20 {
  constructor(tokenAddress: Address, client: EVMClient, walletSigner?: WalletSigner) {
    super(tokenAddress, client, walletSigner); // Pass to parent
  }

  /**
   * Check if an address is a super account
   */
  async getAccountType(account: Address): Promise<AccountType> {
    try {
      const result = (await this.client.readContract({
        address: this.tokenAddress,
        abi: ParametricTokenAbi,
        functionName: 'accountType',
        args: [account],
      })) as number;
      return result as AccountType;
    } catch {
      return AccountType.Normal; // Default to normal if not parametric
    }
  }

  /**
   * Check if an address is a super account (convenience method)
   */
  async isSuperAccount(account: Address): Promise<boolean> {
    return (await this.getAccountType(account)) === AccountType.Super;
  }

  /**
   * Convert a normal account to super account
   */
  async convertToSuper(account: Address): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'convertToSuper',
      args: [account],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Create a new sub-account for a super account
   */
  async createSubAccount(account: Address): Promise<number> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request, result } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'createSubAccount',
      args: [account],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });

    return Number(result); // uint48 -> number
  }

  /**
   * Get balance of a specific sub-account
   */
  async balanceOfSub(superAccount: Address, subId: number): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'balanceOfSub',
      args: [superAccount, subId],
    }) as Promise<bigint>;
  }

  /**
   * Get number of sub-accounts for a super account
   */
  async subsCountOf(superAccount: Address): Promise<number> {
    const result = await this.client.readContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'subsCountOf',
      args: [superAccount],
    });
    return Number(result);
  }

  /**
   * Get parameter for a specific sub-account
   */
  async getSubParameter(superAccount: Address, subId: number): Promise<`0x${string}`> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'getSubParameter',
      args: [superAccount, subId],
    }) as Promise<`0x${string}`>;
  }

  /**
   * Get allowance for a specific sub-account
   */
  async allowanceForSub(owner: Address, subId: number, spender: Address): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'allowanceForSub',
      args: [owner, subId, spender],
    }) as Promise<bigint>;
  }

  /**
   * Transfer from normal account to specific sub-account
   */
  async transferToSub(toSuper: Address, toSubId: number, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'transferToSub',
      args: [toSuper, toSubId, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Transfer from specific sub-account to normal account
   */
  async transferFromSub(fromSubId: number, to: Address, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'transferFromSub',
      args: [fromSubId, to, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Transfer between sub-accounts of the same super account
   */
  async transferBetweenSubs(fromSubId: number, toSubId: number, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'transferBetweenSubs',
      args: [fromSubId, toSubId, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Approve spender for a specific sub-account
   */
  async approveForSub(ownerSubId: number, spender: Address, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'approveForSub',
      args: [ownerSubId, spender, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Approved transfer from normal to sub-account
   */
  async approvedTransferToSub(from: Address, toSuper: Address, toSubId: number, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'approvedTransferToSub',
      args: [from, toSuper, toSubId, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }

  /**
   * Approved transfer from one sub-account to another
   */
  async approvedTransferFromSubToSub(fromSuper: Address, fromSubId: number, toSuper: Address, toSubId: number, amount: bigint): Promise<string> {
    if (!this.walletSigner) {
      throw new Error('Wallet signer required');
    }

    const { request } = await this.client.simulateContract({
      address: this.tokenAddress,
      abi: ParametricTokenAbi,
      functionName: 'approvedTransferFromSubToSub',
      args: [fromSuper, fromSubId, toSuper, toSubId, amount],
      account: this.walletSigner.account!.address,
    });

    const hash = await this.walletSigner.writeContract(request);
    await this.client.waitForTransactionReceipt({ hash });
    return hash;
  }
}

/**
 * Create a new parametric ERC20 contract instance
 */
export function newParametricToken(tokenAddress: Address, client: EVMClient, walletSigner?: WalletSigner): ParametricToken {
  return new ParametricToken(tokenAddress, client, walletSigner);
}
