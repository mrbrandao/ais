package mem

import "testing"

func TestAddTask(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	t.Run("adds task with sequential ID", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}

		if err := AddTask(cfg, dir, "proj", "First task"); err != nil {
			t.Fatalf("AddTask: %v", err)
		}
		if err := AddTask(cfg, dir, "proj", "Second task"); err != nil {
			t.Fatalf("AddTask: %v", err)
		}

		layout := NewLayout(cfg, dir)
		tasks, err := loadTasks(layout, "proj")
		if err != nil {
			t.Fatalf("loadTasks: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("task count: got %d, want 2", len(tasks))
		}
		if tasks[0].ID != "t001" {
			t.Errorf("first ID: got %q, want t001", tasks[0].ID)
		}
		if tasks[1].ID != "t002" {
			t.Errorf("second ID: got %q, want t002", tasks[1].ID)
		}
		if tasks[0].Status != "todo" {
			t.Errorf("status: got %q, want todo", tasks[0].Status)
		}
	})
}

func TestDoneTask(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	t.Run("marks existing task done", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}
		if err := AddTask(cfg, dir, "proj", "A task"); err != nil {
			t.Fatalf("AddTask: %v", err)
		}
		if err := DoneTask(cfg, dir, "proj", "t001"); err != nil {
			t.Fatalf("DoneTask: %v", err)
		}

		layout := NewLayout(cfg, dir)
		tasks, _ := loadTasks(layout, "proj")
		if tasks[0].Status != "done" {
			t.Errorf("status: got %q, want done", tasks[0].Status)
		}
		if tasks[0].Completed == "" {
			t.Error("completed date should be set")
		}
	})

	t.Run("error for unknown task ID", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := Init(cfg, dir, "proj"); err != nil {
			t.Fatalf("Init: %v", err)
		}
		if err := DoneTask(cfg, dir, "proj", "t999"); err == nil {
			t.Error("expected error for unknown ID")
		}
	})
}

func TestNextTaskID(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tests := []struct {
		name  string
		tasks []Task
		want  string
	}{
		{"empty list", nil, "t001"},
		{"one task", []Task{{ID: "t001"}}, "t002"},
		{"gap in sequence", []Task{{ID: "t001"}, {ID: "t003"}}, "t004"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := nextTaskID(cfg, tc.tasks)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
