package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/domain"
)

const (
	// defaultConfigPath is the default path to the .skillspkg.toml configuration file
	defaultConfigPath = ".skillspkg.toml"
)

// InitCmd represents the init command
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 12.1, 12.2, 12.3, 12.4
type InitCmd struct {
	InstallDir []string `help:"Custom install directory (can be specified multiple times)" short:"d"`
	Agent      string   `help:"Agent name (e.g., 'claude') to use default directory" short:"a"`
}

// Run executes the init command
// This method initializes a new .skillspkg.toml configuration file with the specified install directories.
// It handles custom install directories (--install-dir) and agent-specific directories (--agent).
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 12.1, 12.2, 12.3, 12.4
func (c *InitCmd) Run(ctx *kong.Context) error {
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
func (c *InitCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Display progress information (requirement 12.1)
	logger.Info("Initializing project with .skillspkg.toml")

	// Build install targets list (requirements 1.2, 1.3)
	installTargets, err := c.buildInstallTargets(logger)
	if err != nil {
		// Distinguish error types and provide cause and recommended action (requirements 12.2, 12.3)
		logger.Error("Failed to build install targets: %v", err)
		return err
	}

	logger.Verbose("Install targets: %v", installTargets)

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Initialize configuration file (requirement 1.1, 1.5)
	if err := configManager.Initialize(context.Background(), installTargets); err != nil {
		// Handle different error types with appropriate messages (requirements 12.2, 12.3)
		if errors.Is(err, domain.ErrConfigExists) {
			// Configuration file already exists (requirement 1.4)
			logger.Error("Configuration file already exists at %s", configPath)
			logger.Error("Remove the existing file or use a different path")
			return err
		}

		// File system error - distinguish and report (requirements 12.2, 12.3)
		logger.Error("Failed to create configuration file: %v", err)
		logger.Error("Check file permissions and try again")
		return err
	}

	// Success message (requirement 12.1)
	logger.Info("Successfully initialized .skillspkg.toml")
	if len(installTargets) > 0 {
		logger.Info("Install targets:")
		for _, target := range installTargets {
			logger.Info("  - %s", target)
		}
	} else {
		logger.Info("No install targets configured. Add them later with 'skills-pkg add'")
	}

	return nil
}

// buildInstallTargets constructs the list of install target directories
// from custom directories (--install-dir) and agent-specific directories (--agent).
// Requirements: 1.2, 1.3, 10.3, 10.4
func (c *InitCmd) buildInstallTargets(logger *Logger) ([]string, error) {
	installTargets := make([]string, 0)

	// Add custom install directories (requirement 1.2)
	if len(c.InstallDir) > 0 {
		logger.Verbose("Adding custom install directories: %v", c.InstallDir)
		installTargets = append(installTargets, c.InstallDir...)
	}

	// Add agent-specific directory if --agent is specified (requirement 1.3)
	if c.Agent != "" {
		logger.Verbose("Resolving agent directory for: %s", c.Agent)

		// Use AgentProvider to resolve agent directory (requirements 10.3, 10.4)
		agentProvider := adapter.NewClaudeAgentAdapter()
		agentDir, err := agentProvider.ResolveAgentDir(c.Agent)
		if err != nil {
			// Report unsupported agent error with cause and recommended action (requirements 12.2, 12.3)
			return nil, fmt.Errorf("failed to resolve agent directory: %w. Supported agents: claude", err)
		}

		logger.Verbose("Resolved agent directory: %s", agentDir)
		installTargets = append(installTargets, agentDir)
	}

	return installTargets, nil
}
