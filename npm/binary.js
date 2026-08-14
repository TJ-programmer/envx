'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');

const IS_WINDOWS = process.platform === 'win32';
const binName = IS_WINDOWS ? 'envx.exe' : 'envx';

function repo() {
  return process.env.ENVX_REPO || 'TJ-programmer/envx';
}

function downloadBase() {
  return (
    process.env.ENVX_DOWNLOAD_BASE ||
    `https://github.com/${repo()}/releases/download`
  );
}

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

function resolveVersion(pkgVersion) {
  return process.env.ENVX_VERSION || pkgVersion;
}

function archiveFor(version) {
  const osName = mapPlatform(process.platform);
  const arch = mapArch(process.arch);
  const ext = osName === 'windows' ? 'zip' : 'tar.gz';
  return `envx_${version}_${osName}_${arch}.${ext}`;
}

function cacheDir() {
  if (IS_WINDOWS) {
    return path.join(
      process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'),
      'envx'
    );
  }
  return path.join(
    process.env.XDG_CACHE_HOME || path.join(os.homedir(), '.cache'),
    'envx'
  );
}

function binaryInCache(version) {
  return path.join(cacheDir(), 'bin', version, binName);
}

function binaryInPackage(pkgDir) {
  return path.join(pkgDir, 'vendor', binName);
}

async function fetchBuffer(url) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status} ${res.statusText}: ${url}`);
  }
  return Buffer.from(await res.arrayBuffer());
}

async function installBinary(destDir, version) {
  const archive = archiveFor(version);
  const base = `${downloadBase()}/v${version}`;
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

  fs.mkdirSync(destDir, { recursive: true });
  const dest = path.join(destDir, binName);
  fs.copyFileSync(path.join(tmp, binName), dest);
  fs.rmSync(tmp, { recursive: true, force: true });
  if (!IS_WINDOWS) {
    fs.chmodSync(dest, 0o755);
  }
  fs.writeFileSync(path.join(destDir, 'version'), version);
  return dest;
}

module.exports = {
  IS_WINDOWS,
  binName,
  cacheDir,
  binaryInCache,
  binaryInPackage,
  installBinary,
  resolveVersion,
  archiveFor,
};
