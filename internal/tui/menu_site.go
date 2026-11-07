package tui

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/sitemanager"
)

func runSiteMenu(reader *bufio.Reader) {
	mgr := sitemanager.New()

	for {
		clearScreen()
		cfg := config.GetConfig()

		fmt.Println("==================================================")
		fmt.Println("       🌐 Site Manager (Virtual Host & SSL)       ")
		fmt.Println("==================================================")
		if cfg.Environment == config.ModeLocal {
			fmt.Println(" Mode: 💻 LOCAL (Domain *.test & localhost supported)")
		} else {
			serverIP := cfg.ServerIP
			if serverIP == "" {
				serverIP = config.DetectPublicIP()
			}
			fmt.Printf(" Mode: 🌐 PRODUCTION VPS (IP: %s)\n", serverIP)
		}
		fmt.Println("--------------------------------------------------")

		sites, err := mgr.ListSites()
		if err != nil {
			fmt.Printf("[-] Error listing sites: %v\n", err)
		} else if len(sites) == 0 {
			fmt.Println("No sites configured yet.")
		} else {
			for i, s := range sites {
				statusBadge := "[DISABLED]"
				if s.Enabled {
					statusBadge = "[ACTIVE  ]"
				}
				sslBadge := "[HTTP ]"
				if s.HasSSL {
					sslBadge = "[HTTPS]"
				}
				locBadge := ""
				if s.IsLocal {
					locBadge = "[LOCAL]"
				}
				fmt.Printf("[%d] %s %s %s %-20s (Type: %s)\n", i+1, statusBadge, sslBadge, locBadge, s.Domain, s.Type)
			}
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("[A] Add New Site (Wizard)")
		if config.CanOpenBrowser() {
			fmt.Println("[O] Buka Domain di Browser Lokal (xdg-open)")
		}
		fmt.Println("[E] Enable a Site")
		fmt.Println("[D] Disable a Site")
		fmt.Println("[S] Issue / Renew Certbot SSL (Production FQDN only)")
		fmt.Println("[X] Delete a Site")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println("==================================================")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "A":
			wizardNewSite(mgr, reader)
		case "O":
			fmt.Print("Masukkan nama domain yang ingin dibuka di browser: ")
			dom, _ := reader.ReadString('\n')
			dom = strings.TrimSpace(dom)
			if dom != "" {
				fmt.Printf("[*] Membuka %s di browser lokal...\n", dom)
				if err := mgr.OpenSiteInBrowser(dom); err != nil {
					fmt.Printf("[-] Gagal membuka browser: %v\n", err)
				} else {
					fmt.Printf("[+] Halaman %s telah dibuka di browser!\n", dom)
				}
				waitEnter(reader)
			}
		case "E":
			fmt.Print("Enter domain to enable: ")
			dom, _ := reader.ReadString('\n')
			dom = strings.TrimSpace(dom)
			if dom != "" {
				if err := mgr.EnableSite(dom); err != nil {
					fmt.Printf("[-] Error: %v\n", err)
				} else {
					fmt.Println("[+] Site enabled and Nginx reloaded successfully!")
				}
				waitEnter(reader)
			}
		case "D":
			fmt.Print("Enter domain to disable: ")
			dom, _ := reader.ReadString('\n')
			dom = strings.TrimSpace(dom)
			if dom != "" {
				if err := mgr.DisableSite(dom); err != nil {
					fmt.Printf("[-] Error: %v\n", err)
				} else {
					fmt.Println("[+] Site disabled and Nginx reloaded successfully!")
				}
				waitEnter(reader)
			}
		case "S":
			fmt.Print("Enter domain to install SSL: ")
			dom, _ := reader.ReadString('\n')
			dom = strings.TrimSpace(dom)
			if dom != "" {
				if sitemanager.IsLocalDomain(dom) {
					fmt.Printf("[-] Domain '%s' adalah domain lokal/testing. Let's Encrypt hanya mendukung domain publik FQDN yang terdaftar di DNS publik.\n", dom)
					waitEnter(reader)
					continue
				}

				fmt.Print("Enter email for Let's Encrypt (optional): ")
				email, _ := reader.ReadString('\n')
				email = strings.TrimSpace(email)

				fmt.Printf("\n[*] Requesting Certbot SSL for %s...\n", dom)
				if err := mgr.RequestSSL(dom, email); err != nil {
					fmt.Printf("[-] Error: %v\n", err)
				} else {
					fmt.Println("[+] SSL certificate issued and activated successfully!")
				}
				waitEnter(reader)
			}
		case "X":
			fmt.Print("Enter domain to DELETE: ")
			dom, _ := reader.ReadString('\n')
			dom = strings.TrimSpace(dom)
			if dom != "" {
				fmt.Printf("Are you sure you want to delete %s? (y/N): ", dom)
				confirm, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
					if err := mgr.DeleteSite(dom); err != nil {
						fmt.Printf("[-] Error: %v\n", err)
					} else {
						fmt.Println("[+] Site configuration removed.")
					}
					waitEnter(reader)
				}
			}
		case "0":
			return
		}
	}
}

func wizardNewSite(mgr *sitemanager.Manager, reader *bufio.Reader) {
	clearScreen()
	cfg := config.GetConfig()

	fmt.Println("==================================================")
	fmt.Println("          🪄 New Virtual Host Setup Wizard         ")
	fmt.Println("==================================================")

	exampleDomain := "myapp.example.com"
	if cfg.Environment == config.ModeLocal {
		exampleDomain = "myapp.test"
		fmt.Println(" [TIPS LOCAL]: Gunakan domain berakhiran .test (contoh: proyekku.test)")
		fmt.Println("               Aplikasi akan otomatis mendaftarkannya ke /etc/hosts")
	} else {
		serverIP := cfg.ServerIP
		if serverIP == "" {
			serverIP = config.DetectPublicIP()
		}
		fmt.Println(" [TIPS PROD]: Pastikan domain sudah memiliki A-Record DNS yang")
		fmt.Printf("              mengarah ke IP VPS Anda: %s\n", serverIP)
	}
	fmt.Println("--------------------------------------------------")

	fmt.Printf("1. Enter Domain / Subdomain (e.g. %s): ", exampleDomain)
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)
	if domain == "" {
		fmt.Println("[-] Domain cannot be empty.")
		waitEnter(reader)
		return
	}

	addToHosts := false
	if cfg.Environment == config.ModeLocal || sitemanager.IsLocalDomain(domain) {
		fmt.Printf("   Tambahkan %s ke /etc/hosts (127.0.0.1) otomatis? [Y/n]: ", domain)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		addToHosts = (ans == "" || ans == "y" || ans == "yes")
	}

	fmt.Println("\n2. Select Application Type:")
	fmt.Println("   [1] PHP-FPM (Laravel, WordPress, Yii, CodeIgniter)")
	fmt.Println("   [2] Node.js Reverse Proxy (Next.js, Express, Nest.js)")
	fmt.Println("   [3] Static SPA (HTML, React, Vue build)")
	fmt.Print("   Select [1-3] (default 1): ")
	typeInput, _ := reader.ReadString('\n')
	typeInput = strings.TrimSpace(typeInput)

	siteType := sitemanager.SiteTypePHP
	phpVer := "8.3"
	nodePort := 3000
	docRoot := ""

	switch typeInput {
	case "2":
		siteType = sitemanager.SiteTypeNode
		fmt.Print("   Enter Local Port to Proxy (default 3000): ")
		portStr, _ := reader.ReadString('\n')
		portStr = strings.TrimSpace(portStr)
		if portStr != "" {
			nodePort, _ = strconv.Atoi(portStr)
		}
	case "3":
		siteType = sitemanager.SiteTypeStatic
		fmt.Printf("   Enter Document Root (default /var/www/%s): ", domain)
		rootStr, _ := reader.ReadString('\n')
		docRoot = strings.TrimSpace(rootStr)
	default:
		siteType = sitemanager.SiteTypePHP
		fmt.Print("   Enter PHP Version [8.2, 8.3, 8.4, 8.5] (default 8.3): ")
		verStr, _ := reader.ReadString('\n')
		verStr = strings.TrimSpace(verStr)
		if verStr != "" {
			phpVer = verStr
		}
		fmt.Printf("   Enter Document Root (default /var/www/%s/public): ", domain)
		rootStr, _ := reader.ReadString('\n')
		docRoot = strings.TrimSpace(rootStr)
	}

	wantSSL := false
	sslEmail := ""

	// SSL only offered if not local testing domain and in production mode
	if cfg.Environment == config.ModeProduction && !sitemanager.IsLocalDomain(domain) {
		fmt.Print("\n3. Automatically request Let's Encrypt SSL now? (y/N): ")
		sslInput, _ := reader.ReadString('\n')
		wantSSL = strings.ToLower(strings.TrimSpace(sslInput)) == "y"
		if wantSSL {
			fmt.Print("   Enter admin email for SSL renewal alerts: ")
			emailStr, _ := reader.ReadString('\n')
			sslEmail = strings.TrimSpace(emailStr)
		}
	}

	siteCfg := sitemanager.SiteConfig{
		Domain:        domain,
		Type:          siteType,
		DocumentRoot:  docRoot,
		PHPVersion:    phpVer,
		NodeProxyPort: nodePort,
		EnableSSL:     wantSSL,
		SSLEmail:      sslEmail,
		AddToHosts:    addToHosts,
	}

	fmt.Println("\n[*] Creating and activating Nginx virtual host...")
	if err := mgr.CreateSite(siteCfg); err != nil {
		fmt.Printf("[-] Setup failed: %v\n", err)
	} else {
		fmt.Printf("[+] Virtual Host for %s successfully created and active!\n", domain)
		if addToHosts {
			fmt.Printf("[+] %s telah dipetakan ke 127.0.0.1 di /etc/hosts\n", domain)
		}

		if config.CanOpenBrowser() {
			fmt.Print("\nBuka situs baru di browser lokal sekarang? [Y/n]: ")
			openNow, _ := reader.ReadString('\n')
			openNow = strings.ToLower(strings.TrimSpace(openNow))
			if openNow == "" || openNow == "y" || openNow == "yes" {
				_ = mgr.OpenSiteInBrowser(domain)
			}
		}
	}
	waitEnter(reader)
}
