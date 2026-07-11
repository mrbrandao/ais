package model

import "time"

// Search type constants control query depth.
const (
	// TypeSmart runs fast first, falls back to deep.
	TypeSmart = "smart"
	// TypeFast searches title and directory only.
	TypeFast = "fast"
	// TypeDeep searches message content as well.
	TypeDeep = "deep"
)

// Query carries all search dimensions for a session
// lookup across any provider.
type Query struct {
	Strings  []string  // -s / --string (repeatable)
	Dir      string    // --dir
	ID       string    // --id
	Model    string    // --model
	Branch   string    // --branch
	Repo     string    // --repo
	DateFrom time.Time // --date-from
	DateTo   time.Time // --date-to
	Type     string    // --type: smart|fast|deep
}
