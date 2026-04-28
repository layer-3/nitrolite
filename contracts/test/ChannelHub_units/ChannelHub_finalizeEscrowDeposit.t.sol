// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {TestUtils} from "../TestUtils.sol";
import {Utils} from "../../src/Utils.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {State, ChannelDefinition, StateIntent, Ledger} from "../../src/interfaces/Types.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_finalizeEscrowDeposit is ChannelHubTest_Base {
    ChannelDefinition internal def;
    bytes32 internal channelId;
    State internal initState;

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
    }

    // ========== StateIntent ==========

    function test_revert_homeChain_ifWrongIntent() public {
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.finalizeEscrowDeposit(channelId, bytes32(0), state);
    }

    function test_revert_nonHomeChain_ifWrongIntent() public {
        // Use a different nonce so this channel does not exist on the current chain (non-home path).
        ChannelDefinition memory altDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE + 1,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
        bytes32 altChannelId = Utils.getChannelId(altDef, CHANNEL_HUB_VERSION);

        // Current chain acts as non-home chain; NON_HOME_CHAIN_ID is the home chain.
        State memory escrowInitState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
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
        escrowInitState = mutualSignStateBothWithEcdsaValidator(escrowInitState, altChannelId, ALICE_PK);

        // Channel is VOID here, so initiateEscrowDeposit takes the non-home path and writes metadata.
        cHub.initiateEscrowDeposit(altDef, escrowInitState);
        bytes32 escrowId = Utils.getEscrowId(altChannelId, escrowInitState.version);

        // Finalize with wrong intent — must exercise the non-home metadata path and revert.
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.finalizeEscrowDeposit(altChannelId, escrowId, state);
    }
}
