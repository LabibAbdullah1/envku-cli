package tui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/installer"
)

func runInstallerMenu(inst *installer.Installer, reader *bufio.Reader) {
	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Println("       📦 App & Package Downloader / Manager      ")
		fmt.Println("==================================================")

		nginxSt := inst.Nginx.Status()
		mysqlSt := inst.MySQL.Status()
		redisSt := inst.Redis.Status()
		nodeSt := inst.Node.Status()

		fmt.Printf("[1] Nginx Web Server      : %s\n", renderBadge(nginxSt.Installed, nginxSt.Version))
		fmt.Printf("[2] MySQL Database        : %s\n", renderBadge(mysqlSt.Installed, mysqlSt.Version))
		fmt.Printf("[3] Redis In-Memory Cache : %s\n", renderBadge(redisSt.Installed, redisSt.Version))
		fmt.Printf("[4] Node.js & NVM         : %s\n", renderBadge(nodeSt.Installed, nodeSt.Version))
		fmt.Println("[5] PHP Multi-Version Installer (8.2, 8.3, 8.4, 8.5)")
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println("==================================================")
		fmt.Print("Select option [0-5]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("\n[*] Installing / Verifying Nginx...")
			if err := inst.Nginx.Install(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Println("[+] Nginx installed and running successfully!")
			}
			waitEnter(reader)
		case "2":
			fmt.Println("\n[*] Installing / Verifying MySQL Server...")
			if err := inst.MySQL.Install(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Println("[+] MySQL Server installed and running successfully!")
			}
			waitEnter(reader)
		case "3":
			fmt.Println("\n[*] Installing / Verifying Redis Server...")
			if err := inst.Redis.Install(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Println("[+] Redis installed and running successfully!")
			}
			waitEnter(reader)
		case "4":
			fmt.Println("\n[*] Installing NVM and Node.js LTS...")
			if err := inst.Node.Install(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Println("[+] Node.js LTS and NVM setup successfully!")
			}
			waitEnter(reader)
		case "5":
			runPHPInstallSubmenu(inst.PHP, reader)
		case "0":
			return
		}
	}
}

func runPHPInstallSubmenu(phpInst *installer.PHPInstaller, reader *bufio.Reader) {
	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Println("         🐘 PHP Multi-Version Installer           ")
		fmt.Println("==================================================")
		for i, ver := range installer.SupportedPHPVersions {
			st := phpInst.Status(ver)
			fmt.Printf("[%d] Install PHP %s-FPM Stack : %s\n", i+1, ver, renderBadge(st.Installed, st.Version))
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Back to Installer Menu")
		fmt.Println("==================================================")
		fmt.Print("Select PHP version to install [0-4]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "0" {
			return
		}

		var targetVer string
		switch choice {
		case "1":
			targetVer = "8.2"
		case "2":
			targetVer = "8.3"
		case "3":
			targetVer = "8.4"
		case "4":
			targetVer = "8.5"
		default:
			continue
		}

		fmt.Printf("\n[*] Installing PHP %s with FPM and core extensions...\n", targetVer)
		if err := phpInst.InstallVersion(targetVer); err != nil {
			fmt.Printf("[-] Error: %v\n", err)
		} else {
			fmt.Printf("[+] PHP %s-FPM stack installed and active!\n", targetVer)
		}
		waitEnter(reader)
	}
}

func renderBadge(installed bool, version string) string {
	if installed {
		if version != "" {
			return fmt.Sprintf("[INSTALLED - v%s]", version)
		}
		return "[INSTALLED]"
	}
	return "[NOT INSTALLED]"
}
