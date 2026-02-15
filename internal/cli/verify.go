package cli

import (
	"context"
	"errors"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// VerifyCmd represents the verify command
type VerifyCmd struct {
}

// Run executes the verify command
// Requirements: 5.4, 5.5, 5.6, 12.1, 12.2, 12.3
func (c *VerifyCmd) Run(ctx *kong.Context) error {
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
// This method verifies all skills' hash values.
// Requirements: 5.4, 5.5, 5.6, 12.1, 12.2, 12.3
func (c *VerifyCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting
	logger := NewLogger(verbose)

	return c.runWithLogger(configPath, logger)
}

// runWithLogger executes the verify command with a custom logger (for testing)
// Requirements: 5.4, 5.5, 5.6, 12.1, 12.2, 12.3
func (c *VerifyCmd) runWithLogger(configPath string, logger *Logger) error {
	// Display progress information (requirement 12.1)
	logger.Info("Verifying skill integrity...")
	logger.Verbose("Loading configuration from %s", configPath)

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Create HashService
	hashService := adapter.NewDirhashService()

	// Create HashVerifier
	hashVerifier := domain.NewHashVerifier(configManager, hashService)

	// Verify all skills (requirements 5.4, 5.6)
	logger.Verbose("Starting verification of all skills")
	summary, err := hashVerifier.VerifyAll(context.Background())
	if err != nil {
		// Handle different error types with appropriate messages (requirements 12.2, 12.3)
		if errors.Is(err, domain.ErrConfigNotFound) {
			// Configuration file not found
			logger.Error("Configuration file not found at %s", configPath)
			logger.Error("Run 'skills-pkg init' to create a configuration file")
			return err
		}

		// File system error or other errors - distinguish and report (requirements 12.2, 12.3)
		logger.Error("Failed to verify skills: %v", err)
		logger.Error("Check file permissions and configuration")
		return err
	}

	// Check if there are no skills to verify
	if summary.TotalSkills == 0 {
		logger.Info("")
		logger.Info("No skills to verify")
		logger.Info("Use 'skills-pkg add <name> --source <type> --url <url>' to add skills")
		return nil
	}

	// Display verification results
	logger.Info("")
	logger.Info("Verification Results:")
	logger.Info("%s", "--------------------------------------------------------------------------------")

	// Display details for each verification (requirements 5.5)
	for _, result := range summary.Results {
		if result.Match {
			logger.Verbose("✓ %s (in %s): Hash verified", result.SkillName, result.InstallDir)
		} else {
			// Display warning for hash mismatch (requirement 5.5)
			logger.Error("⚠ WARNING: Hash mismatch for skill '%s' in %s", result.SkillName, result.InstallDir)
			logger.Error("  Expected: %s", result.Expected)
			logger.Error("  Actual:   %s", result.Actual)
			logger.Error("  The skill may have been tampered with or modified")
		}
	}

	// Display summary (requirement 5.6)
	logger.Info("")
	logger.Info("Verification complete:")
	logger.Info("  Total skills verified: %d", summary.TotalSkills)
	logger.Info("  Successful: %d", summary.SuccessCount)
	logger.Info("  Failed: %d", summary.FailureCount)

	if summary.FailureCount > 0 {
		logger.Info("")
		logger.Error("⚠ Warning: %d skill(s) failed verification", summary.FailureCount)
		logger.Error("This may indicate tampering or corruption")
		logger.Error("Consider reinstalling the affected skills with 'skills-pkg install'")
	}

	return nil
}
