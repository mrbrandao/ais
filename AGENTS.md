# AGENTS.md

Guidance for coding agents and contributors
working in this repository.

## What is mental

`mental` is a cross-session memory and AI session manager CLI.
It persists LLM context across sessions, tracks tasks, and lets
multiple agents share knowledge through a file-based protocol.
The extension architecture supports built-in and external plugins.

Binary: `mental` | Module: `github.com/mrbrandao/mental`

## Architecture

```
cmd/                Cobra commands — thin, no logic
internal/
  model/            Session search types (Session, Query)
  provider/         Provider interface + per-assistant pkg
    opencode/       OpenCode SQLite backend
  output/           Formatters: table (pterm), json, plain
  config/           XDG resolution + viper config
  extensions/       Extension system
    extension.go    Extension interface + manifest types
    manager.go      Registry: internal + XDG external scan
    runner.go       External subprocess execution
    mem/            Built-in mem extension
      config.yaml   Embedded default layout config
      types.go      Checkpoint, Task, Topic, ProjectContext
```

## How to add a new assistant backend

1. Create `internal/provider/<name>/` package
2. Implement the `Provider` interface:
   ```go
   func (p *Provider) Name() string
   func (p *Provider) Search(
       ctx context.Context,
       q model.Query,
   ) ([]model.Session, error)
   ```
3. Register in `cmd/search.go` `resolveProvider()`
4. Add table-driven tests in `<name>_test.go`

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
