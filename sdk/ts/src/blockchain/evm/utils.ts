/**
 * Utility functions for EVM blockchain interactions
 */

import { Address, Hex, hexToBytes } from 'viem';
import Decimal from 'decimal.js';
import * as core from '../../core/types';
import { decimalToBigInt, generateChannelMetadata, getStateTransitionHash } from '../../core/utils';
import { ChannelDefinition, Ledger, State } from './types';

/**
 * hexToBytes32 converts a hex string to a 32-byte array
 */
export function hexToBytes32(s: string): Uint8Array {
  const bytes = hexToBytes(s as Hex);
  if (bytes.length !== 32) {
    throw new Error(`invalid length: expected 32 bytes, got ${bytes.length}`);
  }
  return bytes;
}

/**
 * coreDefToContractDef converts a core channel definition to a contract channel definition
 */
export function coreDefToContractDef(def: core.ChannelDefinition, asset: string, userWallet: Address, nodeAddress: Address): ChannelDefinition {
  return {
    challengeDuration: def.challenge,
    user: userWallet,
    node: nodeAddress,
    nonce: def.nonce,
    approvedSignatureValidators: BigInt(def.approvedSigValidators || '0x00'),
    metadata: generateChannelMetadata(asset) as `0x${string}`,
  };
}

/**
 * coreStateToContractState converts a core state to a contract state
 */
export async function coreStateToContractState(state: core.State, tokenGetter: (blockchainId: bigint, tokenAddress: Address) => Promise<number>): Promise<State> {
  const homeDecimals = await tokenGetter(state.homeLedger.blockchainId, state.homeLedger.tokenAddress);

  const homeLedger = coreLedgerToContractLedger(state.homeLedger, homeDecimals);

  let nonHomeLedger: Ledger;
  if (state.escrowLedger) {
    const nonHomeDecimals = await tokenGetter(state.escrowLedger.blockchainId, state.escrowLedger.tokenAddress);
    nonHomeLedger = coreLedgerToContractLedger(state.escrowLedger, nonHomeDecimals);
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

  const intent = transitionTypeToIntent(state.transition.type);

  const metadata = getStateTransitionHash(state.transition) as `0x${string}`;

  const userSig = state.userSig ? (state.userSig as `0x${string}`) : ('0x' as `0x${string}`);
  const nodeSig = state.nodeSig ? (state.nodeSig as `0x${string}`) : ('0x' as `0x${string}`);

  return {
    version: state.version,
    intent,
    metadata,
    homeLedger: homeLedger,
    nonHomeLedger: nonHomeLedger,
    userSig,
    nodeSig,
  };
}

/**
 * coreLedgerToContractLedger converts a core ledger to a contract ledger
 */
export function coreLedgerToContractLedger(ledger: core.Ledger, decimals: number): Ledger {
  const userAllocation = decimalToBigInt(ledger.userBalance, decimals);
  const userNetFlow = decimalToBigInt(ledger.userNetFlow, decimals);
  const nodeAllocation = decimalToBigInt(ledger.nodeBalance, decimals);
  const nodeNetFlow = decimalToBigInt(ledger.nodeNetFlow, decimals);

  return {
    chainId: ledger.blockchainId,
    token: ledger.tokenAddress,
    decimals,
    userAllocation,
    userNetFlow,
    nodeAllocation,
    nodeNetFlow,
  };
}

/**
 * contractStateToCoreState converts a contract state to a core state
 */
export function contractStateToCoreState(contractState: State, homeChannelId: string, escrowChannelId?: string): core.State {
  const homeLedger = contractLedgerToCoreLedger(contractState.homeLedger);

  let escrowLedger: core.Ledger | undefined;
  if (contractState.nonHomeLedger.chainId !== 0n) {
    escrowLedger = contractLedgerToCoreLedger(contractState.nonHomeLedger);
  }

  const homeChannelIdPtr = homeChannelId || undefined;
  const escrowChannelIdPtr = escrowChannelId || undefined;

  let userSig: Hex | undefined;
  let nodeSig: Hex | undefined;

  if (contractState.userSig && contractState.userSig !== '0x') {
    userSig = contractState.userSig as Hex;
  }
  if (contractState.nodeSig && contractState.nodeSig !== '0x') {
    nodeSig = contractState.nodeSig as Hex;
  }

  return {
    id: '',
    transition: { type: core.TransitionType.Void, txId: '', amount: new Decimal(0) },
    asset: '',
    userWallet: '0x0000000000000000000000000000000000000000' as Address,
    epoch: 0n,
    version: contractState.version,
    homeChannelId: homeChannelIdPtr,
    escrowChannelId: escrowChannelIdPtr,
    homeLedger,
    escrowLedger,
    userSig,
    nodeSig,
  };
}

/**
 * contractLedgerToCoreLedger converts a contract ledger to a core ledger
 */
export function contractLedgerToCoreLedger(ledger: Ledger): core.Ledger {
  const exp = -ledger.decimals;
  return {
    blockchainId: ledger.chainId,
    tokenAddress: ledger.token,
    userBalance: new Decimal(ledger.userAllocation.toString()).mul(Decimal.pow(10, exp)),
    userNetFlow: new Decimal(ledger.userNetFlow.toString()).mul(Decimal.pow(10, exp)),
    nodeBalance: new Decimal(ledger.nodeAllocation.toString()).mul(Decimal.pow(10, exp)),
    nodeNetFlow: new Decimal(ledger.nodeNetFlow.toString()).mul(Decimal.pow(10, exp)),
  };
}

/**
 * transitionTypeToIntent maps a transition type to an intent value for the contract
 */
function transitionTypeToIntent(transitionType: core.TransitionType): number {
  switch (transitionType) {
    case core.TransitionType.HomeDeposit:
      return core.INTENT_DEPOSIT;
    case core.TransitionType.HomeWithdrawal:
      return core.INTENT_WITHDRAW;
    case core.TransitionType.Finalize:
      return core.INTENT_CLOSE;
    case core.TransitionType.EscrowDeposit:
      return core.INTENT_INITIATE_ESCROW_DEPOSIT;
    case core.TransitionType.EscrowWithdraw:
      return core.INTENT_INITIATE_ESCROW_WITHDRAWAL;
    case core.TransitionType.Migrate:
      return core.INTENT_INITIATE_MIGRATION;
    default:
      return core.INTENT_OPERATE;
  }
}
