package builder

import (
	"byto/internal/deps"
	"byto/internal/domain"
	"fmt"
	"runtime"
	"strings"
)

type YTDLPState struct {
	Newline            bool
	Update             bool
	IgnoreErrors       bool
	WriteSubtitles     bool
	WriteAutoSubtitles bool

	Format                string
	ProgressTemplate      string
	HasProgressTemplate   bool
	DownloadPath          string
	HasDownloadPath       bool
	URL                   string
	HasURL                bool
	Cookies               string
	HasCookies            bool
	CookiesFromBrowser    string
	HasCookiesFromBrowser bool
	JsRuntimes            []string
	ExtractorArgs         []string
	DownloadSections      []string
	SleepSubtitles        int
	HasSleepSubtitles     bool
	SubtitlesFormat       bool
	SubtitlesLanguages    []string
	SafeFilenames         bool
	Playlist              domain.PlaylistSelection
	HasPlaylist           bool
}

type YTDLPBuilder struct {
	ytdlpPath string
	state     YTDLPState
}

func NewYTDLPBuilder() *YTDLPBuilder {
	return NewYTDLPBuilderWithDeps(nil)
}

// NewYTDLPBuilderWithDeps builds a YTDLPBuilder using the yt-dlp path from the deps manager only.
func NewYTDLPBuilderWithDeps(m *deps.Manager) *YTDLPBuilder {
	return &YTDLPBuilder{
		ytdlpPath: ytDlpPathFromManager(m),
	}
}

func ytDlpPathFromManager(m *deps.Manager) string {
	if m == nil {
		return ""
	}
	if p, ok := m.GetPath("yt-dlp"); ok {
		return p
	}
	return ""
}

// YtDlpPath overrides the yt-dlp executable path (e.g. for tests).
func (y *YTDLPBuilder) YtDlpPath(path string) *YTDLPBuilder {
	if path != "" {
		y.ytdlpPath = path
	}
	return y
}

// GetYtDlpPath returns the path to yt-dlp executable
func (y *YTDLPBuilder) GetYtDlpPath() string {
	return y.ytdlpPath
}

// "[byto:title] %(info.title)s [byto:downloaded_bytes] %(progress.downloaded_bytes)d [byto:total_bytes] %(progress.total_bytes)d"
func (y *YTDLPBuilder) ProgressTemplate(template string) *YTDLPBuilder {
	y.state.ProgressTemplate = template
	y.state.HasProgressTemplate = true
	return y
}

// Newline forces a newline character at the end of each progress line
func (y *YTDLPBuilder) Newline() *YTDLPBuilder {
	y.state.Newline = true
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
		y.state.Format = "bestvideo[height<=360]+bestaudio/best[height<=360]/best"

	case domain.Quality480p:
		y.state.Format = "bestvideo[height<=480]+bestaudio/best[height<=480]/best"

	case domain.Quality720p:
		y.state.Format = "bestvideo[height<=720]+bestaudio/best[height<=720]/best"

	case domain.Quality1080p:
		y.state.Format = "bestvideo[height<=1080]+bestaudio/best[height<=1080]/best"

	case domain.Quality1440p:
		y.state.Format = "bestvideo[height<=1440]+bestaudio/best[height<=1440]/best"

	case domain.Quality2160p:
		y.state.Format = "bestvideo[height<=2160]+bestaudio/best[height<=2160]/best"

	default:
		y.state.Format = "bestvideo+bestaudio/best"
	}

	return y
}

func (y *YTDLPBuilder) Audio() *YTDLPBuilder {
	y.state.Format = "bestaudio/best"
	return y
}

func (y *YTDLPBuilder) DownloadPath(path string) *YTDLPBuilder {
	y.state.DownloadPath = path
	y.state.HasDownloadPath = true
	return y
}

// SafeFilenames adds platform-appropriate filename restrictions
func (y *YTDLPBuilder) SafeFilenames() *YTDLPBuilder {
	y.state.SafeFilenames = true
	return y
}

func (y *YTDLPBuilder) URL(url string) *YTDLPBuilder {
	y.state.URL = url
	y.state.HasURL = true
	return y
}

func (y *YTDLPBuilder) Update() *YTDLPBuilder {
	y.state.Update = true
	return y
}

func (y *YTDLPBuilder) Playlist(playlist domain.PlaylistSelection) *YTDLPBuilder {
	if err := playlist.Validate(); err != nil {
		return y
	}

	y.state.Playlist = playlist
	y.state.HasPlaylist = true
	return y
}

func (y *YTDLPBuilder) Cookies(path string) *YTDLPBuilder {
	y.state.Cookies = path
	y.state.HasCookies = true
	return y
}

func (y *YTDLPBuilder) CookiesFromBrowser(browser string) *YTDLPBuilder {
	y.state.CookiesFromBrowser = browser
	y.state.HasCookiesFromBrowser = true
	return y
}

func (y *YTDLPBuilder) JsRuntimes(jsRuntime string) *YTDLPBuilder {
	y.state.JsRuntimes = append(y.state.JsRuntimes, jsRuntime)
	return y
}

func (y *YTDLPBuilder) ExtractorArgs(args string) *YTDLPBuilder {
	y.state.ExtractorArgs = append(y.state.ExtractorArgs, args)
	return y
}

func (y *YTDLPBuilder) DownloadSection(timerange domain.TimeRange) *YTDLPBuilder {
	y.state.DownloadSections = append(
		y.state.DownloadSections,
		fmt.Sprintf("*%s-%s", timerange.Start, timerange.End),
	)

	return y
}

func (y *YTDLPBuilder) IgnoreErrors() *YTDLPBuilder {
	y.state.IgnoreErrors = true
	return y
}

func (y *YTDLPBuilder) WriteSubtitles() *YTDLPBuilder {
	y.state.WriteSubtitles = true
	return y
}

func (y *YTDLPBuilder) WriteAutoSubtitles() *YTDLPBuilder {
	y.state.WriteAutoSubtitles = true
	return y
}

func (y *YTDLPBuilder) SleepSubtitles(seconds int) *YTDLPBuilder {
	y.state.SleepSubtitles = seconds
	y.state.HasSleepSubtitles = true
	return y
}

func (y *YTDLPBuilder) SubtitlesFormat() *YTDLPBuilder {
	y.state.SubtitlesFormat = true
	return y
}

func (y *YTDLPBuilder) SubtitlesLanguages(languageCodes []string) *YTDLPBuilder {
	if len(languageCodes) > 0 {
		y.state.SubtitlesLanguages = languageCodes
	}
	return y
}

func (y *YTDLPBuilder) Build() []string {
	var args []string

	if y.state.IgnoreErrors {
		args = append(args, "--ignore-errors")
	}

	if y.state.Newline {
		args = append(args, "--newline")
	}

	if y.state.Update {
		args = append(args, "--update")
	}

	if y.state.WriteSubtitles {
		args = append(args, "--write-subs")
	}

	if y.state.WriteAutoSubtitles {
		args = append(args, "--write-auto-subs")
	}

	if y.state.Format != "" {
		args = append(args, "-f", y.state.Format)
	}

	if y.state.HasCookies {
		args = append(args, "--cookies", y.state.Cookies)
	}

	if y.state.HasCookiesFromBrowser {
		args = append(args, "--cookies-from-browser", y.state.CookiesFromBrowser)
	}

	if y.state.HasPlaylist {
		switch y.state.Playlist.Type {
		case domain.SelectionRange:
			args = append(
				args,
				"--playlist-items",
				fmt.Sprintf(
					"%d-%d",
					y.state.Playlist.StartIndex,
					y.state.Playlist.EndIndex,
				),
			)

		case domain.SelectionItems:
			args = append(
				args,
				"--playlist-items",
				y.state.Playlist.Items,
			)
		}
	}

	if y.state.HasURL {
		args = append(args, y.state.URL)
	}

	if y.state.HasDownloadPath {
		args = append(
			args,
			"-o",
			y.state.DownloadPath+"/%(title).100s.%(ext)s",
		)
	}

	if y.state.HasProgressTemplate {
		args = append(args, "--progress-template", y.state.ProgressTemplate)
	}

	for _, jsRuntime := range y.state.JsRuntimes {
		args = append(args, "--js-runtimes", jsRuntime)
	}

	for _, extractorArgs := range y.state.ExtractorArgs {
		args = append(args, "--extractor-args", extractorArgs)
	}

	for _, section := range y.state.DownloadSections {
		args = append(args, "--download-sections", section)
	}

	if y.state.HasSleepSubtitles {
		args = append(
			args,
			"--sleep-subtitles",
			fmt.Sprintf("%d", y.state.SleepSubtitles),
		)
	}

	if y.state.SubtitlesFormat {
		args = append(args, "--sub-format", "srt/vtt/best")
	}

	if len(y.state.SubtitlesLanguages) > 0 {
		args = append(
			args,
			"--sub-langs",
			strings.Join(y.state.SubtitlesLanguages, ","),
		)
	}

	if y.state.SafeFilenames {
		if runtime.GOOS == "windows" {
			args = append(args, "--windows-filenames")
		} else {
			args = append(args, "--restrict-filenames")
		}
	}

	return args
}
