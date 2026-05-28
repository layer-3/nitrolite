package stress

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// RunTest executes totalReqs calls of fn distributed across the client pool.
// Each connection sends requests sequentially — waits for a response before
// sending the next one. Multiple connections run in parallel.
func RunTest(ctx context.Context, totalReqs int, clients []*sdk.Client, fn MethodFunc) ([]Result, time.Duration) {
	numClients := len(clients)
	results := make([]Result, totalReqs)
	var completed int64

	// Split requests evenly across connections
	work := make([][]int, numClients)
	for i := range totalReqs {
		ci := i % numClients
		work[ci] = append(work[ci], i)
	}

	start := time.Now()

	var wg sync.WaitGroup
	for ci := range numClients {
		wg.Add(1)
		go func(clientIdx int) {
			defer wg.Done()
			client := clients[clientIdx]
			for _, idx := range work[clientIdx] {
				if ctx.Err() != nil {
					results[idx] = Result{Err: ctx.Err()}
					atomic.AddInt64(&completed, 1)
					continue
				}

				reqStart := time.Now()
				err := fn(ctx, client)
				results[idx] = Result{Duration: time.Since(reqStart), Err: err}

				c := atomic.AddInt64(&completed, 1)
				step := int64(totalReqs)/20 + 1
				if c%step == 0 || c == int64(totalReqs) {
					pct := float64(c) / float64(totalReqs) * 100
					fmt.Printf("\r  Progress: %d/%d (%.0f%%)  ", c, totalReqs, pct)
				}
			}
		}(ci)
	}

	wg.Wait()
	totalTime := time.Since(start)
	fmt.Println()

	return results, totalTime
}
