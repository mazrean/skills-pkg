package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter/pkgmanager"
	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// UpdateCmd represents the update command
type UpdateCmd struct {
	Skills []string `arg:"" optional:"" help:"Skill names to update (if not specified, updates all skills to their latest versions)"`
	Source string   `optional:"" help:"Filter by source type (git or go-mod)" default:""`
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
// Requirements: 1.1, 1.3, 1.4, 1.5, 6.2, 6.3, 6.4, 7.2, 7.3, 7.4, 7.5
func (c *UpdateCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 6.4)
	logger := NewLogger(verbose)

	// Validate Source field (requirement 7.5)
	if c.Source != "" && c.Source != "git" && c.Source != "go-mod" {
		logger.Error("Invalid source type: %q", c.Source)
		logger.Error("Valid source types: git, go-mod")
		return domain.ErrInvalidSource
	}

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

	// Update skills with source filter (requirement 1.1, 1.3, 1.4)
	// Domain layer prints the target skill list before updating (requirement 6.1)
	results, err := skillManager.Update(context.Background(), c.Skills, c.Source)
	if err != nil {
		// Critical error handling (requirement 7.2, 7.4)
		if errors.Is(err, domain.ErrConfigNotFound) {
			logger.Error("Configuration file not found at %s", configPath)
			logger.Error("Run 'skills-pkg init' to create a configuration file")
			return err
		}
		logger.Error("Failed to update skills: %v", err)
		return err
	}

	// Per-skill result display (requirement 6.2)
	var updatedCount, skippedCount, failedCount int
	var failedResults []*domain.UpdateResult

	for _, result := range results {
		if result.Skipped {
			logger.Info("Skill '%s': already at latest (%s), skipped", result.SkillName, result.OldVersion)
			skippedCount++
		} else if result.Err != nil {
			logger.Error("Skill '%s': FAILED - %v", result.SkillName, result.Err)
			failedResults = append(failedResults, result)
			failedCount++
		} else {
			logger.Info("Skill '%s': %s → %s", result.SkillName, result.OldVersion, result.NewVersion)
			updatedCount++
		}
	}

	// Summary display (requirement 6.3)
	logger.Info("Update complete: %d updated, %d skipped, %d failed", updatedCount, skippedCount, failedCount)

	// Error aggregate display (requirement 7.3)
	if len(failedResults) > 0 {
		logger.Error("The following skills failed to update:")
		for _, result := range failedResults {
			c.handleUpdateError(logger, result.SkillName, configPath, result.Err)
		}
		return fmt.Errorf("update completed with %d error(s)", failedCount)
	}

	return nil
}

// handleUpdateError handles different types of errors that can occur during skill update.
// It provides appropriate error messages with causes and recommended actions.
// Requirements: 12.2, 12.3
func (c *UpdateCmd) handleUpdateError(logger *Logger, skillName string, configPath string, err error) {
	// Configuration file not found
	if errors.Is(err, domain.ErrConfigNotFound) {
		logger.Error("Configuration file not found at %s", configPath)
		logger.Error("Run 'skills-pkg init' to create a configuration file")
		return
	}

	// Skill not found in configuration
	if errors.Is(err, domain.ErrSkillNotFound) {
		logger.Error("Skill '%s' not found in configuration", skillName)
		logger.Error("Use 'skills-pkg add %s --source <type> --url <url>' to add it first", skillName)
		return
	}

	// Network, file system, or other errors - distinguish and report (requirements 12.2, 12.3)
	if skillName == "" {
		logger.Error("Failed to update skills: %v", err)
	} else {
		logger.Error("Failed to update skill '%s': %v", skillName, err)
	}
	logger.Error("Check network connection, file permissions, and try again")
}
