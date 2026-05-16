#!/usr/bin/env node
'use strict';

/**
 * Downloads the pre-built mcpsync binary for the current platform from GitHub Releases.
 * Called automatically by `npm install` via the "postinstall" script.
 *
 * Does nothing if the binary already exists (idempotent).
 * Exits 0 even on failure — the error message tells the user what to do next.
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');

const pkg = require('../package.json');
const VERSION = pkg.version;
const REPO = 'anush-data-portfolio/MCPSync';

// ── Platform detection ────────────────────────────────────────────────────────

function getPlatformBinaryName() {
  const platform = process.platform; // darwin | linux | win32
  const arch = process.arch;         // arm64  | x64

  const platformMap = {
    'darwin-arm64': 'mcpsync-darwin-arm64',
    'darwin-x64':   'mcpsync-darwin-amd64',
    'linux-x64':    'mcpsync-linux-amd64',
    'linux-arm64':  'mcpsync-linux-arm64',
    'win32-x64':    'mcpsync-windows-amd64.exe',
  };

  const key = `${platform}-${arch}`;
  const name = platformMap[key];
  if (!name) {
    throw new Error(
      `Unsupported platform: ${key}\n` +
      `Build from source: https://github.com/${REPO}#quick-start`
    );
  }
  return name;
}

// ── Download ──────────────────────────────────────────────────────────────────

function download(url, destPath, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) return reject(new Error('Too many redirects'));

    https.get(url, { headers: { 'User-Agent': 'mcpsync-postinstall' } }, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        const location = res.headers.location;
        if (!location) return reject(new Error('Redirect with no location header'));
        const redirectUrl = new URL(location, url);
        const origHost = new URL(url).hostname;
        if (redirectUrl.hostname !== origHost && !redirectUrl.hostname.endsWith('.githubusercontent.com')) {
          return reject(new Error(`Redirect to untrusted host: ${redirectUrl.hostname}`));
        }
        return resolve(download(redirectUrl.href, destPath, redirects + 1));
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode} from ${url}`));
      }

      const tmp = destPath + '.download';
      const file = fs.createWriteStream(tmp);
      res.pipe(file);

      file.on('finish', () => {
        file.close(() => {
          fs.renameSync(tmp, destPath);
          resolve();
        });
      });
      file.on('error', (err) => {
        fs.unlink(tmp, () => {});
        reject(err);
      });
    }).on('error', reject);
  });
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  const binaryName = process.platform === 'win32' ? 'mcpsync.exe' : 'mcpsync';
  const destPath = path.join(__dirname, '..', 'bin', binaryName);

  // Already installed — nothing to do.
  if (fs.existsSync(destPath)) {
    return;
  }

  let remoteName;
  try {
    remoteName = getPlatformBinaryName();
  } catch (e) {
    warn(e.message);
    return;
  }

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${remoteName}`;
  process.stdout.write(`mcpsync: downloading binary for ${process.platform}/${process.arch}...\n`);

  try {
    await download(url, destPath);
    fs.chmodSync(destPath, 0o755);
    process.stdout.write(`mcpsync: installed to ${destPath}\n`);
  } catch (e) {
    warn(
      `Could not download binary: ${e.message}\n` +
      `\n` +
      `To install manually, choose one option:\n` +
      `\n` +
      `  1. Build from source (requires Go 1.21+):\n` +
      `       git clone https://github.com/${REPO}\n` +
      `       cd mcpsync && go build -o mcpsync . && sudo mv mcpsync /usr/local/bin/\n` +
      `\n` +
      `  2. Download a binary from:\n` +
      `       https://github.com/${REPO}/releases/tag/v${VERSION}\n`
    );
  }
}

function warn(msg) {
  process.stderr.write(`\nWarning: ${msg}\n\n`);
}

main();
