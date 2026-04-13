// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestChannelHub} from "../TestChannelHub.sol";
import {MockERC20} from "../mocks/MockERC20.sol";
import {ECDSAValidator} from "../../src/sigValidators/ECDSAValidator.sol";
import {EscrowStatus} from "../../src/interfaces/Types.sol";

abstract contract ChannelHubTest_EscrowDepositPurge_Base is Test {
    TestChannelHub public cHub;
    MockERC20 public token;

    address public node;
    address public user;

    uint256 private _escrowCounter;

    uint64 internal constant UNLOCK_DELAY = 3 hours;
    uint256 internal constant LOCKED_AMOUNT = 500;

    function setUp() public virtual {
        node = makeAddr("node");
        user = makeAddr("user");

        cHub = new TestChannelHub(new ECDSAValidator(), node);
        token = new MockERC20("Test Token", "TST", 18);
    }

    /// @dev Creates a unique escrow ID for each call, deterministic per test
    function _nextEscrowId() internal returns (bytes32) {
        return bytes32(++_escrowCounter);
    }

    /// @dev Adds a single escrow entry to the hub storage and the purge queue
    function _addEscrow(EscrowStatus status, uint64 unlockAt, uint64 challengeExpireAt, uint256 lockedAmount)
        internal
        returns (bytes32 escrowId)
    {
        escrowId = _nextEscrowId();
        cHub.workaround_setEscrowDeposit(
            escrowId, bytes32(0), status, user, node, unlockAt, challengeExpireAt, lockedAmount, address(token)
        );
        cHub.workaround_addEscrowDepositId(escrowId);
    }

    /// @dev Shorthand: INITIALIZED escrow with unlock time in the past (purgeable)
    function _addUnlockable(uint256 lockedAmount) internal returns (bytes32) {
        return _addEscrow(EscrowStatus.INITIALIZED, uint64(block.timestamp) - 1, 0, lockedAmount);
    }

    /// @dev Shorthand: INITIALIZED escrow with unlock time in the future (not yet purgeable)
    function _addNotYetUnlockable(uint256 lockedAmount) internal returns (bytes32) {
        return _addEscrow(EscrowStatus.INITIALIZED, uint64(block.timestamp) + UNLOCK_DELAY, 0, lockedAmount);
    }

    /// @dev Shorthand: FINALIZED escrow
    function _addFinalized() internal returns (bytes32) {
        return _addEscrow(EscrowStatus.FINALIZED, 0, 0, 0);
    }

    /// @dev Shorthand: DISPUTED escrow (challenge still active)
    function _addDisputed(uint64 challengeExpireAt) internal returns (bytes32) {
        return _addEscrow(EscrowStatus.DISPUTED, uint64(block.timestamp) - 1, challengeExpireAt, LOCKED_AMOUNT);
    }
}
