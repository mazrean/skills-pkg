package cli

import (
	"context"
	"errors"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// ListCmd represents the list command
type ListCmd struct {
}

// Run executes the list command
// Requirements: 8.1, 8.2, 8.3, 8.4, 12.1, 12.2, 12.3
func (c *ListCmd) Run(ctx *kong.Context) error {
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
// This method lists all skills from the configuration file.
// Requirements: 8.1, 8.2, 8.3, 8.4, 12.1, 12.2, 12.3
func (c *ListCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting
	logger := NewLogger(verbose)

	return c.runWithLogger(configPath, logger)
}

// runWithLogger executes the list command with a custom logger (for testing)
// Requirements: 8.1, 8.2, 8.3, 8.4, 12.1, 12.2, 12.3
func (c *ListCmd) runWithLogger(configPath string, logger *Logger) error {
	// Display progress information (requirement 12.1)
	logger.Verbose("Loading skills from configuration")

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Load all skills (requirements 8.1, 8.2)
	skills, err := configManager.ListSkills(context.Background())
	if err != nil {
		// Handle different error types with appropriate messages (requirements 12.2, 12.3)
		if err, ok := errors.AsType[*domain.ErrorConfigNotFound](err); ok {
			// Configuration file not found
			logger.Error("Configuration file not found at %s", err.Path)
			logger.Error("Run 'skills-pkg init' to create a configuration file")
			return err
		}

		// File system error or other errors - distinguish and report (requirements 12.2, 12.3)
		logger.Error("Failed to load skills from configuration: %v", err)
		logger.Error("Check file permissions and try again")
		return err
	}

	// Check if skills list is empty (requirement 8.4)
	if len(skills) == 0 {
		logger.Info("No skills installed")
		logger.Info("Use 'skills-pkg add <name> --source <type> --url <url>' to add skills")
		return nil
	}

	// Display skills in a table format (requirements 8.2, 8.3)
	logger.Info("")
	logger.Info("Installed Skills:")
	logger.Info("%-20s %-15s %-30s", "NAME", "SOURCE", "VERSION")
	logger.Info("%s", "--------------------------------------------------------------------------------")

	for _, skill := range skills {
		logger.Info("%-20s %-15s %-30s", skill.Name, skill.Source, skill.Version)
	}

	logger.Info("")
	logger.Info("Total: %d skill(s)", len(skills))

	return nil
}
