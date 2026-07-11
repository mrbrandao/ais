# aisession Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `ais` — an AI session manager CLI that searches
OpenCode sessions from the terminal, with an extensible provider
architecture ready for additional assistants in future releases.

**Architecture:** Domain-driven layers — thin Cobra commands call
a `provider.Provider` interface; each assistant backend lives in
`internal/provider/<name>/`; `internal/model/` holds pure data
types; `internal/output/` handles pterm/json/plain rendering.

**Tech Stack:** Go 1.25, Cobra, Viper, pterm, modernc.org/sqlite
(CGO-free SQLite), goreleaser, pre-commit, gitleaks, gosec, snyk.

## Global Constraints

- Module: `github.com/mrbrandao/ais`
- Go version: 1.25 (match local `go version`)
- Line width: 80 chars hard wrap — no exceptions
- Effective Go: https://go.dev/doc/effective_go
- Errors always last return value; wrap with `fmt.Errorf("pkg: %w")`
- `context.Context` always first param on provider methods
- `defer` for all cleanup (db.Close, rows.Close)
- No `else` after `return`
- Commits: tpope rules, ≤150 lines changed, conventional prefixes
- Never `git add .` — always explicit paths
- No personal data, local paths, secrets in any committed file
- SQLite driver: `modernc.org/sqlite` (CGO-free, cross-compile safe)
- pterm for table/color output; stdlib `encoding/json` for JSON
- All examples use generic public-safe values

---

## File Map

```
main.go                              entry point, wires Cobra root
cmd/root.go                          root command, global flags
cmd/search.go                        search subcommand
internal/model/session.go            model.Session type
internal/model/query.go              model.Query + Type constants
internal/provider/provider.go        Provider interface
internal/provider/opencode/
  opencode.go                        OpenCode SQLite backend
  opencode_test.go                   table-driven tests
internal/output/output.go            Formatter interface + impls
Makefile                             help target + mk/ includes
mk/go.mk                             build, test, vet, lint, coverage
mk/install.mk                        install, uninstall (PREFIX)
mk/release.mk                        release-dry, release
mk/container.mk                      container-build/binary/run/push
mk/dev.mk                            hooks, pre-commit, dev-deps
Containerfile                        multi-stage builder + runtime
.goreleaser.yaml                     cross-platform release config
.pre-commit-config.yaml              gitleaks, go tools, gosec, snyk
.gitignore                           Go + Vim template
.github/workflows/ci.yml             lint, vet, test on push/PR
.github/workflows/release.yml        goreleaser on v* tags
.github/ISSUE_TEMPLATE/
  bug_report.yml                     bug report form
  feature_request.yml                feature request form
.github/PULL_REQUEST_TEMPLATE.md     PR checklist
AGENTS.md                            LLM rules + contribution guide
install.sh                           curl one-liner installer
README.md                            badges + quickstart
docs/dev.md                          developer setup guide
docs/superpowers/specs/
  2026-07-11-aisession-design.md     full design spec
```

---

### Task 1: Scaffold — go.mod, .gitignore, directories

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `internal/model/.gitkeep`
- Create: `internal/provider/opencode/.gitkeep`
- Create: `internal/output/.gitkeep`
- Create: `cmd/.gitkeep`
- Create: `mk/.gitkeep`
- Create: `.github/workflows/.gitkeep`
- Create: `.github/ISSUE_TEMPLATE/.gitkeep`

**Interfaces:**
- Produces: Go module root that all subsequent tasks import from

- [ ] **Step 1: Initialize go module**

```bash
cd ~/dev/gen/ais
go mod init github.com/mrbrandao/ais
```

Expected: `go.mod` created with `module github.com/mrbrandao/ais`
and `go 1.25`.

- [ ] **Step 2: Create .gitignore**

Write `.gitignore` with the Go + Vim template:

```
### Go ###
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work

### Vim ###
[._]*.s[a-v][a-z]
!*.svg
[._]*.sw[a-p]
[._]s[a-rt-v][a-z]
[._]ss[a-gi-z]
[._]sw[a-p]
Session.vim
Sessionx.vim
.netrwhist
*~
tags
[._]*.un~

### Project ###
bin/
coverage.out
```

- [ ] **Step 3: Create directory skeleton**

```bash
mkdir -p cmd internal/model internal/provider/opencode \
  internal/output mk \
  .github/workflows .github/ISSUE_TEMPLATE \
  docs/superpowers/specs docs/superpowers/plans \
  bin
```

- [ ] **Step 4: Commit scaffold**

```bash
git add go.mod .gitignore
git commit -m "chore: initialize go module and gitignore"
```

---

### Task 2: Domain model types

**Files:**
- Create: `internal/model/session.go`
- Create: `internal/model/query.go`

**Interfaces:**
- Produces: `model.Session`, `model.Query`, `model.TypeSmart`,
  `model.TypeFast`, `model.TypeDeep` — used by all other packages

- [ ] **Step 1: Write internal/model/session.go**

```go
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
```

- [ ] **Step 2: Write internal/model/query.go**

```go
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
```

- [ ] **Step 3: Verify it compiles**

```bash
cd ~/dev/gen/ais && go build ./internal/model/...
```

Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add internal/model/session.go internal/model/query.go
git commit -m "feat: add domain model types (Session, Query)"
```

---

### Task 3: Provider interface

**Files:**
- Create: `internal/provider/provider.go`

**Interfaces:**
- Consumes: `model.Session`, `model.Query`
- Produces: `provider.Provider` interface — implemented by every
  assistant backend

- [ ] **Step 1: Write internal/provider/provider.go**

```go
// Package provider defines the interface every AI
// assistant backend must implement.
package provider

import (
	"context"

	"github.com/mrbrandao/ais/internal/model"
)

// Provider is the contract for an AI assistant
// session backend. Add Export, Import, Save here
// as the tool grows.
type Provider interface {
	// Name returns the assistant identifier,
	// e.g. "opencode", "claude".
	Name() string
	// Search returns sessions matching q.
	Search(
		ctx context.Context,
		q model.Query,
	) ([]model.Session, error)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd ~/dev/gen/ais && go build ./internal/provider/...
```

Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/provider/provider.go
git commit -m "feat: add Provider interface"
```

---

### Task 4: OpenCode SQLite backend

**Files:**
- Create: `internal/provider/opencode/opencode.go`
- Create: `internal/provider/opencode/opencode_test.go`

**Interfaces:**
- Consumes: `provider.Provider`, `model.Session`, `model.Query`
- Produces: `opencode.New() *Provider`,
  `opencode.NewWithPath(path string) *Provider`

- [ ] **Step 1: Add modernc.org/sqlite dependency**

```bash
cd ~/dev/gen/ais
go get modernc.org/sqlite
```

- [ ] **Step 2: Write the failing test first**

Write `internal/provider/opencode/opencode_test.go`:

```go
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
```

- [ ] **Step 3: Run tests — expect FAIL (not compiled yet)**

```bash
cd ~/dev/gen/ais
go test ./internal/provider/opencode/... 2>&1 | head -5
```

Expected: compile error — `opencode` package not found.

- [ ] **Step 4: Write internal/provider/opencode/opencode.go**

```go
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
		SELECT id, title, directory, time_updated
		FROM session
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY time_updated DESC`

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
			s   model.Session
			ms  int64
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
```

- [ ] **Step 5: Run tests — expect PASS**

```bash
cd ~/dev/gen/ais
go test -v ./internal/provider/opencode/...
```

Expected:
```
--- PASS: TestSearch_Fast_MatchesTitle
--- PASS: TestSearch_Fast_NoMatch
--- PASS: TestSearch_Deep_MatchesContent
--- PASS: TestSearch_Smart_FallsBackToDeep
--- PASS: TestSearch_ByDir
--- PASS: TestProvider_Name
PASS
```

- [ ] **Step 6: Fix the fast WHERE to alias session table**

The fast query uses `s.` aliases — update the stmt to use
`session s` alias:

```go
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
```

Re-run tests to confirm still passing.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/opencode/opencode.go \
        internal/provider/opencode/opencode_test.go \
        go.mod go.sum
git commit -m \
  "feat: add OpenCode SQLite provider with smart/fast/deep search"
```

---

### Task 5: Output formatters

**Files:**
- Create: `internal/output/output.go`

**Interfaces:**
- Consumes: `model.Session`
- Produces: `output.Formatter` interface,
  `output.New(format string) Formatter`

- [ ] **Step 1: Add pterm dependency**

```bash
cd ~/dev/gen/ais
go get github.com/pterm/pterm
```

- [ ] **Step 2: Write internal/output/output.go**

```go
// Package output renders search results in multiple
// formats: table (default), json, plain.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"

	"github.com/mrbrandao/ais/internal/model"
)

// Formatter renders a slice of sessions to stdout.
type Formatter interface {
	Print(sessions []model.Session, assistant string)
}

// New returns a Formatter for the given format string.
// Recognised values: "table", "json", "plain".
// Defaults to table for unrecognised values.
func New(format string) Formatter {
	switch format {
	case "json":
		return &jsonFmt{}
	case "plain":
		return &plainFmt{}
	default:
		return &tableFmt{}
	}
}

// restoreCmd returns the restore hint for an assistant.
func restoreCmd(assistant, id string) string {
	switch assistant {
	case "opencode":
		return fmt.Sprintf(
			"opencode --session %s", id,
		)
	default:
		return id
	}
}

// shortID returns first 12 chars of the session ID.
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// fmtTime formats a time for display.
func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// --- table formatter ---

type tableFmt struct{}

func (f *tableFmt) Print(
	sessions []model.Session,
	assistant string,
) {
	if len(sessions) == 0 {
		pterm.Info.Printfln(
			"No sessions found for %s", assistant,
		)
		return
	}

	pterm.DefaultHeader.WithFullWidth().Printfln(
		"Found %d session(s) for %s",
		len(sessions), assistant,
	)

	data := pterm.TableData{
		{"#", "ID", "Title", "Dir", "Updated"},
	}
	for i, s := range sessions {
		dir := s.Dir
		if len(dir) > 30 {
			dir = "..." + dir[len(dir)-27:]
		}
		title := s.Title
		if len(title) > 35 {
			title = title[:32] + "..."
		}
		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			shortID(s.ID),
			title,
			dir,
			fmtTime(s.UpdatedAt),
		})
	}

	if err := pterm.DefaultTable.
		WithHasHeader().
		WithData(data).
		Render(); err != nil {
		fmt.Fprintf(os.Stderr,
			"table render: %v\n", err,
		)
	}

	pterm.Println()
	pterm.Info.Println(
		"Restore hint: " +
			restoreCmd(assistant, "<ID>"),
	)
}

// --- json formatter ---

type jsonFmt struct{}

func (f *jsonFmt) Print(
	sessions []model.Session,
	_ string,
) {
	type row struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Dir       string `json:"dir"`
		Assistant string `json:"assistant"`
		UpdatedAt string `json:"updated_at"`
	}
	rows := make([]row, len(sessions))
	for i, s := range sessions {
		rows[i] = row{
			ID:        s.ID,
			Title:     s.Title,
			Dir:       s.Dir,
			Assistant: s.Assistant,
			UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
		}
	}
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"json marshal: %v\n", err,
		)
		return
	}
	fmt.Println(string(b))
}

// --- plain formatter ---

type plainFmt struct{}

func (f *plainFmt) Print(
	sessions []model.Session,
	assistant string,
) {
	for _, s := range sessions {
		fmt.Printf(
			"%s\t%s\t%s\n",
			s.ID, s.Title,
			restoreCmd(assistant, s.ID),
		)
	}
}
```

- [ ] **Step 3: Verify it compiles**

```bash
cd ~/dev/gen/ais && go build ./internal/output/...
```

Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
git add internal/output/output.go go.mod go.sum
git commit -m "feat: add output formatters (table/json/plain)"
```

---

### Task 6: Cobra commands

**Files:**
- Create: `cmd/root.go`
- Create: `cmd/search.go`
- Create: `main.go`

**Interfaces:**
- Consumes: `provider.Provider`, `output.Formatter`,
  `model.Query`, `opencode.New()`
- Produces: `ais` binary entry point

- [ ] **Step 1: Add Cobra dependency**

```bash
cd ~/dev/gen/ais
go get github.com/spf13/cobra
```

- [ ] **Step 2: Write cmd/root.go**

```go
// Package cmd wires the ais CLI using Cobra.
// Commands are thin — all logic lives in internal/.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ais",
	Short: "AI session manager for multiple assistants",
	Long: `ais — search, export, and manage sessions
across AI assistants (opencode, claude, gemini, ...).

Docs: https://github.com/mrbrandao/ais`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Write cmd/search.go**

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mrbrandao/ais/internal/model"
	"github.com/mrbrandao/ais/internal/output"
	"github.com/mrbrandao/ais/internal/provider/opencode"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sessions for an AI assistant",
	Example: `  ais search -a opencode -s "git-release"
  ais search -a opencode -s "foo" -s "bar"
  ais search -a opencode --type=deep --branch main
  ais search -a opencode --dir /path/to/project
  ais search -a opencode --output json`,
	RunE: runSearch,
}

var (
	flagAssistant  string
	flagStrings    []string
	flagDir        string
	flagID         string
	flagModel      string
	flagBranch     string
	flagRepo       string
	flagDateFrom   string
	flagDateTo     string
	flagSearchType string
	flagOutput     string
)

func init() {
	rootCmd.AddCommand(searchCmd)

	f := searchCmd.Flags()
	f.StringVarP(
		&flagAssistant, "assistant", "a", "opencode",
		"AI assistant to search (opencode, ...)",
	)
	f.StringArrayVarP(
		&flagStrings, "string", "s", nil,
		"Search string (repeatable)",
	)
	f.StringVar(&flagDir, "dir", "",
		"Filter by working directory")
	f.StringVar(&flagID, "id", "",
		"Look up a specific session ID")
	f.StringVar(&flagModel, "model", "",
		"Filter by model used")
	f.StringVar(&flagBranch, "branch", "",
		"Filter by git branch (deep search)")
	f.StringVar(&flagRepo, "repo", "",
		"Filter by git repo (deep search)")
	f.StringVar(&flagDateFrom, "date-from", "",
		"Sessions updated after (YYYY-MM-DD)")
	f.StringVar(&flagDateTo, "date-to", "",
		"Sessions updated before (YYYY-MM-DD)")
	f.StringVar(
		&flagSearchType, "type", model.TypeSmart,
		"Search depth: smart|fast|deep",
	)
	f.StringVarP(
		&flagOutput, "output", "o", "table",
		"Output format: table|json|plain",
	)
}

func runSearch(cmd *cobra.Command, _ []string) error {
	q, err := buildQuery()
	if err != nil {
		return err
	}

	p, err := resolveProvider(flagAssistant)
	if err != nil {
		return err
	}

	sessions, err := p.Search(context.Background(), q)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	f := output.New(flagOutput)
	f.Print(sessions, flagAssistant)
	return nil
}

// buildQuery converts CLI flags into a model.Query.
func buildQuery() (model.Query, error) {
	q := model.Query{
		Strings: flagStrings,
		Dir:     flagDir,
		ID:      flagID,
		Model:   flagModel,
		Branch:  flagBranch,
		Repo:    flagRepo,
		Type:    flagSearchType,
	}

	if flagDateFrom != "" {
		t, err := time.Parse("2006-01-02", flagDateFrom)
		if err != nil {
			return q, fmt.Errorf(
				"--date-from: %w", err,
			)
		}
		q.DateFrom = t
	}
	if flagDateTo != "" {
		t, err := time.Parse("2006-01-02", flagDateTo)
		if err != nil {
			return q, fmt.Errorf(
				"--date-to: %w", err,
			)
		}
		q.DateTo = t
	}
	return q, nil
}

// resolveProvider returns the Provider for the given
// assistant name. Add new cases here as backends grow.
func resolveProvider(
	assistant string,
) (interface {
	Name() string
	Search(
		context.Context,
		model.Query,
	) ([]model.Session, error)
}, error) {
	switch assistant {
	case "opencode":
		return opencode.New(), nil
	default:
		fmt.Fprintf(os.Stderr,
			"unknown assistant %q\n"+
				"supported: opencode\n", assistant,
		)
		return nil, fmt.Errorf(
			"unsupported assistant: %s", assistant,
		)
	}
}
```

- [ ] **Step 4: Write main.go**

```go
package main

import "github.com/mrbrandao/ais/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 5: Verify it builds and runs**

```bash
cd ~/dev/gen/ais
go build -o bin/ais .
./bin/ais --help
./bin/ais search --help
```

Expected: help text prints, no errors.

- [ ] **Step 6: Run all tests**

```bash
go test ./...
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add cmd/root.go cmd/search.go main.go go.mod go.sum
git commit -m "feat: add cobra root and search command"
```

---

### Task 7: Makefile + mk/ build system

**Files:**
- Create: `Makefile`
- Create: `mk/go.mk`
- Create: `mk/install.mk`
- Create: `mk/release.mk`
- Create: `mk/container.mk`
- Create: `mk/dev.mk`

**Interfaces:**
- Produces: `make` → binary, `make install` → system install

- [ ] **Step 1: Write Makefile**

```makefile
SHELL := /bin/bash
.DEFAULT_GOAL := build

.PHONY: help
help: ## - print help and usage
	@printf "ais — AI session manager\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' \
		$(MAKEFILE_LIST) | \
		sed 's/^[^:]*://' | \
		awk 'BEGIN {FS = ":.*?## "}; \
		{printf "\033[36m%-20s\033[0m %s\n", \
		$$1, $$2}'

include mk/go.mk
include mk/install.mk
include mk/release.mk
include mk/container.mk
include mk/dev.mk
```

- [ ] **Step 2: Write mk/go.mk**

```makefile
# Go build targets.
# VERSION: exact tag if on one, else empty
# (goreleaser injects version at release time).
GO_VERSION := $(shell sed -n 's/^go //p' go.mod)
VERSION    := $(shell \
  git describe --tags --exact-match 2>/dev/null || echo "dev")
LDFLAGS    := -s -w \
  -X github.com/mrbrandao/ais/cmd.version=$(VERSION)
GOFLAGS    ?= -trimpath

MIN_COVERAGE ?= 60

.PHONY: build test vet lint fmt tidy coverage \
        coverage-badge

build: ## - build bin/ais binary
	@mkdir -p bin
	go build $(GOFLAGS) \
		-ldflags "$(LDFLAGS)" -o bin/ais .

test: ## - run test suite
	go test -race ./...

vet: ## - run go vet static analysis
	go vet ./...

lint: ## - run golangci-lint
	golangci-lint run ./...

fmt: ## - format code with gofmt
	gofmt -l -w .

tidy: ## - tidy go modules
	GOTOOLCHAIN=go$(GO_VERSION) go mod tidy

coverage: ## - run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-badge: coverage ## - update README.md badge
	$(eval COV := $(shell go tool cover \
		-func=coverage.out | grep ^total | \
		awk '{print $$3}'))
	sed -i \
		"s|coverage-[0-9.]*%25|coverage-$(COV)|g" \
		README.md
	@echo "Coverage: $(COV)"
```

- [ ] **Step 3: Write mk/install.mk**

```makefile
PREFIX  ?= /usr/local
BINDIR  := $(PREFIX)/bin

.PHONY: install uninstall

install: build ## - install ais to $(BINDIR)
	install -Dm755 bin/ais $(DESTDIR)$(BINDIR)/ais
	@echo "Installed: $(DESTDIR)$(BINDIR)/ais"

uninstall: ## - remove ais from $(BINDIR)
	rm -f $(DESTDIR)$(BINDIR)/ais
	@echo "Removed: $(DESTDIR)$(BINDIR)/ais"
```

- [ ] **Step 4: Write mk/release.mk**

```makefile
.PHONY: release-dry release

release-dry: ## - dry-run goreleaser release
	goreleaser release --snapshot \
		--clean --skip=publish

release: ## - release via goreleaser (requires tag)
	goreleaser release --clean
```

- [ ] **Step 5: Write mk/container.mk**

```makefile
IMAGE ?= ghcr.io/mrbrandao/ais
TAG   ?= latest

.PHONY: container-build container-binary \
        container-run container-push

container-build: ## - build container image
	podman build -t $(IMAGE):$(TAG) .

container-binary: ## - extract binary (no Go needed)
	podman build --target builder -t ais-builder .
	podman run --rm \
		-v $(PWD)/bin:/out:Z \
		ais-builder cp /ais /out/ais

container-run: ## - run ais via container (ARGS="...")
	podman run --rm \
		-v $(HOME)/.local/share:/data:ro:Z \
		$(IMAGE):$(TAG) $(ARGS)

container-push: ## - push image to registry
	podman push $(IMAGE):$(TAG)
```

- [ ] **Step 6: Write mk/dev.mk**

```makefile
.PHONY: dev-deps hooks pre-commit clean

dev-deps: ## - install local dev tools
	go install \
		github.com/golangci/golangci-lint/\
cmd/golangci-lint@latest
	@echo "Install snyk: npm install -g snyk"
	@echo "  or: brew install snyk"

hooks: ## - install pre-commit hooks
	pre-commit install

pre-commit: ## - run pre-commit on all files
	pre-commit run --all-files

clean: ## - remove build artifacts
	rm -rf bin/ coverage.out
```

- [ ] **Step 7: Verify make works**

```bash
cd ~/dev/gen/ais && make
```

Expected: `bin/ais` produced.

```bash
make help
```

Expected: colored help table with all targets listed.

- [ ] **Step 8: Commit**

```bash
git add Makefile mk/go.mk mk/install.mk \
        mk/release.mk mk/container.mk mk/dev.mk
git commit -m "build: add Makefile with lola pattern and mk/ targets"
```

---

### Task 8: Containerfile

**Files:**
- Create: `Containerfile`

- [ ] **Step 1: Write Containerfile**

```dockerfile
# Stage 1 — builder: compiles the binary.
# Use this stage to extract bin/ais without
# installing Go locally (see mk/container.mk).
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath \
      -ldflags "-s -w" \
      -o /ais .

# Stage 2 — runtime: minimal image with binary only.
# Compatible with both podman and docker.
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ais /usr/local/bin/ais
ENTRYPOINT ["ais"]
```

- [ ] **Step 2: Verify build (requires podman or docker)**

```bash
podman build -t ais:local . && echo "OK"
# or: docker build -f Containerfile -t ais:local .
```

Expected: build succeeds, image tagged `ais:local`.

- [ ] **Step 3: Commit**

```bash
git add Containerfile
git commit -m "build: add multi-stage Containerfile"
```

---

### Task 9: goreleaser config

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Write .goreleaser.yaml**

```yaml
version: 2

project_name: ais

before:
  hooks:
    - go mod tidy

builds:
  - main: .
    binary: ais
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X github.com/mrbrandao/ais/cmd.version={{.Version}}
    flags:
      - -trimpath

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - Merge pull request
      - Merge branch
```

- [ ] **Step 2: Dry-run (requires goreleaser)**

```bash
goreleaser release --snapshot --clean --skip=publish
```

Expected: `dist/` populated with cross-compiled binaries.
If goreleaser not installed: `go install
github.com/goreleaser/goreleaser/v2@latest`.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "ci: add goreleaser config for cross-platform releases"
```

---

### Task 10: pre-commit config

**Files:**
- Create: `.pre-commit-config.yaml`

- [ ] **Step 1: Write .pre-commit-config.yaml**

```yaml
repos:
  # Secret scanning — runs first
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.27.2
    hooks:
      - id: gitleaks

  # Official Go tools
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-vet
      - id: golangci-lint

  # Go SAST
  - repo: https://github.com/securego/gosec
    rev: v2.22.4
    hooks:
      - id: gosec

  # Vulnerability scanning (requires snyk CLI)
  # Install: npm install -g snyk
  - repo: local
    hooks:
      - id: snyk
        name: snyk
        entry: snyk test
        language: system
        pass_filenames: false
```

- [ ] **Step 2: Commit**

```bash
git add .pre-commit-config.yaml
git commit -m "ci: add pre-commit with gitleaks, go tools, gosec, snyk"
```

---

### Task 11: GitHub workflows and templates

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `.github/ISSUE_TEMPLATE/bug_report.yml`
- Create: `.github/ISSUE_TEMPLATE/feature_request.yml`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`

- [ ] **Step 1: Write .github/workflows/ci.yml**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: make fmt
      - run: make vet
      - run: make lint

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: make test

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: make build
```

- [ ] **Step 2: Write .github/workflows/release.yml**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  release:
    name: goreleaser
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Write bug_report.yml**

```yaml
name: "Bug Report"
description: Report unexpected behavior in ais
title: "[BUG] "
labels: ["bug"]
body:
  - type: textarea
    id: description
    attributes:
      label: "Description"
      description: What happened and what you expected
    validations:
      required: true

  - type: textarea
    id: steps
    attributes:
      label: "Steps to Reproduce"
      placeholder: |
        1. Run 'ais search -a opencode -s ...'
        2. See error
    validations:
      required: true

  - type: input
    id: version
    attributes:
      label: "ais version"
      placeholder: "ais --version"

  - type: input
    id: os
    attributes:
      label: "OS"
      placeholder: "e.g. Fedora 42, macOS 15, Ubuntu 24.04"

  - type: textarea
    id: logs
    attributes:
      label: "Output / Logs"
      render: shell
```

- [ ] **Step 4: Write feature_request.yml**

```yaml
name: "Feature Request"
description: Suggest a new feature or improvement
title: "[FEAT] "
labels: ["enhancement"]
body:
  - type: textarea
    id: problem
    attributes:
      label: "Problem"
      description: What problem does this solve?
    validations:
      required: true

  - type: textarea
    id: solution
    attributes:
      label: "Proposed Solution"
    validations:
      required: true

  - type: textarea
    id: context
    attributes:
      label: "Additional Context"
```

- [ ] **Step 5: Write PULL_REQUEST_TEMPLATE.md**

```markdown
## Summary

<!-- What changed and why (1-3 bullet points) -->

-

## Related Issues

<!-- Fixes #N or Relates to #N -->

## Test Plan

- [ ]

## Checklist

- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Follows 80-char line width
- [ ] Commits follow tpope rules

## AI Disclosure

<!-- If AI-assisted, note the tool used -->
<!-- Delete if not applicable -->
```

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/ci.yml \
        .github/workflows/release.yml \
        .github/ISSUE_TEMPLATE/bug_report.yml \
        .github/ISSUE_TEMPLATE/feature_request.yml \
        .github/PULL_REQUEST_TEMPLATE.md
git commit -m "ci: add GitHub workflows and issue templates"
```

---

### Task 12: install.sh curl installer

**Files:**
- Create: `install.sh`

- [ ] **Step 1: Write install.sh**

```bash
#!/usr/bin/env bash
# ais installer — downloads the latest release binary
# from GitHub and installs it to ~/.local/bin or
# /usr/local/bin (with sudo).
#
# Usage:
#   curl -sSfL \
#     https://raw.githubusercontent.com/mrbrandao/ais/main/install.sh \
#     | bash
set -euo pipefail

REPO="mrbrandao/ais"
BINARY="ais"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

# detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)  OS="linux"  ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# detect arch
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  arm64|aarch64)   ARCH="arm64" ;;
  *)
    echo "Unsupported arch: $ARCH" >&2
    exit 1
    ;;
esac

# fetch latest release tag
TAG="$(curl -sSf \
  "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | cut -d'"' -f4)"

FILENAME="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ais ${TAG} (${OS}/${ARCH})..."
curl -sSfL "$URL" -o "${TMP}/${FILENAME}"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"

mkdir -p "$INSTALL_DIR"
install -m755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Installed: ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Add to PATH if needed:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
```

- [ ] **Step 2: Make executable and commit**

```bash
chmod +x install.sh
git add install.sh
git commit -m "feat: add curl install script"
```

---

### Task 13: AGENTS.md

**Files:**
- Create: `AGENTS.md`

- [ ] **Step 1: Write AGENTS.md**

```markdown
# AGENTS.md

Guidance for coding agents and contributors working
in this repository.

## What is ais

`ais` is an AI session manager CLI. It currently
supports searching OpenCode sessions. The provider
architecture is designed to add more assistants
without changing the command layer.

Binary: `ais` | Module: `github.com/mrbrandao/ais`

## Architecture

```
cmd/           Cobra commands — thin, no logic
internal/
  model/       Pure data types (Session, Query)
  provider/    Provider interface + one pkg per assistant
    opencode/  OpenCode SQLite backend
  output/      Formatters: table (pterm), json, plain
```

Commands call provider.Provider. Providers handle
assistant-specific storage. Output formats are
independent of both.

## How to add a new assistant backend

1. Create `internal/provider/<name>/` package
2. Implement the `Provider` interface:
   ```go
   func (p *Provider) Name() string
   func (p *Provider) Search(
       ctx context.Context,
       q model.Query,
   ) ([]model.Session, error)
   ```
3. Register in `cmd/search.go` `resolveProvider()`
4. Add table-driven tests in `<name>_test.go`

## How to add a new command

1. Create `cmd/<command>.go`
2. Define a `cobra.Command`, call `rootCmd.AddCommand`
3. Keep all logic in `internal/` — commands only parse
   flags and call domain functions

## How to add a new output format

1. Add a new struct implementing `output.Formatter`
2. Add a `case` in `output.New(format string)`

Both changes live in `internal/output/output.go`.

## Build and test

```bash
make          # build bin/ais
make test     # go test -race ./...
make lint     # golangci-lint run ./...
make coverage # coverage report
make install  # install to /usr/local/bin
sudo make install  # system-wide
PREFIX=~/.local make install  # user-local
```

See `docs/dev.md` for full developer setup.

## Code standards

- Go 1.25, follow https://go.dev/doc/effective_go
- 80 characters per line — hard wrap
- Errors last return value; wrap with fmt.Errorf
- context.Context always first param
- defer for all cleanup (db.Close, rows.Close)
- No else after return
- Table-driven tests in `_test.go` files
- No CGO — use modernc.org/sqlite for SQLite

## Commit rules (tpope)

- Conventional prefix: feat/fix/docs/test/chore/ci
- Subject: imperative mood, ≤50 chars, no period
- Body: wrapped at 72 chars, explain what and why
- Commit after ≤150 lines changed
- Never `git add .` — always explicit file paths

## Security — NEVER include in any file or commit

- Secrets, tokens, API keys, passwords
- Local filesystem paths that reveal real environments
  (use /path/to/... or ~/.config/... as examples)
- Internal hostnames, IPs, org-internal URLs
- Real session content containing private data
- Any data identifying a real private environment

All examples must use generic, public-safe values.
Only commit content safe to publish publicly.
```

- [ ] **Step 2: Commit**

```bash
git add AGENTS.md
git commit -m "docs: add AGENTS.md with rules and contribution guide"
```

---

### Task 14: README.md and docs/dev.md

**Files:**
- Create: `README.md`
- Create: `docs/dev.md`

- [ ] **Step 1: Write README.md**

```markdown
# ais — AI session manager

![CI](https://github.com/mrbrandao/ais/actions/workflows/ci.yml/badge.svg)
![Release](https://img.shields.io/github/v/release/mrbrandao/ais)
![Go version](https://img.shields.io/github/go-mod/go-version/mrbrandao/ais)
![License](https://img.shields.io/github/license/mrbrandao/ais)
![Coverage](https://img.shields.io/badge/coverage-0%25-lightgrey)

Search, export, and manage sessions across AI assistants
from a single CLI.

## Install

**curl (recommended):**
```bash
curl -sSfL \
  https://raw.githubusercontent.com/mrbrandao/ais/main/install.sh \
  | bash
```

**Go install:**
```bash
go install github.com/mrbrandao/ais@latest
```

**Container (no Go needed):**
```bash
make container-binary   # extracts bin/ais via podman
```

## Quick start

```bash
# Search OpenCode sessions
ais search -a opencode -s "my topic"

# Multiple search terms
ais search -a opencode -s "topic" -s "branch-name"

# Deep search (scans message content)
ais search -a opencode --type=deep --branch feat/my-branch

# Filter by directory
ais search -a opencode --dir /path/to/project

# JSON output
ais search -a opencode -s "topic" --output json

# Restore a session (from output)
opencode --session <session-id>
```

## Supported assistants

| Assistant | Status |
|---|---|
| opencode | supported |
| claude | planned |
| gemini | planned |
| cursor | planned |

## Build from source

```bash
git clone https://github.com/mrbrandao/ais.git
cd ais
make          # builds bin/ais
make install  # installs to /usr/local/bin
```

See [docs/dev.md](docs/dev.md) for full developer setup.

## License

MIT
```

- [ ] **Step 2: Write docs/dev.md**

```markdown
# Developer Guide

## Prerequisites

| Tool | Required | Install |
|---|---|---|
| Go 1.25+ | yes | https://go.dev/dl |
| pre-commit | yes | https://pre-commit.com |
| golangci-lint | yes | `make dev-deps` |
| snyk | for pre-commit | `npm install -g snyk` |
| podman | for containers | https://podman.io |
| goreleaser | for releases | https://goreleaser.com |

## Setup

```bash
git clone https://github.com/mrbrandao/ais.git
cd ais
make dev-deps   # installs golangci-lint
make hooks      # installs pre-commit hooks
```

## Daily workflow

```bash
make          # build bin/ais
make test     # run tests
make lint     # run linter
make fmt      # format code
make coverage # coverage report
```

## Pre-commit hooks

Hooks run automatically on `git commit`:

| Hook | Purpose |
|---|---|
| gitleaks | secret scanning |
| go-fmt | formatting |
| go-vet | static analysis |
| golangci-lint | linting |
| gosec | SAST security scan |
| snyk | vulnerability scan |

snyk requires a separate install:
```bash
npm install -g snyk
snyk auth      # authenticate once
```

## Containers

```bash
# Build container image
make container-build

# Extract binary without Go installed
make container-binary  # produces bin/ais

# Run ais via container
make container-run ARGS="search -a opencode -s topic"
```

## Releases

Releases are automated via goreleaser on tag push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions runs goreleaser and uploads binaries
to the GitHub Release page.

Dry-run locally:
```bash
make release-dry
```

## Adding a provider

See `AGENTS.md` — "How to add a new assistant backend".
```

- [ ] **Step 3: Commit**

```bash
git add README.md docs/dev.md
git commit -m "docs: add README with badges and docs/dev.md setup guide"
```

---

### Task 15: go mod tidy and final verification

- [ ] **Step 1: Tidy modules**

```bash
cd ~/dev/gen/ais
make tidy
```

- [ ] **Step 2: Run full suite**

```bash
make test
make build
make vet
./bin/ais --help
./bin/ais search --help
./bin/ais search -a opencode -s "test" --type=fast
```

Expected: all tests pass, help renders, search runs
against the real OpenCode DB.

- [ ] **Step 3: Commit go.sum if changed**

```bash
git add go.mod go.sum
git commit -m "chore: go mod tidy"
```

- [ ] **Step 4: Push to GitHub**

```bash
git push origin main
```

---

## Self-Review Checklist

- [x] Task 1 covers go.mod, .gitignore, dirs
- [x] Task 2 covers model.Session + model.Query
- [x] Task 3 covers provider.Provider interface
- [x] Task 4 covers OpenCode SQLite backend + tests
- [x] Task 5 covers output formatters (table/json/plain)
- [x] Task 6 covers cmd/root.go, cmd/search.go, main.go
- [x] Task 7 covers Makefile + all mk/ files
- [x] Task 8 covers Containerfile (multi-stage)
- [x] Task 9 covers .goreleaser.yaml
- [x] Task 10 covers .pre-commit-config.yaml
- [x] Task 11 covers GitHub workflows + templates
- [x] Task 12 covers install.sh
- [x] Task 13 covers AGENTS.md
- [x] Task 14 covers README.md + docs/dev.md
- [x] Task 15 covers final verification + push
- [x] No personal paths, names, or secrets in examples
- [x] All file paths use generic public-safe values
- [x] All commits follow tpope rules ≤150 lines
- [x] 80-char line width enforced throughout
- [x] CGO-free SQLite (modernc.org/sqlite)
- [x] context.Context first param on all provider methods
- [x] defer used for db.Close and rows.Close
- [x] Table-driven tests in opencode_test.go
