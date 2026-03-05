package channel_v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func toCoreState(state rpc.StateV1) (core.State, error) {
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

	homeLedger, err := toCoreLedger(&state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse home ledger: %w", err)
	}

	escrowLedger, err := toCoreLedger(state.EscrowLedger)
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

func toCoreLedger(ledger *rpc.LedgerV1) (*core.Ledger, error) {
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

// toCoreChannelDefinition converts RPC channel definition to core type.
func toCoreChannelDefinition(def rpc.ChannelDefinitionV1) (core.ChannelDefinition, error) {
	nonce, err := strconv.ParseUint(def.Nonce, 10, 64)
	if err != nil {
		return core.ChannelDefinition{}, fmt.Errorf("failed to parse nonce: %w", err)
	}

	return core.ChannelDefinition{
		Nonce:                 nonce,
		Challenge:             def.Challenge,
		ApprovedSigValidators: def.ApprovedSigValidators,
	}, nil
}

// channelTypeToString converts core.ChannelType to its string representation
func channelTypeToString(t core.ChannelType) string {
	switch t {
	case core.ChannelTypeHome:
		return "home"
	case core.ChannelTypeEscrow:
		return "escrow"
	default:
		return "unknown"
	}
}

// channelStatusToString converts core.ChannelStatus to its string representation
func channelStatusToString(s core.ChannelStatus) string {
	switch s {
	case core.ChannelStatusVoid:
		return "void"
	case core.ChannelStatusOpen:
		return "open"
	case core.ChannelStatusChallenged:
		return "challenged"
	case core.ChannelStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// coreChannelToRPC converts a core.Channel to rpc.ChannelV1
func coreChannelToRPC(channel core.Channel) rpc.ChannelV1 {
	return rpc.ChannelV1{
		ChannelID:             channel.ChannelID,
		UserWallet:            channel.UserWallet,
		Asset:                 channel.Asset,
		Type:                  channelTypeToString(channel.Type),
		BlockchainID:          strconv.FormatUint(channel.BlockchainID, 10),
		TokenAddress:          channel.TokenAddress,
		ChallengeDuration:     channel.ChallengeDuration,
		ChallengeExpiresAt:    channel.ChallengeExpiresAt,
		Nonce:                 strconv.FormatUint(channel.Nonce, 10),
		Status:                channelStatusToString(channel.Status),
		StateVersion:          strconv.FormatUint(channel.StateVersion, 10),
		ApprovedSigValidators: channel.ApprovedSigValidators,
	}
}

// coreStateToRPC converts a core.State to rpc.StateV1
func coreStateToRPC(state core.State) rpc.StateV1 {
	transition := rpc.TransitionV1{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    state.Transition.Amount.String(),
	}

	homeLedger := rpc.LedgerV1{
		TokenAddress: state.HomeLedger.TokenAddress,
		BlockchainID: strconv.FormatUint(state.HomeLedger.BlockchainID, 10),
		UserBalance:  state.HomeLedger.UserBalance.String(),
		UserNetFlow:  state.HomeLedger.UserNetFlow.String(),
		NodeBalance:  state.HomeLedger.NodeBalance.String(),
		NodeNetFlow:  state.HomeLedger.NodeNetFlow.String(),
	}

	var escrowLedger *rpc.LedgerV1
	if state.EscrowLedger != nil {
		escrowLedger = &rpc.LedgerV1{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: strconv.FormatUint(state.EscrowLedger.BlockchainID, 10),
			UserBalance:  state.EscrowLedger.UserBalance.String(),
			UserNetFlow:  state.EscrowLedger.UserNetFlow.String(),
			NodeBalance:  state.EscrowLedger.NodeBalance.String(),
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow.String(),
		}
	}

	return rpc.StateV1{
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
}

// unmapChannelSessionKeyStateV1 converts an RPC ChannelSessionKeyStateV1 to a core.ChannelSessionKeyStateV1.
func unmapChannelSessionKeyStateV1(state *rpc.ChannelSessionKeyStateV1) (core.ChannelSessionKeyStateV1, error) {
	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.ChannelSessionKeyStateV1{}, fmt.Errorf("invalid version: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(state.ExpiresAt, 10, 64)
	if err != nil {
		return core.ChannelSessionKeyStateV1{}, fmt.Errorf("invalid expires_at: %w", err)
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

// mapChannelSessionKeyStateV1 converts a core.ChannelSessionKeyStateV1 to an RPC ChannelSessionKeyStateV1.
func mapChannelSessionKeyStateV1(state *core.ChannelSessionKeyStateV1) rpc.ChannelSessionKeyStateV1 {
	return rpc.ChannelSessionKeyStateV1{
		UserAddress: state.UserAddress,
		SessionKey:  state.SessionKey,
		Version:     strconv.FormatUint(state.Version, 10),
		Assets:      state.Assets,
		ExpiresAt:   strconv.FormatInt(state.ExpiresAt.Unix(), 10),
		UserSig:     state.UserSig,
	}
}
