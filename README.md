# skills-pkg

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/mazrean/skills-pkg)](https://goreportcard.com/report/github.com/mazrean/skills-pkg)

**skills-pkg** is a CLI package manager for agent skills — reusable instruction sets and prompt libraries for AI coding agents such as Claude Code, Codex CLI, Cursor, and more.

## Features

- **Unified skill management** — one config file works across multiple agents
- **Multiple source types** — install from Git repositories or Go module paths
- **Hash-based integrity verification** — detect tampered or corrupted skills
- **Agent-aware install paths** — automatically resolves per-agent directories
- **Multi-target installs** — deploy a skill to several agent directories at once
- **Go module integration** — version is resolved from `go.mod` automatically, keeping skills in sync with library dependencies

## Quick Start

```sh
# 1. Create a configuration file for Claude Code (project-level)
skills-pkg init --agent claude

# 2. Add a skill from a Git repository
skills-pkg add frontend-design --url https://github.com/anthropics/skills

# 3. Install all skills listed in the configuration
skills-pkg install

# 4. Verify integrity of installed skills
skills-pkg verify
```

After running `init`, a `.skillspkg.toml` file is created in the current directory. This file tracks all skills and their install targets.

## Supported Agents

| Agent | Project-level path | Global path (`--global`) |
|---|---|---|
| `claude` | `.claude/skills/` | `~/.claude/skills/` |
| `codex` | `.codex/skills/` | `~/.codex/skills/` |
| `cursor` | `.cursor/skills/` | `~/.cursor/rules/` |
| `copilot` | `.copilot/skills/` | `~/.github/skills/` |
| `goose` | `.goose/skills/` | `~/.config/goose/skills/` |
| `opencode` | `.opencode/skills/` | `~/.config/opencode/skill/` |
| `gemini` | `.gemini/skills/` | `~/.gemini/skills/` |
| `amp` | `.amp/skills/` | `~/.config/agents/skills/` |
| `factory` | `.factory/skills/` | `~/.factory/skills/` |

## Installation

### Homebrew (macOS / Linux)

```sh
brew install mazrean/tap/skills-pkg
```

### Pre-built binaries

Download the latest binary for your platform from the [Releases](https://github.com/mazrean/skills-pkg/releases) page.

**Linux / macOS**

```sh
# Extract and move to PATH
tar -xzf skills-pkg_Linux_x86_64.tar.gz
sudo mv skills-pkg /usr/local/bin/
```

**Windows**

Extract `skills-pkg_Windows_x86_64.zip` and place `skills-pkg.exe` somewhere on your `PATH`.

### Debian / Ubuntu (.deb)

```sh
# Download the .deb package from the Releases page, then:
sudo dpkg -i skills-pkg_amd64.deb
```

### RHEL / Fedora (.rpm)

```sh
sudo rpm -i skills-pkg_amd64.rpm
```

### Alpine Linux (.apk)

```sh
sudo apk add --allow-untrusted skills-pkg_amd64.apk
```

### Build from source

```sh
git clone https://github.com/mazrean/skills-pkg.git
cd skills-pkg
go build -o skills-pkg .
```

## Commands

| Command | Description |
|---|---|
| `init` | Create a new `.skillspkg.toml` configuration file |
| `add <name>` | Add a skill to configuration and install it |
| `install [names...]` | Install skills from configuration |
| `update [names...]` | Update skills to their latest versions |
| `uninstall <name>` | Remove a skill from configuration and all install targets |
| `list` | List all configured skills |
| `verify` | Verify the integrity of all installed skills |
| `setup-ci` | Generate CI configuration for automated skill updates (GitHub Actions and/or Renovate) |

Use `skills-pkg <command> --help` for detailed options.

## Configuration File

skills-pkg is configured via `.skillspkg.toml` in the project root:

```toml
install_targets = ['./.claude/skills']

[[skills]]
name    = "my-skill"
source  = "git"
url     = "https://github.com/example/skills-repo"
version = "v1.2.0"
subdir  = "skills/my-skill"
hash_value = "h1:abc123..."
```

See [docs/configuration.md](docs/configuration.md) for the full reference.

## Documentation

- [Getting Started](docs/getting-started.md) — step-by-step guide for first-time users
- [Configuration Reference](docs/configuration.md) — all fields in `.skillspkg.toml`
- [Command Reference](docs/commands.md) — detailed options for every command
- [Go Module Integration](docs/go-module-integration.md) — version resolution, `GOPROXY` support, and packaging skills as a Go module

## License

MIT — see [LICENSE](LICENSE)
