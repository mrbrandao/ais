package memx

import "time"

// nowString returns the current UTC time in ISO 8601 format.
// Extracted so tests can verify time-dependent output patterns.
func nowString() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05")
}
