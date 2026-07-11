// Package opencode implements the provider.Provider
// interface for OpenCode sessions stored in SQLite.
package opencode

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/mrbrandao/ais/internal/model"
)

const providerName = "opencode"

// defaultDBPath returns ~/.local/share/opencode/opencode.db.
func defaultDBPath() string {
	return filepath.Join(
		os.Getenv("HOME"),
		".local", "share", "opencode", "opencode.db",
	)
}

// Provider searches OpenCode sessions via SQLite.
type Provider struct {
	path string
}

// New returns a Provider using the default DB path.
func New() *Provider {
	return &Provider{path: defaultDBPath()}
}

// NewWithPath returns a Provider with a custom path.
// Useful for testing.
func NewWithPath(path string) *Provider {
	return &Provider{path: path}
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return providerName }

// Search implements provider.Provider.
// Uses smart mode by default: fast first, then deep.
func (p *Provider) Search(
	ctx context.Context,
	q model.Query,
) ([]model.Session, error) {
	db, err := sql.Open("sqlite", p.path)
	if err != nil {
		return nil, fmt.Errorf(
			"opencode: open: %w", err,
		)
	}
	defer db.Close()

	switch q.Type {
	case model.TypeFast:
		return p.fast(ctx, db, q)
	case model.TypeDeep:
		return p.deep(ctx, db, q)
	default:
		res, err := p.fast(ctx, db, q)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			return res, nil
		}
		return p.deep(ctx, db, q)
	}
}

// fast searches session title and directory fields.
func (p *Provider) fast(
	ctx context.Context,
	db *sql.DB,
	q model.Query,
) ([]model.Session, error) {
	clauses, args := fastWhere(q)
	if len(clauses) == 0 {
		return nil, nil
	}
	stmt := `
		SELECT s.id, s.title, s.directory,
		       s.time_updated
		FROM session s
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY s.time_updated DESC`

	rows, err := db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"opencode: fast query: %w", err,
		)
	}
	defer rows.Close()
	return scan(rows)
}

// deep searches message part JSON content.
func (p *Provider) deep(
	ctx context.Context,
	db *sql.DB,
	q model.Query,
) ([]model.Session, error) {
	clauses, args := deepWhere(q)
	if len(clauses) == 0 {
		return nil, nil
	}
	stmt := `
		SELECT DISTINCT s.id, s.title,
		       s.directory, s.time_updated
		FROM session s
		JOIN message m ON m.session_id = s.id
		JOIN part pt ON pt.message_id = m.id
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY s.time_updated DESC`

	rows, err := db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"opencode: deep query: %w", err,
		)
	}
	defer rows.Close()
	return scan(rows)
}

// fastWhere builds WHERE clauses for the session table.
func fastWhere(
	q model.Query,
) (clauses []string, args []any) {
	for _, s := range q.Strings {
		clauses = append(clauses,
			"(s.title LIKE ? OR s.directory LIKE ?)",
		)
		like := "%" + s + "%"
		args = append(args, like, like)
	}
	if q.Dir != "" {
		clauses = append(clauses, "s.directory LIKE ?")
		args = append(args, "%"+q.Dir+"%")
	}
	if q.ID != "" {
		clauses = append(clauses, "s.id = ?")
		args = append(args, q.ID)
	}
	if !q.DateFrom.IsZero() {
		clauses = append(
			clauses, "s.time_updated >= ?",
		)
		args = append(args, q.DateFrom.UnixMilli())
	}
	if !q.DateTo.IsZero() {
		clauses = append(
			clauses, "s.time_updated <= ?",
		)
		args = append(args, q.DateTo.UnixMilli())
	}
	return clauses, args
}

// deepWhere builds WHERE clauses joining parts.
func deepWhere(
	q model.Query,
) (clauses []string, args []any) {
	for _, s := range q.Strings {
		clauses = append(clauses,
			"(s.title LIKE ? OR s.directory LIKE ?"+
				" OR pt.data LIKE ?)",
		)
		like := "%" + s + "%"
		args = append(args, like, like, like)
	}
	if q.Branch != "" {
		clauses = append(clauses, "pt.data LIKE ?")
		args = append(args, "%"+q.Branch+"%")
	}
	if q.Repo != "" {
		clauses = append(clauses, "pt.data LIKE ?")
		args = append(args, "%"+q.Repo+"%")
	}
	if q.Dir != "" {
		clauses = append(clauses, "s.directory LIKE ?")
		args = append(args, "%"+q.Dir+"%")
	}
	if q.ID != "" {
		clauses = append(clauses, "s.id = ?")
		args = append(args, q.ID)
	}
	return clauses, args
}

// scan reads session rows into a slice.
func scan(rows *sql.Rows) ([]model.Session, error) {
	var sessions []model.Session
	for rows.Next() {
		var (
			s  model.Session
			ms int64
		)
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Dir, &ms,
		); err != nil {
			return nil, fmt.Errorf(
				"opencode: scan: %w", err,
			)
		}
		s.UpdatedAt = time.UnixMilli(ms)
		s.Assistant = providerName
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"opencode: rows: %w", err,
		)
	}
	return sessions, nil
}
