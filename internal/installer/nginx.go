package installer

import (
	"fmt"
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

	if err := AptUpdate(); err != nil {
		return fmt.Errorf("failed to update apt repository: %w", err)
	}

	if err := AptInstall("nginx"); err != nil {
		return fmt.Errorf("failed to install nginx package: %w", err)
	}

	// Validate default configuration before starting
	if err := n.ValidateConfig(); err != nil {
		return fmt.Errorf("nginx default configuration invalid: %w", err)
	}

	if err := EnableAndStartService("nginx"); err != nil {
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
