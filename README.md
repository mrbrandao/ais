# mental — cross-session memory and AI session manager

![CI](https://github.com/mrbrandao/mental/actions/workflows/ci.yml/badge.svg)
![Release](https://img.shields.io/github/v/release/mrbrandao/mental)
![Go version](https://img.shields.io/github/go-mod/go-version/mrbrandao/mental)
![License](https://img.shields.io/github/license/mrbrandao/mental)
![Coverage](https://img.shields.io/badge/coverage-35.0%-lightgrey)

Cross-session memory management and AI session search for LLM workflows.
mental persists context across sessions, tracks tasks, and lets multiple
agents share knowledge through a simple file-based protocol.

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

```bash
# Search OpenCode sessions
mental search -a opencode -s "my topic"

# Multiple search terms
mental search -a opencode -s "topic" -s "branch-name"

# Deep search (scans message content)
mental search -a opencode --type=deep --branch feat/my-branch

# Filter by directory
mental search -a opencode --dir /path/to/project

# JSON output
mental search -a opencode -s "topic" --output json
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
git clone https://github.com/mrbrandao/mental.git
cd mental
make          # builds bin/mental
make install  # installs to /usr/local/bin
```

See [docs/dev.md](docs/dev.md) for full developer setup.

## License

Apache 2.0
