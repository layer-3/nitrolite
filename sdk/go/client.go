package sdk

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/layer-3/nitrolite/pkg/blockchain/evm"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// Client provides a unified interface for interacting with Clearnode.
// It combines state-building operations (Deposit, Withdraw, Transfer) with a single
// Checkpoint method for blockchain settlement, plus low-level RPC access for advanced use cases.
//
// The two-step pattern for blockchain operations:
//  1. Build and co-sign the state off-chain (Deposit, Withdraw, CloseHomeChannel, etc.)
//  2. Settle on-chain via Checkpoint
//
// High-level example:
//
//	stateSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
//	txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)
//	client, _ := sdk.NewClient(
//	    "wss://clearnode.example.com/ws",
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
//	defer client.Close()
//
//	// Deposit: build state, then settle on-chain
//	state, _ := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	txHash, _ := client.Checkpoint(ctx, "usdc")
//
//	// Transfer: off-chain only, no Checkpoint needed for existing channels
//	state, _ = client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//
//	// Low-level operations
//	config, _ := client.GetConfig(ctx)
//	balances, _ := client.GetBalances(ctx, walletAddress)
type Client struct {
	rpcClient         *rpc.Client
	config            Config
	exitCh            chan struct{}
	closeOnce         sync.Once
	blockchainClients map[uint64]core.Client
	homeBlockchains   map[string]uint64
	stateSigner       core.ChannelSigner
	rawSigner         sign.Signer
	assetStore        *clientAssetStore
}

// NewClient creates a new Clearnode client with both high-level and low-level methods.
// This is the recommended constructor for most use cases.
//
// Parameters:
//   - wsURL: WebSocket URL of the Clearnode server (e.g., "wss://clearnode.example.com/ws")
//   - stateSigner: core.ChannelSigner for signing channel states (use sign.NewEthereumMsgSigner)
//   - txSigner: sign.Signer for signing blockchain transactions (use sign.NewEthereumRawSigner)
//   - opts: Optional configuration (WithBlockchainRPC, WithHandshakeTimeout, etc.)
//
// Returns:
//   - Configured Client ready for operations
//   - Error if connection or initialization fails
//
// Example:
//
//	stateSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
//	txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)
//	client, err := sdk.NewClient(
//	    "wss://clearnode.example.com/ws",
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
func NewClient(wsURL string, stateSigner core.ChannelSigner, rawSigner sign.Signer, opts ...Option) (*Client, error) {
	// Build config starting with defaults
	config := DefaultConfig
	config.URL = wsURL

	// Apply user options
	for _, opt := range opts {
		opt(&config)
	}

	// Create WebSocket dialer with configuration
	dialerConfig := rpc.DefaultWebsocketDialerConfig
	dialerConfig.HandshakeTimeout = config.HandshakeTimeout
	dialerConfig.PingTimeout = config.PingTimeout

	dialer := rpc.NewWebsocketDialer(dialerConfig)
	rpcClient := rpc.NewClient(dialer)

	// Create client instance
	client := &Client{
		rpcClient:         rpcClient,
		config:            config,
		exitCh:            make(chan struct{}),
		blockchainClients: make(map[uint64]core.Client),
		homeBlockchains:   make(map[string]uint64),
		stateSigner:       stateSigner,
		rawSigner:         rawSigner,
	}

	// Create asset store
	client.assetStore = newClientAssetStore(client)

	// Error handler wrapper
	handleError := func(err error) {
		if config.ErrorHandler != nil {
			config.ErrorHandler(err)
		}
		client.doClose()
	}

	// Establish connection
	err := rpcClient.Start(context.Background(), wsURL, handleError)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clearnode: %w", err)
	}

	return client, nil
}

// SetHomeBlockchain configures the primary blockchain network for a specific asset.
// This is required for operations like Transfer which may trigger channel creation
// but do not accept a blockchain ID as a parameter.
//
// Validation:
//   - Checks if the asset is actually supported on the specified blockchain.
//   - Verifies that a home blockchain hasn't already been set for this asset.
//
// Constraints:
//   - This mapping is immutable once set for the client instance.
//   - To move an asset to a different blockchain, use the Migrate() method instead.
//
// Parameters:
//   - asset: The asset symbol (e.g., "usdc")
//   - blockchainId: The chain ID to associate with the asset (e.g., 80002)
//
// Example:
//
//	// Set USDC to settle on Polygon Amoy
//	if err := client.SetHomeBlockchain("usdc", 80002); err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) SetHomeBlockchain(asset string, blockchainId uint64) error {
	blockchainID, homeBlockchainIsSet := c.homeBlockchains[asset]
	if homeBlockchainIsSet {
		return fmt.Errorf("home blockchain is already set for asset %s to %d, please use Migrate() if you want to change home blockchain", asset, blockchainID)
	}
	ok, err := c.assetStore.AssetExistsOnBlockchain(blockchainId, asset)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("asset %s not supported on blockchain %d", asset, blockchainId)
	}
	c.homeBlockchains[asset] = blockchainId
	return nil
}

// ============================================================================
// Connection & Lifecycle Methods
// ============================================================================

// Close cleanly shuts down the client connection.
// It's recommended to defer this call after creating the client.
//
// Example:
//
//	client, err := NewClient(...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func (c *Client) Close() error {
	c.doClose()
	return nil
}

// doClose closes exitCh exactly once, safe for concurrent callers.
func (c *Client) doClose() {
	c.closeOnce.Do(func() {
		close(c.exitCh)
	})
}

// WaitCh returns a channel that closes when the connection is lost or closed.
// This is useful for monitoring connection health in long-running applications.
//
// Example:
//
//	go func() {
//	    <-client.WaitCh()
//	    log.Println("Connection closed")
//	}()
func (c *Client) WaitCh() <-chan struct{} {
	return c.exitCh
}

// ============================================================================
// Shared Helper Methods
// ============================================================================

// SignState signs a channel state by packing it, hashing it, and signing the hash.
// Returns the signature as a hex-encoded string (with 0x prefix).
//
// This is a low-level method exposed for advanced users who want to manually
// construct and sign states. Most users should use the high-level methods like
// Transfer, Deposit, and Withdraw instead.
func (c *Client) SignState(state *core.State) (string, error) {
	if state == nil {
		return "", fmt.Errorf("state cannot be nil")
	}

	// Pack the state into ABI-encoded bytes
	packedState, err := core.PackState(*state, c.assetStore)
	if err != nil {
		return "", fmt.Errorf("failed to pack state: %w", err)
	}

	// Sign the hash
	signature, err := c.stateSigner.Sign(packedState)
	if err != nil {
		return "", fmt.Errorf("failed to sign state hash: %w", err)
	}

	// Return hex-encoded signature with 0x prefix
	return hexutil.Encode(signature), nil
}

// GetUserAddress returns the Ethereum address associated with the signer.
// This is useful for identifying the current user's wallet address.
func (c *Client) GetUserAddress() string {
	return c.rawSigner.PublicKey().Address().String()
}

// signAndSubmitState is a helper that signs a state and submits it to the node.
// It returns the node's signature.
func (c *Client) signAndSubmitState(ctx context.Context, state *core.State) (string, error) {
	// Sign state
	sig, err := c.SignState(state)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}
	state.UserSig = &sig

	// Submit to node
	nodeSig, err := c.submitState(ctx, *state)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}

	// Update state with node signature
	state.NodeSig = &nodeSig

	return nodeSig, nil
}

// ============================================================================
// Blockchain Configuration
// ============================================================================

// WithBlockchainRPC returns an Option that configures a blockchain RPC client for a specific chain.
// This is required for the Checkpoint method which settles states on-chain.
//
// Parameters:
//   - chainID: The blockchain network ID (e.g., 80002 for Polygon Amoy testnet)
//   - rpcURL: The RPC endpoint URL (e.g., "https://polygon-amoy.alchemy.com/v2/KEY")
//
// Example:
//
//	client, err := sdk.NewClient(
//	    wsURL,
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	    sdk.WithBlockchainRPC(84532, "https://base-sepolia.alchemy.com/v2/KEY"),
//	)
func WithBlockchainRPC(chainID uint64, rpcURL string) Option {
	return func(c *Config) {
		// Store blockchain RPC config for later initialization
		if c.BlockchainRPCs == nil {
			c.BlockchainRPCs = make(map[uint64]string)
		}
		c.BlockchainRPCs[chainID] = rpcURL
	}
}

// initializeBlockchainClient initializes a blockchain client for a specific chain.
// This is called lazily when a blockchain operation is needed.
func (c *Client) initializeBlockchainClient(ctx context.Context, chainID uint64) error {
	// Check if already initialized
	if _, exists := c.blockchainClients[chainID]; exists {
		return nil
	}

	// Get RPC URL from config
	rpcURL, exists := c.config.BlockchainRPCs[chainID]
	if !exists {
		return fmt.Errorf("blockchain RPC not configured for chain %d (use WithBlockchainRPC)", chainID)
	}

	// Get contract address for this blockchain
	contractAddress, err := c.getContractAddress(ctx, chainID)
	if err != nil {
		return err
	}

	// Get node address
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return err
	}

	// Connect to blockchain
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}

	// Create blockchain client using the user's signer and node address
	evmClient, err := evm.NewClient(
		common.HexToAddress(contractAddress),
		ethClient,
		c.rawSigner,
		chainID,
		nodeAddress,
		c.assetStore,
	)

	if err != nil {
		return fmt.Errorf("failed to create blockchain client: %w", err)
	}

	c.blockchainClients[chainID] = evmClient
	return nil
}

// generateNonce generates a random 8-byte nonce for channel creation.
func generateNonce() uint64 {
	return uint64(time.Now().UnixNano())
}

// getTokenAddress looks up the token address for an asset on a specific blockchain.
func (c *Client) getTokenAddress(ctx context.Context, blockchainID uint64, asset string) (string, error) {
	assets, err := c.GetAssets(ctx, &blockchainID)
	if err != nil {
		return "", fmt.Errorf("failed to get assets: %w", err)
	}

	for _, a := range assets {
		if strings.EqualFold(a.Symbol, asset) {
			// Find token for this blockchain
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return token.Address, nil
				}
			}
			return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
		}
	}

	return "", fmt.Errorf("asset %s not found", asset)
}

// getContractAddress retrieves the channel hub contract address for a specific blockchain from node config.
func (c *Client) getContractAddress(ctx context.Context, blockchainID uint64) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}

	for _, bc := range nodeConfig.Blockchains {
		if bc.ID == blockchainID {
			return bc.ChannelHubAddress, nil
		}
	}

	return "", fmt.Errorf("blockchain %d not found in node config", blockchainID)
}

// getNodeAddress retrieves the node's Ethereum address from the node config.
func (c *Client) getNodeAddress(ctx context.Context) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}
	return nodeConfig.NodeAddress, nil
}

// getSupportedSigValidatorsBitmap fetches the node config and builds a hex bitmap
// from the supported signature validators. This bitmap is used in ChannelDefinition.
func (c *Client) getSupportedSigValidatorsBitmap(ctx context.Context) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}
	return core.BuildSigValidatorsBitmap(nodeConfig.SupportedSigValidators), nil
}
