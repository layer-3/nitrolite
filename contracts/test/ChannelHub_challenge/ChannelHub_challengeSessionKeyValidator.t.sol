// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Challenge_Base} from "./ChannelHub_Challenge_Base.t.sol";

import {Utils} from "../../src/Utils.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "../TestUtils.sol";
import {ChannelDefinition, State, StateIntent, Ledger, ParticipantIndex} from "../../src/interfaces/Types.sol";
import {SessionKeyValidator, SessionKeyAuthorization} from "../../src/sigValidators/SessionKeyValidator.sol";

/*
 * @dev Black-box tests verifying that all three challenge functions revert with
 *      ChallengeWithSessionKeyNotSupported when the challenger signature uses the
 *      SessionKeyValidator. The channel/escrow must have the SessionKeyValidator
 *      approved (bit SESSION_KEY_VALIDATOR_ID set in approvedSignatureValidators) so
 *      that _extractValidator routes to the validator before the revert is triggered.
 */

abstract contract ChannelHubTest_Challenge_SkApproved_Base is ChannelHubTest_Challenge_Base {
    function setUp() public virtual override {
        super.setUp();

        def = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 1 << SESSION_KEY_VALIDATOR_ID,
            metadata: bytes32(0)
        });
        channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);
    }

    function buildSkChallengerSig(bytes32 channelId_, State memory state, uint256 signerPk, address skAddress)
        internal
        pure
        returns (bytes memory)
    {
        bytes32 metadataHash = keccak256("metadata");
        SessionKeyAuthorization memory skAuth = TestUtils.buildAndSignSkAuth(vm, skAddress, metadataHash, signerPk);
        // no need to change challenge signature in any way — just well-formatted would suffice
        return TestUtils.signStateEip191WithSkValidator(vm, channelId_, state, ALICE_SK1_PK, skAuth);
    }
}

// forge-lint: disable-start(unsafe-typecast)

// ─────────────────────────────────────────────────────────────────────────────
// challengeChannel
// ─────────────────────────────────────────────────────────────────────────────

contract ChannelHubTest_Challenge_HomeChain_SessionKeyValidator is ChannelHubTest_Challenge_SkApproved_Base {
    function setUp() public override {
        super.setUp();
        createChannelWithDeposit();
    }

    function test_revert_challengeChannel_withSkValidator_asUser() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initState, ALICE_PK, aliceSk1);

        vm.prank(alice);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.USER);
    }

    function test_revert_challengeChannel_withSkValidator_asNode() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initState, NODE_PK, aliceSk1);

        vm.prank(node);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// challengeEscrowDeposit
// ─────────────────────────────────────────────────────────────────────────────

contract ChannelHubTest_Challenge_NonHomeChain_EscrowDeposit_SessionKeyValidator is
    ChannelHubTest_Challenge_SkApproved_Base
{
    uint64 constant ESCROW_VERSION = 1;
    uint256 constant ESCROW_AMOUNT = 500;

    bytes32 escrowId;
    State initiateEscrowDepositState;

    function setUp() public override {
        super.setUp();

        // Non-home chain: NON_HOME_CHAIN_ID (42) is the home chain, block.chainid is the non-home chain.
        // No on-chain channel exists on the current chain — initiateEscrowDeposit takes the non-home path.
        initiateEscrowDepositState = State({
            version: ESCROW_VERSION,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: ESCROW_AMOUNT,
                nodeNetFlow: int256(ESCROW_AMOUNT)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: ESCROW_AMOUNT,
                userNetFlow: int256(ESCROW_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        initiateEscrowDepositState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowDepositState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initiateEscrowDepositState);

        escrowId = Utils.getEscrowId(channelId, ESCROW_VERSION);
    }

    function test_revert_challengeEscrowDeposit_withSkValidator_asUser() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initiateEscrowDepositState, ALICE_PK, aliceSk1);

        vm.prank(alice);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeEscrowDeposit(escrowId, challengerSig, ParticipantIndex.USER);
    }

    function test_revert_challengeEscrowDeposit_withSkValidator_asNode() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initiateEscrowDepositState, NODE_PK, aliceSk1);

        vm.prank(node);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeEscrowDeposit(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// challengeEscrowWithdrawal
// ─────────────────────────────────────────────────────────────────────────────

contract ChannelHubTest_Challenge_NonHomeChain_EscrowWithdrawal_SessionKeyValidator is
    ChannelHubTest_Challenge_SkApproved_Base
{
    uint64 constant WITHDRAWAL_VERSION = 1;
    uint256 constant WITHDRAWAL_AMOUNT = 300;

    bytes32 escrowId;
    State initiateEscrowWithdrawalState;

    function setUp() public override {
        super.setUp();

        // Non-home chain: NON_HOME_CHAIN_ID (42) is the home chain, block.chainid is the non-home chain.
        initiateEscrowWithdrawalState = State({
            version: WITHDRAWAL_VERSION,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: WITHDRAWAL_AMOUNT,
                nodeNetFlow: int256(WITHDRAWAL_AMOUNT)
            }),
            userSig: "",
            nodeSig: ""
        });
        initiateEscrowWithdrawalState =
            mutualSignStateBothWithEcdsaValidator(initiateEscrowWithdrawalState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initiateEscrowWithdrawalState);

        escrowId = Utils.getEscrowId(channelId, WITHDRAWAL_VERSION);
    }

    function test_revert_challengeEscrowWithdrawal_withSkValidator_asUser() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initiateEscrowWithdrawalState, ALICE_PK, aliceSk1);

        vm.prank(alice);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.USER);
    }

    function test_revert_challengeEscrowWithdrawal_withSkValidator_asNode() public {
        bytes memory challengerSig = buildSkChallengerSig(channelId, initiateEscrowWithdrawalState, NODE_PK, aliceSk1);

        vm.prank(node);
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.NODE);
    }
}
// forge-lint: disable-end(unsafe-typecast)
