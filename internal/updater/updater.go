package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
		Timeout: 3 * time.Second,
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Envku-CLI")
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

	// Determine matching asset binary name
	expectedAsset := fmt.Sprintf("envku-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	for _, a := range rel.Assets {
		if a.Name == expectedAsset {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}

	// Fallback direct URL if asset list is not parsed
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/%s", RepoOwner, RepoName, expectedAsset)
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

// PerformSelfUpdate downloads the latest release binary and atomically replaces current binary
func PerformSelfUpdate(info *UpdateInfo) error {
	if info == nil || info.DownloadURL == "" {
		return fmt.Errorf("invalid update information or missing download URL")
	}

	// Determine destination path
	destPath, err := os.Executable()
	if err != nil {
		destPath = "/usr/local/bin/envku"
	}
	destPath, err = filepath.EvalSymlinks(destPath)
	if err != nil {
		destPath = "/usr/local/bin/envku"
	}

	destDir := filepath.Dir(destPath)
	tmpFile, err := os.CreateTemp(destDir, "envku-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file in %s (permission denied? Try running with sudo): %w", destDir, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download binary with progress timeout
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(info.DownloadURL)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download binary from %s: %w", info.DownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("download failed with HTTP status: %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write binary content: %w", err)
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	// Atomically replace target binary
	if err := os.Rename(tmpPath, destPath); err != nil {
		// On Windows or cross-device, try copy fallback
		return fmt.Errorf("failed to replace binary %s: %w", destPath, err)
	}

	return nil
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
