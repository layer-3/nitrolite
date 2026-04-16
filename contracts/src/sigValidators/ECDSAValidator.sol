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
 * @title ECDSAValidator
 * @notice Default signature validator supporting EIP-191 and raw ECDSA signatures
 * @dev Automatically tries both signature formats:
 *      1. EIP-191: Prefixed with Ethereum signed message header (tried first)
 *      2. Raw ECDSA: Direct signature without prefix (tried if EIP-191 fails)
 *
 * The validator prepends channelId to the signingData to construct the full message.
 */
contract ECDSAValidator is ISignatureValidator {
    /**
     * @notice Validates a single participant's signature
     * @dev Constructs the full message by prepending channelId to signingData, then tries EIP-191 recovery first, then raw ECDSA if that fails
     * @param channelId The channel identifier to include in the signed message
     * @param signingData The encoded state data (without channelId or signatures)
     * @param signature The signature to validate (format: [r: 32][s: 32][v: 1], 65 bytes)
     * @param participant The expected signer's address
     * @return result VALIDATION_SUCCESS if valid, VALIDATION_FAILURE otherwise
     */
    function validateSignature(bytes32 channelId, bytes memory signingData, bytes memory signature, address participant)
        public
        pure
        returns (ValidationResult)
    {
        require(channelId != bytes32(0), EmptyChannelId());
        require(participant != address(0), InvalidSignerAddress());

        bytes memory message = Utils.pack(channelId, signingData);
        if (EcdsaSignatureUtils.validateEcdsaSigner(message, signature, participant)) {
            return VALIDATION_SUCCESS;
        } else {
            return VALIDATION_FAILURE;
        }
    }

    /**
     * @notice Validates a challenge signature
     * @dev Appends "challenge" to the packed message and verifies the ECDSA signature.
     * @param channelId The channel identifier to include in the signed message
     * @param signingData The encoded state data (without channelId, signatures, or challenge suffix)
     * @param signature The challenge signature to validate (format: [r: 32][s: 32][v: 1], 65 bytes)
     * @param participant The expected challenger's address
     * @return result VALIDATION_SUCCESS if valid, VALIDATION_FAILURE otherwise
     */
    function validateChallengeSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external pure returns (ValidationResult) {
        return validateSignature(channelId, abi.encodePacked(signingData, "challenge"), signature, participant);
    }
}
