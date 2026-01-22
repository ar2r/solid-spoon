package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// CommandCount represents command usage statistics
type CommandCount struct {
	Command string
	Count   int64
}

// StatsRepository handles command statistics persistence
type StatsRepository struct {
	db *sql.DB
}

// NewStatsRepository creates a new StatsRepository
func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// RecordCommand records a command execution
func (r *StatsRepository) RecordCommand(userID int64, command string) error {
	query := `INSERT INTO command_stats (user_id, command, executed_at) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, userID, command, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record command: %w", err)
	}
	return nil
}

// GetCommandCount returns total commands executed by a user
func (r *StatsRepository) GetCommandCount(userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM command_stats WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetTotalCommands returns total commands executed by all users
func (r *StatsRepository) GetTotalCommands() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM command_stats").Scan(&count)
	return count, err
}

// GetPopularCommands returns most popular commands (top N)
func (r *StatsRepository) GetPopularCommands(limit int) ([]CommandCount, error) {
	query := `
		SELECT command, COUNT(*) as count
		FROM command_stats
		GROUP BY command
		ORDER BY count DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular commands: %w", err)
	}
	defer rows.Close()

	var results []CommandCount
	for rows.Next() {
		var item CommandCount
		if err := rows.Scan(&item.Command, &item.Count); err != nil {
			return nil, fmt.Errorf("failed to scan command count: %w", err)
		}
		results = append(results, item)
	}

	return results, rows.Err()
}
