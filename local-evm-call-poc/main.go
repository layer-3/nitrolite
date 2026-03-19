// local-evm-call-poc demonstrates executing a pure Solidity function both
// via a real JSON-RPC eth_call and via a standalone, in-process go-ethereum
// EVM instance, then compares the results.
//
// Usage:
//
//	go run ./local-evm-call-poc \
//	  -rpc   https://sepolia.infura.io/v3/<key> \
//	  -addr  0x<deployed Manipulator address> \
//	  -a     7 \
//	  -b     3
package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

// manipulatorABI is the ABI for the Manipulator contract.
// It contains a single pure function: manipulate(uint256,uint256) → uint256.
const manipulatorABI = `[
  {
    "type": "function",
    "name": "manipulate",
    "inputs": [
      { "name": "a", "type": "uint256" },
      { "name": "b", "type": "uint256" }
    ],
    "outputs": [
      { "name": "result", "type": "uint256" }
    ],
    "stateMutability": "pure"
  }
]`

func main() {
	// ── CLI flags ──────────────────────────────────────────────────────
	rpcURL := flag.String("rpc", os.Getenv("RPC_URL"), "Ethereum JSON-RPC URL (or set RPC_URL env var)")
	addrHex := flag.String("addr", os.Getenv("CONTRACT_ADDR"), "Deployed Manipulator contract address (or set CONTRACT_ADDR env var)")
	argA := flag.String("a", "7", "First uint256 argument")
	argB := flag.String("b", "3", "Second uint256 argument")
	flag.Parse()

	if *rpcURL == "" || *addrHex == "" {
		fmt.Fprintln(os.Stderr, "Error: -rpc and -addr flags (or RPC_URL / CONTRACT_ADDR env vars) are required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the two uint256 arguments.
	a, ok := new(big.Int).SetString(*argA, 10)
	if !ok {
		fatal("invalid value for -a: %s", *argA)
	}
	b, ok := new(big.Int).SetString(*argB, 10)
	if !ok {
		fatal("invalid value for -b: %s", *argB)
	}

	contractAddr := common.HexToAddress(*addrHex)

	// ── Parse the ABI and build calldata ───────────────────────────────
	parsed, err := abi.JSON(strings.NewReader(manipulatorABI))
	if err != nil {
		fatal("parse ABI: %v", err)
	}

	calldata, err := parsed.Pack("manipulate", a, b)
	if err != nil {
		fatal("pack calldata: %v", err)
	}

	// ── Connect to the RPC ─────────────────────────────────────────────
	ctx := context.Background()

	client, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		fatal("dial RPC: %v", err)
	}
	defer client.Close()

	// ── Step 1: Real eth_call ──────────────────────────────────────────
	fmt.Println("Step 1: Executing real eth_call via RPC...")

	rpcStart := time.Now()
	rpcResult, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: calldata,
	}, nil) // nil = latest block
	rpcDuration := time.Since(rpcStart)

	if err != nil {
		fatal("eth_call: %v", err)
	}

	// Decode the uint256 return value.
	rpcValues, err := parsed.Unpack("manipulate", rpcResult)
	if err != nil {
		fatal("unpack RPC result: %v", err)
	}
	rpcAnswer := rpcValues[0].(*big.Int)
	fmt.Printf("  RPC result: %s  (took %s)\n\n", rpcAnswer, rpcDuration)

	// ── Step 2: Fetch the deployed runtime bytecode ────────────────────
	fmt.Println("Step 2: Fetching runtime bytecode from RPC...")

	bytecode, err := client.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		fatal("CodeAt: %v", err)
	}
	if len(bytecode) == 0 {
		fatal("no bytecode at address %s — is the contract deployed?", contractAddr.Hex())
	}
	fmt.Printf("  Bytecode length: %d bytes\n\n", len(bytecode))

	// ── Step 2b: Validate bytecode is self-contained ───────────────────
	fmt.Println("Step 2b: Scanning bytecode for DELEGATECALL...")

	if hasDelegateCall(bytecode) {
		fatal("REJECTED: Contract uses DELEGATECALL (Proxy/Clone/External Library). Immutable logic cannot be guaranteed.")
	}
	fmt.Println("  Validation passed: No DELEGATECALL detected. Logic is self-contained.\n")

	// ── Step 3: Set up a standalone, in-memory EVM ─────────────────────
	fmt.Println("Step 3: Setting up local in-memory EVM...")

	// 3a. Create an in-memory state database.
	//     rawdb.NewMemoryDatabase → triedb.NewDatabase → state.NewDatabase → state.New
	memDB := rawdb.NewMemoryDatabase()
	tdb := triedb.NewDatabase(memDB, nil)
	sdb := state.NewDatabase(tdb, nil)

	stateDB, err := state.New(common.Hash{}, sdb) // empty state root = fresh state
	if err != nil {
		fatal("state.New: %v", err)
	}

	// 3b. Deploy the fetched bytecode into the local state at a dummy address.
	dummyAddr := common.HexToAddress("0x00000000000000000000000000000000DeaDBeef")
	stateDB.CreateAccount(dummyAddr)
	stateDB.SetCode(dummyAddr, bytecode, tracing.CodeChangeUnspecified)

	// 3c. Build minimal block and transaction contexts.
	//     Random must be non-nil so the EVM recognises this as a post-Merge block.
	//     Without it, isMerge is false and Shanghai opcodes (PUSH0, etc.) are disabled.
	dummyRandom := common.Hash{0x01}
	blockCtx := vm.BlockContext{
		CanTransfer: func(_ vm.StateDB, _ common.Address, _ *uint256.Int) bool { return true },
		Transfer:    func(_ vm.StateDB, _, _ common.Address, _ *uint256.Int) {},
		GetHash:     func(_ uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		GasLimit:    30_000_000,
		BlockNumber: big.NewInt(1),
		Time:        uint64(time.Now().Unix()),
		Difficulty:  big.NewInt(0),
		BaseFee:     big.NewInt(0),
		Random:      &dummyRandom,
	}

	// 3d. Create the EVM instance.
	//     Use AllDevChainProtocolChanges so all fork opcodes (PUSH0, etc.) are enabled
	//     from timestamp 0. MainnetChainConfig has specific activation timestamps that
	//     may not align with our dummy block context.
	evmInstance := vm.NewEVM(blockCtx, stateDB, params.AllDevChainProtocolChanges, vm.Config{})

	// Set a minimal transaction context.
	callerAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	evmInstance.SetTxContext(vm.TxContext{
		Origin:   callerAddr,
		GasPrice: uint256.NewInt(0),
	})

	fmt.Println("  Local EVM ready.\n")

	// ── Step 4: Execute the same calldata locally ──────────────────────
	fmt.Println("Step 4: Executing calldata on local EVM...")

	localStart := time.Now()
	ret, _, evmErr := evmInstance.Call(
		callerAddr,  // caller
		dummyAddr,   // contract address in local state
		calldata,    // exact same calldata as the RPC call
		30_000_000,  // gas
		uint256.NewInt(0), // value (no ETH transfer)
	)
	localDuration := time.Since(localStart)

	if evmErr != nil {
		fatal("local EVM call: %v", evmErr)
	}

	// Decode the result the same way.
	localValues, err := parsed.Unpack("manipulate", ret)
	if err != nil {
		fatal("unpack local result: %v", err)
	}
	localAnswer := localValues[0].(*big.Int)
	fmt.Printf("  Local result: %s  (took %s)\n\n", localAnswer, localDuration)

	// ── Step 5: Report ─────────────────────────────────────────────────
	match := rpcAnswer.Cmp(localAnswer) == 0

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("                    RESULTS REPORT")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("  Contract:       %s\n", contractAddr.Hex())
	fmt.Printf("  Function:       manipulate(uint256,uint256)\n")
	fmt.Printf("  Inputs:         a = %s, b = %s\n", a, b)
	fmt.Printf("  Expected:       (a * b) + a = %s\n", new(big.Int).Add(new(big.Int).Mul(a, b), a))
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Printf("  RPC result:     %s\n", rpcAnswer)
	fmt.Printf("  Local result:   %s\n", localAnswer)
	fmt.Printf("  Match:          %t\n", match)
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Printf("  RPC time:       %s\n", rpcDuration)
	fmt.Printf("  Local time:     %s\n", localDuration)

	if rpcDuration > localDuration {
		fmt.Printf("  Speedup:        local is ~%.1fx faster\n", float64(rpcDuration)/float64(localDuration))
	} else {
		fmt.Printf("  Speedup:        RPC is ~%.1fx faster\n", float64(localDuration)/float64(rpcDuration))
	}
	fmt.Println("═══════════════════════════════════════════════════════")

	if !match {
		fmt.Fprintln(os.Stderr, "\n⚠ MISMATCH: RPC and local EVM results differ!")
		os.Exit(1)
	}
}

// hasDelegateCall scans raw EVM bytecode for the DELEGATECALL opcode (0xF4).
// It properly skips over PUSH1..PUSH32 (0x60..0x7F) immediate data bytes so
// that embedded constants (addresses, hashes, etc.) are never misidentified
// as opcodes.
func hasDelegateCall(bytecode []byte) bool {
	const (
		opPUSH1       = 0x60
		opPUSH32      = 0x7F
		opDELEGATECALL = 0xF4
	)

	for i := 0; i < len(bytecode); {
		op := bytecode[i]

		if op == opDELEGATECALL {
			return true
		}

		// PUSH1 (0x60) pushes 1 byte, PUSH2 (0x61) pushes 2 bytes, …, PUSH32 (0x7F) pushes 32 bytes.
		// Skip past the immediate data so we only inspect actual opcodes.
		if op >= opPUSH1 && op <= opPUSH32 {
			dataBytes := int(op-opPUSH1) + 1
			i += 1 + dataBytes // skip the PUSH opcode + its data
		} else {
			i++
		}
	}

	return false
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
