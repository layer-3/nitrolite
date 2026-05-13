#!/usr/bin/env node
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const packageRoot = resolve(scriptDir, '..');

function readJson(path) {
  return JSON.parse(readFileSync(path, 'utf-8'));
}

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

function run(command, args) {
  return execFileSync(command, args, {
    cwd: packageRoot,
    encoding: 'utf-8',
    stdio: ['ignore', 'pipe', 'pipe'],
  }).trim();
}

function npmVersionExists(packageName, version) {
  let resolved;
  try {
    resolved = run('npm', ['view', `${packageName}@${version}`, 'version']);
  } catch {
    throw new Error(`${packageName}@${version} is not available on npm`);
  }
  assert(resolved === version, `${packageName}@${version} is not available on npm`);
}

function goTagExists(tag) {
  try {
    run('git', ['ls-remote', '--exit-code', 'origin', `refs/tags/${tag}`]);
  } catch {
    throw new Error(`Go module tag ${tag} is not available on origin`);
  }
}

const packageJson = readJson(join(packageRoot, 'package.json'));
const serverJson = readJson(join(packageRoot, 'server.json'));
const releaseJson = readJson(join(packageRoot, 'content/release.json'));
const version = packageJson.version;
const expectedMcpTag = `mcp-v${version}`;
const expectedGoTag = `v${version}`;

assert(serverJson.version === version, 'server.json version must match package.json version');
const npmPackage = serverJson.packages.find((entry) => entry.registryType === 'npm');
assert(npmPackage?.version === version, 'server.json npm package version must match package.json version');
assert(releaseJson.serverVersion === version, 'release.json serverVersion must match package.json version');
assert(releaseJson.sdkVersion === version, 'release.json sdkVersion must match package.json version');
assert(releaseJson.compatVersion === version, 'release.json compatVersion must match package.json version');
assert(releaseJson.goModuleVersion === expectedGoTag, `release.json goModuleVersion must be ${expectedGoTag}`);
assert(releaseJson.versionPolicy === 'strict-mirror', 'release.json versionPolicy must be strict-mirror');

if (process.env.GITHUB_REF_NAME) {
  assert(process.env.GITHUB_REF_NAME === expectedMcpTag, `release tag ${process.env.GITHUB_REF_NAME} must be ${expectedMcpTag}`);
}

goTagExists(expectedGoTag);
npmVersionExists('@yellow-org/sdk', version);
npmVersionExists('@yellow-org/sdk-compat', version);

console.error(`Release preflight passed for ${packageJson.name}@${version}`);
