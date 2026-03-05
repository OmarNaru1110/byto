package builder

import (
	"byto/internal/domain"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type YTDLPBuilder struct {
	args      []string
	ytdlpPath string
}

func NewYTDLPBuilder() *YTDLPBuilder {
	return &YTDLPBuilder{
		ytdlpPath: findYtDlpPath(),
	}
}

// YtDlpPath sets the path to the yt-dlp executable (e.g. from deps manager).
func (y *YTDLPBuilder) YtDlpPath(path string) *YTDLPBuilder {
	if path != "" {
		y.ytdlpPath = path
	}
	return y
}

func findYtDlpPath() string {
	execPath, err := os.Executable()
	if err == nil {
		appDir := filepath.Dir(execPath)
		ytdlpName := "yt-dlp"
		if runtime.GOOS == "windows" {
			ytdlpName = "yt-dlp.exe"
		}
		bundledPath := filepath.Join(appDir, ytdlpName)
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath
		}
	}

	globalPath, err := exec.LookPath("yt-dlp")
	if err == nil {
		return globalPath
	}

	return "yt-dlp"
}

// GetYtDlpPath returns the path to yt-dlp executable
func (y *YTDLPBuilder) GetYtDlpPath() string {
	return y.ytdlpPath
}

// "[byto:title] %(info.title)s [byto:downloaded_bytes] %(progress.downloaded_bytes)d [byto:total_bytes] %(progress.total_bytes)d"
func (y *YTDLPBuilder) ProgressTemplate(template string) *YTDLPBuilder {
	y.args = append(y.args, "--progress-template", template)
	return y
}

// Newline forces a newline character at the end of each progress line
func (y *YTDLPBuilder) Newline() *YTDLPBuilder {
	y.args = append(y.args, "--newline")
	return y
}

func (y *YTDLPBuilder) Video(quality domain.VideoQuality) *YTDLPBuilder {
	// Use format selection with fallback to best available
	// "bestvideo[height<=X]+bestaudio/best[height<=X]/best" means:
	// 1. Try best video up to X height + best audio
	// 2. Fall back to combined best up to X height
	// 3. Fall back to absolute best available
	switch quality {
	case domain.Quality360p:
		y.args = append(y.args, "-f", "bestvideo[height<=360]+bestaudio/best[height<=360]/best")
	case domain.Quality480p:
		y.args = append(y.args, "-f", "bestvideo[height<=480]+bestaudio/best[height<=480]/best")
	case domain.Quality720p:
		y.args = append(y.args, "-f", "bestvideo[height<=720]+bestaudio/best[height<=720]/best")
	case domain.Quality1080p:
		y.args = append(y.args, "-f", "bestvideo[height<=1080]+bestaudio/best[height<=1080]/best")
	case domain.Quality1440p:
		y.args = append(y.args, "-f", "bestvideo[height<=1440]+bestaudio/best[height<=1440]/best")
	case domain.Quality2160p:
		y.args = append(y.args, "-f", "bestvideo[height<=2160]+bestaudio/best[height<=2160]/best")
	default:
		y.args = append(y.args, "-f", "bestvideo+bestaudio/best")
	}
	return y
}

func (y *YTDLPBuilder) Audio() *YTDLPBuilder {
	y.args = append(y.args, "-f", "bestaudio/best")
	return y
}

func (y *YTDLPBuilder) DownloadPath(path string) *YTDLPBuilder {
	y.args = append(y.args, "-o", path+"/%(title).100s.%(ext)s")
	return y
}

// SafeFilenames adds platform-appropriate filename restrictions
func (y *YTDLPBuilder) SafeFilenames() *YTDLPBuilder {
	if runtime.GOOS == "windows" {
		y.args = append(y.args, "--windows-filenames")
	} else {
		y.args = append(y.args, "--restrict-filenames")
	}
	return y
}

func (y *YTDLPBuilder) URL(url string) *YTDLPBuilder {
	y.args = append(y.args, url)
	return y
}

func (y *YTDLPBuilder) Update() *YTDLPBuilder {
	y.args = append(y.args, "--update")
	return y
}

func (y *YTDLPBuilder) Playlist(playlist domain.PlaylistSelection) *YTDLPBuilder {
	if err := playlist.Validate(); err != nil {
		return y
	}
	switch playlist.Type {
	case domain.SelectionRange:
		y.args = append(y.args, "--playlist-items", fmt.Sprintf("%d-%d", playlist.StartIndex, playlist.EndIndex))
	case domain.SelectionItems:
		y.args = append(y.args, "--playlist-items", playlist.Items)
	}
	return y
}

func (y *YTDLPBuilder) Build() []string {
	return y.args
}
