package app_session_v1

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// CreateAppSession creates a new application session between participants.
// App sessions are created with 0 allocations by default, as per V1 API specification.
// Deposits must be done through the submit_deposit_state endpoint.
func (h *Handler) CreateAppSession(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1CreateAppSessionRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if len(reqPayload.Definition.Participants) > h.maxParticipants {
		c.Fail(rpc.Errorf("participants array exceeds maximum length of %d", h.maxParticipants), "")
		return
	}
	if len(reqPayload.QuorumSigs) > len(reqPayload.Definition.Participants) {
		c.Fail(rpc.Errorf("quorum_sigs count (%d) exceeds participants count (%d)", len(reqPayload.QuorumSigs), len(reqPayload.Definition.Participants)), "")
		return
	}
	if len(reqPayload.SessionData) > h.maxSessionData {
		c.Fail(rpc.Errorf("session_data exceeds maximum length of %d", h.maxSessionData), "")
		return
	}

	if reqPayload.Definition.Application == "" {
		c.Fail(nil, "application id is required")
		return
	}

	appDef, err := unmapAppDefinitionV1(reqPayload.Definition)
	if err != nil {
		c.Fail(err, "invalid app definition")
		return
	}

	logger.Debug("processing app session creation request",
		"application", reqPayload.Definition.Application,
		"participantsNum", len(reqPayload.Definition.Participants),
		"quorum", reqPayload.Definition.Quorum,
		"nonce", reqPayload.Definition.Nonce)

	// Validate nonce
	if reqPayload.Definition.Nonce == "" || reqPayload.Definition.Nonce == "0" {
		c.Fail(nil, "nonce is zero or not provided")
		return
	}

	// Validate quorum is greater than zero
	if reqPayload.Definition.Quorum == 0 {
		c.Fail(nil, "quorum must be greater than zero")
		return
	}

	// Validate quorum against total weights and check for duplicate participants
	var totalWeights uint8
	participantWeights := make(map[string]uint8)
	for _, participant := range reqPayload.Definition.Participants {
		participantWallet := strings.ToLower(participant.WalletAddress)

		// Check for duplicate participant addresses
		if _, exists := participantWeights[participantWallet]; exists {
			c.Fail(rpc.Errorf("duplicate participant address: %s", participant.WalletAddress), "")
			return
		}
		totalWeights += participant.SignatureWeight
		participantWeights[participantWallet] = participant.SignatureWeight
	}

	if reqPayload.Definition.Quorum > totalWeights {
		c.Fail(rpc.Errorf("target quorum (%d) cannot be greater than total sum of weights (%d)",
			reqPayload.Definition.Quorum, totalWeights), "")
		return
	}

	// Validate signatures and quorum
	if len(reqPayload.QuorumSigs) == 0 {
		c.Fail(nil, "no signatures provided")
		return
	}

	// Pack the request for signature verification
	packedRequest, err := app.PackCreateAppSessionRequestV1(appDef, reqPayload.SessionData)
	if err != nil {
		c.Fail(rpc.Errorf("failed to pack request: %v", err), "")
		return
	}

	// Generate app session ID (deterministic)
	appSessionID, err := app.GenerateAppSessionIDV1(appDef)
	if err != nil {
		c.Fail(rpc.Errorf("failed to generate app session ID: %v", err), "")
		return
	}

	err = h.useStoreInTx(func(tx Store) error {
		if h.appRegistryEnabled {
			registeredApp, err := tx.GetApp(appDef.ApplicationID)
			if err != nil {
				return rpc.Errorf("failed to look up application: %v", err)
			}

			// App must be registered regardless of CreationApprovalNotRequired flag.
			if registeredApp == nil {
				return rpc.Errorf("application %s is not registered", appDef.ApplicationID)
			}

			if !registeredApp.App.CreationApprovalNotRequired {
				if reqPayload.OwnerSig == "" {
					return rpc.Errorf("owner_sig is required for this application")
				}

				sigBytes, err := hexutil.Decode(reqPayload.OwnerSig)
				if err != nil {
					return rpc.Errorf("failed to decode signature: %v", err)
				}
				if len(sigBytes) == 0 {
					return rpc.Errorf("empty owner_sig after decode")
				}

				sigType := app.AppSessionSignerTypeV1(sigBytes[0])
				appSessionSignerValidator := app.NewAppSessionKeySigValidatorV1(
					func(sessionKeyAddr string) (string, error) {
						return tx.GetAppSessionKeyOwner(sessionKeyAddr, appSessionID)
					},
				)
				recoveredOwnerWallet, err := appSessionSignerValidator.Recover(packedRequest, sigBytes)
				if err != nil {
					h.metrics.IncAppSessionUpdateSigValidation(appSessionID, sigType, false)
					return rpc.Errorf("failed to recover user wallet: %v", err)
				}
				h.metrics.IncAppSessionUpdateSigValidation(appSessionID, sigType, true)

				if !strings.EqualFold(recoveredOwnerWallet, registeredApp.App.OwnerWallet) {
					return rpc.Errorf("invalid owner signature: signer %s is not the app owner", recoveredOwnerWallet)
				}
			}

			err = h.actionGateway.AllowAction(tx, registeredApp.App.OwnerWallet, core.GatedActionAppSessionCreation)
			if err != nil {
				return rpc.NewError(err)
			}
		}

		if err := h.verifyQuorum(tx, appSessionID, appDef.ApplicationID, participantWeights, appDef.Quorum, packedRequest, reqPayload.QuorumSigs); err != nil {
			return err
		}

		// Create app session with 0 allocations
		appSession := app.AppSessionV1{
			SessionID:     appSessionID,
			ApplicationID: appDef.ApplicationID,
			Participants:  appDef.Participants,
			Quorum:        appDef.Quorum,
			Nonce:         appDef.Nonce,
			Status:        app.AppSessionStatusOpen,
			Version:       1,
			SessionData:   reqPayload.SessionData,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := tx.CreateAppSession(appSession); err != nil {
			return rpc.Errorf("failed to create app session: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to create app session", "error", err)
		c.Fail(err, "failed to create app session")
		return
	}

	resp := rpc.AppSessionsV1CreateAppSessionResponse{
		AppSessionID: appSessionID,
		Version:      "1",
		Status:       app.AppSessionStatusOpen.String(),
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully created app session", "appSessionID", appSessionID)
}
