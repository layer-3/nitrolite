package channel_v1

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// RequestCreation processes channel creation requests from users.
// It validates the channel definition and initial state, checks for existing channels,
// signs the state with the node's key, and persists the new pending state and channel.
func (h *Handler) RequestCreation(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.ChannelsV1RequestCreationRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var err error
	reqPayload.State.UserWallet, err = core.NormalizeHexAddress(reqPayload.State.UserWallet)
	if err != nil {
		c.Fail(rpc.Errorf("invalid user_wallet: %v", err), "")
		return
	}

	incomingState, err := toCoreState(reqPayload.State)
	if err != nil {
		c.Fail(err, "failed to parse state")
		return
	}

	channelDef, err := toCoreChannelDefinition(reqPayload.ChannelDefinition)
	if err != nil {
		c.Fail(err, "failed to parse channel definition")
		return
	}

	if channelDef.Nonce == 0 {
		c.Fail(nil, "nonce must be non-zero")
		return
	}
	if channelDef.Challenge < h.minChallenge {
		c.Fail(rpc.Errorf("challenge period must be at least %d seconds, but got %d", h.minChallenge, channelDef.Challenge), "")
		return
	}
	if channelDef.Challenge > h.maxChallenge {
		c.Fail(rpc.Errorf("challenge period must be at most %d seconds, but got %d", h.maxChallenge, channelDef.Challenge), "")
		return
	}

	logger = logger.
		WithKV("userWallet", incomingState.UserWallet).
		WithKV("asset", incomingState.Asset)

	ok, err := h.memoryStore.IsAssetSupported(incomingState.Asset, incomingState.HomeLedger.TokenAddress, incomingState.HomeLedger.BlockchainID)
	if err != nil {
		c.Fail(err, "failed to check asset support")
		return
	}
	if !ok {
		c.Fail(rpc.Errorf(
			"asset %s is not supported on blockchain %d with token address %s",
			incomingState.Asset,
			incomingState.HomeLedger.BlockchainID,
			incomingState.HomeLedger.TokenAddress), "")
		return
	}

	ok = core.SignerValidatorsSupported(reqPayload.ChannelDefinition.ApprovedSigValidators)
	if !ok {
		c.Fail(nil, "one or more of the provided approved signature validators are not supported")
		return
	}

	var nodeSig string
	err = h.useStoreInTx(func(tx Store) error {
		_, err := tx.LockUserState(incomingState.UserWallet, incomingState.Asset)
		if err != nil {
			return rpc.Errorf("failed to lock user state: %v", err)
		}

		// Check if channel already exists
		currentState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, false)
		if err != nil {
			return rpc.Errorf("failed to check existing channel: %v", err)
		}
		// User has no previous state
		if currentState == nil {
			logger.Debug("no previous state found, issuing a void state")
			currentState = core.NewVoidState(incomingState.Asset, incomingState.UserWallet)
		}

		// Calculate home channel ID
		homeChannelID, err := core.GetHomeChannelID(
			h.nodeAddress,
			incomingState.UserWallet,
			incomingState.Asset,
			channelDef.Nonce,
			channelDef.Challenge,
			channelDef.ApprovedSigValidators,
		)
		if err != nil {
			return rpc.Errorf("failed to calculate channel ID: %v", err)
		}

		// Validate the home channel ID in the state
		if incomingState.HomeChannelID == nil || !strings.EqualFold(*incomingState.HomeChannelID, homeChannelID) {
			return rpc.Errorf("incoming state home_channel_id is invalid")
		}

		if currentState.HomeChannelID != nil {
			isFinal := currentState.IsFinal()
			if !isFinal {
				return rpc.Errorf("channel is already initialized")
			}
			if isFinal && strings.EqualFold(*incomingState.HomeChannelID, *currentState.HomeChannelID) {
				return rpc.Errorf("cannot use same home channel id")
			}
		}

		logger.Debug("processing channel creation request", "incomingVersion", incomingState.Version)

		if err := h.stateAdvancer.ValidateAdvancement(*currentState, incomingState); err != nil {
			return rpc.Errorf("invalid state: %v", err)
		}

		// Pack and validate user signature
		packedState, err := h.statePacker.PackState(incomingState)
		if err != nil {
			return rpc.Errorf("failed to pack state: %v", err)
		}

		if incomingState.UserSig == nil {
			return rpc.Errorf("missing user signature")
		}
		userSigBytes, err := hexutil.Decode(*incomingState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode user signature: %v", err)
		}

		sigType, err := core.GetSignerType(userSigBytes)
		if err != nil {
			return rpc.Errorf("failed to get user signature type: %v", err)
		}

		if !core.IsChannelSignerSupported(channelDef.ApprovedSigValidators, sigType) {
			return rpc.Errorf("user signature type '%d' is not supported by channel", sigType)
		}

		sigValidator := h.getChannelSigValidator(tx, incomingState.Asset)
		if err := sigValidator.Verify(incomingState.UserWallet, packedState, userSigBytes); err != nil {
			h.metrics.IncChannelStateSigValidation(sigType, false)
			return rpc.Errorf("invalid incoming state user signature: %v", err)
		}
		h.metrics.IncChannelStateSigValidation(sigType, true)

		newHomeChannel := core.NewChannel(
			homeChannelID,
			incomingState.UserWallet,
			incomingState.Asset,
			core.ChannelTypeHome,
			incomingState.HomeLedger.BlockchainID,
			incomingState.HomeLedger.TokenAddress,
			channelDef.Nonce,
			channelDef.Challenge,
			channelDef.ApprovedSigValidators,
		)

		// Create the home channel entity
		if err := tx.CreateChannel(*newHomeChannel); err != nil {
			return rpc.Errorf("failed to create channel: %v", err)
		}

		// Provide node's signature
		_nodeSig, err := h.nodeSigner.Sign(packedState)
		if err != nil {
			return rpc.Errorf("failed to sign state: %v", err)
		}
		nodeSig = _nodeSig.String()
		incomingState.NodeSig = &nodeSig

		// Store the pending state
		if err := tx.StoreUserState(incomingState); err != nil {
			return rpc.Errorf("failed to store state: %v", err)
		}

		incomingTransition := incomingState.Transition

		if incomingTransition.Type != core.TransitionTypeAcknowledgement {
			var transaction *core.Transaction

			switch incomingTransition.Type {
			case core.TransitionTypeVoid:
				return rpc.Errorf("incoming state has no transitions")

			case core.TransitionTypeHomeDeposit, core.TransitionTypeHomeWithdrawal:
				transaction, err = core.NewTransactionFromTransition(&incomingState, nil, incomingTransition)
				if err != nil {
					return rpc.Errorf("failed to create transaction: %v", err)
				}

				// We return Node's signature, the user is expected to submit this on blockchain.
			case core.TransitionTypeTransferSend:
				newReceiverState, err := h.issueTransferReceiverState(ctx, tx, incomingState)
				if err != nil {
					return rpc.Errorf("failed to issue receiver state: %v", err)
				}
				transaction, err = core.NewTransactionFromTransition(&incomingState, newReceiverState, incomingTransition)
				if err != nil {
					return rpc.Errorf("failed to create transaction: %v", err)
				}
			default:
				return rpc.Errorf("transition '%s' is not supported by this endpoint", incomingTransition.Type.String())
			}

			if err := tx.RecordTransaction(*transaction); err != nil {
				return rpc.Errorf("failed to record transaction")
			}

			logger.Info("recorded transaction",
				"txID", transaction.ID,
				"txType", transaction.TxType.String(),
				"from", transaction.FromAccount,
				"to", transaction.ToAccount,
				"asset", transaction.Asset,
				"amount", transaction.Amount.String())
		}

		logger.Info("channel creation request processed",
			"homeChannelID", homeChannelID,
			"nonce", channelDef.Nonce,
			"challengeDuration", time.Duration(channelDef.Challenge)*time.Second,
			"incomingVersion", incomingState.Version)
		return nil
	})
	if err != nil {
		logger.Error("failed to process channel creation request", "error", err)
		c.Fail(err, "failed to process channel creation request")
		return
	}

	resp := rpc.ChannelsV1RequestCreationResponse{
		Signature: nodeSig,
	}
	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Debug("channel creation request completed", "userWallet", incomingState.UserWallet, "asset", incomingState.Asset)
}
