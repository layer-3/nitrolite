// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Challenge_Base} from "./ChannelHub_Challenge_Base.t.sol";

// forge-lint: disable-start(unsafe-typecast)

import {Utils} from "../../src/Utils.sol";
import {TestUtils} from "../TestUtils.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    ChannelStatus,
    ParticipantIndex
} from "../../src/interfaces/Types.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {ChannelEngine} from "../../src/ChannelEngine.sol";

/*
 * @dev This file uses integration / blackbox testing through ChannelHub to verify
 *    critical end-to-end challenge flows (signature validation, fund movements, storage updates, events).
 *    Complex state machine logic and edge cases are tested exhaustively in dedicated engine unit tests
 *    (ChannelEngine.t.sol, EscrowDepositEngine.t.sol, EscrowWithdrawalEngine.t.sol) for faster execution
 *    and better isolation.
 */
contract ChannelHubTest_Challenge_HomeChain_NormalOperation is ChannelHubTest_Challenge_Base {
    /*
    Test cases:
    - a channel can be challenged with a newer state, which is enforced during challenge
    - a channel can be challenged with existing state, which is NOT enforced the second time during challenge
    - a channel can NOT be challenged with CLOSE intent (must use closeChannel function for that)
    - challenge is finalized (funds can be withdrawn) after `challengeExpireAt` time expires
    - challenged "operating" state can be resolved with a newer state until `challengeExpireAt` time has NOT passed
    - challenged state can NOT be resolved after `challengeExpireAt` time has passed
    - a channel can NOT be challenged again during a challenge
    - a channel can NOT be challenged with an earlier state
    - a non-yet-on-chain channel can NOT be challenged
    */

    function setUp() public override {
        super.setUp();
        createChannelWithDeposit();
    }

    function test_challengeWithNewerState_enforcesState() public {
        // Off-chain: user transfers 100 to node
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Off-chain: user transfers another 50 to node
        State memory stateV2 =
            TestUtils.nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Node challenges with newer state V2, which should be enforced during challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV2, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV2, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            2,
            block.timestamp + CHALLENGE_DURATION,
            "State V2 should be enforced during challenge"
        );
        verifyChannelState(
            channelId,
            [uint256(850), uint256(0)],
            [int256(1000), int256(-150)],
            "State V2 should be enforced during challenge"
        );
    }

    function test_challengeWithExistingState_notEnforcedAgain() public {
        // Checkpoint a new state
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);

        // Verify state V1 is on-chain
        (,, State memory latestStateBefore,,) = cHub.getChannelData(channelId);
        assertEq(latestStateBefore.version, 1, "State version should be 1 before challenge");

        // Node challenges with the same state V1 (already on-chain)
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            1,
            block.timestamp + CHALLENGE_DURATION,
            "State V1 should be enforced during challenge"
        );
        verifyChannelState(
            channelId,
            [uint256(900), uint256(0)],
            [int256(1000), int256(-100)],
            "State V1 should be enforced during challenge"
        );
    }

    function test_revert_challengeWithCloseIntent() public {
        State memory closeState =
            TestUtils.nextState(initState, StateIntent.CLOSE, [uint256(0), uint256(0)], [int256(0), int256(0)]);
        closeState = mutualSignStateBothWithEcdsaValidator(closeState, channelId, ALICE_PK);

        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, closeState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.challengeChannel(channelId, closeState, challengerSig, ParticipantIndex.NODE);
    }

    function test_challengeFinalization_afterTimeout() public {
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge with current state
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        vm.warp(block.timestamp + CHALLENGE_DURATION + 1);

        uint256 aliceBalanceBefore = token.balanceOf(alice);
        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));

        // Finalize challenge by closing the channel (unilateral closure)
        // When doing unilateral closure after timeout, any state works
        vm.prank(alice);
        cHub.closeChannel(channelId, initState);

        // Verify channel is CLOSED and funds were distributed according to last enforced state (V1)
        verifyChannelData(
            channelId, ChannelStatus.CLOSED, 1, 0, "Channel should be CLOSED after challenge finalization"
        );

        uint256 aliceBalanceAfter = token.balanceOf(alice);
        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token));

        assertEq(aliceBalanceAfter, aliceBalanceBefore + 900, "Alice should receive her allocation");
        // Node balance should remain unchanged because:
        // 1. The node already received its 100 when the challenge was processed (nodeNetFlow -100 released funds)
        // 2. During unilateral closure, node gets nodeAllocation (0)
        assertEq(
            nodeBalanceAfter,
            nodeBalanceBefore,
            "Node balance should remain unchanged (already received net flow during challenge)"
        );
    }

    function test_resolveChallenge_withNewerState_beforeTimeout() public {
        // State V1: user transfers 100
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge with stateV1
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            1,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED after challenge"
        );

        // State V2: user transfers another 50 (newer state to resolve challenge)
        State memory stateV2 =
            TestUtils.nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Resolve challenge by checkpointing newer state (before timeout)
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV2);

        verifyChannelData(
            channelId,
            ChannelStatus.OPERATING,
            2,
            0,
            "Channel should be OPERATING after resolving challenge with newer state"
        );
        verifyChannelState(
            channelId,
            [uint256(850), uint256(0)],
            [int256(1000), int256(-150)],
            "State V2 should be enforced after resolving challenge with newer state"
        );
    }

    function test_revert_resolveChallenge_withOlderState_beforeTimeout() public {
        // State V1: user transfers 100
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // State V2: user receives 50 back
        State memory stateV2 =
            TestUtils.nextState(stateV1, StateIntent.OPERATE, [uint256(950), uint256(0)], [int256(1000), int256(-50)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Challenge with stateV2
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV2, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV2, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            2,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED after challenge"
        );

        // Try to resolve with older state V1 (should fail)
        vm.expectRevert(ChannelEngine.IncorrectStateVersion.selector);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);
    }

    function test_revert_resolveChallenge_withNewerState_afterTimeout() public {
        // State V1
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        vm.warp(block.timestamp + CHALLENGE_DURATION + 1);

        // State V2: user transfers another 50 (newer state to resolve challenge)
        State memory stateV2 =
            TestUtils.nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Cannot resolve challenge after timeout - must close channel instead
        vm.expectRevert(ChannelEngine.ChallengeExpired.selector);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV2);
    }

    function test_revert_challengeAlreadyChallengedChannel() public {
        // First challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            0,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED after first challenge"
        );

        // Try to challenge again (should fail)
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        bytes memory challengerSig2 = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.challengeChannel(channelId, stateV1, challengerSig2, ParticipantIndex.NODE);
    }

    function test_revert_challengeWithOlderState() public {
        // State V1
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Checkpoint V1
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);

        // Try to challenge with older state (initial) (should fail)
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.ChallengerVersionTooLow.selector);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);
    }

    function test_revert_challengeNonExistingChannel() public {
        ChannelDefinition memory newDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE + 42,
            approvedSignatureValidators: 0,
            metadata: bytes32("42")
        });
        bytes32 newChannelId = Utils.getChannelId(newDef, CHANNEL_HUB_VERSION);

        // Off-chain: user transfers 100 to node
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, newChannelId, ALICE_PK);

        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(newChannelId, stateV1, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.challengeChannel(newChannelId, stateV1, challengerSig, ParticipantIndex.NODE);
    }
}

contract ChannelHubTest_Challenge_HomeChain_EscrowDeposit is ChannelHubTest_Challenge_Base {
    /*
    Test cases:
    - a channel can be challenged with a newer state, which is enforced during challenge:
        (new: InitiateEscrowDeposit, FinalizeEscrowDeposit)
    - a channel can be challenged with existing state, which is NOT enforced the second time during challenge:
        (existing: InitiateEscrowDeposit, FinalizeEscrowDeposit)
    - a challenged channel can be resolved with "InitiateEscrowDeposit" / "FinalizeEscrowDeposit" state until `challengeExpireAt` time has NOT passed
    */

    bytes32 escrowId;

    uint64 initiateEscrowDepositVersion = 1;
    State initiateEscrowDepositState;
    uint64 finalizeEscrowDepositVersion = 2;
    State finalizeEscrowDepositState;

    function setUp() public override {
        super.setUp();
        createChannelWithDeposit();

        initiateEscrowDepositState = TestUtils.nextState(
            initState,
            StateIntent.INITIATE_ESCROW_DEPOSIT,
            [uint256(1000), uint256(500)],
            [int256(1000), int256(500)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(500), uint256(0)],
            [int256(500), int256(0)]
        );
        initiateEscrowDepositState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowDepositState, channelId, ALICE_PK);

        escrowId = Utils.getEscrowId(channelId, initiateEscrowDepositVersion);

        finalizeEscrowDepositState = TestUtils.nextState(
            initiateEscrowDepositState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(1500), uint256(0)],
            [int256(1000), int256(500)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(0), uint256(0)],
            [int256(500), int256(-500)]
        );
        finalizeEscrowDepositState =
            mutualSignStateBothWithEcdsaValidator(finalizeEscrowDepositState, channelId, ALICE_PK);
    }

    function test_challenge_initiateEscrowDeposit_asNew() public {
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and initiateEscrowDepositState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateEscrowDepositVersion,
            block.timestamp + CHALLENGE_DURATION,
            "InitiateEscrowDepositState should be enforced"
        );
        verifyChannelState(
            channelId,
            [uint256(1000), uint256(500)],
            [int256(1000), int256(500)],
            "InitiateEscrowDepositState should be enforced"
        );
    }

    function test_challenge_initiateEscrowDeposit_asExisting() public {
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Challenge with already enforced initiateEscrowDepositState state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Verify state is still initiateEscrowDepositState
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateEscrowDepositVersion,
            block.timestamp + CHALLENGE_DURATION,
            "State should not be re-enforced"
        );
        verifyChannelState(
            channelId, [uint256(1000), uint256(500)], [int256(1000), int256(500)], "State should not be re-enforced"
        );
    }

    function test_challenge_initiateEscrowDeposit_resolve() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with newer initiateEscrowDepositState state (before timeout)
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, initiateEscrowDepositVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(1000), uint256(500)],
            [int256(1000), int256(500)],
            "initiateEscrowDepositState should be enforced"
        );
    }

    function test_challenge_finalizeEscrowDeposit_asNew() public {
        // First enforce INITIATE_ESCROW_DEPOSIT on-chain (required for FINALIZE to be valid)
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Now challenge with FINALIZE_ESCROW_DEPOSIT
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, finalizeEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, finalizeEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and finalizeEscrowDepositState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            finalizeEscrowDepositVersion,
            block.timestamp + CHALLENGE_DURATION,
            "FinalizeEscrowDepositState should be enforced"
        );
        verifyChannelState(
            channelId,
            [uint256(1500), uint256(0)],
            [int256(1000), int256(500)],
            "finalizeEscrowDepositState should be enforced"
        );
    }

    function test_challenge_finalizeEscrowDeposit_asExisting() public {
        // First enforce INITIATE_ESCROW_DEPOSIT on-chain
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Then enforce FINALIZE_ESCROW_DEPOSIT on-chain
        vm.prank(alice);
        cHub.finalizeEscrowDeposit(channelId, escrowId, finalizeEscrowDepositState);

        // Challenge with already enforced finalizeEscrowDepositState state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, finalizeEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, finalizeEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Verify state is still finalizeEscrowDepositState
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            finalizeEscrowDepositVersion,
            block.timestamp + CHALLENGE_DURATION,
            "State should not be re-enforced"
        );
        verifyChannelState(
            channelId, [uint256(1500), uint256(0)], [int256(1000), int256(500)], "State should not be re-enforced"
        );
    }

    function test_challenge_finalizeEscrowDeposit_resolve() public {
        // First enforce INITIATE_ESCROW_DEPOSIT on-chain
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Challenge with older initiate state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with newer finalizeEscrowDepositState state (before timeout)
        vm.prank(alice);
        cHub.finalizeEscrowDeposit(channelId, escrowId, finalizeEscrowDepositState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, finalizeEscrowDepositVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(1500), uint256(0)],
            [int256(1000), int256(500)],
            "finalizeEscrowDepositState should be enforced"
        );
    }

    function test_finalizeEscrowDeposit_resolve_newlyChallenged_initializeEscrowDeposit() public {
        // Challenge with INITIATE_ESCROW_DEPOSIT state (without enforcing it on-chain first)
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowDepositState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with finalizeEscrowDepositState state (before timeout)
        vm.prank(alice);
        cHub.finalizeEscrowDeposit(channelId, escrowId, finalizeEscrowDepositState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, finalizeEscrowDepositVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(1500), uint256(0)],
            [int256(1000), int256(500)],
            "finalizeEscrowDepositState should be enforced"
        );
    }

    function test_revert_onChallengeEscrowDeposit() public {
        // First enforce INITIATE_ESCROW_DEPOSIT on-chain
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        // Challenge with INITIATE_ESCROW_DEPOSIT state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.NoChannelIdFoundForEscrow.selector);
        cHub.challengeEscrowDeposit(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}

contract ChannelHubTest_Challenge_HomeChain_EscrowWithdrawal is ChannelHubTest_Challenge_Base {
    /*
    Test cases:
    - a channel can be challenged with a newer state, which is enforced during challenge:
        (new: InitiateEscrowWithdrawal, FinalizeEscrowWithdrawal)
    - a channel can be challenged with existing state, which is NOT enforced the second time during challenge:
        (existing: InitiateEscrowWithdrawal, FinalizeEscrowWithdrawal)
    - a challenged channel can be resolved with "InitiateEscrowWithdrawal" / "FinalizeEscrowWithdrawal" state until `challengeExpireAt` time has NOT passed
    */

    bytes32 escrowId;

    uint64 initiateEscrowWithdrawalVersion = 1;
    State initiateEscrowWithdrawalState;
    uint64 finalizeEscrowWithdrawalVersion = 2;
    State finalizeEscrowWithdrawalState;

    function setUp() public override {
        super.setUp();
        createChannelWithDeposit();

        initiateEscrowWithdrawalState = TestUtils.nextState(
            initState,
            StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            [uint256(1000), uint256(0)],
            [int256(1000), int256(0)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(0), uint256(300)],
            [int256(0), int256(300)]
        );
        initiateEscrowWithdrawalState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowWithdrawalState, channelId, ALICE_PK);

        escrowId = Utils.getEscrowId(channelId, initiateEscrowWithdrawalVersion);

        finalizeEscrowWithdrawalState = TestUtils.nextState(
            initiateEscrowWithdrawalState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(0), uint256(0)],
            [int256(-300), int256(300)]
        );
        finalizeEscrowWithdrawalState =
            mutualSignStateBothWithEcdsaValidator(finalizeEscrowWithdrawalState, channelId, ALICE_PK);
    }

    function test_challenge_initiateEscrowWithdrawal_asNew() public {
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowWithdrawalState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and initiateEscrowWithdrawalState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateEscrowWithdrawalVersion,
            block.timestamp + CHALLENGE_DURATION,
            "InitiateEscrowWithdrawalState should be enforced"
        );
        verifyChannelState(
            channelId,
            [uint256(1000), uint256(0)],
            [int256(1000), int256(0)],
            "InitiateEscrowWithdrawalState should be enforced"
        );
    }

    function test_challenge_initiateEscrowWithdrawal_asExisting() public {
        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initiateEscrowWithdrawalState);

        // Challenge with already enforced initiateEscrowWithdrawalState state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowWithdrawalState, challengerSig, ParticipantIndex.NODE);

        // Verify state is still initiateEscrowWithdrawalState
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateEscrowWithdrawalVersion,
            block.timestamp + CHALLENGE_DURATION,
            "State should not be re-enforced"
        );
        verifyChannelState(
            channelId, [uint256(1000), uint256(0)], [int256(1000), int256(0)], "State should not be re-enforced"
        );
    }

    function test_challenge_initiateEscrowWithdrawal_resolve() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with newer initiateEscrowWithdrawalState state (before timeout)
        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initiateEscrowWithdrawalState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, initiateEscrowWithdrawalVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(1000), uint256(0)],
            [int256(1000), int256(0)],
            "initiateEscrowWithdrawalState should be enforced"
        );
    }

    function test_challenge_finalizeEscrowWithdrawal_asNew() public {
        // INITIATE_ESCROW_WITHDRAWAL is NOT required to be enforced first on-chain

        // Challenge with FINALIZE_ESCROW_WITHDRAWAL
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, finalizeEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, finalizeEscrowWithdrawalState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and finalizeEscrowWithdrawalState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            finalizeEscrowWithdrawalVersion,
            block.timestamp + CHALLENGE_DURATION,
            "FinalizeEscrowWithdrawalState should be enforced"
        );
        verifyChannelState(
            channelId,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            "finalizeEscrowWithdrawalState should be enforced"
        );
    }

    function test_challenge_finalizeEscrowWithdrawal_asExisting() public {
        // INITIATE_ESCROW_WITHDRAWAL is NOT required to be enforced first on-chain

        // Enforce FINALIZE_ESCROW_WITHDRAWAL on-chain
        vm.prank(alice);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, finalizeEscrowWithdrawalState);

        // Challenge with already enforced finalizeEscrowWithdrawalState state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, finalizeEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, finalizeEscrowWithdrawalState, challengerSig, ParticipantIndex.NODE);

        // Verify state is still finalizeEscrowWithdrawalState
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            finalizeEscrowWithdrawalVersion,
            block.timestamp + CHALLENGE_DURATION,
            "State should not be re-enforced"
        );
        verifyChannelState(
            channelId, [uint256(700), uint256(0)], [int256(1000), int256(-300)], "State should not be re-enforced"
        );
    }

    function test_challenge_finalizeEscrowWithdrawal_resolve() public {
        // INITIATE_ESCROW_WITHDRAWAL is NOT required to be enforced first on-chain

        // Challenge with older initiate state
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with newer finalizeEscrowWithdrawalState state (before timeout)
        vm.prank(alice);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, finalizeEscrowWithdrawalState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, finalizeEscrowWithdrawalVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            "finalizeEscrowWithdrawalState should be enforced"
        );
    }

    function test_finalizeEscrowWithdrawal_resolve_newlyChallenged_initializeEscrowWithdrawal() public {
        // Challenge with INITIATE_ESCROW_WITHDRAWAL state (without enforcing it on-chain first)
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateEscrowWithdrawalState, challengerSig, ParticipantIndex.NODE);

        // Resolve challenge with finalizeEscrowWithdrawalState state (before timeout)
        vm.prank(alice);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, finalizeEscrowWithdrawalState);

        // Verify challenge was resolved
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, finalizeEscrowWithdrawalVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            "finalizeEscrowWithdrawalState should be enforced"
        );
    }

    function test_revert_onChallengeEscrowWithdrawal() public {
        // First enforce INITIATE_ESCROW_WITHDRAWAL on-chain
        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initiateEscrowWithdrawalState);

        // Challenge with INITIATE_ESCROW_WITHDRAWAL state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.NoChannelIdFoundForEscrow.selector);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}

contract ChannelHubTest_Challenge_HomeChain_HomeMigration is ChannelHubTest_Challenge_Base {
    /*
    Test cases:
    - a channel in Operate status can be challenged with initiated migration state
    - a channel challenged with "InitiateMigration" state can be checkpointed calling "finalizeMigration" (-> MigratedOut status)
    - a channel challenged with "InitiateMigration" state can be resolved with "operation" state
        (although this should not happen in practice since the node should finalize migration instead of resolving with an older state, but just to be safe)
    - a channel can NOT be challenged when in MIGRATED_OUT status
    - a channel can NOT be challenged with FINALIZE_MIGRATION intent on home chain (must use finalizeMigration function for that)
    */

    uint64 initiateMigrationVersion = 1;
    State initiateMigrationState;
    uint64 finalizeMigrationVersion = 2;
    State finalizeMigrationState;
    uint64 operateAfterMigrationInitVersion = 2;
    State operateAfterMigrationInitState;

    // New channel for testing NEW home chain behavior
    ChannelDefinition newHomeDef;
    bytes32 newHomeChannelId;
    State newHomeInitiateMigrationState;
    uint64 newHomeOperateVersion = 3;
    State newHomeOperateState;

    function setUp() public override {
        super.setUp();
        createChannelWithDeposit();

        // INITIATE_MIGRATION state:
        initiateMigrationState = TestUtils.nextState(
            initState,
            StateIntent.INITIATE_MIGRATION,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(0), uint256(700)], // Node locks user allocation on new home
            [int256(0), int256(700)]
        );
        initiateMigrationState = mutualSignStateBothWithEcdsaValidator(initiateMigrationState, channelId, ALICE_PK);

        // FINALIZE_MIGRATION state: Allocations zero out on old home, user receives allocation on new home
        finalizeMigrationState = TestUtils.nextState(
            initiateMigrationState,
            StateIntent.FINALIZE_MIGRATION,
            [uint256(0), uint256(0)], // Old home: allocations zero out
            [int256(1000), int256(-1000)], // Old home: net flows balance
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(700), uint256(0)], // New home: user receives allocation
            [int256(0), int256(700)]
        );
        // Swap home and non-home states as per migration protocol
        Ledger memory temp = finalizeMigrationState.homeLedger;
        finalizeMigrationState.homeLedger = finalizeMigrationState.nonHomeLedger;
        finalizeMigrationState.nonHomeLedger = temp;
        finalizeMigrationState = mutualSignStateBothWithEcdsaValidator(finalizeMigrationState, channelId, ALICE_PK);

        // OPERATE state after migration initiation (for resolving challenge)
        operateAfterMigrationInitState = TestUtils.nextState(
            initiateMigrationState, StateIntent.OPERATE, [uint256(650), uint256(0)], [int256(1000), int256(-350)]
        );
        operateAfterMigrationInitState =
            mutualSignStateBothWithEcdsaValidator(operateAfterMigrationInitState, channelId, ALICE_PK);
    }

    function test_challenge_initiateMigration_fromOperating() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initiateMigrationState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateMigrationState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and initiateMigrationState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateMigrationVersion,
            block.timestamp + CHALLENGE_DURATION,
            "InitiateMigrationState should be enforced"
        );
        verifyChannelState(
            channelId,
            [uint256(700), uint256(0)],
            [int256(1000), int256(-300)],
            "InitiateMigrationState should be enforced"
        );
    }

    function test_challenge_initiateMigration_resolve_withFinalizeMigration() public {
        // Challenge with INITIATE_MIGRATION
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initiateMigrationState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateMigrationState, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateMigrationVersion,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED"
        );

        // Resolve challenge with FINALIZE_MIGRATION (before timeout)
        vm.prank(alice);
        cHub.finalizeMigration(channelId, finalizeMigrationState);

        // Verify channel is MIGRATED_OUT and initiateMigrationState was enforced
        verifyChannelData(
            channelId,
            ChannelStatus.MIGRATED_OUT,
            finalizeMigrationVersion,
            0,
            "finalizeMigration should resolve the challenge"
        );
    }

    function test_challenge_initiateMigration_resolve_withOperate() public {
        // Challenge with INITIATE_MIGRATION
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initiateMigrationState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initiateMigrationState, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            initiateMigrationVersion,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED"
        );

        // Resolve challenge with newer OPERATE state (before timeout)
        // This is technically possible but shouldn't happen in practice as participants should NOT sign OPERATE state as direct successor of INITIATE_MIGRATION
        vm.prank(alice);
        cHub.checkpointChannel(channelId, operateAfterMigrationInitState);

        // Verify channel is back to OPERATING
        verifyChannelData(
            channelId, ChannelStatus.OPERATING, operateAfterMigrationInitVersion, 0, "Challenge should be resolved"
        );
        verifyChannelState(
            channelId,
            [uint256(650), uint256(0)],
            [int256(1000), int256(-350)],
            "operateAfterMigrationInitState should be enforced"
        );
    }

    function test_revert_challenge_migratedOut() public {
        // First initiate migration
        vm.prank(alice);
        cHub.initiateMigration(def, initiateMigrationState);

        // Then finalize migration to put channel in MIGRATED_OUT status
        vm.prank(alice);
        cHub.finalizeMigration(channelId, finalizeMigrationState);

        // Verify channel is in MIGRATED_OUT status
        verifyChannelData(
            channelId, ChannelStatus.MIGRATED_OUT, finalizeMigrationVersion, 0, "Channel should be MIGRATED_OUT"
        );

        // Try to challenge channel in MIGRATED_OUT status (should fail)
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, finalizeMigrationState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.challengeChannel(channelId, finalizeMigrationState, challengerSig, ParticipantIndex.NODE);
    }

    function test_revert_challengeWithFinalizeMigrationIntent() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, finalizeMigrationState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.challengeChannel(channelId, finalizeMigrationState, challengerSig, ParticipantIndex.NODE);
    }
}
// forge-lint: disable-end(unsafe-typecast)
