# Suggested Contract Design for Nitrolite Protocol v1

## Context: What Changed from v0.5

The Custody.sol implementation handles a relatively simple protocol: channels can be created, funded, checkpointed, challenged, resized, and closed. All operations happen on a single chain, and the adjudicator is responsible for validating state transitions.

The new Nitrolite protocol introduces significant complexity:

- **Cross-chain state**: Channels now have both home and non-home states
- **Escrow operations**: Two-phase bridging (deposit and withdrawal) with separate lifecycle management
- **Migration**: Channels can move their home chain
- **Multiple entity types**: Channels, escrow deposits, and escrow withdrawals each have their own storage and lifecycle

If you implement this using the Custody.sol pattern (each entrypoint contains its own validation logic), you'll face several problems:

1. **Contract size limits**: The new protocol has many more operations. You'll easily exceed the 24KB limit.
2. **Logic duplication**: Common validations (version checks, allocation consistency, signature verification) will be repeated across 15+ functions.
3. **Maintenance burden**: When you need to fix an invariant check, you'll need to update it in multiple places.
4. **Testing complexity**: Each entrypoint needs its own test suite, making it harder to ensure comprehensive coverage.
5. **Audit difficulty**: Auditors must verify that the same invariants are checked consistently across all entrypoints.

## The Core Insight

All state channel operations are fundamentally the same thing: **advancing from one signed state to another, with side effects**.

Whether you're creating a channel, depositing funds, withdrawing, checkpointing, or closing, the pattern is identical:

1. Validate the transition is legal
2. Calculate what needs to change (fund movements, status updates)
3. Apply those changes
4. Emit an event

The only differences are:
- Which intent-specific validations apply
- Which side effects occur

Instead of scattering this logic across many functions, you can centralize it into a unified transition function.

## The Proposed Architecture

### Single-node deployment model

Each `ChannelHub` deployment serves exactly **one node**, identified by the `NODE` immutable set at construction time. The protocol design is node-agnostic (any address can be a node), but this implementation intentionally restricts one hub to one node. Operators who wish to run multiple nodes must deploy a separate `ChannelHub` per node.

### High-Level Structure

The main contract, `ChannelHub.sol`, remains the single source of storage and fund transfers. It contains:

- All storage mappings (channels, escrow deposits, escrow withdrawals, node balances)
- All public entrypoints (thin wrappers of 10-15 lines each)
- Fund transfer helpers (pull/push functions)

The business logic is extracted into library functions organized by entity type:

- `ChannelEngine.validateTransition()` handles all channel operations (CREATE, DEPOSIT, WITHDRAW, OPERATE, CLOSE)
- `EscrowDepositEngine.validateTransition()` handles escrow deposit operations (INITIATE, CHALLENGE, FINALIZE)
- `EscrowWithdrawalEngine.validateTransition()` handles escrow withdrawal operations (INITIATE, CHALLENGE, FINALIZE)

These libraries are deployed as separate contracts with external/public functions, meaning they don't count toward the main contract's 24KB size limit. The main contract calls them via DELEGATECALL, executing their logic in the context of the main contract's storage.

### Why One Function Per Entity Type?

You might ask: why not one giant `validateTransition()` function for everything?

The answer is conceptual boundaries. Channels, escrow deposits, and escrow withdrawals are different entities with:

- Different storage locations (different mappings)
- Different lifecycles (channels operate continuously, escrows are temporary)
- Different context requirements (escrows have unlock expiry, channels don't)

Mixing them in one function would create a 1000+ line monstrosity that's hard to understand and audit. By separating them, each unified function is 200-300 lines and handles one coherent entity type.

### How It Works: The Unified Transition Pattern

Every public entrypoint follows the same mechanical pattern:

1. **Build context**: Gather current state from storage (status, previous state, locked funds, etc.)
2. **Validate and calculate**: Call the unified transition function, which returns the effects to apply
3. **Verify signatures**: Check that both parties signed the state
4. **Apply effects**: Execute fund transfers and storage updates based on the calculated effects
5. **Emit event**: Log the operation

The unified transition function itself has three phases:

**Phase 1: Universal validation** - Checks that apply to ALL transitions regardless of intent:
- Version must increase
- Chain ID must match
- Allocations must equal net flows
- State structure must be valid

**Phase 2: Intent-specific calculation** - Routes based on the intent field and calculates effects:
- For CREATE: requires version 0, empty non-home state, calculates user deposit amount
- For DEPOSIT: requires positive user delta, calculates both user and node fund movements
- For WITHDRAW: requires negative user delta, calculates fund release
- For OPERATE: requires zero user delta, used for checkpointing
- For CLOSE: calculates final payouts

**Phase 3: Universal invariants** - Verifies that the calculated effects maintain consistency:
- Locked funds won't go negative
- Allocations equal locked funds after all movements
- Node has sufficient balance if locking more funds

The function returns a `TransitionEffects` struct containing all the changes to apply, without actually applying them. This separation enables pure functions (no storage access) that are easier to test and can be called off-chain for simulation.

### Example: How Channel Deposit Works

In Custody.sol, the deposit logic is embedded in the entrypoint function along with validation, state updates, and fund transfers.

In the new design:

The public `depositToChannel()` function is thin (about 10 lines). It builds a context struct containing the current channel status, previous state, locked funds, and node available balance. It calls `ChannelEngine.validateTransition()` passing this context and the candidate state. The engine validates the transition, confirms the user delta is positive, checks the node has sufficient funds, and returns effects indicating how much to pull from user and lock from node vault. The public function then verifies signatures, applies the effects (pull funds, update storage, adjust balances), and emits the event.

This means all the business logic lives in `ChannelEngine.validateTransition()`, which handles CREATE, DEPOSIT, WITHDRAW, OPERATE, and CLOSE in one place. The public functions are just mechanical wrappers that build context, call the engine, and apply effects.

### What About Escrow Operations?

Escrow deposits and withdrawals follow the exact same pattern, but with their own unified functions.

For example, `initiateEscrowDeposit()` builds an escrow context, calls `EscrowDepositEngine.validateTransition()`, which validates that this is an initiate intent, checks the user is locking funds on the non-home chain, calculates the unlock expiry timestamp, and returns effects. The public function applies these effects to the escrow storage.

Later, `finalizeEscrowDeposit()` builds context with the current escrow state, calls the same `EscrowDepositEngine.validateTransition()` function (now with a finalize intent), which validates the swap is correct and returns different effects. The public function applies those effects.

The key is that all escrow deposit logic—whether initiate, challenge, or finalize—flows through one validation function.

## Comparison with Custody.sol

### What's Similar

- Single main contract holding all storage
- Public functions are the API surface
- No proxy/upgrade complexity
- Fund transfers handled by internal helpers

### What's Different

**Custody.sol approach**: Each public function (`create`, `join`, `close`, `challenge`, `checkpoint`, `resize`) contains its own validation logic. Common checks are repeated. The adjudicator is called to validate state transitions.

**New approach**: Each public function is a thin wrapper. All validation logic for an entity type is consolidated in one pure function. No external adjudicator—business logic is self-contained in the engine libraries.

**Custody.sol**: Business logic is scattered across multiple functions. To understand what validations apply to a state transition, you must read the specific function handling that intent.

**New approach**: Business logic is co-located in the engine. To understand what validations apply, you read one function with intent-based routing.

**Custody.sol**: Testing requires separate test cases for each public function. Fuzzing is harder because you must fuzz each entrypoint independently.

**New approach**: Testing focuses on the engine functions. You can fuzz `ChannelEngine.validateTransition()` with arbitrary context and candidate state, automatically testing all intents. The public functions are so thin they're almost trivial to test.

**Custody.sol**: Invariants (like allocation consistency) must be checked in every function. Easy to miss in one place.

**New approach**: Invariants are checked in one place (the "universal invariants" phase), impossible to miss.

## Benefits of This Design

### Testability

The unified transition functions are pure—they take memory structs and return memory structs without accessing storage. This means you can unit test them in isolation without deploying contracts or setting up complex state.

You can write comprehensive fuzz tests that generate random contexts and candidate states, call the validation function, and verify invariants hold. This is much harder with Custody.sol's approach where each function accesses storage.

For example, you can fuzz test `ChannelEngine.validateTransition()` with thousands of random inputs covering CREATE, DEPOSIT, WITHDRAW, OPERATE, and CLOSE intents, verifying that allocations always equal locked funds, versions always increase, and no transition causes negative balances.

### Maintainability

When you need to add a new validation rule or fix an invariant check, you modify one function instead of ten. When you add a new intent, you add one branch instead of creating a new public function with duplicated validation logic.

All channel-related business logic lives in `ChannelEngine`, all escrow deposit logic in `EscrowDepositEngine`. This makes the codebase much easier to navigate and understand.

### Auditability

Auditors can focus on three critical functions (`ChannelEngine.validateTransition`, `EscrowDepositEngine.validateTransition`, `EscrowWithdrawalEngine.validateTransition`) knowing that all business logic flows through them.

They can verify that universal validations (version, allocations, signatures) are consistently applied to all intents. They can trace through each intent branch to understand the specific rules.

In Custody.sol, auditors must read every public function and verify that common validations are applied consistently across all of them.

### Off-Chain Simulation

Because the validation functions are pure (they don't modify state) and return effects without applying them, you can call them off-chain to preview what a transaction will do before submitting it.

Your off-chain systems can build a context, create a candidate state, and call `ChannelEngine.validateTransition()` as a view function (since external library functions can be called directly as well as via DELEGATECALL). This shows exactly what fund movements and state changes will occur. This is valuable for UX and for catching errors before spending gas.

Note that while the library functions are called via DELEGATECALL when invoked from the main contract (so they execute in the context of the main contract's storage), they can also be called directly from off-chain code for simulation purposes.

### Gas Efficiency and Contract Size Trade-offs

When using libraries in Solidity, there's an important distinction between internal and external library functions that affects both gas costs and contract size:

**Internal library functions** are embedded directly into your contract bytecode at compile time. They use JUMP instructions and have lower gas costs (no external call overhead), but they DO count toward your contract's 24KB size limit. The Solidity compiler only includes the internal functions you actually use, so unused functions are excluded, but all used functions add to your contract size.

**External or public library functions** are deployed as separate contracts and called via DELEGATECALL. They have a higher gas cost (approximately 700 additional gas per call for the DELEGATECALL), but they DON'T count toward your contract's 24KB size limit. The library is deployed once and can be shared across multiple contracts.

For the Nitrolite protocol, you should use **external functions** in your engine libraries. Here's why:

The cross-chain validation logic will be extensive (hundreds of lines for channel operations, escrow deposits, and escrow withdrawals). If you embedded this as internal functions, you would quickly hit the 24KB limit. By using external functions, the engine logic lives in separate library contracts and doesn't bloat your main `ChannelHub` contract.

The gas overhead of approximately 700 gas per DELEGATECALL is negligible compared to:
- The cost of fund transfers (21,000+ gas for ETH transfers)
- Storage updates (5,000-20,000 gas per slot)
- The actual ERC20 token transfers

The benefits of avoiding contract size issues, maintaining testability, and enabling code reuse far outweigh the minor gas cost increase.

**Important**: When implementing the engine libraries, make sure to declare the `validateTransition()` functions as `external` or `public`, not `internal`. This ensures they're deployed separately and don't count toward your main contract's size limit.

## Testing Strategy

Your test suite becomes much more powerful with this design:

**Unit tests** focus on the engine libraries. You test `ChannelEngine.validateTransition()` with various contexts and candidate states, verifying that:
- Invalid transitions revert with appropriate errors
- Valid transitions return correct effects
- Invariants always hold
- Edge cases (zero amounts, boundary values) are handled

**Fuzz tests** generate random inputs and verify properties:
- Allocations always equal locked funds after transitions
- Versions are monotonically increasing
- No transition causes negative balances
- Effects sum to zero (conservation of funds)

**Integration tests** verify the full flow with actual storage:
- Create a channel, deposit, checkpoint, withdraw, close
- Verify storage is updated correctly
- Verify funds move as expected
- Verify events are emitted

**Scenario tests** cover cross-chain flows:
- Initiate escrow deposit on non-home chain, finalize on home chain
- Challenge during escrow, resolve with newer state
- Migrate channel from one home chain to another

The unified design makes all of these easier to write and more comprehensive.

## Why This is Better Than Alternatives

**Alternative 1: Monolithic contract with per-entrypoint logic (Custody.sol style)**
- Hits contract size limits quickly
- Duplicates validation logic
- Harder to test comprehensively
- Invariants can be missed

**Alternative 2: Diamond pattern with facets**
- Much more complexity (proxy, storage slots, facet routing)
- Harder to audit (must understand diamond standard)
- Gas overhead from proxy calls
- No clear benefit over libraries for this use case

**Alternative 3: Multiple separate contracts**
- Must handle cross-contract calls
- Shared storage patterns become complex
- No clear entity boundaries
- Harder to ensure atomic operations

**Proposed approach: Unified transition functions in external libraries**
- Clean conceptual model
- Maximum testability
- Small gas overhead (~700 gas per DELEGATECALL)
- No contract size issues (external functions don't count toward 24KB limit)
- Single source of truth per entity
- Libraries can be shared across multiple contracts

## Conclusion

The unified transition function approach is a natural fit for state channel protocols because fundamentally, everything is a state transition with different intents and effects.

By centralizing the validation logic in pure library functions organized by entity type, you get a codebase that is easier to implement, test, audit, and maintain than the Custody.sol approach.

The public contract remains simple—just storage, thin entrypoint wrappers, and fund transfer helpers. All the complex business logic lives in the engines where it can be thoroughly tested in isolation.

This design scales to handle the cross-chain complexity of the new protocol without becoming unwieldy, while maintaining the benefits of a monolithic storage contract.
