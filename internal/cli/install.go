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

// InstallCmd represents the install command
type InstallCmd struct {
	Skills []string `arg:"" optional:"" help:"Skill names to install (if not specified, installs all skills from configuration)"`
}

// Run executes the install command
// Requirements: 6.1, 6.2, 6.3, 12.1, 12.2, 12.3, 12.4
func (c *InstallCmd) Run(ctx *kong.Context) error {
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
// This method installs skills from the configuration file.
// Requirements: 6.1, 6.2, 6.3, 12.1, 12.2, 12.3, 12.4
func (c *InstallCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Display progress information (requirement 12.1)
	if len(c.Skills) == 0 {
		logger.Info("Installing all skills from configuration")
	} else {
		logger.Info("Installing skills: %v", c.Skills)
	}

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Create HashService
	hashService := adapter.NewDirhashService()

	// Create PackageManagers
	packageManagers := []port.PackageManager{
		adapter.NewGitAdapter(),
		adapter.NewGoModAdapter(),
	}

	// Create SkillManager
	skillManager := domain.NewSkillManager(configManager, hashService, packageManagers)

	// Determine what to install (requirements 6.1, 6.2)
	if len(c.Skills) == 0 {
		// Install all skills (requirement 6.1)
		logger.Verbose("Installing all skills")
		if err := skillManager.Install(context.Background(), ""); err != nil {
			c.handleInstallError(logger, "", configPath, err)
			return err
		}
		logger.Info("Successfully installed all skills")
	} else {
		// Install specific skills (requirement 6.2)
		for _, skillName := range c.Skills {
			logger.Verbose("Installing skill: %s", skillName)
			if err := skillManager.Install(context.Background(), skillName); err != nil {
				c.handleInstallError(logger, skillName, configPath, err)
				return err
			}
			logger.Info("Successfully installed skill '%s'", skillName)
		}
	}

	// Success message (requirement 12.1)
	logger.Info("Installation complete")

	return nil
}

// handleInstallError handles different types of errors that can occur during skill installation.
// It provides appropriate error messages with causes and recommended actions.
// Requirements: 6.3, 12.2, 12.3
func (c *InstallCmd) handleInstallError(logger *Logger, skillName string, configPath string, err error) {
	// Configuration file not found
	if errors.Is(err, domain.ErrConfigNotFound) {
		logger.Error("Configuration file not found at %s", configPath)
		logger.Error("Run 'skills-pkg init' to create a configuration file")
		return
	}

	// Skill not found in configuration (requirement 6.3)
	if errors.Is(err, domain.ErrSkillNotFound) {
		logger.Error("Skill '%s' not found in configuration", skillName)
		logger.Error("Use 'skills-pkg add %s --source <type> --url <url>' to add it first", skillName)
		return
	}

	// Network, file system, or other errors - distinguish and report (requirements 12.2, 12.3)
	if skillName == "" {
		logger.Error("Failed to install skills: %v", err)
	} else {
		logger.Error("Failed to install skill '%s': %v", skillName, err)
	}
	logger.Error("Check network connection, file permissions, and try again")
}
