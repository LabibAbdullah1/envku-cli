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
	ID             string
	IDLike         string
	VersionID      string
	Name           string
	Codename       string
	UbuntuCodename string
	DebianCodename string
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
		} else if strings.HasPrefix(line, "ID_LIKE=") {
			info.IDLike = strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			info.VersionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		} else if strings.HasPrefix(line, "NAME=") {
			info.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
		} else if strings.HasPrefix(line, "VERSION_CODENAME=") {
			info.Codename = strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), `"`)
		} else if strings.HasPrefix(line, "UBUNTU_CODENAME=") {
			info.UbuntuCodename = strings.Trim(strings.TrimPrefix(line, "UBUNTU_CODENAME="), `"`)
		} else if strings.HasPrefix(line, "DEBIAN_CODENAME=") {
			info.DebianCodename = strings.Trim(strings.TrimPrefix(line, "DEBIAN_CODENAME="), `"`)
		}
	}

	return info, scanner.Err()
}

// IsDebian returns true if the OS is pure Debian or LMDE
func (d *DistroInfo) IsDebian() bool {
	if d.ID == "debian" {
		return true
	}
	if d.ID == "linuxmint" && strings.Contains(strings.ToUpper(d.Name), "LMDE") {
		return true
	}
	if strings.Contains(d.IDLike, "debian") && !strings.Contains(d.IDLike, "ubuntu") {
		return true
	}
	return false
}

// GetEffectiveUbuntuCodename resolves the underlying Ubuntu LTS codename for Ubuntu and derivatives (Linux Mint, Pop!_OS, Zorin, etc.)
func (d *DistroInfo) GetEffectiveUbuntuCodename() string {
	if d.UbuntuCodename != "" {
		return strings.ToLower(d.UbuntuCodename)
	}

	codename := strings.ToLower(d.Codename)

	// Linux Mint codename & version mapping
	if d.ID == "linuxmint" || strings.Contains(d.IDLike, "linuxmint") {
		// Mint 20.x -> focal (Ubuntu 20.04)
		if strings.HasPrefix(d.VersionID, "20") || strings.Contains(codename, "ulyana") ||
			strings.Contains(codename, "ulyssa") || strings.Contains(codename, "uma") || strings.Contains(codename, "una") {
			return "focal"
		}
		// Mint 21.x -> jammy (Ubuntu 22.04)
		if strings.HasPrefix(d.VersionID, "21") || strings.Contains(codename, "vanessa") ||
			strings.Contains(codename, "vera") || strings.Contains(codename, "victoria") || strings.Contains(codename, "virginia") {
			return "jammy"
		}
		// Mint 22.x (zena, wilma, xia) -> noble (Ubuntu 24.04)
		return "noble"
	}

	// Version ID prefix heuristics
	if strings.HasPrefix(d.VersionID, "24") {
		return "noble"
	}
	if strings.HasPrefix(d.VersionID, "22") {
		return "jammy"
	}
	if strings.HasPrefix(d.VersionID, "20") {
		return "focal"
	}

	// If already an Ubuntu codename
	knownUbuntu := map[string]bool{
		"noble": true, "jammy": true, "focal": true, "bionic": true, "mantic": true, "lunar": true, "kinetic": true,
	}
	if knownUbuntu[codename] {
		return codename
	}

	return "noble"
}

// GetEffectiveDebianCodename resolves Debian release codename (bookworm, bullseye, buster)
func (d *DistroInfo) GetEffectiveDebianCodename() string {
	if d.DebianCodename != "" {
		return strings.ToLower(d.DebianCodename)
	}

	codename := strings.ToLower(d.Codename)
	knownDebian := map[string]bool{
		"bookworm": true, "bullseye": true, "buster": true, "trixie": true, "sid": true,
	}
	if knownDebian[codename] {
		return codename
	}

	if strings.HasPrefix(d.VersionID, "12") {
		return "bookworm"
	}
	if strings.HasPrefix(d.VersionID, "11") {
		return "bullseye"
	}

	return "bookworm"
}

// CleanBrokenSources removes faulty third-party repository lists that commonly fail on Linux Mint / Ubuntu
func CleanBrokenSources() {
	// Remove invalid sury repositories referencing Linux Mint codenames (e.g. zena, wilma, virginia)
	_ = exec.Command("bash", "-c", "rm -f /etc/apt/sources.list.d/*ondrej*.list* /etc/apt/sources.list.d/php.list 2>/dev/null || true").Run()
	_ = exec.Command("bash", "-c", "grep -l 'packages.sury.org/php zena' /etc/apt/sources.list.d/* 2>/dev/null | xargs rm -f 2>/dev/null || true").Run()
	_ = exec.Command("bash", "-c", "grep -l 'packages.sury.org/php wilma' /etc/apt/sources.list.d/* 2>/dev/null | xargs rm -f 2>/dev/null || true").Run()
	_ = exec.Command("bash", "-c", "grep -l 'packages.sury.org/php virginia' /etc/apt/sources.list.d/* 2>/dev/null | xargs rm -f 2>/dev/null || true").Run()
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

// AptUpdate executes apt-get update safely, cleaning broken sources first
func AptUpdate() error {
	CleanBrokenSources()
	err := RunCommand("apt-get", "update", "-y")
	if err != nil {
		// Retry with allow-releaseinfo-change if repo codenames recently changed
		_ = RunCommand("apt-get", "update", "-y", "--allow-releaseinfo-change")
	}
	return nil
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
