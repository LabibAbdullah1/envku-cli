package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
	"github.com/LabibAbdullah1/envku-cli/internal/servicectl"
)

// Daemon manages background health checks, automated watchdog, and systemd service registration
type Daemon struct {
	controller *servicectl.Controller
}

// New creates a new Daemon instance
func New() *Daemon {
	return &Daemon{
		controller: servicectl.New(),
	}
}

const ServiceUnitContent = `[Unit]
Description=Envku CLI Background Environment Monitor & Daemon
After=network.target nginx.service mysql.service

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/envku-cli daemon run
Restart=always
RestartSec=10s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

// InstallService installs the systemd unit file and enables the service
func (d *Daemon) InstallService() error {
	// Clean up legacy envku.service if exists
	_ = exec.Command("systemctl", "stop", "envku").Run()
	_ = exec.Command("systemctl", "disable", "envku").Run()
	_ = os.Remove("/etc/systemd/system/envku.service")

	servicePath := config.ServiceFile
	if err := os.WriteFile(servicePath, []byte(ServiceUnitContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service unit: %w", err)
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	if err := exec.Command("systemctl", "enable", "--now", config.ServiceName).Run(); err != nil {
		return fmt.Errorf("failed to enable and start %s service: %w", config.ServiceName, err)
	}

	return nil
}

// UninstallService disables and removes the envku-cli systemd unit
func (d *Daemon) UninstallService() error {
	_ = exec.Command("systemctl", "stop", config.ServiceName).Run()
	_ = exec.Command("systemctl", "disable", config.ServiceName).Run()
	_ = exec.Command("systemctl", "stop", "envku").Run()
	_ = exec.Command("systemctl", "disable", "envku").Run()

	_ = os.Remove(config.ServiceFile)
	_ = os.Remove("/etc/systemd/system/envku.service")

	return exec.Command("systemctl", "daemon-reload").Run()
}

// Run starts the daemon watchdog loop and waits for termination signals
func (d *Daemon) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("[Envku CLI Daemon] Background service watchdog started.")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case sig := <-sigChan:
			log.Printf("[Envku CLI Daemon] Received signal %v. Shutting down gracefully...", sig)
			return nil
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			d.performHealthCheck()
		}
	}
}

// performHealthCheck inspects critical services and logs their health status
func (d *Daemon) performHealthCheck() {
	services := d.controller.ListManagedServices()
	for _, s := range services {
		if !s.Active {
			log.Printf("[Envku CLI Watchdog WARNING] Service %s (%s) is INACTIVE!", s.Name, s.UnitName)
		}
	}
}
