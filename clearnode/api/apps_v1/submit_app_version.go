package apps_v1

import (
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// SubmitAppVersion updates an entry in the app registry.
func (h *Handler) SubmitAppVersion(c *rpc.Context) {
	var req rpc.AppsV1SubmitAppVersionRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if !app.AppIDV1Regex.MatchString(req.App.ID) {
		c.Fail(rpc.Errorf("invalid app ID: should match regex %s", app.AppIDV1Regex.String()), "")
		return
	}
	if req.App.OwnerWallet == "" {
		c.Fail(nil, "owner_wallet is required")
		return
	}
	if req.OwnerSig == "" {
		c.Fail(nil, "owner_sig is required")
		return
	}
	if len(req.App.Metadata) > h.maxAppMetadataLen {
		c.Fail(rpc.Errorf("metadata exceeds maximum length of %d characters", h.maxAppMetadataLen), "")
		return
	}

	version, err := strconv.ParseUint(req.App.Version, 10, 64)
	if err != nil {
		c.Fail(err, "invalid version")
		return
	}

	// Only creation (version 1) is supported for now
	if version != 1 {
		c.Fail(nil, "only version 1 (creation) is currently supported")
		return
	}

	appEntry := app.AppV1{
		ID:                          strings.ToLower(req.App.ID),
		OwnerWallet:                 strings.ToLower(req.App.OwnerWallet),
		Metadata:                    req.App.Metadata,
		Version:                     version,
		CreationApprovalNotRequired: req.App.CreationApprovalNotRequired,
	}

	packedApp, err := app.PackAppV1(appEntry)
	if err != nil {
		c.Fail(rpc.Errorf("failed to pack app data: %v", err), "")
		return
	}

	sigBytes, err := hexutil.Decode(req.OwnerSig)
	if err != nil {
		c.Fail(rpc.Errorf("failed to decode owner signature: %v", err), "")
		return
	}

	sigValidator, err := sign.NewSigValidator(sign.TypeEthereumMsg)
	if err != nil {
		c.Fail(rpc.Errorf("failed to create signature validator: %v", err), "")
		return
	}

	if err := sigValidator.Verify(appEntry.OwnerWallet, packedApp, sigBytes); err != nil {
		c.Fail(rpc.Errorf("invalid owner signature: %v", err), "")
		return
	}

	if err := h.store.CreateApp(appEntry); err != nil {
		c.Fail(err, "failed to create app")
		return
	}

	resp := rpc.AppsV1SubmitAppVersionResponse{}
	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
