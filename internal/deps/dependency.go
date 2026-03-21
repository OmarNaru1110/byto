package deps

import "time"

type ProgressEvent struct {
	Name            string
	DownloadedBytes int64
	TotalBytes      int64
	Percentage      int
}

type ProgressCallback func(event ProgressEvent)
type DownloadProgressCallback func(downloaded, total int64)

type Dependency interface {
	GetName() string
	Path() string
	TTL() time.Duration
	CheckInstalled() (bool, error)
	Install(progress DownloadProgressCallback) error
	Update(progress DownloadProgressCallback) error
	Version() (string, error)
}
