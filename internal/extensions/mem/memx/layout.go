package memx

import (
	"os"
	"path/filepath"
	"time"
)

// Layout resolves all file and directory paths for a project.
// All paths are derived from cfg.Layout — changing mem.config.yaml
// changes the paths returned here; no other code needs to change.
type Layout struct {
	cfg *Config
	// mentalDir is the resolved MENTAL_DIR ($MENTAL_DIR or XDG default).
	mentalDir string
}

// NewLayout constructs a Layout from the given config and MENTAL_DIR.
func NewLayout(cfg *Config, mentalDir string) *Layout {
	return &Layout{cfg: cfg, mentalDir: mentalDir}
}

// ProjectDir returns the root directory for a project.
func (l *Layout) ProjectDir(project string) string {
	return filepath.Join(
		l.mentalDir,
		l.cfg.Layout.Root,
		project,
	)
}

// MemoryFile returns the path to MEMORY.md for a project.
func (l *Layout) MemoryFile(project string) string {
	return filepath.Join(
		l.ProjectDir(project),
		l.cfg.Layout.Project.Memory,
	)
}

// TasksFile returns the path to tasks.yaml for a project.
func (l *Layout) TasksFile(project string) string {
	return filepath.Join(
		l.ProjectDir(project),
		l.cfg.Layout.Project.Tasks,
	)
}

// TopicsFile returns the path to topics.yaml for a project.
func (l *Layout) TopicsFile(project string) string {
	return filepath.Join(
		l.ProjectDir(project),
		l.cfg.Layout.Project.Topics,
	)
}

// CheckpointsDir returns the checkpoints directory for a project.
func (l *Layout) CheckpointsDir(project string) string {
	return filepath.Join(
		l.ProjectDir(project),
		l.cfg.Layout.Project.Checkpoints.Dir,
	)
}

// CheckpointFile returns the path for a new checkpoint file.
// The filename encodes the given time using the configured format.
func (l *Layout) CheckpointFile(project string, t time.Time) string {
	name := t.Format(l.cfg.Layout.Project.Checkpoints.Format) +
		l.cfg.Layout.Project.Checkpoints.Ext
	return filepath.Join(l.CheckpointsDir(project), name)
}

// EnsureProjectDirs creates all directories for a project.
// It is idempotent — calling it on an existing project is safe.
func (l *Layout) EnsureProjectDirs(project string) error {
	dirs := []string{
		l.ProjectDir(project),
		l.CheckpointsDir(project),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
