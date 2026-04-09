package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/layer-3/nitrolite/clearnode/api"
	"github.com/layer-3/nitrolite/clearnode/event_handlers"
	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/clearnode/store/database"
	"github.com/layer-3/nitrolite/clearnode/stress"
	"github.com/layer-3/nitrolite/pkg/blockchain/evm"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "stress-test" {
		os.Exit(stress.Run(os.Args[2:]))
	}

	if len(os.Args) > 1 && os.Args[1] == "operator" {
		runOperatorCommand(os.Args[2:])
		return
	}

	bb := InitBackbone()
	logger := bb.Logger
	rootCtx := context.Background()
	blockchainCtx, cancelBlockchain := context.WithCancel(rootCtx)
	ctx := rootCtx

	vl := bb.ValidationLimits
	rpcRouterCfg := api.RPCRouterConfig{
		NodeVersion:               bb.NodeVersion,
		MinChallenge:              bb.ChannelMinChallengeDuration,
		MaxChallenge:              bb.ChannelMaxChallengeDuration,
		MaxParticipants:           vl.MaxParticipants,
		MaxSessionDataLen:         vl.MaxSessionDataLen,
		MaxAppMetadataLen:         vl.MaxAppMetadataLen,
		MaxRebalanceSignedUpdates: vl.MaxSignedUpdates,
		MaxSessionKeyIDs:          vl.MaxSessionKeyIDs,
		RateLimitPerSec:           bb.RateLimitPerSec,
		RateLimitBurst:            bb.RateLimitBurst,
	}
	api.NewRPCRouter(rpcRouterCfg, bb.RpcNode, bb.StateSigner, bb.DbStore, bb.MemoryStore, bb.ActionGateway, bb.RuntimeMetrics, bb.Logger)

	rpcListenAddr := ":7824"
	rpcListenEndpoint := "/ws"
	rpcMux := http.NewServeMux()
	rpcMux.HandleFunc(rpcListenEndpoint, bb.RpcNode.ServeHTTP)

	rpcServer := &http.Server{
		Addr:    rpcListenAddr,
		Handler: rpcMux,
	}

	blockchains, err := bb.MemoryStore.GetBlockchains()
	if err != nil {
		logger.Fatal("failed to get blockchains from memory store", "error", err)
	}

	wrapInTx := func(handler func(database.DatabaseStore) error) error {
		return bb.DbStore.ExecuteInTransaction(handler)
	}
	useEHV1StoreInTx := func(h event_handlers.StoreTxHandler) error {
		return wrapInTx(func(s database.DatabaseStore) error { return h(s) })
	}

	eventHandlerService := event_handlers.NewEventHandlerService(useEHV1StoreInTx, logger)

	for _, b := range blockchains {
		rpcURL, ok := bb.BlockchainRPCs[b.ID]
		if !ok {
			logger.Fatal("no RPC URL configured for blockchain", "blockchainID", b.ID)
		}

		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			logger.Fatal("failed to connect to EVM Node")
		}

		if b.ChannelHubAddress != "" {
			// For the node itself, the node address is the signer's address
			nodeAddress := bb.StateSigner.PublicKey().Address().String()

			clientOpts := []evm.ClientOption{
				evm.ClientBalanceCheck{RequireBalanceCheck: false},
				evm.ClientAllowanceCheck{RequireAllowanceCheck: false},
			}

			blockchainClient, err := evm.NewBlockchainClient(common.HexToAddress(b.ChannelHubAddress), client, bb.TxSigner, b.ID, nodeAddress, bb.MemoryStore, clientOpts...)
			if err != nil {
				logger.Fatal("failed to create EVM client")
			}

			sigValidators, err := bb.MemoryStore.GetChannelSigValidators(b.ID)
			if err != nil {
				logger.Fatal("failed to get channel signature validators from memory store", "error", err, "blockchainID", b.ID)
			}

			if err := ensureSigValidatorsRegistered(blockchainClient, sigValidators, true); err != nil {
				logger.Fatal("failed to ensure signature validators are registered", "error", err, "blockchainID", b.ID)
			}

			reactor := evm.NewChannelHubReactor(b.ID, eventHandlerService, bb.DbStore.StoreContractEvent)
			reactor.SetOnEventProcessed(bb.RuntimeMetrics.IncBlockchainEvent)
			l := evm.NewListener(common.HexToAddress(b.ChannelHubAddress), client, b.ID, b.BlockStep, logger, reactor.HandleEvent, bb.DbStore)
			l.Listen(blockchainCtx, func(err error) {
				if err != nil {
					logger.Fatal("blockchain listener stopped", "error", err, "blockchainID", b.ID)
				}
			})

			worker := NewBlockchainWorker(b.ID, blockchainClient, bb.DbStore, logger, bb.RuntimeMetrics)
			worker.Start(blockchainCtx, func(err error) {
				if err != nil {
					logger.Fatal("blockchain worker stopped", "error", err, "blockchainID", b.ID)
				}
			})
		} else {
			logger.Info("channel hub address is not configured for blockchain", "blockchainID", b.ID)
		}

		if b.LockingContractAddress != "" {
			appRegistryClient, err := evm.NewLockingClient(common.HexToAddress(b.LockingContractAddress), client, b.ID)
			if err != nil {
				logger.Fatal("failed to create locking client", "error", err, "blockchainID", b.ID)
			}

			reactor, err := evm.NewLockingContractReactor(b.ID, eventHandlerService, appRegistryClient.GetTokenDecimals, bb.DbStore.StoreContractEvent)
			if err != nil {
				logger.Fatal("failed to create app registry reactor", "error", err, "blockchainID", b.ID)
			}

			reactor.SetOnEventProcessed(bb.RuntimeMetrics.IncBlockchainEvent)
			l := evm.NewListener(common.HexToAddress(b.LockingContractAddress), client, b.ID, b.BlockStep, logger, reactor.HandleEvent, bb.DbStore)
			l.Listen(blockchainCtx, func(err error) {
				if err != nil {
					logger.Fatal("blockchain listener stopped", "error", err, "blockchainID", b.ID)
				}
			})
		}
	}

	go runStoreMetricsExporter(ctx, 30*time.Second, bb.DbStore, bb.StoreMetrics, logger)

	metricsListenAddr := ":4242"
	metricsEndpoint := "/metrics"
	// Set up a separate mux for metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle(metricsEndpoint, promhttp.Handler())

	// Start metrics server on a separate port
	metricsServer := &http.Server{
		Addr:    metricsListenAddr,
		Handler: metricsMux,
	}

	go func() {
		logger.Info("prometheus metrics available", "listenAddr", metricsListenAddr, "endpoint", metricsEndpoint)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("metrics server failure", "error", err)
		}
	}()

	// Start the main HTTP server.
	go func() {
		logger.Info("RPC server available", "listenAddr", rpcListenAddr, "endpoint", rpcListenEndpoint)
		if err := rpcServer.ListenAndServe(); err != nil {
			logger.Fatal("RPC server failure", "error", err)
		}
	}()

	// Wait for shutdown signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")

	// Shutdown metrics server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shut down metrics server", "error", err)
	}

	// Shutdown RPC server
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rpcServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shut down RPC server", "error", err)
	}

	logger.Info("stopping blockchain listeners and workers")
	cancelBlockchain()

	// Close backbone resources
	if err := bb.Close(); err != nil {
		logger.Error("failed to close backbone resources", "error", err)
	}

	logger.Info("shutdown complete")
}

func runOperatorCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: clearnode operator <command>")
		fmt.Println("Commands:")
		fmt.Println("  address               Print the operator address")
		fmt.Println("  register-validators   Register signature validators on-chain")
		os.Exit(1)
	}

	switch args[0] {
	case "address":
		runOperatorAddress()
	case "register-validators":
		runRegisterValidators()
	default:
		fmt.Printf("Unknown operator command: %s\n", args[0])
		os.Exit(1)
	}
}

func runOperatorAddress() {
	bb := InitBackbone()
	defer bb.Close()

	fmt.Println(bb.StateSigner.PublicKey().Address().String())
	time.Sleep(5 * time.Second)
}

func runRegisterValidators() {
	bb := InitBackbone()
	defer bb.Close()
	logger := bb.Logger

	blockchains, err := bb.MemoryStore.GetBlockchains()
	if err != nil {
		logger.Fatal("failed to get blockchains from memory store", "error", err)
	}

	for _, b := range blockchains {
		if b.ChannelHubAddress == "" {
			continue
		}

		rpcURL, ok := bb.BlockchainRPCs[b.ID]
		if !ok {
			logger.Fatal("no RPC URL configured for blockchain", "blockchainID", b.ID)
		}

		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			logger.Fatal("failed to connect to EVM Node", "blockchainID", b.ID)
		}

		nodeAddress := bb.StateSigner.PublicKey().Address().String()
		clientOpts := []evm.ClientOption{
			evm.ClientBalanceCheck{RequireBalanceCheck: false},
			evm.ClientAllowanceCheck{RequireAllowanceCheck: false},
		}

		blockchainClient, err := evm.NewBlockchainClient(common.HexToAddress(b.ChannelHubAddress), client, bb.TxSigner, b.ID, nodeAddress, bb.MemoryStore, clientOpts...)
		if err != nil {
			logger.Fatal("failed to create EVM client", "blockchainID", b.ID)
		}

		sigValidators, err := bb.MemoryStore.GetChannelSigValidators(b.ID)
		if err != nil {
			logger.Fatal("failed to get channel signature validators from memory store", "error", err, "blockchainID", b.ID)
		}

		if err := ensureSigValidatorsRegistered(blockchainClient, sigValidators, false); err != nil {
			logger.Fatal("failed to register signature validators", "error", err, "blockchainID", b.ID)
		}

		logger.Info("signature validators registered successfully", "blockchainID", b.ID)
	}

	logger.Info("all signature validators registered")
}

func ensureSigValidatorsRegistered(client core.BlockchainClient, validators map[uint8]string, checkOnly bool) error {
	for id, addr := range validators {
		if err := client.EnsureSigValidatorRegistered(id, addr, checkOnly); err != nil {
			return err
		}
	}

	return nil
}

func runStoreMetricsExporter(
	ctx context.Context,
	fetchInterval time.Duration,
	store interface {
		GetChannelsCountByLabels() ([]database.ChannelCount, error)
		GetAppSessionsCountByLabels() ([]database.AppSessionCount, error)
		GetTotalValueLocked() ([]database.TotalValueLocked, error)
		CountActiveUsers(window time.Duration) ([]database.ActiveCountByLabel, error)
		CountActiveAppSessions(window time.Duration) ([]database.ActiveCountByLabel, error)
	},
	metricExported metrics.StoreMetricExporter, logger log.Logger) {
	logger = logger.WithName("store-metrics")
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			timeSpans := []struct {
				label    string
				duration time.Duration
			}{
				{"day", 24 * time.Hour},
				{"week", 7 * 24 * time.Hour},
				{"month", 30 * 24 * time.Hour},
			}

			channelCounts, err := store.GetChannelsCountByLabels()
			if err != nil {
				logger.Error("failed to get channel counts by labels", "error", err)
			} else {
				for _, c := range channelCounts {
					metricExported.SetChannels(c.Asset, c.Status, c.Count)
				}
			}

			appSessionCounts, err := store.GetAppSessionsCountByLabels()
			if err != nil {
				logger.Error("failed to get app sessions counts by labels", "error", err)
			} else {
				for _, c := range appSessionCounts {
					metricExported.SetAppSessions(c.Application, c.Status, c.Count)
				}
			}

			tvlCounts, err := store.GetTotalValueLocked()
			if err != nil {
				logger.Error("failed to get total value locked", "error", err)
			} else {
				for _, c := range tvlCounts {
					metricExported.SetTotalValueLocked(c.Domain, c.Asset, c.Value.InexactFloat64())
				}
			}

			for _, tw := range timeSpans {
				if counts, err := store.CountActiveUsers(tw.duration); err != nil {
					logger.Error("failed to count active users", "timeframe", tw.label, "error", err)
				} else {
					for _, c := range counts {
						metricExported.SetActiveUsers(c.Label, tw.label, c.Count)
					}
				}
			}

			for _, tw := range timeSpans {
				if counts, err := store.CountActiveAppSessions(tw.duration); err != nil {
					logger.Error("failed to count active app sessions", "timeframe", tw.label, "error", err)
				} else {
					for _, c := range counts {
						metricExported.SetActiveAppSessions(c.Label, tw.label, c.Count)
					}
				}
			}

		case <-ctx.Done():
			logger.Info("stopping store metrics exporter")
			return
		}
	}
}
