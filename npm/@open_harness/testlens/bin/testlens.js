#!/usr/bin/env node
'use strict';

const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');

const PLATFORMS = {
  'linux-x64':    '@open_harness/testlens-linux-x64',
  'darwin-arm64': '@open_harness/testlens-darwin-arm64',
  'darwin-x64':   '@open_harness/testlens-darwin-x64',
  'win32-x64':    '@open_harness/testlens-win32-x64',
};

function getBinaryPath() {
  const platform = `${os.platform()}-${os.arch()}`;
  const pkg = PLATFORMS[platform];

  if (!pkg) {
    throw new Error(`testlens: plataforma no soportada: ${platform}`);
  }

  try {
    const pkgDir = path.dirname(require.resolve(`${pkg}/package.json`));
    const ext = os.platform() === 'win32' ? '.exe' : '';
    return path.join(pkgDir, 'bin', `testlens${ext}`);
  } catch {
    throw new Error(
      `testlens: paquete de plataforma "${pkg}" no instalado.\n` +
      `Intenta: npm install --save-dev @open_harness/testlens`
    );
  }
}

try {
  const bin = getBinaryPath();
  const args = process.argv.slice(2);
  execFileSync(bin, args, { stdio: 'inherit' });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  console.error(err.message);
  process.exit(1);
}
