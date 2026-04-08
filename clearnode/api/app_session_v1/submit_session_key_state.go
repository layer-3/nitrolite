package app_session_v1

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// SubmitSessionKeyState processes session key state submissions for registration and updates.
func (h *Handler) SubmitSessionKeyState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1SubmitSessionKeyStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if len(reqPayload.State.ApplicationIDs) > h.maxSessionKeyIDs {
		c.Fail(rpc.Errorf("application_ids array exceeds maximum length of %d", h.maxSessionKeyIDs), "")
		return
	}
	if len(reqPayload.State.AppSessionIDs) > h.maxSessionKeyIDs {
		c.Fail(rpc.Errorf("app_session_ids array exceeds maximum length of %d", h.maxSessionKeyIDs), "")
		return
	}

	logger.Debug("processing session key state submission",
		"userAddress", reqPayload.State.UserAddress,
		"sessionKey", reqPayload.State.SessionKey,
		"version", reqPayload.State.Version)

	// Convert RPC type to core type
	coreState, err := unmapSessionKeyStateV1(&reqPayload.State)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: %v", err), "")
		return
	}

	// Validate required fields
	coreState.UserAddress, err = core.NormalizeHexAddress(coreState.UserAddress)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: invalid user_address: %v", err), "")
		return
	}

	coreState.SessionKey, err = core.NormalizeHexAddress(coreState.SessionKey)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: invalid session_key: %v", err), "")
		return
	}

	if coreState.Version == 0 {
		c.Fail(rpc.Errorf("invalid_session_key_state: version must be greater than 0"), "")
		return
	}
	if coreState.ExpiresAt.Before(time.Now()) {
		c.Fail(rpc.Errorf("invalid_session_key_state: expires_at must be in the future"), "")
		return
	}
	if coreState.UserSig == "" {
		c.Fail(rpc.Errorf("invalid_session_key_state: user_sig is required"), "")
		return
	}

	// Pack the session key state for signature verification (ABI encoding)
	packedState, err := app.PackAppSessionKeyStateV1(coreState)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: failed to pack state: %v", err), "")
		return
	}

	// Decode the user signature
	sigBytes, err := hexutil.Decode(coreState.UserSig)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: failed to decode user_sig: %v", err), "")
		return
	}

	// Recover signer address from signature using ECDSA recovery
	ethMsgRecoverer, err := sign.NewSigValidator(sign.TypeEthereumMsg)
	if err != nil {
		c.Fail(rpc.Errorf("internal_error: failed to create signature validator: %v", err), "")
		return
	}

	recoveredAddress, err := ethMsgRecoverer.Recover(packedState, sigBytes)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: failed to recover signer: %v", err), "")
		return
	}

	// Verify the recovered address matches user_address
	if !strings.EqualFold(recoveredAddress, coreState.UserAddress) {
		c.Fail(rpc.Errorf("invalid_session_key_state: signature does not match user_address"), "")
		return
	}

	// Validate version and store the session key state
	err = h.useStoreInTx(func(tx Store) error {
		// Check the latest version for this (user_address, session_key) pair; 0 means no state exists
		latestVersion, err := tx.GetLastAppSessionKeyVersion(coreState.UserAddress, coreState.SessionKey)
		if err != nil {
			return rpc.Errorf("failed to check existing session key state: %v", err)
		}

		if coreState.Version != latestVersion+1 {
			return rpc.Errorf("invalid_session_key_state: expected version %d, got %d", latestVersion+1, coreState.Version)
		}

		return tx.StoreAppSessionKeyState(coreState)
	})

	if err != nil {
		logger.Error("failed to store session key state", "error", err)
		c.Fail(err, "failed to store session key state")
		return
	}

	resp := rpc.AppSessionsV1SubmitSessionKeyStateResponse{}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully stored session key state",
		"userAddress", coreState.UserAddress,
		"sessionKey", coreState.SessionKey,
		"version", coreState.Version)
}
