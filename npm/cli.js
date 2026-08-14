#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

const pkg = require('./package.json');
const bin = require('./binary.js');

async function resolveBin() {
  const version = bin.resolveVersion(pkg.version);
  const cached = bin.binaryInCache(version);
  if (fs.existsSync(cached)) {
    return cached;
  }
  const vendored = bin.binaryInPackage(__dirname);
  if (fs.existsSync(vendored)) {
    return vendored;
  }
  console.error(`envx: prebuilt binary not found — downloading ${bin.archiveFor(version)} …`);
  return bin.installBinary(path.join(bin.cacheDir(), 'bin', version), version);
}

async function main() {
  const binary = await resolveBin();
  const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });

  child.on('exit', (code, signal) => {
    if (signal && !bin.IS_WINDOWS) {
      process.kill(process.pid, signal);
      return;
    }
    process.exit(code == null ? 1 : code);
  });
}

main().catch((err) => {
  console.error(`envx: ${err.message}`);
  process.exit(1);
});
