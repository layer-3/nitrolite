// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {State, StateIntent} from "../../src/interfaces/Types.sol";

contract ChannelHubTest_checkpointChannel is ChannelHubTest_Base {
    // ========== StateIntent ==========

    function test_revert_ifWrongIntent() public {
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.checkpointChannel(bytes32(0), state);
    }

    // ========== Payable ==========

    // The EVM rejects ETH at the dispatcher level before any Solidity code runs,
    // producing an empty revert (no error selector).

    function test_revert_ifETHSent() public {
        State memory state;

        vm.deal(address(this), 1);
        (bool success, bytes memory returnData) =
            address(cHub).call{value: 1}(abi.encodeCall(cHub.checkpointChannel, (bytes32(0), state)));

        assertFalse(success, "checkpointChannel must not accept ETH");
        assertEq(returnData.length, 0, "Non-payable rejection produces no error data");
    }
}
