package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
	"github.com/LabibAbdullah1/envku-cli/internal/servicectl"
)

// Run initializes and executes the interactive Terminal UI
func Run() error {
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
		fmt.Println("[3] ⚙️  PHP Manager (php.ini Tweaks & Extensions)")
		fmt.Println("[4] 🌐 Site Manager (Virtual Host & Certbot SSL)")
		fmt.Println("[5] ⚡ Service Manager (systemctl & Log Inspector)")
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] 🚪 Exit")
		fmt.Println("==================================================")
		fmt.Print("Select an option [0-5]: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		choice = strings.TrimSpace(choice)

		switch choice {
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
		case "0":
			clearScreen()
			fmt.Println("Thank you for using Envku CLI. Goodbye!")
			return nil
		}
	}
}

func renderDashboard(ctl *servicectl.Controller) {
	metrics := GetSystemMetrics()

	fmt.Println("==================================================")
	fmt.Printf("   🚀 ENVKU CLI v%s - Linux Server Manager\n", config.AppVersion)
	fmt.Println("==================================================")
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
