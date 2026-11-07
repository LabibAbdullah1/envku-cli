package tui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/firewall"
)

func runEnvironmentMenu(reader *bufio.Reader) {
	fw := firewall.New()

	for {
		clearScreen()
		cfg := config.GetConfig()
		currentMode := cfg.Environment

		fmt.Println("==================================================")
		fmt.Println("       🔀 ENVIRONMENT & OPERATING PROFILE        ")
		fmt.Println("==================================================")
		if currentMode == config.ModeLocal {
			fmt.Println("  Mode Aktif: 💻 LOCAL DEVELOPMENT (PC / Desktop / Laptop)")
		} else {
			fmt.Println("  Mode Aktif: 🌐 PRODUCTION SERVER (VPS / Cloud Public)")
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("[1] 💻 Alihkan ke Mode LOCAL (Desktop / Laptop / WSL)")
		fmt.Println("    • Ramah pengembang: bisa langsung buka phpMyAdmin di browser")
		fmt.Println("    • Integrasi otomatis domain lokal (.test) ke /etc/hosts")
		fmt.Println("    • Firewall fleksibel (tidak memblokir port development)")
		fmt.Println("")
		fmt.Println("[2] 🌐 Alihkan ke Mode PRODUCTION (VPS / Cloud)")
		fmt.Println("    • Proteksi ketat: phpMyAdmin dibatasi firewall & IP whitelist")
		fmt.Println("    • Akses via Public IP atau Domain yang terarah ke VPS")
		fmt.Println("    • Integrasi Certbot SSL Let's Encrypt resmi")
		fmt.Println("    • Database MySQL & Redis tertutup dari publik")
		fmt.Println("")
		fmt.Println("[3] 📋 Lihat Rekomendasi Fitur Lengkap (Prod vs Local)")
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Kembali ke Menu Utama")
		fmt.Println("==================================================")
		fmt.Print("Pilih opsi [0-3]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := config.SetEnvironment(config.ModeLocal); err != nil {
				fmt.Printf("[-] Gagal mengubah mode: %v\n", err)
			} else {
				fmt.Println("[+] Berhasil beralih ke Mode LOCAL DEVELOPMENT.")
				_ = fw.ApplyLocalProfile(cfg.PMAPort)
			}
			waitEnter(reader)

		case "2":
			if err := config.SetEnvironment(config.ModeProduction); err != nil {
				fmt.Printf("[-] Gagal mengubah mode: %v\n", err)
			} else {
				fmt.Println("[+] Berhasil beralih ke Mode PRODUCTION SERVER.")
				fmt.Println("[*] Mengaktifkan profil firewall ketat UFW...")
				_ = fw.ApplyProductionProfile(cfg.PMAPort, cfg.AllowedAdminIPs)
			}
			waitEnter(reader)

		case "3":
			showFeatureRecommendations(reader)

		case "0":
			return
		}
	}
}

func showFeatureRecommendations(reader *bufio.Reader) {
	clearScreen()
	fmt.Println("================================================================================")
	fmt.Println("        📋 REKOMENDASI FITUR & PERBEDAAN (LOCAL VS PRODUCTION)                 ")
	fmt.Println("================================================================================")
	fmt.Println("")
	fmt.Println("1. 🌐 AKSES BROWSER & phpMyAdmin:")
	fmt.Println("   • Local : Mendeteksi browser GUI (xdg-open, wslview, chrome, firefox).")
	fmt.Println("             Satu klik langsung membuka http://localhost:8080 di browser.")
	fmt.Println("   • Prod  : phpMyAdmin dibatasi oleh Firewall UFW dan Nginx IP Whitelist.")
	fmt.Println("             Akses wajib lewat Public IP/Domain VPS yang di-whitelist,")
	fmt.Println("             atau via SSH Tunnel aman: ssh -N -L 8080:127.0.0.1:8080 user@vps.")
	fmt.Println("")
	fmt.Println("2. 🖥️ DOMAIN & VIRTUAL HOST:")
	fmt.Println("   • Local : Mendukung domain development lokal (.test, .local).")
	fmt.Println("             Otomatis mendaftarkan domain ke /etc/hosts (127.0.0.1).")
	fmt.Println("   • Prod  : Wajib menggunakan domain publik (FQDN) dengan A-Record DNS")
	fmt.Println("             yang mengarah ke IP publik VPS. Dilengkapi Certbot Let's Encrypt.")
	fmt.Println("")
	fmt.Println("3. 🛡️ KEAMANAN & FIREWALL:")
	fmt.Println("   • Local : Permissive. Port 80, 443, 3306, 6379, 8080 terbuka di loopback")
	fmt.Println("             agar aplikasi lokal seperti DBeaver, TablePlus, Postman mudah konek.")
	fmt.Println("   • Prod  : Strict UFW Firewall. Port database (3306) & Redis (6379) ditutup")
	fmt.Println("             total dari publik. Hanya port 22, 80, 443 yang terbuka bebas.")
	fmt.Println("")
	fmt.Println("4. ⚙️ PROFIL PHP & RESOURCE MANAGEMENT:")
	fmt.Println("   • Local : Profil development (display_errors = On, error_reporting = E_ALL).")
	fmt.Println("             Ramah baterai laptop (watchdog service tidak boros siklus CPU).")
	fmt.Println("   • Prod  : Profil production (display_errors = Off, opcache enabled,")
	fmt.Println("             memory limit disesuaikan dengan RAM VPS, watchdog aktif 24/7).")
	fmt.Println("================================================================================")
	waitEnter(reader)
}

