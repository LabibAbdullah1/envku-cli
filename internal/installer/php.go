package installer

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// PHPInstaller coordinates PHP multi-version installations via Ondrej Sury repository
type PHPInstaller struct{}

func NewPHPInstaller() *PHPInstaller {
	return &PHPInstaller{}
}

// SetupRepository adds the official Ondřej Surý PHP repository correctly tailored for Debian, Ubuntu, and Ubuntu derivatives (Linux Mint, Pop!_OS, Zorin)
func (p *PHPInstaller) SetupRepository() error {
	osInfo, err := DetectOS()
	if err != nil {
		osInfo = &DistroInfo{ID: "ubuntu", Codename: "noble"}
	}

	// 1. Clean up legacy or broken Sury/Launchpad entries that failed previously (e.g. invalid 'zena' codename on Linux Mint)
	cleanCmd := "rm -f /etc/apt/sources.list.d/php.list /etc/apt/sources.list.d/*ondrej* 2>/dev/null || true"
	_ = exec.Command("bash", "-c", cleanCmd).Run()

	// 2. Install prerequisites
	packages := []string{"software-properties-common", "apt-transport-https", "ca-certificates", "curl", "gnupg", "lsb-release"}
	_ = AptInstall(packages...)

	_ = os.MkdirAll("/etc/apt/keyrings", 0755)
	_ = os.MkdirAll("/etc/apt/trusted.gpg.d", 0755)

	// 3. Configure repository according to distribution type
	if osInfo.IsDebian() {
		// Pure Debian or LMDE (Linux Mint Debian Edition)
		debianCodename := osInfo.GetEffectiveDebianCodename()
		fmt.Printf("[*] Terdeteksi sistem Debian (Codename basis: %s)...\n", debianCodename)

		cmdGPG := "curl -sSLo /etc/apt/trusted.gpg.d/php.gpg https://packages.sury.org/php/apt.gpg && cp /etc/apt/trusted.gpg.d/php.gpg /etc/apt/keyrings/sury-php.gpg 2>/dev/null || true"
		if err := exec.Command("bash", "-c", cmdGPG).Run(); err != nil {
			return fmt.Errorf("gagal mengunduh GPG key Sury Debian: %w", err)
		}

		cmdRepo := fmt.Sprintf(`echo "deb [signed-by=/etc/apt/trusted.gpg.d/php.gpg] https://packages.sury.org/php/ %s main" > /etc/apt/sources.list.d/php.list`, debianCodename)
		if err := exec.Command("bash", "-c", cmdRepo).Run(); err != nil {
			return fmt.Errorf("gagal menulis repositori Debian php.list: %w", err)
		}
	} else {
		// Ubuntu & Ubuntu Derivatives (Linux Mint, Pop!_OS, Zorin OS, Elementary OS)
		ubuntuCodename := osInfo.GetEffectiveUbuntuCodename()
		fmt.Printf("[*] Terdeteksi sistem basis Ubuntu/Linux Mint (Codename basis: %s)...\n", ubuntuCodename)

		// Fetch official Launchpad PPA GPG key for Ondrej PHP
		cmdGPG := `curl -fsSL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x14AA40EC0831756756D7F66C4F4EA0AAE5267A6C" | gpg --dearmor --yes -o /etc/apt/trusted.gpg.d/ondrej-ubuntu-php.gpg 2>/dev/null || true`
		_ = exec.Command("bash", "-c", cmdGPG).Run()

		// Write canonical Launchpad PPA entry with resolved Ubuntu codename (fixing 404 on Linux Mint 'zena'/'wilma')
		cmdRepo := fmt.Sprintf(`echo "deb [signed-by=/etc/apt/trusted.gpg.d/ondrej-ubuntu-php.gpg] https://ppa.launchpadcontent.net/ondrej/php/ubuntu %s main" > /etc/apt/sources.list.d/ondrej-ubuntu-php.list`, ubuntuCodename)
		if err := exec.Command("bash", "-c", cmdRepo).Run(); err != nil {
			// Fallback: try add-apt-repository utility directly
			if errPPA := EnsurePPA("ppa:ondrej/php"); errPPA != nil {
				return fmt.Errorf("gagal menambahkan repositori PPA Ondrej untuk Ubuntu/Mint (%s): %w", ubuntuCodename, errPPA)
			}
			return nil
		}
	}

	return AptUpdate()
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
