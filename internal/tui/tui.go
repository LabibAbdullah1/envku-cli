package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
	"github.com/LabibAbdullah1/envku-cli/internal/servicectl"
	"github.com/LabibAbdullah1/envku-cli/internal/updater"
)

var (
	cachedUpdateInfo *updater.UpdateInfo
	updateOnce       sync.Once
)

func checkBackgroundUpdate() {
	updateOnce.Do(func() {
		go func() {
			if info, err := updater.CheckUpdate(); err == nil && info.Available {
				cachedUpdateInfo = info
			}
		}()
	})
}

// Run initializes and executes the interactive Terminal UI
func Run() error {
	checkBackgroundUpdate()
	reader := bufio.NewReader(os.Stdin)
	inst := installer.New()
	ctl := servicectl.New()

	for {
		clearScreen()
		renderDashboard(ctl)

		fmt.Println("==================================================")
		fmt.Println("                   MAIN MENU                      ")
		fmt.Println("==================================================")
		fmt.Println("[1] 📊 Refresh Dashboard")
		fmt.Println("[2] 📦 App & Package Downloader (Nginx, MySQL, Redis, PHP, Node)")
		fmt.Println("[3] ⚙️ PHP Manager (php.ini Tweaks & Extensions)")
		fmt.Println("[4] 🌐 Site Manager (Virtual Host & Certbot SSL)")
		fmt.Println("[5] ⚡ Service Manager (systemctl & Log Inspector)")
		if cachedUpdateInfo != nil && cachedUpdateInfo.Available {
			fmt.Printf("[U] 🚀 Update Envku CLI ke v%s [UPDATE READY]\n", cachedUpdateInfo.LatestVersion)
		} else {
			fmt.Println("[U] 🔍 Periksa / Update Envku CLI")
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] 🚪 Exit")
		fmt.Println("==================================================")
		fmt.Print("Select an option [0-5, U]: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		choice = strings.TrimSpace(choice)

		switch strings.ToUpper(choice) {
		case "1":
			// Loop will naturally clear screen and refresh
			continue
		case "2":
			runInstallerMenu(inst, reader)
		case "3":
			runPHPMenu(reader)
		case "4":
			runSiteMenu(reader)
		case "5":
			runServiceMenu(reader)
		case "U":
			handleTUIUpdate(reader)
		case "0":
			clearScreen()
			fmt.Println("Thank you for using Envku CLI. Goodbye!")
			return nil
		}
	}
}

func handleTUIUpdate(reader *bufio.Reader) {
	clearScreen()
	fmt.Println("==================================================")
	fmt.Println("         🚀 ENVKU CLI UPDATE MANAGER              ")
	fmt.Println("==================================================")
	fmt.Printf(" Versi Saat Ini : v%s\n", config.AppVersion)

	info := cachedUpdateInfo
	if info == nil {
		fmt.Println("[*] Menghubungi GitHub untuk mengecek pembaruan...")
		var err error
		info, err = updater.CheckUpdate()
		if err != nil {
			fmt.Printf("[-] Gagal memeriksa pembaruan: %v\n", err)
			waitEnter(reader)
			return
		}
		cachedUpdateInfo = info
	}

	if !info.Available {
		fmt.Println("[+] Aplikasi Anda sudah menggunakan versi terbaru!")
		waitEnter(reader)
		return
	}

	fmt.Printf(" Versi Terbaru  : v%s\n", info.LatestVersion)
	if info.ReleaseName != "" {
		fmt.Printf(" Rilis          : %s\n", info.ReleaseName)
	}
	fmt.Println("--------------------------------------------------")
	if info.ReleaseNotes != "" {
		fmt.Println(" Catatan Rilis:")
		fmt.Printf(" %s\n", strings.TrimSpace(info.ReleaseNotes))
		fmt.Println("--------------------------------------------------")
	}

	fmt.Print("Perbarui Envku CLI sekarang? [Y/n]: ")
	ans, _ := reader.ReadString('\n')
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans != "" && ans != "y" && ans != "yes" {
		fmt.Println("[*] Pembaruan dibatalkan.")
		waitEnter(reader)
		return
	}

	fmt.Printf("\n[*] Mengunduh dan memasang v%s dari GitHub...\n", info.LatestVersion)
	if err := updater.PerformSelfUpdate(info); err != nil {
		fmt.Printf("[-] Pembaruan gagal: %v\n", err)
		if !installer.IsRoot() {
			fmt.Println("[-] Tips: Menimpa binary sistem membutuhkan hak akses root. Jalankan: sudo envku update")
		}
	} else {
		fmt.Printf("==================================================\n")
		fmt.Printf(" [BERHASIL] Envku CLI telah diperbarui ke v%s!\n", info.LatestVersion)
		fmt.Println(" Silakan jalankan ulang aplikasi untuk menikmati fitur baru.")
		fmt.Printf("==================================================\n")
	}
	waitEnter(reader)
}

func renderDashboard(ctl *servicectl.Controller) {
	metrics := GetSystemMetrics()

	fmt.Println("==================================================")
	fmt.Printf("   🚀 ENVKU CLI v%s - Linux Server Manager\n", config.AppVersion)
	fmt.Println("==================================================")
	if !installer.IsRoot() {
		fmt.Println("  ⚠️  Notice: Running as non-root user.")
		fmt.Println("     Run with 'sudo envku' to manage/install system services.")
		fmt.Println("--------------------------------------------------")
	}
	if cachedUpdateInfo != nil && cachedUpdateInfo.Available {
		fmt.Printf("  🔔 UPDATE TERSEDIA: v%s (Versi Anda: v%s)\n", cachedUpdateInfo.LatestVersion, config.AppVersion)
		fmt.Println("     Pilih opsi [U] untuk memperbarui aplikasi otomatis!")
		fmt.Println("--------------------------------------------------")
	}
	fmt.Printf(" CPU Model : %s\n", metrics.CPUModel)
	if metrics.RAMTotalMB > 0 {
		fmt.Printf(" RAM Usage : %d MB / %d MB (%.1f%%)\n", metrics.RAMUsedMB, metrics.RAMTotalMB, metrics.RAMUsedPerc)
	}
	fmt.Printf(" Uptime    : %s\n", metrics.Uptime)
	fmt.Println("--------------------------------------------------")
	fmt.Println(" Service Health Status:")

	services := ctl.ListManagedServices()
	if len(services) == 0 {
		fmt.Println("   (No managed services registered yet)")
	} else {
		for _, s := range services {
			badge := "[INACTIVE]"
			if s.Active {
				badge = "[RUNNING ]"
			}
			fmt.Printf("   %s %-18s (%s)\n", badge, s.Name, s.UnitName)
		}
	}
}

func clearScreen() {
	// ANSI clear screen and reset cursor
	fmt.Print("\033[H\033[2J")
}

func waitEnter(reader *bufio.Reader) {
	fmt.Print("\nPress [Enter] to return...")
	_, _ = reader.ReadString('\n')
}
