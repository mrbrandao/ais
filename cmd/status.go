package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd prints the current project memory state: MEMORY.md
// summary, task counts, and most recent checkpoint timestamp.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current project memory status",
	Long: `Display the current state of the active project's memory:
MEMORY.md summary, task counts by status, and the most recent
checkpoint. Useful at session start to orient quickly.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("mental status (not yet implemented)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
