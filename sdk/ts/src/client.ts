/**
 * Main Nitrolite SDK Client
 * Provides a unified interface for interacting with Nitrolite payment channels.
 * Combines both high-level operations (Deposit, Withdraw, Transfer) and
 * low-level RPC access for advanced use cases.
 */

import { Address, Hex, createPublicClient, createWalletClient, http, custom, verifyMessage } from 'viem';
import { Decimal } from 'decimal.js';
import * as core from './core/index.js';
import * as app from './app/index.js';
import * as API from './rpc/api.js';
import { StateV1, ChannelDefinitionV1, ChannelSessionKeyStateV1, AppV1, AppInfoV1 } from './rpc/types.js';
import { RPCClient } from './rpc/client.js';
import { WebsocketDialer } from './rpc/dialer.js';
import { ClientAssetStore } from './asset_store.js';
import { appendApplicationIDQueryParam, Config, DefaultConfig, Option } from './config.js';
import {
  generateNonce,
  transformNodeConfig,
  transformAssets,
  transformBalances,
  transformChannel,
  transformState,
  transformTransaction,
  transformPaginationMetadata,
  transformAppDefinitionToRPC,
  transformAppStateUpdateToRPC,
  transformSignedAppStateUpdateToRPC,
  transformAppSessionInfo,
  transformAppDefinitionFromRPC,
  transformActionAllowance,
} from './utils.js';
import * as blockchain from './blockchain/index.js';
import { nextState, applyChannelCreation, applyAcknowledgementTransition, applyHomeDepositTransition, applyHomeWithdrawalTransition, applyTransferSendTransition, applyFinalizeTransition, applyCommitTransition } from './core/state.js';
import { newVoidState } from './core/types.js';
import { packState, packChallengeState } from './core/state_packer.js';
import { StateSigner, TransactionSigner } from './signers.js';

/**
 * Default challenge period for channels (1 day in seconds)
 */
export const DEFAULT_CHALLENGE_PERIOD = 86400;

/**
 * Strip the channel signer type prefix byte from a signature.
 * Session key registration requires a raw EIP-191 wallet signature,
 * so the ChannelDefaultSigner's 0x00 prefix must be removed.
 * Throws if the signature doesn't start with the expected 0x00 prefix.
 */
function stripSignerTypePrefix(sig: Hex): Hex {
  if (sig.length < 6) {
    throw new Error(`signature too short to contain a signer type prefix: ${sig}`);
  }
  const prefixByte = parseInt(sig.slice(2, 4), 16);
  if (prefixByte !== core.ChannelSignerType.Default) {
    throw new Error(
      `expected ChannelDefaultSigner prefix 0x00, got 0x${prefixByte.toString(16).padStart(2, '0')}; ` +
      `session key signing requires the default wallet signer, not a session key signer`
    );
  }
  return `0x${sig.slice(4)}` as Hex;
}

// Re-export signer interfaces for convenience
export type { StateSigner, TransactionSigner };

/**
 * Client provides a unified interface for interacting with Nitrolite.
 * It combines both high-level operations (Deposit, Withdraw, Transfer) and
 * low-level RPC access for advanced use cases.
 *
 * @example
 * ```typescript
 * import { Client, withBlockchainRPC } from '@nitrolite/sdk';
 * import { privateKeyToAccount } from 'viem/accounts';
 *
 * // Create signers
 * const account = privateKeyToAccount('0x...');
 * const stateSigner = new WalletStateSigner(walletClient);
 *
 * // Create client
 * const client = await Client.create(
 *   'wss://node.nitrolite.com/ws',
 *   stateSigner,
 *   walletClient,
 *   withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
 * );
 *
 * // High-level operations
 * const state = await client.deposit(80002n, 'usdc', new Decimal(100));
 * const txHash = await client.checkpoint('usdc');
 * const transferState = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));
 *
 * // Low-level operations
 * const config = await client.getConfig();
 * const balances = await client.getBalances('0x1234...');
 * ```
 */
export class Client {
  private rpcClient: RPCClient;
  private config: Config;
  private exitPromise: Promise<void>;
  private exitResolve?: () => void;
  private blockchainClients: Map<bigint, blockchain.evm.Client>;
  private blockchainLockingClients: Map<bigint, blockchain.evm.LockingClient>;
  private homeBlockchains: Map<string, bigint>;
  private stateSigner: StateSigner;
  private txSigner: TransactionSigner;
  private assetStore: ClientAssetStore;
  private stateAdvancer: core.StateAdvancerV1;

  private constructor(
    rpcClient: RPCClient,
    config: Config,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    assetStore: ClientAssetStore
  ) {
    this.rpcClient = rpcClient;
    this.config = config;
    this.stateSigner = stateSigner;
    this.txSigner = txSigner;
    this.assetStore = assetStore;
    this.blockchainClients = new Map();
    this.blockchainLockingClients = new Map();
    this.homeBlockchains = new Map();
    this.stateAdvancer = new core.StateAdvancerV1(assetStore);

    // Create exit promise
    this.exitPromise = new Promise((resolve) => {
      this.exitResolve = resolve;
    });
  }

  /**
   * Create a new Nitrolite client with both high-level and low-level methods.
   * This is the recommended constructor for most use cases.
   *
   * @param wsURL - WebSocket URL of the Nitrolite server (e.g., "wss://clearnode-sandbox.yellow.org/v1/ws")
   * @param stateSigner - Signer for signing channel states (EthereumMsgSigner)
   * @param txSigner - Signer for blockchain transactions (EthereumRawSigner)
   * @param opts - Optional configuration (withBlockchainRPC, withHandshakeTimeout, etc.)
   * @returns Configured Client ready for operations
   *
   * @example
   * ```typescript
   * import { createSigners } from '@nitrolite/sdk';
   *
   * const { stateSigner, txSigner } = createSigners('0x...');
   * const client = await Client.create(
   *   'wss://node.nitrolite.com/ws',
   *   stateSigner,
   *   txSigner,
   *   withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
   * );
   * ```
   */
  static async create(
    wsURL: string,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    ...opts: Option[]
  ): Promise<Client> {
    // Build config starting with defaults
    const config: Config = {
      url: wsURL,
      handshakeTimeout: DefaultConfig.handshakeTimeout,
      pingInterval: DefaultConfig.pingInterval,
      errorHandler: DefaultConfig.errorHandler,
      blockchainRPCs: DefaultConfig.blockchainRPCs || new Map(),
    };

    // Apply user options
    for (const opt of opts) {
      opt(config);
    }

    // Create WebSocket dialer
    const dialer = new WebsocketDialer();
    const rpcClient = new RPCClient(dialer);

    // Declare client variable for use in asset store
    let client: Client;

    // Create asset store (will be initialized with client methods)
    const assetStore = new ClientAssetStore(async () => {
      return await client.getAssets();
    });

    // Create client instance
    client = new Client(rpcClient, config, stateSigner, txSigner, assetStore);

    // Error handler wrapper
    const handleError = (err?: Error) => {
      if (err && config.errorHandler) {
        config.errorHandler(err);
      }
      client.exitResolve?.();
    };

    // Establish connection (append app_id query param if configured)
    const dialURL = appendApplicationIDQueryParam(wsURL, config.applicationID);
    await rpcClient.start(dialURL, handleError);

    return client;
  }

  // ============================================================================
  // Home Blockchain Management
  // ============================================================================

  /**
   * SetHomeBlockchain configures the primary blockchain network for a specific asset.
   * This is required for operations like Transfer which may trigger channel creation
   * but do not accept a blockchain ID as a parameter.
   *
   * @param asset - The asset symbol (e.g., "usdc")
   * @param blockchainId - The chain ID to associate with the asset (e.g., 80002n)
   *
   * @example
   * ```typescript
   * // Set USDC to settle on Polygon Amoy
   * await client.setHomeBlockchain('usdc', 80002n);
   * ```
   */
  async setHomeBlockchain(asset: string, blockchainId: bigint): Promise<void> {
    const existingBlockchainId = this.homeBlockchains.get(asset);
    if (existingBlockchainId !== undefined) {
      throw new Error(
        `home blockchain is already set for asset ${asset} to ${existingBlockchainId}, please use Migrate() if you want to change home blockchain`
      );
    }

    const exists = await this.assetStore.assetExistsOnBlockchain(blockchainId, asset);
    if (!exists) {
      throw new Error(`asset ${asset} not supported on blockchain ${blockchainId}`);
    }

    this.homeBlockchains.set(asset, blockchainId);
  }

  // ============================================================================
  // Connection & Lifecycle Methods
  // ============================================================================

  /**
   * Close cleanly shuts down the client connection.
   * It's recommended to call this when done using the client.
   *
   * @example
   * ```typescript
   * await client.close();
   * ```
   */
  async close(): Promise<void> {
    this.exitResolve?.();
  }

  /**
   * WaitForClose returns a promise that resolves when the connection is lost or closed.
   * This is useful for monitoring connection health in long-running applications.
   *
   * @example
   * ```typescript
   * client.waitForClose().then(() => {
   *   console.log('Connection closed');
   * });
   * ```
   */
  waitForClose(): Promise<void> {
    return this.exitPromise;
  }

  // ============================================================================
  // Shared Helper Methods
  // ============================================================================

  /**
   * SignState signs a channel state by packing it, hashing it, and signing the hash.
   * Returns the signature as a hex-encoded string (with 0x prefix).
   *
   * This is a low-level method exposed for advanced users who want to manually
   * construct and sign states. Most users should use the high-level methods like
   * transfer, deposit, and withdraw instead.
   */
  async signState(state: core.State): Promise<Hex> {
    // Pack the state into ABI-encoded bytes
    const packed = await packState(state, this.assetStore);

    // Sign the packed state using the state signer (adds Ethereum message prefix and hashes internally)
    const signature = await this.stateSigner.signMessage(packed);

    return signature;
  }

  /**
   * GetUserAddress returns the Ethereum address associated with the signer.
   * This is useful for identifying the current user's wallet address.
   */
  getUserAddress(): Address {
    return this.stateSigner.getAddress();
  }

  /**
   * ValidateAndSignState validates that the proposed state is a valid advancement of the
   * current state, then signs it. Returns the signature as a hex-encoded string (with 0x prefix).
   *
   * This is a low-level method exposed for advanced users who want to manually
   * construct and sign states. Most users should use the high-level methods like
   * transfer, deposit, and withdraw instead.
   */
  async validateAndSignState(currentState: core.State, proposedState: core.State): Promise<Hex> {
    await this.stateAdvancer.validateAdvancement(currentState, proposedState);
    return this.signState(proposedState);
  }

  /**
   * SignAndSubmitState is a helper that validates, signs a state and submits it to the node.
   * Returns the node's signature.
   */
  private async signAndSubmitState(currentState: core.State, proposedState: core.State): Promise<Hex> {
    // Validate and sign state
    const sig = await this.validateAndSignState(currentState, proposedState);
    proposedState.userSig = sig;

    // Submit to node
    const nodeSig = await this.submitState(proposedState);

    // Update state with node signature
    proposedState.nodeSig = nodeSig as Hex;

    return nodeSig as Hex;
  }

  // ============================================================================
  // High-Level Operations
  // ============================================================================

  /**
   * Deposit prepares a deposit state for the user's channel.
   * This method handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel with the initial deposit
   * 2. If channel exists: Advances the state with a deposit transition
   *
   * The returned state is signed by both the user and the node, but has not yet been
   * submitted to the blockchain. Use {@link checkpoint} to execute the on-chain transaction.
   *
   * @param blockchainId - The blockchain network ID (e.g., 80002n for Polygon Amoy)
   * @param asset - The asset symbol to deposit (e.g., "usdc")
   * @param amount - The amount to deposit
   * @returns The co-signed state ready for on-chain checkpoint
   *
   * @example
   * ```typescript
   * const state = await client.deposit(80002n, 'usdc', new Decimal(100));
   * const txHash = await client.checkpoint('usdc');
   * console.log('Deposit transaction:', txHash);
   * ```
   */
  async deposit(blockchainId: bigint, asset: string, amount: Decimal): Promise<core.State> {
    const userWallet = this.getUserAddress();

    // Get node address
    const nodeAddress = await this.getNodeAddress();
    if (!nodeAddress) {
      throw new Error('node address is undefined - ensure node config is properly loaded');
    }

    // Get token address for this asset on this blockchain
    const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
    if (!tokenAddress) {
      throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
    }

    // Try to get latest state to determine if channel exists
    let state: core.State | null = null;
    let channelIsOpen = false;
    try {
      state = await this.getLatestState(userWallet, asset, false);

      // If state has a home channel ID, check if it's usable
      if (state && state.homeChannelId) {
        // Check if state has a finalize transition (channel is being closed)
        const hasFinalize = state.transition.type === core.TransitionType.Finalize;
        // If no finalize transition, channel is still open and usable
        channelIsOpen = !hasFinalize;
      }
    } catch (err) {
      // Channel doesn't exist, will create it
    }

    // Scenario A: Channel doesn't exist or is closed - create it
    if (!state || !channelIsOpen) {
      // Get supported sig validators bitmap from node config
      const bitmap = await this.getSupportedSigValidatorsBitmap();

      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
        approvedSigValidators: bitmap,
      };

      // homeChannelId is intentionally not checked here: a non-null state with
      // an undefined homeChannelId is valid — it represents a user who received
      // funds but has not yet opened a channel. Only replace with a void state
      // when there is truly no prior state at all.
      if (!state) {
        state = newVoidState(asset, userWallet);
      }
      const newState = nextState(state!);

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyHomeDepositTransition(newState, amount);

      // Sign state
      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      return newState;
    }

    // Scenario B: Channel exists and is open - advance state with deposit
    const newState = nextState(state);
    applyHomeDepositTransition(newState, amount);

    // Sign and submit state to node
    await this.signAndSubmitState(state, newState);

    return newState;
  }

  /**
   * Withdraw prepares a withdrawal state to remove funds from the user's channel.
   * This operation handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel with the withdrawal transition
   * 2. If channel exists: Advances the state with a withdrawal transition
   *
   * The returned state is signed by both the user and the node, but has not yet been
   * submitted to the blockchain. Use {@link checkpoint} to execute the on-chain transaction.
   *
   * @param blockchainId - The blockchain network ID (e.g., 80002n for Polygon Amoy)
   * @param asset - The asset symbol to withdraw (e.g., "usdc")
   * @param amount - The amount to withdraw
   * @returns The co-signed state ready for on-chain checkpoint
   *
   * @example
   * ```typescript
   * const state = await client.withdraw(80002n, 'usdc', new Decimal(25));
   * const txHash = await client.checkpoint('usdc');
   * console.log('Withdrawal transaction:', txHash);
   * ```
   */
  async withdraw(blockchainId: bigint, asset: string, amount: Decimal): Promise<core.State> {
    const userWallet = this.getUserAddress();

    // Get node address
    const nodeAddress = await this.getNodeAddress();
    if (!nodeAddress) {
      throw new Error('node address is undefined - ensure node config is properly loaded');
    }

    // Get token address for this asset on this blockchain
    const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
    if (!tokenAddress) {
      throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
    }

    // Try to get latest state to determine if channel exists
    let state: core.State | null = null;
    let channelIsOpen = false;
    try {
      state = await this.getLatestState(userWallet, asset, false);

      // If state has a home channel ID, check if it's usable
      if (state && state.homeChannelId) {
        // Check if state has a finalize transition (channel is being closed)
        const hasFinalize = state.transition.type === core.TransitionType.Finalize;
        // If no finalize transition, channel is still open and usable
        channelIsOpen = !hasFinalize;
      }
    } catch (err) {
      // Channel doesn't exist, will create it
    }

    // Channel doesn't exist or is closed - create it and withdraw
    if (!state || !channelIsOpen) {
      // Get supported sig validators bitmap from node config
      const bitmap = await this.getSupportedSigValidatorsBitmap();

      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
        approvedSigValidators: bitmap,
      };

      // homeChannelId is intentionally not checked here: a non-null state with
      // an undefined homeChannelId is valid — it represents a user who received
      // funds but has not yet opened a channel. Only replace with a void state
      // when there is truly no prior state at all.
      if (!state) {
        state = newVoidState(asset, userWallet);
      }
      const newState = nextState(state!);

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyHomeWithdrawalTransition(newState, amount);

      // Sign state
      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      return newState;
    }

    // Create next state
    const newState = nextState(state);
    applyHomeWithdrawalTransition(newState, amount);

    // Sign and submit state to node
    await this.signAndSubmitState(state, newState);

    return newState;
  }

  /**
   * Transfer prepares a transfer state to send funds to another wallet address.
   * This method handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel with the transfer transition
   * 2. If channel exists: Advances the state with a transfer send transition
   *
   * The returned state is signed by both the user and the node. For existing channels,
   * no blockchain interaction is needed. For new channels, use {@link checkpoint} to create
   * the channel on-chain.
   *
   * @param recipientWallet - The recipient's wallet address (e.g., "0x1234...")
   * @param asset - The asset symbol to transfer (e.g., "usdc")
   * @param amount - The amount to transfer
   * @returns The co-signed state with the transfer transition applied
   *
   * @example
   * ```typescript
   * const state = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));
   * console.log('Transfer tx ID:', state.transition.txId);
   * ```
   */
  async transfer(recipientWallet: string, asset: string, amount: Decimal): Promise<core.State> {
    const senderWallet = this.getUserAddress();

    // Get sender's latest state
    let state: core.State | null = null;
    try {
      state = await this.getLatestState(senderWallet, asset, false);
    } catch (err) {
      // Channel doesn't exist
    }

    // homeChannelId is intentionally not checked here: a non-null state with
    // an undefined homeChannelId is valid — it represents a user who received
    // funds but has not yet opened a channel. Only enter the creation path when
    // there is truly no prior state at all.
    if (!state) {
      // Get supported sig validators bitmap from node config
      const bitmap = await this.getSupportedSigValidatorsBitmap();

      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
        approvedSigValidators: bitmap,
      };

      if (!state) {
        state = newVoidState(asset, senderWallet);
      }
      const newState = nextState(state!);

      let blockchainId = this.homeBlockchains.get(asset);
      if (!blockchainId) {
        if (state.homeLedger.blockchainId !== 0n) {
          blockchainId = state.homeLedger.blockchainId;
        } else {
          blockchainId = await this.assetStore.getSuggestedBlockchainId(asset);
        }
      }

      // Get node address
      const nodeAddress = await this.getNodeAddress();
      if (!nodeAddress) {
        throw new Error('node address is undefined - ensure node config is properly loaded');
      }

      // Get token address for this asset on this blockchain
      const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
      if (!tokenAddress) {
        throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
      }

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyTransferSendTransition(newState, recipientWallet, amount);

      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      return newState;
    }

    // Create next state
    const newState = nextState(state);
    applyTransferSendTransition(newState, recipientWallet, amount);

    // Sign and submit state
    await this.signAndSubmitState(state, newState);

    return newState;
  }

  /**
   * Acknowledge prepares an acknowledgement state for the given asset.
   * This is used when a user receives a transfer but hasn't yet acknowledged the state,
   * or to acknowledge channel creation without a deposit.
   *
   * This method handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel with the acknowledgement transition
   * 2. If channel exists: Advances the state with an acknowledgement transition
   *
   * The returned state is signed by both the user and the node.
   *
   * @param asset - The asset symbol to acknowledge (e.g., "usdc")
   * @returns The co-signed state with the acknowledgement transition applied
   *
   * @example
   * ```typescript
   * const state = await client.acknowledge('usdc');
   * ```
   */
  async acknowledge(asset: string): Promise<core.State> {
    const userWallet = this.getUserAddress();

    // Try to get latest state to determine if channel exists
    let state: core.State | null = null;
    try {
      state = await this.getLatestState(userWallet, asset, false);
    } catch (err) {
      // No state exists
    }

    // homeChannelId is intentionally not checked here: a non-null state with
    // an undefined homeChannelId is valid — it represents a user who received
    // funds but has not yet opened a channel. Only enter the creation path when
    // there is truly no prior state at all.
    // No channel path - create channel with acknowledgement
    if (!state) {
      // Get supported sig validators bitmap from node config
      const bitmap = await this.getSupportedSigValidatorsBitmap();

      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
        approvedSigValidators: bitmap,
      };

      if (!state) {
        state = newVoidState(asset, userWallet);
      }
      const newState = nextState(state);

      let blockchainId = this.homeBlockchains.get(asset);
      if (!blockchainId) {
        if (state.homeLedger.blockchainId !== 0n) {
          blockchainId = state.homeLedger.blockchainId;
        } else {
          blockchainId = await this.assetStore.getSuggestedBlockchainId(asset);
        }
      }

      const nodeAddress = await this.getNodeAddress();
      if (!nodeAddress) {
        throw new Error('node address is undefined - ensure node config is properly loaded');
      }

      const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
      if (!tokenAddress) {
        throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
      }

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyAcknowledgementTransition(newState);

      const sig = await this.signState(newState);
      newState.userSig = sig;

      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      return newState;
    }

    if (state.userSig) {
      throw new Error('state already acknowledged by user');
    }

    // Has channel path - submit acknowledgement
    const newState = nextState(state);
    applyAcknowledgementTransition(newState);

    await this.signAndSubmitState(state, newState);

    return newState;
  }

  /**
   * CloseHomeChannel prepares a finalize state to close the user's channel for a specific asset.
   * This creates a final state with zero user balance and submits it to the node.
   *
   * The returned state is signed by both the user and the node, but has not yet been
   * submitted to the blockchain. Use {@link checkpoint} to execute the on-chain close.
   *
   * @param asset - The asset symbol to close (e.g., "usdc")
   * @returns The co-signed finalize state ready for on-chain close
   *
   * @example
   * ```typescript
   * const state = await client.closeHomeChannel('usdc');
   * const txHash = await client.checkpoint('usdc');
   * console.log('Close transaction:', txHash);
   * ```
   */
  async closeHomeChannel(asset: string): Promise<core.State> {
    const senderWallet = this.getUserAddress();

    const state = await this.getLatestState(senderWallet, asset, false);

    if (!state.homeChannelId) {
      throw new Error(`no channel exists for asset ${asset}`);
    }

    // Create next state
    const newState = nextState(state);
    applyFinalizeTransition(newState);

    // Sign and submit state
    await this.signAndSubmitState(state, newState);

    return newState;
  }

  /**
   * Checkpoint executes the blockchain transaction for the latest signed state.
   * It fetches the latest co-signed state and, based on the transition type and on-chain
   * channel status, calls the appropriate blockchain method.
   *
   * This is the only method that interacts with the blockchain. It should be called after
   * any state-building method (deposit, withdraw, closeHomeChannel, etc.) to settle
   * the state on-chain. It can also be used as a recovery mechanism if a previous
   * blockchain transaction failed (e.g., due to gas issues or network problems).
   *
   * Blockchain method mapping:
   * - Channel not yet on-chain (status Void): Creates the channel via blockchainClient.create
   * - HomeDeposit/HomeWithdrawal on existing channel: Checkpoints via blockchainClient.checkpoint
   * - Finalize: Closes the channel via blockchainClient.close
   *
   * @param asset - The asset symbol (e.g., "usdc")
   * @returns Transaction hash of the blockchain transaction
   *
   * @example
   * ```typescript
   * const state = await client.deposit(80002n, 'usdc', new Decimal(100));
   * const txHash = await client.checkpoint('usdc');
   * console.log('On-chain transaction:', txHash);
   * ```
   */
  async checkpoint(asset: string): Promise<string> {
    const userWallet = this.getUserAddress();

    // Get latest signed state (both user and node signatures must be present)
    const state = await this.getLatestState(userWallet, asset, true);

    if (!state.homeChannelId) {
      // NOTE: this should never happen, because signed state MUST have a channel ID
      throw new Error(`no channel exists for asset ${asset}`);
    }

    const blockchainId = state.homeLedger.blockchainId;

    // Initialize blockchain client if needed
    await this.initializeBlockchainClient(blockchainId);
    const blockchainClient = this.blockchainClients.get(blockchainId)!;

    // Get home channel info to determine on-chain status
    const channel = await this.getHomeChannel(userWallet, asset);

    switch (state.transition.type) {
      case core.TransitionType.Acknowledgement:
      case core.TransitionType.HomeDeposit:
      case core.TransitionType.HomeWithdrawal:
      case core.TransitionType.TransferSend:
      case core.TransitionType.TransferReceive:
      case core.TransitionType.Commit:
      case core.TransitionType.Release:
      {
        if (channel.status === core.ChannelStatus.Void) {
          // Channel not yet created on-chain, reconstruct definition and call Create
          const channelDef: core.ChannelDefinition = {
            nonce: channel.nonce,
            challenge: channel.challengeDuration,
            approvedSigValidators: channel.approvedSigValidators,
          };
          return await blockchainClient.create(channelDef, state);
        }

        // Checkpoint existing channel
        return await blockchainClient.checkpoint(state);
      }

      case core.TransitionType.Finalize: {
        return await blockchainClient.close(state);
      }

      default:
        throw new Error(
          `transition type ${state.transition.type} does not require a blockchain operation`
        );
    }
  }

  /**
   * Challenge submits an on-chain challenge for a channel using a co-signed state.
   * The state must have both user and node signatures, which are validated before
   * the challenge transaction is submitted.
   *
   * A challenge initiates a dispute period on-chain. If the counterparty does not
   * respond with a higher-versioned state before the challenge period expires,
   * the channel can be closed with the challenged state.
   *
   * @param state - A co-signed state (both userSig and nodeSig must be present)
   * @returns Transaction hash of the on-chain challenge transaction
   *
   * @example
   * ```typescript
   * const state = await client.getLatestState(wallet, 'usdc', true);
   * const txHash = await client.challenge(state);
   * console.log('Challenge transaction:', txHash);
   * ```
   */
  async challenge(state: core.State): Promise<string> {
    if (!state.userSig || !state.nodeSig) {
      throw new Error('state must have both user and node signatures');
    }

    if (!state.homeChannelId) {
      throw new Error('state must have a home channel ID');
    }

    // Pack state for signature verification
    const packedState = await packState(state, this.assetStore);

    // Strip the signer type byte (first byte) from signatures before verification
    const userSigRaw = `0x${state.userSig.slice(4)}` as Hex; // skip 0x + type byte
    const userValid = await verifyMessage({
      address: state.userWallet,
      message: { raw: packedState },
      signature: userSigRaw,
    });
    if (!userValid) {
      throw new Error('invalid user signature');
    }

    const nodeAddress = await this.getNodeAddress();
    const nodeSigRaw = `0x${state.nodeSig.slice(4)}` as Hex;
    const nodeValid = await verifyMessage({
      address: nodeAddress,
      message: { raw: packedState },
      signature: nodeSigRaw,
    });
    if (!nodeValid) {
      throw new Error('invalid node signature');
    }

    // Create the challenge signature
    const challengeData = await packChallengeState(state, this.assetStore);
    const challengerSig = await this.stateSigner.signMessage(challengeData);

    // Initialize blockchain client and submit
    const blockchainId = state.homeLedger.blockchainId;
    await this.initializeBlockchainClient(blockchainId);
    const blockchainClient = this.blockchainClients.get(blockchainId)!;

    return await blockchainClient.challenge(state, challengerSig);
  }

  /**
   * Approve the ChannelHub contract to spend tokens on behalf of the user.
   * This is required before depositing ERC-20 tokens. Native tokens (e.g., ETH)
   * do not require approval and will throw an error if attempted.
   *
   * @param chainId - The blockchain network ID (e.g., 11155111n for Sepolia)
   * @param asset - The asset symbol to approve (e.g., "usdc")
   * @param amount - The amount to approve for spending
   * @returns Transaction hash of the approval transaction
   */
  async approveToken(chainId: bigint, asset: string, amount: Decimal): Promise<string> {
    await this.initializeBlockchainClient(chainId);
    const blockchainClient = this.blockchainClients.get(chainId)!;

    return await blockchainClient.approve(asset, amount);
  }

  /**
   * Query the on-chain token balance (ERC-20 or native ETH) for a wallet on a specific blockchain.
   *
   * @param chainId - The blockchain network ID (e.g., 11155111n for Sepolia)
   * @param asset - The asset symbol to query (e.g., "usdc")
   * @param wallet - The wallet address to query
   * @returns The balance adjusted using token decimals for that chain/token
   */
  async getOnChainBalance(chainId: bigint, asset: string, wallet: Address): Promise<Decimal> {
    await this.initializeBlockchainClient(chainId);
    const blockchainClient = this.blockchainClients.get(chainId)!;

    return await blockchainClient.getTokenBalance(asset, wallet);
  }

  /**
   * Check token allowance for a specific chain and token
   * @param chainId - The blockchain ID
   * @param tokenAddress - The ERC20 token contract address
   * @param owner - The owner address
   * @returns Current allowance amount (in smallest unit)
   */
  async checkTokenAllowance(
    chainId: bigint,
    tokenAddress: string,
    owner: string
  ): Promise<bigint> {
    await this.initializeBlockchainClient(chainId);
    const blockchainClient = this.blockchainClients.get(chainId)!;

    return await blockchainClient.checkAllowanceByAddress(
      tokenAddress as `0x${string}`,
      owner as `0x${string}`
    );
  }

  // ============================================================================
  // Locking On-Chain Methods
  // ============================================================================

  /**
   * Lock tokens into the Locking contract on the specified blockchain.
   * The tokens are locked for the specified target address. Before calling this method,
   * you must approve the Locking contract to spend your tokens using approveSecurityToken.
   *
   * @param targetWalletAddress - The Ethereum address to lock tokens for
   * @param blockchainId - The blockchain network ID
   * @param amount - The amount of tokens to lock (in human-readable decimals, e.g., 100.5 USDC)
   * @returns Transaction hash
   */
  async escrowSecurityTokens(targetWalletAddress: string, blockchainId: bigint, amount: Decimal): Promise<string> {
    await this.initializeLockingClient(blockchainId);
    return this.blockchainLockingClients.get(blockchainId)!.lock(
      targetWalletAddress as Address,
      amount,
    );
  }

  /**
   * Initiate the unlock process for locked tokens in the Locking contract.
   * After the unlock period elapses, withdrawSecurityTokens can be called to retrieve the tokens.
   *
   * @param blockchainId - The blockchain network ID
   * @returns Transaction hash
   */
  async initiateSecurityTokensWithdrawal(blockchainId: bigint): Promise<string> {
    await this.initializeLockingClient(blockchainId);
    return this.blockchainLockingClients.get(blockchainId)!.unlock();
  }

  /**
   * Re-lock tokens that are currently in the unlocking state,
   * cancelling the pending unlock and returning them to the locked state.
   *
   * @param blockchainId - The blockchain network ID
   * @returns Transaction hash
   */
  async cancelSecurityTokensWithdrawal(blockchainId: bigint): Promise<string> {
    await this.initializeLockingClient(blockchainId);
    return this.blockchainLockingClients.get(blockchainId)!.relock();
  }

  /**
   * Withdraw unlocked tokens from the Locking contract to the specified destination.
   * Can only be called after the unlock period has fully elapsed.
   *
   * @param blockchainId - The blockchain network ID
   * @param destinationWalletAddress - The Ethereum address to receive the withdrawn tokens
   * @returns Transaction hash
   */
  async withdrawSecurityTokens(blockchainId: bigint, destinationWalletAddress: string): Promise<string> {
    await this.initializeLockingClient(blockchainId);
    return this.blockchainLockingClients.get(blockchainId)!.withdraw(
      destinationWalletAddress as Address,
    );
  }

  /**
   * Approve the Locking contract to spend tokens on behalf of the caller.
   * This must be called before escrowSecurityTokens.
   *
   * @param chainId - The blockchain network ID
   * @param amount - The amount of tokens to approve
   * @returns Transaction hash
   */
  async approveSecurityToken(chainId: bigint, amount: Decimal): Promise<string> {
    await this.initializeLockingClient(chainId);
    return this.blockchainLockingClients.get(chainId)!.approveToken(amount);
  }

  /**
   * Get the locked balance of a user in the Locking contract.
   *
   * @param chainId - The blockchain network ID
   * @param wallet - The Ethereum address to check
   * @returns The locked balance as a Decimal (adjusted for token decimals)
   */
  async getLockedBalance(chainId: bigint, wallet: string): Promise<Decimal> {
    await this.initializeLockingClient(chainId);
    return this.blockchainLockingClients.get(chainId)!.getBalance(wallet as Address);
  }

  // ============================================================================
  // Node Information Methods
  // ============================================================================

  /**
   * Ping performs a health check on the node connection.
   *
   * @example
   * ```typescript
   * await client.ping();
   * console.log('Node is healthy');
   * ```
   */
  async ping(): Promise<void> {
    await this.rpcClient.nodeV1Ping();
  }

  /**
   * GetConfig retrieves the node configuration including supported blockchains.
   *
   * @returns Node configuration with blockchain list
   *
   * @example
   * ```typescript
   * const config = await client.getConfig();
   * console.log('Node version:', config.nodeVersion);
   * console.log('Supported blockchains:', config.blockchains);
   * ```
   */
  async getConfig(): Promise<core.NodeConfig> {
    const resp = await this.rpcClient.nodeV1GetConfig();
    return transformNodeConfig(resp);
  }

  /**
   * GetBlockchains retrieves the list of supported blockchains.
   *
   * @returns Array of supported blockchains
   *
   * @example
   * ```typescript
   * const blockchains = await client.getBlockchains();
   * for (const chain of blockchains) {
   *   console.log(`${chain.name} (${chain.blockchainId})`);
   * }
   * ```
   */
  async getBlockchains(): Promise<core.Blockchain[]> {
    const config = await this.getConfig();
    return config.blockchains;
  }

  /**
   * GetAssets retrieves the list of supported assets, optionally filtered by blockchain.
   *
   * @param blockchainId - Optional blockchain ID to filter assets
   * @returns Array of supported assets
   *
   * @example
   * ```typescript
   * // Get all assets
   * const allAssets = await client.getAssets();
   *
   * // Get assets for specific blockchain
   * const polygonAssets = await client.getAssets(80002n);
   * ```
   */
  async getAssets(blockchainId?: bigint): Promise<core.Asset[]> {
    const req: API.NodeV1GetAssetsRequest = {};
    if (blockchainId !== undefined) {
      req.blockchain_id = blockchainId;
    }
    const resp = await this.rpcClient.nodeV1GetAssets(req);
    return transformAssets(resp.assets);
  }

  // ============================================================================
  // User Query Methods
  // ============================================================================

  /**
   * GetBalances retrieves the balance information for a user's wallet.
   *
   * @param wallet - The user's wallet address
   * @returns Array of balance entries for each asset
   *
   * @example
   * ```typescript
   * const balances = await client.getBalances('0x1234...');
   * for (const entry of balances) {
   *   console.log(`${entry.asset}: ${entry.balance}`);
   * }
   * ```
   */
  async getBalances(wallet: Address): Promise<core.BalanceEntry[]> {
    const req: API.UserV1GetBalancesRequest = {
      wallet,
    };
    const resp = await this.rpcClient.userV1GetBalances(req);
    return transformBalances(resp.balances);
  }

  /**
   * GetTransactions retrieves the transaction history for a user's wallet.
   *
   * @param wallet - The user's wallet address
   * @param options - Optional filters (asset, pagination)
   * @returns Array of transactions and pagination metadata
   *
   * @example
   * ```typescript
   * const { transactions, metadata } = await client.getTransactions('0x1234...', {
   *   asset: 'usdc',
   *   page: 1,
   *   pageSize: 10,
   * });
   * ```
   */
  async getTransactions(
    wallet: Address,
    options?: {
      asset?: string;
      txType?: core.TransactionType;
      fromTime?: bigint;
      toTime?: bigint;
      page?: number;
      pageSize?: number;
    }
  ): Promise<{ transactions: core.Transaction[]; metadata: core.PaginationMetadata }> {
    const req: API.UserV1GetTransactionsRequest = {
      wallet,
      asset: options?.asset,
      tx_type: options?.txType,
      from_time: options?.fromTime,
      to_time: options?.toTime,
      pagination: options?.page && options?.pageSize ? {
        offset: (options.page - 1) * options.pageSize,
        limit: options.pageSize,
      } : undefined,
    };
    const resp = await this.rpcClient.userV1GetTransactions(req);
    return {
      transactions: resp.transactions.map(transformTransaction),
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  /**
   * GetActionAllowances retrieves the action allowances for a user based on their staking level.
   *
   * @param wallet - The user's wallet address
   * @returns Array of action allowances for each gated action
   *
   * @example
   * ```typescript
   * const allowances = await client.getActionAllowances('0x1234...');
   * for (const a of allowances) {
   *   console.log(`${a.gatedAction}: ${a.used}/${a.allowance} (${a.timeWindow})`);
   * }
   * ```
   */
  async getActionAllowances(wallet: Address): Promise<core.ActionAllowance[]> {
    const req: API.UserV1GetActionAllowancesRequest = { wallet };
    const resp = await this.rpcClient.userV1GetActionAllowances(req);
    return resp.allowances.map(transformActionAllowance);
  }

  // ============================================================================
  // Channel Query Methods
  // ============================================================================

  /**
   * GetChannels retrieves all channels for a user with optional filtering.
   *
   * @param wallet - The user's wallet address
   * @param options - Optional filters: status, asset, pagination
   * @returns Object with channels array and pagination metadata
   *
   * @example
   * ```typescript
   * const result = await client.getChannels('0x1234...');
   * for (const ch of result.channels) {
   *   console.log(`${ch.channelId}: ${ch.status}`);
   * }
   * ```
   */
  async getChannels(
    wallet: Address,
    options?: { status?: string; asset?: string; channelType?: string; pagination?: core.PaginationParams }
  ): Promise<{ channels: core.Channel[]; metadata: core.PaginationMetadata }> {
    const req: API.ChannelsV1GetChannelsRequest = {
      wallet,
      status: options?.status,
      asset: options?.asset,
      channel_type: options?.channelType,
      pagination: options?.pagination
        ? {
            offset: options.pagination.offset,
            limit: options.pagination.limit,
          }
        : undefined,
    };
    const resp = await this.rpcClient.channelsV1GetChannels(req);
    return {
      channels: resp.channels.map(transformChannel),
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  /**
   * GetHomeChannel retrieves home channel information for a user's asset.
   *
   * @param wallet - The user's wallet address
   * @param asset - The asset symbol
   * @returns Channel information for the home channel
   *
   * @example
   * ```typescript
   * const channel = await client.getHomeChannel('0x1234...', 'usdc');
   * console.log(`Channel: ${channel.channelId} (Version: ${channel.stateVersion})`);
   * ```
   */
  async getHomeChannel(wallet: Address, asset: string): Promise<core.Channel> {
    const req: API.ChannelsV1GetHomeChannelRequest = {
      wallet,
      asset,
    };
    const resp = await this.rpcClient.channelsV1GetHomeChannel(req);
    return transformChannel(resp.channel);
  }

  /**
   * GetEscrowChannel retrieves escrow channel information for a specific channel ID.
   *
   * @param escrowChannelId - The escrow channel ID to query
   * @returns Channel information for the escrow channel
   *
   * @example
   * ```typescript
   * const channel = await client.getEscrowChannel('0x1234...');
   * console.log(`Channel: ${channel.channelId} (Version: ${channel.stateVersion})`);
   * ```
   */
  async getEscrowChannel(escrowChannelId: string): Promise<core.Channel> {
    const req: API.ChannelsV1GetEscrowChannelRequest = {
      escrow_channel_id: escrowChannelId,
    };
    const resp = await this.rpcClient.channelsV1GetEscrowChannel(req);
    return transformChannel(resp.channel);
  }

  /**
   * GetLatestState retrieves the latest state for a user's asset.
   *
   * @param wallet - The user's wallet address
   * @param asset - The asset symbol (e.g., "usdc")
   * @param onlySigned - If true, returns only the latest signed state
   * @returns State containing all state information
   *
   * @example
   * ```typescript
   * const state = await client.getLatestState('0x1234...', 'usdc', false);
   * console.log(`Version: ${state.version}, Balance: ${state.homeLedger.userBalance}`);
   * ```
   */
  async getLatestState(wallet: Address, asset: string, onlySigned: boolean): Promise<core.State> {
    const req: API.ChannelsV1GetLatestStateRequest = {
      wallet,
      asset,
      only_signed: onlySigned,
    };
    const resp = await this.rpcClient.channelsV1GetLatestState(req);
    return transformState(resp.state);
  }

  // ============================================================================
  // App Session Methods
  // ============================================================================

  /**
   * GetAppSessions retrieves application sessions for the user.
   *
   * @param options - Optional filters (appSessionId, wallet, status, pagination)
   * @returns Array of app session info and pagination metadata
   *
   * @example
   * ```typescript
   * const { sessions, metadata } = await client.getAppSessions({
   *   wallet: '0x1234...',
   *   status: 'open',
   *   page: 1,
   *   pageSize: 10,
   * });
   * ```
   */
  async getAppSessions(options?: {
    appSessionId?: string;
    wallet?: Address;
    status?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ sessions: app.AppSessionInfoV1[]; metadata: core.PaginationMetadata }> {
    const req: API.AppSessionsV1GetAppSessionsRequest = {
      app_session_id: options?.appSessionId,
      participant: options?.wallet,
      status: options?.status,
      pagination: options?.page && options?.pageSize ? {
        offset: (options.page - 1) * options.pageSize,
        limit: options.pageSize,
      } : undefined,
    };
    const resp = await this.rpcClient.appSessionsV1GetAppSessions(req);
    return {
      sessions: (resp.app_sessions as any[]).map(transformAppSessionInfo),
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  /**
   * GetAppDefinition retrieves the definition for a specific app session.
   *
   * @param appSessionId - The app session ID
   * @returns App session definition
   *
   * @example
   * ```typescript
   * const definition = await client.getAppDefinition('0x1234...');
   * console.log('Participants:', definition.participants);
   * ```
   */
  async getAppDefinition(appSessionId: string): Promise<app.AppDefinitionV1> {
    const req: API.AppSessionsV1GetAppDefinitionRequest = {
      app_session_id: appSessionId,
    };
    const resp = await this.rpcClient.appSessionsV1GetAppDefinition(req);
    return transformAppDefinitionFromRPC(resp.definition);
  }

  /**
   * CreateAppSession creates a new application session between participants.
   *
   * @param definition - The app definition with participants, quorum, application ID
   * @param sessionData - Optional JSON stringified session data
   * @param quorumSigs - Participant signatures for the app session creation
   * @returns Object with appSessionId, version, and status
   *
   * @example
   * ```typescript
   * const definition: app.AppDefinitionV1 = {
   *   application: 'chess-v1',
   *   participants: [
   *     { walletAddress: '0x1234...', signatureWeight: 1 },
   *     { walletAddress: '0x5678...', signatureWeight: 1 },
   *   ],
   *   quorum: 2,
   *   nonce: 1n,
   * };
   * const { appSessionId, version, status } = await client.createAppSession(
   *   definition,
   *   '{}',
   *   ['sig1', 'sig2']
   * );
   * console.log('Created session:', appSessionId);
   * ```
   */
  async createAppSession(
    definition: app.AppDefinitionV1,
    sessionData: string,
    quorumSigs: string[],
    opts?: { ownerSig?: string }
  ): Promise<{ appSessionId: string; version: string; status: string }> {
    const req: API.AppSessionsV1CreateAppSessionRequest = {
      definition: transformAppDefinitionToRPC(definition) as any, // RPC type
      session_data: sessionData,
      quorum_sigs: quorumSigs,
    };
    if (opts?.ownerSig) {
      req.owner_sig = opts.ownerSig;
    }
    const resp = await this.rpcClient.appSessionsV1CreateAppSession(req);
    return {
      appSessionId: resp.app_session_id,
      version: resp.version,
      status: resp.status,
    };
  }

  /**
   * SubmitAppSessionDeposit submits a deposit to an app session.
   * This updates both the app session state and the user's channel state.
   *
   * @param appStateUpdate - The app state update with deposit intent
   * @param quorumSigs - Participant signatures for the app state update
   * @param asset - The asset to deposit
   * @param depositAmount - Amount to deposit
   * @returns Node's signature for the state
   *
   * @example
   * ```typescript
   * const appUpdate: app.AppStateUpdateV1 = {
   *   appSessionId: 'session123',
   *   intent: app.AppStateUpdateIntent.Deposit,
   *   version: 2n,
   *   allocations: [
   *     { participant: '0x1234...', asset: 'usdc', amount: new Decimal(100) },
   *   ],
   *   sessionData: '{}',
   * };
   * const nodeSig = await client.submitAppSessionDeposit(
   *   appUpdate,
   *   ['sig1'],
   *   'usdc',
   *   new Decimal(100)
   * );
   * ```
   */
  async submitAppSessionDeposit(
    appStateUpdate: app.AppStateUpdateV1,
    quorumSigs: string[],
    asset: string,
    depositAmount: Decimal
  ): Promise<string> {
    // Get current state
    const currentState = await this.getLatestState(this.getUserAddress(), asset, false);

    // Create next state with commit transition (use app session ID as account ID)
    const newState = nextState(currentState);
    applyCommitTransition(newState, appStateUpdate.appSessionId, depositAmount);

    // Transform to RPC format after applying the commit transition
    const appUpdate = transformAppStateUpdateToRPC(appStateUpdate);

    // Sign the state
    const stateSig = await this.signState(newState);
    newState.userSig = stateSig;

    // Submit deposit
    const req: API.AppSessionsV1SubmitDepositStateRequest = {
      app_state_update: appUpdate as any, // RPC type
      quorum_sigs: quorumSigs,
      user_state: this.transformStateToRPC(newState),
    };

    const resp = await this.rpcClient.appSessionsV1SubmitDepositState(req);
    return resp.signature;
  }

  /**
   * SubmitAppState submits an app session state update.
   * This method handles operate, withdraw, and close intents.
   * For deposits, use submitAppSessionDeposit instead.
   *
   * @param appStateUpdate - The app state update (intent: operate, withdraw, or close)
   * @param quorumSigs - Participant signatures for the app state update
   *
   * @example
   * ```typescript
   * const appUpdate: app.AppStateUpdateV1 = {
   *   appSessionId: 'session123',
   *   intent: app.AppStateUpdateIntent.Operate,
   *   version: 3n,
   *   allocations: [
   *     { participant: '0x1234...', asset: 'usdc', amount: new Decimal(50) },
   *     { participant: '0x5678...', asset: 'usdc', amount: new Decimal(50) },
   *   ],
   *   sessionData: '{"move": "e4"}',
   * };
   * await client.submitAppState(appUpdate, ['sig1', 'sig2']);
   * ```
   */
  async submitAppState(
    appStateUpdate: app.AppStateUpdateV1,
    quorumSigs: string[]
  ): Promise<void> {
    const appUpdate = transformAppStateUpdateToRPC(appStateUpdate);

    const req: API.AppSessionsV1SubmitAppStateRequest = {
      app_state_update: appUpdate as any, // RPC type
      quorum_sigs: quorumSigs,
    };

    await this.rpcClient.appSessionsV1SubmitAppState(req);
  }

  /**
   * RebalanceAppSessions rebalances multiple application sessions atomically.
   *
   * This method performs atomic rebalancing across multiple app sessions, ensuring
   * that funds are redistributed consistently without the risk of partial updates.
   *
   * @param signedUpdates - Array of signed app state updates to apply atomically
   * @returns BatchID for tracking the rebalancing operation
   *
   * @example
   * ```typescript
   * const updates: app.SignedAppStateUpdateV1[] = [
   *   {
   *     appStateUpdate: { appSessionId: 'session1', intent: app.AppStateUpdateIntent.Rebalance, ... },
   *     quorumSigs: ['sig1', 'sig2'],
   *   },
   *   {
   *     appStateUpdate: { appSessionId: 'session2', intent: app.AppStateUpdateIntent.Rebalance, ... },
   *     quorumSigs: ['sig3', 'sig4'],
   *   },
   * ];
   * const batchId = await client.rebalanceAppSessions(updates);
   * console.log('Rebalance batch ID:', batchId);
   * ```
   */
  async rebalanceAppSessions(
    signedUpdates: app.SignedAppStateUpdateV1[]
  ): Promise<string> {
    // Transform SDK types to RPC types
    const rpcUpdates = signedUpdates.map(transformSignedAppStateUpdateToRPC);

    const req: API.AppSessionsV1RebalanceAppSessionsRequest = {
      signed_updates: rpcUpdates as any, // RPC type
    };

    const resp = await this.rpcClient.appSessionsV1RebalanceAppSessions(req);
    return resp.batch_id;
  }

  // ============================================================================
  // App Registry Methods
  // ============================================================================

  /**
   * GetApps retrieves registered applications with optional filtering.
   *
   * @param options - Optional filters (appId, ownerWallet, pagination)
   * @returns Array of registered apps and pagination metadata
   *
   * @example
   * ```typescript
   * const { apps, metadata } = await client.getApps({ ownerWallet: '0x1234...' });
   * for (const app of apps) {
   *   console.log(`${app.id}: owned by ${app.owner_wallet}`);
   * }
   * ```
   */
  async getApps(options?: {
    appId?: string;
    ownerWallet?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ apps: AppInfoV1[]; metadata: core.PaginationMetadata }> {
    const req: API.AppsV1GetAppsRequest = {
      app_id: options?.appId,
      owner_wallet: options?.ownerWallet,
      pagination: options?.page && options?.pageSize ? {
        offset: (options.page - 1) * options.pageSize,
        limit: options.pageSize,
      } : undefined,
    };
    const resp = await this.rpcClient.appsV1GetApps(req);
    return {
      apps: resp.apps,
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  /**
   * RegisterApp registers a new application in the app registry.
   * Currently only version 1 (creation) is supported.
   *
   * The method builds the app definition from the provided parameters,
   * using the client's signer address as the owner wallet and version 1.
   * It then packs and signs the definition automatically.
   *
   * Session key signers are not allowed to perform this action; the main
   * wallet signer must be used.
   *
   * @param appID - The application identifier
   * @param metadata - The application metadata
   * @param creationApprovalNotRequired - Whether sessions can be created without owner approval
   *
   * @example
   * ```typescript
   * await client.registerApp('my-app', '{"name": "My App"}', false);
   * ```
   */
  async registerApp(appID: string, metadata: string, creationApprovalNotRequired: boolean): Promise<void> {
    const appDef: AppV1 = {
      id: appID,
      owner_wallet: this.txSigner.getAddress(),
      metadata,
      version: '1',
      creation_approval_not_required: creationApprovalNotRequired,
    };

    const packed = app.packAppV1(appDef);
    if (!this.txSigner.signPersonalMessage) {
      throw new Error('TransactionSigner must implement signPersonalMessage for app registration');
    }
    const ownerSig = await this.txSigner.signPersonalMessage(packed);

    const req: API.AppsV1SubmitAppVersionRequest = {
      app: appDef,
      owner_sig: ownerSig,
    };
    await this.rpcClient.appsV1SubmitAppVersion(req);
  }

  // ============================================================================
  // Channel Session Key Methods
  // ============================================================================

  /**
   * Sign a channel session key state using the client's state signer.
   * This creates a properly formatted EIP-191 signature that can be set on the state's
   * user_sig field before submitting via submitChannelSessionKeyState.
   *
   * @param state - The channel session key state to sign (user_sig field is excluded from signing)
   * @returns The hex-encoded signature string
   */
  async signChannelSessionKeyState(state: ChannelSessionKeyStateV1): Promise<Hex> {
    const metadataHash = core.getChannelSessionKeyAuthMetadataHashV1(
      BigInt(state.version),
      state.assets,
      BigInt(state.expires_at)
    );
    const packed = core.packChannelKeyStateV1(
      state.session_key as Address,
      metadataHash
    );
    const channelSig = await this.stateSigner.signMessage(packed);
    return stripSignerTypePrefix(channelSig);
  }

  /**
   * Submit a channel session key state for registration or update.
   * The state must be signed by the user's wallet to authorize the session key delegation.
   *
   * @param state - The channel session key state containing delegation information
   */
  async submitChannelSessionKeyState(state: ChannelSessionKeyStateV1): Promise<void> {
    const req: API.ChannelsV1SubmitSessionKeyStateRequest = {
      state,
    };
    await this.rpcClient.channelsV1SubmitSessionKeyState(req);
  }

  /**
   * Retrieve the latest channel session key states for a user.
   *
   * @param userAddress - The user's wallet address
   * @param sessionKey - Optional session key address to filter by
   * @returns List of active channel session key states
   */
  async getLastChannelKeyStates(
    userAddress: string,
    sessionKey?: string
  ): Promise<ChannelSessionKeyStateV1[]> {
    const req: API.ChannelsV1GetLastKeyStatesRequest = {
      user_address: userAddress,
      session_key: sessionKey,
    };
    const resp = await this.rpcClient.channelsV1GetLastKeyStates(req);
    return resp.states;
  }

  // ============================================================================
  // App Session Key Methods
  // ============================================================================

  /**
   * Sign an app session key state using the client's state signer.
   * This creates a properly formatted EIP-191 signature that can be set on the state's
   * user_sig field before submitting via submitSessionKeyState.
   *
   * @param state - The app session key state to sign (user_sig field is excluded from signing)
   * @returns The hex-encoded signature string
   */
  async signSessionKeyState(state: app.AppSessionKeyStateV1): Promise<Hex> {
    const packed = app.packAppSessionKeyStateV1(state);
    const channelSig = await this.stateSigner.signMessage(packed);
    return stripSignerTypePrefix(channelSig);
  }

  /**
   * Submit an app session key state for registration or update.
   * The state must be signed by the user's wallet to authorize the session key delegation.
   *
   * @param state - The session key state containing delegation information
   */
  async submitSessionKeyState(state: app.AppSessionKeyStateV1): Promise<void> {
    const req: API.AppSessionsV1SubmitSessionKeyStateRequest = {
      state,
    };
    await this.rpcClient.appSessionsV1SubmitSessionKeyState(req);
  }

  /**
   * Retrieve the latest session key states for a user.
   *
   * @param userAddress - The user's wallet address
   * @param sessionKey - Optional session key address to filter by
   * @returns List of active session key states
   */
  async getLastKeyStates(
    userAddress: string,
    sessionKey?: string
  ): Promise<app.AppSessionKeyStateV1[]> {
    const req: API.AppSessionsV1GetLastKeyStatesRequest = {
      user_address: userAddress,
      session_key: sessionKey,
    };
    const resp = await this.rpcClient.appSessionsV1GetLastKeyStates(req);
    return resp.states;
  }

  // ============================================================================
  // Private Helper Methods
  // ============================================================================

  /**
   * Get the RPC URL and blockchain info for a specific chain, validating that
   * the chain is configured and supported by the node.
   */
  private async getBlockchainRPCInfo(chainId: bigint): Promise<{ rpcUrl: string; blockchainInfo: core.Blockchain; config: core.NodeConfig }> {
    const rpcUrl = this.config.blockchainRPCs?.get(chainId);
    if (!rpcUrl) {
      throw new Error(
        `blockchain RPC not configured for chain ${chainId} (use withBlockchainRPC)`
      );
    }

    const config = await this.getConfig();
    const blockchainInfo = config.blockchains.find((b) => b.id === chainId);
    if (!blockchainInfo) {
      throw new Error(`blockchain ${chainId} not found in node config`);
    }

    return { rpcUrl, blockchainInfo, config };
  }

  /**
   * Create viem public and wallet clients for a specific chain and RPC URL.
   */
  private createEVMClients(chainId: bigint, rpcUrl: string): { publicClient: blockchain.evm.EVMClient; walletClient: blockchain.evm.WalletSigner | null } {
    const publicClient = createPublicClient({
      transport: http(rpcUrl),
    }) as blockchain.evm.EVMClient;

    const chain = {
      id: Number(chainId),
      name: `Chain ${chainId}`,
      nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
      rpcUrls: {
        default: { http: [rpcUrl] },
        public: { http: [rpcUrl] },
      },
    };

    const isBrowser = typeof window !== 'undefined' && typeof (window as any).ethereum !== 'undefined';

    let walletClient: blockchain.evm.WalletSigner | null = null;
    if (isBrowser) {
      walletClient = createWalletClient({
        chain,
        transport: custom((window as any).ethereum),
        account: this.txSigner.getAddress(),
      }) as blockchain.evm.WalletSigner;
    } else {
      const account = this.txSigner.getAccount ? this.txSigner.getAccount() : undefined;
      if (account) {
        walletClient = createWalletClient({
          chain,
          transport: http(rpcUrl),
          account,
        }) as blockchain.evm.WalletSigner;
      }
    }

    return { publicClient, walletClient };
  }

  /**
   * Initialize a blockchain client for a specific chain.
   * This is called lazily when a blockchain operation is needed.
   */
  private async initializeBlockchainClient(chainId: bigint): Promise<void> {
    if (this.blockchainClients.has(chainId)) {
      return;
    }

    const { rpcUrl, blockchainInfo, config } = await this.getBlockchainRPCInfo(chainId);

    if (!blockchainInfo.channelHubAddress) {
      throw new Error(`channel hub address not configured for blockchain ${chainId}`);
    }

    const { publicClient, walletClient } = this.createEVMClients(chainId, rpcUrl);

    if (!walletClient) {
      throw new Error('Node.js environment requires a TransactionSigner that implements getAccount() (e.g., EthereumRawSigner)');
    }

    const blockchainClient = new blockchain.evm.Client(
      blockchainInfo.channelHubAddress,
      publicClient,
      walletClient,
      chainId,
      config.nodeAddress,
      this.assetStore
    );

    this.blockchainClients.set(chainId, blockchainClient);
  }

  /**
   * Initialize a Locking contract client for a specific chain.
   * This is called lazily when a locking operation is needed.
   */
  private async initializeLockingClient(chainId: bigint): Promise<void> {
    if (this.blockchainLockingClients.has(chainId)) {
      return;
    }

    const { rpcUrl, blockchainInfo } = await this.getBlockchainRPCInfo(chainId);

    if (!blockchainInfo.lockingContractAddress) {
      throw new Error(`locking contract address not configured for blockchain ${chainId}`);
    }

    const { publicClient, walletClient } = this.createEVMClients(chainId, rpcUrl);

    const lockingClient = new blockchain.evm.LockingClient(
      blockchainInfo.lockingContractAddress,
      publicClient,
      walletClient || undefined,
    );

    this.blockchainLockingClients.set(chainId, lockingClient);
  }

  /**
   * Build a hex bitmap string from an array of signer type numbers.
   * Each signer type sets a bit at its corresponding position in a 256-bit value.
   */
  private buildSigValidatorsBitmap(signerTypes: number[]): string {
    const bitmap = new Uint8Array(32);
    for (const t of signerTypes) {
      const idx = t & 0xff;
      bitmap[31 - Math.floor(idx / 8)] |= 1 << (idx % 8);
    }
    // Trim leading zeros for compact hex representation
    for (let i = 0; i < 32; i++) {
      if (bitmap[i] !== 0) {
        const hex = Array.from(bitmap.slice(i))
          .map((b) => b.toString(16).padStart(2, '0'))
          .join('');
        return '0x' + hex;
      }
    }
    return '0x00';
  }

  /**
   * Get supported sig validators bitmap from node config.
   */
  private async getSupportedSigValidatorsBitmap(): Promise<string> {
    const config = await this.getConfig();
    return this.buildSigValidatorsBitmap(config.supportedSigValidators);
  }

  /**
   * Get the node address from the config.
   */
  private async getNodeAddress(): Promise<Address> {
    const config = await this.getConfig();
    return config.nodeAddress;
  }

  /**
   * Submit a signed state update to the node.
   */
  private async submitState(state: core.State): Promise<string> {
    const req: API.ChannelsV1SubmitStateRequest = {
      state: this.transformStateToRPC(state),
    };
    const resp = await this.rpcClient.channelsV1SubmitState(req);
    return resp.signature;
  }

  /**
   * Request the node to co-sign a channel creation state.
   * Used when creating a new channel (via deposit, withdraw, transfer, or acknowledge).
   */
  private async requestChannelCreation(
    state: core.State,
    channelDef: core.ChannelDefinition
  ): Promise<string> {
    const req: API.ChannelsV1RequestCreationRequest = {
      state: this.transformStateToRPC(state),
      channel_definition: this.transformChannelDefinitionToRPC(channelDef),
    };
    const resp = await this.rpcClient.channelsV1RequestCreation(req);
    return resp.signature;
  }

  /**
   * Transform core State to RPC StateV1
   */
  private transformStateToRPC(state: core.State): StateV1 {
    // This is a simplified version - you'll need to implement the full transformation
    return {
      id: state.id,
      transition: {
        type: state.transition.type,
        tx_id: state.transition.txId,
        account_id: state.transition.accountId || '',
        amount: state.transition.amount.toString(),
      },
      asset: state.asset,
      user_wallet: state.userWallet,
      epoch: state.epoch.toString(), // Convert bigint to string
      version: state.version.toString(), // Convert bigint to string
      home_channel_id: state.homeChannelId,
      escrow_channel_id: state.escrowChannelId,
      home_ledger: {
        token_address: state.homeLedger.tokenAddress,
        blockchain_id: state.homeLedger.blockchainId.toString(),
        user_balance: state.homeLedger.userBalance.toString(),
        user_net_flow: state.homeLedger.userNetFlow.toString(),
        node_balance: state.homeLedger.nodeBalance.toString(),
        node_net_flow: state.homeLedger.nodeNetFlow.toString(),
      },
      escrow_ledger: state.escrowLedger
        ? {
            token_address: state.escrowLedger.tokenAddress,
            blockchain_id: state.escrowLedger.blockchainId.toString(),
            user_balance: state.escrowLedger.userBalance.toString(),
            user_net_flow: state.escrowLedger.userNetFlow.toString(),
            node_balance: state.escrowLedger.nodeBalance.toString(),
            node_net_flow: state.escrowLedger.nodeNetFlow.toString(),
          }
        : undefined,
      user_sig: state.userSig,
      node_sig: state.nodeSig,
    };
  }

  /**
   * Transform core ChannelDefinition to RPC ChannelDefinitionV1
   */
  private transformChannelDefinitionToRPC(def: core.ChannelDefinition): ChannelDefinitionV1 {
    return {
      nonce: def.nonce.toString(),
      challenge: def.challenge,
      approved_sig_validators: def.approvedSigValidators,
    };
  }

  /**
   * Convert transition type enum to string
   */
  private transitionTypeToString(type: core.TransitionType): string {
    const typeMap: Record<core.TransitionType, string> = {
      [core.TransitionType.Void]: 'void',
      [core.TransitionType.Acknowledgement]: 'acknowledgement',
      [core.TransitionType.HomeDeposit]: 'home_deposit',
      [core.TransitionType.HomeWithdrawal]: 'home_withdrawal',
      [core.TransitionType.EscrowDeposit]: 'escrow_deposit',
      [core.TransitionType.EscrowWithdraw]: 'escrow_withdraw',
      [core.TransitionType.TransferSend]: 'transfer_send',
      [core.TransitionType.TransferReceive]: 'transfer_receive',
      [core.TransitionType.Commit]: 'commit',
      [core.TransitionType.Release]: 'release',
      [core.TransitionType.Migrate]: 'migrate',
      [core.TransitionType.EscrowLock]: 'escrow_lock',
      [core.TransitionType.MutualLock]: 'mutual_lock',
      [core.TransitionType.Finalize]: 'finalize',
    };
    return typeMap[type] || 'home_deposit';
  }
}
