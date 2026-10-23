package tui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/installer"
	"github.com/LabibAbdullah1/envku-cli/internal/phpmanager"
)

func runPHPMenu(reader *bufio.Reader) {
	for {
		clearScreen()
		fmt.Println("==================================================")
		fmt.Println("       ⚙️ PHP Manager (php.ini & Extensions)       ")
		fmt.Println("==================================================")
		fmt.Println("Select installed PHP version to manage:")

		for i, v := range installer.SupportedPHPVersions {
			mgr := phpmanager.New(v)
			cfg, err := mgr.ReadINIConfig()
			status := "[Not Configured]"
			if err == nil && cfg != nil {
				status = fmt.Sprintf("[Active - Memory: %s]", cfg.MemoryLimit)
			}
			fmt.Printf("[%d] PHP %s %s\n", i+1, v, status)
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println("==================================================")
		fmt.Print("Select option [0-4]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "0" {
			return
		}

		var selectedVer string
		switch choice {
		case "1":
			selectedVer = "8.2"
		case "2":
			selectedVer = "8.3"
		case "3":
			selectedVer = "8.4"
		case "4":
			selectedVer = "8.5"
		default:
			continue
		}

		runPHPVersionSubmenu(selectedVer, reader)
	}
}

func runPHPVersionSubmenu(version string, reader *bufio.Reader) {
	mgr := phpmanager.New(version)

	for {
		clearScreen()
		fmt.Printf("==================================================\n")
		fmt.Printf("        ⚙️ Managing PHP %s Configuration\n", version)
		fmt.Printf("==================================================\n")
		fmt.Println("[1] View / Tweak php.ini Directives")
		fmt.Println("[2] Toggle PHP Extensions (phpenmod / phpdismod)")
		fmt.Println("[3] Test Syntax & Reload PHP-FPM Service")
		fmt.Println("--------------------------------------------------")
		fmt.Println("[0] Back")
		fmt.Println("==================================================")
		fmt.Print("Select option [0-3]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			tweakPHPINI(mgr, reader)
		case "2":
			manageExtensions(mgr, reader)
		case "3":
			fmt.Printf("\n[*] Validating and reloading php%s-fpm...\n", version)
			if err := mgr.ReloadFPM(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
			} else {
				fmt.Printf("[+] php%s-fpm reloaded successfully!\n", version)
			}
			waitEnter(reader)
		case "0":
			return
		}
	}
}

func tweakPHPINI(mgr *phpmanager.Manager, reader *bufio.Reader) {
	clearScreen()
	cfg, err := mgr.ReadINIConfig()
	if err != nil {
		fmt.Printf("[-] Could not read php.ini for PHP %s: %v\n", mgr.Version, err)
		waitEnter(reader)
		return
	}

	fmt.Println("==================================================")
	fmt.Printf(" Current php.ini Directives (PHP %s-FPM)\n", mgr.Version)
	fmt.Println("==================================================")
	fmt.Printf("1. memory_limit        : %s\n", cfg.MemoryLimit)
	fmt.Printf("2. upload_max_filesize : %s\n", cfg.UploadMaxFilesize)
	fmt.Printf("3. post_max_size       : %s\n", cfg.PostMaxSize)
	fmt.Printf("4. max_execution_time  : %s\n", cfg.MaxExecutionTime)
	fmt.Printf("5. max_input_vars      : %s\n", cfg.MaxInputVars)
	fmt.Printf("6. date.timezone       : %s\n", cfg.Timezone)
	fmt.Println("--------------------------------------------------")
	fmt.Println("Press [Enter] to keep current value, or type new value.")

	fmt.Print("New memory_limit (e.g. 512M): ")
	val, _ := reader.ReadString('\n')
	if strings.TrimSpace(val) != "" {
		cfg.MemoryLimit = strings.TrimSpace(val)
	}

	fmt.Print("New upload_max_filesize (e.g. 64M): ")
	val, _ = reader.ReadString('\n')
	if strings.TrimSpace(val) != "" {
		cfg.UploadMaxFilesize = strings.TrimSpace(val)
	}

	fmt.Print("New post_max_size (e.g. 64M): ")
	val, _ = reader.ReadString('\n')
	if strings.TrimSpace(val) != "" {
		cfg.PostMaxSize = strings.TrimSpace(val)
	}

	fmt.Print("New max_execution_time (e.g. 300): ")
	val, _ = reader.ReadString('\n')
	if strings.TrimSpace(val) != "" {
		cfg.MaxExecutionTime = strings.TrimSpace(val)
	}

	fmt.Println("\n[*] Applying changes to php.ini (with auto-backup & validation)...")
	if err := mgr.UpdateINIConfig(*cfg); err != nil {
		fmt.Printf("[-] Error updating configuration: %v\n", err)
	} else {
		fmt.Println("[+] php.ini updated and PHP-FPM service reloaded successfully!")
	}
	waitEnter(reader)
}

func manageExtensions(mgr *phpmanager.Manager, reader *bufio.Reader) {
	for {
		clearScreen()
		exts, err := mgr.ListExtensions()
		if err != nil {
			fmt.Printf("[-] Failed to list extensions: %v\n", err)
			waitEnter(reader)
			return
		}

		fmt.Println("==================================================")
		fmt.Printf(" PHP %s Extensions Manager\n", mgr.Version)
		fmt.Println("==================================================")
		for i, ext := range exts {
			badge := "[OFF]"
			if ext.EnabledFPM {
				badge = "[ON ]"
			}
			fmt.Printf("[%2d] %s %-12s : %s\n", i+1, badge, ext.Name, ext.Description)
		}
		fmt.Println("--------------------------------------------------")
		fmt.Println("Type number to toggle extension, or [0] to exit")
		fmt.Print("Select extension: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" || input == "" {
			return
		}

		var idx int
		fmt.Sscanf(input, "%d", &idx)
		if idx >= 1 && idx <= len(exts) {
			target := exts[idx-1]
			newState := !target.EnabledFPM
			fmt.Printf("\n[*] Toggling %s to %v...\n", target.Name, newState)
			if err := mgr.ToggleExtension(target.Name, newState); err != nil {
				fmt.Printf("[-] Error toggling extension: %v\n", err)
				waitEnter(reader)
			}
		}
	}
}
