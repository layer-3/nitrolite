package user_v1

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetTransactions_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	handler := &Handler{
		store: mockStore,
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	otherWallet := "0x9876543210987654321098765432109876543210"
	asset := "usdc"

	senderStateID := "state123"
	receiverStateID := "state456"

	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	transactions := []core.Transaction{
		{
			ID:                 "tx1",
			Asset:              asset,
			TxType:             core.TransactionTypeTransfer,
			FromAccount:        userWallet,
			ToAccount:          otherWallet,
			SenderNewStateID:   &senderStateID,
			ReceiverNewStateID: &receiverStateID,
			Amount:             decimal.NewFromFloat(100.50),
			CreatedAt:          createdAt,
		},
		{
			ID:          "tx2",
			Asset:       asset,
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: userWallet,
			ToAccount:   userWallet,
			Amount:      decimal.NewFromFloat(500.00),
			CreatedAt:   createdAt.Add(time.Hour),
		},
	}

	metadata := core.PaginationMetadata{
		Page:       1,
		PerPage:    10,
		TotalCount: 2,
		PageCount:  1,
	}

	mockStore.On("GetUserTransactions", userWallet, &asset, (*core.TransactionType)(nil), (*uint64)(nil), (*uint64)(nil), &core.PaginationParams{}).Return(transactions, metadata, nil)

	// Create RPC request
	reqPayload := rpc.UserV1GetTransactionsRequest{
		Wallet: userWallet,
		Asset:  &asset,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "user.v1.get_transactions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetTransactions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.UserV1GetTransactionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Transactions, 2)

	// Verify first transaction
	assert.Equal(t, "tx1", response.Transactions[0].ID)
	assert.Equal(t, asset, response.Transactions[0].Asset)
	assert.Equal(t, core.TransactionTypeTransfer, response.Transactions[0].TxType)
	assert.Equal(t, userWallet, response.Transactions[0].FromAccount)
	assert.Equal(t, otherWallet, response.Transactions[0].ToAccount)
	assert.Equal(t, &senderStateID, response.Transactions[0].SenderNewStateID)
	assert.Equal(t, &receiverStateID, response.Transactions[0].ReceiverNewStateID)
	assert.Equal(t, "100.5", response.Transactions[0].Amount)

	// Verify second transaction
	assert.Equal(t, "tx2", response.Transactions[1].ID)
	assert.Equal(t, core.TransactionTypeHomeDeposit, response.Transactions[1].TxType)
	assert.Equal(t, "500", response.Transactions[1].Amount)

	// Verify metadata
	assert.Equal(t, uint32(1), response.Metadata.Page)
	assert.Equal(t, uint32(10), response.Metadata.PerPage)
	assert.Equal(t, uint32(2), response.Metadata.TotalCount)
	assert.Equal(t, uint32(1), response.Metadata.PageCount)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}
