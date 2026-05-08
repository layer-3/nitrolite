#!/usr/bin/env node
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { cpSync, existsSync, lstatSync, mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { dirname, isAbsolute, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const packageRoot = resolve(scriptDir, '..');
const repoRoot = resolve(packageRoot, '../..');
const contentRoot = join(packageRoot, 'content');
const goModulePath = 'github.com/layer-3/nitrolite';

const copySpecs = [
  { from: 'docs/api.yaml', to: 'docs/api.yaml', required: true },
  { from: 'docs/protocol', to: 'docs/protocol', extensions: ['.md'], required: true },
  { from: 'sdk/ts/package.json', to: 'sdk/ts/package.json', required: true },
  { from: 'sdk/ts/src', to: 'sdk/ts/src', extensions: ['.ts'], required: true },
  { from: 'sdk/ts-compat/package.json', to: 'sdk/ts-compat/package.json', required: true },
  { from: 'sdk/ts-compat/src', to: 'sdk/ts-compat/src', extensions: ['.ts'], required: true },
  { from: 'sdk/ts-compat/docs', to: 'sdk/ts-compat/docs', extensions: ['.md'], required: true },
  { from: 'sdk/go', to: 'sdk/go', extensions: ['.go', '.md'], required: true },
  { from: 'pkg/app', to: 'pkg/app', extensions: ['.go'], required: true },
  { from: 'pkg/core', to: 'pkg/core', extensions: ['.go', '.md'], required: true },
  { from: 'pkg/rpc', to: 'pkg/rpc', extensions: ['.go', '.md'], required: true },
];

function toPosixPath(path) {
  return path.split(sep).join('/');
}

function readJson(path) {
  return JSON.parse(readFileSync(path, 'utf-8'));
}

function assertWithin(root, path) {
  const rel = relative(root, path);
  if (rel === '' || (!rel.startsWith('..') && !isAbsolute(rel))) return;
  throw new Error(`Refusing to write outside ${relative(repoRoot, root)}: ${path}`);
}

function getSourceCommit() {
  if (process.env.GITHUB_SHA) return process.env.GITHUB_SHA;
  try {
    return execFileSync('git', ['rev-parse', 'HEAD'], {
      cwd: repoRoot,
      encoding: 'utf-8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
  } catch {
    return 'unknown';
  }
}

function shouldCopy(filePath, extensions) {
  if (filePath.endsWith('_test.go')) return false;
  if (filePath.endsWith('.test.ts') || filePath.endsWith('.spec.ts')) return false;
  if (!extensions) return true;
  return extensions.some((extension) => filePath.endsWith(extension));
}

function copyFilteredDirectory(sourceDir, targetDir, extensions) {
  let copied = 0;
  for (const entry of readdirSync(sourceDir).sort()) {
    if (entry === 'node_modules' || entry === 'dist') continue;
    const sourcePath = join(sourceDir, entry);
    const targetPath = join(targetDir, entry);
    assertWithin(contentRoot, targetPath);
    const linkStats = lstatSync(sourcePath);
    if (linkStats.isSymbolicLink()) {
      throw new Error(`Refusing to package symlink: ${relative(repoRoot, sourcePath)}`);
    }
    const stats = statSync(sourcePath);
    if (stats.isDirectory()) {
      copied += copyFilteredDirectory(sourcePath, targetPath, extensions);
      continue;
    }
    if (!stats.isFile() || !shouldCopy(sourcePath, extensions)) continue;
    mkdirSync(dirname(targetPath), { recursive: true });
    cpSync(sourcePath, targetPath);
    copied += 1;
  }
  return copied;
}

function collectContentFiles(root) {
  const files = [];
  function walk(dir) {
    for (const entry of readdirSync(dir).sort()) {
      const path = join(dir, entry);
      const stats = statSync(path);
      if (stats.isDirectory()) {
        walk(path);
        continue;
      }
      if (!stats.isFile()) continue;
      const bytes = readFileSync(path);
      files.push({
        path: toPosixPath(relative(root, path)),
        size: stats.size,
        sha256: createHash('sha256').update(bytes).digest('hex'),
      });
    }
  }
  walk(root);
  return files;
}

function countByPrefix(files) {
  const counts = {};
  for (const file of files) {
    const [first, second] = file.path.split('/');
    const prefix = second ? `${first}/${second}` : first;
    counts[prefix] = (counts[prefix] ?? 0) + 1;
  }
  return counts;
}

rmSync(contentRoot, { recursive: true, force: true });
mkdirSync(contentRoot, { recursive: true });

const copyResults = [];
for (const spec of copySpecs) {
  const sourcePath = resolve(repoRoot, spec.from);
  const targetPath = resolve(contentRoot, spec.to);
  assertWithin(repoRoot, sourcePath);
  assertWithin(contentRoot, targetPath);
  if (!existsSync(sourcePath)) {
    if (spec.required) {
      throw new Error(`Missing required content source: ${relative(repoRoot, sourcePath)}`);
    }
    copyResults.push({ from: spec.from, to: spec.to, required: false, files: 0 });
    continue;
  }
  const stats = statSync(sourcePath);
  const linkStats = lstatSync(sourcePath);
  if (linkStats.isSymbolicLink()) {
    throw new Error(`Refusing to package symlink: ${relative(repoRoot, sourcePath)}`);
  }
  let files = 0;
  if (stats.isDirectory()) {
    files = copyFilteredDirectory(sourcePath, targetPath, spec.extensions);
  } else {
    mkdirSync(dirname(targetPath), { recursive: true });
    cpSync(sourcePath, targetPath);
    files = 1;
  }
  if (spec.required && files === 0) {
    throw new Error(`Required content source copied no files: ${relative(repoRoot, sourcePath)}`);
  }
  copyResults.push({ from: spec.from, to: spec.to, required: spec.required !== false, files });
}

const mcpPackage = readJson(join(packageRoot, 'package.json'));
const sdkPackage = readJson(resolve(repoRoot, 'sdk/ts/package.json'));
const compatPackage = readJson(resolve(repoRoot, 'sdk/ts-compat/package.json'));

if (mcpPackage.version !== sdkPackage.version || mcpPackage.version !== compatPackage.version) {
  throw new Error(
    `Strict MCP version policy requires package versions to match: ${mcpPackage.name}@${mcpPackage.version}, ${sdkPackage.name}@${sdkPackage.version}, ${compatPackage.name}@${compatPackage.version}`,
  );
}

const release = {
  schemaVersion: 1,
  serverPackage: mcpPackage.name,
  mcpName: mcpPackage.mcpName,
  serverVersion: mcpPackage.version,
  sdkPackage: sdkPackage.name,
  sdkVersion: sdkPackage.version,
  compatPackage: compatPackage.name,
  compatVersion: compatPackage.version,
  goModule: goModulePath,
  goModuleVersion: `v${mcpPackage.version}`,
  sourceCommit: getSourceCommit(),
  generatedAt: new Date().toISOString(),
  contentMode: 'packaged-snapshot',
  versionPolicy: 'strict-mirror',
};

writeFileSync(join(contentRoot, 'release.json'), `${JSON.stringify(release, null, 2)}\n`);

const files = collectContentFiles(contentRoot);
const manifest = {
  schemaVersion: 1,
  generatedAt: release.generatedAt,
  sourceCommit: release.sourceCommit,
  contentMode: release.contentMode,
  versionPolicy: release.versionPolicy,
  copyResults,
  counts: {
    totalFiles: files.length,
    byPrefix: countByPrefix(files),
  },
  files,
};

writeFileSync(join(contentRoot, 'manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`);

console.error(`Prepared MCP package content in ${relative(packageRoot, contentRoot)} with ${files.length} files`);
