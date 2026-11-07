#!/usr/bin/env bash
# Envku CLI - Uninstaller Script for Ubuntu/Debian Linux
set -e

echo "=================================================="
echo "        Envku CLI - Uninstaller Script            "
echo "=================================================="

# Check Root Privileges
if [ "$EUID" -ne 0 ]; then
  echo "[-] Error: Harap jalankan dengan hak akses root (e.g. sudo bash uninstall.sh)"
  exit 1
fi

UNINSTALL_MODE=""
for arg in "$@"; do
  case $arg in
    --safe)
      UNINSTALL_MODE="safe"
      ;;
    --purge)
      UNINSTALL_MODE="purge"
      ;;
  esac
done

if [ -z "$UNINSTALL_MODE" ]; then
  TTY_SRC=""
  if [ -t 0 ]; then
    TTY_SRC="/dev/stdin"
  elif [ -r /dev/tty ]; then
    TTY_SRC="/dev/tty"
  fi

  if [ -n "$TTY_SRC" ]; then
    echo ""
    echo "PILIH METODE UNINSTALL:"
    echo "  [1] 🛡️  HAPUS ENVKU CLI SAJA (DATA AMAN - Rekomendasi)"
    echo "      • Menghapus binary /usr/local/bin/envku-cli, alias envku, dan daemon Systemd"
    echo "      • Database MySQL, file website (/var/www), Nginx, PHP, dan SSL"
    echo "        TETAP AMAN 100% dan web server tetap berjalan normal."
    echo ""
    echo "  [2] 💣 HAPUS BERSIH SELURUHNYA (PURGE / COMPLETE WIPE)"
    echo "      • Menghapus Envku CLI, web server Nginx, MySQL Database, PHP, Redis,"
    echo "        phpMyAdmin, file website (/var/www), dan database (/var/lib/mysql)."
    echo "      • ⚠️  PERINGATAN: Seluruh website & database akan TERHAPUS PERMANEN!"
    echo ""
    printf "Pilihan Anda [1-2] (default: 1): "
    read -r USER_CHOICE < "$TTY_SRC" || USER_CHOICE="1"
    USER_CHOICE=$(echo "$USER_CHOICE" | tr -d '[:space:]')
    case "$USER_CHOICE" in
      2|purge)
        UNINSTALL_MODE="purge"
        ;;
      *)
        UNINSTALL_MODE="safe"
        ;;
    esac
  else
    UNINSTALL_MODE="safe"
  fi
fi

if [ "$UNINSTALL_MODE" = "safe" ]; then
  echo ""
  echo "[*] Menjalankan Uninstall Mode Aman (Data & Konfigurasi Aman)..."

  # Stop & remove daemon
  if command -v systemctl >/dev/null 2>&1; then
    echo "[*] Menghentikan & menonaktifkan service Systemd envku-cli..."
    systemctl stop envku-cli 2>/dev/null || true
    systemctl disable envku-cli 2>/dev/null || true
    systemctl stop envku 2>/dev/null || true
    systemctl disable envku 2>/dev/null || true
    rm -f /etc/systemd/system/envku-cli.service /etc/systemd/system/envku.service
    systemctl daemon-reload 2>/dev/null || true
  fi

  # Remove internal config
  rm -rf /etc/envku-cli

  # Remove binary & symlink
  echo "[*] Menghapus binary /usr/local/bin/envku-cli & symlink envku..."
  rm -f /usr/local/bin/envku-cli /usr/local/bin/envku

  echo ""
  echo "=================================================="
  echo "  [SUKSES] Envku CLI telah berhasil dicopot!     "
  echo "  ✅ Seluruh data website (/var/www), database,   "
  echo "     konfigurasi Nginx, dan PHP tetap UTUH & AMAN."
  echo "=================================================="
  exit 0
fi

if [ "$UNINSTALL_MODE" = "purge" ]; then
  echo ""
  echo "⚠️  PERINGATAN TINGKAT TINGGI: ANDA MEMILIH PURGE!"
  echo "Seluruh file website (/var/www), database MySQL, Nginx, dan PHP"
  echo "akan DIHAPUS PERMANEN dari server Anda!"
  echo ""
  TTY_SRC=""
  if [ -t 0 ]; then
    TTY_SRC="/dev/stdin"
  elif [ -r /dev/tty ]; then
    TTY_SRC="/dev/tty"
  fi

  if [ -n "$TTY_SRC" ]; then
    printf "Ketik 'HAPUS' (huruf besar) untuk melanjutkan: "
    read -r CONFIRM_TEXT < "$TTY_SRC" || CONFIRM_TEXT=""
    CONFIRM_TEXT=$(echo "$CONFIRM_TEXT" | tr -d '[:space:]')
    if [ "$CONFIRM_TEXT" != "HAPUS" ] && [ "$CONFIRM_TEXT" != "PURGE" ]; then
      echo "[-] Konfirmasi tidak sesuai. Penghapusan dibatalkan."
      exit 1
    fi
  fi

  echo ""
  echo "[*] Menghentikan seluruh service web stack..."
  systemctl stop envku-cli envku nginx mysql redis-server php*-fpm 2>/dev/null || true
  systemctl disable envku-cli envku nginx mysql redis-server php*-fpm 2>/dev/null || true

  echo "[*] Menghapus binary Envku CLI & service systemd..."
  rm -f /etc/systemd/system/envku-cli.service /etc/systemd/system/envku.service
  rm -f /usr/local/bin/envku-cli /usr/local/bin/envku
  rm -rf /etc/envku-cli
  systemctl daemon-reload 2>/dev/null || true

  echo "[*] Menghapus instalasi phpMyAdmin..."
  rm -f /etc/nginx/sites-enabled/phpmyadmin.conf
  rm -f /etc/nginx/sites-available/phpmyadmin.conf
  rm -rf /usr/share/phpmyadmin

  echo "[*] Melakukan apt-get purge pada software web stack..."
  export DEBIAN_FRONTEND=noninteractive
  apt-get purge -y --auto-remove nginx* mysql* php* redis* certbot* 2>/dev/null || true

  echo "[*] Menghapus direktori data web (/var/www)..."
  rm -rf /var/www

  echo "[*] Menghapus direktori database MySQL (/var/lib/mysql, /etc/mysql)..."
  rm -rf /var/lib/mysql /etc/mysql

  echo "[*] Menghapus direktori konfigurasi Nginx dan PHP..."
  rm -rf /etc/nginx /etc/php

  echo ""
  echo "=================================================="
  echo "  [SELESAI] Server telah dibersihkan sepenuhnya.  "
  echo "=================================================="
  exit 0
fi
