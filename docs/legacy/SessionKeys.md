# Session Keys

Session keys are delegated keys that enable applications to perform operations on behalf of a user's wallet with specified spending limits, permissions, and expiration times. They provide a secure way to grant limited access to applications without exposing the main wallet's private key.

> **Important:** Session keys are **no longer used as on-chain channel participant addresses** for new channels created after the v0.5.0 release. For all new channels, the wallet address is used directly as the participant address. However, session keys still function correctly for channels that were created before v0.5.0, ensuring backward compatibility.

## Core Concepts

### General Rules

> **Important:** When authenticating with an already registered session key, you must still provide all parameters in the `auth_request`. However, the configuration values (`application`, `allowances`, `scope`, and `expires_at`) from the request will be ignored, as the system uses the settings from the initial registration. You may provide arbitrary values for these fields, as they are required by the request format but will not be used.

### Applications

Each session key is associated with a specific **application name**, which identifies the application or service that will use the session key. The application name is also used to identify **app sessions** that are created using that session key.

This association serves several purposes:

- **Application Isolation**: Different applications get separate session keys, preventing one application from using another's delegated access
- **Access Control**: Operations performed with a session key are validated against the application specified during registration
- **Single Active Key**: Only one session key can be active per wallet+application combination. Registering a new session key for the same application automatically invalidates any existing session key for that application

> **Important:** Only one session key is allowed per wallet+application combination. If you register a new session key for the same application, the old one is automatically invalidated and removed from the database.

#### Special Application: "clearnode"

Session keys registered with the application name `"clearnode"` receive special treatment:

- **Root Access**: These session keys bypass spending allowance validation and application restrictions
- **Full Permissions**: They can perform any operation the wallet itself could perform
- **Backward Compatibility**: This special behavior facilitates migration from older versions
- **Expiration Still Applies**: Even with root access, the session key expires according to its `expires_at` timestamp

> **Note:** The "clearnode" application name is primarily for backward compatibility and will be deprecated after a migration period for developers.

### Expiration

All session keys must have an **expiration timestamp** (`expires_at`) that defines when the session key becomes invalid:

- **Future Timestamp Required**: The expiration time must be set to a future date when registering a session key
- **Automatic Invalidation**: Once the expiration time passes, the session key can no longer be used for any operations
- **No Re-registration**: It is not possible to re-register an expired session key. You must create a new session key instead
- **Applies to All Keys**: Even "clearnode" application session keys must respect the expiration timestamp

### Allowances

Allowances define **spending limits** for session keys, specifying which assets the session key can spend and how much:

```json
{
  "allowances": [
    {
      "asset": "usdc",
      "amount": "100.0"
    },
    {
      "asset": "eth",
      "amount": "0.5"
    }
  ]
}
```

#### Allowance Validation

- **Supported Assets Only**: All assets specified in allowances must be supported by the system. Unsupported assets cause authentication to fail
- **Usage Tracking**: The system tracks spending per session key by recording which session key was used for each ledger debit operation
- **Spending Limits**: Once a session key reaches its spending cap for an asset, further operations requiring that asset are rejected with: `"operation denied: insufficient session key allowance: X required, Y available"`
- **Empty Allowances**: Providing an empty `allowances` array (`[]`) means zero spending allowed for all assets—any operation attempting to spend funds will be rejected

#### Allowances for "clearnode" Application

Session keys with `application: "clearnode"` are exempt from allowance enforcement:

- **No Spending Limits**: Allowance checks are bypassed entirely
- **Full Financial Access**: These keys can spend any amount of any supported asset
- **Expiration Still Matters**: Even without allowance restrictions, the session key still expires according to its `expires_at` timestamp

## How to Manage Session Keys

### Clearnode

#### Create and Configure

To create a session key, use the `auth_request` method during authentication. This registers the session key with its configuration:

**Request:**

```json
{
  "req": [
    1,
    "auth_request",
    {
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "application": "Chess Game",
      "allowances": [
        {
          "asset": "usdc",
          "amount": "100.0"
        },
        {
          "asset": "eth",
          "amount": "0.5"
        }
      ],
      "scope": "app.create",
      "expires_at": 1762417328
    },
    1619123456789
  ],
  "sig": ["0x5432abcdef..."]
}
```

**Parameters:**

- `address` (required): The wallet address that owns this session key
- `session_key` (required): The address of the session key to register
- `application` (optional): Name of the application using this session key (defaults to "clearnode" if not provided)
- `allowances` (optional): Array of asset allowances specifying spending limits
- `scope` (optional): Permission scope (e.g., "app.create", "ledger.readonly"). **Note:** This feature is not yet implemented
- `expires_at` (required): Unix timestamp (in seconds) when this session key expires

> **Note:** When authenticating with an already registered session key, you must still fill in all fields in the request, at least with arbitrary values. This is required by the request itself, however, the values will be ignored as the system uses the session key configuration stored during initial registration. This behavior will be improved in future versions.

#### List Active Session Keys

Use the `get_session_keys` method to retrieve all active (non-expired) session keys for the authenticated user:

**Request:**

```json
{
  "req": [1, "get_session_keys", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [
    1,
    "get_session_keys",
    {
      "session_keys": [
        {
          "id": 1,
          "session_key": "0xabcdef1234567890...",
          "application": "Chess Game",
          "allowances": [
            {
              "asset": "usdc",
              "allowance": "100.0",
              "used": "45.0"
            },
            {
              "asset": "eth",
              "allowance": "0.5",
              "used": "0.0"
            }
          ],
          "scope": "app.create",
          "expires_at": "2024-12-31T23:59:59Z",
          "created_at": "2024-01-01T00:00:00Z"
        }
      ]
    },
    1619123456789
  ],
  "sig": ["0xabcd1234..."]
}
```

**Response Fields:**

- `id`: Unique identifier for the session key record
- `session_key`: The address of the session key
- `application`: Application name this session key is authorized for
- `allowances`: Array of allowances with usage tracking:
  - `asset`: Symbol of the asset (e.g., "usdc", "eth")
  - `allowance`: Maximum amount the session key can spend
  - `used`: Amount already spent by this session key
- `scope`: Permission scope (omitted if empty)
- `expires_at`: When this session key expires (ISO 8601 format)
- `created_at`: When the session key was created (ISO 8601 format)

#### Revoke a Session Key

To immediately invalidate a session key, use the `revoke_session_key` method:

**Request:**

```json
{
  "req": [
    1,
    "revoke_session_key",
    {
      "session_key": "0xabcdef1234567890..."
    },
    1619123456789
  ],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [
    1,
    "revoke_session_key",
    {
      "session_key": "0xabcdef1234567890..."
    },
    1619123456789
  ],
  "sig": ["0xabcd1234..."]
}
```

**Permission Rules:**

- A wallet can revoke any of its session keys
- A session key can revoke itself
- A session key with `application: "clearnode"` can revoke other session keys belonging to the same wallet
- A non-"clearnode" session key cannot revoke other session keys (only itself)

**Important Notes:**

- Revocation is **immediate and cannot be undone**
- After revocation, any operations attempted with the revoked session key will fail with a validation error
- The revoked session key will no longer appear in the `get_session_keys` response
- Revocation is useful for security purposes when a session key may have been compromised

**Error Cases:**

- Session key does not exist, belongs to another wallet, or is expired: `"operation denied: provided address is not an active session key of this user"`
- Non-"clearnode" session key attempting to revoke another session key: `"operation denied: insufficient permissions for the active session key"`

### Nitrolite V1 API

The V1 API provides versioned session key state management with ABI-encoded signature verification, scoping to applications and app sessions, and versioned state updates.

#### Submit Session Key State

Register or update a session key state via `app_sessions.v1.submit_session_key_state`. Each update creates a new version. The state must be signed by the user's wallet.

**State Fields:**

- `user_address`: The wallet address owning the session key
- `session_key`: The delegated key address
- `version`: Sequential version (starts at 1, increments by 1)
- `application_id`: Array of application IDs the key is authorized for
- `app_session_id`: Array of app session IDs the key is authorized for
- `expires_at`: Unix timestamp (seconds) for expiration
- `user_sig`: ECDSA signature over ABI-encoded state (excluding user_sig)

The signature is verified by ABI-encoding `(address, address, uint64, bytes32[], bytes32[], uint64)` and recovering the signer from the Keccak256 hash.

#### Get Last Key States

Retrieve active session key states via `app_sessions.v1.get_last_key_states`:

- Returns the latest version per session key for a given user
- Excludes expired states
- Optionally filter by `session_key` address

### Nitrolite SDK

The SDKs provide high-level methods for session key management:

**Go SDK:**

```go
// Build, sign, and submit a session key state
state, err := client.BuildSessionKeyState(ctx, "0xSessionKey...", appIDs, sessionIDs, expiresAt)
sig, err := client.SignSessionKeyState(state)
state.UserSig = sig
err = client.SubmitSessionKeyState(ctx, state)

// Query active session key states
states, err := client.GetLastKeyStates(ctx, userAddress, nil)
```

**TypeScript SDK:**

```typescript
// Submit a session key state
await client.submitSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  application_id: ['app1'],
  app_session_id: [],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: '0x...',
});

// Query active session key states
const states = await client.getLastKeyStates('0x1234...');
```

For detailed SDK documentation, refer to the [Go SDK README](../../sdk/go/README.md) and [TypeScript SDK README](../../sdk/ts/README.md).
