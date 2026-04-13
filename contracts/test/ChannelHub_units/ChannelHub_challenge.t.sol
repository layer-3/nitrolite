// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {TestUtils} from "../TestUtils.sol";
import {Utils} from "../../src/Utils.sol";
import {ChannelHub} from "../../src/ChannelHub.sol";
import {State, ChannelDefinition, StateIntent, Ledger, ChannelStatus, ParticipantIndex} from "../../src/interfaces/Types.sol";

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
