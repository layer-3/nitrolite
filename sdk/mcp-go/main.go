// Nitrolite Go SDK MCP Server
//
// Exposes the Nitrolite protocol and Go SDK API surface to AI agents and IDEs
// via the Model Context Protocol. Reads protocol docs and Go SDK source at
// startup to build structured knowledge.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ---------------------------------------------------------------------------
// Paths — resolved relative to repo root
// ---------------------------------------------------------------------------

var (
	repoRoot     string
	protocolDocs map[string]string
	concepts     map[string]string
	rpcMethods   []rpcMethodDoc
	goMethods    []methodInfo
)

type rpcMethodDoc struct {
	Method      string
	Description string
	Access      string
}

type methodInfo struct {
	Name      string
	Signature string
	Comment   string
	Category  string
}

// ---------------------------------------------------------------------------
// File helpers
// ---------------------------------------------------------------------------

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func findRepoRoot() string {
	// Walk up from executable location looking for go.mod with "layer-3/nitrolite"
	dir, _ := os.Getwd()
	for {
		content := readFile(filepath.Join(dir, "go.mod"))
		if strings.Contains(content, "layer-3/nitrolite") && !strings.Contains(content, "sdk/mcp-go") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fallback: assume we're in sdk/mcp-go/
	return filepath.Join(".", "..", "..")
}

// ---------------------------------------------------------------------------
// Loaders
// ---------------------------------------------------------------------------

func loadProtocolDocs() {
	protocolDocs = make(map[string]string)
	files := []string{
		"overview.md", "terminology.md", "cryptography.md", "state-model.md",
		"channel-protocol.md", "enforcement.md", "cross-chain-and-assets.md",
		"interactions.md", "security-and-limitations.md",
	}
	dir := filepath.Join(repoRoot, "docs", "protocol")
	for _, f := range files {
		key := strings.TrimSuffix(f, ".md")
		content := readFile(filepath.Join(dir, f))
		if content != "" {
			protocolDocs[key] = content
		}
	}
}

func loadTerminology() {
	concepts = make(map[string]string)
	content, ok := protocolDocs["terminology"]
	if !ok {
		return
	}
	sections := strings.Split(content, "### ")
	for _, section := range sections[1:] {
		lines := strings.SplitN(strings.TrimSpace(section), "\n", 2)
		name := strings.TrimSpace(lines[0])
		body := ""
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
			// Trim at next section
			if idx := strings.Index(body, "\n## "); idx >= 0 {
				body = body[:idx]
			}
		}
		if name != "" && body != "" {
			concepts[strings.ToLower(name)] = fmt.Sprintf("**%s**\n\n%s", name, body)
		}
	}
}

func loadClearnodeAPI() {
	content := readFile(filepath.Join(repoRoot, "clearnode", "docs", "API.md"))
	if content == "" {
		return
	}
	// Parse the API endpoint table
	re := regexp.MustCompile("(?m)^\\| `([^`]+)` +\\| (.+?) +\\| (\\w+) +\\|$")
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		rpcMethods = append(rpcMethods, rpcMethodDoc{
			Method:      m[1],
			Description: strings.TrimSpace(m[2]),
			Access:      strings.TrimSpace(m[3]),
		})
	}
}

func loadGoSDKMethods() {
	// Parse exported methods from Go SDK source files
	files := []struct {
		path     string
		category string
	}{
		{"sdk/go/channel.go", ""},
		{"sdk/go/node.go", "Node & Config"},
		{"sdk/go/user.go", "User Queries"},
		{"sdk/go/app_session.go", "App Sessions"},
		{"sdk/go/app_registry.go", "App Registry"},
		{"sdk/go/client.go", ""},
	}

	methodRe := regexp.MustCompile(`(?m)((?:^//[^\n]*\n)+)func \(c \*Client\) (\w+)\(([^)]*)\)\s*(.*)`)
	for _, f := range files {
		content := readFile(filepath.Join(repoRoot, f.path))
		if content == "" {
			continue
		}
		for _, m := range methodRe.FindAllStringSubmatch(content, -1) {
			// Join multi-line // comments into a single description
			rawComment := strings.TrimSpace(m[1])
			var commentLines []string
			for _, line := range strings.Split(rawComment, "\n") {
				line = strings.TrimPrefix(strings.TrimSpace(line), "// ")
				if line != "" {
					commentLines = append(commentLines, line)
				}
			}
			comment := strings.Join(commentLines, " ")
			name := m[2]
			params := m[3]
			returns := strings.TrimSpace(m[4])

			// Skip unexported
			if name[0] >= 'a' && name[0] <= 'z' {
				continue
			}

			cat := f.category
			if cat == "" {
				cat = categorizeGoMethod(name)
			}

			sig := fmt.Sprintf("func (c *Client) %s(%s) %s", name, params, returns)
			goMethods = append(goMethods, methodInfo{
				Name:      name,
				Signature: sig,
				Comment:   comment,
				Category:  cat,
			})
		}
	}

	// Also parse constructor and standalone functions
	clientContent := readFile(filepath.Join(repoRoot, "sdk/go/client.go"))
	if strings.Contains(clientContent, "func NewClient(") {
		goMethods = append(goMethods, methodInfo{
			Name:      "NewClient",
			Signature: "func NewClient(wsURL string, stateSigner core.ChannelSigner, rawSigner sign.Signer, opts ...Option) (*Client, error)",
			Comment:   "Creates a new Nitrolite SDK client connected to a clearnode",
			Category:  "Connection",
		})
	}
}

func categorizeGoMethod(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "channel") || strings.Contains(lower, "deposit") ||
		strings.Contains(lower, "withdraw") || strings.Contains(lower, "transfer") ||
		strings.Contains(lower, "checkpoint") || strings.Contains(lower, "challenge") ||
		strings.Contains(lower, "acknowledge") || strings.Contains(lower, "close"):
		return "Channels & Transactions"
	case strings.Contains(lower, "appsession") || strings.Contains(lower, "appstate") ||
		strings.Contains(lower, "appdef") || strings.Contains(lower, "rebalance"):
		return "App Sessions"
	case strings.Contains(lower, "sessionkey") || strings.Contains(lower, "keystate"):
		return "Session Keys"
	case strings.Contains(lower, "escrow") || strings.Contains(lower, "security") ||
		strings.Contains(lower, "locked"):
		return "Security Tokens"
	case strings.Contains(lower, "app") && strings.Contains(lower, "register"):
		return "App Registry"
	case strings.Contains(lower, "balance") || strings.Contains(lower, "transaction") ||
		strings.Contains(lower, "allowance") || strings.Contains(lower, "user"):
		return "User Queries"
	case strings.Contains(lower, "config") || strings.Contains(lower, "blockchain") ||
		strings.Contains(lower, "asset") || strings.Contains(lower, "ping"):
		return "Node & Config"
	default:
		return "Other"
	}
}

// ---------------------------------------------------------------------------
// Resource content builders
// ---------------------------------------------------------------------------

func goAPIMethodsContent() string {
	grouped := make(map[string][]methodInfo)
	for _, m := range goMethods {
		grouped[m.Category] = append(grouped[m.Category], m)
	}

	var b strings.Builder
	b.WriteString("# Nitrolite Go SDK — Client Methods\n\n")
	b.WriteString("Package: `github.com/layer-3/nitrolite/sdk/go`\n\n")
	for _, cat := range []string{"Connection", "Channels & Transactions", "App Sessions", "Session Keys", "Security Tokens", "App Registry", "User Queries", "Node & Config", "Other"} {
		ms := grouped[cat]
		if len(ms) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("## %s\n\n", cat))
		for _, m := range ms {
			b.WriteString(fmt.Sprintf("### `%s`\n```go\n%s\n```\n%s\n\n", m.Name, m.Signature, m.Comment))
		}
	}
	return b.String()
}

func rpcMethodsContent() string {
	var b strings.Builder
	b.WriteString("# Clearnode RPC Methods\n\nAll available RPC methods exposed by a clearnode.\n\n")
	b.WriteString("| Method | Description | Access |\n|---|---|---|\n")
	for _, m := range rpcMethods {
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", m.Method, m.Description, m.Access))
	}
	b.WriteString("\n## Message Format\n\n")
	b.WriteString("**Request:** `{ \"req\": [REQUEST_ID, METHOD, PARAMS, TIMESTAMP], \"sig\": [\"SIGNATURE\"] }`\n\n")
	b.WriteString("**Response:** `{ \"res\": [REQUEST_ID, METHOD, DATA, TIMESTAMP], \"sig\": [\"SIGNATURE\"] }`\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// Server setup
// ---------------------------------------------------------------------------

func main() {
	repoRoot = findRepoRoot()

	// Load all data
	loadProtocolDocs()
	loadTerminology()
	loadClearnodeAPI()
	loadGoSDKMethods()

	s := server.NewMCPServer(
		"nitrolite-go-sdk",
		"0.1.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)

	// ======================== API RESOURCES ================================

	addResource(s, "api-methods", "nitrolite://api/methods",
		"All Go SDK Client methods organized by category", goAPIMethodsContent)

	// ======================== PROTOCOL RESOURCES ===========================

	addProtocolResource(s, "protocol-overview", "nitrolite://protocol/overview",
		"Protocol overview, design goals, system roles", "overview")

	addProtocolResource(s, "protocol-terminology", "nitrolite://protocol/terminology",
		"Canonical definitions of all protocol terms", "terminology")

	addProtocolResource(s, "protocol-wire-format", "nitrolite://protocol/wire-format",
		"Message envelope structure and message types", "interactions")

	addResource(s, "protocol-rpc-methods", "nitrolite://protocol/rpc-methods",
		"All clearnode RPC methods with descriptions", rpcMethodsContent)

	addProtocolResource(s, "protocol-channel-lifecycle", "nitrolite://protocol/channel-lifecycle",
		"Channel states, transitions, and advancement rules", "channel-protocol")

	addProtocolResource(s, "protocol-state-model", "nitrolite://protocol/state-model",
		"State structure, versioning, and consistency rules", "state-model")

	addProtocolResource(s, "protocol-enforcement", "nitrolite://protocol/enforcement",
		"On-chain enforcement, checkpoints, and dispute resolution", "enforcement")

	// Auth flow (synthesized)
	s.AddResource(mcp.Resource{
		URI:         "nitrolite://protocol/auth-flow",
		Name:        "protocol-auth-flow",
		Description: "Challenge-response authentication sequence",
		MIMEType:    "text/markdown",
	}, handleStaticResource(authFlowContent))

	// ======================== CLEARNODE RESOURCES ============================

	addResource(s, "clearnode-entities", "nitrolite://clearnode/entities",
		"Clearnode data entities: Channel, Asset, Account, Transaction", func() string {
			content := readFile(filepath.Join(repoRoot, "clearnode", "docs", "Entities.md"))
			if content == "" {
				return "# Entities\n\nFile not found."
			}
			return content
		})

	addResource(s, "clearnode-session-keys", "nitrolite://clearnode/session-keys",
		"Session key delegation, spending caps, and application isolation", func() string {
			content := readFile(filepath.Join(repoRoot, "clearnode", "docs", "SessionKeys.md"))
			if content == "" {
				return "# Session Keys\n\nFile not found."
			}
			return content
		})

	addResource(s, "clearnode-protocol", "nitrolite://clearnode/protocol",
		"Clearnode protocol: channel creation, virtual apps, ledger model", func() string {
			content := readFile(filepath.Join(repoRoot, "clearnode", "docs", "Clearnode.protocol.md"))
			if content == "" {
				return "# Clearnode Protocol\n\nFile not found."
			}
			return content
		})

	// ======================== SECURITY RESOURCES ===========================

	addProtocolResource(s, "security-overview", "nitrolite://security/overview",
		"Security guarantees and trust assumptions", "security-and-limitations")

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://security/app-session-patterns",
		Name:        "security-app-session-patterns",
		Description: "Quorum design, challenge periods, decentralization patterns",
		MIMEType:    "text/markdown",
	}, handleStaticResource(appSessionPatternsContent))

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://security/state-invariants",
		Name:        "security-state-invariants",
		Description: "Fund conservation, version ordering, signature rules",
		MIMEType:    "text/markdown",
	}, handleStaticResource(stateInvariantsContent))

	// ======================== USE CASES ====================================

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://use-cases",
		Name:        "use-cases",
		Description: "What you can build: payments, gaming, escrow, AI agents",
		MIMEType:    "text/markdown",
	}, handleStaticResource(useCasesContent))

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://use-cases/ai-agents",
		Name:        "use-cases-ai-agents",
		Description: "AI agent payment patterns and framework integration",
		MIMEType:    "text/markdown",
	}, handleStaticResource(aiAgentUseCasesContent))

	// ======================== GO EXAMPLES ==================================

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://examples/full-transfer-script",
		Name:        "examples-full-transfer",
		Description: "Complete Go transfer script: connect, deposit, transfer, close",
		MIMEType:    "text/markdown",
	}, handleStaticResource(goTransferExampleContent))

	s.AddResource(mcp.Resource{
		URI:         "nitrolite://examples/full-app-session-script",
		Name:        "examples-full-app-session",
		Description: "Complete Go app session script: create, update, close",
		MIMEType:    "text/markdown",
	}, handleStaticResource(goAppSessionExampleContent))

	// ======================== TOOLS =======================================

	// search_api
	s.AddTool(mcp.NewTool("search_api",
		mcp.WithDescription("Fuzzy search across all Go SDK methods"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g. \"transfer\", \"session\", \"balance\")")),
	), handleSearchAPI)

	// lookup_method
	s.AddTool(mcp.NewTool("lookup_method",
		mcp.WithDescription("Look up a specific Go SDK Client method by name"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Method name (e.g. \"Transfer\", \"Deposit\", \"GetConfig\")")),
	), handleLookupMethod)

	// explain_concept
	s.AddTool(mcp.NewTool("explain_concept",
		mcp.WithDescription("Plain-English explanation of a Nitrolite protocol concept"),
		mcp.WithString("concept", mcp.Required(), mcp.Description("Concept name (e.g. \"state channel\", \"app session\", \"challenge period\")")),
	), handleExplainConcept)

	// lookup_rpc_method
	s.AddTool(mcp.NewTool("lookup_rpc_method",
		mcp.WithDescription("Look up a clearnode RPC method — returns description and access level"),
		mcp.WithString("method", mcp.Required(), mcp.Description("RPC method name (e.g. \"get_channels\", \"transfer\", \"auth_request\")")),
	), handleLookupRPCMethod)

	// scaffold_project
	s.AddTool(mcp.NewTool("scaffold_project",
		mcp.WithDescription("Generate a starter Go project structure for a new Nitrolite app"),
		mcp.WithString("template", mcp.Required(), mcp.Description("Project template: \"transfer-app\", \"app-session\", or \"ai-agent\"")),
	), handleScaffoldProject)

	// ======================== PROMPTS =====================================

	s.AddPrompt(mcp.Prompt{
		Name:        "create-channel-app",
		Description: "Step-by-step guide to build a Go state channel app",
	}, handleCreateChannelAppPrompt)

	s.AddPrompt(mcp.Prompt{
		Name:        "build-ai-agent-app",
		Description: "Guided conversation for building a Go AI agent with Nitrolite payments",
	}, handleBuildAIAgentPrompt)

	// Start
	if err := server.ServeStdio(s); err != nil {
		log.Fatal("Fatal:", err)
	}
}

// ---------------------------------------------------------------------------
// Resource helpers
// ---------------------------------------------------------------------------

func addResource(s *server.MCPServer, name, uri, description string, contentFn func() string) {
	s.AddResource(mcp.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    "text/markdown",
	}, func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{URI: uri, MIMEType: "text/markdown", Text: contentFn()},
		}, nil
	})
}

func addProtocolResource(s *server.MCPServer, name, uri, description, docKey string) {
	s.AddResource(mcp.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    "text/markdown",
	}, func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		text, ok := protocolDocs[docKey]
		if !ok {
			text = fmt.Sprintf("# %s\n\nProtocol doc '%s' not found.", name, docKey)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{URI: uri, MIMEType: "text/markdown", Text: text},
		}, nil
	})
}

func handleStaticResource(content string) server.ResourceHandlerFunc {
	return func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "text/markdown", Text: content},
		}, nil
	}
}

// ---------------------------------------------------------------------------
// Tool handlers
// ---------------------------------------------------------------------------

func handleSearchAPI(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := strings.ToLower(req.GetString("query", ""))

	var matches []string
	for _, m := range goMethods {
		if strings.Contains(strings.ToLower(m.Name), query) ||
			strings.Contains(strings.ToLower(m.Comment), query) ||
			strings.Contains(strings.ToLower(m.Category), query) {
			matches = append(matches, fmt.Sprintf("- `%s` — %s (%s)", m.Name, m.Comment, m.Category))
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No methods matching \"%s\". Try broader terms.", query)), nil
	}

	text := fmt.Sprintf("# Search results for \"%s\"\n\n## Methods (%d matches)\n\n%s",
		query, len(matches), strings.Join(matches, "\n"))
	return mcp.NewToolResultText(text), nil
}

func handleLookupMethod(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := strings.ToLower(req.GetString("name", ""))

	var matches []string
	for _, m := range goMethods {
		if strings.Contains(strings.ToLower(m.Name), name) {
			matches = append(matches, fmt.Sprintf("## %s\n**Signature:**\n```go\n%s\n```\n**Category:** %s\n**Description:** %s",
				m.Name, m.Signature, m.Category, m.Comment))
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No method matching \"%s\" found. %d methods indexed.", name, len(goMethods))), nil
	}
	return mcp.NewToolResultText(strings.Join(matches, "\n\n---\n\n")), nil
}

func handleExplainConcept(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := strings.ToLower(strings.TrimSpace(req.GetString("concept", "")))

	// Direct match
	if def, ok := concepts[query]; ok {
		return mcp.NewToolResultText(def), nil
	}

	// Fuzzy match
	var matches []string
	for key, def := range concepts {
		if strings.Contains(key, query) || strings.Contains(query, key) {
			matches = append(matches, def)
		}
	}
	if len(matches) > 0 {
		return mcp.NewToolResultText(strings.Join(matches, "\n\n---\n\n")), nil
	}

	// Word-level fuzzy
	words := strings.Fields(query)
	for key, def := range concepts {
		for _, w := range words {
			if strings.Contains(key, w) {
				matches = append(matches, def)
				break
			}
		}
	}
	if len(matches) > 0 {
		if len(matches) > 5 {
			matches = matches[:5]
		}
		return mcp.NewToolResultText(fmt.Sprintf("No exact match for \"%s\". Related:\n\n%s", query, strings.Join(matches, "\n\n---\n\n"))), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("No concept matching \"%s\" found. %d concepts indexed.", query, len(concepts))), nil
}

func handleLookupRPCMethod(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := strings.ToLower(strings.TrimSpace(req.GetString("method", "")))

	// Direct match
	for _, m := range rpcMethods {
		if m.Method == query {
			return mcp.NewToolResultText(fmt.Sprintf("## RPC: `%s`\n\n**Description:** %s\n**Access:** %s",
				m.Method, m.Description, m.Access)), nil
		}
	}

	// Fuzzy match
	var matches []string
	for _, m := range rpcMethods {
		if strings.Contains(m.Method, query) || strings.Contains(query, m.Method) {
			matches = append(matches, fmt.Sprintf("- `%s` — %s (%s)", m.Method, m.Description, m.Access))
		}
	}
	if len(matches) > 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Matching RPC methods:\n\n%s", strings.Join(matches, "\n"))), nil
	}

	methods := make([]string, 0, len(rpcMethods))
	for _, m := range rpcMethods {
		methods = append(methods, m.Method)
	}
	return mcp.NewToolResultText(fmt.Sprintf("Unknown RPC method \"%s\". Available: %s", query, strings.Join(methods, ", "))), nil
}

func handleScaffoldProject(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	template := req.GetString("template", "")

	templates := map[string]string{
		"transfer-app": goScaffoldTransfer,
		"app-session":  goScaffoldAppSession,
		"ai-agent":     goScaffoldAIAgent,
	}

	code, ok := templates[template]
	if !ok {
		return mcp.NewToolResultText(fmt.Sprintf("Unknown template \"%s\". Available: transfer-app, app-session, ai-agent", template)), nil
	}

	goMod := fmt.Sprintf(`module my-nitrolite-%s

go 1.25.0

require (
	github.com/layer-3/nitrolite v0.0.0
	github.com/shopspring/decimal v1.4.0
)`, template)

	text := fmt.Sprintf("# Scaffold: %s\n\n## go.mod\n```\n%s\n```\n\n## main.go\n```go\n%s\n```\n\n## .env.example\n```\nPRIVATE_KEY=your_hex_key\nCLEARNODE_URL=wss://clearnode.example.com/ws\nRPC_URL=https://rpc.sepolia.org\n```\n\n## Setup\n```bash\ngo mod tidy\ngo run .\n```",
		template, goMod, code)

	return mcp.NewToolResultText(text), nil
}

// ---------------------------------------------------------------------------
// Prompt handlers
// ---------------------------------------------------------------------------

func handleCreateChannelAppPrompt(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{{
			Role: mcp.RoleUser,
			Content: mcp.TextContent{
				Type: "text",
				Text: `Guide me through building a Nitrolite state channel application in Go. Cover:

1. **Setup** — Install dependencies, create signers with sign.NewEthereumMsgSigner and sign.NewEthereumRawSigner
2. **Client Creation** — sdk.NewClient with functional options (WithBlockchainRPC, WithHandshakeTimeout)
3. **Channel Lifecycle** — Deposit (creates channel), Transfer, Checkpoint (on-chain), CloseHomeChannel + Checkpoint
4. **App Sessions** — CreateAppSession, SubmitAppState (Operate/Withdraw/Close intents), SubmitAppSessionDeposit
5. **Error Handling** — Go error patterns, context.WithTimeout, defer client.Close()
6. **Testing** — Standard Go test patterns with *_test.go files

For each step, show complete Go code examples using the latest SDK from github.com/layer-3/nitrolite/sdk/go.
Use github.com/shopspring/decimal for amounts. Use context.Context for all async operations.`,
			},
		}},
	}, nil
}

func handleBuildAIAgentPrompt(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{{
			Role: mcp.RoleUser,
			Content: mcp.TextContent{
				Type: "text",
				Text: `I want to build an AI agent in Go that uses Nitrolite state channels for payments. Guide me through:

1. **Agent Wallet Setup** — Create signers from a private key, configure the SDK client
2. **Channel Management** — Open a channel, deposit funds for the agent
3. **Automated Payments** — A goroutine-safe payment function using context and mutexes
4. **Session Key Delegation** — Set up session keys with spending caps
5. **Agent-to-Agent Payments** — Transfer between two autonomous Go agents
6. **Graceful Shutdown** — Handle OS signals, defer client.Close(), WaitCh()
7. **Error Handling** — Wrapped errors, retry patterns, connection recovery

For each step, show complete Go code using github.com/layer-3/nitrolite/sdk/go.
Use context.Context for timeouts. Use decimal.Decimal for amounts. Follow standard Go patterns.`,
			},
		}},
	}, nil
}

// ---------------------------------------------------------------------------
// Static content
// ---------------------------------------------------------------------------

const authFlowContent = `# Authentication Flow

The clearnode uses a challenge-response mechanism based on Ethereum signatures, with optional JWT session management.

## Flow

1. **auth_request** — Client sends address + optional session key parameters
2. **auth_challenge** — Server responds with a random challenge token (UUID)
3. **auth_verify** — Client signs the challenge with their private key and sends it back
4. **JWT issued** — Server responds with a JWT token (valid 24h by default)

## Session Keys

Authentication supports delegated session keys with spending caps:
- Specify a ` + "`session_key`" + ` address in auth_request to enable delegation
- Set per-asset ` + "`allowances`" + ` to limit spending (e.g., 100 USDC max)
- Session keys expire at the specified ` + "`expires_at`" + ` timestamp

## Example (JSON-RPC)

` + "```json" + `
// 1. Request
{ "req": [1, "auth_request", { "address": "0x...", "session_key": "0x..." }, 1619123456789], "sig": ["0x..."] }

// 2. Challenge
{ "res": [1, "auth_challenge", { "challenge_message": "550e8400-..." }, 1619123456789], "sig": ["0x..."] }

// 3. Verify
{ "req": [2, "auth_verify", { "challenge": "550e8400-..." }, 1619123456789], "sig": ["0x..."] }

// 4. Success + JWT
{ "res": [2, "auth_verify", { "address": "0x...", "success": true, "jwt_token": "eyJ..." }, 1619123456789], "sig": ["0x..."] }
` + "```" + `
`

const appSessionPatternsContent = `# App Session Security Patterns

Best practices for building secure, decentralization-ready app sessions in Go.

## Quorum Design

App sessions use weight-based quorum for governance:

` + "```go" + `
definition := app.AppDefinitionV1{
    ApplicationID: "my-app",
    Participants: []app.AppParticipantV1{
        {WalletAddress: addr1, SignatureWeight: 50},
        {WalletAddress: addr2, SignatureWeight: 50},
    },
    Quorum: 100, // Both must agree
    Nonce:  1,
}
` + "```" + `

### Recommended Patterns

- **Equal 2-of-2:** weights [50, 50], quorum 100 — both must agree
- **2-of-3 with arbitrator:** weights [50, 50, 50], quorum 100 — any two can authorize
- **Weighted operator:** weights [70, 30], quorum 70 — operator has majority

## Challenge Periods

- **Short (1 hour):** Low-value, high-frequency
- **Medium (24 hours):** Recommended default
- **Long (7 days):** High-value operations

## State Invariants

1. **Fund conservation:** Total allocations == committed amount
2. **Version ordering:** Each version = previous + 1
3. **Signature requirements:** Meet quorum threshold
4. **Non-negative allocations:** No participant below zero

## Decentralization-Ready Patterns

1. Use quorum >= any single participant's weight
2. Always use challenge periods
3. Keep state transitions deterministic
4. Support unilateral on-chain enforcement
5. Use session keys with spending caps
`

const stateInvariantsContent = `# State Invariants

Critical invariants that MUST hold across all state transitions.

## Ledger Invariant (Fund Conservation)

` + "```" + `
UserAllocation + NodeAllocation == UserNetFlow + NodeNetFlow
` + "```" + `

No assets created or destroyed through state transitions.

## Allocation Non-Negativity

All allocation values MUST be non-negative. Net flow values MAY be negative.

## Version Ordering

- **Off-chain:** Each new version = previous + 1
- **On-chain:** Submitted version must be strictly greater than current on-chain version

## Signature Requirements

- **Mutually signed** (user + node) = enforceable on-chain
- **Node-only pending** = NOT enforceable until user acknowledges

## Locked Funds

Unless closing: UserAllocation + NodeAllocation == LockedFunds
`

const useCasesContent = `# Nitrolite Use Cases

## Peer-to-Peer Payments
Instant, gas-free transfers. Open channel, transfer, settle on-chain when needed.
**Go SDK:** ` + "`client.Deposit()`" + `, ` + "`client.Transfer()`" + `, ` + "`client.CloseHomeChannel()`" + `

## Gaming (Real-Time Wagering)
Turn-based or real-time games with token wagering via app sessions.
**Go SDK:** ` + "`client.CreateAppSession()`" + `, ` + "`client.SubmitAppState()`" + ` (close via Close intent)

## AI Agent Payments
Autonomous agents making payments via state channels. See ` + "`nitrolite://use-cases/ai-agents`" + `.
**Go SDK:** ` + "`sdk.NewClient()`" + `, ` + "`client.Transfer()`" + `

## Multi-Party Escrow
Multiple parties commit funds, release on quorum. Custom weights for governance.

## Cross-Chain Operations
Move assets between blockchains through the escrow mechanism.

## Streaming Payments
Continuous micro-transfers (pay-per-second for compute, bandwidth, content).
`

const aiAgentUseCasesContent = `# AI Agent Use Cases (Go)

## Why State Channels for AI Agents?

AI agents need frequent, small payments. On-chain is too slow and expensive. State channels provide instant finality and near-zero cost.

## Agent Setup

` + "```go" + `
import (
    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
)

stateSigner, _ := sign.NewEthereumMsgSigner(agentPrivateKey)
txSigner, _ := sign.NewEthereumRawSigner(agentPrivateKey)

client, _ := sdk.NewClient(clearnodeURL, stateSigner, txSigner,
    sdk.WithBlockchainRPC(chainID, rpcURL),
)
defer client.Close()
` + "```" + `

## Agent-to-Agent Payments

Two Go agents transact directly through state channels:
1. Both connect to the same clearnode
2. Agent A calls ` + "`client.Transfer(ctx, agentB, \"usdc\", amount)`" + `
3. Agent B receives instantly — no on-chain tx needed

## Session Key Delegation

Use session keys with spending caps for safe autonomous operation:
- Main wallet authorizes a session key during auth
- Session key has max spending allowance
- Once cap reached, operations are rejected

## Graceful Shutdown

` + "```go" + `
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigCh
    client.Close()
}()

<-client.WaitCh() // Block until connection closes
` + "```" + `
`

const goTransferExampleContent = `# Complete Go Transfer Script

` + "```go" + `
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    recipient := os.Getenv("RECIPIENT")
    var chainID uint64 = 11155111 // Sepolia

    // 1. Create signers
    stateSigner, err := sign.NewEthereumMsgSigner(privateKey)
    if err != nil {
        log.Fatal(err)
    }
    txSigner, err := sign.NewEthereumRawSigner(privateKey)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Create client
    client, err := sdk.NewClient(clearnodeURL, stateSigner, txSigner,
        sdk.WithBlockchainRPC(chainID, rpcURL),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // 3. Approve token spending (one-time)
    _, err = client.ApproveToken(ctx, chainID, "usdc", decimal.NewFromInt(1000))
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Token approved")

    // 4. Deposit — creates channel if needed
    state, err := client.Deposit(ctx, chainID, "usdc", decimal.NewFromInt(10))
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Deposited 10 USDC, state version: %d", state.Version)

    // 5. Checkpoint on-chain
    txHash, err := client.Checkpoint(ctx, "usdc")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("On-chain tx: %s", txHash)

    // 6. Transfer
    _, err = client.Transfer(ctx, recipient, "usdc", decimal.NewFromInt(5))
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Transferred 5 USDC")

    // 7. Close channel — two steps: prepare finalize state, then checkpoint on-chain
    _, err = client.CloseHomeChannel(ctx, "usdc")
    if err != nil {
        log.Fatal(err)
    }
    closeTx, err := client.Checkpoint(ctx, "usdc")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Channel closed, tx: %s", closeTx)
}
` + "```" + `

## Environment Variables

- ` + "`PRIVATE_KEY`" + ` — Hex private key (without 0x prefix)
- ` + "`CLEARNODE_URL`" + ` — WebSocket URL
- ` + "`RPC_URL`" + ` — Ethereum RPC endpoint
- ` + "`RECIPIENT`" + ` — Recipient address
`

const goAppSessionExampleContent = `# Complete Go App Session Script

` + "```go" + `
package main

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/layer-3/nitrolite/pkg/app"
    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    privateKey := os.Getenv("PRIVATE_KEY")
    clearnodeURL := os.Getenv("CLEARNODE_URL")
    rpcURL := os.Getenv("RPC_URL")
    peerAddr := os.Getenv("PEER_ADDRESS")
    var chainID uint64 = 11155111

    stateSigner, _ := sign.NewEthereumMsgSigner(privateKey)
    txSigner, _ := sign.NewEthereumRawSigner(privateKey)

    client, err := sdk.NewClient(clearnodeURL, stateSigner, txSigner,
        sdk.WithBlockchainRPC(chainID, rpcURL),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    myAddr := client.GetUserAddress()

    // 1. Create app session
    definition := app.AppDefinitionV1{
        ApplicationID: "my-game",
        Participants: []app.AppParticipantV1{
            {WalletAddress: myAddr, SignatureWeight: 50},
            {WalletAddress: peerAddr, SignatureWeight: 50},
        },
        Quorum: 100,
        Nonce:  1,
    }

    // Collect quorum signatures from participants (off-band signing)
    quorumSigs := []string{"0xMySignature...", "0xPeerSignature..."}

    sessionID, versionStr, status, err := client.CreateAppSession(ctx, definition, "{}", quorumSigs)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Session %s created (version: %s, status: %s)", sessionID, versionStr, status)

    // Parse the initial version — updates must increment from here
    initVersion, _ := strconv.ParseUint(versionStr, 10, 64)

    // 2. Submit state update (version = initial + 1)
    update := app.AppStateUpdateV1{
        AppSessionID: sessionID,
        Intent:       app.AppStateUpdateIntentOperate,
        Version:      initVersion + 1,
        Allocations: []app.AppAllocationV1{
            {Participant: myAddr, Asset: "usdc", Amount: decimal.NewFromInt(15)},
            {Participant: peerAddr, Asset: "usdc", Amount: decimal.NewFromInt(5)},
        },
    }
    operateSigs := []string{"0xMySig...", "0xPeerSig..."}
    err = client.SubmitAppState(ctx, update, operateSigs)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("State updated")

    // 3. Close session — submit with Close intent (version = initial + 2)
    closeUpdate := update
    closeUpdate.Intent = app.AppStateUpdateIntentClose
    closeUpdate.Version = initVersion + 2
    closeSigs := []string{"0xMyCloseSig...", "0xPeerCloseSig..."}
    err = client.SubmitAppState(ctx, closeUpdate, closeSigs)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Session closed")
}
` + "```" + `
`

// ---------------------------------------------------------------------------
// Scaffold templates
// ---------------------------------------------------------------------------

const goScaffoldTransfer = `package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
	txSigner, _ := sign.NewEthereumRawSigner(os.Getenv("PRIVATE_KEY"))

	client, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
		sdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Deposit + checkpoint on-chain
	_, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(10))
	if err != nil {
		log.Fatal(err)
	}
	txHash, err := client.Checkpoint(ctx, "usdc")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deposited 10 USDC, tx: %s", txHash)

	// Transfer
	_, err = client.Transfer(ctx, os.Getenv("RECIPIENT"), "usdc", decimal.NewFromInt(5))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Transferred 5 USDC")
}
`

const goScaffoldAppSession = `package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
	txSigner, _ := sign.NewEthereumRawSigner(os.Getenv("PRIVATE_KEY"))

	client, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
		sdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	myAddr := client.GetUserAddress()
	peer := os.Getenv("PEER_ADDRESS")

	def := app.AppDefinitionV1{
		ApplicationID: "my-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: myAddr, SignatureWeight: 50},
			{WalletAddress: peer, SignatureWeight: 50},
		},
		Quorum: 100,
		Nonce:  1,
	}

	// Collect quorum signatures from participants (off-band signing)
	quorumSigs := []string{"0xMySig...", "0xPeerSig..."}

	sessionID, versionStr, _, err := client.CreateAppSession(ctx, def, "{}", quorumSigs)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Session created: %s", sessionID)

	// Parse initial version — updates must increment from here
	initVersion, _ := strconv.ParseUint(versionStr, 10, 64)

	update := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      initVersion + 1,
		Allocations: []app.AppAllocationV1{
			{Participant: myAddr, Asset: "usdc", Amount: decimal.NewFromInt(12)},
			{Participant: peer, Asset: "usdc", Amount: decimal.NewFromInt(8)},
		},
	}
	operateSigs := []string{"0xMySig...", "0xPeerSig..."}
	if err := client.SubmitAppState(ctx, update, operateSigs); err != nil {
		log.Fatal(err)
	}
	log.Println("State updated")

	// Close session — submit with Close intent (version incremented)
	update.Intent = app.AppStateUpdateIntentClose
	update.Version = initVersion + 2
	closeSigs := []string{"0xMyCloseSig...", "0xPeerCloseSig..."}
	if err := client.SubmitAppState(ctx, update, closeSigs); err != nil {
		log.Fatal(err)
	}
	log.Println("Session closed")
}
`

const goScaffoldAIAgent = `package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("AGENT_PRIVATE_KEY"))
	txSigner, _ := sign.NewEthereumRawSigner(os.Getenv("AGENT_PRIVATE_KEY"))

	client, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner,
		sdk.WithBlockchainRPC(11155111, os.Getenv("RPC_URL")),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down agent...")
		client.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Deposit funds for the agent
	_, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(50))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Agent funded with 50 USDC")

	// Payment loop
	recipients := []string{"0x1111...", "0x2222..."}
	for _, r := range recipients {
		_, err := client.Transfer(ctx, r, "usdc", decimal.NewFromFloat(0.10))
		if err != nil {
			log.Printf("Payment to %s failed: %v", r, err)
			continue
		}
		log.Printf("Paid 0.10 USDC to %s", r)
	}

	log.Println("Agent payments complete")
	<-client.WaitCh()
}
`
