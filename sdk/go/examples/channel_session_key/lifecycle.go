package main

// Example: Channel Session Key Lifecycle
//
// Requirements to run this example:
//
//  1. A reachable nitronode WebSocket endpoint (set via wsURL below).
//     The default points at the stress environment.
//
//  2. One EVM wallet with a hex private key (replace the placeholder below).
//
//  3. Minimum off-chain (channel) balances on the node:
//       - 0.00005 YUSD   (one deposit + one withdraw via session key)
//       - 0.00005 YELLOW (one deposit + one withdraw via session key)
//
//     An open channel is NOT a hard prerequisite. If the wallet already has
//     funds on the node but no acknowledged channel yet, Acknowledge is run
//     first to open one.
//
//  4. chainID below must match the asset's home blockchain for your target
//     nitronode deployment, and rpcURL must point at a JSON-RPC endpoint for
//     that chain. Both Deposit and Withdraw are followed by an on-chain
//     Checkpoint; the example then polls GetHomeChannel until the node has
//     observed the checkpoint event before moving on. Without a working RPC
//     these calls fail.
//
// This example demonstrates:
//  1. Open YUSD and YELLOW channels for the wallet (Acknowledge)
//  2. Generate a fresh session key
//  3. Register session key v1 with both assets allowed
//  4. Deposit YUSD and YELLOW via a session-key-backed client (success)
//  5. Update session key v2 -> [YELLOW] only
//  6. Withdraw YELLOW (success); attempt YUSD withdraw via session key (expected fail)
//  7. Update session key v3 -> [YUSD] only
//  8. Withdraw YUSD (success); attempt YELLOW deposit via session key (expected fail)
//  9. Revoke session key v4 -> []
// 10. Attempt YUSD deposit, YELLOW deposit, and channel closure via session key
//     (all expected to fail)

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

func main() {
	ctx := context.Background()
	wsURL := "wss://nitronode-stress.yellow.org/v1/ws"

	// Replace with your hex private key. The wallet must have minimum off-chain
	// balance for YUSD and YELLOW; channels are auto-opened below if missing.
	walletPrivateKey := "0x7d6071201765d2630ca9eb83cbe3e2e2e76f9b56ea3ed13a49a00208ebcdf843"

	// chainID is the home blockchain ID used for Deposit / Withdraw calls. Set
	// it to the asset's home chain on the target nitronode deployment. 11155111
	// is Ethereum Sepolia (the stress environment).
	chainID := uint64(11155111)

	// rpcURL is a JSON-RPC endpoint for chainID. Replace with your own provider
	// if the public endpoint is rate-limited.
	rpcURL := "https://sepolia.drpc.org"

	// --- Setup wallet signers + wallet-backed SDK client ---
	walletRawSigner, err := sign.NewEthereumRawSigner(walletPrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet private key: %v", err)
	}
	walletMsgSigner, err := sign.NewEthereumMsgSignerFromRaw(walletRawSigner)
	if err != nil {
		log.Fatalf("Failed to create wallet msg signer: %v", err)
	}
	walletChannelSigner, err := core.NewChannelDefaultSigner(walletMsgSigner)
	if err != nil {
		log.Fatalf("Failed to create wallet channel signer: %v", err)
	}
	walletAddress := walletRawSigner.PublicKey().Address().String()
	fmt.Printf("Wallet: %s\n\n", walletAddress)

	walletClient, err := sdk.NewClient(wsURL, walletChannelSigner, walletRawSigner, sdk.WithBlockchainRPC(chainID, rpcURL))
	if err != nil {
		log.Fatalf("Failed to create wallet client: %v", err)
	}
	defer walletClient.Close()

	// --- Step 1: ensure YUSD and YELLOW channels are open ---
	fmt.Println("=== Step 1: Ensuring channels are open ===")
	ensureChannelOpen(ctx, walletClient, "yusd")
	ensureChannelOpen(ctx, walletClient, "yellow")
	fmt.Println()

	// --- Step 2: generate a fresh session key ---
	fmt.Println("=== Step 2: Generating session key ===")
	sessionKeyRawSigner, sessionKeyMsgSigner := generateSessionKey()
	sessionKeyAddress := sessionKeyRawSigner.PublicKey().Address().String()
	fmt.Printf("Session key: %s\n\n", sessionKeyAddress)

	// --- Step 3: register session key v1 with both assets allowed ---
	fmt.Println("=== Step 3: Registering session key v1 ([yusd, yellow]) ===")
	stateV1 := submitSessionKey(ctx, walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 1, []string{"yusd", "yellow"})
	fmt.Println("✓ v1 registered")
	fmt.Println()

	// --- Step 4: deposit YUSD and YELLOW via session-key client ---
	fmt.Println("=== Step 4: Depositing via session-key client (v1) ===")
	skClient1 := newSessionKeyClient(wsURL, walletRawSigner, sessionKeyMsgSigner, stateV1, sdk.WithBlockchainRPC(chainID, rpcURL))
	yusdDepositState, err := skClient1.Deposit(ctx, chainID, "yusd", decimal.NewFromFloat(0.00001))
	if err != nil {
		log.Fatalf("YUSD deposit via v1 failed: %v", err)
	}
	fmt.Println("✓ YUSD deposited via session key")
	checkpointAndWait(ctx, skClient1, "yusd", yusdDepositState.Version)

	yellowDepositState, err := skClient1.Deposit(ctx, chainID, "yellow", decimal.NewFromFloat(0.00001))
	if err != nil {
		log.Fatalf("YELLOW deposit via v1 failed: %v", err)
	}
	fmt.Println("✓ YELLOW deposited via session key")
	checkpointAndWait(ctx, skClient1, "yellow", yellowDepositState.Version)
	skClient1.Close()
	fmt.Println()

	// --- Step 5: update session key v2 -> [yellow] ---
	fmt.Println("=== Step 5: Updating session key v2 ([yellow]) ===")
	stateV2 := submitSessionKey(ctx, walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 2, []string{"yellow"})
	skClient2 := newSessionKeyClient(wsURL, walletRawSigner, sessionKeyMsgSigner, stateV2, sdk.WithBlockchainRPC(chainID, rpcURL))
	fmt.Println("✓ v2 registered")
	fmt.Println()

	// --- Step 6: withdraw YELLOW (ok); attempt YUSD withdraw (fail) ---
	fmt.Println("=== Step 6: Withdraw via v2 (yellow only) ===")
	yellowWithdrawState, err := skClient2.Withdraw(ctx, chainID, "yellow", decimal.NewFromFloat(0.000005))
	if err != nil {
		log.Fatalf("YELLOW withdraw via v2 failed: %v", err)
	}
	fmt.Println("✓ YELLOW withdrawn via session key")
	checkpointAndWait(ctx, skClient2, "yellow", yellowWithdrawState.Version)
	if _, err := skClient2.Withdraw(ctx, chainID, "yusd", decimal.NewFromFloat(0.000005)); err != nil {
		fmt.Printf("✓ Expected: YUSD withdraw rejected by node: %v\n", err)
	} else {
		fmt.Println("✗ Unexpected: YUSD withdraw succeeded under v2")
	}
	skClient2.Close()
	fmt.Println()

	// --- Step 7: update session key v3 -> [yusd] ---
	fmt.Println("=== Step 7: Updating session key v3 ([yusd]) ===")
	stateV3 := submitSessionKey(ctx, walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 3, []string{"yusd"})
	skClient3 := newSessionKeyClient(wsURL, walletRawSigner, sessionKeyMsgSigner, stateV3, sdk.WithBlockchainRPC(chainID, rpcURL))
	fmt.Println("✓ v3 registered")
	fmt.Println()

	// --- Step 8: withdraw YUSD (ok); attempt YELLOW deposit (fail) ---
	fmt.Println("=== Step 8: Withdraw via v3 (yusd only) ===")
	yusdWithdrawState, err := skClient3.Withdraw(ctx, chainID, "yusd", decimal.NewFromFloat(0.000005))
	if err != nil {
		log.Fatalf("YUSD withdraw via v3 failed: %v", err)
	}
	fmt.Println("✓ YUSD withdrawn via session key")
	checkpointAndWait(ctx, skClient3, "yusd", yusdWithdrawState.Version)
	if _, err := skClient3.Deposit(ctx, chainID, "yellow", decimal.NewFromFloat(0.000005)); err != nil {
		fmt.Printf("✓ Expected: YELLOW deposit rejected by node: %v\n", err)
	} else {
		fmt.Println("✗ Unexpected: YELLOW deposit succeeded under v3")
	}
	skClient3.Close()
	fmt.Println()

	// --- Step 9: revoke session key v4 -> [] ---
	// Empty assets disables every per-asset check on the node, so the next
	// version of the key cannot authorize any channel operation.
	fmt.Println("=== Step 9: Revoking session key v4 (empty assets) ===")
	stateV4 := submitSessionKey(ctx, walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 4, []string{})
	skClient4 := newSessionKeyClient(wsURL, walletRawSigner, sessionKeyMsgSigner, stateV4, sdk.WithBlockchainRPC(chainID, rpcURL))
	fmt.Println("✓ v4 registered (revoked)")
	fmt.Println()

	// --- Step 10: every session-key operation must fail ---
	fmt.Println("=== Step 10: Verifying revoked session key cannot operate ===")
	if _, err := skClient4.Deposit(ctx, chainID, "yusd", decimal.NewFromFloat(0.000005)); err != nil {
		fmt.Printf("✓ Expected: YUSD deposit rejected by node: %v\n", err)
	} else {
		fmt.Println("✗ Unexpected: YUSD deposit succeeded under v4")
	}
	if _, err := skClient4.Deposit(ctx, chainID, "yellow", decimal.NewFromFloat(0.000005)); err != nil {
		fmt.Printf("✓ Expected: YELLOW deposit rejected by node: %v\n", err)
	} else {
		fmt.Println("✗ Unexpected: YELLOW deposit succeeded under v4")
	}
	if _, err := skClient4.CloseHomeChannel(ctx, "yusd"); err != nil {
		fmt.Printf("✓ Expected: YUSD channel close rejected by node: %v\n", err)
	} else {
		fmt.Println("✗ Unexpected: YUSD channel close succeeded under v4")
	}
	skClient4.Close()

	fmt.Println("\n=== Example Complete ===")
}

// ensureChannelOpen guarantees that the wallet has an acknowledged channel
// open for asset. If the node holds no state for the wallet/asset pair, or
// the latest state is still awaiting the user's signature (or has been
// finalized), Acknowledge is invoked to create or progress the channel.
func ensureChannelOpen(ctx context.Context, client *sdk.Client, asset string) {
	wallet := client.GetUserAddress()
	state, err := client.GetLatestState(ctx, wallet, asset, false)
	if err != nil {
		log.Fatalf("failed to get latest %s state: %v", asset, err)
	}

	hasOpenChannel := state != nil &&
		state.HomeChannelID != nil &&
		!state.IsFinal() &&
		state.UserSig != nil
	if hasOpenChannel {
		fmt.Printf("✓ channel already open for %s\n", asset)
		return
	}

	if _, err := client.Acknowledge(ctx, asset); err != nil {
		log.Fatalf("failed to acknowledge %s channel: %v", asset, err)
	}
	fmt.Printf("✓ acknowledged channel for %s\n", asset)
}

// generateSessionKey produces a fresh keypair and returns both the raw signer
// (used as the SDK client's tx signer) and the EIP-191 msg signer (used to
// sign channel session-key authorization payloads and channel states).
func generateSessionKey() (sign.Signer, *sign.EthereumMsgSigner) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("failed to generate session key: %v", err)
	}
	privateKeyHex := hexutil.Encode(crypto.FromECDSA(privateKey))

	rawSigner, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("failed to create session key raw signer: %v", err)
	}
	msgSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("failed to create session key msg signer: %v", err)
	}
	return rawSigner, msgSigner
}

// submitSessionKey signs and submits a (version, assets) update for the
// channel session key using the wallet client. Returns the registered state
// (including UserSig + SessionKeySig) so the caller can derive the matching
// session-key channel signer for subsequent operations.
func submitSessionKey(
	ctx context.Context,
	walletClient *sdk.Client,
	walletAddress, sessionKeyAddress string,
	sessionKeyMsgSigner *sign.EthereumMsgSigner,
	version uint64,
	assets []string,
) core.ChannelSessionKeyStateV1 {
	state := core.ChannelSessionKeyStateV1{
		UserAddress: walletAddress,
		SessionKey:  sessionKeyAddress,
		Version:     version,
		Assets:      assets,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	userSig, err := walletClient.SignChannelSessionKeyState(state)
	if err != nil {
		log.Fatalf("failed to sign session key v%d state: %v", version, err)
	}
	state.UserSig = userSig

	sessionKeySig, err := sdk.SignChannelSessionKeyOwnership(state, sessionKeyMsgSigner)
	if err != nil {
		log.Fatalf("failed to sign session key v%d ownership: %v", version, err)
	}
	state.SessionKeySig = sessionKeySig

	if err := walletClient.SubmitChannelSessionKeyState(ctx, state); err != nil {
		log.Fatalf("failed to submit session key v%d: %v", version, err)
	}
	return state
}

// newSessionKeyClient builds an SDK client whose state signer is the channel
// session key derived from the registered state. All channel state operations
// (Deposit, Withdraw, CloseHomeChannel, ...) issued through this client are
// signed with the session key, and the node validates them against the latest
// registered (user, session_key, version) tuple — including the asset
// allow-list and expiry.
//
// rawSigner must remain the wallet's raw signer: the SDK uses it to derive
// the user address (GetUserAddress) and to look up state for the correct
// owner. Substituting the session key here would point the client at the
// session key's address — channels would be queried/created for the wrong
// owner.
func newSessionKeyClient(
	wsURL string,
	walletRawSigner sign.Signer,
	sessionKeyMsgSigner *sign.EthereumMsgSigner,
	state core.ChannelSessionKeyStateV1,
	opts ...sdk.Option,
) *sdk.Client {
	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(state.UserAddress, state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		log.Fatalf("failed to compute metadata hash for v%d: %v", state.Version, err)
	}

	channelSigner, err := core.NewChannelSessionKeySignerV1(sessionKeyMsgSigner, metadataHash.Hex(), state.UserSig)
	if err != nil {
		log.Fatalf("failed to build session-key channel signer for v%d: %v", state.Version, err)
	}

	client, err := sdk.NewClient(wsURL, channelSigner, walletRawSigner, opts...)
	if err != nil {
		log.Fatalf("failed to create session-key client for v%d: %v", state.Version, err)
	}
	return client
}

// checkpointAndWait runs Checkpoint for asset and polls GetHomeChannel until
// the node's observed on-chain state_version catches up to expectedVersion.
// Without this barrier the next deposit/withdraw can race the node's event
// ingestion and be rejected with "home deposit is still ongoing".
func checkpointAndWait(ctx context.Context, client *sdk.Client, asset string, expectedVersion uint64) {
	txHash, err := client.Checkpoint(ctx, asset)
	if err != nil {
		log.Fatalf("checkpoint %s failed: %v", asset, err)
	}
	fmt.Printf("  ↳ checkpoint %s tx %s submitted; waiting for node to observe state_version=%d...\n", asset, txHash, expectedVersion)

	wallet := client.GetUserAddress()
	deadline := time.Now().Add(2 * time.Minute)
	for {
		channel, err := client.GetHomeChannel(ctx, wallet, asset)
		if err != nil {
			log.Fatalf("failed to get home channel for %s: %v", asset, err)
		}
		if channel != nil && channel.StateVersion >= expectedVersion {
			fmt.Printf("  ↳ node observed state_version=%d for %s\n", channel.StateVersion, asset)
			return
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for %s to reach state_version=%d", asset, expectedVersion)
		}
		time.Sleep(2 * time.Second)
	}
}
