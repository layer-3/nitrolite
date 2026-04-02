// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {ChannelEngine} from "../../src/ChannelEngine.sol";
import {ChannelStatus, State, StateIntent, Ledger} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

contract ChannelEngineTest_ValidateTransition is Test {
    // ======== Helpers ========

    // Use native ETH (address(0), decimals=18) so validateTokenDecimals passes without a deployed contract.
    function _operatingCtx(uint256 lockedFunds) internal view returns (ChannelEngine.TransitionContext memory ctx) {
        ctx.status = ChannelStatus.OPERATING;
        ctx.prevState = State({
            version: 1,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(0),
                decimals: 18,
                userAllocation: lockedFunds,
                userNetFlow: int256(lockedFunds),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: TestUtils.emptyLedger(),
            userSig: "",
            nodeSig: ""
        });
        ctx.lockedFunds = lockedFunds;
        ctx.nodeAvailableFunds = 0;
    }

    // ======== CLOSE: zero-allocation enforcement ========

    function test_revert_close_nonZeroUserAllocation() public {
        ChannelEngine.TransitionContext memory ctx = _operatingCtx(1000);

        // Same net flows as prevState -> userNfDelta == 0, so no funds would move, but
        // userAllocation = 1000 is still positive — previously led to stuck funds.
        State memory candidate =
            TestUtils.nextState(ctx.prevState, StateIntent.CLOSE, [uint256(1000), uint256(0)], [int256(1000), int256(0)]);

        vm.expectRevert(ChannelEngine.IncorrectUserAllocation.selector);
        ChannelEngine.validateTransition(ctx, candidate);
    }

    function test_revert_close_nonZeroNodeAllocation() public {
        ChannelEngine.TransitionContext memory ctx = _operatingCtx(1000);

        // userAllocation = 0 (passes first check), but nodeAllocation = 1000 must fail.
        // netFlowsSum = 0 + 1000 = 1000 = allocsSum (passes universal validation)
        State memory candidate =
            TestUtils.nextState(ctx.prevState, StateIntent.CLOSE, [uint256(0), uint256(1000)], [int256(0), int256(1000)]);

        vm.expectRevert(ChannelEngine.IncorrectNodeAllocation.selector);
        ChannelEngine.validateTransition(ctx, candidate);
    }

    // ======== Token mismatch ========

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
