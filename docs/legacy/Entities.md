# Clearnode: Key Entities

This document outlines the primary data entities in the Clearnode protocol. Understanding these entities and their relationships is essential for developers working with Clearnode.

## Channel

A Channel represents a state channel between participants for token transfers, enabling off-chain transactions with on-chain settlement capabilities.

**Fields:**
- `ChannelID` (string): Unique identifier for the channel
- `ChainID` (uint32): Blockchain network identifier
- `Token` (string): Token address used in this channel
- `Participant` (string): Address of the participant
- `Wallet` (string): Wallet address that owns this channel (may differ from participant if using delegation)
- `RawAmount` (decimal): Current amount in the channel in raw format
- `Status` (enum): Current state of the channel ("joining", "open", "closed")
- `Challenge` (uint64): Challenge period for disputes (in blocks)
- `Nonce` (uint64): Sequence number for state updates
- `Version` (uint64): Version number for tracking protocol changes
- `Adjudicator` (string): Address of the adjudicator contract
- `CreatedAt` (timestamp): When the channel was created
- `UpdatedAt` (timestamp): When the channel was last updated

Channels are uniquely identified by their ChannelID and are associated with specific Assets via Token and ChainID.

## Asset

An Asset represents a cryptocurrency or token that can be used in payment channels.

**Fields:**
- `Token` (string): Contract address of the token
- `ChainID` (uint32): Blockchain network identifier
- `Symbol` (string): Token symbol (e.g., "USDC")
- `Decimals` (uint8): Number of decimal places for the token

Assets are uniquely identified by the combination of their Token address and ChainID.

## AppSession

An AppSession represents a virtual payment application session between multiple participants, enabling complex payment applications beyond simple transfers.

**Fields:**
- `SessionID` (string): Unique identifier for the session
- `Protocol` (string): Protocol version used (must be "NitroRPC/0.2")
- `Challenge` (uint64): Challenge period for disputes
- `Nonce` (uint64): Sequence number for state updates
- `ParticipantWallets` (string[]): List of participant wallet addresses
- `Weights` (int64[]): Voting weights of participants (used for quorum-based decisions)
- `Quorum` (uint64): Required consensus threshold
- `Version` (uint64): Version number
- `Status` (enum): Current state of the session (matches Channel status options)

AppSessions enable multi-party payment applications with consensus mechanisms through weighted signatures. The quorum system allows for flexible governance models where decisions require signatures from participants whose combined weights meet or exceed the quorum threshold.

**Note:** The protocol field is enforced to be "NitroRPC/0.2" for all app sessions. This is the only supported protocol version.

## Ledger Entry

A Ledger Entry records credits and debits for accounts, providing a complete audit trail of all financial operations.

**Fields:**
- `ID` (uint): Unique identifier for the entry
- `AccountID` (string): Identifier of the account
- `AccountType` (enum): Type of the account
- `AssetSymbol` (string): Symbol of the asset being tracked
- `Participant` (string): Address of the participant
- `Credit` (decimal): Amount credited
- `Debit` (decimal): Amount debited
- `CreatedAt` (timestamp): Creation timestamp

The ledger system maintains balances by tracking all credits and debits for each account-asset pair. Importantly, all ledger operations use decimal values to maintain precision, while blockchain-related operations (in Channels and other on-chain entities) use big.Int to ensure consistency with different tokens on different networks, each with their own decimal precision requirements.

## Challenge

A Challenge represents an authentication challenge for verifying address ownership.

**Fields:**
- `Token` (uuid): Random challenge token
- `Address` (string): Address this challenge was created for
- `SessionKey` (string): Optional delegated session key
- `Application` (string): Name of the application which opened the connection
- `Allowances` (Allowance[]): Asset allowances for this session
- `Scope` (string): Permission scope
- `SessionKeyExpiresAt` (uint64): Session key expiration timestamp
- `CreatedAt` (timestamp): When the challenge was created
- `ChallengeExpiresAt` (timestamp): When the challenge expires
- `Completed` (bool): Whether the challenge has been used

Challenges are used in the authentication flow to verify that users own the private keys to their addresses.

## Allowance

An Allowance represents permission to use a specific amount of an asset.

**Fields:**
- `Asset` (string): Asset symbol
- `Amount` (decimal): Allowance amount

Allowances are used in authentication policies to limit what assets and amounts can be used in a session.

## Policy

A Policy represents the permissions and scope for an authenticated session.

**Fields:**
- `Wallet` (string): Main wallet address authorizing the session
- `Participant` (string): Delegated session key address
- `Scope` (string): Permission scope (e.g., "app.create", "ledger.readonly")
- `Application` (string): Application public address
- `Allowances` (Allowance[]): Array of asset allowances
- `ExpiresAt` (timestamp): Expiration timestamp

Policies are used in JWT tokens to define what operations are permitted during a session.

## RPCRecord

An RPCRecord stores the history of RPC messages for auditing and retrieval.

**Fields:**
- `ID` (uint): Unique identifier for the record
- `Sender` (string): Address of the sender
- `ReqID` (uint64): Request identifier
- `Method` (string): RPC method name
- `Params` (bytes): Serialized request parameters
- `Timestamp` (uint64): Unix timestamp
- `ReqSig` (string[]): Request signatures
- `Response` (bytes): Serialized response
- `ResSig` (string[]): Response signatures

RPCRecords provide a complete history of all protocol communications.

## NetworkConfig

A NetworkConfig represents configuration for a blockchain network.

**Fields:**
- `Name` (string): Network name (e.g., "polygon", "celo", "base")
- `ChainID` (uint32): Blockchain network identifier
- `RpcURL` (string): RPC endpoint URL
- `CustodyAddress` (string): Address of the custody contract
- `AdjudicatorAddress` (string): Address of the adjudicator contract

NetworkConfig enables the protocol to interact with different blockchain networks.

## Signature

A Signature represents an ECDSA signature in Ethereum format.

**Fields:**
- `V` (uint8): Recovery identifier (adjusted by +27 per Ethereum convention)
- `R` (string): R component of the signature (32 bytes as hex string)
- `S` (string): S component of the signature (32 bytes as hex string)

Signatures are used throughout the protocol to verify and authorize operations.

## Entity Relationships

- **Channels** reference **Assets** via Token and ChainID.
- **AppSessions** are associated with multiple **Channels** through participant addresses.
- **Ledger Entries** track balances for participants in unified accounts and **AppSessions**.
- **RPCRecords** store the history of RPCMessages.
- **Challenges** are used during authentication and linked to the **Policy** for the session.
- **NetworkConfig** defines parameters for blockchain interactions including **Channel** operations.
- **Signatures** are used to authorize operations across all entities.

## Data Type Conventions

- **Blockchain Operations**: All values related to blockchain operations (Channel amounts, Channel Resize, Channel Close, on-chain transfers) use big.Int to maintain consistency with different tokens across networks, each with their own decimal precision requirements.
- **Ledger Operations**: All internal ledger operations use decimal values to maintain precision for financial accounting purposes.
- **Conversion**: When moving between on-chain and off-chain representations, values are converted using the Asset's `Decimals` field as the scaling factor.
- **RPC Messages**: All RPC messages are signed using ECDSA signatures with participants' private keys.
- **Authentication**: Session management uses both cryptographic challenge-response and JWT tokens.
- **Quorum Signatures**: For operations requiring consensus, the total weight of signers must meet or exceed the quorum threshold defined in the AppSession.