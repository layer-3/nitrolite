// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_EscrowDepositPurge_Base} from "./ChannelHub_EscrowDepositPurge_Base.t.sol";
import {MockERC20} from "../mocks/MockERC20.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {EscrowStatus, State} from "../../src/interfaces/Types.sol";

contract ChannelHubTest_purgeEscrowDeposits is ChannelHubTest_EscrowDepositPurge_Base {
    function _purge(uint256 maxToPurge) internal {
        cHub.harness_purgeEscrowDeposits(maxToPurge);
    }

    function _assertEscrowHead(uint256 expected, string memory message) internal view {
        assertEq(cHub.escrowHead(), expected, message);
    }

    function _assertNodeBalance(address node_, address token_, uint256 expected, string memory message) internal view {
        assertEq(cHub.getAccountBalance(node_, token_), expected, message);
    }

    function _assertNodeBalance(uint256 expected, string memory message) internal view {
        _assertNodeBalance(node, address(token), expected, message);
    }

    function _assertEscrowStatus(bytes32 escrowId, EscrowStatus expected, string memory message) internal view {
        (, EscrowStatus status,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(status), uint8(expected), message);
    }

    function _assertEscrowLockedAmount(bytes32 escrowId, uint256 expected, string memory message) internal view {
        (,,,, uint256 lockedAmount,) = cHub.getEscrowDepositData(escrowId);
        assertEq(lockedAmount, expected, message);
    }

    // ========== Empty queue ==========

    function test_doesNothing_forEmptyQueue() public {
        _purge(type(uint256).max);

        _assertEscrowHead(0, "Head should stay at 0 for empty queue");
    }

    // ========== Single entry ==========

    function test_doesNotPurge_whenSingleInitialized_notYetUnlockable() public {
        bytes32 id = _addNotYetUnlockable(LOCKED_AMOUNT);

        _purge(type(uint256).max);

        _assertEscrowHead(0, "Head should stay at 0");
        _assertEscrowStatus(id, EscrowStatus.INITIALIZED, "Status should remain INITIALIZED");
        _assertNodeBalance(0, "Node balance should be unchanged");
    }

    function test_purges_whenSingleInitialized_unlockable() public {
        bytes32 id = _addUnlockable(LOCKED_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.NodeBalanceUpdated(node, address(token), LOCKED_AMOUNT);
        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(1);

        _purge(type(uint256).max);

        _assertEscrowHead(1, "Head should advance to 1");
        _assertEscrowStatus(id, EscrowStatus.FINALIZED, "Status should be FINALIZED");
        _assertEscrowLockedAmount(id, 0, "Locked amount should be zeroed after purge");
        _assertNodeBalance(LOCKED_AMOUNT, "Node balance should be credited with locked amount");
    }

    function test_skips_whenSingleFinalized() public {
        _addFinalized();

        _purge(type(uint256).max);

        _assertEscrowHead(1, "Head should advance past FINALIZED");
        _assertNodeBalance(0, "Node balance should be unchanged");
    }

    function test_skips_whenSingleDisputed() public {
        _addDisputed(uint64(block.timestamp) + 1 days);

        _purge(type(uint256).max);

        _assertEscrowHead(1, "Head should advance past DISPUTED");
        _assertNodeBalance(0, "Node balance should be unchanged");
    }

    // ========== Multiple entries ==========

    function test_skipsDisputed_andPurgesSubsequentUnlockable() public {
        _addDisputed(uint64(block.timestamp) + 1 days);
        bytes32 id = _addUnlockable(LOCKED_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(1);

        _purge(type(uint256).max);

        _assertEscrowHead(2, "Head should advance past both entries");
        _assertEscrowStatus(id, EscrowStatus.FINALIZED, "Unlockable entry after DISPUTED should be purged");
        _assertNodeBalance(LOCKED_AMOUNT, "Node balance should reflect the purged unlockable entry");
    }

    function test_skipsFinalized_andPurgesSubsequentUnlockable() public {
        _addFinalized();
        bytes32 id = _addUnlockable(LOCKED_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(1);

        _purge(type(uint256).max);

        _assertEscrowHead(2, "Head should advance past both entries");
        _assertEscrowStatus(id, EscrowStatus.FINALIZED, "Unlockable entry after FINALIZED should be purged");
        _assertNodeBalance(LOCKED_AMOUNT, "Node balance should reflect the purged unlockable entry");
    }

    function test_stopsAtNonUnlockable_afterPurgingUnlockable() public {
        bytes32 id1 = _addUnlockable(LOCKED_AMOUNT);
        bytes32 id2 = _addNotYetUnlockable(LOCKED_AMOUNT * 2);

        _purge(type(uint256).max);

        _assertEscrowHead(1, "Head should stop at the not-yet-unlockable entry");
        _assertEscrowStatus(id1, EscrowStatus.FINALIZED, "First entry should be FINALIZED");
        _assertEscrowStatus(id2, EscrowStatus.INITIALIZED, "Second entry should remain INITIALIZED");
        _assertNodeBalance(LOCKED_AMOUNT, "Node balance should reflect only the first purged entry");
    }

    function test_purgesAll_inMultiUnlockableQueue() public {
        uint256 amount1 = 100;
        uint256 amount2 = 200;
        uint256 amount3 = 300;

        _addUnlockable(amount1);
        _addUnlockable(amount2);
        _addUnlockable(amount3);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(3);

        _purge(type(uint256).max);

        _assertEscrowHead(3, "Head should advance past all three entries");
        _assertNodeBalance(amount1 + amount2 + amount3, "Node balance should reflect all purged amounts");
    }

    // ========== maxToPurge limit ==========

    function test_respectsMaxToPurge_stopsAfterLimit() public {
        bytes32 id1 = _addUnlockable(LOCKED_AMOUNT);
        bytes32 id2 = _addUnlockable(LOCKED_AMOUNT);
        bytes32 id3 = _addUnlockable(LOCKED_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(2);

        _purge(2);

        _assertEscrowHead(2, "Head should advance by exactly maxToPurge");
        _assertEscrowStatus(id1, EscrowStatus.FINALIZED, "First entry should be purged");
        _assertEscrowStatus(id2, EscrowStatus.FINALIZED, "Second entry should be purged");
        _assertEscrowStatus(id3, EscrowStatus.INITIALIZED, "Third entry should remain INITIALIZED");
        _assertNodeBalance(LOCKED_AMOUNT * 2, "Node balance should reflect only the two purged entries");
    }

    // ========== Mixed queue ==========

    function test_skipsFinalizedAndDisputed_purgesUnlockable_stopsAtNonUnlockable() public {
        _addFinalized();
        _addDisputed(uint64(block.timestamp) + 1 days);
        bytes32 id3 = _addUnlockable(LOCKED_AMOUNT);
        bytes32 id4 = _addNotYetUnlockable(LOCKED_AMOUNT * 2);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.EscrowDepositsPurged(1);

        _purge(type(uint256).max);

        _assertEscrowHead(3, "Head should stop at the not-yet-unlockable entry");
        _assertEscrowStatus(id3, EscrowStatus.FINALIZED, "Third entry should be purged");
        _assertEscrowStatus(id4, EscrowStatus.INITIALIZED, "Fourth entry should remain INITIALIZED");
        _assertNodeBalance(LOCKED_AMOUNT, "Node balance should reflect only the one purged entry");
    }

    // ========== Node balance accuracy ==========

    function test_creditsCorrectNodeAndToken_whenMultipleNodesExist() public {
        MockERC20 token2 = new MockERC20("Token2", "TK2", 18);
        address node2 = makeAddr("node2");

        // node2/token2 escrow is unlockable; default node/token escrow is not yet unlockable
        bytes32 node2EscrowId = _nextEscrowId();
        cHub.workaround_setEscrowDeposit(
            node2EscrowId,
            bytes32(0),
            EscrowStatus.INITIALIZED,
            user,
            node2,
            uint64(block.timestamp) - 1,
            0,
            LOCKED_AMOUNT,
            address(token2)
        );
        cHub.workaround_addEscrowDepositId(node2EscrowId);

        _addNotYetUnlockable(LOCKED_AMOUNT);

        _purge(type(uint256).max);

        _assertNodeBalance(0, "Default node balance should be unchanged");
        _assertNodeBalance(node2, address(token2), LOCKED_AMOUNT, "node2 balance should be credited for token2");
    }
}
