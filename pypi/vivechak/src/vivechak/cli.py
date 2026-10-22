"""Native binary runner for Vivechak and vck.
All diagnostic and status output is directed strictly to sys.stderr so that
sys.stdout remains 100% clean for MCP JSON-RPC protocol communication.
"""

import os
import sys
import platform
import urllib.request
import tarfile
import zipfile
import subprocess
from pathlib import Path

VERSION = "0.1.0"
REPO = "bhaskarjha-dev/vivechak"

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

def ensure_binary(prog_name):
    os_name, arch_name = get_platform_info()
    is_win = os_name == "windows"
    bin_name = f"{prog_name}.exe" if is_win else prog_name
    cache_dir = get_cache_dir()
    bin_path = cache_dir / bin_name

    if bin_path.exists():
        return str(bin_path)

    # Check local repository development build
    repo_root = Path(__file__).resolve().parent.parent.parent.parent.parent
    local_main_go = repo_root / "cmd" / "vivechak" / "main.go"
    if local_main_go.exists():
        try:
            cache_dir.mkdir(parents=True, exist_ok=True)
            sys.stderr.write(f"[vck] Building local development binary in {cache_dir}...\n")
            sys.stderr.flush()
            build_res = subprocess.call(["go", "build", "-trimpath", "-o", str(bin_path), "./cmd/vivechak"], cwd=str(repo_root), stdout=sys.stderr)
            if build_res == 0 and bin_path.exists():
                if not is_win:
                    bin_path.chmod(0o755)
                return str(bin_path)
        except (OSError, FileNotFoundError):
            # Go compiler not available on this host; fall through to downloading binary
            pass

    # Determine release archive name
    if os_name == "darwin":
        archive = f"vivechak_{VERSION}_darwin_all.tar.gz"
    elif os_name == "windows":
        archive = f"vivechak_{VERSION}_windows_{arch_name}.zip"
    else:
        archive = f"vivechak_{VERSION}_{os_name}_{arch_name}.tar.gz"

    cache_dir.mkdir(parents=True, exist_ok=True)
    archive_path = cache_dir / archive
    url = f"https://github.com/{REPO}/releases/download/v{VERSION}/{archive}"

    sys.stderr.write(f"[vck] Fetching Vivechak v{VERSION} for {os_name}/{arch_name}...\n")
    sys.stderr.flush()

    try:
        req = urllib.request.Request(url, headers={"User-Agent": f"vivechak-runner/{VERSION}"})
        with urllib.request.urlopen(req, timeout=30.0) as resp:
            if resp.status != 200:
                raise RuntimeError(f"Download failed with HTTP status {resp.status} from {url}")
            with open(archive_path, "wb") as f:
                while True:
                    chunk = resp.read(64 * 1024)
                    if not chunk:
                        break
                    f.write(chunk)
        if archive.endswith(".zip"):
            with zipfile.ZipFile(archive_path, "r") as zf:
                zf.extractall(cache_dir)
        else:
            with tarfile.open(archive_path, "r:gz") as tf:
                tf.extractall(cache_dir)
    finally:
        if archive_path.exists():
            try:
                archive_path.unlink()
            except Exception:
                pass

    if not is_win:
        for b in [cache_dir / "vivechak", cache_dir / "vck"]:
            if b.exists():
                b.chmod(0o755)

    if not bin_path.exists():
        alt_name = (f"{'vivechak' if prog_name == 'vck' else 'vck'}.exe") if is_win else ('vivechak' if prog_name == 'vck' else 'vck')
        alt_path = cache_dir / alt_name
        if alt_path.exists():
            import shutil
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

        if sys.platform not in ("win32", "cygwin"):
            os.execv(bin_path, args)
        else:
            code = subprocess.call(args)
            sys.exit(code)
    except KeyboardInterrupt:
        sys.exit(130)
    except Exception as e:
        sys.stderr.write(f"Error: {e}\n")
        sys.exit(1)

if __name__ == "__main__":
    main()
