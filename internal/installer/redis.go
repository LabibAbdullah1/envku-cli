package installer

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// RedisInstaller manages latest Redis server installation via official Redis APT repository
type RedisInstaller struct{}

func NewRedisInstaller() *RedisInstaller {
	return &RedisInstaller{}
}

// Status returns current Redis status
func (r *RedisInstaller) Status() PackageStatus {
	installed := IsBinaryInstalled("redis-server")
	active := false
	version := ""

	if installed {
		active = IsServiceActive("redis-server") || IsServiceActive("redis")
		version = r.GetVersion()
	}

	return PackageStatus{
		Name:      "redis-server",
		Installed: installed,
		Active:    active,
		Version:   version,
	}
}

// GetVersion extracts installed Redis version
func (r *RedisInstaller) GetVersion() string {
	out, err := RunCommandOutput("redis-server", "--version")
	if err != nil {
		return ""
	}
	// Output format: Redis server v=7.2.5 sha=00000000:0 malloc=jemalloc-5.3.0 bits=64 ...
	re := regexp.MustCompile(`v=([0-9.]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(out)
}

// SetupRepository configures the official Redis APT repository
func (r *RedisInstaller) SetupRepository() error {
	if err := AptInstall("lsb-release", "curl", "gpg"); err != nil {
		return fmt.Errorf("failed to install prerequisite packages: %w", err)
	}

	cmdKey := "curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg --yes"
	if err := exec.Command("bash", "-c", cmdKey).Run(); err != nil {
		return fmt.Errorf("failed to import redis gpg key: %w", err)
	}

	cmdRepo := `echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list`
	if err := exec.Command("bash", "-c", cmdRepo).Run(); err != nil {
		return fmt.Errorf("failed to register redis apt repository: %w", err)
	}

	return AptUpdate()
}

// Install installs the latest Redis server from official repository
func (r *RedisInstaller) Install() error {
	if !IsRoot() {
		return fmt.Errorf("root privileges required to install Redis")
	}

	if err := r.SetupRepository(); err != nil {
		return err
	}

	if err := AptInstall("redis"); err != nil {
		return fmt.Errorf("failed to install redis package: %w", err)
	}

	if err := EnableAndStartService("redis-server"); err != nil {
		// Fallback to 'redis' service name if redis-server fails
		_ = EnableAndStartService("redis")
	}

	return nil
}
