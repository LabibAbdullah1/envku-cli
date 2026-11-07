package installer

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
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

	// In Local Mode: configure root with empty password for seamless local development
	if config.IsLocal() {
		fmt.Println("[*] Mode Local terdeteksi: Mengonfigurasi MySQL root tanpa password...")
		if err := m.ConfigureLocalRootNoPassword(); err != nil {
			fmt.Printf("[!] Peringatan konfigurasi root tanpa password: %v\n", err)
		} else {
			fmt.Println("[+] User root MySQL berhasil dikonfigurasi tanpa password (bebas akses di localhost/127.0.0.1).")
		}
	}

	return nil
}

// ConfigureLocalRootNoPassword configures MySQL root account to allow empty password connections from localhost and 127.0.0.1
func (m *MySQLInstaller) ConfigureLocalRootNoPassword() error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to configure MySQL user privileges")
	}

	// 1. Primary query for MySQL 8.0+ / standard auth
	query1 := `
ALTER USER 'root'@'localhost' IDENTIFIED BY '';
CREATE USER IF NOT EXISTS 'root'@'127.0.0.1' IDENTIFIED BY '';
ALTER USER 'root'@'127.0.0.1' IDENTIFIED BY '';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO 'root'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
`
	cmd1 := exec.Command("mysql", "-e", query1)
	if out, err := cmd1.CombinedOutput(); err == nil {
		return nil
	} else {
		// 2. Fallback query using mysql_native_password for compatibility
		query2 := `
ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '';
CREATE USER IF NOT EXISTS 'root'@'127.0.0.1' IDENTIFIED WITH mysql_native_password BY '';
ALTER USER 'root'@'127.0.0.1' IDENTIFIED WITH mysql_native_password BY '';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO 'root'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
`
		cmd2 := exec.Command("mysql", "-e", query2)
		if out2, err2 := cmd2.CombinedOutput(); err2 != nil {
			// 3. Fallback for MariaDB or existing password
			query3 := `
SET PASSWORD FOR 'root'@'localhost' = PASSWORD('');
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
`
			cmd3 := exec.Command("mysql", "-e", query3)
			if out3, err3 := cmd3.CombinedOutput(); err3 != nil {
				return fmt.Errorf("failed configuring root without password: %s / %s / %s",
					strings.TrimSpace(string(out)),
					strings.TrimSpace(string(out2)),
					strings.TrimSpace(string(out3)),
				)
			}
		}
	}

	return nil
}

// CanConnectRootPasswordless checks whether MySQL root can connect without a password
func (m *MySQLInstaller) CanConnectRootPasswordless() bool {
	cmd := exec.Command("mysql", "-h", "127.0.0.1", "-u", "root", "-e", "SELECT 1;")
	return cmd.Run() == nil
}
