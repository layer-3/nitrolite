// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {ChannelDefinition, State, StateIntent} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

contract ChannelHubTest_createChannel is ChannelHubTest_Base {
    ChannelDefinition internal def;

    function setUp() public override {
        super.setUp();
        def = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
    }

    // ========== Payable ==========

    // createChannel is payable to support native ETH deposits (DEPOSIT intent).
    // For WITHDRAW and OPERATE intents, no funds are pulled from the user,
    // so any ETH sent with these intents is explicitly rejected.

    function test_revert_ifETHSent_withdrawIntent() public {
        State memory state;
        state.intent = StateIntent.WITHDRAW;

        vm.deal(address(this), 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        cHub.createChannel{value: 1}(def, state);
    }

    function test_revert_ifETHSent_operateIntent() public {
        State memory state;
        state.intent = StateIntent.OPERATE;

        vm.deal(address(this), 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        cHub.createChannel{value: 1}(def, state);
    }
}
