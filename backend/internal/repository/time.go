package repository

import "time"

// parseTime parses a timestamp string from SQLite, trying RFC3339 first
// (the format used by modernc.org/sqlite) then the bare datetime format.
func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	t, _ := time.Parse(time.DateTime, s)
	return t
}
