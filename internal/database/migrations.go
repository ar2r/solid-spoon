package database

import (
	"fmt"
	"log"
)

// Migrate runs all database migrations
func (db *DB) Migrate() error {
	log.Printf("[DB] Running migrations...")

	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_user_id INTEGER NOT NULL UNIQUE,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			language_code TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)`,

		// Command stats table
		`CREATE TABLE IF NOT EXISTS command_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			command TEXT NOT NULL,
			executed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_command_stats_user_id ON command_stats(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_command_stats_command ON command_stats(command)`,
		`CREATE INDEX IF NOT EXISTS idx_command_stats_executed_at ON command_stats(executed_at)`,

		// Video downloads table
		`CREATE TABLE IF NOT EXISTS video_downloads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			video_id TEXT NOT NULL,
			video_url TEXT NOT NULL,
			video_title TEXT,
			quality TEXT NOT NULL,
			compressed BOOLEAN DEFAULT 0,
			file_size_bytes INTEGER,
			executed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_video_downloads_user_id ON video_downloads(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_downloads_video_id ON video_downloads(video_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_downloads_executed_at ON video_downloads(executed_at)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	log.Printf("[DB] Migrations completed successfully")
	return nil
}
