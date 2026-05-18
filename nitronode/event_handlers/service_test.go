package event_handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/sign"
)

func TestHandleHomeChannelCreated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	event := &core.HomeChannelCreatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(1), "", "").Return(nil)

	// Execute
	err := service.HandleHomeChannelCreated(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelCheckpointed_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	expiryTime := time.Now().Add(time.Hour)

	channel := &core.Channel{
		ChannelID:          channelID,
		UserWallet:         userWallet,
		Asset:              "usdc",
		Type:               core.ChannelTypeHome,
		Status:             core.ChannelStatusChallenged,
		StateVersion:       3,
		ChallengeExpiresAt: &expiryTime,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), "", "").Return(nil)
	// Off-chain head missing → no head-sig backfill.
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(nil, nil)

	// Execute
	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelChallenged_PersistsChallenge(t *testing.T) {
	// Channel must be marked Challenged with the challenge expiry so CheckActiveChannel and
	// RefreshUserEnforcedBalance stop treating it as open. Auto-checkpoint stays disabled:
	// non-checkpointable intents (CLOSE, escrow initiate/finalize, migration) cannot be
	// resolved via ScheduleCheckpoint, so operator action is required.
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    4,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 4 &&
			ch.ChallengeExpiresAt != nil &&
			ch.ChallengeExpiresAt.Unix() == int64(challengeExpiry)
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(4), "", "").Return(nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "GetLastStateByChannelID", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleCheckpoint", mock.Anything, mock.Anything)
}

func TestHandleHomeChannelChallenged_StaleVersionIgnored(t *testing.T) {
	// Per protocol the challenged version cannot be lower than the last known on-chain version.
	// Anomalies (replay, indexer mis-order) must not regress channel state.
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 5,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    3,
		ChallengeExpiry: uint64(time.Now().Add(time.Hour).Unix()),
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "UpdateChannel", mock.Anything)
	mockStore.AssertNotCalled(t, "RefreshUserEnforcedBalance", mock.Anything, mock.Anything)
}

func TestHandleHomeChannelChallenged_ChannelNotFound(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xMissingChannel"

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    1,
		ChallengeExpiry: uint64(time.Now().Add(time.Hour).Unix()),
	}

	mockStore.On("GetChannelByID", channelID).Return(nil, nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelChallenged_TypeMismatch(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xEscrowAsHome"

	channel := &core.Channel{
		ChannelID: channelID,
		Type:      core.ChannelTypeEscrow,
		Status:    core.ChannelStatusOpen,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    1,
		ChallengeExpiry: uint64(time.Now().Add(time.Hour).Unix()),
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelChallenged_FromClosingState(t *testing.T) {
	// Closing → Challenged is an expected transition: a co-signed Finalize may race an
	// on-chain challenge. The chain takes precedence; status must become Challenged.
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusClosing,
		StateVersion: 3,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    4,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 4 &&
			ch.ChallengeExpiresAt != nil &&
			ch.ChallengeExpiresAt.Unix() == int64(challengeExpiry)
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(4), "", "").Return(nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelChallenged_AcquiresUserLockBeforeMutation pins the race fix:
// the handler must call LockUserState(userWallet, asset) before UpdateChannel so an
// in-flight receiver-issuance RPC cannot read Status=Open, node-sign a receiver state
// and commit after the status flip to Challenged. See HandleHomeChannelClosed for the
// same pattern.
func TestHandleHomeChannelChallenged_AcquiresUserLockBeforeMutation(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "usdc"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    4,
		ChallengeExpiry: uint64(time.Now().Add(time.Hour).Unix()),
	}

	var locked bool
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).
		Run(func(mock.Arguments) { locked = true }).
		Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(core.Channel) bool {
		require.True(t, locked, "LockUserState must be called before UpdateChannel")
		return true
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(4), "", "").Return(nil)

	err := service.HandleHomeChannelChallenged(ctx, mockStore, event)
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelClosed_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 5,
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: 10,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 10
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(10), "", "").Return(nil)

	// Execute
	err := service.HandleHomeChannelClosed(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositInitiated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	state := &core.State{
		ID:      "state123",
		Version: 1,
		HomeLedger: core.Ledger{
			BlockchainID: 0,
		},
		EscrowLedger: &core.Ledger{
			BlockchainID: 2,
		},
	}

	event := &core.EscrowDepositInitiatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)
	mockStore.On("GetStateByChannelIDAndVersion", channelID, uint64(1)).Return(state, nil)
	mockStore.On("ScheduleInitiateEscrowDeposit", "state123", uint64(0)).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(1), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowDepositInitiated(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositChallenged_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	state := &core.State{
		ID:      "state123",
		Version: 5,
		EscrowLedger: &core.Ledger{
			BlockchainID: 2,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 3 &&
			ch.ChallengeExpiresAt != nil
	})).Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, true).Return(state, nil)
	mockStore.On("ScheduleFinalizeEscrowDeposit", "state123", uint64(2)).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(3), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositChallenged_NoFinalize_SchedulesHomeChallenge(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	homeChannelID := "0xHomeChannel456"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	homeChannel := &core.Channel{
		ChannelID:    homeChannelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		BlockchainID: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
		},
		EscrowLedger: &core.Ledger{
			BlockchainID: 2,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == escrowChannelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 3
	})).Return(nil)
	// No newer signed FINALIZE state available locally — node only has the INITIATE state.
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("GetChannelByID", homeChannelID).Return(homeChannel, nil)
	mockStore.On("ScheduleChallenge", "initiate-state-id", uint64(1)).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_NoFinalize_HomeChannelNotOpen_SkipsChallenge(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	homeChannelID := "0xHomeChannel456"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	homeChannel := &core.Channel{
		ChannelID:    homeChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		BlockchainID: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("GetChannelByID", homeChannelID).Return(homeChannel, nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_NoLocalState_NoSchedule(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(nil, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(nil, nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_HomeChannelIDNil_SkipsChallenge(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: nil,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_HomeChannelNotFound_SkipsChallenge(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	homeChannelID := "0xHomeChannel456"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("GetChannelByID", homeChannelID).Return((*core.Channel)(nil), nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_GetStateByVersionError_Propagates(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	dbErr := errors.New("db boom")

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(nil, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(nil, dbErr)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.ErrorIs(t, err, dbErr)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "UpdateStateSigsIfMissing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_GetHomeChannelError_Propagates(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	homeChannelID := "0xHomeChannel456"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 1,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	dbErr := errors.New("db boom")

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("GetChannelByID", homeChannelID).Return((*core.Channel)(nil), dbErr)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.ErrorIs(t, err, dbErr)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "UpdateStateSigsIfMissing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositChallenged_HomeBlockchainIDZero_SkipsChallenge(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	escrowChannelID := "0xEscrowChannel123"
	homeChannelID := "0xHomeChannel456"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	escrowChannel := &core.Channel{
		ChannelID:    escrowChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	homeChannel := &core.Channel{
		ChannelID:    homeChannelID,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		BlockchainID: 1,
	}

	initiateState := &core.State{
		ID:            "initiate-state-id",
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			BlockchainID: 0,
		},
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       escrowChannelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	mockStore.On("GetChannelByID", escrowChannelID).Return(escrowChannel, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("GetLastStateByChannelID", escrowChannelID, true).Return(initiateState, nil)
	mockStore.On("GetStateByChannelIDAndVersion", escrowChannelID, uint64(3)).Return(initiateState, nil)
	mockStore.On("GetChannelByID", homeChannelID).Return(homeChannel, nil)
	mockStore.On("UpdateStateSigsIfMissing", escrowChannelID, uint64(3), "", "").Return(nil)

	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "ScheduleChallenge", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "ScheduleFinalizeEscrowDeposit", mock.Anything, mock.Anything)
}

func TestHandleEscrowDepositFinalized_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.EscrowDepositFinalizedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowDepositFinalized(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalInitiated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	event := &core.EscrowWithdrawalInitiatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(1), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalInitiated(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalChallenged_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	state := &core.State{
		ID:      "state123",
		Version: 5,
		EscrowLedger: &core.Ledger{
			BlockchainID: 2,
		},
	}

	event := &core.EscrowWithdrawalChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 3 &&
			ch.ChallengeExpiresAt != nil
	})).Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, true).Return(state, nil)
	mockStore.On("ScheduleFinalizeEscrowWithdrawal", "state123", uint64(2)).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(3), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalChallenged(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalFinalized_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.EscrowWithdrawalFinalizedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), "", "").Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalFinalized(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleUserLockedBalanceUpdated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	blockchainID := uint64(1)
	balance := decimal.NewFromInt(1000)

	event := &core.UserLockedBalanceUpdatedEvent{
		UserAddress:  userWallet,
		BlockchainID: blockchainID,
		Balance:      balance,
	}

	// Mock expectations
	mockStore.On("UpdateUserStaked", userWallet, blockchainID, balance).Return(nil)

	// Execute
	err := service.HandleUserLockedBalanceUpdated(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelCheckpointed_BackfillsUserSig covers the recovery path for the wedge
// scenario: a node-only state was checkpointed on chain (e.g. the receiver of a transfer signed
// the receiver state and submitted it directly). The reactor extracts the user signature from the
// event and the handler must forward it to the store so the local row matches what is enforced
// on chain. Without this, EnsureNoOngoingStateTransitions stays blocked on the now-stale prior
// bilateral state and the channel can only be unblocked via on-chain challenge.
func TestHandleHomeChannelCheckpointed_BackfillsUserSig(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	userSig := "0xabcdef0123456789"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 4,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
		UserSig:      userSig,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), userSig, "").Return(nil)

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "GetLastStateByChannelID", mock.Anything, mock.Anything)
}

// TestHandleHomeChannelCheckpointed_BackfillError surfaces store errors from the backfill so
// the surrounding event-processing transaction rolls back and the event can be retried.
func TestHandleHomeChannelCheckpointed_BackfillError(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 4,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
		UserSig:      "0xdeadbeef",
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "usdc").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), "0xdeadbeef", "").Return(errors.New("db error"))

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)

	require.Error(t, err)
	require.Contains(t, err.Error(), "db error")
	mockStore.AssertExpectations(t)
}

// mockEventHandlerAssetStore implements core.AssetStore for state packing in unit tests.
type mockEventHandlerAssetStore struct{}

func (mockEventHandlerAssetStore) GetAssetDecimals(string) (uint8, error) { return 6, nil }
func (mockEventHandlerAssetStore) GetTokenDecimals(uint64, string) (uint8, error) {
	return 6, nil
}

func newTestEventHandlerService(t *testing.T) (*EventHandlerService, string) {
	t.Helper()
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer, err := sign.NewEthereumMsgSigner(hexutil.Encode(crypto.FromECDSA(key)))
	require.NoError(t, err)
	nodeSigner, err := core.NewChannelDefaultSigner(signer)
	require.NoError(t, err)
	packer := core.NewStatePackerV1(mockEventHandlerAssetStore{})
	return NewEventHandlerService(nodeSigner, packer), signer.PublicKey().Address().String()
}

// TestHandleHomeChannelCheckpointed_BackfillsHeadNodeSig covers the case where a
// challenge is cleared while the off-chain head sits above the checkpointed onchain
// version: a receiver state stored unsigned during the challenge window is now the
// channel's actual latest state. The handler must node-sign that head so future flows
// treat it as fully co-signed; the user signature backfill targets the on-chain version
// separately.
func TestHandleHomeChannelCheckpointed_BackfillsHeadNodeSig(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, nodeAddress := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelIDPtr := channelID
	checkpointVersion := uint64(5)
	headVersion := uint64(7)
	checkpointUserSig := "0xusersighex"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 4,
	}

	// Off-chain head is a during-challenge receiver state above the on-chain checkpoint.
	headState := &core.State{
		ID:            core.GetStateID(userWallet, asset, 1, headVersion),
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       headVersion,
		HomeChannelID: &homeChannelIDPtr,
		Transition:    core.Transition{Type: core.TransitionTypeTransferReceive},
		HomeLedger: core.Ledger{
			TokenAddress: "0xtoken",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.Zero,
			NodeBalance:  decimal.Zero,
			NodeNetFlow:  decimal.Zero,
		},
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: checkpointVersion,
		UserSig:      checkpointUserSig,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.Status == core.ChannelStatusOpen && ch.StateVersion == checkpointVersion
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	// User-sig backfill at the on-chain version.
	mockStore.On("UpdateStateSigsIfMissing", channelID, checkpointVersion, checkpointUserSig, "").Return(nil)
	// Off-chain head lookup returns the higher unsigned receiver state.
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(headState, nil)

	var capturedNodeSig string
	mockStore.On("UpdateStateSigsIfMissing", channelID, headVersion, "", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			capturedNodeSig = args.String(3)
		}).Return(nil)

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)
	require.NoError(t, err)
	require.NotEmpty(t, capturedNodeSig, "node signature must be populated on backfill")

	// Verify the produced signature is from the configured node key.
	packer := core.NewStatePackerV1(mockEventHandlerAssetStore{})
	packed, err := packer.PackState(*headState)
	require.NoError(t, err)
	sigBytes, err := hexutil.Decode(capturedNodeSig)
	require.NoError(t, err)
	validator := core.NewChannelSigValidator(nil)
	require.NoError(t, validator.Verify(nodeAddress, packed, sigBytes))

	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelCheckpointed_HeadAlreadySigned_NoBackfill verifies that when
// a challenge clears and the off-chain head is already node-signed (typical case if
// no receiver states were issued during the challenge), the handler does not re-sign.
func TestHandleHomeChannelCheckpointed_HeadAlreadySigned_NoBackfill(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelIDPtr := channelID
	checkpointVersion := uint64(5)
	headVersion := uint64(5)
	existingNodeSig := "0xnodesigalreadyhere"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 4,
	}

	headState := &core.State{
		ID:            core.GetStateID(userWallet, asset, 1, headVersion),
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       headVersion,
		HomeChannelID: &homeChannelIDPtr,
		Transition:    core.Transition{Type: core.TransitionTypeTransferReceive},
		HomeLedger: core.Ledger{
			TokenAddress: "0xtoken",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.Zero,
			NodeBalance:  decimal.Zero,
			NodeNetFlow:  decimal.Zero,
		},
		NodeSig: &existingNodeSig,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: checkpointVersion,
		UserSig:      "0xusersig",
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, checkpointVersion, "0xusersig", "").Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(headState, nil)

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)
	require.NoError(t, err)

	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "UpdateStateSigsIfMissing", channelID, headVersion, "", mock.AnythingOfType("string"))
}

// TestHandleHomeChannelCheckpointed_AcquiresUserLockBeforeMutation pins the race fix:
// the handler must call LockUserState before flipping Status from Challenged to Open
// and backfilling the off-chain head. Otherwise an in-flight receiver-issuance RPC
// can read Status=Challenged, choose to store an unsigned receiver row, and commit
// after the backfill — leaving the latest head unsigned on a now-Open channel.
func TestHandleHomeChannelCheckpointed_AcquiresUserLockBeforeMutation(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "usdc"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 3,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	var locked bool
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).
		Run(func(mock.Arguments) { locked = true }).
		Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(core.Channel) bool {
		require.True(t, locked, "LockUserState must be called before UpdateChannel")
		return true
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, uint64(5), "", "").Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(nil, nil)

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelClosed_ChallengeRescue_Squash exercises the path where a channel
// is closed onchain while still Challenged: any unsigned receiver-state credits accrued
// during the challenge window are squashed into a single ChallengeRescue state on the
// user's ledger.
func TestHandleHomeChannelClosed_ChallengeRescue_Squash(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	tokenAddress := "0xtoken"
	blockchainID := uint64(1)
	closureVersion := uint64(7)
	rescueAmount := decimal.NewFromInt(150)
	homeChannelIDPtr := channelID

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 5,
	}

	// Latest stored state for the channel — represents the highest-version unsigned
	// receiver state recorded during the challenge window.
	prevState := &core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 9),
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       9,
		HomeChannelID: &homeChannelIDPtr,
		HomeLedger: core.Ledger{
			TokenAddress: tokenAddress,
			BlockchainID: blockchainID,
			UserBalance:  decimal.NewFromInt(50),
			UserNetFlow:  decimal.NewFromInt(50),
			NodeBalance:  decimal.Zero,
			NodeNetFlow:  decimal.Zero,
		},
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: closureVersion,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.Status == core.ChannelStatusClosed && ch.StateVersion == closureVersion
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, closureVersion, "", "").Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(prevState, nil)
	mockStore.On("SumNetTransitionAmountAfterVersion", channelID, closureVersion, prevState.Epoch).Return(rescueAmount, nil)

	var capturedState core.State
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		capturedState = state
		return true
	}), "").Return(nil)
	var capturedTx core.Transaction
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		capturedTx = tx
		return true
	}), "").Return(nil)

	err := service.HandleHomeChannelClosed(ctx, mockStore, event)
	require.NoError(t, err)

	require.Equal(t, core.TransitionTypeChallengeRescue, capturedState.Transition.Type)
	require.Equal(t, channelID, capturedState.Transition.AccountID)
	require.True(t, rescueAmount.Equal(capturedState.Transition.Amount))
	require.Nil(t, capturedState.HomeChannelID, "rescue state must be off-channel")
	require.Equal(t, prevState.Epoch+1, capturedState.Epoch)
	require.Equal(t, uint64(1), capturedState.Version)
	require.Nil(t, capturedState.NodeSig, "rescue state is stored unsigned, like a credit to a user with no open home channel")
	require.True(t, rescueAmount.Equal(capturedState.HomeLedger.UserBalance))
	require.Empty(t, capturedState.HomeLedger.TokenAddress)
	require.Equal(t, uint64(0), capturedState.HomeLedger.BlockchainID)

	require.Equal(t, core.TransactionTypeChallengeRescue, capturedTx.TxType)
	require.Equal(t, channelID, capturedTx.FromAccount)
	require.Equal(t, userWallet, capturedTx.ToAccount)
	require.True(t, rescueAmount.Equal(capturedTx.Amount))

	// TxID is derived deterministically from (fromAccount, newStateID).
	expectedTxID, err := core.GetReceiverTransactionID(channelID, capturedState.ID)
	require.NoError(t, err)
	require.Equal(t, expectedTxID, capturedTx.ID)

	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelClosed_ChallengeRescue_NoCredits covers the path where a channel
// is closed while Challenged but no unsigned receiver credits accrued. A zero-amount
// rescue state is still emitted so the user's latest stored state moves to a fresh
// epoch with HomeChannelID nil; without it, future receiver-state issuance and
// channels.v1.request_creation would stay wedged on the closed channel.
func TestHandleHomeChannelClosed_ChallengeRescue_NoCredits(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	tokenAddress := "0xtoken"
	blockchainID := uint64(1)
	closureVersion := uint64(7)
	homeChannelIDPtr := channelID

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 5,
	}

	prevState := &core.State{
		ID:            core.GetStateID(userWallet, asset, 1, closureVersion),
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       closureVersion,
		HomeChannelID: &homeChannelIDPtr,
		HomeLedger: core.Ledger{
			TokenAddress: tokenAddress,
			BlockchainID: blockchainID,
			UserBalance:  decimal.NewFromInt(50),
			UserNetFlow:  decimal.NewFromInt(50),
		},
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: closureVersion,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, closureVersion, "", "").Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(prevState, nil)
	mockStore.On("SumNetTransitionAmountAfterVersion", channelID, closureVersion, prevState.Epoch).Return(decimal.Zero, nil)

	var capturedState core.State
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		capturedState = state
		return true
	}), "").Return(nil)
	mockStore.On("RecordTransaction", mock.Anything, "").Return(nil)

	err := service.HandleHomeChannelClosed(ctx, mockStore, event)
	require.NoError(t, err)

	require.Equal(t, core.TransitionTypeChallengeRescue, capturedState.Transition.Type)
	require.True(t, capturedState.Transition.Amount.IsZero())
	require.True(t, capturedState.HomeLedger.UserBalance.IsZero())
	require.Nil(t, capturedState.HomeChannelID)
	require.Equal(t, prevState.Epoch+1, capturedState.Epoch)
	require.Equal(t, uint64(1), capturedState.Version)

	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelClosed_ChallengeRescue_NegativeNet_ClampsToZero pins the
// adversarial-rollback case: the user closes at a version where her own channel
// balance was higher than the off-chain head, so the net transition amount above
// closure is negative. The rescue must clamp the credit at zero — onchain has
// already paid above the head value, and docking the user further is not the
// rescue's job. A zero-amount rescue is still issued so the user state head
// advances to a fresh epoch with HomeChannelID nil.
func TestHandleHomeChannelClosed_ChallengeRescue_NegativeNet_ClampsToZero(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	tokenAddress := "0xtoken"
	blockchainID := uint64(1)
	closureVersion := uint64(2)
	homeChannelIDPtr := channelID

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        asset,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusChallenged,
		StateVersion: 1,
	}

	prevState := &core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 5),
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       5,
		HomeChannelID: &homeChannelIDPtr,
		HomeLedger: core.Ledger{
			TokenAddress: tokenAddress,
			BlockchainID: blockchainID,
			UserBalance:  decimal.NewFromInt(51),
			UserNetFlow:  decimal.NewFromInt(51),
		},
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: closureVersion,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, asset).Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, closureVersion, "", "").Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, false).Return(prevState, nil)
	mockStore.On("SumNetTransitionAmountAfterVersion", channelID, closureVersion, prevState.Epoch).Return(decimal.NewFromInt(-49), nil)

	var capturedState core.State
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		capturedState = state
		return true
	}), "").Return(nil)
	mockStore.On("RecordTransaction", mock.Anything, "").Return(nil)

	err := service.HandleHomeChannelClosed(ctx, mockStore, event)
	require.NoError(t, err)

	require.Equal(t, core.TransitionTypeChallengeRescue, capturedState.Transition.Type)
	require.True(t, capturedState.Transition.Amount.IsZero(), "negative net must clamp to zero, got %s", capturedState.Transition.Amount.String())
	require.True(t, capturedState.HomeLedger.UserBalance.IsZero())
	require.Nil(t, capturedState.HomeChannelID)
	require.Equal(t, prevState.Epoch+1, capturedState.Epoch)
	require.Equal(t, uint64(1), capturedState.Version)

	mockStore.AssertExpectations(t)
}

// TestHandleHomeChannelClosed_OpenChannel_NoRescue pins behavior for the normal
// Open → Closed path: the channel was never Challenged before the close event, so the
// rescue branch is short-circuited entirely. The unsigned-receiver-sum query, rescue
// state store, and rescue transaction record are never issued.
func TestHandleHomeChannelClosed_OpenChannel_NoRescue(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service, _ := newTestEventHandlerService(t)

	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	closureVersion := uint64(7)

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Asset:        "USDC",
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 5,
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: closureVersion,
	}

	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("LockUserState", userWallet, "USDC").Return(decimal.Zero, nil)
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "USDC").Return(nil)
	mockStore.On("UpdateStateSigsIfMissing", channelID, closureVersion, "", "").Return(nil)

	err := service.HandleHomeChannelClosed(ctx, mockStore, event)
	require.NoError(t, err)

	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "SumNetTransitionAmountAfterVersion", mock.Anything, mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "StoreUserState", mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything)
}

func TestHandleUserLockedBalanceUpdated_StoreError(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())

	service := &EventHandlerService{}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	blockchainID := uint64(1)
	balance := decimal.NewFromInt(500)

	event := &core.UserLockedBalanceUpdatedEvent{
		UserAddress:  userWallet,
		BlockchainID: blockchainID,
		Balance:      balance,
	}

	// Mock expectations
	mockStore.On("UpdateUserStaked", userWallet, blockchainID, balance).Return(errors.New("db error"))

	// Execute
	err := service.HandleUserLockedBalanceUpdated(ctx, mockStore, event)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "db error")
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositsPurged_ClosesEscrowChannels(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	openEscrow := &core.Channel{
		ChannelID: "0xEscrow001",
		Type:      core.ChannelTypeEscrow,
		Status:    core.ChannelStatusOpen,
	}
	// Already closed — UpdateChannel must NOT be called for it.
	closedEscrow := &core.Channel{
		ChannelID: "0xEscrow002",
		Type:      core.ChannelTypeEscrow,
		Status:    core.ChannelStatusClosed,
	}

	event := &core.EscrowDepositsPurgedEvent{
		EscrowIDs: []string{"0xEscrow001", "0xEscrow002", "0xUnknown003"},
	}

	mockStore.On("GetChannelByID", "0xEscrow001").Return(openEscrow, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == "0xEscrow001" && ch.Status == core.ChannelStatusClosed
	})).Return(nil)
	mockStore.On("GetChannelByID", "0xEscrow002").Return(closedEscrow, nil)
	mockStore.On("GetChannelByID", "0xUnknown003").Return(nil, nil)

	err := service.HandleEscrowDepositsPurged(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositsPurged_StoreError_Propagates(t *testing.T) {
	mockStore := new(MockStore)
	ctx := log.SetContextLogger(context.Background(), log.NewNoopLogger())
	service := &EventHandlerService{}

	event := &core.EscrowDepositsPurgedEvent{
		EscrowIDs: []string{"0xEscrow001"},
	}

	mockStore.On("GetChannelByID", "0xEscrow001").Return(nil, errors.New("db error"))

	err := service.HandleEscrowDepositsPurged(ctx, mockStore, event)

	require.Error(t, err)
	require.Contains(t, err.Error(), "db error")
	mockStore.AssertExpectations(t)
}
