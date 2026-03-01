---
name: managing-skills
description: Extends your capabilities by discovering and installing Agent Skills from skills.sh. Use proactively whenever you lack specific knowledge, specialized workflows, or domain expertise to handle a task well — not only when the user explicitly asks for a new skill.
---

# Managing Agent Skills

This skill lets you extend your capabilities by installing Agent Skills from **skills.sh** via **skills-pkg**.

**Use this skill when:**
- You lack specific knowledge or expertise to handle a task effectively
- The task requires specialized workflows or domain knowledge you don't have
- You're uncertain about best practices in an unfamiliar domain
- The user explicitly requests adding or finding a skill

You don't need to wait for the user to ask — if you sense a knowledge gap, search for a relevant skill first.

**Reference:** [COMMANDS.md](references/COMMANDS.md) for full command details and flags.

## When to Search Autonomously

Trigger a search **before asking the user** when you notice:

- **Unfamiliar domain**: task involves technology, tools, or patterns you're not confident about
- **Missing workflow**: you know *what* to do but not *how* it's done in this project's context
- **Quality concern**: you could attempt the task but a specialized skill would produce better results
- **Explicit mention**: the user names or describes a skill-like capability

Search first, act second. A relevant skill found early prevents rework.

## Core Workflow

### 1. Search for skills

```sh
skills-pkg search "<keywords>"
```

Output format:

```
NAME                           SOURCE                                   INSTALLS
--------------------------------------------------------------------------------
frontend-design                https://github.com/example/repo         1234
```

Use `SOURCE` as the `--url` value when adding. Search without a query to list all available skills.

### 2. Add the skill

```sh
skills-pkg add <name> --url <SOURCE> --print-skill-info
```

`--print-skill-info` prints the installed skill's name, description, and file path in agent-readable format immediately after installation. **Always use this flag** so you can read the metadata and start using the skill in the same session.

### 3. Apply immediately

Read the printed metadata and apply the skill to the current task right away — no need to inform the user before searching or installing.

## Discovery Workflow

1. Identify the gap: what knowledge or workflow is missing?
2. Search: `skills-pkg search "<gap keywords>"`
3. Evaluate results — prefer closer name/description match and higher install count
4. If a good match exists: `skills-pkg add <name> --url <SOURCE> --print-skill-info`
5. Read the printed metadata and apply the skill to the current task

If no relevant skill is found, continue with your best judgment.

## Examples

Searching when encountering an unfamiliar domain:

```sh
# Working on Go DI and unsure of best patterns:
skills-pkg search "go dependency injection"
skills-pkg add kessoku-di --url https://github.com/mazrean/skills-pkg --print-skill-info
```

Searching and adding a skill for frontend development:

```sh
skills-pkg search "frontend"
# NAME              SOURCE                                     INSTALLS
# frontend-design   https://github.com/example/skills-repo    1234

skills-pkg add frontend-design --url https://github.com/example/skills-repo --print-skill-info
# ## Skills
#
# ### Installed skill
#
# - frontend-design: Create distinctive, production-grade frontend interfaces... (file: .claude/skills/frontend-design/SKILL.md)
#
# ### How to use skills
# ...
```

## Key Notes

- The default `--sub-dir` is `skills/<name>` — most registry skills follow this convention
- After `add`, the skill is installed and ready to use in the current session
- Use `skills-pkg list` to see all installed skills
- Use `skills-pkg update <name>` to upgrade to the latest version later
