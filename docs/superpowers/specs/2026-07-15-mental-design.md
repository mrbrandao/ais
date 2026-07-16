# Mental — Design Specification

**Date**: 2026-07-15
**Status**: Approved — implementation in progress
**Author**: Igor Brandao + OpenCode session

---

## 1. Overview

Mental is a cross-session memory and task management CLI for LLM-based workflows.
It solves the context loss problem: when a session ends, the LLM forgets everything.
Mental persists structured memory across sessions, projects, and agents.

### Goals

- Exchange knowledge from one LLM session to another without re-explaining context
- Allow multiple agents working on related tasks to share and cross-reference knowledge
- Provide checkpoint-based project continuity for long-running multi-session work
- Be portable across LLM clients: OpenCode, Claude, Cursor, any tool that reads files
- Extend via a plugin system so external memory backends can replace or augment the default

### Non-Goals (for this version)

- Real-time communication between concurrent agents
- Cloud sync (files are local; git handles distribution)
- Automatic session end detection (triggers are explicit or defined in AGENTS.md)

---

## 2. CLI Command Structure

```
mental
├── session
│   ├── search          # search OpenCode/Claude session history
│   └── (copy)          # future: copy session context to another session
│
├── mem                 # built-in memory extension
│   ├── init <project>  # create XDG memory structure for a project
│   ├── load <project>  # read MEMORY.md + tasks.yaml into LLM context
│   ├── save            # rewrite MEMORY.md + write checkpoint + update topics.yaml
│   └── search <topic>  # search topics.yaml, return matching checkpoints
│   └── task
│       ├── add <title> # append a task to tasks.yaml
│       ├── done <id>   # mark a task done
│       └── list        # list all tasks for the current project
│
├── extensions
│   ├── list            # list installed extensions (internal + external)
│   └── describe <name> # show extension manifest
│
└── status              # current project state: memory summary + task counts
```

---

## 3. Memory Protocol

The memory protocol defines how mental stores and retrieves context. It is
intentionally file-based so any LLM tool can read and write it without a
running mental binary.

### 3.1 Directory Layout

```
$MENTAL_DIR/                            # default: ~/.local/share/mental
└── projects/
    └── <project-name>/
        ├── MEMORY.md                   # rolling summary — always loaded at session start
        ├── tasks.yaml                  # task contract — shared between sessions and agents
        ├── topics.yaml                 # search index — maps topics to checkpoint files
        └── checkpoints/
            └── YYYY-MM-DD-HH-MM-SS.md # one file per session
```

`MENTAL_DIR` can be overridden by the environment variable `MENTAL_DIR`.
Default follows XDG Base Directory Specification: `~/.local/share/mental`.

### 3.2 MEMORY.md — Rolling Summary

Always loaded at session start. Rewritten (not appended) on every `mental mem save`.
Maximum 50 lines. Contains the current state of the project.

```markdown
# <project-name>
status: active | paused | done
updated: 2026-07-15T11:00:00
session: <id> | client: opencode | model: claude-sonnet-4-6
dir: /home/user/dev/projects/foo

## Context
<3-5 sentences: what this project is about and current state>

## Decisions
- <decision>: <rationale>

## Next Steps
- [ ] <next action>

## Related
- <other-project>: <why related>

## Search
Topics and past sessions: topics.yaml
Command: mental mem search "<topic>" [--after YYYY-MM-DD]
```

### 3.3 tasks.yaml — Task Contract

The transferable task list. Any session or agent can import, update, and mark tasks.
Updated by `mental mem task` commands. Never rewritten in full — only patched.

```yaml
tasks:
  - id: t001
    title: Write rollback script for auth migration
    description: >
      The auth migration needs a rollback script that reverts schema
      changes if the migration fails in production.
    status: in_progress      # todo | in_progress | blocked | done
    session: abc-123
    client: opencode
    ref:
      - checkpoints/2026-07-15-11-00-00.md

  - id: t002
    title: Test migration in staging
    description: >
      Run the full migration in staging and validate all auth
      endpoints work correctly after the schema change.
    status: blocked
    blocked_by: t001
    ref:
      - checkpoints/2026-07-14-09-15-22.md

  - id: t003
    title: Auth service database schema
    status: done
    completed: 2026-07-14
    subtasks:
      - id: t003a
        title: Design users table
        status: done
      - id: t003b
        title: Design sessions table
        status: done
```

### 3.4 topics.yaml — Search Index

Maps topic keywords to checkpoint files with one-line summaries.
Appended on every `mental mem save`. Never rewritten — only appended.
The `memory` root key is intentional: this file IS the navigable memory index.

```yaml
memory:
  - name: auth service migration
    checkpoint: 2026-07-15-11-00-00.md
    summary: Defined two-phase rollback. Schema finalized. Rollback script incomplete.
    topics:
      - auth migration
      - rollback strategy
      - postgresql schema

  - name: initial schema design
    checkpoint: 2026-07-14-09-15-22.md
    summary: PostgreSQL chosen over MySQL. Auth service owns users table.
    topics:
      - auth migration
      - postgresql schema
      - database design
```

### 3.5 Checkpoint Files — Session Record

One file per session, written once, never modified.
Filename format: `YYYY-MM-DD-HH-MM-SS.md` (sortable, date embedded in name).

```markdown
---
project: foo
date: 2026-07-15T11:00:00
session:
  id: abc-123
  client: opencode        # opencode | claude | cursor | aider | windsurf
  model: claude-sonnet-4-6
  dir: /home/user/dev/foo
topics:
  - auth migration
  - rollback strategy
  - postgresql schema
files:
  - auth/migration.sql
  - auth/rollback.sql
---

## What We Did

## Decisions Made

## Open Questions

## Handoff
<1-2 sentences: exactly where to pick up next session>
```

### 3.6 Configuration — mem.config.yaml

The layout, file names, and structure are driven by `mem.config.yaml`, embedded
in the binary as a default. Override by placing a `config.yaml` at:
`$MENTAL_DIR/extensions/mem/config.yaml`

Changing this file changes mental's behavior without recompilation.
See `internal/extensions/mem/memx/config.yaml` for the fully annotated default.

---

## 4. Extension Architecture

### 4.1 Two Extension Types

**Internal extensions** are compiled into the mental binary. They live in
`internal/extensions/<name>/` and register with the manager at startup.
The `mem` extension is the only built-in. Internal extensions follow the same
interface as external ones, enabling future promotion or demotion.

**External extensions** are independent executables discovered at runtime.
They live in `$MENTAL_DIR/extensions/<name>/` and declare themselves via
an `extension.yaml` manifest. Any language can implement an external extension.

### 4.2 Extension Manifest (extension.yaml)

Each external extension ships an `extension.yaml` in its directory:

```yaml
name: "Hermes Holographic Memory"
type: memory          # memory | task | search
description: "Holographic memory system for mental"
executable: mental-hermes
author: NousResearch
version: "0.1.0"
mode: structured      # structured | passthrough
```

**mode: passthrough** — mental wires stdin/stdout/stderr directly to the terminal.
The plugin owns the output. No data flows back to mental.

**mode: structured** — mental captures stdout as JSON. The plugin returns structured
results that mental can process, format, or pipe to another extension.

### 4.3 JSON Data Exchange Protocol

External extensions in structured mode communicate via JSON on stdin/stdout.

Mental writes to the extension's stdin:
```json
{
  "query": "rollback strategy",
  "project": "foo",
  "mental_dir": "/home/user/.local/share/mental",
  "mental_version": "0.2.0"
}
```

The extension writes results to stdout:
```json
{
  "results": [
    {
      "file": "2026-07-15-11-00-00.md",
      "summary": "Defined two-phase rollback approach",
      "relevance": 0.92
    }
  ]
}
```

Mental reads this JSON and formats it for display or pipes it to another extension.

### 4.4 Environment Variables

Mental injects these into every external extension process:

| Variable | Value |
|----------|-------|
| `MENTAL_DIR` | `$MENTAL_DIR` (resolved XDG path) |
| `MENTAL_PROJECT` | current active project name |
| `MENTAL_VERSION` | mental binary version |
| `MENTAL_CONFIG` | path to mental's viper config file |

### 4.5 External Extension Discovery

At startup, mental scans `$MENTAL_DIR/extensions/` for subdirectories containing
`extension.yaml`. The directory name is the extension's identifier. Duplicate
names: first found wins, warning displayed.

```
$MENTAL_DIR/extensions/
└── hermes/
    ├── extension.yaml
    └── mental-hermes    (executable)
```

`mental extensions list` shows all discovered extensions with their type and mode.

---

## 5. Code Architecture

### 5.1 Package Layout

```
mental/
├── main.go
├── cmd/                              # Cobra commands — zero business logic
│   ├── root.go
│   ├── session/
│   │   └── search.go
│   ├── extensions.go
│   └── status.go
└── internal/
    ├── model/                        # session search types (unchanged from ais)
    │   ├── query.go
    │   └── session.go
    ├── output/                       # session output formatters
    │   └── output.go
    ├── provider/                     # session search backends
    │   ├── provider.go
    │   └── opencode/
    ├── config/                       # XDG resolution + viper config
    │   └── config.go
    └── extensions/                   # extension system
        ├── extension.go              # Extension interface + manifest types
        ├── manager.go                # registry: internal + XDG external scan
        ├── runner.go                 # external subprocess execution
        └── mem/                      # built-in mem extension
            ├── mem.go
            ├── config.go
            ├── config.yaml           # embedded default (go:embed)
            ├── types.go              # Checkpoint, Task, Topic, ProjectContext
            ├── init.go
            ├── load.go
            ├── save.go
            ├── search.go
            └── task.go
```

### 5.2 Key Design Decisions

- `internal/` throughout — all packages are internal; external plugins communicate
  via JSON/stdin/stdout, not Go imports
- `internal/model/` kept for session search types (future refactor noted)
- Mem extension types are local to `internal/extensions/mem/` — not shared
- `mem.config.yaml` drives layout, file names, and structure — no hardcoded paths
- All config via Viper — `$MENTAL_DIR`, `MENTAL_PROJECT`, and extension config
- Functions follow Effective Go: small, single-purpose, clear names
- Every exported symbol has a godoc comment
- Table-driven tests for all logic

---

## 6. Release Pipeline

| Component | Tool | Output |
|-----------|------|--------|
| Binary archives + checksums | goreleaser | Attached to GitHub release |
| GitHub release page notes | `changelog.use: github-native` | Auto-generated, PR-linked |
| `CHANGELOG.md` in repo | git-cliff + `cliff.toml` | Keep a Changelog format |
| Release trigger | `push tag v*.*.*` | Starts CI workflow |
| Tag creation | github-release skill | Signed annotated tag |

---

## 7. Implementation Phases

| Phase | Branch | Key deliverables |
|-------|--------|-----------------|
| 0 | — | Bootstrap `~/dev/gen/tasks/mental/` memory files |
| 1 | `feat/rename-ais-to-mental` | All ais→mental renames, goreleaser github-native, cliff.toml, CHANGELOG.md |
| 2 | `feat/command-structure` | `session search`, extension stubs, viper config |
| 3 | `feat/extension-architecture` | Extension interface, manager, runner, mem config.yaml, dev guide docs |
| 4 | `feat/mem-extension` | All `mental mem` subcommands implemented |
| 5 | `feat/external-extensions` | External subprocess runner, JSON protocol, discovery |
| 6 | `feat/skill-and-docs` | SKILL.md, updated AGENTS.md |

**Workflow per phase:**
```bash
git checkout -b feat/<name>
# commits (max 150 lines each)
git checkout main
git rebase feat/<name>
git branch -d feat/<name>
```

Single tag applied by maintainer after all phases complete.

---

## 8. LLM Trigger Vocabulary

Phrases that activate memory operations when the mental SKILL.md is loaded:

| User says | mental command |
|-----------|---------------|
| "save memory", "checkpoint", "wrapping up" | `mental mem save` |
| "load context", "what were we doing" | `mental mem load <project>` |
| "search memory for X" | `mental mem search "X"` |
| "add task", "create task for..." | `mental mem task add` |
| "mark done", "completed X" | `mental mem task done <id>` |
| "show tasks", "what's pending" | `mental mem task list` |
| "init memory", "start tracking" | `mental mem init <project>` |

---

## 9. Future Work (not in scope)

- `internal/model/` refactor into focused importable packages
- Semantic search via LanceDB or pgvector over checkpoint files
- Session end auto-detection (fsnotify + OpenCode SQLite watch)
- `mental session copy` — transfer context between sessions
- Managed memory backends (Zep Go SDK, mem0 REST sidecar)
- MCP server mode (`mental serve`) for tool-based LLM integration
