// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

import {TestUtils} from "../test/TestUtils.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../src/Utils.sol";

/**
 * @title RegisterNodeValidator
 * @notice Forge script to register a signature validator for a node
 * @dev This script signs the registration message and submits it to ChannelHub
 *
 * Usage:
 *   forge script script/RegisterNodeValidator.s.sol:RegisterNodeValidator \
 *     --rpc-url <RPC_URL> \
 *     --broadcast \
 *     --private-key <SUBMITTER_PRIVATE_KEY> \
 *     --sig "run(address,uint256,uint8,address)" \
 *     <CHANNEL_HUB_ADDRESS> <NODE_PRIVATE_KEY> <VALIDATOR_ID> <VALIDATOR_ADDRESS>
 *
 * Environment variables (alternative to command-line args):
 *   CHANNEL_HUB_ADDRESS - Address of the ChannelHub contract
 *   NODE_PRIVATE_KEY - Private key of the node (for signing, not for submitting tx)
 *   VALIDATOR_ID - The validator ID to register (1-255, 0 is reserved)
 *   VALIDATOR_ADDRESS - Address of the validator contract
 */
contract RegisterNodeValidator is Script {
    using MessageHashUtils for bytes;

    function run() external {
        address channelHubAddress = vm.envAddress("CHANNEL_HUB_ADDRESS");
        uint256 nodePrivateKey = vm.envUint("NODE_PRIVATE_KEY");
        uint8 validatorId = uint8(vm.envUint("VALIDATOR_ID"));
        address validatorAddress = vm.envAddress("VALIDATOR_ADDRESS");

        run(channelHubAddress, nodePrivateKey, validatorId, validatorAddress);
    }

    function run(address channelHubAddress, uint256 nodePrivateKey, uint8 validatorId, address validatorAddress)
        public
    {
        require(channelHubAddress != address(0), "Invalid ChannelHub address");
        require(validatorAddress != address(0), "Invalid validator address");
        require(validatorId != 0, "Validator ID 0 is reserved");
        require(nodePrivateKey != 0, "Invalid node private key");

        ChannelHub channelHub = ChannelHub(payable(channelHubAddress));
        address nodeAddress = vm.addr(nodePrivateKey);

        console.log("=== Register Node Validator ===");
        console.log("ChannelHub:", channelHubAddress);
        console.log("Node address:", nodeAddress);
        console.log("Validator ID:", validatorId);
        console.log("Validator address:", validatorAddress);
        console.log("Chain ID:", block.chainid);

        bytes memory message = Utils.getValidatorRegistrationMessage(channelHubAddress, validatorId, validatorAddress);
        console.log("Message hash:", vm.toString(keccak256(message)));

        bytes memory signature = TestUtils.signEip191(vm, nodePrivateKey, message);
        console.log("Signature:", vm.toString(signature));

        vm.broadcast();
        channelHub.registerNodeValidator(validatorId, ISignatureValidator(validatorAddress), signature);

        console.log("Validator registered successfully!");
    }
}
