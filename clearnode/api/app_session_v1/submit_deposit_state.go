package app_session_v1

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

// SubmitDepositState processes app session deposit state submissions.
func (h *Handler) SubmitDepositState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1SubmitDepositStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if len(reqPayload.AppStateUpdate.SessionData) > h.maxSessionData {
		c.Fail(rpc.Errorf("session_data exceeds maximum length of %d", h.maxSessionData), "")
		return
	}

	logger.Debug("processing app session deposit request",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"version", reqPayload.AppStateUpdate.Version)

	appStateUpd, err := unmapAppStateUpdateV1(&reqPayload.AppStateUpdate)
	if err != nil {
		c.Fail(err, "failed to parse app state update")
		return
	}
	userState, err := unmapStateV1(reqPayload.UserState)
	if err != nil {
		c.Fail(err, "failed to parse user state")
		return
	}

	var nodeSig string
	err = h.useStoreInTx(func(tx Store) error {
		appSession, err := tx.GetAppSession(appStateUpd.AppSessionID)
		if err != nil {
			return rpc.Errorf("app session not found: %v", err)
		}
		if appSession == nil {
			return rpc.Errorf("app session not found")
		}
		if len(reqPayload.QuorumSigs) > len(appSession.Participants) {
			return rpc.Errorf("quorum_sigs count (%d) exceeds participants count (%d)", len(reqPayload.QuorumSigs), len(appSession.Participants))
		}
		if appSession.Status == app.AppSessionStatusClosed {
			return rpc.Errorf("app session is already closed")
		}
		if appStateUpd.Version != appSession.Version+1 {
			return rpc.Errorf("invalid app session version: expected %d, got %d", appSession.Version+1, appStateUpd.Version)
		}

		if appStateUpd.Intent != app.AppStateUpdateIntentDeposit {
			return rpc.Errorf("invalid intent: expected 'deposit', got '%s'", appStateUpd.Intent)
		}

		participantWeights := getParticipantWeights(appSession.Participants)

		if len(reqPayload.QuorumSigs) == 0 {
			return rpc.Errorf("no signatures provided")
		}

		if h.appRegistryEnabled {
			registeredApp, err := tx.GetApp(appSession.ApplicationID)
			if err != nil {
				return rpc.Errorf("failed to look up application: %v", err)
			}
			if registeredApp == nil {
				return rpc.Errorf("application %s is not registered", appSession.ApplicationID)
			}

			err = h.actionGateway.AllowAction(tx, registeredApp.App.OwnerWallet, appStateUpd.Intent.GatedAction())
			if err != nil {
				return rpc.NewError(err)
			}
		}

		// Lock the user's state to prevent concurrent modifications
		_, err = tx.LockUserState(userState.UserWallet, userState.Asset)
		if err != nil {
			return rpc.Errorf("failed to lock user state: %v", err)
		}

		lastTransition := userState.Transition
		if lastTransition.Type != core.TransitionTypeCommit {
			return rpc.Errorf("user state transition must have 'commit' type, got '%s'", lastTransition.Type.String())
		}

		approvedSigValidators, userHasOpenChannel, err := tx.CheckOpenChannel(userState.UserWallet, userState.Asset)
		if err != nil {
			return rpc.Errorf("failed to check open channel: %v", err)
		}
		if !userHasOpenChannel {
			return rpc.Errorf("user has no open channel")
		}

		if lastTransition.AccountID != appStateUpd.AppSessionID {
			return rpc.Errorf("user state transition account ID '%s' does not match app session ID '%s'",
				lastTransition.AccountID, appStateUpd.AppSessionID)
		}

		// Validate user signature on user state
		if userState.UserSig == nil {
			return rpc.Errorf("missing user signature on user state")
		}

		currentState, err := tx.GetLastUserState(userState.UserWallet, userState.Asset, false)
		if err != nil {
			return rpc.Errorf("failed to get last user state: %v", err)
		}
		if currentState == nil {
			currentState = core.NewVoidState(userState.Asset, userState.UserWallet)
		} else {
			if err := tx.EnsureNoOngoingStateTransitions(userState.UserWallet, userState.Asset); err != nil {
				return rpc.Errorf("ongoing state transitions check failed: %v", err)
			}
		}

		if err := h.stateAdvancer.ValidateAdvancement(*currentState, userState); err != nil {
			return rpc.Errorf("invalid state transitions: %v", err)
		}

		packedUserState, err := h.statePacker.PackState(userState)
		if err != nil {
			return rpc.Errorf("failed to pack user state: %v", err)
		}

		userSigBytes, err := hexutil.Decode(*userState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode user signature: %v", err)
		}

		sigType, err := core.GetSignerType(userSigBytes)
		if err != nil {
			return rpc.Errorf("failed to get user signature type: %v", err)
		}
		if !core.IsChannelSignerSupported(approvedSigValidators, sigType) {
			return rpc.Errorf("user signature type '%d' is not supported by channel", sigType)
		}
		sigValidator := core.NewChannelSigValidator(func(walletAddr, sessionKeyAddr, metadataHash string) (bool, error) {
			return tx.ValidateChannelSessionKeyForAsset(walletAddr, sessionKeyAddr, userState.Asset, metadataHash)
		})
		err = sigValidator.Verify(userState.UserWallet, packedUserState, userSigBytes)
		if err != nil {
			h.metrics.IncChannelStateSigValidation(sigType, false)
			return rpc.Errorf("failed to validate signature: %v", err)
		}
		h.metrics.IncChannelStateSigValidation(sigType, true)

		// Pack the app state update for signature verification
		packedStateUpdate, err := app.PackAppStateUpdateV1(appStateUpd)
		if err != nil {
			return rpc.Errorf("failed to pack app state update: %v", err)
		}

		if err := h.verifyQuorum(tx, appStateUpd.AppSessionID, appSession.ApplicationID, participantWeights, appSession.Quorum, packedStateUpdate, reqPayload.QuorumSigs); err != nil {
			return err
		}

		currentAllocations, err := tx.GetParticipantAllocations(appSession.SessionID)
		if err != nil {
			return rpc.Errorf("failed to get current allocations: %v", err)
		}

		// Track total deposit amount to validate against transition amount
		totalDepositAmount := decimal.Zero

		incomingAllocations := make(map[string]map[string]decimal.Decimal)
		for _, alloc := range appStateUpd.Allocations {
			if alloc.Amount.IsNegative() {
				return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
			}

			decimals, err := h.assetStore.GetAssetDecimals(alloc.Asset)
			if err != nil {
				return rpc.Errorf("failed to get asset decimals: %v", err)
			}

			if err := core.ValidateDecimalPrecision(alloc.Amount, decimals); err != nil {
				return rpc.Errorf("invalid amount for allocation with asset %s and participant %s: %w", alloc.Asset, alloc.Participant, err)
			}

			// Reject duplicate (participant, asset) entries
			if incomingAllocations[alloc.Participant] != nil {
				if _, exists := incomingAllocations[alloc.Participant][alloc.Asset]; exists {
					return rpc.Errorf("duplicate allocation for participant %s, asset %s", alloc.Participant, alloc.Asset)
				}
			}

			participantAllocs := currentAllocations[alloc.Participant]
			if participantAllocs == nil {
				participantAllocs = make(map[string]decimal.Decimal, 0)
			}
			currentAmount := participantAllocs[alloc.Asset]

			if alloc.Amount.LessThan(currentAmount) {
				return rpc.Errorf("decreased allocation for %s for participant %s", alloc.Asset, alloc.Participant)
			}

			if alloc.Amount.GreaterThan(currentAmount) {
				// Validate participant
				if _, ok := participantWeights[alloc.Participant]; !ok {
					return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
				}

				// Validate that allocation asset matches user state asset
				if alloc.Asset != userState.Asset {
					return rpc.Errorf("app session deposit allocation for asset '%s' does not match user channel state asset '%s'", alloc.Asset, userState.Asset)
				}

				depositAmount := alloc.Amount.Sub(currentAmount)

				// Accumulate total deposit amount
				totalDepositAmount = totalDepositAmount.Add(depositAmount)

				if err := tx.RecordLedgerEntry(alloc.Participant, appSession.SessionID, alloc.Asset, depositAmount); err != nil {
					return rpc.Errorf("failed to record ledger entry: %v", err)
				}
			}

			// Store in incoming allocations map
			if incomingAllocations[alloc.Participant] == nil {
				incomingAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
			}
			incomingAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
		}

		// Verify all session balances are accounted for
		for participant, assets := range currentAllocations {
			for asset, currentAmount := range assets {
				if currentAmount.IsZero() {
					continue
				}
				if asset == userState.Asset {
					// Skip asset being deposited to avoid double-checking
					continue
				}

				// Check if this participant+asset is included in the incoming request
				incomingAmount, found := incomingAllocations[participant][asset]
				if !found {
					return rpc.Errorf("deposit intent missing allocation for participant %s, asset %s with current amount %s",
						participant, asset, currentAmount.String())
				}

				// Verify amounts match exactly
				if !incomingAmount.Equal(currentAmount) {
					return rpc.Errorf("deposit intent requires non-deposited asset allocations to match current state: participant %s, asset %s, current %s, provided %s",
						participant, asset, currentAmount.String(), incomingAmount.String())
				}
			}
		}

		// Validate that total deposit amount matches the transition amount
		if !totalDepositAmount.Equal(lastTransition.Amount) {
			return rpc.Errorf("total deposit amount %s does not match transition amount %s", totalDepositAmount.String(), lastTransition.Amount.String())
		}

		// Update app session version
		appSession.Version++
		// Overwrite session data if provided
		if reqPayload.AppStateUpdate.SessionData != "" {
			appSession.SessionData = reqPayload.AppStateUpdate.SessionData
		}
		appSession.UpdatedAt = time.Now()

		if err := tx.UpdateAppSession(*appSession); err != nil {
			return rpc.Errorf("failed to update app session: %v", err)
		}

		// Sign the user state with node's signature
		// TODO:create a function to handle state signing
		_nodeSig, err := h.signer.Sign(packedUserState)
		if err != nil {
			return rpc.Errorf("failed to sign user state: %v", err)
		}
		nodeSig = _nodeSig.String()
		userState.NodeSig = &nodeSig

		if err := tx.StoreUserState(userState, appSession.ApplicationID); err != nil {
			return rpc.Errorf("failed to store user state: %v", err)
		}

		transaction, err := core.NewTransactionFromTransition(&userState, nil, lastTransition)
		if err != nil {
			return rpc.Errorf("failed to create transaction: %v", err)
		}

		if err := tx.RecordTransaction(*transaction, appSession.ApplicationID); err != nil {
			return rpc.Errorf("failed to record transaction: %v", err)
		}
		logger.Info("recorded transaction",
			"txID", transaction.ID,
			"txType", transaction.TxType.String(),
			"from", transaction.FromAccount,
			"to", transaction.ToAccount,
			"asset", transaction.Asset,
			"amount", transaction.Amount.String())

		return nil
	})

	if err != nil {
		logger.Error("failed to process deposit state", "error", err)
		c.Fail(err, "failed to process deposit state")
		return
	}

	resp := rpc.AppSessionsV1SubmitDepositStateResponse{
		StateNodeSig: nodeSig,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully processed deposit state",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"appSessionVersion", appStateUpd.Version,
		"userWallet", userState.UserWallet,
		"userStateVersion", userState.Version,
		"asset", userState.Asset,
		"amount", userState.Transition.Amount)
}
