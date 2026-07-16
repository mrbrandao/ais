package memx

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	t.Run("loads memory and tasks", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		ctx, err := Load(cfg, dir, "proj")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if ctx.Project != "proj" {
			t.Errorf("Project: got %q, want %q", ctx.Project, "proj")
		}
		if ctx.Memory == "" {
			t.Error("Memory should not be empty after init")
		}
	})

	t.Run("error when project does not exist", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		_, err := Load(cfg, dir, "missing")
		if err == nil {
			t.Error("expected error for missing project")
		}
	})

	t.Run("reads tasks from tasks.yaml", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "task-proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		layout := NewLayout(cfg, dir)
		taskYAML := `tasks:
  - id: t001
    title: Test task
    status: todo
`
		if err := os.WriteFile(
			layout.TasksFile("task-proj"),
			[]byte(taskYAML), 0o644,
		); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		ctx, err := Load(cfg, dir, "task-proj")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(ctx.Tasks) != 1 {
			t.Errorf("tasks: got %d, want 1", len(ctx.Tasks))
		}
		if ctx.Tasks[0].ID != "t001" {
			t.Errorf("task ID: got %q, want %q",
				ctx.Tasks[0].ID, "t001",
			)
		}
	})
}
