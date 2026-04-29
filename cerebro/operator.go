package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

type Operator struct {
	wsURL     string
	configDir string
	store     *Storage
	client    *sdk.Client
	exitCh    chan struct{}
}

func NewOperator(wsURL, configDir string, store *Storage) (*Operator, error) {
	op := &Operator{
		wsURL:     wsURL,
		configDir: configDir,
		store:     store,
		exitCh:    make(chan struct{}),
	}

	if err := op.connect(); err != nil {
		return nil, err
	}

	return op, nil
}

// buildStateSigner creates the appropriate ChannelSigner based on whether
// a session key is stored. Returns ChannelSessionKeySignerV1 if a session key
// is configured, otherwise returns ChannelDefaultSigner.
func (o *Operator) buildStateSigner(walletPrivateKey string) (core.ChannelSigner, error) {
	// Check if a session key is stored
	skPrivateKey, metadataHash, authSig, err := o.store.GetSessionKey()
	if err == nil && skPrivateKey != "" {
		// Use session key signer
		sessionMsgSigner, err := sign.NewEthereumMsgSigner(skPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create session key signer: %w", err)
		}
		signer, err := core.NewChannelSessionKeySignerV1(sessionMsgSigner, metadataHash, authSig)
		if err != nil {
			return nil, fmt.Errorf("failed to create session key channel signer: %w", err)
		}
		sessionRawSigner, _ := sign.NewEthereumRawSigner(skPrivateKey)
		fmt.Printf("INFO: Using session key for state signing: %s\n", sessionRawSigner.PublicKey().Address().String())
		return signer, nil
	}

	// Use default signer
	ethMsgSigner, err := sign.NewEthereumMsgSigner(walletPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}
	signer, err := core.NewChannelDefaultSigner(ethMsgSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel signer: %w", err)
	}
	return signer, nil
}

// connect creates the SDK client with the appropriate signer.
// If no wallet is configured, a new one is automatically generated.
func (o *Operator) connect() error {
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		// Auto-generate a new wallet for first-time users
		privateKey, err = generatePrivateKey()
		if err != nil {
			return fmt.Errorf("failed to generate wallet: %w", err)
		}
		if err := o.store.SetPrivateKey(privateKey); err != nil {
			return fmt.Errorf("failed to save generated wallet: %w", err)
		}
		signer, err := sign.NewEthereumRawSigner(privateKey)
		if err != nil {
			return fmt.Errorf("failed to create signer: %w", err)
		}
		fmt.Println()
		fmt.Println("Welcome! No wallet imported. A new wallet has been generated for you.")
		fmt.Printf("Address: %s\n", signer.PublicKey().Address().String())
		fmt.Println()
		fmt.Println("IMPORTANT: Run 'config wallet export' to save your private key to a file.")
		fmt.Println("INFO: You can import a different wallet anytime with 'config wallet import'.")
		fmt.Println()
	}

	stateSigner, err := o.buildStateSigner(privateKey)
	if err != nil {
		return err
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create tx signer: %w", err)
	}

	rpcs, err := o.store.GetAllRPCs()
	if err != nil {
		rpcs = make(map[uint64]string)
	}

	opts := []sdk.Option{}
	for chainID, rpcURL := range rpcs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, rpcURL))
	}

	client, err := sdk.NewClient(o.wsURL, stateSigner, txSigner, opts...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	o.client = client

	// Monitor WebSocket connection — only exit if this client is still active
	go func() {
		<-client.WaitCh()
		if o.client != client {
			return // replaced by reconnect, ignore
		}
		fmt.Println("\nWARNING: WebSocket connection lost. Exiting...")
		select {
		case <-o.exitCh:
		default:
			close(o.exitCh)
		}
	}()

	return nil
}

// reconnect closes the current client and creates a new one with the
// appropriate signer (session key if stored, default otherwise).
func (o *Operator) reconnect() error {
	old := o.client
	o.client = nil // mark stale so the WaitCh goroutine ignores the close
	old.Close()
	return o.connect()
}

func (o *Operator) Wait() <-chan struct{} {
	return o.exitCh
}

func (o *Operator) Complete(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(o.complete(d), d.GetWordBeforeCursor(), true)
}

func (o *Operator) complete(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	// First level commands
	if len(args) < 2 {
		return []prompt.Suggest{
			// Setup
			{Text: "help", Description: "Show help information"},
			{Text: "config", Description: "Configuration (wallet, rpc, node, session-key)"},

			// High-level operations
			{Text: "token-balance", Description: "Check on-chain token balance"},
			{Text: "approve", Description: "Approve token spending for deposits"},
			{Text: "deposit", Description: "Deposit funds to channel"},
			{Text: "withdraw", Description: "Withdraw funds from channel"},
			{Text: "transfer", Description: "Transfer funds to another wallet"},
			{Text: "close-channel", Description: "Close home channel on-chain"},
			{Text: "acknowledge", Description: "Acknowledge transfer or channel creation"},
			{Text: "checkpoint", Description: "Submit latest state on-chain"},

			// Node information
			{Text: "ping", Description: "Test node connection"},
			{Text: "chains", Description: "List supported blockchains"},
			{Text: "assets", Description: "List supported assets"},

			// User queries
			{Text: "balances", Description: "Get user balances"},
			{Text: "transactions", Description: "Get transaction history"},
			{Text: "action-allowances", Description: "Get action allowances"},

			// State management
			{Text: "state", Description: "Get latest state"},
			{Text: "home-channel", Description: "Get home channel"},
			{Text: "escrow-channel", Description: "Get escrow channel"},

			// App registry
			{Text: "app-info", Description: "Show application details"},
			{Text: "my-apps", Description: "List your registered applications"},
			{Text: "register-app", Description: "Register a new application"},

			// App sessions (Base Client - Low-level)
			{Text: "app-sessions", Description: "List app sessions"},

			// Security token operations
			{Text: "security-token", Description: "Security token operations"},

{Text: "exit", Description: "Exit the CLI"},
		}
	}

	// Second level
	if len(args) < 3 {
		switch args[0] {
		case "config":
			return []prompt.Suggest{
				{Text: "wallet", Description: "Wallet management"},
				{Text: "rpc", Description: "RPC management"},
				{Text: "node", Description: "Node info and connection"},
				{Text: "session-key", Description: "Session key management"},
			}
		case "close-channel", "acknowledge", "checkpoint":
			return o.getAssetSuggestions()
		case "token-balance", "approve", "deposit", "withdraw":
			return o.getChainSuggestions()
		case "security-token":
			return []prompt.Suggest{
				{Text: "approve", Description: "Approve security token spending"},
				{Text: "balance", Description: "Check escrowed security token balance"},
				{Text: "escrow", Description: "Escrow security tokens"},
				{Text: "initiate-withdrawal", Description: "Start unlock period"},
				{Text: "cancel-withdrawal", Description: "Cancel unlock and re-lock"},
				{Text: "withdraw", Description: "Withdraw unlocked security tokens"},
			}
		case "state", "home-channel":
			// state [wallet] <asset>, home-channel [wallet] <asset>
			// Suggest asset first (common case), wallet can be typed manually
			return o.getAssetSuggestions()
		case "transfer":
			// transfer <recipient> <asset> <amount>
			return o.getWalletSuggestion()
		case "balances", "transactions", "action-allowances":
			return o.getWalletSuggestion()
		case "assets":
			return o.getChainSuggestions()
		}
	}

	// Third level
	if len(args) < 4 {
		switch args[0] {
		case "config":
			switch args[1] {
			case "wallet":
				return []prompt.Suggest{
					{Text: "import", Description: "Import existing private key"},
					{Text: "generate", Description: "Generate new wallet"},
					{Text: "export", Description: "Export private key to file"},
				}
			case "rpc":
				return []prompt.Suggest{
					{Text: "import", Description: "Import blockchain RPC URL"},
				}
			case "node":
				return []prompt.Suggest{
					{Text: "set-ws-url", Description: "Set nitronode WebSocket URL"},
					{Text: "set-home-blockchain", Description: "Set home blockchain for channels"},
				}
			case "session-key":
				return []prompt.Suggest{
					{Text: "generate", Description: "Generate new session key"},
					{Text: "import", Description: "Import existing session key"},
					{Text: "clear", Description: "Clear session key, revert to default signer"},
					{Text: "register-channel-key", Description: "Register channel session key"},
					{Text: "channel-keys", Description: "List active channel session keys"},
					{Text: "register-app-key", Description: "Register app session key"},
					{Text: "app-keys", Description: "List active app session keys"},
				}
			}
		case "token-balance":
			// token-balance <chain_id> <asset>
			return o.getAssetSuggestions()
		case "approve", "deposit", "withdraw":
			// approve/deposit/withdraw <chain_id> <asset> <amount>
			return o.getAssetSuggestions()
		case "transfer":
			// transfer <recipient> <asset> <amount>
			return o.getAssetSuggestions()
		case "state", "home-channel":
			// state [wallet] <asset> — if wallet was explicitly provided, suggest asset
			return o.getAssetSuggestions()
		case "security-token":
			// security-token <subcommand> <chain_id> ...
			return o.getChainSuggestions()
		case "escrow-channel":
			// Escrow channel ID (no suggestion)
			return nil
		}
	}

	// Fourth level
	if len(args) < 5 {
		switch args[0] {
		case "config":
			switch args[1] {
			case "rpc":
				if args[2] == "import" {
					return o.getChainSuggestions()
				}
			case "node":
				if args[2] == "set-home-blockchain" {
					return o.getAssetSuggestions()
				}
			}
		}
	}

	// Fifth level
	if len(args) < 6 {
		if args[0] == "config" && args[1] == "node" && args[2] == "set-home-blockchain" {
			return o.getChainSuggestions()
		}
	}

	return nil
}

func (o *Operator) Execute(s string) {
	args := strings.Fields(s)
	if s == "" || len(args) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "help":
		o.showHelp()
	case "config":
		if len(args) < 2 {
			o.showConfig()
			return
		}
		switch args[1] {
		case "wallet":
			if len(args) < 3 {
				o.showWallet(ctx)
				return
			}
			switch args[2] {
			case "import":
				o.importWallet()
			case "generate":
				o.generateWallet()
			case "export":
				if len(args) < 4 {
					fmt.Println("ERROR: Usage: config wallet export <path>")
					return
				}
				o.exportWallet(args[3])
			default:
				fmt.Printf("ERROR: Unknown wallet command: %s\n", args[2])
				fmt.Println("Usage: config wallet [import|generate|export]")
			}
		case "rpc":
			if len(args) < 3 {
				fmt.Println("ERROR: Usage: config rpc import <chain_id> <rpc_url>")
				return
			}
			switch args[2] {
			case "import":
				if len(args) < 5 {
					fmt.Println("ERROR: Usage: config rpc import <chain_id> <rpc_url>")
					return
				}
				o.importRPC(ctx, args[3], args[4])
			default:
				fmt.Printf("ERROR: Unknown rpc command: %s\n", args[2])
				fmt.Println("Usage: config rpc [import]")
			}
		case "node":
			if len(args) < 3 {
				o.nodeInfo(ctx)
				return
			}
			switch args[2] {
			case "set-ws-url":
				if len(args) < 4 {
					fmt.Println("ERROR: Usage: config node set-ws-url <url>")
					return
				}
				o.setWSURL(args[3])
			case "set-home-blockchain":
				if len(args) < 5 {
					fmt.Println("ERROR: Usage: config node set-home-blockchain <asset> <chain_id>")
					return
				}
				o.setHomeBlockchain(ctx, args[3], args[4])
			default:
				fmt.Printf("ERROR: Unknown node command: %s\n", args[2])
				fmt.Println("Usage: config node [set-ws-url|set-home-blockchain]")
			}
		case "session-key":
			if len(args) < 3 {
				o.showSessionKey()
				return
			}
			switch args[2] {
			case "generate":
				o.generateSessionKey()
			case "import":
				o.importSessionKey()
			case "clear":
				o.clearSessionKey()
			case "register-channel-key":
				if len(args) < 6 {
					fmt.Println("ERROR: Usage: config session-key register-channel-key <session_key_address> <expires_hours> <assets>")
					fmt.Println("INFO: Assets are comma-separated, e.g. usdc,weth")
					return
				}
				o.createChannelSessionKey(ctx, args[3], args[4], args[5])
			case "channel-keys":
				wallet := o.getImportedWalletAddress()
				if wallet == "" {
					fmt.Println("ERROR: No wallet configured. Use 'config wallet import' first.")
					return
				}
				o.listChannelSessionKeys(ctx, wallet)
			case "register-app-key":
				if len(args) < 5 {
					fmt.Println("ERROR: Usage: config session-key register-app-key <session_key_address> <expires_hours> [app_ids] [session_ids]")
					fmt.Println("INFO: IDs are comma-separated. app_ids and session_ids are optional.")
					return
				}
				appIDs := ""
				sessionIDs := ""
				if len(args) >= 6 {
					appIDs = args[5]
				}
				if len(args) >= 7 {
					sessionIDs = args[6]
				}
				o.createAppSessionKey(ctx, args[3], args[4], appIDs, sessionIDs)
			case "app-keys":
				wallet := o.getImportedWalletAddress()
				if wallet == "" {
					fmt.Println("ERROR: No wallet configured. Use 'config wallet import' first.")
					return
				}
				o.listAppSessionKeys(ctx, wallet)
			default:
				fmt.Printf("ERROR: Unknown session-key command: %s\n", args[2])
				fmt.Println("Usage: config session-key [generate|import|clear|register-channel-key|channel-keys|register-app-key|app-keys]")
			}
		default:
			fmt.Printf("ERROR: Unknown config command: %s\n", args[1])
			fmt.Println("Usage: config [wallet|rpc|node|session-key]")
		}
	// High-level operations
	case "token-balance":
		if len(args) < 3 {
			fmt.Println("ERROR: Usage: token-balance <chain_id> <asset>")
			return
		}
		o.tokenBalance(ctx, args[1], args[2])
	case "approve":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: approve <chain_id> <asset> <amount>")
			return
		}
		o.approveToken(ctx, args[1], args[2], args[3])
	case "deposit":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: deposit <chain_id> <asset> <amount>")
			return
		}
		o.deposit(ctx, args[1], args[2], args[3])
	case "withdraw":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: withdraw <chain_id> <asset> <amount>")
			return
		}
		o.withdraw(ctx, args[1], args[2], args[3])
	case "transfer":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: transfer <recipient_address> <asset> <amount>")
			return
		}
		o.transfer(ctx, args[1], args[2], args[3])
	case "close-channel":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: close-channel <asset>")
			return
		}
		o.closeChannel(ctx, args[1])
	case "acknowledge":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: acknowledge <asset>")
			return
		}
		o.acknowledge(ctx, args[1])
	case "checkpoint":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: checkpoint <asset>")
			return
		}
		o.checkpoint(ctx, args[1])

	// Node information
	case "ping":
		o.ping(ctx)
	case "chains":
		o.listChains(ctx)
	case "assets":
		chainID := ""
		if len(args) >= 2 {
			chainID = args[1]
		}
		o.listAssets(ctx, chainID)

	// User queries
	case "balances":
		wallet := ""
		if len(args) >= 2 {
			wallet = args[1]
		} else {
			// Auto-fill with imported wallet
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: balances <wallet_address>")
				fmt.Println("INFO: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
				return
			}
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		}
		o.getBalances(ctx, wallet)
	case "transactions":
		wallet := ""
		if len(args) >= 2 {
			wallet = args[1]
		} else {
			// Auto-fill with imported wallet
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: transactions <wallet_address>")
				fmt.Println("INFO: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
				return
			}
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		}
		o.listTransactions(ctx, wallet)

	// State management (low-level)
	case "state":
		wallet := ""
		asset := ""
		if len(args) >= 3 {
			wallet = args[1]
			asset = args[2]
		} else if len(args) == 2 {
			// Auto-fill wallet, user provided asset
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: state <wallet_address> <asset>")
				fmt.Println("INFO: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
				return
			}
			asset = args[1]
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		} else {
			fmt.Println("ERROR: Usage: state <wallet_address> <asset>")
			fmt.Println("INFO: Or: state <asset> (uses configured wallet)")
			return
		}
		o.getLatestState(ctx, wallet, asset)
	case "home-channel":
		wallet := ""
		asset := ""
		if len(args) >= 3 {
			wallet = args[1]
			asset = args[2]
		} else if len(args) == 2 {
			// Auto-fill wallet, user provided asset
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: home-channel <wallet_address> <asset>")
				fmt.Println("INFO: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
				return
			}
			asset = args[1]
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		} else {
			fmt.Println("ERROR: Usage: home-channel <wallet_address> <asset>")
			fmt.Println("INFO: Or: home-channel <asset> (uses configured wallet)")
			return
		}
		o.getHomeChannel(ctx, wallet, asset)
	case "escrow-channel":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: escrow-channel <escrow_channel_id>")
			return
		}
		o.getEscrowChannel(ctx, args[1])

	// App registry
	case "app-info":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: app-info <app_id>")
			return
		}
		o.getApps(ctx, &args[1], nil)

	case "my-apps":
		wallet := o.getImportedWalletAddress()
		if wallet == "" {
			fmt.Println("ERROR: No wallet configured. Use 'config wallet import' first.")
			return
		}
		o.getApps(ctx, nil, &wallet)

	case "register-app":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: register-app <app_id> [no-approval]")
			fmt.Println("INFO: Pass 'no-approval' as second arg to allow session creation without owner approval")
			return
		}
		noApproval := len(args) >= 3 && args[2] == "no-approval"
		o.registerApp(ctx, args[1], "", noApproval)

	// User action allowances
	case "action-allowances":
		wallet := ""
		if len(args) >= 2 {
			wallet = args[1]
		} else {
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: action-allowances <wallet_address>")
				fmt.Println("INFO: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
				return
			}
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		}
		o.getActionAllowances(ctx, wallet)

	// App sessions
	case "app-sessions":
		wallet := o.getImportedWalletAddress()
		o.listAppSessions(ctx, wallet)


	// Security token operations
	case "security-token":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: security-token <command> ...")
			fmt.Println("Commands: approve, balance, escrow, initiate-withdrawal, cancel-withdrawal, withdraw")
			return
		}
		switch args[1] {
		case "approve":
			if len(args) < 4 {
				fmt.Println("ERROR: Usage: security-token approve <chain_id> <amount>")
				return
			}
			o.approveSecurityToken(ctx, args[2], args[3])
		case "escrow":
			if len(args) < 4 {
				fmt.Println("ERROR: Usage: security-token escrow <chain_id> [target_address] <amount>")
				fmt.Println("INFO: If target_address is omitted, your own wallet is used.")
				return
			}
			if len(args) >= 5 {
				o.escrowSecurityTokens(ctx, args[2], args[3], args[4])
			} else {
				o.escrowSecurityTokens(ctx, args[2], "", args[3])
			}
		case "initiate-withdrawal":
			if len(args) < 3 {
				fmt.Println("ERROR: Usage: security-token initiate-withdrawal <chain_id>")
				return
			}
			o.initiateSecurityWithdrawal(ctx, args[2])
		case "cancel-withdrawal":
			if len(args) < 3 {
				fmt.Println("ERROR: Usage: security-token cancel-withdrawal <chain_id>")
				return
			}
			o.cancelSecurityWithdrawal(ctx, args[2])
		case "withdraw":
			if len(args) < 4 {
				fmt.Println("ERROR: Usage: security-token withdraw <chain_id> <destination_address>")
				return
			}
			o.withdrawSecurityTokens(ctx, args[2], args[3])
		case "balance":
			if len(args) < 3 {
				fmt.Println("ERROR: Usage: security-token balance <chain_id> [wallet_address]")
				return
			}
			wallet := ""
			if len(args) >= 4 {
				wallet = args[3]
			} else {
				wallet = o.getImportedWalletAddress()
				if wallet == "" {
					fmt.Println("ERROR: No wallet configured. Use 'config wallet import' first or specify a wallet address.")
					return
				}
				fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
			}
			o.securityBalance(ctx, args[2], wallet)
		default:
			fmt.Printf("ERROR: Unknown security-token command: %s\n", args[1])
			fmt.Println("Commands: approve, balance, escrow, initiate-withdrawal, cancel-withdrawal, withdraw")
		}

	case "exit":
		fmt.Println("Exiting...")
		close(o.exitCh)
	default:
		fmt.Printf("ERROR: Unknown command: %s (type 'help' for available commands)\n", args[0])
	}
}

func (o *Operator) getChainSuggestions() []prompt.Suggest {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chains, err := o.client.GetBlockchains(ctx)
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, len(chains))
	for i, chain := range chains {
		suggestions[i] = prompt.Suggest{
			Text:        fmt.Sprintf("%d", chain.ID),
			Description: fmt.Sprintf("%s (ID: %d)", chain.Name, chain.ID),
		}
	}
	return suggestions
}

func (o *Operator) getAssetSuggestions() []prompt.Suggest {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assets, err := o.client.GetAssets(ctx, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, len(assets))
	for i, asset := range assets {
		suggestions[i] = prompt.Suggest{
			Text:        asset.Symbol,
			Description: fmt.Sprintf("%s (%d tokens)", asset.Name, len(asset.Tokens)),
		}
	}
	return suggestions
}

func (o *Operator) getWalletSuggestion() []prompt.Suggest {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return nil
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil
	}

	address := signer.PublicKey().Address().String()

	return []prompt.Suggest{
		{
			Text:        address,
			Description: "Your imported wallet",
		},
	}
}

func (o *Operator) parseChainID(chainIDStr string) (uint64, error) {
	chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}
	return chainID, nil
}

func (o *Operator) parseAmount(amountStr string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid amount: %s", amountStr)
	}
	return amount, nil
}

func (o *Operator) getImportedWalletAddress() string {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return ""
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return ""
	}

	return signer.PublicKey().Address().String()
}
