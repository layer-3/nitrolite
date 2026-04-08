// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Vm} from "forge-std/Vm.sol";

import {ChannelHubTest_Base} from "./ChannelHub_Base.t.sol";
import {TestUtils} from "./TestUtils.sol";

import {Utils} from "../src/Utils.sol";
import {ChannelHub} from "../src/ChannelHub.sol";
import {ChannelDefinition, State, StateIntent, Ledger, ParticipantIndex} from "../src/interfaces/Types.sol";
import {EscrowWithdrawalEngine} from "../src/EscrowWithdrawalEngine.sol";

/**
 * Black-box tests verifying that NodeBalanceUpdated is emitted on every operation
 * that mutates internal node vault balance (_nodeBalances), and is NOT emitted when
 * no mutation occurs.
 *
 * # Scope
 *
 * "Node balance" here means the internal vault balance tracked by _nodeBalances,
 * i.e. the value returned by getAccountBalance(). It does NOT include funds pushed
 * directly to the node's address (e.g. nodeAllocation paid out on channel close),
 * because those bypass the vault and require no event.
 *
 * # Off-chain batching
 *
 * The protocol allows multiple off-chain transfers to be batched into a single on-chain
 * state update (checkpoint). From the contract's perspective this is indistinguishable
 * from a single transfer of the same net amount. Batching correctness is an
 * off-chain concern and belongs in off-chain unit tests.
 */

// forge-lint: disable-start(unsafe-typecast)
contract ChannelHubTest_emitsNodeBalanceUpdated is ChannelHubTest_Base {
    /**
     * Emits NodeBalanceUpdated:
     *   - depositToVault                            — direct vault deposit by node
     *   - withdrawFromVault                         — direct vault withdrawal by node
     *   - createChannel (DEPOSIT intent, both lock) — node locks funds into channel
     *   - createChannel (WITHDRAW intent)           — node locks funds into channel
     *   - depositToChannel (both lock)              — node locks funds into channel
     *   - withdrawFromChannel                       — node unlocks funds from channel
     *   - checkpoint (with node fund change)        — node balance changes due to off-chain transfer(s)
     *   - closeChannel cooperative (CLOSE intent)   — node unlocks funds from channel
     *   - challengeChannel with newer state         — when newer state carries non-zero node delta
     *   - initiateEscrowWithdrawal (non-home chain) — node locks liquidity for cross-chain withdrawal
     *   - finalizeEscrowDeposit (non-home chain)    — node releases locked liquidity after swap
     *   - finalizeEscrowWithdrawal (non-home chain, timeout) — node reclaims locked liquidity after challenge timeout
     *   - purgeEscrowDeposits                       — expired escrow deposits released back to node vault
     *
     * Does NOT emit NodeBalanceUpdated:
     *   - createChannel (DEPOSIT intent, only user deposits) - status change only, no node fund movement
     *   - depositToChannel (no change from Node)    - no fund movement
     *   - checkpoint with no node fund change       - no fund movement
     *   - initiateEscrowDeposit (non-home chain)    — status change only, no fund movement
     *   - challengeEscrowDeposit                    — status change only, no fund movement
     *   - challengeEscrowWithdrawal                 — status change only, no fund movement
     */
    // ======== State ========

    ChannelDefinition internal def;
    bytes32 internal channelId;

    // Used for non-home chain escrow tests (bob = user, node = node)
    ChannelDefinition internal bobDef;
    bytes32 internal bobChannelId;

    bytes32 constant NODE_BALANCE_UPDATED_SIG = keccak256("NodeBalanceUpdated(address,uint256)");

    // Non-home chain constants (fake foreign chain)
    uint64 constant FOREIGN_CHAIN_ID = 42;
    address constant FOREIGN_TOKEN = address(42);

    Ledger EMPTY_LEDGER = Ledger({
        chainId: 0, token: address(0), decimals: 0, userAllocation: 0, userNetFlow: 0, nodeAllocation: 0, nodeNetFlow: 0
    });

    // ======== Setup ========

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

        bobDef = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: bob,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });
        bobChannelId = Utils.getChannelId(bobDef, CHANNEL_HUB_VERSION);
    }

    // ======== Helpers ========

    /// @dev Expects the next NodeBalanceUpdated(token, expectedBalance) emission.
    function _expectEmitNodeBalanceUpdated(uint256 expectedBalance) internal {
        vm.expectEmit(true, true, true, true, address(cHub));
        emit ChannelHub.NodeBalanceUpdated(address(token), expectedBalance);
    }

    /// @dev Asserts NodeBalanceUpdated was NOT emitted in the logs recorded since the last vm.recordLogs().
    function _assertNoEmitNodeBalanceUpdated() internal view {
        Vm.Log[] memory logs = vm.getRecordedLogs();
        for (uint256 i = 0; i < logs.length; i++) {
            assertNotEq(logs[i].topics[0], NODE_BALANCE_UPDATED_SIG, "NodeBalanceUpdated was unexpectedly emitted");
        }
    }

    /// @dev Creates a channel for alice where node contributes nothing (nodeNetFlow = 0).
    ///      Returns the signed initial state.
    function _createSimpleChannel() internal returns (State memory state) {
        state = State({
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
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.createChannel(def, state);
    }

    /// @dev Creates a channel via OPERATE intent where node locks DEPOSIT_AMOUNT from vault.
    ///      Returns the signed initial state.
    function _createChannelNodeLocks() internal returns (State memory state) {
        state = State({
            version: 0,
            intent: StateIntent.OPERATE,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.createChannel(def, state);
    }

    /// @dev Sets up an escrow deposit on the non-home chain for bob (current chain = non-home).
    ///      Returns (escrowId, initState). Node vault unchanged; bob's DEPOSIT_AMOUNT is locked.
    function _initiateEscrowDeposit() internal returns (bytes32 escrowId, State memory initState) {
        initState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: FOREIGN_CHAIN_ID,
                token: FOREIGN_TOKEN,
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: DEPOSIT_AMOUNT,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        initState = mutualSignStateBothWithEcdsaValidator(initState, bobChannelId, BOB_PK);
        escrowId = Utils.getEscrowId(bobChannelId, initState.version);
        vm.prank(bob);
        cHub.initiateEscrowDeposit(bobDef, initState);
    }

    /// @dev Sets up an escrow withdrawal on the non-home chain for bob.
    ///      Returns (escrowId, initState). Node locks DEPOSIT_AMOUNT from vault.
    function _initiateEscrowWithdrawal() internal returns (bytes32 escrowId, State memory initState) {
        initState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: FOREIGN_CHAIN_ID,
                token: FOREIGN_TOKEN,
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: DEPOSIT_AMOUNT,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            userSig: "",
            nodeSig: ""
        });
        initState = mutualSignStateBothWithEcdsaValidator(initState, bobChannelId, BOB_PK);
        escrowId = Utils.getEscrowId(bobChannelId, initState.version);
        vm.prank(bob);
        cHub.initiateEscrowWithdrawal(bobDef, initState);
    }

    // ======== Tests: emits NodeBalanceUpdated ========

    function test_success_onDepositToVault() public {
        token.mint(node, DEPOSIT_AMOUNT);
        vm.startPrank(node);
        token.approve(address(cHub), DEPOSIT_AMOUNT);
        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE + DEPOSIT_AMOUNT);
        cHub.depositToVault(node, address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE + DEPOSIT_AMOUNT);
    }

    function test_success_onWithdrawFromVault() public {
        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE - DEPOSIT_AMOUNT);
        vm.prank(node);
        cHub.withdrawFromVault(node, address(token), DEPOSIT_AMOUNT);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }

    function test_success_onCreateChannel_depositIntent_bothDeposit() public {
        // both deposit
        State memory state = State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: DEPOSIT_AMOUNT,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE - DEPOSIT_AMOUNT);
        vm.prank(alice);
        cHub.createChannel(def, state);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }

    function test_success_onCreateChannel_withdrawIntent() public {
        // both deposit, node immediately transfers some funds for user to withdraw
        State memory state = State({
            version: 0,
            intent: StateIntent.WITHDRAW,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 500,
                userNetFlow: -500,
                nodeAllocation: 0,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE - DEPOSIT_AMOUNT);
        vm.prank(alice);
        cHub.createChannel(def, state);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }

    function test_success_onDepositToChannel_bothDeposit() public {
        // Setup: channel with user=DA, node=0
        State memory prevState = _createSimpleChannel();

        // Deposit: both User and Node specify amounts
        State memory candidate = TestUtils.nextState(
            prevState,
            StateIntent.DEPOSIT,
            [DEPOSIT_AMOUNT * 2, DEPOSIT_AMOUNT],
            [int256(DEPOSIT_AMOUNT) * 2, int256(DEPOSIT_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE - DEPOSIT_AMOUNT);
        vm.prank(alice);
        cHub.depositToChannel(channelId, candidate);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }

    function test_success_onWithdrawFromChannel() public {
        // Setup: channel via OPERATE where node locks DEPOSIT_AMOUNT (vault = INITIAL_BALANCE - DA)
        State memory prevState = _createChannelNodeLocks();

        // User withdraws 500
        State memory candidate = State({
            version: prevState.version + 1,
            intent: StateIntent.WITHDRAW,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: -500,
                nodeAllocation: 0,
                nodeNetFlow: 500
            }),
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        uint256 expectedBalance = INITIAL_BALANCE - DEPOSIT_AMOUNT + 500;
        _expectEmitNodeBalanceUpdated(expectedBalance);
        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, candidate);

        assertEq(cHub.getAccountBalance(node, address(token)), expectedBalance);
    }

    function test_success_onCheckpointChannel_withNodeFundChange() public {
        State memory prevState = _createSimpleChannel();

        // Off-chain: user transferred 500 to node.
        State memory candidate = TestUtils.nextState(
            prevState, StateIntent.OPERATE, [DEPOSIT_AMOUNT - 500, 0], [int256(DEPOSIT_AMOUNT), -500]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        uint256 expectedBalance = INITIAL_BALANCE + 500;
        _expectEmitNodeBalanceUpdated(expectedBalance);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, candidate);

        assertEq(cHub.getAccountBalance(node, address(token)), expectedBalance);
    }

    function test_success_onCloseChannel() public {
        // Setup: channel via OPERATE where node locks DEPOSIT_AMOUNT (vault = INITIAL_BALANCE - DEPOSIT_AMOUNT)
        State memory prevState = _createChannelNodeLocks();

        // Close: node balance returns to initial balance
        State memory candidate = State({
            version: prevState.version + 1,
            intent: StateIntent.CLOSE,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: EMPTY_LEDGER,
            userSig: "",
            nodeSig: ""
        });
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE);
        vm.prank(alice);
        cHub.closeChannel(channelId, candidate);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_success_onChallengeChannel_newerStateChangesNodeFunds() public {
        // Setup: simple channel; node vault = INITIAL_BALANCE, lockedFunds = DEPOSIT_AMOUNT (user's)
        State memory initState = _createSimpleChannel();

        // Off-chain: user transferred 500 to node (nodeNF goes from 0 to -500)
        // Enforce via challenge: nodeFundsDelta = -500 - 0 = -500 → vault += 500
        State memory stateV1 = TestUtils.nextState(
            initState, StateIntent.OPERATE, [DEPOSIT_AMOUNT - 500, uint256(0)], [int256(DEPOSIT_AMOUNT), -500]
        );
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        bytes memory sig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE + 500);
        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, sig, ParticipantIndex.NODE);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE + 500);
    }

    function test_success_onInitiateEscrowWithdrawal_nonHome() public {
        // Non-home chain (current): node locks DEPOSIT_AMOUNT from vault to fund user withdrawal
        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE - DEPOSIT_AMOUNT);
        _initiateEscrowWithdrawal();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }

    function test_success_onFinalizeEscrowDeposit_nonHome() public {
        // Setup: bob deposits DEPOSIT_AMOUNT into escrow (node vault unchanged = INITIAL_BALANCE)
        (bytes32 escrowId, State memory initState) = _initiateEscrowDeposit();

        // Finalize: DEPOSIT_AMOUNT flows from escrow (user's locked funds) to node vault
        State memory finalizeState = State({
            version: initState.version + 1,
            intent: StateIntent.FINALIZE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: FOREIGN_CHAIN_ID,
                token: FOREIGN_TOKEN,
                decimals: 18,
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: int256(DEPOSIT_AMOUNT)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: -int256(DEPOSIT_AMOUNT)
            }),
            userSig: "",
            nodeSig: ""
        });
        finalizeState = mutualSignStateBothWithEcdsaValidator(finalizeState, bobChannelId, BOB_PK);

        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE + DEPOSIT_AMOUNT);
        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, finalizeState);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE + DEPOSIT_AMOUNT);
    }

    function test_success_onFinalizeEscrowWithdrawal_nonHome_afterChallengeTimeout() public {
        // Setup: node locks DEPOSIT_AMOUNT → vault = INITIAL_BALANCE - DA
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawal();

        // Challenge: INITIALIZED → DISPUTED
        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);
        vm.prank(bob);
        cHub.challengeEscrowWithdrawal(escrowId, sig, ParticipantIndex.USER);

        // Expire the challenge
        vm.warp(block.timestamp + EscrowWithdrawalEngine.CHALLENGE_DURATION + 1);

        // Finalize via timeout: node reclaims DEPOSIT_AMOUNT → vault = INITIAL_BALANCE
        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE);
        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, initState);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_success_onPurgeEscrowDeposits() public {
        // Setup: bob deposits DEPOSIT_AMOUNT into escrow (node vault unchanged = INITIAL_BALANCE)
        _initiateEscrowDeposit();

        // Wait past unlock delay: escrow becomes unlockable
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        // Purge: DEPOSIT_AMOUNT flows from expired escrow to node vault
        _expectEmitNodeBalanceUpdated(INITIAL_BALANCE + DEPOSIT_AMOUNT);
        cHub.purgeEscrowDeposits(1);

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE + DEPOSIT_AMOUNT);
    }

    // ======== Tests: does NOT emit NodeBalanceUpdated ========

    function test_noEmit_onCreateChannel_depositIntent_onlyUserDeposits() public {
        // channel is created in _createSimpleChannel; re-verify logs from that call are cleared
        vm.recordLogs();

        _createSimpleChannel();

        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_noEmit_onDepositToChannel_noNodeChange() public {
        // Setup: simple channel, nodeNetFlow = 0
        State memory prevState = _createSimpleChannel();

        // Deposit: only user adds funds, nodeNetFlow stays at 0
        State memory candidate = TestUtils.nextState(
            prevState, StateIntent.DEPOSIT, [DEPOSIT_AMOUNT * 2, uint256(0)], [int256(DEPOSIT_AMOUNT) * 2, int256(0)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        vm.recordLogs();
        vm.prank(alice);
        cHub.depositToChannel(channelId, candidate);
        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_noEmit_onCheckpointChannel_noNodeChange() public {
        // Setup: simple channel, nodeNetFlow = 0
        State memory prevState = _createSimpleChannel();

        // Checkpoint: nodeNetFlow stays at 0, userNetFlow unchanged (OPERATE requires userNfDelta == 0)
        State memory candidate = TestUtils.nextState(
            prevState, StateIntent.OPERATE, [DEPOSIT_AMOUNT, uint256(0)], [int256(DEPOSIT_AMOUNT), int256(0)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);

        vm.recordLogs();
        vm.prank(alice);
        cHub.checkpointChannel(channelId, candidate);
        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_noEmit_onInitiateEscrowDeposit_nonHome() public {
        // Non-home chain initiate: only user funds move (userFundsDelta > 0, nodeFundsDelta = 0)
        vm.recordLogs();
        _initiateEscrowDeposit();
        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_noEmit_onChallengeEscrowDeposit() public {
        // Setup: bob deposits DEPOSIT_AMOUNT (node vault = INITIAL_BALANCE, no change)
        (bytes32 escrowId, State memory initState) = _initiateEscrowDeposit();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);

        vm.recordLogs();
        vm.prank(bob);
        cHub.challengeEscrowDeposit(escrowId, sig, ParticipantIndex.USER);
        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE);
    }

    function test_noEmit_onChallengeEscrowWithdrawal() public {
        // Setup: node locks DEPOSIT_AMOUNT (vault = INITIAL_BALANCE - DEPOSIT_AMOUNT)
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawal();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);

        vm.recordLogs();
        vm.prank(bob);
        cHub.challengeEscrowWithdrawal(escrowId, sig, ParticipantIndex.USER);
        _assertNoEmitNodeBalanceUpdated();

        assertEq(cHub.getAccountBalance(node, address(token)), INITIAL_BALANCE - DEPOSIT_AMOUNT);
    }
}
// forge-lint: disable-end(unsafe-typecast)
