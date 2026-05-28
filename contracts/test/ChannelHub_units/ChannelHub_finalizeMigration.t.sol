// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {State, StateIntent} from "../../src/interfaces/Types.sol";

contract ChannelHubTest_finalizeMigration is ChannelHubTest_Base {
    // ========== StateIntent ==========

    function test_revert_ifWrongIntent() public {
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.finalizeMigration(bytes32(0), state);
    }
}
