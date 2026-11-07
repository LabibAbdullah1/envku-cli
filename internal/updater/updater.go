package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LabibAbdullah1/envku-cli/internal/config"
)

const (
	RepoOwner = "LabibAbdullah1"
	RepoName  = "envku-cli"
)

// UpdateInfo holds details about available updates
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseName    string
	ReleaseNotes   string
	DownloadURL    string
}

type gitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckUpdate checks GitHub Releases API for a newer version
func CheckUpdate() (*UpdateInfo, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "envku-cli")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var rel gitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}

	cleanLatest := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	cleanCurrent := strings.TrimPrefix(strings.TrimSpace(config.AppVersion), "v")

	isNewer := compareSemVer(cleanLatest, cleanCurrent) > 0

	// Determine matching asset binary name (prioritize envku-cli-, fallback to envku-)
	expectedPrimary := fmt.Sprintf("envku-cli-%s-%s", runtime.GOOS, runtime.GOARCH)
	expectedLegacy := fmt.Sprintf("envku-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	for _, a := range rel.Assets {
		if a.Name == expectedPrimary {
			downloadURL = a.BrowserDownloadURL
			break
		}
		if a.Name == expectedLegacy && downloadURL == "" {
			downloadURL = a.BrowserDownloadURL
		}
	}

	// Fallback direct URL if asset list is not parsed
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/%s", RepoOwner, RepoName, expectedPrimary)
	}

	return &UpdateInfo{
		Available:      isNewer,
		CurrentVersion: cleanCurrent,
		LatestVersion:  cleanLatest,
		ReleaseName:    rel.Name,
		ReleaseNotes:   rel.Body,
		DownloadURL:    downloadURL,
	}, nil
}

// PerformSelfUpdate downloads the latest release binary and safely replaces current binary with backup & rollback
func PerformSelfUpdate(info *UpdateInfo) error {
	if info == nil || info.DownloadURL == "" {
		return fmt.Errorf("invalid update information or missing download URL")
	}

	// 1. Determine destination executable path
	destPath, err := os.Executable()
	if err != nil {
		destPath = config.DefaultBinPath
	}
	if evaluated, errEval := filepath.EvalSymlinks(destPath); errEval == nil {
		destPath = evaluated
	}

	destDir := filepath.Dir(destPath)

	// 2. Prepare temporary download file in same directory (ensures atomic rename on same filesystem)
	tmpFile, err := os.CreateTemp(destDir, "envku-cli-update-*")
	if err != nil {
		return fmt.Errorf("gagal membuat file sementara di %s (izin akses ditolak? Jalankan dengan sudo): %w", destDir, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 3. Download binary safely
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Get(info.DownloadURL)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("gagal mengunduh binary pembaruan dari %s: %w", info.DownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("unduhan gagal dengan status HTTP: %s", resp.Status)
	}

	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("gagal menulis konten binary: %w", err)
	}
	_ = tmpFile.Close()

	// 4. Integrity & Valid Executable Check (Security Guard)
	if written < 500*1024 {
		return fmt.Errorf("validasi gagal: binary yang diunduh terlalu kecil (%d bytes), kemungkinan file corrupt atau halaman error", written)
	}

	// Check magic bytes header
	if err := validateBinaryHeader(tmpPath); err != nil {
		return fmt.Errorf("validasi format binary gagal: %w", err)
	}

	// Set executable permission
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("gagal menyetel hak akses eksekusi: %w", err)
	}

	// 5. Check if background daemon is active and temporarily stop it to avoid ETXTBSY (text file busy)
	daemonService := config.ServiceName
	restartDaemon := false
	if isServiceActive(daemonService) {
		_ = exec.Command("systemctl", "stop", daemonService).Run()
		restartDaemon = true
	} else if isServiceActive("envku") {
		_ = exec.Command("systemctl", "stop", "envku").Run()
		daemonService = "envku"
		restartDaemon = true
	}

	// 6. Safe Backup Existing Binary before replacement
	bakPath := destPath + ".bak"
	_ = os.Remove(bakPath)
	if _, err := os.Stat(destPath); err == nil {
		if err := os.Rename(destPath, bakPath); err != nil {
			// On Windows if rename fails while running, continue
			_ = copyFile(destPath, bakPath)
		}
	}

	// 7. Atomically place new binary
	if err := os.Rename(tmpPath, destPath); err != nil {
		// Rollback if failed
		if _, statErr := os.Stat(bakPath); statErr == nil {
			_ = os.Rename(bakPath, destPath)
		}
		if restartDaemon {
			_ = exec.Command("systemctl", "start", daemonService).Run()
		}
		return fmt.Errorf("gagal menimpa binary (rollback telah dilakukan): %w", err)
	}

	// Remove backup on success
	_ = os.Remove(bakPath)

	// Ensure convenience symlink /usr/local/bin/envku -> /usr/local/bin/envku-cli
	if destPath == config.DefaultBinPath {
		_ = os.Remove(config.SymlinkBinPath)
		_ = os.Symlink(config.DefaultBinPath, config.SymlinkBinPath)
	}

	// 8. Restart background daemon if it was running
	if restartDaemon {
		_ = exec.Command("systemctl", "start", daemonService).Run()
	}

	return nil
}

// validateBinaryHeader checks whether the file has valid OS executable magic bytes
func validateBinaryHeader(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 4)
	if _, err := io.ReadFull(f, header); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "linux":
		// ELF header: 0x7F, 'E', 'L', 'F'
		if !bytes.Equal(header, []byte{0x7F, 'E', 'L', 'F'}) {
			return fmt.Errorf("bukan binary ELF Linux yang valid")
		}
	case "windows":
		// PE executable: 'M', 'Z'
		if header[0] != 'M' || header[1] != 'Z' {
			return fmt.Errorf("bukan executable Windows PE yang valid")
		}
	}

	return nil
}

func isServiceActive(service string) bool {
	out, err := exec.Command("systemctl", "is-active", service).Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// compareSemVer returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareSemVer(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}

	for i := 0; i < 3; i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}
	return 0
}
