// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title ReentrantERC20
 * @notice Mock ERC20 token whose transferFrom calls back into a configured target with configured
 *         calldata before completing the transfer. Used to simulate ERC777-style `tokensToSend`
 *         hooks for reentrancy testing of ChannelHub lifecycle functions.
 * @dev The reentry fires exactly once: the contract clears its reentry config on each transferFrom
 *      so the inner call (which itself triggers transferFrom on guarded paths) does not recurse
 *      indefinitely. The reentry call's success/return data is captured in
 *      `lastReentryReturnData` / `lastReentrySucceeded` so tests can assert on the inner outcome.
 */
contract ReentrantERC20 is ERC20 {
    address public reentryTarget;
    bytes public reentryCalldata;
    bool public lastReentrySucceeded;
    bytes public lastReentryReturnData;
    bool public reentryArmed;

    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    /// @notice Arm the token to perform a single reentry call into `target` with `data` on the
    /// next `transferFrom` invocation. The reentry hook fires before the underlying ERC20 state
    /// is mutated, mirroring an ERC777 `tokensToSend` callback.
    function armReentry(address target, bytes calldata data) external {
        reentryTarget = target;
        reentryCalldata = data;
        reentryArmed = true;
    }

    function transferFrom(address from, address to, uint256 amount) public override returns (bool) {
        if (reentryArmed) {
            reentryArmed = false;
            address target = reentryTarget;
            bytes memory data = reentryCalldata;
            (bool ok, bytes memory ret) = target.call(data);
            lastReentrySucceeded = ok;
            lastReentryReturnData = ret;
        }
        return super.transferFrom(from, to, amount);
    }
}
