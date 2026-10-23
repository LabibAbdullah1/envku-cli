package phpmanager

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// INIConfig represents the most commonly tuned php.ini parameters
type INIConfig struct {
	MemoryLimit       string `json:"memory_limit"`        // e.g. "512M"
	UploadMaxFilesize string `json:"upload_max_filesize"` // e.g. "64M"
	PostMaxSize       string `json:"post_max_size"`       // e.g. "64M"
	MaxExecutionTime  string `json:"max_execution_time"`  // e.g. "300"
	MaxInputVars      string `json:"max_input_vars"`      // e.g. "3000"
	Timezone          string `json:"timezone"`            // e.g. "UTC" or "Asia/Jakarta"
}

// ReadINIConfig parses the current active directives from php.ini
func (m *Manager) ReadINIConfig() (*INIConfig, error) {
	filePath := m.FPMIniPath()
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open php.ini at %s: %w", filePath, err)
	}
	defer file.Close()

	cfg := &INIConfig{
		MemoryLimit:       "128M",
		UploadMaxFilesize: "2M",
		PostMaxSize:       "8M",
		MaxExecutionTime:  "30",
		MaxInputVars:      "1000",
		Timezone:          "UTC",
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove inline comments
		if idx := strings.Index(val, ";"); idx != -1 {
			val = strings.TrimSpace(val[:idx])
		}
		val = strings.Trim(val, `"`)

		switch key {
		case "memory_limit":
			cfg.MemoryLimit = val
		case "upload_max_filesize":
			cfg.UploadMaxFilesize = val
		case "post_max_size":
			cfg.PostMaxSize = val
		case "max_execution_time":
			cfg.MaxExecutionTime = val
		case "max_input_vars":
			cfg.MaxInputVars = val
		case "date.timezone":
			cfg.Timezone = val
		}
	}

	return cfg, scanner.Err()
}

// UpdateINIConfig safely updates directives in php.ini with automatic backup and rollback
func (m *Manager) UpdateINIConfig(cfg INIConfig) error {
	backupPath, err := m.BackupINI()
	if err != nil {
		return fmt.Errorf("pre-modification backup failed: %w", err)
	}

	filePath := m.FPMIniPath()
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	targets := map[string]string{
		"memory_limit":        cfg.MemoryLimit,
		"upload_max_filesize": cfg.UploadMaxFilesize,
		"post_max_size":       cfg.PostMaxSize,
		"max_execution_time":  cfg.MaxExecutionTime,
		"max_input_vars":      cfg.MaxInputVars,
		"date.timezone":       cfg.Timezone,
	}

	found := make(map[string]bool)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for key, newVal := range targets {
			if newVal == "" {
				continue
			}

			// Match active directive e.g. "memory_limit = 128M" or commented default ";memory_limit = ..."
			pattern := fmt.Sprintf(`^\s*;?\s*%s\s*=.*$`, regexp.QuoteMeta(key))
			matched, _ := regexp.MatchString(pattern, trimmed)
			if matched && !found[key] {
				lines[i] = fmt.Sprintf("%s = %s", key, newVal)
				found[key] = true
				break
			}
		}
	}

	// Append any directives not found in the original file
	for key, newVal := range targets {
		if newVal != "" && !found[key] {
			lines = append(lines, fmt.Sprintf("%s = %s", key, newVal))
		}
	}

	// Write modified content
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		_ = m.RestoreINI(backupPath)
		return fmt.Errorf("failed to write updated php.ini: %w", err)
	}

	// Validate syntax
	if err := m.ValidateFPM(); err != nil {
		// Auto-rollback on validation failure
		_ = m.RestoreINI(backupPath)
		return fmt.Errorf("validation failed, configuration rolled back: %w", err)
	}

	// Safely reload service
	if err := m.ReloadFPM(); err != nil {
		return fmt.Errorf("configuration applied but service reload failed: %w", err)
	}

	return nil
}
