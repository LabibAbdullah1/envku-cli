package servicectl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// LogSource specifies the origin of the log data
type LogSource string

const (
	LogSourceJournal LogSource = "journal"
	LogSourceFile    LogSource = "file"
)

// GetJournalLogs fetches the most recent N lines of journalctl for a unit
func (c *Controller) GetJournalLogs(service string, lines int) ([]string, error) {
	if lines <= 0 {
		lines = 50
	}

	unit := c.normalizeUnit(service)
	cmd := exec.Command("journalctl", "-u", unit, "-n", strconv.Itoa(lines), "--no-pager")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("journalctl failed for %s: %w", unit, err)
	}

	rawLines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	var cleaned []string
	for _, l := range rawLines {
		if strings.TrimSpace(l) != "" {
			cleaned = append(cleaned, l)
		}
	}
	return cleaned, nil
}

// KnownLogPaths returns common direct log file paths for key stack components
func KnownLogPaths() map[string][]string {
	return map[string][]string{
		"nginx": {
			"/var/log/nginx/error.log",
			"/var/log/nginx/access.log",
		},
		"mysql": {
			"/var/log/mysql/error.log",
		},
		"redis-server": {
			"/var/log/redis/redis-server.log",
		},
		"php8.2-fpm": {
			"/var/log/php8.2-fpm.log",
		},
		"php8.3-fpm": {
			"/var/log/php8.3-fpm.log",
		},
		"php8.4-fpm": {
			"/var/log/php8.4-fpm.log",
		},
		"php8.5-fpm": {
			"/var/log/php8.5-fpm.log",
		},
	}
}

// TailFile reads the last N lines from a physical log file
func TailFile(filePath string, maxLines int) ([]string, error) {
	if maxLines <= 0 {
		maxLines = 50
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file %s: %w", filePath, err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	if len(allLines) <= maxLines {
		return allLines, nil
	}

	return allLines[len(allLines)-maxLines:], nil
}
