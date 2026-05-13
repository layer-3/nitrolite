package event_handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(1), "").Return(nil)

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

	service := &EventHandlerService{}

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
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(5), "").Return(nil)

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
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 4 &&
			ch.ChallengeExpiresAt != nil &&
			ch.ChallengeExpiresAt.Unix() == int64(challengeExpiry)
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(4), "").Return(nil)

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
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 4 &&
			ch.ChallengeExpiresAt != nil &&
			ch.ChallengeExpiresAt.Unix() == int64(challengeExpiry)
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(4), "").Return(nil)

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
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 10
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(10), "").Return(nil)

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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(1), "").Return(nil)

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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(3), "").Return(nil)

	// Execute
	err := service.HandleEscrowDepositChallenged(ctx, mockStore, event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(5), "").Return(nil)

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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(1), "").Return(nil)

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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(3), "").Return(nil)

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
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(5), "").Return(nil)

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
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.StateVersion == 5
	})).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(5), userSig).Return(nil)

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
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
	mockStore.On("UpdateChannel", mock.Anything).Return(nil)
	mockStore.On("RefreshUserEnforcedBalance", userWallet, "usdc").Return(nil)
	mockStore.On("UpdateStateUserSigIfMissing", channelID, uint64(5), "0xdeadbeef").Return(errors.New("db error"))

	err := service.HandleHomeChannelCheckpointed(ctx, mockStore, event)

	require.Error(t, err)
	require.Contains(t, err.Error(), "db error")
	mockStore.AssertExpectations(t)
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
