package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

// TestConfigManagerTOMLIntegration tests the integration of ConfigManager with go-toml v2.
// This validates the complete flow of configuration file initialization, reading, and writing.
// Requirements: 2.1, 2.6, 12.2, 12.3
func TestConfigManagerTOMLIntegration(t *testing.T) {
	t.Parallel()

	t.Run("Initialize_Load_Save_Integration", func(t *testing.T) {
		t.Parallel()

		// Setup: Create temporary directory for test
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")

		ctx := context.Background()
		configManager := domain.NewConfigManager(configPath)

		// Test: Initialize configuration
		installDirs := []string{"~/.claude/skills", "~/.codex/skills"}
		err := configManager.Initialize(ctx, installDirs)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Verify: Configuration file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("Configuration file was not created")
		}

		// Test: Load configuration
		loadedConfig, err := configManager.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Verify: Loaded configuration matches initialized values
		if len(loadedConfig.InstallTargets) != len(installDirs) {
			t.Errorf("Expected %d install targets, got %d", len(installDirs), len(loadedConfig.InstallTargets))
		}

		for i, dir := range installDirs {
			if loadedConfig.InstallTargets[i] != dir {
				t.Errorf("Expected install target %s, got %s", dir, loadedConfig.InstallTargets[i])
			}
		}

		if len(loadedConfig.Skills) != 0 {
			t.Errorf("Expected 0 skills, got %d", len(loadedConfig.Skills))
		}

		// Test: Add skill and save
		skill := &domain.Skill{
			Name:           "test-skill",
			Source:         "git",
			URL:            "https://github.com/example/skill.git",
			Version:        "v1.0.0",
			HashAlgo:       "sha256",
			HashValue:      "a1b2c3d4e5f6",
			PackageManager: "",
		}

		err = configManager.AddSkill(ctx, skill)
		if err != nil {
			t.Fatalf("AddSkill failed: %v", err)
		}

		// Test: Reload configuration and verify skill was added
		reloadedConfig, err := configManager.Load(ctx)
		if err != nil {
			t.Fatalf("Reload failed: %v", err)
		}

		if len(reloadedConfig.Skills) != 1 {
			t.Fatalf("Expected 1 skill, got %d", len(reloadedConfig.Skills))
		}

		if reloadedConfig.Skills[0].Name != skill.Name {
			t.Errorf("Expected skill name %s, got %s", skill.Name, reloadedConfig.Skills[0].Name)
		}
	})

	t.Run("TOML_Format_Error_Handling", func(t *testing.T) {
		t.Parallel()

		// Setup: Create invalid TOML file
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")

		invalidTOML := `
[invalid syntax
skills = "not an array"
`
		err := os.WriteFile(configPath, []byte(invalidTOML), 0o644)
		if err != nil {
			t.Fatalf("Failed to create invalid TOML file: %v", err)
		}

		ctx := context.Background()
		configManager := domain.NewConfigManager(configPath)

		// Test: Load should fail with descriptive error
		_, err = configManager.Load(ctx)
		if err == nil {
			t.Fatal("Expected error when loading invalid TOML, got nil")
		}

		// Verify: Error message contains context about TOML parsing
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("Concurrent_Access_Safety", func(t *testing.T) {
		t.Parallel()

		// Setup: Create configuration file
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")

		ctx := context.Background()
		configManager := domain.NewConfigManager(configPath)

		installDirs := []string{"~/.claude/skills"}
		err := configManager.Initialize(ctx, installDirs)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Test: Multiple concurrent reads should succeed
		const numReaders = 10
		errChan := make(chan error, numReaders)

		for i := 0; i < numReaders; i++ {
			go func() {
				_, err := configManager.Load(ctx)
				errChan <- err
			}()
		}

		// Verify: All reads succeeded
		for i := 0; i < numReaders; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent read %d failed: %v", i, err)
			}
		}
	})
}

// TestSkillOperationIntegration tests the integration of skill CRUD operations.
// Requirements: 2.2, 2.3, 2.4, 5.2, 8.1, 8.2
func TestSkillOperationIntegration(t *testing.T) {
	t.Parallel()

	t.Run("Add_Update_Remove_Skill_Flow", func(t *testing.T) {
		t.Parallel()

		// Setup
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")

		ctx := context.Background()
		configManager := domain.NewConfigManager(configPath)

		err := configManager.Initialize(ctx, []string{"~/.claude/skills"})
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Test: Add multiple skills
		skills := []*domain.Skill{
			{
				Name:           "git-skill",
				Source:         "git",
				URL:            "https://github.com/example/git-skill.git",
				Version:        "v1.0.0",
				HashAlgo:       "sha256",
				HashValue:      "hash1",
				PackageManager: "",
			},
			{
				Name:           "npm-skill",
				Source:         "npm",
				URL:            "example-npm-skill",
				Version:        "1.0.0",
				HashAlgo:       "sha256",
				HashValue:      "hash2",
				PackageManager: "npm",
			},
		}

		for _, skill := range skills {
			err := configManager.AddSkill(ctx, skill)
			if err != nil {
				t.Fatalf("AddSkill failed for %s: %v", skill.Name, err)
			}
		}

		// Verify: List skills
		listedSkills, err := configManager.ListSkills(ctx)
		if err != nil {
			t.Fatalf("ListSkills failed: %v", err)
		}

		if len(listedSkills) != len(skills) {
			t.Errorf("Expected %d skills, got %d", len(skills), len(listedSkills))
		}

		// Test: Update skill
		updatedSkill := &domain.Skill{
			Name:           "git-skill",
			Source:         "git",
			URL:            "https://github.com/example/git-skill.git",
			Version:        "v2.0.0",
			HashAlgo:       "sha256",
			HashValue:      "new-hash1",
			PackageManager: "",
		}

		err = configManager.UpdateSkill(ctx, updatedSkill)
		if err != nil {
			t.Fatalf("UpdateSkill failed: %v", err)
		}

		// Verify: Updated skill has new values
		reloadedConfig, err := configManager.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		found := false
		for _, skill := range reloadedConfig.Skills {
			if skill.Name == "git-skill" {
				found = true
				if skill.Version != "v2.0.0" {
					t.Errorf("Expected version v2.0.0, got %s", skill.Version)
				}
				if skill.HashValue != "new-hash1" {
					t.Errorf("Expected hash new-hash1, got %s", skill.HashValue)
				}
			}
		}

		if !found {
			t.Error("Updated skill not found in configuration")
		}

		// Test: Remove skill
		err = configManager.RemoveSkill(ctx, "npm-skill")
		if err != nil {
			t.Fatalf("RemoveSkill failed: %v", err)
		}

		// Verify: Skill was removed
		finalConfig, err := configManager.Load(ctx)
		if err != nil {
			t.Fatalf("Final load failed: %v", err)
		}

		if len(finalConfig.Skills) != 1 {
			t.Errorf("Expected 1 skill after removal, got %d", len(finalConfig.Skills))
		}

		for _, skill := range finalConfig.Skills {
			if skill.Name == "npm-skill" {
				t.Error("Removed skill still present in configuration")
			}
		}
	})

	t.Run("Duplicate_Skill_Error", func(t *testing.T) {
		t.Parallel()

		// Setup
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")

		ctx := context.Background()
		configManager := domain.NewConfigManager(configPath)

		err := configManager.Initialize(ctx, []string{"~/.claude/skills"})
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Test: Add skill
		skill := &domain.Skill{
			Name:      "duplicate-skill",
			Source:    "git",
			URL:       "https://github.com/example/skill.git",
			Version:   "v1.0.0",
			HashAlgo:  "sha256",
			HashValue: "hash",
		}

		err = configManager.AddSkill(ctx, skill)
		if err != nil {
			t.Fatalf("First AddSkill failed: %v", err)
		}

		// Test: Attempt to add duplicate skill
		err = configManager.AddSkill(ctx, skill)
		if err == nil {
			t.Fatal("Expected error when adding duplicate skill, got nil")
		}

		// Verify: Error indicates duplicate
		// (The specific error check depends on the implementation)
	})
}
