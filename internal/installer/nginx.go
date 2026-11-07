package installer

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// NginxInstaller manages Nginx web server installation and verification
type NginxInstaller struct{}

func NewNginxInstaller() *NginxInstaller {
	return &NginxInstaller{}
}

// Status returns current installation and runtime status of Nginx
func (n *NginxInstaller) Status() PackageStatus {
	installed := IsBinaryInstalled("nginx")
	active := false
	version := ""

	if installed {
		active = IsServiceActive("nginx")
		version = n.GetVersion()
	}

	return PackageStatus{
		Name:      "nginx",
		Installed: installed,
		Active:    active,
		Version:   version,
	}
}

// GetVersion extracts the installed Nginx version string
func (n *NginxInstaller) GetVersion() string {
	cmd := exec.Command("nginx", "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Output format: "nginx version: nginx/1.24.0"
	re := regexp.MustCompile(`nginx/([0-9.]+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(string(out))
}

// Install provisions Nginx and starts the systemd service
func (n *NginxInstaller) Install() error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to install Nginx")
	}

	// 1. Clean any broken sources (such as old invalid php.list on Linux Mint) to allow apt update
	CleanBrokenSources()

	// 2. Resolve conflicting web servers (e.g. Apache2 commonly active in Linux Mint / Ubuntu Desktop)
	if IsServiceActive("apache2") {
		fmt.Println("[!] Terdeteksi Apache2 aktif di port 80. Menghentikan Apache2 agar Nginx dapat berjalan...")
		_ = RunCommand("systemctl", "stop", "apache2")
		_ = RunCommand("systemctl", "disable", "apache2")
	}

	// 3. Update apt safely
	if err := AptUpdate(); err != nil {
		return fmt.Errorf("failed to update apt repository: %w", err)
	}

	// 4. Install nginx package
	if err := AptInstall("nginx", "nginx-common"); err != nil {
		return fmt.Errorf("failed to install nginx package: %w", err)
	}

	// 5. Ensure sites-available and sites-enabled directory structure exist
	_ = os.MkdirAll("/etc/nginx/sites-available", 0755)
	_ = os.MkdirAll("/etc/nginx/sites-enabled", 0755)
	_ = os.MkdirAll("/var/www/html", 0755)

	// 6. Disable distro default "Welcome to nginx" virtual host if present to prevent conflicting default server block
	if _, err := os.Lstat("/etc/nginx/sites-enabled/default"); err == nil {
		_ = os.Remove("/etc/nginx/sites-enabled/default")
	}

	// 7. Validate configuration before starting
	if err := n.ValidateConfig(); err != nil {
		return fmt.Errorf("nginx configuration invalid: %w", err)
	}

	// 8. Enable and start Nginx service
	if err := EnableAndStartService("nginx"); err != nil {
		// Diagnostic message if port 80 is occupied
		out, _ := exec.Command("ss", "-tlpn", "sport = :80").Output()
		if len(out) > 0 {
			return fmt.Errorf("gagal menyalakan Nginx, port 80 tampaknya sedang digunakan oleh proses lain: %s (%w)", strings.TrimSpace(string(out)), err)
		}
		return fmt.Errorf("failed to enable and start nginx service: %w", err)
	}

	return nil
}

// ValidateConfig runs `nginx -t` to check syntax
func (n *NginxInstaller) ValidateConfig() error {
	cmd := exec.Command("nginx", "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("syntax test failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Reload safely validates syntax and reloads Nginx
func (n *NginxInstaller) Reload() error {
	if err := n.ValidateConfig(); err != nil {
		return err
	}
	return RunCommand("systemctl", "reload", "nginx")
}
