# Getting Started

This guide walks you through installing **skills-pkg** and setting up skill management for your project from scratch.

## Prerequisites

- A terminal / command prompt
- One or more AI coding agents installed (e.g., Claude Code, Codex CLI)

## 1. Install skills-pkg

### Homebrew (macOS / Linux)

```sh
brew install mazrean/tap/skills-pkg
```

### Pre-built binary (all platforms)

Go to the [Releases page](https://github.com/mazrean/skills-pkg/releases), download the archive for your OS and architecture, and extract the `skills-pkg` binary to a directory on your `PATH`.

Verify the installation:

```sh
skills-pkg --help
```

---

## 2. Initialize a project

Navigate to the root of your project and run:

```sh
skills-pkg init
```

This creates `.skillspkg.toml` with a default install target of `./.skills`:

```toml
install_targets = ['./.skills']
skills = []
```

### Initialize for a specific agent

Use `--agent` to register the agent's default skill directory as an install target.

```sh
# Project-level: installs to ./.claude/skills
skills-pkg init --agent claude

# User-level (global): installs to ~/.claude/skills
skills-pkg init --agent claude --global
```

You can specify multiple agents at once:

```sh
skills-pkg init --agent claude --agent codex
```

### Use a custom directory

```sh
skills-pkg init --install-dir ./shared/skills
```

You can combine `--agent` and `--install-dir` freely.

---

## 3. Add your first skill

```sh
skills-pkg add my-skill --url https://github.com/example/skills-repo
```

This command:
1. Downloads the repository (defaults to the latest tag)
2. Extracts the subdirectory `skills/my-skill` (default subdir convention)
3. Copies the files to every directory listed in `install_targets`
4. Records the skill and its hash in `.skillspkg.toml`

### Specify a version

```sh
skills-pkg add my-skill --url https://github.com/example/skills-repo --version v2.0.0
```

### Install from a Go module

```sh
skills-pkg add my-skill --source go-mod --url github.com/example/skills-module
```

When using `go-mod`, the version defaults to the one resolved by your `go.mod` file.

---

## 4. Install skills on another machine

If you checked `.skillspkg.toml` into version control, team members can reproduce the installation with:

```sh
skills-pkg install
```

This installs all skills listed in the config, resolving to the pinned `version` and verifying `hash_value`.

---

## 5. Keep skills up to date

```sh
# Update all skills to their latest versions
skills-pkg update

# Update a specific skill
skills-pkg update my-skill
```

---

## 6. Verify integrity

Run `verify` at any time to check that installed files match the recorded hashes:

```sh
skills-pkg verify
```

If a hash mismatch is detected, skills-pkg reports the affected skill so you can investigate or reinstall it.

---

## 7. Remove a skill

```sh
skills-pkg uninstall my-skill
```

This removes the skill entry from `.skillspkg.toml` and deletes the installed files from all targets.

---

## Next steps

- [Configuration Reference](configuration.md) — learn all options in `.skillspkg.toml`
- [Command Reference](commands.md) — full flags and usage for every command
