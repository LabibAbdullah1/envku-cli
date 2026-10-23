# 🚀 Envku CLI (Linux & VPS Server Environment Manager)

**Envku CLI** adalah *lightweight, standalone server environment manager* berbasis terminal (TUI) yang dibangun secara native menggunakan **Go (Golang 1.22+)** tanpa dependensi eksternal (*Zero Third-Party Dependencies*). 

Dirancang khusus untuk mengelola VPS dan server Linux (Ubuntu/Debian) secara efisien, Envku mengotomatiskan instalasi web stack (*Nginx, MySQL, PHP Multi-Version, Redis, Node.js*), manajemen `php.ini`, otomatisasi Virtual Host (*Site Manager*), serta penerbitan sertifikat SSL Let's Encrypt (Certbot) dalam satu antarmuka terminal yang ramah pengguna.

---

## 🌟 Keunggulan Utama

* **⚡ Pure Native Go (Zero Dependencies):** Dibuat 100% menggunakan Standard Library Go. Tanpa framework berat, ukuran binary sangat ramping (~10 MB), dan konsumsi RAM sangat hemat.
* **🛡️ Safe & Atomic Operations:** Setiap perubahan konfigurasi (seperti `php.ini` atau Nginx Server Block) dilengkapi dengan *auto-backup*, validasi sintaks otomatis (`nginx -t` / `php-fpm -t`), dan *auto-rollback* jika terjadi kesalahan konfigurasi.
* **📦 Non-Interactive & Automated:** Semua perintah instalasi dieksekusi dengan flag *non-interactive* (`DEBIAN_FRONTEND=noninteractive`), mencegah proses instalasi berhenti di tengah jalan karena menunggu input dialog OS.
* **🔄 Background Daemon Integration:** Terintegrasi langsung dengan Systemd Linux (`envku.service`) untuk menjalankan monitoring di latar belakang.

---

## 🖥️ Kebutuhan Sistem (System Requirements)

* **Sistem Operasi:**
  * Ubuntu 20.04 / 22.04 / 24.04 LTS
  * Debian 11 (Bullseye) / 12 (Bookworm)
  * WSL 2 (Windows Subsystem for Linux) untuk pengujian lokal
* **Arsitektur:** `amd64` (x86_64) atau `arm64` (aarch64)
* **Hak Akses:** Root atau pengguna dengan akses `sudo`

---

## 📥 Panduan Instalasi di Server Linux (VPS)

Pilih salah satu metode instalasi yang paling sesuai dengan kebutuhan Anda:

### Metode 1: One-Line Installer (Paling Praktis) ⭐

Jalankan perintah berikut langsung di terminal server VPS Anda:

```bash
curl -fsSL https://raw.githubusercontent.com/LabibAbdullah1/envku-cli/main/scripts/install.sh | sudo bash
```

**Apa yang dilakukan skrip ini?**
1. Mendeteksi arsitektur server Anda (`amd64` atau `arm64`).
2. Mengunduh binary rilis terbaru `envku` dari GitHub.
3. Memasang binary ke direktori sistem: `/usr/local/bin/envku`.
4. Mendaftarkan dan mengaktifkan service background Systemd (`envku.service`).

---

### Metode 2: Upload Binary dari Komputer Lokal (Windows / macOS)

Jika Anda sudah mengompilasi binary di komputer lokal:

1. **Kompilasi binary untuk Linux di komputer lokal:**
   * Di Git Bash / Linux / macOS:
     ```bash
     GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o envku ./cmd/envku
     ```
   * Di Windows PowerShell:
     ```powershell
     $env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o envku ./cmd/envku
     ```

2. **Upload binary ke server VPS menggunakan SCP:**
   ```bash
   scp envku user@IP_SERVER_ANDA:/tmp/envku
   ```

3. **Login ke server VPS dan pindahkan binary:**
   ```bash
   ssh user@IP_SERVER_ANDA
   sudo mv /tmp/envku /usr/local/bin/envku
   sudo chmod +x /usr/local/bin/envku
   sudo envku daemon install
   ```

---

### Metode 3: Kompilasi dari Source di Server

Jika server VPS Anda sudah terpasang Golang (v1.22+):

```bash
# 1. Clone repository
git clone https://github.com/LabibAbdullah1/envku-cli.git
cd envku-cli

# 2. Build binary
go build -ldflags="-s -w" -o envku ./cmd/envku

# 3. Pindahkan ke direktori bin & daftarkan daemon
sudo mv envku /usr/local/bin/
sudo chmod +x /usr/local/bin/envku
sudo envku daemon install
```

---

## 🚀 Cara Menjalankan & Penggunaan

### 1. Membuka Terminal UI (TUI)
Untuk membuka dashboard dan menu interaktif, jalankan perintah berikut dengan hak akses `sudo`:

```bash
sudo envku
```

### 2. Perintah CLI Tambahan

| Perintah | Deskripsi |
| :--- | :--- |
| `envku` | Membuka antarmuka interaktif TUI utama |
| `sudo envku daemon install` | Mendaftarkan dan mengaktifkan service background Systemd |
| `sudo envku daemon run` | Menjalankan proses daemon secara langsung di foreground |
| `sudo envku daemon uninstall` | Menghapus dan menonaktifkan service Systemd `envku` |
| `envku version` | Melihat versi aplikasi yang sedang terpasang |
| `envku help` | Menampilkan panduan bantuan perintah CLI |

---

## 📋 Fitur-Fitur Menu Interaktif

```text
==================================================
                   MAIN MENU                      
==================================================
[1] 📊 Refresh Dashboard
[2] 📦 App & Package Downloader (Nginx, MySQL, Redis, PHP, Node)
[3] ⚙️  PHP Manager (php.ini Tweaks & Extensions)
[4] 🌐 Site Manager (Virtual Host & Certbot SSL)
[5] ⚡ Service Manager (systemctl & Log Inspector)
--------------------------------------------------
[0] 🚪 Exit
==================================================
```

### 1. 📊 Dashboard
* Menampilkan informasi CPU Model, penggunaan RAM, persentase disk, dan Uptime server.
* Memonitor status kesehatan service aktif (Nginx, MySQL, Redis, PHP-FPM).

### 2. 📦 App & Package Downloader
* **Nginx:** Web server & reverse proxy berkinerja tinggi.
* **MySQL:** Database server relational.
* **Redis:** In-memory cache server dari repository resmi Redis (`packages.redis.io`) versi terbaru.
* **PHP Multi-Version:** Instalasi versi PHP 8.2, 8.3, 8.4, dan 8.5 via PPA Ondrej lengkap dengan ekstensi esensial (`fpm`, `mysql`, `mbstring`, `curl`, `gd`, `zip`, dll.).
* **Node.js & NVM:** Pemasangan otomatis NVM dan Node.js versi LTS.

### 3. ⚙️ PHP Manager
* **Tweaks `php.ini`:** Mengubah konfigurasi penting seperti `memory_limit`, `upload_max_filesize`, `post_max_size`, `max_execution_time`, `max_input_vars`, dan `date.timezone`.
* **Atomic Backup & Auto-Rollback:** Membuat salinan cadangan sebelum mengubah file konfigurasi. Jika validasi sintaks `php-fpm -t` gagal, sistem akan otomatis melakukan *rollback* ke konfigurasi semula.
* **Extension Manager:** Mengaktifkan atau menonaktifkan ekstensi PHP secara dinamis per versi menggunakan `phpenmod` dan `phpdismod`.

### 4. 🌐 Site Manager (Virtual Host & SSL)
* **Server Block Generator:** Wizard pembuatan virtual host Nginx untuk aplikasi berbasis **PHP-FPM** maupun **Node.js Reverse Proxy**.
* **Otomatisasi Penuh:** Mengatur direktori root, membuat konfigurasi di `/etc/nginx/sites-available/`, membuat symlink ke `sites-enabled/`, validasi sintaks `nginx -t`, dan melakukan reload service.
* **1-Klik SSL Let's Encrypt:** Penerbitan dan pembaruan sertifikat HTTPS gratis secara otomatis melalui integrasi Certbot.

### 5. ⚡ Service Manager & Log Inspector
* Kontrol cepat untuk `start`, `stop`, dan `restart` service web stack.
* Fitur *Log Inspector* untuk memantau log error/akses (misalnya `/var/log/nginx/error.log`) langsung dari terminal tanpa harus keluar dari aplikasi.

---

## 🏗️ Struktur Proyek

```text
envku-cli/
├── cmd/
│   └── envku/
│       └── main.go         # Titik masuk utama aplikasi (CLI Entry Point)
├── internal/
│   ├── config/             # Konstanta versi & path direktori sistem
│   ├── daemon/             # Integrasi service background Systemd
│   ├── installer/          # Modul instalasi software (Nginx, MySQL, PHP, Redis, Node)
│   ├── phpmanager/         # Parser & editor php.ini, toggle ekstensi PHP
│   ├── sitemanager/        # Generator Virtual Host Nginx & Certbot SSL
│   ├── servicectl/         # Pengendali service systemd & pembaca log
│   └── tui/                # Antarmuka Terminal UI murni (ANSI / Standard Library)
├── scripts/
│   └── install.sh          # Skrip bootstrap installer satu baris untuk Linux
├── go.mod                  # Definisi Go Module (Zero External Dependency)
└── README.md               # Dokumentasi lengkap proyek
```

---

## 📄 Lisensi

Didistribusikan di bawah lisensi MIT. Silakan gunakan dan kembangkan sesuai kebutuhan server dan infrastruktur Anda.
