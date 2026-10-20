#!/usr/bin/env node
'use strict';
const { spawnSync } = require('node:child_process');
const { join } = require('node:path');
const platforms = {
  'linux-x64': 'linux-amd64/SINet',
  'darwin-arm64': 'darwin-arm64/SINet',
  'win32-x64': 'windows-amd64/SINet.exe',
};
const target = platforms[`${process.platform}-${process.arch}`];
if (!target) {
  console.error(`SINet does not support ${process.platform}/${process.arch}. Supported: Linux x64, macOS arm64, Windows x64.`);
  process.exit(1);
}
const result = spawnSync(join(__dirname, 'binaries', target), process.argv.slice(2), { stdio: 'inherit' });
if (result.error) {
  console.error(`Unable to start SINet: ${result.error.message}`);
  process.exit(1);
}
if (result.signal) process.kill(process.pid, result.signal);
else process.exit(result.status ?? 1);
