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
 *      broadcast batch via the CREATE2 factory (0x4e59b44847b379578588920ca78fbf26c0b4956c,
 *      salt = bytes32(0)). Their addresses are deterministic and are logged in the
 *      summary below. Tx hashes for all deployments are written to the broadcast JSON
 *      after the script completes.
 *
 * Usage:
 *   DEFAULT_VALIDATOR_ADDR=<addr>     Address of an already-deployed ISignatureValidator.
 *                                     Leave unset to deploy a fresh ECDSAValidator.
 *
 *   CHANNEL_ENGINE_ADDR=<addr>        Address of an already-deployed ChannelEngine library.
 *   ESCROW_DEPOSIT_ENGINE_ADDR=<addr> Address of an already-deployed EscrowDepositEngine library.
 *   ESCROW_WITHDRAWAL_ENGINE_ADDR=<addr> Address of an already-deployed EscrowWithdrawalEngine library.
 *
 *   When all three library addresses are provided, Foundry skips their deployment (code
 *   already exists at those addresses). You must also pass the --libraries flag so the
 *   linker wires ChannelHub to the existing addresses:
 *
 *     --libraries src/ChannelEngine.sol:ChannelEngine:<addr> \
 *     --libraries src/EscrowDepositEngine.sol:EscrowDepositEngine:<addr> \
 *     --libraries src/EscrowWithdrawalEngine.sol:EscrowWithdrawalEngine:<addr>
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

        // Optional: reuse existing library deployments; leave unset to deploy via CREATE2
        address channelEngineAddr = vm.envOr("CHANNEL_ENGINE_ADDR", address(0));
        address escrowDepositAddr = vm.envOr("ESCROW_DEPOSIT_ENGINE_ADDR", address(0));
        address escrowWithdrawalAddr = vm.envOr("ESCROW_WITHDRAWAL_ENGINE_ADDR", address(0));

        run(defaultValidatorAddr, nodeAddr, channelEngineAddr, escrowDepositAddr, escrowWithdrawalAddr);
    }

    function run(
        address defaultValidatorAddr,
        address nodeAddr,
        address channelEngineAddr,
        address escrowDepositAddr,
        address escrowWithdrawalAddr
    ) public {
        // msg.sender is set by Foundry to the address derived from --private-key
        address deployer = msg.sender;

        console.log("=== Deploy ChannelHub ===");
        console.log("Deployer:          ", deployer);
        console.log("Chain ID:          ", block.chainid);

        // ----------------------------------------------------------------
        // Resolve library addresses: use provided addresses or fall back to
        // the deterministic CREATE2 addresses Foundry would deploy to.
        // ----------------------------------------------------------------
        bool deployChannelEngine = channelEngineAddr == address(0);
        bool deployEscrowDeposit = escrowDepositAddr == address(0);
        bool deployEscrowWithdrawal = escrowWithdrawalAddr == address(0);

        if (deployChannelEngine) {
            channelEngineAddr = _computeLibraryAddress("ChannelEngine.sol:ChannelEngine");
        }
        if (deployEscrowDeposit) {
            escrowDepositAddr = _computeLibraryAddress("EscrowDepositEngine.sol:EscrowDepositEngine");
        }
        if (deployEscrowWithdrawal) {
            escrowWithdrawalAddr = _computeLibraryAddress("EscrowWithdrawalEngine.sol:EscrowWithdrawalEngine");
        }

        // Validate that provided addresses actually have code
        if (!deployChannelEngine) {
            require(channelEngineAddr.code.length > 0, "DeployChannelHub: CHANNEL_ENGINE_ADDR has no code");
        }
        if (!deployEscrowDeposit) {
            require(escrowDepositAddr.code.length > 0, "DeployChannelHub: ESCROW_DEPOSIT_ENGINE_ADDR has no code");
        }
        if (!deployEscrowWithdrawal) {
            require(escrowWithdrawalAddr.code.length > 0, "DeployChannelHub: ESCROW_WITHDRAWAL_ENGINE_ADDR has no code");
        }

        uint64 nonce = vm.getNonce(deployer);

        bool deployValidator = defaultValidatorAddr == address(0);
        if (deployValidator) {
            console.log("ECDSAValidator:    ", vm.computeCreateAddress(deployer, nonce));
            nonce++;
        } else {
            console.log("DefaultValidator:  ", defaultValidatorAddr);
        }
        console.log(
            deployChannelEngine ? "ChannelEngine (new):      " : "ChannelEngine (existing): ", channelEngineAddr
        );
        console.log(
            deployEscrowDeposit ? "EscrowDepositEngine (new):         " : "EscrowDepositEngine (existing):    ",
            escrowDepositAddr
        );
        console.log(
            deployEscrowWithdrawal ? "EscrowWithdrawalEngine (new):      " : "EscrowWithdrawalEngine (existing): ",
            escrowWithdrawalAddr
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
        //    When library env vars are unset, Foundry detects unlinked library
        //    references and inserts their deployment transactions before this one
        //    in the broadcast batch (CREATE2, salt=0). When addresses are provided
        //    via env vars, Foundry finds code already at those addresses and skips
        //    their redeployment.
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
        string memory broadcastFile =
            string.concat("broadcast/DeployChannelHub.s.sol/", vm.toString(block.chainid), "/run-latest.json");

        console.log("");
        console.log("=== Deployment complete ===");
        console.log("DefaultSigValidator:", defaultValidatorAddr);
        console.log("Node:               ", nodeAddr);
        console.log("ChannelHub:         ", address(hub));
        console.log("");
        console.log("=== Libraries ===");
        console.log(
            deployChannelEngine ? "  ChannelEngine (deployed):         " : "  ChannelEngine (reused):           ",
            channelEngineAddr
        );
        console.log(
            deployEscrowDeposit ? "  EscrowDepositEngine (deployed):    " : "  EscrowDepositEngine (reused):      ",
            escrowDepositAddr
        );
        console.log(
            deployEscrowWithdrawal ? "  EscrowWithdrawalEngine (deployed): " : "  EscrowWithdrawalEngine (reused):   ",
            escrowWithdrawalAddr
        );
        console.log("");
        console.log(string.concat("  (tx hashes: ", broadcastFile, ")"));
    }

    // ----------------------------------------------------------------
    // Internal helpers
    // ----------------------------------------------------------------

    /// @dev Compute the deterministic CREATE2 address Foundry uses when
    ///      auto-deploying an unlinked library: CREATE2_FACTORY, salt=bytes32(0).
    function _computeLibraryAddress(string memory artifact) internal view returns (address) {
        bytes memory creationCode = vm.getCode(artifact);
        return vm.computeCreate2Address(bytes32(0), keccak256(creationCode), CREATE2_FACTORY);
    }
}
