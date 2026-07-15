# Developer Guide

## Prerequisites

| Tool | Required | Install |
|---|---|---|
| Go 1.25+ | yes | https://go.dev/dl |
| pre-commit | yes | https://pre-commit.com |
| golangci-lint | yes | `make dev-deps` |
| git-cliff | for releases | https://git-cliff.org |
| snyk | for pre-commit | `npm install -g snyk` |
| podman | for containers | https://podman.io |
| goreleaser | for releases | https://goreleaser.com |

## Setup

```bash
git clone https://github.com/mrbrandao/mental.git
cd mental
make dev-deps   # installs golangci-lint
make hooks      # installs pre-commit hooks
```

## Daily workflow

```bash
make          # build bin/mental
make test     # run tests
make lint     # run linter
make fmt      # format code
make coverage # coverage report
```

## Pre-commit hooks

Hooks run automatically on `git commit`:

| Hook | Purpose |
|---|---|
| gitleaks | secret scanning |
| go-fmt | formatting |
| go-vet | static analysis |
| golangci-lint | linting |
| gosec | SAST security scan |
| snyk | vulnerability scan |

snyk requires a separate install:
```bash
npm install -g snyk
snyk auth      # authenticate once
```

## Containers

```bash
# Build container image
make container-build

# Extract binary without Go installed
make container-binary  # produces bin/mental

# Run mental via container
make container-run ARGS="search -a opencode -s topic"
```

## Releases

Releases are automated via goreleaser on tag push.
The release workflow runs git-cliff to update CHANGELOG.md
(Keep a Changelog format), then goreleaser to build binaries
and create the GitHub release with auto-generated notes.

Dry-run locally:
```bash
make release-dry
```

## Branch workflow

```bash
# Create feature branch
git checkout -b feat/my-feature

# ... make commits (max 150 lines each) ...

# Integrate to main (linear history, no merge commits)
git checkout main
git rebase feat/my-feature
git branch -d feat/my-feature
```

## Adding a provider

See `AGENTS.md` — "How to add a new assistant backend".

## Adding an extension

See `AGENTS.md` — "How to create an external extension"
and `docs/dev-guide/extension-development.md`.
