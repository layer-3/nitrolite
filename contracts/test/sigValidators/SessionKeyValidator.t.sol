// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestUtils} from "../TestUtils.sol";

import {
    SessionKeyValidator,
    SessionKeyAuthorization,
    toSigningData
} from "../../src/sigValidators/SessionKeyValidator.sol";
import {ValidationResult, VALIDATION_SUCCESS, VALIDATION_FAILURE} from "../../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../../src/Utils.sol";

contract SessionKeyValidatorTest_Base is Test {
    SessionKeyValidator public validator;

    uint256 constant USER_PK = 1;
    uint256 constant NODE_PK = 2;
    uint256 constant OTHER_SIGNER_PK = 3;
    uint256 constant SESSION_KEY1_PK = 4;
    uint256 constant SESSION_KEY2_PK = 5;

    address user;
    address node;
    address otherSigner;
    address sessionKey1;
    address sessionKey2;

    bytes32 constant CHANNEL_ID = keccak256("test-channel");
    bytes32 constant OTHER_CHANNEL_ID = keccak256("other-channel");
    bytes constant SIGNING_DATA = hex"1234567890abcdef";
    bytes constant OTHER_SIGNING_DATA = hex"abcdef1234567890";
    bytes32 constant METADATA_HASH = keccak256("metadata");
    bytes32 constant OTHER_METADATA_HASH = keccak256("other-metadata");

    function setUp() public virtual {
        validator = new SessionKeyValidator();

        user = vm.addr(USER_PK);
        node = vm.addr(NODE_PK);
        otherSigner = vm.addr(OTHER_SIGNER_PK);
        sessionKey1 = vm.addr(SESSION_KEY1_PK);
        sessionKey2 = vm.addr(SESSION_KEY2_PK);
    }

    function createSkAuth(address sessionKey, bytes32 metadataHash, uint256 authorizerPk, bool useEip191)
        internal
        pure
        returns (SessionKeyAuthorization memory)
    {
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: ""})
        );
        bytes memory authSignature;

        if (useEip191) {
            authSignature = TestUtils.signEip191(vm, authorizerPk, authMessage);
        } else {
            authSignature = TestUtils.signRaw(vm, authorizerPk, authMessage);
        }

        return
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: authSignature});
    }

    function signStateWithSk(bytes32 channelId, bytes memory signingData, uint256 skPk, bool useEip191)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory stateMessage = Utils.pack(channelId, signingData);

        if (useEip191) {
            return TestUtils.signEip191(vm, skPk, stateMessage);
        } else {
            return TestUtils.signRaw(vm, skPk, stateMessage);
        }
    }

    function signChallengeWithSk(bytes32 channelId, bytes memory signingData, uint256 skPk, bool useEip191)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory challengeMessage = abi.encodePacked(Utils.pack(channelId, signingData), "challenge");

        if (useEip191) {
            return TestUtils.signEip191(vm, skPk, challengeMessage);
        } else {
            return TestUtils.signRaw(vm, skPk, challengeMessage);
        }
    }
}

contract SessionKeyValidatorTest_validateSignature is SessionKeyValidatorTest_Base {
    function test_success_withBothEip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withBothRaw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withAuthEip191SkSigRaw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withAuthRawSkSigEip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_withSkAuthNotSignedByParticipant_eip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, OTHER_SIGNER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSkAuthNotSignedByParticipant_raw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, OTHER_SIGNER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSigningDataNotSignedBySessionKey_eip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY2_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSigningDataNotSignedBySessionKey_raw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY2_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherMetadataHash_eip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, OTHER_METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        skAuth.metadataHash = METADATA_HASH;
        signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherMetadataHash_raw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, OTHER_METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        skAuth.metadataHash = METADATA_HASH;
        signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherSigningData_eip191() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, OTHER_SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherSigningData_raw() public view {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, OTHER_SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}

contract SessionKeyValidatorTest_validateChallengeSignature is SessionKeyValidatorTest_Base {
    function test_revert_challengeWithSessionKeyNotSupported() public {
        // Signature contents are irrelevant — the method always reverts regardless of input.
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signChallengeWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        validator.validateChallengeSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
    }

    function test_revert_challengeWithSessionKeyNotSupported_emptySignature() public {
        vm.expectRevert(SessionKeyValidator.ChallengeWithSessionKeyNotSupported.selector);
        validator.validateChallengeSignature(CHANNEL_ID, SIGNING_DATA, "", user);
    }
}
