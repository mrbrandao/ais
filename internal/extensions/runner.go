package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	// runTimeout is the maximum time an external extension may run.
	runTimeout = 30 * time.Second

	// describeFlag is passed to extensions to request their manifest.
	describeFlag = "--mental-describe"
)

// RunRequest is the JSON payload written to an external extension's stdin
// when running in structured mode.
type RunRequest struct {
	Command    string `json:"command"`
	Project    string `json:"project,omitempty"`
	MentalDir  string `json:"mental_dir"`
	MentalVer  string `json:"mental_version"`
	Args       []string `json:"args,omitempty"`
}

// RunResponse is the JSON payload read from an external extension's stdout
// when running in structured mode.
type RunResponse struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ExternalExtension wraps a binary discovered from an extension.yaml.
// It implements the Extension interface.
type ExternalExtension struct {
	manifest Manifest
	binary   string // resolved absolute path
	mentalDir string
	version   string
}

// NewExternalExtension constructs an ExternalExtension from a manifest
// and the resolved binary path.
func NewExternalExtension(
	m Manifest,
	binaryPath,
	mentalDir,
	version string,
) *ExternalExtension {
	return &ExternalExtension{
		manifest:  m,
		binary:    binaryPath,
		mentalDir: mentalDir,
		version:   version,
	}
}

// Info returns the extension manifest.
func (e *ExternalExtension) Info() Manifest {
	return e.manifest
}

// Run executes the external extension.
// In passthrough mode, stdin/stdout/stderr are wired to the terminal.
// In structured mode, a JSON request is written to stdin and the JSON
// response is read from stdout.
func (e *ExternalExtension) Run(
	ctx context.Context,
	args []string,
) error {
	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	switch e.manifest.Mode {
	case ModeStructured:
		return e.runStructured(ctx, args)
	default:
		return e.runPassthrough(ctx, args)
	}
}

// runPassthrough execs the binary with terminal passthrough.
// A broken plugin exits non-zero — mental handles the error cleanly
// without panicking.
func (e *ExternalExtension) runPassthrough(
	ctx context.Context,
	args []string,
) error {
	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = append(os.Environ(), e.envVars()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return e.wrapExitErr(err)
	}
	return nil
}

// runStructured writes a JSON request to stdin and reads a JSON response
// from stdout. The plugin's stderr is captured for error reporting.
func (e *ExternalExtension) runStructured(
	ctx context.Context,
	args []string,
) error {
	req := RunRequest{
		MentalDir: e.mentalDir,
		MentalVer: e.version,
		Args:      args,
	}
	if len(args) > 0 {
		req.Command = args[0]
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var errBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = append(os.Environ(), e.envVars()...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stderr = &errBuf

	out, err := cmd.Output()
	if err != nil {
		stderr := errBuf.String()
		if stderr != "" {
			return fmt.Errorf(
				"plugin %q: %w: %s",
				e.manifest.Name, err, stderr,
			)
		}
		return e.wrapExitErr(err)
	}

	var resp RunResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		// Non-JSON output: print as-is.
		fmt.Print(string(out))
		return nil
	}

	if resp.Error != "" {
		return fmt.Errorf("plugin %q: %s", e.manifest.Name, resp.Error)
	}
	if resp.Output != "" {
		fmt.Print(resp.Output)
	}
	return nil
}

// Describe calls the binary with --mental-describe and parses the
// returned JSON manifest. Falls back to the on-disk manifest if the
// binary does not support the flag.
func (e *ExternalExtension) Describe(
	ctx context.Context,
) (Manifest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.binary, describeFlag)
	cmd.Env = append(os.Environ(), e.envVars()...)

	out, err := cmd.Output()
	if err != nil {
		// Binary does not support --mental-describe; use on-disk manifest.
		return e.manifest, nil
	}

	var m Manifest
	if err := json.Unmarshal(out, &m); err != nil {
		return e.manifest, nil
	}
	return m, nil
}

// envVars returns the environment variables injected into every plugin.
func (e *ExternalExtension) envVars() []string {
	return []string{
		"MENTAL_DIR=" + e.mentalDir,
		"MENTAL_VERSION=" + e.version,
	}
}

// wrapExitErr wraps exec.ExitError with the plugin name for clarity.
func (e *ExternalExtension) wrapExitErr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf(
			"plugin %q exited with code %d",
			e.manifest.Name, exitErr.ExitCode(),
		)
	}
	return fmt.Errorf("plugin %q: %w", e.manifest.Name, err)
}
