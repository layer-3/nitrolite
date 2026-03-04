package app_session_v1

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// RebalanceAppSessions processes multi-session rebalancing operations atomically.
// Rebalancing redistributes funds across multiple app sessions in a single atomic operation,
// potentially involving multiple assets. Each asset's balance changes must sum to zero (conservation).
func (h *Handler) RebalanceAppSessions(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1RebalanceAppSessionsRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if len(reqPayload.SignedUpdates) > h.maxSignedUpdates {
		c.Fail(rpc.Errorf("signed_updates array exceeds maximum length of %d", h.maxSignedUpdates), "")
		return
	}

	if len(reqPayload.SignedUpdates) < 2 {
		c.Fail(rpc.Errorf("rebalancing requires at least 2 sessions"), "")
		return
	}
	logger.Debug("processing app session rebalancing request", "sessionCount", len(reqPayload.SignedUpdates))

	// Parse and validate all app state updates
	updates := make([]app.SignedAppStateUpdateV1, len(reqPayload.SignedUpdates))
	seenSessions := make(map[string]bool)

	for i, signedUpdate := range reqPayload.SignedUpdates {
		if len(signedUpdate.AppStateUpdate.SessionData) > h.maxSessionData {
			c.Fail(rpc.Errorf("signed_updates[%d].session_data exceeds maximum length of %d", i, h.maxSessionData), "")
			return
		}

		update, err := unmapSignedAppStateUpdateV1(&signedUpdate)
		if err != nil {
			c.Fail(err, fmt.Sprintf("failed to parse app state update %d", i))
			return
		}

		// Validate intent is rebalance
		if update.AppStateUpdate.Intent != app.AppStateUpdateIntentRebalance {
			c.Fail(rpc.Errorf("all updates must have 'rebalance' intent, got '%s' for session %s",
				update.AppStateUpdate.Intent.String(), update.AppStateUpdate.AppSessionID), "")
			return
		}

		// Only one app state update per app session is allowed
		if seenSessions[update.AppStateUpdate.AppSessionID] {
			c.Fail(rpc.Errorf("duplicate session in rebalance: %s", update.AppStateUpdate.AppSessionID), "")
			return
		}
		seenSessions[update.AppStateUpdate.AppSessionID] = true

		updates[i] = update

		logger.Debug("parsed rebalance update",
			"sessionID", update.AppStateUpdate.AppSessionID,
			"version", update.AppStateUpdate.Version,
			"allocations", len(update.AppStateUpdate.Allocations))
	}

	var batchID string
	err := h.useStoreInTx(func(tx Store) error {
		// Generate deterministic batch ID from session IDs and versions
		sessionVersions := make([]app.AppSessionVersionV1, len(updates))
		for i, u := range updates {
			sessionVersions[i] = app.AppSessionVersionV1{
				SessionID: u.AppStateUpdate.AppSessionID,
				Version:   u.AppStateUpdate.Version,
			}
		}
		var err error
		batchID, err = app.GenerateRebalanceBatchIDV1(sessionVersions)
		if err != nil {
			return rpc.Errorf("failed to generate batch ID: %v", err)
		}

		// Track all balance changes per session per participant per asset
		balanceChanges := make(map[string]map[string]map[string]decimal.Decimal) // sessionID -> participant -> asset -> change
		assetTotalDiff := make(map[string]decimal.Decimal)                       // asset -> total change

		// Validate and process each session
		for _, update := range updates {
			appSession, err := tx.GetAppSession(update.AppStateUpdate.AppSessionID)
			if err != nil {
				return rpc.Errorf("failed to get app session %s: %v", update.AppStateUpdate.AppSessionID, err)
			}
			if appSession == nil {
				return rpc.Errorf("app session not found: %s", update.AppStateUpdate.AppSessionID)
			}
			registeredApp, err := tx.GetApp(appSession.ApplicationID)
			if err != nil {
				return rpc.Errorf("failed to look up application: %v", err)
			}
			if registeredApp == nil {
				return rpc.Errorf("application %s is not registered", appSession.ApplicationID)
			}

			err = h.actionGateway.AllowAction(tx, registeredApp.App.OwnerWallet, update.AppStateUpdate.Intent.GatedAction())
			if err != nil {
				return rpc.NewError(err)
			}
			if len(update.QuorumSigs) > len(appSession.Participants) {
				return rpc.Errorf("quorum_sigs count (%d) exceeds participants count (%d)", len(update.QuorumSigs), len(appSession.Participants))
			}
			if appSession.Status == app.AppSessionStatusClosed {
				return rpc.Errorf("app session %s is already closed", update.AppStateUpdate.AppSessionID)
			}
			if update.AppStateUpdate.Version != appSession.Version+1 {
				return rpc.Errorf("invalid version for session %s: expected %d, got %d",
					update.AppStateUpdate.AppSessionID, appSession.Version+1, update.AppStateUpdate.Version)
			}

			// Verify quorum
			participantWeights := getParticipantWeights(appSession.Participants)
			if len(update.QuorumSigs) == 0 {
				return rpc.Errorf("no signatures provided for session %s", update.AppStateUpdate.AppSessionID)
			}

			packedStateUpdate, err := app.PackAppStateUpdateV1(update.AppStateUpdate)
			if err != nil {
				return rpc.Errorf("failed to pack app state update for session %s: %v", update.AppStateUpdate.AppSessionID, err)
			}

			if err := h.verifyQuorum(tx, update.AppStateUpdate.AppSessionID, appSession.ApplicationID, participantWeights, appSession.Quorum, packedStateUpdate, update.QuorumSigs); err != nil {
				return rpc.Errorf("quorum verification failed for session %s: %v", update.AppStateUpdate.AppSessionID, err)
			}

			// Get current allocations
			currentAllocations, err := tx.GetParticipantAllocations(update.AppStateUpdate.AppSessionID)
			if err != nil {
				return rpc.Errorf("failed to get current allocations for session %s: %v", update.AppStateUpdate.AppSessionID, err)
			}

			// Build map of new allocations
			newAllocations := make(map[string]map[string]decimal.Decimal) // participant -> asset -> amount
			for _, alloc := range update.AppStateUpdate.Allocations {
				// Validate participant exists
				if _, ok := participantWeights[alloc.Participant]; !ok {
					return rpc.Errorf("allocation to non-participant %s in session %s", alloc.Participant, update.AppStateUpdate.AppSessionID)
				}

				if alloc.Amount.IsNegative() {
					return rpc.Errorf("negative allocation: %s for asset %s in session %s",
						alloc.Amount, alloc.Asset, update.AppStateUpdate.AppSessionID)
				}

				if newAllocations[alloc.Participant] == nil {
					newAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
				}
				newAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
			}

			// Calculate balance changes for this session per participant
			sessionChanges := make(map[string]map[string]decimal.Decimal) // participant -> asset -> change

			// Process all assets in current allocations
			for participant, assets := range currentAllocations {
				for asset, currentAmount := range assets {
					newAmount := decimal.Zero
					if newAllocations[participant] != nil {
						newAmount = newAllocations[participant][asset]
					}

					change := newAmount.Sub(currentAmount)
					if !change.IsZero() {
						if sessionChanges[participant] == nil {
							sessionChanges[participant] = make(map[string]decimal.Decimal)
						}
						sessionChanges[participant][asset] = change
						assetTotalDiff[asset] = assetTotalDiff[asset].Add(change)
					}
				}
			}

			// Process any new assets in new allocations
			for participant, assets := range newAllocations {
				for asset, newAmount := range assets {
					if currentAllocations[participant] == nil || currentAllocations[participant][asset].IsZero() {
						if !newAmount.IsZero() {
							if sessionChanges[participant] == nil {
								sessionChanges[participant] = make(map[string]decimal.Decimal)
							}
							sessionChanges[participant][asset] = newAmount
							assetTotalDiff[asset] = assetTotalDiff[asset].Add(newAmount)
						}
					}
				}
			}

			// Store session changes
			balanceChanges[update.AppStateUpdate.AppSessionID] = sessionChanges

			// Update app session
			appSession.Version++
			if update.AppStateUpdate.SessionData != "" {
				appSession.SessionData = update.AppStateUpdate.SessionData
			}
			appSession.UpdatedAt = time.Now()

			if err := tx.UpdateAppSession(*appSession); err != nil {
				return rpc.Errorf("failed to update app session %s: %v", update.AppStateUpdate.AppSessionID, err)
			}
		}

		// Validate conservation: sum of changes must be zero for each asset
		for asset, total := range assetTotalDiff {
			if !total.IsZero() {
				return rpc.Errorf("conservation violation for asset %s: total change is %s (must be 0)",
					asset, total.String())
			}
		}

		// Record ledger entries per participant
		for sessionID, participantChanges := range balanceChanges {
			for participant, assetChanges := range participantChanges {
				for asset, diff := range assetChanges {
					if diff.IsZero() {
						continue
					}

					// Record ledger entry with user wallet (participant)
					if err := tx.RecordLedgerEntry(participant, sessionID, asset, diff); err != nil {
						return rpc.Errorf("failed to record ledger entry for session %s: %v", sessionID, err)
					}
				}
			}
		}

		// Calculate aggregate changes per session per asset and record transactions
		sessionAssetChanges := make(map[string]map[string]decimal.Decimal) // sessionID -> asset -> aggregated change
		for sessionID, participantChanges := range balanceChanges {
			sessionAssetChanges[sessionID] = make(map[string]decimal.Decimal)
			for _, assetChanges := range participantChanges {
				for asset, diff := range assetChanges {
					sessionAssetChanges[sessionID][asset] = sessionAssetChanges[sessionID][asset].Add(diff)
				}
			}
		}

		// Record one transaction per session per asset
		for sessionID, assetChanges := range sessionAssetChanges {
			for asset, diff := range assetChanges {
				if diff.IsZero() {
					continue
				}

				// Record transaction for the aggregated change
				var fromAccount, toAccount string
				var amount decimal.Decimal

				if diff.IsPositive() {
					// Session gaining funds: batch -> session
					fromAccount = batchID
					toAccount = sessionID
					amount = diff
				} else {
					// Session losing funds: session -> batch
					fromAccount = sessionID
					toAccount = batchID
					amount = diff.Abs()
				}

				txID, err := app.GenerateRebalanceTransactionIDV1(batchID, sessionID, asset)
				if err != nil {
					return rpc.Errorf("failed to generate transaction ID for session %s: %v", sessionID, err)
				}

				transaction := core.NewTransaction(
					txID,
					asset,
					core.TransactionTypeRebalance,
					fromAccount,
					toAccount,
					nil, // No sender state for app sessions
					nil, // No receiver state for app sessions
					amount,
				)

				if err := tx.RecordTransaction(*transaction); err != nil {
					return rpc.Errorf("failed to record transaction for session %s: %v", sessionID, err)
				}

				logger.Info("recorded transaction",
					"txID", transaction.ID,
					"txType", transaction.TxType.String(),
					"from", transaction.FromAccount,
					"to", transaction.ToAccount,
					"asset", transaction.Asset,
					"amount", transaction.Amount.String())
			}
		}

		logger.Info("processed app session rebalancing",
			"batchID", batchID,
			"sessionCount", len(updates),
			"assetCount", len(assetTotalDiff))

		return nil
	})

	if err != nil {
		logger.Error("failed to process rebalancing", "error", err)
		c.Fail(err, "failed to process rebalancing")
		return
	}

	resp := rpc.AppSessionsV1RebalanceAppSessionsResponse{
		BatchID: batchID,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
