// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

/**
 * @title Deposit Interface
 * @notice Interface for contracts that manage token deposits and withdrawals
 * @dev Handles funds that can be allocated to state channels
 */
interface IVault {
    /**
     * @notice Emitted when tokens are deposited into the contract
     * @param wallet Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens deposited
     */
    event Deposited(address indexed wallet, address indexed token, uint48 indexed subId, uint256 amount);

    /**
     * @notice Emitted when tokens are withdrawn from the contract
     * @param wallet Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens withdrawn
     */
    event Withdrawn(address indexed wallet, address indexed token, uint48 indexed subId, uint256 amount);

    /**
     * @notice Gets the balances of multiple accounts for multiple tokens
     * @dev Returns a 2D array where each inner array corresponds to the balances of the tokens for each account
     * @param account Address of the account to check balance for
     * @param token Token address to check balance for (use address(0) for native tokens)
     * @return The balance of the specified token for the specified account
     */
    function getAccountBalance(address account, address token, uint48 subId) external view returns (uint256);

    /**
     * @notice Deposits tokens into the contract
     * @dev For native tokens, the value should be sent with the transaction
     * @param account Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to deposit
     */
    function depositToVault(address account, address token, uint48 subId, uint256 amount) external payable;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Can only withdraw available (not locked in channels) funds
     * @param account Address of the account to send tokens to
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to withdraw
     */
    function withdrawFromVault(address account, address token, uint48 subId, uint256 amount) external;
}
