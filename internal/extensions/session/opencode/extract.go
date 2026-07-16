// Package opencode provides session extraction from OpenCode's SQLite database.
// It reads session metadata, auto-generated titles, and file change summaries
// to produce raw checkpoints without requiring an external LLM.
//
// # Data sources
//
//   - session.title        — OpenCode auto-generates this via a hidden title agent.
//     Used as the checkpoint summary (already a quality one-liner).
//   - session.directory    — working directory; used to infer the project name.
//   - session.time_updated — session timestamp.
//   - session_diff/*.json  — per-session git diff summary; provides the files list.
//
// # Topic extraction
//
// Without an LLM, topics are extracted from the session title by splitting on
// whitespace and punctuation, lowercasing, and filtering words ≤2 characters
// and common stop words. The result is imperfect but searchable.
package opencode

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	_ "modernc.org/sqlite"

	"github.com/mrbrandao/mental/internal/extensions/mem/memx"
)

const clientName = "opencode"

// defaultDBPath returns ~/.local/share/opencode/opencode.db.
func defaultDBPath() string {
	return filepath.Join(
		os.Getenv("HOME"),
		".local", "share", "opencode", "opencode.db",
	)
}

// defaultDiffDir returns ~/.local/share/opencode/storage/session_diff/.
func defaultDiffDir() string {
	return filepath.Join(
		os.Getenv("HOME"),
		".local", "share", "opencode", "storage", "session_diff",
	)
}

// sessionRow holds raw data from the session table.
type sessionRow struct {
	id        string
	title     string
	directory string
	updatedMS int64
}

// diffEntry is one file record from a session_diff JSON file.
// OpenCode stores git diff info per session as JSON arrays.
type diffEntry struct {
	File string `json:"file"`
}

// Extract fetches session data from OpenCode and returns a SaveInput
// suitable for writing a raw checkpoint (no LLM synthesis).
//
// The returned SaveInput has:
//   - Summary: session title (OpenCode-generated)
//   - Files: from session_diff JSON if available
//   - Topics: keywords extracted from the session title
//   - Memory: empty (caller decides whether to update MEMORY.md)
func Extract(
	sessionID,
	project,
	dbPath,
	diffDir string,
) (memx.SaveInput, error) {
	if dbPath == "" {
		dbPath = defaultDBPath()
	}
	if diffDir == "" {
		diffDir = defaultDiffDir()
	}

	row, err := fetchSession(dbPath, sessionID)
	if err != nil {
		return memx.SaveInput{}, fmt.Errorf(
			"opencode extract: %w", err,
		)
	}

	// Infer project from session directory if not provided.
	if project == "" {
		project = filepath.Base(row.directory)
	}

	files := extractFiles(diffDir, sessionID)
	topics := extractTopics(row.title)

	return memx.SaveInput{
		Project: project,
		Session: memx.SessionMeta{
			ID:     row.id,
			Client: clientName,
			Model:  "unknown", // model not stored per-session in current schema
			Dir:    row.directory,
		},
		Topics:  topics,
		Files:   files,
		Summary: row.title,
		// Memory and Body are left empty for raw checkpoints.
		// The caller writes the checkpoint body as a session snapshot.
	}, nil
}

// fetchSession retrieves a single session row from OpenCode's SQLite.
func fetchSession(dbPath, sessionID string) (sessionRow, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return sessionRow{}, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	var row sessionRow
	err = db.QueryRow(
		`SELECT id, title, directory, time_updated
		 FROM session WHERE id = ?`,
		sessionID,
	).Scan(&row.id, &row.title, &row.directory, &row.updatedMS)

	switch {
	case err == sql.ErrNoRows:
		return sessionRow{}, fmt.Errorf(
			"session %q not found in OpenCode database", sessionID,
		)
	case err != nil:
		return sessionRow{}, fmt.Errorf("query session: %w", err)
	}
	return row, nil
}

// extractFiles reads session_diff/<sessionID>.json and returns
// the list of changed file paths. Returns nil if not found.
func extractFiles(diffDir, sessionID string) []string {
	path := filepath.Join(diffDir, sessionID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // diff file absent — not an error
	}

	var entries []diffEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.File != "" {
			files = append(files, e.File)
		}
	}
	return files
}

// stopWords is the set of words filtered from topic extraction.
// These are common English words that carry no topical meaning.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "in": true,
	"on": true, "at": true, "to": true, "of": true, "and": true,
	"or": true, "for": true, "with": true, "from": true, "by": true,
	"as": true, "be": true, "it": true, "its": true, "we": true,
	"add": true, "use": true, "fix": true, "run": true, "get": true,
	"set": true, "new": true, "all": true, "not": true,
}

// extractTopics splits the session title into searchable keywords.
// Words ≤2 characters and stop words are filtered out.
// This is intentionally simple — LLM synthesis produces better topics.
func extractTopics(title string) []string {
	// Split on whitespace and common punctuation.
	words := strings.FieldsFunc(title, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == '.' ||
			r == ':' || r == ';' || r == '-' || r == '_'
	})

	seen := make(map[string]bool)
	var topics []string
	for _, w := range words {
		w = strings.ToLower(strings.Trim(w, "\"'()[]{}"))
		if len(w) <= 2 || stopWords[w] {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		topics = append(topics, w)
	}
	return topics
}
