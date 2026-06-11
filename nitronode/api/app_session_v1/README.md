# App Session V1 API Implementation

This directory contains the V1 API handlers for app session management, implementing the `create_app_session`, `submit_deposit_state`, `submit_app_state`, `get_app_sessions`, `get_app_definition`, `submit_session_key_state`, and `get_last_key_states` endpoints.


## Architecture


### API Layer (`nitronode/api/app_session_v1`)
- **Thin RPC handlers** that parse requests and format responses
- Delegates all business logic to `pkg/app`
- **Separate file per endpoint** (following channel_v1 pattern):
  - `handler.go` - Handler struct and constructor
  - `create_app_session.go` - Create app session endpoint
  - `submit_deposit_state.go` - Submit deposit state endpoint
  - `submit_app_state.go` - Submit app state endpoint (operate, withdraw, close)
  - `get_app_sessions.go` - Get app sessions endpoint (with filtering)
  - `get_app_definition.go` - Get app definition endpoint
  - `submit_session_key_state.go` - Submit session key state endpoint
  - `get_last_key_states.go` - Get last session key states endpoint
  - `app_session.go` - Package documentation

### Business Logic Layer (`pkg/app`)
- **AppSessionServiceV1**: Core business logic for app sessions (~528 lines)
- **AppSessionV1**: Type definitions for app sessions
- All validation, state management, and persistence logic


## Endpoints

### 1. `app_sessions.v1.create_app_session`

**Purpose**: Creates a new application session between participants with 0 allocations by default.

**Key Changes from Legacy API**:
- ✅ **App sessions are created with 0 allocations**
- No initial deposits during creation
- Simplified creation flow

**Request**:
```json
{
  "definition": {
    "application": "string",
    "participants": [
      {
        "wallet_address": "string",
        "signature_weight": 1
      }
    ],
    "quorum": 1,
    "nonce": 12345
  },
  "signatures": ["0x...", "0x..."],
  "session_data": "optional json string"
}
```

**Response**:
```json
{
  "app_session_id": "0x...",
  "version": "1",
  "status": "open"
}
```

**Validation**:
- At least 1 participant required
- Nonce must be non-zero
- Quorum cannot exceed total signature weights
- All weights must be non-negative
- Signatures must be provided and valid
- Achieved quorum must meet the required quorum threshold
- Each signature must be from a participant in the session

**Signature Verification**:
- Uses ABI encoding via `PackCreateAppSessionRequestV1` to create a deterministic hash
- Recovers signer addresses from ECDSA signatures
- Validates that signers are participants
- Accumulates signature weights to verify quorum is met

### 2. `app_sessions.v1.submit_deposit_state`

**Purpose**: Submits an app session deposit state update along with the associated user channel state.

**This endpoint performs TWO operations:**

1. **Channel State Operation** (UserState):
   - Processes the user's channel state (similar to `submit_state`)
   - **Validates that the last transition type is "commit"**
   - Validates user signature
   - Validates state transitions
   - Signs and stores the channel state
   - Records the commit transaction

2. **App Session Operation** (AppStateUpdate + SigQuorum):
   - Processes deposits into the app session
   - Updates app session version
   - Records ledger entries for deposits

**Key Features**:
- Combines channel state management with app session deposits
- Ensures atomicity between channel commit and app deposits
- Validates signatures on both app state and user state
- Ensures no conflicting channel operations

**Request**:
```json
{
  "app_state_update": {
    "app_session_id": "0x...",
    "intent": "deposit",
    "version": 2,
    "allocations": [
      {
        "participant": "0x...",
        "asset": "usdc",
        "amount": "1000"
      }
    ],
    "session_data": "optional json string",
    "signatures": ["0x...", "0x..."]
  },
  "sig_quorum": 1,
  "user_state": {
    // StateV1 object with user's channel state
  }
}
```

**Response**:
```json
{
  "signature": "0x..."  // Node's signature for the deposit state
}
```

**Validation**:

*Channel State Validation:*
- **Last transition must be "commit"**
- User state signature must be valid
- Channel state transitions must be valid
- User must have an open channel
- No ongoing state transitions
- Transition account ID must match app session ID
- **Total deposit amount must equal the commit transition amount**

*App Session Validation:*
- App session must exist and be open
- App session version must be sequential (current + 1)
- Intent must be "deposit"
- **Signatures must be provided and valid**
- **Achieved quorum must meet the required quorum threshold**
- Each signature must be from a participant in the session
- Allocations can only increase (deposits only, no decreases)
- **Allocation asset must match user state asset**
- Participant must have sufficient balance
- No challenged channels
- No conflicting allocations in other sessions

**Signature Verification**:
- Uses ABI encoding via `PackAppStateUpdateV1` to create a deterministic hash
- App session ID encoded as `bytes32`
- Allocation amounts encoded as `string` representation of decimals
- Recovers signer addresses from ECDSA signatures
- Validates that signers are participants
- Accumulates signature weights to verify quorum is met

## Flow Diagram

### `submit_deposit_state` Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                    submit_deposit_state Request                      │
│  ┌────────────────────┐         ┌───────────────────────────────┐  │
│  │   UserState        │         │   AppStateUpdate              │  │
│  │ (StateV1)          │         │ + SigQuorum                   │  │
│  │                    │         │                               │  │
│  │ - transitions      │         │ - app_session_id              │  │
│  │ - last = "commit"  │         │ - intent = "deposit"          │  │
│  │ - user_sig         │         │ - version                     │  │
│  └────────────────────┘         │ - allocations                 │  │
│                                  └───────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌───────────────────────────────────────────────────────┐
        │         PART 1: Channel State Operation                │
        │         (similar to submit_state)                      │
        └───────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌────────────────────────────────────────────────────┐
        │ 1. Parse & validate UserState                      │
        │ 2. Verify last transition = "commit"               │
        │ 3. Validate user signature                         │
        │ 4. Get current state from DB                       │
        │ 5. Validate state transitions                      │
        │ 6. Check open channel exists                       │
        │ 7. Ensure no ongoing transitions                   │
        └────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌───────────────────────────────────────────────────────┐
        │         PART 2: App Session Operation                  │
        │         (process deposits)                             │
        └───────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌────────────────────────────────────────────────────┐
        │ 1. Get & validate app session                      │
        │ 2. Verify version is sequential                    │
        │ 3. Verify intent = "deposit"                       │
        │ 4. Validate signatures & verify quorum:            │
        │    - Pack app state update using ABI encoding      │
        │    - Recover signer addresses from signatures      │
        │    - Verify signers are participants               │
        │    - Accumulate weights & check quorum met         │
        │ 5. Process each allocation:                        │
        │    - Check participant is valid                    │
        │    - Verify allocation asset matches user asset    │
        │    - Verify allocation increased (no decreases)    │
        │    - Check sufficient balance                      │
        │    - Accumulate total deposit amount               │
        │    - Record ledger entries                         │
        │ 6. Verify total deposits = commit amount           │
        │ 7. Update app session version                      │
        └────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌───────────────────────────────────────────────────────┐
        │         PART 3: Sign & Store Channel State            │
        └───────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌────────────────────────────────────────────────────┐
        │ 1. Sign UserState with node signature              │
        │ 2. Store UserState to DB                           │
        │ 3. Create transaction from "commit" transition     │
        │ 4. Record channel transaction                      │
        └────────────────────────────────────────────────────┘
                                    │
                                    ▼
        ┌────────────────────────────────────────────────────┐
        │ Response: { signature: "0x..." }                   │
        │ (Node's signature for the user state)              │
        └────────────────────────────────────────────────────┘
```

### 3. `app_sessions.v1.submit_app_state`

**Purpose**: Processes app session state updates for operate, withdraw, and close intents. This endpoint handles non-deposit state changes within an app session.

**Supported Intents**:
- **operate**: Redistribute funds between participants (total per asset remains constant)
- **withdraw**: Decrease participant allocations and return funds to channels
- **close**: Release all funds and mark session as closed

**Key Features**:
- Validates quorum-based consensus with participant signatures
- Enforces intent-specific validation rules
- Issues channel states for withdrawn/released funds
- Prevents deposit intents (must use `submit_deposit_state`)

**Request**:
```json
{
  "app_state_update": {
    "app_session_id": "0x...",
    "intent": "operate|withdraw|close",
    "version": 3,
    "allocations": [
      {
        "participant": "0x...",
        "asset": "usdc",
        "amount": "750"
      }
    ],
    "session_data": "optional json string"
  },
  "signatures": ["0x...", "0x..."]
}
```

**Response**:
```json
{
  "signature": ""  // Empty for operate/withdraw/close intents
}
```

**Validation**:

*Common Validation (All Intents):*
- App session must exist and be open
- App session version must be sequential (current + 1)
- Intent must be operate, withdraw, or close (not deposit)
- Signatures must be provided and valid
- Achieved quorum must meet the required quorum threshold
- Each signature must be from a participant in the session
- All allocations must be non-negative
- All allocations must be to valid participants
- **All allocation amounts must respect asset decimal precision** (validated via AssetStore)

*Intent-Specific Validation:*

**Operate Intent:**
- All current non-zero allocations must be included in request
- Total allocations per asset must match session balance exactly
- Allows redistribution between participants
- Records ledger entries for allocation changes
- Validates decimal precision for each allocation amount

**Withdraw Intent:**
- All current non-zero allocations must be included in request
- Allocations can only decrease or stay the same (no increases)
- Validates decimal precision for calculated withdrawal amounts
- Records negative ledger entries for withdrawals
- Issues channel states for participants receiving withdrawn funds
- Cannot add allocations for new participants

**Close Intent:**
- All current allocations must match exactly (no changes allowed)
- Releases ALL funds from session to participants
- Records negative ledger entries for all releases
- Issues channel states for all participants with allocations
- Marks app session as closed (IsClosed = true)
- Cannot have extra allocations not in current state

**Signature Verification**:
- Uses ABI encoding via `PackAppStateUpdateV1` to create a deterministic hash
- App session ID encoded as `bytes32`
- Allocation amounts encoded as `string` representation of decimals
- Recovers signer addresses from ECDSA signatures
- Validates that signers are participants
- Accumulates signature weights to verify quorum is met

**State Issuance for Withdrawals and Close**:
For withdraw and close intents, the handler issues new channel states to participants receiving funds:
- Creates a `release` transition in the participant's channel
- Signs the new state with node signature (unless last signed state was a lock)
- Stores the new channel state
- Records the release transaction

### 4. `app_sessions.v1.get_app_sessions`

**Purpose**: Retrieves application sessions with optional filtering by participant or app session ID. Includes participant allocations for each session.

**Key Features**:
- Query by app session ID or participant wallet address
- Optional status filtering (open/closed)
- Pagination support
- Returns current allocations for each participant

**Request**:
```json
{
  "app_session_id": "0x...",  // Optional: filter by app session ID
  // OR
  "participant": "0x...",      // Optional: filter by participant wallet
 
  "status": "open",            // Optional: filter by status (open/closed)
  "pagination": {              // Optional: pagination parameters
    "offset": 0,
    "limit": 10,
    "sort": "created_at DESC"
  }
}
```

**Response**:
```json
{
  "app_sessions": [
    {
      "app_session_id": "0x...",
      "status": "open",
      "participants": [
        {
          "wallet_address": "0x...",
          "signature_weight": 1
        }
      ],
      "quorum": 2,
      "version": 3,
      "nonce": 12345,
      "session_data": "optional json string",
      "allocations": [
        {
          "participant": "0x...",
          "asset": "usdc",
          "amount": "1000"
        }
      ]
    }
  ],
  "metadata": {
    "page": 1,
    "per_page": 10,
    "total_count": 25,
    "page_count": 3
  }
}
```

**Validation**:
- At least one of `app_session_id` or `participant` must be provided
- Status filter must be "open" or "closed" if provided
- Pagination parameters are optional

**Implementation Notes**:
- Fetches allocations for each session using `GetParticipantAllocations`
- Returns empty allocations array if no allocations exist
- Status is converted to string representation ("open"/"closed")
- SessionData is null if empty string

### 5. `app_sessions.v1.get_app_definition`

**Purpose**: Retrieves the application definition for a specific app session. Returns the immutable configuration established at session creation.

**Key Features**:
- Returns core session definition without state information
- Includes participants, quorum, and nonce
- Useful for signature verification and session validation

**Request**:
```json
{
  "app_session_id": "0x..."
}
```

**Response**:
```json
{
  "definition": {
    "application": "game",
    "participants": [
      {
        "wallet_address": "0x...",
        "signature_weight": 1
      }
    ],
    "quorum": 2,
    "nonce": 12345
  }
}
```

**Validation**:
- App session must exist
- Returns error if session not found

**Implementation Notes**:
- Definition includes the immutable session parameters
- Does not include dynamic state like version, status, or allocations
- Nonce is from the session definition (not current version)

### 6. `app_sessions.v1.submit_session_key_state`

**Purpose**: Submits a session key state for registration, rotation/update, or revocation. Session keys allow delegated signing for app sessions, enabling applications to sign on behalf of a user's wallet.

**Submit semantics**:
- **Registration**: first submit for a `(user, session_key)` pair (version=1, future `expires_at`).
- **Rotation/update**: bump version with a future `expires_at` to change scopes or extend lifetime.
- **Revocation**: bump version with `expires_at <= now`. The auth path stops accepting state signed by the key and the slot is freed against the per-user cap.

**Key Features**:
- Versioned session key states (each update increments the version)
- ABI-encoded signature verification ensures only the wallet owner can register session keys
- Supports scoping session keys to specific applications and app sessions
- Expiration enforcement

**Request**:
```json
{
  "state": {
    "user_address": "0x1234...",
    "session_key": "0xabcd...",
    "version": "1",
    "application_id": ["app1", "app2"],
    "app_session_id": ["0xSession1..."],
    "expires_at": "1762417328",
    "user_sig": "0x..."
  }
}
```

**Response**:
```json
{}
```

**Validation**:
- `user_address` must be a valid hex address
- `session_key` must be a valid hex address
- `version` must be greater than 0
- `expires_at` may be in the past — past values express revocation (the key is retired and the slot is freed)
- `user_sig` is required
- `application_ids` entries must match `^[a-z0-9_-]{1,66}$` (lowercase letters, digits, dashes, underscores, max 66 chars); malformed entries are rejected before signature verification
- `app_session_ids` entries must be 0x-prefixed 32-byte hashes in lowercase canonical form (`^0x[0-9a-f]{64}$`); checksummed/uppercase hex is rejected before signature verification
- Version must be sequential (latest_version + 1)
- Signature must recover to `user_address`
- The per-user cap (`NITRONODE_MAX_SESSION_KEYS_PER_USER`, default 100) is enforced whenever the submit transitions the slot from inactive to active: a brand-new key (no prior state) or a reactivation (previous latest state's `expires_at` was already in the past). Rotation/update against a still-active key, and revocation submits, are not subject to the cap. A value `<= 0` disables the cap entirely (unlimited session keys per user).

**Concurrency**: A `SELECT ... FOR UPDATE` is taken on a per-(user, session_key, kind) pointer row in `current_session_key_states_v1` so concurrent submits for the same key serialize and report a clean "expected version" error instead of racing on the history table's UNIQUE constraint.

**Signature Verification**:
- Uses ABI encoding via `PackAppSessionKeyStateV1` to create a deterministic hash
- Encodes: user_address (address), session_key (address), version (uint64), application_ids (bytes32[]), app_session_ids (bytes32[]), expires_at (uint64)
- Each application_id and app_session_id string is converted to bytes32 via `keccak256(utf8(id))` before ABI encoding
- The `user_sig` field is excluded from packing (it is the signature itself)
- Recovers signer address from ECDSA signature
- Validates that recovered address matches `user_address`

### 7. `app_sessions.v1.get_last_key_states`

**Purpose**: Retrieves the latest non-expired session key states for a user, with optional filtering by session key address.

**Request**:
```json
{
  "user_address": "0x1234...",
  "session_key": "0xabcd..."  // Optional filter
}
```

**Response**:
```json
{
  "states": [
    {
      "user_address": "0x1234...",
      "session_key": "0xabcd...",
      "version": "3",
      "application_id": ["app1"],
      "app_session_id": [],
      "expires_at": "1762417328",
      "user_sig": "0x..."
    }
  ]
}
```

**Validation**:
- `user_address` is required
- Returns only the latest version per session key
- Excludes expired session key states

**Pagination**: Optional `pagination` block (`limit`, `offset`); response includes a `pagination` block with `current_page`, `page_count`, `per_page`, `total_items`. Server-side default and max `limit` are both 10.

**Read path**: Filters `current_session_key_states_v1` by (user_address, kind=app_session) and JOINs the history table on (user_address, session_key, version). Per-request DB work is bounded by the number of distinct session keys for the user, regardless of version churn in history.

## Implementation Details

### Files

**API Layer** (`nitronode/api/app_session_v1/`):
- `handler.go` - Handler struct with signature validators and signer
- `create_app_session.go` - Create app session endpoint handler
- `submit_deposit_state.go` - Submit deposit state endpoint handler
- `submit_app_state.go` - Submit app state endpoint handler (operate, withdraw, close)
- `get_app_sessions.go` - Get app sessions endpoint handler (with filtering and pagination)
- `get_app_definition.go` - Get app definition endpoint handler
- `submit_session_key_state.go` - Submit session key state endpoint handler
- `get_last_key_states.go` - Get last session key states endpoint handler
- `interface.go` - Store and signature validator interfaces
- `utils.go` - Mapping functions between RPC and core types

**Business Logic** (`pkg/app/`):
- `app_session_v1.go` - Type definitions and ABI encoding functions

### ABI Encoding Functions

The implementation uses Ethereum ABI encoding for deterministic hashing and signature verification:

#### `GenerateAppSessionIDV1(definition AppDefinitionV1) (string, error)`
- Generates a deterministic app session ID using ABI encoding
- Encodes: application (string), participants (address[], uint8[]), quorum (uint64), nonce (uint64)
- Returns Keccak256 hash as hex string

#### `PackCreateAppSessionRequestV1(definition AppDefinitionV1, sessionData string) ([]byte, error)`
- Packs app session creation request for signature verification
- Encodes: application, participants, quorum, nonce, sessionData
- Returns Keccak256 hash of ABI-encoded data
- Used in `create_app_session` to verify participant signatures

#### `PackAppStateUpdateV1(stateUpdate AppStateUpdateV1) ([]byte, error)`
- Packs app state update for signature verification
- Encodes:
  - `appSessionID` as `bytes32` (using `common.HexToHash`)
  - `intent` as `uint8`
  - `version` as `uint64`
  - `allocations` as array of tuples (address, string, string)
  - `sessionData` as `string`
- Amount encoded as string representation of decimal for precision
- Returns Keccak256 hash of ABI-encoded data
- Used in `submit_deposit_state` to verify participant signatures

#### `GenerateSessionKeyStateIDV1(userAddress, sessionKey string, version uint64) (string, error)`
- Generates a deterministic ID from user_address, session_key, and version
- Encodes: user_address (address), session_key (address), version (uint64)
- Returns Keccak256 hash as hex string
- Used as the primary key for session key state records

#### `PackAppSessionKeyStateV1(state AppSessionKeyStateV1) ([]byte, error)`
- Packs session key state for signature verification using ABI encoding
- Encodes: user_address (address), session_key (address), version (uint64), application_ids (bytes32[]), app_session_ids (bytes32[]), expires_at (uint64)
- Each `application_id` and `app_session_id` string is converted to bytes32 via `keccak256(utf8(id))`, providing deterministic, collision-resistant identifiers in practice
- Excludes the `user_sig` field (it is the signature itself)
- Returns Keccak256 hash of ABI-encoded data
- Used in `submit_session_key_state` to verify the user's signature

### Dependencies

The implementation uses:
- `pkg/core` - State management, validation, and decimal precision utilities
- `pkg/rpc` - RPC types and framework
- `pkg/sign` - Cryptographic signing
- `pkg/log` - Logging
- `github.com/shopspring/decimal` - Precise decimal arithmetic
- `github.com/ethereum/go-ethereum/crypto` - Ethereum cryptography

### AssetStore Interface

The handler requires an `AssetStore` interface for asset metadata operations:

```go
type AssetStore interface {
    GetAssetDecimals(asset string) (uint8, error)
}
```

This is used for:
- **Decimal precision validation** - Ensures allocation amounts don't exceed the asset's decimal precision
- Asset-specific validation during state updates

### Store Interface

The service requires an `Store` interface for persistence operations:

```go
type Store interface {
    // App session operations
    CreateAppSession(session app.AppSessionV1) error
    GetAppSession(sessionID string) (*app.AppSessionV1, error)
    GetAppSessions(appSessionID *string, participant *string, status *string, params *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error)
    UpdateAppSession(session app.AppSessionV1) error
    GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error)
    GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error)

    // Ledger operations
    RecordLedgerEntry(accountID, asset string, amount decimal.Decimal, sessionKey *string) error
    GetAccountBalance(accountID, asset string) (decimal.Decimal, error)

    // Transaction recording
    RecordTransaction(tx core.Transaction) error

    // Channel state operations (used by submit_deposit_state)
    CheckOpenChannel(wallet, asset string) (bool, error)
    GetLastUserState(wallet, asset string, signed bool) (*core.State, error)
    StoreUserState(state core.State) error
    EnsureNoOngoingStateTransitions(wallet, asset string) error

    // Session key state operations
    StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error
    GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error)
    GetLastAppSessionKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error)
    GetAppSessionKeyOwner(sessionKey, appSessionId, applicationId string) (string, error)
}
```

**Key Interface Notes:**
- `GetAppSession`: Retrieves a single app session by ID (used by `get_app_definition` and `submit_app_state`)
- `GetAppSessions`: Retrieves multiple app sessions with filtering and pagination (used by `get_app_sessions`)
- `GetParticipantAllocations`: Returns current allocations per participant per asset (used by `get_app_sessions`)
- `RecordTransaction`: Records channel state transactions (commit transitions from submit_deposit_state)
- Channel state operations are needed because `submit_deposit_state` handles both channel and app session state

## Usage Example

### Initializing the Handler

```go
// Create signature validators
sigValidators := map[app_session_v1.SigType]app_session_v1.SigValidator{
    app_session_v1.EcdsaSigType: ecdsaValidator,
}

// Create the handler
handler := app_session_v1.NewHandler(
    storeTxProvider,  // StoreTxProvider (wraps store operations in transactions)
    assetStore,      // AssetStore (provides asset metadata like decimals)
    signer,          // sign.Signer
    stateAdvancer,   // core.StateAdvancer
    statePacker,     // core.StatePacker (packs state for signatures)
    sigValidators,   // map[SigType]SigValidator
    nodeAddress,     // string (node's wallet address)
)

// Register with RPC router
router.Register(rpc.AppSessionsV1CreateAppSessionMethod, handler.CreateAppSession)
router.Register(rpc.AppSessionsV1SubmitDepositStateMethod, handler.SubmitDepositState)
router.Register(rpc.AppSessionsV1SubmitAppStateMethod, handler.SubmitAppState)
router.Register(rpc.AppSessionsV1GetAppSessionsMethod, handler.GetAppSessions)
router.Register(rpc.AppSessionsV1GetAppDefinitionMethod, handler.GetAppDefinition)
router.Register(rpc.AppSessionsV1SubmitSessionKeyStateMethod, handler.SubmitSessionKeyState)
router.Register(rpc.AppSessionsV1GetLastKeyStatesMethod, handler.GetLastKeyStates)
```

## Key Implementation Decisions

### 1. App Session Creation with Zero Allocations

**As required**: App sessions are created with **0 allocations by default**.

Previous implementation allowed deposits during creation. New implementation:
- Creates session with empty allocations
- Deposits must be done through `submit_deposit_state`
- Simplifies creation logic
- Separates concerns

### 2. Dual Operation Flow in `submit_deposit_state`

**Critical Design**: The `submit_deposit_state` endpoint performs **TWO operations in sequence**:

#### Part 1: Channel State Operation (UserState)
- Similar to the `submit_state` channel endpoint
- **Validates last transition type = "commit"** (required)
- Processes the user's channel state
- Signs with node signature
- Stores channel state
- Records commit transaction

#### Part 2: App Session Operation (AppStateUpdate + SigQuorum)
- Processes deposits into the app session
- Validates quorum requirements
- Updates allocations
- Records ledger entries
- Updates app session version

This dual nature ensures atomicity between channel commits and app session deposits.

### 3. Cryptographic Security

**ABI Encoding for Signatures**:
- All signature verification uses Ethereum ABI encoding for deterministic hashing
- `GenerateAppSessionIDV1`: Uses ABI encoding to generate deterministic session IDs
- `PackCreateAppSessionRequestV1`: ABI-encodes session creation data for signature verification
- `PackAppStateUpdateV1`: ABI-encodes state updates with proper type handling:
  - App session ID as `bytes32` (not string) for efficient on-chain compatibility
  - Amounts as `string` representation for decimal precision
  - Addresses as native `address` type

**Quorum-Based Consensus**:
- Supports weighted signature schemes
- Each participant has a configurable signature weight
- Quorum threshold ensures sufficient consensus
- Prevents replay attacks through unique nonces
- ECDSA signature recovery for address validation

### 4. Architecture Pattern

Following `channel_v1` structure:
- Separate file per endpoint
- Business logic in `pkg/app`
- Thin API handlers
- Clear separation of concerns

## Differences from Legacy API

| Aspect | Legacy API | New V1 API |
|--------|-----------|-----------|
| App session creation | With initial allocations | **With 0 allocations** |
| Deposit handling | Part of creation | Separate `submit_deposit_state` |
| Channel state | Separate from app ops | **Integrated with deposits** |
| Transition validation | Basic | **Requires "commit" transition** |
| Signature verification | Custom/varied | **ABI encoding (Ethereum-compatible)** |
| Session ID generation | Hash of JSON | **ABI-encoded deterministic hash** |
| Amount handling | Varies | **String representation for precision** |
| Quorum validation | Not implemented | **Weighted signature quorum** |
| Deposit validation | Basic | **Asset matching + amount validation** |
| Architecture | Mixed concerns | **Clean separation** |
| File structure | Single file | **Separate file per endpoint** |

## Testing

The implementation includes comprehensive test coverage:

```bash
# Run all tests
cd nitronode/api/app_session_v1
go test -v

# Run specific test suites
go test -v -run TestSubmitAppState_.*     # All submit_app_state tests
go test -v -run TestSubmitDepositState_.* # All submit_deposit_state tests
go test -v -run TestCreateAppSession_.*   # All create_app_session tests
```




## References

- API Specification: `docs/api.yaml`
- RPC Types: `pkg/rpc/`
- Application Types: `pkg/app/`
- Core Package: `pkg/core/`
- Channel V1 Reference: `nitronode/api/channel_v1/`
