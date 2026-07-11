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
