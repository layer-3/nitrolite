// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestChannelHub} from "./TestChannelHub.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {NonReturningERC20} from "./mocks/NonReturningERC20.sol";
import {RevertingERC20} from "./mocks/RevertingERC20.sol";
import {GasConsumingERC20} from "./mocks/GasConsumingERC20.sol";
import {MalformedReturningERC20} from "./mocks/MalformedReturningERC20.sol";
import {DonatingERC20} from "./mocks/DonatingERC20.sol";
import {RevertingEthReceiver} from "./mocks/RevertingEthReceiver.sol";
import {GasConsumingEthReceiver} from "./mocks/GasConsumingEthReceiver.sol";

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";

/**
 * @notice Simple contract that can receive ETH for testing normal transfers
 */
contract SimpleReceiver {
    receive() external payable {}
}

contract ChannelHubTest_nonRevertingPushFunds is Test {
    TestChannelHub public cHub;
    MockERC20 public normalToken;
    NonReturningERC20 public nonReturningToken;
    RevertingERC20 public revertingToken;
    GasConsumingERC20 public gasConsumingToken;
    MalformedReturningERC20 public malformedToken;

    SimpleReceiver public simpleReceiver;
    RevertingEthReceiver public revertingReceiver;
    GasConsumingEthReceiver public gasConsumingReceiver;

    address public recipient;

    uint256 constant TRANSFER_AMOUNT = 1000 ether;
    uint256 constant BALANCE_AMOUNT = TRANSFER_AMOUNT * 10;

    function setUp() public {
        cHub = new TestChannelHub(new ECDSAValidator(), makeAddr("node"));
        normalToken = new MockERC20("Normal Token", "NRM", 18);
        nonReturningToken = new NonReturningERC20();
        revertingToken = new RevertingERC20();
        gasConsumingToken = new GasConsumingERC20();
        malformedToken = new MalformedReturningERC20();

        simpleReceiver = new SimpleReceiver();
        revertingReceiver = new RevertingEthReceiver();
        gasConsumingReceiver = new GasConsumingEthReceiver();

        recipient = makeAddr("recipient");

        vm.deal(address(cHub), BALANCE_AMOUNT);

        normalToken.mint(address(cHub), BALANCE_AMOUNT);
        nonReturningToken.mint(address(cHub), BALANCE_AMOUNT);
        revertingToken.mint(address(cHub), BALANCE_AMOUNT);
        gasConsumingToken.mint(address(cHub), BALANCE_AMOUNT);
        malformedToken.mint(address(cHub), BALANCE_AMOUNT);
    }

    function _verifyTransferSuccess(address user, address tokenAddr, uint256 transferredAmount) internal view {
        if (tokenAddr == address(0)) {
            uint256 userBalanceAfter = user.balance;
            assertEq(userBalanceAfter, transferredAmount, "Transfer amount mismatch");
            uint256 channelHubBalanceAfter = address(cHub).balance;
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT - transferredAmount, "ChannelHub balance mismatch");
        } else {
            IERC20 token = IERC20(tokenAddr);
            uint256 userBalanceAfter = token.balanceOf(user);
            assertEq(userBalanceAfter, transferredAmount, "Transfer amount mismatch");
            uint256 channelHubBalanceAfter = token.balanceOf(address(cHub));
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT - transferredAmount);
        }

        assertEq(cHub.getReclaimBalance(user, tokenAddr), 0, "Reclaim amount should be zero");
    }

    function _verifyBalancesNotChanged(address user, address tokenAddr, uint256 expectedReclaimAmount) internal view {
        if (tokenAddr == address(0)) {
            uint256 userBalanceAfter = user.balance;
            assertEq(userBalanceAfter, 0, "User balance should not change");
            uint256 channelHubBalanceAfter = address(cHub).balance;
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT, "ChannelHub balance should not change");
        } else {
            IERC20 token = IERC20(tokenAddr);
            uint256 userBalanceAfter = token.balanceOf(user);
            assertEq(userBalanceAfter, 0, "User balance should not change");
            uint256 channelHubBalanceAfter = token.balanceOf(address(cHub));
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT, "ChannelHub balance should not change");
        }

        assertEq(cHub.getReclaimBalance(user, tokenAddr), expectedReclaimAmount, "Reclaim amount mismatch");
    }

    // ========== Normal ERC20 Tests ==========

    function test_succeeds_withNormalERC20() public {
        cHub.exposed_nonRevertingPushFunds(recipient, address(normalToken), TRANSFER_AMOUNT);
        _verifyTransferSuccess(recipient, address(normalToken), TRANSFER_AMOUNT);
    }

    function test_succeeds_withZeroAmount() public {
        // Should not revert, should be a no-op
        cHub.exposed_nonRevertingPushFunds(recipient, address(normalToken), 0);
        _verifyBalancesNotChanged(recipient, address(normalToken), 0);
    }

    // ========== Non-Returning ERC20 Tests ==========

    function test_succeeds_withNonReturningERC20() public {
        cHub.exposed_nonRevertingPushFunds(recipient, address(nonReturningToken), TRANSFER_AMOUNT);
        _verifyTransferSuccess(recipient, address(nonReturningToken), TRANSFER_AMOUNT);
    }

    // ========== False-Returning ERC20 Tests ==========

    function test_accumulatesReclaims_whenERC20ReturnsFalse() public {
        normalToken.setFailTransfers(true);

        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(recipient, address(normalToken), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(normalToken), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(recipient, address(normalToken), TRANSFER_AMOUNT);
    }

    // ========== Reverting ERC20 Tests ==========

    function test_accumulatesReclaims_whenERC20Reverts() public {
        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(recipient, address(revertingToken), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(revertingToken), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(recipient, address(revertingToken), TRANSFER_AMOUNT);
    }

    function test_accumulatesReclaims_multipleFailedTransfers() public {
        cHub.exposed_nonRevertingPushFunds(recipient, address(revertingToken), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(revertingToken), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(recipient, address(revertingToken), TRANSFER_AMOUNT * 2);
    }

    // ========== Gas Consuming ERC20 Tests ==========

    function test_accumulatesReclaims_whenERC20ConsumesAllGas() public {
        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(recipient, address(gasConsumingToken), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(gasConsumingToken), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(recipient, address(gasConsumingToken), TRANSFER_AMOUNT);
    }

    // ========== Malformed Returning ERC20 Tests ==========

    function test_accumulatesReclaims_whenERC20ReturnsMalformedData() public {
        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(recipient, address(malformedToken), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(malformedToken), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(recipient, address(malformedToken), TRANSFER_AMOUNT);
    }

    // ========== ERC777 Donation-Back Tests ==========

    function test_succeeds_whenERC777DonatesBack() public {
        uint256 donationAmount = 1 ether;
        DonatingERC20 donatingToken = new DonatingERC20(address(cHub), donationAmount);
        donatingToken.mint(address(cHub), BALANCE_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(recipient, address(donatingToken), TRANSFER_AMOUNT);

        // Recipient received the transferred amount
        assertEq(donatingToken.balanceOf(recipient), TRANSFER_AMOUNT, "Recipient should have received tokens");
        // No false reclaim: old balance-check code would have incorrectly added TRANSFER_AMOUNT here
        assertEq(cHub.getReclaimBalance(recipient, address(donatingToken)), 0, "No reclaim should be created");
    }

    // ========== Native ETH Tests ==========

    function test_succeeds_withNativeETH() public {
        cHub.exposed_nonRevertingPushFunds(recipient, address(0), TRANSFER_AMOUNT);

        _verifyTransferSuccess(recipient, address(0), TRANSFER_AMOUNT);
    }

    function test_succeeds_withNativeETH_toContract() public {
        cHub.exposed_nonRevertingPushFunds(address(simpleReceiver), address(0), TRANSFER_AMOUNT);

        _verifyTransferSuccess(address(simpleReceiver), address(0), TRANSFER_AMOUNT);
    }

    // ========== Reverting ETH Receiver Tests ==========

    function test_accumulatesReclaims_whenETHReceiverReverts() public {
        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(address(revertingReceiver), address(0), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(address(revertingReceiver), address(0), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(address(revertingReceiver), address(0), TRANSFER_AMOUNT);
    }

    // ========== Gas Consuming ETH Receiver Tests ==========

    function test_accumulatesReclaims_whenETHReceiverConsumesAllGas() public {
        vm.expectEmit(true, true, false, true);
        emit ChannelHub.TransferFailed(address(gasConsumingReceiver), address(0), TRANSFER_AMOUNT);

        cHub.exposed_nonRevertingPushFunds(address(gasConsumingReceiver), address(0), TRANSFER_AMOUNT);

        _verifyBalancesNotChanged(address(gasConsumingReceiver), address(0), TRANSFER_AMOUNT);
    }
}
