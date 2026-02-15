package cli

import (
	"context"
	"errors"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// UpdateCmd represents the update command
type UpdateCmd struct {
	Skills []string `arg:"" optional:"" help:"Skill names to update (empty for all)"`
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

	// Display progress information (requirement 12.1)
	if len(c.Skills) == 0 {
		logger.Info("Updating all skills to latest versions")
	} else {
		logger.Info("Updating skills: %v", c.Skills)
	}

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Create HashService
	hashService := adapter.NewDirhashService()

	// Create PackageManagers
	packageManagers := []port.PackageManager{
		adapter.NewGitAdapter(),
		adapter.NewNpmAdapter(),
		adapter.NewGoModAdapter(),
		adapter.NewPipAdapter(),
		adapter.NewCargoAdapter(),
	}

	// Create SkillManager
	skillManager := domain.NewSkillManager(configManager, hashService, packageManagers)

	// Determine what to update (requirements 7.1, 7.2)
	if len(c.Skills) == 0 {
		// Update all skills (requirement 7.1)
		logger.Verbose("Updating all skills")
		result, err := skillManager.Update(context.Background(), "")
		if err != nil {
			c.handleUpdateError(logger, "", configPath, err)
			return err
		}
		if result != nil {
			// Display update result (requirement 7.6)
			logger.Info("Successfully updated all skills")
		}
	} else {
		// Update specific skills (requirement 7.2)
		for _, skillName := range c.Skills {
			logger.Verbose("Updating skill: %s", skillName)
			result, err := skillManager.Update(context.Background(), skillName)
			if err != nil {
				c.handleUpdateError(logger, skillName, configPath, err)
				return err
			}
			if result != nil {
				// Display update result (requirement 7.6)
				logger.Info("Successfully updated skill '%s' from %s to %s", result.SkillName, result.OldVersion, result.NewVersion)
			}
		}
	}

	// Success message (requirement 12.1)
	logger.Info("Update complete")

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
