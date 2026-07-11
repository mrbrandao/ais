#!/usr/bin/env bash
# ais installer
# Detects OS/arch, downloads the latest release
# binary from GitHub, installs to ~/.local/bin.
#
# Usage:
#   curl -sSfL \
#     https://raw.githubusercontent.com/mrbrandao/ais/main/install.sh \
#     | bash
set -euo pipefail

REPO="mrbrandao/ais"
BINARY="ais"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

# detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)  OS="linux"  ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# detect arch
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported arch: $ARCH" >&2
    exit 1
    ;;
esac

# fetch latest release tag
TAG="$(curl -sSf \
  "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | cut -d'"' -f4)"

FILENAME="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/\
download/${TAG}/${FILENAME}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ais ${TAG} (${OS}/${ARCH})..."
curl -sSfL "$URL" -o "${TMP}/${FILENAME}"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"

mkdir -p "$INSTALL_DIR"
install -m755 "${TMP}/${BINARY}" \
  "${INSTALL_DIR}/${BINARY}"

echo "Installed: ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Add to PATH if needed:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
