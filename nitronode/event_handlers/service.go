package event_handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

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

// HandleNodeBalanceUpdated processes the NodeBalanceUpdated event emitted when the node's
// on-chain liquidity changes. It records the new node liquidity for the (blockchain, asset)
// pair via SetNodeBalance; this is observability data only and does not affect user staking
// state.
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
// A legitimate Created event is observed exactly once per channel, when the local row is still in
// the ChannelStatusVoid state seeded by CreateChannel. If the channel has already advanced past
// Void, this handler is being re-fired (indexer replay, chain reorg, block reprocessing) and any
// mutation here would regress the post-Open lifecycle — most importantly resetting a Closing
// channel back to Open and erasing the Finalize marker that gates CheckActiveChannel.
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
	if channel.Status >= core.ChannelStatusOpen {
		logger.Warn("ignoring replayed HomeChannelCreated event on already-initialized channel",
			"channelId", chanID, "currentStatus", channel.Status, "currentStateVersion", channel.StateVersion, "eventStateVersion", event.StateVersion)
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
// the Challenged status if present, returning the channel to Open — unless the local DB already
// holds a co-signed Finalize for this channel, in which case the post-Finalize Closing marker
// is restored instead. Without that restore, a Closing → Challenged → Open sequence driven by
// on-chain events would erase the fact that the node has already signed a finalized state, and
// CheckActiveChannel would let the user submit further transitions past the finalized state.
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

	// Acquire the user's balance-row lock before mutating channel status or
	// backfilling the off-chain head. Receiver issuance paths lock the same row
	// and then re-check Status; without this lock an RPC can read Status=Challenged,
	// decide to store an unsigned receiver row, but commit after we flip to Open
	// and backfill the prior head — leaving the latest head unsigned on an Open
	// channel. See HandleHomeChannelClosed for the same pattern.
	if _, err := tx.LockUserState(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	channel.StateVersion = event.StateVersion

	wasChallenged := channel.Status == core.ChannelStatusChallenged
	if wasChallenged {
		// Reconstruct the post-Finalize Closing marker from channel_states: if the node
		// has already signed a Finalize state for this channel, the off-chain close is
		// still pending and the channel must not return to Open. See the doc comment on
		// HandleHomeChannelChallenged for the round-trip rationale.
		finalized, err := tx.HasSignedFinalize(chanID)
		if err != nil {
			return err
		}
		if finalized {
			channel.Status = core.ChannelStatusClosing
		} else {
			channel.Status = core.ChannelStatusOpen
		}
		// The challenge is resolved on chain; the expiry timestamp is no longer relevant
		// and would otherwise surface as a stale deadline through the channel API.
		channel.ChallengeExpiresAt = nil
	}

	err = tx.UpdateChannel(*channel)
	if err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	// Always backfill the user signature at the on-chain checkpointed version so the
	// row matches what is enforced on chain.
	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	// When a challenge is cleared, the off-chain head may sit above event.StateVersion:
	// any receiver state issued during the challenge was stored unsigned and is now the
	// channel's actual latest state. Backfill the node signature on that head so future
	// flows treat it as fully co-signed. On normal Open→Open checkpoints the head row
	// is already node-signed via the RPC path and this is a no-op.
	if wasChallenged {
		if err := s.backfillOffChainHeadNodeSig(ctx, tx, event.ChannelID); err != nil {
			return err
		}
	}

	logger.Info("handled HomeChannelCheckpointed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// backfillOffChainHeadNodeSig loads the off-chain head state for channelID (the highest
// stored version, regardless of signature status) and node-signs it when the row is
// present and the node signature is missing. The user signature is intentionally left
// untouched: when the head was created during a challenge it carries no user signature,
// and the user must countersign and acknowledge it via the regular RPC flow.
func (s *EventHandlerService) backfillOffChainHeadNodeSig(ctx context.Context, tx core.ChannelHubEventHandlerStore, channelID string) error {
	head, err := tx.GetLastStateByChannelID(channelID, false)
	if err != nil {
		return err
	}
	if head == nil || head.NodeSig != nil {
		return nil
	}
	// Per the challenge-clearance spec the only states accumulated during the dispute
	// window are receiver credits (transfer_receive, release) — user-initiated ops are
	// rejected upstream while the channel is Challenged. If the head is some other
	// transition kind, the invariant has broken upstream and we must not silently
	// node-sign it. Log it and bail so the caller surfaces the inconsistency.
	if head.Transition.Type != core.TransitionTypeTransferReceive &&
		head.Transition.Type != core.TransitionTypeRelease {
		log.FromContext(ctx).Debug("off-chain head after challenge clearance is not a receiver state, skipping node-sig backfill",
			"channelId", channelID,
			"transitionType", head.Transition.Type,
			"version", head.Version,
		)
		return nil
	}
	packed, err := s.statePacker.PackState(*head)
	if err != nil {
		return err
	}
	sig, err := s.nodeSigner.Sign(packed)
	if err != nil {
		return err
	}
	if err := tx.UpdateStateSigsIfMissing(channelID, head.Version, "", sig.String()); err != nil {
		return err
	}
	log.FromContext(ctx).Info("backfilled missing node signature on off-chain head state",
		"channelId", channelID, "stateVersion", head.Version)
	return nil
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

	// Acquire the user's balance-row lock before mutating channel status. Receiver
	// issuance paths (issueTransferReceiverState / issueReleaseReceiverState) lock
	// the same row up front and then re-check Status via CheckActiveChannel; without
	// this lock an in-flight RPC can read Status=Open, node-sign a receiver state,
	// and commit after we flip to Challenged — leaving a node-signed higher-version
	// receiver state on a disputed channel. See HandleHomeChannelClosed for the same
	// pattern.
	if _, err := tx.LockUserState(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	if event.StateVersion < channel.StateVersion {
		// Per protocol the challenged version cannot be lower than the last known on-chain version.
		// Treat as an anomaly (replay, indexer mis-order, contract bug): warn and skip persistence.
		logger.Warn("challenged state version is less than current channel state version, ignoring", "channelId", chanID, "currentStateVersion", channel.StateVersion, "challengedStateVersion", event.StateVersion)
		return nil
	}

	channel.StateVersion = event.StateVersion
	// Closing → Challenged is an expected transition: a co-signed Finalize may race an
	// on-chain challenge. The status field is intentionally overwritten — the chain takes
	// precedence while the dispute is live. The post-Finalize fact is not lost: it is
	// shadowed in channel_states (the latest fully-signed row carries TransitionTypeFinalize)
	// and HandleHomeChannelCheckpointed restores ChannelStatusClosing from there when the
	// challenge resolves, so the off-chain close flow resumes instead of silently regressing
	// to Open.
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
//
// Additionally, when the closing path was a challenge resolution (channel was Challenged
// and the closing state is not a finalize), the handler issues a single ChallengeRescue
// state for the user that squashes the sum of receiver-state credits accrued during the
// challenge window into an off-channel ledger entry tied to the closed channel's ID.
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

	// Acquire the user's balance-row lock before mutating channel status or summing
	// receiver credits. issueTransferReceiverState / issueReleaseReceiverState lock
	// the same row up front, so this serializes the close against any in-flight RPC
	// receiver-issuance for the same user: either the RPC commits its unsigned row
	// before we sum (it lands in the rescue), or it blocks until we set the channel
	// to Closed and then sees that via its own re-check.
	if _, err := tx.LockUserState(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	wasChallenged := channel.Status == core.ChannelStatusChallenged

	channel.StateVersion = event.StateVersion
	channel.Status = core.ChannelStatusClosed
	// Channel is terminal; any pending challenge deadline is no longer meaningful.
	channel.ChallengeExpiresAt = nil

	if err := tx.UpdateChannel(*channel); err != nil {
		return err
	}

	if err := tx.RefreshUserEnforcedBalance(channel.UserWallet, channel.Asset); err != nil {
		return err
	}

	if err := tx.UpdateStateSigsIfMissing(event.ChannelID, event.StateVersion, event.UserSig, ""); err != nil {
		return err
	}

	if wasChallenged {
		if err := s.issueChallengeRescue(ctx, tx, channel, event.StateVersion); err != nil {
			return err
		}
	}

	logger.Info("handled HomeChannelClosed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)
	return nil
}

// issueChallengeRescue emits a ChallengeRescue state on the user's ledger after a
// challenged-channel close. The state is issued unconditionally so the user's latest
// stored state moves to a fresh epoch with HomeChannelID nil; without it future
// receiver-state issuance and channels.v1.request_creation would stay wedged on the
// closed channel. The rescue amount is the NET effect on the user's home-channel
// balance of transitions stored strictly above closureVersion — receives
// (TransferReceive, Release) credit the user, sends (TransferSend, Commit) debit,
// everything else is excluded because it requires onchain backing the chain didn't
// enforce or belongs to a different ledger. Signed (Open-time) and unsigned
// (during-challenge) rows both contribute. The result is clamped at zero so an
// adversarial close at a version where the user's own balance was higher than the
// off-chain head can't dock the user further. AccountID is the closed channel's ID;
// HomeChannelID is nil — the shape of a credit to a user with no open home channel.
func (s *EventHandlerService) issueChallengeRescue(ctx context.Context, tx core.ChannelHubEventHandlerStore, channel *core.Channel, closureVersion uint64) error {
	logger := log.FromContext(ctx)

	prev, err := tx.GetLastStateByChannelID(channel.ChannelID, false)
	if err != nil {
		return err
	}
	if prev == nil {
		// Should not happen for a channel that reached Challenged → Closed: at least the
		// closure state itself must be on file. Surface the inconsistency rather than
		// silently dropping the rescue.
		return fmt.Errorf("no state found for closed challenged channel %s", channel.ChannelID)
	}

	// Strict `>` against closureVersion: the row at the closure version itself is the
	// closing state and must be excluded — only transitions issued strictly after the
	// dispute version are unenforced. The epoch filter pins the sum to prev.Epoch (the
	// closed channel's epoch) as a defense against any future DB inconsistency.
	net, err := tx.SumNetTransitionAmountAfterVersion(channel.ChannelID, closureVersion, prev.Epoch)
	if err != nil {
		return err
	}

	// Negative net is only reachable when the user closed at a version where her own
	// channel balance was higher than the off-chain head (adversarial rollback of her
	// own sends/commits). Onchain has already paid her above the head value; rescue
	// must not dock further. Honest challenges typically have receives dominating,
	// net >= 0. Clamp defensively and log.
	total := net
	if net.IsNegative() {
		logger.Warn("challenge_rescue net is negative, clamping to zero",
			"channelId", channel.ChannelID,
			"netAmount", net.String(),
			"closureVersion", closureVersion,
			"prevVersion", prev.Version,
		)
		total = decimal.Zero
	}

	// Invariant: any row included in the sum sits strictly above closureVersion, so the
	// channel's off-chain head must too. A non-zero net with prev at or below closure
	// means the state chain disagrees with itself — surface it before issuing the rescue.
	if !total.IsZero() && prev.Version <= closureVersion {
		return fmt.Errorf("challenge_rescue: non-zero net (%s) but prev v=%d <= closure v=%d on channel %s",
			total.String(), prev.Version, closureVersion, channel.ChannelID)
	}

	rescue, err := core.NewChallengeRescueState(*prev, total)
	if err != nil {
		return err
	}

	// The rescue state is off-channel (HomeChannelID == nil) and is therefore not
	// node-signed via the channel packer. It is treated like a credit to a user with no
	// open home channel: the value is recorded in the user's state chain and will be
	// folded into a properly signed state when the user next opens a channel.
	if err := tx.StoreUserState(*rescue, ""); err != nil {
		return err
	}

	txn, err := core.NewTransactionFromTransition(nil, rescue, rescue.Transition)
	if err != nil {
		return err
	}
	if err := tx.RecordTransaction(*txn, ""); err != nil {
		return err
	}

	logger.Info("issued challenge_rescue state",
		"channelId", channel.ChannelID,
		"userWallet", channel.UserWallet,
		"asset", channel.Asset,
		"amount", total.String(),
		"newStateID", rescue.ID,
		"txID", txn.ID,
	)
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
	// Channel is terminal; any pending challenge deadline is no longer meaningful.
	channel.ChallengeExpiresAt = nil

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
	// Channel is terminal; any pending challenge deadline is no longer meaningful.
	channel.ChallengeExpiresAt = nil

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
