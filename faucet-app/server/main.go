package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"faucet-server/internal/nitronode"
	"faucet-server/internal/config"
	"faucet-server/internal/logger"
	"faucet-server/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Nitrolite Faucet Server")
	logger.Infof("Configuration loaded: Server port=%s, Nitronode URL=%s",
		cfg.ServerPort, cfg.NitronodeURL)

	client, err := nitronode.NewClient(cfg.OwnerPrivateKey, cfg.NitronodeURL, cfg.TokenSymbol, cfg.StandardTipAmountDecimal, cfg.MinTransferCount)
	if err != nil {
		logger.Fatalf("Failed to create Nitronode client: %v", err)
	}

	if err := client.EnsureOperational(); err != nil {
		logger.Fatalf("Operational check failed: %v", err)
	}

	logger.Infof("Faucet owner address: %s", client.GetOwnerAddress())
	logger.Info("Successfully connected to Nitronode")

	httpServer := server.NewServer(cfg, client)

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	logger.Info("Faucet server is ready to serve requests")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if err := client.Close(); err != nil {
		logger.Errorf("Error closing Nitronode connection: %v", err)
	}

	logger.Info("Server shutdown complete")
}
