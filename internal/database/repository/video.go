package repository

import (
	"database/sql"
	"fmt"

	"github.com/artur/solid-spoon/internal/database/models"
)

// VideoRepository handles video download persistence
type VideoRepository struct {
	db *sql.DB
}

// NewVideoRepository creates a new VideoRepository
func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// RecordDownload records a video download
func (r *VideoRepository) RecordDownload(download *models.VideoDownload) error {
	query := `
		INSERT INTO video_downloads
		(user_id, video_id, video_url, video_title, quality, compressed, file_size_bytes, executed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		download.UserID,
		download.VideoID,
		download.VideoURL,
		download.VideoTitle,
		download.Quality,
		download.Compressed,
		download.FileSizeBytes,
		download.ExecutedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record video download: %w", err)
	}

	return nil
}

// GetUserDownloadCount returns total downloads for a user
func (r *VideoRepository) GetUserDownloadCount(userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM video_downloads WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetTotalDownloads returns total downloads by all users
func (r *VideoRepository) GetTotalDownloads() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM video_downloads").Scan(&count)
	return count, err
}

// PopularVideo represents a video with download count
type PopularVideo struct {
	VideoID       string
	VideoTitle    string
	DownloadCount int64
}

// GetPopularVideos returns most downloaded videos (top N)
func (r *VideoRepository) GetPopularVideos(limit int) ([]PopularVideo, error) {
	query := `
		SELECT video_id, video_title, COUNT(*) as download_count
		FROM video_downloads
		GROUP BY video_id
		ORDER BY download_count DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular videos: %w", err)
	}
	defer rows.Close()

	var videos []PopularVideo
	for rows.Next() {
		var video PopularVideo
		var title sql.NullString
		if err := rows.Scan(&video.VideoID, &title, &video.DownloadCount); err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		video.VideoTitle = title.String
		videos = append(videos, video)
	}

	return videos, rows.Err()
}
