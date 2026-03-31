// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "../ChannelHub_Base.t.sol";

import {Utils} from "../../src/Utils.sol";
import {State, ChannelDefinition, StateIntent, Ledger, ChannelStatus} from "../../src/interfaces/Types.sol";
import {SessionKeyAuthorization} from "../../src/sigValidators/SessionKeyValidator.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "../TestUtils.sol";

contract ChannelHubTest_SingleChain_Lifecycle is ChannelHubTest_Base {
    function test_happyPath() public {
        // Approve SessionKeyValidator (ID 1) for user signatures by setting bit 1
        uint256 approvedValidators = 1 << SESSION_KEY_VALIDATOR_ID; // Bit 1 = 1

        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: approvedValidators,
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
                userAllocation: 1000,
                userNetFlow: 1000,
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

        // both sign with default validator
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        // Verify user balance after channel creation (deposited 1000)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1000, "User balance after channel creation");

        // transfer 42 (allocation decreases by 42, node net flow decreases by 42)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(958), uint256(0)], [int256(1000), int256(-42)]);

        // NOTE: user registers a Session Key
        SessionKeyAuthorization memory skAuth = TestUtils.buildAndSignSkAuth(vm, aliceSk1, bytes32(0), ALICE_PK);

        // NOTE: user signs with Session key validator
        state = mutualSignStateUserWithSkValidator(state, channelId, ALICE_SK1_PK, skAuth);

        // invoke a checkpoint
        // Expected: user allocation = 958, user net flow = 1000, node allocation = 0, node net flow = -42
        vm.prank(alice);
        cHub.checkpointChannel(channelId, state);
        verifyChannelState(channelId, [uint256(958), uint256(0)], [int256(1000), int256(-42)], "after checkpoint");

        // receive 24 (allocation increases by 24, node net flow increases by 24)
        state = TestUtils.nextState(state, StateIntent.OPERATE, [uint256(982), uint256(0)], [int256(1000), int256(-18)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // invoke a deposit (500)
        // Expected: user allocation = 1482, user net flow = 1500, node allocation = 0, node net flow = -18
        state =
            TestUtils.nextState(state, StateIntent.DEPOSIT, [uint256(1482), uint256(0)], [int256(1500), int256(-18)]);
        // NOTE: both sign with ECDSA validator
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.depositToChannel(channelId, state);
        verifyChannelState(channelId, [uint256(1482), uint256(0)], [int256(1500), int256(-18)], "after deposit");

        // Verify user balance after first deposit (deposited 500 more)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1500, "User balance after first deposit");

        // transfer 1
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1481), uint256(0)], [int256(1500), int256(-19)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // transfer 2
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1479), uint256(0)], [int256(1500), int256(-21)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // invoke a withdrawal (100)
        // Expected: user allocation = 1379, user net flow = 1400, node allocation = 0, node net flow = -21
        state =
            TestUtils.nextState(state, StateIntent.WITHDRAW, [uint256(1379), uint256(0)], [int256(1400), int256(-21)]);
        // NOTE: user signs with Session key validator
        state = mutualSignStateUserWithSkValidator(state, channelId, ALICE_SK1_PK, skAuth);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(channelId, [uint256(1379), uint256(0)], [int256(1400), int256(-21)], "after withdrawal");

        // Verify user balance after first withdrawal (withdrew 100)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1400, "User balance after first withdrawal");

        // transfer 3
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1376), uint256(0)], [int256(1400), int256(-24)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 10
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1386), uint256(0)], [int256(1400), int256(-14)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // invoke a deposit (200)
        // Expected: user allocation = 1586, user net flow = 1600, node allocation = 0, node net flow = -14
        state =
            TestUtils.nextState(state, StateIntent.DEPOSIT, [uint256(1586), uint256(0)], [int256(1600), int256(-14)]);
        // NOTE: user signs with Session key validator
        state = mutualSignStateUserWithSkValidator(state, channelId, ALICE_SK1_PK, skAuth);

        vm.prank(alice);
        cHub.depositToChannel(channelId, state);
        verifyChannelState(channelId, [uint256(1586), uint256(0)], [int256(1600), int256(-14)], "after second deposit");

        // Verify user balance after second deposit (deposited 200 more)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1600, "User balance after second deposit");

        // receive 1
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1587), uint256(0)], [int256(1600), int256(-13)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // transfer 2
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1585), uint256(0)], [int256(1600), int256(-15)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 3
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1588), uint256(0)], [int256(1600), int256(-12)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // transfer 4
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1584), uint256(0)], [int256(1600), int256(-16)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 5
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1589), uint256(0)], [int256(1600), int256(-11)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // withdraw (300)
        // Expected: user allocation = 1289, user net flow = 1300, node allocation = 0, node net flow = -11
        state =
            TestUtils.nextState(state, StateIntent.WITHDRAW, [uint256(1289), uint256(0)], [int256(1300), int256(-11)]);
        // NOTE: user signs with ECDSA validator
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(
            channelId, [uint256(1289), uint256(0)], [int256(1300), int256(-11)], "after second withdrawal"
        );

        // Verify user balance after second withdrawal (withdrew 300)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1300, "User balance after second withdrawal");

        // transfer 1
        // Expected: user allocation = 1288, user net flow = 1300, node allocation = 0, node net flow = -12
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1288), uint256(0)], [int256(1300), int256(-12)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // receive 2
        // Expected: user allocation = 1290, user net flow = 1300, node allocation = 0, node net flow = -10
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1290), uint256(0)], [int256(1300), int256(-10)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // transfer 3
        // Expected: user allocation = 1287, user net flow = 1300, node allocation = 0, node net flow = -13
        state =
            TestUtils.nextState(state, StateIntent.OPERATE, [uint256(1287), uint256(0)], [int256(1300), int256(-13)]);
        state = mutualSignStateBothWithEcdsaValidator(state, channelId, ALICE_PK);

        // close channel
        // Expected: allocations = 0, user net flow 13, node net flow -13
        state = TestUtils.nextState(state, StateIntent.CLOSE, [uint256(0), uint256(0)], [int256(13), int256(-13)]);

        // NOTE: user signs with Channel validator
        state = mutualSignStateUserWithSkValidator(state, channelId, ALICE_SK1_PK, skAuth);

        vm.prank(alice);
        cHub.closeChannel(channelId, state);

        // Check CLOSED status after channel closure
        (ChannelStatus finalStatus,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(finalStatus), uint8(ChannelStatus.CLOSED), "Channel should be CLOSED after closure");

        // Verify user balance after channel closure (received back 1287)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1300 + 1287, "User balance after channel closure");
    }

    function test_create_withOperateIntent() public {
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
        (ChannelStatus statusBefore,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(statusBefore), uint8(ChannelStatus.VOID), "Channel should be VOID before creation");

        // Verify user balance before channel creation
        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance before channel creation");

        // Non-initial state: imagine Alice has received some funds off-chain before channel creation
        State memory state = State({
            version: 16,
            intent: StateIntent.OPERATE,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 1000,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 1000
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

        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance after OPERATE creation stays the same");
        verifyChannelState(
            channelId, [uint256(1000), uint256(0)], [int256(0), int256(1000)], "after create with OPERATE intent"
        );
        (ChannelStatus status,, State memory latestState,, uint256 lockedFunds) = cHub.getChannelData(channelId);
        assertEq(
            uint8(status), uint8(ChannelStatus.OPERATING), "Channel created with OPERATE intent should be OPERATING"
        );
        assertEq(latestState.version, 16, "Channel created with OPERATE intent should have correct version");
        assertEq(
            uint8(latestState.intent),
            uint8(StateIntent.OPERATE),
            "Channel created with OPERATE intent should have correct intent"
        );
        assertEq(lockedFunds, 1000, "Channel created with OPERATE intent should have correct locked funds");
    }

    function test_create_withDepositIntent() public {
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
        (ChannelStatus statusBefore,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(statusBefore), uint8(ChannelStatus.VOID), "Channel should be VOID before creation");

        // Verify user balance before channel creation
        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance before channel creation");

        // Non-initial state: imagine Alice has received some funds off-chain before channel creation
        State memory state = State({
            version: 42,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 1500,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 1000
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

        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 500, "User balance after DEPOSIT creation decreases");
        verifyChannelState(
            channelId, [uint256(1500), uint256(0)], [int256(500), int256(1000)], "after create with DEPOSIT intent"
        );
        (ChannelStatus status,, State memory latestState,, uint256 lockedFunds) = cHub.getChannelData(channelId);
        assertEq(
            uint8(status), uint8(ChannelStatus.OPERATING), "Channel created with DEPOSIT intent should be OPERATING"
        );
        assertEq(latestState.version, 42, "Channel created with DEPOSIT intent should have correct version");
        assertEq(
            uint8(latestState.intent),
            uint8(StateIntent.DEPOSIT),
            "Channel created with DEPOSIT intent should have correct intent"
        );
        assertEq(lockedFunds, 1500, "Channel created with DEPOSIT intent should have correct locked funds");
    }

    function test_create_withWithdrawIntent() public {
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
        (ChannelStatus statusBefore,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(statusBefore), uint8(ChannelStatus.VOID), "Channel should be VOID before creation");

        // Verify user balance before channel creation
        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance before channel creation");

        // Non-initial state: imagine Alice has received some funds off-chain before channel creation
        State memory state = State({
            version: 24,
            intent: StateIntent.WITHDRAW,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 500,
                userNetFlow: -500,
                nodeAllocation: 0,
                nodeNetFlow: 1000
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

        assertEq(token.balanceOf(alice), INITIAL_BALANCE + 500, "User balance after WITHDRAW creation increases");
        verifyChannelState(
            channelId, [uint256(500), uint256(0)], [int256(-500), int256(1000)], "after create with WITHDRAW intent"
        );
        (ChannelStatus status,, State memory latestState,, uint256 lockedFunds) = cHub.getChannelData(channelId);
        assertEq(
            uint8(status), uint8(ChannelStatus.OPERATING), "Channel created with WITHDRAW intent should be OPERATING"
        );
        assertEq(latestState.version, 24, "Channel created with WITHDRAW intent should have correct version");
        assertEq(
            uint8(latestState.intent),
            uint8(StateIntent.WITHDRAW),
            "Channel created with WITHDRAW intent should have correct intent"
        );
        assertEq(lockedFunds, 500, "Channel created with WITHDRAW intent should have correct locked funds");
    }
}
