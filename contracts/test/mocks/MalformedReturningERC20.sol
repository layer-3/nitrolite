// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MalformedReturningERC20
 * @notice Mock ERC20 token that returns malformed data (less than 32 bytes)
 * @dev This simulates a non-compliant token that could cause abi.decode to revert
 */
contract MalformedReturningERC20 is ERC20 {
    constructor() ERC20("Malformed Token", "MLF") {
        _mint(msg.sender, 1000000 * 10 ** 18);
    }

    function transfer(address, uint256) public pure override returns (bool) {
        // Return only 1 byte instead of 32 (malformed) WITHOUT actually transferring
        // This simulates a malicious/buggy token that returns invalid data
        assembly {
            // mstore8 writes a single byte (0x01) to address 0.
            // mstore(0, 1) would write the value 1 as a 32-byte big-endian word,
            // placing 0x01 at address 31 and 0x00 at address 0 — so return(0, 1)
            // would yield 0x00, not 0x01. mstore8 avoids this by writing exactly one byte.
            mstore8(0, 1)
            return(0, 1)
        }
    }

    function transferFrom(address, address, uint256) public pure override returns (bool) {
        // Return only 1 byte instead of 32 (malformed) WITHOUT actually transferring
        // This simulates a malicious/buggy token that returns invalid data
        assembly {
            // mstore8 writes a single byte (0x01) to address 0.
            // mstore(0, 1) would write the value 1 as a 32-byte big-endian word,
            // placing 0x01 at address 31 and 0x00 at address 0 — so return(0, 1)
            // would yield 0x00, not 0x01. mstore8 avoids this by writing exactly one byte.
            mstore8(0, 1)
            return(0, 1)
        }
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
