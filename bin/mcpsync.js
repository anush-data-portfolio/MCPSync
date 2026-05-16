#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const binaryName = process.platform === 'win32' ? 'mcpsync.exe' : 'mcpsync';
// Binaries land in <package-root>/bin/ after postinstall downloads them
const binaryPath = path.join(__dirname, binaryName);

if (!fs.existsSync(binaryPath)) {
  process.stderr.write(
    [
      '',
      'mcpsync: binary not found.',
      '',
      'This usually means postinstall was skipped (e.g. --ignore-scripts).',
      'Fix it with one of:',
      '',
      '  npm install -g @anushkrishnav/mcpsync   # re-run postinstall',
      '  npx @anushkrishnav/mcpsync             # use npx (downloads automatically)',
      '',
      'Or build from source (requires Go 1.21+):',
      '  git clone https://github.com/anush-data-portfolio/MCPSync',
      '  cd mcpsync && go build -o mcpsync . && sudo mv mcpsync /usr/local/bin/',
      '',
    ].join('\n')
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status ?? 1);
