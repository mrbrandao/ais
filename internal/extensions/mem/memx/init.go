package memx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// memoryTemplate is the initial MEMORY.md content for a new project.
// The project name is substituted at creation time.
const memoryTemplate = `# {project}
status: active
updated: {date}

## Context
<Describe what this project is about and its current state.>

## Decisions
<Key decisions made and their rationale.>

## Next Steps
- [ ] <First action for the next session>

## Related
<Links to related projects or tasks, if any.>

## Search
Topics and past sessions: topics.yaml
Command: mental mem search "<topic>"
`

// tasksTemplate is the initial tasks.yaml content for a new project.
const tasksTemplate = `tasks: []
`

// topicsTemplate is the initial topics.yaml content for a new project.
const topicsTemplate = `memory: []
`

// Init creates the project directory structure under mentalDir.
// Returns an error if the project already exists.
func Init(cfg *Config, mentalDir, project string) error {
	layout := NewLayout(cfg, mentalDir)
	projectDir := layout.ProjectDir(project)

	if _, err := os.Stat(projectDir); err == nil {
		return fmt.Errorf(
			"project %q already exists at %s",
			project, projectDir,
		)
	}

	if err := layout.EnsureProjectDirs(project); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	if err := writeInitialFiles(layout, project); err != nil {
		return fmt.Errorf("write initial files: %w", err)
	}

	fmt.Printf(
		"Initialised project %q at %s\n",
		project, projectDir,
	)
	return nil
}

// writeInitialFiles writes MEMORY.md, tasks.yaml, and topics.yaml.
func writeInitialFiles(l *Layout, project string) error {
	now := nowString()
	content := strings.NewReplacer(
		"{project}", project,
		"{date}", now,
	).Replace(memoryTemplate)

	files := []struct {
		path    string
		content string
	}{
		{l.MemoryFile(project), content},
		{l.TasksFile(project), tasksTemplate},
		{l.TopicsFile(project), topicsTemplate},
	}

	for _, f := range files {
		if err := writeFile(f.path, f.content); err != nil {
			return err
		}
	}
	return nil
}

// writeFile writes content to path, creating parent directories as needed.
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
