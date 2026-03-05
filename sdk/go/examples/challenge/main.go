package main

// Channel Challenge Test Flow
// This example simulates a malicious user depositing funds into a channel,
// then transferring those funds out of the channel, and finally submitting a challenge on-chain
// with the pre-transfer state to try to recover their balance.
//
// This example demonstrates:
// 1. Set up wallet and create an SDK client
// 2. Deposit funds into a channel and checkpoint on-chain
// 3. Save the post-deposit state (pre-transfer state)
// 4. Transfer the user's balance to another random wallet
// 5. Submit a challenge on-chain with the saved pre-transfer state

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	wsURL := "wss://deployment.yellow.org/ws"
	privateKeyHex := "0x7d6..."
	chainID := uint64(11155111)
	rpcUrl := "https://sepolia.drpc.org"

	stateMsgSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to create state signer: %v", err)
	}

	channelSigner, err := core.NewChannelDefaultSigner(stateMsgSigner)
	if err != nil {
		log.Fatalf("Failed to create channel signer: %v", err)
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to create tx signer: %v", err)
	}

	client, err := sdk.NewClient(
		wsURL,
		channelSigner,
		txSigner,
		sdk.WithBlockchainRPC(chainID, rpcUrl),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	walletAddress := client.GetUserAddress()
	fmt.Printf("Wallet Address: %s\n\n", walletAddress)

	// --- 1. Deposit funds into a channel ---
	fmt.Println("=== Step 1: Depositing USDC ===")

	depositAmount := decimal.NewFromFloat(1.0)
	_, err = client.Deposit(ctx, chainID, "usdc", depositAmount)
	if err != nil {
		log.Fatalf("Deposit failed: %v", err)
	}
	fmt.Printf("Deposit state prepared: %s USDC\n", depositAmount)

	// --- 2. Checkpoint deposit on-chain ---
	fmt.Println("=== Step 2: Checkpointing deposit on-chain ===")

	txHash, err := client.Checkpoint(ctx, "usdc")
	if err != nil {
		log.Fatalf("Checkpoint failed: %v", err)
	}
	fmt.Printf("Deposit checkpointed on-chain: %s\n\n", txHash)

	// --- 3. Save the post-deposit state (this is the state we'll challenge with) ---
	fmt.Println("=== Step 3: Saving pre-transfer state ===")

	preTransferState, err := client.GetLatestState(ctx, walletAddress, "usdc", true)
	if err != nil {
		log.Fatalf("Failed to get latest state: %v", err)
	}
	fmt.Printf("Saved state at version %d (balance: %s USDC)\n\n",
		preTransferState.Version, preTransferState.HomeLedger.UserBalance)

	<-time.After(10 * time.Second)

	// --- 4. Transfer balance to a random wallet ---
	fmt.Println("=== Step 4: Transferring balance to random wallet ===")

	recipientKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate recipient key: %v", err)
	}
	recipientAddress := crypto.PubkeyToAddress(recipientKey.PublicKey).Hex()
	fmt.Printf("Random recipient: %s\n", recipientAddress)

	transferAmount := depositAmount // transfer entire balance
	_, err = client.Transfer(ctx, recipientAddress, "usdc", transferAmount)
	if err != nil {
		log.Fatalf("Transfer failed: %v", err)
	}
	fmt.Printf("Transferred %s USDC to %s\n\n", transferAmount, recipientAddress)

	// --- 5. Submit challenge on-chain with the pre-transfer state ---
	fmt.Println("=== Step 5: Submitting challenge with pre-transfer state ===")

	challengeTxHash, err := client.Challenge(ctx, *preTransferState)
	if err != nil {
		log.Fatalf("Challenge failed: %v", err)
	}
	fmt.Printf("Challenge submitted on-chain: %s\n", challengeTxHash)
	fmt.Printf("Challenged with state version %d (balance: %s USDC)\n\n",
		preTransferState.Version, preTransferState.HomeLedger.UserBalance)

	fmt.Println("=== Example Complete ===")
	fmt.Println("The challenge period is now active on-chain.")
	fmt.Println("If the counterparty does not respond with a higher-versioned state,")
	fmt.Println("the channel can be closed with the challenged state after the period expires.")
}
