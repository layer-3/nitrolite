// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_EscrowDepositPurge_Base} from "./ChannelHub_EscrowDepositPurge_Base.t.sol";

contract ChannelHubTest_getUnlockableEscrowDepositStats is ChannelHubTest_EscrowDepositPurge_Base {
    function _getAndAssertStats(uint256 expectedCount, uint256 expectedTotal, string memory message) internal view {
        (uint256 count, uint256 totalAmount) = cHub.getUnlockableEscrowDepositStats();
        assertEq(count, expectedCount, message);
        assertEq(totalAmount, expectedTotal, message);
    }

    // ========== Empty queue ==========

    function test_returns_zero_forEmptyQueue() public view {
        _getAndAssertStats(0, 0, "Empty queue should return zeroes");
    }

    // ========== Single-entry queue ==========

    function test_returns_zero_whenSingleInitialized_notYetUnlockable() public {
        _addNotYetUnlockable(LOCKED_AMOUNT);

        _getAndAssertStats(0, 0, "Single initialized not yet unlockable should return zeroes");
    }

    function test_returns_one_whenSingleInitialized_unlockable() public {
        _addUnlockable(LOCKED_AMOUNT);

        _getAndAssertStats(1, LOCKED_AMOUNT, "Single initialized unlockable should return one count and amount");
    }

    function test_returns_zero_whenSingleFinalized() public {
        _addFinalized();

        _getAndAssertStats(0, 0, "Single finalized should return zeroes");
    }

    function test_returns_zero_whenSingleDisputed() public {
        _addDisputed(uint64(block.timestamp) + 1 days);

        _getAndAssertStats(0, 0, "Single disputed should return zeroes");
    }

    // ========== FINALIZED / DISPUTED blocking ==========

    function test_skipsFinalized_andCountsSubsequentUnlockable() public {
        _addFinalized();
        _addUnlockable(LOCKED_AMOUNT);

        _getAndAssertStats(1, LOCKED_AMOUNT, "Should skip finalized and count subsequent unlockable");
    }

    function test_skipsDisputed_andCountsSubsequentUnlockable() public {
        _addDisputed(uint64(block.timestamp) + 1 days);
        _addUnlockable(LOCKED_AMOUNT);

        _getAndAssertStats(1, LOCKED_AMOUNT, "Should skip disputed and count subsequent unlockable");
    }

    function test_skipsFinalizedAndDisputed_andCountsSubsequentUnlockable() public {
        _addFinalized();
        _addDisputed(uint64(block.timestamp) + 1 days);
        _addUnlockable(LOCKED_AMOUNT);

        _getAndAssertStats(1, LOCKED_AMOUNT, "Should skip finalized and disputed, and count subsequent unlockable");
    }

    // ========== Stop condition ==========

    function test_stopsAtNonUnlockable_afterCountingUnlockable() public {
        _addUnlockable(LOCKED_AMOUNT);
        _addNotYetUnlockable(LOCKED_AMOUNT * 2);

        _getAndAssertStats(1, LOCKED_AMOUNT, "Should count first unlockable and stop at subsequent non-unlockable");
    }

    // ========== Multiple entries ==========

    function test_countsAllUnlockable_inMultiEntryQueue() public {
        uint256 amount1 = 100;
        uint256 amount2 = 200;
        uint256 amount3 = 300;

        _addUnlockable(amount1);
        _addUnlockable(amount2);
        _addUnlockable(amount3);

        _getAndAssertStats(3, amount1 + amount2 + amount3, "Should count all unlockable entries in multi-entry queue");
    }

    // ========== escrowHead offset ==========

    /// @dev After the purge advances escrowHead past the first entry, stats must start from the new head
    function test_startsFromEscrowHead_ignoringAlreadyAdvancedEntries() public {
        // Position 0: unlockable — will be consumed by the purge call below
        _addUnlockable(LOCKED_AMOUNT);
        // Position 1: not yet unlockable — will be the first visible entry after head advances
        _addNotYetUnlockable(LOCKED_AMOUNT);

        // Advance escrowHead to 1 by purging position 0
        cHub.harness_purgeEscrowDeposits(1);
        assertEq(cHub.escrowHead(), 1);

        _getAndAssertStats(0, 0, "After advancing head, should return zeroes since next entry is not unlockable");
    }
}
