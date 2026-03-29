// SPDX-License-Identifier: GPL-3.0-or-later
pragma solidity ^0.8.22;

import { ERC20, ERC20Capped } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Capped.sol";

contract PremintERC20 is ERC20Capped {
    uint8 private immutable DECIMALS;

    constructor(
        string memory name,
        string memory symbol,
        uint8 decimals_,
        address beneficiary,
        uint256 cap
    ) ERC20(name, symbol) ERC20Capped(cap) {
        DECIMALS = decimals_;
        _mint(beneficiary, cap);
    }

    function decimals() public view virtual override returns (uint8) {
        return DECIMALS;
    }
}
