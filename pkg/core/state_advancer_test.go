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

func TestValidateAdvancement_RejectsNonPositiveAmount(t *testing.T) {
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

	for _, tc := range []struct {
		name   string
		amount decimal.Decimal
	}{
		{"zero", decimal.Zero},
		{"negative", decimal.NewFromInt(-1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			current := newCurrentState()
			proposed := current.NextState()
			proposed.Transition = *NewTransition(TransitionTypeTransferSend, "0xTxID", "0xRecipient", tc.amount)
			proposed.UserSig = &sig

			err := advancer.ValidateAdvancement(current, *proposed)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must be positive")
		})
	}
}
