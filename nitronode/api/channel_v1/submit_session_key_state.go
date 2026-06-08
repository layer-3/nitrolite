package channel_v1

import (
	"errors"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/nitronode/store/database"
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

	if strings.EqualFold(coreState.UserAddress, coreState.SessionKey) {
		c.Fail(rpc.Errorf("invalid_session_key_state: session_key must differ from user_address"), "")
		return
	}

	if coreState.Version == 0 {
		c.Fail(rpc.Errorf("invalid_session_key_state: version must be greater than 0"), "")
		return
	}
	// Past expires_at is permitted as a revocation signal. The auth path filters
	// expires_at > now so a past timestamp deactivates the key immediately while keeping
	// the same monotonic version sequence (a later submit with a future expires_at can
	// re-activate the key). A negative unix timestamp is rejected because the
	// metadata-hash packer casts int64 -> uint64 (wraps to a huge future value), which
	// would cause the user-signed payload to disagree with the value persisted in the
	// database — defense-in-depth even though the DB filter is the source of truth.
	if coreState.ExpiresAt.Unix() < 0 {
		c.Fail(rpc.Errorf("invalid_session_key_state: expires_at must be non-negative"), "")
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

	// A submit with expires_at after now activates, extends, or rotates the key and requires
	// both signatures. A submit with expires_at <= now is a revocation: it only deactivates an
	// existing delegation, so the wallet's user_sig alone authorizes it. Requiring session_key_sig
	// on the revocation path would let a lost, unavailable, or malicious session key veto its own
	// revocation, stranding the user until the prior expires_at naturally passes. now is captured
	// here and reused for the cap/version logic inside the transaction so the active/inactive
	// decision is consistent across both.
	now := time.Now()
	if coreState.ExpiresAt.After(now) {
		if coreState.SessionKeySig == "" {
			c.Fail(rpc.Errorf("invalid_session_key_state: session_key_sig is required"), "")
			return
		}
		// Validate both signatures: wallet's user_sig and session-key holder's session_key_sig.
		if err := core.ValidateChannelSessionKeyStateV1(coreState); err != nil {
			c.Fail(rpc.Errorf("invalid_session_key_state: %v", err), "")
			return
		}
	} else {
		// Revocation only deactivates an existing delegation. Version 1 means there is no prior
		// delegation, so reject it before LockSessionKeyState can seed a permanent ownership
		// claim for a session_key the caller never proved possession of: a legitimate revoke is
		// always version >= 2, since registration at version 1 is future-dated and requires both
		// signatures.
		if coreState.Version == 1 {
			c.Fail(rpc.Errorf("invalid_session_key_state: cannot revoke a session key with no prior delegation"), "")
			return
		}
		// Validate only the wallet's user_sig. Any session_key_sig present is ignored, so clear
		// it before persisting to keep stored revocation rows canonical.
		if err := core.ValidateChannelSessionKeyStateUserSigV1(coreState); err != nil {
			c.Fail(rpc.Errorf("invalid_session_key_state: %v", err), "")
			return
		}
		coreState.SessionKeySig = ""
	}

	// revoked is true only when this submit transitions an active key to inactive,
	// so the revocation log is not emitted for inactive-to-inactive updates.
	var revoked bool
	err = h.useStoreInTx(func(tx Store) error {
		// Lock the (user, session_key, channel) pointer row for the duration of the tx so that
		// concurrent submits for the same (user, session_key) serialize cleanly and report a
		// proper "expected version" error rather than racing on the history UNIQUE constraint.
		latestVersion, latestExpiresAt, err := tx.LockSessionKeyState(coreState.UserAddress, coreState.SessionKey, database.SessionKeyKindChannel)
		if err != nil {
			if errors.Is(err, database.ErrSessionKeyNotAllowed) {
				logger.Warn("session key registration collision",
					"userAddress", coreState.UserAddress,
					"sessionKey", coreState.SessionKey,
					"kind", database.SessionKeyKindChannel)
				return rpc.Errorf("invalid_session_key_state: session_key not allowed")
			}
			return rpc.Errorf("failed to lock session key state: %v", err)
		}

		// Enforce the per-user cap whenever this submit transitions the slot from inactive
		// to active — i.e. a brand-new key (latestVersion == 0) or a reactivation where the
		// previous latest state was already past its expires_at. A rotation/update against a
		// still-active key is not counted again so legitimate rotation is never blocked, and
		// a revoke submit (submitted expires_at <= now) decreases the active count so it is
		// not subject to the cap either.
		//
		// Without the reactivation check a user at the cap can revoke key A, register fresh
		// key B into the freed slot, then re-submit key A with a future expires_at — the
		// `latestVersion > 0` branch would skip the cap check and leave the user above the
		// cap.
		//
		// TODO(MF-H01-followup): the row lock above only serializes submits for the same
		// (user, session_key, kind), so two concurrent submits registering *different* new keys
		// for the same user can both observe the same count and both pass the check, ending up
		// at most maxSessionKeysPerUser + (concurrent new-key writers - 1) keys. The cap is a
		// soft DOS bound, not a hard quota — a small over-shoot under genuine concurrency is
		// acceptable. If a hard quota is ever required, take a per-user advisory lock here
		// (pg_advisory_xact_lock(hashtext(user_address))) before counting.
		prevActive := latestVersion > 0 && latestExpiresAt.After(now)
		submittedActive := coreState.ExpiresAt.After(now)
		revoked = prevActive && !submittedActive
		if !prevActive && submittedActive && h.maxSessionKeysPerUser > 0 {
			count, err := tx.CountSessionKeysForUser(coreState.UserAddress)
			if err != nil {
				return rpc.Errorf("failed to count session keys for user: %v", err)
			}
			if count >= uint32(h.maxSessionKeysPerUser) {
				return rpc.Errorf("invalid_session_key_state: user has reached the session key limit of %d", h.maxSessionKeysPerUser)
			}
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
	if revoked {
		logger.Info("channel session key revoked",
			"userAddress", coreState.UserAddress,
			"sessionKey", coreState.SessionKey,
			"version", coreState.Version,
			"expiresAt", coreState.ExpiresAt)
		return
	}
	logger.Info("successfully stored channel session key state",
		"userAddress", coreState.UserAddress,
		"sessionKey", coreState.SessionKey,
		"version", coreState.Version)
}
