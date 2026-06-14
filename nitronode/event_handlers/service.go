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

// guardEventVersionMonotonic returns true (drop=true) when the incoming event's
// StateVersion is strictly less than the row's current StateVersion. The caller
// must `return nil` immediately when drop is true; the helper logs a structured
// warning identifying which event intent arrived stale.
//
// Intent is a short stable string describing the on-chain transition the dropped
// event would have applied (e.g. "checkpointed", "closed", "escrow_deposit_initiated").
// It surfaces in the warn log so operators can correlate drops with reentrancy
// scenarios in MF3-L19 / nitronode-event-monotonicity.md.
//
// Once §B lands, this helper is the single chokepoint where the on-drop
// chain-state refresh will be invoked: every dropped event triggers exactly one
// RefreshChannelFromChain call before the function returns drop=true, so the
// refresh policy stays consistent across all six handlers.
func guardEventVersionMonotonic(
	ctx context.Context,
	logger log.Logger,
	chanID string,
	intent string,
	eventVersion uint64,
	currentVersion uint64,
) (drop bool) {
	if eventVersion >= currentVersion {
		return false
	}
	logger.Warn("event state version is less than current channel state version, ignoring",
		"channelId", chanID,
		"intent", intent,
		"currentStateVersion", currentVersion,
		"eventStateVersion", eventVersion)
	return true
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
	// Acquire the user's balance-row lock and read the channel under it before mutating status:
	// the lock guards against a concurrent submit_state flipping channel status (e.g. Void→Open
	// receiver issuance) between the read and our write. See HandleHomeChannelCheckpointed and
	// HandleHomeChannelClosed for the same pattern.
	channel, err := tx.LockUserStateForHomeChannel(chanID)
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

	// Backfill the node signature on any unsigned receiver-credit state that
	// landed while the channel was still Void. An unsigned head is possible when
	// an incoming transfer or release was stored by a concurrent RPC before this
	// Created event arrived and flipped the channel to Open. The guard inside
	// backfillOffChainHeadNodeSig ensures only transfer_receive / release heads
	// are signed, so this is a no-op for the normal case where the head is the
	// CREATE state itself (already node-signed via the RPC path).
	if err := s.backfillOffChainHeadNodeSig(ctx, tx, event.ChannelID); err != nil {
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
	// Acquire the user's balance-row lock and read the channel under it before mutating
	// channel status or backfilling the off-chain head. Reading the channel before the
	// lock would race: a concurrent submit_state can co-sign a Finalize and flip the
	// channel Open→Closing between the read and the lock, and the non-challenged path
	// below persists the channel snapshot verbatim — a pre-lock Open snapshot would
	// silently reopen the finalized channel. Receiver issuance paths lock the same row
	// and then re-check Status; without this lock an RPC can read Status=Challenged,
	// decide to store an unsigned receiver row, but commit after we flip to Open and
	// backfill the prior head — leaving the latest head unsigned on an Open channel. See
	// HandleHomeChannelClosed for the same pattern.
	channel, err := tx.LockUserStateForHomeChannel(chanID)
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

	// Per protocol the checkpointed version cannot be lower than the last known on-chain
	// version. This branch is reachable when contract reentrancy emits an inner
	// higher-version event before the outer ChannelCheckpointed (see MF3-L19 /
	// nitronode-event-monotonicity.md scenarios 1–3). Drop the event so we do not
	// regress channel.StateVersion and, critically, so the wasChallenged branch below
	// does not flip a live challenge back to Open based on a stale version.
	if guardEventVersionMonotonic(ctx, logger, chanID, "checkpointed", event.StateVersion, channel.StateVersion) {
		return nil
	}

	channel.StateVersion = event.StateVersion

	// Snapshot the pre-checkpoint status once and derive both transition flags from it, so the
	// Void and Challenged branches below are independent of each other's mutation order: either
	// branch may reassign channel.Status without silently making the other unreachable.
	prevStatus := channel.Status
	wasVoid := prevStatus == core.ChannelStatusVoid
	wasChallenged := prevStatus == core.ChannelStatusChallenged

	// ChannelHub.createChannel can emit ChannelCheckpointed before ChannelCreated for an
	// initial non-deposit/non-withdraw state. If this checkpoint is processed while the
	// local row is still the Void seed from CreateChannel, the on-chain checkpoint is
	// sufficient evidence that the channel has been materialized: promote Void to Open
	// here instead of leaving a bumped state_version on a Void channel until the later
	// ChannelCreated event replays. That replay then no-ops via the Status >= Open guard
	// in HandleHomeChannelCreated.
	if wasVoid {
		channel.Status = core.ChannelStatusOpen
	}

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
	// flows treat it as fully co-signed. The same applies to a Void→Open promotion: a
	// concurrent RPC may have stored an unsigned receiver head while the channel was
	// still Void (mirrors HandleHomeChannelCreated). On normal Open→Open checkpoints the
	// head row is already node-signed via the RPC path and this is a no-op.
	if wasChallenged || wasVoid {
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
// untouched: the user must countersign and acknowledge it via the regular RPC flow.
//
// Called from two contexts:
//   - challenge clearance (HandleHomeChannelCheckpointed): the head is an unsigned
//     receiver credit accumulated during the dispute window.
//   - channel open (HandleHomeChannelCreated): the head is an unsigned receiver credit
//     stored by a concurrent RPC while the channel was still Void.
func (s *EventHandlerService) backfillOffChainHeadNodeSig(ctx context.Context, tx core.ChannelHubEventHandlerStore, channelID string) error {
	head, err := tx.GetLastStateByChannelID(channelID, false)
	if err != nil {
		return err
	}
	if head == nil || head.NodeSig != nil {
		return nil
	}
	// Only receiver credits (transfer_receive, release) should appear as unsigned heads
	// in either call context. Any other transition kind means an invariant broke upstream;
	// do not silently node-sign it — log and bail so the caller surfaces the inconsistency.
	if head.Transition.Type != core.TransitionTypeTransferReceive &&
		head.Transition.Type != core.TransitionTypeRelease {
		log.FromContext(ctx).Warn("off-chain head is not a receiver state, skipping node-sig backfill",
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
	// Acquire the user's balance-row lock and read the channel under it before mutating
	// channel status. Receiver issuance paths (issueTransferReceiverState /
	// issueReleaseReceiverState) lock the same row up front and then re-check Status via
	// CheckActiveChannel; without this lock an in-flight RPC can read Status=Open,
	// node-sign a receiver state, and commit after we flip to Challenged — leaving a
	// node-signed higher-version receiver state on a disputed channel. The lock+read is a
	// single store call so the channel snapshot is consistent with the lock. See
	// HandleHomeChannelClosed for the same pattern.
	channel, err := tx.LockUserStateForHomeChannel(chanID)
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
// Additionally, when the channel was locally Challenged at the time of close, the handler
// issues a ChallengeRescue state crediting the user the net receiver-minus-sender balance
// accrued strictly above the closure version. The rescue runs unconditionally on the
// Challenged → Closed transition: both path-1 (timeout on a stale candidate) and path-2
// (cooperative close on a signed Finalize) routes pass through this branch, and the
// constructor + prev source pick the correct placement for the rescue row.
//
// MF3-I01 recovery anchor. This handler is the single recovery point for the wedge
// state described in audit finding MF3-I01: a receiver credit issued during the
// challenge window inherits HomeChannelID from currentState via NextState(), so the
// user's latest stored state can transiently point at a channel that closes via path-1
// before the next receiver-credit issuance reads currentState again. The listener
// ordering & idempotency invariant (pkg/blockchain/evm/listener.go, see processEvents
// doc) guarantees HandleHomeChannelChallenged has already run for any path-1 close, so
// wasChallenged is true here and the rescue advances the user past the closed channel.
// Subsequent receiver-credit issuance reads the rescue row as currentState and no
// longer carries the closed channel reference, so request_creation can reopen on the
// same (wallet, asset) through the normal flow.
func (s *EventHandlerService) HandleHomeChannelClosed(ctx context.Context, tx core.ChannelHubEventHandlerStore, event *core.HomeChannelClosedEvent) error {
	logger := log.FromContext(ctx)
	chanID := event.ChannelID
	// Acquire the user's balance-row lock and read the channel under it before mutating
	// channel status or summing receiver credits. issueTransferReceiverState /
	// issueReleaseReceiverState lock the same row up front, so this serializes the close
	// against any in-flight RPC receiver-issuance for the same user: either the RPC commits
	// its unsigned row before we sum (it lands in the rescue), or it blocks until we set the
	// channel to Closed and then sees that via its own re-check. The lock+read is a single
	// store call so the channel snapshot is consistent with the lock.
	channel, err := tx.LockUserStateForHomeChannel(chanID)
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

	// Drop stale Closed events that would regress state_version (MF3-L19).
	if guardEventVersionMonotonic(ctx, logger, chanID, "closed", event.StateVersion, channel.StateVersion) {
		return nil
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
// challenged-channel close. The state advances the user's chain past the closed
// channel so future receiver issuance and channels.v1.request_creation no longer
// wedge on it.
//
// Amount: NET effect on the user's home-channel balance of transitions stored
// strictly above closureVersion — receives (TransferReceive, Release) credit the
// user, sends (TransferSend, Commit) debit; everything else is excluded because it
// requires onchain backing the chain didn't enforce or belongs to a different
// ledger. Signed (Open-time) and unsigned (during-challenge) rows both contribute.
// Clamped at zero so an adversarial close at a version where the user's own balance
// was higher than the off-chain head can't dock the user further.
//
// Placement: prev is the user's latest state (across both channel-attached and
// detached rows). When prev is the in-channel head, the rescue wraps to a fresh
// epoch at (E+1, 0). When prev is a detached tip — the case where a node-signed
// Finalize already advanced the user via NextState() and post-Finalize receiver
// credits live at (E+1, v=0..M) with HomeChannelID nil — the rescue appends at
// (E+1, M+1), inheriting prev's ledger. NewChallengeRescueState picks the branch.
//
// AccountID on the rescue transition is the closed channel's ID; the rescue row
// itself has HomeChannelID nil — the shape of a credit to a user with no open home
// channel, to be folded into a signed state when the user next opens one.
func (s *EventHandlerService) issueChallengeRescue(ctx context.Context, tx core.ChannelHubEventHandlerStore, channel *core.Channel, closureVersion uint64) error {
	logger := log.FromContext(ctx)

	// Strict `>` against closureVersion: the row at the closure version itself is the
	// closing state and must be excluded — only transitions issued strictly after the
	// dispute version are unenforced. A channel's in-channel rows live at a single
	// epoch; detached post-Finalize rows have HomeChannelID NULL and are already
	// excluded by the channel_id predicate, so no epoch filter is needed.
	net, err := tx.SumNetTransitionAmountAfterVersion(channel.ChannelID, closureVersion)
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
		)
		total = decimal.Zero
	}

	// prev is the user's latest state across both channel-attached and detached rows.
	// When a node-signed Finalize already advanced the user via NextState() at sign
	// time, prev is the detached tip and the rescue appends after it. Otherwise prev
	// is the channel's own head and the rescue wraps to a fresh epoch.
	prev, err := tx.GetLastUserState(channel.UserWallet, channel.Asset, false)
	if err != nil {
		return err
	}
	if prev == nil {
		// Should not happen for a channel that reached Challenged → Closed: at least the
		// closure state itself must be on file. Surface the inconsistency rather than
		// silently dropping the rescue.
		return fmt.Errorf("no state found for closed challenged channel %s", channel.ChannelID)
	}

	rescue, err := core.NewChallengeRescueState(*prev, channel.ChannelID, total)
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

	if guardEventVersionMonotonic(ctx, logger, chanID, "escrow_deposit_initiated", event.StateVersion, channel.StateVersion) {
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

	if guardEventVersionMonotonic(ctx, logger, chanID, "escrow_deposit_finalized", event.StateVersion, channel.StateVersion) {
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

	if guardEventVersionMonotonic(ctx, logger, chanID, "escrow_withdrawal_initiated", event.StateVersion, channel.StateVersion) {
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

	if guardEventVersionMonotonic(ctx, logger, chanID, "escrow_withdrawal_finalized", event.StateVersion, channel.StateVersion) {
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
