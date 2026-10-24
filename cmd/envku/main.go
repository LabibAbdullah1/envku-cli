package main

import (
	"fmt"
	"os"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/daemon"
	"github.com/LabibAbdullah1/envku-cli/internal/tui"
	"github.com/LabibAbdullah1/envku-cli/internal/updater"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "daemon":
			handleDaemon(args[1:])
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
		fmt.Println("    To update now, run: sudo envku update")
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
		fmt.Println("Usage: envku daemon [install|uninstall|run]")
		os.Exit(1)
	}

	switch subArgs[0] {
	case "install":
		if err := d.InstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install daemon service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Envku background daemon service installed and started.")
	case "uninstall":
		if err := d.UninstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall daemon service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Envku background daemon service removed.")
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
	fmt.Println("  envku                  Launch the interactive Terminal UI (TUI)")
	fmt.Println("  envku update           Update Envku CLI to the latest release")
	fmt.Println("  envku check-update     Check if a newer version is available")
	fmt.Println("  envku daemon install   Register and enable the Systemd background service")
	fmt.Println("  envku daemon uninstall Remove the background service")
	fmt.Println("  envku daemon run       Run background watchdog process in foreground")
	fmt.Println("  envku version          Show application version")
	fmt.Println("  envku help             Show this help message")
}
