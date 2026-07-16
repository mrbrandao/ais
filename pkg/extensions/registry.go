package extensions

import (
	"context"
	"fmt"

	"github.com/mrbrandao/mental/internal/extensions/mem/memx"
	"github.com/mrbrandao/mental/internal/extensions/session/opencode"
)

// builtinMem is the internal memx memory engine registered at startup.
type builtinMem struct{}

func (b *builtinMem) Info() Manifest {
	return Manifest{
		Name:        "memx",
		Kind:        "mem",
		Types:       []string{"init", "load", "save", "search", "task"},
		Description: "Built-in file-based memory engine (MEMORY.md protocol)",
		Version:     "0.1.0",
	}
}

// Run is not used for internal extensions — their commands are wired
// directly into Cobra via cmd/mem.go. This satisfies the interface.
func (b *builtinMem) Run(_ context.Context, _ []string) error {
	return fmt.Errorf("memx: use mental mem <subcommand> directly")
}

// builtinOpenCode is the internal OpenCode session provider.
type builtinOpenCode struct {
	mentalDir string
}

func (b *builtinOpenCode) Info() Manifest {
	return Manifest{
		Name:        "opencode",
		Kind:        "session",
		Types:       []string{"search", "extract"},
		Description: "Built-in OpenCode session provider (SQLite backend)",
		Version:     "0.1.0",
	}
}

func (b *builtinOpenCode) Run(
	ctx context.Context,
	args []string,
) error {
	if len(args) == 0 {
		return fmt.Errorf("opencode: requires a subcommand")
	}
	switch args[0] {
	case "extract":
		if len(args) < 2 {
			return fmt.Errorf(
				"opencode extract: requires session ID",
			)
		}
		_, err := opencode.Extract(args[1], "", "", "")
		return err
	default:
		return fmt.Errorf(
			"opencode: unknown subcommand %q", args[0],
		)
	}
}

// Validate that builtinMem implements Extension at compile time.
// This catches future interface drift early.
var _ Extension = (*builtinMem)(nil)
var _ Extension = (*builtinOpenCode)(nil)

// builtinMentout is the built-in pterm output extension.
// It provides table, json, plain, and wide output formats.
// External output extensions can register alongside it as kind: output.
type builtinMentout struct{}

func (b *builtinMentout) Info() Manifest {
	return Manifest{
		Name:        "mentout",
		Kind:        "output",
		Types:       []string{"table", "wide", "json", "plain"},
		Description: "Built-in pterm output extension (mental default)",
		Version:     "0.1.0",
	}
}

func (b *builtinMentout) Run(_ context.Context, _ []string) error {
	return fmt.Errorf("mentout: output is selected via -o flag on data commands")
}

var _ Extension = (*builtinMentout)(nil)

// RegisterBuiltins registers all built-in extensions with the global
// manager. Call once from cmd/root.go at startup.
func RegisterBuiltins(mentalDir string) error {
	builtins := []Extension{
		&builtinMem{},
		&builtinOpenCode{mentalDir: mentalDir},
		&builtinMentout{},
	}

	for _, ext := range builtins {
		if err := Global.Register(ext); err != nil {
			return fmt.Errorf(
				"register builtin %q: %w",
				ext.Info().Name, err,
			)
		}
	}
	return nil
}

// VerifyConfig checks that memx package is importable by calling
// LoadConfig — used to surface embedding errors early at startup.
func VerifyConfig() error {
	_, err := memx.LoadConfig()
	return err
}
