package core

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChannel(t *testing.T) {
	t.Parallel()
	channelID := "0xChanID"
	userWallet := "0xUser"
	token := "0xToken"
	ch := NewChannel(channelID, userWallet, "USDC", ChannelTypeHome, 1, token, 1, 100, "0x1")

	assert.Equal(t, channelID, ch.ChannelID)
	assert.Equal(t, userWallet, ch.UserWallet)
	assert.Equal(t, "USDC", ch.Asset)
	assert.Equal(t, ChannelTypeHome, ch.Type)
	assert.Equal(t, uint64(1), ch.BlockchainID)
	assert.Equal(t, token, ch.TokenAddress)
	assert.Equal(t, uint64(1), ch.Nonce)
	assert.Equal(t, uint32(100), ch.ChallengeDuration)
	assert.Equal(t, "0x1", ch.ApprovedSigValidators)
	assert.Equal(t, ChannelStatusVoid, ch.Status)
	assert.Equal(t, uint64(0), ch.StateVersion)
}

func TestNewVoidState(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	assert.Equal(t, "USDC", state.Asset)
	assert.Equal(t, "0xUser", state.UserWallet)
	assert.True(t, state.HomeLedger.UserBalance.IsZero())
}

func TestState_NextState(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	state.Version = 1
	state.Epoch = 1

	// Normal next state
	next := state.NextState()
	assert.Equal(t, uint64(2), next.Version)
	assert.Equal(t, uint64(1), next.Epoch)
	assert.Nil(t, next.EscrowLedger)

	// Final state next state (new epoch)
	state.Transition.Type = TransitionTypeFinalize
	nextFinal := state.NextState()
	assert.Equal(t, uint64(0), nextFinal.Version)
	assert.Equal(t, uint64(2), nextFinal.Epoch)

	// With Escrow Ledger
	state.Transition.Type = TransitionTypeVoid
	state.EscrowLedger = &Ledger{BlockchainID: 2}
	nextEscrow := state.NextState()
	assert.NotNil(t, nextEscrow.EscrowLedger)
	assert.Equal(t, uint64(2), nextEscrow.EscrowLedger.BlockchainID)

	// Escrow cleanup transitions
	state.Transition.Type = TransitionTypeEscrowDeposit
	nextEscrowCleanup := state.NextState()
	assert.Nil(t, nextEscrowCleanup.EscrowLedger)
}

func TestState_ApplyChannelCreation(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	def := ChannelDefinition{Nonce: 1, Challenge: 100, ApprovedSigValidators: "0x1"}

	id, err := state.ApplyChannelCreation(def, 137, "0xToken", "0xNode")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, *state.HomeChannelID)
	assert.Equal(t, uint64(137), state.HomeLedger.BlockchainID)
	assert.Equal(t, "0xToken", state.HomeLedger.TokenAddress)
}

func TestState_ApplyAcknowledgementTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")

	// Success
	transition, err := state.ApplyAcknowledgementTransition()
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeAcknowledgement, transition.Type)
	assert.Equal(t, TransitionTypeAcknowledgement, state.Transition.Type)

	// Fail if transition exists
	_, err = state.ApplyAcknowledgementTransition()
	assert.Error(t, err)
}

func TestState_ApplyHomeDepositTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	chanID := "0xChan"
	state.HomeChannelID = &chanID
	state.ID = "0xStateID" // Required for txID gen

	amount := decimal.NewFromInt(100)

	transition, err := state.ApplyHomeDepositTransition(amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeHomeDeposit, transition.Type)
	assert.Equal(t, amount.String(), state.HomeLedger.UserBalance.String())
	assert.Equal(t, amount.String(), state.HomeLedger.UserNetFlow.String())

	// Fail missing channel ID
	state = NewVoidState("USDC", "0xUser")
	_, err = state.ApplyHomeDepositTransition(amount)
	assert.Error(t, err)
}

func TestState_ApplyHomeWithdrawalTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	chanID := "0xChan"
	state.HomeChannelID = &chanID
	state.ID = "0xStateID"

	// Setup initial balance
	state.HomeLedger.UserBalance = decimal.NewFromInt(100)
	state.HomeLedger.UserNetFlow = decimal.NewFromInt(100)

	amount := decimal.NewFromInt(50)
	transition, err := state.ApplyHomeWithdrawalTransition(amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeHomeWithdrawal, transition.Type)
	assert.Equal(t, "50", state.HomeLedger.UserBalance.String())
	assert.Equal(t, "50", state.HomeLedger.UserNetFlow.String())
}

func TestState_ApplyTransferSendTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	state.ID = "0xStateID"
	state.HomeLedger.UserBalance = decimal.NewFromInt(100)

	amount := decimal.NewFromInt(10)
	recipient := "0xRecipient"

	transition, err := state.ApplyTransferSendTransition(recipient, amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeTransferSend, transition.Type)
	assert.Equal(t, "90", state.HomeLedger.UserBalance.String())
	assert.Equal(t, "-10", state.HomeLedger.NodeNetFlow.String())
}

func TestState_ApplyTransferReceiveTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	amount := decimal.NewFromInt(10)
	sender := "0xSender"
	txID := "0xTx"

	transition, err := state.ApplyTransferReceiveTransition(sender, amount, txID)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeTransferReceive, transition.Type)
	assert.Equal(t, "10", state.HomeLedger.UserBalance.String())
	assert.Equal(t, "10", state.HomeLedger.NodeNetFlow.String())
}

func TestState_ApplyCommitTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	state.ID = "0xStateID"
	state.HomeLedger.UserBalance = decimal.NewFromInt(100)

	amount := decimal.NewFromInt(10)
	appSessionID := "0xAppSession"

	transition, err := state.ApplyCommitTransition(appSessionID, amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeCommit, transition.Type)
	assert.Equal(t, "90", state.HomeLedger.UserBalance.String())
}

func TestState_ApplyReleaseTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	state.ID = "0xStateID"

	amount := decimal.NewFromInt(10)
	appSessionID := "0xAppSession"

	transition, err := state.ApplyReleaseTransition(appSessionID, amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeRelease, transition.Type)
	assert.Equal(t, "10", state.HomeLedger.UserBalance.String())
}

func TestState_ApplyMutualLockTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	chanID := "0xChan"
	state.HomeChannelID = &chanID
	state.ID = "0xStateID"
	state.HomeLedger.UserBalance = decimal.NewFromInt(100)

	amount := decimal.NewFromInt(10)

	transition, err := state.ApplyMutualLockTransition(2, "0xForeignToken", amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeMutualLock, transition.Type)
	assert.NotNil(t, state.EscrowLedger)
	assert.NotNil(t, state.EscrowChannelID)
	assert.Equal(t, "10", state.HomeLedger.NodeBalance.String())
	assert.Equal(t, "10", state.EscrowLedger.UserBalance.String())

	// Failures
	state.Transition.Type = TransitionTypeVoid
	state.HomeChannelID = nil
	_, err = state.ApplyMutualLockTransition(2, "0xT", amount)
	assert.Error(t, err)
}

func TestState_ApplyEscrowDepositTransition(t *testing.T) {
	t.Parallel()

	t.Run("user_balance_increases_escrow_balance_decreases", func(t *testing.T) {
		t.Parallel()
		state := NewVoidState("USDC", "0xUser")
		state.ID = "0xStateID"

		escrowID := "0xEscrow"
		state.EscrowChannelID = &escrowID
		state.EscrowLedger = &Ledger{UserBalance: decimal.NewFromInt(100)}

		amount := decimal.NewFromInt(10)
		transition, err := state.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)
		assert.Equal(t, TransitionTypeEscrowDeposit, transition.Type)
		assert.Equal(t, "10", state.HomeLedger.UserBalance.String())
		assert.Equal(t, "90", state.EscrowLedger.UserBalance.String())
	})

	t.Run("clears_node_balance_and_leaves_net_flows_unchanged", func(t *testing.T) {
		t.Parallel()
		state := NewVoidState("USDC", "0xUser")
		state.ID = "0xStateID"

		escrowID := "0xEscrow"
		state.EscrowChannelID = &escrowID

		// Realistic precondition: state after ApplyMutualLockTransition where the
		// node has locked funds on the home chain.
		state.HomeLedger.NodeBalance = decimal.NewFromInt(10)
		state.HomeLedger.NodeNetFlow = decimal.NewFromInt(10)
		state.EscrowLedger = &Ledger{
			UserBalance: decimal.NewFromInt(10),
			UserNetFlow: decimal.NewFromInt(10),
		}

		prevNodeNetFlow := state.HomeLedger.NodeNetFlow
		prevUserNetFlow := state.HomeLedger.UserNetFlow

		amount := decimal.NewFromInt(10)
		_, err := state.ApplyEscrowDepositTransition(amount)
		require.NoError(t, err)

		// User receives funds on the home chain.
		assert.Equal(t, "10", state.HomeLedger.UserBalance.String())

		// Node's locked allocation must be cleared (on-chain: nodeAllocation == 0).
		assert.Equal(t, "0", state.HomeLedger.NodeBalance.String())

		// Net flows must not change (on-chain: nodeNfDelta == 0, userNfDelta == 0).
		assert.True(t, state.HomeLedger.NodeNetFlow.Equal(prevNodeNetFlow),
			"NodeNetFlow must remain unchanged: got %s, want %s",
			state.HomeLedger.NodeNetFlow.String(), prevNodeNetFlow.String())
		assert.True(t, state.HomeLedger.UserNetFlow.Equal(prevUserNetFlow),
			"UserNetFlow must remain unchanged: got %s, want %s",
			state.HomeLedger.UserNetFlow.String(), prevUserNetFlow.String())
	})
}

func TestState_ApplyEscrowLockTransition(t *testing.T) {
	t.Parallel()

	t.Run("success_creates_escrow_ledger_and_clears_node_balance", func(t *testing.T) {
		t.Parallel()
		state := NewVoidState("USDC", "0xUser")
		chanID := "0xChan"
		state.HomeChannelID = &chanID
		state.ID = "0xStateID"
		state.HomeLedger.UserBalance = decimal.NewFromInt(50)
		state.HomeLedger.NodeBalance = decimal.NewFromInt(5)

		amount := decimal.NewFromInt(10)
		transition, err := state.ApplyEscrowLockTransition(2, "0xT", amount)
		require.NoError(t, err)
		assert.Equal(t, TransitionTypeEscrowLock, transition.Type)
		assert.NotNil(t, state.EscrowLedger)
		assert.Equal(t, "10", state.EscrowLedger.NodeBalance.String())
		assert.True(t, state.HomeLedger.NodeBalance.IsZero(),
			"home NodeBalance must be cleared to zero, got %s", state.HomeLedger.NodeBalance.String())
	})

	t.Run("reject_insufficient_user_balance", func(t *testing.T) {
		t.Parallel()
		state := NewVoidState("USDC", "0xUser")
		chanID := "0xChan"
		state.HomeChannelID = &chanID
		state.ID = "0xStateID"
		state.HomeLedger.UserBalance = decimal.NewFromInt(5)

		amount := decimal.NewFromInt(10)
		_, err := state.ApplyEscrowLockTransition(2, "0xT", amount)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient user balance for escrow lock")
	})
}

func TestState_ApplyEscrowWithdrawTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	state.ID = "0xStateID"

	escrowID := "0xEscrow"
	state.EscrowChannelID = &escrowID
	state.EscrowLedger = &Ledger{NodeBalance: decimal.NewFromInt(100)}
	state.HomeLedger.UserBalance = decimal.NewFromInt(100)

	amount := decimal.NewFromInt(10)
	transition, err := state.ApplyEscrowWithdrawTransition(amount)
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeEscrowWithdraw, transition.Type)
	assert.Equal(t, "90", state.HomeLedger.UserBalance.String())
	assert.Equal(t, "90", state.EscrowLedger.NodeBalance.String())
}

func TestState_ApplyMigrateTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	_, err := state.ApplyMigrateTransition(decimal.Zero)
	assert.Error(t, err) // Not implemented
}

func TestState_ApplyFinalizeTransition(t *testing.T) {
	t.Parallel()
	state := NewVoidState("USDC", "0xUser")
	chanID := "0xChan"
	state.HomeChannelID = &chanID
	state.ID = "0xStateID"

	state.HomeLedger.UserBalance = decimal.NewFromInt(100)
	state.HomeLedger.UserNetFlow = decimal.NewFromInt(200)

	transition, err := state.ApplyFinalizeTransition()
	require.NoError(t, err)
	assert.Equal(t, TransitionTypeFinalize, transition.Type)
	assert.Equal(t, "0", state.HomeLedger.UserBalance.String())
	assert.Equal(t, "100", state.HomeLedger.UserNetFlow.String())
}

func TestLedger_Equal_Validate(t *testing.T) {
	t.Parallel()
	l1 := Ledger{
		TokenAddress: "0xT",
		BlockchainID: 1,
		UserBalance:  decimal.NewFromInt(10),
		NodeBalance:  decimal.NewFromInt(10),
		UserNetFlow:  decimal.NewFromInt(10),
		NodeNetFlow:  decimal.NewFromInt(10),
	}
	l2 := l1
	assert.NoError(t, l1.Equal(l2))
	assert.NoError(t, l1.Validate())

	l2.TokenAddress = "0xOther"
	assert.Error(t, l1.Equal(l2))

	lInvalid := l1
	lInvalid.TokenAddress = ""
	assert.Error(t, lInvalid.Validate())

	lInvalid = l1
	lInvalid.BlockchainID = 0
	assert.Error(t, lInvalid.Validate())

	lInvalid = l1
	lInvalid.UserBalance = decimal.NewFromInt(-1)
	assert.Error(t, lInvalid.Validate())

	lInvalid = l1
	lInvalid.UserNetFlow = decimal.NewFromInt(999) // Mismatch
	assert.Error(t, lInvalid.Validate())
}

func TestTransactionType_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "transfer", TransactionTypeTransfer.String())
	assert.Equal(t, "unknown", TransactionType(255).String())
}

func TestNewTransaction(t *testing.T) {
	t.Parallel()
	tx := NewTransaction("id", "asset", TransactionTypeTransfer, "from", "to", nil, nil, decimal.Zero)
	assert.Equal(t, "id", tx.ID)
	assert.False(t, tx.CreatedAt.IsZero())
}

func TestNewTransactionFromTransition(t *testing.T) {
	t.Parallel()
	senderState := NewVoidState("A", "U")
	senderState.HomeChannelID = new(string)
	*senderState.HomeChannelID = "HC"
	senderState.EscrowChannelID = new(string)
	*senderState.EscrowChannelID = "EC"

	// HomeDeposit
	transition := Transition{Type: TransitionTypeHomeDeposit, Amount: decimal.NewFromInt(10)}
	tx, err := NewTransactionFromTransition(senderState, nil, transition)
	require.NoError(t, err)
	assert.Equal(t, TransactionTypeHomeDeposit, tx.TxType)

	// TransferSend
	transition = Transition{Type: TransitionTypeTransferSend, Amount: decimal.NewFromInt(10), AccountID: "REC"}
	receiverState := NewVoidState("A", "REC")
	tx, err = NewTransactionFromTransition(senderState, receiverState, transition)
	require.NoError(t, err)
	assert.Equal(t, TransactionTypeTransfer, tx.TxType)

	// TransferSend error: nil receiverState
	_, err = NewTransactionFromTransition(senderState, nil, transition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver state must not be nil")

	// Invalid
	transition = Transition{Type: 255}
	_, err = NewTransactionFromTransition(senderState, nil, transition)
	assert.Error(t, err)
}

func TestPaginationParams_GetOffsetAndLimit(t *testing.T) {
	t.Parallel()
	var p *PaginationParams
	o, l := p.GetOffsetAndLimit(10, 100)
	assert.Equal(t, uint32(0), o)
	assert.Equal(t, uint32(10), l)

	off := uint32(5)
	lim := uint32(50)
	p = &PaginationParams{Offset: &off, Limit: &lim}
	o, l = p.GetOffsetAndLimit(10, 100)
	assert.Equal(t, uint32(5), o)
	assert.Equal(t, uint32(50), l)

	lim = 200
	o, l = p.GetOffsetAndLimit(10, 100)
	assert.Equal(t, uint32(100), l) // Capped at max
}
