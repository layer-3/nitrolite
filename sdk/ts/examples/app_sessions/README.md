# App Sessions Lifecycle Example

This example demonstrates the complete lifecycle of Nitrolite app sessions, including:

1. **Create first app session** for wallet 1
2. **Deposit USDC** into first app session by wallet 1
3. **Create second app session** for wallet 2 with wallet 3 as a participant
4. **Deposit WETH** into second app session by wallet 2
5. **Redistribute app state** within app session so that participant with wallet 3 also has some allocation
6. **Rebalance 2 app sessions atomically**
7. **Wallet 3 withdraws** from his app session
8. **Close both app sessions**

## Prerequisites

- Node.js 18+ installed
- Three wallets with private keys
- Access to a Nitrolite node (default: `wss://nitronode-sandbox.yellow.org/v1/ws`)

## Setup

1. Install dependencies:

```bash
npm install
```

2. Edit `lifecycle.ts` and replace the placeholder private keys with your actual private keys:

```typescript
const wallet1PrivateKey = '0x7d60...'; // Replace with actual private key
const wallet2PrivateKey = '0x9b65...'; // Replace with actual private key
const wallet3PrivateKey = '0xf636...'; // Replace with actual private key
```

## Running the Example

```bash
npm run lifecycle
```

## What It Does

This script performs a complete app session lifecycle test:

### Step 1: Create App Session 1 (Single Participant)
- Creates an app session with wallet 1 as the only participant
- Uses a 100% quorum (only wallet 1 needs to sign)

### Step 2: Deposit USDC into Session 1
- Deposits 0.0001 USDC into the first app session
- Updates the app session state with the deposit

### Step 3: Create App Session 2 (Multi-Party)
- Creates an app session with wallet 2 and wallet 3 as participants
- Each participant has 50% signature weight
- Requires 100% quorum (both wallets must sign)

### Step 4: Deposit WETH into Session 2
- Deposits 0.015 WETH into the second app session by wallet 2
- Both participants sign the deposit operation

### Step 5: Redistribute Funds in Session 2
- Redistributes WETH within session 2
- Wallet 2 gets 0.01 WETH
- Wallet 3 gets 0.005 WETH
- This is an "operate" intent state update

### Step 6: Atomic Rebalance Across Sessions
- Performs an atomic rebalance across both app sessions
- Session 1: Wallet 1 gets 0.005 WETH and 0.00005 USDC
- Session 2: Wallet 2 gets 0.00005 USDC and 0.005 WETH, Wallet 3 gets 0.005 WETH
- All updates happen atomically or none at all

### Step 7: Wallet 3 Withdraws from Session 2
- Wallet 3 withdraws 0.004 WETH back to their channel
- Final allocations: Wallet 2 (0.00005 USDC, 0.005 WETH), Wallet 3 (0.001 WETH)

### Step 8: Close Both App Sessions
- Closes session 1 with wallet 1's signature
- Closes session 2 with both wallet 2 and wallet 3's signatures
- Finalizes all app session state

## Key Concepts Demonstrated

### App Definitions
- Application ID
- Participants with signature weights
- Quorum requirements
- Unique nonce for session creation

### App State Updates
- Different intents: Deposit, Operate, Withdraw, Close, Rebalance
- Version tracking
- Allocations per participant per asset
- Session data (JSON string)

### Multi-Signature Operations
- Creating app sessions with multiple participants
- State updates requiring quorum signatures
- Atomic operations across multiple sessions

### State Packing and Signing
- `packCreateAppSessionRequestV1` for session creation
- `packAppStateUpdateV1` for state updates
- EIP-191 message signing for participant signatures

## TypeScript SDK Features Used

- `Client.create()` - Create SDK clients
- `createSigners()` - Create both state and transaction signers from a private key
- `packCreateAppSessionRequestV1()` - Pack session creation for signing
- `packAppStateUpdateV1()` - Pack state updates for signing
- `createAppSession()` - Create new app sessions
- `submitAppSessionDeposit()` - Submit deposits to app sessions
- `submitAppState()` - Submit state updates (operate, withdraw, close)
- `rebalanceAppSessions()` - Atomic rebalancing across sessions
- `getAppSessions()` - Query app session information

## Comparison with Go SDK

This TypeScript example is a direct port of the Go SDK's `lifecycle.go` example:

| Feature | Go SDK | TypeScript SDK |
|---------|--------|----------------|
| Signing | `sign.NewEthereumMsgSigner()` | `createSigners()` |
| Packing | `app.PackCreateAppSessionRequestV1()` | `packCreateAppSessionRequestV1()` |
| Decimal handling | `decimal.NewFromFloat()` | `new Decimal()` |
| BigInt | `uint64(time.Now().UnixNano())` | `BigInt(Date.now() * 1000000)` |
| Client creation | `sdk.NewClient()` | `await Client.create()` |
| Error handling | Go error returns | Try-catch blocks |

## Notes

- All amounts are in their base units (e.g., USDC with 6 decimals)
- Session data is JSON-stringified
- Nonces are generated using nanosecond timestamps
- Version numbers must increment sequentially
- Allocations must be signed by all required participants based on quorum
