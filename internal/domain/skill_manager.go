package domain

import (
	"context"
	"fmt"

	"github.com/mazrean/skills-pkg/internal/port"
)

// SkillManager manages skill installation, updates, and removal.
// It orchestrates package managers, config management, and hash services
// to provide a complete skill management solution.
// Requirements: 3.1-4.6, 6.1-7.6, 9.1-9.4, 10.2, 10.5, 11.4, 11.5
type SkillManager interface {
	// Install installs the specified skill. If skillName is empty, installs all skills.
	Install(ctx context.Context, skillName string) error

	// Update updates the specified skill. If skillName is empty, updates all skills.
	Update(ctx context.Context, skillName string) (*UpdateResult, error)

	// Uninstall removes the specified skill.
	Uninstall(ctx context.Context, skillName string) error
}

// UpdateResult represents the result of a skill update operation.
// It contains information about the old and new versions.
// Requirements: 7.6
type UpdateResult struct {
	SkillName  string // Name of the updated skill
	OldVersion string // Previous version
	NewVersion string // New version after update
}

// skillManagerImpl is the concrete implementation of SkillManager.
// It integrates ConfigManager, HashService, and PackageManager implementations.
// Requirements: 11.4, 11.5, 12.2, 12.3
type skillManagerImpl struct {
	configManager   *ConfigManager
	hashService     port.HashService
	packageManagers []port.PackageManager
}

// NewSkillManager creates a new SkillManager instance.
// It requires a ConfigManager for configuration persistence, a HashService for integrity verification,
// and a list of PackageManager implementations for downloading skills from various sources.
// Requirements: 11.4
func NewSkillManager(
	configManager *ConfigManager,
	hashService port.HashService,
	packageManagers []port.PackageManager,
) SkillManager {
	return &skillManagerImpl{
		configManager:   configManager,
		hashService:     hashService,
		packageManagers: packageManagers,
	}
}

// selectPackageManager selects the appropriate package manager based on the source type.
// It returns ErrInvalidSource if the source type is not supported.
// Requirements: 11.4, 11.5, 12.2, 12.3
func (s *skillManagerImpl) selectPackageManager(sourceType string) (port.PackageManager, error) {
	// Validate that source type is not empty
	if sourceType == "" {
		return nil, fmt.Errorf("%w: source type is empty. Supported types: git, npm, go-module", ErrInvalidSource)
	}

	// Find the package manager that matches the source type
	for _, pm := range s.packageManagers {
		if pm.SourceType() == sourceType {
			return pm, nil
		}
	}

	// No matching package manager found
	return nil, fmt.Errorf("%w: source type '%s' is not supported. Supported types: git, npm, go-module", ErrInvalidSource, sourceType)
}

// Install installs the specified skill.
// If skillName is empty, it installs all skills from the configuration.
// This is a placeholder implementation for task 6.1.
// Full implementation will be provided in task 6.2.
// Requirements: 6.1, 6.2
func (s *skillManagerImpl) Install(ctx context.Context, skillName string) error {
	// Placeholder implementation
	// Full implementation in task 6.2
	return nil
}

// Update updates the specified skill to the latest version.
// If skillName is empty, it updates all skills from the configuration.
// This is a placeholder implementation for task 6.1.
// Full implementation will be provided in task 6.3.
// Requirements: 7.1, 7.2
func (s *skillManagerImpl) Update(ctx context.Context, skillName string) (*UpdateResult, error) {
	// Placeholder implementation
	// Full implementation in task 6.3
	return nil, nil
}

// Uninstall removes the specified skill from all installation targets.
// This is a placeholder implementation for task 6.1.
// Full implementation will be provided in task 6.4.
// Requirements: 9.1, 9.2, 9.3, 9.4
func (s *skillManagerImpl) Uninstall(ctx context.Context, skillName string) error {
	// Placeholder implementation
	// Full implementation in task 6.4
	return nil
}
