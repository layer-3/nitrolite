// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {Utils} from "../../src/Utils.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    ChannelStatus,
    EscrowStatus
} from "../../src/interfaces/Types.sol";

// These tests verify that escrow deposit and escrow withdrawal can be finalized on the non-home chain
// even after a migration has been initiated to that same chain (making it `MIGRATING_IN` and treated as
// home chain). Without the metadata-based routing fix, finalizeEscrowDeposit and
// finalizeEscrowWithdrawal would take the home chain path after migration, bypassing escrow metadata
// and permanently locking funds.
//
// The signing order in both tests reflects the realistic off-chain protocol scenario:
// the finalize state (v11) is pre-signed as the execution phase immediately after the initiate
// state (v10), before migration is ever signed. This means only ONE off-chain rule is broken
// (rule 21: node must not issue new states during a cross-chain op); version monotonicity is
// preserved in the signing sequence (v10 → v11 → v43). Migration is submitted on-chain before
// the pre-signed finalize is submitted, which triggers the bug the fix addresses.
//
// Signing order:    v10 → v11 (pre-signed, not yet submitted) → v43
// Submission order: v10 → v43 → v11

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_CrossChain_EscrowAfterMigration is ChannelHubTest_Base {
    bytes32 bobChannelId;
    ChannelDefinition bobDef;

    function setUp() public override {
        super.setUp();

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

    function test_depositEscrow_nonHomeChain_thenMigration() public {
        // ====== Step 1: Initiate escrow deposit on non-home chain ======
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID before escrow");

        uint256 bobBalanceBefore = token.balanceOf(bob);
        uint256 nodeVaultBefore = cHub.getAccountBalance(node, address(token));

        State memory escrowInitiateState = State({
            version: 10,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(42),
                decimals: 18,
                userAllocation: 958,
                userNetFlow: 1000,
                nodeAllocation: 500,
                nodeNetFlow: 458
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        escrowInitiateState = mutualSignStateBothWithEcdsaValidator(escrowInitiateState, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, escrowInitiateState.version);

        vm.prank(bob);
        cHub.initiateEscrowDeposit(bobDef, escrowInitiateState);

        assertEq(token.balanceOf(bob), bobBalanceBefore - 500, "User balance after escrow deposit initiation");
        (, EscrowStatus escrowStatus,,, uint256 lockedAmount,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        assertEq(lockedAmount, 500, "Escrow locked amount should be 500");

        // Pre-sign the finalize state (v11) as the execution phase — before migration is signed.
        // This preserves version monotonicity in the signing sequence (v10 → v11 → v43).
        // The state is not submitted yet; it will be submitted after migration is on-chain.
        State memory escrowFinalizeState = nextState(
            escrowInitiateState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(1458), uint256(0)],
            [int256(1000), int256(458)],
            uint64(block.chainid),
            address(token),
            18,
            [uint256(0), uint256(0)],
            [int256(500), int256(-500)]
        );
        escrowFinalizeState = mutualSignStateBothWithEcdsaValidator(escrowFinalizeState, bobChannelId, BOB_PK);

        // ====== Step 2: Initiate migration to this chain ======
        // Node breaks the "Flow suspension" rule by signing and submitting migration while escrow is pending.
        // After this call, _isChannelHomeChain returns true for bobChannelId on the current chain,
        // because the stored state has homeLedger.chainId == block.chainid (states are swapped on store).
        State memory migrationState = State({
            version: 43,
            intent: StateIntent.INITIATE_MIGRATION,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(42),
                decimals: 18,
                userAllocation: 469,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: -281
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 469,
                nodeNetFlow: 469
            }),
            userSig: "",
            nodeSig: ""
        });
        migrationState = mutualSignStateBothWithEcdsaValidator(migrationState, bobChannelId, BOB_PK);

        vm.prank(bob);
        cHub.initiateMigration(bobDef, migrationState);

        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.MIGRATING_IN), "Channel should be MIGRATING_IN after migration");
        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBefore - 469,
            "Node vault after migration (locked 469 for migration)"
        );

        // ====== Step 3: Finalize escrow deposit ======
        // _isChannelHomeChain now returns true, but _isEscrowDepositHomeChain returns false because
        // escrow metadata exists (channelId != 0). The non-home path is taken, correctly releasing
        // the locked 500 back to the node vault.
        // The finalize state was pre-signed before migration (see above); only the submission is here.
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        uint256 nodeVaultBeforeFinalize = cHub.getAccountBalance(node, address(token));

        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, escrowFinalizeState);

        assertEq(token.balanceOf(bob), bobBalanceBefore - 500, "User balance unchanged after escrow finalization");
        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBeforeFinalize + 500,
            "Node vault after escrow deposit finalization (500 returned)"
        );

        (, EscrowStatus finalStatus,,, uint256 finalLocked,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(finalLocked, 0, "Escrow locked amount should be 0");

        // Non-home path does not update channel state
        verifyChannelData(bobChannelId, ChannelStatus.MIGRATING_IN, 43, 0, "Channel should still be MIGRATING_IN");
    }

    function test_withdrawalEscrow_nonHomeChain_thenMigration() public {
        // ====== Step 1: Initiate escrow withdrawal on non-home chain ======
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID before escrow");

        uint256 bobBalanceBefore = token.balanceOf(bob);
        uint256 nodeVaultBefore = cHub.getAccountBalance(node, address(token));

        State memory escrowInitiateState = State({
            version: 10,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(42),
                decimals: 18,
                userAllocation: 1217,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: 467
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 750,
                nodeNetFlow: 750
            }),
            userSig: "",
            nodeSig: ""
        });
        escrowInitiateState = mutualSignStateBothWithEcdsaValidator(escrowInitiateState, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, escrowInitiateState.version);

        vm.prank(bob);
        cHub.initiateEscrowWithdrawal(bobDef, escrowInitiateState);

        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBefore - 750,
            "Node vault after escrow withdrawal initiation (locked 750)"
        );
        (, EscrowStatus escrowStatus,, uint256 lockedAmount,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        assertEq(lockedAmount, 750, "Escrow locked amount should be 750");

        // Pre-sign the finalize state (v11) as the execution phase — before migration is signed.
        // This preserves version monotonicity in the signing sequence (v10 → v11 → v43).
        // The state is not submitted yet; it will be submitted after migration is on-chain.
        State memory escrowFinalizeState = nextState(
            escrowInitiateState,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(467), uint256(0)],
            [int256(750), int256(-283)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [int256(-750), int256(750)]
        );
        escrowFinalizeState = mutualSignStateBothWithEcdsaValidator(escrowFinalizeState, bobChannelId, BOB_PK);

        // ====== Step 2: Initiate migration to this chain ======
        // Node breaks the "Flow suspension" rule by signing and submitting migration while escrow is pending.
        // After this call, _isChannelHomeChain returns true for bobChannelId on the current chain.
        State memory migrationState = State({
            version: 43,
            intent: StateIntent.INITIATE_MIGRATION,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(42),
                decimals: 18,
                userAllocation: 469,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: -281
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 469,
                nodeNetFlow: 469
            }),
            userSig: "",
            nodeSig: ""
        });
        migrationState = mutualSignStateBothWithEcdsaValidator(migrationState, bobChannelId, BOB_PK);

        vm.prank(bob);
        cHub.initiateMigration(bobDef, migrationState);

        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.MIGRATING_IN), "Channel should be MIGRATING_IN after migration");
        assertEq(
            cHub.getAccountBalance(node, address(token)),
            nodeVaultBefore - 750 - 469,
            "Node vault after migration (escrow 750 + migration 469)"
        );

        // ====== Step 3: Finalize escrow withdrawal ======
        // _isChannelHomeChain now returns true, but _isEscrowWithdrawalHomeChain returns false because
        // escrow metadata exists (channelId != 0). The non-home path is taken, correctly pushing the
        // locked 750 to the user.
        // The finalize state was pre-signed before migration (see above); only the submission is here.
        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, escrowFinalizeState);

        assertEq(token.balanceOf(bob), bobBalanceBefore + 750, "User balance after escrow withdrawal finalization");

        (, EscrowStatus finalStatus,, uint256 finalLocked,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(finalLocked, 0, "Escrow locked amount should be 0");

        // Non-home path does not update channel state; channel remains MIGRATING_IN
        verifyChannelData(bobChannelId, ChannelStatus.MIGRATING_IN, 43, 0, "Channel should still be MIGRATING_IN");
    }
}
