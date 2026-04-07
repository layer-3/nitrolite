import { Address } from 'viem';
import { Decimal } from 'decimal.js';
import {
  State,
  Ledger,
  Transition,
  TransitionType,
  ChannelDefinition,
  newTransition,
} from './types.js';
import { getHomeChannelId, getEscrowChannelId, getStateId, getSenderTransactionId, getReceiverTransactionId } from './utils.js';

// ============================================================================
// State Query Methods
// ============================================================================

/**
 * GetLastTransition returns the transition, or null if the state has a void transition
 * or if the transition is a receive type (TransferReceive or Release)
 * @param state - The state to query
 * @returns The transition or null
 */
export function getLastTransition(state: State): Transition | null {
  if (state.transition.type === TransitionType.Void) {
    return null;
  }

  if (
    state.transition.type === TransitionType.TransferReceive ||
    state.transition.type === TransitionType.Release
  ) {
    return null;
  }

  return state.transition;
}

/**
 * IsFinal checks if the state has been finalized
 * @param state - The state to check
 * @returns True if the state is final
 */
export function isFinal(state: State): boolean {
  return state.transition.type === TransitionType.Finalize;
}

// ============================================================================
// State Generation
// ============================================================================

/**
 * NextState generates the next state after incrementing epoch or version
 * @param state - Current state
 * @returns New state for next transition
 */
export function nextState(state: State): State {
  let newState: State;

  const voidTransition: Transition = { type: TransitionType.Void, txId: '', amount: new Decimal(0) };

  if (isFinal(state)) {
    // After finalization, increment epoch and reset version
    newState = {
      id: '',
      transition: voidTransition,
      asset: state.asset,
      userWallet: state.userWallet,
      epoch: state.epoch + 1n,
      version: 0n,
      homeChannelId: undefined,
      escrowChannelId: undefined,
      homeLedger: {
        tokenAddress: '0x0' as Address,
        blockchainId: 0n,
        userBalance: new Decimal(0),
        userNetFlow: new Decimal(0),
        nodeBalance: new Decimal(0),
        nodeNetFlow: new Decimal(0),
      },
      escrowLedger: undefined,
    };
  } else {
    // Normal advancement, increment version
    newState = {
      id: '',
      transition: voidTransition,
      asset: state.asset,
      userWallet: state.userWallet,
      epoch: state.epoch,
      version: state.version + 1n,
      homeChannelId: state.homeChannelId,
      escrowChannelId: state.escrowChannelId,
      homeLedger: { ...state.homeLedger },
      escrowLedger: undefined,
    };
  }

  // Copy escrow ledger if it exists
  if (state.escrowLedger) {
    newState.escrowLedger = { ...state.escrowLedger };

    if (!state.userSig) {
      // Copy transition if user hasn't signed yet
      newState.transition = { ...state.transition };
    } else {
      if (
        state.transition.type === TransitionType.EscrowDeposit ||
        state.transition.type === TransitionType.EscrowWithdraw
      ) {
        // Clear escrow channel and ledger after escrow operations complete
        newState.escrowChannelId = undefined;
        newState.escrowLedger = undefined;
      }
    }
  }

  // Calculate state ID
  newState.id = getStateId(newState.userWallet, newState.asset, newState.epoch, newState.version);

  return newState;
}

// ============================================================================
// State Mutations
// ============================================================================

/**
 * ApplyChannelCreation applies channel creation parameters to the state and returns the calculated home channel ID
 * @param state - State to modify (mutated in place)
 * @param channelDef - Channel definition
 * @param blockchainId - Blockchain ID (uint64)
 * @param tokenAddress - Token address
 * @param nodeAddress - Node address
 * @returns Home channel ID
 */
export function applyChannelCreation(
  state: State,
  channelDef: ChannelDefinition,
  blockchainId: bigint,
  tokenAddress: Address,
  nodeAddress: Address
): string {
  // Set home ledger
  state.homeLedger.tokenAddress = tokenAddress;
  state.homeLedger.blockchainId = blockchainId;

  // Calculate home channel ID
  const homeChannelId = getHomeChannelId(
    nodeAddress,
    state.userWallet,
    state.asset,
    channelDef.nonce,
    channelDef.challenge,
    channelDef.approvedSigValidators
  );

  state.homeChannelId = homeChannelId;

  return homeChannelId;
}

/**
 * ApplyReceiverTransitions applies multiple receiver transitions (TransferReceive, Release)
 * @param state - State to modify (mutated in place)
 * @param transitions - Transitions to apply
 */
export function applyReceiverTransitions(state: State, ...transitions: Transition[]): void {
  for (const transition of transitions) {
    switch (transition.type) {
      case TransitionType.TransferReceive:
        if (!transition.accountId) {
          throw new Error('missing account ID for transfer receive transition');
        }
        applyTransferReceiveTransition(state, transition.accountId, transition.amount, transition.txId);
        break;
      case TransitionType.Release:
        if (!transition.accountId) {
          throw new Error('missing account ID for release transition');
        }
        applyReleaseTransition(state, transition.accountId, transition.amount);
        break;
      default:
        throw new Error(`transition '${transition.type}' cannot be applied by receiver`);
    }
  }
}

// ============================================================================
// Individual Transition Applications
// ============================================================================

/**
 * ApplyAcknowledgementTransition applies an acknowledgement transition with zero amount
 * and placeholder txId/accountId. Used when receiving a transfer without an existing state,
 * or to acknowledge channel creation without a deposit.
 * @param state - State to modify (mutated in place)
 * @returns The created transition
 */
export function applyAcknowledgementTransition(state: State): Transition {
  if (state.transition.type !== TransitionType.Void) {
    throw new Error(`state already has a transition: ${state.transition.type}`);
  }

  const zeroHash = '0x0000000000000000000000000000000000000000000000000000000000000000';
  const newTransitionObj = newTransition(TransitionType.Acknowledgement, zeroHash, zeroHash, new Decimal(0));
  state.transition = newTransitionObj;

  return newTransitionObj;
}

/**
 * ApplyHomeDepositTransition applies a home deposit transition
 * @param state - State to modify (mutated in place)
 * @param amount - Amount to deposit
 * @returns The created transition
 */
export function applyHomeDepositTransition(state: State, amount: Decimal): Transition {
  if (!state.homeChannelId) {
    throw new Error('missing home channel ID');
  }

  const accountId = state.homeChannelId;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.HomeDeposit, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userNetFlow = state.homeLedger.userNetFlow.add(newTransitionObj.amount);
  state.homeLedger.userBalance = state.homeLedger.userBalance.add(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyHomeWithdrawalTransition applies a home withdrawal transition
 * @param state - State to modify (mutated in place)
 * @param amount - Amount to withdraw
 * @returns The created transition
 */
export function applyHomeWithdrawalTransition(state: State, amount: Decimal): Transition {
  if (!state.homeChannelId) {
    throw new Error('missing home channel ID');
  }

  const accountId = state.homeChannelId;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.HomeWithdrawal, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userNetFlow = state.homeLedger.userNetFlow.sub(newTransitionObj.amount);
  state.homeLedger.userBalance = state.homeLedger.userBalance.sub(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyTransferSendTransition applies a transfer send transition
 * @param state - State to modify (mutated in place)
 * @param recipient - Recipient wallet address
 * @param amount - Amount to send
 * @returns The created transition
 */
export function applyTransferSendTransition(
  state: State,
  recipient: string,
  amount: Decimal
): Transition {
  const accountId = recipient;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.TransferSend, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userBalance = state.homeLedger.userBalance.sub(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.sub(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyTransferReceiveTransition applies a transfer receive transition
 * @param state - State to modify (mutated in place)
 * @param sender - Sender wallet address
 * @param amount - Amount received
 * @param txId - Transaction ID
 * @returns The created transition
 */
export function applyTransferReceiveTransition(
  state: State,
  sender: string,
  amount: Decimal,
  txId: string
): Transition {
  const accountId = sender;

  const newTransitionObj = newTransition(TransitionType.TransferReceive, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userBalance = state.homeLedger.userBalance.add(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.add(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyCommitTransition applies a commit transition (lock funds in app session)
 * @param state - State to modify (mutated in place)
 * @param accountId - App session ID
 * @param amount - Amount to commit
 * @returns The created transition
 */
export function applyCommitTransition(state: State, accountId: string, amount: Decimal): Transition {
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.Commit, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userBalance = state.homeLedger.userBalance.sub(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.sub(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyReleaseTransition applies a release transition (unlock funds from app session)
 * @param state - State to modify (mutated in place)
 * @param accountId - App session ID
 * @param amount - Amount to release
 * @returns The created transition
 */
export function applyReleaseTransition(state: State, accountId: string, amount: Decimal): Transition {
  const txId = getReceiverTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.Release, txId, accountId, amount);
  state.transition = newTransitionObj;
  state.homeLedger.userBalance = state.homeLedger.userBalance.add(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.add(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyMutualLockTransition applies a mutual lock transition (initiate escrow deposit)
 * @param state - State to modify (mutated in place)
 * @param blockchainId - Blockchain ID for escrow (uint64)
 * @param tokenAddress - Token address for escrow
 * @param amount - Amount to lock
 * @returns The created transition
 */
export function applyMutualLockTransition(
  state: State,
  blockchainId: bigint,
  tokenAddress: Address,
  amount: Decimal
): Transition {
  if (!state.homeChannelId) {
    throw new Error('missing home channel ID');
  }
  if (blockchainId === 0n) {
    throw new Error('invalid blockchain ID');
  }
  if (!tokenAddress || tokenAddress === ('0x0' as Address)) {
    throw new Error('invalid token address');
  }

  const escrowChannelId = getEscrowChannelId(state.homeChannelId, state.version);
  state.escrowChannelId = escrowChannelId;
  const accountId = escrowChannelId;

  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.MutualLock, txId, accountId, amount);
  state.transition = newTransitionObj;

  state.homeLedger.nodeBalance = state.homeLedger.nodeBalance.add(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.add(newTransitionObj.amount);

  state.escrowLedger = {
    blockchainId,
    tokenAddress,
    userBalance: new Decimal(0).add(newTransitionObj.amount),
    userNetFlow: new Decimal(0).add(newTransitionObj.amount),
    nodeBalance: new Decimal(0),
    nodeNetFlow: new Decimal(0),
  };

  return newTransitionObj;
}

/**
 * ApplyEscrowDepositTransition applies an escrow deposit transition (complete escrow deposit)
 * @param state - State to modify (mutated in place)
 * @param amount - Amount to deposit from escrow
 * @returns The created transition
 */
export function applyEscrowDepositTransition(state: State, amount: Decimal): Transition {
  if (!state.escrowChannelId) {
    throw new Error('internal error: escrow channel ID is nil');
  }
  if (!state.escrowLedger) {
    throw new Error('escrow ledger is nil');
  }

  const accountId = state.escrowChannelId;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.EscrowDeposit, txId, accountId, amount);
  state.transition = newTransitionObj;

  state.homeLedger.userBalance = state.homeLedger.userBalance.add(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.add(newTransitionObj.amount);

  state.escrowLedger.userBalance = state.escrowLedger.userBalance.sub(newTransitionObj.amount);
  state.escrowLedger.nodeNetFlow = state.escrowLedger.nodeNetFlow.sub(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyEscrowLockTransition applies an escrow lock transition (initiate escrow withdrawal)
 * @param state - State to modify (mutated in place)
 * @param blockchainId - Blockchain ID for escrow (uint64)
 * @param tokenAddress - Token address for escrow
 * @param amount - Amount to lock for withdrawal
 * @returns The created transition
 */
export function applyEscrowLockTransition(
  state: State,
  blockchainId: bigint,
  tokenAddress: Address,
  amount: Decimal
): Transition {
  if (!state.homeChannelId) {
    throw new Error('missing home channel ID');
  }
  if (blockchainId === 0n) {
    throw new Error('invalid blockchain ID');
  }
  if (!tokenAddress || tokenAddress === ('0x0' as Address)) {
    throw new Error('invalid token address');
  }

  const escrowChannelId = getEscrowChannelId(state.homeChannelId, state.version);
  state.escrowChannelId = escrowChannelId;
  const accountId = escrowChannelId;

  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.EscrowLock, txId, accountId, amount);
  state.transition = newTransitionObj;

  state.escrowLedger = {
    blockchainId,
    tokenAddress,
    userBalance: new Decimal(0),
    userNetFlow: new Decimal(0),
    nodeBalance: new Decimal(0).add(newTransitionObj.amount),
    nodeNetFlow: new Decimal(0).add(newTransitionObj.amount),
  };

  return newTransitionObj;
}

/**
 * ApplyEscrowWithdrawTransition applies an escrow withdrawal transition (complete escrow withdrawal)
 * @param state - State to modify (mutated in place)
 * @param amount - Amount to withdraw to escrow
 * @returns The created transition
 */
export function applyEscrowWithdrawTransition(state: State, amount: Decimal): Transition {
  if (!state.escrowChannelId) {
    throw new Error('internal error: escrow channel ID is nil');
  }
  if (!state.escrowLedger) {
    throw new Error('escrow ledger is nil');
  }

  const accountId = state.escrowChannelId;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.EscrowWithdraw, txId, accountId, amount);
  state.transition = newTransitionObj;

  state.homeLedger.userBalance = state.homeLedger.userBalance.sub(newTransitionObj.amount);
  state.homeLedger.nodeNetFlow = state.homeLedger.nodeNetFlow.sub(newTransitionObj.amount);

  state.escrowLedger.userNetFlow = state.escrowLedger.userNetFlow.sub(newTransitionObj.amount);
  state.escrowLedger.nodeBalance = state.escrowLedger.nodeBalance.sub(newTransitionObj.amount);

  return newTransitionObj;
}

/**
 * ApplyMigrateTransition applies a migrate transition (not implemented yet)
 * @param state - State to modify (mutated in place)
 * @param amount - Amount to migrate
 * @returns The created transition
 */
export function applyMigrateTransition(state: State, amount: Decimal): Transition {
  throw new Error('migrate transition not implemented yet');
}

/**
 * ApplyFinalizeTransition applies a finalize transition (close channel)
 * @param state - State to modify (mutated in place)
 * @returns The created transition
 */
export function applyFinalizeTransition(state: State): Transition {
  if (!state.homeChannelId) {
    throw new Error('missing home channel ID');
  }

  const accountId = state.homeChannelId;
  const amount = state.homeLedger.userBalance;
  const txId = getSenderTransactionId(accountId, state.id);

  const newTransitionObj = newTransition(TransitionType.Finalize, txId, accountId, amount);
  state.transition = newTransitionObj;

  state.homeLedger.userNetFlow = state.homeLedger.userNetFlow.sub(state.homeLedger.userBalance);
  state.homeLedger.userBalance = new Decimal(0);

  return newTransitionObj;
}
