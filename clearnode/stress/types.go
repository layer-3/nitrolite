package stress

import (
	"context"
	"time"

	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// Runner is the unified signature for executing a stress test.
// Every method in the registry uses this — pool-based or custom.
type Runner func(ctx context.Context, cfg *Config, spec TestSpec) (Report, error)

// MethodFunc is the signature for a single stress test request against one client.
type MethodFunc func(ctx context.Context, client *sdk.Client) error

// Factory parses method-specific args and returns a MethodFunc.
type Factory func(args []string, walletAddress string) (MethodFunc, error)

// Result captures the outcome of a single request.
type Result struct {
	Duration time.Duration
	Err      error
}

// Report contains the full aggregated report after a stress test run.
type Report struct {
	Method      string
	TotalReqs   int
	Connections int
	Successful  int
	Failed      int
	TotalTime   time.Duration

	MinLatency    time.Duration
	MaxLatency    time.Duration
	AvgLatency    time.Duration
	MedianLatency time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration

	RequestsPerSec float64
	ErrorBreakdown map[string]int
}

// TestSpec represents a parsed test specification from CLI args.
type TestSpec struct {
	Method      string
	TotalReqs   int
	Connections int // 0 means use default from config
	ExtraArgs   []string
}
