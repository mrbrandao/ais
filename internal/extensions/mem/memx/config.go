// Package memx is the built-in memory engine for mental for mental.
//
// It implements the extensions.Extension interface and manages
// cross-session memory using the plain-file protocol defined in
// mem.config.yaml (embedded as the default configuration).
//
// # Memory Protocol
//
// All data lives under $MENTAL_DIR/projects/<project>/:
//
//	MEMORY.md          rolling summary, rewritten on every save
//	tasks.yaml         task contract shared across sessions
//	topics.yaml        search index: topic → checkpoint files
//	checkpoints/       one file per session, never modified
//
// # Configuration
//
// The directory layout, file names, and validation rules are driven by
// mem.config.yaml. Changing the YAML changes the behaviour without
// recompilation. The binary embeds the default; users can override by
// placing config.yaml at $MENTAL_DIR/extensions/mem/config.yaml.
//
// To change the layout:
//   - Edit layout.* keys in mem.config.yaml.
//   - No Go code changes are required.
//
// To add a required checkpoint section:
//   - Add the section name to checkpoint.sections.required.
//   - The save command validates it automatically.
//
// To add a task status:
//   - Append the status string to tasks.statuses.
//   - The task commands accept it automatically.
package memx

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var defaultConfigBytes []byte

// Config is the runtime configuration for the mem extension.
// It is populated from mem.config.yaml (default) merged with any
// user override at $MENTAL_DIR/extensions/mem/config.yaml.
type Config struct {
	Layout     LayoutConfig     `yaml:"layout"`
	Memory     MemoryConfig     `yaml:"memory"`
	Checkpoint CheckpointConfig `yaml:"checkpoint"`
	Tasks      TasksConfig      `yaml:"tasks"`
}

// LayoutConfig defines where mental stores files on disk.
// All values are relative to $MENTAL_DIR.
type LayoutConfig struct {
	Root    string        `yaml:"root"`
	Project ProjectLayout `yaml:"project"`
}

// ProjectLayout defines the file names within a project directory.
type ProjectLayout struct {
	Memory      string            `yaml:"memory"`
	Tasks       string            `yaml:"tasks"`
	Topics      string            `yaml:"topics"`
	Checkpoints CheckpointLayout  `yaml:"checkpoints"`
}

// CheckpointLayout controls checkpoint file naming.
type CheckpointLayout struct {
	Dir    string `yaml:"dir"`
	Format string `yaml:"format"` // Go time layout string
	Ext    string `yaml:"ext"`
}

// MemoryConfig controls MEMORY.md structure and size constraints.
type MemoryConfig struct {
	MaxLines int             `yaml:"max_lines"`
	Sections []SectionConfig `yaml:"sections"`
}

// SectionConfig describes a single section in MEMORY.md.
type SectionConfig struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
}

// CheckpointConfig defines the frontmatter and body section contract
// for checkpoint files.
type CheckpointConfig struct {
	Frontmatter FrontmatterConfig  `yaml:"frontmatter"`
	Sections    SectionsConfig     `yaml:"sections"`
}

// FrontmatterConfig lists required and optional YAML frontmatter fields.
type FrontmatterConfig struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// SectionsConfig lists required and optional Markdown body sections.
type SectionsConfig struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// TasksConfig controls task ID generation and valid status values.
type TasksConfig struct {
	IDPrefix  string   `yaml:"id_prefix"`
	IDPadding int      `yaml:"id_padding"`
	Statuses  []string `yaml:"statuses"`
}

// LoadConfig parses the embedded default config.yaml.
// Future: merge with a user-supplied override from MENTAL_DIR.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(defaultConfigBytes, &cfg); err != nil {
		return nil, fmt.Errorf("parse mem config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid mem config: %w", err)
	}

	return &cfg, nil
}

// validate checks that all required fields are non-empty.
func (c *Config) validate() error {
	switch {
	case c.Layout.Root == "":
		return fmt.Errorf("layout.root must not be empty")
	case c.Layout.Project.Memory == "":
		return fmt.Errorf(
			"layout.project.memory must not be empty",
		)
	case c.Layout.Project.Tasks == "":
		return fmt.Errorf(
			"layout.project.tasks must not be empty",
		)
	case c.Layout.Project.Topics == "":
		return fmt.Errorf(
			"layout.project.topics must not be empty",
		)
	case c.Layout.Project.Checkpoints.Dir == "":
		return fmt.Errorf(
			"layout.project.checkpoints.dir must not be empty",
		)
	case c.Layout.Project.Checkpoints.Format == "":
		return fmt.Errorf(
			"layout.project.checkpoints.format must not be empty",
		)
	}
	return nil
}
