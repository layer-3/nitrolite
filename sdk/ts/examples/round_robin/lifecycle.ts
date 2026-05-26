/**
 * Example: Round Robin Test
 *
 * Exercises every basic channel action (deposit, transfer, close, withdraw)
 * for one asset across every chain on which that asset is supported.
 *
 * ---------------------------------------------------------------------------
 * Flow
 * ---------------------------------------------------------------------------
 *
 *  1. Preparation
 *     a. Set up wallet signers for private keys A and B.
 *     b. Build tokenSet: every supported token for the configured asset,
 *        paired with its chain's JSON-RPC endpoint. tokenSet[i].blockchainId
 *        must have an entry in chainRPCs.
 *     c. Ensure signer A holds >= minNativeBalances[chainId] of the native
 *        gas token on every chain in tokenSet. Each chain pays three
 *        checkpoint transactions per run (deposit + close + withdraw).
 *     d. Ensure signer A holds >= transferAmount of the configured asset on
 *        tokenSet[0].blockchainId. Subsequent iterations are seeded by the
 *        previous iteration's withdrawal — no other chain needs to be
 *        pre-funded with the asset.
 *     e. If sessionKeyPriv is non-empty, register it as a channel session
 *        key for signer A and use the session-key-backed client for every
 *        channel operation in the loop.
 *
 *  2. For i = 0 .. tokenSet.length-1 (let next = (i + 1) % tokenSet.length):
 *     a. Signer A deposits transferAmount of tokenSet[i].
 *     b. Signer A transfers transferAmount to signer B (off-chain).
 *     c. Signer A closes the home channel for the asset.
 *     d. Signer B transfers transferAmount back to signer A (lands on a
 *        void state, no chain attached).
 *     e. Signer A withdraws transferAmount on tokenSet[next].blockchainId,
 *        which auto-creates a new channel on that chain. On the last
 *        iteration this wraps to tokenSet[0], closing the loop.
 *
 * Together (2.a–2.e) exercise deposit / transfer-send / close /
 * transfer-receive / withdraw on every chain in tokenSet.
 *
 * ---------------------------------------------------------------------------
 * Operational notes
 * ---------------------------------------------------------------------------
 *
 *   - wsURL must point at a reachable nitronode WebSocket endpoint.
 *   - The nitronode must maintain reserves of the asset on every chain in
 *     tokenSet. 2.e withdraws tokens signer A never deposited on that chain,
 *     which only works if the node holds liquidity there.
 *   - tokenSet is derived at runtime from getAssets(asset).tokens; the
 *     iteration order matches the order returned by the node.
 *   - The example exits non-zero on the first failure. It does not reset
 *     state, so re-running on the same wallets resumes from whatever state
 *     the node holds. Fresh wallets give the cleanest preflight numbers.
 */

// Node <22 does not expose a stable global WebSocket. The SDK dialer uses
// `new WebSocket(url)` directly, so we polyfill the global from the `ws`
// package before any client is constructed.
import WebSocket from 'ws';
if (typeof (globalThis as { WebSocket?: unknown }).WebSocket === 'undefined') {
  (globalThis as { WebSocket: unknown }).WebSocket = WebSocket;
}

import Decimal from 'decimal.js';
import { Address, createPublicClient, http, Hex } from 'viem';
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';

import { Client } from '../../src/client';
import {
  createSigners,
  EthereumMsgSigner,
  EthereumRawSigner,
  ChannelSessionKeyStateSigner,
} from '../../src/signers';
import { withBlockchainRPC } from '../../src/config';
import { ChannelSessionKeyStateV1 } from '../../src/rpc/types';
import { getChannelSessionKeyAuthMetadataHashV1 } from '../../src/core/utils';
import { ChannelStatus, Token } from '../../src/core/types';

// ============================================================================
// Configuration
// ============================================================================

const wsURL = 'wss://nitronode-sandbox.yellow.org/v1/ws';

// Replace with your hex private keys. privA performs every channel
// operation; privB only receives transfers and sends them back.
const privA = '0x7d6071...' as Hex;
const privB = '0xf63695...' as Hex;

// Empty string disables the session-key path; the wallet client performs
// channel operations directly. Otherwise this key is registered as a channel
// session key for privA and used for the loop.
const sessionKeyPriv = '' as Hex | '';

// Asset symbol to test. tokenSet is built from this asset's tokens as
// returned by the node.
const asset = 'yusd';

// transferAmount is the value used for every deposit / transfer / withdraw
// in the loop. Keep this small for testnets.
const transferAmount = new Decimal('0.00001');

// chainRPCs maps blockchain id -> JSON-RPC endpoint. Must cover every chain
// in tokenSet (derived at runtime). Missing entries fail preflight.
const chainRPCs: Record<string, string> = {
  '11155111': 'https://sepolia.drpc.org',              // Ethereum Sepolia
  '84532':    'https://sepolia.base.org',              // Base Sepolia
  '80002':    'https://rpc-amoy.polygon.technology',   // Polygon Amoy
  '59141':    'https://rpc.sepolia.linea.build',       // Linea Sepolia
};

// minNativeBalances maps blockchain id -> minimum native gas balance for
// privA. Sized to cover three checkpoint transactions (deposit + close +
// withdraw). Tune per chain based on observed gas costs.
const minNativeBalances: Record<string, Decimal> = {
  '11155111': new Decimal('0.01'),
  '84532':    new Decimal('0.005'),
  '80002':    new Decimal('0.05'),
  '59141':    new Decimal('0.005'),
};

// ============================================================================
// Main
// ============================================================================

async function main() {
  // --- Build wallet clients for A and B with every RPC pre-registered ---
  const signersA = createSigners(privA);
  const signersB = createSigners(privB);
  const addrA = signersA.stateSigner.getAddress();
  const addrB = signersB.stateSigner.getAddress();
  console.log(`Wallet A: ${addrA}\nWallet B: ${addrB}\n`);

  const walletA = await newClient(signersA.stateSigner, signersA.txSigner);
  const walletB = await newClient(signersB.stateSigner, signersB.txSigner);

  let opsClient = walletA;
  try {
    // --- 1.b: discover tokenSet for the configured asset ---
    console.log('=== 1.b: Discovering tokenSet ===');
    const tokenSet = await discoverTokenSet(walletA, asset);
    tokenSet.forEach((t, i) =>
      console.log(`  [${i}] ${t.symbol} on chain ${t.blockchainId} (${t.address})`)
    );
    console.log();

    // --- 1.c + 1.d: preflight native and asset balances ---
    console.log('=== 1.c + 1.d: Preflight ===');
    await preflight(walletA, addrA, tokenSet);
    console.log('✓ preflight passed\n');

    // --- 1.e: optional session-key registration ---
    if (sessionKeyPriv !== '') {
      console.log('=== 1.e: Registering channel session key ===');
      opsClient = await setupSessionKeyClient(walletA, addrA);
      console.log('✓ session-key client ready\n');
    }

    // --- 2: round-robin loop ---
    console.log('=== 2: Round robin ===');
    for (let i = 0; i < tokenSet.length; i++) {
      const next = (i + 1) % tokenSet.length;
      await runIteration(i, opsClient, walletB, addrA, addrB, tokenSet[i], tokenSet[next]);
    }

    console.log('\n=== Example Complete ===');
  } finally {
    if (opsClient !== walletA) await opsClient.close();
    await walletA.close();
    await walletB.close();
  }
}

// ============================================================================
// Iteration
// ============================================================================

async function runIteration(
  i: number,
  opsClient: Client,
  walletB: Client,
  addrA: Address,
  addrB: Address,
  cur: Token,
  next: Token
): Promise<void> {
  console.log(`--- iter ${i}: deposit on chain ${cur.blockchainId}, withdraw on chain ${next.blockchainId} ---`);

  // 2.a A deposits transferAmount on cur.blockchainId.
  await opsClient.approveToken(cur.blockchainId, asset, transferAmount);
  const depositState = await opsClient.deposit(cur.blockchainId, asset, transferAmount);
  console.log(`  ✓ A deposited ${transferAmount} ${asset} on chain ${cur.blockchainId}`);
  await checkpointAndWait(opsClient, asset, depositState.version);

  // 2.b A -> B (off-chain).
  await opsClient.transfer(addrB, asset, transferAmount);
  console.log(`  ✓ A transferred ${transferAmount} ${asset} to B (off-chain)`);

  // 2.c A closes home channel for asset on cur.blockchainId.
  await opsClient.closeHomeChannel(asset);
  console.log(`  ✓ A closed home channel for ${asset}`);
  await closeAndWait(opsClient, asset);

  // 2.d B -> A (off-chain credit, lands on void state, no chain attached).
  await walletB.transfer(addrA, asset, transferAmount);
  console.log(`  ✓ B transferred ${transferAmount} ${asset} back to A (off-chain credit)`);

  // 2.e A withdraws on next.blockchainId. Withdraw auto-creates a new channel
  // on next.blockchainId because A's latest state is void after close +
  // transfer-receive.
  const withdrawState = await opsClient.withdraw(next.blockchainId, asset, transferAmount);
  console.log(`  ✓ A withdrew ${transferAmount} ${asset} on chain ${next.blockchainId}`);
  await checkpointAndWait(opsClient, asset, withdrawState.version);

  await waitForOnChain(opsClient, next.blockchainId, asset, addrA, transferAmount);
  console.log();
}

// ============================================================================
// Preflight
// ============================================================================

async function preflight(walletA: Client, addrA: Address, tokenSet: Token[]): Promise<void> {
  const shortfalls: string[] = [];

  for (const t of tokenSet) {
    const key = t.blockchainId.toString();
    if (!(key in chainRPCs)) shortfalls.push(`chainRPCs missing entry for chain ${key}`);
    if (!(key in minNativeBalances)) shortfalls.push(`minNativeBalances missing entry for chain ${key}`);
  }
  if (shortfalls.length > 0) {
    console.log('Configuration shortfalls:');
    shortfalls.forEach((s) => console.log(`  - ${s}`));
    process.exit(1);
  }

  // 1.c: native balance on every chain >= minNativeBalances.
  console.log('Native balance check:');
  for (const t of tokenSet) {
    const key = t.blockchainId.toString();
    const have = await nativeBalance(chainRPCs[key], addrA);
    const need = minNativeBalances[key];
    const marker = have.gte(need) ? '✓' : '✗';
    if (have.lt(need)) {
      shortfalls.push(`chain ${key} native: need >= ${need}, have ${have}`);
    }
    console.log(`  ${marker} chain ${key}: need >= ${need}, have ${have}`);
  }

  // 1.d: privA holds >= transferAmount of asset on tokenSet[0].blockchainId.
  const seedChain = tokenSet[0].blockchainId;
  const have = await walletA.getOnChainBalance(seedChain, asset, addrA);
  const marker = have.gte(transferAmount) ? '✓' : '✗';
  if (have.lt(transferAmount)) {
    shortfalls.push(`chain ${seedChain} ${asset}: need >= ${transferAmount}, have ${have}`);
  }
  console.log(`Asset balance check:\n  ${marker} chain ${seedChain} ${asset}: need >= ${transferAmount}, have ${have}`);

  if (shortfalls.length > 0) {
    console.log('\nPreflight failed. Resolve the following before re-running:');
    shortfalls.forEach((s) => console.log(`  - ${s}`));
    process.exit(1);
  }
}

async function nativeBalance(rpcURL: string, addr: Address): Promise<Decimal> {
  const pc = createPublicClient({ transport: http(rpcURL) });
  const wei = await pc.getBalance({ address: addr });
  return new Decimal(wei.toString()).div(new Decimal(10).pow(18));
}

// ============================================================================
// Token discovery
// ============================================================================

async function discoverTokenSet(client: Client, symbol: string): Promise<Token[]> {
  const assets = await client.getAssets();
  const found = assets.find((a) => a.symbol.toLowerCase() === symbol.toLowerCase());
  if (!found) throw new Error(`asset ${symbol} not supported by node`);
  if (found.tokens.length === 0) throw new Error(`asset ${symbol} has no supported tokens`);
  return found.tokens;
}

// ============================================================================
// Client construction
// ============================================================================

async function newClient(stateSigner: any, txSigner: any): Promise<Client> {
  const opts = Object.entries(chainRPCs).map(([id, url]) => withBlockchainRPC(BigInt(id), url));
  return await Client.create(wsURL, stateSigner, txSigner, ...opts);
}

// ============================================================================
// Session key
// ============================================================================

async function setupSessionKeyClient(walletA: Client, addrA: Address): Promise<Client> {
  const sessionKey = sessionKeyPriv === '' ? generatePrivateKey() : (sessionKeyPriv as Hex);
  const skAccount = privateKeyToAccount(sessionKey);
  const skAddress = skAccount.address;
  const skMsg = new EthereumMsgSigner(sessionKey);
  console.log(`Session key: ${skAddress}`);

  const expiresAt = BigInt(Math.floor(Date.now() / 1000) + 24 * 60 * 60);
  const state: ChannelSessionKeyStateV1 = {
    user_address: addrA,
    session_key: skAddress,
    version: '1',
    assets: [asset],
    expires_at: expiresAt.toString(),
    user_sig: '',
    session_key_sig: '',
  };

  state.user_sig = await walletA.signChannelSessionKeyState(state);
  state.session_key_sig = await walletA.signChannelSessionKeyOwnership(state, skMsg);
  await walletA.submitChannelSessionKeyState(state);

  const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
    addrA,
    BigInt(state.version),
    state.assets,
    BigInt(state.expires_at)
  );

  // user_sig has been stripped of any ChannelSigner type-byte prefix by
  // signChannelSessionKeyState before submission, so it is the raw EIP-191
  // signature expected by the session-key signer.
  const stateSigner = new ChannelSessionKeyStateSigner(
    sessionKey,
    addrA,
    metadataHash,
    state.user_sig as Hex
  );

  // txSigner stays as the wallet's key: the SDK uses it to derive the user
  // address and to sign on-chain checkpoint transactions.
  const txSigner = new EthereumRawSigner(privA);

  const opts = Object.entries(chainRPCs).map(([id, url]) => withBlockchainRPC(BigInt(id), url));
  return await Client.create(wsURL, stateSigner, txSigner, ...opts);
}

// ============================================================================
// Wait helpers
// ============================================================================

/**
 * Runs checkpoint after a finalize transition and polls getHomeChannel until
 * the node observes the on-chain close. Closure is signalled either by the
 * home-channel row dropping out (null) or by its status being reset to Void.
 */
async function closeAndWait(client: Client, asset: string): Promise<void> {
  const txHash = await client.checkpoint(asset);
  console.log(`    ↳ checkpoint ${asset} tx ${txHash}; waiting for channel close (null or status=Void)...`);

  const wallet = client.getUserAddress();
  const deadline = Date.now() + 2 * 60 * 1000;
  while (true) {
    const channel = await client.getHomeChannel(wallet, asset);
    if (channel === null || channel.status === ChannelStatus.Void) return;
    if (Date.now() > deadline) {
      throw new Error(`timed out waiting for ${asset} channel to close (last status=${channel.status})`);
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
}

/**
 * Runs checkpoint and polls getHomeChannel until the node observes the
 * expected post-checkpoint state.
 *
 * When expectedVersion > 0 the helper waits for channel.stateVersion to catch
 * up to expectedVersion. When expectedVersion === 0n — which happens for the
 * channel-creation transitions issued by deposit / withdraw on a void state —
 * the state_version stays at 0 even after the checkpoint, so the helper
 * instead waits for channel.status === Open.
 */
async function checkpointAndWait(client: Client, asset: string, expectedVersion: bigint): Promise<void> {
  const txHash = await client.checkpoint(asset);
  if (expectedVersion === 0n) {
    console.log(`    ↳ checkpoint ${asset} tx ${txHash}; waiting for channel status=Open...`);
  } else {
    console.log(`    ↳ checkpoint ${asset} tx ${txHash}; waiting for state_version=${expectedVersion}...`);
  }

  const wallet = client.getUserAddress();
  const deadline = Date.now() + 2 * 60 * 1000;
  while (true) {
    const channel = await client.getHomeChannel(wallet, asset);
    if (channel !== null) {
      if (expectedVersion === 0n && channel.status === ChannelStatus.Open) return;
      if (expectedVersion > 0n && channel.stateVersion >= expectedVersion) return;
    }
    if (Date.now() > deadline) {
      if (expectedVersion === 0n) {
        throw new Error(`timed out waiting for ${asset} channel to reach status=Open`);
      }
      throw new Error(`timed out waiting for ${asset} to reach state_version=${expectedVersion}`);
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
}

async function waitForOnChain(
  client: Client,
  chainId: bigint,
  asset: string,
  addr: Address,
  minAmount: Decimal
): Promise<void> {
  const deadline = Date.now() + 2 * 60 * 1000;
  while (true) {
    const have = await client.getOnChainBalance(chainId, asset, addr);
    if (have.gte(minAmount)) {
      console.log(`    ↳ on-chain balance on chain ${chainId} settled: ${have} ${asset}`);
      return;
    }
    if (Date.now() > deadline) {
      throw new Error(`timed out waiting for chain ${chainId} ${asset} balance >= ${minAmount} (have ${have})`);
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
}

main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
