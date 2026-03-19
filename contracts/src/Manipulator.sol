// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IManipulator} from "./interfaces/IManipulator.sol";

/// @title Manipulator
/// @notice A minimal contract that implements a pure math function for PoC testing.
contract Manipulator is IManipulator {
    /// @inheritdoc IManipulator
    function manipulate(uint256 a, uint256 b) external pure override returns (uint256 result) {
        result = (a * b) + a;
    }
}
