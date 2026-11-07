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
* **🔄 Safe Atomic Self-Updater:** Update otomatis divalidasi dengan pengecekan ukuran berkas, ELF header check, auto-backup, dan rollback jika gagal.
* **🗑️ Fitur Pencopotan Bersih (Uninstall Manager):** Fleksibel memilih apakah hanya mencopot Envku CLI dengan data aman, atau membersihkan seluruh web stack (*purge*).

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
| **Akses Database & phpMyAdmin** | **User `root` tanpa password** (`AllowNoPassword = true`), langsung masuk di `http://localhost:8080` | Root password diamankan, phpMyAdmin dibatasi firewall & IP whitelist |
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

## 🚀 Perintah CLI & Penggunaan

| Perintah | Deskripsi |
| :--- | :--- |
| `envku` | Membuka antarmuka interaktif TUI utama |
| `envku mode` | Melihat profil environment yang sedang aktif |
| `sudo envku mode local` | Mengalihkan sistem ke mode Local Development |
| `sudo envku mode prod` | Mengalihkan sistem ke mode Production VPS (mengaktifkan firewall ketat) |
| `envku pma status` | Memeriksa status instalasi & URL phpMyAdmin |
| `envku pma open` | Membuka phpMyAdmin langsung di browser desktop lokal |
| `sudo envku update` | Memperbarui binary Envku CLI secara aman (dengan auto-backup & rollback) |
| `sudo envku uninstall` | Membuka wizard pencopotan (Pilihan: Data Aman vs Purge Bersih) |
| `sudo envku uninstall --safe` | Menghapus Envku CLI saja, seluruh data web & database tetap aman |
| `sudo envku daemon install` | Mendaftarkan & mengaktifkan service background Systemd |
| `sudo envku daemon run` | Menjalankan proses monitoring watchdog di foreground |
| `envku version` | Melihat versi aplikasi yang sedang terpasang |

---

## 🗑️ Panduan Uninstall (Pencopotan)

Envku CLI menyediakan dua opsi pencopotan yang aman:

### Opsi 1: Hapus Envku CLI Saja (DATA AMAN - Rekomendasi)
Hanya menghapus binary `envku` dan background service Systemd.
* **File website Anda (`/var/www/`)**: ✅ **AMAN (Tidak disentuh)**
* **Database MySQL (`/var/lib/mysql/`)**: ✅ **AMAN (Tidak disentuh)**
* **Web Server & PHP (`nginx`, `php`, `redis`)**: ✅ **TETAP BERJALAN NORMAL**

Perintah:
```bash
sudo envku uninstall --safe
# atau via skrip mandiri:
sudo bash scripts/uninstall.sh --safe
```

### Opsi 2: Hapus Bersih Seluruhnya (PURGE / COMPLETE WIPE)
Menghapus binary Envku, service daemon, serta seluruh paket dan direktori data (`/var/www`, `/var/lib/mysql`, `/etc/nginx`, `/etc/php`). Mode ini memerlukan konfirmasi pengetikan kata `HAPUS` untuk mencegah penghapusan yang tidak disengaja.

Perintah:
```bash
sudo envku uninstall
# Pilih opsi [2] Purge
# atau via skrip:
sudo bash scripts/uninstall.sh --purge
```

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
│   ├── uninstaller/        # Modul uninstaller (Safe Uninstall & Complete Purge)
│   └── updater/            # Self-update aman dengan validasi ELF, backup & rollback
├── scripts/
│   ├── install.sh          # Skrip bootstrap installer dengan pemilihan mode
│   └── uninstall.sh        # Skrip mandiri uninstaller (Data Aman / Purge)
├── go.mod                  # Definisi Go Module (Zero External Dependency)
└── README.md               # Dokumentasi lengkap proyek
```

---

## 📄 Lisensi

Didistribusikan di bawah lisensi MIT. Silakan gunakan dan kembangkan sesuai kebutuhan server dan infrastruktur Anda.
