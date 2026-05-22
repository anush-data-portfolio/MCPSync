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
const crypto = require('crypto');

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

// ── Checksum verification ─────────────────────────────────────────────────────

/** Download a URL to a string (for small text files like checksums.txt). */
function downloadToString(url, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) return reject(new Error('Too many redirects'));
    https.get(url, { headers: { 'User-Agent': 'mcpsync-postinstall' } }, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        const location = res.headers.location;
        if (!location) return reject(new Error('Redirect with no location header'));
        const redirectUrl = new URL(location, url);
        if (!redirectUrl.hostname.endsWith('.githubusercontent.com') &&
            redirectUrl.hostname !== new URL(url).hostname) {
          return reject(new Error(`Redirect to untrusted host: ${redirectUrl.hostname}`));
        }
        return resolve(downloadToString(redirectUrl.href, redirects + 1));
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      let body = '';
      res.on('data', (chunk) => { body += chunk; });
      res.on('end', () => resolve(body));
      res.on('error', reject);
    }).on('error', reject);
  });
}

/** Parse a checksums.txt line like "<hex>  <filename>" and return the hex for a given filename. */
function parseChecksum(checksumText, filename) {
  for (const line of checksumText.split('\n')) {
    const parts = line.trim().split(/\s+/);
    if (parts.length >= 2 && parts[1] === filename) {
      return parts[0];
    }
  }
  return null;
}

/** Compute SHA-256 of a local file and return hex string. */
function fileChecksum(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash('sha256');
    const stream = fs.createReadStream(filePath);
    stream.on('data', (chunk) => hash.update(chunk));
    stream.on('end', () => resolve(hash.digest('hex')));
    stream.on('error', reject);
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
  const checksumsUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt`;
  process.stdout.write(`mcpsync: downloading binary for ${process.platform}/${process.arch}...\n`);

  // Fetch checksums file — hard-fail if unavailable (security: no verification = no install).
  let checksumText;
  try {
    checksumText = await downloadToString(checksumsUrl);
  } catch (e) {
    // Remove the binary if it was somehow partially written before we got here.
    if (fs.existsSync(destPath)) fs.unlinkSync(destPath);
    throw new Error(
      `Could not fetch checksums.txt — aborting install for security.\n` +
      `Error: ${e.message}\n\n` +
      `To install manually:\n` +
      `  1. Download a binary from: https://github.com/${REPO}/releases/tag/v${VERSION}\n` +
      `  2. Verify its SHA-256 against checksums.txt from the same release\n`
    );
  }

  try {
    await download(url, destPath);

    // Verify checksum.
    const expected = parseChecksum(checksumText, remoteName);
    if (expected) {
        const actual = await fileChecksum(destPath);
        if (actual !== expected) {
          fs.unlinkSync(destPath);
          throw new Error(
            `Checksum mismatch for ${remoteName}:\n` +
            `  expected: ${expected}\n` +
            `  actual:   ${actual}`
          );
        }
        process.stdout.write(`mcpsync: checksum verified ✓\n`);
      } else {
        warn(`No checksum entry for ${remoteName} in checksums.txt — skipping verification`);
      }

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
