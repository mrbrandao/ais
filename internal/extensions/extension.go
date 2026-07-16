// Package extensions defines the contract for mental extensions and
// manages their lifecycle — both built-in and external.
//
// # Architecture
//
// Mental supports two kinds of extensions:
//
//   - Internal: compiled into the binary, registered at startup.
//     Code lives in internal/extensions/<kind>/<name>/.
//   - External: standalone executables discovered at runtime from
//     $MENTAL_DIR/extensions/<name>/ via an extension.yaml manifest.
//
// All extensions satisfy the Extension interface, enabling uniform
// registration, listing, and dispatch regardless of kind.
//
// # Extension Classification
//
// Extensions are classified by two fields in their Manifest:
//
//   - Kind: the command group the extension belongs to.
//     Built-in kinds: "mem" (memory engines), "session" (AI providers).
//     Future kinds may be added without changing this package.
//
//   - Types: the specific operations this extension implements within
//     its kind. A mem extension may implement [init, load, save, search,
//     task]; a session extension may implement [search, extract].
//
// This two-field model maps directly to the CLI:
//
//	mental mem save --engine=memx    → kind=mem, type=save, engine=memx
//	mental session search -a opencode → kind=session, type=search
//
// # Adding an Internal Extension
//
//  1. Create a package under internal/extensions/<kind>/<name>/.
//  2. Implement the Extension interface.
//  3. Register via Manager.Register in your package's init or New.
//  4. Call RegisterBuiltins() from cmd/root.go at startup.
//
// # Data Exchange (External Extensions)
//
// External extensions in "structured" mode communicate via JSON on
// stdin/stdout. mental writes a JSON request to the extension's stdin
// and reads a JSON response from its stdout. See runner.go for details.
//
// Extensions in "passthrough" mode own the terminal directly; mental
// does not capture their output.
package extensions

import "context"

// Mode controls how mental communicates with an external extension.
type Mode string

const (
	// ModePassthrough wires the extension directly to the terminal.
	// mental does not capture output. Use for display-only extensions.
	ModePassthrough Mode = "passthrough"

	// ModeStructured captures the extension's stdout as JSON.
	// mental reads the response for further processing or formatting.
	ModeStructured Mode = "structured"
)

// Manifest holds the metadata declared in an extension.yaml file.
// Internal extensions populate this programmatically; external
// extensions have it parsed from their manifest file.
//
// # Kind values
//
//	"mem"     — manages memory storage (MEMORY.md, tasks.yaml, topics.yaml)
//	"session" — connects to an AI assistant to extract session data
//
// # Types values (by kind)
//
//	kind=mem:     init, load, save, search, task
//	kind=session: search, extract
type Manifest struct {
	// Name is the extension identifier used in CLI flags.
	// For --engine: matches a kind=mem extension name.
	// For -a/--agent: matches a kind=session extension name.
	Name string `yaml:"name"`

	// Kind is the command group this extension belongs to.
	// Valid values: "mem", "session" (more kinds may be added).
	Kind string `yaml:"kind"`

	// Types lists the specific operations this extension implements
	// within its kind. Used for capability routing and documentation.
	Types []string `yaml:"types"`

	// Description is a one-line summary shown in "mental extensions list".
	Description string `yaml:"description"`

	// Executable is the binary name for external extensions.
	// Empty for internal extensions.
	Executable string `yaml:"executable,omitempty"`

	// Author is the extension author (name or org).
	Author string `yaml:"author,omitempty"`

	// Version is the extension version string (semver recommended).
	Version string `yaml:"version,omitempty"`

	// Mode controls how mental communicates with an external extension.
	// Internal extensions ignore this field.
	Mode Mode `yaml:"mode,omitempty"`
}

// Extension is the contract all mental extensions implement.
// Built-in extensions call Register on the global Manager.
type Extension interface {
	// Info returns the extension manifest for listing and introspection.
	Info() Manifest

	// Run executes the extension with the given arguments.
	// ctx carries a deadline; extensions must respect cancellation.
	Run(ctx context.Context, args []string) error
}
