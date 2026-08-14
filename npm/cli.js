#!/usr/bin/env node
'use strict';

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const bin = path.join(
  __dirname,
  'vendor',
  process.platform === 'win32' ? 'envx.exe' : 'envx'
);

if (!fs.existsSync(bin)) {
  console.error('envx: binary not found. Reinstall the package to download it.');
  process.exit(1);
}

const child = spawn(bin, process.argv.slice(2), { stdio: 'inherit' });

child.on('exit', (code, signal) => {
  if (signal && process.platform !== 'win32') {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code == null ? 1 : code);
});
