package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
			// In sudo sessions, DISPLAY may be stripped by env_reset in /etc/sudoers.
			// Check if X11 socket exists or user desktop runtime dir exists.
			if _, err := os.Stat("/tmp/.X11-unix/X0"); err == nil {
				hasDisplay = true
			} else if sudoUser := getDesktopUser(); sudoUser != "" {
				uid, _ := getUserInfo(sudoUser)
				userBus := fmt.Sprintf("/run/user/%s/bus", uid)
				if _, err := os.Stat(userBus); err == nil {
					hasDisplay = true
				}
			}
		}

		if !hasDisplay {
			return false
		}

		// Check for browser launcher utilities or browsers
		candidates := []string{
			"xdg-open", "gio", "x-www-browser", "sensible-browser",
			"firefox", "google-chrome", "chromium", "chromium-browser", "brave-browser",
		}
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

		desktopUser := getDesktopUser()

		type candidate struct {
			bin  string
			args []string
		}

		candidates := []candidate{
			{bin: "xdg-open", args: []string{url}},
			{bin: "gio", args: []string{"open", url}},
			{bin: "x-www-browser", args: []string{url}},
			{bin: "sensible-browser", args: []string{url}},
			{bin: "firefox", args: []string{url}},
			{bin: "google-chrome", args: []string{url}},
			{bin: "chromium", args: []string{url}},
			{bin: "chromium-browser", args: []string{url}},
			{bin: "brave-browser", args: []string{url}},
		}

		var lastErr error
		for _, c := range candidates {
			binPath, err := exec.LookPath(c.bin)
			if err != nil {
				continue
			}

			var cmd *exec.Cmd
			if desktopUser != "" && os.Geteuid() == 0 {
				// Running with sudo/root: Launch as desktop user with their GUI session environment.
				// Browsers (Firefox, Chrome) and D-Bus reject execution as root.
				cmd = buildUserDesktopCommand(desktopUser, binPath, c.args)
			} else {
				cmd = exec.Command(binPath, c.args...)
			}

			if err := tryLaunch(cmd); err == nil {
				return nil
			} else {
				lastErr = err
			}
		}

		if lastErr != nil {
			return fmt.Errorf("gagal membuka browser grafis: %w", lastErr)
		}
		return fmt.Errorf("tidak ditemukan utilitas browser lokal (seperti xdg-open atau firefox)")
	default:
		return fmt.Errorf("unsupported operating system for browser opening: %s", runtime.GOOS)
	}
}

// getDesktopUser detects the actual logged in GUI user if running under sudo or pkexec
func getDesktopUser() string {
	if u := os.Getenv("SUDO_USER"); u != "" && u != "root" {
		return u
	}
	if u := os.Getenv("DOAS_USER"); u != "" && u != "root" {
		return u
	}
	if u := os.Getenv("PKEXEC_UID"); u != "" && u != "0" {
		out, err := exec.Command("id", "-nu", u).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

// getUserInfo returns user UID and home directory
func getUserInfo(username string) (string, string) {
	uid := "1000"
	home := "/home/" + username

	outUID, err := exec.Command("id", "-u", username).Output()
	if err == nil && len(outUID) > 0 {
		uid = strings.TrimSpace(string(outUID))
	}

	outPasswd, err := exec.Command("getent", "passwd", username).Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(outPasswd)), ":")
		if len(parts) >= 6 && parts[5] != "" {
			home = parts[5]
		}
	}

	return uid, home
}

// buildUserDesktopCommand wraps a command to run with dropped privileges as the logged-in desktop user
func buildUserDesktopCommand(username, bin string, args []string) *exec.Cmd {
	uid, home := getUserInfo(username)

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")

	xauth := os.Getenv("XAUTHORITY")
	if xauth == "" {
		defaultXauth := filepath.Join(home, ".Xauthority")
		if _, err := os.Stat(defaultXauth); err == nil {
			xauth = defaultXauth
		}
	}

	runtimeDir := fmt.Sprintf("/run/user/%s", uid)
	dbusBus := fmt.Sprintf("unix:path=%s/bus", runtimeDir)

	envVars := []string{
		fmt.Sprintf("HOME=%s", home),
		fmt.Sprintf("USER=%s", username),
		fmt.Sprintf("LOGNAME=%s", username),
		fmt.Sprintf("DISPLAY=%s", display),
		fmt.Sprintf("XDG_RUNTIME_DIR=%s", runtimeDir),
		fmt.Sprintf("DBUS_SESSION_BUS_ADDRESS=%s", dbusBus),
	}
	if waylandDisplay != "" {
		envVars = append(envVars, fmt.Sprintf("WAYLAND_DISPLAY=%s", waylandDisplay))
	}
	if xauth != "" {
		envVars = append(envVars, fmt.Sprintf("XAUTHORITY=%s", xauth))
	}

	// Prepare sudo -u <username> -H env <envVars...> <bin> <args...>
	sudoArgs := []string{"-u", username, "-H", "env"}
	sudoArgs = append(sudoArgs, envVars...)
	sudoArgs = append(sudoArgs, bin)
	sudoArgs = append(sudoArgs, args...)

	return exec.Command("sudo", sudoArgs...)
}

// tryLaunch executes a command and verifies it does not immediately fail/crash within 250ms
func tryLaunch(cmd *exec.Cmd) error {
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// If command finished within 250ms with an error, it failed
		if err != nil {
			return err
		}
		// Finished with exit 0 (e.g. xdg-open opened URL in existing browser instance)
		return nil
	case <-time.After(250 * time.Millisecond):
		// Process is still alive and running after 250ms (e.g. browser launched new window)
		return nil
	}
}
