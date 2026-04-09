package deps

import (
	"archive/zip"
	"bytes"
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

type DenoDependency struct {
	binPath        string
	ttl            time.Duration
	denoReleaseURL string
}

func NewDenoDependency(dirPath string, ttl time.Duration) *DenoDependency {
	denoName := "deno"
	if runtime.GOOS == "windows" {
		denoName = "deno.exe"
	}
	binPath := filepath.Join(dirPath, denoName)
	return &DenoDependency{
		binPath:        binPath,
		ttl:            ttl,
		denoReleaseURL: "https://api.github.com/repos/denoland/deno/releases/latest",
	}
}

func (d *DenoDependency) GetName() string {
	return "deno"
}

func (d *DenoDependency) Path() string {
	return d.binPath
}

func (d *DenoDependency) TTL() time.Duration {
	return d.ttl
}

func (d *DenoDependency) CheckInstalled() (bool, error) {
	_, err := os.Stat(d.binPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DenoDependency) assetNameForPlatform() string {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "arm64" {
			return "deno-aarch64-pc-windows-msvc.zip"
		}
		return "deno-x86_64-pc-windows-msvc.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "deno-aarch64-apple-darwin.zip"
		}
		return "deno-x86_64-apple-darwin.zip"
	default:
		if runtime.GOARCH == "arm64" {
			return "deno-aarch64-unknown-linux-gnu.zip"
		}
		return "deno-x86_64-unknown-linux-gnu.zip"
	}
}

func (d *DenoDependency) Version() (string, error) {
	cmd := exec.Command(d.binPath, "--version")
	HideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	// "deno 1.x.y (release, ...)\nv8 ...\ntypescript ..."
	lines := strings.Split(string(bytes.TrimSpace(out.Bytes())), "\n")
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 2 {
			return fields[1], nil
		}
	}
	return "unknown", nil
}

func (d *DenoDependency) Install(progress DownloadProgressCallback) (err error) {
	slog.Info("installing deno", "path", d.binPath)
	defer func() {
		if err != nil {
			os.Remove(d.binPath)
		}
	}()

	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 5,
	}
	httpClient := &http.Client{
		Timeout:   5 * time.Minute,
		Transport: transport,
	}
	resp, err := httpClient.Get(d.denoReleaseURL)
	if err != nil {
		return fmt.Errorf("deno: failed to fetch releases from %s: %w", d.denoReleaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deno: releases API returned %d: %s", resp.StatusCode, resp.Status)
	}

	var release struct {
		Assets []githubAsset `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("deno: failed to parse release JSON: %w", err)
	}

	var downloadURL string
	assetName := d.assetNameForPlatform()

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("deno: no download found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	const maxRetries = 3
	var dlResp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		dlResp, err = httpClient.Get(downloadURL)
		if err != nil {
			if attempt == maxRetries-1 {
				return fmt.Errorf("deno: download failed after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		if dlResp.StatusCode == http.StatusOK {
			break
		}
		statusErr := fmt.Errorf("status %d: %s", dlResp.StatusCode, dlResp.Status)
		dlResp.Body.Close()
		if attempt == maxRetries-1 {
			return fmt.Errorf("deno: download failed after %d attempts: %w", maxRetries, statusErr)
		}
	}
	defer dlResp.Body.Close()

	tmpFile, err := os.CreateTemp("", "deno-*.zip")
	if err != nil {
		return fmt.Errorf("deno: failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	var downloaded int64
	buf := make([]byte, 256*1024)
	for {
		n, err := dlResp.Body.Read(buf)
		if n > 0 {
			tmpFile.Write(buf[:n])
			downloaded += int64(n)
		}
		if progress != nil {
			progress(downloaded, dlResp.ContentLength)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			return fmt.Errorf("deno: download interrupted at %d bytes: %w", downloaded, err)
		}
	}
	tmpFile.Close()

	if err := d.extractFromZip(tmpPath, runtime.GOOS == "windows"); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(d.binPath, 0755); err != nil {
			return fmt.Errorf("deno: failed to make executable: %w", err)
		}
	}

	return nil
}

func (d *DenoDependency) extractFromZip(zipPath string, isWindows bool) error {
	r, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("deno: failed to open zip: %w", err)
	}
	defer r.Close()
	stat, err := r.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(r, stat.Size())
	if err != nil {
		return fmt.Errorf("deno: invalid zip: %w", err)
	}
	exeName := "deno"
	if isWindows {
		exeName = "deno.exe"
	}
	for _, file := range zr.File {
		if file.Name == exeName {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			out, err := os.Create(d.binPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("deno: %s not found in zip", exeName)
}

func (d *DenoDependency) Update(progress DownloadProgressCallback) error {
	slog.Info("updating deno", "path", d.binPath)
	if progress != nil {
		progress(0, 0)
	}
	cmd := exec.Command(d.binPath, "upgrade")
	HideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		slog.Warn("deno upgrade failed, falling back to re-download", "error", err, "output", out.String())
		return d.Install(progress)
	}
	if progress != nil {
		progress(1, 1)
	}
	return nil
}
