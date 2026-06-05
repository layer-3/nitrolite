package sdk

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
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
// Returns (nil, nil) when no app session exists for the given ID — absence is
// a successful response, not an error.
//
// Parameters:
//   - appSessionID: The application session ID
//
// Returns:
//   - app.AppDefinitionV1 with participants, quorum, and application info, or
//     nil if absent
//   - Error if the request fails
//
// Example:
//
//	def, err := client.GetAppDefinition(ctx, "session123")
//	if err != nil { return err }
//	if def == nil { /* not found */ }
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

	if resp.Definition == nil {
		return nil, nil
	}

	def, err := transformAppDefinition(*resp.Definition)
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

// ============================================================================
// Session Key Methods
// ============================================================================

// SubmitAppSessionKeyState submits a session key state for registration, update,
// revocation, or re-activation. The state must carry both the wallet's UserSig
// (authorizing the delegation) and the session-key holder's SessionKeySig (proving
// possession of the key); submits without a valid SessionKeySig are rejected on every
// path, including revocation — the session key must co-sign its own deactivation.
// Wallet-only revocation (for a lost or compromised key) is not supported by this
// method.
//
// Set state.ExpiresAt to a future time to register or update the key. Set it to a
// value at or before time.Now() to revoke the key — the auth path filters by
// expires_at, so the key is deactivated immediately while keeping the same monotonic
// version sequence. A later submit with the next version and a future ExpiresAt
// re-activates the same session key address. Negative unix timestamps are rejected
// by the server.
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
//	}
//	state.UserSig, _ = client.SignSessionKeyState(state)
//	state.SessionKeySig, _ = sdk.SignAppSessionKeyOwnership(state, sessionKeySigner)
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
	// IncludeInactive, when set to true, includes latest states whose expires_at is in
	// the past (expired or revoked). Defaults to false (active-only) when nil or false.
	IncludeInactive *bool
}

// GetLastAppKeyStates retrieves the latest app session key states for a user.
// By default only currently active (non-expired) states are returned; set
// opts.IncludeInactive to true to include expired or revoked latest states.
//
// Parameters:
//   - userAddress: The user's wallet address
//   - opts: Optional filters (pass nil for active-only with no session-key filter)
//
// Returns:
//   - Slice of AppSessionKeyStateV1 with the latest session key states matching the filter
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
		req.IncludeInactive = opts.IncludeInactive
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

// SignSessionKeyState produces the wallet UserSig over the session key state using the
// client's state signer. Set the returned hex on state.UserSig before submit. The matching
// session-key-holder SessionKeySig must also be populated (see SignAppSessionKeyOwnership)
// — submits with only one of the two are rejected.
//
// Parameters:
//   - state: The session key state to sign (UserSig and SessionKeySig fields are excluded from signing)
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
//	state.UserSig, _ = client.SignSessionKeyState(state)
//	state.SessionKeySig, _ = sdk.SignAppSessionKeyOwnership(state, sessionKeySigner)
//	err = client.SubmitAppSessionKeyState(ctx, state)
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

// SignAppSessionKeyOwnership produces the session-key holder's ownership signature over the
// packed app-session key state. The signer must be the holder of the session key being
// registered; the resulting hex-encoded signature is intended to populate state.SessionKeySig
// before submitting via SubmitAppSessionKeyState. The packed state already binds user_address,
// so replay across wallets is not possible.
//
// The parameter is narrowed to *sign.EthereumMsgSigner because the server recovers
// SessionKeySig under sign.TypeEthereumMsg — a broader signer interface could produce a
// signature without the EIP-191 prefix (or with extra wrapper bytes) that the server rejects.
func SignAppSessionKeyOwnership(state app.AppSessionKeyStateV1, sessionKeySigner *sign.EthereumMsgSigner) (string, error) {
	packed, err := app.PackAppSessionKeyStateV1(state)
	if err != nil {
		return "", fmt.Errorf("failed to pack session key state: %w", err)
	}

	sig, err := sessionKeySigner.Sign(packed)
	if err != nil {
		return "", fmt.Errorf("failed to sign session key ownership: %w", err)
	}

	return hexutil.Encode(sig), nil
}
