package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mrbrandao/mental/internal/config"
	"github.com/mrbrandao/mental/internal/extensions/mem"
)

// memCmd is the "mental mem" command group.
// The mem extension manages cross-session memory: checkpoints,
// tasks, and topic-indexed search across project history.
var memCmd = &cobra.Command{
	Use:   "mem",
	Short: "Manage cross-session memory",
	Long: `The mem extension persists context across LLM sessions.

Memory is stored under MENTAL_DIR (default: ~/.local/share/mental)
using a plain-file protocol: MEMORY.md, tasks.yaml, topics.yaml,
and dated checkpoint files.

See: mental mem --help for available subcommands.`,
}

var memInitCmd = &cobra.Command{
	Use:   "init <project>",
	Short: "Initialise memory structure for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		return mem.Init(cfg, mentalDir, args[0])
	},
}

// loadMemConfig loads both the app config (for MENTAL_DIR) and
// the mem extension config. Called by all mem subcommands.
func loadMemConfig() (*mem.Config, string, error) {
	appCfg, err := config.Load()
	if err != nil {
		return nil, "", fmt.Errorf("config: %w", err)
	}
	memCfg, err := mem.LoadConfig()
	if err != nil {
		return nil, "", fmt.Errorf("mem config: %w", err)
	}
	return memCfg, appCfg.Dir(), nil
}

var memLoadCmd = &cobra.Command{
	Use:   "load <project>",
	Short: "Load project memory into session context",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		ctx, err := mem.Load(cfg, mentalDir, args[0])
		if err != nil {
			return err
		}
		mem.PrintContext(ctx)
		return nil
	},
}

var memSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save current session as a checkpoint",
	Long: `Save reads session details from stdin and writes:
- An updated MEMORY.md for the project
- A new timestamped checkpoint file
- Updated topics.yaml with new topic entries

Input format (via stdin):
  project: <name>
  session.id: <id>
  session.client: opencode|claude|cursor
  session.model: <model>
  session.dir: /path/to/project
  topics: topic one, topic two
  files: path/to/file.go, other/file.go
  summary: One-sentence session description.
  memory:
  # <project>
  <full MEMORY.md content>
  ---
  ## What We Did
  <checkpoint body>`,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		input, err := mem.ReadSaveInput(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		return mem.Save(cfg, mentalDir, input)
	},
}

var memSearchCmd = &cobra.Command{
	Use:   "search <topic>",
	Short: "Search topics index for matching checkpoints",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		results, err := mem.Search(cfg, mentalDir, project, args[0])
		if err != nil {
			return err
		}
		mem.PrintSearchResults(results, args[0])
		return nil
	},
}

// memTaskCmd is the "mental mem task" subgroup for task management.
var memTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks within a project",
}

var memTaskAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a task to the current project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		return mem.AddTask(cfg, mentalDir, project,
			strings.Join(args, " "),
		)
	},
}

var memTaskDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		return mem.DoneTask(cfg, mentalDir, project, args[0])
	},
}

var memTaskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks for the current project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		return mem.ListTasks(cfg, mentalDir, project)
	},
}

func init() {
	// --project flag required for all task subcommands.
	for _, cmd := range []*cobra.Command{
		memTaskAddCmd, memTaskDoneCmd, memTaskListCmd,
	} {
		cmd.Flags().String(
			"project", "",
			"Project name (required)",
		)
		_ = cmd.MarkFlagRequired("project")
	}

	memTaskCmd.AddCommand(memTaskAddCmd)
	memTaskCmd.AddCommand(memTaskDoneCmd)
	memTaskCmd.AddCommand(memTaskListCmd)

	memSearchCmd.Flags().String(
		"project", "",
		"Project to search (required)",
	)
	_ = memSearchCmd.MarkFlagRequired("project")

	memCmd.AddCommand(memInitCmd)
	memCmd.AddCommand(memLoadCmd)
	memCmd.AddCommand(memSaveCmd)
	memCmd.AddCommand(memSearchCmd)
	memCmd.AddCommand(memTaskCmd)

	rootCmd.AddCommand(memCmd)
}
