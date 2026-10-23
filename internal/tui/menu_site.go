package tui

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/sitemanager"
)

func runSiteMenu(reader *bufio.Reader) {
	mgr := sitemanager.New()

	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Println("       🌐 Site Manager (Virtual Host & SSL)       ")
		fmt.Println("==================================================")

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
				fmt.Printf("[%d] %s %s %-20s (Type: %s)\n", i+1, statusBadge, sslBadge, s.Domain, s.Type)
			}
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("[A] Add New Site (Wizard)")
		fmt.Println("[E] Enable a Site")
		fmt.Println("[D] Disable a Site")
		fmt.Println("[S] Issue / Renew Certbot SSL")
		fmt.Println("[X] Delete a Site")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println("==================================================")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "A":
			wizardNewSite(mgr, reader)
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
			fmt.Print("Enter email for Let's Encrypt (optional): ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)
			if dom != "" {
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
	fmt.Println("==================================================")
	fmt.Println("          🪄 New Virtual Host Setup Wizard         ")
	fmt.Println("==================================================")

	fmt.Print("1. Enter Domain / Subdomain (e.g. app.example.com): ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)
	if domain == "" {
		fmt.Println("[-] Domain cannot be empty.")
		waitEnter(reader)
		return
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

	fmt.Print("\n3. Automatically request Let's Encrypt SSL now? (y/N): ")
	sslInput, _ := reader.ReadString('\n')
	wantSSL := strings.ToLower(strings.TrimSpace(sslInput)) == "y"
	sslEmail := ""
	if wantSSL {
		fmt.Print("   Enter admin email for SSL renewal alerts: ")
		emailStr, _ := reader.ReadString('\n')
		sslEmail = strings.TrimSpace(emailStr)
	}

	cfg := sitemanager.SiteConfig{
		Domain:        domain,
		Type:          siteType,
		DocumentRoot:  docRoot,
		PHPVersion:    phpVer,
		NodeProxyPort: nodePort,
		EnableSSL:     wantSSL,
		SSLEmail:      sslEmail,
	}

	fmt.Println("\n[*] Creating and activating Nginx virtual host...")
	if err := mgr.CreateSite(cfg); err != nil {
		fmt.Printf("[-] Setup failed: %v\n", err)
	} else {
		fmt.Printf("[+] Virtual Host for %s successfully created and active!\n", domain)
	}
	waitEnter(reader)
}
