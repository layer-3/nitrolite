# Signature Validators

This document describes the pluggable signature validation system in the Nitrolite protocol.

---

## Overview

The protocol supports flexible signature validation through the `ISignatureValidator` interface. All validators implement:

- `validateSignature(channelId, signingData, signature, participant)` — Validates a state signature
- `validateChallengeSignature(channelId, signingData, signature, participant)` — Validates a challenge signature

Validators receive the core state data (`signingData`) and `channelId` separately, allowing them to construct the full message according to their signing scheme.

For challenge signatures, ChannelHub calls `validateChallengeSignature` rather than `validateSignature`. Each validator is responsible for constructing the challenge message and enforcing any validator-specific constraints (e.g., temporal bounds for session keys).

---

## Validator Selection

### Node Validator Registry

The protocol uses a **validator registry** system. NODE can register signature validators and assign them 1-byte identifiers (0x01-0xFF).

**Design rationale:** In the Nitrolite off-chain protocol, the node acts as the orchestrator and decides which signature validators are supported for **node signatures**. This ensures:

- NODE can enforce its security requirements for its own signatures
- NODE benefits from flexible validator implementations (SessionKey, multi-sig, etc.)
- Cross-chain compatibility (validator addresses don't affect channelId or signature verification)

### Security Consideration: Preventing Signature Forgery

**Identified Vulnerability:** If both user and node signatures use validators from a node-controlled registry, a malicious node can register a custom validator that always returns `VALIDATION_SUCCESS`, allowing it to forge user signatures and unilaterally execute state transitions (withdrawals, closures, etc.).

**Solution Approaches:**

Three approaches were considered to prevent this attack while maintaining cross-chain compatibility and minimal transaction overhead:

1. **Approved Validators Bitmask (Current Implementation):** Store allowed validator IDs in the `ChannelDefinition.approvedSignatureValidators` field (uint256 bitmask). Since this field is part of `channelId` computation, it cannot be changed during cross-chain operations without invalidating signatures. The default ECDSA validator (0x00) is always available. Additional validators from the node's registry can be approved by setting corresponding bits (e.g., bit 42 set = validator ID 42 approved).

2. **Per-Signature Authorization:** Users include an ECDSA authorization signature alongside each state signature when using non-default validators. The authorization proves the user consents to using that specific validator for that specific channel. This adds ~65 bytes overhead per signature but provides maximum flexibility without requiring additional transactions.

3. **Hybrid Approach (Planned):** Combine approaches 1 and 2. Users can pre-approve validators in the `approvedSignatureValidators` bitmask (zero overhead), or use per-signature authorization for validators not in the bitmask (+65 bytes overhead). This provides both efficiency for common cases and flexibility for advanced use cases. Will be implemented in a future version.

### User Validator Selection via Approved Validators Bitmask

To prevent nodes from forging user signatures, validator selection is controlled via the `ChannelDefinition.approvedSignatureValidators` field (uint256 bitmask). This mitigates the signature forgery vulnerability by ensuring users control which validators can verify signatures.

**How it works:**

- The `approvedSignatureValidators` field in `ChannelDefinition` is part of the `channelId` computation
- The default ECDSA validator (0x00) is **always** available, regardless of the bitmask value
- The bitmask specifies which additional validators from the node's registry are agreed validators
- Each bit corresponds to a validator ID: if bit N is set to 1, validator ID N is approved (e.g., bit 42 = 1 means validator ID 42 is approved)
- Since `channelId` is included in all signatures, the agreed validators cannot be changed during cross-chain operations without invalidating signatures

This approach provides security (users control the agreed validators), cross-chain compatibility (approvedSignatureValidators is in channelId), zero transaction overhead (no separate validator registration needed), and guaranteed access to the default ECDSA validator.

### Validator Registration

NODE registers validators by providing a signature over the validator configuration. This allows node operators to use cold storage or hardware wallets without exposing private keys to send transactions.

**Registration message:**

```solidity
bytes memory message = abi.encode(validatorId, validatorAddress, block.chainid);
```

The signature is verified using ECDSA recovery:

1. Try EIP-191 recovery first (standard for wallet software)
2. Fall back to raw ECDSA if needed
3. Verify recovered address matches the node address

**Key properties:**

- Includes `block.chainid` for cross-chain replay protection (registrations are chain-specific)

- Anyone can submit a registration transaction (relayer-friendly)
- Node's private key only signs, never sends transactions
- Validator ID 0x00 is reserved for the default validator
- Registration is immutable (cannot change once set)
- 255 validators (0x01-0xFF)

### Signature Format

All signatures in the protocol follow this structure:

```txt
[validator_id: 1 byte][signature_data: variable length]
```

- `0x00` = Use ChannelHub's default validator
- `0x01-0xFF` = Look up validator in node's registry

The first byte determines which validator verifies the signature. The remaining bytes are passed to the selected validator for verification.

---

## Domain Separation: ChannelHub vs Validators

The protocol maintains clear separation between protocol concerns and cryptographic concerns:

### ChannelHub Responsibilities

- Define protocol message structure (when and how channelId binds to states)
- Manage channel lifecycle and state transitions
- Select appropriate validators based on signature first byte
- Handle validator registration (infrastructure concern)

### Validator Responsibilities

- Verify cryptographic signatures using specific schemes (ECDSA, multi-sig, session keys, etc.)
- Support different signature formats and recovery mechanisms
- Remain agnostic to protocol-level message structure

### Why This Matters

**State validation** requires channelId binding for security. Validators receive `channelId` and `signingData` separately because ChannelHub controls *when* and *how* channelId is included in signed messages. This is a protocol-level security requirement, not a cryptographic concern.

**Validator registration** is an infrastructure operation that happens outside the channel state validation flow. It uses direct ECDSA recovery in ChannelHub rather than going through the validator abstraction, because:

- Registration has no channelId (different domain)
- Registration is operational setup, not protocol state transition
- All node operators use ECDSA-capable wallets for registration
- Keeps `ISignatureValidator` focused on its primary purpose

This separation ensures validators remain pluggable for state verification while keeping protocol-level concerns within ChannelHub.

---

## Cross-Chain Compatibility

The node validator registry design solves a critical cross-chain problem: validator contracts may not deploy to the same address on all chains.
This enables true cross-chain operation without requiring deterministic deployment (CREATE2) across all chains.

---

## ECDSAValidator

**Location:** `src/sigValidators/ECDSAValidator.sol`

### Description

Default validator supporting standard ECDSA signatures. Automatically tries both EIP-191 (with Ethereum prefix) and raw ECDSA formats.

### Signature Format

65 bytes: `[r: 32 bytes][s: 32 bytes][v: 1 byte]`

### Validation Logic

1. Try EIP-191 recovery first (most wallets use this)
2. If fails, try raw ECDSA recovery
3. Return `VALIDATION_SUCCESS` if recovered address matches participant, `VALIDATION_FAILURE` otherwise

### Challenge Validation

`validateChallengeSignature` delegates to `validateSignature` with the signing data extended by a `"challenge"` suffix. The signer must sign `pack(channelId, signingData || "challenge")`. No temporal or scope checks apply — ECDSA keys do not expire.

### Use Cases

- Standard wallet signatures (MetaMask, WalletConnect, hardware wallets)
- Most common validator for all channels
- Recommended for both users and nodes

---

## SessionKeyValidator

**Location:** `src/sigValidators/SessionKeyValidator.sol`

### Description

Enables delegation to temporary session keys. Useful for hot wallets, time-limited access, and gasless transactions.

### Signature Format

```solidity
struct SessionKeyAuthorization {
    address sessionKey;      // Delegated signer
    bytes32 metadataHash;    // Hashed expiration, permissions, etc.
    bytes authSignature;     // Participant's authorization (65 bytes)
}

bytes sigBody = abi.encode(SessionKeyAuthorization, bytes sessionKeySignature)
```

### Validation Logic

**Two checks:**

1. Participant authorized the session key: `authData = abi.encode(sessionKey, metadataHash)`
2. Session key signed the state

Both use EIP-191 first, then raw ECDSA if that fails.

### Metadata

Application-defined data encoding expiration, allowed channels, and permissions. **Validated off-chain by Clearnode, not on-chain.**

### Challenge Signatures

`validateChallengeSignature` is **not supported** and always reverts with `ChallengeWithSessionKeyNotSupported`.

This is to prevent a vulnerability: since session key authorizations are permanently valid on-chain (expiration is opaque in metadataHash), allowing session keys to challenge would let any expired or revoked key put channels into `DISPUTED` state unilaterally, bypassing Clearnode's off-chain enforcement and causing a DoS on the channel.

---

## SECURITY: SessionKeyValidator Usage

⚠️ **CRITICAL: SessionKeyValidator is for USER usage only, NOT for nodes.**

### Users: Safe ✅

**Note:** SessionKeyValidator can be enabled by setting the corresponding bit in `ChannelDefinition.approvedSignatureValidators` (e.g., if SessionKeyValidator has ID 1, set bit 1 to approve it).

It is safe for users because:

- Clearnode validates metadata (expiration, scope, permissions)
- Node must countersign (provides protection)
- Limited damage if compromised (Clearnode rejects invalid requests)
- Revocable (switch to main key anytime)

### Nodes: Unsafe ⚠️

**Not recommended for node usage:**

- User has no off-chain validation mechanism
- User cannot verify metadata constraints (only hash is checked on-chain)
- If node's session key is compromised: unlimited, irrevocable authority
- User must blindly trust node's session key
