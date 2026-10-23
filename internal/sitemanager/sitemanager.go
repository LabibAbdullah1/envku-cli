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

	// Optional SSL setup via Certbot
	if cfg.EnableSSL {
		if err := m.RequestSSL(cfg.Domain, cfg.SSLEmail); err != nil {
			return fmt.Errorf("site created, but SSL setup failed: %w", err)
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

// DeleteSite completely removes site configuration files
func (m *Manager) DeleteSite(domain string) error {
	_ = m.DisableSite(domain)
	availFile := filepath.Join(config.NginxSitesAvail, domain+".conf")
	if err := os.Remove(availFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete configuration file %s: %w", availFile, err)
	}
	return m.ReloadNginx()
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
		})
	}

	return sites, nil
}
