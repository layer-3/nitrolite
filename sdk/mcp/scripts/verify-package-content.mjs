#!/usr/bin/env node
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const packageRoot = resolve(scriptDir, '..');
const packJsonPath = resolve(process.cwd(), process.argv[2] ?? 'pack.json');

function readJson(path) {
  return JSON.parse(readFileSync(path, 'utf-8'));
}

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

const packageJson = readJson(join(packageRoot, 'package.json'));
const serverJson = readJson(join(packageRoot, 'server.json'));
const releaseJson = readJson(join(packageRoot, 'content/release.json'));
const manifestJson = readJson(join(packageRoot, 'content/manifest.json'));
const packJson = readJson(packJsonPath);
const pack = Array.isArray(packJson) ? packJson[0] : packJson;
const paths = new Set(pack.files.map((file) => file.path));

const requiredPackageFiles = [
  'dist/index.js',
  'dist/index.d.ts',
  'content/release.json',
  'content/manifest.json',
  'content/docs/api.yaml',
  'content/docs/protocol/overview.md',
  'content/docs/protocol/cryptography.md',
  'content/docs/protocol/security-and-limitations.md',
  'content/sdk/ts/package.json',
  'content/sdk/ts/src/client.ts',
  'content/sdk/ts/src/index.ts',
  'content/sdk/ts/src/core/types.ts',
  'content/sdk/ts/src/rpc/methods.ts',
  'content/sdk/ts-compat/package.json',
  'content/sdk/ts-compat/src/client.ts',
  'content/sdk/ts-compat/src/index.ts',
  'content/sdk/ts-compat/docs/migration-overview.md',
  'content/sdk/go/client.go',
  'content/sdk/go/README.md',
  'content/pkg/app/app_v1.go',
  'content/pkg/core/types.go',
  'content/pkg/rpc/methods.go',
  'README.md',
  'package.json',
  'server.json',
];

for (const path of requiredPackageFiles) {
  assert(paths.has(path), `missing package file: ${path}`);
}

// Defense-in-depth against a future package.json files allowlist expansion.
for (const path of paths) {
  assert(!path.startsWith('src/'), `source-only file leaked into package: ${path}`);
  assert(!path.startsWith('scripts/'), `script file leaked into package: ${path}`);
  assert(!path.endsWith('.test.ts'), `test file leaked into package: ${path}`);
  assert(!path.endsWith('.spec.ts'), `test file leaked into package: ${path}`);
  assert(!path.endsWith('_test.go'), `test file leaked into package: ${path}`);
}

assert(serverJson.name === packageJson.mcpName, `server.json name ${serverJson.name} does not match package.json mcpName ${packageJson.mcpName}`);
const npmPackage = serverJson.packages.find((entry) => entry.registryType === 'npm');
assert(npmPackage, 'server.json is missing an npm package entry');
assert(npmPackage.identifier === packageJson.name, `server.json package ${npmPackage.identifier} does not match package.json name ${packageJson.name}`);
assert(serverJson.version === packageJson.version, 'server.json version must match package.json version');
assert(npmPackage.version === packageJson.version, 'server.json npm package version must match package.json version');

assert(releaseJson.serverPackage === packageJson.name, 'release.json serverPackage must match package.json name');
assert(releaseJson.mcpName === packageJson.mcpName, 'release.json mcpName must match package.json mcpName');
assert(releaseJson.serverVersion === packageJson.version, 'release.json serverVersion must match package.json version');
assert(releaseJson.sdkVersion === packageJson.version, 'release.json sdkVersion must match MCP package version under strict-mirror policy');
assert(releaseJson.compatVersion === packageJson.version, 'release.json compatVersion must match MCP package version under strict-mirror policy');
assert(releaseJson.goModule === 'github.com/layer-3/nitrolite', 'release.json goModule is unexpected');
assert(releaseJson.goModuleVersion === `v${packageJson.version}`, `release.json goModuleVersion must be v${packageJson.version}`);
assert(releaseJson.contentMode === 'packaged-snapshot', 'release.json contentMode must be packaged-snapshot');
assert(releaseJson.versionPolicy === 'strict-mirror', 'release.json versionPolicy must be strict-mirror');
assert(typeof releaseJson.sourceCommit === 'string' && releaseJson.sourceCommit.length > 0, 'release.json sourceCommit must be present');

assert(manifestJson.contentMode === 'packaged-snapshot', 'manifest contentMode must be packaged-snapshot');
assert(manifestJson.versionPolicy === 'strict-mirror', 'manifest versionPolicy must be strict-mirror');
assert(Array.isArray(manifestJson.files), 'manifest files must be an array');
assert(manifestJson.files.length >= 70, `manifest contains too few files: ${manifestJson.files.length}`);
assert(manifestJson.counts?.totalFiles === manifestJson.files.length, 'manifest totalFiles must match files length');

for (const file of manifestJson.files) {
  const packagePath = `content/${file.path}`;
  assert(paths.has(packagePath), `manifest file missing from package: ${packagePath}`);
  assert(typeof file.sha256 === 'string' && /^[a-f0-9]{64}$/.test(file.sha256), `manifest file has invalid sha256: ${file.path}`);
  assert(typeof file.size === 'number' && file.size >= 0, `manifest file has invalid size: ${file.path}`);
}

console.error(`Verified ${pack.filename} with ${manifestJson.files.length} manifest files`);
