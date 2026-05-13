#!/usr/bin/env node
import { chmodSync, existsSync, lstatSync, statSync } from 'node:fs';
import { isAbsolute, relative, resolve } from 'node:path';

const target = process.argv[2];

if (!target) {
  throw new Error('Usage: node scripts/set-executable.mjs <path>');
}

const targetPath = resolve(process.cwd(), target);
const rel = relative(process.cwd(), targetPath);

if (rel === '' || rel.startsWith('..') || isAbsolute(rel)) {
  throw new Error(`Executable target must stay inside the package directory: ${target}`);
}

if (!existsSync(targetPath)) {
  throw new Error(`Cannot make missing file executable: ${target}`);
}

if (lstatSync(targetPath).isSymbolicLink()) {
  throw new Error(`Executable target must not be a symlink: ${target}`);
}

if (!statSync(targetPath).isFile()) {
  throw new Error(`Executable target is not a file: ${target}`);
}

chmodSync(targetPath, 0o755);
