package opencode

import (
	"fmt"
	"strings"
)

// ErrMultipleMatches is returned when a session ID prefix matches
// more than one session. The user must use the full session ID.
// Run mental session search -o wide to see full IDs.
type ErrMultipleMatches struct {
	Prefix  string
	IDs     []string
	Titles  []string
}

func (e *ErrMultipleMatches) Error() string {
	return fmt.Sprintf(
		"session prefix %q matches %d sessions — use full ID"+
			" (run: mental session search -o wide)",
		e.Prefix, len(e.IDs),
	)
}

// Detail returns a human-readable list of all matching sessions.
func (e *ErrMultipleMatches) Detail() string {
	var b strings.Builder
	for i, id := range e.IDs {
		title := ""
		if i < len(e.Titles) {
			title = e.Titles[i]
		}
		fmt.Fprintf(&b, "  %s  %s\n", id, title)
	}
	return b.String()
}
