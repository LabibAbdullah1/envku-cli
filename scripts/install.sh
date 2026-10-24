#!/usr/bin/env bash
# Envku CLI - One-line Bootstrap Installer for Ubuntu/Debian VPS
set -e

echo "=================================================="
echo "          Envku CLI - Bootstrap Installer         "
echo "=================================================="

# Check Root Privileges
if [ "$EUID" -ne 0 ]; then
  echo "[-] Error: Please run as root (e.g. sudo bash install.sh)"
  exit 1
fi

# Detect Architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64)
    BIN_ARCH="amd64"
    ;;
  aarch64|arm64)
    BIN_ARCH="arm64"
    ;;
  *)
    echo "[-] Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

REPO="LabibAbdullah1/envku-cli"
BIN_NAME="envku-linux-$BIN_ARCH"
TARGET_BIN="/usr/local/bin/envku"

echo "[+] Detected architecture: linux-$BIN_ARCH"

TMP_BIN="/tmp/envku-install.$$"
cleanup() {
  rm -f "$TMP_BIN"
}
trap cleanup EXIT

# If background daemon is currently running, stop it temporarily during update
RESTART_SERVICE=0
if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet envku 2>/dev/null; then
  echo "[*] Stopping running Envku background daemon for binary update..."
  systemctl stop envku 2>/dev/null || true
  RESTART_SERVICE=1
fi

# If local binary exists in build or dist, install that; otherwise download from GitHub Releases
if [ -f "./dist/$BIN_NAME" ]; then
  echo "[+] Installing local binary from ./dist/$BIN_NAME..."
  install -m 755 "./dist/$BIN_NAME" "$TARGET_BIN"
elif [ -f "./$BIN_NAME" ]; then
  echo "[+] Installing local binary from ./$BIN_NAME..."
  install -m 755 "./$BIN_NAME" "$TARGET_BIN"
elif [ -f "./envku" ]; then
  echo "[+] Installing local binary from ./envku..."
  install -m 755 "./envku" "$TARGET_BIN"
else
  echo "[+] Fetching latest release binary from GitHub ($REPO)..."
  DIRECT_URL="https://github.com/$REPO/releases/latest/download/$BIN_NAME"
  if curl -fsSL "$DIRECT_URL" -o "$TMP_BIN" 2>/dev/null; then
    echo "[+] Downloaded $BIN_NAME from $DIRECT_URL"
    install -m 755 "$TMP_BIN" "$TARGET_BIN"
  else
    LATEST_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep "browser_download_url.*$BIN_NAME" | cut -d '"' -f 4)
    if [ -n "$LATEST_URL" ] && curl -fsSL "$LATEST_URL" -o "$TMP_BIN"; then
      echo "[+] Downloaded $BIN_NAME from $LATEST_URL"
      install -m 755 "$TMP_BIN" "$TARGET_BIN"
    else
      echo "[-] Warning: No pre-compiled release found on GitHub. Checking if Go is installed to compile from source..."
      if command -v go >/dev/null 2>&1; then
        echo "[+] Building binary with Go..."
        go build -ldflags="-s -w" -o "$TMP_BIN" ./cmd/envku
        install -m 755 "$TMP_BIN" "$TARGET_BIN"
      else
        echo "[-] Error: Release binary not found on GitHub and Go is not installed."
        exit 1
      fi
    fi
  fi
fi

if [ "$RESTART_SERVICE" -eq 1 ]; then
  echo "[+] Restarting Envku background daemon..."
  systemctl start envku 2>/dev/null || true
fi

chmod +x "$TARGET_BIN"
echo "[+] Binary successfully installed at $TARGET_BIN"

# Register & Enable Background Systemd Daemon
echo "[+] Registering Envku Systemd background service..."
if "$TARGET_BIN" daemon install; then
  echo "[+] Envku daemon installed and running."
else
  echo "[!] Note: Daemon could not be automatically registered (systemd might not be active in containers/WSL1)."
fi

echo "=================================================="
echo "  [SUCCESS] Envku CLI installation complete!      "
echo "  Launch the manager anytime by typing: envku     "
echo "=================================================="
