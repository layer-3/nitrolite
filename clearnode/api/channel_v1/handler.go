package channel_v1

import (
	"context"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// Handler manages channel state transitions and provides RPC endpoints for state submission.
type Handler struct {
	useStoreInTx     StoreTxProvider
	memoryStore      MemoryStore
	actionGateway    ActionGateway
	nodeSigner       *core.ChannelDefaultSigner
	stateAdvancer    core.StateAdvancer
	statePacker      core.StatePacker
	nodeAddress      string // Node's wallet address for channel ID calculation
	minChallenge     uint32
	metrics          metrics.RuntimeMetricExporter
	maxSessionKeyIDs int
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	memoryStore MemoryStore,
	actionGateway ActionGateway,
	nodeSigner *core.ChannelDefaultSigner,
	stateAdvancer core.StateAdvancer,
	statePacker core.StatePacker,
	nodeAddress string,
	minChallenge uint32,
	m metrics.RuntimeMetricExporter,
	maxSessionKeyIDs int,
) *Handler {
	return &Handler{
		stateAdvancer:    stateAdvancer,
		statePacker:      statePacker,
		useStoreInTx:     useStoreInTx,
		memoryStore:      memoryStore,
		actionGateway:    actionGateway,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          m,
		maxSessionKeyIDs: maxSessionKeyIDs,
	}
}

func (h *Handler) getChannelSigValidator(tx Store, asset string) *core.ChannelSigValidator {
	return core.NewChannelSigValidator(func(walletAddr, sessionKeyAddr, metadataHash string) (bool, error) {
		return tx.ValidateChannelSessionKeyForAsset(walletAddr, sessionKeyAddr, asset, metadataHash)
	})
}

// issueTransferReceiverState creates and stores a new state for the receiver of a transfer.
// It reads the receiver's current state, applies a transfer_receive transition with the same
// amount and tx hash, signs it with the node's key, and persists it.
func (h *Handler) issueTransferReceiverState(ctx context.Context, tx Store, senderState core.State) (*core.State, error) {
	logger := log.FromContext(ctx)

	incomingTransition := senderState.Transition
	if incomingTransition.Type != core.TransitionTypeTransferSend {
		return nil, rpc.Errorf("incoming state doesn't have 'transfer_send' transition")
	}
	receiverWallet := incomingTransition.AccountID
	if senderState.UserWallet == receiverWallet {
		return nil, rpc.Errorf("sender and receiver wallets are the same")
	}

	logger = logger.
		WithKV("sender", senderState.UserWallet).
		WithKV("receiver", receiverWallet).
		WithKV("asset", senderState.Asset)

	logger.Debug("issuing transfer receiver state")

	// Lock the receiver's state to prevent concurrent modifications
	if _, err := tx.LockUserState(receiverWallet, senderState.Asset); err != nil {
		return nil, rpc.Errorf("failed to lock receiver state: %v", err)
	}

	currentState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, false)
	if err != nil {
		return nil, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}
	if currentState == nil {
		currentState = core.NewVoidState(senderState.Asset, receiverWallet)
	}
	newState := currentState.NextState()

	_, err = newState.ApplyTransferReceiveTransition(
		senderState.UserWallet,
		incomingTransition.Amount,
		incomingTransition.TxID)
	if err != nil {
		return nil, err
	}

	lastSignedState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, true)
	if err != nil {
		return nil, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}

	// TODO: move to DB query
	shouldSign := true
	if lastSignedState != nil {
		lastStateTransition := lastSignedState.Transition
		if lastStateTransition.Type == core.TransitionTypeMutualLock ||
			lastStateTransition.Type == core.TransitionTypeEscrowLock {
			shouldSign = false
		}
	}

	if newState.HomeChannelID != nil && shouldSign {
		packedState, err := h.statePacker.PackState(*newState)
		if err != nil {
			return nil, rpc.Errorf("failed to pack receiver state: %v", err)
		}

		_nodeSig, err := h.nodeSigner.Sign(packedState)
		if err != nil {
			return nil, rpc.Errorf("failed to sign receiver state")
		}
		nodeSig := _nodeSig.String()
		newState.NodeSig = &nodeSig
	}
	if err := tx.StoreUserState(*newState); err != nil {
		return nil, rpc.Errorf("failed to store receiver state")
	}

	logger.Info("issued transfer receiver state", "receiverStateVersion", newState.Version)
	return newState, nil
}

// issueExtraState creates an additional state by reapplying unsigned transitions to a newly signed state.
// When a user submits a signed state (e.g., after escrow_deposit or escrow_withdraw), any pending
// unsigned transitions from the previous state are reapplied to create a new unsigned state.
// This ensures that pending operations are preserved across state updates that require user signatures.
// func (h *Handler) issueExtraState(ctx context.Context, tx Store, incomingState core.State, extraTransitions []core.Transition) (*core.State, error) {
// 	logger := log.FromContext(ctx)

// 	lastTransition := incomingState.GetLastTransition()
// 	if lastTransition == nil {
// 		return nil, rpc.Errorf("incoming state has no transitions")
// 	}

// 	if len(extraTransitions) == 0 {
// 		return &incomingState, nil
// 	}

// 	logger = logger.
// 		WithKV("userWallet", incomingState.UserWallet).
// 		WithKV("asset", incomingState.Asset)

// 	extraState := incomingState.NextState()
// 	logger.Debug("issuing extra state", "extraStateVersion", extraState.Version)

// 	if err := extraState.ApplyReceiverTransitions(extraTransitions...); err != nil {
// 		return nil, err
// 	}

// 	packedState, err := h.statePacker.PackState(*extraState)
// 	if err != nil {
// 		return nil, rpc.Errorf("failed to pack extra state")
// 	}

// 	_nodeSig, err := h.signer.Sign(packedState)
// 	if err != nil {
// 		return nil, rpc.Errorf("failed to sign extra state")
// 	}
// 	nodeSig := _nodeSig.String()
// 	extraState.NodeSig = &nodeSig

// 	if err := tx.StoreUserState(*extraState); err != nil {
// 		return nil, rpc.Errorf("failed to store extra state")
// 	}

// 	logger.Info("issued extra state", "extraStateVersion", extraState.Version)
// 	return extraState, nil
// }
