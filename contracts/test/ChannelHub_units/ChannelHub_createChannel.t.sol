// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {Utils} from "../../src/Utils.sol";
import {ChannelDefinition, ChannelStatus, State, StateIntent, Ledger} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

contract ChannelHubTest_createChannel is ChannelHubTest_Base {
    ChannelDefinition internal def;
    bytes32 internal channelId;

    State initialDepositState;

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
        channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);

        initialDepositState = State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: TestUtils.emptyLedger(),
            userSig: "",
            nodeSig: ""
        });
    }

    // ========== Success: create on VOID channel ==========

    function test_createChannel_depositIntent() public {
        State memory state = mutualSignStateBothWithEcdsaValidator(initialDepositState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        verifyChannelData(
            channelId,
            ChannelStatus.OPERATING,
            0,
            0,
            "Channel status should be OPERATING after createChannel with DEPOSIT intent"
        );
    }

    function test_createChannel_withdrawIntent() public {
        // Represents an off-chain state (version > 0) where the node funded the user directly.
        // userNetFlow < 0: net funds flowed out to user; nodeNetFlow > 0: node vault funds that.
        // allocsSum must equal netFlowsSum (= 0), so both allocations are 0.
        State memory state = TestUtils.nextState(
            initialDepositState,
            StateIntent.WITHDRAW,
            [uint256(0), 0],
            [-int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        verifyChannelData(
            channelId,
            ChannelStatus.OPERATING,
            1,
            0,
            "Channel status should be OPERATING after createChannel with WITHDRAW intent"
        );
    }

    function test_createChannel_operateIntent() public {
        // Represents an off-chain state (version > 0) with no on-chain fund movement yet.
        State memory state =
            TestUtils.nextState(initialDepositState, StateIntent.OPERATE, [uint256(0), 0], [int256(0), 0]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        verifyChannelData(
            channelId,
            ChannelStatus.OPERATING,
            1,
            0,
            "Channel status should be OPERATING after createChannel with OPERATE intent"
        );
    }

    // ========== Revert: createChannel on existing channel ==========

    function _createDefaultChannel() internal {
        State memory state = mutualSignStateBothWithEcdsaValidator(initialDepositState, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.createChannel(def, state);
    }

    function test_revert_ifChannelExists_depositIntent() public {
        _createDefaultChannel();

        State memory state = TestUtils.nextState(
            initialDepositState,
            StateIntent.DEPOSIT,
            [DEPOSIT_AMOUNT, 0],
            [-int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.createChannel(def, state);
    }

    function test_revert_ifChannelExists_withdrawIntent() public {
        _createDefaultChannel();

        State memory state = TestUtils.nextState(
            initialDepositState,
            StateIntent.WITHDRAW,
            [DEPOSIT_AMOUNT, 0],
            [-int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.createChannel(def, state);
    }

    function test_revert_ifChannelExists_operateIntent() public {
        _createDefaultChannel();

        State memory state =
            TestUtils.nextState(initialDepositState, StateIntent.OPERATE, [uint256(0), 0], [int256(0), 0]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.createChannel(def, state);
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
