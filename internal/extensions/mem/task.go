package mem

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// AddTask appends a new task to tasks.yaml with an auto-generated ID.
// The ID is generated from the configured prefix, padding, and the
// next available sequence number.
func AddTask(cfg *Config, mentalDir, project, title string) error {
	layout := NewLayout(cfg, mentalDir)

	tasks, err := loadTasks(layout, project)
	if err != nil {
		return err
	}

	id := nextTaskID(cfg, tasks)
	task := Task{
		ID:     id,
		Title:  title,
		Status: "todo",
	}

	tasks = append(tasks, task)

	if err := saveTasks(layout, project, tasks); err != nil {
		return err
	}

	fmt.Printf("Added task #%s: %s\n", id, title)
	return nil
}

// DoneTask marks the task with the given id as done.
// Returns an error if the id is not found.
func DoneTask(cfg *Config, mentalDir, project, id string) error {
	layout := NewLayout(cfg, mentalDir)

	tasks, err := loadTasks(layout, project)
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Status = "done"
			tasks[i].Completed = nowString()[:10] // YYYY-MM-DD
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task %q not found in project %q",
			id, project,
		)
	}

	if err := saveTasks(layout, project, tasks); err != nil {
		return err
	}

	fmt.Printf("Marked #%s as done\n", id)
	return nil
}

// ListTasks reads and prints all tasks for a project to stdout.
func ListTasks(cfg *Config, mentalDir, project string) error {
	layout := NewLayout(cfg, mentalDir)

	tasks, err := loadTasks(layout, project)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Printf("No tasks for project %q\n", project)
		return nil
	}

	fmt.Printf("Tasks for project %q:\n\n", project)
	for _, t := range tasks {
		marker := "[ ]"
		if t.Status == "done" {
			marker = "[x]"
		}
		fmt.Printf("%s #%s %s (%s)\n",
			marker, t.ID, t.Title, t.Status,
		)
		for _, sub := range t.Subtasks {
			subMarker := "  [ ]"
			if sub.Status == "done" {
				subMarker = "  [x]"
			}
			fmt.Printf("%s #%s %s\n",
				subMarker, sub.ID, sub.Title,
			)
		}
	}
	return nil
}

// loadTasks reads and parses tasks.yaml for a project.
func loadTasks(l *Layout, project string) ([]Task, error) {
	data, err := os.ReadFile(l.TasksFile(project))
	if err != nil {
		return nil, fmt.Errorf(
			"read tasks for %q: %w — run mental mem init first",
			project, err,
		)
	}
	var tf TasksFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parse tasks.yaml: %w", err)
	}
	return tf.Tasks, nil
}

// saveTasks writes the task slice back to tasks.yaml.
func saveTasks(l *Layout, project string, tasks []Task) error {
	tf := TasksFile{Tasks: tasks}
	data, err := yaml.Marshal(tf)
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}
	return os.WriteFile(l.TasksFile(project), data, 0o644)
}

// nextTaskID generates the next task ID using the configured prefix
// and zero-padding. Example: "t001", "t002", etc.
func nextTaskID(cfg *Config, tasks []Task) string {
	max := 0
	prefix := cfg.Tasks.IDPrefix

	for _, t := range tasks {
		raw := strings.TrimPrefix(t.ID, prefix)
		n := 0
		_, _ = fmt.Sscanf(raw, "%d", &n)
		if n > max {
			max = n
		}
	}

	format := fmt.Sprintf(
		"%s%%0%dd",
		prefix,
		cfg.Tasks.IDPadding,
	)
	return fmt.Sprintf(format, max+1)
}
