// Package rpc provides a high-level client for interacting with the Nitrolite Node RPC server.
//
// This file implements the V1 API client with versioned request/response types
// following the api.yaml specification.
package rpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Client provides a high-level interface for interacting with the Nitrolite Node V1 RPC API.
// It wraps a Dialer to provide convenient methods for all V1 RPC operations.
//
// The Client supports:
//   - Channel management (home and escrow channels)
//   - State queries and submissions
//   - Application session operations
//   - Session key management
//   - User balance and transaction queries
//   - Node configuration and asset queries
//
// Example usage:
//
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
//
//	// Connect to the server
//	err := client.Start(ctx, "wss://server.example.com/ws", handleError)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Make RPC calls
//	config, err := client.NodeV1GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
type Client struct {
	dialer Dialer
}

// NewClient creates a new V1 RPC client using the provided dialer.
// The dialer must be connected before making RPC calls.
func NewClient(dialer Dialer) *Client {
	return &Client{
		dialer: dialer,
	}
}

// Start establishes a connection to the RPC server.
// This is a convenience method that wraps the dialer's Dial method.
func (c *Client) Start(ctx context.Context, url string, handleClosure func(err error)) error {
	return c.dialer.Dial(ctx, url, handleClosure)
}

// ============================================================================
// Channels Group - V1 API Methods
// ============================================================================

// ChannelsV1GetHomeChannel retrieves current on-chain home channel information.
func (c *Client) ChannelsV1GetHomeChannel(ctx context.Context, req ChannelsV1GetHomeChannelRequest) (ChannelsV1GetHomeChannelResponse, error) {
	var resp ChannelsV1GetHomeChannelResponse
	if err := c.call(ctx, ChannelsV1GetHomeChannelMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1GetEscrowChannel retrieves current on-chain escrow channel information.
func (c *Client) ChannelsV1GetEscrowChannel(ctx context.Context, req ChannelsV1GetEscrowChannelRequest) (ChannelsV1GetEscrowChannelResponse, error) {
	var resp ChannelsV1GetEscrowChannelResponse
	if err := c.call(ctx, ChannelsV1GetEscrowChannelMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1GetChannels retrieves all channels for a user with optional filtering.
func (c *Client) ChannelsV1GetChannels(ctx context.Context, req ChannelsV1GetChannelsRequest) (ChannelsV1GetChannelsResponse, error) {
	var resp ChannelsV1GetChannelsResponse
	if err := c.call(ctx, ChannelsV1GetChannelsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1GetLatestState retrieves the current state of the user stored on the Node.
func (c *Client) ChannelsV1GetLatestState(ctx context.Context, req ChannelsV1GetLatestStateRequest) (ChannelsV1GetLatestStateResponse, error) {
	var resp ChannelsV1GetLatestStateResponse
	if err := c.call(ctx, ChannelsV1GetLatestStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1GetStates retrieves state history for a user with optional filtering.
func (c *Client) ChannelsV1GetStates(ctx context.Context, req ChannelsV1GetStatesRequest) (ChannelsV1GetStatesResponse, error) {
	var resp ChannelsV1GetStatesResponse
	if err := c.call(ctx, ChannelsV1GetStatesMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1RequestCreation requests the creation of a channel from Node.
func (c *Client) ChannelsV1RequestCreation(ctx context.Context, req ChannelsV1RequestCreationRequest) (ChannelsV1RequestCreationResponse, error) {
	var resp ChannelsV1RequestCreationResponse
	if err := c.call(ctx, ChannelsV1RequestCreationMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1SubmitState submits a cross-chain state.
func (c *Client) ChannelsV1SubmitState(ctx context.Context, req ChannelsV1SubmitStateRequest) (ChannelsV1SubmitStateResponse, error) {
	var resp ChannelsV1SubmitStateResponse
	if err := c.call(ctx, ChannelsV1SubmitStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1SubmitSessionKeyState submits a channel session key state for registration or update.
func (c *Client) ChannelsV1SubmitSessionKeyState(ctx context.Context, req ChannelsV1SubmitSessionKeyStateRequest) (ChannelsV1SubmitSessionKeyStateResponse, error) {
	var resp ChannelsV1SubmitSessionKeyStateResponse
	if err := c.call(ctx, ChannelsV1SubmitSessionKeyStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ChannelsV1GetLastKeyStates retrieves the latest channel session key states for a user.
func (c *Client) ChannelsV1GetLastKeyStates(ctx context.Context, req ChannelsV1GetLastKeyStatesRequest) (ChannelsV1GetLastKeyStatesResponse, error) {
	var resp ChannelsV1GetLastKeyStatesResponse
	if err := c.call(ctx, ChannelsV1GetLastKeyStatesMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ============================================================================
// App Sessions Group - V1 API Methods
// ============================================================================

// AppSessionsV1SubmitDepositState submits an application session state update.
func (c *Client) AppSessionsV1SubmitDepositState(ctx context.Context, req AppSessionsV1SubmitDepositStateRequest) (AppSessionsV1SubmitDepositStateResponse, error) {
	var resp AppSessionsV1SubmitDepositStateResponse
	if err := c.call(ctx, AppSessionsV1SubmitDepositStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1SubmitAppState submits an application session state update.
func (c *Client) AppSessionsV1SubmitAppState(ctx context.Context, req AppSessionsV1SubmitAppStateRequest) (AppSessionsV1SubmitAppStateResponse, error) {
	var resp AppSessionsV1SubmitAppStateResponse
	if err := c.call(ctx, AppSessionsV1SubmitAppStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1GetAppDefinition retrieves the application definition for a specific app session.
func (c *Client) AppSessionsV1GetAppDefinition(ctx context.Context, req AppSessionsV1GetAppDefinitionRequest) (AppSessionsV1GetAppDefinitionResponse, error) {
	var resp AppSessionsV1GetAppDefinitionResponse
	if err := c.call(ctx, AppSessionsV1GetAppDefinitionMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1GetAppSessions lists all application sessions for a participant with optional filtering.
func (c *Client) AppSessionsV1GetAppSessions(ctx context.Context, req AppSessionsV1GetAppSessionsRequest) (AppSessionsV1GetAppSessionsResponse, error) {
	var resp AppSessionsV1GetAppSessionsResponse
	if err := c.call(ctx, AppSessionsV1GetAppSessionsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1CreateAppSession creates a new application session between participants.
func (c *Client) AppSessionsV1CreateAppSession(ctx context.Context, req AppSessionsV1CreateAppSessionRequest) (AppSessionsV1CreateAppSessionResponse, error) {
	var resp AppSessionsV1CreateAppSessionResponse
	if err := c.call(ctx, AppSessionsV1CreateAppSessionMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1RebalanceAppSessions rebalances multiple application sessions atomically.
func (c *Client) AppSessionsV1RebalanceAppSessions(ctx context.Context, req AppSessionsV1RebalanceAppSessionsRequest) (AppSessionsV1RebalanceAppSessionsResponse, error) {
	var resp AppSessionsV1RebalanceAppSessionsResponse
	if err := c.call(ctx, AppSessionsV1RebalanceAppSessionsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1Register initiates session key registration.
func (c *Client) AppSessionsV1SubmitSessionKeyState(ctx context.Context, req AppSessionsV1SubmitSessionKeyStateRequest) (AppSessionsV1SubmitSessionKeyStateRequest, error) {
	var resp AppSessionsV1SubmitSessionKeyStateRequest
	if err := c.call(ctx, AppSessionsV1SubmitSessionKeyStateMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppSessionsV1GetSessionKeys retrieves all active session keys for the authenticated user.
func (c *Client) AppSessionsV1GetLastKeyStates(ctx context.Context, req AppSessionsV1GetLastKeyStatesRequest) (AppSessionsV1GetLastKeyStatesResponse, error) {
	var resp AppSessionsV1GetLastKeyStatesResponse
	if err := c.call(ctx, AppSessionsV1GetLastKeyStatesMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ============================================================================
// Apps Group - V1 API Methods
// ============================================================================

// AppsV1GetApps retrieves registered applications with optional filtering.
func (c *Client) AppsV1GetApps(ctx context.Context, req AppsV1GetAppsRequest) (AppsV1GetAppsResponse, error) {
	var resp AppsV1GetAppsResponse
	if err := c.call(ctx, AppsV1GetAppsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// AppsV1SubmitAppVersion submits a new application version (currently only creation is supported).
func (c *Client) AppsV1SubmitAppVersion(ctx context.Context, req AppsV1SubmitAppVersionRequest) (AppsV1SubmitAppVersionResponse, error) {
	var resp AppsV1SubmitAppVersionResponse
	if err := c.call(ctx, AppsV1SubmitAppVersionMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ============================================================================
// User Group - V1 API Methods
// ============================================================================

// UserV1GetBalances retrieves the balances of the user in YN.
func (c *Client) UserV1GetBalances(ctx context.Context, req UserV1GetBalancesRequest) (UserV1GetBalancesResponse, error) {
	var resp UserV1GetBalancesResponse
	if err := c.call(ctx, UserV1GetBalancesMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// UserV1GetTransactions retrieves ledger transaction history with optional filtering.
func (c *Client) UserV1GetTransactions(ctx context.Context, req UserV1GetTransactionsRequest) (UserV1GetTransactionsResponse, error) {
	var resp UserV1GetTransactionsResponse
	if err := c.call(ctx, UserV1GetTransactionsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// UserV1GetActionAllowances retrieves the user's current action allowances for channels and app sessions.
func (c *Client) UserV1GetActionAllowances(ctx context.Context, req UserV1GetActionAllowancesRequest) (UserV1GetActionAllowancesResponse, error) {
	var resp UserV1GetActionAllowancesResponse
	if err := c.call(ctx, UserV1GetActionAllowancesMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ============================================================================
// Node Group - V1 API Methods
// ============================================================================

// NodeV1Ping sends a ping request to the server to check connectivity.
func (c *Client) NodeV1Ping(ctx context.Context) error {
	req := NodeV1PingRequest{}
	var resp NodeV1PingResponse
	return c.call(ctx, NodeV1PingMethod, req, &resp)
}

// NodeV1GetConfig retrieves broker configuration and supported networks.
func (c *Client) NodeV1GetConfig(ctx context.Context) (NodeV1GetConfigResponse, error) {
	req := NodeV1GetConfigRequest{}
	var resp NodeV1GetConfigResponse
	if err := c.call(ctx, NodeV1GetConfigMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// NodeV1GetAssets retrieves all supported assets with optional chain filter.
func (c *Client) NodeV1GetAssets(ctx context.Context, req NodeV1GetAssetsRequest) (NodeV1GetAssetsResponse, error) {
	var resp NodeV1GetAssetsResponse
	if err := c.call(ctx, NodeV1GetAssetsMethod, req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ============================================================================
// Internal Helper Methods
// ============================================================================

// call is an internal helper that makes an RPC call with the given method and parameters.
func (c *Client) call(ctx context.Context, method Method, reqParams any, respParams any) error {
	params, err := NewPayload(reqParams)
	if err != nil {
		return fmt.Errorf("failed to create payload: %w", err)
	}

	req := NewRequest(
		uint64(uuid.New().ID()),
		method.String(),
		params,
	)

	res, err := c.dialer.Call(ctx, &req)
	if err != nil {
		return fmt.Errorf("rpc call failed: %w", err)
	}

	if err := res.Error(); err != nil {
		return fmt.Errorf("rpc returned error: %w", err)
	}

	if err := res.Payload.Translate(respParams); err != nil {
		return fmt.Errorf("failed to translate response: %w", err)
	}

	return nil
}
