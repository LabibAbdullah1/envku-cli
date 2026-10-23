package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SystemMetrics holds basic host utilization stats
type SystemMetrics struct {
	Uptime      string
	CPUModel    string
	RAMTotalMB  uint64
	RAMUsedMB   uint64
	RAMUsedPerc float64
	DiskTotalGB uint64
	DiskUsedGB  uint64
	DiskUsedPerc float64
}

// GetSystemMetrics extracts CPU, RAM, and Uptime info from Linux /proc
func GetSystemMetrics() SystemMetrics {
	m := SystemMetrics{
		Uptime:   "N/A",
		CPUModel: "Linux Host",
	}

	// Read Uptime from /proc/uptime
	if uptimeData, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(uptimeData))
		if len(parts) > 0 {
			if secs, err := strconv.ParseFloat(parts[0], 64); err == nil {
				d := time.Duration(secs) * time.Second
				days := int(d.Hours()) / 24
				hours := int(d.Hours()) % 24
				mins := int(d.Minutes()) % 60
				if days > 0 {
					m.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
				} else {
					m.Uptime = fmt.Sprintf("%dh %dm", hours, mins)
				}
			}
		}
	}

	// Read RAM from /proc/meminfo
	if memFile, err := os.Open("/proc/meminfo"); err == nil {
		defer memFile.Close()
		scanner := bufio.NewScanner(memFile)
		var totalKB, availKB uint64
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				totalKB = parseMemKB(line)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				availKB = parseMemKB(line)
			}
		}
		if err := scanner.Err(); err == nil && totalKB > 0 {
			m.RAMTotalMB = totalKB / 1024
			usedKB := totalKB - availKB
			m.RAMUsedMB = usedKB / 1024
			m.RAMUsedPerc = (float64(usedKB) / float64(totalKB)) * 100
		}
	}

	// Read CPU Model from /proc/cpuinfo
	if cpuFile, err := os.Open("/proc/cpuinfo"); err == nil {
		defer cpuFile.Close()
		scanner := bufio.NewScanner(cpuFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					m.CPUModel = strings.TrimSpace(parts[1])
					break
				}
			}
		}
		_ = scanner.Err()
	}

	return m
}

func parseMemKB(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		return val
	}
	return 0
}
