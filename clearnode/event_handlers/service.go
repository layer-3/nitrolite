package event_handlers

import (
	"context"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

var _ core.ChannelHubEventHandler = &EventHandlerService{}
var _ core.LockingContractEventHandler = &EventHandlerService{}

// EventHandlerService processes blockchain events and updates the local database state accordingly.
// It handles events from both home channels (user state channels) and escrow channels (temporary lock channels).
type EventHandlerService struct {
}

// NewEventHandlerService creates a new EventHandlerService instance.
func NewEventHandlerService() *EventHandlerService {
	return &EventHandlerService{}
}

// HandleNodeBalanceUpdated processes the NodeBalanceUpdated event emitted when a node's balance is updated on-chain.
// It updates the user's staked balance for the specified blockchain in the database.
func (s *EventHandlerService) HandleNodeBalanceUpdated(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.NodeBalanceUpdatedEvent) error {
	logger := log.FromContext(ctx)

	if err := tx.SetNodeBalance(event.BlockchainID, event.Asset, event.Balance); err != nil {
		return err
	}

	logger.Info("handled NodeBalanceUpdated event", "blockchainID", event.BlockchainID, "asset", event.Asset, "balance", event.Balance)
	return nil
}

// HandleHomeChannelCreated processes the HomeChannelCreated event emitted when a home channel
// is successfully created on-chain. It updates the channel status to Open and sets the state version.
// The channel must exist in the database with type ChannelTypeHome, otherwise a warning is logged.
func (s *EventHandlerService) HandleHomeChannelCreated(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelCreatedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Warn("channel not found in DB during HomeChannelCreated event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeHome {
		logger.Warn("channel type mismatch during HomeChannelCreated event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
		return nil
	}
	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusOpen

	err = tx.UpdateChannel(*channel)
	if err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	logger.Info("handled HomeChannelCreated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleHomeChannelMigrated processes the HomeChannelMigrated event emitted when a home channel
// is migrated to a new version or blockchain. This is currently not implemented and logs a warning.
// TODO: Implement HomeChannelMigrated handler logic
func (s *EventHandlerService) HandleHomeChannelMigrated(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelMigratedEvent) error {
	logger := log.FromContext(ctx)
	logger.Warn("unexpected HomeChannelMigrated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion)
	return nil
}

// HandleHomeChannelCheckpointed processes the HomeChannelCheckpointed event emitted when a channel
// state is successfully checkpointed on-chain. It updates the channel's state version and clears
// the Challenged status if present, returning the channel to Open status.
func (s *EventHandlerService) HandleHomeChannelCheckpointed(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelCheckpointedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during HomeChannelCheckpointed event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeHome {
		logger.Warn("channel type mismatch during HomeChannelCheckpointed event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
		return nil
	}
	channel.StateVersion = event.StateVersion

	if channel.Status == core.ChannelStatusChallenged {
		channel.Status = core.ChannelStatusOpen
	}

	err = tx.UpdateChannel(*channel)
	if err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	logger.Info("handled HomeChannelCheckpointed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleHomeChannelChallenged processes the HomeChannelChallenged event emitted when a potentially
// stale state is submitted on-chain. It updates the channel status to Challenged, sets the challenge
// expiration time, and automatically schedules a checkpoint of the latest signed state if available
// to resolve the challenge.
func (s *EventHandlerService) HandleHomeChannelChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelChallengedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during HomeChannelChallenged event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeHome {
		logger.Warn("channel type mismatch during HomeChannelChallenged event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
		return nil
	}

	if event.StateVersion < channel.StateVersion {
		logger.Error("challenged state version is less than current channel state version", "channelId", chanID, "currentStateVersion", channel.StateVersion, "challengedStateVersion", event.StateVersion)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusChallenged

	expirationTime := time.Unix(int64(event.ChallengeExpiry), 0)
	channel.ChallengeExpiresAt = &expirationTime

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	lastSignedState, err := tx.GetLastStateByChannelID(chanID, true)
	if err != nil {
		return err
	}
	if lastSignedState == nil {
		logger.Warn("no state found for channel during HomeChannelChallenged event", "channelId", chanID)
	} else if lastSignedState.Version <= event.StateVersion {
		logger.Warn("last signed state version is not greater than challenged state version", "channelId", chanID, "lastSignedStateVersion", lastSignedState.Version, "challengedStateVersion", event.StateVersion)
	} else {
		if err := tx.ScheduleCheckpoint(lastSignedState.ID, lastSignedState.HomeLedger.BlockchainID); err != nil {
			return err
		}
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	logger.Info("handled HomeChannelChallenged event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleHomeChannelClosed processes the HomeChannelClosed event emitted when a home channel is
// finalized and closed on-chain. It updates the channel status to Closed and sets the final state version.
// Once closed, no further state updates are possible for this channel.
func (s *EventHandlerService) HandleHomeChannelClosed(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelClosedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during HomeChannelClosed event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeHome {
		logger.Warn("channel type mismatch during HomeChannelClosed event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusClosed

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	logger.Info("handled HomeChannelClosed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowDepositInitiated processes the EscrowDepositInitiated event emitted when an escrow
// deposit operation begins on-chain. It updates the escrow channel status to Open, sets the state
// version, and schedules a checkpoint to finalize the deposit if a matching state exists in the database.
func (s *EventHandlerService) HandleEscrowDepositInitiated(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowDepositInitiatedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowDepositInitiated event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowDepositInitiated event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusOpen

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	state, err := tx.GetStateByChannelIDAndVersion(chanID, event.StateVersion)
	if err != nil {
		return err
	}
	if state == nil {
		logger.Warn("no state found for channel during EscrowDepositInitiated event", "channelId", chanID)
	} else {
		if err := tx.ScheduleInitiateEscrowDeposit(state.ID, state.HomeLedger.BlockchainID); err != nil {
			return err
		}
	}

	logger.Info("handled EscrowDepositInitiated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowDepositChallenged processes the EscrowDepositChallenged event emitted when an escrow
// deposit is challenged on-chain. Similar to home channel challenges, it marks the channel as Challenged,
// sets the expiration time, and automatically schedules a checkpoint with the latest signed state
// to resolve the challenge.
func (s *EventHandlerService) HandleEscrowDepositChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowDepositChallengedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowDepositChallenged event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowDepositChallenged event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	if event.StateVersion < channel.StateVersion {
		logger.Error("challenged escrow deposit state version is less than current channel state version", "channelId", chanID, "currentStateVersion", channel.StateVersion, "challengedStateVersion", event.StateVersion)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusChallenged

	expirationTime := time.Unix(int64(event.ChallengeExpiry), 0)
	channel.ChallengeExpiresAt = &expirationTime

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	lastSignedState, err := tx.GetLastStateByChannelID(chanID, true)
	if err != nil {
		return err
	}
	if lastSignedState == nil {
		logger.Warn("no state found for channel during EscrowDepositChallenged event", "channelId", chanID)
	} else if lastSignedState.Version <= event.StateVersion {
		logger.Warn("last signed state version is not greater than challenged state version", "channelId", chanID, "lastSignedStateVersion", lastSignedState.Version, "challengedStateVersion", event.StateVersion)
	} else {
		if lastSignedState.EscrowLedger == nil {
			logger.Warn("last signed state has no escrow ledger during EscrowDepositChallenged event", "channelId", chanID)
		} else {
			if err := tx.ScheduleFinalizeEscrowDeposit(lastSignedState.ID, lastSignedState.EscrowLedger.BlockchainID); err != nil {
				return err
			}
		}
	}

	logger.Info("handled EscrowDepositChallenged event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowDepositFinalized processes the EscrowDepositFinalized event emitted when an escrow
// deposit is successfully finalized on-chain. It updates the channel status to Closed and sets
// the final state version, completing the deposit lifecycle.
func (s *EventHandlerService) HandleEscrowDepositFinalized(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowDepositFinalizedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowDepositFinalized event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowDepositFinalized event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusClosed

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	logger.Info("handled EscrowDepositFinalized event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowWithdrawalInitiated processes the EscrowWithdrawalInitiated event emitted when an escrow
// withdrawal operation begins on-chain. It updates the escrow channel status to Open and sets the state
// version to reflect the initiated withdrawal.
func (s *EventHandlerService) HandleEscrowWithdrawalInitiated(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowWithdrawalInitiatedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowWithdrawalInitiated event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowWithdrawalInitiated event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusOpen

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	logger.Info("handled EscrowWithdrawalInitiated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowWithdrawalChallenged processes the EscrowWithdrawalChallenged event emitted when an escrow
// withdrawal is challenged on-chain. It marks the channel as Challenged, sets the expiration time,
// and schedules a checkpoint for escrow withdrawal with the latest signed state to resolve the challenge.
func (s *EventHandlerService) HandleEscrowWithdrawalChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowWithdrawalChallengedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowWithdrawalChallenged event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowWithdrawalChallenged event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	if event.StateVersion < channel.StateVersion {
		logger.Error("challenged escrow withdrawal state version is less than current channel state version", "channelId", chanID, "currentStateVersion", channel.StateVersion, "challengedStateVersion", event.StateVersion)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusChallenged

	expirationTime := time.Unix(int64(event.ChallengeExpiry), 0)
	channel.ChallengeExpiresAt = &expirationTime

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	lastSignedState, err := tx.GetLastStateByChannelID(chanID, true)
	if err != nil {
		return err
	}
	if lastSignedState == nil {
		logger.Warn("no state found for channel during EscrowWithdrawalChallenged event", "channelId", chanID)
	} else if lastSignedState.Version <= event.StateVersion {
		logger.Warn("last signed state version is not greater than challenged state version", "channelId", chanID, "lastSignedStateVersion", lastSignedState.Version, "challengedStateVersion", event.StateVersion)
	} else {
		if lastSignedState.EscrowLedger == nil {
			logger.Warn("last signed state has no escrow ledger during EscrowWithdrawalChallenged event", "channelId", chanID)
		} else {
			if err := tx.ScheduleFinalizeEscrowWithdrawal(lastSignedState.ID, lastSignedState.EscrowLedger.BlockchainID); err != nil {
				return err
			}
		}
	}

	logger.Info("handled EscrowWithdrawalChallenged event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowWithdrawalFinalized processes the EscrowWithdrawalFinalized event emitted when an escrow
// withdrawal is successfully finalized on-chain. It updates the channel status to Closed and sets
// the final state version, completing the withdrawal lifecycle.
func (s *EventHandlerService) HandleEscrowWithdrawalFinalized(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowWithdrawalFinalizedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	channel, err := tx.GetChannelByID(chanID)
	if err != nil {
		return err
	}
	if channel == nil {
		logger.Debug("channel not found in DB during EscrowWithdrawalFinalized event", "channelId", chanID)
		return nil
	}
	if channel.Type != core.ChannelTypeEscrow {
		logger.Warn("channel type mismatch during EscrowWithdrawalFinalized event", "channelId", chanID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
		return nil
	}

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusClosed

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	logger.Info("handled EscrowWithdrawalFinalized event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

func (s *EventHandlerService) HandleUserLockedBalanceUpdated(ctx context.Context, tx core.LockingContractEventHandlerStore, event *core.UserLockedBalanceUpdatedEvent) error {
	logger := log.FromContext(ctx)
	err := tx.UpdateUserStaked(event.UserAddress, event.BlockchainID, event.Balance)
	if err != nil {
		return err
	}

	logger.Info("handled UserLockedBalanceUpdatedEvent event", "userWallet", event.UserAddress, "blockchainID", event.BlockchainID, "balance", event.Balance)
	return nil
}
