package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/erc7824/nitrolite/clearnode/api"
	"github.com/erc7824/nitrolite/clearnode/event_handlers"
	"github.com/erc7824/nitrolite/clearnode/metrics"
	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/clearnode/stress"
	"github.com/erc7824/nitrolite/pkg/blockchain/evm"
	"github.com/erc7824/nitrolite/pkg/log"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "stress-test" {
		os.Exit(stress.Run(os.Args[2:]))
	}

	bb := InitBackbone()
	logger := bb.Logger
	ctx := context.Background()

	vl := bb.ValidationLimits
	api.NewRPCRouter(bb.NodeVersion, bb.ChannelMinChallengeDuration,
		bb.RpcNode, bb.StateSigner, bb.DbStore, bb.MemoryStore, bb.RuntimeMetrics, bb.Logger,
		vl.MaxParticipants, vl.MaxSessionDataLen, vl.MaxAppMetadataLen, vl.MaxSignedUpdates, vl.MaxSessionKeyIDs)

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
		reactor := evm.NewReactor(b.ID, eventHandlerService, bb.DbStore.StoreContractEvent)
		reactor.SetOnEventProcessed(bb.RuntimeMetrics.IncBlockchainEvent)
		l := evm.NewListener(common.HexToAddress(b.ChannelHubAddress), client, b.ID, b.BlockStep, logger, reactor.HandleEvent, bb.DbStore.GetLatestEvent)
		l.Listen(ctx, func(err error) {
			if err != nil {
				logger.Fatal("blockchain listener stopped", "error", err, "blockchainID", b.ID)
			}
		})

		// For the node itself, the node address is the signer's address
		nodeAddress := bb.StateSigner.PublicKey().Address().String()

		clientOpts := []evm.ClientOption{
			evm.ClientBalanceCheck{RequireBalanceCheck: false},
			evm.ClientAllowanceCheck{RequireAllowanceCheck: false},
		}

		blockchainClient, err := evm.NewClient(common.HexToAddress(b.ChannelHubAddress), client, bb.TxSigner, b.ID, nodeAddress, bb.MemoryStore, clientOpts...)
		if err != nil {
			logger.Fatal("failed to create EVM client")
		}

		worker := NewBlockchainWorker(b.ID, blockchainClient, bb.DbStore, logger, bb.RuntimeMetrics)
		worker.Start(ctx, func(err error) {
			if err != nil {
				logger.Fatal("blockchain worker stopped", "error", err, "blockchainID", b.ID)
			}
		})
	}

	go runStoreMetricsExporter(ctx, 10*time.Second, bb.DbStore, bb.StoreMetrics, logger)

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

	// Close backbone resources
	if err := bb.Close(); err != nil {
		logger.Error("failed to close backbone resources", "error", err)
	}

	// TODO: gracefully stop blockchain listeners and workers
	logger.Info("shutdown complete")
}

func runStoreMetricsExporter(
	ctx context.Context,
	fetchInterval time.Duration,
	store interface {
		CountAppSessionsByStatus() ([]database.AppSessionCount, error)
		CountChannelsByStatus() ([]database.ChannelCount, error)
	},
	metricExported metrics.StoreMetricExporter, logger log.Logger) {
	logger = logger.WithName("store-metrics")
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if counts, err := store.CountAppSessionsByStatus(); err != nil {
				logger.Error("failed to count app sessions", "error", err)
			} else {
				for _, c := range counts {
					metricExported.SetAppSessions(c.Application, c.Status, c.Count)
				}
			}

			if counts, err := store.CountChannelsByStatus(); err != nil {
				logger.Error("failed to count channels", "error", err)
			} else {
				for _, c := range counts {
					metricExported.SetChannels(c.Asset, c.Status, c.Count)
				}
			}
		case <-ctx.Done():
			logger.Info("stopping store metrics exporter")
			return
		}
	}
}
