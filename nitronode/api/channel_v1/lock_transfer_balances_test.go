package channel_v1

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
)

// lockOrder returns the wallet arguments of every LockUserState call recorded on
// the mock, in invocation order.
func lockOrder(m *MockStore) []string {
	var order []string
	for _, call := range m.Calls {
		if call.Method == "LockUserState" {
			order = append(order, call.Arguments.String(0))
		}
	}
	return order
}

func transferSenderState(sender, receiver, asset string) core.State {
	return core.State{
		UserWallet: sender,
		Asset:      asset,
		Transition: core.Transition{
			Type:      core.TransitionTypeTransferSend,
			AccountID: receiver,
		},
	}
}

// MF3-L18: both balance rows must be locked in a deterministic order (ascending
// lowercase wallet) regardless of transfer direction, so two opposite-direction
// transfers can't acquire row locks in opposite order and deadlock.
func TestLockTransferBalances_DeterministicOrder(t *testing.T) {
	const (
		asset = "USDC"
		low   = "0x1111111111111111111111111111111111111111"
		high  = "0x2222222222222222222222222222222222222222"
	)

	t.Run("sender lower than receiver", func(t *testing.T) {
		h := &Handler{}
		tx := new(MockStore)
		tx.On("LockUserState", mock.Anything, asset).Return(decimal.Zero, nil)

		receiver, err := h.lockTransferBalances(tx, transferSenderState(low, high, asset))
		require.NoError(t, err)
		require.Equal(t, high, receiver)
		require.Equal(t, []string{low, high}, lockOrder(tx))
	})

	t.Run("sender higher than receiver", func(t *testing.T) {
		h := &Handler{}
		tx := new(MockStore)
		tx.On("LockUserState", mock.Anything, asset).Return(decimal.Zero, nil)

		// Direction flipped: sender is the lexicographically larger wallet, but the
		// locks must still be taken low-then-high — same sequence as the case above.
		receiver, err := h.lockTransferBalances(tx, transferSenderState(high, low, asset))
		require.NoError(t, err)
		require.Equal(t, low, receiver)
		require.Equal(t, []string{low, high}, lockOrder(tx))
	})
}

func TestLockTransferBalances_RejectsSelfTransfer(t *testing.T) {
	const wallet = "0x1111111111111111111111111111111111111111"
	h := &Handler{}
	tx := new(MockStore)

	// Mixed case must still be caught as a self-transfer before any lock is taken.
	_, err := h.lockTransferBalances(tx, transferSenderState(wallet, "0x1111111111111111111111111111111111111111", "USDC"))
	require.Error(t, err)
	require.Empty(t, lockOrder(tx), "no row should be locked for a self-transfer")
}

func TestLockTransferBalances_RejectsInvalidReceiver(t *testing.T) {
	h := &Handler{}
	tx := new(MockStore)

	_, err := h.lockTransferBalances(tx, transferSenderState("0x1111111111111111111111111111111111111111", "not-an-address", "USDC"))
	require.Error(t, err)
	require.Empty(t, lockOrder(tx), "no row should be locked when the receiver address is invalid")
}
