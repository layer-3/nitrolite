package sdk

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

// ============================================================================
// NodeConfig and Blockchain Transformations
// ============================================================================

// transformNodeConfig converts an RPC NodeV1GetConfigResponse to SDK NodeConfig type.
func transformNodeConfig(resp rpc.NodeV1GetConfigResponse) (*core.NodeConfig, error) {
	blockchains := make([]core.Blockchain, 0, len(resp.Blockchains))
	for _, info := range resp.Blockchains {
		blockchainID, err := strconv.ParseUint(info.BlockchainID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse blockchain ID: %w", err)
		}

		blockchains = append(blockchains, core.Blockchain{
			Name:              info.Name,
			ID:                blockchainID,
			ChannelHubAddress: info.ChannelHubAddress,
			BlockStep:         0, // Not provided in RPC response
		})
	}

	return &core.NodeConfig{
		NodeAddress:            resp.NodeAddress,
		NodeVersion:            resp.NodeVersion,
		SupportedSigValidators: resp.SupportedSigValidators,
		Blockchains:            blockchains,
	}, nil
}

// ============================================================================
// Asset and Token Transformations
// ============================================================================

// transformAssets converts RPC AssetV1 slice to core.Asset slice.
func transformAssets(assets []rpc.AssetV1) ([]core.Asset, error) {
	result := make([]core.Asset, 0, len(assets))
	for _, asset := range assets {
		tokens := make([]core.Token, 0, len(asset.Tokens))
		for _, token := range asset.Tokens {
			blockchainID, err := strconv.ParseUint(token.BlockchainID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse blockchain ID: %w", err)
			}

			tokens = append(tokens, core.Token{
				Name:         token.Name,
				Symbol:       token.Symbol,
				Address:      token.Address,
				BlockchainID: blockchainID,
				Decimals:     token.Decimals,
			})
		}
		suggestedBlockchainID, err := strconv.ParseUint(asset.SuggestedBlockchainID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse suggested blockchain ID: %w", err)
		}

		result = append(result, core.Asset{
			Name:                  asset.Name,
			Symbol:                asset.Symbol,
			Decimals:              asset.Decimals,
			SuggestedBlockchainID: suggestedBlockchainID,
			Tokens:                tokens,
		})
	}
	return result, nil
}

// ============================================================================
// Balance Transformations
// ============================================================================

// transformBalances converts RPC BalanceEntryV1 slice to core.BalanceEntry slice.
func transformBalances(balances []rpc.BalanceEntryV1) ([]core.BalanceEntry, error) {
	result := make([]core.BalanceEntry, 0, len(balances))
	for _, balance := range balances {
		amount, err := decimal.NewFromString(balance.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to parse balance amount: %w", err)
		}
		result = append(result, core.BalanceEntry{
			Asset:   balance.Asset,
			Balance: amount,
		})
	}
	return result, nil
}

// ============================================================================
// Channel Transformations
// ============================================================================

// transformChannel converts a single RPC ChannelV1 to core.Channel.
func transformChannel(channel rpc.ChannelV1) (core.Channel, error) {
	// Parse channel type
	var channelType core.ChannelType
	switch channel.Type {
	case "home":
		channelType = core.ChannelTypeHome
	case "escrow":
		channelType = core.ChannelTypeEscrow
	}

	// Parse channel status
	var channelStatus core.ChannelStatus
	switch channel.Status {
	case "void":
		channelStatus = core.ChannelStatusVoid
	case "open":
		channelStatus = core.ChannelStatusOpen
	case "challenged":
		channelStatus = core.ChannelStatusChallenged
	case "closed":
		channelStatus = core.ChannelStatusClosed
	}

	// Parse state version (it's a string in RPC, convert to uint64)
	var stateVersion uint64
	if channel.StateVersion != "" {
		parsed, err := strconv.ParseUint(channel.StateVersion, 10, 64)
		if err == nil {
			stateVersion = parsed
		}
	}

	blockchainID, err := strconv.ParseUint(channel.BlockchainID, 10, 64)
	if err != nil {
		return core.Channel{}, fmt.Errorf("failed to parse blockchain ID: %w", err)
	}

	nonce, err := strconv.ParseUint(channel.Nonce, 10, 64)
	if err != nil {
		return core.Channel{}, fmt.Errorf("failed to parse nonce: %w", err)
	}

	return core.Channel{
		ChannelID:             channel.ChannelID,
		UserWallet:            channel.UserWallet,
		Asset:                 channel.Asset,
		Type:                  channelType,
		BlockchainID:          blockchainID,
		TokenAddress:          channel.TokenAddress,
		ChallengeDuration:     channel.ChallengeDuration,
		Nonce:                 nonce,
		ApprovedSigValidators: channel.ApprovedSigValidators,
		Status:                channelStatus,
		StateVersion:          stateVersion,
	}, nil
}

// ============================================================================
// Transaction Transformations
// ============================================================================

// transformTransactions converts RPC TransactionV1 slice to core.Transaction slice.
func transformTransactions(transactions []rpc.TransactionV1) ([]core.Transaction, error) {
	result := make([]core.Transaction, 0, len(transactions))
	for _, tx := range transactions {
		amount, err := decimal.NewFromString(tx.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to parse balance amount: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339, tx.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %w", err)
		}
		result = append(result, core.Transaction{
			ID:                 tx.ID,
			Asset:              tx.Asset,
			TxType:             tx.TxType,
			FromAccount:        tx.FromAccount,
			ToAccount:          tx.ToAccount,
			SenderNewStateID:   tx.SenderNewStateID,
			ReceiverNewStateID: tx.ReceiverNewStateID,
			Amount:             amount,
			CreatedAt:          createdAt,
		})
	}
	return result, nil
}

// ============================================================================
// Pagination Transformations
// ============================================================================

// transformPaginationMetadata converts RPC PaginationMetadataV1 to core.PaginationMetadata.
func transformPaginationMetadata(meta rpc.PaginationMetadataV1) core.PaginationMetadata {
	return core.PaginationMetadata{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}

// transformPaginationParams converts core.PaginationParams to RPC PaginationParamsV1.
func transformPaginationParams(params *core.PaginationParams) *rpc.PaginationParamsV1 {
	if params == nil {
		return nil
	}
	return &rpc.PaginationParamsV1{
		Offset: params.Offset,
		Limit:  params.Limit,
		Sort:   params.Sort,
	}
}

// ============================================================================
// State Management Transformations
// ============================================================================

// transformState converts RPC StateV1 to core.State.
func transformState(state rpc.StateV1) (core.State, error) {
	// Parse numeric strings
	epoch, err := strconv.ParseUint(state.Epoch, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse epoch: %w", err)
	}

	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse version: %w", err)
	}

	// Transform transition
	transitionAmount, err := decimal.NewFromString(state.Transition.Amount)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse transition amount: %w", err)
	}
	transition := core.Transition{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    transitionAmount,
	}

	// Transform ledgers
	homeLedger, err := transformLedger(state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to transform home ledger: %w", err)
	}

	var escrowLedger *core.Ledger
	if state.EscrowLedger != nil {
		el, err := transformLedger(*state.EscrowLedger)
		if err != nil {
			return core.State{}, fmt.Errorf("failed to transform escrow ledger: %w", err)
		}
		escrowLedger = &el
	}

	result := core.State{
		ID:              state.ID,
		Transition:      transition,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           epoch,
		Version:         version,
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
		// Note: IsFinal is computed from transitions, not stored
	}

	return result, nil
}

// transformStateToRPC converts core.State to RPC StateV1.
func transformStateToRPC(state core.State) rpc.StateV1 {
	// Transform transition
	transition := rpc.TransitionV1{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    state.Transition.Amount.String(),
	}

	// Transform ledgers
	homeLedger := transformLedgerToRPC(state.HomeLedger)

	var escrowLedger *rpc.LedgerV1
	if state.EscrowLedger != nil {
		el := transformLedgerToRPC(*state.EscrowLedger)
		escrowLedger = &el
	}

	result := rpc.StateV1{
		ID:              state.ID,
		Transition:      transition,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           strconv.FormatUint(state.Epoch, 10),
		Version:         strconv.FormatUint(state.Version, 10),
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
	}

	return result
}

// transformLedger converts RPC LedgerV1 to core.Ledger.
func transformLedger(ledger rpc.LedgerV1) (core.Ledger, error) {
	blockchainID, err := strconv.ParseUint(ledger.BlockchainID, 10, 64)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse blockchain ID: %w", err)
	}

	userBalance, err := decimal.NewFromString(ledger.UserBalance)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse user balance: %w", err)
	}

	userNetFlow, err := decimal.NewFromString(ledger.UserNetFlow)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse user net flow: %w", err)
	}

	nodeBalance, err := decimal.NewFromString(ledger.NodeBalance)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse node balance: %w", err)
	}

	nodeNetFlow, err := decimal.NewFromString(ledger.NodeNetFlow)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse node net flow: %w", err)
	}

	return core.Ledger{
		TokenAddress: ledger.TokenAddress,
		BlockchainID: blockchainID,
		UserBalance:  userBalance,
		UserNetFlow:  userNetFlow,
		NodeBalance:  nodeBalance,
		NodeNetFlow:  nodeNetFlow,
	}, nil
}

// transformLedgerToRPC converts core.Ledger to RPC LedgerV1.
func transformLedgerToRPC(ledger core.Ledger) rpc.LedgerV1 {
	return rpc.LedgerV1{
		TokenAddress: ledger.TokenAddress,
		BlockchainID: strconv.FormatUint(ledger.BlockchainID, 10),
		UserBalance:  ledger.UserBalance.String(),
		UserNetFlow:  ledger.UserNetFlow.String(),
		NodeBalance:  ledger.NodeBalance.String(),
		NodeNetFlow:  ledger.NodeNetFlow.String(),
	}
}

// transformChannelDefinitionToRPC converts core.ChannelDefinition to RPC ChannelDefinitionV1.
func transformChannelDefinitionToRPC(def core.ChannelDefinition) rpc.ChannelDefinitionV1 {
	return rpc.ChannelDefinitionV1{
		Nonce:                 strconv.FormatUint(def.Nonce, 10),
		Challenge:             def.Challenge,
		ApprovedSigValidators: def.ApprovedSigValidators,
	}
}

// ============================================================================
// App Session Transformations
// ============================================================================

// transformAppSessions converts RPC AppSessionInfoV1 slice to app.AppSessionInfoV1 slice.
func transformAppSessions(sessions []rpc.AppSessionInfoV1) ([]app.AppSessionInfoV1, error) {
	result := make([]app.AppSessionInfoV1, 0, len(sessions))
	for _, s := range sessions {
		appDef, err := transformAppDefinition(s.AppDefinitionV1)
		if err != nil {
			return nil, fmt.Errorf("failed to transform app definition: %w", err)
		}

		// Transform allocations
		allocations := make([]app.AppAllocationV1, 0, len(s.Allocations))
		for _, a := range s.Allocations {
			amount, err := decimal.NewFromString(a.Amount)
			if err != nil {
				return nil, fmt.Errorf("failed to parse allocation amount: %w", err)
			}

			allocations = append(allocations, app.AppAllocationV1{
				Participant: a.Participant,
				Asset:       a.Asset,
				Amount:      amount,
			})
		}

		// Parse status - RPC uses string, app uses IsClosed bool
		isClosed := (s.Status == "closed")

		// Handle session data - RPC uses *string, app uses string
		sessionData := ""
		if s.SessionData != nil {
			sessionData = *s.SessionData
		}

		version, err := strconv.ParseUint(s.Version, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version: %w", err)
		}

		result = append(result, app.AppSessionInfoV1{
			AppSessionID:  s.AppSessionID,
			AppDefinition: appDef,
			IsClosed:      isClosed,
			SessionData:   sessionData,
			Version:       version,
			Allocations:   allocations,
		})
	}
	return result, nil
}

// transformAppDefinition converts RPC AppDefinitionV1 to app.AppDefinitionV1.
func transformAppDefinition(def rpc.AppDefinitionV1) (app.AppDefinitionV1, error) {
	participants := make([]app.AppParticipantV1, 0, len(def.Participants))
	for _, p := range def.Participants {
		participants = append(participants, app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		})
	}

	nonce, err := strconv.ParseUint(def.Nonce, 10, 64)
	if err != nil {
		return app.AppDefinitionV1{}, fmt.Errorf("failed to parse nonce: %w", err)
	}

	return app.AppDefinitionV1{
		ApplicationID: def.Application,
		Participants:  participants,
		Quorum:        def.Quorum,
		Nonce:         nonce,
	}, nil
}

// transformAppDefinitionToRPC converts app.AppDefinitionV1 to RPC AppDefinitionV1.
func transformAppDefinitionToRPC(def app.AppDefinitionV1) rpc.AppDefinitionV1 {
	participants := make([]rpc.AppParticipantV1, 0, len(def.Participants))
	for _, p := range def.Participants {
		participants = append(participants, rpc.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		})
	}

	return rpc.AppDefinitionV1{
		Application:  def.ApplicationID,
		Participants: participants,
		Quorum:       def.Quorum,
		Nonce:        strconv.FormatUint(def.Nonce, 10),
	}
}

// transformAppStateUpdateToRPC converts app.AppStateUpdateV1 to RPC AppStateUpdateV1.
func transformAppStateUpdateToRPC(update app.AppStateUpdateV1) rpc.AppStateUpdateV1 {
	allocations := make([]rpc.AppAllocationV1, 0, len(update.Allocations))
	for _, a := range update.Allocations {
		allocations = append(allocations, rpc.AppAllocationV1{
			Participant: a.Participant,
			Asset:       a.Asset,
			Amount:      a.Amount.String(),
		})
	}

	return rpc.AppStateUpdateV1{
		AppSessionID: update.AppSessionID,
		Intent:       update.Intent,
		Version:      strconv.FormatUint(update.Version, 10),
		Allocations:  allocations,
		SessionData:  update.SessionData,
	}
}

// transformSignedAppStateUpdateToRPC converts app.SignedAppStateUpdateV1 to RPC SignedAppStateUpdateV1.
func transformSignedAppStateUpdateToRPC(signed app.SignedAppStateUpdateV1) rpc.SignedAppStateUpdateV1 {
	return rpc.SignedAppStateUpdateV1{
		AppStateUpdate: transformAppStateUpdateToRPC(signed.AppStateUpdate),
		QuorumSigs:     signed.QuorumSigs,
	}
}

// ============================================================================
// Channel Session Key State Transformations
// ============================================================================

// transformChannelSessionKeyStateToRPC converts core.ChannelSessionKeyStateV1 to RPC ChannelSessionKeyStateV1.
func transformChannelSessionKeyStateToRPC(state core.ChannelSessionKeyStateV1) rpc.ChannelSessionKeyStateV1 {
	return rpc.ChannelSessionKeyStateV1{
		UserAddress: state.UserAddress,
		SessionKey:  state.SessionKey,
		Version:     strconv.FormatUint(state.Version, 10),
		Assets:      state.Assets,
		ExpiresAt:   strconv.FormatInt(state.ExpiresAt.Unix(), 10),
		UserSig:     state.UserSig,
	}
}

// transformChannelSessionKeyState converts RPC ChannelSessionKeyStateV1 to core.ChannelSessionKeyStateV1.
func transformChannelSessionKeyState(state rpc.ChannelSessionKeyStateV1) (core.ChannelSessionKeyStateV1, error) {
	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.ChannelSessionKeyStateV1{}, fmt.Errorf("failed to parse version: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(state.ExpiresAt, 10, 64)
	if err != nil {
		return core.ChannelSessionKeyStateV1{}, fmt.Errorf("failed to parse expires_at: %w", err)
	}

	assets := state.Assets
	if assets == nil {
		assets = []string{}
	}

	return core.ChannelSessionKeyStateV1{
		UserAddress: strings.ToLower(state.UserAddress),
		SessionKey:  strings.ToLower(state.SessionKey),
		Version:     version,
		Assets:      assets,
		ExpiresAt:   time.Unix(expiresAtUnix, 0),
		UserSig:     state.UserSig,
	}, nil
}

// transformChannelSessionKeyStates converts a slice of RPC ChannelSessionKeyStateV1 to core.ChannelSessionKeyStateV1.
func transformChannelSessionKeyStates(states []rpc.ChannelSessionKeyStateV1) ([]core.ChannelSessionKeyStateV1, error) {
	result := make([]core.ChannelSessionKeyStateV1, 0, len(states))
	for _, s := range states {
		state, err := transformChannelSessionKeyState(s)
		if err != nil {
			return nil, fmt.Errorf("failed to transform channel session key state: %w", err)
		}
		result = append(result, state)
	}
	return result, nil
}

// ============================================================================
// App Session Key State Transformations
// ============================================================================

// transformSessionKeyStateToRPC converts app.AppSessionKeyStateV1 to RPC AppSessionKeyStateV1.
func transformSessionKeyStateToRPC(state app.AppSessionKeyStateV1) rpc.AppSessionKeyStateV1 {
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

// transformSessionKeyState converts RPC AppSessionKeyStateV1 to app.AppSessionKeyStateV1.
func transformSessionKeyState(state rpc.AppSessionKeyStateV1) (app.AppSessionKeyStateV1, error) {
	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return app.AppSessionKeyStateV1{}, fmt.Errorf("failed to parse version: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(state.ExpiresAt, 10, 64)
	if err != nil {
		return app.AppSessionKeyStateV1{}, fmt.Errorf("failed to parse expires_at: %w", err)
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

// transformSessionKeyStates converts a slice of RPC AppSessionKeyStateV1 to app.AppSessionKeyStateV1.
func transformSessionKeyStates(states []rpc.AppSessionKeyStateV1) ([]app.AppSessionKeyStateV1, error) {
	result := make([]app.AppSessionKeyStateV1, 0, len(states))
	for _, s := range states {
		state, err := transformSessionKeyState(s)
		if err != nil {
			return nil, fmt.Errorf("failed to transform session key state: %w", err)
		}
		result = append(result, state)
	}
	return result, nil
}
