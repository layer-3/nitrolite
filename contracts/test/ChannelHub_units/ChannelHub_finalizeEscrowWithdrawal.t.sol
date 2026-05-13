// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {TestUtils} from "../TestUtils.sol";
import {Utils} from "../../src/Utils.sol";

import {ChannelHub} from "../../src/ChannelHub.sol";
import {EscrowWithdrawalEngine} from "../../src/EscrowWithdrawalEngine.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    EscrowStatus,
    ParticipantIndex
} from "../../src/interfaces/Types.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_finalizeEscrowWithdrawal is ChannelHubTest_Base {
    ChannelDefinition internal def;
    bytes32 internal channelId;
    State internal initState;

    uint256 constant WITHDRAWAL_AMOUNT = 500;
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
        cHub.finalizeEscrowWithdrawal(channelId, bytes32(0), state);
    }

    // ========== Challenge Expiry Clearing ==========

    // Regression test: cooperative finalization from DISPUTED must zero out challengeExpireAt.
    // Before the fix, _applyEscrowWithdrawalEffects used `if (effects.newChallengeExpiry > 0)` which
    // skipped the write when the finalize effects left newChallengeExpiry at 0, leaving a stale
    // non-zero value observable via getEscrowWithdrawalData().
    function test_cooperativeFinalize_fromDISPUTED_clearsChallengeExpiry() public {
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
        // Node locks WITHDRAWAL_AMOUNT from its vault on this (non-home) chain.
        State memory escrowInitState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: WITHDRAWAL_AMOUNT,
                userNetFlow: int256(WITHDRAWAL_AMOUNT),
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
        escrowInitState = mutualSignStateBothWithEcdsaValidator(escrowInitState, altChannelId, ALICE_PK);

        cHub.initiateEscrowWithdrawal(altDef, escrowInitState);
        bytes32 escrowId = Utils.getEscrowId(altChannelId, escrowInitState.version);

        // Challenge → status becomes DISPUTED, challengeExpireAt set to a future timestamp.
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(altChannelId, escrowInitState, ALICE_PK);
        cHub.challengeEscrowWithdrawal(escrowId, challengerSig, ParticipantIndex.USER);

        (, EscrowStatus disputedStatus, uint64 challengeExpiryAfterChallenge,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(disputedStatus), uint8(EscrowStatus.DISPUTED), "Should be DISPUTED after challenge");
        assertGt(challengeExpiryAfterChallenge, 0, "challengeExpireAt should be non-zero after challenge");

        // Cooperatively finalize before the challenge period expires.
        // User allocation on home decreases by WITHDRAWAL_AMOUNT; node releases locked funds to user.
        State memory finalizeState = TestUtils.nextState(
            escrowInitState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(0), uint256(0)],
            [int256(WITHDRAWAL_AMOUNT), -int256(WITHDRAWAL_AMOUNT)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [-int256(WITHDRAWAL_AMOUNT), int256(WITHDRAWAL_AMOUNT)]
        );
        finalizeState = mutualSignStateBothWithEcdsaValidator(finalizeState, altChannelId, ALICE_PK);
        cHub.finalizeEscrowWithdrawal(altChannelId, escrowId, finalizeState);

        // Assert: status FINALIZED and challengeExpireAt cleared to zero.
        (, EscrowStatus finalStatus, uint64 finalChallengeExpiry,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(
            uint8(finalStatus), uint8(EscrowStatus.FINALIZED), "Should be FINALIZED after cooperative finalization"
        );
        assertEq(
            finalChallengeExpiry, 0, "challengeExpireAt must be cleared after cooperative finalization from DISPUTED"
        );
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
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: NON_HOME_CHAIN_ID,
                token: NON_HOME_TOKEN,
                decimals: 18,
                userAllocation: WITHDRAWAL_AMOUNT,
                userNetFlow: int256(WITHDRAWAL_AMOUNT),
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
        escrowInitState = mutualSignStateBothWithEcdsaValidator(escrowInitState, altChannelId, ALICE_PK);

        // Channel is VOID here, so initiateEscrowWithdrawal takes the non-home path and writes metadata.
        cHub.initiateEscrowWithdrawal(altDef, escrowInitState);
        bytes32 escrowId = Utils.getEscrowId(altChannelId, escrowInitState.version);

        // Finalize with wrong intent — must exercise the non-home metadata path and revert.
        State memory state;
        state.intent = StateIntent.DEPOSIT;

        vm.expectRevert(ChannelHub.IncorrectStateIntent.selector);
        cHub.finalizeEscrowWithdrawal(altChannelId, escrowId, state);
    }
}
