package nitronode

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
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

	reconnectMu sync.Mutex // serialises reconnect attempts; not held during I/O
	tokenMu     sync.Mutex // serialises GetAssets; prevents N goroutines racing to validate

	logger           log.Logger
	ownerAddress     string
	tokenSymbol      string
	tipAmount        decimal.Decimal
	minTransferCount int
	tokenSupported   bool // cached per connection; reset in reconnect
}

// NewClient creates a Client that wraps the Nitrolite SDK for faucet operations.
// privateKeyHex drives both message signing and tx signing. nitronodeURL is the
// WebSocket endpoint. The client is immediately connected and ready to use.
func NewClient(logger log.Logger, privateKeyHex, nitronodeURL, tokenSymbol string, tipAmount decimal.Decimal, minTransferCount int) (*Client, error) {
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
				logger.Error("nitronode connection error", "error", err)
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
		logger:           logger,
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
	// Fast path: check WaitCh under read lock.
	c.mu.RLock()
	waitCh := c.sdkClient.WaitCh()
	c.mu.RUnlock()

	select {
	case <-waitCh:
		// Connection lost; fall through.
	default:
		return nil
	}

	// Serialise reconnect attempts; only one goroutine does the work at a time.
	c.reconnectMu.Lock()
	defer c.reconnectMu.Unlock()

	// Double-check under read lock — another goroutine may have reconnected while
	// we waited for reconnectMu.
	c.mu.RLock()
	waitCh = c.sdkClient.WaitCh()
	c.mu.RUnlock()

	select {
	case <-waitCh:
		// Still disconnected; proceed.
	default:
		return nil
	}

	return c.reconnect()
}

// reconnect retries SDK connection with exponential backoff.
// reconnectMu must be held by the caller; c.mu is NOT held here so I/O
// (dial + ping) does not stall readers or writers.
func (c *Client) reconnect() error {
	delay := reconnectInitDelay
	var lastErr error

	for attempt := 1; attempt <= reconnectAttempts; attempt++ {
		c.logger.Info("reconnecting to nitronode", "attempt", attempt, "max", reconnectAttempts)

		newClient, err := c.newSDKClient()
		if err != nil {
			lastErr = err
			c.logger.Warn("reconnect attempt failed", "attempt", attempt, "max", reconnectAttempts, "error", err)
		} else {
			// Ping without holding any lock.
			ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
			pingErr := newClient.Ping(ctx)
			cancel()

			if pingErr == nil {
				// Swap under write lock — fast, no I/O.
				c.mu.Lock()
				old := c.sdkClient
				c.sdkClient = newClient
				c.tokenSupported = false // re-validate on new connection
				c.mu.Unlock()

				// Close old client outside the lock.
				if err := old.Close(); err != nil {
					c.logger.Error("error closing stale nitronode client", "error", err)
				}
				c.logger.Info("reconnected to nitronode", "attempt", attempt)
				return nil
			}

			lastErr = fmt.Errorf("ping failed: %w", pingErr)
			c.logger.Warn("reconnect ping failed", "attempt", attempt, "max", reconnectAttempts, "error", pingErr)
			if err := newClient.Close(); err != nil {
				c.logger.Warn("error closing failed reconnect client", "error", err)
			}
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
	// Fast path: already confirmed for this connection.
	c.mu.RLock()
	if c.tokenSupported {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// Serialise the GetAssets call so only one goroutine fetches at a time.
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring tokenMu.
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
			c.logger.Debug("token supported by nitronode", "token", tokenSymbol)
			c.mu.Lock()
			if c.sdkClient == cl { // guard against reconnect between fetch and write
				c.tokenSupported = true
			}
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
			c.logger.Info("sufficient balance",
				"token", tokenSymbol,
				"balance", balance.Balance.String(),
			)
			if balance.Enforced.IsPositive() && balance.Enforced.LessThan(balance.Balance) {
				c.logger.Warn("enforced balance below channel balance; consider checkpointing",
					"token", tokenSymbol,
					"enforced", balance.Enforced.String(),
					"channel", balance.Balance.String(),
				)
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
// Uses write lock to serialise with reconnect and prevent closing a freshly installed client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sdkClient.Close()
}
