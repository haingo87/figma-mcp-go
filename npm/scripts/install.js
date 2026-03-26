#!/usr/bin/env node
'use strict';

const https = require('node:https');
const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const os = require('node:os');

const pkg = require('../package.json');
const version = pkg.version;

const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64',
};

const goos = PLATFORM_MAP[process.platform];
const goarch = ARCH_MAP[process.arch];

if (!goos || !goarch) {
  console.error(`[figma-mcp-go] Unsupported platform: ${process.platform}/${process.arch}`);
  process.exit(1);
}

const isWindows = goos === 'windows';
const binaryName = isWindows ? 'figma-mcp-go.exe' : 'figma-mcp-go';
const ext = isWindows ? '.zip' : '.tar.gz';
const archiveName = `figma-mcp-go_${goos}_${goarch}${ext}`;
const downloadUrl = `https://github.com/vkhanhqui/figma-mcp-go/releases/download/v${version}/${archiveName}`;
const binDir = path.join(__dirname, '..', 'bin');
const tmpFile = path.join(os.tmpdir(), archiveName);

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    function get(u) {
      https.get(u, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          file.close();
          fs.rmSync(dest, { force: true });
          reject(new Error(`HTTP ${res.statusCode} downloading ${u}`));
          return;
        }
        res.pipe(file);
        file.on('finish', () => file.close(resolve));
        file.on('error', reject);
      }).on('error', reject);
    }
    get(url);
  });
}

async function main() {
  console.log(`[figma-mcp-go] Downloading ${archiveName}...`);

  try {
    await download(downloadUrl, tmpFile);
  } catch (err) {
    console.error(`[figma-mcp-go] Failed to download: ${err.message}`);
    console.error(`[figma-mcp-go] URL: ${downloadUrl}`);
    process.exit(1);
  }

  fs.mkdirSync(binDir, { recursive: true });

  try {
    if (isWindows) {
      execFileSync('powershell', [
        '-NoProfile', '-Command',
        `Expand-Archive -Force -Path "${tmpFile}" -DestinationPath "${binDir}"`,
      ]);
    } else {
      execFileSync('tar', ['xzf', tmpFile, '-C', binDir, binaryName]);
      fs.chmodSync(path.join(binDir, binaryName), 0o755);
    }
  } catch (err) {
    console.error(`[figma-mcp-go] Failed to extract: ${err.message}`);
    process.exit(1);
  } finally {
    fs.rmSync(tmpFile, { force: true });
  }

  console.log('[figma-mcp-go] Installed successfully.');
}

main();
