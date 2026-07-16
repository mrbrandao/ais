package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// extensionsCmd is the "mental extensions" command group.
// It lists and describes both internal and external extensions.
var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "List and inspect mental extensions",
	Long: `Extensions add capabilities to mental. Built-in extensions
are compiled into the binary. External extensions are discovered
from $MENTAL_DIR/extensions/ via extension.yaml manifests.`,
}

var extensionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed extensions",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("mental extensions list (not yet implemented)")
		return nil
	},
}

var extensionsDescribeCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details for an extension",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf(
			"mental extensions describe: %s (not yet implemented)\n",
			args[0],
		)
		return nil
	},
}

func init() {
	extensionsCmd.AddCommand(extensionsListCmd)
	extensionsCmd.AddCommand(extensionsDescribeCmd)
	rootCmd.AddCommand(extensionsCmd)
}
