package installer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DistroInfo contains detected Linux distribution metadata
type DistroInfo struct {
	ID        string
	VersionID string
	Name      string
	Codename  string
}

// DetectOS parses /etc/os-release to detect Linux distribution details
func DetectOS() (*DistroInfo, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	info := &DistroInfo{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ID=") {
			info.ID = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			info.VersionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		} else if strings.HasPrefix(line, "NAME=") {
			info.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
		} else if strings.HasPrefix(line, "VERSION_CODENAME=") {
			info.Codename = strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), `"`)
		} else if strings.HasPrefix(line, "UBUNTU_CODENAME=") && info.Codename == "" {
			info.Codename = strings.Trim(strings.TrimPrefix(line, "UBUNTU_CODENAME="), `"`)
		}
	}

	return info, scanner.Err()
}

// RunCommand executes a command with non-interactive flags set for debian/ubuntu
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCommandOutput executes a command and returns trimmed stdout as string
func RunCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// AptUpdate executes apt-get update safely
func AptUpdate() error {
	return RunCommand("apt-get", "update", "-y")
}

// AptInstall installs one or more packages non-interactively
func AptInstall(packages ...string) error {
	args := append([]string{"install", "-y", "--no-install-recommends"}, packages...)
	return RunCommand("apt-get", args...)
}

// IsServiceActive checks if a systemd unit is active
func IsServiceActive(service string) bool {
	out, err := RunCommandOutput("systemctl", "is-active", service)
	return err == nil && out == "active"
}

// EnableAndStartService enables and starts a systemd unit
func EnableAndStartService(service string) error {
	return RunCommand("systemctl", "enable", "--now", service)
}

// IsBinaryInstalled checks if an executable exists in PATH
func IsBinaryInstalled(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// EnsurePPA ensures software-properties-common is present and adds a PPA (Ubuntu)
func EnsurePPA(ppa string) error {
	if !IsBinaryInstalled("add-apt-repository") {
		if err := AptInstall("software-properties-common"); err != nil {
			return fmt.Errorf("failed to install software-properties-common: %w", err)
		}
	}
	if err := RunCommand("add-apt-repository", "-y", ppa); err != nil {
		return fmt.Errorf("failed to add repository %s: %w", ppa, err)
	}
	return AptUpdate()
}
