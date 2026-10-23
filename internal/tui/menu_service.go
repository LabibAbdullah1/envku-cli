package tui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/servicectl"
)

func runServiceMenu(reader *bufio.Reader) {
	ctl := servicectl.New()

	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Println("       ⚡ Service Manager & Log Inspector        ")
		fmt.Println("==================================================")

		services := ctl.ListManagedServices()
		if len(services) == 0 {
			fmt.Println("No managed services detected on this system.")
		} else {
			for i, s := range services {
				statusBadge := "[INACTIVE]"
				if s.Active {
					statusBadge = "[RUNNING ]"
				}
				fmt.Printf("[%d] %s %-20s (Unit: %s)\n", i+1, statusBadge, s.Name, s.UnitName)
			}
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("Enter number [1-N] to manage service, or:")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println("==================================================")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "0" || choice == "" {
			return
		}

		var idx int
		fmt.Sscanf(choice, "%d", &idx)
		if idx >= 1 && idx <= len(services) {
			target := services[idx-1]
			runServiceActionSubmenu(ctl, target, reader)
		}
	}
}

func runServiceActionSubmenu(ctl *servicectl.Controller, s servicectl.ServiceState, reader *bufio.Reader) {
	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Printf(" Manage Service: %s (%s)\n", s.Name, s.UnitName)
		fmt.Println("==================================================")
		statusBadge := "[INACTIVE]"
		if s.Active {
			statusBadge = "[RUNNING ]"
		}
		fmt.Printf("Current State : %s (SubState: %s)\n", statusBadge, s.SubState)
		fmt.Printf("Boot Autostart: %v\n", s.Enabled)
		fmt.Println("--------------------------------------------------")
		fmt.Println("[1] Start Service")
		fmt.Println("[2] Stop Service")
		fmt.Println("[3] Restart Service")
		fmt.Println("[4] Reload Service")
		fmt.Println("[5] View Recent Journal Logs (last 40 lines)")
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Back")
		fmt.Println("==================================================")
		fmt.Print("Select action [0-5]: ")

		actionChoice, _ := reader.ReadString('\n')
		actionChoice = strings.TrimSpace(actionChoice)

		switch actionChoice {
		case "1":
			fmt.Printf("\n[*] Starting %s...\n", s.UnitName)
			if err := ctl.Control(s.UnitName, servicectl.ActionStart); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Printf("[+] %s started.\n", s.UnitName)
			}
			waitEnter(reader)
			return
		case "2":
			fmt.Printf("\n[*] Stopping %s...\n", s.UnitName)
			if err := ctl.Control(s.UnitName, servicectl.ActionStop); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Printf("[+] %s stopped.\n", s.UnitName)
			}
			waitEnter(reader)
			return
		case "3":
			fmt.Printf("\n[*] Restarting %s...\n", s.UnitName)
			if err := ctl.Control(s.UnitName, servicectl.ActionRestart); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Printf("[+] %s restarted.\n", s.UnitName)
			}
			waitEnter(reader)
			return
		case "4":
			fmt.Printf("\n[*] Reloading %s...\n", s.UnitName)
			if err := ctl.Control(s.UnitName, servicectl.ActionReload); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Printf("[+] %s reloaded.\n", s.UnitName)
			}
			waitEnter(reader)
			return
		case "5":
			clearScreen()
			fmt.Printf("=== Recent Logs for %s ===\n\n", s.UnitName)
			logs, err := ctl.GetJournalLogs(s.UnitName, 40)
			if err != nil {
				fmt.Printf("[-] Failed to read logs: %v\n", err)
			} else {
				for _, line := range logs {
					fmt.Println(line)
				}
			}
			fmt.Println("\n==========================")
			waitEnter(reader)
		case "0":
			return
		}
	}
}
