package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const AppVersion = "3.1.0"

const (
	GitHubOwner = "OmarNaru1110"
	GitHubRepo  = "byto"
)

type VersionInfo struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
	Changelog   string `json:"changelog"`
	Downloads   struct {
		Windows string `json:"windows"`
		Darwin  string `json:"darwin"`
		Linux   string `json:"linux"`
	} `json:"downloads"`
	MinVersion string `json:"min_version"`
}

type UpdateResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	CurrentVersion string `json:"current_version,omitempty"`
	LatestVersion  string `json:"latest_version,omitempty"`
	HasUpdate      bool   `json:"has_update,omitempty"`
	Changelog      string `json:"changelog,omitempty"`
	DownloadURL    string `json:"download_url,omitempty"`
}

type Updater struct {
	httpClient *http.Client
}

func NewUpdater() *Updater {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 5,
	}

	return &Updater{
		httpClient: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: transport,
		},
	}
}

func (u *Updater) GetAppVersion() string {
	return AppVersion
}

func (u *Updater) CheckAppUpdate() UpdateResult {
	versionURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/main/version.json",
		GitHubOwner, GitHubRepo,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", versionURL, nil)
	if err != nil {
		return UpdateResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to create request: %v", err),
			CurrentVersion: AppVersion,
		}
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return UpdateResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to check for updates: %v", err),
			CurrentVersion: AppVersion,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to fetch version info: HTTP %d", resp.StatusCode),
			CurrentVersion: AppVersion,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UpdateResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to read response: %v", err),
			CurrentVersion: AppVersion,
		}
	}

	var versionInfo VersionInfo
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		return UpdateResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to parse version info: %v", err),
			CurrentVersion: AppVersion,
		}
	}

	hasUpdate := compareVersions(versionInfo.Version, AppVersion) > 0

	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = versionInfo.Downloads.Windows
	case "darwin":
		downloadURL = versionInfo.Downloads.Darwin
	default:
		downloadURL = versionInfo.Downloads.Linux
	}

	return UpdateResult{
		Success:        true,
		Message:        "Version check completed",
		CurrentVersion: AppVersion,
		LatestVersion:  versionInfo.Version,
		HasUpdate:      hasUpdate,
		Changelog:      versionInfo.Changelog,
		DownloadURL:    downloadURL,
	}
}

func (u *Updater) DownloadAppUpdate(downloadURL string, progressCallback func(downloaded, total int64)) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL provided")
	}

	resp, err := u.httpClient.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	downloadsDir := filepath.Join(homeDir, "Downloads")

	filename := filepath.Base(downloadURL)
	if filename == "" || filename == "." {
		filename = "byto-update.exe"
	}
	destPath := filepath.Join(downloadsDir, filename)

	file, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 256*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return "", fmt.Errorf("failed to write file: %v", writeErr)
			}
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to download: %v", err)
		}
	}

	return destPath, nil
}

func (u *Updater) LaunchInstaller(installerPath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", installerPath)
		hideWindow(cmd)
	case "darwin":
		cmd = exec.Command("open", installerPath)
	default:
		cmd = exec.Command("xdg-open", installerPath)
	}

	return cmd.Start()
}

func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &num1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &num2)
		}

		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}

	return 0
}
