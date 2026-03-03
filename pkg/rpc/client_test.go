package rpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// Test helpers for V1 client
var (
	testCtxV1      = context.Background()
	testWalletV1   = "0x1234"
	testWallet2V1  = "0x5678"
	testChainIDV1  = "1"
	testTokenV1    = "0xUSDC"
	testSymbolV1   = "USDC"
	testAssetV1    = "usdc"
	testChannelID  = "ch123"
	testAppSession = "app123"
)

// setupClient creates a test V1 client with mock dialer
func setupClient() (*rpc.Client, *MockDialer) {
	mockDialer := NewMockDialer()
	client := rpc.NewClient(mockDialer)
	return client, mockDialer
}

// createResponseV1 creates an RPC response with the given data
func createResponseV1[T any](method string, data T) (*rpc.Message, error) {
	params, err := rpc.NewPayload(data)
	if err != nil {
		return nil, err
	}
	res := rpc.NewResponse(0, method, params)
	return &res, nil
}

// registerSimpleHandlerV1 registers a handler that returns the given response
func registerSimpleHandlerV1[T any](dialer *MockDialer, method string, response T) {
	dialer.RegisterHandler(rpc.Method(method), func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		return createResponseV1(method, response)
	})
}

// ============================================================================
// Node Group Tests
// ============================================================================

func TestClientV1_NodeV1Ping(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	registerSimpleHandlerV1(dialer, "node.v1.ping", rpc.NodeV1PingResponse{})

	err := client.NodeV1Ping(testCtxV1)
	assert.NoError(t, err)
}

func TestClientV1_NodeV1GetConfig(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	config := rpc.NodeV1GetConfigResponse{
		NodeAddress: testWalletV1,
		Blockchains: []rpc.BlockchainInfoV1{
			{BlockchainID: testChainIDV1, ChannelHubAddress: "0xContract"},
		},
	}

	registerSimpleHandlerV1(dialer, "node.v1.get_config", config)

	resp, err := client.NodeV1GetConfig(testCtxV1)
	require.NoError(t, err)
	assert.Equal(t, testWalletV1, resp.NodeAddress)
	assert.Len(t, resp.Blockchains, 1)
}

func TestClientV1_NodeV1GetAssets(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	assets := []rpc.AssetV1{
		{
			Name:                  "USD Coin",
			Symbol:                testSymbolV1,
			SuggestedBlockchainID: testChainIDV1,
			Tokens: []rpc.TokenV1{
				{Name: "USDC on Ethereum", Symbol: testSymbolV1, Address: testTokenV1, BlockchainID: testChainIDV1, Decimals: 6},
				{Name: "USDC on Polygon", Symbol: testSymbolV1, Address: "0xUSDC2", BlockchainID: "137", Decimals: 6},
			},
		},
		{
			Name:                  "Ethereum",
			Symbol:                "ETH",
			SuggestedBlockchainID: testChainIDV1,
			Tokens: []rpc.TokenV1{
				{Name: "ETH on Ethereum", Symbol: "ETH", Address: "0xETH", BlockchainID: testChainIDV1, Decimals: 18},
			},
		},
		{
			Name:                  "DAI",
			Symbol:                "DAI",
			SuggestedBlockchainID: "137",
			Tokens: []rpc.TokenV1{
				{Name: "DAI on Polygon", Symbol: "DAI", Address: "0xDAI", BlockchainID: "137", Decimals: 18},
			},
		},
	}

	registerSimpleHandlerV1(dialer, "node.v1.get_assets", rpc.NodeV1GetAssetsResponse{Assets: assets})

	resp, err := client.NodeV1GetAssets(testCtxV1, rpc.NodeV1GetAssetsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Assets, 3)
	assert.Equal(t, "USD Coin", resp.Assets[0].Name)
	assert.Len(t, resp.Assets[0].Tokens, 2)
}

// ============================================================================
// Channels Group Tests
// ============================================================================

func TestClientV1_ChannelsV1GetHomeChannel(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	channel := rpc.ChannelsV1GetHomeChannelResponse{
		Channel: rpc.ChannelV1{
			ChannelID:         testChannelID,
			UserWallet:        testWalletV1,
			Type:              "home",
			BlockchainID:      testChainIDV1,
			TokenAddress:      testTokenV1,
			ChallengeDuration: 3600,
			Nonce:             "1",
			Status:            "open",
			StateVersion:      "1",
		},
	}

	registerSimpleHandlerV1(dialer, "channels.v1.get_home_channel", channel)

	resp, err := client.ChannelsV1GetHomeChannel(testCtxV1, rpc.ChannelsV1GetHomeChannelRequest{
		Wallet: testWalletV1,
		Asset:  testAssetV1,
	})
	require.NoError(t, err)
	assert.Equal(t, testChannelID, resp.Channel.ChannelID)
	assert.Equal(t, "home", resp.Channel.Type)
}

func TestClientV1_ChannelsV1GetEscrowChannel(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	channel := rpc.ChannelsV1GetEscrowChannelResponse{
		Channel: rpc.ChannelV1{
			ChannelID:    testChannelID,
			Type:         "escrow",
			BlockchainID: testChainIDV1,
			Status:       "open",
		},
	}

	registerSimpleHandlerV1(dialer, "channels.v1.get_escrow_channel", channel)

	resp, err := client.ChannelsV1GetEscrowChannel(testCtxV1, rpc.ChannelsV1GetEscrowChannelRequest{
		EscrowChannelID: testChannelID,
	})
	require.NoError(t, err)
	assert.Equal(t, "escrow", resp.Channel.Type)
}

func TestClientV1_ChannelsV1GetChannels(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	channels := rpc.ChannelsV1GetChannelsResponse{
		Channels: []rpc.ChannelV1{
			{ChannelID: "ch1", UserWallet: testWalletV1, Status: "open"},
			{ChannelID: "ch2", UserWallet: testWalletV1, Status: "open"},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 2,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, "channels.v1.get_channels", channels)

	resp, err := client.ChannelsV1GetChannels(testCtxV1, rpc.ChannelsV1GetChannelsRequest{
		Wallet: testWalletV1,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Channels, 2)
	assert.Equal(t, uint32(2), resp.Metadata.TotalCount)
}

func TestClientV1_ChannelsV1GetLatestState(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	state := rpc.ChannelsV1GetLatestStateResponse{
		State: rpc.StateV1{
			ID:         "state123",
			Asset:      testAssetV1,
			UserWallet: testWalletV1,
			Epoch:      "1",
			Version:    "5",
			HomeLedger: rpc.LedgerV1{
				TokenAddress: testTokenV1,
				BlockchainID: testChainIDV1,
				UserBalance:  "1000",
				NodeBalance:  "500",
			},
		},
	}

	registerSimpleHandlerV1(dialer, "channels.v1.get_latest_state", state)

	resp, err := client.ChannelsV1GetLatestState(testCtxV1, rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     testWalletV1,
		Asset:      testAssetV1,
		OnlySigned: false,
	})
	require.NoError(t, err)
	assert.Equal(t, "state123", resp.State.ID)
	assert.Equal(t, testAssetV1, resp.State.Asset)
}

func TestClientV1_ChannelsV1GetStates(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	states := rpc.ChannelsV1GetStatesResponse{
		States: []rpc.StateV1{
			{ID: "state1", Version: "1", Asset: testAssetV1},
			{ID: "state2", Version: "2", Asset: testAssetV1},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 2,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, "channels.v1.get_states", states)

	resp, err := client.ChannelsV1GetStates(testCtxV1, rpc.ChannelsV1GetStatesRequest{
		Wallet:     testWalletV1,
		Asset:      testAssetV1,
		OnlySigned: false,
	})
	require.NoError(t, err)
	assert.Len(t, resp.States, 2)
}

func TestClientV1_ChannelsV1RequestCreation(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.ChannelsV1RequestCreationResponse{
		Signature: "0xsig123",
	}

	registerSimpleHandlerV1(dialer, "channels.v1.request_creation", response)

	resp, err := client.ChannelsV1RequestCreation(testCtxV1, rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:         "state123",
			UserWallet: testWalletV1,
			Asset:      testAssetV1,
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:     "1",
			Challenge: 3600,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "0xsig123", resp.Signature)
}

func TestClientV1_ChannelsV1SubmitState(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.ChannelsV1SubmitStateResponse{
		Signature: "0xsig456",
	}

	registerSimpleHandlerV1(dialer, "channels.v1.submit_state", response)

	resp, err := client.ChannelsV1SubmitState(testCtxV1, rpc.ChannelsV1SubmitStateRequest{
		State: rpc.StateV1{
			ID:      "state123",
			Version: "2",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "0xsig456", resp.Signature)
}

// ============================================================================
// App Sessions Group Tests
// ============================================================================

func TestClientV1_AppSessionsV1GetAppDefinition(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	definition := rpc.AppSessionsV1GetAppDefinitionResponse{
		Definition: rpc.AppDefinitionV1{
			Application: "game",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: testWalletV1, SignatureWeight: 1},
				{WalletAddress: testWallet2V1, SignatureWeight: 1},
			},
			Quorum: 2,
			Nonce:  "1",
		},
	}

	registerSimpleHandlerV1(dialer, "app_sessions.v1.get_app_definition", definition)

	resp, err := client.AppSessionsV1GetAppDefinition(testCtxV1, rpc.AppSessionsV1GetAppDefinitionRequest{
		AppSessionID: testAppSession,
	})
	require.NoError(t, err)
	assert.Equal(t, "game", resp.Definition.Application)
	assert.Len(t, resp.Definition.Participants, 2)
}

func TestClientV1_AppSessionsV1GetAppSessions(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	sessions := rpc.AppSessionsV1GetAppSessionsResponse{
		AppSessions: []rpc.AppSessionInfoV1{
			{
				AppSessionID: testAppSession,
				Status:       "open",
				Participants: []rpc.AppParticipantV1{
					{WalletAddress: testWalletV1, SignatureWeight: 1},
				},
				Quorum:  1,
				Version: "1",
				Nonce:   "1",
			},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 1,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, "app_sessions.v1.get_app_sessions", sessions)

	resp, err := client.AppSessionsV1GetAppSessions(testCtxV1, rpc.AppSessionsV1GetAppSessionsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.AppSessions, 1)
	assert.Equal(t, testAppSession, resp.AppSessions[0].AppSessionID)
}

func TestClientV1_AppSessionsV1CreateAppSession(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppSessionsV1CreateAppSessionResponse{
		AppSessionID: testAppSession,
		Version:      "1",
		Status:       "open",
	}

	registerSimpleHandlerV1(dialer, "app_sessions.v1.create_app_session", response)

	resp, err := client.AppSessionsV1CreateAppSession(testCtxV1, rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "game",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: testWalletV1, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "1",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, testAppSession, resp.AppSessionID)
	assert.Equal(t, "open", resp.Status)
}

func TestClientV1_AppSessionsV1SubmitDepositState(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppSessionsV1SubmitDepositStateResponse{
		StateNodeSig: "0xsig789",
	}

	registerSimpleHandlerV1(dialer, "app_sessions.v1.submit_deposit_state", response)

	resp, err := client.AppSessionsV1SubmitDepositState(testCtxV1, rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: testAppSession,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      "2",
		},
		UserState: rpc.StateV1{ID: "state123"},
	})
	require.NoError(t, err)
	assert.Equal(t, "0xsig789", resp.StateNodeSig)
}

func TestClientV1_AppSessionsV1SubmitAppState(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppSessionsV1SubmitAppStateResponse{}

	registerSimpleHandlerV1(dialer, "app_sessions.v1.submit_app_state", response)

	_, err := client.AppSessionsV1SubmitAppState(testCtxV1, rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: testAppSession,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "3",
		},
		QuorumSigs: []string{"0xsig1", "0xsig2"},
	})
	require.NoError(t, err)
}

func TestClientV1_AppSessionsV1SubmitSessionKeyState(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppSessionsV1SubmitSessionKeyStateResponse{}

	registerSimpleHandlerV1(dialer, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), response)

	_, err := client.AppSessionsV1SubmitSessionKeyState(testCtxV1, rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    testWalletV1,
			SessionKey:     "0xsession_key_1",
			ApplicationIDs: []string{"0xapp_1"},
			AppSessionIDs:  []string{"0xapp_session_2"},
			Version:        "1",
			ExpiresAt:      fmt.Sprintf("%d", time.Now().Add(24*time.Hour).UTC().Unix()),
		},
	})
	require.NoError(t, err)
}

func TestClientV1_AppSessionsV1GetLastKeyStates(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppSessionsV1GetLastKeyStatesResponse{
		States: []rpc.AppSessionKeyStateV1{
			{
				UserAddress:    testWalletV1,
				SessionKey:     "0xsession_key_1",
				ApplicationIDs: []string{"0xapp_1"},
				AppSessionIDs:  []string{"0xapp_session_2"},
				Version:        "1",
				ExpiresAt:      fmt.Sprintf("%d", time.Now().Add(24*time.Hour).UTC().Unix()),
			},
		},
	}

	registerSimpleHandlerV1(dialer, rpc.AppSessionsV1GetLastKeyStatesMethod.String(), response)

	res, err := client.AppSessionsV1GetLastKeyStates(testCtxV1, rpc.AppSessionsV1GetLastKeyStatesRequest{
		UserAddress: testWalletV1,
		SessionKey:  nil,
	})
	require.NoError(t, err)

	assert.Len(t, res.States, 1)
	assert.Equal(t, "0xsession_key_1", res.States[0].SessionKey)
}

// ============================================================================
// Apps Group Tests
// ============================================================================

func TestClientV1_AppsV1GetApps(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppsV1GetAppsResponse{
		Apps: []rpc.AppInfoV1{
			{
				AppV1: rpc.AppV1{
					ID:                          "my-app",
					OwnerWallet:                 testWalletV1,
					Metadata:                    `{"name": "My App"}`,
					Version:                     "1",
					CreationApprovalNotRequired: true,
				},
				CreatedAt: "1700000000",
				UpdatedAt: "1700000001",
			},
			{
				AppV1: rpc.AppV1{
					ID:                          "another-app",
					OwnerWallet:                 testWallet2V1,
					Metadata:                    `{"name": "Another App"}`,
					Version:                     "1",
					CreationApprovalNotRequired: false,
				},
				CreatedAt: "1700000100",
				UpdatedAt: "1700000200",
			},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 2,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, rpc.AppsV1GetAppsMethod.String(), response)

	resp, err := client.AppsV1GetApps(testCtxV1, rpc.AppsV1GetAppsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Apps, 2)
	assert.Equal(t, "my-app", resp.Apps[0].ID)
	assert.Equal(t, testWalletV1, resp.Apps[0].OwnerWallet)
	assert.Equal(t, `{"name": "My App"}`, resp.Apps[0].Metadata)
	assert.Equal(t, "1", resp.Apps[0].Version)
	assert.True(t, resp.Apps[0].CreationApprovalNotRequired)
	assert.Equal(t, "1700000000", resp.Apps[0].CreatedAt)
	assert.Equal(t, "1700000001", resp.Apps[0].UpdatedAt)
	assert.Equal(t, uint32(2), resp.Metadata.TotalCount)
}

func TestClientV1_AppsV1GetApps_WithFilters(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	appID := "my-app"
	ownerWallet := testWalletV1

	response := rpc.AppsV1GetAppsResponse{
		Apps: []rpc.AppInfoV1{
			{
				AppV1: rpc.AppV1{
					ID:          appID,
					OwnerWallet: ownerWallet,
					Metadata:    `{}`,
					Version:     "1",
				},
				CreatedAt: "1700000000",
				UpdatedAt: "1700000000",
			},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 1,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, rpc.AppsV1GetAppsMethod.String(), response)

	resp, err := client.AppsV1GetApps(testCtxV1, rpc.AppsV1GetAppsRequest{
		AppID:       &appID,
		OwnerWallet: &ownerWallet,
		Pagination: &rpc.PaginationParamsV1{
			Offset: ptrUint32(0),
			Limit:  ptrUint32(10),
		},
	})
	require.NoError(t, err)
	assert.Len(t, resp.Apps, 1)
	assert.Equal(t, appID, resp.Apps[0].ID)
}

func TestClientV1_AppsV1SubmitAppVersion(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	response := rpc.AppsV1SubmitAppVersionResponse{}

	registerSimpleHandlerV1(dialer, rpc.AppsV1SubmitAppVersionMethod.String(), response)

	_, err := client.AppsV1SubmitAppVersion(testCtxV1, rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:                          "my-app",
			OwnerWallet:                 testWalletV1,
			Metadata:                    `{"name": "My App"}`,
			Version:                     "1",
			CreationApprovalNotRequired: false,
		},
		OwnerSig: "0xsig123",
	})
	require.NoError(t, err)
}

// ============================================================================
// User Group Tests
// ============================================================================

func TestClientV1_UserV1GetBalances(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	balances := rpc.UserV1GetBalancesResponse{
		Balances: []rpc.BalanceEntryV1{
			{Asset: testAssetV1, Amount: "1000"},
			{Asset: "eth", Amount: "5"},
		},
	}

	registerSimpleHandlerV1(dialer, "user.v1.get_balances", balances)

	resp, err := client.UserV1GetBalances(testCtxV1, rpc.UserV1GetBalancesRequest{
		Wallet: testWalletV1,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Balances, 2)
	assert.Equal(t, "1000", resp.Balances[0].Amount)
}

func TestClientV1_UserV1GetTransactions(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	transactions := rpc.UserV1GetTransactionsResponse{
		Transactions: []rpc.TransactionV1{
			{
				ID:          "tx1",
				Asset:       testAssetV1,
				FromAccount: testWalletV1,
				ToAccount:   testWallet2V1,
				Amount:      "100",
				CreatedAt:   "2025-01-01T00:00:00Z",
			},
		},
		Metadata: rpc.PaginationMetadataV1{
			Page:       1,
			PerPage:    10,
			TotalCount: 1,
			PageCount:  1,
		},
	}

	registerSimpleHandlerV1(dialer, "user.v1.get_transactions", transactions)

	resp, err := client.UserV1GetTransactions(testCtxV1, rpc.UserV1GetTransactionsRequest{
		Wallet: testWalletV1,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Transactions, 1)
	assert.Equal(t, "100", resp.Transactions[0].Amount)
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestClientV1_ErrorHandling(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	// No handler registered
	_, err := client.NodeV1GetConfig(testCtxV1)
	assert.Contains(t, err.Error(), "method not found")

	// Handler returns error response
	dialer.RegisterHandler(rpc.Method("node.v1.get_assets"), func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		res := rpc.NewErrorResponse(0, "node.v1.get_assets", "internal server error")
		return &res, nil
	})

	_, err = client.NodeV1GetAssets(testCtxV1, rpc.NodeV1GetAssetsRequest{})
	assert.Contains(t, err.Error(), "internal server error")
}

// ============================================================================
// Test Helpers
// ============================================================================

func ptrUint32(v uint32) *uint32 {
	return &v
}
