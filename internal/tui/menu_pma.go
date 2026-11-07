package tui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/firewall"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
)

func runPHPMyAdminMenu(pma *installer.PHPMyAdminInstaller, reader *bufio.Reader) {
	fw := firewall.New()
	mysqlInst := installer.NewMySQLInstaller()

	for {
		clearScreen()
		cfg := config.GetConfig()
		st := pma.Status()

		fmt.Println("==================================================")
		fmt.Println("         🗄️  phpMyAdmin & Database Access         ")
		fmt.Println("==================================================")
		fmt.Printf(" Status phpMyAdmin : %s\n", renderBadge(st.Installed, st.Version))
		fmt.Printf(" Mode Environment  : %s\n", strings.ToUpper(string(cfg.Environment)))
		fmt.Printf(" Target Port       : %d\n", cfg.PMAPort)

		if st.Installed {
			url := pma.GetAccessURL()
			fmt.Printf(" URL Akses         : %s\n", url)
			if cfg.Environment == config.ModeLocal {
				fmt.Println(" Akses Lokal       : ✅ Browser Lokal didukung (xdg-open / wslview)")
				fmt.Println(" Kredensial MySQL  : 👉 User: root | Password: (Kosong / Tanpa Password)")
			} else {
				fmt.Println(" Proteksi Produksi : 🛡️ Dibatasi oleh Firewall & Nginx allowlist")
				if len(cfg.AllowedAdminIPs) > 0 {
					fmt.Printf(" Allowed Admin IPs : %s\n", strings.Join(cfg.AllowedAdminIPs, ", "))
				}
			}
		}
		fmt.Println("--------------------------------------------------")

		if !st.Installed {
			fmt.Println("[1] 📥 Install phpMyAdmin")
			if cfg.Environment == config.ModeProduction {
				fmt.Println("    (⚠️ Perhatian: Di mode Production, akses akan otomatis dilindungi firewall)")
			} else {
				fmt.Println("    (Otomatis disiapkan untuk akses http://localhost:8080 & root tanpa password)")
			}
		} else {
			if cfg.Environment == config.ModeLocal {
				fmt.Println("[1] 🌐 Buka phpMyAdmin di Browser (xdg-open)")
				fmt.Println("[P] 🔓 Setel / Reset MySQL root Tanpa Password")
			} else {
				fmt.Println("[1] ℹ️  Cara Akses phpMyAdmin di Production (VPS)")
				fmt.Println("[2] 🛡️  Konfigurasi IP Whitelist Admin")
				fmt.Println("[3] 🔒 Terapkan Ulang Proteksi Firewall UFW")
			}
			fmt.Println("[R] 🔄 Re-install / Update Konfigurasi phpMyAdmin")
			fmt.Println("[X] 🗑️  Uninstall phpMyAdmin")
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Kembali ke Menu Utama")
		fmt.Println("==================================================")
		fmt.Print("Pilih opsi: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "1":
			if !st.Installed {
				fmt.Println("\n[*] Memulai instalasi phpMyAdmin...")
				if err := pma.Install(); err != nil {
					fmt.Printf("[-] Instalasi gagal: %v\n", err)
				} else {
					fmt.Println("[+] phpMyAdmin berhasil dipasang!")
					if cfg.Environment == config.ModeProduction {
						fmt.Println("[*] Menerapkan perlindungan firewall production...")
						_ = fw.ApplyProductionProfile(cfg.PMAPort, cfg.AllowedAdminIPs)
					} else if config.CanOpenBrowser() {
						fmt.Println("[*] Membuka phpMyAdmin di browser lokal...")
						_ = pma.OpenInBrowser()
					}
				}
				waitEnter(reader)
			} else if cfg.Environment == config.ModeLocal {
				fmt.Println("\n[*] Membuka browser lokal...")
				if err := pma.OpenInBrowser(); err != nil {
					fmt.Printf("[-] Gagal membuka browser otomatis: %v\n", err)
					fmt.Printf("    Buka manual di: %s\n", pma.GetAccessURL())
				} else {
					fmt.Printf("[+] Halaman phpMyAdmin dibuka di browser: %s\n", pma.GetAccessURL())
					fmt.Println("    (Gunakan username: root, password: [kosongkan])")
				}
				waitEnter(reader)
			} else {
				showProductionAccessGuide(pma, reader)
			}

		case "P":
			if cfg.Environment == config.ModeLocal {
				fmt.Println("\n[*] Mengonfigurasi MySQL root tanpa password (localhost & 127.0.0.1)...")
				if err := mysqlInst.ConfigureLocalRootNoPassword(); err != nil {
					fmt.Printf("[-] Gagal mengonfigurasi root MySQL: %v\n", err)
				} else {
					fmt.Println("[+] Berhasil! Akun MySQL root kini dapat login tanpa password.")
					fmt.Println("    Anda bisa langsung masuk ke phpMyAdmin atau aplikasi lokal tanpa mengisi password.")
				}
				waitEnter(reader)
			}

		case "2":
			if st.Installed && cfg.Environment == config.ModeProduction {
				configureAdminIPWhitelist(reader)
				_ = pma.Install() // Re-generate nginx config with new IPs
				_ = fw.ApplyProductionProfile(cfg.PMAPort, config.GetConfig().AllowedAdminIPs)
				waitEnter(reader)
			}

		case "3":
			if st.Installed && cfg.Environment == config.ModeProduction {
				fmt.Println("\n[*] Menerapkan proteksi UFW Firewall untuk phpMyAdmin...")
				if err := fw.ApplyProductionProfile(cfg.PMAPort, cfg.AllowedAdminIPs); err != nil {
					fmt.Printf("[-] Gagal mengatur firewall: %v\n", err)
				} else {
					fmt.Println("[+] Firewall UFW berhasil dikonfigurasi! Port hanya terbuka untuk whitelist.")
				}
				waitEnter(reader)
			}

		case "R":
			if st.Installed {
				fmt.Println("\n[*] Memperbarui instalasi & konfigurasi phpMyAdmin...")
				if err := pma.Install(); err != nil {
					fmt.Printf("[-] Error: %v\n", err)
				} else {
					fmt.Println("[+] phpMyAdmin berhasil diperbarui!")
				}
				waitEnter(reader)
			}

		case "X":
			if st.Installed {
				fmt.Print("Yakin ingin menghapus instalasi phpMyAdmin? (y/N): ")
				ans, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(ans)) == "y" {
					if err := pma.Uninstall(); err != nil {
						fmt.Printf("[-] Gagal menghapus: %v\n", err)
					} else {
						fmt.Println("[+] phpMyAdmin telah dihapus dari sistem.")
					}
					waitEnter(reader)
				}
			}

		case "0":
			return
		}
	}
}

func showProductionAccessGuide(pma *installer.PHPMyAdminInstaller, reader *bufio.Reader) {
	clearScreen()
	cfg := config.GetConfig()
	serverIP := cfg.ServerIP
	if serverIP == "" {
		serverIP = config.DetectPublicIP()
	}

	fmt.Println("==================================================")
	fmt.Println("   🛡️ PANDUAN AKSES PHPMYADMIN DI PRODUCTION      ")
	fmt.Println("==================================================")
	fmt.Println("Di server VPS publik, phpMyAdmin TIDAK dibuka bebas ke")
	fmt.Println("publik untuk mencegah serangan brute-force dan eksploitasi.")
	fmt.Println("")
	fmt.Println("PILIHAN METODE AKSES AMAN:")
	fmt.Println("")
	fmt.Println("1. 🔑 METODE TERBAIK: SSH TUNNEL (Paling Aman)")
	fmt.Println("   Jalankan perintah ini di komputer / laptop Anda:")
	fmt.Printf("   ssh -N -L %d:127.0.0.1:%d user@%s\n", cfg.PMAPort, cfg.PMAPort, serverIP)
	fmt.Println("   Lalu buka browser lokal Anda di:")
	fmt.Printf("   👉 http://localhost:%d\n", cfg.PMAPort)
	fmt.Println("")
	fmt.Println("2. 🌐 METODE IP WHITELIST (Akses Langsung via IP/Domain)")
	fmt.Println("   Daftarkan IP Publik koneksi internet Anda di menu:")
	fmt.Println("   '[2] Konfigurasi IP Whitelist Admin'")
	fmt.Println("   Setelah itu, Anda bisa mengakses langsung via:")
	fmt.Printf("   👉 http://%s:%d\n", serverIP, cfg.PMAPort)
	fmt.Println("==================================================")
	waitEnter(reader)
}

func configureAdminIPWhitelist(reader *bufio.Reader) {
	cfg := config.GetConfig()
	fmt.Println("\n--- Konfigurasi IP Whitelist phpMyAdmin ---")
	fmt.Printf("IP aktif saat ini: %s\n", strings.Join(cfg.AllowedAdminIPs, ", "))
	fmt.Print("Masukkan IP Publik admin Anda (pisahkan koma jika lebih dari satu): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "" {
		parts := strings.Split(input, ",")
		var ips []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				ips = append(ips, p)
			}
		}
		if len(ips) > 0 {
			cfg.AllowedAdminIPs = ips
			_ = config.SaveConfig(cfg)
			fmt.Printf("[+] Daftar IP Whitelist berhasil disimpan: %s\n", strings.Join(ips, ", "))
		}
	}
}
