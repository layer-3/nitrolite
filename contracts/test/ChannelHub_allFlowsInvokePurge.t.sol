// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "./ChannelHub_Base.t.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "./TestUtils.sol";
import {TestChannelHub} from "./TestChannelHub.sol";
import {MockERC20} from "./mocks/MockERC20.sol";

import {Utils} from "../src/Utils.sol";
import {ChannelHub} from "../src/ChannelHub.sol";
import {
    ChannelDefinition,
    State,
    StateIntent,
    Ledger,
    ParticipantIndex,
    EscrowStatus
} from "../src/interfaces/Types.sol";
import {EscrowDepositEngine} from "../src/EscrowDepositEngine.sol";
import {EscrowWithdrawalEngine} from "../src/EscrowWithdrawalEngine.sol";

// forge-lint: disable-start(unsafe-typecast)

/**
 * Black-box tests verifying that _purgeEscrowDeposits is invoked on every protocol
 * operation that is expected to call it, and is NOT called by operations that should
 * not touch the purge queue.
 *
 * Detection mechanism: a FINALIZED sentinel entry is injected into the deposit queue
 * immediately before the tested action via _snapshotAndInjectSentinel(). Because
 * FINALIZED entries are skippable, any purge call advances escrowHead past the
 * sentinel. _assertPurgeInvoked() confirms escrowHead moved relative to the snapshot.
 *
 * Invokes purge (_purgeEscrowDeposits is called):
 *   === home chain flows ===
 *   - createChannel
 *   - depositToChannel
 *   - withdrawFromChannel
 *   - checkpointChannel
 *   - closeChannel
 *   - challengeChannel
 *   - initiateEscrowDeposit
 *   - finalizeEscrowDeposit
 *   - initiateEscrowWithdrawal
 *   - finalizeEscrowWithdrawal
 *
 *   === non-home-chain escrow flows ===
 *   - initiateEscrowDeposit
 *   - challengeEscrowDeposit
 *   - finalizeEscrowDeposit (cooperative)
 *   - finalizeEscrowDeposit (unilateral timeout)
 *   - initiateEscrowWithdrawal
 *   - challengeEscrowWithdrawal
 *   - finalizeEscrowWithdrawal (cooperative)
 *   - finalizeEscrowWithdrawal (unilateral timeout)
 *
 *   === migration flows ===
 *   - initiateMigration (home chain OUT)
 *   - initiateMigration (non-home chain IN)
 *   - finalizeMigration (new home chain IN)
 *   - finalizeMigration (old home chain OUT)
 *
 *   - purgeEscrowDeposits (public)
 *
 * Does NOT invoke purge:
 *   - depositToVault
 *   - withdrawFromVault
 */
contract ChannelHubTest_allFlowsInvokePurge is ChannelHubTest_Base {
    TestChannelHub internal tHub;

    ChannelDefinition internal def;
    bytes32 internal channelId;

    // Separate channel for non-home-chain escrow flows
    ChannelDefinition internal bobDef;
    bytes32 internal bobChannelId;

    uint64 constant FOREIGN_CHAIN_ID = 42;
    address constant FOREIGN_TOKEN = address(42);
    uint256 constant WITHDRAWAL_AMOUNT = 300;

    bytes32 constant SENTINEL_ESCROW_ID = keccak256("purge_sentinel_escrow_id");
    uint256 internal _headBeforeAction;

    // ======== setUp ========

    function setUp() public override {
        super.setUp();

        // Deploy TestChannelHub (a ChannelHub subtype) and assign it to the base's
        // cHub field so all inherited helpers keep working.
        tHub = new TestChannelHub(ECDSA_SIG_VALIDATOR, node);
        cHub = tHub;

        // super.setUp() already spent node's INITIAL_BALANCE on the old cHub vault.
        // Re-mint so node can fund the new tHub vault.
        token.mint(node, INITIAL_BALANCE);
        vm.startPrank(node);
        token.approve(address(cHub), INITIAL_BALANCE);
        cHub.depositToVault(node, address(token), INITIAL_BALANCE);
        vm.stopPrank();

        bytes memory skValidatorSig = TestUtils.buildAndSignValidatorRegistration(
            vm, SESSION_KEY_VALIDATOR_ID, address(SK_SIG_VALIDATOR), NODE_PK, address(cHub)
        );
        cHub.registerNodeValidator(node, SESSION_KEY_VALIDATOR_ID, SK_SIG_VALIDATOR, skValidatorSig);

        vm.prank(alice);
        token.approve(address(cHub), INITIAL_BALANCE);

        vm.prank(bob);
        token.approve(address(cHub), INITIAL_BALANCE);

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

    // ======== Purge detection helpers ========

    /// @dev Snapshots the current escrowHead and appends a FINALIZED sentinel to the
    ///      deposit queue at that position. Any subsequent purge call will skip the
    ///      sentinel and advance escrowHead, making the call detectable.
    function _snapshotAndInjectSentinel() internal {
        _headBeforeAction = tHub.escrowHead();
        tHub.workaround_setEscrowDeposit(
            SENTINEL_ESCROW_ID, bytes32(0), EscrowStatus.FINALIZED, address(0), address(0), 0, 0, 0, address(0)
        );
        tHub.workaround_addEscrowDepositId(SENTINEL_ESCROW_ID);
    }

    function _assertPurgeInvoked() internal view {
        assertGt(tHub.escrowHead(), _headBeforeAction, "_purgeEscrowDeposits should have been called");
    }

    function _assertPurgeNotInvoked() internal view {
        assertEq(tHub.escrowHead(), _headBeforeAction, "_purgeEscrowDeposits should NOT have been called");
    }

    // ======== Protocol helpers ========

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
            nonHomeLedger: TestUtils.emptyLedger(),
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.createChannel(def, state);
    }

    /// @dev Submits INITIATE_ESCROW_DEPOSIT on the home chain (uses alice's channel).
    ///      The node locks DEPOSIT_AMOUNT from its vault to back the cross-chain deposit.
    function _initiateEscrowDepositHomeChain() internal returns (bytes32 escrowId, State memory initState) {
        State memory prevState = _createSimpleChannel();
        initState = TestUtils.nextState(
            prevState,
            StateIntent.INITIATE_ESCROW_DEPOSIT,
            [DEPOSIT_AMOUNT, DEPOSIT_AMOUNT],
            [int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(DEPOSIT_AMOUNT), int256(0)]
        );
        initState = mutualSignStateBothWithEcdsaValidator(initState, channelId, ALICE_PK);
        escrowId = Utils.getEscrowId(channelId, initState.version);
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, initState);
    }

    /// @dev Submits INITIATE_ESCROW_WITHDRAWAL on the home chain (uses alice's channel).
    ///      No funds move yet; WITHDRAWAL_AMOUNT will be paid on the foreign chain.
    function _initiateEscrowWithdrawalHomeChain() internal returns (bytes32 escrowId, State memory initState) {
        State memory prevState = _createSimpleChannel();
        initState = TestUtils.nextState(
            prevState,
            StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(DEPOSIT_AMOUNT), int256(0)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [uint256(0), WITHDRAWAL_AMOUNT],
            [int256(0), int256(WITHDRAWAL_AMOUNT)]
        );
        initState = mutualSignStateBothWithEcdsaValidator(initState, channelId, ALICE_PK);
        escrowId = Utils.getEscrowId(channelId, initState.version);
        vm.prank(alice);
        cHub.initiateEscrowWithdrawal(def, initState);
    }

    /// @dev Submits INITIATE_ESCROW_DEPOSIT on the non-home chain (uses bob's channel).
    ///      Returns the escrowId and the signed initState for use in subsequent steps.
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

    /// @dev Submits INITIATE_ESCROW_WITHDRAWAL on the non-home chain (uses bob's channel).
    ///      The node locks DEPOSIT_AMOUNT from its vault on the home chain.
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

    /// @dev Initiates migration OUT on the home chain (alice's channel must already be OPERATING).
    function _initiateMigrationHomeChain(State memory prevState) internal returns (State memory initState) {
        initState = TestUtils.nextState(
            prevState,
            StateIntent.INITIATE_MIGRATION,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(DEPOSIT_AMOUNT), int256(0)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [uint256(0), DEPOSIT_AMOUNT],
            [int256(0), int256(DEPOSIT_AMOUNT)]
        );
        initState = mutualSignStateBothWithEcdsaValidator(initState, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.initiateMigration(def, initState);
    }

    /// @dev Initiates migration IN on the non-home chain (bob's channel, VOID → MIGRATING_IN).
    ///      State is passed with homeLedger=FOREIGN, nonHomeLedger=LOCAL;
    ///      the contract swaps them before storing, so the returned state reflects what is stored
    ///      (homeLedger=LOCAL, nonHomeLedger=FOREIGN) — ready for use as prevState in nextState().
    function _initiateMigrationNonHomeChain() internal returns (State memory storedState) {
        State memory initState = State({
            version: 1,
            intent: StateIntent.INITIATE_MIGRATION,
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
        vm.prank(bob);
        cHub.initiateMigration(bobDef, initState);

        // Mirror the contract's ledger swap so the caller has the stored representation.
        storedState = initState;
        storedState.homeLedger = initState.nonHomeLedger;
        storedState.nonHomeLedger = initState.homeLedger;
        storedState.userSig = "";
        storedState.nodeSig = "";
    }

    // ======== Tests: home chain flows ========

    function test_purgeInvoked_onCreateChannel() public {
        _snapshotAndInjectSentinel();
        _createSimpleChannel();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onDepositToChannel() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();

        State memory candidate = TestUtils.nextState(
            prevState,
            StateIntent.DEPOSIT,
            [DEPOSIT_AMOUNT * 2, DEPOSIT_AMOUNT],
            [int256(DEPOSIT_AMOUNT) * 2, int256(DEPOSIT_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.depositToChannel(channelId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onWithdrawFromChannel() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();

        State memory candidate = TestUtils.nextState(
            prevState,
            StateIntent.WITHDRAW,
            [DEPOSIT_AMOUNT - 500, uint256(0)],
            [int256(DEPOSIT_AMOUNT) - 500, int256(0)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onCheckpointChannel() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();

        State memory candidate = TestUtils.nextState(
            prevState, StateIntent.OPERATE, [DEPOSIT_AMOUNT - 500, uint256(0)], [int256(DEPOSIT_AMOUNT), -500]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onCloseChannel() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();

        State memory candidate =
            TestUtils.nextState(prevState, StateIntent.CLOSE, [uint256(0), uint256(0)], [int256(0), int256(0)]);
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.closeChannel(channelId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onChallengeChannel() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();

        State memory candidate = TestUtils.nextState(
            prevState, StateIntent.OPERATE, [DEPOSIT_AMOUNT - 500, uint256(0)], [int256(DEPOSIT_AMOUNT), -500]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, candidate, NODE_PK);
        vm.prank(node);
        cHub.challengeChannel(channelId, candidate, challengerSig, ParticipantIndex.NODE);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onInitiateEscrowDeposit_homeChain() public {
        _snapshotAndInjectSentinel();
        _initiateEscrowDepositHomeChain();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowDeposit_homeChain() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowDepositHomeChain();
        _snapshotAndInjectSentinel();

        // User allocation increases by DEPOSIT_AMOUNT; node allocation released.
        State memory candidate = TestUtils.nextState(
            initState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [DEPOSIT_AMOUNT * 2, uint256(0)],
            [int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [uint256(0), uint256(0)],
            [int256(DEPOSIT_AMOUNT), -int256(DEPOSIT_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.finalizeEscrowDeposit(channelId, escrowId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onInitiateEscrowWithdrawal_homeChain() public {
        _snapshotAndInjectSentinel();
        _initiateEscrowWithdrawalHomeChain();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowWithdrawal_homeChain() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawalHomeChain();
        _snapshotAndInjectSentinel();

        // User allocation decreases by WITHDRAWAL_AMOUNT; node balance released from channel.
        // nonHomeLedger.chainId must differ from block.chainid (home chain path via ChannelEngine).
        State memory candidate = TestUtils.nextState(
            initState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [DEPOSIT_AMOUNT - WITHDRAWAL_AMOUNT, uint256(0)],
            [int256(DEPOSIT_AMOUNT), -int256(WITHDRAWAL_AMOUNT)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [uint256(0), uint256(0)],
            [-int256(WITHDRAWAL_AMOUNT), int256(WITHDRAWAL_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.finalizeEscrowWithdrawal(channelId, escrowId, candidate);

        _assertPurgeInvoked();
    }

    // ======== Tests: non-home-chain escrow flows ========

    function test_purgeInvoked_onInitiateEscrowDeposit() public {
        _snapshotAndInjectSentinel();
        _initiateEscrowDeposit();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onChallengeEscrowDeposit() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowDeposit();
        _snapshotAndInjectSentinel();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);
        vm.prank(bob);
        cHub.challengeEscrowDeposit(escrowId, sig, ParticipantIndex.USER);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowDeposit_cooperative() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowDeposit();
        _snapshotAndInjectSentinel();

        State memory candidate = TestUtils.nextState(
            initState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(0), int256(DEPOSIT_AMOUNT)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [int256(DEPOSIT_AMOUNT), -int256(DEPOSIT_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, bobChannelId, BOB_PK);
        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowDeposit_unilateral() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowDeposit();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);
        vm.prank(bob);
        cHub.challengeEscrowDeposit(escrowId, sig, ParticipantIndex.USER);

        vm.warp(block.timestamp + EscrowDepositEngine.CHALLENGE_DURATION + 1);
        _snapshotAndInjectSentinel();

        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, initState);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onInitiateEscrowWithdrawal() public {
        _snapshotAndInjectSentinel();
        _initiateEscrowWithdrawal();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onChallengeEscrowWithdrawal() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawal();
        _snapshotAndInjectSentinel();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);
        vm.prank(bob);
        cHub.challengeEscrowWithdrawal(escrowId, sig, ParticipantIndex.USER);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowWithdrawal_cooperative() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawal();
        _snapshotAndInjectSentinel();

        // User allocation drops to 0 (full amount withdrawn); node NF decreases by DEPOSIT_AMOUNT.
        // nonHomeLedger.chainId == block.chainid (non-home chain path via EscrowWithdrawalEngine).
        State memory candidate = TestUtils.nextState(
            initState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(0), uint256(0)],
            [int256(DEPOSIT_AMOUNT), -int256(DEPOSIT_AMOUNT)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [-int256(DEPOSIT_AMOUNT), int256(DEPOSIT_AMOUNT)]
        );
        candidate = mutualSignStateBothWithEcdsaValidator(candidate, bobChannelId, BOB_PK);
        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, candidate);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeEscrowWithdrawal_unilateral() public {
        (bytes32 escrowId, State memory initState) = _initiateEscrowWithdrawal();

        bytes memory sig = signChallengeEip191WithEcdsaValidator(bobChannelId, initState, BOB_PK);
        vm.prank(bob);
        cHub.challengeEscrowWithdrawal(escrowId, sig, ParticipantIndex.USER);

        vm.warp(block.timestamp + EscrowWithdrawalEngine.CHALLENGE_DURATION + 1);
        _snapshotAndInjectSentinel();

        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, initState);

        _assertPurgeInvoked();
    }

    // ======== Tests: migration flows ========

    function test_purgeInvoked_onInitiateMigration_homeChain() public {
        State memory prevState = _createSimpleChannel();
        _snapshotAndInjectSentinel();
        _initiateMigrationHomeChain(prevState);
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onInitiateMigration_nonHomeChain() public {
        _snapshotAndInjectSentinel();
        _initiateMigrationNonHomeChain();
        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeMigration_newHomeChain() public {
        // storedInitState has homeLedger=LOCAL (block.chainid), nonHomeLedger=FOREIGN
        State memory storedInitState = _initiateMigrationNonHomeChain();
        _snapshotAndInjectSentinel();

        // userMigratedAlloc = storedInitState.nonHomeLedger.userAlloc = DEPOSIT_AMOUNT
        // userNfDelta = 0 (userNF must equal storedInitState.homeLedger.userNF = 0)
        // nodeNfDelta = 0 (nodeNF must equal storedInitState.homeLedger.nodeNF = DEPOSIT_AMOUNT)
        State memory finalizeState = TestUtils.nextState(
            storedInitState,
            StateIntent.FINALIZE_MIGRATION,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(0), int256(DEPOSIT_AMOUNT)],
            FOREIGN_CHAIN_ID,
            FOREIGN_TOKEN,
            [uint256(0), uint256(0)],
            [int256(0), int256(0)]
        );
        finalizeState = mutualSignStateBothWithEcdsaValidator(finalizeState, bobChannelId, BOB_PK);
        vm.prank(bob);
        cHub.finalizeMigration(bobChannelId, finalizeState);

        _assertPurgeInvoked();
    }

    function test_purgeInvoked_onFinalizeMigration_oldHomeChain() public {
        State memory prevState = _createSimpleChannel();
        State memory initState = _initiateMigrationHomeChain(prevState);
        _snapshotAndInjectSentinel();

        // Candidate has homeLedger=FOREIGN (new home), nonHomeLedger=LOCAL (block.chainid).
        // Contract detects nonHomeLedger.chainId == block.chainid → swaps before validation.
        // Swap initState ledgers to get a base whose homeLedger.chainId == FOREIGN_CHAIN_ID,
        // so that nextState preserves the correct chainId/token for the finalize candidate.
        State memory swappedBase = initState;
        swappedBase.homeLedger = initState.nonHomeLedger;
        swappedBase.nonHomeLedger = initState.homeLedger;

        State memory finalizeState = TestUtils.nextState(
            swappedBase,
            StateIntent.FINALIZE_MIGRATION,
            [DEPOSIT_AMOUNT, uint256(0)],
            [int256(DEPOSIT_AMOUNT), int256(0)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [int256(0), int256(0)]
        );
        finalizeState = mutualSignStateBothWithEcdsaValidator(finalizeState, channelId, ALICE_PK);
        vm.prank(alice);
        cHub.finalizeMigration(channelId, finalizeState);

        _assertPurgeInvoked();
    }

    // ======== Tests: public purge ========

    function test_purgeInvoked_onPurgeEscrowDeposits() public {
        _snapshotAndInjectSentinel();
        cHub.purgeEscrowDeposits(1);
        _assertPurgeInvoked();
    }

    // ======== Tests: does NOT invoke purge ========

    function test_purgeNotInvoked_onDepositToVault() public {
        _snapshotAndInjectSentinel();

        token.mint(node, DEPOSIT_AMOUNT);
        vm.startPrank(node);
        token.approve(address(cHub), DEPOSIT_AMOUNT);
        cHub.depositToVault(node, address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        _assertPurgeNotInvoked();
    }

    function test_purgeNotInvoked_onWithdrawFromVault() public {
        _snapshotAndInjectSentinel();

        vm.prank(node);
        cHub.withdrawFromVault(node, address(token), DEPOSIT_AMOUNT);

        _assertPurgeNotInvoked();
    }
}

// forge-lint: disable-end(unsafe-typecast)
