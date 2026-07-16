package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mrbrandao/mental/internal/config"
	"github.com/mrbrandao/mental/internal/extensions"
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
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		if err := extensions.DiscoverExternal(
			extensions.Global,
			cfg.Dir(),
			"dev",
		); err != nil {
			return fmt.Errorf("discover extensions: %w", err)
		}

		all := extensions.Global.List()
		if len(all) == 0 {
			fmt.Println("No extensions installed.")
			return nil
		}

		fmt.Printf("%-20s %-12s %s\n", "NAME", "TYPE", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 60))
		for _, ext := range all {
			m := ext.Info()
			fmt.Printf("%-20s %-12s %s\n",
				m.Name, m.Type, m.Description,
			)
		}
		return nil
	},
}

var extensionsDescribeCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details for an extension",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		if err := extensions.DiscoverExternal(
			extensions.Global,
			cfg.Dir(),
			"dev",
		); err != nil {
			return fmt.Errorf("discover extensions: %w", err)
		}

		ext, ok := extensions.Global.Get(args[0])
		if !ok {
			return fmt.Errorf(
				"extension %q not found", args[0],
			)
		}

		m := ext.Info()
		fmt.Printf("Name:        %s\n", m.Name)
		fmt.Printf("Type:        %s\n", m.Type)
		fmt.Printf("Description: %s\n", m.Description)
		if m.Version != "" {
			fmt.Printf("Version:     %s\n", m.Version)
		}
		if m.Author != "" {
			fmt.Printf("Author:      %s\n", m.Author)
		}
		if m.Executable != "" {
			fmt.Printf("Executable:  %s\n", m.Executable)
			fmt.Printf("Mode:        %s\n", m.Mode)
		} else {
			fmt.Println("Kind:        built-in")
		}
		return nil
	},
}

func init() {
	extensionsCmd.AddCommand(extensionsListCmd)
	extensionsCmd.AddCommand(extensionsDescribeCmd)
	rootCmd.AddCommand(extensionsCmd)
}
