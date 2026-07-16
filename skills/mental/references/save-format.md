# mental mem save — Input Format Reference

This file documents the complete format for `mental mem save` stdin input.
Load this reference when building the save block at session end.

## Complete format

```
project: <project-name>
session.id: <current-session-id>
session.client: opencode|claude|cursor|aider|windsurf
session.model: <model-name>
session.dir: /path/to/project
topics: auth migration, rollback strategy, postgresql schema
files: path/to/file.go, other/file.go
summary: One sentence describing what this session accomplished.
memory:
# <project-name>
status: active
updated: 2026-07-15T11:00:00

## Context
<3-5 sentences about the current project state after this session>

## Decisions
- <decision>: <rationale>

## Next Steps
- [ ] <next action>

## Related
<related projects or tasks, if any>

## Search
Topics and past sessions: topics.yaml
Command: mental mem search "<topic>" --project <project-name>
---
## What We Did
<paragraph summary of what happened in this session>

## Decisions Made
- <decision and rationale>

## Open Questions
- <unresolved question if any>

## Handoff
<1-2 sentences: exactly where to pick up in the next session>
```

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `project` | yes | Project name (must match an initialised project) |
| `session.id` | yes | Current session ID |
| `session.client` | yes | AI client being used |
| `session.model` | yes | Model name |
| `session.dir` | yes | Working directory path |
| `topics` | yes | 3-5 comma-separated keywords for search indexing |
| `files` | no | Changed files, comma-separated |
| `summary` | no | One-sentence description (used in topics index) |
| `memory:` | no | If present: full MEMORY.md content to write |
| `---` separator | no | Separates memory content from checkpoint body |
| Body sections | no | Markdown sections after `---` |

## Notes

- The `memory:` section rewrites MEMORY.md completely. Omit it to
  keep MEMORY.md unchanged (e.g., for quick mid-session checkpoints).
- `topics` are used to update `topics.yaml` — make them specific and
  searchable (e.g., "auth migration" not "work done").
- The `---` separator is required only when providing both memory
  content AND a checkpoint body.
- `status` in MEMORY.md can be: `active`, `paused`, or `done`.

## Minimal save (quick checkpoint, no memory update)

```
project: myproject
session.id: abc-123
session.client: opencode
session.model: claude-sonnet-4-6
session.dir: /home/user/dev/myproject
topics: quick checkpoint
summary: Quick mid-session checkpoint.
```

## Provider mode (no stdin needed)

```bash
# Raw checkpoint from OpenCode session — no LLM required
mental mem save -a opencode -s <session-id>

# Print LLM prompt, pipe to LLM, pipe result back to save
mental mem save -a opencode -s <session-id> -p | claude -p | mental mem save

# With Ollama (local LLM)
mental mem save -a opencode -s <session-id> -p | ollama run llama3 | mental mem save
```
