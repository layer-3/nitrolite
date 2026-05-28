// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {TestUtils} from "../TestUtils.sol";
import {Utils} from "../../src/Utils.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    ChannelStatus,
    EscrowStatus
} from "../../src/interfaces/Types.sol";

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

    function test_revert_homeChain_ifETHSent() public {
        vm.deal(node, 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        vm.prank(node);
        cHub.initiateEscrowDeposit{value: 1}(def, escrowState);
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

    // ========== Non-home native deposit ==========

    function test_nonHomeChain_nativeDeposit_acceptsExactETH() public {
        (ChannelDefinition memory nonHomeDef, bytes32 nonHomeChannelId, State memory nativeEscrowState) =
            _buildNonHomeNativeEscrowDeposit();
        bytes32 escrowId = Utils.getEscrowId(nonHomeChannelId, nativeEscrowState.version);

        uint256 hubBalanceBefore = address(cHub).balance;
        vm.deal(alice, ESCROW_AMOUNT);
        vm.prank(alice);
        cHub.initiateEscrowDeposit{value: ESCROW_AMOUNT}(nonHomeDef, nativeEscrowState);

        (, EscrowStatus status,,, uint256 lockedAmount,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(status), uint8(EscrowStatus.INITIALIZED), "Escrow should be initialized");
        assertEq(lockedAmount, ESCROW_AMOUNT, "Escrow should lock native ETH");
        assertEq(address(cHub).balance, hubBalanceBefore + ESCROW_AMOUNT, "Native ETH should be pulled");
    }

    function test_revert_nonHomeChain_nativeDeposit_wrongValue() public {
        (ChannelDefinition memory nonHomeDef,, State memory nativeEscrowState) = _buildNonHomeNativeEscrowDeposit();

        vm.deal(alice, ESCROW_AMOUNT - 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        vm.prank(alice);
        cHub.initiateEscrowDeposit{value: ESCROW_AMOUNT - 1}(nonHomeDef, nativeEscrowState);
    }

    function _buildNonHomeNativeEscrowDeposit()
        internal
        view
        returns (ChannelDefinition memory nonHomeDef, bytes32 nonHomeChannelId, State memory nativeEscrowState)
    {
        nonHomeDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE + 1,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
        nonHomeChannelId = Utils.getChannelId(nonHomeDef, CHANNEL_HUB_VERSION);

        nativeEscrowState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: ESCROW_AMOUNT,
                userNetFlow: int256(ESCROW_AMOUNT),
                nodeAllocation: ESCROW_AMOUNT,
                nodeNetFlow: int256(ESCROW_AMOUNT)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(0),
                decimals: 18,
                userAllocation: ESCROW_AMOUNT,
                userNetFlow: int256(ESCROW_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        nativeEscrowState = mutualSignStateBothWithEcdsaValidator(nativeEscrowState, nonHomeChannelId, ALICE_PK);
    }
}
