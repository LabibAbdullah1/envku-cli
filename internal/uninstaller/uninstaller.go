package uninstaller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/installer"
)

// Uninstaller coordinates safe or complete purge uninstallations
type Uninstaller struct{}

// New creates a new Uninstaller instance
func New() *Uninstaller {
	return &Uninstaller{}
}

// UninstallSafe removes only Envku CLI and its background daemon.
// All user web files (/var/www), MySQL databases, Nginx/PHP configs, and SSL certs remain 100% intact.
func (u *Uninstaller) UninstallSafe() error {
	if !installer.IsRoot() {
		return fmt.Errorf("hak akses root (sudo) diperlukan untuk menghapus Envku CLI")
	}

	fmt.Println("[*] Menghentikan & menghapus service background Envku CLI Systemd...")
	_ = exec.Command("systemctl", "stop", config.ServiceName).Run()
	_ = exec.Command("systemctl", "disable", config.ServiceName).Run()
	_ = exec.Command("systemctl", "stop", "envku").Run()
	_ = exec.Command("systemctl", "disable", "envku").Run()
	_ = os.Remove(config.ServiceFile)
	_ = os.Remove("/etc/systemd/system/envku.service")
	_ = exec.Command("systemctl", "daemon-reload").Run()

	fmt.Println("[*] Menghapus file konfigurasi internal Envku CLI (/etc/envku-cli)...")
	_ = os.RemoveAll(config.DefaultConfigDir)

	fmt.Println("[*] Menghapus binary Envku CLI...")
	binPath, err := os.Executable()
	if err != nil {
		binPath = config.DefaultBinPath
	}
	if evaluated, errEval := filepath.EvalSymlinks(binPath); errEval == nil {
		binPath = evaluated
	}

	_ = os.Remove(binPath)
	_ = os.Remove(config.DefaultBinPath)
	_ = os.Remove(config.SymlinkBinPath)

	return nil
}

// PurgeOptions specifies what additional components to remove in Purge mode
type PurgeOptions struct {
	RemoveWebData     bool // Remove /var/www/
	RemoveDatabases   bool // Remove /var/lib/mysql/
	RemovePackages    bool // Apt purge nginx, mysql, php, redis, phpmyadmin
	RemoveHostsConfig bool // Clean local hosts entries
}

// UninstallPurge completely wipes Envku CLI, web servers, databases, and configuration files
func (u *Uninstaller) UninstallPurge(opts PurgeOptions) error {
	if !installer.IsRoot() {
		return fmt.Errorf("hak akses root (sudo) diperlukan untuk melakukan purge")
	}

	// 1. Remove Envku CLI binary & daemon first
	if err := u.UninstallSafe(); err != nil {
		return err
	}

	// 2. Stop all web stack services
	fmt.Println("[*] Menghentikan service web stack (Nginx, MySQL, Redis, PHP-FPM)...")
	services := []string{"nginx", "mysql", "redis-server", "php8.2-fpm", "php8.3-fpm", "php8.4-fpm", "php8.5-fpm"}
	for _, s := range services {
		_ = exec.Command("systemctl", "stop", s).Run()
		_ = exec.Command("systemctl", "disable", s).Run()
	}

	// 3. Remove phpMyAdmin installation
	fmt.Println("[*] Menghapus instalasi phpMyAdmin & konfigurasi Nginx-nya...")
	_ = os.Remove(config.PMANginxEnable)
	_ = os.Remove(config.PMANginxConf)
	_ = os.RemoveAll(config.PMADir)

	// 4. Optionally purge APT packages
	if opts.RemovePackages {
		fmt.Println("[*] Melakukan apt purge pada paket Nginx, MySQL, PHP, Redis, dan Certbot...")
		packages := []string{
			"nginx", "nginx-common", "nginx-core",
			"mysql-server", "mysql-client", "mysql-common",
			"redis-server",
			"certbot", "python3-certbot-nginx",
		}
		// Add php packages
		for _, ver := range installer.SupportedPHPVersions {
			packages = append(packages, fmt.Sprintf("php%s*", ver))
		}

		args := append([]string{"purge", "-y", "--auto-remove"}, packages...)
		cmd := exec.Command("apt-get", args...)
		cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
		_ = cmd.Run()
	}

	// 5. Optionally remove web files (/var/www)
	if opts.RemoveWebData {
		fmt.Println("[*] Menghapus seluruh direktori data web (/var/www)...")
		_ = os.RemoveAll("/var/www")
	}

	// 6. Optionally remove database storage (/var/lib/mysql)
	if opts.RemoveDatabases {
		fmt.Println("[*] Menghapus seluruh data database MySQL (/var/lib/mysql)...")
		_ = os.RemoveAll("/var/lib/mysql")
		_ = os.RemoveAll("/etc/mysql")
	}

	// 7. Clean Nginx and PHP config directories
	_ = os.RemoveAll("/etc/nginx")
	_ = os.RemoveAll("/etc/php")

	return nil
}
