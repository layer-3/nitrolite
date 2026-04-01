package stress

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// runStorm is the entry point for the "storm" stress test strategy.
// Currently supports: transfers
func runStorm(args []string) int {
	if len(args) == 0 {
		printStormUsage()
		return 1
	}

	parts := strings.Split(args[0], ":")
	if len(parts) < 1 {
		printStormUsage()
		return 1
	}

	switch parts[0] {
	case "transfers":
		return runTransferStorm(parts[1:])
	case "sessions":
		return runSessionStorm(parts[1:])
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Unknown storm method %q\n", parts[0])
		printStormUsage()
		return 1
	}
}

// runTransferStorm implements the binary-tree transfer storm.
//
// Spec: transfers:iterations:cycles:asset:amount
//
// The test creates a binary tree of wallets. In each forward iteration,
// every active wallet transfers to a new child wallet, doubling the active set.
// After reaching the iteration limit, a plateau phase runs the last-layer
// transfers back and forth for the given number of cycles. Finally, the
// reverse phase collects all funds back up the tree to the origin wallet.
func runTransferStorm(parts []string) int {
	if len(parts) < 4 {
		fmt.Fprintf(os.Stderr, "ERROR: transfers requires iterations, cycles, asset, and amount\n")
		fmt.Fprintf(os.Stderr, "Usage: clearnode stress-test storm transfers:<iterations>:<cycles>:<asset>:<amount>\n")
		return 1
	}

	iterations, err := strconv.Atoi(parts[0])
	if err != nil || iterations <= 0 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid iterations %q: must be positive integer\n", parts[0])
		return 1
	}

	cycles, err := strconv.Atoi(parts[1])
	if err != nil || cycles < 0 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid cycles %q: must be non-negative integer\n", parts[1])
		return 1
	}

	asset := parts[2]

	amount, err := decimal.NewFromString(parts[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid amount %q: %v\n", parts[3], err)
		return 1
	}

	if os.Getenv("STRESS_PRIVATE_KEY") == "" {
		fmt.Fprintf(os.Stderr, "ERROR: STRESS_PRIVATE_KEY is required for storm transfers (origin must have funds)\n")
		return 1
	}

	cfg, err := ReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	originAddr, err := cfg.WalletAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	totalNodes := int(math.Pow(2, float64(iterations)))
	derivedWallets := totalNodes - 1
	lastLayerSize := int(math.Pow(2, float64(iterations-1)))
	plateauTransfers := cycles * 2 * lastLayerSize // each cycle = back + forth
	totalTransfers := 2*derivedWallets + plateauTransfers
	originAmount := amount.Mul(decimal.NewFromInt(int64(totalNodes)))

	fmt.Printf("Transfer Storm\n")
	fmt.Printf("  Origin:       %s\n", originAddr)
	fmt.Printf("  Iterations:   %d\n", iterations)
	fmt.Printf("  Cycles:       %d\n", cycles)
	fmt.Printf("  Asset:        %s\n", asset)
	fmt.Printf("  Amount/leaf:  %s\n", amount.String())
	fmt.Printf("  Origin needs: %s %s\n", originAmount.String(), asset)
	fmt.Printf("  Wallets:      %d (derived)\n", derivedWallets)
	fmt.Printf("  Transfers:    %d total (%d forward + %d plateau + %d reverse)\n",
		totalTransfers, derivedWallets, plateauTransfers, derivedWallets)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Build wallet tree: origin at index 0, derived wallets at 1..totalNodes-1.
	// Binary tree: node i has children 2i+1 and 2i+2.
	// Connections are established lazily per-iteration to avoid hitting connection limits.
	wallets := make([]wallet, totalNodes)

	// Derive all wallets upfront (cheap, no connections).
	fmt.Printf("  Deriving %d wallets...\n", derivedWallets)
	wallets[0] = wallet{key: cfg.PrivateKey, addr: originAddr}
	for i := 1; i < totalNodes; i++ {
		key := deriveKey(cfg.PrivateKey, i-1)
		addr, err := walletAddressFromKey(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to derive wallet %d: %v\n", i, err)
			return 1
		}
		wallets[i] = wallet{key: key, addr: addr}
	}

	// Connect origin.
	fmt.Println("  Connecting origin...")
	originClient, err := createClient(cfg.WsURL, cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to connect origin: %v\n", err)
		return 1
	}
	wallets[0].client = originClient
	defer func() {
		for _, w := range wallets {
			if w.client != nil {
				w.client.Close()
			}
		}
	}()

	var results []Result
	start := time.Now()

	// Forward phase: iterations 1..iterations
	// Iteration i: 2^(i-1) senders, each transfers amount * 2^(iterations-i) to its child.
	// Before each iteration, connect the recipients that will be needed.
	for iter := 1; iter <= iterations; iter++ {
		sendersCount := int(math.Pow(2, float64(iter-1)))
		transferAmount := amount.Mul(decimal.NewFromInt(int64(math.Pow(2, float64(iterations-iter)))))

		// Connect recipients for this iteration.
		// Children at iteration i occupy indices 2^(i-1) .. 2^i - 1.
		layerStart := int(math.Pow(2, float64(iter-1)))
		layerEnd := int(math.Pow(2, float64(iter))) - 1
		if layerEnd >= totalNodes {
			layerEnd = totalNodes - 1
		}
		fmt.Printf("  Connecting wallets %d..%d for iteration %d...\n", layerStart, layerEnd, iter)
		for i := layerStart; i <= layerEnd; i++ {
			c, err := createClient(cfg.WsURL, wallets[i].key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: failed to connect wallet %d: %v\n", i, err)
				return 1
			}
			wallets[i].client = c
			time.Sleep(10 * time.Millisecond)
		}

		fmt.Printf("  Forward iteration %d/%d: %d transfers of %s %s\n",
			iter, iterations, sendersCount, transferAmount.String(), asset)

		iterResults := make([]Result, sendersCount)
		var wg sync.WaitGroup

		for s := range sendersCount {
			if ctx.Err() != nil {
				fmt.Fprintf(os.Stderr, "\nERROR: context cancelled: %v\n", ctx.Err())
				return 1
			}

			wg.Add(1)
			go func(senderIdx int) {
				defer wg.Done()

				// Senders at iteration i are all existing nodes: indices 0..2^(i-1)-1.
				// Children at iteration i: indices 2^(i-1)..2^i-1.
				// Sender senderIdx maps to child at 2^(iter-1) + senderIdx.
				senderNodeIdx := senderIdx
				childNodeIdx := int(math.Pow(2, float64(iter-1))) + senderIdx
				if childNodeIdx >= totalNodes {
					return
				}

				t := time.Now()
				_, err := wallets[senderNodeIdx].client.Transfer(ctx, wallets[childNodeIdx].addr, asset, transferAmount)
				iterResults[senderIdx] = Result{Duration: time.Since(t), Err: err}
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n    Transfer failed at iteration %d, sender %d: %v\n", iter, senderIdx, err)
				}
			}(s)
		}
		wg.Wait()

		if failed := collectErrors(iterResults); len(failed) > 0 {
			for _, e := range failed {
				fmt.Fprintf(os.Stderr, "ERROR: transfer failed during forward phase: %v\n", e)
			}
			printStormReport("storm-transfers", results, time.Since(start))
			return 1
		}
		results = append(results, iterResults...)
	}

	// Plateau phase: last-layer transfers bounce back and forth.
	if cycles > 0 {
		fmt.Printf("  Plateau phase: %d cycles of %d back-and-forth transfers...\n", cycles, lastLayerSize)
		parentStart := 0
		childStart := int(math.Pow(2, float64(iterations-1)))

		for c := 1; c <= cycles; c++ {
			// Back: children → parents
			fmt.Printf("  Plateau %d back: %d transfers of %s %s\n", c, lastLayerSize, amount.String(), asset)
			backResults := make([]Result, lastLayerSize)
			var wg sync.WaitGroup
			for s := range lastLayerSize {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					childIdx := childStart + idx
					parentIdx := parentStart + idx
					t := time.Now()
					_, err := wallets[childIdx].client.Transfer(ctx, wallets[parentIdx].addr, asset, amount)
					backResults[idx] = Result{Duration: time.Since(t), Err: err}
				}(s)
			}
			wg.Wait()
			if failed := collectErrors(backResults); len(failed) > 0 {
				for _, e := range failed {
					fmt.Fprintf(os.Stderr, "ERROR: transfer failed during plateau back: %v\n", e)
				}
				printStormReport("storm-transfers", results, time.Since(start))
				return 1
			}
			results = append(results, backResults...)

			// Forth: parents → children
			fmt.Printf("  Plateau %d forth: %d transfers of %s %s\n", c, lastLayerSize, amount.String(), asset)
			forthResults := make([]Result, lastLayerSize)
			for s := range lastLayerSize {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					parentIdx := parentStart + idx
					childIdx := childStart + idx
					t := time.Now()
					_, err := wallets[parentIdx].client.Transfer(ctx, wallets[childIdx].addr, asset, amount)
					forthResults[idx] = Result{Duration: time.Since(t), Err: err}
				}(s)
			}
			wg.Wait()
			if failed := collectErrors(forthResults); len(failed) > 0 {
				for _, e := range failed {
					fmt.Fprintf(os.Stderr, "ERROR: transfer failed during plateau forth: %v\n", e)
				}
				printStormReport("storm-transfers", results, time.Since(start))
				return 1
			}
			results = append(results, forthResults...)
		}
	}

	fmt.Println("  Plateau complete, starting reverse (cleanup)...")

	// Reverse phase: iterations countdown from `iterations` to 1.
	// Each child sends funds back to its parent.
	for iter := iterations; iter >= 1; iter-- {
		sendersCount := int(math.Pow(2, float64(iter-1)))
		transferAmount := amount.Mul(decimal.NewFromInt(int64(math.Pow(2, float64(iterations-iter)))))
		fmt.Printf("  Reverse iteration %d/%d: %d transfers of %s %s\n",
			iter, iterations, sendersCount, transferAmount.String(), asset)

		iterResults := make([]Result, sendersCount)
		var wg sync.WaitGroup

		for s := range sendersCount {
			if ctx.Err() != nil {
				fmt.Fprintf(os.Stderr, "\nERROR: context cancelled: %v\n", ctx.Err())
				return 1
			}

			wg.Add(1)
			go func(senderIdx int) {
				defer wg.Done()

				// Child (sender in reverse) at indices 2^(iter-1)..2^iter-1,
				// parent at index senderIdx (0..2^(iter-1)-1).
				childNodeIdx := int(math.Pow(2, float64(iter-1))) + senderIdx
				parentNodeIdx := senderIdx
				if childNodeIdx >= totalNodes {
					return
				}

				t := time.Now()
				_, err := wallets[childNodeIdx].client.Transfer(ctx, wallets[parentNodeIdx].addr, asset, transferAmount)
				iterResults[senderIdx] = Result{Duration: time.Since(t), Err: err}
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n    Reverse transfer failed at iteration %d, sender %d: %v\n", iter, senderIdx, err)
				}
			}(s)
		}
		wg.Wait()

		// Close children after reverse iteration — they won't be needed again.
		layerStart := int(math.Pow(2, float64(iter-1)))
		layerEnd := int(math.Pow(2, float64(iter))) - 1
		if layerEnd >= totalNodes {
			layerEnd = totalNodes - 1
		}
		for i := layerStart; i <= layerEnd; i++ {
			if wallets[i].client != nil {
				wallets[i].client.Close()
				wallets[i].client = nil
			}
		}

		if failed := collectErrors(iterResults); len(failed) > 0 {
			for _, e := range failed {
				fmt.Fprintf(os.Stderr, "ERROR: transfer failed during reverse phase: %v\n", e)
			}
			printStormReport("storm-transfers", results, time.Since(start))
			return 1
		}
		results = append(results, iterResults...)
	}

	totalTime := time.Since(start)
	fmt.Println("  Cleanup finished.")
	printStormReport("storm-transfers", results, totalTime)

	fmt.Println("PASS")
	return 0
}

func collectErrors(results []Result) []error {
	var errs []error
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, r.Err)
		}
	}
	return errs
}

func printStormReport(method string, results []Result, totalTime time.Duration) {
	report := ComputeReport(method, len(results), 0, results, totalTime)
	PrintReport(report)
}

// stormNode holds a wallet with its pre-computed app session signer.
type stormNode struct {
	wallet
	signer *app.AppSessionSignerV1
}

// runSessionStorm implements the ternary-growth app session storm.
//
// Spec: sessions:iterations:cycles:asset:amount
//
// Tree growth: each existing node opens a 3-participant session with 2 new children per iteration.
// Total nodes = 3^iterations. Indexing: at iteration i, parent p's children are at
// 3^(i-1) + 2*p and 3^(i-1) + 2*p + 1.
//
// Forward: parent deposits, reallocates to children, close.
// Plateau: last-layer sessions bounce back and forth for N cycles.
// Reverse: children deposit back, reallocate to parent, close.
func runSessionStorm(parts []string) int {
	if len(parts) < 4 {
		fmt.Fprintf(os.Stderr, "ERROR: sessions requires iterations, cycles, asset, and amount\n")
		fmt.Fprintf(os.Stderr, "Usage: clearnode stress-test storm sessions:<iterations>:<cycles>:<asset>:<amount>\n")
		return 1
	}

	iterations, err := strconv.Atoi(parts[0])
	if err != nil || iterations <= 0 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid iterations %q: must be positive integer\n", parts[0])
		return 1
	}

	cycles, err := strconv.Atoi(parts[1])
	if err != nil || cycles < 0 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid cycles %q: must be non-negative integer\n", parts[1])
		return 1
	}

	asset := parts[2]

	amount, err := decimal.NewFromString(parts[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid amount %q: %v\n", parts[3], err)
		return 1
	}

	if os.Getenv("STRESS_PRIVATE_KEY") == "" {
		fmt.Fprintf(os.Stderr, "ERROR: STRESS_PRIVATE_KEY is required for storm sessions (origin must have funds)\n")
		return 1
	}

	cfg, err := ReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	originAddr, err := cfg.WalletAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	totalNodes := int(math.Pow(3, float64(iterations)))
	derivedNodes := totalNodes - 1
	// Forward: 4 ops per session (create + deposit + reallocate + close)
	// Reverse: 5 ops per session (create + deposit_child1 + deposit_child2 + reallocate + close)
	forwardSessionsTotal := (totalNodes - 1) / 2 // sum of 3^(i-1) for i=1..iterations
	reverseSessionsTotal := forwardSessionsTotal
	lastLayerParents := int(math.Pow(3, float64(iterations-1)))
	// Plateau: each cycle = back (5 ops/session) + forth (4 ops/session) for lastLayerParents sessions
	plateauSessions := cycles * 2 * lastLayerParents
	plateauOps := cycles * lastLayerParents * (5 + 4)
	forwardOps := forwardSessionsTotal * 4
	reverseOps := reverseSessionsTotal * 5
	totalOps := forwardOps + plateauOps + reverseOps
	originAmount := amount.Mul(decimal.NewFromInt(int64(totalNodes)))

	fmt.Printf("Session Storm\n")
	fmt.Printf("  Origin:       %s\n", originAddr)
	fmt.Printf("  Iterations:   %d\n", iterations)
	fmt.Printf("  Cycles:       %d\n", cycles)
	fmt.Printf("  Asset:        %s\n", asset)
	fmt.Printf("  Amount/leaf:  %s\n", amount.String())
	fmt.Printf("  Origin needs: %s %s\n", originAmount.String(), asset)
	fmt.Printf("  Wallets:      %d (derived)\n", derivedNodes)
	fmt.Printf("  Sessions:     %d forward + %d plateau + %d reverse\n", forwardSessionsTotal, plateauSessions, reverseSessionsTotal)
	fmt.Printf("  Operations:   %d total (%d forward + %d plateau + %d reverse)\n", totalOps, forwardOps, plateauOps, reverseOps)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Derive all nodes upfront (cheap, no connections).
	fmt.Printf("  Deriving %d wallets and signers...\n", derivedNodes)
	nodes := make([]stormNode, totalNodes)

	// Origin at index 0.
	originSigner, err := createAppSessionSigner(cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create origin signer: %v\n", err)
		return 1
	}
	nodes[0] = stormNode{
		wallet: wallet{key: cfg.PrivateKey, addr: originAddr},
		signer: originSigner,
	}

	for i := 1; i < totalNodes; i++ {
		key := deriveKey(cfg.PrivateKey, i-1)
		addr, err := walletAddressFromKey(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to derive wallet %d: %v\n", i, err)
			return 1
		}
		signer, err := createAppSessionSigner(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to create signer for wallet %d: %v\n", i, err)
			return 1
		}
		nodes[i] = stormNode{
			wallet: wallet{key: key, addr: addr},
			signer: signer,
		}
	}

	// Connect origin.
	fmt.Println("  Connecting origin...")
	originClient, err := createClient(cfg.WsURL, cfg.PrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to connect origin: %v\n", err)
		return 1
	}
	nodes[0].client = originClient
	defer func() {
		for i := range nodes {
			if nodes[i].client != nil {
				nodes[i].client.Close()
			}
		}
	}()

	var results []Result
	appID := fmt.Sprintf("stress-storm-%d", time.Now().UnixNano())
	fmt.Printf("  App ID:       %s\n", appID)
	start := time.Now()

	// Forward phase.
	for iter := 1; iter <= iterations; iter++ {
		parentsCount := int(math.Pow(3, float64(iter-1)))
		layerBase := parentsCount // first child index for this iteration
		depositAmount := amount.Mul(decimal.NewFromInt(2)).Mul(decimal.NewFromInt(int64(math.Pow(3, float64(iterations-iter)))))
		reallocAmount := amount.Mul(decimal.NewFromInt(int64(math.Pow(3, float64(iterations-iter)))))

		// Connect children for this iteration.
		childStart := layerBase
		childEnd := layerBase + 2*parentsCount - 1
		fmt.Printf("  Connecting wallets %d..%d for forward iteration %d...\n", childStart, childEnd, iter)
		for i := childStart; i <= childEnd; i++ {
			c, err := createClient(cfg.WsURL, nodes[i].key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: failed to connect wallet %d: %v\n", i, err)
				return 1
			}
			nodes[i].client = c
			time.Sleep(10 * time.Millisecond)
		}

		// Parents that were children in the previous iteration need to
		// acknowledge to open a channel before they can deposit.
		// Previous iteration's children: indices prevLayerBase..layerBase-1.
		if iter >= 2 {
			prevLayerBase := int(math.Pow(3, float64(iter-2)))
			fmt.Printf("  Acknowledging wallets %d..%d...\n", prevLayerBase, layerBase-1)
			for i := prevLayerBase; i < layerBase; i++ {
				if _, err := nodes[i].client.Acknowledge(ctx, asset); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: acknowledge wallet %d: %v\n", i, err)
					return 1
				}
			}
		}

		fmt.Printf("  Forward iteration %d/%d: %d sessions, deposit %s, reallocate %s each\n",
			iter, iterations, parentsCount, depositAmount.String(), reallocAmount.String())

		iterResults := make([]Result, parentsCount*4)
		var wg sync.WaitGroup

		for s := range parentsCount {
			if ctx.Err() != nil {
				fmt.Fprintf(os.Stderr, "\nERROR: context cancelled: %v\n", ctx.Err())
				return 1
			}

			wg.Add(1)
			go func(sessionIdx int) {
				defer wg.Done()
				parentIdx := sessionIdx
				child1Idx := layerBase + 2*sessionIdx
				child2Idx := layerBase + 2*sessionIdx + 1
				base := sessionIdx * 4

				nonce := uint64(iter)*10000 + uint64(sessionIdx)
				results := executeForwardSession(ctx, nodes, parentIdx, child1Idx, child2Idx,
					asset, depositAmount, reallocAmount, appID, nonce)
				copy(iterResults[base:base+4], results)
			}(s)
		}
		wg.Wait()

		if failed := collectErrors(iterResults); len(failed) > 0 {
			for _, e := range failed {
				fmt.Fprintf(os.Stderr, "ERROR: session failed during forward phase: %v\n", e)
			}
			printStormReport("storm-sessions", results, time.Since(start))
			return 1
		}
		results = append(results, iterResults...)
	}

	// One-time acknowledge before plateau/reverse: all last-layer wallets
	// that will need to deposit. Channels stay open once acknowledged.
	// - Leaf children received funds from last forward close.
	// - Non-origin parents need channels for plateau forth deposits.
	{
		layerBase := int(math.Pow(3, float64(iterations-1)))
		childStart := layerBase
		childEnd := layerBase + 2*lastLayerParents - 1
		fmt.Printf("  Acknowledging leaf wallets %d..%d...\n", childStart, childEnd)
		for i := childStart; i <= childEnd; i++ {
			if _, err := nodes[i].client.Acknowledge(ctx, asset); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: acknowledge wallet %d: %v\n", i, err)
				return 1
			}
		}
	}

	// Plateau phase: last-layer sessions bounce back and forth.
	if cycles > 0 {
		layerBase := int(math.Pow(3, float64(iterations-1)))
		childAmount := amount // leaf amount
		depositAmount := amount.Mul(decimal.NewFromInt(2))

		fmt.Printf("  Plateau phase: %d cycles of %d back-and-forth sessions...\n", cycles, lastLayerParents)
		for c := 1; c <= cycles; c++ {
			// Back: children deposit, reallocate to parent, close (5 ops each).
			fmt.Printf("  Plateau %d back: %d sessions, children deposit %s each\n",
				c, lastLayerParents, childAmount.String())
			backResults := make([]Result, lastLayerParents*5)
			var wg sync.WaitGroup
			for s := range lastLayerParents {
				wg.Add(1)
				go func(sessionIdx int) {
					defer wg.Done()
					parentIdx := sessionIdx
					child1Idx := layerBase + 2*sessionIdx
					child2Idx := layerBase + 2*sessionIdx + 1
					base := sessionIdx * 5
					nonce := uint64(100+c)*10000 + uint64(sessionIdx)
					r := executeReverseSession(ctx, nodes, parentIdx, child1Idx, child2Idx,
						asset, childAmount, depositAmount, appID, nonce)
					copy(backResults[base:base+5], r)
				}(s)
			}
			wg.Wait()
			if failed := collectErrors(backResults); len(failed) > 0 {
				for _, e := range failed {
					fmt.Fprintf(os.Stderr, "ERROR: session failed during plateau back: %v\n", e)
				}
				printStormReport("storm-sessions", results, time.Since(start))
				return 1
			}
			results = append(results, backResults...)

			// Forth: parent deposits, reallocates to children, close (4 ops each).
			fmt.Printf("  Plateau %d forth: %d sessions, parent deposits %s each\n",
				c, lastLayerParents, depositAmount.String())
			forthResults := make([]Result, lastLayerParents*4)
			for s := range lastLayerParents {
				wg.Add(1)
				go func(sessionIdx int) {
					defer wg.Done()
					parentIdx := sessionIdx
					child1Idx := layerBase + 2*sessionIdx
					child2Idx := layerBase + 2*sessionIdx + 1
					base := sessionIdx * 4
					nonce := uint64(200+c)*10000 + uint64(sessionIdx)
					r := executeForwardSession(ctx, nodes, parentIdx, child1Idx, child2Idx,
						asset, depositAmount, childAmount, appID, nonce)
					copy(forthResults[base:base+4], r)
				}(s)
			}
			wg.Wait()
			if failed := collectErrors(forthResults); len(failed) > 0 {
				for _, e := range failed {
					fmt.Fprintf(os.Stderr, "ERROR: session failed during plateau forth: %v\n", e)
				}
				printStormReport("storm-sessions", results, time.Since(start))
				return 1
			}
			results = append(results, forthResults...)
		}
	}

	fmt.Println("  Starting reverse (cleanup)...")

	// Reverse phase.
	for iter := iterations; iter >= 1; iter-- {
		parentsCount := int(math.Pow(3, float64(iter-1)))
		layerBase := parentsCount
		childAmount := amount.Mul(decimal.NewFromInt(int64(math.Pow(3, float64(iterations-iter)))))
		collectAmount := childAmount.Mul(decimal.NewFromInt(2))

		fmt.Printf("  Reverse iteration %d/%d: %d sessions, children deposit %s each\n",
			iter, iterations, parentsCount, childAmount.String())

		iterResults := make([]Result, parentsCount*5)
		var wg sync.WaitGroup

		for s := range parentsCount {
			if ctx.Err() != nil {
				fmt.Fprintf(os.Stderr, "\nERROR: context cancelled: %v\n", ctx.Err())
				return 1
			}

			wg.Add(1)
			go func(sessionIdx int) {
				defer wg.Done()
				parentIdx := sessionIdx
				child1Idx := layerBase + 2*sessionIdx
				child2Idx := layerBase + 2*sessionIdx + 1
				base := sessionIdx * 5

				nonce := uint64(iterations+iter)*10000 + uint64(sessionIdx)
				results := executeReverseSession(ctx, nodes, parentIdx, child1Idx, child2Idx,
					asset, childAmount, collectAmount, appID, nonce)
				copy(iterResults[base:base+5], results)
			}(s)
		}
		wg.Wait()

		// Close children after reverse — they won't be needed again.
		childStart := layerBase
		childEnd := layerBase + 2*parentsCount - 1
		for i := childStart; i <= childEnd; i++ {
			if nodes[i].client != nil {
				nodes[i].client.Close()
				nodes[i].client = nil
			}
		}

		if failed := collectErrors(iterResults); len(failed) > 0 {
			for _, e := range failed {
				fmt.Fprintf(os.Stderr, "ERROR: session failed during reverse phase: %v\n", e)
			}
			printStormReport("storm-sessions", results, time.Since(start))
			return 1
		}
		results = append(results, iterResults...)
	}

	totalTime := time.Since(start)
	fmt.Println("  Cleanup finished.")
	printStormReport("storm-sessions", results, totalTime)

	fmt.Println("PASS")
	return 0
}

// executeForwardSession runs a single forward app session lifecycle:
// acknowledge (depositor) → create → deposit (parent) → reallocate (parent → children) → close.
// Returns 4 Results (acknowledge is not measured, it's setup).
func executeForwardSession(
	ctx context.Context,
	nodes []stormNode,
	parentIdx, child1Idx, child2Idx int,
	asset string,
	depositAmount, reallocAmount decimal.Decimal,
	appID string, nonce uint64,
) []Result {
	results := make([]Result, 4)
	parent := &nodes[parentIdx]
	child1 := &nodes[child1Idx]
	child2 := &nodes[child2Idx]

	definition := app.AppDefinitionV1{
		ApplicationID: appID,
		Participants: []app.AppParticipantV1{
			{WalletAddress: parent.addr, SignatureWeight: 1},
			{WalletAddress: child1.addr, SignatureWeight: 1},
			{WalletAddress: child2.addr, SignatureWeight: 1},
		},
		Quorum: 3,
		Nonce:  nonce,
	}

	signers := []*app.AppSessionSignerV1{parent.signer, child1.signer, child2.signer}

	// 1. Create
	createReq, err := app.PackCreateAppSessionRequestV1(definition, "{}")
	if err != nil {
		results[0] = Result{Err: fmt.Errorf("pack create: %w", err)}
		skipResults(results, 1, results[0].Err)
		return results
	}
	createSigs, err := stormSignAll(signers, createReq)
	if err != nil {
		results[0] = Result{Err: fmt.Errorf("sign create: %w", err)}
		skipResults(results, 1, results[0].Err)
		return results
	}

	t := time.Now()
	sessionID, _, _, err := parent.client.CreateAppSession(ctx, definition, "{}", createSigs)
	results[0] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 1, err)
		return results
	}

	// 2. Deposit (parent deposits)
	version := uint64(2)
	depositUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: parent.addr, Asset: asset, Amount: depositAmount},
		},
	}
	depositReq, _ := app.PackAppStateUpdateV1(depositUpdate)
	depositSigs, _ := stormSignAll(signers, depositReq)

	t = time.Now()
	_, err = parent.client.SubmitAppSessionDeposit(ctx, depositUpdate, depositSigs, asset, depositAmount)
	results[1] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 2, err)
		return results
	}
	version++

	// 3. Reallocate (parent → children)
	reallocUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: parent.addr, Asset: asset, Amount: decimal.Zero},
			{Participant: child1.addr, Asset: asset, Amount: reallocAmount},
			{Participant: child2.addr, Asset: asset, Amount: reallocAmount},
		},
	}
	reallocReq, _ := app.PackAppStateUpdateV1(reallocUpdate)
	reallocSigs, _ := stormSignAll(signers, reallocReq)

	t = time.Now()
	err = parent.client.SubmitAppState(ctx, reallocUpdate, reallocSigs)
	results[2] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 3, err)
		return results
	}
	version++

	// 4. Close
	closeUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: parent.addr, Asset: asset, Amount: decimal.Zero},
			{Participant: child1.addr, Asset: asset, Amount: reallocAmount},
			{Participant: child2.addr, Asset: asset, Amount: reallocAmount},
		},
	}
	closeReq, _ := app.PackAppStateUpdateV1(closeUpdate)
	closeSigs, _ := stormSignAll(signers, closeReq)

	t = time.Now()
	err = parent.client.SubmitAppState(ctx, closeUpdate, closeSigs)
	results[3] = Result{Duration: time.Since(t), Err: err}
	return results
}

// executeReverseSession runs a single reverse app session lifecycle:
// create → deposit (child1) → deposit (child2) → reallocate (children → parent) → close.
// Returns 5 Results.
func executeReverseSession(
	ctx context.Context,
	nodes []stormNode,
	parentIdx, child1Idx, child2Idx int,
	asset string,
	childAmount, collectAmount decimal.Decimal,
	appID string, nonce uint64,
) []Result {
	results := make([]Result, 5)
	parent := &nodes[parentIdx]
	child1 := &nodes[child1Idx]
	child2 := &nodes[child2Idx]

	definition := app.AppDefinitionV1{
		ApplicationID: appID,
		Participants: []app.AppParticipantV1{
			{WalletAddress: parent.addr, SignatureWeight: 1},
			{WalletAddress: child1.addr, SignatureWeight: 1},
			{WalletAddress: child2.addr, SignatureWeight: 1},
		},
		Quorum: 3,
		Nonce:  nonce,
	}

	signers := []*app.AppSessionSignerV1{parent.signer, child1.signer, child2.signer}

	// 1. Create
	createReq, err := app.PackCreateAppSessionRequestV1(definition, "{}")
	if err != nil {
		results[0] = Result{Err: fmt.Errorf("pack create: %w", err)}
		skipResults(results, 1, results[0].Err)
		return results
	}
	createSigs, err := stormSignAll(signers, createReq)
	if err != nil {
		results[0] = Result{Err: fmt.Errorf("sign create: %w", err)}
		skipResults(results, 1, results[0].Err)
		return results
	}

	t := time.Now()
	sessionID, _, _, err := parent.client.CreateAppSession(ctx, definition, "{}", createSigs)
	results[0] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 1, err)
		return results
	}

	// 2. Deposit child1
	version := uint64(2)
	deposit1Update := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: child1.addr, Asset: asset, Amount: childAmount},
		},
	}
	deposit1Req, _ := app.PackAppStateUpdateV1(deposit1Update)
	deposit1Sigs, _ := stormSignAll(signers, deposit1Req)

	t = time.Now()
	_, err = child1.client.SubmitAppSessionDeposit(ctx, deposit1Update, deposit1Sigs, asset, childAmount)
	results[1] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 2, err)
		return results
	}
	version++

	// 3. Deposit child2
	deposit2Update := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: child2.addr, Asset: asset, Amount: childAmount},
		},
	}
	deposit2Req, _ := app.PackAppStateUpdateV1(deposit2Update)
	deposit2Sigs, _ := stormSignAll(signers, deposit2Req)

	t = time.Now()
	_, err = child2.client.SubmitAppSessionDeposit(ctx, deposit2Update, deposit2Sigs, asset, childAmount)
	results[2] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 3, err)
		return results
	}
	version++

	// 4. Reallocate (children → parent)
	reallocUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: parent.addr, Asset: asset, Amount: collectAmount},
			{Participant: child1.addr, Asset: asset, Amount: decimal.Zero},
			{Participant: child2.addr, Asset: asset, Amount: decimal.Zero},
		},
	}
	reallocReq, _ := app.PackAppStateUpdateV1(reallocUpdate)
	reallocSigs, _ := stormSignAll(signers, reallocReq)

	t = time.Now()
	err = parent.client.SubmitAppState(ctx, reallocUpdate, reallocSigs)
	results[3] = Result{Duration: time.Since(t), Err: err}
	if err != nil {
		skipResults(results, 4, err)
		return results
	}
	version++

	// 5. Close
	closeUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: parent.addr, Asset: asset, Amount: collectAmount},
			{Participant: child1.addr, Asset: asset, Amount: decimal.Zero},
			{Participant: child2.addr, Asset: asset, Amount: decimal.Zero},
		},
	}
	closeReq, _ := app.PackAppStateUpdateV1(closeUpdate)
	closeSigs, _ := stormSignAll(signers, closeReq)

	t = time.Now()
	err = parent.client.SubmitAppState(ctx, closeUpdate, closeSigs)
	results[4] = Result{Duration: time.Since(t), Err: err}
	return results
}

func createAppSessionSigner(privateKey string) (*app.AppSessionSignerV1, error) {
	msgSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create msg signer: %w", err)
	}
	return app.NewAppSessionWalletSignerV1(msgSigner)
}

func stormSignAll(signers []*app.AppSessionSignerV1, data []byte) ([]string, error) {
	sigs := make([]string, len(signers))
	for i, s := range signers {
		sig, err := s.Sign(data)
		if err != nil {
			return nil, fmt.Errorf("signer %d: %w", i, err)
		}
		sigs[i] = sig.String()
	}
	return sigs, nil
}

func skipResults(results []Result, from int, cause error) {
	for i := from; i < len(results); i++ {
		results[i] = Result{Err: fmt.Errorf("skipped: %w", cause)}
	}
}

func printStormUsage() {
	fmt.Println("Usage: clearnode stress-test storm <method>:<params...>")
	fmt.Println()
	fmt.Println("Available methods:")
	fmt.Println()
	fmt.Println("  transfers:iterations:cycles:asset:amount")
	fmt.Println("    Binary-tree transfer storm. Each iteration doubles active wallets.")
	fmt.Println("    After iterations, plateau cycles bounce last-layer transfers back and forth.")
	fmt.Println("    Origin wallet needs amount * 2^iterations of the asset.")
	fmt.Println()
	fmt.Println("    Example: clearnode stress-test storm transfers:3:2:usdc:1")
	fmt.Println("      Iteration 1: A -> B (4 usdc)")
	fmt.Println("      Iteration 2: A -> C (2), B -> D (2)")
	fmt.Println("      Iteration 3: A -> E (1), B -> F (1), C -> G (1), D -> H (1)")
	fmt.Println("      Plateau 1 back:  E -> A, F -> B, G -> C, H -> D")
	fmt.Println("      Plateau 1 forth: A -> E, B -> F, C -> G, D -> H")
	fmt.Println("      Plateau 2 back:  E -> A, F -> B, G -> C, H -> D")
	fmt.Println("      Plateau 2 forth: A -> E, B -> F, C -> G, D -> H")
	fmt.Println("      Reverse: E -> A, F -> B, G -> C, H -> D, C -> A, D -> B, B -> A")
	fmt.Println()
	fmt.Println("  sessions:iterations:cycles:asset:amount")
	fmt.Println("    Ternary-growth app session storm. Each iteration triples active wallets.")
	fmt.Println("    After iterations, plateau cycles bounce last-layer sessions back and forth.")
	fmt.Println("    Origin wallet needs amount * 3^iterations of the asset.")
	fmt.Println()
	fmt.Println("    Example: clearnode stress-test storm sessions:2:2:usdc:1")
	fmt.Println("      Iteration 1: session(A,B,C) — A deposits 6, reallocates 3 to B, 3 to C")
	fmt.Println("      Iteration 2: session(A,D,E), session(B,F,G), session(C,H,I) — each deposits 2, reallocates 1 each")
	fmt.Println("      Plateau 1 back:  D,E -> A; F,G -> B; H,I -> C")
	fmt.Println("      Plateau 1 forth: A -> D,E; B -> F,G; C -> H,I")
	fmt.Println("      Plateau 2 back:  D,E -> A; F,G -> B; H,I -> C")
	fmt.Println("      Plateau 2 forth: A -> D,E; B -> F,G; C -> H,I")
	fmt.Println("      Reverse 2: D,E -> A; F,G -> B; H,I -> C")
	fmt.Println("      Reverse 1: B,C -> A")
}
