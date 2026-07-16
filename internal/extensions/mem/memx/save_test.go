package memx

import (
	"os"
	"strings"
	"testing"
)

func validInput(project string) SaveInput {
	return SaveInput{
		Project: project,
		Session: SessionMeta{
			ID:     "test-session",
			Client: "opencode",
			Model:  "claude-sonnet",
			Dir:    "/tmp/proj",
		},
		Topics:  []string{"auth migration", "rollback"},
		Files:   []string{"auth/migration.sql"},
		Summary: "Defined rollback approach.",
	}
}

func TestSave(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	t.Run("creates checkpoint and updates topics", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		if err := Save(cfg, dir, validInput("proj")); err != nil {
			t.Fatalf("Save: %v", err)
		}

		layout := NewLayout(cfg, dir)
		cps, err := os.ReadDir(layout.CheckpointsDir("proj"))
		if err != nil {
			t.Fatalf("ReadDir: %v", err)
		}
		if len(cps) != 1 {
			t.Errorf("checkpoints: got %d, want 1", len(cps))
		}
	})

	t.Run("checkpoint contains frontmatter", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "front"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		if err := Save(cfg, dir, validInput("front")); err != nil {
			t.Fatalf("Save: %v", err)
		}

		layout := NewLayout(cfg, dir)
		entries, _ := os.ReadDir(layout.CheckpointsDir("front"))
		data, err := os.ReadFile(
			layout.CheckpointsDir("front") + "/" + entries[0].Name(),
		)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if !strings.Contains(string(data), "project: front") {
			t.Error("checkpoint should contain project in frontmatter")
		}
		if !strings.Contains(string(data), "auth migration") {
			t.Error("checkpoint should contain topics")
		}
	})

	t.Run("topics.yaml updated after save", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "topics"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		if err := Save(cfg, dir, validInput("topics")); err != nil {
			t.Fatalf("Save: %v", err)
		}

		layout := NewLayout(cfg, dir)
		data, err := os.ReadFile(layout.TopicsFile("topics"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if !strings.Contains(string(data), "auth migration") {
			t.Error("topics.yaml should contain saved topics")
		}
	})

	t.Run("validates required fields", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		input := validInput("validate")
		input.Project = ""
		if err := Save(cfg, dir, input); err == nil {
			t.Error("Save should error on empty project")
		}

		input = validInput("validate")
		input.Topics = nil
		if err := Save(cfg, dir, input); err == nil {
			t.Error("Save should error on empty topics")
		}
	})
}

func TestReadSaveInput(t *testing.T) {
	t.Parallel()

	raw := `project: foo
session.id: abc-123
session.client: opencode
session.model: claude-sonnet
session.dir: /home/user/foo
topics: auth, rollback
files: auth/migration.sql
summary: Defined rollback approach.
`
	input, err := ReadSaveInput(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("ReadSaveInput: %v", err)
	}

	if input.Project != "foo" {
		t.Errorf("Project: got %q, want %q", input.Project, "foo")
	}
	if input.Session.ID != "abc-123" {
		t.Errorf("Session.ID: got %q, want %q",
			input.Session.ID, "abc-123",
		)
	}
	if len(input.Topics) != 2 {
		t.Errorf("Topics: got %d, want 2", len(input.Topics))
	}
}
