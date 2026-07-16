// Package output renders search results in multiple
// formats: table (default), json, plain, wide.
//
// The "wide" format is equivalent to "table" but shows full session
// IDs instead of the short (12-char) version. This follows the
// kubectl -o wide convention for showing more detail.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"

	"github.com/mrbrandao/mental/internal/model"
)

// Formatter renders a slice of sessions to stdout.
type Formatter interface {
	Print(sessions []model.Session, assistant string)
}

// New returns a Formatter for the given -o/--output value.
//
// Recognised values:
//   - "table"  — pterm table, short IDs (default)
//   - "wide"   — pterm table, full IDs (kubectl -o wide convention)
//   - "json"   — JSON array, full IDs (machine-readable)
//   - "plain"  — tab-separated, full IDs (pipeline-friendly)
//
// Unknown values fall back to "table".
func New(format string) Formatter {
	switch format {
	case "wide":
		return &tableFmt{wide: true}
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
		return fmt.Sprintf("opencode --session %s", id)
	default:
		return id
	}
}

// shortID returns the first 12 characters of the session ID.
// Used by default to keep tables narrow.
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

// tableFmt renders sessions as a pterm table.
// When wide is true, full session IDs are shown (like kubectl -o wide).
type tableFmt struct {
	wide bool
}

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
		id := shortID(s.ID)
		if f.wide {
			id = s.ID
		}
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
			id,
			title,
			dir,
			fmtTime(s.UpdatedAt),
		})
	}

	if err := pterm.DefaultTable.
		WithHasHeader().
		WithData(data).
		Render(); err != nil {
		fmt.Fprintf(os.Stderr, "table render: %v\n", err)
	}

	pterm.Println()
	pterm.Info.Println(
		"Restore: " + restoreCmd(assistant, "<ID>"),
	)
}

// --- json formatter ---

// jsonFmt renders sessions as a JSON array with full IDs.
// Intended for machine consumption and pipeline use.
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
		fmt.Fprintf(os.Stderr, "json marshal: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

// --- plain formatter ---

// plainFmt renders sessions as tab-separated lines with full IDs.
// Intended for shell pipelines and scripting.
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
