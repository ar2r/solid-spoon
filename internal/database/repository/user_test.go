package repository_test

import (
	"database/sql"
	"testing"

	"github.com/artur/solid-spoon/internal/database"
	"github.com/artur/solid-spoon/internal/database/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	dbWrapper := &database.DB{DB: db}
	if err := dbWrapper.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestUserRepository_UpsertFromTelegram(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	tgUser := &tgbotapi.User{
		ID:           12345,
		FirstName:    "Test",
		LastName:     "User",
		UserName:     "testuser",
		LanguageCode: "en",
	}

	// First insert
	user1, err := repo.UpsertFromTelegram(tgUser)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	if user1 == nil {
		t.Fatal("Expected user to be returned")
	}

	if user1.TelegramUserID != 12345 {
		t.Errorf("Expected telegram_user_id 12345, got %d", user1.TelegramUserID)
	}

	if user1.FirstName != "Test" {
		t.Errorf("Expected first_name 'Test', got %s", user1.FirstName)
	}

	// Update same user
	tgUser.FirstName = "Updated"
	user2, err := repo.UpsertFromTelegram(tgUser)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if user2.ID != user1.ID {
		t.Errorf("User ID should remain same, got %d vs %d", user2.ID, user1.ID)
	}

	if user2.FirstName != "Updated" {
		t.Errorf("Expected first_name 'Updated', got %s", user2.FirstName)
	}
}

func TestUserRepository_UpsertFromTelegram_NilUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	_, err := repo.UpsertFromTelegram(nil)
	if err == nil {
		t.Error("Expected error for nil user")
	}
}

func TestUserRepository_GetByTelegramID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	// Get non-existent user
	user, err := repo.GetByTelegramID(99999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if user != nil {
		t.Errorf("Expected nil for non-existent user")
	}

	// Insert and retrieve
	tgUser := &tgbotapi.User{ID: 12345, FirstName: "Test"}
	_, err = repo.UpsertFromTelegram(tgUser)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	user, err = repo.GetByTelegramID(12345)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if user == nil || user.TelegramUserID != 12345 {
		t.Errorf("Failed to retrieve correct user")
	}
}

func TestUserRepository_GetTotalUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	// Initially zero
	count, err := repo.GetTotalUsers()
	if err != nil {
		t.Fatalf("Failed to get total users: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users, got %d", count)
	}

	// Add users
	repo.UpsertFromTelegram(&tgbotapi.User{ID: 1, FirstName: "User1"})
	repo.UpsertFromTelegram(&tgbotapi.User{ID: 2, FirstName: "User2"})

	count, err = repo.GetTotalUsers()
	if err != nil {
		t.Fatalf("Failed to get total users: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}
