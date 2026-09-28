#!/bin/sh
# Vivechak installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh
set -e

REPO="bhaskarjha-dev/vivechak"
INSTALL_DIR="/usr/local/bin"
FALLBACK_DIR="$HOME/.local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)      echo "Error: unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
esac

# macOS universal binary
if [ "$OS" = "darwin" ]; then
  ARCH="all"
fi

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
echo "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
if [ -z "$TAG" ]; then
  echo "Error: could not determine latest release"
  exit 1
fi
VERSION="${TAG#v}"
echo "Latest version: ${TAG}"

# Determine archive name and URL
ARCHIVE="vivechak_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download archive and checksums
echo "Downloading ${ARCHIVE}..."
curl -fsSL -o "${TMP_DIR}/${ARCHIVE}" "${URL}"
curl -fsSL -o "${TMP_DIR}/checksums.txt" "${CHECKSUM_URL}"

# Verify checksum
echo "Verifying checksum..."
cd "$TMP_DIR"
if command -v sha256sum > /dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | sha256sum -c --quiet
elif command -v shasum > /dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | shasum -a 256 -c --quiet
else
  echo "Warning: no sha256sum or shasum found, skipping checksum verification"
fi

# Extract
echo "Extracting..."
tar xzf "${ARCHIVE}"

# Install
if [ -w "$INSTALL_DIR" ]; then
  TARGET="$INSTALL_DIR"
else
  # Try with sudo
  if command -v sudo > /dev/null 2>&1; then
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp vivechak "${INSTALL_DIR}/vivechak"
    sudo chmod +x "${INSTALL_DIR}/vivechak"
    echo ""
    echo "✓ vivechak ${TAG} installed to ${INSTALL_DIR}/vivechak"
    echo ""
    echo "Next step: configure your AI host:"
    echo "  vivechak mcp-config                          # output universal MCP configuration"
    echo "  vivechak mcp-config --path <file> --write    # write directly to any agent config"
    echo "  vivechak mcp-config --preset <name> --write  # or use desktop shortcut (e.g. cursor, vscode)"
    echo ""
    exit 0
  fi
  # Fallback to user directory
  TARGET="$FALLBACK_DIR"
  mkdir -p "$TARGET"
  echo "Note: installing to ${TARGET} (no sudo available)"
  echo "Make sure ${TARGET} is in your PATH"
fi

cp vivechak "${TARGET}/vivechak"
chmod +x "${TARGET}/vivechak"

echo ""
echo "✓ vivechak ${TAG} installed to ${TARGET}/vivechak"
echo ""
echo "Next step: configure your AI host:"
echo "  vivechak mcp-config                          # output universal MCP configuration"
echo "  vivechak mcp-config --path <file> --write    # write directly to any agent config"
echo "  vivechak mcp-config --preset <name> --write  # or use desktop shortcut (e.g. cursor, vscode)"
echo ""
