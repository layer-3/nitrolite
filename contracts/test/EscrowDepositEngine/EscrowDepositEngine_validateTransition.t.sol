// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {EscrowDepositEngine} from "../../src/EscrowDepositEngine.sol";
import {EscrowStatus, State, StateIntent, Ledger} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

contract EscrowDepositEngineTest_ValidateTransition is Test {
    // nonHomeLedger.token must match the token stored in initState.
    // Fund movement on finalize already reads the token from the stored initState,
    // so no direct drain is possible via this engine. But accepting a mismatched
    // token in the finalize state would corrupt on-chain state consistency.
    function test_revert_tokenMismatch_onFinalize() public {
        uint64 homeChainId = 42; // not the current chain
        address initToken = address(0xAAAA); // token locked at initiation
        address otherToken = address(0); // attacker substitutes native ETH

        EscrowDepositEngine.TransitionContext memory ctx;
        ctx.status = EscrowStatus.INITIALIZED;
        ctx.lockedAmount = 500;
        ctx.initState = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: homeChainId,
                token: address(0xDEAD),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 500,
                nodeNetFlow: int256(500)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: initToken,
                decimals: 18,
                userAllocation: 500,
                userNetFlow: int256(500),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });

        // Finalize candidate uses a different nonHomeLedger token.
        // Universal validation passes (address(0) is valid native token with decimals 18).
        // EscrowTokenMismatch fires inside _calculateFinalizeEffects.
        State memory candidate = TestUtils.nextState(
            ctx.initState,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(500), uint256(0)],
            [int256(0), int256(500)],
            uint64(block.chainid),
            otherToken, // different from initToken
            [uint256(0), uint256(0)],
            [int256(0), int256(0)]
        );

        vm.expectRevert(EscrowDepositEngine.EscrowTokenMismatch.selector);
        EscrowDepositEngine.validateTransition(ctx, candidate);
    }
}
