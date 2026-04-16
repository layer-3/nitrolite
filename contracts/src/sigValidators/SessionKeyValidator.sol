// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {
    ISignatureValidator,
    ValidationResult,
    VALIDATION_SUCCESS,
    VALIDATION_FAILURE
} from "../interfaces/ISignatureValidator.sol";
import {EcdsaSignatureUtils} from "./EcdsaSignatureUtils.sol";
import {Utils} from "../Utils.sol";

/**
 * @notice Authorization struct for delegating signing authority to a session key
 * @dev The participant signs this authorization to allow the session key to sign on their behalf
 * @param sessionKey The address of the delegated session key
 * @param metadataHash Hashed application-specific data (e.g., expiration timestamp, nonce, permissions)
 * @param authSignature The participant's signature authorizing this session key (65 bytes ECDSA)
 */
struct SessionKeyAuthorization {
    address sessionKey;
    bytes32 metadataHash;
    bytes authSignature;
}

function toSigningData(SessionKeyAuthorization memory skAuth) pure returns (bytes memory) {
    return abi.encode(
        skAuth.sessionKey,
        skAuth.metadataHash
        // omit authSignature
    );
}

/**
 * @title SessionKeyValidator
 * @notice Validator supporting session key delegation for temporary signing authority
 * @dev Enables a participant to delegate signing authority to a session key with metadata.
 *      Useful for hot wallets, time-limited access, or gasless transactions.
 *
 * Authorization Flow:
 * 1. Participant signs a SessionKeyAuthorization to delegate to a session key
 * 2. Session key signs the actual state data
 * 3. Both signatures are validated on-chain
 *
 * Signature Format:
 * bytes sigBody = abi.encode(SessionKeyAuthorization skAuthorization, bytes signature)
 *
 * Where signature is a standard 65-byte EIP-191 or raw ECDSA signature of the packed state.
 *
 * Security Model:
 * - Off-chain enforcement (Clearnode) should validate session key expiration and usage limits
 * - On-chain validation only checks cryptographic validity
 * - Participants are responsible for session key management
 */
contract SessionKeyValidator is ISignatureValidator {
    error ChallengeWithSessionKeyNotSupported();

    /**
     * @notice Validates a signature using a delegated session key
     * @dev Validates:
     *      1. participant signed the SessionKeyAuthorization (with channelId binding)
     *      2. sessionKey signed the full state message (channelId + signingData)
     *      Tries EIP-191 recovery first, then raw ECDSA for both signatures.
     * @param channelId The channel identifier to include in state messages
     * @param signingData The encoded state data (without channelId or signatures)
     * @param signature Encoded as abi.encode(SessionKeyAuthorization, bytes signature)
     * @param participant The expected authorizing participant's address
     * @return result VALIDATION_SUCCESS if valid, VALIDATION_FAILURE otherwise
     */
    function validateSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external pure returns (ValidationResult) {
        require(channelId != bytes32(0), EmptyChannelId());
        require(participant != address(0), InvalidSignerAddress());

        (SessionKeyAuthorization memory skAuth, bytes memory skSignature) =
            abi.decode(signature, (SessionKeyAuthorization, bytes));

        // Step 1: Verify participant authorized this session key
        bytes memory authMessage = toSigningData(skAuth);
        bool authResult = EcdsaSignatureUtils.validateEcdsaSigner(authMessage, skAuth.authSignature, participant);

        if (!authResult) {
            return VALIDATION_FAILURE;
        }

        // Step 2: Verify session key signed the full state message
        bytes memory stateMessage = Utils.pack(channelId, signingData);
        if (EcdsaSignatureUtils.validateEcdsaSigner(stateMessage, skSignature, skAuth.sessionKey)) {
            return VALIDATION_SUCCESS;
        } else {
            return VALIDATION_FAILURE;
        }
    }

    /**
     * @notice Challenge signatures via session keys are not supported
     */
    function validateChallengeSignature(bytes32, bytes calldata, bytes calldata, address)
        external
        pure
        returns (ValidationResult)
    {
        revert ChallengeWithSessionKeyNotSupported();
    }
}
