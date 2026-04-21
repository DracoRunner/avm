#!/usr/bin/env node
// Post-install script: downloads the correct avm binary for the current platform

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const os = require('os');

const pkg = require('../package.json');
const REPO = 'DracoRunner/avm';
const VERSION = `v${pkg.version}`;

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  let goOs, goArch;

  if (platform === 'darwin') goOs = 'darwin';
  else if (platform === 'linux') goOs = 'linux';
  else throw new Error(`Unsupported OS: ${platform}`);

  if (arch === 'x64') goArch = 'amd64';
  else if (arch === 'arm64') goArch = 'arm64';
  else throw new Error(`Unsupported architecture: ${arch}`);

  return { goOs, goArch };
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    const request = (reqUrl) => {
      https.get(reqUrl, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          request(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: HTTP ${res.statusCode} for ${reqUrl}`));
          return;
        }
        res.pipe(file);
        file.on('finish', () => file.close(resolve));
      }).on('error', (err) => {
        fs.unlink(dest, () => {});
        reject(err);
      });
    };

    request(url);
  });
}

async function main() {
  try {
    const { goOs, goArch } = getPlatform();
    const archiveName = `avm_${goOs}_${goArch}.tar.gz`;
    const url = `https://github.com/${REPO}/releases/download/${VERSION}/${archiveName}`;

    const binDir = path.join(__dirname);
    const tarPath = path.join(binDir, archiveName);
    const finalBinary = path.join(binDir, 'avm-bin');

    console.log(`Downloading avm ${VERSION} for ${goOs}/${goArch}...`);
    await download(url, tarPath);

    console.log('Extracting...');
    execSync(`tar -xzf "${tarPath}" -C "${binDir}"`);

    // The archive contains a binary named "avm"; rename to "avm-bin"
    const extractedBinary = path.join(binDir, 'avm');
    if (fs.existsSync(extractedBinary)) {
      fs.renameSync(extractedBinary, finalBinary);
    }

    fs.chmodSync(finalBinary, 0o755);
    fs.unlinkSync(tarPath);

    // Create ~/.avm.json if missing
    const globalConfig = path.join(os.homedir(), '.avm.json');
    if (!fs.existsSync(globalConfig)) {
      fs.writeFileSync(globalConfig, '{}\n');
    }

    console.log('✓ avm installed successfully');
    console.log('');
    console.log('To enable avm in your shell, add this to ~/.zshrc or ~/.bashrc:');
    console.log('  eval "$(avm-bin shell-init)"');
    console.log('');
    console.log('Then reload: source ~/.zshrc  # or source ~/.bashrc');
  } catch (err) {
    console.error('avm install failed:', err.message);
    console.error('You can install manually from: https://github.com/DracoRunner/avm/releases');
    process.exit(1);
  }
}

main();
