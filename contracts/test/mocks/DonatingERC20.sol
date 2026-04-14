// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title DonatingERC20
 * @notice Mock ERC20 that simulates an ERC777 tokensReceived hook donating tokens back to ChannelHub
 * @dev During transfer, performs the real transfer and then mints extra tokens to `DONATION_TARGET`,
 *      increasing its balance. This replicates the scenario where a recipient's ERC777 hook sends
 *      tokens back to ChannelHub mid-transfer, causing a balance-based success check to misfire.
 */
contract DonatingERC20 is ERC20 {
    address public immutable DONATION_TARGET;
    uint256 public immutable DONATION_AMOUNT;

    constructor(address donationTarget, uint256 donationAmount) ERC20("Donating Token", "DON") {
        DONATION_TARGET = donationTarget;
        DONATION_AMOUNT = donationAmount;
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        bool success = super.transfer(to, amount);
        // Simulate ERC777 tokensReceived hook: recipient donates tokens back to ChannelHub,
        // increasing ChannelHub's balance above (balanceBefore - amount)
        _mint(DONATION_TARGET, DONATION_AMOUNT);
        return success;
    }
}
