# MCP Vetting Results — Go Server

**38 passed, 0 failed, 38 total**

---

## Inventory

### 0.4: 30 resources registered (expected 30)

**Status:** PASS

```
30
```

### 0.5: 8 tools: explain_concept, get_rpc_method, lookup_method, lookup_rpc_method, lookup_type, scaffold_project, search_api, validate_import

**Status:** PASS

```
explain_concept, get_rpc_method, lookup_method, lookup_rpc_method, lookup_type, scaffold_project, search_api, validate_import
```

### 0.6: 3 prompts: build-ai-agent-app, create-channel-app, migrate-from-v053

**Status:** PASS

```
build-ai-agent-app, create-channel-app, migrate-from-v053
```

## GoSDK

### H.1: lookup_method("Deposit", go) — exists

**Status:** PASS

```
### Deposit
**Signature:**
```go
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (*core.State, error) {
```
**Category:** Channels & Transactions
**Description:** Deposit prepares a deposit state for the user's channel. This method handles two scenarios automatically: 1. If no channel exists: Creates a new channel with the initial deposit 2. If channel exists: Advances the state with a deposit transition The returned state is signed by bot
... (2024 chars total, truncated)
```

### H.2: lookup_method("Transfer", go) — exists

**Status:** PASS

```
### Transfer
**Signature:**
```go
func (c *Client) Transfer(ctx context.Context, recipientWallet string, asset string, amount decimal.Decimal) (*core.State, error) {
```
**Category:** Channels & Transactions
**Description:** Transfer prepares a transfer state to send funds to another wallet address. This method handles two scenarios automatically: 1. If no channel exists: Creates a new channel with the transfer transition 2. If channel exists: Advances the state with a transfer send transition T
... (1135 chars total, truncated)
```

### H.3: lookup_method("CreateAppSession", go) — exists

**Status:** PASS

```
### CreateAppSession
**Signature:**
```go
func (c *Client) CreateAppSession(ctx context.Context, definition app.AppDefinitionV1, sessionData string, quorumSigs []string, opts ...CreateAppSessionOptions) (string, string, string, error) {
```
**Category:** App Sessions
**Description:** CreateAppSession creates a new application session between participants. Parameters: - definition: The app definition with participants, quorum, application ID - sessionData: Optional JSON stringified session data -
... (1000 chars total, truncated)
```

### H.4: lookup_method("SubmitAppState", go) — exists

**Status:** PASS

```
### SubmitAppState
**Signature:**
```go
func (c *Client) SubmitAppState(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string) error {
```
**Category:** App Sessions
**Description:** SubmitAppState submits an app session state update. This method handles operate, withdraw, and close intents. For deposits, use SubmitAppSessionDeposit instead. Parameters: - appStateUpdate: The app state update (intent: operate, withdraw, or close) - quorumSigs: Participant signatures for th
... (795 chars total, truncated)
```

### H.5: lookup_method("CloseHomeChannel", go) — exists

**Status:** PASS

```
### CloseHomeChannel
**Signature:**
```go
func (c *Client) CloseHomeChannel(ctx context.Context, asset string) (*core.State, error) {
```
**Category:** Channels & Transactions
**Description:** CloseHomeChannel prepares a finalize state to close the user's channel for a specific asset. This creates a final state with zero user balance and submits it to the node. The returned state is signed by both the user and the node, but has not yet been submitted to the blockchain. Use Checkpoint to execute 
... (974 chars total, truncated)
```

### H.6: lookup_method("Checkpoint", go) — exists

**Status:** PASS

```
### Checkpoint
**Signature:**
```go
func (c *Client) Checkpoint(ctx context.Context, asset string) (string, error) {
```
**Category:** Channels & Transactions
**Description:** Checkpoint executes the blockchain transaction for the latest signed state. It fetches the latest co-signed state and, based on the transition type and on-chain channel status, calls the appropriate blockchain method. This is the only method that interacts with the blockchain. It should be called after any state-building m
... (1529 chars total, truncated)
```

### H.7: lookup_method("GetBalances", go) — exists

**Status:** PASS

```
### GetBalances
**Signature:**
```go
func (c *Client) GetBalances(ctx context.Context, wallet string) ([]core.BalanceEntry, error) {
```
**Category:** User Queries
**Description:** GetBalances retrieves the balance information for a user. Parameters: - wallet: The user's wallet address Returns: - Slice of Balance containing asset balances - Error if the request fails Example: balances, err := client.GetBalances(ctx, "0x1234567890abcdef...") for _, balance := range balances { fmt.Printf("%s: %s\n
... (536 chars total, truncated)
```

### H.8: lookup_method("GetConfig", go) — exists

**Status:** PASS

```
### GetConfig
**Signature:**
```go
func (c *Client) GetConfig(ctx context.Context) (*core.NodeConfig, error) {
```
**Category:** Node & Config
**Description:** GetConfig retrieves the clearnode configuration including node identity and supported blockchains. Returns: - NodeConfig containing the node address, version, and list of supported blockchain networks - Error if the request fails Example: config, err := client.GetConfig(ctx) if err != nil { log.Fatal(err) } fmt.Printf("Node: %s (v%s)\n", 
... (539 chars total, truncated)
```

### H.9: CloseAppSession — NOT found

**Status:** PASS
**Notes:** Correctly not found

```
No method matching "CloseAppSession" found. Available categories: Other, Node & Queries, Signing, Transactions, Channels, App Sessions
```

### H.10: scaffold go-transfer-app — has Checkpoint

**Status:** PASS

```
# Scaffold: go-transfer-app

## go.mod
```
module my-nitrolite-transfer-app

go 1.25.0

require (
	github.com/layer-3/nitrolite v0.0.0
	github.com/shopspring/decimal v1.4.0
)
```

## main.go
```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
	txSigner, _ := sign.NewEthereumRawSigner(
... (1415 chars total, truncated)
```

### H.11: scaffold go-app-session — quorum sigs, close intent

**Status:** PASS

```
# Scaffold: go-app-session

## go.mod
```
module my-nitrolite-app-session

go 1.25.0

require (
	github.com/layer-3/nitrolite v0.0.0
	github.com/shopspring/decimal v1.4.0
)
```

## main.go
```go
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_
... (2421 chars total, truncated)
```

### H.12: scaffold go-ai-agent

**Status:** PASS

```
# Scaffold: go-ai-agent

## go.mod
```
module my-nitrolite-ai-agent

go 1.25.0

require (
	github.com/layer-3/nitrolite v0.0.0
	github.com/shopspring/decimal v1.4.0
)
```

## main.go
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("AGENT_PRIVATE_KEY"))
	txSigner, _ := sign
... (1667 chars total, truncated)
```

### H.13: full Go transfer — CloseHomeChannel + Checkpoint

**Status:** PASS

```
# Complete Go Transfer Script

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    recipient := os.Getenv("RECIPIENT")
    var chainID uint64 = 11155111 // Sepolia

    stateSigner, err := sign.NewEthereumM
... (2132 chars total, truncated)
```

### H.14: full Go app-session — quorum sigs, version tracking

**Status:** PASS

```
# Complete Go App Session Script

```go
package main

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/layer-3/nitrolite/pkg/app"
    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    peerAddr := os.Getenv("PEER_ADDRESS")
    var chainID uint64 = 
... (2663 chars total, truncated)
```

### H.15: use-cases resource loads

**Status:** PASS

```
# Nitrolite Use Cases

What you can build with the Nitrolite SDK and state channels.

## Peer-to-Peer Payments
Instant, gas-free token transfers between users. Open a channel, transfer any amount instantly, settle on-chain only when needed.
**SDK methods:** `client.deposit()`, `client.transfer()`, `client.closeHomeChannel()`

## Gaming (Real-Time Wagering)
Turn-based or real-time games where players wager tokens. App sessions track game state; winners receive payouts automatically.
**SDK methods
... (2039 chars total, truncated)
```

### H.16: lookup_type("AppSessionV1", go) — found

**Status:** PASS

```
### AppSessionV1 (struct)
**Source:** pkg/app
```go
type AppSessionV1 struct {
SessionID     string
	ApplicationID string
	Participants  []AppParticipantV1
	Quorum        uint8
	Nonce         uint64
	Status        AppSessionStatus
	Version       uint64
	SessionData   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```
```

### H.17: lookup_type("ChannelStatus", go) — enum found

**Status:** PASS

```
### ChannelStatus (enum)
**Source:** pkg/core
**Values:**
- `ChannelStatusVoid`
- `ChannelStatusOpen`
- `ChannelStatusChallenged`
- `ChannelStatusClosed`
```

## Protocol

### F.1g: explain_concept("state channel")

**Status:** PASS

```
**Channel**

A state container shared between a user and a node that allows off-chain state updates while maintaining on-chain security guarantees. Each channel operates on a single unified asset.

---

**State**

An abstract data structure representing the current configuration of a protocol entity at a specific version.
```

### F.2g: explain_concept("app session")

**Status:** PASS

```
No exact match for "app session". Related concepts:

**Session Key**

A delegated signing key authorized by a participant's primary key to sign specific types of state updates on their behalf. Session key authorization MUST be associated with the same address as the channel's user or node participant.

---

**Application Session**

An extension that enables off-chain application functionality. Application sessions hold committed assets and maintain their own state.

---

**Application State**

T
... (685 chars total, truncated)
```

### F.4g: explain_concept("clearnode")

**Status:** PASS

```
No concept matching "clearnode" found. 34 concepts indexed from protocol terminology. Try broader terms like "channel", "state", "session", "escrow", "transfer".
```

### F.5g: explain_concept("made up thing") — graceful

**Status:** PASS

```
No concept matching "made up thing" found. 34 concepts indexed from protocol terminology. Try broader terms like "channel", "state", "session", "escrow", "transfer".
```

### F.6g: resource nitrolite://protocol/overview

**Status:** PASS

```
# Nitrolite Protocol Overview

Nitrolite is a state channel protocol that enables high-speed off-chain interactions between users while preserving on-chain security guarantees.

Users exchange signed state updates off-chain with Nodes that act as a hub connecting network participants. Any user can enforce the latest agreed state on the blockchain layer at any time.

## Table of Contents

1. [Overview](overview.md) — high-level protocol description and design goals
2. [Terminology](terminology.md
... (5058 chars total, truncated)
```

### F.7g: resource nitrolite://protocol/terminology

**Status:** PASS

```
# Terminology

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)

---

This document defines all protocol terms used throughout the Nitrolite protocol documentation.

Each term is defined once. All other documents MUST use these terms consistently.

## Naming Conventions

- Protocol entities use CamelCase (e.g., ChannelState, AppSession)
- Field names use CamelCase (e.g., ChannelId, StateVersion)
- Operations use lowercase with hyphens in document references (e.g., state-
... (7514 chars total, truncated)
```

### F.9g: resource nitrolite://security/overview

**Status:** PASS

```
# Security and Limitations

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)

---

This document describes the security guarantees of the Nitrolite protocol, its current trust assumptions, and the known limitations of the present version.

## Protocol Maturity

The core protocol functionality is implemented and operational. A user MAY operate over a unified asset, deposit and withdraw on any supported blockchain, and conduct the majority of interact
... (5370 chars total, truncated)
```

## RPC

### G.6: lookup_rpc_method("channels.v1.get_home_channel")

**Status:** PASS

```
## V1 RPC Method: `channels.v1.get_home_channel`

**Group:** channels
**Description:** Retrieve current on-chain home channel information

**Request fields:** wallet, asset
**Response fields:** channel

```

### G.7: lookup_rpc_method("app_sessions.v1.create_app_session")

**Status:** PASS

```
## V1 RPC Method: `app_sessions.v1.create_app_session`

**Group:** app_sessions
**Description:** Create a new application session between participants. The application must be registered in the app registry. If the application requires creation approval (creation_approval_not_required is false), an owner signature is required.

**Request fields:** definition, session_data, quorum_sigs, owner_sig
**Response fields:** app_session_id, version, status

```

### G.8: lookup_rpc_method("user.v1.get_balances")

**Status:** PASS

```
## V1 RPC Method: `user.v1.get_balances`

**Group:** user
**Description:** Retrieve the balances of the user in YN

**Request fields:** wallet
**Response fields:** balances

```

## Prompts

### J.4: create-channel-app — mentions Checkpoint

**Status:** PASS

```
Guide me through building a Nitrolite state channel application. Cover:

1. **Setup** — Install dependencies (@yellow-org/sdk, viem), create Client with config
2. **Authentication** — Connect wallet, establish WebSocket, authenticate with clearnode
3. **Channel Lifecycle** — Deposit (auto-creates channel), query channels, close channel
4. **Transfers** — Send tokens to another participant via state channels
5. **App Sessions** — Create sessions for multi-party apps, submit state, close
6. **Erro
... (1675 chars total, truncated)
```

### J.5: build-ai-agent-app

**Status:** PASS

```
I want to build an AI agent that uses Nitrolite state channels for payments. Guide me through:

1. **Agent Wallet Setup** — Create a wallet for the agent, configure the SDK client
2. **Channel Management** — Open a channel, deposit funds for the agent to use
3. **Automated Payments** — Implement a payment function the agent can call autonomously
4. **Session Key Delegation** — Set up a session key with spending caps for security
5. **Agent-to-Agent Payments** — Transfer funds between two autonom
... (1708 chars total, truncated)
```

## Edge

### K.5: scaffold("nonexistent") — error

**Status:** PASS

```
MCP error -32602: Input validation error: Invalid arguments for tool scaffold_project: [
  {
    "received": "nonexistent",
    "code": "invalid_enum_value",
    "options": [
      "transfer-app",
      "app-session",
      "ai-agent",
      "go-transfer-app",
      "go-app-session",
      "go-ai-agent"
    ],
    "path": [
      "template"
    ],
    "message": "Invalid enum value. Expected 'transfer-app' | 'app-session' | 'ai-agent' | 'go-transfer-app' | 'go-app-session' | 'go-ai-agent', recei
... (524 chars total, truncated)
```

### K.6: search_api("zzzzz", go) — no matches

**Status:** PASS

```
# Search results for "zzzzz"

No matches found. Try a broader term.

```

## Sweep

### L.11: resource nitrolite://go-api/methods

**Status:** PASS

```
# Nitrolite Go SDK — Client Methods

Package: `github.com/layer-3/nitrolite/sdk/go`

## Connection

### `NewClient`
```go
func NewClient(wsURL string, stateSigner core.ChannelSigner, rawSigner sign.Signer, opts ...Option) (*Client, error)
```
Creates a new Nitrolite SDK client connected to a clearnode

## Channels & Transactions

### `Deposit`
```go
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (*core.State, error) {
```
Deposit prepares
... (30225 chars total, truncated)
```

### L.12: resource nitrolite://go-api/types

**Status:** PASS

```
# Nitrolite Go SDK — Types

## pkg/app

### `AppStateUpdateIntent` (enum)
**Values:**
- `AppStateUpdateIntentOperate`
- `AppStateUpdateIntentDeposit`
- `AppStateUpdateIntentWithdraw`
- `AppStateUpdateIntentClose`
- `AppStateUpdateIntentRebalance`

### `AppSessionStatus` (enum)
**Values:**
- `AppSessionStatusVoid`
- `AppSessionStatusOpen`
- `AppSessionStatusClosed`

### `AppSessionV1` (struct)
```go
type AppSessionV1 struct {
SessionID     string
	ApplicationID string
	Participants  []AppParticip
... (27072 chars total, truncated)
```

### L.13: resource nitrolite://protocol/enforcement

**Status:** PASS

```
# State Enforcement

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)

---

This document describes how channel states are enforced on the blockchain layer.

## Purpose

Enforcement is the mechanism by which off-chain state is reflected on-chain. It serves two complementary roles:

1. **Regular state synchronization** — participants submit signed states to the blockchain layer to keep the on-chain record up-to-date with the latest off-
... (13071 chars total, truncated)
```

### L.14: resource nitrolite://protocol/auth-flow

**Status:** PASS

```
# Request Signing & Authorization

In v1, every RPC request includes a `sig` field — the client's signature over the entire `req` tuple. This is the authorization mechanism. There is no separate authentication handshake; request signatures are the identity proof.

## Session Keys

Session keys enable delegated signing with scoped permissions. They are managed via:
- `channels.v1.submit_session_key_state` — register/update channel session keys
- `app_sessions.v1.submit_session_key_state` — regist
... (1288 chars total, truncated)
```
