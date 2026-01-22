package models

import "time"

// User represents a Telegram user stored in database
type User struct {
	ID             int64
	TelegramUserID int64
	Username       string
	FirstName      string
	LastName       string
	LanguageCode   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
