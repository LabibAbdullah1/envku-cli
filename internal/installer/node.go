package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const NVMVersion = "v0.40.1"

// NodeInstaller manages NVM (Node Version Manager) and Node.js LTS installations
type NodeInstaller struct{}

func NewNodeInstaller() *NodeInstaller {
	return &NodeInstaller{}
}

// Status returns current Node.js and NVM status
func (n *NodeInstaller) Status() PackageStatus {
	nodeInstalled := IsBinaryInstalled("node")
	version := ""

	if nodeInstalled {
		version = n.GetNodeVersion()
	} else {
		// Check if node exists inside ~/.nvm/versions/node/...
		home, _ := os.UserHomeDir()
		nvmDir := filepath.Join(home, ".nvm")
		if _, err := os.Stat(nvmDir); err == nil {
			version = n.GetNVMNodeVersion()
			if version != "" {
				nodeInstalled = true
			}
		}
	}

	return PackageStatus{
		Name:      "node-nvm",
		Installed: nodeInstalled,
		Active:    nodeInstalled,
		Version:   version,
	}
}

// GetNodeVersion gets Node version via system PATH
func (n *NodeInstaller) GetNodeVersion() string {
	out, err := RunCommandOutput("node", "-v")
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "v")
}

// GetNVMNodeVersion checks Node version through NVM bash sourcing
func (n *NodeInstaller) GetNVMNodeVersion() string {
	script := `export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
node -v 2>/dev/null || true`
	cmd := exec.Command("bash", "-c", script)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
}

// Install provisions NVM and automatically installs Node.js LTS
func (n *NodeInstaller) Install() error {
	// Ensure curl is available
	if !IsBinaryInstalled("curl") {
		if err := AptInstall("curl"); err != nil {
			return fmt.Errorf("failed to install curl for NVM: %w", err)
		}
	}

	// Install NVM script
	nvmScriptURL := fmt.Sprintf("https://raw.githubusercontent.com/nvm-sh/nvm/%s/install.sh", NVMVersion)
	installCmd := fmt.Sprintf("curl -o- %s | bash", nvmScriptURL)

	cmd := exec.Command("bash", "-c", installCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute NVM installer: %w", err)
	}

	// Source NVM and install Node.js LTS
	nodeLTSCmd := `export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
nvm install --lts
nvm use --lts
nvm alias default 'lts/*'`

	cmdLTS := exec.Command("bash", "-c", nodeLTSCmd)
	cmdLTS.Stdout = os.Stdout
	cmdLTS.Stderr = os.Stderr
	if err := cmdLTS.Run(); err != nil {
		return fmt.Errorf("failed to install Node.js LTS via NVM: %w", err)
	}

	return nil
}
