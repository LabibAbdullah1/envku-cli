package servicectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Action represents a systemctl lifecycle operation
type Action string

const (
	ActionStart   Action = "start"
	ActionStop    Action = "stop"
	ActionRestart Action = "restart"
	ActionReload  Action = "reload"
	ActionEnable  Action = "enable"
	ActionDisable Action = "disable"
)

// ServiceState holds the runtime and boot status of a service unit
type ServiceState struct {
	Name        string `json:"name"`
	UnitName    string `json:"unit_name"`
	Active      bool   `json:"active"`
	SubState    string `json:"sub_state"` // e.g. "running", "dead", "exited"
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// Controller manages Linux systemd services and log inspection
type Controller struct{}

// New creates a new Controller instance
func New() *Controller {
	return &Controller{}
}

// Control executes a systemctl action on a service
func (c *Controller) Control(service string, action Action) error {
	unit := c.normalizeUnit(service)

	// Validate action
	switch action {
	case ActionStart, ActionStop, ActionRestart, ActionReload, ActionEnable, ActionDisable:
		// valid
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	cmd := exec.Command("systemctl", string(action), unit)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to %s %s: %s (%w)", action, unit, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// GetStatus queries the systemctl status of a specific service unit
func (c *Controller) GetStatus(service string) (*ServiceState, error) {
	unit := c.normalizeUnit(service)

	// Check if active
	activeOut, _ := c.execQuiet("systemctl", "is-active", unit)
	isActive := strings.TrimSpace(activeOut) == "active"

	// Check if enabled
	enabledOut, _ := c.execQuiet("systemctl", "is-enabled", unit)
	isEnabled := strings.TrimSpace(enabledOut) == "enabled"

	// Query substate
	subStateOut, _ := c.execQuiet("systemctl", "show", "-p", "SubState", "--value", unit)
	subState := strings.TrimSpace(subStateOut)

	// Query description
	descOut, _ := c.execQuiet("systemctl", "show", "-p", "Description", "--value", unit)
	desc := strings.TrimSpace(descOut)

	return &ServiceState{
		Name:        service,
		UnitName:    unit,
		Active:      isActive,
		SubState:    subState,
		Enabled:     isEnabled,
		Description: desc,
	}, nil
}

// ListManagedServices queries the statuses of all web and database services managed by Envku
func (c *Controller) ListManagedServices() []ServiceState {
	candidates := []struct {
		name string
		unit string
	}{
		{"Nginx Web Server", "nginx"},
		{"MySQL Database", "mysql"},
		{"Redis Cache", "redis-server"},
		{"PHP 8.2 FPM", "php8.2-fpm"},
		{"PHP 8.3 FPM", "php8.3-fpm"},
		{"PHP 8.4 FPM", "php8.4-fpm"},
		{"PHP 8.5 FPM", "php8.5-fpm"},
		{"Envku Daemon", "envku"},
	}

	var results []ServiceState
	for _, cand := range candidates {
		// Only include if unit file exists or is loaded
		if !c.unitExists(cand.unit) {
			continue
		}

		st, err := c.GetStatus(cand.unit)
		if err == nil {
			st.Name = cand.name
			results = append(results, *st)
		}
	}

	return results
}

// unitExists checks whether the systemd unit file is present or known to systemd
func (c *Controller) unitExists(unit string) bool {
	norm := c.normalizeUnit(unit)
	out, err := c.execQuiet("systemctl", "list-unit-files", norm)
	if err == nil && strings.Contains(out, norm) {
		return true
	}
	// Fallback check in /etc/systemd/system or /lib/systemd/system
	paths := []string{
		filepath.Join("/etc/systemd/system", norm),
		filepath.Join("/lib/systemd/system", norm),
		filepath.Join("/usr/lib/systemd/system", norm),
	}
	for _, p := range paths {
		matches, _ := filepath.Glob(p)
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

func (c *Controller) normalizeUnit(service string) string {
	if !strings.HasSuffix(service, ".service") {
		return service + ".service"
	}
	return service
}

func (c *Controller) execQuiet(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return string(out), err
}
