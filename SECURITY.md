# Security Policy

Vivechak (विवेचक) is designed as a sovereign, local-first research meta-framework and Model Context Protocol (MCP) server. We take the security, integrity, and privacy of your technical architecture workflows seriously.

---

## 1. Security Architecture & Threat Model

### Workspace Confinement (`os.Root` Jail) & `project_root` Threat Model
- All filesystem read and write operations executed by the MCP server and CLI are confined within the resolved `project_root` workspace directory under `research/` using Go standard library [`os.Root`](https://pkg.go.dev/os#Root) APIs.
- When an agent or user invokes tools, the target `project_root` is evaluated through a strict 7-step resolution ladder; attempts to target filesystem roots (`/`, `C:\`), protected system paths (`/usr`, `/etc`, `C:\Windows`), application folders, or parent directories are rejected.
- Path traversal attacks (e.g. `../../../etc/passwd` or `..\..\Windows\System32`) are blocked at the operating system filesystem handle level.
- Directory creation is confined component-wise (`Workspace.MkdirAll`), eliminating symlink race conditions (TOCTOU).

### Host Configuration Safety (`vck setup`)
- The host configuration command `vck setup` writes and updates MCP server blocks in IDE configuration files.
- Configuration writes are atomic and non-destructive: existing configurations are backed up with immutable timestamps (`<config>.vivechak-bak-<timestamp>`), existing file permissions are preserved (mode `0600` for new files), and identical configurations are not rewritten.

### Zero Telemetry & Zero Network Egress
- **100% Offline & Local-First:** Vivechak contains **zero telemetry**, **zero analytics**, and **zero external network egress**.
- The server does not send code, architecture drafts, prompt outputs, or diagnostic reports over the internet.
- All research documents, ADRs, and DAG states remain 100% private on your local filesystem under git version control.

### Non-Executable Data Handling
- Vivechak treats research artifacts, prompts, and session outputs strictly as structured text (Markdown and YAML frontmatter).
- The server does not evaluate, compile, or execute user-supplied research findings or code blocks.
- Frontmatter parsing is performed via safe deserialization without arbitrary type instantiation.

### Concurrency & Mutation Safety
- Mutating operations (`vivechak_save`, `vivechak_record_decision`, `vivechak_amend`, `vivechak_replan`) acquire cross-process advisory file locks (`flock`) with bounded timeouts.
- File writes employ atomic file generation with monotonic process counters, preventing torn writes and handle collisions on Windows and POSIX filesystems.

### Minimal Dependency Surface
Vivechak maintains an intentionally minimal dependency footprint to eliminate supply-chain vulnerabilities:
- `github.com/modelcontextprotocol/go-sdk` — Official Go SDK for Model Context Protocol.
- `github.com/gofrs/flock` — POSIX and Windows advisory file locking.
- `github.com/tailscale/hujson` — JSONC comment-preserving parser for non-destructive IDE configuration.
- `gopkg.in/yaml.v3` — Standard YAML parsing and serialization.
- Standard Library for all cryptographic, path, AST, and confinement logic.

---

## 2. Supported Versions

| Version | Supported | Security Maintenance |
|---|---|---|
| `v0.1.x` | :white_check_mark: | Active security updates and patches |

---

## 3. Reporting a Vulnerability

If you discover a security vulnerability or sandbox boundary escape in Vivechak, please do **NOT** open a public issue.

Instead, report it directly via private security advisory:
1. Open a **Private Security Advisory** on GitHub: Navigate to the repository's **Security** tab and click **Report a vulnerability**.
2. Or contact the maintainers directly via email: `me@bhaskarjha.dev` (Subject: `[SECURITY] Vivechak Vulnerability Report`).

Please include:
- A description of the vulnerability and attack vector.
- Steps to reproduce or proof-of-concept (PoC).
- Potential impact and affected versions/platforms (Windows, macOS, Linux).

We acknowledge all vulnerability reports within **48 hours** and will provide a coordinated disclosure timeline and fix release.
