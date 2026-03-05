package app_session_v1

import (
	"context"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

// SubmitAppState processes app session state updates for operate, withdraw, and close intents.
// Deposit intents should use the SubmitDepositState endpoint instead.
func (h *Handler) SubmitAppState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1SubmitAppStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if len(reqPayload.AppStateUpdate.SessionData) > h.maxSessionData {
		c.Fail(rpc.Errorf("session_data exceeds maximum length of %d", h.maxSessionData), "")
		return
	}

	logger.Debug("processing app session state update request",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"version", reqPayload.AppStateUpdate.Version,
		"intent", reqPayload.AppStateUpdate.Intent)

	appStateUpd, err := unmapAppStateUpdateV1(&reqPayload.AppStateUpdate)
	if err != nil {
		c.Fail(err, "failed to parse app state update")
		return
	}

	// Ensure this is not a deposit intent (should use submit_deposit_state)
	if appStateUpd.Intent == app.AppStateUpdateIntentDeposit {
		c.Fail(rpc.Errorf("deposit intent must use submit_deposit_state endpoint"), "")
		return
	}

	// Validate intent is valid
	if appStateUpd.Intent != app.AppStateUpdateIntentOperate &&
		appStateUpd.Intent != app.AppStateUpdateIntentWithdraw &&
		appStateUpd.Intent != app.AppStateUpdateIntentClose {
		c.Fail(rpc.Errorf("invalid intent: %s", appStateUpd.Intent.String()), "")
		return
	}

	err = h.useStoreInTx(func(tx Store) error {
		appSession, err := tx.GetAppSession(appStateUpd.AppSessionID)
		if err != nil {
			return rpc.Errorf("app session not found: %v", err)
		}
		if appSession == nil {
			return rpc.Errorf("app session not found")
		}
		if appSession.Status == app.AppSessionStatusClosed {
			return rpc.Errorf("app session is already closed")
		}

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

		if len(reqPayload.QuorumSigs) > len(appSession.Participants) {
			return rpc.Errorf("quorum_sigs count (%d) exceeds participants count (%d)", len(reqPayload.QuorumSigs), len(appSession.Participants))
		}
		if appStateUpd.Version != appSession.Version+1 {
			return rpc.Errorf("invalid app session version: expected %d, got %d", appSession.Version+1, appStateUpd.Version)
		}

		participantWeights := getParticipantWeights(appSession.Participants)

		if len(reqPayload.QuorumSigs) == 0 {
			return rpc.Errorf("no signatures provided")
		}

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

		// Handle different intents
		switch appStateUpd.Intent {
		case app.AppStateUpdateIntentOperate:
			// For operate intent, total allocations per asset must match session balance (redistribution allowed)
			if err := h.handleOperateIntent(ctx, tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}

		case app.AppStateUpdateIntentWithdraw:
			// For withdraw intent, validate and record ledger changes
			if err := h.handleWithdrawIntent(ctx, tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}

		case app.AppStateUpdateIntentClose:
			// For close intent, validate final allocations and mark session as closed
			if err := h.handleCloseIntent(ctx, tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}
			appSession.Status = app.AppSessionStatusClosed
		}

		// Update app session version and data
		appSession.Version++
		if reqPayload.AppStateUpdate.SessionData != "" {
			appSession.SessionData = reqPayload.AppStateUpdate.SessionData
		}
		appSession.UpdatedAt = time.Now()

		if err := tx.UpdateAppSession(*appSession); err != nil {
			return rpc.Errorf("failed to update app session: %v", err)
		}

		logger.Info("processed app state update",
			"appSessionID", appSession.SessionID,
			"appSessionVersion", appSession.Version,
			"intent", appStateUpd.Intent.String(),
			"status", appSession.Status.String())

		return nil
	})

	if err != nil {
		logger.Error("failed to process app state update", "error", err)
		c.Fail(err, "failed to process app state update")
		return
	}

	resp := rpc.AppSessionsV1SubmitAppStateResponse{}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}

// handleOperateIntent processes operate intent by validating total allocations and recording ledger changes.
// Operate intent allows redistribution of funds between participants as long as the total per asset stays the same.
// Requires submitting full list of allocations even if some haven't changed.
func (h *Handler) handleOperateIntent(
	_ context.Context,
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	// Get session balances to verify total allocations
	sessionBalances, err := tx.GetAppSessionBalances(appStateUpd.AppSessionID)
	if err != nil {
		return rpc.Errorf("failed to get app session balances: %v", err)
	}

	// Build a map of incoming allocations for validation and lookup
	incomingAllocations := make(map[string]map[string]decimal.Decimal)
	allocationSum := make(map[string]decimal.Decimal)

	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

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

		// Sum up allocations per asset
		if existing, ok := allocationSum[alloc.Asset]; ok {
			allocationSum[alloc.Asset] = existing.Add(alloc.Amount)
		} else {
			allocationSum[alloc.Asset] = alloc.Amount
		}

		// Store in incoming allocations map
		if incomingAllocations[alloc.Participant] == nil {
			incomingAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
		}
		incomingAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
	}

	// Verify all current allocations are present in the incoming request
	for participant, assets := range currentAllocations {
		for asset, currentAmount := range assets {
			if currentAmount.IsZero() {
				continue
			}

			// Check if this participant+asset is included in the incoming request
			incomingAmount, found := incomingAllocations[participant][asset]
			if !found {
				return rpc.Errorf("operate intent missing allocation for participant %s, asset %s with current amount %s",
					participant, asset, currentAmount.String())
			}

			// Calculate the difference and record ledger entry if changed
			diff := incomingAmount.Sub(currentAmount)
			if !diff.IsZero() {
				if err := tx.RecordLedgerEntry(participant, appStateUpd.AppSessionID, asset, diff); err != nil {
					return rpc.Errorf("failed to record operate ledger entry: %v", err)
				}
			}
		}
	}

	// Record ledger entries for new allocations (participants that didn't have this asset before)
	for participant, assets := range incomingAllocations {
		for asset, incomingAmount := range assets {
			if incomingAmount.IsZero() {
				continue
			}

			// Check if this is a new allocation (not in current allocations)
			currentAmount := decimal.Zero
			if currentAllocations[participant] != nil {
				currentAmount = currentAllocations[participant][asset]
			}

			// If current amount is zero and incoming amount is non-zero, this is a new allocation
			if currentAmount.IsZero() && !incomingAmount.IsZero() {
				if err := tx.RecordLedgerEntry(participant, appStateUpd.AppSessionID, asset, incomingAmount); err != nil {
					return rpc.Errorf("failed to record new allocation ledger entry: %v", err)
				}
			}
		}
	}

	// Verify that total allocations per asset match session balances
	for asset, totalAlloc := range allocationSum {
		sessionBalance, ok := sessionBalances[asset]
		if !ok {
			sessionBalance = decimal.Zero
		}

		if !totalAlloc.Equal(sessionBalance) {
			return rpc.Errorf("operate intent allocation mismatch for asset %s: total allocations %s, session balance %s",
				asset, totalAlloc.String(), sessionBalance.String())
		}
	}

	// Verify all session balances are accounted for
	for asset, sessionBalance := range sessionBalances {
		if sessionBalance.IsZero() {
			continue
		}

		_, ok := allocationSum[asset]
		if !ok {
			return rpc.Errorf("operate intent missing allocations for asset %s with balance %s",
				asset, sessionBalance.String())
		}
	}

	return nil
}

// handleWithdrawIntent processes withdraw intent by validating and recording ledger changes.
// It also issues new channel states for participants receiving withdrawn funds.
// Requires submitting full list of allocations even if some haven't changed.
func (h *Handler) handleWithdrawIntent(
	ctx context.Context,
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	// Build incoming allocations map for validation
	incomingAllocations := make(map[string]map[string]decimal.Decimal)

	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

		if alloc.Amount.IsNegative() {
			return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
		}

		// Check for new allocations (reject if current is zero but incoming is non-zero)
		if !alloc.Amount.IsZero() {
			currentAmount := decimal.Zero
			if currentAllocations[alloc.Participant] != nil {
				currentAmount = currentAllocations[alloc.Participant][alloc.Asset]
			}

			if currentAmount.IsZero() {
				return rpc.Errorf("withdraw intent cannot add new allocation for participant %s, asset %s",
					alloc.Participant, alloc.Asset)
			}
		}

		// Store in incoming allocations map
		if incomingAllocations[alloc.Participant] == nil {
			incomingAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
		}
		incomingAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
	}

	// Verify all current allocations are present and process withdrawals
	for participant, assets := range currentAllocations {
		for asset, currentAmount := range assets {
			if currentAmount.IsZero() {
				continue
			}

			// Check if this participant+asset is included in the incoming request
			incomingAmount, found := incomingAllocations[participant][asset]
			if !found {
				return rpc.Errorf("withdraw intent missing allocation for participant %s, asset %s with current amount %s",
					participant, asset, currentAmount.String())
			}

			// For withdraw, amounts can only decrease or stay the same
			if incomingAmount.GreaterThan(currentAmount) {
				return rpc.Errorf("withdraw intent cannot increase allocations: participant %s, asset %s",
					participant, asset)
			}

			if incomingAmount.LessThan(currentAmount) {
				// Record the withdrawal (negative ledger entry for the session)
				withdrawAmount := currentAmount.Sub(incomingAmount)
				if err := tx.RecordLedgerEntry(participant, appStateUpd.AppSessionID, asset, withdrawAmount.Neg()); err != nil {
					return rpc.Errorf("failed to record withdrawal ledger entry: %v", err)
				}

				decimals, err := h.assetStore.GetAssetDecimals(asset)
				if err != nil {
					return rpc.Errorf("failed to get asset decimals: %v", err)
				}

				if err := core.ValidateDecimalPrecision(withdrawAmount, decimals); err != nil {
					return rpc.Errorf("invalid withdraw amount for allocation with asset %s and participant %s: %w", asset, participant, err)
				}

				// Issue new channel state for participant receiving withdrawn funds
				if err := h.issueReleaseReceiverState(ctx, tx, participant, asset, appStateUpd.AppSessionID, withdrawAmount); err != nil {
					return rpc.Errorf("failed to issue release state for participant %s: %v", participant, err)
				}
			}
		}
	}

	return nil
}

// handleCloseIntent processes close intent by validating that allocations match current state,
// then releasing ALL funds from the session back to participants with channel state issuance.
func (h *Handler) handleCloseIntent(
	ctx context.Context,
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	// Build a map of incoming allocations for easy lookup
	incomingAllocations := make(map[string]map[string]decimal.Decimal)
	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

		if alloc.Amount.IsNegative() {
			return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
		}

		if incomingAllocations[alloc.Participant] == nil {
			incomingAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
		}
		incomingAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
	}

	// Iterate over current allocations (source of truth) and verify they match incoming allocations
	for participant, assets := range currentAllocations {
		for asset, currentAmount := range assets {
			if currentAmount.IsZero() {
				continue
			}

			// Check if this participant+asset is included in the incoming request
			incomingAmount, found := incomingAllocations[participant][asset]
			if !found {
				return rpc.Errorf("close intent missing allocation for participant %s, asset %s with current amount %s",
					participant, asset, currentAmount.String())
			}

			// Verify amounts match exactly
			if !incomingAmount.Equal(currentAmount) {
				return rpc.Errorf("close intent requires allocations to match current state: participant %s, asset %s, current %s, provided %s",
					participant, asset, currentAmount.String(), incomingAmount.String())
			}
		}
	}

	// Verify there are no extra allocations in the request that don't exist in current state
	for participant, assets := range incomingAllocations {
		for asset, incomingAmount := range assets {
			currentAmount := decimal.Zero
			if currentAllocations[participant] != nil {
				currentAmount = currentAllocations[participant][asset]
			}

			// If incoming has an allocation but current doesn't (or is zero), reject
			if currentAmount.IsZero() && !incomingAmount.IsZero() {
				return rpc.Errorf("close intent contains unexpected allocation for participant %s, asset %s with amount %s",
					participant, asset, incomingAmount.String())
			}
		}
	}

	// Iterate over current allocations and release each non-zero amount
	for participant, assets := range currentAllocations {
		for asset, amount := range assets {
			if amount.IsZero() {
				continue
			}

			// Record negative ledger entry (funds leaving the session)
			if err := tx.RecordLedgerEntry(participant, appStateUpd.AppSessionID, asset, amount.Neg()); err != nil {
				return rpc.Errorf("failed to record close ledger entry: %v", err)
			}

			// Issue new channel state for participant receiving funds back
			if err := h.issueReleaseReceiverState(ctx, tx, participant, asset, appStateUpd.AppSessionID, amount); err != nil {
				return rpc.Errorf("failed to issue release state for participant %s: %v", participant, err)
			}
		}
	}

	return nil
}
