// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";

import {WadMath} from "./WadMath.sol";
import {ChannelDefinition, State, Ledger} from "./interfaces/Types.sol";

library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    error DecimalsExceedMaxPrecision();
    error DecimalsMismatch();
    error FailedToFetchDecimals();

    function getChannelId(ChannelDefinition memory def, uint8 version) internal pure returns (bytes32 channelId) {
        bytes32 baseId = keccak256(abi.encode(def));

        assembly ("memory-safe") {
            // Store the version in the first byte (most significant byte) of the channelId
            // Clear the first byte of baseId, then set it to version
            channelId := or(
                and(baseId, 0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff),
                shl(248, version)
            )
        }
    }

    function getEscrowId(bytes32 channelId, uint64 version) internal pure returns (bytes32) {
        // "channelId, (state-)version" pair is unique as long as participants do not reuse versions
        return keccak256(abi.encode(channelId, version));
    }

    function getValidatorRegistrationMessage(address channelHub, uint8 validatorId, address validator)
        internal
        view
        returns (bytes memory)
    {
        return abi.encode(block.chainid, channelHub, validatorId, validator);
    }

    // ========== State ==========

    function pack(State memory state, bytes32 channelId) internal pure returns (bytes memory) {
        return abi.encode(channelId, toSigningData(state));
    }

    function pack(bytes32 channelId, bytes memory signingData) internal pure returns (bytes memory) {
        return abi.encode(channelId, signingData);
    }

    function toSigningData(State memory state) internal pure returns (bytes memory) {
        return abi.encode(
            state.version,
            state.intent,
            state.metadata,
            state.homeLedger,
            state.nonHomeLedger
            // omit signatures
        );
    }

    function isEmpty(State memory state) internal pure returns (bool) {
        return state.homeLedger.chainId == 0 && state.nonHomeLedger.chainId == 0;
    }

    // ========== Ledger ==========

    /**
     * @notice Validates that the ledger's decimals match the token contract's decimals
     * @dev Only validates if on the same chain as the ledger
     * @param ledger The ledger to validate
     */
    function validateTokenDecimals(Ledger memory ledger) internal view {
        if (ledger.decimals > WadMath.MAX_PRECISION) {
            revert DecimalsExceedMaxPrecision();
        }

        if (ledger.chainId == block.chainid) {
            if (ledger.token == address(0)) {
                require(ledger.decimals == 18, DecimalsMismatch());
            } else {
                try IERC20Metadata(ledger.token).decimals() returns (uint8 tokenDecimals) {
                    require(ledger.decimals == tokenDecimals, DecimalsMismatch());
                } catch {
                    revert FailedToFetchDecimals();
                }
            }
        }
    }

    function isEmpty(Ledger memory ledger) internal pure returns (bool) {
        return ledger.chainId == 0;
    }
}
