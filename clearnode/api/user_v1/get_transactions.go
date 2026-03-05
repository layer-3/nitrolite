package user_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetTransactions retrieves transaction history for a user with optional filters.
func (h *Handler) GetTransactions(c *rpc.Context) {
	var req rpc.UserV1GetTransactionsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var paginationParams core.PaginationParams
	if req.Pagination != nil {
		paginationParams.Offset = req.Pagination.Offset
		paginationParams.Limit = req.Pagination.Limit
		paginationParams.Sort = req.Pagination.Sort
	}

	var transactions []core.Transaction
	var metadata core.PaginationMetadata

	transactions, metadata, err := h.store.GetUserTransactions(req.Wallet, req.Asset, req.TxType, req.FromTime, req.ToTime, &paginationParams)
	if err != nil {
		c.Fail(err, "failed to retrieve transactions")
		return
	}

	response := rpc.UserV1GetTransactionsResponse{
		Transactions: []rpc.TransactionV1{},
		Metadata:     *mapPaginationMetadataV1(metadata),
	}

	for _, tx := range transactions {
		response.Transactions = append(response.Transactions, mapTransactionV1(tx))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
