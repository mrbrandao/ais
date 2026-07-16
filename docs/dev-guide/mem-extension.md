# mem Extension — Developer Guide

The `mem` extension is mental's built-in memory system. This guide explains
how it works, how to change its structure, and how to extend it.

---

## How it works

The mem extension stores cross-session memory as plain files under
`$MENTAL_DIR/projects/<project>/`:

```
$MENTAL_DIR/
└── projects/
    └── <project>/
        ├── MEMORY.md          rolling summary, rewritten on every save
        ├── tasks.yaml         task contract shared across sessions
        ├── topics.yaml        search index: topic → checkpoint files
        └── checkpoints/
            └── YYYY-MM-DD-HH-MM-SS.md   one file per session
```

### Session workflow

```
Session start
  mental mem load <project>
    → reads MEMORY.md + tasks.yaml
    → prints formatted context to stdout
    → LLM reads it and picks up where the last session stopped

During session
  mental mem task add "write rollback script"
    → appends task to tasks.yaml with auto-generated ID

Session end
  mental mem save
    → LLM fills in checkpoint content via stdin
    → mental rewrites MEMORY.md (current state, ≤50 lines)
    → writes new checkpoint file (timestamped, never modified again)
    → appends new topics to topics.yaml

Cross-session search
  mental mem search "rollback strategy"
    → parses topics.yaml for matching topic keywords
    → prints checkpoint file names + one-line summaries
    → LLM loads specific checkpoint for full detail
```

---

## Changing the layout

All file and directory names come from `mem.config.yaml`. To change the
name of any file or directory:

1. Open `internal/extensions/mem/config.yaml`.
2. Change the relevant key under `layout:`.
3. Run `make test` — no code changes required.

**Example: rename `topics.yaml` to `index.yaml`:**

```yaml
# in config.yaml
layout:
  project:
    topics: index.yaml  # was: topics.yaml
```

The change takes effect immediately for all new operations. Existing
`topics.yaml` files must be renamed manually in each project directory.

---

## Changing the checkpoint filename format

The `format` field uses Go time layout syntax. See
https://pkg.go.dev/time#Layout for the reference values.

```yaml
# in config.yaml
layout:
  project:
    checkpoints:
      format: "2006-01-02-15-04-05"   # default: 2026-07-15-11-00-00.md
      # format: "2006-01-02T15-04-05"  # ISO variant
      # format: "2006-01-02"           # date-only (one file per day)
```

Changing format affects only new checkpoints. Existing files keep their
original names and are still discoverable via topics.yaml.

---

## Adding a required MEMORY.md section

```yaml
# in config.yaml
memory:
  sections:
    - name: context
      required: true
    - name: blockers       # new section
      required: true
```

The save command reads this list and instructs the LLM to populate
`## Blockers` in MEMORY.md. No Go code changes needed.

---

## Adding a required checkpoint field

```yaml
# in config.yaml
checkpoint:
  frontmatter:
    required:
      - project
      - date
      - session
      - topics
      - files
      - duration_minutes   # new required field
```

The save command validates that `duration_minutes` is present in the
checkpoint's frontmatter before writing. If missing, the save fails with
a descriptive error. The LLM must provide the value.

---

## Adding a task status

```yaml
# in config.yaml
tasks:
  statuses:
    - todo
    - in_progress
    - blocked
    - paused          # new status
    - done
```

The task commands (`mental mem task done`, etc.) accept `paused` immediately.
No code changes needed.

---

## Changing the task ID format

```yaml
# in config.yaml
tasks:
  id_prefix: task-    # was: t
  id_padding: 4       # was: 3
  # produces: task-0001, task-0002, ...
```

Existing task IDs are not renamed. New tasks use the updated format.

---

## Adding a new command to the mem extension

If you need a new operation not covered by the YAML config (e.g.,
`mental mem archive`):

1. Add a new file `internal/extensions/mem/archive.go`.
2. Implement the operation function there.
3. Register the Cobra command in `cmd/mem.go`.
4. Add table-driven tests in `internal/extensions/mem/archive_test.go`.

Follow the same pattern as the existing `init.go`, `load.go`, etc.

---

## File format reference

### MEMORY.md

Max 50 lines (configurable via `memory.max_lines`). Rewritten on every
`mental mem save`. Structure:

```markdown
# <project>
status: active
updated: 2026-07-15T11:00:00
session: <id> | client: opencode | model: <model>
dir: /path/to/project

## Context
<3-5 sentences about current state>

## Decisions
- <decision>: <rationale>

## Next Steps
- [ ] <action>

## Related
- <project>: <why>

## Search
Topics and past sessions: topics.yaml
Command: mental mem search "<topic>"
```

### topics.yaml

Append-only. Never rewritten. Root key is `memory`:

```yaml
memory:
  - name: <session description>
    checkpoint: <filename>
    summary: <one sentence>
    topics:
      - <keyword>
```

### tasks.yaml

Patched in place on task operations. Root key is `tasks`:

```yaml
tasks:
  - id: t001
    title: <title>
    description: <multiline context>
    status: todo | in_progress | blocked | done
    session: <id>        # present when in_progress
    client: opencode     # present when in_progress
    blocked_by: <id>     # present when blocked
    completed: YYYY-MM-DD  # present when done
    ref:
      - <checkpoint or path or URL>
    subtasks:
      - id: t001a
        title: <title>
        status: done
```

### Checkpoint files

Written once, never modified. YAML frontmatter + Markdown body:

```
---
project: <name>
date: 2026-07-15T11:00:00
session:
  id: <id>
  client: opencode
  model: <model>
  dir: /path/to/project
topics:
  - <keyword>
files:
  - <relative/path>
---

## What We Did

## Decisions Made

## Open Questions

## Handoff
<1-2 sentences: exactly where to pick up>
```
