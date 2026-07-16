# AGENTS.md

Guidance for coding agents and contributors
working in this repository.

## Memory (mental)

This project uses mental for cross-session memory.
The mental binary manages memory, tasks, and checkpoints.

**At session start**, load project context:

```bash
mental mem load mental
```

**At session end**, save a checkpoint:

```bash
mental mem save < /tmp/mental-save.txt
```

**Search past sessions**:

```bash
mental mem search "<topic>" --project mental
```

See `.opencode/skills/mental/SKILL.md` for the full trigger
vocabulary and save input format.

## What is mental

`mental` is an AI Session Manager. It searches session history from AI
assistants and manages session context memory across sessions and providers.
Extensible via built-in and external extensions.

Binary: `mental` | Module: `github.com/mrbrandao/mental`

## Architecture

```
cmd/                    Cobra commands — thin, no logic
  session/              mental session search
internal/
  model/                Session search types (Session, Query)
  provider/             Provider interface + per-assistant pkg
    opencode/           OpenCode SQLite backend (session search)
  output/               Formatters: table (pterm), json, plain
  config/               XDG resolution + viper config
  extensions/           Extension system
    extension.go        Extension interface + Manifest{kind, types}
    manager.go          Registry: internal + XDG external scan
    runner.go           External subprocess + JSON protocol
    discover.go         XDG extension discovery
    mem/                kind: mem — memory engine extensions
      memx/             Built-in default memory engine (memx)
        config.yaml     Embedded YAML-driven layout config
        types.go        Checkpoint, Task, Topic, ProjectContext
    session/            kind: session — AI session provider extensions
      opencode/         Built-in OpenCode provider (search + extract)
```

## Documentation rule

Every code change must update corresponding docs in the SAME commit:

| Change | Must update |
|--------|-------------|
| CLI command or flag | README.md quick start |
| Architecture | AGENTS.md architecture section |
| Extension system | docs/dev-guide/extension-development.md |
| mem protocol | docs/dev-guide/mem-extension.md |
| Contributing workflow | docs/dev-guide/contributing.md |

## How to add a new session provider (kind: session)

For session search only:
1. Create `internal/extensions/session/<name>/`
2. Implement `extensions.Extension` with `Kind: "session"`
3. Register in `cmd/root.go` via `RegisterBuiltins()`
4. Wire `-a <name>` routing in `cmd/session/search.go`

For session search that also powers `mental mem save -a <name>`:
- Also implement `Extract(sessionID, project, ...) (memx.SaveInput, error)`
- Wire in `cmd/mem.go` `runSaveProvider()` switch

## How to add a new command

1. Create `cmd/<command>.go`
2. Define a `cobra.Command`, register with
   `rootCmd.AddCommand`
3. All logic goes in `internal/` — commands only
   parse flags and call domain functions

## How to add a new output format

1. Add a struct implementing `output.Formatter`
2. Add a case in `output.New(format string)`

Both changes live in `internal/output/output.go`.

## How to create an external extension

External extensions are standalone executables discovered
in `$MENTAL_DIR/extensions/<name>/`. Each must provide:

1. An `extension.yaml` manifest (name, type, executable, mode)
2. An executable binary that reads JSON from stdin and writes
   JSON to stdout (structured mode), or owns the terminal
   (passthrough mode)

See `docs/dev-guide/extension-development.md` for the full
contract, environment variables, and worked examples.

## Build and test

```bash
make            # build bin/mental
make test       # go test -race ./...
make lint       # golangci-lint run ./...
make coverage   # coverage report
make install    # install to /usr/local/bin
PREFIX=~/.local make install  # user-local
```

See `docs/dev.md` for full developer setup.

## Code standards

- Go 1.25, follow https://go.dev/doc/effective_go
- 80 characters per line — hard wrap
- Errors last return value; wrap with fmt.Errorf
- context.Context always first param
- defer for all cleanup (db.Close, rows.Close)
- No else after return
- Functions small and single-purpose
- Every exported symbol has a godoc comment
- Table-driven tests in `_test.go` files
- No CGO — use modernc.org/sqlite for SQLite

## Commit rules (tpope)

- Conventional prefix: feat/fix/docs/test/chore/ci
- Subject: imperative mood, ≤50 chars, no period
- Body: wrapped at 72 chars, explain what and why
- Commit after ≤150 lines changed
- Never `git add .` — always explicit file paths

## Security — NEVER include in any file or commit

- Secrets, tokens, API keys, passwords
- Local filesystem paths revealing real environments
  (use /path/to/... or ~/.config/... as examples)
- Internal hostnames, IPs, org-internal URLs
- Real session content containing private data
- Personal information beyond public git metadata
- Any data identifying a real private environment

All examples must use generic, public-safe values.
Only commit content safe to publish publicly.
