package memx

import "time"

// Checkpoint represents a single session memory record.
// It maps to the YAML frontmatter of a checkpoint file.
type Checkpoint struct {
	Project string          `yaml:"project"`
	Date    time.Time       `yaml:"date"`
	Session SessionMeta     `yaml:"session"`
	Topics  []string        `yaml:"topics"`
	Files   []string        `yaml:"files"`
	Summary string          `yaml:"summary,omitempty"`
}

// SessionMeta holds identifying information about the LLM session
// that produced a checkpoint. Stored in checkpoint frontmatter.
type SessionMeta struct {
	ID     string `yaml:"id"`
	Client string `yaml:"client"` // opencode | claude | cursor | aider
	Model  string `yaml:"model"`
	Dir    string `yaml:"dir"`
}

// Task represents a unit of work within a project.
// Tasks persist across sessions via tasks.yaml.
type Task struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Status      string    `yaml:"status"`
	Session     string    `yaml:"session,omitempty"`
	Client      string    `yaml:"client,omitempty"`
	BlockedBy   string    `yaml:"blocked_by,omitempty"`
	Completed   string    `yaml:"completed,omitempty"`
	Ref         []string  `yaml:"ref,omitempty"`
	Subtasks    []Subtask `yaml:"subtasks,omitempty"`
}

// Subtask is a lightweight child task. It carries only the
// minimum fields — if a subtask grows complex, promote it to a Task.
type Subtask struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Status string `yaml:"status"`
}

// TopicEntry is one record in topics.yaml, linking a session name
// and its checkpoint file to a list of searchable topic keywords.
type TopicEntry struct {
	Name       string   `yaml:"name"`
	Checkpoint string   `yaml:"checkpoint"`
	Summary    string   `yaml:"summary,omitempty"`
	Topics     []string `yaml:"topics"`
}

// ProjectContext is the combined output of mental mem load.
// It holds the raw MEMORY.md content and the task list for display.
type ProjectContext struct {
	Project string
	Memory  string
	Tasks   []Task
}

// TasksFile is the in-memory representation of tasks.yaml.
type TasksFile struct {
	Tasks []Task `yaml:"tasks"`
}

// TopicsFile is the in-memory representation of topics.yaml.
type TopicsFile struct {
	Memory []TopicEntry `yaml:"memory"`
}
