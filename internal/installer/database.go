package installer

import (
	"fmt"
	"regexp"
	"strings"
)

// MySQLInstaller manages MySQL server installation and status
type MySQLInstaller struct{}

func NewMySQLInstaller() *MySQLInstaller {
	return &MySQLInstaller{}
}

// Status returns installation and service state of MySQL
func (m *MySQLInstaller) Status() PackageStatus {
	installed := IsBinaryInstalled("mysql")
	active := false
	version := ""

	if installed {
		active = IsServiceActive("mysql")
		version = m.GetVersion()
	}

	return PackageStatus{
		Name:      "mysql-server",
		Installed: installed,
		Active:    active,
		Version:   version,
	}
}

// GetVersion extracts the installed MySQL version
func (m *MySQLInstaller) GetVersion() string {
	out, err := RunCommandOutput("mysql", "--version")
	if err != nil {
		return ""
	}
	// Output: mysql  Ver 8.0.39-0ubuntu0.24.04.1 for Linux on x86_64
	re := regexp.MustCompile(`Ver\s+([0-9.]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(out)
}

// Install installs MySQL Server and starts the service
func (m *MySQLInstaller) Install() error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to install MySQL")
	}

	if err := AptUpdate(); err != nil {
		return fmt.Errorf("failed to update apt repository: %w", err)
	}

	if err := AptInstall("mysql-server"); err != nil {
		return fmt.Errorf("failed to install mysql-server: %w", err)
	}

	if err := EnableAndStartService("mysql"); err != nil {
		return fmt.Errorf("failed to start mysql service: %w", err)
	}

	return nil
}
