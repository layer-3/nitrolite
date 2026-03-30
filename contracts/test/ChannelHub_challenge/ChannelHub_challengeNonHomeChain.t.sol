// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Challenge_Base} from "./ChannelHub_Challenge_Base.t.sol";

// forge-lint: disable-start(unsafe-typecast)

import {Utils} from "../../src/Utils.sol";
import {
    ChannelDefinition,
    ChannelStatus,
    State,
    StateIntent,
    Ledger,
    EscrowStatus,
    ParticipantIndex
} from "../../src/interfaces/Types.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {EscrowDepositEngine} from "../../src/EscrowDepositEngine.sol";
import {EscrowWithdrawalEngine} from "../../src/EscrowWithdrawalEngine.sol";

/*
 * @dev This file uses integration / blackbox testing through ChannelHub to verify
 *    critical end-to-end challenge flows (signature validation, fund movements, storage updates, events).
 *    Complex state machine logic and edge cases are tested exhaustively in dedicated engine unit tests
 *    (ChannelEngine.t.sol, EscrowDepositEngine.t.sol, EscrowWithdrawalEngine.t.sol) for faster execution
 *    and better isolation.
 */

contract ChannelHubTest_Challenge_NonHomeChain_EscrowDeposit is ChannelHubTest_Challenge_Base {
    /*
    - reverts on challenging NON-EXISTENT escrow deposit
    - escrow deposit can be challenged until `unlockAt` time has NOT passed
    - escrow deposit can NOT be challenged after `unlockAt` time has passed
    - challenged escrow deposit can be resolved until `challengeExpireAt` time has passed with a newer finalization state, which removes challenge and unlock funds
    - challenged escrow deposit can NOT be resolved if `challengeExpireAt` has passed, but
        can be withdrawn after `challengeExpireAt` time passes
    - reverts on challenging already challenged escrow deposit
    */

    uint64 constant ESCROW_VERSION = 1;
    uint256 constant ESCROW_AMOUNT = 500;

    bytes32 escrowId;
    State initiateEscrowDepositState;
    State finalizeEscrowDepositState;

    function setUp() public override {
        super.setUp();
        // `def` and `channelId` are set by ChannelHubTest_Challenge_Base.setUp()
        // For non-home chain: NON_HOME_CHAIN_ID (42) is the home chain, block.chainid is non-home

        initiateEscrowDepositState = State({
            version: ESCROW_VERSION,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID, // 42 — this IS the home chain (not current chain)
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: ESCROW_AMOUNT, // must equal deposit amount in WAD (same decimals here)
                nodeNetFlow: int256(ESCROW_AMOUNT)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid), // current chain is non-home
                token: address(token),
                decimals: 18,
                userAllocation: ESCROW_AMOUNT,
                userNetFlow: int256(ESCROW_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        initiateEscrowDepositState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowDepositState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        escrowId = Utils.getEscrowId(channelId, ESCROW_VERSION);

        // Finalize state (version = ESCROW_VERSION + 1):
        // home: userAllocation += ESCROW_AMOUNT, nodeAllocation = 0, userNetFlow unchanged
        // non-home: allocations = 0; userNetFlow = +ESCROW_AMOUNT, nodeNetFlow = -ESCROW_AMOUNT
        finalizeEscrowDepositState = nextState(
            initiateEscrowDepositState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(500 + ESCROW_AMOUNT), uint256(0)],
            [int256(500), int256(ESCROW_AMOUNT)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [int256(ESCROW_AMOUNT), -int256(ESCROW_AMOUNT)]
        );
        finalizeEscrowDepositState =
            mutualSignStateBothWithEcdsaValidator(finalizeEscrowDepositState, channelId, ALICE_PK);
    }

    function _challengeEscrowDeposit() internal {
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);
        vm.prank(node);
        cHub.challengeEscrowDeposit(escrowId, challengerSig, ParticipantIndex.NODE);
    }

    function test_revert_challengeEscrowDeposit_nonExistentEscrow() public {
        bytes32 nonExistentEscrowId = Utils.getEscrowId(channelId, 999);
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);
        vm.prank(node);
        vm.expectRevert(ChannelHub.NoChannelIdFoundForEscrow.selector);
        cHub.challengeEscrowDeposit(nonExistentEscrowId, challengerSig, ParticipantIndex.NODE);
    }

    function test_success_challengeEscrowDeposit_beforeUnlockAt() public {
        _challengeEscrowDeposit();

        (, EscrowStatus status,, uint64 challengeExpireAt,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(status), uint8(EscrowStatus.DISPUTED), "Escrow should be DISPUTED after challenge");
        assertEq(
            challengeExpireAt,
            uint64(block.timestamp) + EscrowDepositEngine.CHALLENGE_DURATION,
            "challengeExpireAt should be set to timestamp + CHALLENGE_DURATION"
        );
    }

    function test_revert_challengeEscrowDeposit_afterUnlockAt() public {
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        vm.expectRevert(EscrowDepositEngine.UnlockPeriodPassed.selector);
        _challengeEscrowDeposit();
    }

    function test_resolveChallengedEscrowDeposit_withFinalizeState_beforeChallengeExpiry() public {
        _challengeEscrowDeposit();

        (, EscrowStatus statusAfterChallenge,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(statusAfterChallenge), uint8(EscrowStatus.DISPUTED), "Should be DISPUTED after challenge");

        uint256 nodeVaultBefore = cHub.getAccountBalance(node, address(token));

        // Cooperative finalization with FINALIZE state (before challengeExpireAt)
        vm.prank(node);
        cHub.finalizeEscrowDeposit(channelId, escrowId, finalizeEscrowDepositState);

        (, EscrowStatus statusAfterFinalize,,, uint256 lockedAmount,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(statusAfterFinalize), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(lockedAmount, 0, "Locked amount should be 0 after finalization");

        // Cooperative path: locked funds released to node vault (node earned them for providing cross-chain liquidity)
        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBefore + ESCROW_AMOUNT,
            "Node vault should receive locked amount"
        );
    }

    function test_challengedEscrowDeposit_canNotBeResolved_nodeReclaimsAfterChallengeExpiry() public {
        _challengeEscrowDeposit();

        (, EscrowStatus statusAfterChallenge,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(statusAfterChallenge), uint8(EscrowStatus.DISPUTED), "Should be DISPUTED after challenge");

        vm.warp(block.timestamp + EscrowDepositEngine.CHALLENGE_DURATION + 1);

        uint256 aliceBalanceBefore = token.balanceOf(alice);

        // Unilateral finalization: anyone can call, state is ignored
        vm.prank(node);
        cHub.finalizeEscrowDeposit(channelId, escrowId, initiateEscrowDepositState);

        (, EscrowStatus statusAfterFinalize,,, uint256 lockedAmount,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(statusAfterFinalize), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(lockedAmount, 0, "Locked amount should be 0 after finalization");

        // Deposit Escrow funds are withdrawn to user wallet
        assertEq(token.balanceOf(alice), aliceBalanceBefore + ESCROW_AMOUNT, "User should receive locked amount");
    }

    function test_revert_challengeEscrowDeposit_alreadyChallenged() public {
        _challengeEscrowDeposit();

        // Attempt to challenge the same escrow deposit again
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowDepositState, NODE_PK);
        vm.prank(node);
        vm.expectRevert(EscrowDepositEngine.IncorrectEscrowStatus.selector);
        cHub.challengeEscrowDeposit(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}

contract ChannelHubTest_Challenge_NonHomeChain_EscrowWithdrawal is ChannelHubTest_Challenge_Base {
    /*
    - reverts on challenging NON-EXISTENT escrow withdrawal
    - escrow withdrawal can be challenged
    - challenged escrow withdrawal can be resolved until `challengeExpireAt` time has passed with a newer finalization state, which removes challenge and unlock funds
    - challenged escrow withdrawal can NOT be resolved if `challengeExpireAt` has passed, but
        can be withdrawn after `challengeExpireAt` time passes
    - reverts on challenging already challenged escrow withdrawal
    */

    uint64 constant WITHDRAWAL_VERSION = 1;
    uint256 constant WITHDRAWAL_AMOUNT = 300;

    bytes32 escrowId;
    State initiateEscrowWithdrawalState;
    State finalizeEscrowWithdrawalState;

    function setUp() public override {
        super.setUp();
        // `def` and `channelId` are set by ChannelHubTest_Challenge_Base.setUp()
        // For non-home chain: NON_HOME_CHAIN_ID (42) is the home chain, block.chainid is non-home

        initiateEscrowWithdrawalState = State({
            version: WITHDRAWAL_VERSION,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID, // 42 — this IS the home chain (not current chain)
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 500, // user has enough allocation to withdraw
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid), // current chain is non-home
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: WITHDRAWAL_AMOUNT, // node locks this amount for user's withdrawal
                nodeNetFlow: int256(WITHDRAWAL_AMOUNT)
            }),
            userSig: "",
            nodeSig: ""
        });
        initiateEscrowWithdrawalState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowWithdrawalState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initiateEscrowWithdrawalState);

        escrowId = Utils.getEscrowId(channelId, WITHDRAWAL_VERSION);

        // Finalize state (version = WITHDRAWAL_VERSION + 1):
        // home: userAllocation decreases by WITHDRAWAL_AMOUNT, nodeNetFlow decreases by WITHDRAWAL_AMOUNT
        // non-home: allocations = 0; userNetFlow = -WITHDRAWAL_AMOUNT, nodeNetFlow = +WITHDRAWAL_AMOUNT
        finalizeEscrowWithdrawalState = nextState(
            initiateEscrowWithdrawalState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(500 - WITHDRAWAL_AMOUNT), uint256(0)],
            [int256(500), -int256(WITHDRAWAL_AMOUNT)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [-int256(WITHDRAWAL_AMOUNT), int256(WITHDRAWAL_AMOUNT)]
        );
        finalizeEscrowWithdrawalState =
            mutualSignStateBothWithEcdsaValidator(finalizeEscrowWithdrawalState, channelId, ALICE_PK);
    }

    function _challengeEscrowWithdrawal() internal {
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);
        vm.prank(node);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.NODE);
    }

    function test_revert_challengeEscrowWithdrawal_nonExistentEscrow() public {
        bytes32 nonExistentEscrowId = Utils.getEscrowId(channelId, 999);
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);
        vm.prank(node);
        vm.expectRevert(ChannelHub.NoChannelIdFoundForEscrow.selector);
        cHub.challengeEscrowWithdrawal(nonExistentEscrowId, challengerSig, ParticipantIndex.NODE);
    }

    function test_challengeEscrowWithdrawal() public {
        _challengeEscrowWithdrawal();

        (, EscrowStatus status, uint64 challengeExpireAt,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(status), uint8(EscrowStatus.DISPUTED), "Escrow should be DISPUTED after challenge");
        assertEq(
            challengeExpireAt,
            uint64(block.timestamp) + EscrowWithdrawalEngine.CHALLENGE_DURATION,
            "challengeExpireAt should be set to timestamp + CHALLENGE_DURATION"
        );
    }

    function test_resolveChallengedEscrowWithdrawal_withFinalizeState_beforeChallengeExpiry() public {
        _challengeEscrowWithdrawal();

        (, EscrowStatus statusAfterChallenge,,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(statusAfterChallenge), uint8(EscrowStatus.DISPUTED), "Should be DISPUTED after challenge");

        uint256 aliceBalanceBefore = token.balanceOf(alice);
        uint256 nodeVaultBefore = cHub.getAccountBalance(node, address(token));

        // Cooperative finalization with FINALIZE state (before challengeExpireAt)
        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, finalizeEscrowWithdrawalState);

        (, EscrowStatus statusAfterFinalize,, uint256 lockedAmount,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(statusAfterFinalize), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(lockedAmount, 0, "Locked amount should be 0 after finalization");

        // Cooperative path: locked funds released to user wallet (withdrawal succeeded)
        assertEq(
            token.balanceOf(alice), aliceBalanceBefore + WITHDRAWAL_AMOUNT, "User should receive withdrawal amount"
        );
        // Node vault should be unchanged (locked amount was already deducted at initiation)
        assertEq(cHub.getAccountBalance(node, address(token)), nodeVaultBefore, "Node vault should be unchanged");
    }

    function test_challengedEscrowWithdrawal_canNotBeResolved_nodeReclaimsAfterChallengeExpiry() public {
        _challengeEscrowWithdrawal();

        vm.warp(block.timestamp + EscrowWithdrawalEngine.CHALLENGE_DURATION + 1);

        uint256 aliceBalanceBefore = token.balanceOf(alice);
        uint256 nodeVaultBefore = cHub.getAccountBalance(node, address(token));

        // Attempt cooperative resolution with a valid FINALIZE state after challengeExpireAt
        // The unilateral path intercepts and ignores the candidate state
        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, finalizeEscrowWithdrawalState);

        (, EscrowStatus status,, uint256 lockedAmount,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(status), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(lockedAmount, 0, "Locked amount should be 0");

        // Unilateral path (not cooperative): locked funds returned to node vault (withdrawal failed)
        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBefore + WITHDRAWAL_AMOUNT,
            "Node vault should reclaim locked amount (cooperative resolution bypassed)"
        );
        assertEq(token.balanceOf(alice), aliceBalanceBefore, "User wallet unchanged: withdrawal was not completed");
    }

    function test_revert_challengeEscrowWithdrawal_alreadyChallenged() public {
        _challengeEscrowWithdrawal();

        // Attempt to challenge the same escrow withdrawal again
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(channelId, initiateEscrowWithdrawalState, NODE_PK);
        vm.prank(node);
        vm.expectRevert(EscrowWithdrawalEngine.IncorrectEscrowStatus.selector);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}

contract ChannelHubTest_Challenge_NonHomeChain_HomeMigration is ChannelHubTest_Challenge_Base {
    /*
    Test cases:
    - a channel in Migrating_in status (empty channel after being called with `initiateMigration`) can be challenged with it
    - a channel in Migrating_in status (empty channel after being called with `initiateMigration`) can be challenged with a newer Operation state
    - a channel in Migrating_in status can be challenged with FINALIZE_MIGRATION intent (with version+1)
    */

    // New channel for testing NEW home chain behavior
    ChannelDefinition newHomeDef;
    bytes32 newHomeChannelId;

    uint64 initiateMigrationVersion = 1;
    State initiateMigrationState;
    uint64 finalizeMigrationVersion = 2;
    State finalizeMigrationState;
    uint64 newHomeOperateVersion = 3;
    State newHomeOperateState;

    function setUp() public override {
        super.setUp();

        // Setup for NEW home chain tests (migration IN)
        newHomeDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: uint64(42), // Different nonce to create a new channel
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
        newHomeChannelId = Utils.getChannelId(newHomeDef, CHANNEL_HUB_VERSION);

        // INITIATE_MIGRATION state for NEW home chain (migration IN)
        // homeLedger = OLD home chain (NON_HOME_CHAIN_ID)
        // nonHomeLedger = NEW home chain (current chain)
        initiateMigrationState = State({
            version: initiateMigrationVersion,
            intent: StateIntent.INITIATE_MIGRATION,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 500, // Node locks user allocation on new home
                nodeNetFlow: 500
            }),
            userSig: "",
            nodeSig: ""
        });
        initiateMigrationState =
            mutualSignStateBothWithEcdsaValidator(initiateMigrationState, newHomeChannelId, ALICE_PK);

        finalizeMigrationState = State({
            version: finalizeMigrationVersion,
            intent: StateIntent.FINALIZE_MIGRATION,
            metadata: bytes32(0),
            nonHomeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: -500
            }),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 500
            }),
            userSig: "",
            nodeSig: ""
        });
        finalizeMigrationState =
            mutualSignStateBothWithEcdsaValidator(finalizeMigrationState, newHomeChannelId, ALICE_PK);

        // OPERATE state on NEW home chain after migration
        // After initiateMigration on NEW home, ledgers are swapped, so homeLedger becomes current chain
        // OPERATE requires userNfDelta == 0, so userNetFlow must stay 0
        newHomeOperateState = State({
            version: newHomeOperateVersion,
            intent: StateIntent.OPERATE,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 450,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 450
            }),
            nonHomeLedger: Ledger({
                chainId: 0,
                token: address(0),
                decimals: 0,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        newHomeOperateState = mutualSignStateBothWithEcdsaValidator(newHomeOperateState, newHomeChannelId, ALICE_PK);
    }

    function test_challenge_newHomeChain_withInitiateMigration_asExisting() public {
        // Initiate migration IN on NEW home chain
        vm.prank(alice);
        cHub.initiateMigration(newHomeDef, initiateMigrationState);

        // Verify channel is in MIGRATING_IN status
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.MIGRATING_IN,
            initiateMigrationVersion,
            0,
            "newHomeInitiateMigrationState should be enforced"
        );

        // Challenge with the same INITIATE_MIGRATION state (already enforced)
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(newHomeChannelId, initiateMigrationState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(newHomeChannelId, initiateMigrationState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and state is still version 0
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.DISPUTED,
            initiateMigrationVersion,
            block.timestamp + CHALLENGE_DURATION,
            "initiateMigrationVersion should remain enforced"
        );
    }

    function test_challenge_newHomeChain_withOperate_inMigratingIn() public {
        // Initiate migration IN on NEW home chain
        vm.prank(alice);
        cHub.initiateMigration(newHomeDef, initiateMigrationState);

        // Verify channel is in MIGRATING_IN status
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.MIGRATING_IN,
            initiateMigrationVersion,
            0,
            "newHomeInitiateMigrationState should be enforced"
        );

        // Challenge with newer OPERATE state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(newHomeChannelId, newHomeOperateState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(newHomeChannelId, newHomeOperateState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and newHomeOperateState was enforced
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.DISPUTED,
            newHomeOperateVersion,
            block.timestamp + CHALLENGE_DURATION,
            "newHomeOperateState should start a challenge"
        );
        verifyChannelState(
            newHomeChannelId,
            [uint256(450), uint256(0)],
            [int256(0), int256(450)],
            "newHomeOperateState should be enforced"
        );
    }

    function test_challenge_newHomeChain_withFinalizeMigration() public {
        // Initiate migration IN on NEW home chain
        vm.prank(alice);
        cHub.initiateMigration(newHomeDef, initiateMigrationState);

        // Verify channel is in MIGRATING_IN status
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.MIGRATING_IN,
            initiateMigrationVersion,
            0,
            "newHomeInitiateMigrationState should be enforced"
        );

        // Challenge with newer FINALIZE_MIGRATION state
        bytes memory challengerSig =
            signChallengeEip191WithEcdsaValidator(newHomeChannelId, finalizeMigrationState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(newHomeChannelId, finalizeMigrationState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED and finalizeMigrationState was enforced
        verifyChannelData(
            newHomeChannelId,
            ChannelStatus.DISPUTED,
            finalizeMigrationVersion,
            block.timestamp + CHALLENGE_DURATION,
            "finalizeMigrationState should start a challenge"
        );
        verifyChannelState(
            newHomeChannelId,
            [uint256(500), uint256(0)],
            [int256(0), int256(500)],
            "finalizeMigrationState should be enforced"
        );
    }
}
// forge-lint: disable-end(unsafe-typecast)
