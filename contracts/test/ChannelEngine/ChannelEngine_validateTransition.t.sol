// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {ChannelEngine} from "../../src/ChannelEngine.sol";
import {ChannelStatus, State, StateIntent, Ledger} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

contract ChannelEngineTest_ValidateTransition is Test {
    function test_revert_tokenMismatch() public {
        address tokenA = address(0xAAAA);
        address tokenB = address(0xBBBB);

        ChannelEngine.TransitionContext memory ctx;
        ctx.status = ChannelStatus.OPERATING;
        ctx.prevState = State({
            version: 1,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: tokenA,
                decimals: 18,
                userAllocation: 1000,
                userNetFlow: int256(1000),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: TestUtils.emptyLedger(),
            userSig: "",
            nodeSig: ""
        });
        ctx.lockedFunds = 1000;
        ctx.nodeAvailableFunds = 10000;

        // Candidate switches to tokenB while keeping the same amount accounting,
        // which would drain tokenB from the hub's balance.
        State memory candidate =
            TestUtils.nextState(ctx.prevState, StateIntent.WITHDRAW, [uint256(0), uint256(0)], [int256(0), int256(0)]);
        candidate.homeLedger.token = tokenB;

        vm.expectRevert(ChannelEngine.TokenMismatch.selector);
        ChannelEngine.validateTransition(ctx, candidate);
    }
}
