---
name: mental
description: >-
  Use this skill to manage cross-session memory and AI session history
  for LLM workflows. Activate when the user says: "save memory",
  "checkpoint", "wrapping up", "load context", "what were we doing",
  "search memory", "add task", "mark done", "show tasks", "init memory",
  "start tracking", "mental mem", or asks about past sessions or session
  history — even without mentioning "mental" by name.
  SKIP: if mental binary is not in PATH (check with: which mental).
license: Apache-2.0
compatibility: >-
  Requires mental binary in PATH.
  Install: go install github.com/mrbrandao/mental@latest
  MENTAL_DIR defaults to ~/.local/share/mental.
allowed-tools: Bash
metadata:
  author: mrbrandao
  version: "0.2.0"
---

# mental

AI Session Manager — search session history and manage
cross-session memory for LLM workflows.

## Before you start

Check the binary exists:

```bash
which mental || echo "mental not found — install it first"
```

If not found, stop and tell the user to install:
```bash
go install github.com/mrbrandao/mental@latest
```

Determine the project name from the current directory if not obvious:
```bash
basename $(git rev-parse --show-toplevel 2>/dev/null || pwd)
```

## Step 1 — Session Start: Load Context

When starting work on a known project:

```bash
mental mem load <project>
```

Read the output carefully — it contains the current MEMORY.md state and
task list. Use it to orient the session without asking the user to
re-explain context. If the output says "not found", run init first.

## Step 2 — During Session: Manage Tasks

Add tasks as they are identified:

```bash
mental mem task add --project <project> "Task description"
# Output: Added task #t001: Task description
```

Mark tasks done when completed:

```bash
mental mem task done --project <project> t001
```

List all tasks at any time:

```bash
mental mem task list --project <project>
```

## Step 3 — Session End: Save Checkpoint

When the user says "wrapping up", "done for today", or "save memory":

1. Build the save input block (see `references/save-format.md` for
   the full annotated format).
2. Write it to a temp file and pipe to mental:

```bash
mental mem save < /tmp/mental-save.txt
```

3. Verify the save succeeded:

```bash
mental mem load <project> | head -5
# If the updated timestamp appears, save was successful.
```

**Alternatively — pipe through an installed LLM for synthesis:**

```bash
# Generate a prompt for an LLM, pipe to it, save the result:
mental mem save -a opencode -s <session-id> -p | claude -p | mental mem save

# Raw checkpoint from OpenCode (no LLM, no synthesis):
mental mem save -a opencode -s <session-id>

# Find the current OpenCode session ID:
mental session search -a opencode -s "$(basename $(pwd))" --output json \
  | head -5
```

## Step 4 — Search Past Sessions

```bash
# Search memory by topic
mental mem search "rollback strategy" --project <project>

# Search OpenCode session history
mental session search -a opencode -s "topic keyword"
mental session search -a opencode --type=deep --branch feat/my-feature
```

## Step 5 — Init a New Project

When starting a project for the first time:

```bash
mental mem init <project>
# Creates: ~/.local/share/mental/projects/<project>/
# Files:   MEMORY.md, tasks.yaml, topics.yaml, checkpoints/
```

## Step 6 — Extensions

```bash
mental extensions list           # shows all installed extensions
mental extensions describe memx  # details for built-in memx engine
```

## Gotchas

- If `mental` is not in PATH, stop immediately — do not guess commands.
- The `--project` flag is required for search, task, and list commands.
- In STDIN mode, `mental mem save` reads until EOF — always pipe from
  a file or an LLM output, never type interactively.
- When using provider mode (`-a opencode -s <id>`), MEMORY.md is NOT
  updated — only the checkpoint and topics.yaml are written.
- `MENTAL_DIR` overrides the data directory for all commands.
- Load `references/save-format.md` for the complete checkpoint format
  when building the save input block.
