package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// CanOpenBrowser checks if the current runtime environment supports opening a desktop browser
func CanOpenBrowser() bool {
	switch runtime.GOOS {
	case "windows", "darwin":
		return true
	case "linux":
		// Check WSL
		if os.Getenv("WSL_DISTRO_NAME") != "" {
			if _, err := exec.LookPath("wslview"); err == nil {
				return true
			}
		}

		// Check for X11 / Wayland graphical session
		hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
		if !hasDisplay {
			return false
		}

		// Check for browser launcher utilities
		candidates := []string{"xdg-open", "gio", "gnome-open", "x-www-browser", "firefox", "google-chrome", "chromium"}
		for _, bin := range candidates {
			if _, err := exec.LookPath(bin); err == nil {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// OpenBrowser opens a given URL in the default local system browser
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		if os.Getenv("WSL_DISTRO_NAME") != "" {
			if _, err := exec.LookPath("wslview"); err == nil {
				return exec.Command("wslview", url).Start()
			}
		}

		candidates := []string{"xdg-open", "gio", "gnome-open", "x-www-browser", "firefox", "google-chrome", "chromium"}
		for _, bin := range candidates {
			if _, err := exec.LookPath(bin); err == nil {
				if bin == "gio" {
					return exec.Command("gio", "open", url).Start()
				}
				return exec.Command(bin, url).Start()
			}
		}
		return fmt.Errorf("no desktop browser or xdg-open found (system might be a headless server)")
	default:
		return fmt.Errorf("unsupported operating system for browser opening: %s", runtime.GOOS)
	}
}

