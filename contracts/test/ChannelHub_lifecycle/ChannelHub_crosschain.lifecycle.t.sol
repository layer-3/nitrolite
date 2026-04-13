// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";
import {MockERC20} from "../mocks/MockERC20.sol";

import {Utils} from "../../src/Utils.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    ChannelStatus,
    EscrowStatus
} from "../../src/interfaces/Types.sol";
import {TestUtils} from "../TestUtils.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_CrossChain_Lifecycle is ChannelHubTest_Base {
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

    function test_happyPath_homeChain() public {
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });

        bytes32 channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);

        // Check VOID status before channel creation
        (ChannelStatus status,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID before creation");

        // Verify user balance before channel creation
        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance before channel creation");

        // Initial state: alice deposits 1000
        // Expected: user allocation = 1000, user net flow = 1000, node allocation = 0, node net flow = 0
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
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: 0,
                token: address(0),
                decimals: 0,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });

        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        // Verify user balance after channel creation (deposited 1000)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1000, "User balance after channel creation");

        // transfer 42 (allocation decreases by 42, node net flow decreases by 42)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(958), uint256(0)], [int256(1000), int256(-42)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // deposit from another chain
        state = TestUtils.nextState(
            state,
            StateIntent.INITIATE_ESCROW_DEPOSIT,
            // user amounts stay the same, node amounts increase by 500
            [uint256(958), uint256(500)],
            [int256(1000), int256(458)],
            42,
            address(42), // chainId 42, token address 42 for simplicity
            // user deposit amount appear in allocation and net flow on non-home side
            [uint256(500), uint256(0)],
            [int256(500), int256(0)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // on chainId 42:
        // channelsHub.initiateEscrowDeposit(channelId, state)
        // NOTE: see a `test_depositEscrow_nonHomeChain` test for that

        // initiate escrow deposit on home chain
        // Expected: user allocation = 958, user net flow = 1000, node allocation = 0, node net flow = -42
        vm.prank(node);
        cHub.initiateEscrowDeposit(def, state);
        verifyChannelState(
            channelId, [uint256(958), uint256(500)], [int256(1000), int256(458)], "after cross chain deposit"
        );

        // finalize escrow deposit
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            // user allocation amount increases by cross-chain deposit, node allocation goes to 0
            [uint256(1458), uint256(0)],
            [int256(1000), int256(458)],
            42,
            address(42),
            // user deposit amount is zeroed, and withdrawn (unlocked) by node via net flow on non-home side
            [uint256(0), uint256(0)],
            [int256(500), int256(-500)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 24 (allocation increases by 24, node net flow increases by 24)
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1482), uint256(0)], [int256(1000), int256(482)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // send 12 (allocation decreases by 12, node net flow decreases by 12)
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1470), uint256(0)], [int256(1000), int256(470)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // withdraw 250 on home chain
        // Expected: user allocation = 1220, user net flow = 750, node allocation = 0, node net flow = 470
        state =
            TestUtils.nextState(state, StateIntent.WITHDRAW, [uint256(1220), uint256(0)], [int256(750), int256(470)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(channelId, [uint256(1220), uint256(0)], [int256(750), int256(470)], "after withdrawal");

        // Verify user balance after withdrawal (withdrew 250)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 750, "User balance after withdrawal");

        // send 2 (allocation decreases by 2, node net flow decreases by 2)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1218), uint256(0)], [int256(750), int256(468)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 3 (allocation increases by 3, node net flow increases by 3)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1221), uint256(0)], [int256(750), int256(471)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // send 4 (allocation decreases by 4, node net flow decreases by 4)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1217), uint256(0)], [int256(750), int256(467)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // withdrawal to another chain
        state = TestUtils.nextState(
            state,
            StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            // home chain stays the same
            [uint256(1217), uint256(0)],
            [int256(750), int256(467)],
            42,
            address(42), // chainId 42, token address 42 for simplicity
            // node deposits withdrawal amount
            [uint256(0), uint256(750)],
            [int256(0), int256(750)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // on chainId 42:
        // channelsHub.initiateEscrowWithdrawal(channelId, state)
        // NOTE: see a `test_withdrawalEscrow_nonHomeChain` test for that

        // finalize escrow withdrawal on another chain
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            // user allocation decreases by withdrawal amount, node allocation stays 0, node net flow decreases by withdrawal amount
            [uint256(467), uint256(0)],
            [int256(750), int256(-283)],
            42,
            address(42), // chainId 42, token address 42 for simplicity
            // user withdraws the amount (negative net flow)
            [uint256(0), uint256(0)],
            [int256(-750), int256(750)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 10 (allocation increases by 10, node net flow increases by 10)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(477), uint256(0)], [int256(750), int256(-273)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // checkpoint on home chain
        vm.prank(alice);
        cHub.checkpointChannel(channelId, state);
        verifyChannelState(channelId, [uint256(477), uint256(0)], [int256(750), int256(-273)], "after checkpoint");

        // Verify user balance hasn't changed
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 750, "User balance after checkpoint");

        // send 9 (allocation decreases by 9, node net flow decreases by 9)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(468), uint256(0)], [int256(750), int256(-282)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 8 (allocation increases by 8, node net flow increases by 8)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(476), uint256(0)], [int256(750), int256(-274)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // send 7 (allocation decreases by 7, node net flow decreases by 7)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(469), uint256(0)], [int256(750), int256(-281)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // migrate channel
        state = TestUtils.nextState(
            state,
            StateIntent.INITIATE_MIGRATION,
            // home chain stays the same
            [uint256(469), uint256(0)],
            [int256(750), int256(-281)],
            42,
            address(42), // chainId 42, token address 42 for simplicity
            // node deposits full user allocation amount
            [uint256(0), uint256(469)],
            [int256(0), int256(469)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // on chainId 42:
        // channelsHub.initiateMigrationIn(channelId, state)
        // NOTE: see a `test_migration_nonHomeChain` test for that

        // finalize migration on old home chain
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_MIGRATION,
            // channel closes on old home chain, allocations go to 0, net flows balance out
            [uint256(0), uint256(0)],
            [int256(750), int256(-750)],
            42,
            address(42), // chainId 42, token address 42 for simplicity
            // user receives allocation on new home chain
            [uint256(469), uint256(0)],
            [int256(0), int256(469)]
        );
        // home state and non-home state are swapped
        Ledger memory temp = state.homeLedger;
        state.homeLedger = state.nonHomeLedger;
        state.nonHomeLedger = temp;
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(node);
        cHub.finalizeMigration(channelId, state);

        // Verify channel is migrated out
        verifyChannelState(channelId, [uint256(0), uint256(0)], [int256(750), int256(-750)], "after migration");

        // Verify user balance hasn't changed (migration doesn't move funds on home chain)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 750, "User balance after migration");

        // Check MIGRATED_OUT status after channel was migrated
        (ChannelStatus finalStatus,,,,) = cHub.getChannelData(channelId);
        assertEq(
            uint8(finalStatus), uint8(ChannelStatus.MIGRATED_OUT), "Channel should be MIGRATED_OUT after migration"
        );
    }

    function test_depositEscrow_nonHomeChain() public {
        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        // Verify user balance before deposit
        assertEq(token.balanceOf(bob), INITIAL_BALANCE, "User balance before escrow deposit");

        // state from the "happyPath" test, but with home and nonHome states swapped
        State memory state = State({
            version: 42,
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
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state.version);

        // verify no escrow struct exists yet
        (, EscrowStatus escrowStatus,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        vm.prank(bob);
        cHub.initiateEscrowDeposit(bobDef, state);

        // Verify user balance after deposit (deposited 500)
        assertEq(token.balanceOf(bob), INITIAL_BALANCE - 500, "User balance after escrow deposit");

        // Verify escrow struct is updated on ChannelsHub
        (
            ,
            EscrowStatus finalEscrowStatus,
            uint64 unlockAt,
            uint64 challengeExpiresAt,
            uint256 lockedAmount,
            State memory initState
        ) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        uint64 expectedUnlockAt = uint64(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY());
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt is incorrect");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 500, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        // ====== finalize escrow deposit ======
        // this is an explicit action
        // escrow deposit locked funds should also be unlocked after `unlockAt` time passes alongside any other on-chain call
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));

        // state from the "happyPath" test, but with home and nonHome states swapped
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(1458), uint256(0)],
            [int256(1000), int256(458)],
            uint64(block.chainid),
            address(token),
            18, // nonHomeDecimals
            [uint256(0), uint256(0)],
            [int256(500), int256(-500)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, state);

        // Verify user balance after deposit finalized has NOT changed
        assertEq(token.balanceOf(bob), INITIAL_BALANCE - 500, "User balance after escrow deposit finalized");

        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore + 500, "Node balance after escrow deposit finalized");

        // Verify escrow struct is updated on ChannelsHub
        (, finalEscrowStatus, unlockAt, challengeExpiresAt, lockedAmount, initState) =
            cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt should remain unchanged");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount should have been zeroed");
        assertEq(initState.version, 42, "Escrow initState version should remain unchanged");
    }

    function test_depositEscrow_nonHomeChain_diffDecimals() public {
        // Home chain: 6 decimals (USDC), Non-home chain (current test chain): 14 decimals

        // Setup: Deploy 14-decimal token for non-home chain (this chain)
        MockERC20 token14dec = new MockERC20("Token14", "TK14", 14);
        token14dec.mint(bob, 1000 * 1e14);
        vm.prank(bob);
        token14dec.approve(address(cHub), 1000 * 1e14);

        // Check VOID status on non-home chain
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        // Verify user balance before deposit
        assertEq(token14dec.balanceOf(bob), 1000 * 1e14, "User balance before escrow deposit");

        // Create initial state representing home chain (chainId 42, 6 decimals USDC)
        // Home chain has existing channel with 50 USDC (50e6)
        // Bob wants to deposit 10 tokens on non-home chain (10e14 with 14 decimals = 10e6 with 6 decimals on home)
        State memory state = State({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(0x42), // USDC token on home chain
                decimals: 6,
                userAllocation: 50 * 1e6, // Existing user allocation
                userNetFlow: int256(50 * 1e6),
                nodeAllocation: 10 * 1e6, // Node locks 10 USDC for cross-chain deposit
                nodeNetFlow: int256(10 * 1e6)
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token14dec),
                decimals: 14,
                userAllocation: 10 * 1e14, // User deposits 10 tokens (14 decimals)
                userNetFlow: int256(10 * 1e14),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state.version);

        // Verify no escrow struct exists yet
        (, EscrowStatus escrowStatus,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        // Initiate escrow deposit on non-home chain
        vm.prank(bob);
        cHub.initiateEscrowDeposit(bobDef, state);

        // Verify user balance after deposit (deposited 10 tokens with 14 decimals)
        assertEq(token14dec.balanceOf(bob), 990 * 1e14, "User balance after escrow deposit");

        // Verify escrow struct is updated on ChannelsHub
        (
            ,
            EscrowStatus finalEscrowStatus,
            uint64 unlockAt,
            uint64 challengeExpiresAt,
            uint256 lockedAmount,
            State memory initState
        ) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        uint64 expectedUnlockAt = uint64(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY());
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt is incorrect");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 10 * 1e14, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        // ====== Finalize escrow deposit ======
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token14dec));

        // After finalization, home chain user allocation increases, non-home releases funds to node
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            [uint256(60 * 1e6), uint256(0)], // Home: user allocation increases by 10 USDC
            [int256(50 * 1e6), int256(10 * 1e6)], // Home: net flows stay same
            uint64(block.chainid),
            address(token14dec),
            14, // nonHomeDecimals
            [uint256(0), uint256(0)], // Non-home: allocations zero out
            [int256(10 * 1e14), int256(-10 * 1e14)] // Non-home: node releases deposited amount
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(node);
        cHub.finalizeEscrowDeposit(bobChannelId, escrowId, state);

        // Verify user balance after deposit finalized has NOT changed
        assertEq(token14dec.balanceOf(bob), 990 * 1e14, "User balance after escrow deposit finalized");

        // Verify node received the deposited tokens
        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token14dec));
        assertEq(nodeBalanceAfter, nodeBalanceBefore + 10 * 1e14, "Node balance after escrow deposit finalized");

        // Verify escrow struct is updated on ChannelsHub
        (, finalEscrowStatus, unlockAt, challengeExpiresAt, lockedAmount, initState) =
            cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt should remain unchanged");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount should have been zeroed");
        assertEq(initState.version, 42, "Escrow initState version should remain unchanged");
    }

    function test_withdrawalEscrow_nonHomeChain() public {
        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));

        // state from the "happyPath" test, but with home and nonHome states swapped
        State memory state = State({
            version: 42,
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
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state.version);

        // verify no escrow struct exists yet
        (, EscrowStatus escrowStatus,,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        vm.prank(bob);
        cHub.initiateEscrowWithdrawal(bobDef, state);

        // Verify user node's after deposit (deposited 500)
        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore - 750, "Node balance after escrow withdrawal");

        // Verify escrow struct is updated on ChannelsHub: escrow data exists, `locked` equals to withdrawalAmount
        (, EscrowStatus finalEscrowStatus, uint64 challengeExpireAt, uint256 lockedAmount, State memory initState) =
            cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 750, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        uint256 bobBalanceBefore = token.balanceOf(bob);

        // finalize escrow withdrawal on another chain
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(467), uint256(0)],
            [int256(750), int256(-283)],
            uint64(block.chainid),
            address(token),
            [uint256(0), uint256(0)],
            [int256(-750), int256(750)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, state);

        // Verify user balance after withdrawal (withdrew 750)
        uint256 bobBalanceAfter = token.balanceOf(bob);
        assertEq(bobBalanceAfter, bobBalanceBefore + 750, "User balance after escrow withdrawal");

        // Verify escrow struct is updated on ChannelsHub: escrow data exists, has status "finalized", `locked` equals to 0
        (, finalEscrowStatus, challengeExpireAt, lockedAmount, initState) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount is incorrect");
        assertEq(initState.version, 42, "Escrow initState  should not have changed");
    }

    function test_withdrawalEscrow_nonHomeChain_diffDecimals() public {
        // Home chain: 2 decimals, Non-home chain (current test chain): 8 decimals

        // Setup: Deploy 8-decimal token for non-home chain (this chain)
        MockERC20 token8dec = new MockERC20("Token8", "TK8", 8);
        token8dec.mint(bob, 1000 * 1e8);
        vm.prank(bob);
        token8dec.approve(address(cHub), 1000 * 1e8);

        // Setup: Node needs funds in 8-decimal token for non-home chain
        vm.startPrank(node);
        token8dec.mint(node, 100 * 1e8);
        token8dec.approve(address(cHub), 100 * 1e8);
        cHub.depositToNode(address(token8dec), 100 * 1e8);
        vm.stopPrank();

        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token8dec));

        // Bob wants to withdraw 5 tokens on non-home chain (5e8 with 8 decimals = 5e2 with 2 decimals)
        State memory state = State({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: 42,
                token: address(0x42), // Some token on home chain
                decimals: 2,
                userAllocation: 10 * 1e2,
                userNetFlow: int256(10 * 1e2),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token8dec),
                decimals: 8,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 5 * 1e8, // Node locks 5 tokens (5e8 with 8 decimals)
                nodeNetFlow: int256(5 * 1e8)
            }),
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state.version);

        // Verify no escrow exists yet
        (, EscrowStatus escrowStatus,,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        // Initiate escrow withdrawal on non-home chain
        vm.prank(bob);
        cHub.initiateEscrowWithdrawal(bobDef, state);

        // Verify node locked the withdrawal amount
        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token8dec));
        assertEq(nodeBalanceAfter, nodeBalanceBefore - 5 * 1e8, "Node balance after escrow withdrawal initiation");

        // Verify escrow struct is created
        (, EscrowStatus finalEscrowStatus, uint64 challengeExpireAt, uint256 lockedAmount, State memory initState) =
            cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 5 * 1e8, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        uint256 bobBalanceBefore = token8dec.balanceOf(bob);

        // Finalize escrow withdrawal on non-home chain
        // After withdrawal, user allocation on home decreases by 5 tokens (5e2 with 2 decimals)
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(5 * 1e2), uint256(0)], // Home: user allocation decreased by 5 tokens
            [int256(10 * 1e2), int256(-5 * 1e2)], // Home: node net flow decreased by 5 tokens
            uint64(block.chainid),
            address(token8dec),
            8, // nonHomeDecimals
            [uint256(0), uint256(0)],
            [int256(-5 * 1e8), int256(5 * 1e8)] // Non-home: user withdraws, node releases
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(bobChannelId, escrowId, state);

        // Verify user received the withdrawal
        uint256 bobBalanceAfter = token8dec.balanceOf(bob);
        assertEq(bobBalanceAfter, bobBalanceBefore + 5 * 1e8, "User balance after withdrawal");

        // Verify escrow is finalized
        (, finalEscrowStatus, challengeExpireAt, lockedAmount, initState) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount should be zero");
        assertEq(initState.version, 42, "Escrow initState should not have changed");
    }

    function test_migration_nonHomeChain() public {
        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        uint256 nodeBalanceBefore = cHub.getNodeBalance(address(token));
        uint256 userBalanceBefore = token.balanceOf(bob);

        // state from the "happyPath" test
        State memory state = State({
            version: 42,
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
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(bob);
        cHub.initiateMigration(bobDef, state);

        // Verify node's balance after migration (should have locked 469)
        uint256 nodeBalanceAfter = cHub.getNodeBalance(address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore - 469, "Node balance after migration initiation");

        // user balance should not have changed
        uint256 userBalanceAfter = token.balanceOf(bob);
        assertEq(userBalanceAfter, userBalanceBefore, "User balance after migration initiation");

        // Check MIGRATING_IN status
        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.MIGRATING_IN), "Channel should be MIGRATING_IN after migration");

        // sign finalize migration state by swapping homeLedger and nonHomeLedger, and swapping allocations
        state = State({
            version: 43,
            intent: StateIntent.FINALIZE_MIGRATION,
            metadata: bytes32(0),
            nonHomeLedger: Ledger({
                chainId: 42,
                token: address(42),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: -750
            }),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 469,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 469
            }),
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // perform some operations to verify channel is operating as normal
        // send 9 (allocation decreases by 9, node net flow decreases by 9)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(460), uint256(0)], [int256(0), int256(460)]);
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // receive 8 (allocation increases by 8, node net flow increases by 8)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(468), uint256(0)], [int256(0), int256(468)]);
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // send 7 (allocation decreases by 7, node net flow decreases by 7)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(461), uint256(0)], [int256(0), int256(461)]);
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // withdraw 400 on home chain
        // Expected: user allocation = 61, user net flow = -400, node allocation = 0, node net flow = 461
        state = TestUtils.nextState(state, StateIntent.WITHDRAW, [uint256(61), uint256(0)], [int256(-400), int256(461)]);
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(bob);
        cHub.withdrawFromChannel(bobChannelId, state);
        verifyChannelState(bobChannelId, [uint256(61), uint256(0)], [int256(-400), int256(461)], "after withdrawal");

        // Verify user balance after withdrawal (withdrew 400)
        assertEq(token.balanceOf(bob), userBalanceBefore + 400, "User balance after withdrawal");

        // Verify channel is still operating after migration and withdrawal
        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.OPERATING), "Channel should be OPERATING after withdrawal");
    }

    function test_migration_nonHomeChain_DiffDecimals() public {
        // Setup: Deploy token with 10 decimals on Old Home Chain
        MockERC20 token10dec = new MockERC20("Token10", "TK10", 10);
        token10dec.mint(bob, 1000 * 1e10);
        vm.prank(bob);
        token10dec.approve(address(cHub), 1000 * 1e10);

        // Setup: Node needs funds in both tokens
        vm.startPrank(node);
        token10dec.mint(node, 100 * 1e10);
        token10dec.approve(address(cHub), 100 * 1e10);
        cHub.depositToNode(address(token10dec), 100 * 1e10);
        vm.stopPrank();

        // 1. Create Channel with 10-decimal token on Old Home Chain
        // Bob deposits 50 tokens (50e10)
        State memory state = State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token10dec),
                decimals: 10,
                userAllocation: 50 * 1e10,
                userNetFlow: int256(50 * 1e10),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
                chainId: 0,
                token: address(0),
                decimals: 0,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        vm.prank(bob);
        cHub.createChannel(bobDef, state);

        // Verify channel is created
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.OPERATING), "Channel should be OPERATING");

        // 2. Perform some operations to build up channel state
        // Transfer 5 tokens (allocation decreases, node net flow decreases)
        state = TestUtils.nextState(
            state, StateIntent.OPERATE, [uint256(45 * 1e10), uint256(0)], [int256(50 * 1e10), int256(-5 * 1e10)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // 3. Initiate Migration to New Home Chain (14 decimals)
        // Deploy 14-decimal token for new home chain
        MockERC20 token14dec = new MockERC20("Token14", "TK14", 14);

        // Node needs funds in 14-decimal token for new home chain
        vm.startPrank(node);
        token14dec.mint(node, 100 * 1e14);
        token14dec.approve(address(cHub), 100 * 1e14);
        cHub.depositToNode(address(token14dec), 100 * 1e14);
        vm.stopPrank();

        // Initiate migration: Old home has 45 tokens (45e10 with 10 decimals)
        // New home will have 45 tokens (45e14 with 14 decimals)
        // Node must lock equivalent value: 45e10 in WAD = 45e18, 45e14 in WAD = 45e18 ✓
        state = TestUtils.nextState(
            state,
            StateIntent.INITIATE_MIGRATION,
            [uint256(45 * 1e10), uint256(0)], // Old home stays the same
            [int256(50 * 1e10), int256(-5 * 1e10)],
            42,
            address(token14dec),
            14, // nonHomeDecimals
            [uint256(0), uint256(45 * 1e14)], // Node locks user allocation on new home
            [int256(0), int256(45 * 1e14)]
        );
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // Submit on old home chain
        vm.prank(bob);
        cHub.initiateMigration(bobDef, state);

        // Verify state is updated correctly on old home chain
        (,, State memory latestState,,) = cHub.getChannelData(bobChannelId);
        assertEq(latestState.version, 2, "State version should be 2");
        assertEq(latestState.homeLedger.userAllocation, 45 * 1e10, "User allocation unchanged on old home");
        assertEq(latestState.nonHomeLedger.nodeAllocation, 45 * 1e14, "Node locked 45e14 on new home");

        // 4. Finalize Migration on Old Home Chain
        // After migration completes, allocations zero out on old home
        state = TestUtils.nextState(
            state,
            StateIntent.FINALIZE_MIGRATION,
            [uint256(0), uint256(0)], // Old home: allocations zero out
            [int256(50 * 1e10), int256(-50 * 1e10)], // Old home: net flows balance
            42,
            address(token14dec),
            14, // nonHomeDecimals
            [uint256(45 * 1e14), uint256(0)], // New home: user receives allocation
            [int256(0), int256(45 * 1e14)]
        );
        // Swap home and non-home states as per migration protocol
        Ledger memory temp = state.homeLedger;
        state.homeLedger = state.nonHomeLedger;
        state.nonHomeLedger = temp;
        state = mutualSignStateBothWithEcdsaValidator(state, bobChannelId, BOB_PK);

        // Submit finalization on old home chain
        vm.prank(node);
        cHub.finalizeMigration(bobChannelId, state);

        // Verify channel is migrated out
        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.MIGRATED_OUT), "Channel should be MIGRATED_OUT");

        // Verify final state on old home chain
        (,, latestState,,) = cHub.getChannelData(bobChannelId);
        assertEq(latestState.version, 3, "State version should be 3");
        assertEq(latestState.homeLedger.userAllocation, 0, "User allocation should be 0 on old home");
        assertEq(latestState.homeLedger.nodeAllocation, 0, "Node allocation should be 0 on old home");
    }
}
