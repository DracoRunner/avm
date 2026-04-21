#!/usr/bin/env node
// Wrapper that executes the downloaded avm binary
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const binaryPath = path.join(__dirname, 'avm-bin');

if (!fs.existsSync(binaryPath)) {
  console.error('avm binary not found. Try reinstalling: npm install -g @dracorunner/avm');
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
child.on('exit', (code) => process.exit(code ?? 1));
child.on('error', (err) => {
  console.error('Failed to execute avm:', err.message);
  process.exit(1);
});
