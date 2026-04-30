#!/usr/bin/env node
import { cpSync, existsSync, mkdirSync, readdirSync, rmSync, statSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const packageRoot = resolve(scriptDir, '..');
const repoRoot = resolve(packageRoot, '../..');
const contentRoot = join(packageRoot, 'content');

const copySpecs = [
  { from: 'docs/api.yaml', to: 'docs/api.yaml' },
  { from: 'docs/protocol', to: 'docs/protocol', extensions: ['.md'] },
  { from: 'sdk/ts/package.json', to: 'sdk/ts/package.json' },
  { from: 'sdk/ts/src', to: 'sdk/ts/src', extensions: ['.ts'] },
  { from: 'sdk/ts-compat/package.json', to: 'sdk/ts-compat/package.json' },
  { from: 'sdk/ts-compat/src', to: 'sdk/ts-compat/src', extensions: ['.ts'] },
  { from: 'sdk/ts-compat/docs', to: 'sdk/ts-compat/docs', extensions: ['.md'] },
  { from: 'sdk/go', to: 'sdk/go', extensions: ['.go', '.md'] },
  { from: 'pkg/app', to: 'pkg/app', extensions: ['.go'] },
  { from: 'pkg/core', to: 'pkg/core', extensions: ['.go', '.md'] },
  { from: 'pkg/rpc', to: 'pkg/rpc', extensions: ['.go', '.md'] },
];

function shouldCopy(filePath, extensions) {
  if (filePath.endsWith('_test.go')) return false;
  if (filePath.endsWith('.test.ts') || filePath.endsWith('.spec.ts')) return false;
  if (!extensions) return true;
  return extensions.some((extension) => filePath.endsWith(extension));
}

function copyFilteredDirectory(sourceDir, targetDir, extensions) {
  for (const entry of readdirSync(sourceDir)) {
    if (entry === 'node_modules' || entry === 'dist') continue;
    const sourcePath = join(sourceDir, entry);
    const targetPath = join(targetDir, entry);
    const stats = statSync(sourcePath);
    if (stats.isDirectory()) {
      copyFilteredDirectory(sourcePath, targetPath, extensions);
      continue;
    }
    if (!stats.isFile() || !shouldCopy(sourcePath, extensions)) continue;
    mkdirSync(dirname(targetPath), { recursive: true });
    cpSync(sourcePath, targetPath);
  }
}

rmSync(contentRoot, { recursive: true, force: true });
mkdirSync(contentRoot, { recursive: true });

for (const spec of copySpecs) {
  const sourcePath = resolve(repoRoot, spec.from);
  const targetPath = resolve(contentRoot, spec.to);
  if (!existsSync(sourcePath)) {
    console.warn(`Skipping missing content source: ${relative(repoRoot, sourcePath)}`);
    continue;
  }
  const stats = statSync(sourcePath);
  if (stats.isDirectory()) {
    copyFilteredDirectory(sourcePath, targetPath, spec.extensions);
  } else {
    mkdirSync(dirname(targetPath), { recursive: true });
    cpSync(sourcePath, targetPath);
  }
}

console.error(`Prepared MCP package content in ${relative(packageRoot, contentRoot)}`);
