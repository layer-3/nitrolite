package app_session_v1

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

func unmapAppDefinitionV1(def rpc.AppDefinitionV1) (app.AppDefinitionV1, error) {
	participants := make([]app.AppParticipantV1, len(def.Participants))
	for i, p := range def.Participants {
		participants[i] = app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	nonce, err := strconv.ParseUint(def.Nonce, 10, 64)
	if err != nil {
		return app.AppDefinitionV1{}, fmt.Errorf("invalid nonce: %w", err)
	}

	return app.AppDefinitionV1{
		ApplicationID: def.Application,
		Participants:  participants,
		Quorum:        def.Quorum,
		Nonce:         nonce,
	}, nil
}

// unmapStateV1 converts an RPC StateV1 to a core.State.
func unmapStateV1(state rpc.StateV1) (core.State, error) {
	decimalTxAmount, err := decimal.NewFromString(state.Transition.Amount)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse amount: %w", err)
	}

	coreTransition := core.Transition{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    decimalTxAmount,
	}

	epoch, err := strconv.ParseUint(state.Epoch, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse epoch: %w", err)
	}

	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse version: %w", err)
	}

	homeLedger, err := unmapLedgerV1(&state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse home ledger: %w", err)
	}

	escrowLedger, err := unmapLedgerV1(state.EscrowLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse escrow ledger: %w", err)
	}

	return core.State{
		ID:              state.ID,
		Transition:      coreTransition,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           epoch,
		Version:         version,
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      *homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
	}, nil
}

func unmapLedgerV1(ledger *rpc.LedgerV1) (*core.Ledger, error) {
	if ledger == nil {
		return nil, nil
	}

	userBalance, err := decimal.NewFromString(ledger.UserBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user balance: %w", err)
	}

	userNetFlow, err := decimal.NewFromString(ledger.UserNetFlow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user net-flow: %w", err)
	}

	nodeBalance, err := decimal.NewFromString(ledger.NodeBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node balance: %w", err)
	}

	nodeNetFlow, err := decimal.NewFromString(ledger.NodeNetFlow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node net-flow: %w", err)
	}

	blockchainID, err := strconv.ParseUint(ledger.BlockchainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse blockchain ID: %w", err)
	}

	return &core.Ledger{
		BlockchainID: blockchainID,
		TokenAddress: ledger.TokenAddress,
		UserBalance:  userBalance,
		UserNetFlow:  userNetFlow,
		NodeBalance:  nodeBalance,
		NodeNetFlow:  nodeNetFlow,
	}, nil
}

func unmapAppStateUpdateV1(upd *rpc.AppStateUpdateV1) (app.AppStateUpdateV1, error) {
	allocations := make([]app.AppAllocationV1, len(upd.Allocations))
	for i, alloc := range upd.Allocations {
		decAmount, err := decimal.NewFromString(alloc.Amount)
		if err != nil {
			return app.AppStateUpdateV1{}, fmt.Errorf("failed to parse amount: %w", err)
		}

		allocations[i] = app.AppAllocationV1{
			Participant: strings.ToLower(alloc.Participant),
			Asset:       alloc.Asset,
			Amount:      decAmount,
		}
	}

	version, err := strconv.ParseUint(upd.Version, 10, 64)
	if err != nil {
		return app.AppStateUpdateV1{}, fmt.Errorf("failed to parse version: %w", err)
	}

	return app.AppStateUpdateV1{
		AppSessionID: upd.AppSessionID,
		Intent:       upd.Intent,
		Version:      version,
		Allocations:  allocations,
		SessionData:  upd.SessionData,
	}, nil
}

func unmapSignedAppStateUpdateV1(signedUpd *rpc.SignedAppStateUpdateV1) (app.SignedAppStateUpdateV1, error) {
	appStateUpd, err := unmapAppStateUpdateV1(&signedUpd.AppStateUpdate)
	if err != nil {
		return app.SignedAppStateUpdateV1{}, err
	}

	return app.SignedAppStateUpdateV1{
		AppStateUpdate: appStateUpd,
		QuorumSigs:     signedUpd.QuorumSigs,
	}, nil
}

// getParticipantWeights creates a map of participant wallet addresses to their weights.
func getParticipantWeights(participants []app.AppParticipantV1) map[string]uint8 {
	weights := make(map[string]uint8, len(participants))
	for _, p := range participants {
		weights[strings.ToLower(p.WalletAddress)] = p.SignatureWeight
	}
	return weights
}

func mapAppSessionInfoV1(session app.AppSessionV1, allocations map[string]map[string]decimal.Decimal) rpc.AppSessionInfoV1 {
	participants := make([]rpc.AppParticipantV1, len(session.Participants))
	for i, p := range session.Participants {
		participants[i] = rpc.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	var sessionData *string
	if session.SessionData != "" {
		sessionData = &session.SessionData
	}

	// Convert allocations map to RPC format
	rpcAllocations := []rpc.AppAllocationV1{}
	for participant, assetMap := range allocations {
		for asset, amount := range assetMap {
			rpcAllocations = append(rpcAllocations, rpc.AppAllocationV1{
				Participant: participant,
				Asset:       asset,
				Amount:      amount.String(),
			})
		}
	}
	slices.SortFunc(rpcAllocations, func(a, b rpc.AppAllocationV1) int {
		if a.Asset > b.Asset {
			return 1
		} else if a.Asset < b.Asset {
			return -1
		}

		if a.Participant > b.Participant {
			return 1
		} else if a.Participant < b.Participant {
			return -1
		}

		return 0
	})

	return rpc.AppSessionInfoV1{
		AppSessionID: session.SessionID,
		Status:       session.Status.String(),
		AppDefinitionV1: rpc.AppDefinitionV1{
			Application:  session.ApplicationID,
			Participants: participants,
			Quorum:       session.Quorum,
			Nonce:        strconv.FormatUint(session.Nonce, 10),
		},
		SessionData: sessionData,
		Version:     strconv.FormatUint(session.Version, 10),
		Allocations: rpcAllocations,
	}
}

func mapPaginationMetadataV1(meta core.PaginationMetadata) rpc.PaginationMetadataV1 {
	return rpc.PaginationMetadataV1{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}

// unmapSessionKeyStateV1 converts an RPC AppSessionKeyStateV1 to a core app.AppSessionKeyStateV1.
func unmapSessionKeyStateV1(state *rpc.AppSessionKeyStateV1) (app.AppSessionKeyStateV1, error) {
	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return app.AppSessionKeyStateV1{}, fmt.Errorf("invalid version: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(state.ExpiresAt, 10, 64)
	if err != nil {
		return app.AppSessionKeyStateV1{}, fmt.Errorf("invalid expires_at: %w", err)
	}

	applicationIDs := state.ApplicationIDs
	if applicationIDs == nil {
		applicationIDs = []string{}
	}

	appSessionIDs := state.AppSessionIDs
	if appSessionIDs == nil {
		appSessionIDs = []string{}
	}

	return app.AppSessionKeyStateV1{
		UserAddress:    strings.ToLower(state.UserAddress),
		SessionKey:     strings.ToLower(state.SessionKey),
		Version:        version,
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      time.Unix(expiresAtUnix, 0),
		UserSig:        state.UserSig,
	}, nil
}

// mapSessionKeyStateV1 converts a core app.AppSessionKeyStateV1 to an RPC AppSessionKeyStateV1.
func mapSessionKeyStateV1(state *app.AppSessionKeyStateV1) rpc.AppSessionKeyStateV1 {
	return rpc.AppSessionKeyStateV1{
		UserAddress:    state.UserAddress,
		SessionKey:     state.SessionKey,
		Version:        strconv.FormatUint(state.Version, 10),
		ApplicationIDs: state.ApplicationIDs,
		AppSessionIDs:  state.AppSessionIDs,
		ExpiresAt:      strconv.FormatInt(state.ExpiresAt.Unix(), 10),
		UserSig:        state.UserSig,
	}
}

// toRPCState converts a core.State to rpc.StateV1 for testing
func toRPCState(state core.State) rpc.StateV1 {
	transition := rpc.TransitionV1{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    state.Transition.Amount.String(),
	}

	rpcState := rpc.StateV1{
		ID:              state.ID,
		Transition:      transition,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           strconv.FormatUint(state.Epoch, 10),
		Version:         strconv.FormatUint(state.Version, 10),
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger: rpc.LedgerV1{
			TokenAddress: state.HomeLedger.TokenAddress,
			BlockchainID: strconv.FormatUint(state.HomeLedger.BlockchainID, 10),
			UserBalance:  state.HomeLedger.UserBalance.String(),
			UserNetFlow:  state.HomeLedger.UserNetFlow.String(),
			NodeBalance:  state.HomeLedger.NodeBalance.String(),
			NodeNetFlow:  state.HomeLedger.NodeNetFlow.String(),
		},
		UserSig: state.UserSig,
		NodeSig: state.NodeSig,
	}

	if state.EscrowLedger != nil {
		rpcState.EscrowLedger = &rpc.LedgerV1{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: strconv.FormatUint(state.EscrowLedger.BlockchainID, 10),
			UserBalance:  state.EscrowLedger.UserBalance.String(),
			UserNetFlow:  state.EscrowLedger.UserNetFlow.String(),
			NodeBalance:  state.EscrowLedger.NodeBalance.String(),
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow.String(),
		}
	}

	return rpcState
}
