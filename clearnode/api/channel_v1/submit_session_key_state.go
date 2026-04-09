package channel_v1

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// SubmitSessionKeyState processes channel session key state submissions for registration and updates.
func (h *Handler) SubmitSessionKeyState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.ChannelsV1SubmitSessionKeyStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	logger.Debug("processing channel session key state submission",
		"userAddress", reqPayload.State.UserAddress,
		"sessionKey", reqPayload.State.SessionKey,
		"version", reqPayload.State.Version)

	// Convert RPC type to core type
	coreState, err := unmapChannelSessionKeyStateV1(&reqPayload.State)
	if err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: %v", err), "")
		return
	}

	// Validate required fields
	if !common.IsHexAddress(coreState.UserAddress) {
		c.Fail(rpc.Errorf("invalid_session_key_state: invalid user_address"), "")
		return
	}
	if !common.IsHexAddress(coreState.SessionKey) {
		c.Fail(rpc.Errorf("invalid_session_key_state: invalid session_key"), "")
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
	if len(coreState.Assets) > h.maxSessionKeyIDs {
		c.Fail(rpc.Errorf("invalid_session_key_state: assets array exceeds maximum length of %d", h.maxSessionKeyIDs), "")
		return
	}
	if coreState.UserSig == "" {
		c.Fail(rpc.Errorf("invalid_session_key_state: user_sig is required"), "")
		return
	}

	// Validate user's signature over the session key state
	if err := core.ValidateChannelSessionKeyAuthSigV1(coreState); err != nil {
		c.Fail(rpc.Errorf("invalid_session_key_state: %v", err), "")
		return
	}

	// Validate version and store the session key state
	err = h.useStoreInTx(func(tx Store) error {
		// Check the latest version for this (user_address, session_key) pair; 0 means no state exists
		latestVersion, err := tx.GetLastChannelSessionKeyVersion(coreState.UserAddress, coreState.SessionKey)
		if err != nil {
			return rpc.Errorf("failed to check existing session key state: %v", err)
		}

		if coreState.Version != latestVersion+1 {
			return rpc.Errorf("invalid_session_key_state: expected version %d, got %d", latestVersion+1, coreState.Version)
		}

		return tx.StoreChannelSessionKeyState(coreState)
	})

	if err != nil {
		logger.Error("failed to store channel session key state", "error", err)
		c.Fail(err, "failed to store channel session key state")
		return
	}

	resp := rpc.ChannelsV1SubmitSessionKeyStateResponse{}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully stored channel session key state",
		"userAddress", coreState.UserAddress,
		"sessionKey", coreState.SessionKey,
		"version", coreState.Version)
}
