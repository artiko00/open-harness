#!/usr/bin/env node
'use strict';

const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');

const PLATFORMS = {
  'linux-x64':    '@open-harness/secretlens-linux-x64',
  'darwin-arm64': '@open-harness/secretlens-darwin-arm64',
  'darwin-x64':   '@open-harness/secretlens-darwin-x64',
  'win32-x64':    '@open-harness/secretlens-win32-x64',
};

function getBinaryPath() {
  const platform = `${os.platform()}-${os.arch()}`;
  const pkg = PLATFORMS[platform];

  if (!pkg) {
    throw new Error(`secretlens: plataforma no soportada: ${platform}`);
  }

  try {
    const pkgDir = path.dirname(require.resolve(`${pkg}/package.json`));
    const ext = os.platform() === 'win32' ? '.exe' : '';
    return path.join(pkgDir, 'bin', `secretlens${ext}`);
  } catch {
    throw new Error(
      `secretlens: paquete de plataforma "${pkg}" no instalado.\n` +
      `Intenta: npm install --save-dev @open-harness/secretlens`
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
