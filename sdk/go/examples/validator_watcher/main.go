package main

// Validator Registration Monitor
//
// This example shows how to watch for ValidatorRegistered events on the ChannelHub
// contract. App builders should run this monitoring loop and alert users whenever an
// unexpected validator is registered — users then have a 1-day window
// (VALIDATOR_ACTIVATION_DELAY) to revoke their ERC20 approvals before the validator
// becomes usable. See contracts/SECURITY.md for the full security context.
//
// The RPC URL must be a WebSocket endpoint (wss://) because event subscriptions
// require a persistent connection. HTTP endpoints are not supported.
//
// Gap-free monitoring: each event carries a BlockNumber. On reconnect the example
// passes lastBlock+1 as fromBlock so any events emitted during the outage are
// replayed before live events resume — the 1-day safety window is preserved even
// across network interruptions.

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	nitronodeURL := "wss://nitronode-sandbox.yellow.org/v1/ws"
	// WebSocket RPC is required for event subscriptions.
	wsRPCURL := "wss://sepolia.drpc.org"
	chainID := uint64(11155111)

	// Load private key from environment to avoid accidental exposure in error messages.
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY env var not set")
	}

	stateSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("failed to create state signer: %v", err)
	}
	channelSigner, err := core.NewChannelDefaultSigner(stateSigner)
	if err != nil {
		log.Fatalf("failed to create channel signer: %v", err)
	}
	txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("failed to create tx signer: %v", err)
	}

	client, err := sdk.NewClient(
		nitronodeURL,
		channelSigner,
		txSigner,
		sdk.WithBlockchainRPC(chainID, wsRPCURL),
	)
	if err != nil {
		log.Fatalf("failed to create SDK client: %v", err)
	}
	defer client.Close()

	fmt.Printf("Monitoring ValidatorRegistered events on chain %d...\n", chainID)
	fmt.Println("Press Ctrl+C to stop.")

	// fromBlock tracks where to resume on reconnect. Zero means "start from latest"
	// on the first call; subsequent reconnects pass lastBlock+1 to replay any events
	// emitted while the subscription was down.
	var fromBlock uint64

	for ctx.Err() == nil {
		events, err := client.WatchValidatorRegistered(ctx, chainID, fromBlock)
		if err != nil {
			log.Printf("failed to start validator watcher (fromBlock=%d): %v — retrying in 5s", fromBlock, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		for ev := range events {
			handleValidatorRegistered(ev)
			// Advance fromBlock so the next reconnect replays from the block after this event.
			fromBlock = ev.BlockNumber + 1
		}

		if ctx.Err() != nil {
			break
		}
		log.Printf("Validator watcher subscription lost (will resume from block %d) — resubscribing in 5s", fromBlock)
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
	}

	fmt.Println("Validator watcher stopped.")
}

// handleValidatorRegistered is called for every ValidatorRegistered event.
// Replace with your own alerting, logging, or user-notification logic.
func handleValidatorRegistered(ev *core.ValidatorRegisteredEvent) {
	fmt.Printf(
		"[ALERT] New validator registered on chain %d at block %d: ID=%d address=%s\n",
		ev.BlockchainID, ev.BlockNumber, ev.ValidatorID, ev.Validator,
	)
	fmt.Println("Action required: verify this validator is expected.")
	fmt.Println("If unexpected, revoke all ERC20 approvals to the ChannelHub contract immediately.")
	fmt.Println("You have 1 day (VALIDATOR_ACTIVATION_DELAY) before the validator becomes active.")
}
