// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

import {Manipulator} from "../src/Manipulator.sol";

/**
 * @title DeployManipulator
 * @notice Forge script to deploy the Manipulator contract.
 *
 * Usage:
 *   forge script script/DeployManipulator.s.sol:DeployManipulator \
 *     --rpc-url <RPC_URL> \
 *     --private-key <DEPLOYER_PK> \
 *     --broadcast \
 *     [-vvvv]
 */
contract DeployManipulator is Script {
    function run() external {
        address deployer = msg.sender;

        console.log("=== Deploy Manipulator ===");
        console.log("Deployer: ", deployer);
        console.log("Chain ID: ", block.chainid);

        uint64 nonce = vm.getNonce(deployer);
        console.log("Predicted:", vm.computeCreateAddress(deployer, nonce));

        vm.startBroadcast();

        Manipulator manipulator = new Manipulator();

        vm.stopBroadcast();

        console.log("");
        console.log("=== Deployment complete ===");
        console.log("Manipulator:", address(manipulator));
    }
}
