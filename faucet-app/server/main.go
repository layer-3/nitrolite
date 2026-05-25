package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/layer-3/nitrolite/faucet-app/server/internal/config"
	"github.com/layer-3/nitrolite/faucet-app/server/internal/nitronode"
	"github.com/layer-3/nitrolite/faucet-app/server/internal/server"
	"github.com/layer-3/nitrolite/pkg/log"
)

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "[invalid URL]"
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := log.NewZapLogger(cfg.Log).WithName("faucet")

	logger.Info("starting nitrolite faucet server")
	logger.Info("configuration loaded",
		"server_port", cfg.ServerPort,
		"nitronode_url", redactURL(cfg.NitronodeURL),
	)

	client, err := nitronode.NewClient(logger.WithName("nitronode"), cfg.OwnerPrivateKey, cfg.NitronodeURL, cfg.TokenSymbol, cfg.StandardTipAmountDecimal, cfg.MinTransferCount)
	if err != nil {
		logger.Fatal("failed to create nitronode client", "error", err)
	}

	if err := client.EnsureOperational(); err != nil {
		logger.Fatal("operational check failed", "error", err)
	}

	logger.Info("connected to nitronode", "owner_address", client.GetOwnerAddress())

	httpServer := server.NewServer(logger.WithName("http"), cfg, client)

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal("failed to start HTTP server", "error", err)
		}
	}()

	logger.Info("faucet server ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")

	if err := client.Close(); err != nil {
		logger.Error("error closing nitronode connection", "error", err)
	}

	logger.Info("shutdown complete")
}
