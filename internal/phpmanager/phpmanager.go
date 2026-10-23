package phpmanager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles PHP-FPM configuration, extensions, and service operations
type Manager struct {
	Version string // e.g. "8.2", "8.3", "8.4", "8.5"
}

// New creates a new Manager instance for a specific PHP version
func New(version string) *Manager {
	return &Manager{Version: version}
}

// FPMIniPath returns the path to php.ini for PHP-FPM
func (m *Manager) FPMIniPath() string {
	return filepath.Join("/etc/php", m.Version, "fpm", "php.ini")
}

// CLIIniPath returns the path to php.ini for PHP CLI
func (m *Manager) CLIIniPath() string {
	return filepath.Join("/etc/php", m.Version, "cli", "php.ini")
}

// ServiceName returns the systemd service name for this PHP version
func (m *Manager) ServiceName() string {
	return fmt.Sprintf("php%s-fpm", m.Version)
}

// FPMBinary returns the binary name used for testing configuration
func (m *Manager) FPMBinary() string {
	return fmt.Sprintf("php-fpm%s", m.Version)
}

// BackupINI creates a timestamped backup of the current FPM php.ini file
func (m *Manager) BackupINI() (string, error) {
	srcPath := m.FPMIniPath()
	if _, err := os.Stat(srcPath); err != nil {
		return "", fmt.Errorf("php.ini not found at %s: %w", srcPath, err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.bak.%s", srcPath, timestamp)

	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy php.ini to backup: %w", err)
	}

	return backupPath, nil
}

// RestoreINI restores the php.ini from a given backup path
func (m *Manager) RestoreINI(backupPath string) error {
	dstPath := m.FPMIniPath()

	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file %s: %w", backupPath, err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to php.ini: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to restore php.ini: %w", err)
	}

	return nil
}

// ValidateFPM runs `php-fpm<version> -t` to check syntax integrity
func (m *Manager) ValidateFPM() error {
	bin := m.FPMBinary()
	cmd := exec.Command(bin, "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PHP-FPM syntax validation failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// ReloadFPM validates configuration syntax first, then reloads the php-fpm systemd service
func (m *Manager) ReloadFPM() error {
	if err := m.ValidateFPM(); err != nil {
		return fmt.Errorf("cannot reload service, validation error: %w", err)
	}

	cmd := exec.Command("systemctl", "reload", m.ServiceName())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload %s: %s", m.ServiceName(), strings.TrimSpace(string(out)))
	}
	return nil
}
