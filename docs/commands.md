# Command Reference

All commands share the following global flags:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--help` | | | Show help |

The global `-v` flag can also be set via the `SKILLSPKG_VERBOSE` environment variable.

---

## `init`

Initialize a new `.skillspkg.toml` configuration file.

```
skills-pkg init [flags]
```

### Flags

| Flag | Short | Description |
|---|---|---|
| `--agent <name>` | `-a` | Add the agent's default skill directory as an install target. Can be specified multiple times. Valid values: `claude`, `codex`, `cursor`, `copilot`, `goose`, `opencode`, `gemini`, `amp`, `factory` |
| `--install-dir <path>` | `-d` | Add a custom directory as an install target. Can be specified multiple times |
| `--global` | `-g` | Use the agent's user-level (global) directory instead of the project-level one. Requires `--agent` |

### Behavior

- Writes `.skillspkg.toml` in the current directory
- Fails if the file already exists
- If neither `--agent` nor `--install-dir` is given, defaults to `./.skills`
- With `--agent` (no `--global`), adds `./.{agent}/skills` (e.g., `./.claude/skills`)
- With `--agent --global`, resolves the agent's global path (e.g., `~/.claude/skills`)

### Examples

```sh
# Default (project-level .skills directory)
skills-pkg init

# Claude Code, project-level
skills-pkg init --agent claude

# Claude Code, user-level
skills-pkg init --agent claude --global

# Multiple agents
skills-pkg init --agent claude --agent codex

# Custom directory
skills-pkg init --install-dir ./shared/skills

# Mix of agent and custom
skills-pkg init --agent claude --install-dir ./shared/skills
```

---

## `add`

Add a skill to the configuration and install it immediately.

```
skills-pkg add <name> --url <url> [flags]
```

### Arguments

| Argument | Description |
|---|---|
| `<name>` | Unique name for this skill in the configuration |

### Flags

| Flag | Default | Description |
|---|---|---|
| `--url <url>` | *(required)* | Git remote URL or Go module path |
| `--source <type>` | `git` | Source type: `git` or `go-mod` |
| `--version <ver>` | | Pinned version. For `git`: tag, branch, or commit SHA; defaults to the latest tag. For `go-mod`: semver or pseudo-version; defaults to the version found in the nearest `go.mod`, then falls back to the latest from the module proxy |
| `--sub-dir <path>` | `skills/<name>` | Subdirectory within the source that contains the skill files |
| `--print-skill-info` | `false` | After installation, print skill name, description, and file path in agent-readable format (Codex-compatible) |

### Behavior

1. Reads the existing `.skillspkg.toml` (fails if not found — run `init` first)
2. Checks that `<name>` is not already registered (fails if duplicate)
3. Downloads and copies the skill files to all `install_targets`
4. Records `hash_value` and saves the updated config

If installation fails, the skill entry is **not** written to the config, leaving the file unchanged.

### Examples

```sh
# Add from Git (latest tag)
skills-pkg add my-skill --url https://github.com/example/skills-repo

# Add and print skill metadata for agent awareness
skills-pkg add my-skill --url https://github.com/example/skills-repo --print-skill-info

# Add a specific version
skills-pkg add my-skill --url https://github.com/example/skills-repo --version v2.0.0

# Custom subdirectory
skills-pkg add my-skill --url https://github.com/example/skills-repo --sub-dir prompts/my-skill

# From Go module (version resolved from go.mod if present, otherwise latest from proxy)
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills

# From Go module — always use latest, ignoring go.mod
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills --version latest

# From Go module with pinned version
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills --version v1.3.0
```

> **Go Module version resolution:** When `--source go-mod` is used without `--version`, skills-pkg first searches for the module in the nearest `go.mod` file (walking up the directory tree). If found, that version is used so the skill stays in sync with your Go dependency graph. If not found, the latest version is fetched from the module proxy. See [Go Module Integration](go-module-integration.md) for more details.

---

## `install`

Install skills from the configuration file.

```
skills-pkg install [names...] [flags]
```

### Arguments

| Argument | Description |
|---|---|
| `[names...]` | Skill names to install. If omitted, all skills in the config are installed |

### Behavior

- For each specified (or all) skill, downloads the files at the pinned `version`
- Copies the files to all `install_targets`
- Verifies the hash after copying; fails if there is a mismatch
- Does **not** modify `.skillspkg.toml`

### Examples

```sh
# Install all skills
skills-pkg install

# Install specific skills only
skills-pkg install my-skill other-skill
```

---

## `update`

Update skills to their latest versions.

```
skills-pkg update [names...] [flags]
```

### Arguments

| Argument | Description |
|---|---|
| `[names...]` | Skill names to update. If omitted, all skills are updated |

### Flags

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Show what would be updated without making any changes |
| `--output <format>` | `text` | Output format: `text` (human-readable) or `json` (machine-readable, written to stdout) |

### Behavior

- For each target skill, resolves the latest available version (latest Git tag, or latest module version)
- Downloads and installs the new version
- Updates `version` and `hash_value` in `.skillspkg.toml`
- With `--dry-run`, no files or config are modified; results are printed only
- With `--output json`, the result is written to **stdout** as a JSON object; progress messages go to stderr

### JSON output schema

```json
{
  "updates": [
    {
      "skill_name": "my-skill",
      "current_version": "v1.0.0",
      "latest_version": "v2.0.0",
      "has_update": true,
      "file_diffs": [
        { "path": "SKILL.md", "status": "modified", "patch": "..." }
      ]
    }
  ]
}
```

`file_diffs[].status` is one of `added`, `removed`, or `modified`.

### Examples

```sh
# Update all skills
skills-pkg update

# Update a specific skill
skills-pkg update my-skill

# Check for updates without applying them (text output)
skills-pkg update --dry-run

# Check for updates and emit JSON (suitable for scripting or CI)
skills-pkg update --dry-run --output json > updates.json
```

---

## `uninstall`

Remove a skill from the configuration and delete its installed files.

```
skills-pkg uninstall <name> [flags]
```

### Arguments

| Argument | Description |
|---|---|
| `<name>` | Name of the skill to remove |

### Behavior

- Deletes the skill's subdirectory from every `install_target`
- Removes the `[[skills]]` entry from `.skillspkg.toml`

### Example

```sh
skills-pkg uninstall my-skill
```

---

## `list`

List all skills configured in `.skillspkg.toml`.

```
skills-pkg list [flags]
```

Prints each skill's name, source type, URL, and pinned version.

### Example

```sh
skills-pkg list
```

---

## `verify`

Verify the integrity of installed skills by comparing their content against the recorded hashes.

```
skills-pkg verify [flags]
```

### Behavior

- Reads `hash_value` for each skill from `.skillspkg.toml`
- Recomputes the hash of the files currently in each `install_target`
- Reports any mismatch
- Exits with code `1` if any skill fails verification; `0` if all pass

### Example

```sh
skills-pkg verify
```

---

## `setup-ci`

Generate CI configuration for automated skill updates.

```
skills-pkg setup-ci [flags]
```

At least one flag must be specified.

### Flags

| Flag | Description |
|---|---|
| `--github-actions` | Create `.github/workflows/update-skills.yml` — a GitHub Actions workflow that detects and applies skill updates via pull requests |
| `--renovate` | Add a JSONata custom manager entry to `renovate.json` that tracks skill versions through the Renovate bot |

### `--github-actions`

Creates `.github/workflows/update-skills.yml` with the following workflow:

1. **Detect updates** — runs `skills-pkg update --dry-run --output json` and identifies skills that have a newer version available
2. **Update in parallel** — for each skill with an update, creates a dedicated Git branch using `git worktree` and runs `skills-pkg update <skill>` in it concurrently
3. **Open PRs** — creates one pull request per updated skill via `gh pr create`

The workflow is triggered on a weekly schedule (every Monday at 00:00 UTC) and can also be triggered manually via `workflow_dispatch`.

Required repository permissions (set automatically in the generated workflow):

| Permission | Reason |
|---|---|
| `contents: write` | Push update branches |
| `pull-requests: write` | Open pull requests |

### `--renovate`

Adds an entry to `renovate.json` (creating the file with a minimal schema stub if it does not exist) under `customManagers`:

```json
{
  "customType": "jsonata",
  "fileFormat": "toml",
  "managerFilePatterns": ["(^|/)\\.skillspkg\\.toml$"],
  "matchStrings": [
    "skills[source = \"git\" and $startsWith(url, \"https://github.com/\")].{\"depName\": $replace($substringAfter(url, \"https://github.com/\"), /\\.git$/, \"\"), \"currentValue\": version}"
  ],
  "datasourceTemplate": "github-tags",
  "versioningTemplate": "semver-coerced"
}
```

This instructs Renovate to:

- Parse every `.skillspkg.toml` in the repository as TOML
- Extract skills whose `source` is `git` and whose `url` starts with `https://github.com/`
- Look up the latest GitHub tag for each extracted `owner/repo` pair
- Open a PR when a newer tag is found

Running `setup-ci --renovate` a second time is a no-op if the entry already exists (detected by `managerFilePatterns`).

**Limitations**

- `renovate.json` must be valid JSON; JSONC (comments, trailing commas) is not supported.
- Only `https://github.com/` URLs are tracked. SSH remotes (`git@github.com:...`) are skipped.
- Skills with no `version` field are ignored by Renovate.

### Examples

```sh
# Generate the GitHub Actions workflow only
skills-pkg setup-ci --github-actions

# Add the Renovate custom manager only
skills-pkg setup-ci --renovate

# Generate both at once
skills-pkg setup-ci --github-actions --renovate
```

---

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error (any kind) |
