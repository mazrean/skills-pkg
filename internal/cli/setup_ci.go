package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/alecthomas/kong"
)

// SetupCICmd represents the setup-ci command
type SetupCICmd struct {
	GitHubActions bool `help:"Generate GitHub Actions workflow for skills auto-update" name:"github-actions"`
	Renovate      bool `help:"Add Renovate custom manager configuration for skills auto-update" name:"renovate"`
}

// Run executes the setup-ci command
func (c *SetupCICmd) Run(ctx *kong.Context) error {
	verbose := false
	if model := ctx.Model; model != nil && model.Target.IsValid() {
		if verboseField := model.Target.FieldByName("Verbose"); verboseField.IsValid() && verboseField.Kind() == reflect.Bool {
			verbose = verboseField.Bool()
		}
	}

	return c.run(
		filepath.Join(".github", "workflows", "update-skills.yml"),
		"renovate.json",
		verbose,
	)
}

func (c *SetupCICmd) run(workflowPath, renovatePath string, verbose bool) error {
	if !c.GitHubActions && !c.Renovate {
		return fmt.Errorf("specify at least one of --github-actions or --renovate")
	}

	logger := NewLogger(verbose)

	if c.GitHubActions {
		if err := c.setupGitHubActions(logger, workflowPath); err != nil {
			return fmt.Errorf("failed to set up GitHub Actions: %w", err)
		}
	}

	if c.Renovate {
		if err := c.setupRenovate(logger, renovatePath); err != nil {
			return fmt.Errorf("failed to set up Renovate: %w", err)
		}
	}

	return nil
}

const (
	// setupCIDirPerm is the permission for directories created by setup-ci (rwxr-xr-x).
	setupCIDirPerm = 0o755
	// setupCIFilePerm is the permission for files written by setup-ci (rw-r--r--).
	setupCIFilePerm = 0o644
)

// updateSkillsWorkflow is the GitHub Actions workflow template for auto-updating skills.
// It runs a dry-run to detect available updates, then uses git worktrees to apply each
// skill's update in parallel and creates a separate PR per skill.
const updateSkillsWorkflow = `name: Update Skills

on:
  schedule:
    - cron: '0 0 * * 1'
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

jobs:
  update-skills:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Install skills-pkg
        run: go install github.com/mazrean/skills-pkg@latest

      - name: Detect updates (dry-run)
        id: detect
        run: |
          skills-pkg update --dry-run --output json > dry-run-output.json
          UPDATE_SKILLS=$(jq -c '[.updates[] | select(.has_update == true) | .skill_name]' dry-run-output.json)
          echo "skills=$UPDATE_SKILLS" >> "$GITHUB_OUTPUT"

      - name: Update skills with worktrees
        if: steps.detect.outputs.skills != '[]'
        run: |
          set -euo pipefail
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git config user.name "github-actions[bot]"

          pids=()
          for skill in $(echo '${{ steps.detect.outputs.skills }}' | jq -r '.[]'); do
            branch="update-skill/${skill}"
            git worktree add "worktrees/${skill}" -B "${branch}"
            (
              set -euo pipefail
              cd "worktrees/${skill}"
              skills-pkg update "${skill}"
              git add -A
              if git diff --cached --quiet; then
                echo "No changes for ${skill}, skipping commit"
                exit 0
              fi
              git commit -m "chore(skills): update ${skill}"
              git push --force-with-lease origin "${branch}"
            ) &
            pids+=($!)
          done

          failed=0
          for pid in "${pids[@]}"; do
            wait "$pid" || failed=1
          done
          if [ "${failed}" -ne 0 ]; then
            echo "One or more skill updates failed" >&2
            exit 1
          fi

      - name: Create PRs
        if: steps.detect.outputs.skills != '[]'
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          for skill in $(echo '${{ steps.detect.outputs.skills }}' | jq -r '.[]'); do
            branch="update-skill/${skill}"
            gh pr create \
              --title "chore(skills): update ${skill}" \
              --body "Automated skill update for ${skill}." \
              --head "${branch}" \
              --label "dependencies" \
              || true
          done
`

func (c *SetupCICmd) setupGitHubActions(logger *Logger, workflowPath string) error {
	workflowDir := filepath.Dir(workflowPath)
	if err := os.MkdirAll(workflowDir, setupCIDirPerm); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	if err := os.WriteFile(workflowPath, []byte(updateSkillsWorkflow), setupCIFilePerm); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	logger.Info("Created %s", workflowPath)
	return nil
}

// renovateCustomManager holds the Renovate custom manager configuration for skills-pkg.
// It uses JSONata to parse .skillspkg.toml and extracts git-source skills,
// mapping GitHub repository URLs and versions to the github-tags datasource.
// Fields are ordered to minimise GC-scannable pointer bytes (strings before slices).
type renovateCustomManager struct {
	CustomType          string   `json:"customType"`
	FileFormat          string   `json:"fileFormat"`
	DatasourceTemplate  string   `json:"datasourceTemplate"`
	VersioningTemplate  string   `json:"versioningTemplate"`
	ManagerFilePatterns []string `json:"managerFilePatterns"`
	MatchStrings        []string `json:"matchStrings"`
}

var skillspkgRenovateManager = renovateCustomManager{
	CustomType: "jsonata",
	FileFormat: "toml",
	// managerFilePatterns uses plain regex strings (no /.../ delimiters).
	ManagerFilePatterns: []string{`(^|/)\.skillspkg\.toml$`},
	// Extract depName (GitHub owner/repo) and currentValue (version tag) from git-source skills.
	// Only skills sourced from https://github.com/ are matched.
	// $replace strips an optional .git suffix so "owner/repo.git" becomes "owner/repo".
	MatchStrings: []string{
		`skills[source = "git" and $startsWith(url, "https://github.com/")].{"depName": $replace($substringAfter(url, "https://github.com/"), /\.git$/, ""), "currentValue": version}`,
	},
	DatasourceTemplate: "github-tags",
	VersioningTemplate: "semver-coerced",
}

func (c *SetupCICmd) setupRenovate(logger *Logger, renovatePath string) error {
	var rawConfig map[string]json.RawMessage

	data, err := os.ReadFile(renovatePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read %s: %w", renovatePath, err)
		}
		schemaVal, marshalErr := json.Marshal("https://docs.renovatebot.com/renovate-schema.json")
		if marshalErr != nil {
			return fmt.Errorf("failed to build default renovate config: %w", marshalErr)
		}
		rawConfig = map[string]json.RawMessage{
			"$schema": schemaVal,
		}
	} else {
		if unmarshalErr := json.Unmarshal(data, &rawConfig); unmarshalErr != nil {
			return fmt.Errorf("failed to parse %s: %w", renovatePath, unmarshalErr)
		}
	}

	var existingManagers []json.RawMessage
	if rawManagers, ok := rawConfig["customManagers"]; ok {
		if unmarshalErr := json.Unmarshal(rawManagers, &existingManagers); unmarshalErr != nil {
			return fmt.Errorf("failed to parse customManagers in %s: %w", renovatePath, unmarshalErr)
		}
	}

	// Skip if the skills-pkg manager is already registered (identified by managerFilePatterns).
	for _, m := range existingManagers {
		var mgr map[string]json.RawMessage
		if unmarshalErr := json.Unmarshal(m, &mgr); unmarshalErr != nil {
			continue
		}
		patternsRaw, ok := mgr["managerFilePatterns"]
		if !ok {
			continue
		}
		var patterns []string
		if unmarshalErr := json.Unmarshal(patternsRaw, &patterns); unmarshalErr != nil {
			continue
		}
		if slices.Contains(patterns, skillspkgRenovateManager.ManagerFilePatterns[0]) {
			logger.Info("Renovate custom manager for skills-pkg already configured in %s", renovatePath)
			return nil
		}
	}

	managerJSON, err := json.Marshal(skillspkgRenovateManager)
	if err != nil {
		return fmt.Errorf("failed to marshal skills-pkg custom manager: %w", err)
	}
	existingManagers = append(existingManagers, managerJSON)

	updatedManagers, err := json.Marshal(existingManagers)
	if err != nil {
		return fmt.Errorf("failed to marshal customManagers: %w", err)
	}
	rawConfig["customManagers"] = updatedManagers

	outData, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", renovatePath, err)
	}

	if err := os.WriteFile(renovatePath, append(outData, '\n'), setupCIFilePerm); err != nil {
		return fmt.Errorf("failed to write %s: %w", renovatePath, err)
	}

	logger.Info("Updated %s", renovatePath)
	return nil
}
