// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title IParametricToken
 * @dev Extension of ERC20 that allows a single address to manage multiple
 *      sub-accounts (partitions), each with its own parameters (e.g., mint time)
 */
interface IParametricToken is IERC20 {
    // Account types
    enum AccountType {
        Normal,
        Super
    }

    // Events
    event AccountConvertedToSuper(address indexed account);
    event SubAccountCreated(address indexed superAccount, uint48 indexed subId);
    event TransferToSub(address indexed from, address indexed toSuper, uint48 indexed toSubId, uint256 amount);
    event TransferFromSub(address indexed fromSuper, uint48 indexed fromSubId, address indexed to, uint256 amount);
    event TransferBetweenSubs(address indexed superAccount, uint48 indexed fromSubId, uint48 indexed toSubId, uint256 amount);
    event TransferFromSubToSub(address indexed fromSuper, uint48 indexed fromSubId, address indexed toSuper, uint48 toSubId, uint256 amount);
    event ApprovalForSub(address indexed owner, uint48 indexed subId, address indexed spender, uint256 amount);

    // Account management
    function convertToSuper(address account) external returns (bool);

    function createSubAccount(address account) external returns (uint48);

    function accountType(address account) external view returns (AccountType);

    // Sub-account queries
    function balanceOfSub(address superAccount, uint48 subId) external view returns (uint256);

    function subsCountOf(address superAccount) external view returns (uint48);

    function numberOfParameters() external pure returns (uint8);

    function parameterOf(uint8 paramIndex, address account) external view returns (uint64);

    function parameterOfSub(uint8 paramIndex, address account, uint48 subId) external view returns (uint64);

    function allowanceForSub(address owner, uint48 subId, address spender) external view returns (uint256);

    // Parametric transfers
    function transferToSub(address toSuper, uint48 toSubId, uint256 amount) external returns (bool);

    function transferFromSub(uint48 fromSubId, address to, uint256 amount) external returns (bool);

    function transferBetweenSubs(uint48 fromSubId, uint48 toSubId, uint256 amount) external returns (bool);

    // Approved parametric transfers
    function approveForSub(uint48 ownerSubId, address spender, uint256 amount) external returns (bool);

    function approvedTransferToSub(address from, address toSuper, uint48 toSubId, uint256 amount) external returns (bool);

    function approvedTransferFromSubToSub(address fromSuper, uint48 fromSubId, address toSuper, uint48 toSubId, uint256 amount) external returns (bool);
}
