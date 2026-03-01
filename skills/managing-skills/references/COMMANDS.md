# skills-pkg Command Reference

Full reference for commands used by the managing-skills skill.

## Install skills-pkg

### Homebrew (macOS / Linux)

```sh
brew install mazrean/tap/skills-pkg
```

### Pre-built binary

Download from [GitHub Releases](https://github.com/mazrean/skills-pkg/releases), extract, and place on `PATH`.

Verify:

```sh
skills-pkg --version
```

---

## `search`

Search for available skills on [skills.sh](https://skills.sh).

```sh
skills-pkg search [query] [--limit N]
```

| Argument / Flag | Default | Description |
|---|---|---|
| `[query]` | *(empty)* | Keywords to filter skills. Omit to list all |
| `--limit N` | `10` | Maximum number of results |

**Output columns:**

| Column | Description |
|---|---|
| `NAME` | Skill identifier used as `<name>` in `add` |
| `SOURCE` | URL passed as `--url` to `add` |
| `INSTALLS` | Total install count (higher = more trusted) |

**Examples:**

```sh
# Search by keyword
skills-pkg search "frontend"

# List all available skills
skills-pkg search

# Increase result count
skills-pkg search "go" --limit 20
```

---

## `add`

Add a skill to the configuration and install it immediately.

```sh
skills-pkg add <name> --url <url> [flags]
```

| Argument / Flag | Default | Description |
|---|---|---|
| `<name>` | *(required)* | Unique name for this skill in the config |
| `--url <url>` | *(required)* | Git remote URL or Go module path |
| `--source <type>` | `git` | Source type: `git` or `go-mod` |
| `--version <ver>` | *(latest)* | Pinned version: tag, branch, commit SHA (git) or semver (go-mod) |
| `--sub-dir <path>` | `skills/<name>` | Subdirectory within the source containing skill files |

**Behavior:**

1. Reads `.skillspkg.toml` — fails if not found (run `init` first)
2. Checks `<name>` is not already registered
3. Downloads and copies skill files to all `install_targets`
4. Records `hash_value` and saves the updated config

**Examples:**

```sh
# From Git, latest tag (most common)
skills-pkg add my-skill --url https://github.com/example/skills-repo

# Pinned version
skills-pkg add my-skill --url https://github.com/example/skills-repo --version v2.0.0

# From Go module
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills

# Custom subdirectory
skills-pkg add my-skill --url https://github.com/example/repo --sub-dir prompts/my-skill
```

---

## `init`

Initialize `.skillspkg.toml` in the current directory.

```sh
skills-pkg init [flags]
```

| Flag | Short | Description |
|---|---|---|
| `--agent <name>` | `-a` | Add the agent's default directory. Valid: `claude`, `codex`, `cursor`, `copilot`, `goose`, `opencode`, `gemini`, `amp` |
| `--install-dir <path>` | `-d` | Add a custom install directory |
| `--global` | `-g` | Use agent's user-level (global) directory. Requires `--agent` |

**Examples:**

```sh
# Default (.skills directory)
skills-pkg init

# Claude Code project-level
skills-pkg init --agent claude

# Claude Code user-level (global)
skills-pkg init --agent claude --global

# Multiple agents
skills-pkg init --agent claude --agent codex
```

---

## `list`

List all installed skills in the current configuration.

```sh
skills-pkg list
```

---

## `update`

Update skills to their latest versions.

```sh
skills-pkg update [names...]
```

---

## `uninstall`

Remove a skill from configuration and delete installed files.

```sh
skills-pkg uninstall <name>
```

---

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error |
