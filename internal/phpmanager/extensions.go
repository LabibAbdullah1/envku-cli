package phpmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtensionInfo represents the state of a PHP extension
type ExtensionInfo struct {
	Name        string `json:"name"`
	EnabledFPM  bool   `json:"enabled_fpm"`
	EnabledCLI  bool   `json:"enabled_cli"`
	Description string `json:"description"`
}

// CommonExtensions contains descriptions for frequently used PHP extensions
var CommonExtensions = map[string]string{
	"pdo_mysql": "MySQL database PDO driver",
	"mbstring":  "Multibyte string support for international text",
	"curl":      "HTTP client and URL data transfer library",
	"gd":        "Image processing and manipulation library",
	"zip":       "Zip archive compression and extraction",
	"bcmath":    "Arbitrary precision mathematics calculation",
	"intl":      "Internationalization and localization functions",
	"redis":     "Fast in-memory Redis database client",
	"xml":       "XML document parsing and DOM manipulation",
	"soap":      "SOAP protocol web services support",
	"imagick":   "ImageMagick advanced image processing",
	"opcache":   "Bytecode caching and execution accelerator",
	"sqlite3":   "Embedded SQLite3 database support",
}

// ListExtensions discovers available extensions and their active status
func (m *Manager) ListExtensions() ([]ExtensionInfo, error) {
	modsDir := filepath.Join("/etc/php", m.Version, "mods-available")
	fpmConfDir := filepath.Join("/etc/php", m.Version, "fpm", "conf.d")
	cliConfDir := filepath.Join("/etc/php", m.Version, "cli", "conf.d")

	var list []ExtensionInfo

	// Read available modules directory
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		// If directory cannot be read, populate from CommonExtensions
		for name, desc := range CommonExtensions {
			list = append(list, ExtensionInfo{
				Name:        name,
				EnabledFPM:  m.isExtensionEnabled(fpmConfDir, name),
				EnabledCLI:  m.isExtensionEnabled(cliConfDir, name),
				Description: desc,
			})
		}
		return list, nil
	}

	seen := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ini") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".ini")
		desc := CommonExtensions[name]
		if desc == "" {
			desc = fmt.Sprintf("PHP %s extension module", name)
		}
		list = append(list, ExtensionInfo{
			Name:        name,
			EnabledFPM:  m.isExtensionEnabled(fpmConfDir, name),
			EnabledCLI:  m.isExtensionEnabled(cliConfDir, name),
			Description: desc,
		})
		seen[name] = true
	}

	// Add any common extensions not yet seen
	for name, desc := range CommonExtensions {
		if !seen[name] {
			list = append(list, ExtensionInfo{
				Name:        name,
				EnabledFPM:  m.isExtensionEnabled(fpmConfDir, name),
				EnabledCLI:  m.isExtensionEnabled(cliConfDir, name),
				Description: desc,
			})
		}
	}

	return list, nil
}

// isExtensionEnabled checks if a symlink exists for the extension in conf.d
func (m *Manager) isExtensionEnabled(confDir, extName string) bool {
	matches, err := filepath.Glob(filepath.Join(confDir, fmt.Sprintf("*-%s.ini", extName)))
	if err == nil && len(matches) > 0 {
		return true
	}
	matchesDirect, err := filepath.Glob(filepath.Join(confDir, fmt.Sprintf("%s.ini", extName)))
	return err == nil && len(matchesDirect) > 0
}

// ToggleExtension enables or disables an extension using phpenmod/phpdismod
func (m *Manager) ToggleExtension(ext string, enable bool) error {
	action := "phpenmod"
	if !enable {
		action = "phpdismod"
	}

	cmd := exec.Command(action, "-v", m.Version, ext)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to %s extension %s: %s", action, ext, strings.TrimSpace(string(out)))
	}

	// Validate configuration syntax
	if err := m.ValidateFPM(); err != nil {
		// Revert action if validation fails
		revertAction := "phpdismod"
		if !enable {
			revertAction = "phpenmod"
		}
		_ = exec.Command(revertAction, "-v", m.Version, ext).Run()
		return fmt.Errorf("validation failed after toggling extension %s, reverted: %w", ext, err)
	}

	// Reload FPM to apply changes
	if err := m.ReloadFPM(); err != nil {
		return fmt.Errorf("extension toggled but failed to reload PHP-FPM: %w", err)
	}

	return nil
}
