package deps

import (
	"archive/tar"
	"archive/zip"
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

	"github.com/ulikunitz/xz"
)

var ffmpegDownloadURLs = map[string]string{
	"windows": "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
	"darwin":  "https://evermeet.cx/ffmpeg/ffmpeg.zip",
	"linux":   "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz",
}

type FfmpegDependency struct {
	binPath string
	ttl     time.Duration
}

func NewFfmpegDependency(dirPath string, ttl time.Duration) *FfmpegDependency {
	ffmpegName := "ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
	}
	binPath := filepath.Join(dirPath, ffmpegName)
	return &FfmpegDependency{
		binPath: binPath,
		ttl:     ttl,
	}
}

func (f *FfmpegDependency) GetName() string {
	return "ffmpeg"
}

func (f *FfmpegDependency) Path() string {
	return f.binPath
}

func (f *FfmpegDependency) TTL() time.Duration {
	return f.ttl
}

func (f *FfmpegDependency) CheckInstalled() (bool, error) {
	_, err := os.Stat(f.binPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *FfmpegDependency) Version() (string, error) {
	cmd := exec.Command(f.binPath, "-version")
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	lines := strings.SplitN(out.String(), "\n", 2)
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 3 {
			return fields[2], nil
		}
	}
	return "unknown", nil
}

func (f *FfmpegDependency) Install(progress DownloadProgressCallback) (err error) {
	slog.Info("installing ffmpeg", "path", f.binPath)
	defer func() {
		if err != nil {
			os.Remove(f.binPath)
		}
	}()

	url, ok := ffmpegDownloadURLs[runtime.GOOS]
	if !ok {
		return fmt.Errorf("ffmpeg: auto-download not supported for %s", runtime.GOOS)
	}

	httpClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("ffmpeg: failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ffmpeg: download returned %d: %s", resp.StatusCode, resp.Status)
	}

	ext := ".zip"
	if strings.HasSuffix(url, ".tar.xz") {
		ext = ".tar.xz"
	}
	tmpFile, err := os.CreateTemp("", "ffmpeg-*"+ext)
	if err != nil {
		return fmt.Errorf("ffmpeg: failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 256*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			tmpFile.Write(buf[:n])
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ffmpeg: download interrupted: %w", err)
		}
	}
	tmpFile.Close()

	osName := runtime.GOOS
	switch osName {
	case "windows":
		if err := f.extractFromZip(tmpPath, true); err != nil {
			return err
		}
	case "darwin":
		if err := f.extractFromZip(tmpPath, false); err != nil {
			return err
		}
		if err := os.Chmod(f.binPath, 0755); err != nil {
			return fmt.Errorf("ffmpeg: failed to make executable: %w", err)
		}
	case "linux":
		if err := f.extractFromTarXZ(tmpPath); err != nil {
			return err
		}
		if err := os.Chmod(f.binPath, 0755); err != nil {
			return fmt.Errorf("ffmpeg: failed to make executable: %w", err)
		}
	default:
		return fmt.Errorf("ffmpeg: unsupported OS %s", osName)
	}

	return nil
}

func (f *FfmpegDependency) extractFromZip(zipPath string, isWindows bool) error {
	r, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("ffmpeg: failed to open zip: %w", err)
	}
	defer r.Close()
	stat, err := r.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(r, stat.Size())
	if err != nil {
		return fmt.Errorf("ffmpeg: invalid zip: %w", err)
	}
	exeName := "ffmpeg"
	if isWindows {
		exeName = "ffmpeg.exe"
	}
	for _, file := range zr.File {
		if strings.HasSuffix(file.Name, "/"+exeName) || file.Name == exeName {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			out, err := os.Create(f.binPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("ffmpeg: %s not found in zip", exeName)
}

func (f *FfmpegDependency) extractFromTarXZ(tarxzPath string) error {
	file, err := os.Open(tarxzPath)
	if err != nil {
		return fmt.Errorf("ffmpeg: failed to open tar.xz: %w", err)
	}
	defer file.Close()
	xzReader, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("ffmpeg: failed to decompress xz: %w", err)
	}
	tarReader := tar.NewReader(xzReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(hdr.Name, "/ffmpeg") || hdr.Name == "ffmpeg" {
			out, err := os.Create(f.binPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, tarReader)
			return err
		}
	}
	return fmt.Errorf("ffmpeg: ffmpeg not found in tar.xz")
}

func (f *FfmpegDependency) Update(progress DownloadProgressCallback) error {
	// TTL 0 means no auto-update; this is a no-op but required by interface
	return f.Install(progress)
}
