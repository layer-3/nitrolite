// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {ChannelStatus, State, StateIntent} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {WadMath} from "./WadMath.sol";

/**
 * @title ChannelEngine
 * @notice Unified validation and calculation engine for all channel state transitions
 * @dev REQUIRES `state.homeLedger` to ALWAYS point to this (execution) chain. Otherwise, delta calculations will be incorrect.
 */
library ChannelEngine {
    using SafeCast for int256;
    using SafeCast for uint256;
    using WadMath for uint256;
    using WadMath for int256;

    error IncorrectHomeChainId();
    error IncorrectNonHomeChainId();
    error NonHomeStateRequired();
    error NonHomeStateMustBeEmpty();

    error IncorrectStateIntent();
    error IncorrectPreviousStateIntent();
    error IncorrectChannelStatus();
    error IncorrectStateVersion();
    error ChallengeExpired();

    error IncorrectUserAllocation();
    error IncorrectNodeAllocation();
    error IncorrectUserNetFlow();
    error IncorrectNodeNetFlow();

    error IncorrectUserNetFlowDelta();
    error IncorrectNodeNetFlowDelta();
    error NegativeNetFlowSum();
    error IncorrectAllocationSum();

    error AllocationExceedsLockedFunds();
    error UserAllocationDeltaMismatch();
    error UserNodeAllocationsMismatch();

    error InsufficientLockedFunds();
    error LockedFundsConsistencyViolation();
    error InsufficientNodeBalance();

    // ========== Structs ==========

    struct TransitionContext {
        ChannelStatus status;
        State prevState;
        uint256 lockedFunds;
        uint256 nodeAvailableFunds;
        uint64 challengeExpiry;
    }

    struct TransitionEffects {
        // Fund movements (positive = pull/lock, negative = push/release)
        int256 userFundsDelta; // Funds to pull from user (>0) or push to user (<0)
        int256 nodeFundsDelta; // Funds to lock from node vault (>0) or release (<0)

        // State updates
        ChannelStatus newStatus;
        uint64 newChallengeExpiry;
        bool updateLastState;
        bool closeChannel;
    }

    // ========== Public Functions ==========

    /**
     * @notice Unified validation and calculation for all channel state transitions
     * @dev Three phases: universal validation → intent-specific calculation → universal invariants
     * @param ctx Current channel context from storage
     * @param candidate New state to transition to
     * @return effects The calculated effects to apply
     */
    function validateTransition(TransitionContext memory ctx, State memory candidate)
        external
        view
        returns (TransitionEffects memory effects)
    {
        // Phase 1: Universal validation
        _validateUniversal(ctx, candidate);

        // Phase 2: Intent-specific calculation
        effects = _calculateEffectsByIntent(ctx, candidate);

        // Phase 3: Universal invariants
        _validateInvariants(ctx, candidate, effects);

        return effects;
    }

    // ========== Internal: Phase 1 - Universal Validation ==========

    function _validateUniversal(TransitionContext memory ctx, State memory candidate) internal view {
        // homeLedger always represents current chain
        require(candidate.homeLedger.chainId == block.chainid, IncorrectHomeChainId());
        require(candidate.version > ctx.prevState.version || Utils.isEmpty(ctx.prevState), IncorrectStateVersion());

        // Validate token decimals for homeLedger
        Utils.validateTokenDecimals(candidate.homeLedger);

        // Cross-chain escrow and migration operations require nonHomeLedger
        if (
            candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT
                || candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT
                || candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL
                || candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL
                || candidate.intent == StateIntent.INITIATE_MIGRATION
                || candidate.intent == StateIntent.FINALIZE_MIGRATION
        ) {
            require(!Utils.isEmpty(candidate.nonHomeLedger), NonHomeStateRequired());
            require(candidate.nonHomeLedger.chainId != block.chainid, IncorrectNonHomeChainId());
        } else {
            require(Utils.isEmpty(candidate.nonHomeLedger), NonHomeStateMustBeEmpty());
        }

        uint256 allocsSum = candidate.homeLedger.userAllocation + candidate.homeLedger.nodeAllocation;
        int256 netFlowsSum = candidate.homeLedger.userNetFlow + candidate.homeLedger.nodeNetFlow;

        require(netFlowsSum >= 0, NegativeNetFlowSum());
        require(allocsSum == netFlowsSum.toUint256(), IncorrectAllocationSum());

        // If channel is DISPUTED, check that challenge hasn't expired
        if (ctx.status == ChannelStatus.DISPUTED) {
            require(block.timestamp <= ctx.challengeExpiry, ChallengeExpired());
        }
    }

    // ========== Internal: Phase 2 - Intent-Specific Calculation ==========

    function _calculateEffectsByIntent(TransitionContext memory ctx, State memory candidate)
        internal
        view
        returns (TransitionEffects memory effects)
    {
        int256 userNfDelta = candidate.homeLedger.userNetFlow - ctx.prevState.homeLedger.userNetFlow;
        int256 nodeNfDelta = candidate.homeLedger.nodeNetFlow - ctx.prevState.homeLedger.nodeNetFlow;

        StateIntent intent = candidate.intent;

        if (intent == StateIntent.DEPOSIT) {
            effects = _calculateDepositEffects(ctx, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.WITHDRAW) {
            effects = _calculateWithdrawEffects(ctx, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.OPERATE) {
            effects = _calculateOperateEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.CLOSE) {
            effects = _calculateCloseEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_ESCROW_DEPOSIT) {
            effects = _calculateInitiateEscrowDepositEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_ESCROW_DEPOSIT) {
            effects = _calculateFinalizeEscrowDepositEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL) {
            effects = _calculateInitiateEscrowWithdrawalEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL) {
            effects = _calculateFinalizeEscrowWithdrawalEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_MIGRATION) {
            effects = _calculateInitiateMigrationEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_MIGRATION) {
            effects = _calculateFinalizeMigrationEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else {
            revert IncorrectStateIntent();
        }

        effects.updateLastState = true;
        return effects;
    }

    function _calculateDepositEffects(TransitionContext memory ctx, int256 userNfDelta, int256 nodeNfDelta)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // DEPOSIT-specific validations
        require(
            ctx.status == ChannelStatus.VOID || ctx.status == ChannelStatus.OPERATING
                || ctx.status == ChannelStatus.DISPUTED || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta > 0, IncorrectUserNetFlowDelta());

        // Calculate effects
        effects.userFundsDelta = userNfDelta; // Pull deposit from user
        effects.nodeFundsDelta = nodeNfDelta; // May lock more from node or release
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateWithdrawEffects(TransitionContext memory ctx, int256 userNfDelta, int256 nodeNfDelta)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // WITHDRAW-specific validations
        require(
            ctx.status == ChannelStatus.VOID || ctx.status == ChannelStatus.OPERATING
                || ctx.status == ChannelStatus.DISPUTED || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta < 0, IncorrectUserNetFlowDelta());

        // Calculate effects
        effects.userFundsDelta = userNfDelta; // Negative = push to user
        effects.nodeFundsDelta = nodeNfDelta;
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateOperateEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // OPERATE-specific validations (checkpoint)
        require(
            ctx.status == ChannelStatus.VOID || ctx.status == ChannelStatus.OPERATING
                || ctx.status == ChannelStatus.DISPUTED || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta == 0, IncorrectUserNetFlowDelta());
        require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateCloseEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // CLOSE-specific validations
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );

        uint256 allocsSum = candidate.homeLedger.userAllocation + candidate.homeLedger.nodeAllocation;
        require(allocsSum <= ctx.lockedFunds, AllocationExceedsLockedFunds());

        // Ensure final locked funds will be sufficient for special nodeAllocation handling
        int256 finalLockedFunds = ctx.lockedFunds.toInt256() + userNfDelta + nodeNfDelta;
        require(finalLockedFunds >= 0, InsufficientLockedFunds());
        require(finalLockedFunds >= candidate.homeLedger.nodeAllocation.toInt256(), InsufficientLockedFunds());

        // Calculate effects
        // Push allocations to parties (negative = push out from channel)
        effects.userFundsDelta = userNfDelta;
        effects.nodeFundsDelta = nodeNfDelta;
        effects.newStatus = ChannelStatus.CLOSED;
        effects.newChallengeExpiry = 0;
        effects.closeChannel = true;

        return effects;
    }

    function _calculateInitiateEscrowDepositEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // INITIATE_ESCROW_DEPOSIT-specific validations (Home Chain)
        // Node locks liquidity in channel for cross-chain deposit
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta == 0, IncorrectUserNetFlowDelta()); // no user funds movement
        // node fund movement may accommodate transfers, thus it can be both positive or negative
        // node allocation may not have changed if previous operation is also initiate escrow deposit

        // Check home - non-home state consistency
        uint256 depositAmount = candidate.nonHomeLedger.userAllocation;
        require(depositAmount > 0, IncorrectUserAllocation());
        require(
            candidate.homeLedger.nodeAllocation.toWad(candidate.homeLedger.decimals)
                == depositAmount.toWad(candidate.nonHomeLedger.decimals),
            IncorrectNodeAllocation()
        );
        require(candidate.nonHomeLedger.userNetFlow == depositAmount.toInt256(), IncorrectUserNetFlow());

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateFinalizeEscrowDepositEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // FINALIZE_ESCROW_DEPOSIT-specific validations (Home Chain)
        // Previous on-chain state MUST be INITIATE_ESCROW_DEPOSIT
        // Funds stay in channel, just move from node allocation to user allocation
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );

        // Check home - non-home state consistency
        require(ctx.prevState.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, IncorrectPreviousStateIntent());
        require(candidate.version == ctx.prevState.version + 1, IncorrectStateVersion());
        require(candidate.nonHomeLedger.userAllocation == 0, IncorrectUserAllocation());
        require(candidate.nonHomeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

        require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());
        // nothing changes from initiate escrow deposit state
        require(userNfDelta == 0, IncorrectUserNetFlowDelta());
        require(nodeNfDelta == 0, IncorrectNodeNetFlowDelta());

        uint256 depositAmount = ctx.prevState.nonHomeLedger.userAllocation;
        require(candidate.nonHomeLedger.userNetFlow == depositAmount.toInt256(), IncorrectUserNetFlow());
        require(candidate.nonHomeLedger.nodeNetFlow == -depositAmount.toInt256(), IncorrectNodeNetFlow());

        uint256 userAllocDelta = candidate.homeLedger.userAllocation - ctx.prevState.homeLedger.userAllocation;
        require(
            userAllocDelta.toWad(candidate.homeLedger.decimals)
                == depositAmount.toWad(ctx.prevState.nonHomeLedger.decimals),
            UserAllocationDeltaMismatch()
        );

        // Calculate effects - funds stay in channel, no external movement
        effects.userFundsDelta = 0;
        effects.nodeFundsDelta = 0;
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateInitiateEscrowWithdrawalEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // INITIATE_ESCROW_WITHDRAWAL-specific validations (Home Chain)
        // Previous on-chain state can be anything, so validate like an OPERATE state + non-home State
        // Prepare for cross-chain withdrawal (state validation only)
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta == 0, IncorrectUserNetFlowDelta()); // no user funds movement
        require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

        // Check home - non-home state consistency
        require(candidate.nonHomeLedger.userNetFlow == 0, IncorrectUserNetFlow());
        require(candidate.nonHomeLedger.userAllocation == 0, IncorrectUserAllocation());
        require(
            candidate.nonHomeLedger.nodeAllocation.toInt256() == candidate.nonHomeLedger.nodeNetFlow,
            IncorrectNodeNetFlow()
        );

        // Calculate effects - no immediate fund movement
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateFinalizeEscrowWithdrawalEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // FINALIZE_ESCROW_WITHDRAWAL-specific validations (Home Chain)
        // Previous on-chain state can be anything, so validate like an OPERATE state + non-home State
        // Decrease user allocation after cross-chain withdrawal completes
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            IncorrectChannelStatus()
        );
        require(userNfDelta == 0, IncorrectUserNetFlowDelta()); // no user funds movement
        require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

        // Check home - non-home state consistency
        require(candidate.nonHomeLedger.userAllocation == 0, IncorrectUserAllocation());
        require(candidate.nonHomeLedger.nodeAllocation == 0, IncorrectNodeAllocation());
        require(candidate.nonHomeLedger.userNetFlow == -candidate.nonHomeLedger.nodeNetFlow, IncorrectUserNetFlow());

        // TODO: provide V-1 state (INITIATE_ESCROW_WITHDRAWAL) to validate against?

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.newChallengeExpiry = 0;

        return effects;
    }

    function _calculateInitiateMigrationEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal view returns (TransitionEffects memory effects) {
        // INITIATE_MIGRATION: Can be called on both home and non-home chain

        if (ctx.status == ChannelStatus.VOID || ctx.status == ChannelStatus.MIGRATED_OUT) {
            // NON-HOME CHAIN (IN): Create MIGRATING_IN channel
            // homeLedger represents new home (current chain)

            uint256 userNonHomeAlloc = candidate.nonHomeLedger.userAllocation;
            require(userNonHomeAlloc > 0, IncorrectUserAllocation());
            require(candidate.nonHomeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

            require(candidate.homeLedger.userAllocation == 0, IncorrectUserAllocation());
            require(
                candidate.homeLedger.nodeAllocation.toWad(candidate.homeLedger.decimals)
                    == userNonHomeAlloc.toWad(candidate.nonHomeLedger.decimals),
                UserNodeAllocationsMismatch()
            );
            require(
                candidate.homeLedger.nodeNetFlow.toWad(candidate.homeLedger.decimals)
                    == userNonHomeAlloc.toInt256().toWad(candidate.nonHomeLedger.decimals),
                IncorrectNodeNetFlow()
            );
            require(candidate.homeLedger.userNetFlow == 0, IncorrectUserNetFlow());

            // Calculate effects - lock node funds
            // No delta calculation needed - creating fresh channel
            effects.nodeFundsDelta = candidate.homeLedger.nodeAllocation.toInt256();
            effects.newStatus = ChannelStatus.MIGRATING_IN;
        } else if (ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED) {
            // HOME CHAIN (OUT): Update state
            require(candidate.homeLedger.chainId == block.chainid, IncorrectHomeChainId());

            require(userNfDelta == 0, IncorrectUserNetFlowDelta()); // no user funds movement

            // Validate homeLedger (current chain)
            uint256 userHomeAlloc = candidate.homeLedger.userAllocation;
            require(userHomeAlloc > 0, IncorrectUserAllocation());
            require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

            // Validate nonHomeLedger (target chain)
            require(candidate.nonHomeLedger.userAllocation == 0, IncorrectUserAllocation());
            require(
                candidate.nonHomeLedger.nodeAllocation.toWad(candidate.nonHomeLedger.decimals)
                    == userHomeAlloc.toWad(candidate.homeLedger.decimals),
                UserNodeAllocationsMismatch()
            );
            require(
                candidate.nonHomeLedger.nodeNetFlow.toWad(candidate.nonHomeLedger.decimals)
                    == userHomeAlloc.toInt256().toWad(candidate.homeLedger.decimals),
                IncorrectNodeNetFlow()
            );
            require(candidate.nonHomeLedger.userNetFlow == 0, IncorrectUserNetFlow());

            // Calculate effects - may adjust node vault based on net flow delta
            effects.nodeFundsDelta = nodeNfDelta;
            effects.newStatus = ChannelStatus.OPERATING;
            effects.newChallengeExpiry = 0;
        } else {
            revert IncorrectChannelStatus();
        }

        return effects;
    }

    function _calculateFinalizeMigrationEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal view returns (TransitionEffects memory effects) {
        // FINALIZE_MIGRATION: Can be called on both new home and old home chain

        if (ctx.status == ChannelStatus.MIGRATING_IN) {
            // NEW HOME CHAIN (IN): Move MIGRATING_IN → OPERATING
            // The homeLedger represents the new home (current chain) in candidate and prevState
            require(candidate.homeLedger.chainId == block.chainid, IncorrectHomeChainId());
            require(ctx.prevState.intent == StateIntent.INITIATE_MIGRATION, IncorrectPreviousStateIntent());
            require(candidate.version == ctx.prevState.version + 1, IncorrectStateVersion());

            // nonHomeLedger = old home (holds the user's migrated allocation)
            uint256 userMigratedAlloc = ctx.prevState.nonHomeLedger.userAllocation;

            // Validate that this completes the migration: user receives their migrated allocation on new home
            require(candidate.homeLedger.userAllocation == userMigratedAlloc, IncorrectUserAllocation());
            require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());
            require(candidate.nonHomeLedger.userAllocation == 0, IncorrectUserAllocation());
            require(candidate.nonHomeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

            // Special delta calculation: previous state was swapped during INITIATE_MIGRATION
            // So prevState.homeLedger represents new home (current chain)
            // Calculate deltas normally - no special handling needed since state was swapped on storage
            require(userNfDelta == 0, IncorrectUserNetFlowDelta());
            require(nodeNfDelta == 0, IncorrectNodeNetFlowDelta());

            // Calculate effects - just status change
            effects.newStatus = ChannelStatus.OPERATING;
        } else if (ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED) {
            // OLD HOME CHAIN (OUT): Release funds and move to MIGRATED_OUT
            // homeLedger represents old home (current chain)

            // Validate homeLedger
            require(candidate.homeLedger.userAllocation == 0, IncorrectUserAllocation());
            require(candidate.homeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

            // Validate nonHomeLedger (new home)
            require(candidate.nonHomeLedger.userAllocation > 0, IncorrectUserAllocation());
            require(candidate.nonHomeLedger.nodeAllocation == 0, IncorrectNodeAllocation());

            // Calculate effects - release all currently locked funds to node vault
            effects.nodeFundsDelta = nodeNfDelta;
            effects.newStatus = ChannelStatus.MIGRATED_OUT;
            effects.newChallengeExpiry = 0;
            effects.closeChannel = true;
        } else {
            revert IncorrectChannelStatus();
        }

        return effects;
    }

    // ========== Internal: Phase 3 - Universal Invariants ==========

    function _validateInvariants(TransitionContext memory ctx, State memory candidate, TransitionEffects memory effects)
        internal
        pure
    {
        int256 expectedLocked = ctx.lockedFunds.toInt256() + effects.userFundsDelta + effects.nodeFundsDelta;
        require(expectedLocked >= 0, InsufficientLockedFunds());

        // Check that allocations equal expected locked funds (unless deleting)
        if (!effects.closeChannel) {
            uint256 allocsSum = candidate.homeLedger.userAllocation + candidate.homeLedger.nodeAllocation;
            require(allocsSum == expectedLocked.toUint256(), LockedFundsConsistencyViolation());
        }

        // Check node has sufficient funds for positive nodeNfDelta
        if (effects.nodeFundsDelta > 0) {
            require(ctx.nodeAvailableFunds >= effects.nodeFundsDelta.toUint256(), InsufficientNodeBalance());
        }
    }
}
