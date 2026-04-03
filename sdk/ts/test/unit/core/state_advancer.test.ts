import Decimal from 'decimal.js';
import { Address, Hex } from 'viem';
import { StateAdvancerV1 } from '../../../src/core/state_advancer';
import { AssetStore } from '../../../src/core/interface';
import { State, TransitionType, newVoidState, newTransition } from '../../../src/core/types';
import {
  nextState,
  applyMutualLockTransition,
  applyEscrowDepositTransition,
  applyHomeDepositTransition,
} from '../../../src/core/state';
import { getStateId } from '../../../src/core/utils';

class MockAssetStore implements AssetStore {
  async getAssetDecimals(_asset: string): Promise<number> {
    return 0;
  }

  async getTokenDecimals(_blockchainId: bigint, _tokenAddress: Address): Promise<number> {
    return 0;
  }
}

const USER_WALLET = '0x1111111111111111111111111111111111111111' as Address;
const HOME_CHANNEL_ID = '0x2222222222222222222222222222222222222222222222222222222222222222';
const TOKEN_ADDRESS = '0x3333333333333333333333333333333333333333' as Address;
const FOREIGN_TOKEN_ADDRESS = '0x4444444444444444444444444444444444444444' as Address;

function newMutualLockState(amount: Decimal): State {
  const state = newVoidState('USDC', USER_WALLET);
  state.version = 5n;
  state.homeChannelId = HOME_CHANNEL_ID;
  state.id = getStateId(USER_WALLET, 'USDC', 0n, 5n);
  state.homeLedger.tokenAddress = TOKEN_ADDRESS;
  state.homeLedger.blockchainId = 1n;

  applyMutualLockTransition(state, 2n, FOREIGN_TOKEN_ADDRESS, amount);

  state.userSig = '0xSig' as Hex;
  state.nodeSig = '0xSig' as Hex;

  return state;
}

describe('ValidateAdvancement_EscrowDeposit', () => {
  const advancer = new StateAdvancerV1(new MockAssetStore());
  const amount = new Decimal(10);

  test('success_valid_escrow_deposit', async () => {
    const mutualLockState = newMutualLockState(amount);
    const proposed = nextState(mutualLockState);
    applyEscrowDepositTransition(proposed, amount);

    await expect(advancer.validateAdvancement(mutualLockState, proposed)).resolves.toBeUndefined();
  });

  test('reject_tampered_home_node_balance', async () => {
    const mutualLockState = newMutualLockState(amount);
    const proposed = nextState(mutualLockState);
    applyEscrowDepositTransition(proposed, amount);
    proposed.homeLedger.nodeBalance = proposed.homeLedger.nodeBalance.add(new Decimal(1));

    await expect(advancer.validateAdvancement(mutualLockState, proposed)).rejects.toThrow(
      'home ledger mismatch'
    );
  });

  test('reject_increased_home_node_net_flow', async () => {
    const mutualLockState = newMutualLockState(amount);
    const proposed = nextState(mutualLockState);
    applyEscrowDepositTransition(proposed, amount);
    proposed.homeLedger.nodeNetFlow = proposed.homeLedger.nodeNetFlow.add(amount);

    await expect(advancer.validateAdvancement(mutualLockState, proposed)).rejects.toThrow(
      'home ledger mismatch'
    );
  });

  test('reject_amount_mismatch_with_mutual_lock', async () => {
    const mutualLockState = newMutualLockState(amount);
    const proposed = nextState(mutualLockState);
    applyEscrowDepositTransition(proposed, amount);
    proposed.transition.amount = new Decimal(99);

    await expect(advancer.validateAdvancement(mutualLockState, proposed)).rejects.toThrow(
      'escrow deposit amount must be the same as mutual lock amount'
    );
  });

  test('reject_escrow_deposit_not_after_mutual_lock', async () => {
    const mutualLockState = newMutualLockState(amount);

    const homeDepositState = nextState(mutualLockState);
    applyHomeDepositTransition(homeDepositState, amount);
    homeDepositState.userSig = '0xSig' as Hex;

    const proposed = nextState(homeDepositState);
    applyEscrowDepositTransition(proposed, amount);

    await expect(advancer.validateAdvancement(homeDepositState, proposed)).rejects.toThrow(
      'escrow deposit transition must follow a mutual lock transition'
    );
  });
});

describe('ValidateAdvancement_RejectsInvalidAmount', () => {
  const advancer = new StateAdvancerV1(new MockAssetStore());
  const chanId = '0xChannel';

  const newCurrentState = (): State => {
    const s = newVoidState('USDC', USER_WALLET);
    s.homeChannelId = chanId;
    s.id = getStateId(USER_WALLET, 'USDC', 0n, 0n);
    return s;
  };

  const cases: {
    name: string;
    transitionType: TransitionType;
    invalidAmounts: Decimal[];
    errContains: string;
  }[] = [
    // Acknowledgement: amount must be exactly zero
    {
      name: 'Acknowledgement',
      transitionType: TransitionType.Acknowledgement,
      invalidAmounts: [new Decimal(1), new Decimal(-1)],
      errContains: 'must be zero',
    },
    // Finalize: amount must not be negative (zero and positive are allowed)
    {
      name: 'Finalize',
      transitionType: TransitionType.Finalize,
      invalidAmounts: [new Decimal(-1)],
      errContains: 'must not be negative',
    },
    // All remaining transitions: amount must be strictly positive
    { name: 'HomeDeposit', transitionType: TransitionType.HomeDeposit, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'HomeWithdrawal', transitionType: TransitionType.HomeWithdrawal, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'TransferSend', transitionType: TransitionType.TransferSend, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'Commit', transitionType: TransitionType.Commit, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'MutualLock', transitionType: TransitionType.MutualLock, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'EscrowDeposit', transitionType: TransitionType.EscrowDeposit, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'EscrowLock', transitionType: TransitionType.EscrowLock, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'EscrowWithdraw', transitionType: TransitionType.EscrowWithdraw, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
    { name: 'Migrate', transitionType: TransitionType.Migrate, invalidAmounts: [new Decimal(0), new Decimal(-1)], errContains: 'must be positive' },
  ];

  cases.forEach((tc) => {
    tc.invalidAmounts.forEach((invalidAmount) => {
      test(`${tc.name}/${invalidAmount.toString()}`, async () => {
        const current = newCurrentState();
        const proposed = nextState(current);
        proposed.transition = newTransition(tc.transitionType, '0xTxID', '0xAccountID', invalidAmount);

        await expect(advancer.validateAdvancement(current, proposed)).rejects.toThrow(tc.errContains);
      });
    });
  });
});
