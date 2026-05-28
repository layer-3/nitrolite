package stress

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

type wallet struct {
	key    string
	addr   string
	client *sdk.Client
}

func deriveKey(masterKey string, index int) string {
	data := masterKey + ":receiver:" + strconv.Itoa(index)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func walletAddressFromKey(privateKey string) (string, error) {
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return "", err
	}
	return signer.PublicKey().Address().String(), nil
}

func createClient(wsURL, privateKey string) (*sdk.Client, error) {
	ethMsgSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}
	stateSigner, err := core.NewChannelDefaultSigner(ethMsgSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel signer: %w", err)
	}
	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	const maxRetries = 3
	var client *sdk.Client
	var connectErr error
	for attempt := range maxRetries + 1 {
		client, connectErr = sdk.NewClient(wsURL, stateSigner, txSigner, sdk.WithErrorHandler(func(_ error) {}))
		if connectErr == nil {
			return client, nil
		}
		if attempt < maxRetries {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			time.Sleep(backoff)
		}
	}
	return nil, connectErr
}

// RunTransferStress runs a parallel transfer stress test.
//
// Spec format: transfer-roundtrip:rounds:wallets:asset[:amount]
//   - wallets: number of derived wallets (must be even, each pair runs in parallel)
//   - rounds: number of back-and-forth rounds per pair
//
// Phases:
//  1. Fund distribution: sender sends amount to each derived wallet
//  2. Stress test: pairs of wallets transfer back and forth in parallel
//  3. Fund collection: all wallets return funds to sender
func RunTransferStress(ctx context.Context, cfg *Config, spec TestSpec) (Report, error) {
	if os.Getenv("STRESS_PRIVATE_KEY") == "" {
		return Report{}, fmt.Errorf("STRESS_PRIVATE_KEY is required for transfer-roundtrip (sender must have funds)")
	}

	rounds := spec.TotalReqs
	numWallets := spec.Connections
	if numWallets < 2 {
		numWallets = 2
	}
	if numWallets%2 != 0 {
		numWallets++
	}
	numPairs := numWallets / 2

	if len(spec.ExtraArgs) < 1 {
		return Report{}, fmt.Errorf("asset required: transfer-roundtrip:rounds:wallets:asset[:amount]")
	}
	asset := spec.ExtraArgs[0]

	amount := decimal.NewFromFloat(0.000001)
	if len(spec.ExtraArgs) > 1 {
		parsed, err := decimal.NewFromString(spec.ExtraArgs[1])
		if err != nil {
			return Report{}, fmt.Errorf("invalid amount %q: %w", spec.ExtraArgs[1], err)
		}
		amount = parsed
	}

	senderAddr, err := cfg.WalletAddress()
	if err != nil {
		return Report{}, err
	}

	// Derive wallets
	wallets := make([]wallet, numWallets)
	for i := range numWallets {
		key := deriveKey(cfg.PrivateKey, i)
		addr, err := walletAddressFromKey(key)
		if err != nil {
			return Report{}, fmt.Errorf("failed to derive wallet %d: %w", i, err)
		}
		wallets[i] = wallet{key: key, addr: addr}
	}

	totalStressOps := numWallets * rounds // each pair does rounds*2, numPairs pairs
	fmt.Printf("  Sender:      %s\n", senderAddr)
	fmt.Printf("  Wallets:     %d (%d pairs)\n", numWallets, numPairs)
	fmt.Printf("  Asset:       %s\n", asset)
	fmt.Printf("  Amount:      %s per transfer\n", amount.String())
	fmt.Printf("  Rounds/pair: %d\n", rounds)
	fmt.Printf("  Transfers:   %d (stress phase)\n", totalStressOps)

	// Connect sender
	fmt.Println("  Connecting sender...")
	senderClient, err := createClient(cfg.WsURL, cfg.PrivateKey)
	if err != nil {
		return Report{}, fmt.Errorf("failed to connect sender: %w", err)
	}
	defer senderClient.Close()

	// Connect all wallets
	fmt.Printf("  Connecting %d wallets...\n", numWallets)
	for i := range wallets {
		c, err := createClient(cfg.WsURL, wallets[i].key)
		if err != nil {
			// Clean up already-connected
			for _, w := range wallets[:i] {
				if w.client != nil {
					w.client.Close()
				}
			}
			return Report{}, fmt.Errorf("failed to connect wallet %d: %w", i, err)
		}
		wallets[i].client = c
		if (i+1)%10 == 0 || i+1 == numWallets {
			fmt.Printf("\r    Connected: %d/%d  ", i+1, numWallets)
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()
	defer func() {
		for _, w := range wallets {
			if w.client != nil {
				w.client.Close()
			}
		}
	}()

	// Phase 1: Distribute funds from sender to all wallets
	fmt.Printf("  Phase 1: Distributing funds to %d wallets...\n", numWallets)
	for i, w := range wallets {
		if ctx.Err() != nil {
			return Report{}, ctx.Err()
		}
		_, err := senderClient.Transfer(ctx, w.addr, asset, amount)
		if err != nil {
			return Report{}, fmt.Errorf("failed to fund wallet %d (%s): %w", i, w.addr, err)
		}
		if (i+1)%10 == 0 || i+1 == numWallets {
			fmt.Printf("\r    Funded: %d/%d  ", i+1, numWallets)
		}
	}
	fmt.Println()

	// Phase 2: Parallel stress test
	// Pairs: (wallet[0], wallet[1]), (wallet[2], wallet[3]), ...
	fmt.Printf("  Phase 2: Stress test (%d pairs x %d rounds)...\n", numPairs, rounds)

	results := make([]Result, totalStressOps)
	var completed int64
	start := time.Now()

	var wg sync.WaitGroup
	for p := range numPairs {
		wg.Add(1)
		go func(pairIdx int) {
			defer wg.Done()
			a := &wallets[pairIdx*2]
			b := &wallets[pairIdx*2+1]
			base := pairIdx * rounds * 2 // index into results

			for r := range rounds {
				if ctx.Err() != nil {
					break
				}

				// A → B
				idx := base + r*2
				t := time.Now()
				_, err := a.client.Transfer(ctx, b.addr, asset, amount)
				results[idx] = Result{Duration: time.Since(t), Err: err}
				c := atomic.AddInt64(&completed, 1)
				step := int64(totalStressOps)/20 + 1
				if c%step == 0 || c == int64(totalStressOps) {
					fmt.Printf("\r    Progress: %d/%d (%.0f%%)  ",
						c, totalStressOps, float64(c)/float64(totalStressOps)*100)
				}

				if ctx.Err() != nil {
					break
				}

				// B → A
				idx = base + r*2 + 1
				t = time.Now()
				_, err = b.client.Transfer(ctx, a.addr, asset, amount)
				results[idx] = Result{Duration: time.Since(t), Err: err}
				c = atomic.AddInt64(&completed, 1)
				if c%step == 0 || c == int64(totalStressOps) {
					fmt.Printf("\r    Progress: %d/%d (%.0f%%)  ",
						c, totalStressOps, float64(c)/float64(totalStressOps)*100)
				}
			}
		}(p)
	}

	wg.Wait()
	totalTime := time.Since(start)
	fmt.Println()

	// Phase 3: Collect funds back to sender
	fmt.Printf("  Phase 3: Collecting funds from %d wallets...\n", numWallets)
	collected := 0
	for i, w := range wallets {
		if ctx.Err() != nil {
			break
		}
		_, err := w.client.Transfer(ctx, senderAddr, asset, amount)
		if err != nil {
			fmt.Printf("\n    WARNING: Failed to collect from wallet %d (%s): %v", i, w.addr, err)
			continue
		}
		collected++
		if collected%10 == 0 || i+1 == numWallets {
			fmt.Printf("\r    Collected: %d/%d  ", collected, numWallets)
		}
	}
	fmt.Println()

	if collected < numWallets {
		fmt.Printf("  WARNING: Only collected from %d/%d wallets\n", collected, numWallets)
	}

	return ComputeReport("transfer-roundtrip", totalStressOps, numPairs, results, totalTime), nil
}
