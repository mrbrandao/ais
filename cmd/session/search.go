package session

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mrbrandao/mental/internal/model"
	"github.com/mrbrandao/mental/internal/output"
	"github.com/mrbrandao/mental/internal/provider/opencode"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sessions for an AI assistant",
	Example: `  mental session search -a opencode -s "git-release"
  mental session search -a opencode -s "foo" -s "bar"
  mental session search -a opencode --type=deep --branch main
  mental session search -a opencode --dir /path/to/project
  mental session search -a opencode --output json`,
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
	Cmd.AddCommand(searchCmd)

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
			return q, fmt.Errorf("--date-from: %w", err)
		}
		q.DateFrom = t
	}
	if flagDateTo != "" {
		t, err := time.Parse("2006-01-02", flagDateTo)
		if err != nil {
			return q, fmt.Errorf("--date-to: %w", err)
		}
		q.DateTo = t
	}
	return q, nil
}

// resolveProvider returns the Provider for the given assistant name.
// Add new cases here as backends grow.
func resolveProvider(
	assistant string,
) (interface {
	Name() string
	Search(context.Context, model.Query) ([]model.Session, error)
}, error) {
	switch assistant {
	case "opencode":
		return opencode.New(), nil
	default:
		fmt.Fprintf(os.Stderr,
			"unknown assistant %q\nsupported: opencode\n",
			assistant,
		)
		return nil, fmt.Errorf(
			"unsupported assistant: %s", assistant,
		)
	}
}
