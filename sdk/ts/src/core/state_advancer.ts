import { State, TransitionType, ledgerEqual, validateLedger, transitionsEqual } from './types.js';
import { StateAdvancer, AssetStore } from './interface.js';
import { validateDecimalPrecision } from './utils.js';
import {
  nextState,
  applyAcknowledgementTransition,
  applyHomeDepositTransition,
  applyHomeWithdrawalTransition,
  applyTransferSendTransition,
  applyCommitTransition,
  applyMutualLockTransition,
  applyEscrowDepositTransition,
  applyEscrowLockTransition,
  applyEscrowWithdrawTransition,
  applyMigrateTransition,
  applyFinalizeTransition,
} from './state.js';

export class StateAdvancerV1 implements StateAdvancer {
  constructor(private assetStore: AssetStore) {}

  // ValidateAdvancement validates that the proposed state is a valid advancement of the current state
  //
  // NOTE: User signature is not validated here
  //
  // TODO: Add shared JSON fixture suite consumed by both Go and TS test suites to guarantee validation parity
  async validateAdvancement(currentState: State, proposedState: State): Promise<void> {
    const expectedState = nextState(currentState);

    if (!proposedState.homeChannelId) {
      throw new Error('home channel ID cannot be nil');
    }

    if (!expectedState.homeChannelId) {
      expectedState.homeChannelId = proposedState.homeChannelId;
      expectedState.homeLedger.blockchainId = proposedState.homeLedger.blockchainId;
      expectedState.homeLedger.tokenAddress = proposedState.homeLedger.tokenAddress;
    }

    if (proposedState.homeChannelId !== expectedState.homeChannelId) {
      throw new Error(
        `home channel ID mismatch: expected=${expectedState.homeChannelId}, proposed=${proposedState.homeChannelId}`
      );
    }

    if (proposedState.version !== expectedState.version) {
      throw new Error(
        `version mismatch: expected=${expectedState.version}, proposed=${proposedState.version}`
      );
    }

    if (proposedState.userWallet.toLowerCase() !== expectedState.userWallet.toLowerCase()) {
      throw new Error(
        `user wallet mismatch: expected=${expectedState.userWallet}, proposed=${proposedState.userWallet}`
      );
    }

    if (proposedState.asset !== expectedState.asset) {
      throw new Error(
        `asset mismatch: expected=${expectedState.asset}, proposed=${proposedState.asset}`
      );
    }

    if (proposedState.epoch !== expectedState.epoch) {
      throw new Error(
        `epoch mismatch: expected=${expectedState.epoch}, proposed=${proposedState.epoch}`
      );
    }

    if (proposedState.id !== expectedState.id) {
      throw new Error(
        `state ID mismatch: expected=${expectedState.id}, proposed=${proposedState.id}`
      );
    }

    const newTransition = proposedState.transition;

    const decimals = await this.assetStore.getAssetDecimals(proposedState.asset);
    validateDecimalPrecision(newTransition.amount, decimals);

    switch (newTransition.type) {
      case TransitionType.Acknowledgement:
        if (!newTransition.amount.isZero()) {
          throw new Error(
            `transition amount must be zero, got ${newTransition.amount.toString()}`
          );
        }
        break;
      case TransitionType.Finalize:
        if (newTransition.amount.isNegative()) {
          throw new Error(
            `transition amount must not be negative, got ${newTransition.amount.toString()}`
          );
        }
        break;
      default:
        if (newTransition.amount.isNegative() || newTransition.amount.isZero()) {
          throw new Error(
            `transition amount must be positive, got ${newTransition.amount.toString()}`
          );
        }
    }

    const lastTransition = currentState.transition;

    switch (newTransition.type) {
      case TransitionType.Void:
        throw new Error('cannot apply void transition as new transition');
      case TransitionType.Acknowledgement:
        if (currentState.userSig) {
          throw new Error('current state is already acknowledged');
        }
        applyAcknowledgementTransition(expectedState);
        break;
      case TransitionType.HomeDeposit:
        applyHomeDepositTransition(expectedState, newTransition.amount);
        break;
      case TransitionType.HomeWithdrawal:
        applyHomeWithdrawalTransition(expectedState, newTransition.amount);
        break;
      case TransitionType.TransferSend:
        if (!newTransition.accountId) {
          throw new Error('missing account ID for transfer send transition');
        }
        applyTransferSendTransition(expectedState, newTransition.accountId, newTransition.amount);
        break;
      case TransitionType.Commit:
        if (!newTransition.accountId) {
          throw new Error('missing account ID for commit transition');
        }
        applyCommitTransition(expectedState, newTransition.accountId, newTransition.amount);
        break;
      case TransitionType.MutualLock:
        if (!proposedState.escrowLedger) {
          throw new Error('proposed state escrow ledger is nil');
        }
        applyMutualLockTransition(
          expectedState,
          proposedState.escrowLedger.blockchainId,
          proposedState.escrowLedger.tokenAddress,
          newTransition.amount
        );
        break;
      case TransitionType.EscrowDeposit:
        if (lastTransition.type === TransitionType.MutualLock) {
          if (!lastTransition.amount.equals(newTransition.amount)) {
            throw new Error('escrow deposit amount must be the same as mutual lock amount');
          }
          applyEscrowDepositTransition(expectedState, newTransition.amount);
        } else {
          throw new Error('escrow deposit transition must follow a mutual lock transition');
        }
        break;
      case TransitionType.EscrowLock:
        if (!proposedState.escrowLedger) {
          throw new Error('proposed state escrow ledger is nil');
        }
        applyEscrowLockTransition(
          expectedState,
          proposedState.escrowLedger.blockchainId,
          proposedState.escrowLedger.tokenAddress,
          newTransition.amount
        );
        break;
      case TransitionType.EscrowWithdraw:
        if (lastTransition.type === TransitionType.EscrowLock) {
          if (!lastTransition.amount.equals(newTransition.amount)) {
            throw new Error('escrow withdraw amount must be the same as escrow lock amount');
          }
          applyEscrowWithdrawTransition(expectedState, newTransition.amount);
        } else {
          throw new Error('escrow withdraw transition must follow an escrow lock transition');
        }
        break;
      case TransitionType.Migrate:
        applyMigrateTransition(expectedState, newTransition.amount);
        break;
      case TransitionType.Finalize:
        applyFinalizeTransition(expectedState);
        break;
      default:
        throw new Error(`unsupported type for new transition: ${newTransition.type}`);
    }

    const expectedTransition = expectedState.transition;
    const transitionMismatch = transitionsEqual(expectedTransition, newTransition);
    if (transitionMismatch) {
      throw new Error(`new transition does not match expected: ${transitionMismatch}`);
    }

    const homeLedgerMismatch = ledgerEqual(expectedState.homeLedger, proposedState.homeLedger);
    if (homeLedgerMismatch) {
      throw new Error(`home ledger mismatch: ${homeLedgerMismatch}`);
    }
    validateLedger(proposedState.homeLedger);

    const expectedHasEscrowId = expectedState.escrowChannelId !== undefined;
    const proposedHasEscrowId = proposedState.escrowChannelId !== undefined;
    if (expectedHasEscrowId !== proposedHasEscrowId) {
      throw new Error('escrow channel ID presence mismatch');
    }

    if (expectedState.escrowChannelId && proposedState.escrowChannelId) {
      if (expectedState.escrowChannelId !== proposedState.escrowChannelId) {
        throw new Error(
          `escrow channel ID mismatch: expected=${expectedState.escrowChannelId}, proposed=${proposedState.escrowChannelId}`
        );
      }
    }

    const expectedHasEscrowLedger = expectedState.escrowLedger !== undefined;
    const proposedHasEscrowLedger = proposedState.escrowLedger !== undefined;
    if (expectedHasEscrowLedger !== proposedHasEscrowLedger) {
      throw new Error('escrow ledger presence mismatch');
    }

    if (expectedState.escrowLedger && proposedState.escrowLedger) {
      const escrowLedgerMismatch = ledgerEqual(
        expectedState.escrowLedger,
        proposedState.escrowLedger
      );
      if (escrowLedgerMismatch) {
        throw new Error(`escrow ledger mismatch: expected=${JSON.stringify(expectedState.escrowLedger)}, proposed=${JSON.stringify(proposedState.escrowLedger)}: ${escrowLedgerMismatch}`);
      }
      validateLedger(proposedState.escrowLedger);

      if (proposedState.escrowLedger.blockchainId === proposedState.homeLedger.blockchainId) {
        throw new Error('escrow ledger blockchain ID cannot match home ledger blockchain ID');
      }
    }
  }
}
