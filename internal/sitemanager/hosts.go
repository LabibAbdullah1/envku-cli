package sitemanager

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const HostsFile = "/etc/hosts"

// EnsureHostsEntry ensures that a domain points to 127.0.0.1 in /etc/hosts for local development
func EnsureHostsEntry(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" || domain == "localhost" || domain == "127.0.0.1" {
		return nil
	}

	file, err := os.Open(HostsFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", HostsFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			for _, host := range parts[1:] {
				if strings.EqualFold(host, domain) {
					// Entry already exists
					return nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed scanning %s: %w", HostsFile, err)
	}

	// Append entry to /etc/hosts
	entry := fmt.Sprintf("\n127.0.0.1\t%s # Added by Envku CLI (Local Mode)\n", domain)
	f, err := os.OpenFile(HostsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to append to %s (requires root): %w", HostsFile, err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed writing to %s: %w", HostsFile, err)
	}

	return nil
}

// RemoveHostsEntry removes a domain from /etc/hosts if it was configured for local mode
func RemoveHostsEntry(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil
	}

	content, err := os.ReadFile(HostsFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	modified := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, domain) && (strings.HasPrefix(trimmed, "127.0.0.1") || strings.Contains(trimmed, "Envku CLI")) {
			modified = true
			continue
		}
		newLines = append(newLines, line)
	}

	if modified {
		return os.WriteFile(HostsFile, []byte(strings.Join(newLines, "\n")), 0644)
	}

	return nil
}

