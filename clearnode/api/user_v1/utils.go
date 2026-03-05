package user_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func mapTransactionV1(tx core.Transaction) rpc.TransactionV1 {
	return rpc.TransactionV1{
		ID:                 tx.ID,
		Asset:              tx.Asset,
		TxType:             tx.TxType,
		FromAccount:        tx.FromAccount,
		ToAccount:          tx.ToAccount,
		SenderNewStateID:   tx.SenderNewStateID,
		ReceiverNewStateID: tx.ReceiverNewStateID,
		Amount:             tx.Amount.String(),
		CreatedAt:          tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapBalanceEntryV1(entry core.BalanceEntry) rpc.BalanceEntryV1 {
	return rpc.BalanceEntryV1{
		Asset:  entry.Asset,
		Amount: entry.Balance.String(),
	}
}

func mapPaginationMetadataV1(meta core.PaginationMetadata) *rpc.PaginationMetadataV1 {
	return &rpc.PaginationMetadataV1{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}
