package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
)

// databaseChannelToCore converts database.Channel to core.Channel
func databaseChannelToCore(dbChannel *Channel) *core.Channel {
	return &core.Channel{
		ChannelID:             dbChannel.ChannelID,
		UserWallet:            dbChannel.UserWallet,
		Asset:                 dbChannel.Asset,
		Type:                  dbChannel.Type,
		BlockchainID:          dbChannel.BlockchainID,
		TokenAddress:          dbChannel.Token,
		ApprovedSigValidators: dbChannel.ApprovedSigValidators,
		ChallengeDuration:     dbChannel.ChallengeDuration,
		ChallengeExpiresAt:    dbChannel.ChallengeExpiresAt,
		Nonce:                 dbChannel.Nonce,
		Status:                dbChannel.Status,
		StateVersion:          dbChannel.StateVersion,
	}
}

// databaseAppSessionToCore converts database.AppSessionV1 to app.AppSessionV1
func databaseAppSessionToCore(dbSession *AppSessionV1) *app.AppSessionV1 {
	participants := make([]app.AppParticipantV1, len(dbSession.Participants))
	for i, p := range dbSession.Participants {
		participants[i] = app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	return &app.AppSessionV1{
		SessionID:    dbSession.ID,
		Application:  dbSession.ApplicationID,
		Participants: participants,
		Quorum:       dbSession.Quorum,
		Nonce:        dbSession.Nonce,
		Status:       dbSession.Status,
		Version:      dbSession.Version,
		SessionData:  dbSession.SessionData,
		CreatedAt:    dbSession.CreatedAt,
		UpdatedAt:    dbSession.UpdatedAt,
	}
}

// databaseStateToCore converts database.State to core.State
func databaseStateToCore(dbState *State) (*core.State, error) {
	// Build home ledger with blockchain ID and token address from joined channels
	homeLedger := core.Ledger{
		UserBalance: dbState.HomeUserBalance,
		UserNetFlow: dbState.HomeUserNetFlow,
		NodeBalance: dbState.HomeNodeBalance,
		NodeNetFlow: dbState.HomeNodeNetFlow,
	}

	// If home channel ID exists, blockchain ID and token address must be present
	if dbState.HomeChannelID != nil {
		if dbState.HomeBlockchainID == nil || dbState.HomeTokenAddress == nil {
			return nil, fmt.Errorf("home channel %s exists but blockchain ID or token address is missing", *dbState.HomeChannelID)
		}
		homeLedger.BlockchainID = *dbState.HomeBlockchainID
		homeLedger.TokenAddress = *dbState.HomeTokenAddress
	}

	transition := core.Transition{
		Type:      core.TransitionType(dbState.TransitionType),
		TxID:      dbState.TransitionTxID,
		AccountID: dbState.TransitionAccountID,
		Amount:    dbState.TransitionAmount,
	}

	state := &core.State{
		ID:              dbState.ID,
		Transition:      transition,
		Asset:           dbState.Asset,
		UserWallet:      dbState.UserWallet,
		Epoch:           dbState.Epoch,
		Version:         dbState.Version,
		HomeChannelID:   dbState.HomeChannelID,
		EscrowChannelID: dbState.EscrowChannelID,
		HomeLedger:      homeLedger,
	}

	// If escrow channel ID exists, blockchain ID and token address must be present
	if dbState.EscrowChannelID != nil {
		if dbState.EscrowBlockchainID == nil || dbState.EscrowTokenAddress == nil {
			return nil, fmt.Errorf("escrow channel %s exists but blockchain ID or token address is missing", *dbState.EscrowChannelID)
		}
		state.EscrowLedger = &core.Ledger{
			BlockchainID: *dbState.EscrowBlockchainID,
			TokenAddress: *dbState.EscrowTokenAddress,
			UserBalance:  dbState.EscrowUserBalance,
			UserNetFlow:  dbState.EscrowUserNetFlow,
			NodeBalance:  dbState.EscrowNodeBalance,
			NodeNetFlow:  dbState.EscrowNodeNetFlow,
		}
	}

	if dbState.UserSig != nil {
		state.UserSig = dbState.UserSig
	}
	if dbState.NodeSig != nil {
		state.NodeSig = dbState.NodeSig
	}

	return state, nil
}

// coreStateToDB converts core.State to database.State
func coreStateToDB(state *core.State) (*State, error) {
	dbState := &State{
		ID:                  strings.ToLower(state.ID),
		TransitionType:      uint8(state.Transition.Type),
		TransitionTxID:      strings.ToLower(state.Transition.TxID),
		TransitionAccountID: strings.ToLower(state.Transition.AccountID),
		TransitionAmount:    state.Transition.Amount,
		Asset:               state.Asset,
		UserWallet:          strings.ToLower(state.UserWallet),
		Epoch:               state.Epoch,
		Version:             state.Version,
		HomeUserBalance:     state.HomeLedger.UserBalance,
		HomeUserNetFlow:     state.HomeLedger.UserNetFlow,
		HomeNodeBalance:     state.HomeLedger.NodeBalance,
		HomeNodeNetFlow:     state.HomeLedger.NodeNetFlow,
		CreatedAt:           time.Now(),
	}
	if state.HomeChannelID != nil {
		lowered := strings.ToLower(*state.HomeChannelID)
		dbState.HomeChannelID = &lowered
	}
	if state.EscrowChannelID != nil {
		lowered := strings.ToLower(*state.EscrowChannelID)
		dbState.EscrowChannelID = &lowered
	}

	if state.EscrowLedger != nil {
		dbState.EscrowUserBalance = state.EscrowLedger.UserBalance
		dbState.EscrowUserNetFlow = state.EscrowLedger.UserNetFlow
		dbState.EscrowNodeBalance = state.EscrowLedger.NodeBalance
		dbState.EscrowNodeNetFlow = state.EscrowLedger.NodeNetFlow
	}

	if state.UserSig != nil {
		dbState.UserSig = state.UserSig
	}
	if state.NodeSig != nil {
		dbState.NodeSig = state.NodeSig
	}

	return dbState, nil
}

func toCoreTransaction(dbTx *Transaction) *core.Transaction {
	return &core.Transaction{
		ID:                 dbTx.ID,
		Asset:              dbTx.AssetSymbol,
		TxType:             dbTx.Type,
		FromAccount:        dbTx.FromAccount,
		ToAccount:          dbTx.ToAccount,
		SenderNewStateID:   dbTx.SenderNewStateID,
		ReceiverNewStateID: dbTx.ReceiverNewStateID,
		Amount:             dbTx.Amount,
		CreatedAt:          dbTx.CreatedAt,
	}
}

// calculatePaginationMetadata computes pagination metadata from total count, offset, and limit
func calculatePaginationMetadata(totalCount int64, offset, limit uint32) core.PaginationMetadata {
	pageCount := uint32((totalCount + int64(limit) - 1) / int64(limit))
	currentPage := uint32(1)
	if limit > 0 {
		currentPage = offset/limit + 1
	}

	return core.PaginationMetadata{
		Page:       currentPage,
		PerPage:    limit,
		TotalCount: uint32(totalCount),
		PageCount:  pageCount,
	}
}
