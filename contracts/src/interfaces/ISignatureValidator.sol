// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

/**
 * @dev ValidationResult is a custom type representing the outcome of signature validation.
 * A non-zero value indicates successful validation, while zero indicates failure.
 * The exact non-zero value can encode additional validation metadata if needed.
 */
type ValidationResult is uint256;

ValidationResult constant VALIDATION_FAILURE = ValidationResult.wrap(0);
ValidationResult constant VALIDATION_SUCCESS = ValidationResult.wrap(1);

/**
 * @title ISignatureValidator
 * @notice Interface for pluggable signature validation in the Nitrolite protocol
 *
 * @dev The Nitrolite protocol supports flexible signature validation through a validator module system.
 * This enables channels to use custom signature schemes beyond standard ECDSA/EIP-191.
 *
 * Validators receive the core state data (signingData) and channelId separately, allowing them to
 * construct the full message according to their specific signing scheme.
 */
interface ISignatureValidator {
    error EmptyChannelId();
    error InvalidSignerAddress();

    /**
     * @notice Validates a participant's signature
     * @param channelId The channel identifier to be included in the signed message
     * @param signingData The encoded state data (without channelId or signatures)
     * @param signature The signature to validate
     * @param participant The expected signer's address
     * @return result ValidationResult indicating success or failure
     */
    function validateSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external view returns (ValidationResult);

    /**
     * @notice Validates a challenge signature
     * @dev The validator constructs the challenge message internally (e.g., appending "challenge"
     *      and any validator-specific data to signingData). This allows each validator to define
     *      its own challenge signature format and enforce validator-specific constraints.
     * @param channelId The channel identifier to be included in the signed message
     * @param signingData The encoded state data
     * @param signature The challenge signature to validate
     * @param participant The expected challenger's address
     * @return result ValidationResult indicating success or failure
     */
    function validateChallengeSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external view returns (ValidationResult);
}
