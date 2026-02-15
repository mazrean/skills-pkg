package domain

import (
	"context"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// ConfigManager manages the reading and writing of the .skillspkg.toml configuration file.
// It provides methods for initializing, loading, and saving configuration.
// Requirements: 1.1-1.5, 2.1-2.6, 8.1-8.4, 10.1, 11.4
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new ConfigManager instance.
// The configPath parameter specifies the path to the .skillspkg.toml file.
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{configPath: configPath}
}

// Initialize creates a new .skillspkg.toml file with the specified install directories.
// It returns ErrConfigExists if the configuration file already exists.
// Requirements: 1.1, 1.4, 1.5, 12.2, 12.3
func (m *ConfigManager) Initialize(ctx context.Context, installDirs []string) error {
	// Check if config file already exists (requirement 1.4)
	if _, err := os.Stat(m.configPath); err == nil {
		// File exists - return error with clear message
		return fmt.Errorf("%w: configuration file already exists at %s. Remove the existing file or use a different path", ErrConfigExists, m.configPath)
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking file existence
		return fmt.Errorf("failed to check configuration file existence: %w", err)
	}

	// Create empty config with specified install directories (requirement 1.5)
	config := &Config{
		Skills:         []*Skill{},
		InstallTargets: installDirs,
	}

	// Marshal config to TOML format
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write config file (requirement 1.1)
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file to %s: %w. Check file permissions and directory existence", m.configPath, err)
	}

	return nil
}
