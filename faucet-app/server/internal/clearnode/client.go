package clearnode

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"

	"faucet-server/internal/logger"
)

// TransferResult holds the result of a token transfer.
type TransferResult struct {
	TxID   string
	Amount string
	Asset  string
}

// Client wraps the Nitrolite SDK client for faucet operations.
type Client struct {
	mu           sync.RWMutex
	sdkClient    *sdk.Client
	newSDKClient func() (*sdk.Client, error) // captures parsed signers; no raw key hex stored

	tokenSymbol      string
	tipAmount        decimal.Decimal
	minTransferCount int
}

func NewClient(privateKeyHex, clearnodeURL, tokenSymbol string, tipAmount decimal.Decimal, minTransferCount int) (*Client, error) {
	// Parse signers once — raw key hex is used here and not retained on the struct.
	msgSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to create message signer: %w", err)
	}

	stateSigner, err := core.NewChannelDefaultSigner(msgSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	// factory captures already-parsed signers so reconnects don't need the raw key.
	factory := func() (*sdk.Client, error) {
		cl, err := sdk.NewClient(clearnodeURL, stateSigner, txSigner)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Clearnode: %w", err)
		}
		return cl, nil
	}

	sdkClient, err := factory()
	if err != nil {
		return nil, err
	}

	return &Client{
		sdkClient:        sdkClient,
		newSDKClient:     factory,
		tokenSymbol:      tokenSymbol,
		tipAmount:        tipAmount,
		minTransferCount: minTransferCount,
	}, nil
}

// GetOwnerAddress returns the faucet owner's Ethereum address.
func (c *Client) GetOwnerAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sdkClient.GetUserAddress()
}

// EnsureConnected checks the connection and reconnects if necessary.
func (c *Client) EnsureConnected() error {
	// Fast path: read WaitCh under read lock.
	c.mu.RLock()
	waitCh := c.sdkClient.WaitCh()
	c.mu.RUnlock()

	select {
	case <-waitCh:
		// Connection lost; fall through to reconnect.
	default:
		return nil
	}

	// Slow path: write lock with double-check to prevent thundering-herd reconnects.
	c.mu.Lock()

	select {
	case <-c.sdkClient.WaitCh():
		// Still disconnected; reconnect now.
	default:
		c.mu.Unlock()
		return nil // Another goroutine already reconnected while we waited for the lock.
	}

	logger.Info("Connection lost, reconnecting to Clearnode...")
	newClient, err := c.newSDKClient()
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to reconnect: %w", err)
	}
	oldClient := c.sdkClient
	c.sdkClient = newClient
	c.mu.Unlock() // Release before closing old client to avoid holding lock during I/O.

	if err := oldClient.Close(); err != nil {
		logger.Errorf("Error closing stale Clearnode client: %v", err)
	}
	logger.Info("Successfully reconnected to Clearnode")
	return nil
}

// EnsureOperational validates token support and sufficient balance.
func (c *Client) EnsureOperational() error {
	if err := c.validateTokenSupport(c.tokenSymbol); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if err := c.validateFaucetBalance(c.tokenSymbol, c.tipAmount, c.minTransferCount); err != nil {
		return fmt.Errorf("balance check failed: %w", err)
	}

	return nil
}

const sdkCallTimeout = 30 * time.Second

func (c *Client) validateTokenSupport(tokenSymbol string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sdkCallTimeout)
	defer cancel()

	c.mu.RLock()
	cl := c.sdkClient
	c.mu.RUnlock()

	assets, err := cl.GetAssets(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch supported assets: %w", err)
	}

	for _, asset := range assets {
		if strings.EqualFold(asset.Symbol, tokenSymbol) {
			logger.Debugf("Token '%s' is supported by Clearnode", tokenSymbol)
			return nil
		}
	}

	return fmt.Errorf("token '%s' is not supported by Clearnode", tokenSymbol)
}

func (c *Client) validateFaucetBalance(tokenSymbol string, tipAmount decimal.Decimal, minTransferCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), sdkCallTimeout)
	defer cancel()

	c.mu.RLock()
	cl := c.sdkClient
	c.mu.RUnlock()

	ownerAddress := cl.GetUserAddress()

	balances, err := cl.GetBalances(ctx, ownerAddress)
	if err != nil {
		return fmt.Errorf("failed to fetch faucet balance: %w", err)
	}

	minRequired := tipAmount.Mul(decimal.NewFromInt(int64(minTransferCount)))

	for _, balance := range balances {
		if strings.EqualFold(balance.Asset, tokenSymbol) {
			if balance.Balance.LessThan(minRequired) {
				return fmt.Errorf("insufficient %s balance: %s (required: %s for %d transfers)",
					tokenSymbol, balance.Balance.String(), minRequired.String(), minTransferCount)
			}
			logger.Infof("✓ Sufficient %s balance: %s", tokenSymbol, balance.Balance.String())
			return nil
		}
	}

	return fmt.Errorf("insufficient %s balance: 0 (required: %s for %d transfers)",
		tokenSymbol, minRequired.String(), minTransferCount)
}

// Transfer sends tokens to the destination address.
func (c *Client) Transfer(destination, asset string, amount decimal.Decimal) (*TransferResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sdkCallTimeout)
	defer cancel()

	c.mu.RLock()
	cl := c.sdkClient
	c.mu.RUnlock()

	state, err := cl.Transfer(ctx, destination, asset, amount)
	if err != nil {
		return nil, fmt.Errorf("transfer failed: %w", err)
	}

	result := &TransferResult{
		TxID:   state.Transition.TxID,
		Amount: state.Transition.Amount.String(),
		Asset:  state.Asset,
	}

	if result.Amount == "" || result.Amount == "0" {
		result.Amount = amount.String()
	}
	if result.Asset == "" {
		result.Asset = asset
	}

	return result, nil
}

// Close shuts down the Clearnode connection.
func (c *Client) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sdkClient.Close()
}
