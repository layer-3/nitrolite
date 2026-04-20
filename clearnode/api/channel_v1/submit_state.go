package channel_v1

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// SubmitState processes user-submitted state transitions, validates them against the current state,
// verifies user signatures, signs the new state with the node's key, and persists changes.
// For transfer transitions, it automatically creates corresponding receiver states.
// For certain transitions (escrow lock, etc.), it schedules blockchain actions.
func (h *Handler) SubmitState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.ChannelsV1SubmitStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	incomingState, err := toCoreState(reqPayload.State)
	if err != nil {
		c.Fail(err, "failed to parse state")
		return
	}

	applicationID := rpc.GetApplicationID(c)

	var nodeSig string
	incomingTransition := incomingState.Transition
	err = h.useStoreInTx(func(tx Store) error {
		err := h.actionGateway.AllowAction(tx, incomingState.UserWallet, incomingState.Transition.Type.GatedAction())
		if err != nil {
			return rpc.NewError(err)
		}

		_, err = tx.LockUserState(incomingState.UserWallet, incomingState.Asset)
		if err != nil {
			return rpc.Errorf("failed to lock user state: %v", err)
		}

		approvedSigValidators, userHasOpenChannel, err := tx.CheckOpenChannel(incomingState.UserWallet, incomingState.Asset)
		if err != nil {
			return rpc.Errorf("failed to check open channel: %v", err)
		}
		if !userHasOpenChannel {
			return rpc.Errorf("user has no open channel")
		}

		logger.Debug("processing incoming state",
			"userWallet", incomingState.UserWallet,
			"asset", incomingState.Asset,
			"incomingTransition", incomingTransition.Type.String())

		currentState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, false)
		if err != nil {
			return rpc.Errorf("failed to get last user state: %v", err)
		}

		// FIXME:
		// var extraTransitions []core.Transition
		switch incomingTransition.Type {
		case core.TransitionTypeEscrowDeposit, core.TransitionTypeEscrowWithdraw, core.TransitionTypeMigrate:
			return rpc.Errorf("transition is not supported yet")
			// latestStateVersion := currentState.Version
			// extraTransitions = currentState.Transitions

			// currentState, err = tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, true)
			// if err != nil {
			// 	return rpc.Errorf("failed to get last user state: %v", err)
			// }

			// // User has no signed previous state
			// if currentState == nil {
			// 	return rpc.Errorf("no signed previous state found for escrow/migrate transition")
			// }
			// if currentState.Version < latestStateVersion {
			// 	currentState.Version = latestStateVersion
			// } else if currentState.Version == latestStateVersion {
			// 	extraTransitions = nil // no extra transitions to reapply
			// }
		default:
			// User has no previous state
			if currentState == nil {
				logger.Debug("no previous state found, issuing a void state")
				currentState = core.NewVoidState(incomingState.Asset, incomingState.UserWallet)
			}
		}

		if err := tx.EnsureNoOngoingStateTransitions(currentState.UserWallet, currentState.Asset); err != nil {
			return rpc.Errorf("ongoing state transitions check failed: %v", err)
		}

		if err := h.stateAdvancer.ValidateAdvancement(*currentState, incomingState); err != nil {
			return rpc.Errorf("invalid state transition: %w", err)
		}

		packedState, err := h.statePacker.PackState(incomingState)
		if err != nil {
			return rpc.Errorf("failed to pack state: %v", err)
		}

		// Validate user's signature
		if incomingState.UserSig == nil {
			return rpc.Errorf("missing incoming state user signature: %v", err)
		}
		userSigBytes, err := hexutil.Decode(*incomingState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode incoming state user signature: %v", err)
		}

		sigType, err := core.GetSignerType(userSigBytes)
		if err != nil {
			return rpc.Errorf("failed to get user signature type: %v", err)
		}
		if !core.IsChannelSignerSupported(approvedSigValidators, sigType) {
			return rpc.Errorf("user signature type '%d' is not supported by channel", sigType)
		}
		sigValidator := h.getChannelSigValidator(tx, incomingState.Asset)
		if err := sigValidator.Verify(incomingState.UserWallet, packedState, userSigBytes); err != nil {
			h.metrics.IncChannelStateSigValidation(sigType, false)
			return rpc.Errorf("invalid incoming state user signature: %v", err)
		}
		h.metrics.IncChannelStateSigValidation(sigType, true)

		// Provide node's signature
		_nodeSig, err := h.nodeSigner.Sign(packedState)
		if err != nil {
			return rpc.Errorf("failed to sign incoming state: %v", err)
		}
		nodeSig = _nodeSig.String()
		incomingState.NodeSig = &nodeSig

		// Store user state early — it's fully validated and signed at this point.
		// The wrapping DB transaction ensures rollback if any subsequent step fails.
		if err := tx.StoreUserState(incomingState, applicationID); err != nil {
			return rpc.Errorf("failed to store user state: %v", err)
		}

		if incomingTransition.Type != core.TransitionTypeAcknowledgement {
			var transaction *core.Transaction
			switch incomingTransition.Type {
			case core.TransitionTypeHomeDeposit, core.TransitionTypeHomeWithdrawal:
				// We return Node's signature, the user is expected to submit this on blockchain.
				transaction, err = core.NewTransactionFromTransition(&incomingState, nil, incomingTransition)
				if err != nil {
					return rpc.Errorf("failed to create transaction: %v", err)
				}

			case core.TransitionTypeTransferSend:
				newReceiverState, err := h.issueTransferReceiverState(ctx, tx, incomingState, applicationID)
				if err != nil {
					return rpc.Errorf("failed to issue receiver state: %v", err)
				}
				transaction, err = core.NewTransactionFromTransition(&incomingState, newReceiverState, incomingTransition)
				if err != nil {
					return rpc.Errorf("failed to create transaction: %v", err)
				}
			case core.TransitionTypeMutualLock:
				return rpc.Errorf("transition is not supported yet")
				// if err := h.createEscrowChannel(tx, incomingState); err != nil {
				// 	return err
				// }

				// transaction, err = core.NewTransactionFromTransition(&incomingState, nil, *incomingTransition)
				// if err != nil {
				// 	return rpc.Errorf("failed to create transaction: %v", err)
				// }
			case core.TransitionTypeEscrowLock:
				return rpc.Errorf("transition is not supported yet")
				// if err := h.createEscrowChannel(tx, incomingState); err != nil {
				// 	return err
				// }

				// if err := tx.ScheduleInitiateEscrowWithdrawal(incomingState.ID, incomingState.EscrowLedger.BlockchainID); err != nil {
				// 	return rpc.Errorf("failed to schedule blockchain action: %v", err)
				// }
				// transaction, err = core.NewTransactionFromTransition(&incomingState, nil, *incomingTransition)
				// if err != nil {
				// 	return rpc.Errorf("failed to create transaction: %v", err)
				// }
			case core.TransitionTypeEscrowDeposit:
				return rpc.Errorf("transition is not supported yet")
				// transaction, err = core.NewTransactionFromTransition(&incomingState, nil, *incomingTransition)
				// if err != nil {
				// 	return rpc.Errorf("failed to create transaction: %v", err)
				// }
				// extraState, err := h.issueExtraState(ctx, tx, incomingState, extraTransitions)
				// if err != nil {
				// 	return rpc.Errorf("failed to issue an extra state: %v", err)
				// }
				// logger.Info("extra state issued", "userID", extraState.UserWallet, "asset", extraState.Asset, "version", extraState.Version)
			case core.TransitionTypeEscrowWithdraw:
				return rpc.Errorf("transition is not supported yet")
				// transaction, err = core.NewTransactionFromTransition(&incomingState, nil, *incomingTransition)
				// if err != nil {
				// 	return rpc.Errorf("failed to create transaction: %v", err)
				// }

				// extraState, err := h.issueExtraState(ctx, tx, incomingState, extraTransitions)
				// if err != nil {
				// 	return rpc.Errorf("failed to issue an extra state: %v", err)
				// }
				// logger.Info("extra state issued", "userID", extraState.UserWallet, "asset", extraState.Asset, "version", extraState.Version)
			case core.TransitionTypeFinalize:
				transaction, err = core.NewTransactionFromTransition(&incomingState, nil, incomingTransition)
				if err != nil {
					return rpc.Errorf("failed to create transaction: %v", err)
				}
			case core.TransitionTypeMigrate:
				return rpc.Errorf("transition is not supported yet")
				// extraState, err := h.issueExtraState(ctx, tx, incomingState)
				// if err != nil {
				// 	return rpc.Errorf("failed to issue extra state: %v", err)
				// }
			default:
				return rpc.Errorf("transition '%s' is not supported by this endpoint", incomingTransition.Type.String())
			}

			if err := tx.RecordTransaction(*transaction, applicationID); err != nil {
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

		// TODO: consider state checkpoint if channel is challenged

		return nil
	})
	if err != nil {
		logger.Error("failed to process incoming state", "error", err)
		c.Fail(err, "failed to process incoming state")
		return
	}

	resp := rpc.ChannelsV1SubmitStateResponse{
		Signature: nodeSig,
	}
	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("processed incoming state",
		"userWallet", incomingState.UserWallet,
		"asset", incomingState.Asset,
		"incomingVersion", incomingState.Version,
		"incomingTransition", incomingTransition.Type.String())
}

func (h *Handler) createEscrowChannel(tx Store, incomingState core.State) error {
	if incomingState.EscrowChannelID == nil {
		return rpc.Errorf("missing escrow channel ID")
	}
	escrowChannelID, err := core.GetEscrowChannelID(*incomingState.HomeChannelID, incomingState.Version) // just to validate format
	if err != nil {
		return rpc.Errorf("failed to calculate escrow channel ID: %v", err)
	}
	if *incomingState.EscrowChannelID != escrowChannelID {
		return rpc.Errorf("incoming state escrow_channel_id is invalid")
	}
	homeChannel, err := tx.GetChannelByID(*incomingState.HomeChannelID)
	if err != nil {
		return rpc.Errorf("failed to get home channel: %v", err)
	}
	if homeChannel == nil {
		return rpc.Errorf("home channel does not exist")
	}

	ok, err := h.memoryStore.IsAssetSupported(incomingState.Asset, incomingState.EscrowLedger.TokenAddress, incomingState.EscrowLedger.BlockchainID)
	if err != nil {
		return err
	}
	if !ok {
		return rpc.Errorf(
			"asset %s is not supported on blockchain %d with token address %s",
			incomingState.Asset,
			incomingState.EscrowLedger.BlockchainID,
			incomingState.EscrowLedger.TokenAddress)
	}

	newEscrowChannel := core.NewChannel(
		escrowChannelID,
		incomingState.UserWallet,
		incomingState.Asset,
		core.ChannelTypeEscrow,
		incomingState.EscrowLedger.BlockchainID,
		incomingState.EscrowLedger.TokenAddress,
		homeChannel.Nonce,
		homeChannel.ChallengeDuration,
		homeChannel.ApprovedSigValidators,
	)

	// Create the escrow channel entity
	err = tx.CreateChannel(*newEscrowChannel)
	if err != nil {
		return rpc.Errorf("failed to create escrow channel: %v", err)
	}
	return nil
}
