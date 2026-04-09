package main

import (
	"byto/internal/builder"
	"byto/internal/command"
	"byto/internal/deps"
	"byto/internal/domain"
	"byto/internal/queue"
	"byto/internal/updater"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	queue         *queue.Queue
	settings      *domain.Setting
	mediaDefaults *domain.MediaDefaults
	updater       *updater.Updater
	depsManager   *deps.Manager
	appConfigDir  string
}

func NewApp() *App {
	userDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting config dir: %v", err)
		return nil
	}
	appConfigDir := filepath.Join(userDir, "byto")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		log.Printf("Error creating config dir: %v", err)
	}

	stateFile := filepath.Join(appConfigDir, "state.json")
	store, err := deps.NewStateStore(stateFile)
	depManager := deps.NewManager(store)

	depManager.Add(deps.NewYTDLPDependency(appConfigDir, time.Hour))
	depManager.Add(deps.NewFfmpegDependency(appConfigDir, 0))
	depManager.Add(deps.NewDenoDependency(appConfigDir, 0))

	return &App{
		queue:         queue.NewQueue(),
		settings:      domain.NewSetting(),
		mediaDefaults: domain.NewMediaDefaults(),
		updater:       updater.NewUpdater(),
		appConfigDir:  appConfigDir,
		depsManager:   depManager,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Byto App started")
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetSettings() *domain.Setting {
	return a.settings
}

func (a *App) UpdateSettings(parallelDownloads int) {
	a.settings.Update(parallelDownloads)
	log.Printf("Settings updated in memory: parallel=%d", parallelDownloads)
}

func (a *App) SaveSettings() error {
	log.Println("Saving settings to file")
	return a.settings.Save()
}

func (a *App) GetSupportedBrowsersForCookies() []string {
	return []string{"brave", "chromium", "edge", "firefox", "opera", "vivaldi", "safari", "whale"}
}

func (a *App) SelectCookiesPath(defaultPath string) string {
	if defaultPath == "" {
		home, _ := os.UserHomeDir()
		defaultPath = filepath.Join(home, "Downloads")
	} else {
		defaultPath = filepath.Dir(defaultPath)
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			home, _ := os.UserHomeDir()
			defaultPath = filepath.Join(home, "Downloads")
			log.Printf("Saved path doesn't exist, falling back to: %s", defaultPath)
		}
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Cookies File",
		DefaultDirectory: defaultPath,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
		},
	})
	if err != nil {
		log.Printf("Error selecting cookies file: %v", err)
		return ""
	}
	return path
}

func (a *App) SelectDownloadFolderWithDefault(defaultPath string) string {
	if defaultPath == "" {
		defaultPath = a.mediaDefaults.DownloadPath
	}
	// Check if the path exists, fallback to Downloads if not
	if defaultPath != "" {
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			home, _ := os.UserHomeDir()
			defaultPath = filepath.Join(home, "Downloads")
			log.Printf("Saved path doesn't exist, falling back to: %s", defaultPath)
		}
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Download Folder",
		DefaultDirectory: defaultPath,
	})
	if err != nil {
		log.Printf("Error selecting folder: %v", err)
		return ""
	}
	return path
}

func (a *App) ShowInFolder(filePath string) {
	log.Printf("ShowInFolder called with path: %s", filePath)
	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		// Check if it's a directory or file
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Error checking path: %v", err)
			return
		}
		if info.IsDir() {
			// Open the folder directly
			cmd = exec.Command("explorer", filePath)
		} else {
			// Select the file in explorer
			cmd = exec.Command("explorer", "/select,", filePath)
		}
	case "darwin":
		cmd = exec.Command("open", "-R", filePath)
	default: // Linux
		cmd = exec.Command("xdg-open", filePath)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Error opening folder: %v", err)
	}
}

func (a *App) GetDefaultDownloadPath() string {
	if a.mediaDefaults != nil {
		return a.mediaDefaults.DownloadPath
	}
	return "./downloads"
}

// GetMediaDefaults returns the current media defaults (quality and download path)
func (a *App) GetMediaDefaults() *domain.MediaDefaults {
	return a.mediaDefaults
}

// UpdateMediaDefaults updates the media defaults for new items
func (a *App) UpdateMediaDefaults(quality string, downloadPath string, onlyAudio bool, cookies domain.Cookies) {
	var q domain.VideoQuality
	switch quality {
	case "360p":
		q = domain.Quality360p
	case "480p":
		q = domain.Quality480p
	case "720p":
		q = domain.Quality720p
	case "1080p":
		q = domain.Quality1080p
	case "1440p":
		q = domain.Quality1440p
	case "2160p":
		q = domain.Quality2160p
	default:
		q = domain.Quality1080p
	}
	a.mediaDefaults.Update(q, downloadPath, onlyAudio, cookies)
	log.Printf("Media defaults updated in memory: quality=%s, path=%s, onlyAudio=%v, cookies=%v", quality, downloadPath, onlyAudio, cookies)
}

// SaveMediaDefaults saves the media defaults to file
func (a *App) SaveMediaDefaults() error {
	log.Println("Saving media defaults to file")
	return a.mediaDefaults.Save()
}

func (a *App) AddToQueue(url string, quality string, customPath string, onlyAudio bool, isPlaylist bool, playlistSelection domain.PlaylistSelection, cookies domain.Cookies, timerange domain.TimeRange) string {
	id := uuid.New().String()
	log.Printf("Adding to queue: %s with id: %s", url, id)

	filePath := a.mediaDefaults.DownloadPath
	if customPath != "" {
		filePath = customPath
	}

	// Convert quality string to VideoQuality
	var q domain.VideoQuality
	switch quality {
	case "360p":
		q = domain.Quality360p
	case "480p":
		q = domain.Quality480p
	case "720p":
		q = domain.Quality720p
	case "1080p":
		q = domain.Quality1080p
	case "1440p":
		q = domain.Quality1440p
	case "2160p":
		q = domain.Quality2160p
	default:
		q = domain.Quality1080p
	}

	a.queue.Add(&domain.Media{
		ID:                id,
		URL:               url,
		Title:             "Pending...",
		FilePath:          filePath,
		Quality:           q,
		OnlyAudio:         onlyAudio,
		Status:            domain.Pending,
		IsPlaylist:        isPlaylist,
		PlaylistSelection: playlistSelection,
		Cookies:           cookies,
		TimeRange:         timerange,
		Progress: domain.DownloadProgress{
			Percentage:      0,
			DownloadedBytes: 0,
			Logs:            []string{},
		},
	})
	return id
}

func (a *App) RemoveFromQueue(id string) error {
	log.Printf("Removing from queue: %s", id)
	a.PauseSingleDownload(id)
	return a.queue.Remove(id)
}

func (a *App) GetQueue() []*domain.Media {
	return a.queue.GetAll()
}

func (a *App) StartDownloads() {
	log.Println("Starting downloads")
	if a.settings == nil {
		a.settings = domain.NewSetting()
	}

	queueItems := a.queue.GetAll()
	semaphore := make(chan struct{}, a.settings.ParallelDownloads)

	// Collect pending/failed/paused items in order
	var pendingItems []*domain.Media
	for _, media := range queueItems {
		if media.Status == domain.Pending || media.Status == domain.Failed || media.Status == domain.Paused {
			// Create a context for cancellation
			ctx, cancelFunc := context.WithCancel(context.Background())
			media.Ctx = ctx
			media.CancelFunc = cancelFunc

			// Attach callbacks - capture the media ID to avoid closure issues
			mediaID := media.ID
			media.OnProgress = func(id string, progress domain.DownloadProgress) {
				// Get the current media state to include title
				currentMedia, err := a.queue.Get(id)
				title := "Pending..."
				totalBytes := int64(0)
				if err == nil && currentMedia != nil {
					title = currentMedia.Title
					totalBytes = currentMedia.TotalBytes
				}
				runtime.EventsEmit(a.ctx, "download_progress", map[string]interface{}{
					"id":          id,
					"title":       title,
					"total_bytes": totalBytes,
					"progress":    progress,
				})
			}

			media.OnStatusChange = func(id string, status domain.DownloadStatus) {
				runtime.EventsEmit(a.ctx, "download_status", map[string]interface{}{
					"id":     id,
					"status": status,
				})
			}

			media.OnTitleChange = func(id string, title string) {
				runtime.EventsEmit(a.ctx, "download_title", map[string]interface{}{
					"id":    id,
					"title": title,
				})
			}

			pendingItems = append(pendingItems, media)
			_ = mediaID // used in callbacks via closure
		}
	}

	// Use a job channel to maintain FIFO order
	jobs := make(chan *domain.Media, len(pendingItems))
	for _, media := range pendingItems {
		jobs <- media
	}
	close(jobs)

	// Start workers that pull from the job channel in order
	for i := 0; i < a.settings.ParallelDownloads; i++ {
		go func() {
			for m := range jobs {
				semaphore <- struct{}{}

				m.SetStatus(domain.InProgress)
				log.Printf("Processing item: %s", m.URL)

				// Initialize builder - use media's own FilePath and Quality
				b := builder.NewYTDLPBuilderWithDeps(a.depsManager)
				b = b.URL(m.URL).
					DownloadPath(m.FilePath).
					SafeFilenames()
				if m.Cookies.IsAllowed {
					switch m.Cookies.Type {
					case domain.CookiesTypeFile:
						b = b.Cookies(m.Cookies.Path)
					case domain.CookiesTypeBrowser:
						browser := strings.ToLower(m.Cookies.Browser)
						b = b.CookiesFromBrowser(browser)
					}
				}
				if m.OnlyAudio {
					b = b.Audio()
				} else {
					b = b.Video(m.Quality)
				}
				if m.IsPlaylist {
					b = b.Playlist(m.PlaylistSelection)
				}
				if m.TimeRange.IsAllowed && m.TimeRange.Validate() == nil {
					b = b.DownloadSection(m.TimeRange)
				}
				cmd := &command.DownloadCommand{
					Builder: b,
				}

				if err := cmd.Execute(m); err != nil {
					if err == context.Canceled {
						// Download was paused, set status to Paused
						m.SetStatus(domain.Paused)
						log.Printf("Download paused for %s", m.URL)
					} else {
						m.SetStatus(domain.Failed)
						log.Printf("Download failed for %s: %v", m.URL, err)
						m.AppendLog(fmt.Sprintf("Download failed: %v", err))
					}
				} else {
					log.Printf("Download completed: %s", m.URL)
				}

				<-semaphore
			}
		}()
	}
}

func (a *App) PauseDownloads() {
	log.Println("Pausing all downloads")
	queueItems := a.queue.GetAll()

	for _, media := range queueItems {
		if media.Status == domain.InProgress {
			media.Cancel()
		}
	}
}

func (a *App) StartSingleDownload(id string) {
	log.Printf("Starting single download: %s", id)
	media, err := a.queue.Get(id)
	if err != nil {
		log.Printf("Error getting media from queue: %v", err)
		return
	}

	if media.Status != domain.Pending && media.Status != domain.Failed && media.Status != domain.Paused {
		log.Printf("Media %s is not in a startable state (status: %d)", id, media.Status)
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	media.Ctx = ctx
	media.CancelFunc = cancelFunc

	media.OnProgress = func(id string, progress domain.DownloadProgress) {
		currentMedia, err := a.queue.Get(id)
		title := "Pending..."
		totalBytes := int64(0)
		if err == nil && currentMedia != nil {
			title = currentMedia.Title
			totalBytes = currentMedia.TotalBytes
		}
		runtime.EventsEmit(a.ctx, "download_progress", map[string]interface{}{
			"id":          id,
			"title":       title,
			"total_bytes": totalBytes,
			"progress":    progress,
		})
	}

	media.OnStatusChange = func(id string, status domain.DownloadStatus) {
		runtime.EventsEmit(a.ctx, "download_status", map[string]interface{}{
			"id":     id,
			"status": status,
		})
	}

	media.OnTitleChange = func(id string, title string) {
		runtime.EventsEmit(a.ctx, "download_title", map[string]interface{}{
			"id":    id,
			"title": title,
		})
	}

	go func() {
		media.SetStatus(domain.InProgress)
		log.Printf("Processing item: %s", media.URL)

		b := builder.NewYTDLPBuilderWithDeps(a.depsManager)
		b = b.URL(media.URL).
			DownloadPath(media.FilePath).
			SafeFilenames()
		if media.Cookies.IsAllowed {
			switch media.Cookies.Type {
			case domain.CookiesTypeFile:
				b = b.Cookies(media.Cookies.Path)
			case domain.CookiesTypeBrowser:
				browser := strings.ToLower(media.Cookies.Browser)
				b = b.CookiesFromBrowser(browser)
			}
		}
		if media.OnlyAudio {
			b = b.Audio()
		} else {
			b = b.Video(media.Quality)
		}
		if media.IsPlaylist {
			b = b.Playlist(media.PlaylistSelection)
		}
		if media.TimeRange.IsAllowed && media.TimeRange.Validate() == nil {
			b = b.DownloadSection(media.TimeRange)
		}
		cmd := &command.DownloadCommand{
			Builder: b,
		}

		if err := cmd.Execute(media); err != nil {
			if err == context.Canceled {
				media.SetStatus(domain.Paused)
				log.Printf("Download paused for %s", media.URL)
			} else {
				media.SetStatus(domain.Failed)
				log.Printf("Download failed for %s: %v", media.URL, err)
				media.AppendLog(fmt.Sprintf("Download failed: %v", err))
			}
		} else {
			log.Printf("Download completed: %s", media.URL)
		}
	}()
}

func (a *App) PauseSingleDownload(id string) {
	log.Printf("Pausing single download: %s", id)
	media, err := a.queue.Get(id)
	if err != nil {
		log.Printf("Error getting media from queue: %v", err)
		return
	}

	if media.Status == domain.InProgress {
		media.Cancel()
	}
}

func (a *App) GetAppVersion() string {
	return a.updater.GetAppVersion()
}

func (a *App) CheckAppUpdate() updater.UpdateResult {
	log.Println("Checking for app updates...")
	result := a.updater.CheckAppUpdate()
	if result.Success {
		log.Printf("App update check: current=%s, latest=%s, hasUpdate=%v",
			result.CurrentVersion, result.LatestVersion, result.HasUpdate)
	} else {
		log.Printf("App update check failed: %s", result.Message)
	}
	return result
}

func (a *App) DownloadAppUpdate(downloadURL string) (string, error) {
	log.Printf("Downloading app update from: %s", downloadURL)

	// Emit progress events
	progressCallback := func(downloaded, total int64) {
		var percentage float64
		if total > 0 {
			percentage = float64(downloaded) / float64(total) * 100
		}
		runtime.EventsEmit(a.ctx, "update_download_progress", map[string]interface{}{
			"downloaded": downloaded,
			"total":      total,
			"percentage": percentage,
		})
	}

	installerPath, err := a.updater.DownloadAppUpdate(downloadURL, progressCallback)
	if err != nil {
		log.Printf("Failed to download update: %v", err)
		return "", err
	}

	log.Printf("Update downloaded to: %s", installerPath)
	return installerPath, nil
}

func (a *App) LaunchInstaller(installerPath string) error {
	log.Printf("Launching installer: %s", installerPath)
	err := a.updater.LaunchInstaller(installerPath)
	if err != nil {
		log.Printf("Failed to launch installer: %v", err)
		return err
	}
	a.ShutDown()
	return nil
}

func (a *App) PerformFullUpdate() map[string]interface{} {
	log.Println("Performing full update check...")

	// Step 1: Update dependencies (yt-dlp, ffmpeg) via deps
	runtime.EventsEmit(a.ctx, "update_status", map[string]interface{}{
		"step":    "deps",
		"message": "Updating dependencies...",
	})
	depsErr := a.depsManager.Bootstrap(nil)
	depsSuccess := depsErr == nil
	depsMessage := "Dependencies up to date"
	if depsErr != nil {
		depsMessage = depsErr.Error()
	}

	// Step 2: Check for app updates
	runtime.EventsEmit(a.ctx, "update_status", map[string]interface{}{
		"step":    "app_check",
		"message": "Checking for app updates...",
	})
	appResult := a.updater.CheckAppUpdate()

	return map[string]interface{}{
		"deps": map[string]interface{}{
			"success": depsSuccess,
			"message": depsMessage,
		},
		"app": map[string]interface{}{
			"success":         appResult.Success,
			"message":         appResult.Message,
			"current_version": appResult.CurrentVersion,
			"latest_version":  appResult.LatestVersion,
			"has_update":      appResult.HasUpdate,
			"changelog":       appResult.Changelog,
			"download_url":    appResult.DownloadURL,
		},
	}
}

func (a *App) ShutDown() {
	log.Println("Shutting down Byto App")
	runtime.Quit(a.ctx)
}

func (a *App) SetupDependencies() {
	err := a.depsManager.Bootstrap(func(event deps.ProgressEvent) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "dependency_progress", event)
		}
	})

	if err != nil {
		log.Printf("Error bootstrapping deps: %v", err)
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "dependency_bootstrap_failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	} else if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "dependency_bootstrap_complete", nil)
	}
}

func (a *App) CheckDependencies() []deps.DependencyState {
	return a.depsManager.DependenciesState()
}
