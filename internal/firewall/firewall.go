package firewall

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
)

// Controller manages Linux firewall rules (UFW)
type Controller struct{}

// New creates a new Firewall Controller instance
func New() *Controller {
	return &Controller{}
}

// IsInstalled checks if UFW (Uncomplicated Firewall) binary is available
func (f *Controller) IsInstalled() bool {
	return installer.IsBinaryInstalled("ufw")
}

// IsActive checks if UFW is currently active and enforcing rules
func (f *Controller) IsActive() bool {
	if !f.IsInstalled() {
		return false
	}
	out, err := f.runQuiet("ufw", "status")
	return err == nil && strings.Contains(strings.ToLower(out), "status: active")
}

// EnsureInstalled installs ufw package if not already present
func (f *Controller) EnsureInstalled() error {
	if f.IsInstalled() {
		return nil
	}
	if !installer.IsRoot() {
		return fmt.Errorf("root privileges required to install ufw")
	}
	return installer.AptInstall("ufw")
}

// ApplyProductionProfile applies strict security rules suited for Public VPS environments
func (f *Controller) ApplyProductionProfile(pmaPort int, allowedAdminIPs []string) error {
	if err := f.EnsureInstalled(); err != nil {
		return err
	}

	if !installer.IsRoot() {
		return fmt.Errorf("root privileges required to configure firewall")
	}

	// Set baseline defaults
	_, _ = f.runQuiet("ufw", "default", "deny", "incoming")
	_, _ = f.runQuiet("ufw", "default", "allow", "outgoing")

	// Essential public ports
	_, _ = f.runQuiet("ufw", "allow", "22/tcp")  // SSH
	_, _ = f.runQuiet("ufw", "allow", "80/tcp")  // HTTP
	_, _ = f.runQuiet("ufw", "allow", "443/tcp") // HTTPS

	// Block internal database & cache ports from direct public exposure
	_, _ = f.runQuiet("ufw", "deny", "3306/tcp") // MySQL
	_, _ = f.runQuiet("ufw", "deny", "6379/tcp") // Redis

	// Protect phpMyAdmin port (e.g. 8080)
	if pmaPort <= 0 {
		pmaPort = 8080
	}

	// Remove older blanket allow for PMA port if exists
	_, _ = f.runQuiet("ufw", "delete", "allow", fmt.Sprintf("%d/tcp", pmaPort))

	// Allow specific admin IPs if configured
	hasAllowedIP := false
	for _, ip := range allowedAdminIPs {
		ip = strings.TrimSpace(ip)
		if ip != "" && ip != "127.0.0.1" && ip != "localhost" {
			_, _ = f.runQuiet("ufw", "allow", "from", ip, "to", "any", "port", fmt.Sprintf("%d", pmaPort), "proto", "tcp")
			hasAllowedIP = true
		}
	}

	// If no specific IP allowed, block PMA port from public ingress entirely
	if !hasAllowedIP {
		_, _ = f.runQuiet("ufw", "deny", fmt.Sprintf("%d/tcp", pmaPort))
	}

	// Enable UFW non-interactively
	cmd := exec.Command("ufw", "--force", "enable")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable ufw: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// ApplyLocalProfile sets a permissive developer profile for local Linux desktop/WSL
func (f *Controller) ApplyLocalProfile(pmaPort int) error {
	if !f.IsInstalled() {
		// On local machines, UFW not being installed is normal and safe
		return nil
	}

	if !installer.IsRoot() {
		return fmt.Errorf("root privileges required to configure firewall")
	}

	if pmaPort <= 0 {
		pmaPort = 8080
	}

	// Allow web, PMA, and database on local loopback and subnet
	_, _ = f.runQuiet("ufw", "allow", "80/tcp")
	_, _ = f.runQuiet("ufw", "allow", "443/tcp")
	_, _ = f.runQuiet("ufw", "allow", fmt.Sprintf("%d/tcp", pmaPort))
	_, _ = f.runQuiet("ufw", "allow", "3306/tcp")
	_, _ = f.runQuiet("ufw", "allow", "6379/tcp")

	return nil
}

// AllowPort opens a specific TCP port with a label
func (f *Controller) AllowPort(port int, comment string) error {
	args := []string{"allow", fmt.Sprintf("%d/tcp", port)}
	if comment != "" {
		args = append(args, "comment", comment)
	}
	_, err := f.runQuiet("ufw", args...)
	return err
}

// AllowPortForIP allows access to a specific port only from a specified client IP
func (f *Controller) AllowPortForIP(port int, clientIP string) error {
	_, err := f.runQuiet("ufw", "allow", "from", clientIP, "to", "any", "port", fmt.Sprintf("%d", port), "proto", "tcp")
	return err
}

// DenyPort blocks ingress to a specific port
func (f *Controller) DenyPort(port int) error {
	_, err := f.runQuiet("ufw", "deny", fmt.Sprintf("%d/tcp", port))
	return err
}

// GetStatusDetails returns human-readable status and rule summary
func (f *Controller) GetStatusDetails() (bool, string, error) {
	if !f.IsInstalled() {
		return false, "UFW is not installed (no active firewall)", nil
	}
	out, err := f.runQuiet("ufw", "status", "numbered")
	if err != nil {
		return false, "", err
	}
	isActive := strings.Contains(strings.ToLower(out), "status: active")
	return isActive, strings.TrimSpace(out), nil
}

// ProtectPHPMyAdmin applies firewall protection according to the active environment mode
func (f *Controller) ProtectPHPMyAdmin(pmaPort int) error {
	cfg := config.GetConfig()
	if cfg.Environment == config.ModeProduction {
		return f.ApplyProductionProfile(pmaPort, cfg.AllowedAdminIPs)
	}
	return f.ApplyLocalProfile(pmaPort)
}

func (f *Controller) runQuiet(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
