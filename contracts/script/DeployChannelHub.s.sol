// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";

/**
 * @title DeployChannelHub
 * @notice Forge script to deploy engine libraries and ChannelHub
 * @dev Foundry automatically deploys unlinked libraries (ChannelEngine,
 *      EscrowDepositEngine, EscrowWithdrawalEngine) before ChannelHub in the
 *      broadcast batch. Their addresses appear in the broadcast JSON output.
 *
 * Usage:
 *   DEFAULT_VALIDATOR_ADDR=<addr>   Address of an already-deployed ISignatureValidator.
 *                                   Leave unset to deploy a fresh ECDSAValidator.
 *
 *   forge script script/DeployChannelHub.s.sol:DeployChannelHub \
 *     --rpc-url <RPC_URL> \
 *     --private-key <DEPLOYER_PK> \
 *     --broadcast \
 *     [-vvvv]
 *
 */
contract DeployChannelHub is Script {
    function run() external {
        // Optional: reuse an existing validator or deploy a fresh ECDSAValidator
        address defaultValidatorAddr = vm.envOr("DEFAULT_VALIDATOR_ADDR", address(0));
        address nodeAddr = vm.envAddress("NODE_ADDR");
        run(defaultValidatorAddr, nodeAddr);
    }

    function run(address defaultValidatorAddr, address nodeAddr) public {
        // msg.sender is set by Foundry to the address derived from --private-key
        address deployer = msg.sender;

        console.log("=== Deploy ChannelHub ===");
        console.log("Deployer:          ", deployer);
        console.log("Chain ID:          ", block.chainid);

        // ----------------------------------------------------------------
        // Predict addresses for informational logging.
        // NOTE: Unlinked libraries (ChannelEngine, EscrowDepositEngine,
        // EscrowWithdrawalEngine) are deployed by Foundry via the CREATE2
        // deployer (0x4e59b44847b379578588920ca78fbf26c0b4956c) and do NOT
        // consume the deployer's CREATE nonce.  Their exact addresses appear
        // in the broadcast JSON after the run.
        // ----------------------------------------------------------------
        uint64 nonce = vm.getNonce(deployer);

        bool deployValidator = defaultValidatorAddr == address(0);
        if (deployValidator) {
            console.log("ECDSAValidator:    ", vm.computeCreateAddress(deployer, nonce));
            nonce++;
        } else {
            console.log("DefaultValidator:  ", defaultValidatorAddr);
        }

        // Libraries are deployed via the CREATE2 deployer; they do not affect
        // the deployer's nonce sequence, so ChannelHub lands at nonce `nonce`.
        console.log(
            "ChannelEngine/EscrowDepositEng/EscrowWithdrawEng: deployed via CREATE2 deployer (see broadcast JSON)"
        );
        console.log("ChannelHub:        ", vm.computeCreateAddress(deployer, nonce));

        vm.startBroadcast();

        // 1. Deploy default signature validator if not provided
        if (deployValidator) {
            ECDSAValidator ecdsaValidator = new ECDSAValidator();
            defaultValidatorAddr = address(ecdsaValidator);
            console.log("Deployed ECDSAValidator:", defaultValidatorAddr);
        }

        // 2. Deploy ChannelHub.
        //    Foundry detects unlinked library references (ChannelEngine,
        //    EscrowDepositEngine, EscrowWithdrawalEngine) and inserts their
        //    deployment transactions before this one in the broadcast batch.
        require(
            defaultValidatorAddr.code.length > 0,
            "DeployChannelHub: DEFAULT_VALIDATOR_ADDR has no code - must be a deployed contract"
        );
        require(nodeAddr != address(0), "DeployChannelHub: NODE_ADDR must be set");
        ChannelHub hub = new ChannelHub(ISignatureValidator(defaultValidatorAddr), nodeAddr);

        vm.stopBroadcast();

        // ----------------------------------------------------------------
        // Summary
        // ----------------------------------------------------------------
        console.log("");
        console.log("=== Deployment complete ===");
        console.log("DefaultSigValidator:", defaultValidatorAddr);
        console.log("Node:               ", nodeAddr);
        console.log("ChannelHub:         ", address(hub));
        console.log("(Library addresses are logged in the broadcast JSON)");
    }
}
