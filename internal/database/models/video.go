package models

import "time"

// VideoDownload represents a video download record
type VideoDownload struct {
	ID            int64
	UserID        int64
	VideoID       string
	VideoURL      string
	VideoTitle    string
	Quality       string
	Compressed    bool
	FileSizeBytes int64
	ExecutedAt    time.Time
}
