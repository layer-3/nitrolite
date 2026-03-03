package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erc7824/nitrolite/clearnode/metrics"
	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/clearnode/store/memory"
	"github.com/erc7824/nitrolite/pkg/blockchain/evm"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/erc7824/nitrolite/pkg/sign/kms/gcp"
)

//go:embed config/migrations/*/*.sql
var embedMigrations embed.FS

var Version = "v1.0.0" // set at build time with -ldflags "-X main.Version=x.y.z"

type Backbone struct {
	NodeVersion                 string
	ChannelMinChallengeDuration uint32
	BlockchainRPCs              map[uint64]string
	ValidationLimits            ValidationLimits

	DbStore        database.DatabaseStore
	MemoryStore    memory.MemoryStore
	RpcNode        rpc.Node
	StateSigner    sign.Signer
	TxSigner       sign.Signer
	Logger         log.Logger
	RuntimeMetrics metrics.RuntimeMetricExporter
	StoreMetrics   metrics.StoreMetricExporter
	closers        []func() error
}

// Close releases resources held by the backbone (e.g., KMS client connections).
func (b *Backbone) Close() error {
	var firstErr error
	for _, fn := range b.closers {
		if err := fn(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type Config struct {
	Database                    database.DatabaseConfig
	ChannelMinChallengeDuration uint32           `yaml:"channel_min_challenge_duration" env:"CLEARNODE_CHANNEL_MIN_CHALLENGE_DURATION" env-default:"86400"` // 24 hours
	SignerType                  string           `yaml:"signer_type" env:"CLEARNODE_SIGNER_TYPE" env-default:"key"`                                         // "key" or "gcp-kms"
	SignerKey                   string           `yaml:"signer_key" env:"CLEARNODE_SIGNER_KEY"`                                                             // required when signer_type=key
	GCPKMSKeyName               string           `yaml:"gcp_kms_key_name" env:"CLEARNODE_GCP_KMS_KEY_NAME"`                                                 // required when signer_type=gcp-kms
	ValidationLimits            ValidationLimits `yaml:"validation_limits"`
}

// ValidationLimits defines configurable upper bounds for dynamic-length request fields.
type ValidationLimits struct {
	MaxParticipants   int `yaml:"max_participants" env:"CLEARNODE_MAX_PARTICIPANTS" env-default:"32"`
	MaxSessionDataLen int `yaml:"max_session_data_len" env:"CLEARNODE_MAX_SESSION_DATA_LEN" env-default:"1024"`
	MaxAppMetadataLen int `yaml:"max_app_metadata_len" env:"CLEARNODE_MAX_APP_METADATA_LEN" env-default:"1024"`
	MaxSessionKeyIDs  int `yaml:"max_session_key_ids" env:"CLEARNODE_MAX_SESSION_KEY_IDS" env-default:"256"`
	MaxSignedUpdates  int `yaml:"max_signed_updates" env:"CLEARNODE_MAX_SIGNED_UPDATES" env-default:"0"`
}

// InitBackbone initializes the backbone components of the application.
func InitBackbone() *Backbone {
	// ------------------------------------------------
	// Logger
	// ------------------------------------------------

	var loggerConf log.Config
	if err := cleanenv.ReadEnv(&loggerConf); err != nil {
		panic("failed to read logger config from env: " + err.Error())
	}
	logger := log.NewZapLogger(loggerConf)
	logger = logger.WithName("main")

	// ------------------------------------------------
	// (Preparation)
	// ------------------------------------------------

	configDirPath := os.Getenv("CLEARNODE_CONFIG_DIR_PATH")
	if configDirPath == "" {
		configDirPath = "."
	}

	configDotEnvPath := filepath.Join(configDirPath, ".env")
	logger.Info("loading .env file", "path", configDotEnvPath)
	if err := godotenv.Load(configDotEnvPath); err != nil {
		logger.Warn(".env file not found")
	}

	var conf Config
	if err := cleanenv.ReadEnv(&conf); err != nil {
		logger.Fatal("failed to read env", "err", err)
	}

	logger.Info("config loaded", "version", Version)

	// ------------------------------------------------
	// Database Store
	// ------------------------------------------------

	db, err := database.ConnectToDB(conf.Database, embedMigrations)
	if err != nil {
		logger.Fatal("failed to load database store", "error", err)
	}
	dbStore := database.NewDBStore(db)

	// ------------------------------------------------
	// Memory Store
	// ------------------------------------------------

	memoryStore, err := memory.NewMemoryStoreV1FromConfig(configDirPath)
	if err != nil {
		logger.Fatal("failed to load blockchains", "error", err)
	}

	// ------------------------------------------------
	// Signer
	// ------------------------------------------------

	var (
		stateSigner, txSigner sign.Signer
		signerErr             error
		closers               []func() error
	)

	switch conf.SignerType {
	case "key":
		if conf.SignerKey == "" {
			logger.Fatal("CLEARNODE_SIGNER_KEY is required when CLEARNODE_SIGNER_TYPE=key")
		}
		txSigner, signerErr = sign.NewEthereumRawSigner(conf.SignerKey)
		if signerErr != nil {
			logger.Fatal("failed to initialise tx signer", "error", signerErr)
		}
	case "gcp-kms":
		if conf.GCPKMSKeyName == "" {
			logger.Fatal("CLEARNODE_GCP_KMS_KEY_NAME is required when CLEARNODE_SIGNER_TYPE=gcp-kms")
		}
		kmsCtx, kmsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer kmsCancel()
		kmsSigner, kmsErr := gcp.NewSigner(kmsCtx, conf.GCPKMSKeyName)
		if kmsErr != nil {
			logger.Fatal("failed to initialise GCP KMS signer", "error", kmsErr)
		}
		closers = append(closers, kmsSigner.Close)
		txSigner = kmsSigner
	default:
		logger.Fatal("unsupported CLEARNODE_SIGNER_TYPE", "type", conf.SignerType)
	}
	stateSigner, signerErr = sign.NewEthereumMsgSignerFromRaw(txSigner)
	if signerErr != nil {
		logger.Fatal("failed to wrap KMS signer as state signer", "error", signerErr)
	}

	logger.Info("signer initialized", "type", conf.SignerType, "address", stateSigner.PublicKey().Address())

	// ------------------------------------------------
	// Metrics
	// ------------------------------------------------

	runtimeMetrics, err := metrics.NewRuntimeMetricExporter(prometheus.DefaultRegisterer)
	if err != nil {
		logger.Fatal("failed to initialize runtime metric exporter", "error", err)
	}
	storeMetrics, err := metrics.NewStoreMetricExporter(prometheus.DefaultRegisterer)
	if err != nil {
		logger.Fatal("failed to initialize store metric exporter", "error", err)
	}

	// ------------------------------------------------
	// RPC Node
	// ------------------------------------------------

	rpcNode, err := rpc.NewWebsocketNode(rpc.WebsocketNodeConfig{
		Logger:             logger,
		ObserveConnections: runtimeMetrics.SetRPCConnections,
	})
	if err != nil {
		logger.Fatal("failed to initialize RPC node", "error", err)
	}

	// ------------------------------------------------
	// Blockchain RPCs
	// ------------------------------------------------

	blockchains, err := memoryStore.GetBlockchains()
	if err != nil {
		logger.Fatal("failed to get blockchains", "error", err)
	}

	blockchainRPCs := make(map[uint64]string)
	for _, bc := range blockchains {
		envVarName := "CLEARNODE_BLOCKCHAIN_RPC_" + strings.ToUpper(bc.Name)
		rpcURL := os.Getenv(envVarName)
		if rpcURL == "" {
			logger.Fatal("blockchain RPC URL not set in env", "blockchainID", bc.ID, "env_var", envVarName)
		}

		// Test connection
		if err := checkChainId(rpcURL, bc.ID); err != nil {
			logger.Fatal("failed to verify blockchain RPC", "blockchainID", bc.ID, "error", err)
		}

		// Verify ChannelHub version
		channelHubAddress := common.HexToAddress(bc.ChannelHubAddress)
		if err := checkChannelHubVersion(rpcURL, channelHubAddress, core.ChannelHubVersion); err != nil {
			logger.Fatal("failed to verify ChannelHub version", "blockchainID", bc.ID, "address", bc.ChannelHubAddress, "error", err)
		}

		blockchainRPCs[bc.ID] = rpcURL
	}

	return &Backbone{
		NodeVersion:                 Version,
		ChannelMinChallengeDuration: conf.ChannelMinChallengeDuration,
		BlockchainRPCs:              blockchainRPCs,
		ValidationLimits:            conf.ValidationLimits,

		DbStore:        dbStore,
		MemoryStore:    memoryStore,
		RpcNode:        rpcNode,
		StateSigner:    stateSigner,
		TxSigner:       txSigner,
		Logger:         logger,
		RuntimeMetrics: runtimeMetrics,
		StoreMetrics:   storeMetrics,
		closers:        closers,
	}
}

// checkChainId connects to an RPC endpoint and verifies it returns the expected chain ID.
// This ensures the RPC URL points to the correct blockchain network.
// The function uses a 5-second timeout for the connection and chain ID query.
func checkChainId(blockchainRPC string, expectedChainID uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, blockchainRPC)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID from blockchain RPC: %w", err)
	}

	if chainID.Uint64() != expectedChainID {
		return fmt.Errorf("unexpected chain ID from blockchain RPC: got %d, want %d", chainID.Uint64(), expectedChainID)
	}

	return nil
}

// checkChannelHubVersion verifies that the ChannelHub contract at the given address
// has the expected VERSION constant value.
// The function uses a 5-second timeout for the connection and contract calls.
func checkChannelHubVersion(blockchainRPC string, channelHubAddress common.Address, expectedVersion uint8) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, blockchainRPC)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}
	defer client.Close()

	channelHub, err := evm.NewChannelHubCaller(channelHubAddress, client)
	if err != nil {
		return fmt.Errorf("failed to create ChannelHub caller: %w", err)
	}

	fetchedVersion, err := channelHub.VERSION(nil)
	if err != nil {
		return fmt.Errorf("failed to get ChannelHub version: %w", err)
	}

	if fetchedVersion != expectedVersion {
		return fmt.Errorf("configured and fetched ChannelHub version mismatch: got %d, want %d", fetchedVersion, expectedVersion)
	}

	return nil
}
