// Package entry defines the core work log entry type and its categories.
package entry

import "time"

// Type identifies the category of a work log entry.
type Type string

const (
	// Highlight marks a positive entry - a win, accomplishment, or good moment.
	Highlight Type = "highlight"
	// Lowlight marks a negative entry - a miss, mistake, or difficult moment.
	Lowlight Type = "lowlight"
)

// Entry is a single work log record.
type Entry struct {
	ID        int64
	Type      Type
	Body      string
	CreatedAt time.Time
}
