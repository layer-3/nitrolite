// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/// @title IManipulator
/// @notice Interface for a simple pure math manipulator.
interface IManipulator {
    /// @notice Computes a deterministic math operation on two unsigned integers.
    /// @param a The first operand.
    /// @param b The second operand.
    /// @return result The computed value: (a * b) + a.
    function manipulate(uint256 a, uint256 b) external pure returns (uint256 result);
}
