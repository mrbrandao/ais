# ais — AI session manager

![CI](https://github.com/mrbrandao/ais/actions/workflows/ci.yml/badge.svg)
![Release](https://img.shields.io/github/v/release/mrbrandao/ais)
![Go version](https://img.shields.io/github/go-mod/go-version/mrbrandao/ais)
![License](https://img.shields.io/github/license/mrbrandao/ais)
![Coverage](https://img.shields.io/badge/coverage-35.0%-lightgrey)

Search, export, and manage sessions across AI assistants
from a single CLI.

## Install

**curl (recommended):**
```bash
curl -sSfL \
  https://raw.githubusercontent.com/mrbrandao/ais/main/install.sh \
  | bash
```

**Go install:**
```bash
go install github.com/mrbrandao/ais@latest
```

**Container (no Go needed):**
```bash
make container-binary   # extracts bin/ais via podman
```

## Quick start

```bash
# Search OpenCode sessions
ais search -a opencode -s "my topic"

# Multiple search terms
ais search -a opencode -s "topic" -s "branch-name"

# Deep search (scans message content)
ais search -a opencode --type=deep --branch feat/my-branch

# Filter by directory
ais search -a opencode --dir /path/to/project

# JSON output
ais search -a opencode -s "topic" --output json

# Restore a session (from output)
opencode --session <session-id>
```

## Supported assistants

| Assistant | Status    |
|-----------|-----------|
| opencode  | supported |
| claude    | planned   |
| gemini    | planned   |
| cursor    | planned   |

## Build from source

```bash
git clone https://github.com/mrbrandao/ais.git
cd ais
make          # builds bin/ais
make install  # installs to /usr/local/bin
```

See [docs/dev.md](docs/dev.md) for full developer setup.

## License

Apache 2.0
