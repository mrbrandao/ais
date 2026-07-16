package memx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	t.Run("creates project structure", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "test-project"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		paths := []string{
			layout.ProjectDir("test-project"),
			layout.MemoryFile("test-project"),
			layout.TasksFile("test-project"),
			layout.TopicsFile("test-project"),
			layout.CheckpointsDir("test-project"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err != nil {
				t.Errorf("expected %s to exist: %v", p, err)
			}
		}
	})

	t.Run("memory file contains project name", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "my-proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		data, err := os.ReadFile(layout.MemoryFile("my-proj"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data[:10]) != "# my-proj\n" {
			t.Errorf(
				"MEMORY.md should start with project name, got: %q",
				string(data[:20]),
			)
		}
	})

	t.Run("returns error if project exists", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "dup"); err != nil {
			t.Fatalf("first Init: %v", err)
		}
		if err := Init(cfg, dir, "dup"); err == nil {
			t.Error("second Init should return an error")
		}
	})

	t.Run("tasks.yaml has empty list", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "tasks-proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		data, err := os.ReadFile(layout.TasksFile("tasks-proj"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != "tasks: []\n" {
			t.Errorf(
				"tasks.yaml: got %q, want %q",
				string(data), "tasks: []\n",
			)
		}
	})

	t.Run("layout paths use config names", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "layout-proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		// Verify the filename matches config (not hardcoded)
		want := filepath.Join(
			dir, cfg.Layout.Root, "layout-proj",
			cfg.Layout.Project.Memory,
		)
		got := layout.MemoryFile("layout-proj")
		if got != want {
			t.Errorf("MemoryFile: got %q, want %q", got, want)
		}
	})
}
