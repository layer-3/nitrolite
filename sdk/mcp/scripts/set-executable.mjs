#!/usr/bin/env node
import { chmodSync, existsSync, statSync } from 'node:fs';
import { resolve } from 'node:path';

const target = process.argv[2];

if (!target) {
  throw new Error('Usage: node scripts/set-executable.mjs <path>');
}

const targetPath = resolve(process.cwd(), target);

if (!existsSync(targetPath)) {
  throw new Error(`Cannot make missing file executable: ${target}`);
}

if (!statSync(targetPath).isFile()) {
  throw new Error(`Executable target is not a file: ${target}`);
}

chmodSync(targetPath, 0o755);
