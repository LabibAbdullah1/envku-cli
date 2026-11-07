package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/daemon"
	"github.com/LabibAbdullah1/envku-cli/internal/firewall"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
	"github.com/LabibAbdullah1/envku-cli/internal/tui"
	"github.com/LabibAbdullah1/envku-cli/internal/uninstaller"
	"github.com/LabibAbdullah1/envku-cli/internal/updater"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "daemon":
			handleDaemon(args[1:])
			return
		case "mode", "env":
			handleMode(args[1:])
			return
		case "pma", "phpmyadmin":
			handlePMA(args[1:])
			return
		case "uninstall":
			handleUninstall(args[1:])
			return
		case "update":
			handleUpdate()
			return
		case "check-update":
			handleCheckUpdate()
			return
		case "version", "-v", "--version":
			fmt.Printf("%s version %s\n", config.AppName, config.AppVersion)
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	// Default: Launch interactive Terminal UI
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Envku CLI: %v\n", err)
		os.Exit(1)
	}
}

func handleUninstall(subArgs []string) {
	uninst := uninstaller.New()

	isSafe := false
	isPurge := false
	isForce := false

	for _, a := range subArgs {
		switch a {
		case "--safe":
			isSafe = true
		case "--purge":
			isPurge = true
		case "--force", "-f":
			isForce = true
		}
	}

	if isSafe {
		fmt.Println("[*] Menjalankan Uninstall Mode Aman (Data website & database tetap utuh)...")
		if err := uninst.UninstallSafe(); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Gagal uninstall: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[+] Envku CLI dan service background berhasil dicopot.")
		fmt.Println("    File website (/var/www), database MySQL, Nginx & PHP tetap aman.")
		return
	}

	if isPurge {
		if !isForce {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("⚠️  PERINGATAN: Mode PURGE akan menghapus seluruh data web, database, Nginx, dan PHP.")
			fmt.Print("Ketik 'HAPUS' untuk konfirmasi: ")
			ans, _ := reader.ReadString('\n')
			if strings.TrimSpace(ans) != "HAPUS" && strings.TrimSpace(ans) != "PURGE" {
				fmt.Println("[-] Konfirmasi dibatalkan.")
				return
			}
		}
		opts := uninstaller.PurgeOptions{
			RemoveWebData:     true,
			RemoveDatabases:   true,
			RemovePackages:    true,
			RemoveHostsConfig: true,
		}
		if err := uninst.UninstallPurge(opts); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Gagal melakukan purge: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[+] Seluruh sistem telah dibersihkan sepenuhnya.")
		return
	}

	// Interactive prompt if no flags provided
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("==================================================")
	fmt.Println("           🗑️  UNINSTALL ENVKU CLI               ")
	fmt.Println("==================================================")
	fmt.Println("Pilih salah satu metode pencopotan aplikasi:")
	fmt.Println("  [1] 🛡️  Hapus Envku CLI Saja (DATA & WEBSITE AMAN)")
	fmt.Println("      • Menghapus binary 'envku-cli' (dan alias 'envku') serta background service Systemd.")
	fmt.Println("      • Database MySQL, file website (/var/www), Nginx, PHP, dan SSL")
	fmt.Println("        TETAP AMAN 100% dan web server tetap berjalan normal.")
	fmt.Println("")
	fmt.Println("  [2] 💣 Hapus Bersih Seluruhnya (PURGE / COMPLETE WIPE)")
	fmt.Println("      • Menghapus Envku CLI dan SELURUH paket web stack.")
	fmt.Println("      • Menghapus direktori website (/var/www) dan database (/var/lib/mysql).")
	fmt.Println("      • ⚠️  PERINGATAN: Semua data dan website akan HILANG PERMANEN!")
	fmt.Println("--------------------------------------------------")
	fmt.Print("Pilihan Anda [1-2] (default: 1): ")

	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(ans)

	if ans == "2" || strings.ToLower(ans) == "purge" {
		fmt.Print("\nKetik 'HAPUS' (huruf besar) untuk melanjutkan penghapusan total: ")
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(confirm) != "HAPUS" && strings.TrimSpace(confirm) != "PURGE" {
			fmt.Println("[-] Konfirmasi tidak sesuai. Penghapusan dibatalkan.")
			return
		}
		opts := uninstaller.PurgeOptions{
			RemoveWebData:     true,
			RemoveDatabases:   true,
			RemovePackages:    true,
			RemoveHostsConfig: true,
		}
		if err := uninst.UninstallPurge(opts); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[+] Seluruh sistem telah dibersihkan sepenuhnya.")
	} else {
		if err := uninst.UninstallSafe(); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[+] Envku CLI dan service background berhasil dicopot.")
		fmt.Println("    File website (/var/www), database MySQL, Nginx & PHP tetap aman.")
	}
}

func handleMode(subArgs []string) {
	cfg := config.GetConfig()

	if len(subArgs) == 0 {
		fmt.Printf("Current Environment Mode: %s\n", strings.ToUpper(string(cfg.Environment)))
		if cfg.Environment == config.ModeLocal {
			fmt.Println("Profile: Local Development (Linux Desktop / Laptop / WSL)")
			fmt.Println("Browser: Local browser direct access enabled")
		} else {
			fmt.Println("Profile: Production Server (VPS / Cloud)")
			fmt.Println("Security: Strict firewall and public IP access rules enabled")
		}
		fmt.Println("\nTo switch mode: sudo envku-cli mode [local|production]")
		return
	}

	target := strings.ToLower(subArgs[0])
	fw := firewall.New()

	switch target {
	case "local", "dev", "development":
		if err := config.SetEnvironment(config.ModeLocal); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Failed to set local mode: %v\n", err)
			os.Exit(1)
		}
		_ = fw.ApplyLocalProfile(cfg.PMAPort)
		fmt.Println("[+] Successfully switched to Mode: LOCAL DEVELOPMENT")
		fmt.Println("    • phpMyAdmin and local sites will open directly in local browser")
		fmt.Println("    • Relaxed developer firewall profile active")
	case "prod", "production", "server", "vps":
		if err := config.SetEnvironment(config.ModeProduction); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Failed to set production mode: %v\n", err)
			os.Exit(1)
		}
		_ = fw.ApplyProductionProfile(cfg.PMAPort, cfg.AllowedAdminIPs)
		fmt.Println("[+] Successfully switched to Mode: PRODUCTION SERVER (VPS)")
		fmt.Println("    • Strict UFW firewall profile enforced")
		fmt.Println("    • phpMyAdmin access restricted to IP whitelist / SSH tunnel")
	default:
		fmt.Printf("Unknown environment mode: %s. Use 'local' or 'production'.\n", target)
		os.Exit(1)
	}
}

func handlePMA(subArgs []string) {
	pma := installer.NewPHPMyAdminInstaller()
	cfg := config.GetConfig()

	if len(subArgs) == 0 || subArgs[0] == "status" {
		st := pma.Status()
		fmt.Printf("phpMyAdmin Status: %s\n", renderBadgeCli(st.Installed, st.Version))
		fmt.Printf("Environment Mode : %s\n", strings.ToUpper(string(cfg.Environment)))
		if st.Installed {
			fmt.Printf("Access URL       : %s\n", pma.GetAccessURL())
		}
		return
	}

	switch subArgs[0] {
	case "open", "browser":
		if !pma.IsInstalled() {
			fmt.Println("[-] phpMyAdmin is not installed. Run 'sudo envku-cli' -> menu [6] to install.")
			os.Exit(1)
		}
		fmt.Println("[*] Opening phpMyAdmin in default browser...")
		if err := pma.OpenInBrowser(); err != nil {
			fmt.Fprintf(os.Stderr, "[-] Could not open browser: %v\n", err)
			fmt.Printf("    Open manually at: %s\n", pma.GetAccessURL())
			os.Exit(1)
		}
		fmt.Println("[+] Opened successfully.")
	case "url":
		fmt.Println(pma.GetAccessURL())
	default:
		fmt.Println("Usage: envku-cli pma [status|open|url]")
	}
}

func renderBadgeCli(installed bool, version string) string {
	if installed {
		if version != "" {
			return fmt.Sprintf("INSTALLED (v%s)", version)
		}
		return "INSTALLED"
	}
	return "NOT INSTALLED"
}

func handleCheckUpdate() {
	fmt.Printf("[*] Checking for updates (current version: v%s)...\n", config.AppVersion)
	info, err := updater.CheckUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Failed to check for updates: %v\n", err)
		os.Exit(1)
	}

	if info.Available {
		fmt.Printf("[+] New version available: v%s!\n", info.LatestVersion)
		if info.ReleaseName != "" {
			fmt.Printf("    Release: %s\n", info.ReleaseName)
		}
		fmt.Println("    To update now, run: sudo envku-cli update")
	} else {
		fmt.Printf("[+] Envku CLI is up to date (v%s).\n", config.AppVersion)
	}
}

func handleUpdate() {
	fmt.Printf("[*] Checking for latest release from GitHub...\n")
	info, err := updater.CheckUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Failed to check updates: %v\n", err)
		os.Exit(1)
	}

	if !info.Available {
		fmt.Printf("[+] You are already using the latest version of Envku CLI (v%s).\n", config.AppVersion)
		return
	}

	fmt.Printf("[+] Found update: v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)
	fmt.Printf("[*] Downloading binary from %s...\n", info.DownloadURL)

	if err := updater.PerformSelfUpdate(info); err != nil {
		fmt.Fprintf(os.Stderr, "[-] Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("==================================================\n")
	fmt.Printf(" [SUCCESS] Envku CLI updated to v%s!\n", info.LatestVersion)
	fmt.Printf("==================================================\n")
}

func handleDaemon(subArgs []string) {
	d := daemon.New()

	if len(subArgs) == 0 {
		fmt.Println("Usage: envku-cli daemon [install|uninstall|run]")
		os.Exit(1)
	}

	switch subArgs[0] {
	case "install":
		if err := d.InstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install daemon service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Envku CLI background daemon service installed and started.")
	case "uninstall":
		if err := d.UninstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall daemon service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Envku CLI background daemon service removed.")
	case "run":
		if err := d.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon runtime error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown daemon action: %s. Use install, uninstall, or run.\n", subArgs[0])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("%s - Lightweight Linux & VPS Environment Manager (v%s)\n\n", config.AppName, config.AppVersion)
	fmt.Println("Usage:")
	fmt.Println("  envku-cli                  Launch the interactive Terminal UI (TUI)")
	fmt.Println("  envku-cli mode [local|prod]View or switch environment mode (Local vs Production)")
	fmt.Println("  envku-cli pma [open|status]Manage or open phpMyAdmin in local desktop browser")
	fmt.Println("  envku-cli uninstall        Uninstall Envku CLI (with Safe or Purge data options)")
	fmt.Println("  envku-cli update           Update Envku CLI to the latest release")
	fmt.Println("  envku-cli check-update     Check if a newer version is available")
	fmt.Println("  envku-cli daemon install   Register and enable the Systemd background service")
	fmt.Println("  envku-cli daemon uninstall Remove the background service")
	fmt.Println("  envku-cli daemon run       Run background watchdog process in foreground")
	fmt.Println("  envku-cli version          Show application version")
	fmt.Println("  envku-cli help             Show this help message")
	fmt.Println("\nAlias note: 'envku' can be used as a convenient shorthand for 'envku-cli'.")
}
