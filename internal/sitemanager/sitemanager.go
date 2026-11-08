package sitemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
)

// SiteType represents the kind of application being served
type SiteType string

const (
	SiteTypePHP    SiteType = "php"
	SiteTypeNode   SiteType = "node"
	SiteTypeStatic SiteType = "static"
)

// SiteConfig represents parameters for creating a new site
type SiteConfig struct {
	Domain        string   `json:"domain"`
	Type          SiteType `json:"type"`
	DocumentRoot  string   `json:"document_root"`
	PHPVersion    string   `json:"php_version,omitempty"`
	NodeProxyPort int      `json:"node_port,omitempty"`
	EnableSSL     bool     `json:"enable_ssl"`
	SSLEmail      string   `json:"ssl_email,omitempty"`
	AddToHosts    bool     `json:"add_to_hosts,omitempty"` // Local /etc/hosts resolution
}

// SiteInfo represents discovered metadata of an existing site
type SiteInfo struct {
	Domain        string   `json:"domain"`
	Type          SiteType `json:"type"`
	DocumentRoot  string   `json:"document_root"`
	PHPVersion    string   `json:"php_version,omitempty"`
	NodeProxyPort int      `json:"node_port,omitempty"`
	Enabled       bool     `json:"enabled"`
	HasSSL        bool     `json:"has_ssl"`
	IsLocal       bool     `json:"is_local"`
}

// Manager controls Nginx Virtual Host configurations and SSL lifecycles
type Manager struct{}

// New creates a new Site Manager instance
func New() *Manager {
	return &Manager{}
}

// ValidateNginx executes `nginx -t` to test configuration syntax
func (m *Manager) ValidateNginx() error {
	cmd := exec.Command("nginx", "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx syntax test failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// ReloadNginx validates syntax and issues `systemctl reload nginx`
func (m *Manager) ReloadNginx() error {
	if err := m.ValidateNginx(); err != nil {
		return err
	}
	cmd := exec.Command("systemctl", "reload", "nginx")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload nginx: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// CreateSite generates, links, tests, and activates a new Nginx virtual host
func (m *Manager) CreateSite(cfg SiteConfig) error {
	if cfg.Domain == "" {
		return fmt.Errorf("domain name cannot be empty")
	}

	// Prepare directories
	_ = os.MkdirAll(config.NginxSitesAvail, 0755)
	_ = os.MkdirAll(config.NginxSitesEnable, 0755)

	// Create document root directory if applicable
	if cfg.Type == SiteTypePHP || cfg.Type == SiteTypeStatic {
		docRoot := cfg.DocumentRoot
		if docRoot == "" {
			if cfg.Type == SiteTypePHP {
				docRoot = fmt.Sprintf("/var/www/%s/public", cfg.Domain)
			} else {
				docRoot = fmt.Sprintf("/var/www/%s", cfg.Domain)
			}
		}
		if err := os.MkdirAll(docRoot, 0755); err != nil {
			return fmt.Errorf("failed to create document root directory %s: %w", docRoot, err)
		}

		// Ensure starter index file exists so Nginx doesn't return 403 Forbidden
		if err := EnsureStarterFile(cfg, docRoot); err != nil {
			fmt.Printf("[!] Warning: could not create starter file: %v\n", err)
		}

		// Set directory ownership to www-data and permissions 755
		baseDir := fmt.Sprintf("/var/www/%s", cfg.Domain)
		_ = exec.Command("chown", "-R", "www-data:www-data", baseDir).Run()
		_ = exec.Command("chmod", "-R", "755", baseDir).Run()
	}

	availFile := filepath.Join(config.NginxSitesAvail, cfg.Domain+".conf")
	enableFile := filepath.Join(config.NginxSitesEnable, cfg.Domain+".conf")

	// Render Nginx configuration
	nginxContent := GenerateNginxConfig(cfg)
	if err := os.WriteFile(availFile, []byte(nginxContent), 0644); err != nil {
		return fmt.Errorf("failed to write site config to %s: %w", availFile, err)
	}

	// Create symlink to enable site
	_ = os.Remove(enableFile)
	if err := os.Symlink(availFile, enableFile); err != nil {
		_ = os.Remove(availFile)
		return fmt.Errorf("failed to symlink site config: %w", err)
	}

	// Syntax validation
	if err := m.ValidateNginx(); err != nil {
		// Rollback files if invalid
		_ = os.Remove(enableFile)
		_ = os.Remove(availFile)
		return fmt.Errorf("site configuration invalid, rolled back: %w", err)
	}

	// Reload web server
	if err := m.ReloadNginx(); err != nil {
		return err
	}

	// Verify PHP-FPM socket if it's a PHP site
	if cfg.Type == SiteTypePHP {
		phpVer := cfg.PHPVersion
		if phpVer == "" {
			phpVer = "8.3"
		}
		sockPath := fmt.Sprintf("/run/php/php%s-fpm.sock", phpVer)
		if _, err := os.Stat(sockPath); os.IsNotExist(err) {
			// Try starting the service in case it's installed but stopped
			serviceName := fmt.Sprintf("php%s-fpm", phpVer)
			_ = exec.Command("systemctl", "start", serviceName).Run()
			if _, errAfter := os.Stat(sockPath); os.IsNotExist(errAfter) {
				fmt.Printf("\n[!] PERINGATAN: Socket PHP-FPM tidak ditemukan di %s\n", sockPath)
				fmt.Printf("    Layanan PHP %s-FPM belum terpasang atau belum berjalan.\n", phpVer)
				fmt.Println("    Jika situs dibuka di browser sekarang, Nginx akan menampilkan '502 Bad Gateway'.")
				fmt.Printf("    -> Solusi: Pasang PHP %s melalui Menu [2] Install & Manage Stack.\n\n", phpVer)
			}
		}
	}

	// Auto-resolve to /etc/hosts for local development if requested or if domain is local
	if cfg.AddToHosts || IsLocalDomain(cfg.Domain) {
		_ = EnsureHostsEntry(cfg.Domain)
	}

	// Optional SSL setup via Certbot (Production only)
	if cfg.EnableSSL {
		if IsLocalDomain(cfg.Domain) {
			fmt.Printf("[!] Note: SSL Let's Encrypt dilewati karena %s adalah domain lokal.\n", cfg.Domain)
		} else {
			if err := m.RequestSSL(cfg.Domain, cfg.SSLEmail); err != nil {
				return fmt.Errorf("site created, but SSL setup failed: %w", err)
			}
		}
	}

	return nil
}

// EnableSite enables a site by creating a symlink in sites-enabled
func (m *Manager) EnableSite(domain string) error {
	availFile := filepath.Join(config.NginxSitesAvail, domain+".conf")
	enableFile := filepath.Join(config.NginxSitesEnable, domain+".conf")

	if _, err := os.Stat(availFile); err != nil {
		return fmt.Errorf("configuration file for %s does not exist", domain)
	}

	_ = os.Remove(enableFile)
	if err := os.Symlink(availFile, enableFile); err != nil {
		return fmt.Errorf("failed to enable site %s: %w", domain, err)
	}

	if err := m.ValidateNginx(); err != nil {
		_ = os.Remove(enableFile)
		return fmt.Errorf("enabling site %s caused syntax errors, reverted: %w", domain, err)
	}

	return m.ReloadNginx()
}

// DisableSite disables a site by removing its symlink in sites-enabled
func (m *Manager) DisableSite(domain string) error {
	enableFile := filepath.Join(config.NginxSitesEnable, domain+".conf")
	if _, err := os.Lstat(enableFile); err == nil {
		if err := os.Remove(enableFile); err != nil {
			return fmt.Errorf("failed to disable site %s: %w", domain, err)
		}
	}

	if err := m.ValidateNginx(); err != nil {
		return err
	}
	return m.ReloadNginx()
}

// DeleteSite completely removes site configuration files and hosts entries
func (m *Manager) DeleteSite(domain string) error {
	_ = m.DisableSite(domain)
	availFile := filepath.Join(config.NginxSitesAvail, domain+".conf")
	if err := os.Remove(availFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete configuration file %s: %w", availFile, err)
	}
	_ = RemoveHostsEntry(domain)
	return m.ReloadNginx()
}

// OpenSiteInBrowser opens the site URL in the default local desktop browser
func (m *Manager) OpenSiteInBrowser(domain string) error {
	cleanDomain := strings.TrimSpace(domain)
	cleanDomain = strings.TrimPrefix(cleanDomain, "http://")
	cleanDomain = strings.TrimPrefix(cleanDomain, "https://")
	cleanDomain = strings.TrimRight(cleanDomain, "/")

	scheme := "http"
	if m.HasSSL(cleanDomain) {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s", scheme, cleanDomain)
	return config.OpenBrowser(url)
}

// ListSites discovers all sites in sites-available and their current status
func (m *Manager) ListSites() ([]SiteInfo, error) {
	entries, err := os.ReadDir(config.NginxSitesAvail)
	if err != nil {
		if os.IsNotExist(err) {
			return []SiteInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read sites-available directory: %w", err)
	}

	var sites []SiteInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}

		domain := strings.TrimSuffix(entry.Name(), ".conf")
		availPath := filepath.Join(config.NginxSitesAvail, entry.Name())
		enablePath := filepath.Join(config.NginxSitesEnable, entry.Name())

		_, errLink := os.Lstat(enablePath)
		isEnabled := (errLink == nil)

		content, errRead := os.ReadFile(availPath)
		siteType := SiteTypeStatic
		docRoot := ""
		phpVer := ""
		proxyPort := 0

		if errRead == nil {
			text := string(content)
			if strings.Contains(text, "fastcgi_pass") {
				siteType = SiteTypePHP
				rePHP := regexp.MustCompile(`php([0-9.]+)-fpm\.sock`)
				matchPHP := rePHP.FindStringSubmatch(text)
				if len(matchPHP) > 1 {
					phpVer = matchPHP[1]
				}
			} else if strings.Contains(text, "proxy_pass") {
				siteType = SiteTypeNode
				reProxy := regexp.MustCompile(`proxy_pass\s+http://127\.0\.0\.1:([0-9]+);`)
				matchProxy := reProxy.FindStringSubmatch(text)
				if len(matchProxy) > 1 {
					proxyPort, _ = strconv.Atoi(matchProxy[1])
				}
			}

			reRoot := regexp.MustCompile(`root\s+([^;]+);`)
			matchRoot := reRoot.FindStringSubmatch(text)
			if len(matchRoot) > 1 {
				docRoot = strings.TrimSpace(matchRoot[1])
			}
		}

		sites = append(sites, SiteInfo{
			Domain:        domain,
			Type:          siteType,
			DocumentRoot:  docRoot,
			PHPVersion:    phpVer,
			NodeProxyPort: proxyPort,
			Enabled:       isEnabled,
			HasSSL:        m.HasSSL(domain),
			IsLocal:       IsLocalDomain(domain),
		})
	}

	return sites, nil
}

// EnsureStarterFile generates a default welcome index file if document root is empty
func EnsureStarterFile(cfg SiteConfig, docRoot string) error {
	if cfg.Type == SiteTypePHP {
		indexPath := filepath.Join(docRoot, "index.php")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			return os.WriteFile(indexPath, []byte(defaultPHPIndex), 0644)
		}
	} else if cfg.Type == SiteTypeStatic {
		indexPath := filepath.Join(docRoot, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			return os.WriteFile(indexPath, []byte(defaultStaticIndex), 0644)
		}
	}
	return nil
}

const defaultPHPIndex = `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title><?= htmlspecialchars($_SERVER['HTTP_HOST'] ?? 'Envku App') ?> - Envku CLI</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #0b0f19;
            color: #f1f5f9;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            padding: 24px;
        }
        .container {
            background-color: #111827;
            border: 1px solid #1f2937;
            border-radius: 16px;
            padding: 40px;
            max-width: 620px;
            width: 100%;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }
        .badge {
            display: inline-block;
            background: #0284c7;
            color: white;
            font-size: 0.8rem;
            font-weight: 700;
            padding: 6px 12px;
            border-radius: 9999px;
            margin-bottom: 20px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        h1 {
            font-size: 1.75rem;
            color: #ffffff;
            margin-bottom: 12px;
        }
        p {
            color: #94a3b8;
            font-size: 0.95rem;
            line-height: 1.6;
            margin-bottom: 24px;
        }
        .info-box {
            background: #1f2937;
            border-radius: 10px;
            padding: 16px 20px;
            margin-bottom: 24px;
        }
        .row {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #374151;
            font-size: 0.9rem;
        }
        .row:last-child {
            border-bottom: none;
        }
        .row .label {
            color: #9ca3af;
        }
        .row .value {
            font-family: monospace;
            color: #38bdf8;
            font-weight: 600;
        }
        .status-ok {
            color: #22c55e !important;
        }
        .note {
            background: #0f172a;
            border-left: 4px solid #38bdf8;
            padding: 12px 16px;
            font-size: 0.85rem;
            color: #cbd5e1;
            border-radius: 0 8px 8px 0;
            line-height: 1.5;
        }
        .note code {
            color: #f59e0b;
            font-family: monospace;
        }
        .footer {
            margin-top: 28px;
            text-align: center;
            font-size: 0.8rem;
            color: #6b7280;
        }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">&#10003; Berhasil Dikonfigurasi</span>
        <h1>Selamat Datang di <?= htmlspecialchars($_SERVER['HTTP_HOST'] ?? 'Envku') ?>!</h1>
        <p>Virtual Host Nginx Anda telah berhasil dibuat dan aktif berjalan melalui <strong>Envku CLI</strong>.</p>
        
        <div class="info-box">
            <div class="row">
                <span class="label">Status Server:</span>
                <span class="value status-ok">&#9679; Online</span>
            </div>
            <div class="row">
                <span class="label">Versi PHP Aktif:</span>
                <span class="value">PHP <?= phpversion() ?></span>
            </div>
            <div class="row">
                <span class="label">Web Server:</span>
                <span class="value"><?= htmlspecialchars($_SERVER['SERVER_SOFTWARE'] ?? 'Nginx') ?></span>
            </div>
            <div class="row">
                <span class="label">Document Root:</span>
                <span class="value"><?= htmlspecialchars(__DIR__) ?></span>
            </div>
            <div class="row">
                <span class="label">Waktu Server:</span>
                <span class="value"><?= date('Y-m-d H:i:s T') ?></span>
            </div>
        </div>

        <div class="note">
            <strong>Langkah Selanjutnya:</strong> Letakkan source code project Anda (Laravel, WordPress, atau PHP kustom) di dalam folder <code><?= htmlspecialchars(__DIR__) ?></code>.
        </div>

        <div class="footer">Envku CLI &bull; Local &amp; Production Server Manager</div>
    </div>
</body>
</html>`

const defaultStaticIndex = `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Envku CLI - Static Site</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #0b0f19;
            color: #f1f5f9;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            padding: 24px;
        }
        .container {
            background-color: #111827;
            border: 1px solid #1f2937;
            border-radius: 16px;
            padding: 40px;
            max-width: 620px;
            width: 100%;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }
        .badge {
            display: inline-block;
            background: #0284c7;
            color: white;
            font-size: 0.8rem;
            font-weight: 700;
            padding: 6px 12px;
            border-radius: 9999px;
            margin-bottom: 20px;
            text-transform: uppercase;
        }
        h1 { font-size: 1.75rem; color: #ffffff; margin-bottom: 12px; }
        p { color: #94a3b8; font-size: 0.95rem; line-height: 1.6; margin-bottom: 24px; }
        .footer { margin-top: 28px; text-align: center; font-size: 0.8rem; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">&#10003; Situs Statis Aktif</span>
        <h1>Situs Statis Berhasil Dibuat</h1>
        <p>Virtual Host Nginx Anda telah siap. Silakan letakkan file HTML, CSS, JavaScript, atau build SPA React/Vue Anda di direktori ini.</p>
        <div class="footer">Envku CLI &bull; Local &amp; Production Server Manager</div>
    </div>
</body>
</html>`
