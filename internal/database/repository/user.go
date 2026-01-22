package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/artur/solid-spoon/internal/database/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserRepository handles user data persistence
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// UpsertFromTelegram creates or updates user from Telegram user object
func (r *UserRepository) UpsertFromTelegram(tgUser *tgbotapi.User) (*models.User, error) {
	if tgUser == nil {
		return nil, fmt.Errorf("telegram user is nil")
	}

	now := time.Now()

	// Try to insert, on conflict update
	query := `
		INSERT INTO users (telegram_user_id, username, first_name, last_name, language_code, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(telegram_user_id) DO UPDATE SET
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			language_code = excluded.language_code,
			updated_at = excluded.updated_at
	`

	_, err := r.db.Exec(query,
		tgUser.ID,
		tgUser.UserName,
		tgUser.FirstName,
		tgUser.LastName,
		tgUser.LanguageCode,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	// Fetch the user to return
	return r.GetByTelegramID(tgUser.ID)
}

// GetByTelegramID retrieves user by Telegram user ID
func (r *UserRepository) GetByTelegramID(telegramUserID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_user_id, username, first_name, last_name, language_code, created_at, updated_at
		FROM users
		WHERE telegram_user_id = ?
	`

	user := &models.User{}
	var username, firstName, lastName, languageCode sql.NullString

	err := r.db.QueryRow(query, telegramUserID).Scan(
		&user.ID,
		&user.TelegramUserID,
		&username,
		&firstName,
		&lastName,
		&languageCode,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Username = username.String
	user.FirstName = firstName.String
	user.LastName = lastName.String
	user.LanguageCode = languageCode.String

	return user, nil
}

// GetTotalUsers returns total number of unique users
func (r *UserRepository) GetTotalUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
