// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestChannelHub} from "./TestChannelHub.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";
import {SessionKeyValidator} from "../src/sigValidators/SessionKeyValidator.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {DEFAULT_SIG_VALIDATOR_ID} from "../src/interfaces/Types.sol";

abstract contract ChannelHubTest_SigValidator_Base is Test {
    TestChannelHub public cHub;

    uint256 constant NODE_PK = 1;
    address node;

    ISignatureValidator immutable ECDSA_VALIDATOR = new ECDSAValidator();
    ISignatureValidator immutable SK_VALIDATOR = new SessionKeyValidator();

    uint8 constant VALIDATOR_ID = 1;

    function setUp() public virtual {
        node = vm.addr(NODE_PK);
        cHub = new TestChannelHub(ECDSA_VALIDATOR, node);
    }

    // -------- helpers --------

    function _registerValidator(uint8 validatorId, ISignatureValidator validator) internal {
        bytes memory sig = TestUtils.buildAndSignValidatorRegistration(vm, validatorId, address(validator), NODE_PK);
        cHub.registerNodeValidator(node, validatorId, validator, sig);
    }

    function _registerAndActivate(uint8 validatorId, ISignatureValidator validator) internal {
        _registerValidator(validatorId, validator);
        vm.warp(block.timestamp + cHub.VALIDATOR_ACTIVATION_DELAY() + 1);
    }

    /// @dev Build a 1-byte-prefixed signature with the given validatorId and dummy sig data.
    function _sig(uint8 validatorId) internal pure returns (bytes memory) {
        return abi.encodePacked(validatorId, bytes32(0), bytes32(0), uint8(27));
    }
}

contract ChannelHubTest_RegisterNodeValidator is ChannelHubTest_SigValidator_Base {
    function test_success_storesValidatorInfo() public {
        uint256 ts = 1_000_000;
        vm.warp(ts);

        _registerValidator(VALIDATOR_ID, SK_VALIDATOR);

        (ISignatureValidator stored, uint64 registeredAt) = cHub.getNodeValidator(node, VALIDATOR_ID);
        assertEq(address(stored), address(SK_VALIDATOR));
        assertEq(registeredAt, ts);
    }

    function test_success_emitsValidatorRegistered() public {
        bytes memory sig = TestUtils.buildAndSignValidatorRegistration(vm, VALIDATOR_ID, address(SK_VALIDATOR), NODE_PK);

        vm.expectEmit(true, true, true, true);
        emit ChannelHub.ValidatorRegistered(node, VALIDATOR_ID, SK_VALIDATOR);

        cHub.registerNodeValidator(node, VALIDATOR_ID, SK_VALIDATOR, sig);
    }

    function test_revert_defaultValidatorId() public {
        bytes memory sig =
            TestUtils.buildAndSignValidatorRegistration(vm, DEFAULT_SIG_VALIDATOR_ID, address(SK_VALIDATOR), NODE_PK);

        vm.expectRevert(ChannelHub.InvalidValidatorId.selector);
        cHub.registerNodeValidator(node, DEFAULT_SIG_VALIDATOR_ID, SK_VALIDATOR, sig);
    }

    function test_revert_zeroValidatorAddress() public {
        bytes memory sig = TestUtils.buildAndSignValidatorRegistration(vm, VALIDATOR_ID, address(0), NODE_PK);

        vm.expectRevert(ChannelHub.InvalidAddress.selector);
        cHub.registerNodeValidator(node, VALIDATOR_ID, ISignatureValidator(address(0)), sig);
    }

    function test_revert_duplicateValidatorId() public {
        _registerValidator(VALIDATOR_ID, SK_VALIDATOR);

        bytes memory sig2 =
            TestUtils.buildAndSignValidatorRegistration(vm, VALIDATOR_ID, address(SK_VALIDATOR), NODE_PK);

        vm.expectRevert(abi.encodeWithSelector(ChannelHub.ValidatorAlreadyRegistered.selector, node, VALIDATOR_ID));
        cHub.registerNodeValidator(node, VALIDATOR_ID, SK_VALIDATOR, sig2);
    }
}

contract ChannelHubTest_ExtractValidator is ChannelHubTest_SigValidator_Base {
    /// @dev approvedSignatureValidators bitmask with bit `id` set.
    function _approved(uint8 id) internal pure returns (uint256) {
        return 1 << id;
    }

    function test_success_defaultValidator() public view {
        bytes memory sig = _sig(DEFAULT_SIG_VALIDATOR_ID);

        ISignatureValidator result = cHub.exposed_extractValidator(sig, node, 0);

        assertEq(address(result), address(cHub.DEFAULT_SIG_VALIDATOR()));
    }

    function test_success_activeNodeValidator() public {
        _registerAndActivate(VALIDATOR_ID, SK_VALIDATOR);

        bytes memory sig = _sig(VALIDATOR_ID);

        ISignatureValidator result = cHub.exposed_extractValidator(sig, node, _approved(VALIDATOR_ID));

        assertEq(address(result), address(SK_VALIDATOR));
    }

    function test_revert_validatorNotRegistered() public {
        bytes memory sig = _sig(VALIDATOR_ID);

        vm.expectRevert(abi.encodeWithSelector(ChannelHub.ValidatorNotRegistered.selector, node, VALIDATOR_ID));
        cHub.exposed_extractValidator(sig, node, _approved(VALIDATOR_ID));
    }

    function test_revert_validatorNotYetActive() public {
        _registerValidator(VALIDATOR_ID, SK_VALIDATOR);

        // Advance time but stay just inside the delay
        vm.warp(block.timestamp + cHub.VALIDATOR_ACTIVATION_DELAY() - 1);

        bytes memory sig = _sig(VALIDATOR_ID);

        uint64 expectedActivatesAt = uint64(block.timestamp + 1);
        vm.expectRevert(
            abi.encodeWithSelector(ChannelHub.ValidatorNotActive.selector, node, VALIDATOR_ID, expectedActivatesAt)
        );
        cHub.exposed_extractValidator(sig, node, _approved(VALIDATOR_ID));
    }
}
