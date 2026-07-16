package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		fmt.Printf("mental mem init: %s (not yet implemented)\n",
			args[0])
		return nil
	},
}

var memLoadCmd = &cobra.Command{
	Use:   "load <project>",
	Short: "Load project memory into session context",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf("mental mem load: %s (not yet implemented)\n",
			args[0])
		return nil
	},
}

var memSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save current session as a checkpoint",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("mental mem save (not yet implemented)")
		return nil
	},
}

var memSearchCmd = &cobra.Command{
	Use:   "search <topic>",
	Short: "Search topics index for matching checkpoints",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf("mental mem search: %q (not yet implemented)\n",
			args[0])
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
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf("mental mem task add: %q (not yet implemented)\n",
			args[0])
		return nil
	},
}

var memTaskDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf("mental mem task done: %s (not yet implemented)\n",
			args[0])
		return nil
	},
}

var memTaskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks for the current project",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("mental mem task list (not yet implemented)")
		return nil
	},
}

func init() {
	memTaskCmd.AddCommand(memTaskAddCmd)
	memTaskCmd.AddCommand(memTaskDoneCmd)
	memTaskCmd.AddCommand(memTaskListCmd)

	memCmd.AddCommand(memInitCmd)
	memCmd.AddCommand(memLoadCmd)
	memCmd.AddCommand(memSaveCmd)
	memCmd.AddCommand(memSearchCmd)
	memCmd.AddCommand(memTaskCmd)

	rootCmd.AddCommand(memCmd)
}
