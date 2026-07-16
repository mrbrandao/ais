package memx

import (
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	setupProject := func(t *testing.T) (string, string) {
		t.Helper()
		dir := t.TempDir()
		project := "search-proj"

		if err := Init(cfg, dir, project); err != nil {
			t.Fatalf("Init: %v", err)
		}

		for _, input := range []SaveInput{
			{
				Project: project,
				Session: SessionMeta{
					ID:     "s1",
					Client: "opencode",
					Model:  "claude",
					Dir:    "/tmp",
				},
				Topics:  []string{"auth migration", "rollback strategy"},
				Files:   []string{"auth.sql"},
				Summary: "Defined rollback approach.",
			},
			{
				Project: project,
				Session: SessionMeta{
					ID:     "s2",
					Client: "opencode",
					Model:  "claude",
					Dir:    "/tmp",
				},
				Topics:  []string{"database schema", "postgresql"},
				Files:   []string{"schema.sql"},
				Summary: "Schema finalized.",
			},
		} {
			if err := Save(cfg, dir, input); err != nil {
				t.Fatalf("Save: %v", err)
			}
		}
		return dir, project
	}

	t.Run("finds by topic keyword", func(t *testing.T) {
		t.Parallel()
		dir, project := setupProject(t)

		results, err := Search(cfg, dir, project, "rollback")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
	})

	t.Run("finds by summary keyword", func(t *testing.T) {
		t.Parallel()
		dir, project := setupProject(t)

		results, err := Search(cfg, dir, project, "finalized")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		t.Parallel()
		dir, project := setupProject(t)

		results, err := Search(cfg, dir, project, "AUTH")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
	})

	t.Run("no results for unknown topic", func(t *testing.T) {
		t.Parallel()
		dir, project := setupProject(t)

		results, err := Search(cfg, dir, project, "kubernetes")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("got %d results, want 0", len(results))
		}
	})

	t.Run("error when topics file missing", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		_, err := Search(cfg, dir, "missing-proj", "anything")
		if err == nil {
			t.Error("expected error for missing project")
		}
	})

	t.Run("empty topics.yaml returns no results", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "empty"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		_ = os.WriteFile(
			layout.TopicsFile("empty"),
			[]byte("memory: []\n"),
			0o644,
		)

		results, err := Search(cfg, dir, "empty", "anything")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("got %d results, want 0", len(results))
		}
	})
}
