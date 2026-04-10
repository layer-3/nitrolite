package app_session_v1

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// Handler manages app session operations and provides RPC endpoints for app session management.
type Handler struct {
	useStoreInTx     StoreTxProvider
	assetStore       AssetStore
	actionGateway    ActionGateway
	signer           *core.ChannelDefaultSigner
	stateAdvancer    core.StateAdvancer
	statePacker      core.StatePacker
	nodeAddress      string // Node's wallet address
	metrics          metrics.RuntimeMetricExporter
	maxParticipants  int
	maxSessionData   int
	maxSessionKeyIDs int
	maxSignedUpdates int
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	assetStore AssetStore,
	actionGateway ActionGateway,
	signer *core.ChannelDefaultSigner,
	stateAdvancer core.StateAdvancer,
	statePacker core.StatePacker,
	nodeAddress string,
	m metrics.RuntimeMetricExporter,
	maxParticipants, maxSessionData, maxSessionKeyIDs, maxSignedUpdates int,
) *Handler {
	return &Handler{
		useStoreInTx:     useStoreInTx,
		assetStore:       assetStore,
		actionGateway:    actionGateway,
		signer:           signer,
		stateAdvancer:    stateAdvancer,
		statePacker:      statePacker,
		nodeAddress:      nodeAddress,
		metrics:          m,
		maxParticipants:  maxParticipants,
		maxSessionData:   maxSessionData,
		maxSessionKeyIDs: maxSessionKeyIDs,
		maxSignedUpdates: maxSignedUpdates,
	}
}

func (h *Handler) verifyQuorum(tx Store, appSessionId, applicationID string, participantWeights map[string]uint8, requiredQuorum uint8, data []byte, signatures []string) error {
	// Verify signatures and calculate quorum
	signedWeights := make(map[string]bool)
	var achievedQuorum uint8

	appSessionSignerValidator := app.NewAppSessionKeySigValidatorV1(
		func(sessionKeyAddr string) (string, error) {
			return tx.GetAppSessionKeyOwner(sessionKeyAddr, appSessionId)
		},
	)

	for _, sigHex := range signatures {
		sigBytes, err := hexutil.Decode(sigHex)
		if err != nil {
			return rpc.Errorf("failed to decode signature: %v", err)
		}

		sigType := app.AppSessionSignerTypeV1(sigBytes[0])
		userWallet, err := appSessionSignerValidator.Recover(data, sigBytes)
		if err != nil {
			h.metrics.IncAppSessionUpdateSigValidation(applicationID, sigType, false)
			return rpc.Errorf("failed to recover user wallet: %v", err)
		}
		h.metrics.IncAppSessionUpdateSigValidation(applicationID, sigType, true)
		userWallet = strings.ToLower(userWallet)

		// Check if signer is a participant
		weight, isParticipant := participantWeights[userWallet]
		if !isParticipant {
			return rpc.Errorf("signature from non-participant: %s", userWallet)
		}

		// Add weight if not already counted
		if !signedWeights[userWallet] {
			signedWeights[userWallet] = true
			achievedQuorum += weight
		}
	}

	// Check if quorum is met
	if achievedQuorum < requiredQuorum {
		return rpc.Errorf("quorum not met: achieved %d, required %d", achievedQuorum, requiredQuorum)
	}

	return nil
}

// issueReleaseReceiverState creates a new channel state for a participant receiving funds from app session.
// This follows the same pattern as issueTransferReceiverState in channel_v1 for transfer_receive transitions.
func (h *Handler) issueReleaseReceiverState(ctx context.Context, tx Store, receiverWallet, asset, appSessionID string, amount decimal.Decimal) error {
	logger := log.FromContext(ctx)

	// Lock the receiver's state to prevent concurrent modifications
	if _, err := tx.LockUserState(receiverWallet, asset); err != nil {
		return rpc.Errorf("failed to lock receiver state: %v", err)
	}

	// Get the receiver's current state (or create void state if none exists)
	currentState, err := tx.GetLastUserState(receiverWallet, asset, false)
	if err != nil {
		return rpc.Errorf("failed to get receiver state: %v", err)
	}
	if currentState == nil {
		currentState = core.NewVoidState(asset, receiverWallet)
	}

	logger = logger.
		WithKV("userWallet", receiverWallet).
		WithKV("asset", asset)

	// Create next state and apply release transition
	newState := currentState.NextState()
	logger.Debug("issuing app session receiver state",
		"stateVersion", newState.Version,
		"appSessionID", appSessionID,
		"amount", amount.String())

	releaseTransition, err := newState.ApplyReleaseTransition(appSessionID, amount)
	if err != nil {
		return rpc.Errorf("failed to apply release transition: %v", err)
	}

	// Check if we need to sign the state (skip signing if last signed state was a lock)
	lastSignedState, err := tx.GetLastUserState(receiverWallet, asset, true)
	if err != nil {
		return rpc.Errorf("failed to get last signed state: %v", err)
	}

	// TODO: move to DB query
	if lastSignedState != nil && lastSignedState.EscrowChannelID != nil {
		return rpc.Errorf("cannot issue release receiver state: last signed state is a lock with escrow channel %s", *lastSignedState.EscrowChannelID)
	}

	if newState.HomeChannelID != nil {
		// Pack and sign the state
		packedState, err := h.statePacker.PackState(*newState)
		if err != nil {
			return rpc.Errorf("failed to pack receiver state: %v", err)
		}

		nodeSig, err := h.signer.Sign(packedState)
		if err != nil {
			return rpc.Errorf("failed to sign receiver state: %v", err)
		}

		nodeSigStr := nodeSig.String()
		newState.NodeSig = &nodeSigStr
	}

	// Store the new state
	if err := tx.StoreUserState(*newState); err != nil {
		return rpc.Errorf("failed to store receiver state: %v", err)
	}

	transaction, err := core.NewTransactionFromTransition(nil, newState, releaseTransition)
	if err != nil {
		return rpc.Errorf("failed to create transaction: %v", err)
	}

	if err := tx.RecordTransaction(*transaction); err != nil {
		return rpc.Errorf("failed to record transaction: %v", err)
	}
	logger.Info("recorded transaction",
		"txID", transaction.ID,
		"txType", transaction.TxType.String(),
		"from", transaction.FromAccount,
		"to", transaction.ToAccount,
		"asset", transaction.Asset,
		"amount", transaction.Amount.String())

	logger.Info("issued app session receiver state",
		"stateVersion", newState.Version,
		"appSessionID", appSessionID,
		"amount", amount.String())
	return nil
}
