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

// AddSkillToConfig adds a new skill entry to the configuration in memory.
// It returns the updated Config without saving to file.
// This is useful when you want to add a skill and perform additional operations
// (like installation) before saving the configuration.
// It returns ErrSkillExists if a skill with the same name already exists.
// Requirements: 2.2, 2.3, 2.4, 5.2, 12.2, 12.3
func (m *ConfigManager) AddSkillToConfig(ctx context.Context, skill *Skill) (*Config, error) {
	// Validate the skill before adding
	if err := skill.Validate(); err != nil {
		return nil, fmt.Errorf("skill validation failed: %w", err)
	}

	// Load the current config
	config, err := m.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check for duplicate skill names (requirement 2.2)
	if config.HasSkill(skill.Name) {
		return nil, fmt.Errorf("%w: skill '%s' already exists in configuration", ErrSkillExists, skill.Name)
	}

	// Add the skill to the config
	config.Skills = append(config.Skills, skill)

	return config, nil
}

// AddSkill adds a new skill entry to the configuration.
// It returns ErrSkillExists if a skill with the same name already exists.
// Requirements: 2.2, 2.3, 2.4, 5.2, 12.2, 12.3
func (m *ConfigManager) AddSkill(ctx context.Context, skill *Skill) error {
	// Add skill to config (without saving)
	config, err := m.AddSkillToConfig(ctx, skill)
	if err != nil {
		return err
	}

	// Save the updated config
	if err := m.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to save configuration after adding skill '%s': %w", skill.Name, err)
	}

	return nil
}

// UpdateSkill updates an existing skill entry in the configuration.
// It returns ErrSkillNotFound if the skill does not exist.
// Requirements: 2.2, 5.2, 12.2, 12.3
func (m *ConfigManager) UpdateSkill(ctx context.Context, skill *Skill) error {
	// Validate the skill before updating
	if err := skill.Validate(); err != nil {
		return fmt.Errorf("skill validation failed: %w", err)
	}

	// Load the current config
	config, err := m.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the skill to update
	existingSkill := config.FindSkillByName(skill.Name)
	if existingSkill == nil {
		return fmt.Errorf("%w: skill '%s' not found in configuration", ErrSkillNotFound, skill.Name)
	}

	// Update the skill fields
	existingSkill.Source = skill.Source
	existingSkill.URL = skill.URL
	existingSkill.Version = skill.Version
	existingSkill.HashAlgo = skill.HashAlgo
	existingSkill.HashValue = skill.HashValue
	existingSkill.SubDir = skill.SubDir

	// Save the updated config
	if err := m.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to save configuration after updating skill '%s': %w", skill.Name, err)
	}

	return nil
}

// RemoveSkill removes a skill entry from the configuration.
// It returns ErrSkillNotFound if the skill does not exist.
// Requirements: 9.2, 12.2, 12.3
func (m *ConfigManager) RemoveSkill(ctx context.Context, skillName string) error {
	// Load the current config
	config, err := m.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the skill index
	skillIndex := -1
	for i, skill := range config.Skills {
		if skill.Name == skillName {
			skillIndex = i
			break
		}
	}

	// Check if skill exists
	if skillIndex == -1 {
		return fmt.Errorf("%w: skill '%s' not found in configuration", ErrSkillNotFound, skillName)
	}

	// Remove the skill from the slice
	config.Skills = append(config.Skills[:skillIndex], config.Skills[skillIndex+1:]...)

	// Save the updated config
	if err := m.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to save configuration after removing skill '%s': %w", skillName, err)
	}

	return nil
}

// ListSkills returns all skills from the configuration.
// Requirements: 8.1, 8.2, 12.2, 12.3
func (m *ConfigManager) ListSkills(ctx context.Context) ([]*Skill, error) {
	// Load the current config
	config, err := m.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Return the skills list
	return config.Skills, nil
}

// GetInstallTargets returns the list of installation target directories from the configuration.
// This is the single source of truth for where skills should be installed.
// Requirements: 1.2, 2.5, 10.1
func (m *ConfigManager) GetInstallTargets(ctx context.Context) ([]string, error) {
	// Load the current config
	config, err := m.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Return the install targets list
	return config.InstallTargets, nil
}
