package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/mrbrandao/mental/internal/config"
	"github.com/mrbrandao/mental/pkg/extensions"
)

// extensionsCmd is the "mental extensions" command group.
// It lists and describes both internal and external extensions.
var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "List and inspect mental extensions",
	Long: `Extensions add capabilities to mental. Built-in extensions
are compiled into the binary. External extensions are discovered
from $MENTAL_DIR/extensions/ via extension.yaml manifests.

Extension classification:
  Kind  — command group: mem, session, output, ...
  Types — operations within that kind (list, search, save, ...)`,
}

// extListOutput holds the -o flag for extensions list.
var extListOutput string

var extensionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed extensions",
	Long: `List installed extensions — built-in and external.

Columns: NAME, KIND, TYPES/CAPS, SOURCE, DESCRIPTION
Use -o wide to see full type lists. Use -o json for scripting.`,
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
			pterm.Info.Println("No extensions installed.")
			return nil
		}

		switch extListOutput {
		case "json":
			return printExtensionsJSON(all)
		case "plain":
			return printExtensionsPlain(all)
		default:
			return printExtensionsTable(all, extListOutput == "wide")
		}
	},
}

var extensionsDescribeCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show full details for an extension",
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
				"extension %q not found — run mental extensions list",
				args[0],
			)
		}

		printExtensionDetail(ext.Info())
		return nil
	},
}

func init() {
	extensionsListCmd.Flags().StringVarP(
		&extListOutput, "output", "o", "table",
		"Output: table|wide|json|plain",
	)
	extensionsCmd.AddCommand(extensionsListCmd)
	extensionsCmd.AddCommand(extensionsDescribeCmd)
	rootCmd.AddCommand(extensionsCmd)
}

// kindStyle returns a colored string for a Kind value.
func kindStyle(kind string) string {
	switch kind {
	case "mem":
		return pterm.FgLightGreen.Sprint(kind)
	case "session":
		return pterm.FgLightBlue.Sprint(kind)
	case "output":
		return pterm.FgLightYellow.Sprint(kind)
	default:
		return pterm.FgGray.Sprint(kind)
	}
}

// sourceStyle returns a colored string for built-in vs external.
func sourceStyle(isBuiltin bool) string {
	if isBuiltin {
		return pterm.FgGreen.Sprint("built-in")
	}
	return pterm.FgLightBlue.Sprint("external")
}

// typesStr returns the types list formatted for the table.
// When wide is false, truncates at 25 chars with ellipsis.
func typesStr(types []string, wide bool) string {
	s := strings.Join(types, ", ")
	if !wide && len(s) > 25 {
		return s[:22] + "…"
	}
	return s
}

// printExtensionsTable renders the extensions list as a pterm table.
func printExtensionsTable(all []extensions.Extension, wide bool) error {
	data := pterm.TableData{
		{"NAME", "KIND", "TYPES/CAPS", "SOURCE", "DESCRIPTION"},
	}

	for _, ext := range all {
		m := ext.Info()
		isBuiltin := m.Executable == ""
		data = append(data, []string{
			m.Name,
			kindStyle(m.Kind),
			pterm.FgGray.Sprint(typesStr(m.Types, wide)),
			sourceStyle(isBuiltin),
			m.Description,
		})
	}

	if err := pterm.DefaultTable.
		WithHasHeader().
		WithData(data).
		Render(); err != nil {
		fmt.Fprintf(os.Stderr, "table render: %v\n", err)
	}
	return nil
}

// printExtensionsJSON renders extensions as a JSON array.
func printExtensionsJSON(all []extensions.Extension) error {
	type row struct {
		Name        string   `json:"name"`
		Kind        string   `json:"kind"`
		Types       []string `json:"types"`
		Source      string   `json:"source"`
		Description string   `json:"description"`
		Version     string   `json:"version,omitempty"`
		Author      string   `json:"author,omitempty"`
	}
	rows := make([]row, len(all))
	for i, ext := range all {
		m := ext.Info()
		source := "built-in"
		if m.Executable != "" {
			source = "external"
		}
		rows[i] = row{
			Name:        m.Name,
			Kind:        m.Kind,
			Types:       m.Types,
			Source:      source,
			Description: m.Description,
			Version:     m.Version,
			Author:      m.Author,
		}
	}
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// printExtensionsPlain renders extensions as tab-separated lines.
func printExtensionsPlain(all []extensions.Extension) error {
	for _, ext := range all {
		m := ext.Info()
		source := "built-in"
		if m.Executable != "" {
			source = "external"
		}
		fmt.Printf(
			"%s\t%s\t%s\t%s\t%s\n",
			m.Name, m.Kind,
			strings.Join(m.Types, ","),
			source, m.Description,
		)
	}
	return nil
}

// printExtensionDetail renders full extension details using pterm.
func printExtensionDetail(m extensions.Manifest) {
	isBuiltin := m.Executable == ""

	pterm.DefaultSection.Println(m.Name)
	pterm.Println(pterm.Gray("Kind:       ") + kindStyle(m.Kind))
	pterm.Println(pterm.Gray("Types/Caps: ") + strings.Join(m.Types, ", "))
	pterm.Println(pterm.Gray("Description:") + " " + m.Description)
	pterm.Println(pterm.Gray("Source:     ") + sourceStyle(isBuiltin))
	if m.Version != "" {
		pterm.Println(pterm.Gray("Version:    ") + m.Version)
	}
	if m.Author != "" {
		pterm.Println(pterm.Gray("Author:     ") + m.Author)
	}
	if !isBuiltin {
		pterm.Println(pterm.Gray("Executable: ") + m.Executable)
		pterm.Println(pterm.Gray("Mode:       ") + string(m.Mode))
	}
}
