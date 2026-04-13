// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestChannelHub} from "./TestChannelHub.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {RevertingEthReceiver} from "./mocks/RevertingEthReceiver.sol";

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";

contract ChannelHubTest_claimFunds is Test {
    TestChannelHub public cHub;
    MockERC20 public token;
    RevertingEthReceiver public revertingReceiver;

    address public claimer;
    address public destination;

    uint256 constant RECLAIM_AMOUNT = 100 ether;
    uint256 constant BALANCE_AMOUNT = RECLAIM_AMOUNT * 10;

    function setUp() public {
        cHub = new TestChannelHub(new ECDSAValidator(), makeAddr("node"));
        token = new MockERC20("Test Token", "TST", 18);
        revertingReceiver = new RevertingEthReceiver();

        claimer = makeAddr("claimer");
        destination = makeAddr("destination");

        token.mint(address(cHub), BALANCE_AMOUNT);
        vm.deal(address(cHub), BALANCE_AMOUNT);
    }

    function _verifyTransferSuccess(address source, address destination_, address tokenAddr, uint256 transferredAmount)
        internal
        view
    {
        if (tokenAddr == address(0)) {
            uint256 userBalanceAfter = destination_.balance;
            assertEq(userBalanceAfter, transferredAmount, "Transfer amount mismatch");
            uint256 channelHubBalanceAfter = address(cHub).balance;
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT - transferredAmount, "ChannelHub balance mismatch");
        } else {
            IERC20 token_ = IERC20(tokenAddr);
            uint256 userBalanceAfter = token_.balanceOf(destination_);
            assertEq(userBalanceAfter, transferredAmount, "Transfer amount mismatch");
            uint256 channelHubBalanceAfter = token_.balanceOf(address(cHub));
            assertEq(channelHubBalanceAfter, BALANCE_AMOUNT - transferredAmount);
        }

        assertEq(cHub.getReclaimBalance(source, tokenAddr), 0, "Reclaim amount should be zero");
    }

    // ========== ERC20 Claim Tests ==========

    function test_success_erc20_toSameAddress() public {
        cHub.workaround_setReclaim(claimer, address(token), RECLAIM_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.FundsClaimed(claimer, address(token), claimer, RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(token), claimer);

        _verifyTransferSuccess(claimer, claimer, address(token), RECLAIM_AMOUNT);
    }

    function test_success_erc20_toDifferentAddress() public {
        cHub.workaround_setReclaim(claimer, address(token), RECLAIM_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.FundsClaimed(claimer, address(token), destination, RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(token), destination);

        _verifyTransferSuccess(claimer, destination, address(token), RECLAIM_AMOUNT);
    }

    function test_success_erc20_multipleAccumulations() public {
        // Simulate multiple accumulations by setting total accumulated amount
        uint256 totalAccumulated = RECLAIM_AMOUNT * 3;
        cHub.workaround_setReclaim(claimer, address(token), totalAccumulated);

        vm.prank(claimer);
        cHub.claimFunds(address(token), destination);

        _verifyTransferSuccess(claimer, destination, address(token), totalAccumulated);
    }

    // ========== Native ETH Claim Tests ==========

    function test_success_eth_toSameAddress() public {
        cHub.workaround_setReclaim(claimer, address(0), RECLAIM_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.FundsClaimed(claimer, address(0), claimer, RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(0), claimer);

        _verifyTransferSuccess(claimer, claimer, address(0), RECLAIM_AMOUNT);
    }

    function test_success_eth_toDifferentAddress() public {
        cHub.workaround_setReclaim(claimer, address(0), RECLAIM_AMOUNT);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.FundsClaimed(claimer, address(0), destination, RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(0), destination);

        _verifyTransferSuccess(claimer, destination, address(0), RECLAIM_AMOUNT);
    }

    // ========== Revert Tests ==========

    function test_revert_ifDestinationIsZeroAddress() public {
        cHub.workaround_setReclaim(claimer, address(token), RECLAIM_AMOUNT);

        vm.prank(claimer);
        vm.expectRevert(ChannelHub.InvalidAddress.selector);
        cHub.claimFunds(address(token), address(0));
    }

    function test_revert_ifReclaimBalanceIsZero() public {
        // No reclaim balance set

        vm.prank(claimer);
        vm.expectRevert(ChannelHub.IncorrectAmount.selector);
        cHub.claimFunds(address(token), destination);
    }

    function test_revert_ifETHTransferFails() public {
        cHub.workaround_setReclaim(claimer, address(0), RECLAIM_AMOUNT);

        vm.prank(claimer);
        vm.expectRevert(
            abi.encodeWithSelector(ChannelHub.NativeTransferFailed.selector, address(revertingReceiver), RECLAIM_AMOUNT)
        );
        cHub.claimFunds(address(0), address(revertingReceiver));
    }

    // ========== State Change Tests ==========

    function test_onlyClaimerCanClaimTheirReclaims() public {
        cHub.workaround_setReclaim(claimer, address(token), RECLAIM_AMOUNT);

        address otherUser = makeAddr("otherUser");

        // Other user tries to claim
        vm.prank(otherUser);
        vm.expectRevert(ChannelHub.IncorrectAmount.selector);
        cHub.claimFunds(address(token), destination);

        // Verify reclaim still exists for claimer
        assertEq(cHub.getReclaimBalance(claimer, address(token)), RECLAIM_AMOUNT, "Reclaim should still exist");
    }

    function test_canClaimDifferentTokensSeparately() public {
        MockERC20 token2 = new MockERC20("Token 2", "TK2", 18);
        token2.mint(address(cHub), BALANCE_AMOUNT);

        cHub.workaround_setReclaim(claimer, address(token), RECLAIM_AMOUNT);
        cHub.workaround_setReclaim(claimer, address(token2), RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(token), destination);

        _verifyTransferSuccess(claimer, destination, address(token), RECLAIM_AMOUNT);

        vm.prank(claimer);
        cHub.claimFunds(address(token2), destination);

        _verifyTransferSuccess(claimer, destination, address(token2), RECLAIM_AMOUNT);
    }
}
