# Windows Package Manager (Winget) Submission Guide

Vivechak has an authored Winget manifest in [`winget/bhaskarjha-dev.Vivechak.yaml`](../../winget/bhaskarjha-dev.Vivechak.yaml).

## Submission via `wingetcreate` CLI (Recommended)

Once the `v0.1.0` release is published on GitHub, run:

```powershell
# Install wingetcreate if not already installed:
winget install Microsoft.WingetCreate

# Automatically download, compute SHA256 hashes, generate manifests, and submit PR to microsoft/winget-pkgs:
wingetcreate submit https://github.com/bhaskarjha-dev/vivechak/releases/download/v0.1.0/vivechak_0.1.0_windows_amd64.zip
```

When prompted:
- **PackageIdentifier:** `bhaskarjha-dev.Vivechak`
- **Commands:** `vivechak`, `vck`
- **License:** `MIT`
- **Publisher:** `bhaskarjha-dev`
