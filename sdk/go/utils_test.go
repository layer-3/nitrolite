package sdk

import (
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformNodeConfig(t *testing.T) {
	rpcResp := rpc.NodeV1GetConfigResponse{
		NodeAddress:            "0xNodeAddress",
		NodeVersion:            "v1.0.0",
		SupportedSigValidators: []core.ChannelSignerType{core.ChannelSignerType_SessionKey},
		Blockchains: []rpc.BlockchainInfoV1{
			{
				Name:              "Polygon",
				BlockchainID:      "137",
				ChannelHubAddress: "0xHubAddress",
			},
		},
	}

	config, err := transformNodeConfig(rpcResp)
	require.NoError(t, err)
	assert.Equal(t, "0xNodeAddress", config.NodeAddress)
	assert.Equal(t, "v1.0.0", config.NodeVersion)
	assert.Len(t, config.SupportedSigValidators, 1)
	assert.Equal(t, core.ChannelSignerType_SessionKey, config.SupportedSigValidators[0])
	assert.Len(t, config.Blockchains, 1)
	assert.Equal(t, uint64(137), config.Blockchains[0].ID)
	assert.Equal(t, "Polygon", config.Blockchains[0].Name)

	// Test error case
	rpcResp.Blockchains[0].BlockchainID = "invalid"
	_, err = transformNodeConfig(rpcResp)
	assert.Error(t, err)
}

func TestTransformAssets(t *testing.T) {
	rpcAssets := []rpc.AssetV1{
		{
			Name:                  "USDC",
			Symbol:                "USDC",
			Decimals:              6,
			SuggestedBlockchainID: "137",
			Tokens: []rpc.TokenV1{
				{
					BlockchainID: "137",
					Address:      "0xTokenAddress",
					Name:         "USDC Token",
					Symbol:       "USDC",
					Decimals:     6,
				},
			},
		},
	}

	assets, err := transformAssets(rpcAssets)
	require.NoError(t, err)
	assert.Len(t, assets, 1)
	assert.Equal(t, "USDC", assets[0].Symbol)
	assert.Equal(t, uint64(137), assets[0].SuggestedBlockchainID)
	assert.Len(t, assets[0].Tokens, 1)
	assert.Equal(t, uint64(137), assets[0].Tokens[0].BlockchainID)

	// Test error case: invalid blockchain ID
	rpcAssets[0].Tokens[0].BlockchainID = "invalid"
	_, err = transformAssets(rpcAssets)
	assert.Error(t, err)

	// Test error case: invalid suggested blockchain ID
	rpcAssets[0].Tokens[0].BlockchainID = "137" // fix previous error
	rpcAssets[0].SuggestedBlockchainID = "invalid"
	_, err = transformAssets(rpcAssets)
	assert.Error(t, err)
}

func TestTransformBalances(t *testing.T) {
	rpcBalances := []rpc.BalanceEntryV1{
		{
			Asset:  "USDC",
			Amount: "100.50",
		},
	}

	balances, err := transformBalances(rpcBalances)
	require.NoError(t, err)
	assert.Len(t, balances, 1)
	assert.Equal(t, "USDC", balances[0].Asset)
	assert.Equal(t, "100.5", balances[0].Balance.String())

	// Test error case
	rpcBalances[0].Amount = "invalid"
	_, err = transformBalances(rpcBalances)
	assert.Error(t, err)
}

func TestTransformPaginationMetadata(t *testing.T) {
	rpcMeta := rpc.PaginationMetadataV1{
		Page:       1,
		PerPage:    10,
		TotalCount: 100,
		PageCount:  10,
	}

	meta := transformPaginationMetadata(rpcMeta)
	assert.Equal(t, uint32(1), meta.Page)
	assert.Equal(t, uint32(10), meta.PerPage)
	assert.Equal(t, uint32(100), meta.TotalCount)
}

func TestTransformPaginationParams(t *testing.T) {
	limit := uint32(20)
	offset := uint32(0)
	sort := "desc"

	params := &core.PaginationParams{
		Limit:  &limit,
		Offset: &offset,
		Sort:   &sort,
	}

	rpcParams := transformPaginationParams(params)
	assert.Equal(t, limit, *rpcParams.Limit)
	assert.Equal(t, offset, *rpcParams.Offset)
	assert.Equal(t, sort, *rpcParams.Sort)

	// Test nil input
	assert.Nil(t, transformPaginationParams(nil))
}

func TestTransformChannel(t *testing.T) {
	rpcChan := rpc.ChannelV1{
		ChannelID:             "0xChannelID",
		UserWallet:            "0xUserWallet",
		Type:                  "home",
		BlockchainID:          "137",
		TokenAddress:          "0xToken",
		ChallengeDuration:     100,
		Nonce:                 "1",
		ApprovedSigValidators: "0x01",
		Status:                "open",
		StateVersion:          "5",
	}

	ch, err := transformChannel(rpcChan)
	require.NoError(t, err)
	assert.Equal(t, core.ChannelTypeHome, ch.Type)
	assert.Equal(t, core.ChannelStatusOpen, ch.Status)
	assert.Equal(t, uint64(137), ch.BlockchainID)
	assert.Equal(t, uint64(5), ch.StateVersion)

	// Test error cases
	rpcChan.BlockchainID = "invalid"
	_, err = transformChannel(rpcChan)
	assert.Error(t, err)

	rpcChan.BlockchainID = "137"
	rpcChan.Nonce = "invalid"
	_, err = transformChannel(rpcChan)
	assert.Error(t, err)
}

func TestTransformTransactions(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	rpcTxs := []rpc.TransactionV1{
		{
			ID:        "0xTxID",
			Asset:     "USDC",
			Amount:    "50.0",
			CreatedAt: now.Format(time.RFC3339),
		},
	}

	txs, err := transformTransactions(rpcTxs)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "50", txs[0].Amount.String())
	assert.True(t, txs[0].CreatedAt.Equal(now) || txs[0].CreatedAt.Unix() == now.Unix())

	// Test error cases
	rpcTxs[0].Amount = "invalid"
	_, err = transformTransactions(rpcTxs)
	assert.Error(t, err)

	rpcTxs[0].Amount = "50.0"
	rpcTxs[0].CreatedAt = "invalid-time"
	_, err = transformTransactions(rpcTxs)
	assert.Error(t, err)
}

func TestTransformState(t *testing.T) {
	rpcState := rpc.StateV1{
		ID:         "0xStateID",
		Epoch:      "1",
		Version:    "2",
		UserWallet: "0xUser",
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
	}

	state, err := transformState(rpcState)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), state.Epoch)
	assert.Equal(t, uint64(2), state.Version)
	assert.Equal(t, "10", state.Transition.Amount.String())
	assert.Equal(t, uint64(137), state.HomeLedger.BlockchainID)

	// Reverse transform
	rpcStateBack := transformStateToRPC(state)
	assert.Equal(t, rpcState.ID, rpcStateBack.ID)
	assert.Equal(t, rpcState.Epoch, rpcStateBack.Epoch)

	// Test error cases
	badState := rpcState
	badState.Epoch = "invalid"
	_, err = transformState(badState)
	assert.Error(t, err)

	badState = rpcState
	badState.Transition.Amount = "invalid"
	_, err = transformState(badState)
	assert.Error(t, err)

	badState = rpcState
	badState.HomeLedger.BlockchainID = "invalid"
	_, err = transformState(badState)
	assert.Error(t, err)
}

func TestTransformAppSessions(t *testing.T) {
	sessionData := "data"
	rpcSessions := []rpc.AppSessionInfoV1{
		{
			AppSessionID: "0xSessionID",
			AppDefinitionV1: rpc.AppDefinitionV1{
				Participants: []rpc.AppParticipantV1{
					{WalletAddress: "0xA", SignatureWeight: 1},
				},
				Nonce: "1",
			},
			Allocations: []rpc.AppAllocationV1{
				{Participant: "0xA", Asset: "USDC", Amount: "10.0"},
			},
			Status:      "open",
			SessionData: &sessionData,
			Version:     "1",
		},
	}

	sessions, err := transformAppSessions(rpcSessions)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "0xSessionID", sessions[0].AppSessionID)
	assert.False(t, sessions[0].IsClosed)
	assert.Equal(t, "data", sessions[0].SessionData)
	assert.Equal(t, uint64(1), sessions[0].AppDefinition.Nonce)

	// Test error cases
	rpcSessions[0].Allocations[0].Amount = "invalid"
	_, err = transformAppSessions(rpcSessions)
	assert.Error(t, err)

	rpcSessions[0].Allocations[0].Amount = "10.0"
	rpcSessions[0].AppDefinitionV1.Nonce = "invalid"
	_, err = transformAppSessions(rpcSessions)
	assert.Error(t, err)
}

func TestTransformAppDefinition(t *testing.T) {
	rpcDef := rpc.AppDefinitionV1{
		Application: "0xApp",
		Participants: []rpc.AppParticipantV1{
			{WalletAddress: "0xA", SignatureWeight: 1},
		},
		Nonce:  "1",
		Quorum: 1,
	}

	def, err := transformAppDefinition(rpcDef)
	require.NoError(t, err)
	assert.Equal(t, "0xApp", def.ApplicationID)
	assert.Equal(t, uint64(1), def.Nonce)

	// Reverse
	rpcDefBack := transformAppDefinitionToRPC(def)
	assert.Equal(t, rpcDef.Application, rpcDefBack.Application)
	assert.Equal(t, rpcDef.Nonce, rpcDefBack.Nonce)

	// Error case
	rpcDef.Nonce = "invalid"
	_, err = transformAppDefinition(rpcDef)
	assert.Error(t, err)
}

func TestTransformAppStateUpdate(t *testing.T) {
	update := app.AppStateUpdateV1{
		AppSessionID: "0xSession",
		Version:      1,
		Allocations: []app.AppAllocationV1{
			{Participant: "0xA", Asset: "USDC", Amount: decimal.NewFromInt(10)},
		},
	}

	rpcUpdate := transformAppStateUpdateToRPC(update)
	assert.Equal(t, "1", rpcUpdate.Version)
	assert.Equal(t, "10", rpcUpdate.Allocations[0].Amount)

	signedUpdate := app.SignedAppStateUpdateV1{
		AppStateUpdate: update,
		QuorumSigs:     []string{"0xSig"},
	}

	rpcSigned := transformSignedAppStateUpdateToRPC(signedUpdate)
	assert.Equal(t, []string{"0xSig"}, rpcSigned.QuorumSigs)
}

func TestTransformChannelSessionKeyState(t *testing.T) {
	now := time.Now().Unix()
	rpcState := rpc.ChannelSessionKeyStateV1{
		UserAddress: "0xUser",
		SessionKey:  "0xKey",
		Version:     "1",
		Assets:      []string{"USDC"},
		ExpiresAt:   string(decimal.NewFromInt(now).String()),
		UserSig:     "0xSig",
	}

	state, err := transformChannelSessionKeyState(rpcState)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), state.Version)
	assert.Equal(t, now, state.ExpiresAt.Unix())
	// Note: user address and session key are lowercased in transformChannelSessionKeyState
	assert.Equal(t, "0xuser", state.UserAddress)

	// Reverse
	rpcStateBack := transformChannelSessionKeyStateToRPC(state)
	assert.Equal(t, rpcState.Version, rpcStateBack.Version)
	// rpcStateBack.UserAddress might be lowercased now

	// Error cases
	badState := rpcState
	badState.Version = "invalid"
	_, err = transformChannelSessionKeyState(badState)
	assert.Error(t, err)

	badState = rpcState
	badState.ExpiresAt = "invalid"
	_, err = transformChannelSessionKeyState(badState)
	assert.Error(t, err)
}

func TestTransformAppSessionKeyState(t *testing.T) {
	now := time.Now().Unix()
	rpcState := rpc.AppSessionKeyStateV1{
		UserAddress: "0xUser",
		SessionKey:  "0xKey",
		Version:     "1",
		ExpiresAt:   string(decimal.NewFromInt(now).String()),
	}

	state, err := transformSessionKeyState(rpcState)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), state.Version)
	assert.Equal(t, "0xuser", state.UserAddress) // lowercased

	// Reverse
	rpcStateBack := transformSessionKeyStateToRPC(state)
	assert.Equal(t, rpcState.Version, rpcStateBack.Version)

	// Error cases
	badState := rpcState
	badState.Version = "invalid"
	_, err = transformSessionKeyState(badState)
	assert.Error(t, err)
}
