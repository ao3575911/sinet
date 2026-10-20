#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const child = require('node:child_process');
const os = require('node:os');

const targets = {
  'linux-x64': 'SINet-linux-amd64',
  'darwin-arm64': 'SINet-darwin-arm64',
  'win32-x64': 'SINet-windows-amd64.exe',
};

function main() {
  const filename = targets[`${process.platform}-${process.arch}`];
  if (!filename) throw new Error('Supported targets: Linux x64, macOS arm64, Windows x64.');
  const binary = path.join(__dirname, filename);
  const manifest = JSON.parse(fs.readFileSync(path.join(__dirname, 'SHA256.json'), 'utf8'));
  if (!fs.lstatSync(binary).isFile()) throw new Error('SINet binary must be a regular file.');
  const actual = crypto.createHash('sha256').update(fs.readFileSync(binary)).digest('hex');
  if (!/^[0-9a-f]{64}$/.test(manifest[filename] || '') || actual !== manifest[filename]) {
    throw new Error('SINet binary integrity check failed; reinstall the package.');
  }
  const result = child.spawnSync(binary, process.argv.slice(2), {stdio: 'inherit', shell: false});
  if (result.error) throw new Error(`Unable to start SINet: ${result.error.code}`);
  return result.status ?? (128 + (os.constants.signals[result.signal] || 1));
}

try {
  process.exitCode = main();
} catch (error) {
  process.stderr.write(`SINet: ${error.message}\n`);
  process.exitCode = 1;
}
