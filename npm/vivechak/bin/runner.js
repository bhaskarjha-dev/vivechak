#!/usr/bin/env node

/**
 * Vivechak & vck native binary runner.
 * Resolves the native precompiled executable from local cache or downloads it lazily
 * from GitHub Releases on first execution.
 *
 * All diagnostics and progress are written to process.stderr so that process.stdout
 * remains 100% reserved for MCP JSON-RPC protocol communication.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const http = require('http');
const { spawn, spawnSync } = require('child_process');

const VERSION = '0.1.0';
const REPO = 'bhaskarjha-dev/vivechak';

function getPlatformInfo() {
  const p = process.platform;
  const a = process.arch;

  let osName = '';
  if (p === 'darwin') osName = 'darwin';
  else if (p === 'linux') osName = 'linux';
  else if (p === 'win32') osName = 'windows';
  else if (p === 'freebsd') osName = 'freebsd';
  else throw new Error(`Unsupported operating system: ${p}`);

  let archName = '';
  if (a === 'x64') archName = 'amd64';
  else if (a === 'arm64') archName = 'arm64';
  else throw new Error(`Unsupported architecture: ${a}`);

  return { osName, archName };
}

function getCacheDir() {
  if (process.env.VIVECHAK_CACHE_DIR) {
    return path.join(process.env.VIVECHAK_CACHE_DIR, `v${VERSION}`);
  }
  const base = process.platform === 'win32'
    ? (process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'))
    : (process.env.XDG_CACHE_HOME || path.join(os.homedir(), '.cache'));
  return path.join(base, 'vivechak', `v${VERSION}`);
}

function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    const proto = url.startsWith('https') ? https : http;
    const req = proto.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return downloadFile(res.headers.location, destPath).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Download failed with HTTP status ${res.statusCode} from ${url}`));
      }
      const fileStream = fs.createWriteStream(destPath);
      res.pipe(fileStream);
      fileStream.on('finish', () => fileStream.close(resolve));
      fileStream.on('error', reject);
      res.on('error', reject);
    });
    req.setTimeout(30000, () => {
      req.destroy();
      reject(new Error(`Download timed out after 30 seconds from ${url}`));
    });
    req.on('error', reject);
  });
}

function extractArchive(archivePath, targetDir) {
  const ext = path.extname(archivePath).toLowerCase();
  // Reserve stdout exclusively for JSON-RPC in MCP mode; redirect child stdout to stderr
  const childStdio = ['ignore', 2, 2];
  if (ext === '.zip') {
    // Windows 10/11 native tar.exe or PowerShell Expand-Archive
    const res = spawnSync('tar', ['-xf', archivePath, '-C', targetDir], { stdio: childStdio });
    if (res.status !== 0) {
      // Fallback for older Windows without tar.exe
      const ps = spawnSync('powershell', ['-NoProfile', '-Command', `Expand-Archive -Path '${archivePath}' -DestinationPath '${targetDir}' -Force`], { stdio: childStdio });
      if (ps.status !== 0) {
        throw new Error(`Failed to extract ZIP archive ${archivePath}`);
      }
    }
  } else {
    // Unix tar handles tar.gz natively
    const res = spawnSync('tar', ['-xzf', archivePath, '-C', targetDir], { stdio: childStdio });
    if (res.status !== 0) {
      throw new Error(`Failed to extract tar.gz archive ${archivePath}`);
    }
  }
}

async function ensureBinary() {
  const { osName, archName } = getPlatformInfo();
  const cacheDir = getCacheDir();
  const isWin = osName === 'windows';

  // Determine requested binary (vck vs vivechak)
  const invokedName = path.basename(process.argv[1] || '').toLowerCase();
  const prog = invokedName.includes('vck') ? 'vck' : 'vivechak';
  const binName = isWin ? `${prog}.exe` : prog;
  const targetBinPath = path.join(cacheDir, binName);

  if (fs.existsSync(targetBinPath)) {
    return targetBinPath;
  }

  // Check if we are running in the development repository with local Go installed
  const repoRoot = path.resolve(__dirname, '..', '..', '..');
  const localMainGo = path.join(repoRoot, 'cmd', 'vivechak', 'main.go');
  if (fs.existsSync(localMainGo)) {
    const localBinPath = path.join(cacheDir, binName);
    fs.mkdirSync(cacheDir, { recursive: true });
    process.stderr.write(`[vck] Compiling local development binary in ${cacheDir}...\n`);
    const buildRes = spawnSync('go', ['build', '-trimpath', '-o', localBinPath, './cmd/vivechak'], { cwd: repoRoot, stdio: ['ignore', 2, 2] });
    if (buildRes.status === 0 && fs.existsSync(localBinPath)) {
      if (!isWin) fs.chmodSync(localBinPath, 0o755);
      return localBinPath;
    }
  }

  // Determine archive name
  let archiveName = '';
  if (osName === 'darwin') {
    archiveName = `vivechak_${VERSION}_darwin_all.tar.gz`;
  } else if (osName === 'windows') {
    archiveName = `vivechak_${VERSION}_windows_${archName}.zip`;
  } else {
    archiveName = `vivechak_${VERSION}_${osName}_${archName}.tar.gz`;
  }

  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
  fs.mkdirSync(cacheDir, { recursive: true });

  const tempArchive = path.join(cacheDir, archiveName);
  process.stderr.write(`[vck] Fetching Vivechak v${VERSION} for ${osName}/${archName}...\n`);

  try {
    await downloadFile(downloadUrl, tempArchive);
    extractArchive(tempArchive, cacheDir);
    if (!isWin) {
      const vckPath = path.join(cacheDir, 'vck');
      const vivPath = path.join(cacheDir, 'vivechak');
      if (fs.existsSync(vckPath)) fs.chmodSync(vckPath, 0o755);
      if (fs.existsSync(vivPath)) fs.chmodSync(vivPath, 0o755);
    }
    try { fs.unlinkSync(tempArchive); } catch (_) {}
  } catch (err) {
    try { fs.unlinkSync(tempArchive); } catch (_) {}
    throw new Error(`Failed to download and setup Vivechak: ${err.message}`);
  }

  if (!fs.existsSync(targetBinPath)) {
    // If only vivechak was extracted, link/copy to vck or vice-versa
    const altBinName = isWin ? (prog === 'vck' ? 'vivechak.exe' : 'vck.exe') : (prog === 'vck' ? 'vivechak' : 'vck');
    const altPath = path.join(cacheDir, altBinName);
    if (fs.existsSync(altPath)) {
      fs.copyFileSync(altPath, targetBinPath);
      if (!isWin) fs.chmodSync(targetBinPath, 0o755);
    } else {
      throw new Error(`Binary ${binName} was not found after archive extraction in ${cacheDir}`);
    }
  }

  return targetBinPath;
}

async function main() {
  try {
    const binPath = await ensureBinary();
    const args = process.argv.slice(2);
    const child = spawn(binPath, args, { stdio: 'inherit' });

    child.on('error', (err) => {
      process.stderr.write(`[vck] Failed to start native process: ${err.message}\n`);
      process.exit(1);
    });

    const forwardSignal = (sig) => {
      if (child.pid) {
        try { child.kill(sig); } catch (_) {}
      }
    };
    process.on('SIGINT', () => forwardSignal('SIGINT'));
    process.on('SIGTERM', () => forwardSignal('SIGTERM'));

    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      }
      process.exit(code ?? 0);
    });
  } catch (err) {
    process.stderr.write(`Error: ${err.message}\n`);
    process.exit(1);
  }
}

main();
