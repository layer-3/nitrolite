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
	nodeSigner  *core.ChannelDefaultSigner
	statePacker core.StatePacker
}

// NewEventHandlerService creates a new EventHandlerService instance.
// nodeSigner and statePacker are used to backfill the node signature on the
// checkpointed head state when it is missing from the local record.
func NewEventHandlerService(nodeSigner *core.ChannelDefaultSigner, statePacker core.StatePacker) *EventHandlerService {
	return &EventHandlerService{
		nodeSigner:  nodeSigner,
		statePacker: statePacker,
	}
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
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

	wasChallenged := channel.Status == core.ChannelStatusChallenged
	if wasChallenged {
		channel.Status = core.ChannelStatusOpen
	}

	err = tx.UpdateChannel(*channel)
	if err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	// When a challenge is resolved by checkpoint, the head state row may have been
	// persisted without a node signature (e.g. a receiver state stored unsigned during
	// the dispute window that is now the highest version at or below the checkpointed
	// version). Backfill the node signature locally so future flows treat the head as
	// fully co-signed. On normal Open→Open checkpoints the row is already node-signed
	// via the standard RPC path, so this work is skipped. Higher-version unsigned
	// receiver states are intentionally left untouched and reconciled on close.
	nodeSig := ""
	if wasChallenged {
		nodeSig, err = s.buildHeadNodeSig(ctx, tx, event.ChannelID, event.StateVersion)
		if err != nil {
			return err
		}
	}
	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, nodeSig); err != nil {
		return err
	}

	logger.Info("handled HomeChannelCheckpointed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// buildHeadNodeSig returns a node signature for the stored state at (channelID, version)
// when the row is present and lacks a node signature. Returns an empty string when the
// state row is missing or already node-signed; callers pass the result straight through
// to UpdateStateSigsIfMissing which no-ops on empty input.
func (s *EventHandlerService) buildHeadNodeSig(ctx context.Context, tx core.ChannelHubEventHandlerStore, channelID string, version uint64) (string, error) {
	state, err := tx.GetStateByChannelIDAndVersion(channelID, version)
	if err != nil {
		return "", err
	}
	if state == nil || state.NodeSig != nil {
		return "", nil
	}
	packed, err := s.statePacker.PackState(*state)
	if err != nil {
		return "", err
	}
	sig, err := s.nodeSigner.Sign(packed)
	if err != nil {
		return "", err
	}
	log.FromContext(ctx).Info("backfilled missing node signature on head state",
		"channelId", channelID, "stateVersion", version)
	return sig.String(), nil
}

// HandleHomeChannelChallenged processes the HomeChannelChallenged event emitted when a potentially
// stale state is submitted on-chain. It marks the channel as Challenged and persists the challenge
// expiry so subsequent state-submission paths (CheckActiveChannel, RefreshUserEnforcedBalance) stop
// treating the channel as open. Automatic challenge response is intentionally disabled: the latest
// signed state may carry an intent (e.g. CLOSE, escrow initiate/finalize, migration) that cannot
// be resolved via ScheduleCheckpoint, and silently queueing an impossible transaction risks
// letting the challenge expire on a stale state. A warning is emitted so operators submit the
// appropriate on-chain action manually before expiry.
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
		// Per protocol the challenged version cannot be lower than the last known on-chain version.
		// Treat as an anomaly (replay, indexer mis-order, contract bug): warn and skip persistence.
		logger.Warn("challenged state version is less than current channel state version, ignoring", "channelId", chanID, "currentStateVersion", channel.StateVersion, "challengedStateVersion", event.StateVersion)
		return nil
	}

	channel.StateVersion = event.StateVersion
	// Closing → Challenged is an expected transition: a co-signed Finalize may race an
	// on-chain challenge. The chain takes precedence; the off-chain close flow is abandoned
	// and the channel follows the Challenged → Closed path instead.
	channel.Status = core.ChannelStatusChallenged
	expirationTime := time.Unix(int64(event.ChallengeExpiry), 0)
	channel.ChallengeExpiresAt = &expirationTime

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	logger.Warn("home channel challenged",
		"channelId", chanID,
		"userWallet", channel.UserWallet,
		"blockchainID", channel.BlockchainID,
		"asset", channel.Asset,
		"challengedStateVersion", event.StateVersion,
		"challengeExpiry", expirationTime,
	)
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	logger.Info("handled EscrowDepositInitiated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowDepositChallenged processes the EscrowDepositChallenged event emitted when an escrow
// deposit is challenged on-chain. It marks the channel as Challenged and sets the expiration time.
// Resolution policy depends on whether the node holds a newer fully-signed state for this channel:
//   - If a newer signed FINALIZE_ESCROW_DEPOSIT exists, finalize the escrow on the non-home chain.
//   - Otherwise the user is withholding finalize: defend the node allocation on the home chain by
//     scheduling challengeChannel(...) with the INITIATE_ESCROW_DEPOSIT state. Without this, the user
//     can let the non-home challenge expire, recover escrow-chain funds, and still threaten the
//     home-chain finalize path against the node's locked allocation.
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
	if lastSignedState != nil && lastSignedState.Version > event.StateVersion {
		if lastSignedState.EscrowLedger == nil {
			logger.Warn("last signed state has no escrow ledger during EscrowDepositChallenged event", "channelId", chanID)
		} else {
			if err := tx.ScheduleFinalizeEscrowDeposit(lastSignedState.ID, lastSignedState.EscrowLedger.BlockchainID); err != nil {
				return err
			}
		}
	} else {
		if err := s.scheduleHomeChannelChallengeForEscrowDeposit(ctx, tx, chanID, event.StateVersion); err != nil {
			return err
		}
	}

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	logger.Info("handled EscrowDepositChallenged event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// scheduleHomeChannelChallengeForEscrowDeposit queues a challengeChannel(...) submission on the home
// chain using the INITIATE_ESCROW_DEPOSIT state referenced by the escrow event. This anchors the
// home channel in DISPUTED so the user cannot later push a withheld FINALIZE state on home, and
// starts the home-chain challenge timer the operator uses to recover the node allocation.
func (s *EventHandlerService) scheduleHomeChannelChallengeForEscrowDeposit(ctx context.Context, tx core.ChannelHubEventHandlerStore, escrowChanID string, stateVersion uint64) error {
	logger := log.FromContext(ctx)

	initiateState, err := tx.GetStateByChannelIDAndVersion(escrowChanID, stateVersion)
	if err != nil {
		return err
	}
	if initiateState == nil {
		logger.Error("INITIATE_ESCROW_DEPOSIT state missing locally, cannot defend home channel automatically", "escrowChannelId", escrowChanID, "stateVersion", stateVersion)
		return nil
	}
	if initiateState.HomeChannelID == nil {
		logger.Error("INITIATE_ESCROW_DEPOSIT state has no home channel ID, cannot defend home channel automatically", "escrowChannelId", escrowChanID, "stateVersion", stateVersion)
		return nil
	}

	homeChannel, err := tx.GetChannelByID(*initiateState.HomeChannelID)
	if err != nil {
		return err
	}
	if homeChannel == nil {
		logger.Error("home channel not found, cannot defend home channel automatically", "homeChannelId", *initiateState.HomeChannelID, "escrowChannelId", escrowChanID)
		return nil
	}
	if homeChannel.Status != core.ChannelStatusOpen {
		switch homeChannel.Status {
		case core.ChannelStatusChallenged:
			logger.Warn("home channel already Challenged, skipping auto-challenge", "homeChannelId", *initiateState.HomeChannelID, "escrowChannelId", escrowChanID)
		case core.ChannelStatusClosed:
			logger.Error("home channel Closed, defense window passed", "homeChannelId", *initiateState.HomeChannelID, "escrowChannelId", escrowChanID)
		default:
			logger.Warn("home channel not Open, skipping auto-challenge", "homeChannelId", *initiateState.HomeChannelID, "homeStatus", homeChannel.Status, "escrowChannelId", escrowChanID)
		}
		return nil
	}

	if initiateState.HomeLedger.BlockchainID == 0 {
		logger.Error("INITIATE_ESCROW_DEPOSIT state has zero home BlockchainID, cannot defend home channel automatically", "homeChannelId", *initiateState.HomeChannelID, "escrowChannelId", escrowChanID)
		return nil
	}

	if err := tx.ScheduleChallenge(initiateState.ID, initiateState.HomeLedger.BlockchainID); err != nil {
		return err
	}

	logger.Warn("scheduled home-channel challenge to defend node allocation against withheld escrow finalize",
		"homeChannelId", *initiateState.HomeChannelID,
		"escrowChannelId", escrowChanID,
		"stateVersion", stateVersion,
	)
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	logger.Info("handled EscrowDepositFinalized event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// HandleEscrowDepositsPurged processes the EscrowDepositsPurged event emitted when expired escrow deposits
// are finalized by the on-chain purge queue without a signed FINALIZE_ESCROW_DEPOSIT state. It marks each
// corresponding escrow channel as Closed, preserving its existing StateVersion.
//
// TODO: consider scoping the DB transaction per channel update instead of wrapping the whole batch,
// so a single failure does not roll back already-processed channels in the same purge event.
func (s *EventHandlerService) HandleEscrowDepositsPurged(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.EscrowDepositsPurgedEvent) error {
	logger := log.FromContext(ctx)
	closedCount := 0

	for _, escrowID := range event.EscrowIDs {
		channel, err := tx.GetChannelByID(escrowID)
		if err != nil {
			return err
		}
		if channel == nil {
			logger.Debug("channel not found in DB during EscrowDepositsPurged event", "escrowId", escrowID)
			continue
		}
		if channel.Type != core.ChannelTypeEscrow {
			logger.Warn("channel type mismatch during EscrowDepositsPurged event", "escrowId", escrowID, "expectedType", core.ChannelTypeEscrow, "actualType", channel.Type)
			continue
		}
		if channel.Status == core.ChannelStatusClosed {
			continue
		}

		channel.Status = core.ChannelStatusClosed
		if err := tx.UpdateChannel(*channel); err != nil {
			return err
		}
		closedCount++
	}

	logger.Info("handled EscrowDepositsPurged event", "purgedCount", len(event.EscrowIDs), "closedCount", closedCount)
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
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

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
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
