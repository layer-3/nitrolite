// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

interface IChannelHub {
    /// @notice Deposits ERC-20 tokens into the node balance on ChannelHub.
    /// @param token ERC-20 token address.
    /// @param amount Raw token amount in token decimals.
    function depositToNode(address token, uint256 amount) external payable;
}

/**
 * @title DepositToNode
 * @notice Approve a token spend and deposit to the node on the current chain.
 * @dev Single-chain, single-token script. Multi-chain / multi-token orchestration
 *      is handled by batchDepositToNode.sh, which calls this script once per token.
 *
 * Direct usage:
 *   forge script script/DepositToNode.s.sol \
 *     --sig "run(address,address,uint256)" <HUB> <TOKEN> <AMOUNT> \
 *     --rpc-url <RPC_URL> \
 *     --broadcast \
 *     [--account <NAME> | -i | --ledger]
 */
contract DepositToNode is Script {
    using SafeERC20 for IERC20;

    /**
     * @notice Approve hub to spend token then call depositToNode.
     * @param hub    ChannelHub contract address on the current chain.
     * @param token  ERC-20 token address to deposit.
     * @param amount Token amount in the token's native decimals.
     */
    function run(address hub, address token, uint256 amount) external {
        require(hub != address(0), "hub=0");
        require(token != address(0), "token=0");
        require(amount > 0, "amount=0");

        console.log("=== DepositToNode ===");
        console.log("Chain:  ", block.chainid);
        console.log("Hub:    ", hub);
        console.log("Token:  ", token);
        console.log("Amount: ", amount);

        vm.startBroadcast();
        // forceApprove handles non-standard tokens (e.g. USDT) that don't return bool
        IERC20(token).forceApprove(hub, amount);
        IChannelHub(hub).depositToNode(token, amount);
        vm.stopBroadcast();
    }
}
