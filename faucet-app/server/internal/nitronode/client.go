package nitronode

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

const (
	sdkCallTimeout     = 30 * time.Second
	reconnectInitDelay = 300 * time.Millisecond
	reconnectMaxDelay  = 2 * time.Second
	reconnectAttempts  = 3
	pingTimeout        = 3 * time.Second
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

	ownerAddress     string
	tokenSymbol      string
	tipAmount        decimal.Decimal
	minTransferCount int
	tokenSupported   bool // cached after first successful GetAssets; reset on reconnect
}

func NewClient(privateKeyHex, nitronodeURL, tokenSymbol string, tipAmount decimal.Decimal, minTransferCount int) (*Client, error) {
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
		cl, err := sdk.NewClient(nitronodeURL, stateSigner, txSigner,
			sdk.WithApplicationID("faucet"),
			sdk.WithErrorHandler(func(err error) {
				logger.Errorf("Nitronode connection error: %v", err)
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Nitronode: %w", err)
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
		ownerAddress:     sdkClient.GetUserAddress(), // immutable: all factory clients share the same signer
		tokenSymbol:      tokenSymbol,
		tipAmount:        tipAmount,
		minTransferCount: minTransferCount,
	}, nil
}

// GetOwnerAddress returns the faucet owner's Ethereum address.
// Derived from the immutable raw signer; cached at construction so no lock is needed.
func (c *Client) GetOwnerAddress() string {
	return c.ownerAddress
}

// EnsureConnected checks the connection and reconnects with exponential backoff if necessary.
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
	defer c.mu.Unlock()

	select {
	case <-c.sdkClient.WaitCh():
		// Still disconnected; proceed with reconnect.
	default:
		return nil // Another goroutine already reconnected while we waited for the lock.
	}

	return c.reconnectLocked()
}

// reconnectLocked tries to establish a new SDK connection with exponential backoff.
// Must be called with c.mu write lock held.
func (c *Client) reconnectLocked() error {
	delay := reconnectInitDelay
	var lastErr error

	for attempt := 1; attempt <= reconnectAttempts; attempt++ {
		logger.Infof("Reconnecting to Nitronode (attempt %d/%d)...", attempt, reconnectAttempts)

		newClient, err := c.newSDKClient()
		if err != nil {
			lastErr = err
			logger.Warnf("Reconnect attempt %d/%d failed: %v", attempt, reconnectAttempts, err)
		} else {
			// Ping to confirm the new connection is alive before accepting it.
			ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
			pingErr := newClient.Ping(ctx)
			cancel()

			if pingErr == nil {
				oldClient := c.sdkClient
				c.sdkClient = newClient
				c.tokenSupported = false // re-validate token on the new connection

				// doClose is a non-blocking channel close, safe to call under the lock.
				if err := oldClient.Close(); err != nil {
					logger.Errorf("Error closing stale Nitronode client: %v", err)
				}
				logger.Infof("Successfully reconnected to Nitronode on attempt %d", attempt)
				return nil
			}

			lastErr = fmt.Errorf("ping failed: %w", pingErr)
			logger.Warnf("Reconnect attempt %d/%d ping failed: %v", attempt, reconnectAttempts, pingErr)
			_ = newClient.Close()
		}

		if attempt < reconnectAttempts {
			time.Sleep(delay)
			delay = min(delay*2, reconnectMaxDelay)
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %w", reconnectAttempts, lastErr)
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

func (c *Client) validateTokenSupport(tokenSymbol string) error {
	// Fast path: token already confirmed on this connection.
	c.mu.RLock()
	cached := c.tokenSupported
	cl := c.sdkClient
	c.mu.RUnlock()

	if cached {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sdkCallTimeout)
	defer cancel()

	assets, err := cl.GetAssets(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch supported assets: %w", err)
	}

	for _, asset := range assets {
		if strings.EqualFold(asset.Symbol, tokenSymbol) {
			logger.Debugf("Token '%s' is supported by Nitronode", tokenSymbol)
			c.mu.Lock()
			c.tokenSupported = true
			c.mu.Unlock()
			return nil
		}
	}

	return fmt.Errorf("token '%s' is not supported by Nitronode", tokenSymbol)
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
			if balance.Enforced.IsPositive() && balance.Enforced.LessThan(balance.Balance) {
				logger.Warnf("⚠ %s enforced balance (%s) is below channel balance (%s); consider checkpointing",
					tokenSymbol, balance.Enforced.String(), balance.Balance.String())
			}
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

	if result.Amount == "" {
		result.Amount = amount.String()
	}
	if result.Asset == "" {
		result.Asset = asset
	}

	return result, nil
}

// Close shuts down the Nitronode connection.
// Uses write lock to serialize with reconnect and prevent closing a freshly installed client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sdkClient.Close()
}
