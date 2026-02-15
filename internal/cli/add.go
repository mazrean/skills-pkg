package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// AddCmd represents the add command
type AddCmd struct {
	Name    string `arg:"" help:"Skill name"`
	Source  string `default:"git" help:"Source type (git, npm, go-module)"`
	URL     string `required:"" help:"Source URL (Git URL, npm package name, or Go module path)"`
	Version string `default:"latest" help:"Version (tag, commit hash, or semantic version)"`
	SubDir  string `help:"Subdirectory within the source to extract (default: skills/{name})"`
}

// Run executes the add command
// Requirements: 6.3, 12.1, 12.2, 12.3
func (c *AddCmd) Run(ctx *kong.Context) error {
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
// This method adds a skill to the configuration file.
// Requirements: 6.3, 12.1, 12.2, 12.3
func (c *AddCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Display progress information (requirement 12.1)
	logger.Info("Adding skill '%s' to configuration", c.Name)
	logger.Verbose("Source: %s, URL: %s, Version: %s", c.Source, c.URL, c.Version)

	// Validate source type (requirement 6.3)
	validSources := map[string]bool{
		"git":       true,
		"npm":       true,
		"go-module": true,
	}
	if !validSources[c.Source] {
		// Report invalid source type with cause and recommended action (requirements 12.2, 12.3)
		logger.Error("Invalid source type '%s'", c.Source)
		logger.Error("Supported source types: git, npm, go-module")
		return fmt.Errorf("%w: %s. Supported types: git, npm, go-module", domain.ErrInvalidSource, c.Source)
	}

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Determine SubDir (default: skills/{name})
	subDir := c.SubDir
	if subDir == "" {
		subDir = fmt.Sprintf("skills/%s", c.Name)
		logger.Verbose("Using default subdirectory: %s", subDir)
	}

	// Create skill entry
	skill := &domain.Skill{
		Name:      c.Name,
		Source:    c.Source,
		URL:       c.URL,
		Version:   c.Version,
		HashAlgo:  "", // Hash will be set during installation
		HashValue: "", // Hash will be set during installation
		SubDir:    subDir,
	}

	logger.Verbose("Created skill entry: %+v", skill)

	// Add skill to configuration (requirement 6.3)
	if err := configManager.AddSkill(context.Background(), skill); err != nil {
		// Handle different error types with appropriate messages (requirements 12.2, 12.3)
		if errors.Is(err, domain.ErrConfigNotFound) {
			// Configuration file not found
			logger.Error("Configuration file not found at %s", configPath)
			logger.Error("Run 'skills-pkg init' to create a configuration file")
			return err
		}

		if errors.Is(err, domain.ErrSkillExists) {
			// Duplicate skill name (requirement 6.3)
			logger.Error("Skill '%s' already exists in configuration", c.Name)
			logger.Error("Use 'skills-pkg update' to update an existing skill or choose a different name")
			return err
		}

		if errors.Is(err, domain.ErrInvalidSource) {
			// Invalid source type
			logger.Error("Invalid source type '%s'", c.Source)
			logger.Error("Supported source types: git, npm, go-module")
			return err
		}

		// File system error or other errors - distinguish and report (requirements 12.2, 12.3)
		logger.Error("Failed to add skill to configuration: %v", err)
		logger.Error("Check file permissions and try again")
		return err
	}

	// Success message (requirement 12.1)
	logger.Info("Successfully added skill '%s' to configuration", c.Name)
	logger.Info("Use 'skills-pkg install %s' to install the skill", c.Name)

	return nil
}
