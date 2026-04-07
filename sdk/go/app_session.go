package sdk

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

// ============================================================================
// App Session Methods
// ============================================================================

// GetAppSessionsOptions contains optional filters for GetAppSessions.
type GetAppSessionsOptions struct {
	// AppSessionID filters by application session ID
	AppSessionID *string

	// Participant filters by participant wallet address
	Participant *string

	// Status filters by status ("open" or "closed")
	Status *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetAppSessions retrieves application sessions with optional filtering.
//
// Parameters:
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of AppSession
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	sessions, meta, err := client.GetAppSessions(ctx, nil)
//	for _, session := range sessions {
//	    fmt.Printf("Session %s: %d participants\n", session.AppSessionID, len(session.Participants))
//	}
func (c *Client) GetAppSessions(ctx context.Context, opts *GetAppSessionsOptions) ([]app.AppSessionInfoV1, core.PaginationMetadata, error) {
	req := rpc.AppSessionsV1GetAppSessionsRequest{}
	if opts != nil {
		req.AppSessionID = opts.AppSessionID
		req.Participant = opts.Participant
		req.Status = opts.Status
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.AppSessionsV1GetAppSessions(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get app sessions: %w", err)
	}

	appSessions, err := transformAppSessions(resp.AppSessions)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to transform app sessions: %w", err)
	}

	return appSessions, transformPaginationMetadata(resp.Metadata), nil
}

// GetAppDefinition retrieves the definition for a specific app session.
//
// Parameters:
//   - appSessionID: The application session ID
//
// Returns:
//   - app.AppDefinitionV1 with participants, quorum, and application info
//   - Error if the request fails
//
// Example:
//
//	def, err := client.GetAppDefinition(ctx, "session123")
//	fmt.Printf("App: %s, Quorum: %d\n", def.Application, def.Quorum)
func (c *Client) GetAppDefinition(ctx context.Context, appSessionID string) (*app.AppDefinitionV1, error) {
	if appSessionID == "" {
		return nil, fmt.Errorf("app session ID required")
	}
	req := rpc.AppSessionsV1GetAppDefinitionRequest{
		AppSessionID: appSessionID,
	}
	resp, err := c.rpcClient.AppSessionsV1GetAppDefinition(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get app definition: %w", err)
	}

	def, err := transformAppDefinition(resp.Definition)
	if err != nil {
		return nil, fmt.Errorf("failed to transform app definition: %w", err)
	}
	return &def, nil
}

// CreateAppSessionOptions contains optional parameters for CreateAppSession.
type CreateAppSessionOptions struct {
	// OwnerSig is the app owner's signature approving session creation.
	// Required when the app's CreationApprovalNotRequired is false.
	OwnerSig string
}

// CreateAppSession creates a new application session between participants.
//
// Parameters:
//   - definition: The app definition with participants, quorum, application ID
//   - sessionData: Optional JSON stringified session data
//   - quorumSigs: Participant signatures for the app session creation
//   - opts: Optional parameters (owner signature for apps requiring approval)
//
// Returns:
//   - AppSessionID of the created session
//   - Initial version of the session
//   - Status of the session
//   - Error if the request fails
//
// Example:
//
//	def := app.AppDefinitionV1{
//	    Application: "chess-v1",
//	    Participants: []app.AppParticipantV1{...},
//	    Quorum: 2,
//	    Nonce: 1,
//	}
//	sessionID, version, status, err := client.CreateAppSession(ctx, def, "{}", []string{"sig1", "sig2"})
func (c *Client) CreateAppSession(ctx context.Context, definition app.AppDefinitionV1, sessionData string, quorumSigs []string, opts ...CreateAppSessionOptions) (string, string, string, error) {
	req := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition:  transformAppDefinitionToRPC(definition),
		SessionData: sessionData,
		QuorumSigs:  quorumSigs,
	}
	if len(opts) > 0 && opts[0].OwnerSig != "" {
		req.OwnerSig = opts[0].OwnerSig
	}
	resp, err := c.rpcClient.AppSessionsV1CreateAppSession(ctx, req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create app session: %w", err)
	}
	return resp.AppSessionID, resp.Version, resp.Status, nil
}

// SubmitAppSessionDeposit submits a deposit to an app session.
// This updates both the app session state and the user's channel state.
//
// Parameters:
//   - appStateUpdate: The app state update with deposit intent
//   - quorumSigs: Participant signatures for the app state update
//   - userState: The user's updated channel state
//
// Returns:
//   - Node's signature for the state
//   - Error if the request fails
//
// Example:
//
//	appUpdate := app.AppStateUpdateV1{
//	    AppSessionID: "session123",
//	    Intent: app.AppStateUpdateIntentDeposit,
//	    Version: 2,
//	    Allocations: []app.AppAllocationV1{...},
//	}
//	nodeSig, err := client.SubmitAppSessionDeposit(ctx, appUpdate, []string{"sig1"}, userState)
func (c *Client) SubmitAppSessionDeposit(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string, asset string, depositAmount decimal.Decimal) (string, error) {
	// TODO: Would be good to only have appStateUpdate and quorumSigs here, as userState can be built inside.
	appUpdate := transformAppStateUpdateToRPC(appStateUpdate)

	currentState, err := c.GetLatestState(ctx, c.GetUserAddress(), asset, false)
	if err != nil {
		return "", fmt.Errorf("failed to get latest state: %w", err)
	}

	if currentState == nil {
		return "", fmt.Errorf("no channel state to advance for AppSession")
	}

	nextState := currentState.NextState()

	_, err = nextState.ApplyCommitTransition(appUpdate.AppSessionID, depositAmount)
	if err != nil {
		return "", fmt.Errorf("failed to apply commit transition: %w", err)
	}

	stateSig, err := c.ValidateAndSignState(currentState, nextState)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}

	nextState.UserSig = &stateSig
	req := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appUpdate,
		QuorumSigs:     quorumSigs,
		UserState:      transformStateToRPC(*nextState),
	}

	resp, err := c.rpcClient.AppSessionsV1SubmitDepositState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit deposit state: %w", err)
	}
	return resp.StateNodeSig, nil
}

// SubmitAppState submits an app session state update.
// This method handles operate, withdraw, and close intents.
// For deposits, use SubmitAppSessionDeposit instead.
//
// Parameters:
//   - appStateUpdate: The app state update (intent: operate, withdraw, or close)
//   - quorumSigs: Participant signatures for the app state update
//
// Returns:
//   - Error if the request fails
//
// Example:
//
//	appUpdate := app.AppStateUpdateV1{
//	    AppSessionID: "session123",
//	    Intent: app.AppStateUpdateIntentOperate,
//	    Version: 3,
//	    Allocations: []app.AppAllocationV1{...},
//	}
//	err := client.SubmitAppState(ctx, appUpdate, []string{"sig1", "sig2"})
func (c *Client) SubmitAppState(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string) error {
	appUpdate := transformAppStateUpdateToRPC(appStateUpdate)

	req := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: appUpdate,
		QuorumSigs:     quorumSigs,
	}
	_, err := c.rpcClient.AppSessionsV1SubmitAppState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to submit app state: %w", err)
	}
	return nil
}

// RebalanceAppSessions rebalances multiple application sessions atomically.
//
// This method performs atomic rebalancing across multiple app sessions, ensuring
// that funds are redistributed consistently without the risk of partial updates.
//
// Parameters:
//   - signedUpdates: Slice of signed app state updates to apply atomically
//
// Returns:
//   - BatchID for tracking the rebalancing operation
//   - Error if the request fails
//
// Example:
//
//	updates := []app.SignedAppStateUpdateV1{...}
//	batchID, err := client.RebalanceAppSessions(ctx, updates)
//	fmt.Printf("Rebalance batch ID: %s\n", batchID)
func (c *Client) RebalanceAppSessions(ctx context.Context, signedUpdates []app.SignedAppStateUpdateV1) (string, error) {
	// Transform SDK types to RPC types
	rpcUpdates := make([]rpc.SignedAppStateUpdateV1, 0, len(signedUpdates))
	for _, update := range signedUpdates {
		rpcUpdate := transformSignedAppStateUpdateToRPC(update)
		rpcUpdates = append(rpcUpdates, rpcUpdate)
	}

	req := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: rpcUpdates,
	}

	resp, err := c.rpcClient.AppSessionsV1RebalanceAppSessions(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to rebalance app sessions: %w", err)
	}

	return resp.BatchID, nil
}

// ============================================================================
// Session Key Methods
// ============================================================================

// SubmitAppSessionKeyState submits a session key state for registration or update.
// The state must be signed by the user's wallet to authorize the session key delegation.
//
// Parameters:
//   - state: The session key state containing delegation information
//
// Returns:
//   - Error if the request fails
//
// Example:
//
//	state := app.AppSessionKeyStateV1{
//	    UserAddress:    "0x1234...",
//	    SessionKey:     "0xabcd...",
//	    Version:        1,
//	    ApplicationIDs: []string{"app1"},
//	    AppSessionIDs:  []string{},
//	    ExpiresAt:      time.Now().Add(24 * time.Hour),
//	    UserSig:        "0x...",
//	}
//	err := client.SubmitAppSessionKeyState(ctx, state)
func (c *Client) SubmitAppSessionKeyState(ctx context.Context, state app.AppSessionKeyStateV1) error {
	req := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: transformSessionKeyStateToRPC(state),
	}
	_, err := c.rpcClient.AppSessionsV1SubmitSessionKeyState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to submit session key state: %w", err)
	}
	return nil
}

// GetLastKeyStatesOptions contains optional filters for GetLastKeyStates.
type GetLastKeyStatesOptions struct {
	// SessionKey filters by a specific session key address
	SessionKey *string
}

// GetLastAppKeyStates retrieves the latest session key states for a user.
//
// Parameters:
//   - userAddress: The user's wallet address
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of AppSessionKeyStateV1 with the latest non-expired session key states
//   - Error if the request fails
//
// Example:
//
//	states, err := client.GetLastAppKeyStates(ctx, "0x1234...", nil)
//	for _, state := range states {
//	    fmt.Printf("Session key %s expires at %s\n", state.SessionKey, state.ExpiresAt)
//	}
func (c *Client) GetLastAppKeyStates(ctx context.Context, userAddress string, opts *GetLastKeyStatesOptions) ([]app.AppSessionKeyStateV1, error) {
	req := rpc.AppSessionsV1GetLastKeyStatesRequest{
		UserAddress: userAddress,
	}
	if opts != nil {
		req.SessionKey = opts.SessionKey
	}

	resp, err := c.rpcClient.AppSessionsV1GetLastKeyStates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get last key states: %w", err)
	}

	states, err := transformSessionKeyStates(resp.States)
	if err != nil {
		return nil, fmt.Errorf("failed to transform session key states: %w", err)
	}

	return states, nil
}

// SignSessionKeyState signs a session key state using the client's state signer.
// This creates a properly formatted signature that can be set on the state's UserSig field
// before submitting via SubmitSessionKeyState.
//
// Parameters:
//   - state: The session key state to sign (UserSig field is excluded from signing)
//
// Returns:
//   - The hex-encoded signature string
//   - Error if signing fails
//
// Example:
//
//	state := app.AppSessionKeyStateV1{
//	    UserAddress:    client.GetUserAddress(),
//	    SessionKey:     "0xabcd...",
//	    Version:        1,
//	    ApplicationIDs: []string{},
//	    AppSessionIDs:  []string{},
//	    ExpiresAt:      time.Now().Add(24 * time.Hour),
//	}
//	sig, err := client.SignSessionKeyState(state)
//	state.UserSig = sig
//	err = client.SubmitSessionKeyState(ctx, state)
func (c *Client) SignSessionKeyState(state app.AppSessionKeyStateV1) (string, error) {
	packed, err := app.PackAppSessionKeyStateV1(state)
	if err != nil {
		return "", fmt.Errorf("failed to pack session key state: %w", err)
	}

	sig, err := c.stateSigner.Sign(packed)
	if err != nil {
		return "", fmt.Errorf("failed to sign session key state: %w", err)
	}

	// Strip the channel signer type prefix byte; session key auth uses plain EIP-191 signatures
	return hexutil.Encode(sig[1:]), nil
}
