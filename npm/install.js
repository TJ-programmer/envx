#!/usr/bin/env node
'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');

const pkg = require('./package.json');
const pkgDir = __dirname;

const REPO = process.env.ENVX_REPO || 'TJ-programmer/envx';
const DOWNLOAD_BASE =
  process.env.ENVX_DOWNLOAD_BASE || `https://github.com/${REPO}/releases/download`;
const IS_WINDOWS = process.platform === 'win32';

function mapPlatform(p) {
  switch (p) {
    case 'darwin':
      return 'darwin';
    case 'linux':
      return 'linux';
    case 'win32':
      return 'windows';
    default:
      throw new Error(`unsupported platform: ${p}`);
  }
}

function mapArch(a) {
  switch (a) {
    case 'x64':
      return 'amd64';
    case 'arm64':
      return 'arm64';
    default:
      throw new Error(`unsupported architecture: ${a}`);
  }
}

async function fetchBuffer(url) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status} ${res.statusText}: ${url}`);
  }
  return Buffer.from(await res.arrayBuffer());
}

async function main() {
  const version = process.env.ENVX_VERSION || pkg.version;
  const osName = mapPlatform(process.platform);
  const arch = mapArch(process.arch);
  const ext = osName === 'windows' ? 'zip' : 'tar.gz';
  const archive = `envx_${version}_${osName}_${arch}.${ext}`;
  const base = `${DOWNLOAD_BASE}/v${version}`;
  const binName = IS_WINDOWS ? 'envx.exe' : 'envx';

  console.log(`envx: downloading ${archive} …`);
  const archiveBuf = await fetchBuffer(`${base}/${archive}`);
  const checksums = (await fetchBuffer(`${base}/checksums.txt`)).toString('utf8');
  const line = checksums
    .split(/\r?\n/)
    .find((l) => l.trim().endsWith(` ${archive}`));
  if (!line) {
    throw new Error(`checksum for ${archive} not found in checksums.txt`);
  }
  const expected = line.trim().split(/\s+/)[0];
  const actual = crypto.createHash('sha256').update(archiveBuf).digest('hex');
  if (expected !== actual) {
    throw new Error(
      `checksum mismatch for ${archive}\n  expected: ${expected}\n  actual:   ${actual}`
    );
  }
  console.log('envx: checksum OK');

  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'envx-'));
  const archivePath = path.join(tmp, archive);
  fs.writeFileSync(archivePath, archiveBuf);
  try {
    execSync(`tar -xzf "${archivePath}" -C "${tmp}"`, { stdio: 'inherit' });
  } finally {
    fs.rmSync(archivePath, { force: true });
  }

  const vendorDir = path.join(pkgDir, 'vendor');
  fs.mkdirSync(vendorDir, { recursive: true });
  const dest = path.join(vendorDir, binName);
  fs.copyFileSync(path.join(tmp, binName), dest);
  fs.rmSync(tmp, { recursive: true, force: true });
  if (!IS_WINDOWS) {
    fs.chmodSync(dest, 0o755);
  }
  fs.writeFileSync(path.join(vendorDir, 'version'), version);
  console.log(`envx: installed v${version} → ${dest}`);
}

main().catch((err) => {
  console.error(`envx: install failed: ${err.message}`);
  process.exit(1);
});
