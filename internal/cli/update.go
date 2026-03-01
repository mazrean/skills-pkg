package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter/pkgmanager"
	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// UpdateCmd represents the update command
type UpdateCmd struct {
	Output string   `help:"Output format (text, json)" default:"text" enum:"text,json"`
	Skills []string `arg:"" optional:"" help:"Skill names to update (if not specified, updates all skills to their latest versions)"`
	DryRun bool     `help:"Show what would be updated without making changes" name:"dry-run"`
}

// Run executes the update command
// Requirements: 7.1, 7.2, 7.6, 12.1, 12.2, 12.3
func (c *UpdateCmd) Run(ctx *kong.Context) error {
	// Access verbose flag from the parsed CLI model using reflection
	verbose := false
	if model := ctx.Model; model != nil && model.Target.IsValid() {
		// Get the "Verbose" field from the CLI struct
		if verboseField := model.Target.FieldByName("Verbose"); verboseField.IsValid() && verboseField.Kind() == reflect.Bool {
			verbose = verboseField.Bool()
		}
	}

	return c.run(defaultConfigPath, verbose)
}

// run is the internal implementation that can be called from tests with custom parameters
// This method updates skills to their latest versions.
// Requirements: 7.1, 7.2, 7.6, 12.1, 12.2, 12.3
func (c *UpdateCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Create HashService
	hashService := service.NewDirhash()

	// Create PackageManagers
	packageManagers := []port.PackageManager{
		pkgmanager.NewGit(),
		pkgmanager.NewGoMod(),
	}

	// Create SkillManager
	skillManager := domain.NewSkillManager(configManager, hashService, packageManagers)

	// Display progress information (requirement 12.1)
	if c.DryRun {
		logger.Verbose("Checking for updates for skills: %v", c.Skills)
	} else {
		logger.Info("Updating skills: %v", c.Skills)
	}

	// Determine what to update (requirements 7.1, 7.2)
	var allResults []*domain.UpdateResult

	results, err := skillManager.Update(context.Background(), c.Skills, c.DryRun)
	if err != nil {
		c.handleUpdateError(logger, err)
		return err
	}
	allResults = append(allResults, results...)

	// Success message (requirement 12.1)
	logger.Info("Update complete")

	switch c.Output {
	case "json":
		return c.printDryRunJSON(logger, allResults)
	default:
		return c.printDryRunText(logger, allResults)
	}
}

// dryRunOutput is the JSON-serializable structure for dry-run results.
type dryRunOutput struct {
	Updates []*dryRunItem `json:"updates"`
}

type dryRunItem struct {
	SkillName      string            `json:"skill_name"`
	CurrentVersion string            `json:"current_version"`
	LatestVersion  string            `json:"latest_version"`
	FileDiffs      []*dryRunFileDiff `json:"file_diffs,omitempty"`
	HasUpdate      bool              `json:"has_update"`
}

type dryRunFileDiff struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Patch  string `json:"patch,omitempty"`
}

// printDryRunText prints human-readable dry-run results.
func (c *UpdateCmd) printDryRunText(logger *Logger, results []*domain.UpdateResult) error {
	updateCount := 0
	for _, r := range results {
		if r.OldVersion != r.NewVersion {
			logger.Info("  %s: %s → %s (update available)", r.SkillName, r.OldVersion, r.NewVersion)
			updateCount++
		} else {
			logger.Info("  %s: %s (up to date)", r.SkillName, r.OldVersion)
		}

		// Show file-level diffs
		for _, fd := range r.FileDiffs {
			switch domain.FileDiffStatus(fd.Status) {
			case domain.FileDiffAdded:
				logger.Info("    + %s", fd.Path)
			case domain.FileDiffRemoved:
				logger.Info("    - %s", fd.Path)
			case domain.FileDiffModified:
				logger.Info("    ~ %s", fd.Path)
				if fd.Patch != "" {
					for line := range strings.SplitSeq(strings.TrimRight(fd.Patch, "\n"), "\n") {
						logger.Info("      %s", line)
					}
				}
			}
		}
	}

	total := len(results)
	switch updateCount {
	case 0:
		logger.Info("%d skill(s) checked, all up to date", total)
	default:
		logger.Info("%d skill(s) checked, %d update(s) available", total, updateCount)
		logger.Info("Run 'skills-pkg update' to apply updates.")
	}

	return nil
}

// printDryRunJSON prints JSON dry-run results.
func (c *UpdateCmd) printDryRunJSON(logger *Logger, results []*domain.UpdateResult) error {
	items := make([]*dryRunItem, 0, len(results))
	for _, r := range results {
		fileDiffs := make([]*dryRunFileDiff, 0, len(r.FileDiffs))
		for _, fd := range r.FileDiffs {
			fileDiffs = append(fileDiffs, &dryRunFileDiff{
				Path:   fd.Path,
				Status: string(fd.Status),
				Patch:  fd.Patch,
			})
		}
		items = append(items, &dryRunItem{
			SkillName:      r.SkillName,
			CurrentVersion: r.OldVersion,
			LatestVersion:  r.NewVersion,
			HasUpdate:      r.OldVersion != r.NewVersion,
			FileDiffs:      fileDiffs,
		})
	}

	data, err := json.MarshalIndent(dryRunOutput{Updates: items}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON output: %w", err)
	}
	_, err = fmt.Fprintln(logger.dataOut, string(data))
	if err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}

	return nil
}

// handleUpdateError handles different types of errors that can occur during skill update.
// It provides appropriate error messages with causes and recommended actions.
// Requirements: 12.2, 12.3
func (c *UpdateCmd) handleUpdateError(logger *Logger, err error) {
	// Configuration file not found
	if err, ok := errors.AsType[*domain.ErrorConfigNotFound](err); ok {
		logger.Error("Configuration file not found at %s", err.Path)
		logger.Error("Run 'skills-pkg init' to create a configuration file")
		return
	}

	// Skill not found in configuration
	if err, ok := errors.AsType[*domain.ErrorSkillsNotFound](err); ok {
		quatedNames := make([]string, 0, len(err.SkillNames))
		for _, name := range err.SkillNames {
			quatedNames = append(quatedNames, fmt.Sprintf("'%s'", name))
		}

		logger.Error("Skills '%s' not found in configuration", strings.Join(quatedNames, ", "))
		if len(err.SkillNames) == 1 {
			logger.Error("Use 'skills-pkg add %s --source <type> --url <url>' to add it first", err.SkillNames[0])
		} else {
			logger.Error("Use 'skills-pkg add <skill-name> --source <type> --url <url>' to add them first")
		}
		return
	}

	// Network, file system, or other errors - distinguish and report (requirements 12.2, 12.3)
	logger.Error("Failed to update skills: %v", err)
	logger.Error("Check network connection, file permissions, and try again")
}
