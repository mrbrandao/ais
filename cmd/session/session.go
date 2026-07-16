// Package session provides the "mental session" command group.
// Session commands operate on AI assistant session history stored
// by external tools (OpenCode, Claude Code, etc.).
package session

import "github.com/spf13/cobra"

// Cmd is the "mental session" parent command.
// Subcommands are registered by their own init functions.
var Cmd = &cobra.Command{
	Use:   "session",
	Short: "Manage and search AI assistant sessions",
	Long: `Search, inspect, and manage sessions from AI assistants
such as OpenCode, Claude Code, and others.`,
}
