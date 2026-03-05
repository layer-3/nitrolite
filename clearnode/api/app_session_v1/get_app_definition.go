package app_session_v1

import (
	"strconv"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetAppDefinition retrieves the application definition for a specific app session.
func (h *Handler) GetAppDefinition(c *rpc.Context) {
	var req rpc.AppSessionsV1GetAppDefinitionRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var definition rpc.AppDefinitionV1

	err := h.useStoreInTx(func(store Store) error {
		session, err := store.GetAppSession(req.AppSessionID)
		if err != nil {
			return err
		}

		if session == nil {
			return rpc.Errorf("app_session_not_found")
		}

		// Convert participants
		participants := make([]rpc.AppParticipantV1, len(session.Participants))
		for i, p := range session.Participants {
			participants[i] = rpc.AppParticipantV1{
				WalletAddress:   p.WalletAddress,
				SignatureWeight: p.SignatureWeight,
			}
		}

		definition = rpc.AppDefinitionV1{
			Application:  session.ApplicationID,
			Participants: participants,
			Quorum:       session.Quorum,
			Nonce:        strconv.FormatUint(session.Nonce, 10),
		}

		return nil
	})

	if err != nil {
		c.Fail(err, "failed to retrieve app definition")
		return
	}

	response := rpc.AppSessionsV1GetAppDefinitionResponse{
		Definition: definition,
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
