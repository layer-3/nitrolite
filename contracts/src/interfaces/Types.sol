// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

// ========= Channel Types ==========

enum ParticipantIndex {
    USER,
    NODE
}

struct ChannelDefinition {
    uint32 challengeDuration;
    address user;
    address node;
    uint64 nonce;
    uint256 approvedSignatureValidators; // Bitmask of approved validator IDs for user signatures (bit N = validator ID N)
    bytes32 metadata;
}

enum ChannelStatus {
    VOID,
    OPERATING,
    DISPUTED,
    CLOSED,
    MIGRATING_IN,
    MIGRATED_OUT
}

enum EscrowStatus {
    VOID,
    INITIALIZED,
    DISPUTED,
    FINALIZED
}

enum StateIntent {
    OPERATE,
    CLOSE,
    DEPOSIT,
    WITHDRAW,
    INITIATE_ESCROW_DEPOSIT,
    FINALIZE_ESCROW_DEPOSIT,
    INITIATE_ESCROW_WITHDRAWAL,
    FINALIZE_ESCROW_WITHDRAWAL,
    INITIATE_MIGRATION,
    FINALIZE_MIGRATION
}

uint8 constant DEFAULT_SIG_VALIDATOR_ID = 0;

struct State {
    uint64 version;
    StateIntent intent;
    bytes32 metadata;
    // to be added for fees logic:
    // bytes data;

    Ledger homeLedger;
    Ledger nonHomeLedger;
    bytes userSig;
    bytes nodeSig;
}

struct Ledger {
    uint64 chainId;
    address token;
    uint8 decimals;
    uint256 userAllocation; // FIXME: investigate whether naming the same thing differently in different components is good
    int256 userNetFlow; // can be negative as user can withdraw funds without depositing them (e.g., on a non-home chain)
    uint256 nodeAllocation;
    int256 nodeNetFlow; // can be negative as node can withdraw user funds
}
