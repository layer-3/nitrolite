// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

import {ReentrantERC20} from "./mocks/ReentrantERC20.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";
import {
    ChannelDefinition,
    ChannelStatus,
    DEFAULT_SIG_VALIDATOR_ID,
    Ledger,
    ParticipantIndex,
    State,
    StateIntent
} from "../src/interfaces/Types.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../src/Utils.sol";

/**
 * @notice Regression tests for MF3-L19 (audit) — reentrancy via hook-bearing ERC20 tokens during
 *         the inbound `_pullFunds` callback. Each test exercises one of the four scenarios from
 *         `reentrancy-finding-validation.md`:
 *           1. Outer DEPOSIT pull → inner lifecycle call.
 *           2. Outer `challengeChannel(newer-version)` pull → inner lifecycle call.
 *           3. Outer escrow-deposit pull → inner lifecycle call.
 *           4. Outer `createChannel(DEPOSIT)` pull → inner lifecycle call.
 *
 *         The malicious token (`ReentrantERC20`) calls back into `ChannelHub` from inside
 *         `transferFrom` before the outer's state mutations and event emit. The expected behavior
 *         after the MF3-L19 remediation is that OpenZeppelin's `nonReentrant` modifier on the
 *         outer lifecycle function rejects the inner call with `ReentrancyGuardReentrantCall`.
 *
 *         To make these tests target the guard rather than secondary checks, the inner reentry
 *         payloads are calls that would otherwise reach guarded code paths quickly. We do not
 *         require the inner call to be fully signed — the guard fires before any signature
 *         validation.
 */
// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_Reentrancy is Test {
    ChannelHub public cHub;
    ReentrantERC20 public token;

    uint256 constant NODE_PK = 1;
    uint256 constant ALICE_PK = 2;

    address node;
    address alice;

    ISignatureValidator immutable ECDSA_SIG_VALIDATOR = new ECDSAValidator();

    uint8 constant CHANNEL_HUB_VERSION = 1;
    uint32 constant CHALLENGE_DURATION = 86400;
    uint64 constant NONCE = 1;
    uint256 constant INITIAL_BALANCE = 10000;

    bytes4 immutable REENTRANCY_GUARD_SELECTOR = ReentrancyGuard.ReentrancyGuardReentrantCall.selector;

    function setUp() public {
        node = vm.addr(NODE_PK);
        alice = vm.addr(ALICE_PK);

        cHub = new ChannelHub(ECDSA_SIG_VALIDATOR, node);
        token = new ReentrantERC20("Reentrant Token", "REENT");

        token.mint(node, INITIAL_BALANCE);
        token.mint(alice, INITIAL_BALANCE);

        vm.startPrank(node);
        token.approve(address(cHub), INITIAL_BALANCE);
        cHub.depositToNode(address(token), INITIAL_BALANCE);
        vm.stopPrank();

        vm.prank(alice);
        token.approve(address(cHub), INITIAL_BALANCE);
    }

    // ========== Helpers ==========

    function _buildDef() internal view returns (ChannelDefinition memory) {
        return ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 1,
            metadata: bytes32(0)
        });
    }

    function _signMutual(State memory state, bytes32 channelId) internal pure returns (State memory) {
        state.userSig = TestUtils.signStateEip191WithEcdsaValidator(vm, channelId, state, ALICE_PK);
        state.nodeSig = TestUtils.signStateEip191WithEcdsaValidator(vm, channelId, state, NODE_PK);
        return state;
    }

    function _initialDepositState(uint256 amount) internal view returns (State memory) {
        return State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: amount,
                userNetFlow: int256(amount),
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
    }

    function _openChannel(uint256 initialAmount) internal returns (bytes32 channelId, State memory state) {
        ChannelDefinition memory def = _buildDef();
        channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);
        state = _initialDepositState(initialAmount);
        state = _signMutual(state, channelId);

        vm.prank(alice);
        cHub.createChannel(def, state);
    }

    // ========== Scenario 1: createChannel(DEPOSIT) outer → inner depositToChannel ==========

    function test_reentrancy_createChannel_rejectsInnerDepositToChannel() public {
        ChannelDefinition memory def = _buildDef();
        bytes32 channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);
        State memory state = _initialDepositState(1000);
        state = _signMutual(state, channelId);

        // Inner call: depositToChannel with empty State data. The guard fires before any decoding,
        // so the precise inner state shape is irrelevant.
        State memory empty;
        bytes memory innerCalldata = abi.encodeCall(ChannelHub.depositToChannel, (channelId, empty));
        token.armReentry(address(cHub), innerCalldata);

        vm.prank(alice);
        cHub.createChannel(def, state);

        assertFalse(token.lastReentrySucceeded(), "inner depositToChannel must be rejected");
        bytes memory ret = token.lastReentryReturnData();
        assertGe(ret.length, 4, "inner revert returndata must contain a selector");
        bytes4 sel;
        assembly {
            sel := mload(add(ret, 0x20))
        }
        assertEq(sel, REENTRANCY_GUARD_SELECTOR, "inner revert must be ReentrancyGuardReentrantCall");
    }

    // ========== Scenario 2: depositToChannel outer → inner checkpointChannel ==========

    function test_reentrancy_depositToChannel_rejectsInnerCheckpointChannel() public {
        (bytes32 channelId, State memory state) = _openChannel(1000);

        // Build the next state — DEPOSIT bump.
        State memory next =
            TestUtils.nextState(state, StateIntent.DEPOSIT, [uint256(1500), uint256(0)], [int256(1500), int256(0)]);
        next = _signMutual(next, channelId);

        State memory empty;
        bytes memory innerCalldata = abi.encodeCall(ChannelHub.checkpointChannel, (channelId, empty));
        token.armReentry(address(cHub), innerCalldata);

        vm.prank(alice);
        cHub.depositToChannel(channelId, next);

        assertFalse(token.lastReentrySucceeded(), "inner checkpointChannel must be rejected");
        bytes memory ret = token.lastReentryReturnData();
        assertGe(ret.length, 4, "inner revert returndata must contain a selector");
        bytes4 sel;
        assembly {
            sel := mload(add(ret, 0x20))
        }
        assertEq(sel, REENTRANCY_GUARD_SELECTOR, "inner revert must be ReentrancyGuardReentrantCall");
    }

    // ========== Scenario 3: challengeChannel(newer-version) outer → inner closeChannel ==========

    function test_reentrancy_challengeChannel_newVersion_rejectsInnerCloseChannel() public {
        (bytes32 channelId, State memory state) = _openChannel(1000);

        // Newer DEPOSIT state to force pull (`challengeChannel` invokes `_applyTransitionEffects`
        // for higher-version candidates, which calls `_pullFunds` on a DEPOSIT intent).
        State memory newer =
            TestUtils.nextState(state, StateIntent.DEPOSIT, [uint256(1500), uint256(0)], [int256(1500), int256(0)]);
        newer = _signMutual(newer, channelId);

        // Challenger sig is over `signingData || "challenge"` (see ChannelHub_Base helper).
        bytes memory signingData = Utils.toSigningData(newer);
        bytes memory challengerSigningData = abi.encodePacked(signingData, "challenge");
        bytes memory challengerSigPayload =
            TestUtils.signEip191(vm, ALICE_PK, Utils.pack(channelId, challengerSigningData));
        bytes memory challengerSig = abi.encodePacked(DEFAULT_SIG_VALIDATOR_ID, challengerSigPayload);

        State memory empty;
        bytes memory innerCalldata = abi.encodeCall(ChannelHub.closeChannel, (channelId, empty));
        token.armReentry(address(cHub), innerCalldata);

        vm.prank(alice);
        cHub.challengeChannel(channelId, newer, challengerSig, ParticipantIndex.USER);

        assertFalse(token.lastReentrySucceeded(), "inner closeChannel must be rejected");
        bytes memory ret = token.lastReentryReturnData();
        assertGe(ret.length, 4, "inner revert returndata must contain a selector");
        bytes4 sel;
        assembly {
            sel := mload(add(ret, 0x20))
        }
        assertEq(sel, REENTRANCY_GUARD_SELECTOR, "inner revert must be ReentrancyGuardReentrantCall");
    }

    // ========== Scenario 4: initiateEscrowDeposit (non-home chain) outer → inner purgeEscrowDeposits ==========

    function test_reentrancy_initiateEscrowDeposit_nonHome_rejectsInnerPurge() public {
        // Construct an escrow-deposit candidate where homeLedger is on a different chain id so the
        // current chain is the non-home chain (pull path triggered).
        ChannelDefinition memory def = _buildDef();
        bytes32 channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);

        State memory candidate = State({
            version: 1,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid) + 1, // home is a different chain
                token: address(token),
                decimals: 18,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 500,
                nodeNetFlow: -500
            }),
            nonHomeLedger: Ledger({
                chainId: uint64(block.chainid), // non-home is this chain → pull occurs
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
        candidate = _signMutual(candidate, channelId);

        bytes memory innerCalldata = abi.encodeCall(ChannelHub.purgeEscrowDeposits, (uint256(1)));
        token.armReentry(address(cHub), innerCalldata);

        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, candidate);

        assertFalse(token.lastReentrySucceeded(), "inner purgeEscrowDeposits must be rejected");
        bytes memory ret = token.lastReentryReturnData();
        assertGe(ret.length, 4, "inner revert returndata must contain a selector");
        bytes4 sel;
        assembly {
            sel := mload(add(ret, 0x20))
        }
        assertEq(sel, REENTRANCY_GUARD_SELECTOR, "inner revert must be ReentrancyGuardReentrantCall");
    }
}
