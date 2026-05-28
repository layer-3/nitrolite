// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title OversizedReturnERC20
 * @notice Mock ERC20 whose transfer() returns `returnDataSize` bytes with `firstWord` as the first word.
 * @dev Used to verify that _trySafeTransfer caps returndata copy to 32 bytes, preventing
 *      memory-expansion OOG when a token returns an oversized buffer.
 *      The actual transfer is performed only when firstWord != 0, matching normal token semantics
 *      (a zero return value signals failure without a state change).
 */
contract OversizedReturnERC20 is ERC20 {
    uint256 private immutable RETURN_DATA_SIZE;
    uint256 private immutable FIRST_WORD;

    constructor(uint256 returnDataSize, uint256 firstWord) ERC20("Oversized Token", "OVR") {
        RETURN_DATA_SIZE = returnDataSize;
        FIRST_WORD = firstWord;
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        if (FIRST_WORD != 0) {
            _transfer(msg.sender, to, amount);
        }

        uint256 size = RETURN_DATA_SIZE;
        uint256 word = FIRST_WORD;
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, word)
            // EVM zero-initialises newly expanded memory, so bytes [32, size) are 0x00.
            return(ptr, size)
        }
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
