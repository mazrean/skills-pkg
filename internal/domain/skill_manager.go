package domain

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/mazrean/skills-pkg/internal/port"
	"golang.org/x/sync/errgroup"
)

// Directory permission constants
const (
	installDirMode fs.FileMode = 0o755 // User: rwx, Group: rx, Others: rx
)

// SkillManager manages skill installation, updates, and removal.
// It orchestrates package managers, config management, and hash services
// to provide a complete skill management solution.
// Requirements: 3.1-4.6, 6.1-7.6, 9.1-9.4, 10.2, 10.5, 11.4, 11.5
type SkillManager interface {
	// Install installs the specified skill. If skillName is empty, installs all skills.
	Install(ctx context.Context, skillName string) error

	// InstallSingleSkill installs a single skill that has been added to the config.
	// It downloads the skill, calculates the hash, and updates the config.
	// If saveConfig is true, it also saves the configuration file.
	// This is useful when you want to add a skill to the config and install it in one operation.
	InstallSingleSkill(ctx context.Context, config *Config, skill *Skill, saveConfig bool) error

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
		return nil, fmt.Errorf("%w: source type is empty. Supported types: git, go-module", ErrInvalidSource)
	}

	// Find the package manager that matches the source type
	for _, pm := range s.packageManagers {
		if pm.SourceType() == sourceType {
			return pm, nil
		}
	}

	// No matching package manager found
	return nil, fmt.Errorf("%w: source type '%s' is not supported. Supported types: git, go-module", ErrInvalidSource, sourceType)
}

// Install installs the specified skill.
// If skillName is empty, it installs all skills from the configuration.
// Multiple skills are installed concurrently for better performance.
// Requirements: 6.1, 6.2
func (s *skillManagerImpl) Install(ctx context.Context, skillName string) error {
	// Load configuration (Requirement 6.2)
	config, err := s.configManager.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine which skills to install (Requirements 6.1, 6.2)
	var skillsToInstall []*Skill
	if skillName == "" {
		// Install all skills (Requirement 6.1)
		skillsToInstall = config.Skills
	} else {
		// Install specific skill (Requirement 6.2)
		skill := config.FindSkillByName(skillName)
		if skill == nil {
			// Requirement 6.3, 12.2, 12.3
			return fmt.Errorf("%w: skill '%s' not found in configuration. Use 'skills-pkg add %s --source <type> --url <url>' to add it first", ErrSkillNotFound, skillName, skillName)
		}
		skillsToInstall = []*Skill{skill}
	}

	// Install skills concurrently using errgroup
	eg, egCtx := errgroup.WithContext(ctx)
	for _, skill := range skillsToInstall {
		eg.Go(func() error {
			return s.InstallSingleSkill(egCtx, config, skill, false)
		})
	}

	// Wait for all installations to complete
	if err := eg.Wait(); err != nil {
		return err
	}

	// Save configuration once after all skills are installed
	if err := s.configManager.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// copySkillToTargets copies a skill to all install target directories concurrently.
// It creates missing directories automatically and handles errors appropriately.
// Requirements: 3.4, 4.4, 6.6, 10.2, 10.5, 12.2, 12.3
func (s *skillManagerImpl) copySkillToTargets(sourcePath, skillName string, installTargets []string) error {
	var eg errgroup.Group

	for _, target := range installTargets {
		eg.Go(func() error {
			// Create skill directory in target (Requirement 6.6)
			skillDir := target + "/" + skillName

			// Remove existing skill directory if it exists
			if err := os.RemoveAll(skillDir); err != nil {
				return fmt.Errorf("failed to remove existing skill directory at %s: %w", skillDir, err)
			}

			// Create parent directory if it doesn't exist (Requirement 6.6)
			if err := os.MkdirAll(target, installDirMode); err != nil {
				return fmt.Errorf("failed to create install target directory %s: %w", target, err)
			}

			// Copy skill directory
			if err := copyDir(sourcePath, skillDir); err != nil {
				return fmt.Errorf("failed to copy skill to %s: %w", skillDir, err)
			}

			return nil
		})
	}

	return eg.Wait()
}

// verifyInstalledSkill verifies the hash of an installed skill in all target directories concurrently.
// It returns an error if any verification fails.
// Requirements: 6.4, 6.5
func (s *skillManagerImpl) verifyInstalledSkill(ctx context.Context, skill *Skill, installTargets []string) error {
	// Skip verification if HashValue is empty (e.g., when using go.mod version)
	// In this case, integrity is verified by go.sum
	if skill.HashValue == "" {
		return nil
	}

	eg, egCtx := errgroup.WithContext(ctx)

	for _, target := range installTargets {
		eg.Go(func() error {
			skillDir := target + "/" + skill.Name

			// Calculate hash of installed skill
			hashResult, err := s.hashService.CalculateHash(egCtx, skillDir)
			if err != nil {
				return fmt.Errorf("failed to calculate hash for verification in %s: %w", skillDir, err)
			}

			// Compare with expected hash
			if hashResult.Value != skill.HashValue {
				return fmt.Errorf("hash mismatch in %s: expected %s, got %s", skillDir, skill.HashValue, hashResult.Value)
			}

			return nil
		})
	}

	return eg.Wait()
}

// copyDir recursively copies a directory from src to dst.
// It creates the destination directory if it doesn't exist.
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if mkdirErr := os.MkdirAll(dst, srcInfo.Mode()); mkdirErr != nil {
		return mkdirErr
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := src + "/" + entry.Name()
		dstPath := dst + "/" + entry.Name()

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source file info for permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Write destination file
	if err := os.WriteFile(dst, data, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}

// InstallSingleSkill installs a single skill.
// If saveConfig is true, saves the configuration after updating skill metadata.
// This method is public to allow external callers (like add command) to install a single skill.
// Requirements: 3.3, 3.4, 4.3, 4.4, 5.3, 6.2, 6.4, 6.5, 6.6, 10.2, 10.5, 12.1, 12.2, 12.3
func (s *skillManagerImpl) InstallSingleSkill(ctx context.Context, config *Config, skill *Skill, saveConfig bool) error {
	// Progress information (Requirement 12.1)
	fmt.Printf("Installing skill '%s' from %s...\n", skill.Name, skill.Source)

	// Select appropriate package manager (Requirement 11.4)
	pm, err := s.selectPackageManager(skill.Source)
	if err != nil {
		return fmt.Errorf("failed to select package manager for skill '%s': %w", skill.Name, err)
	}

	// Create source from skill
	source := &port.Source{
		Type: skill.Source,
		URL:  skill.URL,
	}

	// Download skill (Requirements 3.3, 4.3)
	fmt.Printf("Downloading skill '%s' version %s...\n", skill.Name, skill.Version)
	downloadResult, err := pm.Download(ctx, source, skill.Version)
	if err != nil {
		return fmt.Errorf("failed to download skill '%s': %w. Check your network connection and source URL", skill.Name, err)
	}

	// Determine the source path to use for installation and hash calculation
	sourcePath := downloadResult.Path
	if skill.SubDir != "" {
		// Use the subdirectory within the downloaded content
		sourcePath = downloadResult.Path + "/" + skill.SubDir

		// Verify that the subdirectory exists
		if _, statErr := os.Stat(sourcePath); statErr != nil {
			if os.IsNotExist(statErr) {
				return fmt.Errorf("subdirectory '%s' not found in downloaded skill '%s'. Available content is in: %s", skill.SubDir, skill.Name, downloadResult.Path)
			}
			return fmt.Errorf("failed to access subdirectory '%s' in skill '%s': %w", skill.SubDir, skill.Name, statErr)
		}
		fmt.Printf("Using subdirectory '%s' from downloaded content...\n", skill.SubDir)
	}

	// Calculate hash only if not from go.mod (Requirement 5.3)
	// When version is resolved from go.mod, rely on go.sum for integrity verification
	if !downloadResult.FromGoMod {
		// Update version
		skill.Version = downloadResult.Version

		fmt.Printf("Calculating hash for skill '%s'...\n", skill.Name)
		hashResult, err := s.hashService.CalculateHash(ctx, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to calculate hash for skill '%s': %w", skill.Name, err)
		}
		skill.HashAlgo = hashResult.Algorithm
		skill.HashValue = hashResult.Value
	} else {
		// Clear version and hash values when using go.mod version
		// Version and hash verification will use go.mod/go.sum instead
		// This ensures go.mod remains the source of truth
		skill.Version = ""
		skill.HashAlgo = ""
		skill.HashValue = ""
	}

	// Save updated configuration if requested (Requirement 5.3)
	if saveConfig {
		if err := s.configManager.Save(ctx, config); err != nil {
			return fmt.Errorf("failed to save configuration after hash calculation: %w", err)
		}
	}

	// Get install targets (Requirement 6.2)
	installTargets := config.InstallTargets
	if len(installTargets) == 0 {
		return fmt.Errorf("no install targets configured. Run 'skills-pkg init --install-dir <dir>' to configure install targets")
	}

	// Install to all targets (Requirements 3.4, 4.4, 10.2, 10.5, 6.6)
	fmt.Printf("Installing skill '%s' to %d target(s)...\n", skill.Name, len(installTargets))
	if err := s.copySkillToTargets(sourcePath, skill.Name, installTargets); err != nil {
		return fmt.Errorf("failed to copy skill '%s' to install targets: %w. Check file permissions", skill.Name, err)
	}

	// Verify hash after installation (Requirements 6.4, 6.5)
	fmt.Printf("Verifying installation of skill '%s'...\n", skill.Name)
	if err := s.verifyInstalledSkill(ctx, skill, installTargets); err != nil {
		// Show warning but continue (Requirement 6.5, 12.1, 12.2)
		fmt.Printf("WARNING: Hash verification failed for skill '%s': %v. The skill may have been tampered with during installation.\n", skill.Name, err)
	}

	fmt.Printf("Successfully installed skill '%s'\n", skill.Name)
	return nil
}

// Update updates the specified skill to the latest version.
// If skillName is empty, it updates all skills from the configuration.
// Requirements: 5.3, 7.1, 7.2, 7.5, 7.6, 12.1, 12.2, 12.3
func (s *skillManagerImpl) Update(ctx context.Context, skillName string) (*UpdateResult, error) {
	// Load configuration (Requirement 7.1)
	config, err := s.configManager.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine which skills to update (Requirements 7.1, 7.2)
	var skillsToUpdate []*Skill
	if skillName == "" {
		// Update all skills (Requirement 7.1)
		skillsToUpdate = config.Skills
	} else {
		// Update specific skill (Requirement 7.2)
		skill := config.FindSkillByName(skillName)
		if skill == nil {
			// Requirement 12.2, 12.3
			return nil, fmt.Errorf("%w: skill '%s' not found in configuration", ErrSkillNotFound, skillName)
		}
		skillsToUpdate = []*Skill{skill}
	}

	// Update skills concurrently using errgroup
	results := make([]*UpdateResult, len(skillsToUpdate))
	eg, egCtx := errgroup.WithContext(ctx)

	for i, skill := range skillsToUpdate {
		eg.Go(func() error {
			result, err := s.updateSingleSkill(egCtx, config, skill, false)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	// Wait for all updates to complete
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Save configuration once after all skills are updated
	if err := s.configManager.Save(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	// Return the last update result
	// For single skill update, this is the result for that skill
	// For all skills update, this is the result for the last skill
	return results[len(results)-1], nil
}

// updateSingleSkill updates a single skill to the latest version.
// If saveConfig is true, saves the configuration after updating skill metadata.
// Requirements: 5.3, 7.1, 7.2, 7.5, 7.6, 12.1, 12.2, 12.3
func (s *skillManagerImpl) updateSingleSkill(ctx context.Context, config *Config, skill *Skill, saveConfig bool) (*UpdateResult, error) {
	// Record old version (Requirement 7.6)
	oldVersion := skill.Version

	// Progress information (Requirement 12.1)
	fmt.Printf("Updating skill '%s' from version %s...\n", skill.Name, oldVersion)

	// Select appropriate package manager (Requirement 11.4)
	pm, err := s.selectPackageManager(skill.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to select package manager for skill '%s': %w", skill.Name, err)
	}

	// Create source from skill
	source := &port.Source{
		Type: skill.Source,
		URL:  skill.URL,
	}

	// Get latest version (Requirement 7.4, 12.1)
	fmt.Printf("Fetching latest version for skill '%s'...\n", skill.Name)
	latestVersion, err := pm.GetLatestVersion(ctx, source)
	if err != nil {
		// Network error handling (Requirement 12.2, 12.3)
		if IsNetworkError(err) {
			return nil, fmt.Errorf("failed to get latest version for skill '%s': %w. Check your network connection and source URL", skill.Name, err)
		}
		return nil, fmt.Errorf("failed to get latest version for skill '%s': %w", skill.Name, err)
	}

	// Download latest version (Requirement 7.4)
	fmt.Printf("Downloading skill '%s' version %s...\n", skill.Name, latestVersion)
	downloadResult, err := pm.Download(ctx, source, latestVersion)
	if err != nil {
		// Network error handling (Requirement 12.2, 12.3)
		if IsNetworkError(err) {
			return nil, fmt.Errorf("failed to download skill '%s': %w. Check your network connection and source URL", skill.Name, err)
		}
		return nil, fmt.Errorf("failed to download skill '%s': %w", skill.Name, err)
	}

	// Determine the source path to use for installation and hash calculation
	sourcePath := downloadResult.Path
	if skill.SubDir != "" {
		// Use the subdirectory within the downloaded content
		sourcePath = downloadResult.Path + "/" + skill.SubDir

		// Verify that the subdirectory exists
		if _, statErr := os.Stat(sourcePath); statErr != nil {
			if os.IsNotExist(statErr) {
				return nil, fmt.Errorf("subdirectory '%s' not found in downloaded skill '%s'. Available content is in: %s", skill.SubDir, skill.Name, downloadResult.Path)
			}
			return nil, fmt.Errorf("failed to access subdirectory '%s' in skill '%s': %w", skill.SubDir, skill.Name, statErr)
		}
		fmt.Printf("Using subdirectory '%s' from downloaded content...\n", skill.SubDir)
	}

	// Calculate hash only if not from go.mod (Requirement 5.3, 7.5)
	// When version is resolved from go.mod, rely on go.sum for integrity verification
	if !downloadResult.FromGoMod {
		// Update version
		skill.Version = downloadResult.Version

		fmt.Printf("Calculating hash for skill '%s'...\n", skill.Name)
		hashResult, err := s.hashService.CalculateHash(ctx, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash for skill '%s': %w", skill.Name, err)
		}
		skill.HashAlgo = hashResult.Algorithm
		skill.HashValue = hashResult.Value
	} else {
		// Clear version and hash values when using go.mod version
		// Version and hash verification will use go.mod/go.sum instead
		// This ensures go.mod remains the source of truth
		skill.Version = ""
		skill.HashAlgo = ""
		skill.HashValue = ""
	}

	// Save updated configuration if requested (Requirement 5.3, 7.5)
	if saveConfig {
		if err := s.configManager.Save(ctx, config); err != nil {
			// Filesystem error handling (Requirement 12.2, 12.3)
			return nil, fmt.Errorf("failed to save configuration after update: %w. Check file permissions", err)
		}
	}

	// Get install targets
	installTargets := config.InstallTargets
	if len(installTargets) > 0 {
		// Install to all targets (Requirements 10.2, 10.5)
		fmt.Printf("Installing updated skill '%s' to %d target(s)...\n", skill.Name, len(installTargets))
		if err := s.copySkillToTargets(sourcePath, skill.Name, installTargets); err != nil {
			// Filesystem error handling (Requirement 12.2, 12.3)
			return nil, fmt.Errorf("failed to copy updated skill '%s' to install targets: %w. Check file permissions", skill.Name, err)
		}
	}

	// Display update information (Requirement 7.6, 12.1)
	fmt.Printf("Successfully updated skill '%s' from %s to %s\n", skill.Name, oldVersion, downloadResult.Version)

	// Return update result (Requirement 7.6)
	return &UpdateResult{
		SkillName:  skill.Name,
		OldVersion: oldVersion,
		NewVersion: downloadResult.Version,
	}, nil
}

// Uninstall removes the specified skill from all installation targets
// and from the configuration file.
// Requirements: 9.1, 9.2, 9.3, 9.4, 12.2
func (s *skillManagerImpl) Uninstall(ctx context.Context, skillName string) error {
	// Progress information (Requirement 12.1)
	fmt.Printf("Uninstalling skill '%s'...\n", skillName)

	// Load configuration (Requirement 9.2)
	config, err := s.configManager.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if skill exists (Requirement 9.3)
	skill := config.FindSkillByName(skillName)
	if skill == nil {
		// Requirement 9.3, 12.2
		return fmt.Errorf("%w: skill '%s' not found in configuration", ErrSkillNotFound, skillName)
	}

	// Remove skill from all install target directories (Requirement 9.1)
	installTargets := config.InstallTargets
	for _, target := range installTargets {
		skillDir := target + "/" + skillName

		// Remove skill directory if it exists
		if err := os.RemoveAll(skillDir); err != nil {
			// Filesystem error handling (Requirement 12.2, 12.3)
			return fmt.Errorf("failed to remove skill directory at %s: %w. Check file permissions", skillDir, err)
		}
		fmt.Printf("Removed skill '%s' from %s\n", skillName, target)
	}

	// Remove skill from configuration (Requirement 9.2)
	if err := s.configManager.RemoveSkill(ctx, skillName); err != nil {
		return fmt.Errorf("failed to remove skill from configuration: %w", err)
	}

	// Success message (Requirement 9.4, 12.2)
	fmt.Printf("Successfully uninstalled skill '%s'\n", skillName)
	return nil
}
