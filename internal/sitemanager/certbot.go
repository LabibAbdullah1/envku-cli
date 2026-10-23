package sitemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureCertbotInstalled checks if Certbot and the Nginx plugin are installed, installing if necessary
func (m *Manager) EnsureCertbotInstalled() error {
	if _, err := exec.LookPath("certbot"); err == nil {
		return nil
	}

	cmd := exec.Command("apt-get", "install", "-y", "certbot", "python3-certbot-nginx")
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install certbot & python3-certbot-nginx: %w", err)
	}
	return nil
}

// RequestSSL issues a Let's Encrypt SSL certificate using certbot --nginx
func (m *Manager) RequestSSL(domain, email string) error {
	if err := m.EnsureCertbotInstalled(); err != nil {
		return err
	}

	args := []string{
		"--nginx",
		"-d", domain,
		"--non-interactive",
		"--agree-tos",
		"--redirect",
	}

	if email != "" {
		args = append(args, "--email", email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	cmd := exec.Command("certbot", args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot failed: %s: %w", strings.TrimSpace(string(out)), err)
	}

	// Validate configuration after certbot modifies server block
	if err := m.ValidateNginx(); err != nil {
		return fmt.Errorf("nginx validation failed after SSL installation: %w", err)
	}

	return m.ReloadNginx()
}

// HasSSL checks if SSL certificates exist for the given domain
func (m *Manager) HasSSL(domain string) bool {
	certPath := filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem")
	if _, err := os.Stat(certPath); err == nil {
		return true
	}

	// Secondary check: examine server block file for ssl_certificate
	availableFile := filepath.Join("/etc/nginx/sites-available", domain+".conf")
	content, err := os.ReadFile(availableFile)
	if err == nil && strings.Contains(string(content), "ssl_certificate") {
		return true
	}

	return false
}
