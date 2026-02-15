package cli

import (
	"context"
	"errors"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter/pkgmanager"
	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// UninstallCmd represents the uninstall command
type UninstallCmd struct {
	SkillName string `arg:"" help:"Name of the skill to remove from configuration and all install targets"`
}

// Run executes the uninstall command
// Requirements: 9.1, 9.2, 9.3, 9.4, 12.1, 12.2, 12.3
func (c *UninstallCmd) Run(ctx *kong.Context) error {
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
// This method uninstalls a skill by removing it from all install targets and the configuration.
// Requirements: 9.1, 9.2, 9.3, 9.4, 12.1, 12.2, 12.3
func (c *UninstallCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Display progress information (requirement 12.1)
	logger.Info("Uninstalling skill '%s'", c.SkillName)
	logger.Verbose("Config path: %s", configPath)

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

	// Execute uninstall (requirements 9.1, 9.2)
	logger.Verbose("Removing skill from install targets and configuration")
	if err := skillManager.Uninstall(context.Background(), c.SkillName); err != nil {
		c.handleUninstallError(logger, c.SkillName, configPath, err)
		return err
	}

	// Success message (requirement 9.4, 12.1)
	logger.Info("Successfully uninstalled skill '%s'", c.SkillName)

	return nil
}

// handleUninstallError handles different types of errors that can occur during skill uninstallation.
// It provides appropriate error messages with causes and recommended actions.
// Requirements: 9.3, 12.2, 12.3
func (c *UninstallCmd) handleUninstallError(logger *Logger, skillName string, configPath string, err error) {
	// Configuration file not found (requirement 12.2, 12.3)
	if errors.Is(err, domain.ErrConfigNotFound) {
		logger.Error("Configuration file not found at %s", configPath)
		logger.Error("Run 'skills-pkg init' to create a configuration file")
		return
	}

	// Skill not found in configuration (requirement 9.3, 12.2, 12.3)
	if errors.Is(err, domain.ErrSkillNotFound) {
		logger.Error("Skill '%s' not found in configuration", skillName)
		logger.Error("Use 'skills-pkg list' to see available skills")
		return
	}

	// File system or other errors - distinguish and report (requirements 12.2, 12.3)
	logger.Error("Failed to uninstall skill '%s': %v", skillName, err)
	logger.Error("Check file permissions and try again")
}
