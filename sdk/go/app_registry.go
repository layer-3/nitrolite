package sdk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// ============================================================================
// App Registry Methods
// ============================================================================

// GetAppsOptions contains optional filters for GetApps.
type GetAppsOptions struct {
	// AppID filters by application ID
	AppID *string

	// OwnerWallet filters by owner wallet address
	OwnerWallet *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetApps retrieves registered applications with optional filtering.
//
// Parameters:
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of AppInfoV1 with application information
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	apps, meta, err := client.GetApps(ctx, nil)
//	for _, a := range apps {
//	    fmt.Printf("App %s owned by %s\n", a.App.ID, a.App.OwnerWallet)
//	}
func (c *Client) GetApps(ctx context.Context, opts *GetAppsOptions) ([]app.AppInfoV1, core.PaginationMetadata, error) {
	req := rpc.AppsV1GetAppsRequest{}
	if opts != nil {
		req.AppID = opts.AppID
		req.OwnerWallet = opts.OwnerWallet
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.AppsV1GetApps(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get apps: %w", err)
	}

	apps, err := transformApps(resp.Apps)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to transform apps: %w", err)
	}

	return apps, transformPaginationMetadata(resp.Metadata), nil
}

// RegisterApp registers a new application in the app registry.
// Currently only version 1 (creation) is supported.
//
// The method builds the app definition from the provided parameters,
// using the client's signer address as the owner wallet and version 1.
// It then packs and signs the definition automatically.
//
// Session key signers are not allowed to perform this action; the main
// wallet signer must be used.
//
// Parameters:
//   - appID: The application identifier
//   - metadata: The application metadata
//   - creationApprovalNotRequired: Whether sessions can be created without owner approval
//
// Returns:
//   - Error if the request fails
//
// Example:
//
//	err := client.RegisterApp(ctx, "my-app", `{"name": "My App"}`, false)
func (c *Client) RegisterApp(ctx context.Context, appID string, metadata string, creationApprovalNotRequired bool) error {
	appDef := app.AppV1{
		ID:                          appID,
		OwnerWallet:                 c.GetUserAddress(),
		Metadata:                    metadata,
		Version:                     1,
		CreationApprovalNotRequired: creationApprovalNotRequired,
	}

	packed, err := app.PackAppV1(appDef)
	if err != nil {
		return fmt.Errorf("failed to pack app: %w", err)
	}

	ethMsgSigner, err := sign.NewEthereumMsgSignerFromRaw(c.rawSigner)
	if err != nil {
		return fmt.Errorf("failed to create Ethereum message signer: %w", err)
	}

	sig, err := ethMsgSigner.Sign(packed)
	if err != nil {
		return fmt.Errorf("failed to sign app data: %w", err)
	}

	req := rpc.AppsV1SubmitAppVersionRequest{
		App:      transformAppToRPC(appDef),
		OwnerSig: sig.String(),
	}
	_, err = c.rpcClient.AppsV1SubmitAppVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register app: %w", err)
	}
	return nil
}

// ============================================================================
// App Registry Transformations
// ============================================================================

// transformApps converts RPC AppInfoV1 slice to app.AppInfoV1 slice.
func transformApps(apps []rpc.AppInfoV1) ([]app.AppInfoV1, error) {
	result := make([]app.AppInfoV1, 0, len(apps))
	for _, a := range apps {
		version, err := strconv.ParseUint(a.Version, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse app version: %w", err)
		}

		createdAtSec, err := strconv.ParseInt(a.CreatedAt, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		updatedAtSec, err := strconv.ParseInt(a.UpdatedAt, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		result = append(result, app.AppInfoV1{
			App: app.AppV1{
				ID:                          a.ID,
				OwnerWallet:                 a.OwnerWallet,
				Metadata:                    a.Metadata,
				Version:                     version,
				CreationApprovalNotRequired: a.CreationApprovalNotRequired,
			},
			CreatedAt: time.Unix(createdAtSec, 0),
			UpdatedAt: time.Unix(updatedAtSec, 0),
		})
	}
	return result, nil
}

// transformAppToRPC converts app.AppV1 to rpc.AppV1.
func transformAppToRPC(a app.AppV1) rpc.AppV1 {
	return rpc.AppV1{
		ID:                          a.ID,
		OwnerWallet:                 a.OwnerWallet,
		Metadata:                    a.Metadata,
		Version:                     strconv.FormatUint(a.Version, 10),
		CreationApprovalNotRequired: a.CreationApprovalNotRequired,
	}
}
