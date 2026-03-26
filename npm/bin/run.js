#!/usr/bin/env node
'use strict';

const { spawnSync } = require('node:child_process');
const { existsSync } = require('node:fs');
const path = require('node:path');

const binaryName = process.platform === 'win32' ? 'figma-mcp-go.exe' : 'figma-mcp-go';
const binaryPath = path.join(__dirname, binaryName);

if (!existsSync(binaryPath)) {
  process.stderr.write(
    '[figma-mcp-go] Binary not found. Try reinstalling: npm install @vkhanhqui/figma-mcp-go\n'
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status ?? 1);
