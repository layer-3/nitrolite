// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import "forge-std/Test.sol";
import "../../src/ParametricToken.sol";

contract ParametricTokenTest is Test {
    ParametricToken public token;
    address public alice = address(0x1);
    address public bob = address(0x2);
    address public charlie = address(0x3);

    uint48 constant SUBID0 = 0;
    uint48 constant SUBID1 = 1;
    uint48 constant SUBID10 = 1000;

    function setUp() public {
        token = new ParametricToken("Shortbit", "sBTC");

        // Alice  mint
        vm.prank(alice);
        vm.warp(1_000_000);
        token.mint(1000 ether);
        console.log("Alice parameter", token.parameterOf(0, alice));
        vm.stopPrank();

        // Bob mint
        vm.prank(bob);
        vm.warp(2_000_000);
        token.mint(400 ether);
        console.log("Bob parameter after mint 1", token.parameterOf(0, bob));

        // Bob mint 2
        vm.prank(bob);
        vm.warp(4_000_000);
        token.mint(100 ether);
        console.log("Bob parameter after mint 2", token.parameterOf(0, bob));

        assertEq(token.parameterOf(0, bob), 2_400_000);
    }

    function testNormalTransfer() public {
        vm.prank(alice);
        token.transfer(bob, 100 ether);

        assertEq(token.balanceOf(alice), 900 ether);
        assertEq(token.balanceOf(bob), 600 ether);
    }

    function testConvertToSuper() public {
        vm.prank(alice);
        token.convertToSuper(alice);

        assertEq(uint8(token.accountType(alice)), uint8(IParametricToken.AccountType.Super));
        assertEq(token.balanceOf(alice), 1000 ether);
        assertEq(token.balanceOfSub(alice, 0), 1000 ether);
    }

    function testCreateSubAccount() public {
        vm.startPrank(alice);
        token.convertToSuper(alice);

        uint48 subId = token.createSubAccount(alice);
        // SubId 3 doesn't exist, should revert
        vm.expectRevert("Sub-account doesn't exist");
        token.balanceOfSub(alice, 3);
        vm.stopPrank();

        assertEq(subId, 1);
        assertEq(token.subsCountOf(alice), 2);
        assertEq(token.balanceOfSub(alice, subId), 0);
    }

    function testTransferToSub() public {
        // Setup
        vm.startPrank(alice);
        token.convertToSuper(alice);
        uint48 subId = token.createSubAccount(alice);
        vm.stopPrank();

        // Bob transfers to alice's sub-account 0
        vm.startPrank(bob);
        token.transferToSub(alice, 0, 100 ether);
        token.transferToSub(alice, subId, 50 ether);
        vm.stopPrank();

        assertEq(token.balanceOfSub(alice, 0), 1100 ether);
        assertEq(token.balanceOfSub(alice, subId), 50 ether);
        assertEq(token.balanceOf(alice), 1150 ether);
        assertEq(token.balanceOf(bob), 350 ether);

        console.log("Alice param sub 0", token.parameterOfSub(0, alice, 0));
        console.log("Alice param sub", subId, token.parameterOfSub(0, alice, subId));
    }

    function testTransferFromSub() public {
        // Setup: alice becomes super, creates sub, funds it
        vm.startPrank(alice);
        token.convertToSuper(alice);
        uint48 subId = token.createSubAccount(alice);
        token.transferBetweenSubs(0, subId, 200 ether);

        // Transfer from sub to bob
        token.transferFromSub(subId, bob, 150 ether);
        vm.stopPrank();

        assertEq(token.balanceOfSub(alice, subId), 50 ether);
        assertEq(token.balanceOf(bob), 650 ether);

        console.log("Alice param sub 0", token.parameterOfSub(0, alice, 0));
        console.log("Alice param sub", subId, token.parameterOfSub(0, alice, subId));
        console.log("Bob param", token.parameterOf(0, bob));
    }
}
