#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { setTimeout as sleep } from 'node:timers/promises';

import { Client, createSigners, withErrorHandler } from '../../sdk/ts/dist/index.js';
import { NitroliteClient } from '../../sdk/ts-compat/dist/index.js';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '../..');
const sdkRequire = createRequire(path.join(repoRoot, 'sdk/ts/package.json'));
const WebSocketCtor = globalThis.WebSocket ?? sdkRequire('ws');
const wsURL = process.env.NITRONODE_RUNTIME_SMOKE_WS_URL ?? 'ws://127.0.0.1:7824/ws';
const readyTimeoutMs = Number(process.env.NITRONODE_RUNTIME_SMOKE_READY_TIMEOUT_MS ?? 15000);
const adversarialMode = process.env.NITRONODE_RUNTIME_SMOKE_ADVERSARIAL ?? '';
const externalLogDirInput = process.env.NITRONODE_RUNTIME_SMOKE_LOG_DIR ?? '';
const useExternalNode = process.env.NITRONODE_RUNTIME_SMOKE_EXTERNAL === '1';
const anvilPrivateKey =
  '0x59c6995e998f97a5a0044966f094538f0d0921e301baca6a9ae52cd7834c90b9';

class SmokeError extends Error {
  constructor(category, message, cause) {
    super(`[${category}] ${message}${cause ? `: ${cause.message ?? cause}` : ''}`);
    this.name = 'SmokeError';
    this.category = category;
    this.cause = cause;
  }
}

function assertSmoke(condition, category, message) {
  if (!condition) {
    throw new SmokeError(category, message);
  }
}

function resolveRepoChildPath(input, label) {
  const resolved = path.resolve(repoRoot, input);
  const relative = path.relative(repoRoot, resolved);
  if (relative === '' || relative.startsWith('..') || path.isAbsolute(relative)) {
    throw new SmokeError('setup', `${label} must resolve inside the repository`);
  }
  return resolved;
}

function privateKeyForMode() {
  const configuredPrivateKey = process.env.NITRONODE_RUNTIME_SMOKE_PRIVATE_KEY;
  if (useExternalNode) {
    if (!configuredPrivateKey) {
      throw new SmokeError(
        'setup',
        'NITRONODE_RUNTIME_SMOKE_PRIVATE_KEY is required when NITRONODE_RUNTIME_SMOKE_EXTERNAL=1'
      );
    }
    return configuredPrivateKey;
  }

  // Well-known Anvil/Hardhat test account #2, used only for isolated local smoke.
  return configuredPrivateKey ?? anvilPrivateKey;
}

function childEnv(configDir) {
  const env = {
    PATH: process.env.PATH,
    HOME: process.env.HOME,
    TMPDIR: process.env.TMPDIR,
    NITRONODE_CONFIG_DIR_PATH: configDir,
  };

  return Object.fromEntries(Object.entries(env).filter(([, value]) => value !== undefined));
}

function logStep(message) {
  console.log(`[runtime-smoke] ${message}`);
}

const externalLogDir = externalLogDirInput
  ? resolveRepoChildPath(externalLogDirInput, 'NITRONODE_RUNTIME_SMOKE_LOG_DIR')
  : '';
const privateKey = privateKeyForMode();

async function withTimeout(label, promise, timeoutMs = 5000) {
  const timeout = sleep(timeoutMs).then(() => {
    throw new SmokeError('timeout', `${label} timed out after ${timeoutMs}ms`);
  });
  return Promise.race([promise, timeout]);
}

function openWebSocket(url, timeoutMs = 500) {
  return new Promise((resolve, reject) => {
    const ws = new WebSocketCtor(url);
    let settled = false;

    const finish = (err) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      try {
        ws.close();
      } catch {
        // Ignore close errors while probing readiness.
      }
      if (err) reject(err);
      else resolve();
    };

    const timer = setTimeout(() => finish(new Error('WebSocket connect timeout')), timeoutMs);
    ws.onopen = () => finish();
    ws.onerror = () => finish(new Error('WebSocket connection error'));
    ws.onclose = () => finish(new Error('WebSocket closed before open'));
  });
}

async function waitForWebSocket(url, child = null, timeoutMs = 15000) {
  const deadline = Date.now() + timeoutMs;
  let lastError = null;

  while (Date.now() < deadline) {
    if (child && child.exitCode !== null) {
      throw new SmokeError(
        'startup',
        `Nitronode exited before readiness with code ${child.exitCode}`
      );
    }

    try {
      await openWebSocket(url);
      return;
    } catch (err) {
      lastError = err;
      await sleep(250);
    }
  }

  throw new SmokeError(
    'connection',
    `Nitronode did not accept WebSocket connections at ${url}`,
    lastError
  );
}

async function stopProcess(child) {
  if (child.exitCode !== null || child.signalCode !== null) return;

  child.kill('SIGTERM');
  const exited = await Promise.race([
    once(child, 'exit').then(() => true),
    sleep(5000).then(() => false),
  ]);
  if (exited) return;

  child.kill('SIGKILL');
}

async function closeClient(client) {
  if (!client) return;

  const closed = await Promise.race([
    client.close().then(
      () => true,
      (err) => {
        console.warn(`[runtime-smoke] client.close failed: ${err.message ?? err}`);
        return true;
      }
    ),
    sleep(3000).then(() => false),
  ]);

  if (!closed) {
    console.warn('[runtime-smoke] client.close timed out; continuing cleanup');
  }
}

async function runCommand(command, args, options, category) {
  return new Promise((resolve, reject) => {
    let stderr = '';
    const child = spawn(command, args, options);
    child.stderr?.on('data', (chunk) => {
      stderr += chunk.toString();
    });
    child.on('error', (err) => reject(new SmokeError(category, `${command} failed to start`, err)));
    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(
        new SmokeError(
          category,
          `${command} ${args.join(' ')} exited with ${signal ?? code}${stderr ? `\n${stderr}` : ''}`
        )
      );
    });
  });
}

async function writeConfig(configDir) {
  await writeFile(
    path.join(configDir, '.env'),
    [
      'NITRONODE_DATABASE_DRIVER=sqlite',
      'NITRONODE_SIGNER_TYPE=key',
      `NITRONODE_SIGNER_KEY=${privateKey}`,
      'NITRONODE_LOG_LEVEL=error',
      '',
    ].join('\n')
  );

  if (adversarialMode === 'bad-config') {
    await writeFile(path.join(configDir, 'blockchains.yaml'), 'blockchains:\n  - name: BAD_NAME\n');
    await writeFile(path.join(configDir, 'assets.yaml'), 'assets: []\n');
    return;
  }

  await writeFile(path.join(configDir, 'blockchains.yaml'), 'blockchains: []\n');
  await writeFile(path.join(configDir, 'assets.yaml'), 'assets: []\n');
}

async function writeFailureLogs(paths, stdout, stderr, summary) {
  await writeFile(paths.stdoutPath, stdout);
  await writeFile(paths.stderrPath, stderr);

  if (!externalLogDir) return;

  await mkdir(externalLogDir, { recursive: true });
  await writeFile(path.join(externalLogDir, 'summary.txt'), summary);
  await writeFile(path.join(externalLogDir, 'nitronode.stdout.log'), stdout);
  await writeFile(path.join(externalLogDir, 'nitronode.stderr.log'), stderr);
}

async function runSmoke() {
  const configDir = await mkdtemp(path.join(tmpdir(), 'nitrolite-runtime-smoke-'));
  const binaryPath = path.join(configDir, 'nitronode-smoke');
  const stdoutPath = path.join(configDir, 'nitronode.stdout.log');
  const stderrPath = path.join(configDir, 'nitronode.stderr.log');
  let stdout = '';
  let stderr = '';
  let client = null;
  let child = null;
  let compatLogLines = [];

  const logs = () => [
    `stdout (${stdoutPath}):`,
    stdout.trim() || '<empty>',
    `stderr (${stderrPath}):`,
    stderr.trim() || '<empty>',
  ].join('\n');

  try {
    if (useExternalNode) {
      logStep(`using external Nitronode at ${wsURL}`);
    } else {
      logStep(`writing isolated config in ${configDir}`);
      await writeConfig(configDir);
      logStep('building temporary Nitronode binary');
      await runCommand('go', ['build', '-o', binaryPath, './nitronode'], { cwd: repoRoot }, 'setup');

      logStep(`starting Nitronode and waiting for ${wsURL}`);
      child = spawn(binaryPath, {
        cwd: repoRoot,
        env: childEnv(configDir),
        stdio: ['ignore', 'pipe', 'pipe'],
      });

      child.stdout.on('data', (chunk) => {
        stdout += chunk.toString();
      });
      child.stderr.on('data', (chunk) => {
        stderr += chunk.toString();
      });
    }

    await waitForWebSocket(wsURL, child, readyTimeoutMs);

    const { stateSigner, txSigner } = createSigners(privateKey);
    const wallet = stateSigner.getAddress();
    logStep(`creating TS SDK client for wallet ${wallet}`);
    client = await withTimeout(
      'Client.create',
      Client.create(wsURL, stateSigner, txSigner, withErrorHandler(() => {}))
    );

    logStep('calling ping');
    await withTimeout('client.ping', client.ping());

    logStep('calling getConfig');
    const config = await withTimeout('client.getConfig', client.getConfig());
    assertSmoke(typeof config.nodeAddress === 'string', 'transform', 'node config nodeAddress is not a string');
    assertSmoke(Array.isArray(config.blockchains), 'transform', 'node config blockchains is not an array');
    assertSmoke(
      Array.isArray(config.supportedSigValidators),
      'transform',
      'node config supportedSigValidators is not an array'
    );
    if (!useExternalNode) {
      assertSmoke(
        config.nodeAddress.toLowerCase() === wallet.toLowerCase(),
        'transform',
        `expected node address ${wallet}, got ${config.nodeAddress}`
      );
      assertSmoke(config.blockchains.length === 0, 'transform', 'runtime smoke config should expose no blockchains');
    }

    logStep('calling getAssets');
    const assets = await withTimeout('client.getAssets', client.getAssets());
    assertSmoke(Array.isArray(assets), 'transform', 'assets response is not an array');
    if (!useExternalNode) {
      assertSmoke(assets.length === 0, 'transform', 'runtime smoke config should expose no assets');
    }

    logStep('calling getAppSessions');
    const appSessions = await withTimeout(
      'client.getAppSessions',
      client.getAppSessions({ wallet })
    );
    assertSmoke(Array.isArray(appSessions.sessions), 'transform', 'app sessions is not an array');
    assertSmoke(appSessions.sessions.length === 0, 'transform', 'expected no app sessions for smoke wallet');

    logStep('calling getLastChannelKeyStates');
    const channelKeyStates = await withTimeout(
      'client.getLastChannelKeyStates',
      client.getLastChannelKeyStates(wallet)
    );
    assertSmoke(Array.isArray(channelKeyStates), 'transform', 'channel key states is not an array');

    logStep('calling getLastKeyStates');
    const appSessionKeyStates = await withTimeout(
      'client.getLastKeyStates',
      client.getLastKeyStates(wallet)
    );
    assertSmoke(Array.isArray(appSessionKeyStates), 'transform', 'app session key states is not an array');

    logStep('validating compat getAppSessionsList mapping');
    const compatClient = Object.create(NitroliteClient.prototype);
    compatClient.userAddress = wallet;
    compatClient.innerClient = client;
    compatClient.assetsBySymbol = new Map();
    compatClient._lastAppSessionsListError = null;
    compatClient._lastAppSessionsListErrorLogged = null;

    const originalInfo = console.info;
    const originalWarn = console.warn;
    let compatSessions;
    try {
      compatLogLines = [];
      console.info = (...args) => compatLogLines.push(['info', ...args].join(' '));
      console.warn = (...args) => compatLogLines.push(['warn', ...args].join(' '));
      compatSessions = await withTimeout(
        'compat.getAppSessionsList',
        compatClient.getAppSessionsList()
      );
    } finally {
      console.info = originalInfo;
      console.warn = originalWarn;
    }
    assertSmoke(Array.isArray(compatSessions), 'compat mapping', 'compat sessions is not an array');
    assertSmoke(compatSessions.length === 0, 'compat mapping', 'expected no compat app sessions');
    assertSmoke(
      compatClient.getLastAppSessionsListError() === null,
      'compat mapping',
      `compat mapping reported ${compatClient.getLastAppSessionsListError()}`
    );

    logStep('runtime smoke passed');
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    const summary = compatLogLines.length > 0
      ? `${message}\n\ncompat logs:\n${compatLogLines.join('\n')}`
      : message;
    await writeFailureLogs({ stdoutPath, stderrPath }, stdout, stderr, summary);
    console.error(message);
    if (compatLogLines.length > 0) {
      console.error(`compat logs:\n${compatLogLines.join('\n')}`);
    }
    if (err instanceof SmokeError) {
      console.error(logs());
    }
    process.exitCode = 1;
  } finally {
    try {
      await closeClient(client);
    } finally {
      if (child) {
        logStep('stopping Nitronode');
        await stopProcess(child);
      }
      if (process.exitCode) {
        console.error(`runtime smoke logs preserved at ${configDir}`);
      } else {
        await rm(configDir, { recursive: true, force: true });
      }
    }
  }
}

await runSmoke();
process.exit(process.exitCode ?? 0);
