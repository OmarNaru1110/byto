package domain

import (
	"context"
	"fmt"
	"sync"
)

type Media struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	TotalBytes        int64             `json:"total_bytes"`
	URL               string            `json:"url"`
	FilePath          string            `json:"file_path"`
	Quality           VideoQuality      `json:"quality"`
	OnlyAudio         bool              `json:"only_audio"`
	Cookies           Cookies           `json:"cookies,omitempty"`
	Status            DownloadStatus    `json:"status"`
	Progress          DownloadProgress  `json:"progress"`
	IsPlaylist        bool              `json:"is_playlist"`
	PlaylistSelection PlaylistSelection `json:"playlist_selection,omitempty"`
	TimeRange         TimeRange         `json:"time_range,omitempty"`
	mu                sync.Mutex
	// Context for cancellation
	Ctx        context.Context    `json:"-"`
	CancelFunc context.CancelFunc `json:"-"`

	OnProgress     func(id string, progress DownloadProgress) `json:"-"`
	OnStatusChange func(id string, status DownloadStatus)     `json:"-"`
	OnTitleChange  func(id string, title string)              `json:"-"`
}

type Cookies struct {
	IsAllowed bool `json:"is_allowed"`
	// either a path to a cookies file or an identifier for browser cookies is used
	Path    string `json:"path,omitempty"`
	Browser string `json:"browser,omitempty"`

	Type CookiesType `json:"type,omitempty"`
}

type CookiesType string

const (
	CookiesTypeFile    CookiesType = "file"
	CookiesTypeBrowser CookiesType = "browser"
)

type PlaylistSelectionType string

const (
	SelectionAll   PlaylistSelectionType = "all"
	SelectionRange PlaylistSelectionType = "range"
	SelectionItems PlaylistSelectionType = "items"
)

type PlaylistSelection struct {
	Type PlaylistSelectionType `json:"type"`

	StartIndex int `json:"start_index"`
	EndIndex   int `json:"end_index"`

	Items string `json:"items"` // Comma-separated list of specific items to download
}

func (ps PlaylistSelection) Validate() error {
	switch ps.Type {
	case SelectionRange:
		if ps.StartIndex < 1 || ps.EndIndex < ps.StartIndex {
			return fmt.Errorf("invalid playlist range: start index %d, end index %d", ps.StartIndex, ps.EndIndex)
		}
		return nil
	case SelectionItems:
		if ps.Items == "" {
			return fmt.Errorf("items selection type requires a non-empty items list")
		}
		return nil
	default:
		return nil
	}
}

type DownloadProgress struct {
	Percentage      int      `json:"percentage"`
	DownloadedBytes int64    `json:"downloaded_bytes"`
	Logs            []string `json:"logs"`
}

func (m *Media) AppendLog(log string) {
	m.mu.Lock()
	m.Progress.Logs = append(m.Progress.Logs, log)
	progress := m.Progress
	id := m.ID
	onProgress := m.OnProgress
	m.mu.Unlock()

	// Emit progress update with new log
	if onProgress != nil {
		go onProgress(id, progress)
	}
}

func (m *Media) SetTitle(title string) {
	m.mu.Lock()
	m.Title = title
	id := m.ID
	onTitleChange := m.OnTitleChange
	m.mu.Unlock()

	if onTitleChange != nil {
		go onTitleChange(id, title)
	}
}

func (m *Media) UpdateProgress(downloaded, total int64, percentage int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Progress.DownloadedBytes = downloaded
	m.TotalBytes = total
	m.Progress.Percentage = percentage

	if m.OnProgress != nil {
		go m.OnProgress(m.ID, m.Progress)
	}
}

func (m *Media) SetStatus(status DownloadStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Status = status
	if m.OnStatusChange != nil {
		go m.OnStatusChange(m.ID, m.Status)
	}
}

func (m *Media) Cancel() {
	m.mu.Lock()
	cancelFunc := m.CancelFunc
	m.mu.Unlock()

	if cancelFunc != nil {
		cancelFunc()
	}
}
