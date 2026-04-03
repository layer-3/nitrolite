// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";

/**
 * @title TestChannelHub
 * @notice Test harness contract that exposes internal ChannelHub functions for testing
 */
contract TestChannelHub is ChannelHub {
    uint48 constant SUB_ID = 0;

    constructor(ISignatureValidator _defaultSigValidator) ChannelHub(_defaultSigValidator) {}

    /**
     * @notice Marks this contract as a test contract for Forge
     * @dev Prevents size limit checks from being enforced on this test harness
     */
    function IS_TEST() external pure {}

    /**
     * @notice Exposed version of _pushFunds for testing
     */
    function exposed_pushFunds(address to, address token, uint256 amount) external payable {
        _pushFunds(SUB_ID, to, token, amount);
    }

    /**
     * @notice Exposed version of _pullFunds for testing
     */
    function exposed_pullFunds(address from, address token, uint256 amount) external payable {
        _pullFunds(from, SUB_ID, token, amount);
    }

    /**
     * @notice Workaround to set reclaim balance for testing
     * @dev Allows tests to set up reclaim state without going through failed transfers
     */
    function workaround_setReclaim(address account, address token, uint256 amount) external {
        _reclaims[account][token] = amount;
    }
}
