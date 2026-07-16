package memx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SaveInput holds the data provided by the LLM for a save operation.
// Passed via stdin as a structured block or built from flags.
type SaveInput struct {
	Project string
	Session SessionMeta
	Topics  []string
	Files   []string
	Summary string
	Memory  string // full MEMORY.md content to write
	Body    string // checkpoint Markdown body (after frontmatter)
}

// RawSave writes a checkpoint file and updates topics.yaml without
// touching MEMORY.md. Use this for provider-extracted checkpoints where
// no LLM synthesis is available. The caller must ensure the project
// directory exists before calling (use Layout.EnsureProjectDirs).
func RawSave(cfg *Config, mentalDir string, input SaveInput) error {
	if err := validateInput(cfg, input); err != nil {
		return fmt.Errorf("invalid save input: %w", err)
	}

	layout := NewLayout(cfg, mentalDir)
	now := time.Now().UTC()

	cpPath := layout.CheckpointFile(input.Project, now)
	if err := writeCheckpoint(cpPath, input, now); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}

	if err := appendTopics(layout, input, cpPath); err != nil {
		return fmt.Errorf("update topics: %w", err)
	}

	fmt.Printf(
		"Saved raw checkpoint: %s\n"+
			"Note: MEMORY.md was not updated. "+
			"Use -p flag to generate a prompt for LLM synthesis.\n",
		cpPath,
	)
	return nil
}

// Save rewrites MEMORY.md, writes a new checkpoint file, and
// appends new topic entries to topics.yaml.
//
// The caller provides a SaveInput with the session details and the
// LLM-authored content. Save validates required fields per the config
// before writing any file.
func Save(cfg *Config, mentalDir string, input SaveInput) error {
	if err := validateInput(cfg, input); err != nil {
		return fmt.Errorf("invalid save input: %w", err)
	}

	layout := NewLayout(cfg, mentalDir)

	if err := layout.EnsureProjectDirs(input.Project); err != nil {
		return fmt.Errorf("ensure dirs: %w", err)
	}

	now := time.Now().UTC()

	if err := writeMemory(layout, input, now); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}

	cpPath := layout.CheckpointFile(input.Project, now)
	if err := writeCheckpoint(cpPath, input, now); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}

	if err := appendTopics(layout, input, cpPath); err != nil {
		return fmt.Errorf("update topics: %w", err)
	}

	fmt.Printf(
		"Saved checkpoint: %s\n",
		layout.CheckpointFile(input.Project, now),
	)
	return nil
}

// ReadSaveInput parses a save payload from r.
// The format is a simple key: value block followed by a blank line,
// then the checkpoint body. Example:
//
//	project: foo
//	session.id: abc-123
//	session.client: opencode
//	session.model: claude-sonnet-4-6
//	session.dir: /home/user/dev/foo
//	topics: auth migration, rollback strategy
//	files: auth/migration.sql, auth/rollback.sql
//	summary: Defined two-phase rollback. Schema finalized.
//	memory:
//	# foo
//	status: active
//	...
//	---
//	## What We Did
//	...
func ReadSaveInput(r io.Reader) (SaveInput, error) {
	scanner := bufio.NewScanner(r)
	var input SaveInput
	var memLines []string
	var bodyLines []string

	inMem := false
	inBody := false

	for scanner.Scan() {
		line := scanner.Text()

		if inBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		if line == "---" && inMem {
			inBody = true
			continue
		}

		if inMem {
			memLines = append(memLines, line)
			continue
		}

		if strings.HasPrefix(line, "memory:") {
			inMem = true
			rest := strings.TrimPrefix(line, "memory:")
			if s := strings.TrimSpace(rest); s != "" {
				memLines = append(memLines, s)
			}
			continue
		}

		parseSaveKV(&input, line)
	}

	if err := scanner.Err(); err != nil {
		return input, fmt.Errorf("read stdin: %w", err)
	}

	input.Memory = strings.Join(memLines, "\n")
	input.Body = strings.Join(bodyLines, "\n")
	return input, nil
}

// parseSaveKV parses a single "key: value" line into input.
func parseSaveKV(input *SaveInput, line string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	switch key {
	case "project":
		input.Project = val
	case "session.id":
		input.Session.ID = val
	case "session.client":
		input.Session.Client = val
	case "session.model":
		input.Session.Model = val
	case "session.dir":
		input.Session.Dir = val
	case "topics":
		input.Topics = splitCSV(val)
	case "files":
		input.Files = splitCSV(val)
	case "summary":
		input.Summary = val
	}
}

// validateInput checks that required fields are present per config.
func validateInput(cfg *Config, input SaveInput) error {
	required := map[string]string{
		"project":        input.Project,
		"session.id":     input.Session.ID,
		"session.client": input.Session.Client,
		"session.model":  input.Session.Model,
		"session.dir":    input.Session.Dir,
	}
	for field, val := range required {
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("field %q is required", field)
		}
	}

	if len(input.Topics) == 0 {
		return fmt.Errorf("at least one topic is required")
	}

	_ = cfg // reserved for future config-driven validation
	return nil
}

// writeMemory rewrites MEMORY.md with the LLM-authored content,
// updating the header timestamp and session metadata.
func writeMemory(l *Layout, input SaveInput, now time.Time) error {
	content := input.Memory
	if content == "" {
		// Preserve existing if LLM did not provide updated memory.
		data, err := os.ReadFile(l.MemoryFile(input.Project))
		if err != nil {
			return fmt.Errorf("read existing memory: %w", err)
		}
		content = string(data)
	}

	return os.WriteFile(l.MemoryFile(input.Project), []byte(content), 0o644)
}

// writeCheckpoint writes the checkpoint YAML frontmatter + Markdown body.
func writeCheckpoint(path string, input SaveInput, now time.Time) error {
	cp := Checkpoint{
		Project: input.Project,
		Date:    now,
		Session: input.Session,
		Topics:  input.Topics,
		Files:   input.Files,
		Summary: input.Summary,
	}

	fm, err := yaml.Marshal(cp)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n\n")

	if input.Body != "" {
		buf.WriteString(strings.TrimSpace(input.Body))
		buf.WriteString("\n")
	} else {
		buf.WriteString("## What We Did\n\n")
		buf.WriteString("## Decisions Made\n\n")
		buf.WriteString("## Open Questions\n\n")
		buf.WriteString("## Handoff\n\n")
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// appendTopics adds new entries to topics.yaml without rewriting it.
func appendTopics(
	l *Layout,
	input SaveInput,
	cpPath string,
) error {
	data, err := os.ReadFile(l.TopicsFile(input.Project))
	if err != nil {
		return fmt.Errorf("read topics: %w", err)
	}

	var tf TopicsFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("parse topics: %w", err)
	}

	entry := TopicEntry{
		Name:       input.Session.ID,
		Checkpoint: cpPath,
		Summary:    input.Summary,
		Topics:     input.Topics,
	}
	tf.Memory = append(tf.Memory, entry)

	out, err := yaml.Marshal(tf)
	if err != nil {
		return fmt.Errorf("marshal topics: %w", err)
	}

	return os.WriteFile(l.TopicsFile(input.Project), out, 0o644)
}

// splitCSV splits a comma-separated string, trimming whitespace.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
