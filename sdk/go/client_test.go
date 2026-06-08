package sdk

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetHomeChannel(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetHomeChannelResponse{
		Channel: &rpc.ChannelV1{
			ChannelID:    "0xChannelID",
			UserWallet:   "0xWallet",
			Type:         "home",
			BlockchainID: "137",
			Status:       "open",
			StateVersion: "1",
			Nonce:        "1",
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetHomeChannelMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetHomeChannel(context.Background(), "0xWallet", "USDC")
	require.NoError(t, err)
	assert.Equal(t, "0xChannelID", ch.ChannelID)
	assert.Equal(t, core.ChannelTypeHome, ch.Type)
}

// TestClient_GetHomeChannel_NilResponse verifies absent-channel responses surface as (nil, nil).
func TestClient_GetHomeChannel_NilResponse(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockDialer.RegisterResponse(rpc.ChannelsV1GetHomeChannelMethod.String(), rpc.ChannelsV1GetHomeChannelResponse{})

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetHomeChannel(context.Background(), "0xWallet", "USDC")
	require.NoError(t, err)
	assert.Nil(t, ch)
}

func TestClient_GetEscrowChannel(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetEscrowChannelResponse{
		Channel: &rpc.ChannelV1{
			ChannelID:    "0xEscrowID",
			UserWallet:   "0xWallet",
			Type:         "escrow",
			BlockchainID: "137",
			Status:       "open",
			StateVersion: "1",
			Nonce:        "1",
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetEscrowChannelMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetEscrowChannel(context.Background(), "0xEscrowID")
	require.NoError(t, err)
	assert.Equal(t, "0xEscrowID", ch.ChannelID)
	assert.Equal(t, core.ChannelTypeEscrow, ch.Type)
}

// TestClient_GetEscrowChannel_NilResponse verifies absent-channel responses surface as (nil, nil).
func TestClient_GetEscrowChannel_NilResponse(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockDialer.RegisterResponse(rpc.ChannelsV1GetEscrowChannelMethod.String(), rpc.ChannelsV1GetEscrowChannelResponse{})

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetEscrowChannel(context.Background(), "0xEscrowID")
	require.NoError(t, err)
	assert.Nil(t, ch)
}

func TestClient_GetLatestState(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetLatestStateResponse{
		State: &rpc.StateV1{
			ID:         "0xStateID",
			Epoch:      "1",
			Version:    "1",
			UserWallet: "0xWallet",
			Asset:      "USDC",
			Transition: rpc.TransitionV1{
				Type:   core.TransitionTypeTransferSend,
				Amount: "10.0",
			},
			HomeLedger: rpc.LedgerV1{
				BlockchainID: "137",
				UserBalance:  "100.0",
				UserNetFlow:  "0",
				NodeBalance:  "200.0",
				NodeNetFlow:  "0",
			},
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetLatestStateMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	state, err := client.GetLatestState(context.Background(), "0xWallet", "USDC", false)
	require.NoError(t, err)
	assert.Equal(t, "0xStateID", state.ID)
	assert.Equal(t, uint64(1), state.Version)
}

// TestClient_GetLatestState_NilResponse verifies absent-state responses surface as (nil, nil).
func TestClient_GetLatestState_NilResponse(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockDialer.RegisterResponse(rpc.ChannelsV1GetLatestStateMethod.String(), rpc.ChannelsV1GetLatestStateResponse{})

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	state, err := client.GetLatestState(context.Background(), "0xWallet", "USDC", false)
	require.NoError(t, err)
	assert.Nil(t, state)
}

func TestClient_GetBalances(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.UserV1GetBalancesResponse{
		Balances: []rpc.BalanceEntryV1{
			{Asset: "USDC", Amount: "100.0", Enforced: "0"},
		},
	}
	mockDialer.RegisterResponse(rpc.UserV1GetBalancesMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	bals, err := client.GetBalances(context.Background(), "0xWallet")
	require.NoError(t, err)
	assert.Len(t, bals, 1)
	assert.Equal(t, "USDC", bals[0].Asset)
	assert.Equal(t, "100", bals[0].Balance.String())
}

func TestClient_GetTransactions(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.UserV1GetTransactionsResponse{
		Transactions: []rpc.TransactionV1{
			{ID: "0xTxID", Asset: "USDC", Amount: "50.0", CreatedAt: "2023-01-01T00:00:00Z"},
		},
		Metadata: rpc.PaginationMetadataV1{
			TotalCount: 1,
		},
	}
	mockDialer.RegisterResponse(rpc.UserV1GetTransactionsMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	txs, meta, err := client.GetTransactions(context.Background(), "0xWallet", nil)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "0xTxID", txs[0].ID)
	assert.Equal(t, uint32(1), meta.TotalCount)
}

func TestClient_GetAppSessions(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1GetAppSessionsResponse{
		AppSessions: []rpc.AppSessionInfoV1{
			{
				AppSessionID: "0xSessionID",
				AppDefinitionV1: rpc.AppDefinitionV1{
					Participants: []rpc.AppParticipantV1{},
					Nonce:        "1",
				},
				Allocations: []rpc.AppAllocationV1{},
				Status:      "open",
				Version:     "1",
			},
		},
		Metadata: rpc.PaginationMetadataV1{TotalCount: 1},
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1GetAppSessionsMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	sessions, meta, err := client.GetAppSessions(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "0xSessionID", sessions[0].AppSessionID)
	assert.Equal(t, uint32(1), meta.TotalCount)
}

func TestClient_GetAppDefinition(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1GetAppDefinitionResponse{
		Definition: &rpc.AppDefinitionV1{
			Application:  "0xApp",
			Participants: []rpc.AppParticipantV1{},
			Nonce:        "1",
			Quorum:       1,
		},
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1GetAppDefinitionMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	def, err := client.GetAppDefinition(context.Background(), "0xSessionID")
	require.NoError(t, err)
	assert.Equal(t, "0xApp", def.ApplicationID)
	assert.Equal(t, uint64(1), def.Nonce)
}

// TestClient_GetAppDefinition_NilResponse verifies absent-definition responses surface as (nil, nil).
func TestClient_GetAppDefinition_NilResponse(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockDialer.RegisterResponse(rpc.AppSessionsV1GetAppDefinitionMethod.String(), rpc.AppSessionsV1GetAppDefinitionResponse{})

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	def, err := client.GetAppDefinition(context.Background(), "0xSessionID")
	require.NoError(t, err)
	assert.Nil(t, def)
}

func TestClient_CreateAppSession(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1CreateAppSessionResponse{
		AppSessionID: "0xSessionID",
		Version:      "1",
		Status:       "closed",
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1CreateAppSessionMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	def := app.AppDefinitionV1{
		ApplicationID: "chess-v1",
		Participants: []app.AppParticipantV1{
			{WalletAddress: "0xAlice", SignatureWeight: 1},
			{WalletAddress: "0xBob", SignatureWeight: 1},
		},
		Quorum: 2,
		Nonce:  1,
	}

	sessionID, version, status, err := client.CreateAppSession(context.Background(), def, "{}", []string{"sig1", "sig2"})
	require.NoError(t, err)
	assert.Equal(t, "0xSessionID", sessionID)
	assert.Equal(t, "1", version)
	assert.Equal(t, "closed", status)
}

func TestClient_SubmitAppSessionDeposit(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	// Mock assets for packing state
	assetsResp := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{
			{
				Name:                  "USDC",
				Symbol:                "USDC",
				Decimals:              6,
				SuggestedBlockchainID: "137",
				Tokens: []rpc.TokenV1{
					{BlockchainID: "137", Address: "0xToken", Decimals: 6},
				},
			},
		},
	}
	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), assetsResp)

	homeChannelID := "0xHomeChannel"
	// Mock latest state
	stateResp := rpc.ChannelsV1GetLatestStateResponse{
		State: &rpc.StateV1{
			ID:            "0xStateID",
			Epoch:         "1",
			Version:       "1",
			UserWallet:    "0xUser",
			Asset:         "USDC",
			HomeChannelID: &homeChannelID,
			Transition: rpc.TransitionV1{
				Type:   core.TransitionTypeTransferSend,
				Amount: "0",
			},
			HomeLedger: rpc.LedgerV1{
				BlockchainID: "137",
				TokenAddress: "0xToken",
				UserBalance:  "100",
				UserNetFlow:  "100",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetLatestStateMethod.String(), stateResp)

	// Mock deposit response
	depositResp := rpc.AppSessionsV1SubmitDepositStateResponse{
		StateNodeSig: "0xNodeSig",
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1SubmitDepositStateMethod.String(), depositResp)

	// Setup client with signer
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))

	rawSigner, err := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err)

	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err)

	stateSigner, err := core.NewChannelDefaultSigner(msgSigner)
	require.NoError(t, err)

	client := &Client{
		rpcClient:   rpc.NewClient(mockDialer),
		stateSigner: stateSigner,
		rawSigner:   rawSigner,
	}
	client.assetStore = newClientAssetStore(client)
	client.stateAdvancer = core.NewStateAdvancerV1(client.assetStore)

	appUpdate := app.AppStateUpdateV1{
		AppSessionID: "0xSessionID",
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      1,
		Allocations:  []app.AppAllocationV1{},
	}

	sig, err := client.SubmitAppSessionDeposit(context.Background(), appUpdate, []string{"sig1"}, "USDC", decimal.NewFromInt(10))
	require.NoError(t, err)
	assert.Equal(t, "0xNodeSig", sig)
}

func TestClient_SubmitAppState(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1SubmitAppStateResponse{}
	mockDialer.RegisterResponse(rpc.AppSessionsV1SubmitAppStateMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	appUpdate := app.AppStateUpdateV1{
		AppSessionID: "0xSessionID",
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations:  []app.AppAllocationV1{},
	}

	err := client.SubmitAppState(context.Background(), appUpdate, []string{"sig1"})
	require.NoError(t, err)
}

func TestClient_SubmitAppSessionKeyState(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1SubmitSessionKeyStateResponse{}
	mockDialer.RegisterResponse(rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	state := app.AppSessionKeyStateV1{
		UserAddress: "0xUser",
		SessionKey:  "0xKey",
		Version:     1,
		ExpiresAt:   time.Now().Add(time.Hour),
		UserSig:     "0xSig",
	}

	err := client.SubmitAppSessionKeyState(context.Background(), state)
	require.NoError(t, err)
}

func TestClient_GetLastAppKeyStates(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	now := time.Now().Unix()
	mockResp := rpc.AppSessionsV1GetLastKeyStatesResponse{
		States: []rpc.AppSessionKeyStateV1{
			{
				UserAddress: "0xUser",
				SessionKey:  "0xkey",
				Version:     "1",
				ExpiresAt:   decimal.NewFromInt(now).String(),
				UserSig:     "0xSig",
			},
		},
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1GetLastKeyStatesMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	states, err := client.GetLastAppKeyStates(context.Background(), "0xUser", nil)
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, "0xkey", states[0].SessionKey)
}

// TestDoCloseConcurrent verifies that calling doClose from multiple goroutines
// simultaneously does not panic. Before the sync.Once fix, the select+close
// pattern on exitCh could race: two goroutines could both see the channel as
// open and both call close(), causing a "close of closed channel" panic.
func TestDoCloseConcurrent(t *testing.T) {
	t.Parallel()
	client := &Client{
		exitCh: make(chan struct{}),
	}

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			client.doClose()
		}()
	}

	wg.Wait()

	select {
	case <-client.exitCh:
		// ok
	default:
		t.Fatal("exitCh should be closed after doClose")
	}
}

// TestCloseAndDoCloseConcurrent simulates the real race: Close() called by
// the application while doClose() is called by the error handler
func TestCloseAndDoCloseConcurrent(t *testing.T) {
	t.Parallel()
	client := &Client{
		exitCh: make(chan struct{}),
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			client.doClose()
		}()
	}

	wg.Wait()

	select {
	case <-client.exitCh:
		// ok
	default:
		t.Fatal("exitCh should be closed")
	}
}

func TestClient_SignSessionKeyState(t *testing.T) {
	t.Parallel()
	// Setup signer
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))

	rawSigner, err := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err)

	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err)

	stateSigner, err := core.NewChannelDefaultSigner(msgSigner)
	require.NoError(t, err)

	client := &Client{
		stateSigner: stateSigner,
		rawSigner:   rawSigner,
	}

	state := app.AppSessionKeyStateV1{
		UserAddress:    rawSigner.PublicKey().Address().String(),
		SessionKey:     "0xSessionKey",
		Version:        1,
		ApplicationIDs: []string{"app1"},
		ExpiresAt:      time.Now().Add(time.Hour),
	}

	sig, err := client.SignSessionKeyState(state)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)

	// Verify signature
	sigBytes, err := hexutil.Decode(sig)
	require.NoError(t, err)

	packed, err := app.PackAppSessionKeyStateV1(state)
	require.NoError(t, err)

	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	require.NoError(t, err)
	recoveredAddr, err := recoverer.RecoverAddress(packed, sigBytes)
	require.NoError(t, err)
	assert.Equal(t, rawSigner.PublicKey().Address().String(), recoveredAddr.String())
}

// newCrossChainTestClient builds a Client wired to a mockDialer pre-stocked
// with the responses needed to reach the Deposit/Withdraw cross-chain guard:
// node config (chain 137 home), assets (asset on both 137 and 8453), and a
// latest state representing an open channel on chain 137. Returns the client
// and the wallet address used to populate the state.
func newCrossChainTestClient(t *testing.T) (*Client, string) {
	t.Helper()

	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))

	rawSigner, err := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err)
	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err)
	stateSigner, err := core.NewChannelDefaultSigner(msgSigner)
	require.NoError(t, err)
	walletAddr := rawSigner.PublicKey().Address().String()

	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockDialer.RegisterResponse(rpc.NodeV1GetConfigMethod.String(), rpc.NodeV1GetConfigResponse{
		NodeAddress: "0xNodeAddress",
		Blockchains: []rpc.BlockchainInfoV1{
			{Name: "Polygon", BlockchainID: "137", ChannelHubAddress: "0xHubAddr137"},
			{Name: "Base", BlockchainID: "8453", ChannelHubAddress: "0xHubAddr8453"},
		},
	})

	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{
			{
				Name:                  "USDC",
				Symbol:                "USDC",
				Decimals:              6,
				SuggestedBlockchainID: "137",
				Tokens: []rpc.TokenV1{
					{BlockchainID: "137", Address: "0xToken137", Decimals: 6},
					{BlockchainID: "8453", Address: "0xToken8453", Decimals: 6},
				},
			},
		},
	})

	homeChannelID := "0xHomeChannel"
	mockDialer.RegisterResponse(rpc.ChannelsV1GetLatestStateMethod.String(), rpc.ChannelsV1GetLatestStateResponse{
		State: &rpc.StateV1{
			ID:            "0xStateID",
			Epoch:         "1",
			Version:       "1",
			UserWallet:    walletAddr,
			Asset:         "USDC",
			HomeChannelID: &homeChannelID,
			Transition: rpc.TransitionV1{
				Type:   core.TransitionTypeHomeDeposit,
				Amount: "10",
			},
			HomeLedger: rpc.LedgerV1{
				BlockchainID: "137",
				TokenAddress: "0xToken137",
				UserBalance:  "10",
				UserNetFlow:  "10",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
	})

	client := &Client{
		rpcClient:       rpc.NewClient(mockDialer),
		stateSigner:     stateSigner,
		rawSigner:       rawSigner,
		homeBlockchains: make(map[string]uint64),
	}
	client.assetStore = newClientAssetStore(client)
	client.stateAdvancer = core.NewStateAdvancerV1(client.assetStore)
	return client, walletAddr
}

func TestClient_Deposit_RejectsForeignChain(t *testing.T) {
	t.Parallel()
	client, _ := newCrossChainTestClient(t)

	_, err := client.Deposit(context.Background(), 8453, "USDC", decimal.NewFromFloat(0.5))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active home channel for asset \"USDC\" is on chain 137")
	assert.Contains(t, err.Error(), "cannot deposit on chain 8453")
}

func TestClient_Withdraw_RejectsForeignChain(t *testing.T) {
	t.Parallel()
	client, _ := newCrossChainTestClient(t)

	_, err := client.Withdraw(context.Background(), 8453, "USDC", decimal.NewFromFloat(0.5))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active home channel for asset \"USDC\" is on chain 137")
	assert.Contains(t, err.Error(), "cannot withdraw on chain 8453")
}
