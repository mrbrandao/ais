package opencode_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/mrbrandao/ais/internal/model"
	"github.com/mrbrandao/ais/internal/provider/opencode"
)

// seedDB creates a minimal OpenCode-shaped SQLite DB
// for testing. Returns the db path.
func seedDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session (
			id           TEXT PRIMARY KEY,
			title        TEXT NOT NULL,
			directory    TEXT NOT NULL,
			time_updated INTEGER NOT NULL
		);
		CREATE TABLE message (
			id         TEXT PRIMARY KEY,
			session_id TEXT NOT NULL
		);
		CREATE TABLE part (
			id         TEXT PRIMARY KEY,
			message_id TEXT NOT NULL,
			data       TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	now := time.Now().UnixMilli()
	_, err = db.Exec(`
		INSERT INTO session VALUES
		  ('ses_abc','Git release work','/dev/git-release',?),
		  ('ses_def','Unrelated session','/dev/other',?);
		INSERT INTO message VALUES
		  ('msg_1','ses_abc'),('msg_2','ses_def');
		INSERT INTO part VALUES
		  ('pt_1','msg_1',
		   '{"type":"text","text":"feat/github-issue branch"}'),
		  ('pt_2','msg_2',
		   '{"type":"text","text":"nothing here"}');
	`, now, now)
	if err != nil {
		t.Fatalf("seed rows: %v", err)
	}
	return path
}

func TestSearch_Fast_MatchesTitle(t *testing.T) {
	path := seedDB(t)
	p := opencode.NewWithPath(path)

	results, err := p.Search(context.Background(), model.Query{
		Strings: []string{"Git release"},
		Type:    model.TypeFast,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].ID != "ses_abc" {
		t.Errorf("want ses_abc, got %s", results[0].ID)
	}
}

func TestSearch_Fast_NoMatch(t *testing.T) {
	path := seedDB(t)
	p := opencode.NewWithPath(path)

	results, err := p.Search(context.Background(), model.Query{
		Strings: []string{"feat/github-issue"},
		Type:    model.TypeFast,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf(
			"want 0 fast results, got %d", len(results),
		)
	}
}

func TestSearch_Deep_MatchesContent(t *testing.T) {
	path := seedDB(t)
	p := opencode.NewWithPath(path)

	results, err := p.Search(context.Background(), model.Query{
		Strings: []string{"feat/github-issue"},
		Type:    model.TypeDeep,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].ID != "ses_abc" {
		t.Errorf("want ses_abc, got %s", results[0].ID)
	}
}

func TestSearch_Smart_FallsBackToDeep(t *testing.T) {
	path := seedDB(t)
	p := opencode.NewWithPath(path)

	// "feat/github-issue" not in title — smart falls back
	results, err := p.Search(context.Background(), model.Query{
		Strings: []string{"feat/github-issue"},
		Type:    model.TypeSmart,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
}

func TestSearch_ByDir(t *testing.T) {
	path := seedDB(t)
	p := opencode.NewWithPath(path)

	results, err := p.Search(context.Background(), model.Query{
		Dir:  "/dev/git-release",
		Type: model.TypeFast,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1, got %d", len(results))
	}
}

func TestProvider_Name(t *testing.T) {
	p := opencode.New()
	if p.Name() != "opencode" {
		t.Errorf("want opencode, got %s", p.Name())
	}
}

// Verify Provider satisfies the interface at compile time.
var _ interface {
	Name() string
	Search(
		context.Context,
		model.Query,
	) ([]model.Session, error)
} = (*opencode.Provider)(nil)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
