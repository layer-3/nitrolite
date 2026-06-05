package core

import (
	"math/big"
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

// TestValidateAdvancement_RejectsOverflowDeposit ensures a home_deposit whose
// scaled amount exceeds uint256 is rejected before storage. Without the bound
// check, the offchain ledger would record the full amount while the signed
// payload truncates to the low 256 bits.
func TestValidateAdvancement_RejectsOverflowDeposit(t *testing.T) {
	t.Parallel()

	store := newMockAssetStore()
	store.AddToken(1, "0xToken", 0) // decimals=0 so amount maps 1:1 to scaled
	advancer := NewStateAdvancerV1(store)

	userWallet := "0xUser"
	asset := "USDC"
	chanID := "0xHomeChannelId"
	sig := "0xSig"

	current := NewVoidState(asset, userWallet)
	current.HomeChannelID = &chanID
	current.ID = GetStateID(userWallet, asset, 0, 0)
	current.HomeLedger.TokenAddress = "0xToken"
	current.HomeLedger.BlockchainID = 1

	// scaled = 2^256, one above the uint256 range.
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)
	amount := decimal.NewFromBigInt(overflow, 0)

	proposed := current.NextState()
	_, err := proposed.ApplyHomeDepositTransition(amount)
	require.NoError(t, err)
	proposed.UserSig = &sig

	err = advancer.ValidateAdvancement(*current, *proposed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid home ledger")
	assert.Contains(t, err.Error(), "uint256")
}

// TestValidateAdvancement_RejectsOverflowNetFlow checks that net-flow values
// exceeding the int256 range are rejected. Reached by combining balances that
// fit uint256 with a net flow that would overflow int256 (the signed type).
func TestValidateAdvancement_RejectsOverflowNetFlow(t *testing.T) {
	t.Parallel()

	store := newMockAssetStore()
	store.AddToken(1, "0xToken", 0)
	advancer := NewStateAdvancerV1(store)

	userWallet := "0xUser"
	asset := "USDC"
	chanID := "0xHomeChannelId"
	sig := "0xSig"

	current := NewVoidState(asset, userWallet)
	current.HomeChannelID = &chanID
	current.ID = GetStateID(userWallet, asset, 0, 0)
	current.HomeLedger.TokenAddress = "0xToken"
	current.HomeLedger.BlockchainID = 1

	// 2^255, one above max int256
	overInt := new(big.Int).Lsh(big.NewInt(1), 255)
	amount := decimal.NewFromBigInt(overInt, 0)

	proposed := current.NextState()
	_, err := proposed.ApplyHomeDepositTransition(amount)
	require.NoError(t, err)
	proposed.UserSig = &sig

	err = advancer.ValidateAdvancement(*current, *proposed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid home ledger")
	assert.Contains(t, err.Error(), "int256")
}

// TestValidateAdvancement_RejectsOverflowEscrowLedger covers the escrow branch
// of the bound check: a state whose EscrowLedger already carries an out-of-range
// UserBalance is rejected even when the home ledger and the new transition are
// otherwise valid.
func TestValidateAdvancement_RejectsOverflowEscrowLedger(t *testing.T) {
	t.Parallel()

	store := newMockAssetStore()
	store.AddToken(1, "0xHomeToken", 0)
	store.AddToken(2, "0xEscrowToken", 0)
	advancer := NewStateAdvancerV1(store)

	userWallet := "0xUser"
	asset := "USDC"
	homeChanID := "0xHomeChannelId"
	escrowChanID := "0xEscrowChannelId"
	sig := "0xSig"

	// 2^256, one above the uint256 range.
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)
	overflowDec := decimal.NewFromBigInt(overflow, 0)

	current := NewVoidState(asset, userWallet)
	current.HomeChannelID = &homeChanID
	current.EscrowChannelID = &escrowChanID
	current.ID = GetStateID(userWallet, asset, 0, 0)
	current.HomeLedger.TokenAddress = "0xHomeToken"
	current.HomeLedger.BlockchainID = 1
	// Last transition is acknowledgement so NextState carries the escrow ledger
	// forward and the new transition is not constrained.
	current.Transition = *NewTransition(TransitionTypeAcknowledgement, "0x0", "0x0", decimal.Zero)
	// Escrow balances sum to net flows so the equality check passes; the
	// uint256 bound is what should reject the state.
	current.EscrowLedger = &Ledger{
		BlockchainID: 2,
		TokenAddress: "0xEscrowToken",
		UserBalance:  overflowDec,
		UserNetFlow:  overflowDec,
		NodeBalance:  decimal.Zero,
		NodeNetFlow:  decimal.Zero,
	}

	proposed := current.NextState()
	_, err := proposed.ApplyHomeDepositTransition(decimal.NewFromInt(1))
	require.NoError(t, err)
	proposed.UserSig = &sig

	err = advancer.ValidateAdvancement(*current, *proposed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid escrow ledger")
	assert.Contains(t, err.Error(), "uint256")
}

// TestValidateAdvancement_RejectsForgedTxID pins the validation boundary that
// NewTransactionFromTransition now relies on: the advancer recomputes the
// canonical TxID from the proposed state and rejects any transition carrying a
// forged TxID, even when the type, account, and amount are otherwise valid.
func TestValidateAdvancement_RejectsForgedTxID(t *testing.T) {
	t.Parallel()

	advancer := NewStateAdvancerV1(newMockAssetStore())
	sig := "0xSig"
	chanID := "0xChannel"

	current := NewVoidState("USDC", "0xUser")
	current.HomeChannelID = &chanID
	current.ID = GetStateID("0xUser", "USDC", 0, 0)

	proposed := current.NextState()
	_, err := proposed.ApplyHomeDepositTransition(decimal.NewFromInt(10))
	require.NoError(t, err)
	proposed.UserSig = &sig

	// Forge the TxID while leaving the type, account, and amount intact, so the
	// only divergence from the recomputed transition is the transaction ID.
	proposed.Transition.TxID = "0xforgedtxid"

	err = advancer.ValidateAdvancement(*current, *proposed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transaction ID mismatch")
}
