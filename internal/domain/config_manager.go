package domain

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// File permission constants for configuration files
const (
	configFileMode fs.FileMode = 0o644 // User: rw, Group: r, Others: r
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

	// Use Save method to write the config file (requirement 1.1)
	return m.Save(ctx, config)
}

// Load reads the .skillspkg.toml file and returns the configuration.
// It returns ErrConfigNotFound if the configuration file does not exist.
// It provides detailed error messages for TOML parse errors (requirement 2.6).
// Requirements: 2.1, 2.6, 12.2, 12.3
func (m *ConfigManager) Load(ctx context.Context) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found - return sentinel error (requirement 12.2, 12.3)
			return nil, fmt.Errorf("%w: configuration file not found at %s. Run 'skills-pkg init' to create one", ErrConfigNotFound, m.configPath)
		}
		// Other file system error (requirement 12.2, 12.3)
		return nil, fmt.Errorf("failed to read configuration file at %s: %w. Check file permissions", m.configPath, err)
	}

	// Parse TOML content
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		// TOML parse error - provide detailed error message (requirement 2.6)
		return nil, fmt.Errorf("failed to parse configuration file at %s: %w. Ensure the file is valid TOML format", m.configPath, err)
	}

	// Validate the loaded configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to the .skillspkg.toml file.
// It provides detailed error messages for file system errors (requirement 12.2, 12.3).
// Requirements: 2.1, 12.2, 12.3
func (m *ConfigManager) Save(ctx context.Context, config *Config) error {
	// Validate the configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Marshal config to TOML format
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write config file
	if err := os.WriteFile(m.configPath, data, configFileMode); err != nil {
		// File system error - provide detailed error message (requirement 12.2, 12.3)
		return fmt.Errorf("failed to write configuration file to %s: %w. Check file permissions and directory existence", m.configPath, err)
	}

	return nil
}
