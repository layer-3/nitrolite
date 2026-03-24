import { Address, Hex, encodeAbiParameters, concat, toHex } from 'viem';
import { State } from './types';
import { AssetStore, StatePacker } from './interface';
import { getStateTransitionHash, transitionToIntent, decimalToBigInt } from './utils';

/**
 * ContractLedger matches Solidity's Ledger struct for ABI encoding
 */
interface ContractLedger {
  chainId: bigint; // uint64
  token: Address;
  decimals: number; // uint8
  userAllocation: bigint; // uint256
  userNetFlow: bigint; // int256
  nodeAllocation: bigint; // uint256
  nodeNetFlow: bigint; // int256
}

/**
 * StatePackerV1 encodes states into ABI-packed bytes for on-chain submission
 */
export class StatePackerV1 implements StatePacker {
  private assetStore: AssetStore;

  constructor(assetStore: AssetStore) {
    this.assetStore = assetStore;
  }

  /**
   * Computes the inner signing data for a state:
   * abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
   * Returns the channelId and signing data.
   */
  private async packSigningData(state: State): Promise<{ channelId: `0x${string}`; signingData: Hex }> {
    if (!state.homeChannelId) {
      throw new Error('state.homeChannelId is required for packing');
    }

    const channelId = state.homeChannelId as `0x${string}`;
    const metadata = getStateTransitionHash(state.transition);

    const homeDecimals = await this.assetStore.getTokenDecimals(state.homeLedger.blockchainId, state.homeLedger.tokenAddress);

    const homeLedger: ContractLedger = {
      chainId: state.homeLedger.blockchainId,
      token: state.homeLedger.tokenAddress,
      decimals: homeDecimals,
      userAllocation: decimalToBigInt(state.homeLedger.userBalance, homeDecimals),
      userNetFlow: decimalToBigInt(state.homeLedger.userNetFlow, homeDecimals),
      nodeAllocation: decimalToBigInt(state.homeLedger.nodeBalance, homeDecimals),
      nodeNetFlow: decimalToBigInt(state.homeLedger.nodeNetFlow, homeDecimals),
    };

    let nonHomeLedger: ContractLedger;

    if (state.escrowLedger) {
      const escrowDecimals = await this.assetStore.getTokenDecimals(state.escrowLedger.blockchainId, state.escrowLedger.tokenAddress);

      nonHomeLedger = {
        chainId: state.escrowLedger.blockchainId,
        token: state.escrowLedger.tokenAddress,
        decimals: escrowDecimals,
        userAllocation: decimalToBigInt(state.escrowLedger.userBalance, escrowDecimals),
        userNetFlow: decimalToBigInt(state.escrowLedger.userNetFlow, escrowDecimals),
        nodeAllocation: decimalToBigInt(state.escrowLedger.nodeBalance, escrowDecimals),
        nodeNetFlow: decimalToBigInt(state.escrowLedger.nodeNetFlow, escrowDecimals),
      };
    } else {
      nonHomeLedger = {
        chainId: 0n,
        token: '0x0000000000000000000000000000000000000000' as Address,
        decimals: 0,
        userAllocation: 0n,
        userNetFlow: 0n,
        nodeAllocation: 0n,
        nodeNetFlow: 0n,
      };
    }

    const intent = transitionToIntent(state.transition);

    const ledgerComponents = [
      { name: 'chainId', type: 'uint64' },
      { name: 'token', type: 'address' },
      { name: 'decimals', type: 'uint8' },
      { name: 'userAllocation', type: 'uint256' },
      { name: 'userNetFlow', type: 'int256' },
      { name: 'nodeAllocation', type: 'uint256' },
      { name: 'nodeNetFlow', type: 'int256' },
    ] as const;

    const signingData = encodeAbiParameters(
      [{ type: 'uint64' }, { type: 'uint8' }, { type: 'bytes32' }, { type: 'tuple', components: ledgerComponents }, { type: 'tuple', components: ledgerComponents }],
      [state.version, intent, metadata as `0x${string}`, homeLedger, nonHomeLedger],
    );

    return { channelId, signingData };
  }

  /**
   * Wraps signing data with channelId: abi.encode(channelId, signingData)
   */
  private packWithChannelId(channelId: `0x${string}`, signingData: Hex): `0x${string}` {
    return encodeAbiParameters([{ type: 'bytes32' }, { type: 'bytes' }], [channelId, signingData]);
  }

  /**
   * PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
   * This matches the Solidity contract's two-step encoding:
   *
   *   signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
   *   message = abi.encode(channelId, signingData)
   *
   * @param state - State to pack
   * @returns Packed bytes as hex string
   */
  async packState(state: State): Promise<`0x${string}`> {
    const { channelId, signingData } = await this.packSigningData(state);
    return this.packWithChannelId(channelId, signingData);
  }

  /**
   * PackChallengeState encodes a state for challenge signature verification.
   * This matches the Solidity contract's challenge validation:
   *
   *   challengerSigningData = abi.encodePacked(abi.encode(version, intent, metadata, homeLedger, nonHomeLedger), "challenge")
   *   message = abi.encode(channelId, challengerSigningData)
   *
   * @param state - State to pack for challenge
   * @returns Packed challenge bytes as hex string
   */
  async packChallengeState(state: State): Promise<`0x${string}`> {
    const { channelId, signingData } = await this.packSigningData(state);
    const challengeSigningData = concat([signingData, toHex('challenge')]);
    return this.packWithChannelId(channelId, challengeSigningData);
  }
}

/**
 * NewStatePackerV1 creates a new state packer instance
 * @param assetStore - Asset store for retrieving token metadata
 * @returns StatePackerV1 instance
 */
export function newStatePackerV1(assetStore: AssetStore): StatePackerV1 {
  return new StatePackerV1(assetStore);
}

/**
 * PackState is a convenience function that creates a StatePackerV1 and packs the state.
 * For production use, create a StatePackerV1 instance and reuse it.
 * @param state - State to pack
 * @param assetStore - Asset store for retrieving token metadata
 * @returns Packed bytes as hex string
 */
export async function packState(state: State, assetStore: AssetStore): Promise<`0x${string}`> {
  const packer = newStatePackerV1(assetStore);
  return packer.packState(state);
}

/**
 * PackChallengeState is a convenience function that creates a StatePackerV1 and packs the challenge state.
 * For production use, create a StatePackerV1 instance and reuse it.
 * @param state - State to pack for challenge
 * @param assetStore - Asset store for retrieving token metadata
 * @returns Packed challenge bytes as hex string
 */
export async function packChallengeState(state: State, assetStore: AssetStore): Promise<`0x${string}`> {
  const packer = newStatePackerV1(assetStore);
  return packer.packChallengeState(state);
}
