package installer

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// PHPInstaller coordinates PHP multi-version installations via Ondrej Sury repository
type PHPInstaller struct{}

func NewPHPInstaller() *PHPInstaller {
	return &PHPInstaller{}
}

// SetupRepository adds the Ondřej Surý PHP repository
func (p *PHPInstaller) SetupRepository() error {
	osInfo, err := DetectOS()
	if err != nil {
		// Fallback to PPA if detection fails
		return EnsurePPA("ppa:ondrej/php")
	}

	switch osInfo.ID {
	case "debian":
		// Debian repository setup
		packages := []string{"apt-transport-https", "lsb-release", "ca-certificates", "curl"}
		if err := AptInstall(packages...); err != nil {
			return err
		}
		// Fetch GPG key
		cmdGPG := "curl -sSLo /etc/apt/trusted.gpg.d/php.gpg https://packages.sury.org/php/apt.gpg"
		if err := exec.Command("bash", "-c", cmdGPG).Run(); err != nil {
			return fmt.Errorf("failed to fetch sury debian gpg key: %w", err)
		}
		// Add sources.list
		cmdRepo := `echo "deb https://packages.sury.org/php/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/php.list`
		if err := exec.Command("bash", "-c", cmdRepo).Run(); err != nil {
			return fmt.Errorf("failed to write sury debian php.list: %w", err)
		}
		return AptUpdate()
	case "ubuntu":
		return EnsurePPA("ppa:ondrej/php")
	default:
		return EnsurePPA("ppa:ondrej/php")
	}
}

// InstallVersion installs PHP and essential extensions for the specified version (e.g. "8.2", "8.3", "8.4", "8.5")
func (p *PHPInstaller) InstallVersion(version string) error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to install PHP %s", version)
	}

	if err := p.SetupRepository(); err != nil {
		return fmt.Errorf("failed to configure PHP repository: %w", err)
	}

	packages := []string{
		fmt.Sprintf("php%s-fpm", version),
		fmt.Sprintf("php%s-cli", version),
		fmt.Sprintf("php%s-common", version),
		fmt.Sprintf("php%s-mysql", version),
		fmt.Sprintf("php%s-mbstring", version),
		fmt.Sprintf("php%s-xml", version),
		fmt.Sprintf("php%s-curl", version),
		fmt.Sprintf("php%s-zip", version),
		fmt.Sprintf("php%s-bcmath", version),
		fmt.Sprintf("php%s-gd", version),
		fmt.Sprintf("php%s-intl", version),
		fmt.Sprintf("php%s-redis", version),
	}

	if err := AptInstall(packages...); err != nil {
		return fmt.Errorf("failed to install PHP %s stack: %w", version, err)
	}

	serviceName := fmt.Sprintf("php%s-fpm", version)
	if err := EnableAndStartService(serviceName); err != nil {
		return fmt.Errorf("failed to enable %s: %w", serviceName, err)
	}

	return nil
}

// Status returns the installation and service state of a specific PHP version
func (p *PHPInstaller) Status(version string) PackageStatus {
	binName := fmt.Sprintf("php%s", version)
	serviceName := fmt.Sprintf("php%s-fpm", version)

	installed := IsBinaryInstalled(binName)
	active := false
	verString := ""

	if installed {
		active = IsServiceActive(serviceName)
		verString = p.GetVersion(binName)
	}

	return PackageStatus{
		Name:      serviceName,
		Installed: installed,
		Active:    active,
		Version:   verString,
	}
}

// ListAllStatuses returns statuses for all supported modern PHP versions (8.2 - 8.5)
func (p *PHPInstaller) ListAllStatuses() []PackageStatus {
	var statuses []PackageStatus
	for _, v := range SupportedPHPVersions {
		statuses = append(statuses, p.Status(v))
	}
	return statuses
}

// GetVersion queries CLI version string from php executable
func (p *PHPInstaller) GetVersion(binName string) string {
	out, err := RunCommandOutput(binName, "-v")
	if err != nil {
		return ""
	}
	// Format: "PHP 8.3.10 (cli)..."
	re := regexp.MustCompile(`PHP\s+([0-9.]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(out)
}
