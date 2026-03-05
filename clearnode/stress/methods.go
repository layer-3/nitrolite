package stress

import (
	"context"
	"fmt"
	"strconv"

	"github.com/layer-3/nitrolite/pkg/core"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// MethodRegistry returns all available stress test methods.
func MethodRegistry() map[string]Runner {
	return map[string]Runner{
		"ping":                   poolRunner(stressPing),
		"get-config":             poolRunner(stressGetConfig),
		"get-blockchains":        poolRunner(stressGetBlockchains),
		"get-assets":             poolRunner(stressGetAssets),
		"get-balances":           poolRunner(stressGetBalances),
		"get-transactions":       poolRunner(stressGetTransactions),
		"get-home-channel":       poolRunner(stressGetHomeChannel),
		"get-escrow-channel":     poolRunner(stressGetEscrowChannel),
		"get-latest-state":       poolRunner(stressGetLatestState),
		"get-channel-key-states": poolRunner(stressGetLastChannelKeyStates),
		"get-app-sessions":       poolRunner(stressGetAppSessions),
		"get-app-key-states":     poolRunner(stressGetLastAppKeyStates),
		"transfer-roundtrip":     RunTransferStress,
		"app-session-lifecycle":  RunAppSessionLifecycleStress,
	}
}

// poolRunner wraps a Factory into a Runner that creates a connection pool,
// runs the test, and returns the report.
func poolRunner(factory Factory) Runner {
	return func(ctx context.Context, cfg *Config, spec TestSpec) (Report, error) {
		walletAddress, err := cfg.WalletAddress()
		if err != nil {
			return Report{}, err
		}

		fn, err := factory(spec.ExtraArgs, walletAddress)
		if err != nil {
			return Report{}, err
		}

		fmt.Printf("Opening %d WebSocket connections to %s...\n", spec.Connections, cfg.WsURL)
		clients, err := CreateClientPool(cfg.WsURL, cfg.PrivateKey, spec.Connections)
		if err != nil {
			return Report{}, fmt.Errorf("failed to create connection pool: %w", err)
		}
		defer CloseClientPool(clients)

		results, totalTime := RunTest(ctx, spec.TotalReqs, clients, fn)
		return ComputeReport(spec.Method, spec.TotalReqs, len(clients), results, totalTime), nil
	}
}

func stressPing(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		return client.Ping(ctx)
	}, nil
}

func stressGetConfig(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetConfig(ctx)
		return err
	}, nil
}

func stressGetBlockchains(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBlockchains(ctx)
		return err
	}, nil
}

func stressGetAssets(args []string, _ string) (MethodFunc, error) {
	var chainID *uint64
	if len(args) >= 1 {
		parsed, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chain_id: %s", args[0])
		}
		chainID = &parsed
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetAssets(ctx, chainID)
		return err
	}, nil
}

func stressGetBalances(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBalances(ctx, wallet)
		return err
	}, nil
}

func stressGetTransactions(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetTransactions(ctx, wallet, opts)
		return err
	}, nil
}

func stressGetHomeChannel(args []string, walletAddress string) (MethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = walletAddress
		if wallet == "" {
			return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: get-home-channel requires asset param, e.g. get-home-channel:1000:10:usdc")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetHomeChannel(ctx, wallet, asset)
		return err
	}, nil
}

func stressGetEscrowChannel(args []string, _ string) (MethodFunc, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: get-escrow-channel requires channel_id param")
	}
	channelID := args[0]

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetEscrowChannel(ctx, channelID)
		return err
	}, nil
}

func stressGetLatestState(args []string, walletAddress string) (MethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = walletAddress
		if wallet == "" {
			return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: get-latest-state requires asset param, e.g. get-latest-state:1000:10:usdc")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLatestState(ctx, wallet, asset, false)
		return err
	}, nil
}

func stressGetLastChannelKeyStates(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastChannelKeyStates(ctx, wallet, nil)
		return err
	}, nil
}

func stressGetAppSessions(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}

	limit := uint32(20)
	opts := &sdk.GetAppSessionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}
	if wallet != "" {
		opts.Participant = &wallet
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetAppSessions(ctx, opts)
		return err
	}, nil
}

func stressGetLastAppKeyStates(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastAppKeyStates(ctx, wallet, nil)
		return err
	}, nil
}
