// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {MockERC20} from "./mocks/MockERC20.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "./TestUtils.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";
import {SessionKeyValidator, SessionKeyAuthorization} from "../src/sigValidators/SessionKeyValidator.sol";
import {ChannelStatus, State, StateIntent, Ledger, DEFAULT_SIG_VALIDATOR_ID} from "../src/interfaces/Types.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../src/Utils.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_Base is Test {
    ChannelHub public cHub;
    MockERC20 public token;

    uint256 constant NODE_PK = 1;
    uint256 constant ALICE_PK = 2;
    uint256 constant ALICE_SK1_PK = 3;
    uint256 constant BOB_PK = 4;

    address node;
    address alice;
    address aliceSk1;
    address bob;

    ISignatureValidator immutable EMPTY_SIG_VALIDATOR = ISignatureValidator(address(0));
    ISignatureValidator immutable ECDSA_SIG_VALIDATOR = new ECDSAValidator();
    ISignatureValidator immutable SK_SIG_VALIDATOR = new SessionKeyValidator();

    uint8 constant CHANNEL_HUB_VERSION = 1;
    uint32 constant CHALLENGE_DURATION = 86400; // 1 day
    uint64 constant NONCE = 1;
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;

    function setUp() public virtual {
        // Deploy contracts
        cHub = new ChannelHub(ECDSA_SIG_VALIDATOR);
        token = new MockERC20("Test Token", "TST", 18);

        node = vm.addr(NODE_PK);
        alice = vm.addr(ALICE_PK);
        aliceSk1 = vm.addr(ALICE_SK1_PK);
        bob = vm.addr(BOB_PK);

        token.mint(node, INITIAL_BALANCE);
        token.mint(alice, INITIAL_BALANCE);
        token.mint(bob, INITIAL_BALANCE);

        vm.startPrank(node);
        token.approve(address(cHub), INITIAL_BALANCE);
        cHub.depositToVault(node, address(token), INITIAL_BALANCE);
        vm.stopPrank();

        // Register SessionKeyValidator for the node
        bytes memory skValidatorSig = TestUtils.buildAndSignValidatorRegistration(
            vm, SESSION_KEY_VALIDATOR_ID, address(SK_SIG_VALIDATOR), NODE_PK
        );
        cHub.registerNodeValidator(node, SESSION_KEY_VALIDATOR_ID, SK_SIG_VALIDATOR, skValidatorSig);

        vm.prank(alice);
        token.approve(address(cHub), INITIAL_BALANCE);

        vm.prank(bob);
        token.approve(address(cHub), INITIAL_BALANCE);
    }

    function mutualSignStateBothWithEcdsaValidator(State memory state, bytes32 channelId, uint256 userPk)
        internal
        pure
        returns (State memory)
    {
        state.userSig = TestUtils.signStateEip191WithEcdsaValidator(vm, channelId, state, userPk);
        state.nodeSig = TestUtils.signStateEip191WithEcdsaValidator(vm, channelId, state, NODE_PK);
        return state;
    }

    function mutualSignStateUserWithSkValidator(
        State memory state,
        bytes32 channelId,
        uint256 userPk,
        SessionKeyAuthorization memory skAuth
    ) internal pure returns (State memory) {
        state.userSig = TestUtils.signStateEip191WithSkValidator(vm, channelId, state, userPk, skAuth);
        state.nodeSig = TestUtils.signStateEip191WithEcdsaValidator(vm, channelId, state, NODE_PK);
        return state;
    }

    function signChallengeEip191WithEcdsaValidator(bytes32 channelId_, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory signingData = Utils.toSigningData(state);
        bytes memory challengerSigningData = abi.encodePacked(signingData, "challenge");
        bytes memory message = Utils.pack(channelId_, challengerSigningData);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, message);
        return abi.encodePacked(DEFAULT_SIG_VALIDATOR_ID, signature);
    }

    function verifyChannelData(
        bytes32 channelId,
        ChannelStatus expectedStatus,
        uint64 expectedVersion,
        uint256 expectedChallengeExpiry,
        string memory description
    ) internal view {
        (ChannelStatus status,, State memory latestState, uint256 challengeExpiry,) = cHub.getChannelData(channelId);
        assertEq(uint8(status), uint8(expectedStatus), string.concat(description, ": Channel status: "));
        assertEq(latestState.version, expectedVersion, string.concat(description, ": Channel version: "));
        assertEq(challengeExpiry, expectedChallengeExpiry, string.concat(description, ": Challenge expiry: "));
    }

    function verifyChannelState(
        bytes32 channelId,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        string memory description
    ) internal view {
        (,, State memory latestState,,) = cHub.getChannelData(channelId);
        assertEq(
            latestState.homeLedger.userAllocation, allocations[0], string.concat(description, ": User allocation: ")
        );
        assertEq(latestState.homeLedger.userNetFlow, netFlows[0], string.concat(description, ": User net flow: "));
        assertEq(
            latestState.homeLedger.nodeAllocation, allocations[1], string.concat(description, ": Node allocation: ")
        );
        assertEq(latestState.homeLedger.nodeNetFlow, netFlows[1], string.concat(description, ": Node net flow: "));

        uint256 nodeBalance = cHub.getAccountBalance(node, address(token));
        uint256 expectedNodeBalance =
            netFlows[1] < 0 ? INITIAL_BALANCE + uint256(-netFlows[1]) : INITIAL_BALANCE - uint256(netFlows[1]);
        assertEq(nodeBalance, expectedNodeBalance, string.concat(description, ": Node balance: "));
    }
}
