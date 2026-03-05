package deps

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type YTDLPDependency struct {
	binPath         string
	ttl             time.Duration
	ytDlpReleaseURL string
}

func NewYTDLPDependency(dirPath string, ttl time.Duration) *YTDLPDependency {
	ytdlpName := "yt-dlp"
	if runtime.GOOS == "windows" {
		ytdlpName = "yt-dlp.exe"
	}
	binPath := filepath.Join(dirPath, ytdlpName)
	return &YTDLPDependency{
		binPath:         binPath,
		ttl:             ttl,
		ytDlpReleaseURL: "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"}
}

func (d *YTDLPDependency) GetName() string {
	return "yt-dlp"
}

func (d *YTDLPDependency) Path() string {
	return d.binPath
}

func (d *YTDLPDependency) TTL() time.Duration {
	return d.ttl
}

func (y *YTDLPDependency) CheckInstalled() (bool, error) {
	_, err := os.Stat(y.binPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (y *YTDLPDependency) assetNamesForPlatform() []string {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "arm64" {
			return []string{"yt-dlp_arm64.exe", "yt-dlp.exe"}
		}
		if runtime.GOARCH == "386" || runtime.GOARCH == "amd64" {
			return []string{"yt-dlp_x86.exe", "yt-dlp.exe"}
		}
		return []string{"yt-dlp.exe"}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return []string{"yt-dlp_macos"}
		}
		return []string{"yt-dlp_macos_legacy", "yt-dlp_macos"}
	default:
		return []string{"yt-dlp"}
	}
}

func (y *YTDLPDependency) Version() (string, error) {
	cmd := exec.Command(y.binPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out.Bytes())), nil
}

func (y *YTDLPDependency) Install(progress DownloadProgressCallback) (err error) {
	slog.Info("installing yt-dlp", "path", y.binPath)
	defer func() {
		if err != nil {
			os.Remove(y.binPath)
		}
	}()
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true, // Faster for binary downloads
		MaxIdleConnsPerHost: 5,
	}
	httpClient := &http.Client{
		Timeout:   5 * time.Minute,
		Transport: transport,
	}
	resp, err := httpClient.Get(y.ytDlpReleaseURL)
	if err != nil {
		return fmt.Errorf("yt-dlp: failed to fetch releases from %s: %w", y.ytDlpReleaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("yt-dlp: releases API returned %d: %s", resp.StatusCode, resp.Status)
	}

	var release struct {
		TagName string        `json:"tag_name"`
		Assets  []githubAsset `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("yt-dlp: failed to parse release JSON: %w", err)
	}

	var downloadURL, assetName string
	assetNames := y.assetNamesForPlatform()

	for _, name := range assetNames {
		for _, asset := range release.Assets {
			if asset.Name == name {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				break
			}
		}
		if downloadURL != "" {
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("yt-dlp: no download found for %s/%s (tried: %v)", runtime.GOOS, runtime.GOARCH, assetNames)
	}

	// Download with retry (3 attempts, exponential backoff)
	const maxRetries = 3
	var dlResp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		dlResp, err = httpClient.Get(downloadURL)
		if err != nil {
			if attempt == maxRetries-1 {
				return fmt.Errorf("yt-dlp: download failed after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		if dlResp.StatusCode == http.StatusOK {
			break
		}
		statusErr := fmt.Errorf("status %d: %s", dlResp.StatusCode, dlResp.Status)
		dlResp.Body.Close()
		if attempt == maxRetries-1 {
			return fmt.Errorf("yt-dlp: download failed after %d attempts: %w", maxRetries, statusErr)
		}
	}
	defer dlResp.Body.Close()

	out, err := os.Create(y.binPath)
	if err != nil {
		return fmt.Errorf("yt-dlp: failed to create %s: %w", y.binPath, err)
	}
	defer out.Close()

	var downloaded int64

	// Use larger buffer for faster downloads (256KB)
	buf := make([]byte, 256*1024)
	for {
		n, err := dlResp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
		}

		if progress != nil {
			progress(downloaded, dlResp.ContentLength)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("yt-dlp: download interrupted at %d bytes: %w", downloaded, err)
		}
	}

	// Verify checksum if available
	if expectedHash := y.fetchExpectedHash(release.Assets, assetName, httpClient); expectedHash != "" {
		if err := y.verifyChecksum(expectedHash); err != nil {
			return err
		}
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(y.binPath, 0755); err != nil {
			return fmt.Errorf("yt-dlp: failed to make executable: %w", err)
		}
	}

	return nil
}

func (y *YTDLPDependency) fetchExpectedHash(assets []githubAsset, assetName string, client *http.Client) string {
	var sumsURL string
	for _, a := range assets {
		if a.Name == "SHA2-256SUMS" {
			sumsURL = a.BrowserDownloadURL
			break
		}
	}
	if sumsURL == "" {
		return ""
	}
	resp, err := client.Get(sumsURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == assetName {
			return parts[0]
		}
	}
	return ""
}

func (y *YTDLPDependency) verifyChecksum(expectedHex string) error {
	f, err := os.Open(y.binPath)
	if err != nil {
		return fmt.Errorf("yt-dlp: failed to verify checksum: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("yt-dlp: checksum read failed: %w", err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expectedHex) {
		return fmt.Errorf("yt-dlp: checksum mismatch (expected %s, got %s)", expectedHex, got)
	}
	return nil
}

func (y *YTDLPDependency) Update(progress DownloadProgressCallback) error {
	slog.Info("updating yt-dlp via self-update", "path", y.binPath)
	if progress != nil {
		progress(0, 0)
	}
	cmd := exec.Command(y.binPath, "-U")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		slog.Warn("yt-dlp self-update failed, falling back to re-download", "error", err, "output", out.String())
		// Self-update can fail if the exe is corrupted (e.g. PyInstaller PKG error).
		// Fall back to full re-download from GitHub.
		return y.Install(progress)
	}
	if progress != nil {
		progress(1, 1)
	}
	return nil
}
