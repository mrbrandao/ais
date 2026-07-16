# Contributing to mental

## Code standards

mental follows [Effective Go](https://go.dev/doc/effective_go) as the
primary standard. The rules below are enforced by pre-commit hooks and
CI.

| Rule | Detail |
|------|--------|
| Go version | 1.25+ |
| Line length | 80 characters — hard wrap |
| Errors | Last return value; wrap with `fmt.Errorf("context: %w", err)` |
| Context | `context.Context` always the first parameter |
| Cleanup | `defer` for all cleanup (`db.Close`, `rows.Close`, `cancel`) |
| Control flow | No `else` after `return` |
| Function size | Small and single-purpose; if a function needs a comment to explain what it does, split it |
| Exports | Every exported symbol has a godoc comment |
| Tests | Table-driven, in `_test.go` files alongside the code |
| SQLite | No CGO — use `modernc.org/sqlite` |

## Commit rules

Commits follow [Conventional Commits](https://www.conventionalcommits.org/)
and the [tpope body format](https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html):

- Prefix: `feat`, `fix`, `docs`, `test`, `chore`, `ci`, `refactor`
- Subject: imperative mood, ≤50 chars, no trailing period
- Body: wrapped at 72 chars, explains what and why (not how)
- Size: ≤150 lines changed per commit
- Staging: never `git add .` — always explicit file paths

```
feat: add mem init command

Creates the project directory structure under MENTAL_DIR including
MEMORY.md, tasks.yaml, topics.yaml, and the checkpoints/ directory.
Returns an error if the project already exists.
```

## Branch and integration workflow

```bash
# Create a feature branch from main
git checkout -b feat/my-feature

# Make commits (max 150 lines each)

# Integrate to main using rebase (linear history, no merge commits)
git checkout main
git rebase feat/my-feature
git branch -d feat/my-feature
```

Tags are applied by the maintainer after all work is complete and CI passes.

## Running tests

```bash
make test       # go test -race ./...
make vet        # go vet ./...
make lint       # golangci-lint run ./...
make coverage   # coverage report
```

All tests must pass before committing. The pre-commit hook runs
`make test` automatically.

## Adding a new package

Follow the existing layout:

- Logic belongs in `internal/` — never in `cmd/`
- Each package has one clear purpose
- Package name matches the directory name (no `_` in names)
- Start with the interface or type definition, then the implementation
- Write tests before or alongside the implementation

## Pre-commit hooks

Install once with `make hooks`. Hooks run on every `git commit`:

| Hook | Purpose |
|------|---------|
| gitleaks | Secret scanning |
| go-fmt | Code formatting |
| go-vet | Static analysis |
| golangci-lint | Linting |
| gosec | SAST security |
| snyk | Vulnerability scan |
