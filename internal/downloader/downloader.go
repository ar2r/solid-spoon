package downloader

// Downloader interface for downloading videos from various sources
type Downloader interface {
	Download(videoID string) (filePath string, err error)
	DownloadWithQuality(videoID string, quality Quality) (filePath string, err error)
	DownloadWithQualityInfo(videoID string, quality Quality) (*VideoInfo, error)
	GetAvailableFormats(videoID string) ([]VideoFormat, error)
}
