package repository_test

import (
	"testing"

	"github.com/artur/solid-spoon/internal/database/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestStatsRepository_RecordCommand(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Create user first
	tgUser := &tgbotapi.User{ID: 12345, FirstName: "Test"}
	user, err := userRepo.UpsertFromTelegram(tgUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Record command
	err = statsRepo.RecordCommand(user.ID, "start")
	if err != nil {
		t.Fatalf("Failed to record command: %v", err)
	}

	// Verify count
	count, err := statsRepo.GetCommandCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestStatsRepository_GetCommandCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Create user
	user, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 12345, FirstName: "Test"})

	// Record multiple commands
	statsRepo.RecordCommand(user.ID, "start")
	statsRepo.RecordCommand(user.ID, "youtube")
	statsRepo.RecordCommand(user.ID, "start")

	count, err := statsRepo.GetCommandCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestStatsRepository_GetTotalCommands(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Create users
	user1, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 1, FirstName: "User1"})
	user2, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 2, FirstName: "User2"})

	// Record commands
	statsRepo.RecordCommand(user1.ID, "start")
	statsRepo.RecordCommand(user2.ID, "start")
	statsRepo.RecordCommand(user1.ID, "youtube")

	total, err := statsRepo.GetTotalCommands()
	if err != nil {
		t.Fatalf("Failed to get total: %v", err)
	}
	if total != 3 {
		t.Errorf("Expected 3 total commands, got %d", total)
	}
}

func TestStatsRepository_GetPopularCommands(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	user, _ := userRepo.UpsertFromTelegram(&tgbotapi.User{ID: 12345, FirstName: "Test"})

	// Record commands with different frequencies
	statsRepo.RecordCommand(user.ID, "start")
	statsRepo.RecordCommand(user.ID, "start")
	statsRepo.RecordCommand(user.ID, "start")
	statsRepo.RecordCommand(user.ID, "youtube")
	statsRepo.RecordCommand(user.ID, "youtube")
	statsRepo.RecordCommand(user.ID, "help")

	popular, err := statsRepo.GetPopularCommands(2)
	if err != nil {
		t.Fatalf("Failed to get popular commands: %v", err)
	}

	if len(popular) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(popular))
	}

	if popular[0].Command != "start" {
		t.Errorf("Expected 'start' as most popular, got %s", popular[0].Command)
	}
	if popular[0].Count != 3 {
		t.Errorf("Expected count 3 for start, got %d", popular[0].Count)
	}

	if popular[1].Command != "youtube" {
		t.Errorf("Expected 'youtube' as second, got %s", popular[1].Command)
	}
}
