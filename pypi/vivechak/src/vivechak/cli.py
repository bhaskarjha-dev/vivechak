"""Native binary runner for Vivechak and vck.
All diagnostic and status output is directed strictly to sys.stderr so that
sys.stdout remains 100% clean for MCP JSON-RPC protocol communication.
"""

import os
import sys
import platform
import hashlib
import urllib.request
import urllib.parse
import urllib.error
import tarfile
import zipfile
import subprocess
import secrets
import shutil
from pathlib import Path

VERSION = "0.1.0"
DEFAULT_DOWNLOAD_BASE = "https://github.com/bhaskarjha-dev/vivechak/releases/download"

def get_platform_info():
    sys_plat = sys.platform
    mach = platform.machine().lower()

    if sys_plat.startswith("linux"):
        os_name = "linux"
    elif sys_plat == "darwin":
        os_name = "darwin"
    elif sys_plat in ("win32", "cygwin"):
        os_name = "windows"
    elif sys_plat.startswith("freebsd"):
        os_name = "freebsd"
    else:
        raise RuntimeError(f"Unsupported operating system: {sys_plat}")

    if mach in ("x86_64", "amd64"):
        arch_name = "amd64"
    elif mach in ("arm64", "aarch64"):
        arch_name = "arm64"
    else:
        raise RuntimeError(f"Unsupported architecture: {mach}")

    return os_name, arch_name

def get_cache_dir():
    cache_base = os.environ.get("VIVECHAK_CACHE_DIR")
    if cache_base:
        return Path(cache_base) / f"v{VERSION}"
    if sys.platform in ("win32", "cygwin"):
        base = Path(os.environ.get("LOCALAPPDATA", Path.home() / "AppData" / "Local"))
    else:
        base = Path(os.environ.get("XDG_CACHE_HOME", Path.home() / ".cache"))
    return base / "vivechak" / f"v{VERSION}"

def get_download_base():
    custom_base = os.environ.get("VIVECHAK_DOWNLOAD_BASE")
    if not custom_base:
        return DEFAULT_DOWNLOAD_BASE
    parsed = urllib.parse.urlparse(custom_base)
    if parsed.scheme == "http":
        host = (parsed.hostname or "").lower()
        if host not in ("127.0.0.1", "localhost"):
            raise RuntimeError(f"Insecure HTTP download base only allowed for 127.0.0.1 or localhost, got: {host}")
    elif parsed.scheme != "https":
        raise RuntimeError(f"Unsupported download protocol: {parsed.scheme}")
    return custom_base.rstrip("/")

class SafeRedirectHandler(urllib.request.HTTPRedirectHandler):
    def __init__(self, max_redirects=5):
        super().__init__()
        self.max_redirects = max_redirects
        self.count = 0

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        self.count += 1
        if self.count > self.max_redirects:
            raise urllib.error.HTTPError(req.full_url, code, f"Too many redirects (limit {self.max_redirects})", headers, fp)
        # Block HTTPS to HTTP downgrade
        if req.full_url.startswith("https://") and newurl.startswith("http://"):
            raise urllib.error.HTTPError(req.full_url, code, f"Insecure redirect downgrade from HTTPS to HTTP forbidden: {newurl}", headers, fp)
        return super().redirect_request(req, fp, code, msg, headers, newurl)

def download_file(url, dest_path):
    opener = urllib.request.build_opener(SafeRedirectHandler(max_redirects=5))
    req = urllib.request.Request(url, headers={"User-Agent": f"vivechak-runner/{VERSION}"})
    try:
        with opener.open(req, timeout=30.0) as resp:
            if resp.status != 200:
                raise RuntimeError(f"Download failed with HTTP status {resp.status} from {url}")
            with open(dest_path, "wb") as f:
                while True:
                    chunk = resp.read(64 * 1024)
                    if not chunk:
                        break
                    f.write(chunk)
    except Exception as e:
        sys.stderr.write("Download failed. Behind a proxy? Use Node 24+ with NODE_USE_ENV_PROXY=1, or install with install.sh.\n")
        raise e

def safe_extract_tar(tar, path):
    # Use PEP 706 filter="data" if supported (Python 3.12+)
    if hasattr(tarfile, "data_filter"):
        tar.extractall(path, filter="data")
    else:
        for member in tar.getmembers():
            member_path = Path(member.name)
            if member_path.is_absolute() or ".." in member_path.parts:
                raise RuntimeError(f"Archive member contains illegal escaping path: {member.name}")
        tar.extractall(path)

def safe_extract_zip(zf, path):
    for member in zf.namelist():
        member_path = Path(member)
        if member_path.is_absolute() or ".." in member_path.parts:
            raise RuntimeError(f"ZIP member contains illegal escaping path: {member}")
    zf.extractall(path)

def ensure_binary(prog_name):
    os_name, arch_name = get_platform_info()
    is_win = os_name == "windows"
    bin_name = f"{prog_name}.exe" if is_win else prog_name
    cache_dir = get_cache_dir()
    bin_path = cache_dir / bin_name

    if bin_path.exists():
        return str(bin_path)

    # Determine release archive name
    if os_name == "darwin":
        archive = f"vivechak_{VERSION}_darwin_all.tar.gz"
    elif os_name == "windows":
        archive = f"vivechak_{VERSION}_windows_{arch_name}.zip"
    else:
        archive = f"vivechak_{VERSION}_{os_name}_{arch_name}.tar.gz"

    cache_dir.mkdir(parents=True, exist_ok=True)
    random_id = secrets.token_hex(6)
    temp_extract_dir = cache_dir / f".tmp-{os.getpid()}-{random_id}"
    archive_path = cache_dir / f"{archive}.{os.getpid()}-{random_id}.tmp"
    checksum_path = cache_dir / f"checksums.{os.getpid()}-{random_id}.tmp"

    base_url = get_download_base()
    url = f"{base_url}/v{VERSION}/{archive}"
    checksum_url = f"{base_url}/v{VERSION}/checksums.txt"

    sys.stderr.write(f"[vck] Fetching Vivechak v{VERSION} for {os_name}/{arch_name}...\n")
    sys.stderr.flush()

    try:
        download_file(url, archive_path)
        download_file(checksum_url, checksum_path)

        # Compute SHA-256 of downloaded archive
        hasher = hashlib.sha256()
        with open(archive_path, "rb") as f:
            while True:
                chunk = f.read(64 * 1024)
                if not chunk:
                    break
                hasher.update(chunk)
        computed_sha = hasher.hexdigest().lower()

        # Parse checksums.txt
        with open(checksum_path, "rb") as f:
            checksum_data = f.read()

        expected_sha = None
        for line in checksum_data.decode("utf-8", errors="replace").splitlines():
            parts = line.strip().split()
            if len(parts) >= 2 and parts[1] == archive:
                expected_sha = parts[0].lower()
                break

        if not expected_sha:
            raise RuntimeError(f"Checksum entry for {archive} not found in checksums.txt")

        if computed_sha != expected_sha:
            raise RuntimeError(f"Checksum mismatch for {archive}! Expected {expected_sha}, computed {computed_sha}")

        sys.stderr.write(f"[vck] Checksum verified: {computed_sha[:8]}...\n")
        sys.stderr.flush()

        temp_extract_dir.mkdir(parents=True, exist_ok=True)
        if archive.endswith(".zip"):
            with zipfile.ZipFile(archive_path, "r") as zf:
                safe_extract_zip(zf, temp_extract_dir)
        else:
            with tarfile.open(archive_path, "r:gz") as tf:
                safe_extract_tar(tf, temp_extract_dir)

        # Set permissions in temporary extraction directory
        if not is_win:
            for b in [temp_extract_dir / "vivechak", temp_extract_dir / "vck"]:
                if b.exists():
                    b.chmod(0o755)

        # Atomically rename or copy files into cache_dir
        for item in temp_extract_dir.iterdir():
            target_file = cache_dir / item.name
            try:
                item.replace(target_file)
            except Exception:
                shutil.copy2(item, target_file)
                try:
                    item.unlink()
                except Exception:
                    pass

    finally:
        for p in (archive_path, checksum_path):
            if p.exists():
                try:
                    p.unlink()
                except Exception:
                    pass
        if temp_extract_dir.exists():
            shutil.rmtree(temp_extract_dir, ignore_errors=True)

    if not is_win:
        for b in [cache_dir / "vivechak", cache_dir / "vck"]:
            if b.exists():
                b.chmod(0o755)

    if not bin_path.exists():
        alt_name = (f"{'vivechak' if prog_name == 'vck' else 'vck'}.exe") if is_win else ('vivechak' if prog_name == 'vck' else 'vck')
        alt_path = cache_dir / alt_name
        if alt_path.exists():
            shutil.copy2(alt_path, bin_path)
            if not is_win:
                bin_path.chmod(0o755)
        else:
            raise RuntimeError(f"Binary {bin_name} was not found after extraction in {cache_dir}")

    return str(bin_path)

def main():
    try:
        raw_cmd = Path(sys.argv[0]).stem.lower()
        prog = "vck" if "vck" in raw_cmd else "vivechak"
        bin_path = ensure_binary(prog)
        args = [bin_path] + sys.argv[1:]

        env = os.environ.copy()
        env["VIVECHAK_LAUNCHER"] = "uvx"

        if sys.platform not in ("win32", "cygwin"):
            os.execvpe(bin_path, args, env)
        else:
            code = subprocess.call(args, env=env)
            sys.exit(code)
    except KeyboardInterrupt:
        sys.exit(130)
    except Exception as e:
        sys.stderr.write(f"Error: {e}\n")
        sys.exit(1)

if __name__ == "__main__":
    main()
