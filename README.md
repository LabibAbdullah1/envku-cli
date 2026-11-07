# 🚀 Envku CLI (Linux & VPS Server Environment Manager)

**Envku CLI** adalah *lightweight, standalone server environment manager* berbasis terminal (TUI) yang dibangun secara native menggunakan **Go (Golang 1.22+)** tanpa dependensi eksternal (*Zero Third-Party Dependencies*). 

Dirancang fleksibel untuk berjalan di dua jenis lingkungan: **Server VPS Publik (Production)** dan **Komputer Linux Pribadi / Laptop / WSL (Local Development)**. Envku mengotomatiskan instalasi web stack (*Nginx, MySQL, PHP Multi-Version, Redis, Node.js, phpMyAdmin*), manajemen firewall, konfigurasi `php.ini`, otomatisasi Virtual Host (*Site Manager*), serta penerbitan sertifikat SSL Let's Encrypt (Certbot) dalam satu antarmuka terminal yang ramah pengguna.

---

## 🌟 Keunggulan Utama

* **⚡ Pure Native Go (Zero Dependencies):** Dibuat 100% menggunakan Standard Library Go. Tanpa framework berat, ukuran binary sangat ramping (~10 MB), dan konsumsi RAM sangat hemat.
* **🌐 Dua Pilihan Environment (Production vs Local):** Menyesuaikan profil keamanan, akses firewall, serta perilaku browser secara otomatis sesuai kebutuhan server.
* **🗄️ Dukungan phpMyAdmin Cerdas:**
  * **Di Local:** Sekali klik langsung pasang dan otomatis buka di browser desktop Linux Anda (`http://localhost:8080`).
  * **Di Production:** Dibatasi oleh firewall UFW dan Nginx IP Whitelist untuk melindungi database dari ancaman brute-force publik.
* **🖥️ Local Domain Resolver:** Otomatis menambahkan domain pengujian (seperti `proyek.test`) ke `/etc/hosts` di Linux pribadi.
* **🛡️ Safe & Atomic Operations:** Setiap perubahan konfigurasi dilengkapi dengan *auto-backup*, validasi sintaks otomatis (`nginx -t` / `php-fpm -t`), dan *auto-rollback*.
* **📦 Non-Interactive & Automated:** Semua perintah instalasi dieksekusi dengan flag non-interaktif (`DEBIAN_FRONTEND=noninteractive`).
* **🔄 Background Daemon Integration:** Terintegrasi dengan Systemd Linux (`envku.service`) sebagai watchdog pemantau kesehatan server.

---

## ⚖️ Perbandingan & Rekomendasi Fitur: Local vs Production

Envku CLI membedakan fungsionalitas antara komputer Linux pribadi dan server VPS publik:

| Fitur / Komponen | 💻 Local Development (Linux Pribadi / Laptop / WSL) | 🌐 Production (Server VPS / Cloud) |
| :--- | :--- | :--- |
| **Tujuan Utama** | Kemudahan ngoding cepat, testing lokal, hemat daya | Keamanan tinggi, uptime 24/7, isolasi akses publik |
| **Akses phpMyAdmin** | **Langsung di Browser Lokal** (`xdg-open http://localhost:8080`) | **Dibatasi Firewall & IP Whitelist** atau via SSH Tunneling |
| **Cara Akses Web** | `http://localhost`, `127.0.0.1`, atau domain `.test` | **Public IP VPS** atau **Domain Publik (FQDN)** dengan DNS A-Record |
| **Domain Virtual Host** | Mendukung `.test` / `.local` + **Auto `/etc/hosts`** | Wajib domain resmi (misal `app.com`) |
| **Sertifikat SSL** | Ditiadakan / Self-signed untuk domain lokal | **1-Klik Let's Encrypt SSL (Certbot)** otomatis dengan auto-renewal |
| **Profil Firewall (UFW)**| **Permissive** (Port 80, 443, 3306, 6379, 8080 terbuka di loopback) | **Strict** (Port 3306, 6379, 8080 diblokir dari publik; hanya 22, 80, 443) |
| **PHP Error Reporting** | `display_errors = On` & `error_reporting = E_ALL` | `display_errors = Off`, `expose_php = Off`, Opcache aktif |
| **Akses Database CLI** | Root password fleksibel untuk tool desktop (TablePlus, DBeaver) | Root password diamankan, dianjurkan user database khusus per aplikasi |
| **Deteksi Browser** | Otomatis deteksi `xdg-open`, `wslview`, Chrome, Firefox | Non-aktif (Headless Server) |

---

## 🖥️ Kebutuhan Sistem (System Requirements)

* **Sistem Operasi:**
  * Ubuntu 20.04 / 22.04 / 24.04 LTS
  * Debian 11 (Bullseye) / 12 (Bookworm)
  * WSL 2 (Windows Subsystem for Linux)
* **Arsitektur:** `amd64` (x86_64) atau `arm64` (aarch64)
* **Hak Akses:** Root atau pengguna dengan akses `sudo`

---

## 📥 Panduan Instalasi

### Metode 1: One-Line Installer Interaktif (Rekomendasi) ⭐

Jalankan perintah berikut di terminal Linux atau server VPS Anda:

```bash
curl -fsSL https://raw.githubusercontent.com/LabibAbdullah1/envku-cli/main/scripts/install.sh | sudo bash
```

Saat instalasi berjalan, skrip akan menanyakan pilihan lingkungan:
```text
Pilih Lingkungan Penggunaan (Target Environment):
  [1] 🌐 Production Server (VPS / Cloud - Proteksi firewall ketat, akses via IP/Domain)
  [2] 💻 Local Development (Komputer / Laptop Linux Pribadi, WSL - Browser lokal & phpMyAdmin)
```

Anda juga dapat menjalankan instalasi otomatis tanpa dialog menggunakan flag:
* **Untuk VPS Production:** `curl -fsSL ... | sudo bash -s -- --production`
* **Untuk Linux Pribadi:** `curl -fsSL ... | sudo bash -s -- --local`

---

### Metode 2: Kompilasi Manual dari Source

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

# 4. Tentukan mode environment (pilih salah satu)
sudo envku mode local        # Untuk PC / Laptop Linux
# atau
sudo envku mode production   # Untuk VPS Server
```

---

## 🚀 Perintah CLI & Penggunaan

### 1. Membuka Dashboard TUI Utama
```bash
sudo envku
```

### 2. Perintah Cepat CLI
| Perintah | Deskripsi |
| :--- | :--- |
| `envku` | Membuka antarmuka interaktif TUI utama |
| `envku mode` | Melihat profil environment yang sedang aktif |
| `sudo envku mode local` | Mengalihkan sistem ke mode Local Development |
| `sudo envku mode prod` | Mengalihkan sistem ke mode Production VPS (mengaktifkan firewall ketat) |
| `envku pma status` | Memeriksa status instalasi & URL phpMyAdmin |
| `envku pma open` | Membuka phpMyAdmin langsung di browser desktop lokal |
| `sudo envku update` | Memperbarui binary Envku CLI langsung dari GitHub Release |
| `sudo envku daemon install` | Mendaftarkan & mengaktifkan service background Systemd |
| `sudo envku daemon run` | Menjalankan proses monitoring watchdog di foreground |
| `envku version` | Melihat versi aplikasi yang sedang terpasang |

---

## 🗄️ Panduan Akses phpMyAdmin

### 1. Pada Komputer Linux Pribadi (Local Mode)
* Pasang phpMyAdmin melalui menu TUI `[6] phpMyAdmin & Database Access` atau `[2] App & Package Downloader`.
* phpMyAdmin akan otomatis dikonfigurasi pada port `8080`.
* Pilih menu `[1] Buka phpMyAdmin di Browser` atau jalankan `envku pma open` di terminal. Sistem akan memanggil browser default Anda (`http://localhost:8080`).

### 2. Pada Server VPS (Production Mode)
Di server VPS publik, phpMyAdmin **tidak dibuka bebas ke publik** demi keamanan data. Anda memiliki dua opsi akses aman:
1. **Opsi 1: SSH Tunnel (Paling Direkomendasikan)**
   Jalankan dari komputer/laptop Anda:
   ```bash
   ssh -N -L 8080:127.0.0.1:8080 user@IP_VPS_ANDA
   ```
   Lalu buka browser Anda di `http://localhost:8080`.
2. **Opsi 2: IP Whitelist Admin**
   Buka menu `[6] -> [2] Konfigurasi IP Whitelist Admin` di Envku TUI, dan masukkan IP publik koneksi internet Anda. Firewall UFW dan Nginx akan mengizinkan akses hanya dari IP tersebut via `http://IP_VPS:8080`.

---

## 🏗️ Struktur Proyek

```text
envku-cli/
├── cmd/
│   └── envku/
│       └── main.go         # Titik masuk utama aplikasi (CLI Entry Point)
├── internal/
│   ├── config/             # Manajemen profil config (Local vs Prod) & browser launcher
│   ├── daemon/             # Integrasi service background Systemd & watchdog
│   ├── firewall/           # Pengendali UFW Firewall & profil keamanan
│   ├── installer/          # Modul instalasi (Nginx, MySQL, PHP, Redis, Node, phpMyAdmin)
│   ├── phpmanager/         # Parser & editor php.ini, toggle ekstensi PHP
│   ├── sitemanager/        # Generator Virtual Host, /etc/hosts helper & Certbot SSL
│   ├── servicectl/         # Pengendali service systemd & pembaca log
│   ├── tui/                # Antarmuka Terminal UI interaktif (ANSI murni)
│   └── updater/            # Self-update updater otomatis dari GitHub Release
├── scripts/
│   └── install.sh          # Skrip bootstrap installer dengan pemilihan mode
├── go.mod                  # Definisi Go Module (Zero External Dependency)
└── README.md               # Dokumentasi lengkap proyek
```

---

## 📄 Lisensi

Didistribusikan di bawah lisensi MIT. Silakan gunakan dan kembangkan sesuai kebutuhan server dan infrastruktur Anda.
