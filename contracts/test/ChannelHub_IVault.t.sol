// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "./ChannelHub_Base.t.sol";
import {RevertingEthReceiver} from "./mocks/RevertingEthReceiver.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {IVault} from "../src/interfaces/IVault.sol";

contract ChannelHubTest_depositToVault is ChannelHubTest_Base {
    // TODO:
}

contract ChannelHubTest_withdrawFromVault is ChannelHubTest_Base {
    RevertingEthReceiver public revertingReceiver;
    address public recipient;

    function setUp() public override {
        super.setUp();

        revertingReceiver = new RevertingEthReceiver();
        recipient = makeAddr("recipient");
    }

    // ========== Input Validation ==========

    function test_reverts_whenToIsZeroAddress() public {
        vm.prank(node);
        vm.expectRevert(ChannelHub.InvalidAddress.selector);
        cHub.withdrawFromVault(address(0), address(token), DEPOSIT_AMOUNT);
    }

    function test_reverts_whenAmountIsZero() public {
        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectAmount.selector);
        cHub.withdrawFromVault(recipient, address(token), 0);
    }

    function test_reverts_whenInsufficientBalance_ERC20() public {
        vm.prank(node);
        vm.expectRevert(ChannelHub.InsufficientBalance.selector);
        cHub.withdrawFromVault(recipient, address(token), INITIAL_BALANCE + 1);
    }

    function test_reverts_whenInsufficientBalance_nativeETH() public {
        vm.prank(node);
        vm.expectRevert(ChannelHub.InsufficientBalance.selector);
        cHub.withdrawFromVault(recipient, address(0), INITIAL_BALANCE + 1);
    }

    // ========== Successful Withdrawal — ERC20 ==========

    function test_succeeds_withERC20() public {
        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(token), DEPOSIT_AMOUNT);

        assertEq(token.balanceOf(recipient), DEPOSIT_AMOUNT, "Recipient should receive tokens");
        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT, "Node vault balance should decrease");
    }

    function test_emitsEvents_withERC20() public {
        vm.expectEmit(true, true, false, true);
        emit IVault.Withdrawn(node, address(token), DEPOSIT_AMOUNT);

        vm.expectEmit(true, true, false, true);
        emit ChannelHub.NodeBalanceUpdated(address(token), INITIAL_BALANCE - DEPOSIT_AMOUNT);

        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(token), DEPOSIT_AMOUNT);
    }

    function test_succeeds_fullBalance_ERC20() public {
        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(token), INITIAL_BALANCE);

        assertEq(token.balanceOf(recipient), INITIAL_BALANCE, "Recipient should receive full balance");
        assertEq(cHub.getAccountBalance(node, address(token)), 0, "Node vault balance should be zero");
    }

    function test_succeeds_toSelf_ERC20() public {
        vm.prank(node);
        cHub.withdrawFromVault(node, address(token), DEPOSIT_AMOUNT);

        assertEq(token.balanceOf(node), DEPOSIT_AMOUNT, "Node should receive tokens");
        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT, "Node vault balance should decrease");
    }

    // ========== Successful Withdrawal — Native ETH ==========

    function test_succeeds_withNativeETH() public {
        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(0), DEPOSIT_AMOUNT);

        assertEq(recipient.balance, DEPOSIT_AMOUNT, "Recipient should receive ETH");
        assertEq(cHub.getAccountBalance(node, address(0)), INITIAL_BALANCE - DEPOSIT_AMOUNT, "Node vault balance should decrease");
    }

    function test_emitsEvents_withNativeETH() public {
        vm.expectEmit(true, true, false, true);
        emit IVault.Withdrawn(node, address(0), DEPOSIT_AMOUNT);

        vm.expectEmit(true, true, false, true);
        emit ChannelHub.NodeBalanceUpdated(address(0), INITIAL_BALANCE - DEPOSIT_AMOUNT);

        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(0), DEPOSIT_AMOUNT);
    }

    function test_succeeds_fullBalance_nativeETH() public {
        vm.prank(node);
        cHub.withdrawFromVault(recipient, address(0), INITIAL_BALANCE);

        assertEq(recipient.balance, INITIAL_BALANCE, "Recipient should receive full ETH balance");
        assertEq(cHub.getAccountBalance(node, address(0)), 0, "Node vault balance should be zero");
    }

    function test_succeedss_toSelf_nativeETH() public {
        vm.prank(node);
        cHub.withdrawFromVault(node, address(0), DEPOSIT_AMOUNT);

        assertEq(node.balance, DEPOSIT_AMOUNT, "Node should receive ETH");
        assertEq(cHub.getAccountBalance(node, address(0)), INITIAL_BALANCE - DEPOSIT_AMOUNT, "Node vault balance should decrease");
    }

    // ========== Atomicity: Revert on Transfer Failure ==========

    function test_reverts_whenERC20TransferFails_balanceUnchanged() public {
        token.setFailTransfers(true);

        uint256 balanceBefore = cHub.getAccountBalance(node, address(token));

        vm.prank(node);
        vm.expectRevert();
        cHub.withdrawFromVault(recipient, address(token), DEPOSIT_AMOUNT);

        assertEq(cHub.getAccountBalance(node, address(token)), balanceBefore, "Vault balance must be unchanged on revert");
        assertEq(cHub.getReclaimBalance(recipient, address(token)), 0, "No reclaim should be created");
    }

    function test_reverts_whenNativeETHTransferFails_balanceUnchanged() public {
        uint256 balanceBefore = cHub.getAccountBalance(node, address(0));

        vm.prank(node);
        vm.expectRevert(abi.encodeWithSelector(ChannelHub.NativeTransferFailed.selector, address(revertingReceiver), DEPOSIT_AMOUNT));
        cHub.withdrawFromVault(address(revertingReceiver), address(0), DEPOSIT_AMOUNT);

        assertEq(cHub.getAccountBalance(node, address(0)), balanceBefore, "Vault balance must be unchanged on revert");
        assertEq(cHub.getReclaimBalance(address(revertingReceiver), address(0)), 0, "No reclaim should be created");
    }
}
