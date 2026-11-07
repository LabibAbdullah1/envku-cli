package installer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
)

// PHPMyAdminInstaller manages the lifecycle of phpMyAdmin
type PHPMyAdminInstaller struct{}

// NewPHPMyAdminInstaller creates a new phpMyAdmin installer
func NewPHPMyAdminInstaller() *PHPMyAdminInstaller {
	return &PHPMyAdminInstaller{}
}

// Status returns current installation and access status of phpMyAdmin
func (p *PHPMyAdminInstaller) Status() PackageStatus {
	installed := p.IsInstalled()
	active := false
	version := ""

	if installed {
		active = IsServiceActive("nginx") && p.isSiteEnabled()
		version = p.GetVersion()
	}

	return PackageStatus{
		Name:      "phpmyadmin",
		Installed: installed,
		Active:    active,
		Version:   version,
	}
}

// IsInstalled checks if phpMyAdmin files and configuration exist
func (p *PHPMyAdminInstaller) IsInstalled() bool {
	indexPath := filepath.Join(config.PMADir, "index.php")
	if _, err := os.Stat(indexPath); err == nil {
		return true
	}
	// Also check fallback /var/www/phpmyadmin/index.php
	if _, err := os.Stat("/var/www/phpmyadmin/index.php"); err == nil {
		return true
	}
	return false
}

func (p *PHPMyAdminInstaller) isSiteEnabled() bool {
	_, err := os.Lstat(config.PMANginxEnable)
	return err == nil
}

// GetVersion extracts phpMyAdmin version from RELEASE-DATE or README file
func (p *PHPMyAdminInstaller) GetVersion() string {
	versionFile := filepath.Join(config.PMADir, "RELEASE-DATE")
	if data, err := os.ReadFile(versionFile); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			return strings.TrimSpace(lines[0])
		}
	}
	return "5.2"
}

// DetectActivePHPSock finds an active PHP-FPM socket in /run/php/
func (p *PHPMyAdminInstaller) DetectActivePHPSock() (string, string) {
	for _, ver := range []string{"8.4", "8.3", "8.2", "8.1", "8.0"} {
		sock := fmt.Sprintf("/run/php/php%s-fpm.sock", ver)
		if _, err := os.Stat(sock); err == nil {
			return sock, ver
		}
	}
	// Fallback check pattern
	matches, err := filepath.Glob("/run/php/php*-fpm.sock")
	if err == nil && len(matches) > 0 {
		return matches[0], "8.3"
	}
	return "/run/php/php8.3-fpm.sock", "8.3"
}

// Install downloads, configures, and activates phpMyAdmin with Nginx
func (p *PHPMyAdminInstaller) Install() error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to install phpMyAdmin")
	}

	cfg := config.GetConfig()

	// 1. Ensure Nginx and PHP are ready
	if !IsBinaryInstalled("nginx") {
		return fmt.Errorf("nginx is required before installing phpMyAdmin")
	}

	phpSock, _ := p.DetectActivePHPSock()

	// 2. Install essential PHP extensions for phpMyAdmin if missing
	_ = AptInstall("php-mbstring", "php-zip", "php-gd", "php-xml", "php-mysql", "curl", "tar")

	// 3. In Local Mode: ensure MySQL root is configured with empty password for instant login
	if cfg.Environment == config.ModeLocal && IsBinaryInstalled("mysql") {
		mysqlInst := NewMySQLInstaller()
		fmt.Println("[*] Mode Local: Mengonfigurasi MySQL root tanpa password...")
		if err := mysqlInst.ConfigureLocalRootNoPassword(); err != nil {
			fmt.Printf("[!] Catatan: Konfigurasi root tanpa password: %v\n", err)
		} else {
			fmt.Println("[+] MySQL root dikonfigurasi tanpa password.")
		}
	}

	// 4. Download and extract phpMyAdmin
	targetDir := config.PMADir
	_ = os.MkdirAll(targetDir, 0755)

	tarFile := "/tmp/phpmyadmin.tar.gz"
	downloadURL := "https://files.phpmyadmin.net/phpMyAdmin/5.2.1/phpMyAdmin-5.2.1-all-languages.tar.gz"

	fmt.Println("[*] Mengunduh phpMyAdmin dari portal resmi...")
	cmdDownload := exec.Command("curl", "-fsSL", downloadURL, "-o", tarFile)
	if err := cmdDownload.Run(); err != nil {
		// Fallback to secondary download mirror if primary fails
		altURL := "https://www.phpmyadmin.net/downloads/phpMyAdmin-latest-all-languages.tar.gz"
		if errAlt := exec.Command("curl", "-fsSL", altURL, "-o", tarFile).Run(); errAlt != nil {
			return fmt.Errorf("failed to download phpMyAdmin: %w", errAlt)
		}
	}

	// Extract tarball
	tmpExtract := "/tmp/pma-extract"
	_ = os.RemoveAll(tmpExtract)
	_ = os.MkdirAll(tmpExtract, 0755)

	cmdExtract := exec.Command("tar", "-xzf", tarFile, "-C", tmpExtract, "--strip-components=1")
	if err := cmdExtract.Run(); err != nil {
		return fmt.Errorf("failed to extract phpMyAdmin archive: %w", err)
	}

	// Move into target directory
	_ = os.RemoveAll(targetDir)
	_ = os.MkdirAll(targetDir, 0755)
	if err := exec.Command("cp", "-r", tmpExtract+"/.", targetDir+"/").Run(); err != nil {
		return fmt.Errorf("failed to copy phpMyAdmin to %s: %w", targetDir, err)
	}

	// Clean temporary files
	_ = os.Remove(tarFile)
	_ = os.RemoveAll(tmpExtract)

	// 5. Generate Blowfish secret & config.inc.php
	secretBytes := make([]byte, 16)
	_, _ = rand.Read(secretBytes)
	blowfishSecret := hex.EncodeToString(secretBytes)

	allowNoPassword := "false"
	extraLocalCfg := ""
	if cfg.Environment == config.ModeLocal {
		allowNoPassword = "true"
		extraLocalCfg = "$cfg['Servers'][$i]['user'] = 'root';\n$cfg['Servers'][$i]['password'] = '';\n"
	}

	configInc := fmt.Sprintf(`<?php
/**
 * phpMyAdmin configuration generated by Envku CLI
 * Mode Environment: %s
 */
$cfg['blowfish_secret'] = '%s';
$i = 0;
$i++;
$cfg['Servers'][$i]['auth_type'] = 'cookie';
$cfg['Servers'][$i]['host'] = '127.0.0.1';
$cfg['Servers'][$i]['compress'] = false;
$cfg['Servers'][$i]['AllowNoPassword'] = %s;
%s$cfg['UploadDir'] = '';
$cfg['SaveDir'] = '';
$cfg['TempDir'] = '/tmp';
`, strings.ToUpper(string(cfg.Environment)), blowfishSecret, allowNoPassword, extraLocalCfg)

	_ = os.WriteFile(filepath.Join(targetDir, "config.inc.php"), []byte(configInc), 0644)

	// Fix directory permissions for www-data
	_ = exec.Command("chown", "-R", "www-data:www-data", targetDir).Run()
	_ = exec.Command("chmod", "-R", "755", targetDir).Run()

	// 6. Generate and install Nginx server block configuration
	pmaPort := cfg.PMAPort
	if pmaPort <= 0 {
		pmaPort = 8080
	}

	nginxConf := p.GenerateNginxConfig(pmaPort, phpSock, cfg)

	_ = os.MkdirAll(config.NginxSitesAvail, 0755)
	_ = os.MkdirAll(config.NginxSitesEnable, 0755)

	if err := os.WriteFile(config.PMANginxConf, []byte(nginxConf), 0644); err != nil {
		return fmt.Errorf("failed to write phpmyadmin nginx config: %w", err)
	}

	_ = os.Remove(config.PMANginxEnable)
	if err := os.Symlink(config.PMANginxConf, config.PMANginxEnable); err != nil {
		return fmt.Errorf("failed to symlink phpmyadmin nginx site: %w", err)
	}

	// Validate Nginx configuration
	cmdTest := exec.Command("nginx", "-t")
	if out, err := cmdTest.CombinedOutput(); err != nil {
		_ = os.Remove(config.PMANginxEnable)
		return fmt.Errorf("nginx syntax error for phpMyAdmin: %s", strings.TrimSpace(string(out)))
	}

	// Reload Nginx
	_ = exec.Command("systemctl", "reload", "nginx").Run()

	return nil
}

// GenerateNginxConfig creates the Nginx configuration block with security rules
func (p *PHPMyAdminInstaller) GenerateNginxConfig(port int, phpSock string, cfg *config.Config) string {
	ipRestriction := ""
	if cfg.Environment == config.ModeProduction && cfg.PMARestrictProd {
		var lines []string
		lines = append(lines, "    # [PRODUCTION SECURITY] phpMyAdmin dibatasi oleh Firewall & IP Whitelist")
		lines = append(lines, "    allow 127.0.0.1;")
		lines = append(lines, "    allow ::1;")
		for _, ip := range cfg.AllowedAdminIPs {
			ip = strings.TrimSpace(ip)
			if ip != "" && ip != "127.0.0.1" && ip != "::1" {
				lines = append(lines, fmt.Sprintf("    allow %s;", ip))
			}
		}
		lines = append(lines, "    deny all;")
		ipRestriction = strings.Join(lines, "\n") + "\n"
	}

	return fmt.Sprintf(`# phpMyAdmin configuration generated by Envku CLI
# Mode Environment: %s
server {
    listen %d;
    listen [::]:%d;
    server_name _;
    root %s;
    index index.php index.html index.htm;

    client_max_body_size 128M;

%s
    location / {
        try_files $uri $uri/ =404;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        fastcgi_pass unix:%s;
        fastcgi_index index.php;
        fastcgi_read_timeout 300;
        fastcgi_hide_header X-Powered-By;
    }

    location ~ /(libraries|setup|templates|locale)/ {
        deny all;
    }

    location ~ /\. {
        deny all;
    }
}
`, strings.ToUpper(string(cfg.Environment)), port, port, config.PMADir, ipRestriction, phpSock)
}

// GetAccessURL returns the recommended URL to access phpMyAdmin
func (p *PHPMyAdminInstaller) GetAccessURL() string {
	cfg := config.GetConfig()
	port := cfg.PMAPort
	if port <= 0 {
		port = 8080
	}

	if cfg.Environment == config.ModeLocal {
		return fmt.Sprintf("http://localhost:%d", port)
	}

	// Production: use server public IP or domain
	serverIP := cfg.ServerIP
	if serverIP == "" {
		serverIP = config.DetectPublicIP()
	}
	return fmt.Sprintf("http://%s:%d", serverIP, port)
}

// OpenInBrowser attempts to open phpMyAdmin in the local browser (if supported)
func (p *PHPMyAdminInstaller) OpenInBrowser() error {
	url := p.GetAccessURL()
	return config.OpenBrowser(url)
}

// Uninstall disables and removes phpMyAdmin configurations
func (p *PHPMyAdminInstaller) Uninstall() error {
	_ = os.Remove(config.PMANginxEnable)
	_ = os.Remove(config.PMANginxConf)
	_ = os.RemoveAll(config.PMADir)
	_ = exec.Command("nginx", "-t").Run()
	_ = exec.Command("systemctl", "reload", "nginx").Run()
	return nil
}
