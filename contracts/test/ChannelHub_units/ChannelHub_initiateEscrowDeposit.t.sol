// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {TestUtils} from "../TestUtils.sol";
import {Utils} from "../../src/Utils.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {State, ChannelDefinition, StateIntent, Ledger, ChannelStatus} from "../../src/interfaces/Types.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_initiateEscrowDeposit is ChannelHubTest_Base {
    ChannelDefinition internal def;
    bytes32 internal channelId;
    State internal initState;
    State internal escrowState;

    uint256 constant ESCROW_AMOUNT = 500;
    uint64 constant NON_HOME_CHAIN_ID = 42;
    address constant NON_HOME_TOKEN = address(42);

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

        initState = State({
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
        initState = mutualSignStateBothWithEcdsaValidator(initState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, initState);

        // Build INITIATE_ESCROW_DEPOSIT: node locks ESCROW_AMOUNT on home chain,
        // user will lock ESCROW_AMOUNT on non-home chain.
        escrowState = TestUtils.nextState(
            initState,
            StateIntent.INITIATE_ESCROW_DEPOSIT,
            [uint256(DEPOSIT_AMOUNT), uint256(ESCROW_AMOUNT)],
            [int256(DEPOSIT_AMOUNT), int256(ESCROW_AMOUNT)],
            NON_HOME_CHAIN_ID,
            NON_HOME_TOKEN,
            [uint256(ESCROW_AMOUNT), uint256(0)],
            [int256(ESCROW_AMOUNT), int256(0)]
        );
        escrowState = mutualSignStateBothWithEcdsaValidator(escrowState, channelId, ALICE_PK);
    }

    // ========== StateIntent ==========

    function test_revert_ifWrongIntent() public {
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.initiateEscrowDeposit(def, state);
    }

    // ========== INITIATE_ESCROW_DEPOSIT caller restriction ==========

    function test_revert_homeChain_callerNotNode() public {
        vm.expectRevert(ChannelHub.IncorrectMsgSender.selector);
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, escrowState);
    }

    function test_homeChain_nodeCanSubmit() public {
        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));

        vm.prank(node);
        cHub.initiateEscrowDeposit(def, escrowState);

        // Channel state advanced and node funds locked
        verifyChannelData(channelId, ChannelStatus.OPERATING, 1, 0, "State should advance after node submits");
        assertEq(
            cHub.getNodeBalance(address(token)),
            nodeBalanceBefore - ESCROW_AMOUNT,
            "Node balance should decrease by escrow amount"
        );
    }
}
