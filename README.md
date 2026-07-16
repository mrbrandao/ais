# mental — AI Session Manager

![CI](https://github.com/mrbrandao/mental/actions/workflows/ci.yml/badge.svg)
![Release](https://img.shields.io/github/v/release/mrbrandao/mental)
![Go version](https://img.shields.io/github/go-mod/go-version/mrbrandao/mental)
![License](https://img.shields.io/github/license/mrbrandao/mental)
![Coverage](https://img.shields.io/badge/coverage-35.0%-lightgrey)

mental manages AI assistant sessions: search session history and manage
session context memory across sessions and providers. Extensible via
built-in and external extensions.

## Install

**curl (recommended):**
```bash
curl -sSfL \
  https://raw.githubusercontent.com/mrbrandao/mental/main/install.sh \
  | bash
```

**Go install:**
```bash
go install github.com/mrbrandao/mental@latest
```

**Container (no Go needed):**
```bash
make container-binary   # extracts bin/mental via podman
```

## Quick start

### Session search

```bash
# Search OpenCode sessions
mental session search -a opencode -s "my topic"

# Deep search (scans message content)
mental session search -a opencode --type=deep --branch feat/my-branch

# Filter by directory, JSON output
mental session search -a opencode --dir /path/to/project --output json
```

### Memory management

```bash
# Initialise memory for a project
mental mem init myproject

# Load context at session start (used by skills/LLMs)
mental mem load myproject

# Save checkpoint — three modes:

# 1. Stdin mode (used by skills and LLM pipes)
mental mem save < /tmp/checkpoint.txt

# 2. Provider mode: raw checkpoint from OpenCode session (no LLM needed)
mental mem save -a opencode -s <session-id>

# 3. Print mode: generate LLM prompt, pipe to any LLM
mental mem save -a opencode -s <session-id> -p | claude -p
mental mem save -a opencode -s <session-id> -p | ollama run llama3

# Full synthesis pipeline
mental mem save -a opencode -s <session-id> -p | claude -p | mental mem save

# Search past sessions by topic
mental mem search "rollback strategy" --project myproject

# Task management
mental mem task add --project myproject "Write rollback script"
mental mem task list --project myproject
mental mem task done --project myproject t001
```

### Extensions

```bash
# List installed extensions (built-in + external)
mental extensions list
mental extensions describe opencode
```

## Supported providers

| Provider | Session search | Memory extract | Status    |
|----------|---------------|----------------|-----------|
| opencode | supported     | supported      | built-in  |
| claude   | —             | —              | planned   |
| cursor   | —             | —              | planned   |

## Skills (LLM integration)

mental ships a skill that teaches any LLM assistant when and how to
invoke it. Install via [lola](https://github.com/mrbrandao/lola):

```bash
lola mod add https://github.com/mrbrandao/mental.git#subdirectory=skills
lola install mental
```

Once installed, your LLM assistant will automatically load context at
session start, manage tasks, and save checkpoints when you say natural
phrases like "wrapping up" or "what were we doing".

## Build from source

```bash
git clone https://github.com/mrbrandao/mental.git
cd mental
make          # builds bin/mental
make install  # installs to /usr/local/bin
```

See [docs/dev.md](docs/dev.md) for full developer setup.

## License

Apache 2.0
