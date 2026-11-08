package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Default path configurations and global settings for Envku CLI
var (
	AppName    = "envku-cli"
	AppVersion = "2.0.0"
)

const (
	DefaultConfigDir = "/etc/envku-cli"
	ConfigFile       = "/etc/envku-cli/config.json"
	ServiceFile      = "/etc/systemd/system/envku-cli.service"
	DefaultBinPath   = "/usr/local/bin/envku-cli"
	SymlinkBinPath   = "/usr/local/bin/envku"
	ServiceName      = "envku-cli"
	NginxSitesAvail  = "/etc/nginx/sites-available"
	NginxSitesEnable = "/etc/nginx/sites-enabled"
	PMADir           = "/usr/share/phpmyadmin"
	PMANginxConf     = "/etc/nginx/sites-available/phpmyadmin.conf"
	PMANginxEnable   = "/etc/nginx/sites-enabled/phpmyadmin.conf"
)

// EnvironmentMode defines whether Envku is operating on a Public VPS or Local Linux
type EnvironmentMode string

const (
	ModeProduction EnvironmentMode = "production"
	ModeLocal      EnvironmentMode = "local"
)

// Config represents persistent application settings
type Config struct {
	Environment        EnvironmentMode `json:"environment"`          // "production" or "local"
	ServerIP           string          `json:"server_ip"`           // Public/Detected IP
	LocalDomainSuffix  string          `json:"local_domain_suffix"` // e.g. ".test"
	PMAPort            int             `json:"pma_port"`            // Default 8080
	PMARestrictProd    bool            `json:"pma_restrict_prod"`   // In production, block PMA by firewall/allowlist
	AllowedAdminIPs    []string        `json:"allowed_admin_ips"`   // Whitelisted IPs for admin tools in production
	AutoOpenBrowser    bool            `json:"auto_open_browser"`   // In local mode, auto open browser on action
	UpdatedAt          string          `json:"updated_at"`
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
)

// DefaultConfig returns default configuration based on environment detection
func DefaultConfig() *Config {
	env := ModeProduction
	// Auto-detect if likely running in a desktop/WSL local environment
	if CanOpenBrowser() || isLocalHostname() {
		env = ModeLocal
	}

	return &Config{
		Environment:       env,
		ServerIP:          "",
		LocalDomainSuffix: ".test",
		PMAPort:           8080,
		PMARestrictProd:   true,
		AllowedAdminIPs:   []string{"127.0.0.1"},
		AutoOpenBrowser:   true,
		UpdatedAt:         time.Now().Format(time.RFC3339),
	}
}

// GetConfig returns loaded config singleton or loads it from disk
func GetConfig() *Config {
	configMutex.RLock()
	if currentConfig != nil {
		cfg := *currentConfig
		configMutex.RUnlock()
		return &cfg
	}
	configMutex.RUnlock()

	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		cfg = DefaultConfig()
	}

	configMutex.Lock()
	currentConfig = cfg
	configMutex.Unlock()

	return cfg
}

// LoadConfig reads config from /etc/envku-cli/config.json or ~/.config/envku-cli/config.json
func LoadConfig() (*Config, error) {
	path := getConfigFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	if cfg.Environment == "" {
		cfg.Environment = ModeProduction
	}
	if cfg.PMAPort <= 0 {
		cfg.PMAPort = 8080
	}
	if cfg.LocalDomainSuffix == "" {
		cfg.LocalDomainSuffix = ".test"
	}

	return &cfg, nil
}

// SaveConfig persists config to disk
func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	cfg.UpdatedAt = time.Now().Format(time.RFC3339)
	path := getConfigFilePath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	configMutex.Lock()
	currentConfig = cfg
	configMutex.Unlock()

	return nil
}

// IsProduction returns true if currently set to production VPS mode
func IsProduction() bool {
	return GetConfig().Environment == ModeProduction
}

// IsLocal returns true if currently set to local development mode
func IsLocal() bool {
	return GetConfig().Environment == ModeLocal
}

// SetEnvironment switches between production and local modes and saves to disk
func SetEnvironment(env EnvironmentMode) error {
	cfg := GetConfig()
	cfg.Environment = env
	return SaveConfig(cfg)
}

func getConfigFilePath() string {
	// If root, use /etc/envku-cli/config.json
	if os.Geteuid() == 0 {
		return ConfigFile
	}
	// Fallback for user: ~/.config/envku-cli/config.json
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".config", "envku-cli", "config.json")
	}
	return ConfigFile
}

func isLocalHostname() bool {
	host, err := os.Hostname()
	if err != nil {
		return false
	}
	host = strings.ToLower(host)
	return strings.Contains(host, "desktop") || strings.Contains(host, "laptop") || strings.Contains(host, "wsl")
}

// DetectPublicIP tries to resolve public VPS IP or primary outbound IP
func DetectPublicIP() string {
	// 1. Try public IP resolver with short timeout (useful for VPS)
	client := http.Client{Timeout: 3 * time.Second}
	endpoints := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	for _, ep := range endpoints {
		resp, err := client.Get(ep)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, rErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if rErr == nil {
				ip := strings.TrimSpace(string(body))
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}
	}

	// 2. Fallback to outbound interface IP
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 2*time.Second)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	return "127.0.0.1"
}
