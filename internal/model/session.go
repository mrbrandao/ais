// Package model defines the core data types shared
// across all providers and output formatters.
package model

import "time"

// Session is a normalized AI assistant session.
// Provider-specific fields go in Metadata.
type Session struct {
	ID        string
	Title     string
	Dir       string
	Assistant string
	UpdatedAt time.Time
	// Metadata holds provider-specific extras.
	Metadata map[string]any
}
