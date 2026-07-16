package memx

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"layout.root", cfg.Layout.Root, "projects"},
		{"layout.project.memory", cfg.Layout.Project.Memory, "MEMORY.md"},
		{"layout.project.tasks", cfg.Layout.Project.Tasks, "tasks.yaml"},
		{"layout.project.topics", cfg.Layout.Project.Topics, "topics.yaml"},
		{
			"layout.project.checkpoints.dir",
			cfg.Layout.Project.Checkpoints.Dir,
			"checkpoints",
		},
		{
			"layout.project.checkpoints.ext",
			cfg.Layout.Project.Checkpoints.Ext,
			".md",
		},
		{"tasks.id_prefix", cfg.Tasks.IDPrefix, "t"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}

	if cfg.Memory.MaxLines != 50 {
		t.Errorf("max_lines: got %d, want 50", cfg.Memory.MaxLines)
	}

	if cfg.Tasks.IDPadding != 3 {
		t.Errorf("id_padding: got %d, want 3", cfg.Tasks.IDPadding)
	}

	if len(cfg.Tasks.Statuses) == 0 {
		t.Error("tasks.statuses must not be empty")
	}
}
