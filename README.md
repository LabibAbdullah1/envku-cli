# 🚀 Envku CLI (Server & VPS Environment Manager)

**Envku CLI** adalah *lightweight CLI & background environment manager* berbasis terminal yang dirancang khusus untuk Linux dan VPS. Aplikasi ini berfungsi sebagai pengontrol server lokal/remote mandiri yang mengotomatiskan instalasi stack web (*Nginx/Apache, MySQL, PHP Multi-Version, Redis, Node.js*), manajemen ekstensi `php.ini`, serta otomatisasi konfigurasi *Virtual Host (Site Manager)* dan SSL Certbot dalam satu antarmuka terminal interaktif (TUI).

---

## 🛠️ Tech Stack & Arsitektur

* **Language:** Go (Golang) — *Single binary execution, zero-runtime dependency, low memory footprint.*
* **Terminal UI (TUI):** `bubbletea` / `lipgloss` (Charm.sh) atau `cobra` + `promptui`
* **Daemon / Service:** Systemd (`/etc/systemd/system/envku.service`)
* **Target OS:** Linux (Ubuntu/Debian, WSL 2)

---

## 🏗️ Project Architecture & Directory Structure

```text
envku-cli/
├── .github/
│   └── workflows/          # CI/CD automation for cross-compiling binary (Linux amd64/arm64)
├── cmd/
│   └── envku/
│       └── main.go         # Application Entry Point
├── internal/               # Core Application Modules
│   ├── config/             # Path & Global Settings Configuration
│   ├── daemon/             # Background Service Logic & Systemd Integration
│   ├── installer/          # Downloader & Package Installers (Nginx, Apache, MySQL, PHP, Redis, NVM)
│   ├── phpmanager/         # php.ini Parsing, Extension Toggles, & FPM Service Reloader
│   ├── sitemanager/        # Virtual Host / Server Block Automator & Certbot SSL
│   ├── servicectl/         # Start/Stop/Restart Services & Log Monitoring
│   └── tui/                # Interactive Terminal UI (Dashboard, Menus, & Prompts)
├── scripts/
│   └── install.sh          # One-line Bootstrap Installer for VPS
├── go.mod                  # Go Module Definition
├── go.sum                  # Checksum Dependencies
└── README.md               # Project Definition & Documentation
```

---

## 📋 Core Features & Requirements

### 1. One-Line Background Installer (`install.sh`)
* Otomatis mendeteksi distro Linux.
* Mengunduh file binary `envku` yang sudah dikompilasi dari GitHub Release.
* Memindahkan binary ke `/usr/local/bin/envku`.
* Mendaftarkan dan mengaktifkan Systemd Daemon agar berjalan terus di latar belakang.

### 2. Interactive Terminal UI (`envku`)
* **[1] 📊 Dashboard:**
  * Monitoring CPU Usage, RAM Usage, Disk Space, dan System Uptime.
  * Real-Time Status Badges (Nginx/Apache, MySQL, Redis, PHP-FPM).

* **[2] 📦 App & Package Downloader:**
  * **Web Server:** Pilihan antara Nginx atau Apache.
  * **Database:** MySQL / PostgreSQL.
  * **PHP Multi-Version:** PHP 8.2, 8.3, 8.4, 8.5 (via PPA `ondrej/php`).
  * **Cache:** Redis.
  * **Runtime:** Node.js via integrasi NVM.

* **[3] ⚙️ PHP Manager (`php.ini`):**
  * TUI Checkbox interaktif untuk mengaktifkan/menonaktifkan ekstensi PHP esensial (`pdo_mysql`, `mbstring`, `bcmath`, `curl`, `gd`, `zip`, dll.).
  * Tweak parameter penting (`memory_limit`, `upload_max_filesize`, `max_execution_time`).
  * Auto-reload service `php-fpm` setelah konfigurasi disimpan.

* **[4] 🌐 Site Manager (Automated Virtual Host):**
  * **Wizard Pembuatan Site Baru:** Masukkan Subdomain/Domain + Document Root Path + Versi PHP / Node Proxy.
  * Generasi otomatis file konfigurasi (`/etc/nginx/sites-available/` atau Apache `.conf`).
  * Otomatisasi Symlink, validasi syntax (`nginx -t`), dan reload web server.
  * Opsi 1-klik pemasangan SSL HTTPS via Certbot (Let's Encrypt).

* **[5] ⚡ Service Manager:**
  * Kontrol `start`, `stop`, `restart`, dan pembacaan log untuk semua servis infrastruktur.

---

## 🚀 Development Workflow

### Local Build & Test

```bash
# Inisialisasi Modul (Langkah Awal Repo Baru)
go mod init github.com/LabibAbdullah1/envku-cli

# Build Binary Lokal
go build -o envku cmd/envku/main.go

# Run Project
./envku
```

---

## 🎯 Rules & Design Principles for AI & Contributors

* **Keep It Lightweight:** Jangan menambahkan dependensi berat yang bisa membengkakkan ukuran binary atau konsumsi RAM di VPS.
* **Safe Manipulation:** Setiap kali program mengubah file konfigurasi sistem (seperti `php.ini`, `nginx.conf`, atau `/etc/systemd/`), selalu lakukan pengecekan syntax validation sebelum melakukan service reload.
* **Clean Code Architecture:** Selalu pisahkan logika System Execution di paket `internal/` dan UI Presentation di paket `internal/tui`.
