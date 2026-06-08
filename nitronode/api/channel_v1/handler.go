package channel_v1

import (
	"context"
	"strings"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// Handler manages channel state transitions and provides RPC endpoints for state submission.
type Handler struct {
	useStoreInTx          StoreTxProvider
	memoryStore           MemoryStore
	nodeSigner            *core.ChannelDefaultSigner
	stateAdvancer         core.StateAdvancer
	statePacker           core.StatePacker
	nodeAddress           string // Node's wallet address for channel ID calculation
	minChallenge          uint32
	maxChallenge          uint32
	metrics               metrics.RuntimeMetricExporter
	maxSessionKeyIDs      int
	maxSessionKeysPerUser int
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	memoryStore MemoryStore,
	nodeSigner *core.ChannelDefaultSigner,
	stateAdvancer core.StateAdvancer,
	statePacker core.StatePacker,
	nodeAddress string,
	minChallenge, maxChallenge uint32,
	m metrics.RuntimeMetricExporter,
	maxSessionKeyIDs int,
	maxSessionKeysPerUser int,
) *Handler {
	return &Handler{
		stateAdvancer:         stateAdvancer,
		statePacker:           statePacker,
		useStoreInTx:          useStoreInTx,
		memoryStore:           memoryStore,
		nodeSigner:            nodeSigner,
		nodeAddress:           nodeAddress,
		minChallenge:          minChallenge,
		maxChallenge:          maxChallenge,
		metrics:               m,
		maxSessionKeyIDs:      maxSessionKeyIDs,
		maxSessionKeysPerUser: maxSessionKeysPerUser,
	}
}

func (h *Handler) getChannelSigValidator(tx Store, asset string) *core.ChannelSigValidator {
	return core.NewChannelSigValidator(func(walletAddr, sessionKeyAddr, metadataHash string) (bool, error) {
		return tx.ValidateChannelSessionKeyForAsset(walletAddr, sessionKeyAddr, asset, metadataHash)
	})
}

// lockTransferBalances normalizes the receiver address, rejects self-transfers,
// and locks the sender's and receiver's (wallet, asset) balance rows in a
// deterministic order (ascending lowercase wallet) to avoid database deadlocks
// when two opposite-direction transfers run concurrently. It returns the
// normalized receiver wallet.
//
// Callers MUST call this before issueTransferReceiverState for a TransferSend
// transition — issueTransferReceiverState assumes the receiver row is already
// locked and does not take the lock itself.
func (h *Handler) lockTransferBalances(tx Store, senderState core.State) (string, error) {
	receiverWallet, err := core.NormalizeHexAddress(senderState.Transition.AccountID)
	if err != nil {
		return "", rpc.Errorf("invalid receiver wallet address: %v", err)
	}
	if strings.EqualFold(senderState.UserWallet, receiverWallet) {
		return "", rpc.Errorf("sender and receiver wallets are the same")
	}

	// Lock both rows in ascending lowercase-wallet order so concurrent
	// opposite-direction transfers acquire the same locks in the same sequence.
	// LockUserState lowercases internally, so the sort key matches the lock key.
	first, second := senderState.UserWallet, receiverWallet
	if strings.ToLower(first) > strings.ToLower(second) {
		first, second = second, first
	}
	if _, err := tx.LockUserState(first, senderState.Asset); err != nil {
		return "", rpc.Errorf("failed to lock user state: %v", err)
	}
	if _, err := tx.LockUserState(second, senderState.Asset); err != nil {
		return "", rpc.Errorf("failed to lock user state: %v", err)
	}
	return receiverWallet, nil
}

// issueTransferReceiverState creates and stores a new state for the receiver of a transfer.
// It reads the receiver's current state, applies a transfer_receive transition with the same
// amount and tx hash, signs it with the node's key, and persists it.
//
// The receiver's (wallet, asset) balance row MUST already be locked by the caller
// (via lockTransferBalances) before this is invoked — it does not take the lock itself.
func (h *Handler) issueTransferReceiverState(ctx context.Context, tx Store, senderState core.State, applicationID string) (*core.State, error) {
	logger := log.FromContext(ctx)

	incomingTransition := senderState.Transition
	if incomingTransition.Type != core.TransitionTypeTransferSend {
		return nil, rpc.Errorf("incoming state doesn't have 'transfer_send' transition")
	}
	receiverWallet, err := core.NormalizeHexAddress(incomingTransition.AccountID)
	if err != nil {
		return nil, rpc.Errorf("invalid receiver wallet address: %v", err)
	}

	if strings.EqualFold(senderState.UserWallet, receiverWallet) {
		return nil, rpc.Errorf("sender and receiver wallets are the same")
	}

	logger = logger.
		WithKV("sender", senderState.UserWallet).
		WithKV("receiver", receiverWallet).
		WithKV("asset", senderState.Asset)

	logger.Debug("issuing transfer receiver state")

	currentState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, false)
	if err != nil {
		return nil, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, receiverWallet)
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

	if err := tx.EnsureNoOngoingEscrowOperation(receiverWallet, senderState.Asset); err != nil {
		return nil, rpc.Errorf("cannot issue transfer receiver state: %v", err)
	}

	if newState.HomeChannelID != nil {
		// CheckActiveChannel returns a non-nil status only when an Open or Void home
		// channel exists for (wallet, asset); Challenged / Closing / Closed channels
		// fall through as nil. The node only attaches a signature on Open: any other
		// status means the channel is mid-dispute or terminal, and node-signing a
		// receiver state on it could turn dust credits into a fresh checkpoint
		// candidate (Challenged) or commit a credit that will never settle (Closed,
		// Closing). The unsigned row is still persisted so the challenge-rescue
		// squash at close can pick it up.
		//
		// MF3-I01: persisting an unsigned row whose HomeChannelID still references
		// a now-Closed channel is safe under the listener ordering & idempotency
		// invariant (pkg/blockchain/evm/listener.go, processEvents doc). For any
		// Path-1 (challenge-timeout) close, HandleHomeChannelChallenged has
		// already run before HandleHomeChannelClosed, so the rescue issued from
		// HandleHomeChannelClosed has overwritten currentState with a fresh-epoch
		// row whose HomeChannelID is nil, and the next call here reads that row
		// rather than the wedge state. The unsigned credit issued here can only
		// land on a Closed channel if it commits before the close handler runs
		// — in which case the close handler's rescue sum picks it up.
		_, channelStatus, err := tx.CheckActiveChannel(receiverWallet, senderState.Asset)
		if err != nil {
			return nil, rpc.Errorf("failed to check receiver active channel: %v", err)
		}
		if channelStatus != nil && *channelStatus == core.ChannelStatusOpen {
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
		} else {
			logger.Info("skipping node signature on receiver state for non-open home channel",
				"homeChannelID", *newState.HomeChannelID)
		}
	}
	if err := tx.StoreUserState(*newState, applicationID); err != nil {
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
