package domain

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// VerifyResult represents the result of verifying a single skill's hash.
// It contains details about the verification including expected and actual hash values.
// Requirements: 5.4, 5.5
type VerifyResult struct {
	SkillName  string // Name of the skill being verified
	InstallDir string // Installation directory path
	Expected   string // Expected hash value from configuration
	Actual     string // Actual hash value calculated from directory
	Match      bool   // Whether the hashes match
}

// VerifySummary represents the summary of verifying all skills.
// It provides statistics about the verification results.
// Requirements: 5.6
type VerifySummary struct {
	Results      []*VerifyResult // Detailed results for each skill
	TotalSkills  int             // Total number of skills verified
	SuccessCount int             // Number of skills with matching hashes
	FailureCount int             // Number of skills with mismatching hashes
}

// HashVerifier manages hash verification for skills.
// It verifies skill integrity by comparing stored hashes with calculated hashes.
// Requirements: 5.4, 5.5, 5.6
type HashVerifier struct {
	configManager *ConfigManager
	hashService   port.HashService
}

// NewHashVerifier creates a new HashVerifier instance.
// The configManager is used to load skill configuration and expected hashes.
// The hashService is used to calculate directory hashes.
func NewHashVerifier(configManager *ConfigManager, hashService port.HashService) *HashVerifier {
	return &HashVerifier{
		configManager: configManager,
		hashService:   hashService,
	}
}

// Verify verifies the hash of a single skill in a specific installation directory.
// It compares the expected hash from configuration with the actual hash of the directory.
// Returns a VerifyResult containing detailed verification information.
// Requirements: 5.4, 5.5
func (v *HashVerifier) Verify(ctx context.Context, skillName string, installDir string) (*VerifyResult, error) {
	// Load configuration to get expected hash
	config, err := v.configManager.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the skill in configuration
	skill := config.FindSkillByName(skillName)
	if skill == nil {
		return nil, fmt.Errorf("%w: skill '%s' not found in configuration", ErrSkillNotFound, skillName)
	}

	// Calculate actual hash of the skill directory
	hashResult, err := v.hashService.CalculateHash(ctx, installDir)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash for skill '%s' in directory %s: %w", skillName, installDir, err)
	}

	// Compare expected and actual hashes
	match := skill.HashValue == hashResult.Value

	return &VerifyResult{
		SkillName:  skillName,
		InstallDir: installDir,
		Expected:   skill.HashValue,
		Actual:     hashResult.Value,
		Match:      match,
	}, nil
}

// VerifyAll verifies the hashes of all skills in all installation target directories.
// It returns a summary containing statistics and detailed results for each verification.
// Requirements: 5.4, 5.6
func (v *HashVerifier) VerifyAll(ctx context.Context) (*VerifySummary, error) {
	// Load configuration
	config, err := v.configManager.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get installation target directories
	installTargets, err := v.configManager.GetInstallTargets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation targets: %w", err)
	}

	// Initialize summary
	summary := &VerifySummary{
		TotalSkills:  0,
		SuccessCount: 0,
		FailureCount: 0,
		Results:      []*VerifyResult{},
	}

	// If there are no skills, return empty summary
	if len(config.Skills) == 0 {
		return summary, nil
	}

	// Verify each skill in each installation target
	for _, skill := range config.Skills {
		for _, installTarget := range installTargets {
			// Construct the skill directory path
			skillDir := filepath.Join(installTarget, skill.Name)

			// Verify the skill
			result, err := v.Verify(ctx, skill.Name, skillDir)
			if err != nil {
				// If verification fails (e.g., directory doesn't exist), record as failure
				result = &VerifyResult{
					SkillName:  skill.Name,
					InstallDir: skillDir,
					Expected:   skill.HashValue,
					Actual:     "",
					Match:      false,
				}
			}

			// Update summary statistics
			summary.TotalSkills++
			if result.Match {
				summary.SuccessCount++
			} else {
				summary.FailureCount++
			}

			// Add result to the list
			summary.Results = append(summary.Results, result)
		}
	}

	return summary, nil
}
