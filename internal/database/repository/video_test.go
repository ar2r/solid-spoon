package repository_test

import (
	"testing"
	"time"

	"github.com/artur/solid-spoon/internal/database/models"
	"github.com/artur/solid-spoon/internal/database/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestVideoRepository_RecordDownload(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	// Create user
	user, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 12345, FirstName: "Test"})

	// Record download
	download := &models.VideoDownload{
		UserID:        user.ID,
		VideoID:       "dQw4w9WgXcQ",
		VideoURL:      "https://youtube.com/watch?v=dQw4w9WgXcQ",
		VideoTitle:    "Test Video",
		Quality:       "720p",
		Compressed:    false,
		FileSizeBytes: 1024000,
		ExecutedAt:    time.Now(),
	}

	err := videoRepo.RecordDownload(download)
	if err != nil {
		t.Fatalf("Failed to record download: %v", err)
	}

	// Verify count
	count, err := videoRepo.GetUserDownloadCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestVideoRepository_GetUserDownloadCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	user, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 12345, FirstName: "Test"})

	// Record multiple downloads
	for i := 0; i < 3; i++ {
		videoRepo.RecordDownload(&models.VideoDownload{
			UserID:     user.ID,
			VideoID:    "video" + string(rune('A'+i)),
			VideoURL:   "https://youtube.com/watch?v=test",
			Quality:    "720p",
			ExecutedAt: time.Now(),
		})
	}

	count, err := videoRepo.GetUserDownloadCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestVideoRepository_GetTotalDownloads(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	user1, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 1, FirstName: "User1"})
	user2, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 2, FirstName: "User2"})

	// Record downloads for different users
	videoRepo.RecordDownload(&models.VideoDownload{
		UserID: user1.ID, VideoID: "v1", VideoURL: "url1", Quality: "720p", ExecutedAt: time.Now(),
	})
	videoRepo.RecordDownload(&models.VideoDownload{
		UserID: user2.ID, VideoID: "v2", VideoURL: "url2", Quality: "1080p", ExecutedAt: time.Now(),
	})

	total, err := videoRepo.GetTotalDownloads()
	if err != nil {
		t.Fatalf("Failed to get total: %v", err)
	}
	if total != 2 {
		t.Errorf("Expected 2 total downloads, got %d", total)
	}
}

func TestVideoRepository_GetPopularVideos(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	user, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 12345, FirstName: "Test"})

	// Record same video multiple times
	for i := 0; i < 3; i++ {
		videoRepo.RecordDownload(&models.VideoDownload{
			UserID:     user.ID,
			VideoID:    "popular",
			VideoURL:   "https://youtube.com/watch?v=popular",
			VideoTitle: "Popular Video",
			Quality:    "720p",
			ExecutedAt: time.Now(),
		})
	}

	// Record another video once
	videoRepo.RecordDownload(&models.VideoDownload{
		UserID:     user.ID,
		VideoID:    "other",
		VideoURL:   "https://youtube.com/watch?v=other",
		VideoTitle: "Other Video",
		Quality:    "720p",
		ExecutedAt: time.Now(),
	})

	popular, err := videoRepo.GetPopularVideos(2)
	if err != nil {
		t.Fatalf("Failed to get popular videos: %v", err)
	}

	if len(popular) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(popular))
	}

	if popular[0].VideoID != "popular" {
		t.Errorf("Expected 'popular' as most downloaded, got %s", popular[0].VideoID)
	}
	if popular[0].DownloadCount != 3 {
		t.Errorf("Expected 3 downloads, got %d", popular[0].DownloadCount)
	}
}
