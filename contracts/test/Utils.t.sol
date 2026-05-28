// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test, console} from "forge-std/Test.sol";

import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "./TestUtils.sol";

import {Utils} from "../src/Utils.sol";
import {ChannelDefinition, State, Ledger, StateIntent} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization, toSigningData} from "../src/sigValidators/SessionKeyValidator.sol";

contract UtilsTest is Test {
    function test_channelId_forDifferentVersions_differ() public pure {
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 86400,
            user: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045,
            node: 0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a,
            nonce: 42,
            approvedSignatureValidators: 0,
            metadata: 0x13730b0d8e1bdbdc000000000000000000000000000000000000000000000000
        });

        bytes32 channelIdV1 = Utils.getChannelId(def, 1);
        bytes32 channelIdV2 = Utils.getChannelId(def, 2);
        bytes32 channelIdV255 = Utils.getChannelId(def, 255);

        // Channel IDs must differ for different versions
        assert(channelIdV1 != channelIdV2);
        assert(channelIdV1 != channelIdV255);
        assert(channelIdV2 != channelIdV255);

        // First byte should match the version
        assert(uint8(uint256(channelIdV1) >> 248) == 1);
        assert(uint8(uint256(channelIdV2) >> 248) == 2);
        assert(uint8(uint256(channelIdV255) >> 248) == 255);

        // All other bytes should be the same (derived from the same base hash)
        bytes32 mask = 0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
        assert(channelIdV1 & mask == channelIdV2 & mask);
        assert(channelIdV1 & mask == channelIdV255 & mask);
    }

    // ========== Packing Tests ==========

    function test_log_packingState() public pure {
        Ledger memory homeLedger = Ledger({
            chainId: 42,
            token: 0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320,
            decimals: 8,
            userAllocation: 300_000_000,
            userNetFlow: 200_000_001,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999
        });

        Ledger memory nonHomeLedger = Ledger({
            chainId: 4242,
            token: 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2,
            decimals: 14,
            userAllocation: 300_000_000_000_000,
            userNetFlow: 200_000_001_000_000,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999_000_000
        });

        State memory state = State({
            version: 24,
            intent: StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            metadata: hex"dbf80153432e3e0c221112f69a7d20e80980ee5bc48b5684d3b47a6cb75192bd",
            homeLedger: homeLedger,
            nonHomeLedger: nonHomeLedger,
            userSig: hex"36954bf8e670eba9044f0f9eccd3c36871b12ca209f033190bbf378747906d697a521dd4a05faa0ddf3183900df6191ee276055d6d8bf39d8eb8a27e71d2b8b11b",
            nodeSig: hex"2c0648f47bbf3d580dd56acf74662d7d984b6f4abefa1a02ffbd561e0e463761462984ac6dbedac5f679ee29ef58bc9db7f0ac7792d9992832af99a9950039a21b"
        });

        bytes32 channelId = 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b;

        bytes memory packed = Utils.pack(state, channelId);

        // 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000007dbf80153432e3e0c221112f69a7d20e80980ee5bc48b5684d3b47a6cb75192bd000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000001092000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000110d9316ec0000000000000000000000000000000000000000000000000000000b5e62103c2400000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffa50cef950240
        console.logBytes(packed);
    }

    function test_log_packingState_emptyNonHome() public pure {
        Ledger memory homeLedger = Ledger({
            chainId: 42,
            token: 0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320,
            decimals: 8,
            userAllocation: 300_000_000,
            userNetFlow: 200_000_001,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999
        });

        // Empty nonHomeLedger (no escrow)
        Ledger memory nonHomeLedger = Ledger({
            chainId: 0,
            token: address(0),
            decimals: 0,
            userAllocation: 0,
            userNetFlow: 0,
            nodeAllocation: 0,
            nodeNetFlow: 0
        });

        State memory state = State({
            version: 24,
            intent: StateIntent.DEPOSIT,
            metadata: hex"6d621872dd3d14fe6f6ddb415d586e62fb584ffda861ac379bf0d0a0e6410bd6",
            homeLedger: homeLedger,
            nonHomeLedger: nonHomeLedger,
            userSig: hex"36954bf8e670eba9044f0f9eccd3c36871b12ca209f033190bbf378747906d697a521dd4a05faa0ddf3183900df6191ee276055d6d8bf39d8eb8a27e71d2b8b11b",
            nodeSig: hex"2c0648f47bbf3d580dd56acf74662d7d984b6f4abefa1a02ffbd561e0e463761462984ac6dbedac5f679ee29ef58bc9db7f0ac7792d9992832af99a9950039a21b"
        });

        bytes32 channelId = 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b;

        bytes memory packed = Utils.pack(state, channelId);

        // 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000220000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000000000026d621872dd3d14fe6f6ddb415d586e62fb584ffda861ac379bf0d0a0e6410bd6000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
        console.log("Packed state with empty nonHomeLedger:");
        console.logBytes(packed);
    }

    function test_log_calculateChannelId() public pure {
        // Generate metadata from asset: first 8 bytes of keccak256("ether"), padded to 32 bytes
        bytes32 assetHash = keccak256("ether");
        bytes32 metadata;
        assembly {
            metadata := shl(192, shr(192, assetHash))
        }

        console.log("asset hash:");
        // 0x13730b0d8e1bdbdc293b62ba010b1eede56b412ea2980defabe3d0b6c7844c3a
        console.logBytes32(assetHash);

        console.log("metadata:");
        // 0x13730b0d8e1bdbdc000000000000000000000000000000000000000000000000
        console.logBytes32(metadata);

        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 86400,
            user: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045,
            node: 0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a,
            nonce: 42,
            approvedSignatureValidators: 24042,
            metadata: metadata
        });

        bytes32 channelId = Utils.getChannelId(def, 1);

        console.log("channel id:");
        // 0x01f7d8fd998edc15e7f76b914bb9b99a11e56faa5f292a56b42288d4deb168b0
        console.logBytes32(channelId);
    }

    function test_getChannelId_matchesAbiEncodePath() public pure {
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 86400,
            user: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045,
            node: 0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a,
            nonce: 42,
            approvedSignatureValidators: 24042,
            metadata: 0x13730b0d8e1bdbdc000000000000000000000000000000000000000000000000
        });
        uint8 version = 1;

        // Reproduce the pre-assembly path: keccak256(abi.encode(def)) + version in first byte
        bytes32 baseId = keccak256(abi.encode(def));
        bytes32 expected = bytes32(
            (uint256(baseId) & 0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff)
                | (uint256(version) << 248)
        );

        assertEq(Utils.getChannelId(def, version), expected);
    }

    function test_log_calculateEscrowId() public pure {
        bytes32 channelId = 0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66;
        uint64 version = 42;

        bytes32 escrowId = Utils.getEscrowId(channelId, version);

        // 0xe4d925dcf63add647f25c757d6ff0e74ba31401da91d8c7bafa4846c97a92ac2
        console.logBytes32(escrowId);
    }

    function test_log_sessionKeyAuthorization() public pure {
        uint256 skPk = 0xb229abea7c47e19293ef4029fe250d679b295a978922cb6933f022e421f0eb7a;
        address sessionKey = vm.addr(skPk);

        SessionKeyAuthorization memory auth = SessionKeyAuthorization({
            sessionKey: sessionKey,
            metadataHash: hex"59c85ca0f1634dd72294eec96f6d613893a9d3add6943b6f32f9cf7df0858239",
            authSignature: hex""
        });

        bytes memory encoded = toSigningData(auth);

        console.log("SessionKeyAuthorization signing data:");
        // 0x251773da8b8949935ef07284d20cc8605ad7d6f4cf6b5e040ce07dae857f0b6c0000000000000000000000003fba6f40f2cd5b4d833e0305061b88f029ea504459c85ca0f1634dd72294eec96f6d613893a9d3add6943b6f32f9cf7df0858239
        console.logBytes(encoded);

        // 0x5C372EBfC9029aFe7F7506a4d9586604A40e117F
        uint256 privKey = 0xefc7c391f4ce326e149bb9943658998227094eb3e0acc92c343241e22fbc7624;
        bytes memory authSignature = TestUtils.signEip191(vm, privKey, encoded);
        console.log("Auth signature:");
        // 0x287f031625a7a3ac8329d7746c5370c30ffad9857b6e0d022296af4bcad9ca4e0b1b680c3231f5508a1a01406697349afedff09b44c4a26a7881aa262c8bbcee1b
        console.logBytes(authSignature);

        auth.authSignature = authSignature;
        bytes memory signingData =
            hex"3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000007dbf80153432e3e0c221112f69a7d20e80980ee5bc48b5684d3b47a6cb75192bd000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000001092000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000110d9316ec0000000000000000000000000000000000000000000000000000000b5e62103c2400000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffa50cef950240";
        bytes memory skStateSignature = TestUtils.signEip191(vm, skPk, signingData);
        console.log("Session key signature on state:");
        // 0xf8fa3bf9f0660ff737a123d33f19d7939c8662e84d6819259b1543aec89d52720c88a40fa4b99197b8258cf19655e49d957149c6c7acec7007a2b3722b8024b31b
        console.logBytes(skStateSignature);

        bytes memory skModuleSig = abi.encodePacked(SESSION_KEY_VALIDATOR_ID, skStateSignature);
        console.log("Encoded signature for channel validator:");
        // 0x01f8fa3bf9f0660ff737a123d33f19d7939c8662e84d6819259b1543aec89d52720c88a40fa4b99197b8258cf19655e49d957149c6c7acec7007a2b3722b8024b31b
        console.logBytes(skModuleSig);
    }

    function test_log_validatorRegistration() public pure {
        uint256 nodePk = 0xfa35c49dc36191b998cc651d95699f9d63959d2112e14cc1d241f522bec2fe62;
        uint8 validatorId = SESSION_KEY_VALIDATOR_ID;
        address validatorAddress = 0xA33882C770F3D56b9a8E56Bc02d6C7068624F384;
        uint64 chainId = 420042;

        // Build the registration message
        bytes memory message = abi.encode(validatorId, validatorAddress, chainId);
        console.log("Registration message (abi.encode(validatorId, validatorAddress, chainId)):");
        // 0x0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000a33882c770f3d56b9a8e56bc02d6c7068624f38400000000000000000000000000000000000000000000000000000000000668ca
        console.logBytes(message);

        // Sign the registration message
        bytes memory signature = TestUtils.signEip191(vm, nodePk, message);
        console.log("EIP-191 signature:");
        // 0xfbee5bdc27ba3abb6d2bf5bb403b4bb79b96ffacfd89caca022e2abe391b42827135297986b6471ab343a57fb6af66608f5e0c749cd1b6b8f19a6ca1eeab50141c
        console.logBytes(signature);
    }
}
