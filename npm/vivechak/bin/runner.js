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
const crypto = require('crypto');
const https = require('https');
const http = require('http');
const { spawn, spawnSync } = require('child_process');

const VERSION = '0.1.0';
const DEFAULT_DOWNLOAD_BASE = 'https://github.com/bhaskarjha-dev/vivechak/releases/download';

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

function getDownloadBase() {
  const customBase = process.env.VIVECHAK_DOWNLOAD_BASE;
  if (!customBase) {
    return DEFAULT_DOWNLOAD_BASE;
  }
  try {
    const parsed = new URL(customBase);
    if (parsed.protocol === 'http:') {
      const host = parsed.hostname.toLowerCase();
      if (host !== '127.0.0.1' && host !== 'localhost') {
        throw new Error(`Insecure HTTP download base only allowed for 127.0.0.1 or localhost, got: ${host}`);
      }
    } else if (parsed.protocol !== 'https:') {
      throw new Error(`Unsupported download protocol: ${parsed.protocol}`);
    }
    return customBase.replace(/\/+$/, '');
  } catch (err) {
    throw new Error(`Invalid VIVECHAK_DOWNLOAD_BASE (${customBase}): ${err.message}`);
  }
}

function downloadFile(url, destPath, redirectCount = 0) {
  if (redirectCount > 5) {
    return Promise.reject(new Error(`Too many redirects (limit 5) downloading ${url}`));
  }

  return new Promise((resolve, reject) => {
    let parsedUrl;
    try {
      parsedUrl = new URL(url);
    } catch (e) {
      return reject(new Error(`Invalid URL: ${url}`));
    }

    const proto = parsedUrl.protocol === 'https:' ? https : http;
    const req = proto.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        const nextUrl = new URL(res.headers.location, url).toString();
        // Never allow downgrade from https: to http:
        if (url.startsWith('https:') && nextUrl.startsWith('http:')) {
          return reject(new Error(`Insecure redirect downgrade from HTTPS to HTTP forbidden: ${nextUrl}`));
        }
        res.resume(); // Discard redirect body
        return downloadFile(nextUrl, destPath, redirectCount + 1).then(resolve).catch(reject);
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

    req.on('error', (err) => {
      process.stderr.write(`Download failed. Behind a proxy? Use Node 24+ with NODE_USE_ENV_PROXY=1, or install with install.sh.\n`);
      reject(err);
    });
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

  // Determine archive name
  let archiveName = '';
  if (osName === 'darwin') {
    archiveName = `vivechak_${VERSION}_darwin_all.tar.gz`;
  } else if (osName === 'windows') {
    archiveName = `vivechak_${VERSION}_windows_${archName}.zip`;
  } else {
    archiveName = `vivechak_${VERSION}_${osName}_${archName}.tar.gz`;
  }

  const base = getDownloadBase();
  const downloadUrl = `${base}/v${VERSION}/${archiveName}`;
  const checksumUrl = `${base}/v${VERSION}/checksums.txt`;

  fs.mkdirSync(cacheDir, { recursive: true });

  const randomId = crypto.randomBytes(6).toString('hex');
  const tempExtractDir = path.join(cacheDir, `.tmp-${process.pid}-${randomId}`);
  const tempArchive = path.join(cacheDir, `${archiveName}.${process.pid}-${randomId}.tmp`);
  const tempChecksum = path.join(cacheDir, `checksums.${process.pid}-${randomId}.tmp`);

  process.stderr.write(`[vck] Fetching Vivechak v${VERSION} for ${osName}/${archName}...\n`);

  try {
    await downloadFile(downloadUrl, tempArchive);
    await downloadFile(checksumUrl, tempChecksum);

    // Compute SHA-256 of downloaded archive
    const fileBuffer = fs.readFileSync(tempArchive);
    const hash = crypto.createHash('sha256').update(fileBuffer).digest('hex');

    // Parse checksums.txt
    const checksums = fs.readFileSync(tempChecksum, 'utf8');
    const lines = checksums.split(/\r?\n/);
    let expectedHash = null;
    for (const line of lines) {
      const parts = line.trim().split(/\s+/);
      if (parts.length >= 2 && parts[1] === archiveName) {
        expectedHash = parts[0];
        break;
      }
    }

    if (!expectedHash) {
      throw new Error(`Checksum entry for ${archiveName} not found in checksums.txt`);
    }

    if (hash.toLowerCase() !== expectedHash.toLowerCase()) {
      throw new Error(`Checksum verification failed for ${archiveName}! Expected: ${expectedHash}, computed: ${hash}`);
    }

    process.stderr.write(`[vck] Checksum verified: ${hash.slice(0, 8)}...\n`);

    // Safe atomic extraction into temporary directory
    fs.mkdirSync(tempExtractDir, { recursive: true });
    extractArchive(tempArchive, tempExtractDir);

    // Set permissions on binaries in temporary directory
    if (!isWin) {
      for (const name of ['vck', 'vivechak']) {
        const bp = path.join(tempExtractDir, name);
        if (fs.existsSync(bp)) fs.chmodSync(bp, 0o755);
      }
    }

    // Atomically move extracted binaries into cacheDir
    const entries = fs.readdirSync(tempExtractDir);
    for (const entry of entries) {
      const srcPath = path.join(tempExtractDir, entry);
      const dstPath = path.join(cacheDir, entry);
      if (fs.statSync(srcPath).isFile()) {
        try {
          fs.renameSync(srcPath, dstPath);
        } catch (renameErr) {
          // If destination exists or cross-device, copy then unlink
          try {
            fs.copyFileSync(srcPath, dstPath);
            fs.unlinkSync(srcPath);
          } catch (_) {}
        }
      }
    }

    if (!isWin) {
      const vckPath = path.join(cacheDir, 'vck');
      const vivPath = path.join(cacheDir, 'vivechak');
      if (fs.existsSync(vckPath)) fs.chmodSync(vckPath, 0o755);
      if (fs.existsSync(vivPath)) fs.chmodSync(vivPath, 0o755);
    }
  } catch (err) {
    throw new Error(`Failed to download and setup Vivechak: ${err.message}`);
  } finally {
    try { fs.unlinkSync(tempArchive); } catch (_) {}
    try { fs.unlinkSync(tempChecksum); } catch (_) {}
    try { fs.rmSync(tempExtractDir, { recursive: true, force: true }); } catch (_) {}
  }

  if (!fs.existsSync(targetBinPath)) {
    // If only vivechak was extracted, link/copy to vck or vice-versa
    const altBinName = isWin ? (prog === 'vck' ? 'vivechak.exe' : 'vck.exe') : (prog === 'vck' ? 'vivechak' : 'vck');
    const altPath = path.join(cacheDir, altBinName);
    if (fs.existsSync(altPath)) {
      try {
        fs.copyFileSync(altPath, targetBinPath);
        if (!isWin) fs.chmodSync(targetBinPath, 0o755);
      } catch (_) {}
    }
    if (!fs.existsSync(targetBinPath)) {
      throw new Error(`Binary ${binName} was not found after archive extraction in ${cacheDir}`);
    }
  }

  return targetBinPath;
}

async function main() {
  try {
    const binPath = await ensureBinary();
    const args = process.argv.slice(2);
    const child = spawn(binPath, args, {
      stdio: 'inherit',
      env: { ...process.env, VIVECHAK_LAUNCHER: 'npx' }
    });

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
      if (signal === 'SIGINT') process.exit(130);
      if (signal === 'SIGTERM') process.exit(143);
      if (signal) process.exit(1);
      process.exit(code ?? 0);
    });
  } catch (err) {
    process.stderr.write(`Error: ${err.message}\n`);
    process.exit(1);
  }
}

main();
