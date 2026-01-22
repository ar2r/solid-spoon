package models

import "time"

// CommandStat represents a command execution record
type CommandStat struct {
	ID         int64
	UserID     int64
	Command    string
	ExecutedAt time.Time
}
