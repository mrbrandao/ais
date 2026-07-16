package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/mrbrandao/mental/internal/config"
	"github.com/mrbrandao/mental/internal/extensions/mem/memx"
	ocpkg "github.com/mrbrandao/mental/internal/extensions/session/opencode"
)

// memEngine is the --engine flag value. Defaults to config value (memx).
// Set via: mental mem <subcommand> --engine <name>
var memEngine string

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

The --engine flag selects which memory engine to use (default: memx).
Override permanently in $MENTAL_DIR/mental.toml: [mem] engine = "hms"

See: mental mem --help for available subcommands.`,
}

var memInitCmd = &cobra.Command{
	Use:   "init [project]",
	Short: "Initialise memory structure for a project",
	Long: `Initialise the memory structure for a project.

If no project name is given, or "." is passed, the current
directory name is used as the project name.

Examples:
  mental mem init           # uses current directory name
  mental mem init .         # same as above
  mental mem init myproject # uses "myproject"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		project, err := resolveProjectName(args)
		if err != nil {
			return err
		}
		cfg, mentalDir, err := loadMemConfig()
		if err != nil {
			return err
		}
		return memx.Init(cfg, mentalDir, project)
	},
}

// resolveProjectName returns the project name from args.
// If args is empty or args[0] is ".", it uses the current
// directory name as the project name.
func resolveProjectName(args []string) (string, error) {
	if len(args) == 0 || args[0] == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		return filepath.Base(wd), nil
	}
	return args[0], nil
}

// loadMemConfig loads both the app config (for MENTAL_DIR) and
// the memx engine config. Called by all mem subcommands.
//
// The --engine flag (or [mem] engine in mental.toml) selects which
// memory engine to use. Currently only "memx" is supported built-in;
// external engines will be routed via the extension manager.
func loadMemConfig() (*memx.Config, string, error) {
	appCfg, err := config.Load()
	if err != nil {
		return nil, "", fmt.Errorf("config: %w", err)
	}

	// Resolve engine: --engine flag > config > default.
	engine := memEngine
	if engine == "" {
		engine = appCfg.MemEngine()
	}
	if engine != "memx" {
		return nil, "", fmt.Errorf(
			"engine %q not yet supported — only memx is built-in; "+
				"install an external extension for other engines",
			engine,
		)
	}

	memCfg, err := memx.LoadConfig()
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
		ctx, err := memx.Load(cfg, mentalDir, args[0])
		if err != nil {
			return err
		}
		memx.PrintContext(ctx)
		return nil
	},
}

var memSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save current session as a checkpoint",
	Long: `Save has three modes selected by flags:

STDIN MODE (default — used by skills and LLM pipes):
  Reads the structured save block from stdin.
  mental mem save < /tmp/checkpoint.txt
  mental mem save -a opencode -s <id> -p | claude -p | mental mem save

PROVIDER MODE (-a and -s flags):
  Extracts session data from an AI provider and writes a raw checkpoint.
  mental mem save -a opencode -s <session-id>
  mental mem save -a opencode -s <session-id> --project myproject

PRINT MODE (-p flag, requires -a and -s):
  Prints an LLM prompt to stdout instead of saving. Pipe to any LLM CLI.
  mental mem save -a opencode -s <session-id> -p | claude -p
  mental mem save -a opencode -s <session-id> -p | ollama run llama3
  mental mem save -a opencode -s <session-id> -p | claude -p | mental mem save

Stdin format (STDIN MODE):
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
	RunE: runSave,
}

// saveFlagAgent is the agent/provider name (-a flag).
// Multiple aliases: --agent, --assistant, --provider all work.
var saveFlagAgent string

// saveFlagSession is the session ID (-s flag).
var saveFlagSession string

// saveFlagPrint enables print mode: output LLM prompt to stdout (-p flag).
var saveFlagPrint bool

// saveFlagProject overrides the project name (inferred from session dir if omitted).
var saveFlagProject string

// runSave dispatches to the correct save mode based on flags.
func runSave(cmd *cobra.Command, _ []string) error {
	cfg, mentalDir, err := loadMemConfig()
	if err != nil {
		return err
	}

	// Provider mode: -a and -s are both set.
	if saveFlagAgent != "" && saveFlagSession != "" {
		return runSaveProvider(cfg, mentalDir)
	}

	// Stdin mode: no provider flags — read structured block from stdin.
	// Return a clear error if stdin is a terminal (would hang).
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf(
			"stdin is a terminal and no provider flags are set\n\n" +
				"Use provider mode:  mental mem save -a opencode -s <session-id>\n" +
				"Use print mode:     mental mem save -a opencode -s <id> -p | claude -p\n" +
				"Use stdin mode:     mental mem save < /tmp/checkpoint.txt",
		)
	}

	input, err := memx.ReadSaveInput(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	return memx.Save(cfg, mentalDir, input)
}

// runSaveProvider extracts session data from the named provider and either
// prints an LLM prompt (-p) or writes a raw checkpoint directly.
func runSaveProvider(cfg *memx.Config, mentalDir string) error {
	switch saveFlagAgent {
	case "opencode":
		input, err := ocpkg.Extract(
			saveFlagSession,
			saveFlagProject,
			"", // use default DB path
			"", // use default diff dir
		)
		if err != nil {
			// Multiple matches: show which sessions match and
			// instruct the user to get the full ID with -o wide.
			var multi *ocpkg.ErrMultipleMatches
			if errors.As(err, &multi) {
				pterm.Warning.Printfln(
					"Session prefix %q is ambiguous — %d sessions match:",
					multi.Prefix, len(multi.IDs),
				)
				pterm.Println(multi.Detail())
				pterm.Info.Println(
					"Run: mental session search -o wide -s " +
						multi.Prefix + "  to see full IDs",
				)
				return fmt.Errorf("ambiguous session ID")
			}
			return fmt.Errorf("extract session: %w", err)
		}

		// Print mode: output LLM prompt for piping.
		if saveFlagPrint {
			fmt.Print(memx.GeneratePrompt(input))
			return nil
		}

		// Raw checkpoint mode: write without LLM synthesis.
		return saveRawCheckpoint(cfg, mentalDir, input)

	default:
		return fmt.Errorf(
			"unknown provider %q — supported: opencode",
			saveFlagAgent,
		)
	}
}

// saveRawCheckpoint writes a minimal checkpoint from provider-extracted data.
// It does NOT update MEMORY.md (requires LLM synthesis for quality).
// It DOES update topics.yaml so raw checkpoints are searchable.
func saveRawCheckpoint(
	cfg *memx.Config,
	mentalDir string,
	input memx.SaveInput,
) error {
	// Build a raw checkpoint body that is honest about its origin.
	input.Body = fmt.Sprintf(
		"## Session Snapshot\n\n"+
			"Source: %s session %s\n"+
			"Summary: %s\n\n"+
			"## What We Did\n\n"+
			"Raw checkpoint from %s session. "+
			"Run `mental mem save -a %s -s %s -p | <llm>` "+
			"to generate a synthesized checkpoint.\n\n"+
			"## Handoff\n\n"+
			"Resume from: %s\n",
		input.Session.Client, input.Session.ID,
		input.Summary,
		input.Session.Client,
		input.Session.Client, input.Session.ID,
		input.Session.Dir,
	)

	// Ensure the project directory exists (auto-init for raw mode).
	if err := memx.NewLayout(cfg, mentalDir).
		EnsureProjectDirs(input.Project); err != nil {
		return fmt.Errorf("ensure project dirs: %w", err)
	}

	return memx.RawSave(cfg, mentalDir, input)
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
		results, err := memx.Search(cfg, mentalDir, project, args[0])
		if err != nil {
			return err
		}
		memx.PrintSearchResults(results, args[0])
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
		return memx.AddTask(cfg, mentalDir, project,
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
		return memx.DoneTask(cfg, mentalDir, project, args[0])
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
		return memx.ListTasks(cfg, mentalDir, project)
	},
}

func init() {
	// --engine persistent flag on mem group — inherited by all subcommands.
	memCmd.PersistentFlags().StringVar(
		&memEngine, "engine", "",
		"Memory engine to use (default from config: memx)",
	)

	// Flags for mental mem save.
	memSaveCmd.Flags().StringVarP(
		&saveFlagAgent, "agent", "a", "",
		"AI provider to extract session from (opencode, ...)",
	)
	// Aliases: --assistant and --provider map to the same variable.
	memSaveCmd.Flags().StringVar(&saveFlagAgent, "assistant", "",
		"alias for --agent")
	memSaveCmd.Flags().StringVar(&saveFlagAgent, "provider", "",
		"alias for --agent")
	memSaveCmd.Flags().StringVarP(
		&saveFlagSession, "session", "s", "",
		"Session ID to extract (requires --agent)",
	)
	memSaveCmd.Flags().BoolVarP(
		&saveFlagPrint, "print", "p", false,
		"Print LLM prompt to stdout instead of saving",
	)
	memSaveCmd.Flags().StringVar(
		&saveFlagProject, "project", "",
		"Project name (inferred from session directory if omitted)",
	)

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
