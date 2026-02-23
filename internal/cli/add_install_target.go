package cli

import (
	"context"
	"errors"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// AddInstallTargetCmd represents the add-install-target command
type AddInstallTargetCmd struct {
	Target string `arg:"" help:"Install target directory path"`
}

// Run executes the add-install-target command
func (c *AddInstallTargetCmd) Run(ctx *kong.Context) error {
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

func (c *AddInstallTargetCmd) run(configPath string, verbose bool) error {
	logger := NewLogger(verbose)

	logger.Info("Adding install target '%s' to configuration", c.Target)
	logger.Verbose("Config path: %s", configPath)

	configManager := domain.NewConfigManager(configPath)

	if err := configManager.AddInstallTarget(context.Background(), c.Target); err != nil {
		if err, ok := errors.AsType[*domain.ErrorConfigNotFound](err); ok {
			logger.Error("Configuration file not found at %s", err.Path)
			logger.Error("Run 'skills-pkg init' to create a configuration file")
			return err
		}

		if err, ok := errors.AsType[*domain.ErrorInstallTargetExists](err); ok {
			logger.Error("Install target '%s' already exists in configuration", err.Target)
			return err
		}

		logger.Error("Failed to add install target: %v", err)
		return err
	}

	logger.Info("Successfully added install target '%s'", c.Target)

	return nil
}
