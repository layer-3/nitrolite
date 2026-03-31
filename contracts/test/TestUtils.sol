// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Vm} from "forge-std/Vm.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

import {DEFAULT_SIG_VALIDATOR_ID, State, Ledger, StateIntent} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization, toSigningData} from "../src/sigValidators/SessionKeyValidator.sol";
import {Utils} from "../src/Utils.sol";

uint8 constant SESSION_KEY_VALIDATOR_ID = 1;

library TestUtils {
    function signRaw(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, keccak256(message));
        return abi.encodePacked(r, s, v);
    }

    function signEip191(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(message);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signStateEip191WithEcdsaValidator(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, packedState);
        return abi.encodePacked(DEFAULT_SIG_VALIDATOR_ID, signature);
    }

    function signStateEip191WithSkValidator(
        Vm vm,
        bytes32 channelId,
        State memory state,
        uint256 skPk,
        SessionKeyAuthorization memory skAuth
    ) internal pure returns (bytes memory) {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory skSig = TestUtils.signEip191(vm, skPk, packedState);
        bytes memory skModuleSig = abi.encode(skAuth, skSig);
        return abi.encodePacked(SESSION_KEY_VALIDATOR_ID, skModuleSig);
    }

    function buildAndSignSkAuth(Vm vm, address sessionKey, bytes32 metadataHash, uint256 authorizerPk)
        internal
        pure
        returns (SessionKeyAuthorization memory)
    {
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: ""})
        );
        bytes memory signature = TestUtils.signEip191(vm, authorizerPk, authMessage);
        return SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: signature});
    }

    function buildAndSignValidatorRegistration(Vm vm, uint8 validatorId, address validatorAddress, uint256 nodePk)
        internal
        view
        returns (bytes memory)
    {
        bytes memory message = abi.encode(validatorId, validatorAddress, block.chainid);
        return signEip191(vm, nodePk, message);
    }

    function emptyLedger() internal pure returns (Ledger memory) {
        return Ledger({
            chainId: 0,
            token: address(0),
            decimals: 0,
            userAllocation: 0,
            userNetFlow: 0,
            nodeAllocation: 0,
            nodeNetFlow: 0
        });
    }

    function nextState(State memory state, StateIntent intent, uint256[2] memory allocations, int256[2] memory netFlows)
        internal
        pure
        returns (State memory)
    {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeLedger: Ledger({
                chainId: state.homeLedger.chainId,
                token: state.homeLedger.token,
                decimals: state.homeLedger.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeLedger: Ledger({
                chainId: 0,
                token: address(0),
                decimals: 0,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function nextState(
        State memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        uint64 nonHomeChainId,
        address nonHomeChainToken,
        uint256[2] memory nonHomeAllocations,
        int256[2] memory nonHomeNetFlows
    ) internal pure returns (State memory) {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeLedger: Ledger({
                chainId: state.homeLedger.chainId,
                token: state.homeLedger.token,
                decimals: state.homeLedger.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeLedger: Ledger({
                chainId: nonHomeChainId,
                token: nonHomeChainToken,
                decimals: 18,
                userAllocation: nonHomeAllocations[0],
                userNetFlow: nonHomeNetFlows[0],
                nodeAllocation: nonHomeAllocations[1],
                nodeNetFlow: nonHomeNetFlows[1]
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function nextState(
        State memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        uint64 nonHomeChainId,
        address nonHomeChainToken,
        uint8 nonHomeDecimals,
        uint256[2] memory nonHomeAllocations,
        int256[2] memory nonHomeNetFlows
    ) internal pure returns (State memory) {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeLedger: Ledger({
                chainId: state.homeLedger.chainId,
                token: state.homeLedger.token,
                decimals: state.homeLedger.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeLedger: Ledger({
                chainId: nonHomeChainId,
                token: nonHomeChainToken,
                decimals: nonHomeDecimals,
                userAllocation: nonHomeAllocations[0],
                userNetFlow: nonHomeNetFlows[0],
                nodeAllocation: nonHomeAllocations[1],
                nodeNetFlow: nonHomeNetFlows[1]
            }),
            userSig: "",
            nodeSig: ""
        });
    }
}
