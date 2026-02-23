# Configuration Reference

skills-pkg is configured via a single TOML file, by default named `.skillspkg.toml` in the project root.

## Top-level fields

| Field | Type | Required | Description |
|---|---|---|---|
| `install_targets` | `[]string` | yes | List of directories where skills are installed |
| `skills` | `[]Skill` | — | List of managed skills (populated by `add`, `update`) |

### `install_targets`

Each entry is a path (absolute or relative to the project root) into which all skill subdirectories are copied.

```toml
install_targets = [
  './.claude/skills',
  './.codex/skills',
]
```

You can point multiple agents at the same shared location, or keep them separate.

---

## Skill entry fields

Each skill is declared as a `[[skills]]` table.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Unique identifier for this skill |
| `source` | `string` | yes | Source type: `"git"` or `"go-mod"` |
| `url` | `string` | yes | Git remote URL or Go module path |
| `version` | `string` | — | Pinned version (tag, commit hash, or semver). Defaults to latest tag for git; resolved from `go.mod` for go-mod |
| `subdir` | `string` | — | Subdirectory within the source that contains the skill files. Defaults to `skills/<name>` |
| `hash_value` | `string` | — | Content hash recorded after installation (format: `h1:<base64>`). Set automatically; do not edit manually |

### `source` values

**`git`** — Clone a Git repository.

- `url`: any Git-compatible URL (HTTPS, SSH)
- `version`: a tag (`v1.0.0`), branch name, or full commit SHA

**`go-mod`** — Fetch a Go module via the module proxy.

- `url`: a Go module path (e.g., `github.com/example/mymodule`)
- `version`: a semver tag (`v1.2.3`) or pseudo-version. When omitted, version resolution follows this priority: (1) the version recorded in the nearest `go.mod` file found by walking up the directory tree, then (2) the latest version from the module proxy

See [Go Module Integration](go-module-integration.md) for detailed behavior including `GOPROXY` support and `direct` mode.

---

## Complete example

```toml
install_targets = [
  './.claude/skills',
  './.codex/skills',
]

[[skills]]
name       = "code-review"
source     = "git"
url        = "https://github.com/example/agent-skills"
version    = "v3.1.0"
subdir     = "skills/code-review"
hash_value = "h1:YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo="

[[skills]]
name       = "test-writer"
source     = "go-mod"
url        = "github.com/example/go-skills"
version    = "v0.5.2"
subdir     = "skills/test-writer"
hash_value = "h1:MTIzNDU2Nzg5MGFiY2RlZmdoaWprbG1ub3Bx"
```

---

## Managing the config file

In typical usage you **do not edit `.skillspkg.toml` by hand**. The CLI commands maintain it for you:

| Command | Effect on config |
|---|---|
| `init` | Creates the file with specified `install_targets` |
| `add` | Appends a `[[skills]]` entry and sets `hash_value` |
| `update` | Updates `version` and `hash_value` for the named skills |
| `uninstall` | Removes the matching `[[skills]]` entry |
| `install` | Reads the file; does not modify it |
| `verify` | Reads `hash_value`; does not modify it |

Commit `.skillspkg.toml` to version control so that all collaborators install the same skill versions.

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `SKILLSPKG_VERBOSE` | `false` | Enable verbose output (equivalent to `-v` / `--verbose`) |
| `GOPROXY` | `https://proxy.golang.org,direct` | Go Module proxy list used when `source = "go-mod"`. Follows the same syntax as the Go toolchain |
| `SKILLSPKG_TEMP_DIR` | OS temp dir | Override the base directory used for temporary module downloads (`go-mod` source only) |
