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
    ParticipantIndex
} from "../../src/interfaces/Types.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_challenge is ChannelHubTest_Base {
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

    function test_revert_closeIntent() public {
        State memory state;
        state.version = 1;
        state.intent = StateIntent.CLOSE;

        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, state, ALICE_PK);

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        vm.prank(alice);
        cHub.challengeChannel(channelId, state, challengerSig, ParticipantIndex.USER);
    }

    function test_revert_finalizeMigrationIntent() public {
        State memory state;
        state.version = 1;
        state.intent = StateIntent.FINALIZE_MIGRATION;

        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, state, ALICE_PK);

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        vm.prank(alice);
        cHub.challengeChannel(channelId, state, challengerSig, ParticipantIndex.USER);
    }

    // ========== Payable ==========

    function test_revert_ifETHSent_sameVersionChallenge() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.deal(node, 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        vm.prank(node);
        cHub.challengeChannel{value: 1}(channelId, initState, challengerSig, ParticipantIndex.NODE);
    }

    function test_revert_challengeChannel_initiateEscrowDepositIntent_ifETHSent() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, escrowState, NODE_PK);

        vm.deal(node, 1);
        vm.expectRevert(ChannelHub.IncorrectValue.selector);
        vm.prank(node);
        cHub.challengeChannel{value: 1}(channelId, escrowState, challengerSig, ParticipantIndex.NODE);
    }

    function test_nativeDepositChallenge_acceptsExactETH() public {
        uint256 depositDelta = 100;
        ChannelDefinition memory nativeDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE + 1,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
        bytes32 nativeChannelId = Utils.getChannelId(nativeDef, CHANNEL_HUB_VERSION);

        State memory nativeInitState = State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(0),
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
        nativeInitState = mutualSignStateBothWithEcdsaValidator(nativeInitState, nativeChannelId, ALICE_PK);

        vm.deal(alice, DEPOSIT_AMOUNT);
        vm.prank(alice);
        cHub.createChannel{value: DEPOSIT_AMOUNT}(nativeDef, nativeInitState);

        State memory depositState = TestUtils.nextState(
            nativeInitState,
            StateIntent.DEPOSIT,
            [uint256(DEPOSIT_AMOUNT + depositDelta), uint256(0)],
            [int256(DEPOSIT_AMOUNT + depositDelta), int256(0)]
        );
        depositState = mutualSignStateBothWithEcdsaValidator(depositState, nativeChannelId, ALICE_PK);
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(nativeChannelId, depositState, NODE_PK);

        uint256 hubBalanceBefore = address(cHub).balance;
        vm.deal(node, depositDelta);
        vm.prank(node);
        cHub.challengeChannel{value: depositDelta}(nativeChannelId, depositState, challengerSig, ParticipantIndex.NODE);

        (ChannelStatus status,, State memory latestState, uint256 challengeExpiry,) =
            cHub.getChannelData(nativeChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.DISPUTED), "Channel should be DISPUTED");
        assertEq(latestState.version, 1, "Native deposit state should be enforced");
        assertEq(
            latestState.homeLedger.userAllocation,
            DEPOSIT_AMOUNT + depositDelta,
            "Native allocation should include challenge deposit"
        );
        assertEq(challengeExpiry, block.timestamp + CHALLENGE_DURATION, "Challenge expiry should be set");
        assertEq(address(cHub).balance, hubBalanceBefore + depositDelta, "Native ETH should be pulled");
    }

    // ========== INITIATE_ESCROW_DEPOSIT caller restriction ==========

    function test_revert_initiateEscrowDeposit_homeChain_callerNotNode() public {
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, escrowState, ALICE_PK);

        vm.expectRevert(ChannelHub.IncorrectMsgSender.selector);
        vm.prank(alice);
        cHub.challengeChannel(channelId, escrowState, challengerSig, ParticipantIndex.USER);
    }

    function test_initiateEscrowDeposit_homeChain_nodeCanChallenge() public {
        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));

        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, escrowState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, escrowState, challengerSig, ParticipantIndex.NODE);

        // State is enforced and channel enters DISPUTED
        verifyChannelData(
            channelId,
            ChannelStatus.DISPUTED,
            1,
            block.timestamp + CHALLENGE_DURATION,
            "Channel should be DISPUTED with escrow state enforced"
        );
        assertEq(
            cHub.getNodeBalance(address(token)),
            nodeBalanceBefore - ESCROW_AMOUNT,
            "Node balance should decrease by escrow amount"
        );
    }
}
