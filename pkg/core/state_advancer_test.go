package core

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMutualLockState(t *testing.T, amount decimal.Decimal) *State {
	t.Helper()
	userWallet := "0xUser"
	asset := "USDC"
	chanID := "0xHomeChannelId"

	state := NewVoidState(asset, userWallet)
	state.Version = 5
	state.HomeChannelID = &chanID
	state.ID = GetStateID(userWallet, asset, 0, 5)
	state.HomeLedger.TokenAddress = "0xToken"
	state.HomeLedger.BlockchainID = 1

	_, err := state.ApplyMutualLockTransition(2, "0xForeignToken", amount)
	require.NoError(t, err)

	sig := "0xSig"
	state.UserSig = &sig
	state.NodeSig = &sig

	return state
}

func newEscrowLockState(t *testing.T, amount decimal.Decimal) *State {
	t.Helper()
	userWallet := "0xUser"
	asset := "USDC"
	chanID := "0xHomeChannelId"

	state := NewVoidState(asset, userWallet)
	state.Version = 5
	state.HomeChannelID = &chanID
	state.ID = GetStateID(userWallet, asset, 0, 5)
	state.HomeLedger.TokenAddress = "0xToken"
	state.HomeLedger.BlockchainID = 1
	state.HomeLedger.UserBalance = amount

	_, err := state.ApplyEscrowLockTransition(2, "0xForeignToken", amount)
	require.NoError(t, err)

	sig := "0xSig"
	state.UserSig = &sig
	state.NodeSig = &sig

	return state
}

func TestValidateAdvancement_StrictTransitionOrdering(t *testing.T) {
	t.Parallel()

	advancer := NewStateAdvancerV1(newMockAssetStore())
	amount := decimal.NewFromInt(10)
	sig := "0xSig"

	t.Run("reject_non_escrow_deposit_after_mutual_lock", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		proposed := mutualLockState.NextState()
		_, err := proposed.ApplyHomeDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		err = advancer.ValidateAdvancement(*mutualLockState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "after mutual lock, only escrow deposit is allowed")
	})

	t.Run("reject_non_escrow_withdraw_after_escrow_lock", func(t *testing.T) {
		t.Parallel()
		escrowLockState := newEscrowLockState(t, amount)

		proposed := escrowLockState.NextState()
		_, err := proposed.ApplyHomeDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		err = advancer.ValidateAdvancement(*escrowLockState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "after escrow lock, only escrow withdraw is allowed")
	})
}

func TestValidateAdvancement_EscrowDeposit(t *testing.T) {
	t.Parallel()

	advancer := NewStateAdvancerV1(newMockAssetStore())
	amount := decimal.NewFromInt(10)
	sig := "0xSig"

	t.Run("success_valid_escrow_deposit", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		proposed := mutualLockState.NextState()
		_, err := proposed.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		err = advancer.ValidateAdvancement(*mutualLockState, *proposed)
		assert.NoError(t, err)
	})

	t.Run("reject_non_zero_home_node_balance", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		proposed := mutualLockState.NextState()
		_, err := proposed.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		proposed.HomeLedger.NodeBalance = amount

		err = advancer.ValidateAdvancement(*mutualLockState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "home ledger mismatch")
	})

	t.Run("reject_increased_home_node_net_flow", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		proposed := mutualLockState.NextState()
		_, err := proposed.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		proposed.HomeLedger.NodeNetFlow = proposed.HomeLedger.NodeNetFlow.Add(amount)

		err = advancer.ValidateAdvancement(*mutualLockState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "home ledger mismatch")
	})

	t.Run("reject_amount_mismatch_with_mutual_lock", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		proposed := mutualLockState.NextState()
		_, err := proposed.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig
		// Attempt escrow deposit with a different amount than the mutual lock.
		proposed.Transition.Amount = decimal.NewFromInt(99)

		err = advancer.ValidateAdvancement(*mutualLockState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "escrow deposit amount must be the same as mutual lock amount")
	})

	t.Run("reject_escrow_deposit_not_after_mutual_lock", func(t *testing.T) {
		t.Parallel()
		mutualLockState := newMutualLockState(t, amount)

		homeDepositState := mutualLockState.NextState()
		_, err := homeDepositState.ApplyHomeDepositTransition(amount)
		require.NoError(t, err)
		homeDepositState.UserSig = &sig

		proposed := homeDepositState.NextState()
		_, err = proposed.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		proposed.UserSig = &sig

		err = advancer.ValidateAdvancement(*homeDepositState, *proposed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "escrow deposit transition must follow a mutual lock transition")
	})
}

func TestValidateAdvancement_RejectsInvalidAmount(t *testing.T) {
	t.Parallel()

	advancer := NewStateAdvancerV1(newMockAssetStore())
	sig := "0xSig"
	chanID := "0xChannel"

	newCurrentState := func() State {
		s := NewVoidState("USDC", "0xUser")
		s.HomeChannelID = &chanID
		s.ID = GetStateID("0xUser", "USDC", 0, 0)
		return *s
	}

	cases := []struct {
		name           string
		transitionType TransitionType
		invalidAmounts []decimal.Decimal
		errContains    string
	}{
		// Acknowledgement: amount must be exactly zero
		{
			"Acknowledgement",
			TransitionTypeAcknowledgement,
			[]decimal.Decimal{decimal.NewFromInt(1), decimal.NewFromInt(-1)},
			"must be zero",
		},
		// Finalize: amount must not be negative (zero and positive are allowed)
		{
			"Finalize",
			TransitionTypeFinalize,
			[]decimal.Decimal{decimal.NewFromInt(-1)},
			"must not be negative",
		},
		// All remaining transitions: amount must be strictly positive
		{"HomeDeposit", TransitionTypeHomeDeposit, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"HomeWithdrawal", TransitionTypeHomeWithdrawal, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"TransferSend", TransitionTypeTransferSend, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"Commit", TransitionTypeCommit, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"MutualLock", TransitionTypeMutualLock, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"EscrowDeposit", TransitionTypeEscrowDeposit, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"EscrowLock", TransitionTypeEscrowLock, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"EscrowWithdraw", TransitionTypeEscrowWithdraw, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
		{"Migrate", TransitionTypeMigrate, []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-1)}, "must be positive"},
	}

	for _, tc := range cases {
		for _, invalidAmount := range tc.invalidAmounts {
			t.Run(tc.name+"/"+invalidAmount.String(), func(t *testing.T) {
				t.Parallel()

				current := newCurrentState()
				proposed := current.NextState()
				proposed.Transition = *NewTransition(tc.transitionType, "0xTxID", "0xAccountID", invalidAmount)
				proposed.UserSig = &sig

				err := advancer.ValidateAdvancement(current, *proposed)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			})
		}
	}
}
