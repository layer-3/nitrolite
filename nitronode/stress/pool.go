package stress

import (
	"fmt"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// CreateClientPool opens up to n WebSocket connections to the nitronode.
// It tolerates individual connection failures and returns whatever connections
// succeeded. Returns an error only if zero connections could be established.
func CreateClientPool(wsURL, privateKey string, n int) ([]*sdk.Client, error) {
	ethMsgSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}
	stateSigner, err := core.NewChannelDefaultSigner(ethMsgSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel signer: %w", err)
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	opts := []sdk.Option{
		sdk.WithErrorHandler(func(_ error) {}),
	}

	const maxRetries = 3

	clients := make([]*sdk.Client, 0, n)
	var lastErr error
	totalRetries := 0

	for i := 0; i < n; i++ {
		var client *sdk.Client
		var connectErr error

		for attempt := range maxRetries + 1 {
			client, connectErr = sdk.NewClient(wsURL, stateSigner, txSigner, opts...)
			if connectErr == nil {
				break
			}
			lastErr = connectErr
			if attempt < maxRetries {
				backoff := min(time.Duration(500*(1<<attempt))*time.Millisecond, 10*time.Second)
				fmt.Printf("\r  Connections: %d/%d (retrying %d/%d, backoff %v)          ",
					len(clients), n, attempt+1, maxRetries, backoff)
				time.Sleep(backoff)
				totalRetries++
			}
		}

		if connectErr != nil {
			fmt.Printf("\n  Connection %d/%d gave up after %d attempts: %s\n",
				i+1, n, maxRetries+1, connectErr.Error())
			continue
		}

		clients = append(clients, client)
		fmt.Printf("\r  Connections: %d/%d  ", len(clients), n)
		// Pace connection attempts to avoid overwhelming the server.
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()

	if totalRetries > 0 {
		fmt.Printf("  (%d retries during connection setup)\n", totalRetries)
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("failed to open any connections: %w", lastErr)
	}

	if len(clients) < n {
		fmt.Printf("WARNING: Only %d/%d connections established\n", len(clients), n)
	}

	return clients, nil
}

// CloseClientPool closes all clients in the pool.
func CloseClientPool(clients []*sdk.Client) {
	for _, c := range clients {
		c.Close()
	}
}
