# Extension Development Guide

mental supports two kinds of extensions: **internal** (compiled into the
binary) and **external** (standalone executables). This guide covers both.

---

## Extension classification

Extensions are classified by two manifest fields:

- **Kind**: the command group the extension belongs to (`mem`, `session`).
- **Types**: the specific operations it implements within that kind.

| Kind | Types | CLI flag |
|------|-------|---------|
| `mem` | init, load, save, search, task | `--engine <name>` |
| `session` | search, extract | `-a/--agent <name>` |

---

## Internal extensions

Internal extensions live in `internal/extensions/<kind>/<name>/` and are
compiled into the mental binary. The built-in `memx` engine is an example.

```
internal/extensions/
├── mem/        ← kind: mem
│   └── memx/   ← built-in memory engine
└── session/    ← kind: session
    └── opencode/ ← built-in OpenCode provider
```

### Implementing an internal extension

1. Create a package directory under the appropriate kind:

   ```
   internal/extensions/<kind>/<name>/
   ├── <name>.go       # implements extensions.Extension
   └── <name>_test.go
   ```

2. Implement the `extensions.Extension` interface:

   ```go
   package myeng

   import (
       "context"
       "github.com/mrbrandao/mental/internal/extensions"
   )

   type MyEng struct{}

   func (e *MyEng) Info() extensions.Manifest {
       return extensions.Manifest{
           Name:        "myeng",
           Kind:        "mem",
           Types:       []string{"init", "load", "save", "search"},
           Description: "My custom memory engine",
           Version:     "0.1.0",
       }
   }

   func (e *MyEng) Run(ctx context.Context, args []string) error {
       // implementation
       return nil
   }
   ```

3. Register with the global manager at startup in `manager.go`:

   ```go
   // internal/extensions/manager.go — in the init/setup function
   if err := Global.Register(&myext.MyExt{}); err != nil {
       return fmt.Errorf("register myext: %w", err)
   }
   ```

4. Wire the Cobra commands in `cmd/`. Internal extension commands follow
   the same pattern as `cmd/mem.go`.

---

## External extensions

External extensions are standalone binaries discovered at runtime from
`$MENTAL_DIR/extensions/<name>/`. They can be written in any language.

### Directory layout

```
$MENTAL_DIR/extensions/
└── hermes/
    ├── extension.yaml   # manifest — required
    └── mental-hermes    # executable — must match manifest.executable
```

### extension.yaml manifest

```yaml
name: hms                                  # identifier for --engine hms
kind: mem                                  # mem | session
types:                                     # operations this extension handles
  - init
  - load
  - save
  - search
description: "Holographic memory engine for mental"
executable: mental-hms
author: YourName
version: "0.1.0"
mode: structured      # structured | passthrough
```

For a session provider extension:

```yaml
name: claude                               # identifier for -a claude
kind: session
types:
  - search
description: "Claude Code session provider for mental"
executable: mental-claude
author: YourName
version: "0.1.0"
mode: structured
```

**mode: passthrough** — mental execs the binary and wires stdin/stdout/
stderr directly to the terminal. The plugin owns the output. Use this for
display-only plugins that do not return data to mental.

**mode: structured** — mental spawns the binary as a subprocess, writes a
JSON request to its stdin, and reads a JSON response from its stdout. Use
this when mental needs to process the plugin's output.

### Environment variables

mental injects these into every external extension process:

| Variable | Value |
|----------|-------|
| `MENTAL_DIR` | resolved data directory (`~/.local/share/mental`) |
| `MENTAL_PROJECT` | current active project name |
| `MENTAL_VERSION` | mental binary version |
| `MENTAL_CONFIG` | path to mental's config file |

### JSON data exchange protocol (structured mode)

**Request** — mental writes this to the extension's stdin:

```json
{
  "command": "search",
  "query": "rollback strategy",
  "project": "foo",
  "mental_dir": "/home/user/.local/share/mental",
  "mental_version": "0.2.0"
}
```

**Response** — extension writes this to stdout:

```json
{
  "results": [
    {
      "file": "2026-07-15-11-00-00.md",
      "summary": "Defined two-phase rollback approach",
      "relevance": 0.92
    }
  ],
  "error": ""
}
```

If `error` is non-empty, mental treats the call as failed and reports it.
The extension must exit 0 for success, non-zero for failure.

### Example external extension in Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

type Request struct {
    Command   string `json:"command"`
    Query     string `json:"query"`
    Project   string `json:"project"`
    MentalDir string `json:"mental_dir"`
}

type Result struct {
    File      string  `json:"file"`
    Summary   string  `json:"summary"`
    Relevance float64 `json:"relevance"`
}

type Response struct {
    Results []Result `json:"results"`
    Error   string   `json:"error"`
}

func main() {
    var req Request
    if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
        respond(Response{Error: err.Error()})
        os.Exit(1)
    }

    results, err := search(req.MentalDir, req.Project, req.Query)
    if err != nil {
        respond(Response{Error: err.Error()})
        os.Exit(1)
    }

    respond(Response{Results: results})
}

func respond(r Response) {
    _ = json.NewEncoder(os.Stdout).Encode(r)
}

func search(dir, project, query string) ([]Result, error) {
    // your search logic here
    return nil, fmt.Errorf("not implemented")
}
```

### Capability discovery

mental queries an extension's capabilities by calling it with
`--mental-describe`. The extension must print its manifest as JSON to
stdout and exit 0:

```json
{
  "name": "Hermes Holographic Memory",
  "type": "memory",
  "mode": "structured",
  "version": "0.1.0",
  "commands": ["search", "store"]
}
```

If the extension does not support `--mental-describe`, mental falls back
to reading the `extension.yaml` manifest from disk.

### Installing an external extension

```bash
# Place the binary and manifest in MENTAL_DIR
mkdir -p ~/.local/share/mental/extensions/hermes
cp mental-hermes ~/.local/share/mental/extensions/hermes/
cp extension.yaml ~/.local/share/mental/extensions/hermes/

# Verify it is discovered
mental extensions list
mental extensions describe hermes
```
