#!/usr/bin/env bash
# Envku CLI - Bootstrap Installer for Ubuntu/Debian Linux (VPS & Local PC)
set -e

echo "=================================================="
echo "          Envku CLI - Bootstrap Installer         "
echo "=================================================="

# Check Root Privileges
if [ "$EUID" -ne 0 ]; then
  echo "[-] Error: Please run as root (e.g. sudo bash install.sh)"
  exit 1
fi

# Parse CLI flags for unattended installation
CHOSEN_MODE=""
for arg in "$@"; do
  case $arg in
    --prod|--production)
      CHOSEN_MODE="production"
      ;;
    --local|--dev)
      CHOSEN_MODE="local"
      ;;
  esac
done

# If mode is not passed via CLI flag, ask interactively
if [ -z "$CHOSEN_MODE" ]; then
  # Determine if an interactive tty is available (even when piped through curl | sudo bash)
  TTY_SRC=""
  if [ -t 0 ]; then
    TTY_SRC="/dev/stdin"
  elif [ -r /dev/tty ]; then
    TTY_SRC="/dev/tty"
  fi

  if [ -n "$TTY_SRC" ]; then
    echo ""
    echo "Pilih Lingkungan Penggunaan (Target Environment):"
    echo "  [1] 🌐 Production Server (VPS / Cloud - Proteksi firewall ketat, akses via IP/Domain)"
    echo "  [2] 💻 Local Development (Komputer / Laptop Linux Pribadi, WSL - Browser lokal & phpMyAdmin)"
    echo ""
    printf "Pilihan Anda [1-2] (default: 1): "
    read -r USER_CHOICE < "$TTY_SRC" || USER_CHOICE="1"
    USER_CHOICE=$(echo "$USER_CHOICE" | tr -d '[:space:]')
    case "$USER_CHOICE" in
      2|local|dev)
        CHOSEN_MODE="local"
        ;;
      *)
        CHOSEN_MODE="production"
        ;;
    esac
  else
    # Non-interactive automated default
    CHOSEN_MODE="production"
  fi
fi

echo "[+] Target Environment Mode: $(echo "$CHOSEN_MODE" | tr '[:lower:]' '[:upper:]')"

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

chmod +x "$TARGET_BIN"
echo "[+] Binary successfully installed at $TARGET_BIN"

# Set configured environment mode
echo "[*] Configuring environment mode: $CHOSEN_MODE..."
"$TARGET_BIN" mode "$CHOSEN_MODE" || true

# Register & Enable Background Systemd Daemon
echo "[+] Registering Envku Systemd background service..."
if "$TARGET_BIN" daemon install; then
  echo "[+] Envku daemon installed and running."
else
  echo "[!] Note: Daemon could not be automatically registered (systemd might not be active in containers/WSL1)."
fi

if [ "$RESTART_SERVICE" -eq 1 ]; then
  echo "[+] Restarting Envku background daemon..."
  systemctl start envku 2>/dev/null || true
fi

echo "=================================================="
echo "  [SUCCESS] Envku CLI installation complete!      "
if [ "$CHOSEN_MODE" = "local" ]; then
  echo "  Mode: 💻 LOCAL DEVELOPMENT (PC / Desktop / Laptop)"
  echo "  • phpMyAdmin siap dipasang & dibuka di browser lokal"
  echo "  • Virtual Host lokal (.test) otomatis dipetakan ke /etc/hosts"
else
  echo "  Mode: 🌐 PRODUCTION SERVER (VPS / Cloud)"
  echo "  • Proteksi firewall UFW aktif untuk keamanan VPS"
  echo "  • phpMyAdmin dibatasi firewall & IP Whitelist"
  echo "  • Akses wajib melalui IP Publik VPS atau Domain terarah"
fi
echo "--------------------------------------------------"
echo "  Launch the manager anytime by typing: sudo envku"
echo "=================================================="
