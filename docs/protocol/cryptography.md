# Cryptography

Previous: [Terminology](terminology.md) | Next: [State Model](state-model.md)

---

This document defines how protocol objects are encoded, hashed, and signed.

All rules are described as algorithms and canonical procedures, independent of any specific programming language.

## Purpose

Cryptography in the Nitrolite protocol serves three functions:

1. **Authentication** — proving that a specific participant authorized a state update
2. **Integrity** — ensuring that signed data has not been modified
3. **Replay protection** — preventing previously signed states from being reused in unintended contexts

## Cryptographic Algorithms

The protocol uses the following cryptographic primitives.

**Signature Algorithm**
ECDSA over the secp256k1 curve, producing a 65-byte signature (r, s, v).

**Hash Function**
Keccak-256, producing a 32-byte digest.

## Canonical Encoding

Protocol objects that require signing MUST be encoded into a canonical binary representation before hashing.

The canonical encoding uses RLP encoding (`abi.encode` in Solidity) as defined in [this paper](https://doi.org/10.48550/arXiv.2009.13769) and by [Ethereum documentation](https://ethereum.org/developers/docs/data-structures-and-encoding/rlp/). This ensures deterministic byte sequences regardless of implementation language.

## Message Digest Construction

The digest of a signable payload is constructed as follows:

1. Encode the object using canonical encoding
2. Prepend the EIP-191 personal message prefix: the ASCII string `"\x19Ethereum Signed Message:\n"` followed by the decimal length of the encoded bytes, then the encoded bytes themselves
3. Compute the Keccak-256 hash of the prefixed message

The resulting 32-byte digest is the value that is signed.

## ECDSA Signature Format

The raw ECDSA signature consists of:

| Field | Size     | Description              |
| ----- | -------- | ------------------------ |
| R     | 32 bytes | ECDSA r component        |
| S     | 32 bytes | ECDSA s component        |
| V     | 1 byte   | Recovery identifier      |

The signer's address is recovered from the signature and the message digest. The protocol does not transmit the signer's public key or address alongside the signature.

## Protocol Signature Envelope

A protocol signature is a wrapper around the raw ECDSA signature that includes a validation mode prefix:

```
ProtocolSignature = ValidationMode || SignatureData
```

The first byte (`ValidationMode`) determines the validation method, which must map to a signature validator registered by the Node on the Smart Contract infrastructure. The remaining bytes (`SignatureData`) contain mode-specific data including the raw signature.

## Signature Validation Modes

The protocol supports multiple signature validation modes to allow different key types and authorization schemes.

**Default Mode (0x00)**
Standard ECDSA signature validation. SignatureData contains the raw ECDSA signature (R, S, V). The signer's address is recovered from the signature. The recovered address MUST match the expected participant address.

**Session Key Mode (0x01)**
Delegated signature validation. SignatureData contains a session key authorization and the session key's ECDSA signature over the state data, ABI-encoded as a tuple. The validator first verifies that the participant authorized the session key, then verifies that the session key produced a valid signature over the state. The session key authorization MUST be associated with the same address as the channel's user or node participant. The recovered session key address MUST match the address authorized by the participant.

Session-key signatures are valid for both off-chain state advancement and on-chain enforcement, provided the session key validation mode is among the channel's approved signature validators.

## Signable Object Classes

The protocol defines a general signing framework that accommodates multiple classes of signable objects:

- **Channel Objects**: primarily, the state of a channel, but also a session key registration and challenger signature
- **Extension Objects**: primarily, the state of an extension entity (such as an application session), signed by the relevant session participants

Please note that channel and extension states are identified by a unique entity identifier and follows the same canonical encoding and digest construction rules.

This framework is extensible: future protocol extensions MAY introduce additional signable object classes without requiring changes to the core signing rules.

## Session Key Authorization

A participant MAY delegate signing authority to a session key.

The authorization is constructed as follows:

1. The participant signs a message containing:
   - the session key address
   - authorization metadata hash (`keccak256` over scope, expiration and possible other data)
2. The authorization signature is produced using the participant's primary key
3. The session key MAY then produce signatures on behalf of the participant within the authorized scope

Session key signatures MUST include the authorization proof alongside the session key signature. The authorization proof is canonically encoded as a tuple containing the session key authorization and the raw signature bytes.

## Replay Protection

The protocol prevents replay attacks through the following mechanisms:

**Entity Identifier**
Each signable entity has a unique identifier derived from its definition. Signed states are bound to a specific entity, preventing a signature over one entity's state from being replayed against another.

**State Version**
Each state includes a monotonically increasing version number. The blockchain layer MUST reject states with a version less than or equal to the currently enforced version.

**Blockchain Identifier**
States include blockchain-specific identifiers preventing cross-chain replay.

**Smart Contract Version**
The channel entity identifier incorporates a contract version (currently as the first byte), preventing replay across different deployments.

---

Previous: [Terminology](terminology.md) | Next: [State Model](state-model.md)
