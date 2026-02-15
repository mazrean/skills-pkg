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

	tests := []struct {
		testFunc  func(t *testing.T, ctx context.Context, configPath string, configManager *domain.ConfigManager)
		setupFunc func(t *testing.T) (ctx context.Context, configPath string, configManager *domain.ConfigManager)
		name      string
	}{
		{
			name: "Initialize_Load_Save_Integration",
			setupFunc: func(t *testing.T) (context.Context, string, *domain.ConfigManager) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)
				return ctx, configPath, configManager
			},
			testFunc: func(t *testing.T, ctx context.Context, configPath string, configManager *domain.ConfigManager) {
				installDirs := []string{"~/.claude/skills", "~/.codex/skills"}
				err := configManager.Initialize(ctx, installDirs)
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}

				if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
					t.Fatalf("Configuration file was not created")
				}

				loadedConfig, err := configManager.Load(ctx)
				if err != nil {
					t.Fatalf("Load failed: %v", err)
				}

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

				skill := &domain.Skill{
					Name:      "test-skill",
					Source:    "git",
					URL:       "https://github.com/example/skill.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "a1b2c3d4e5f6",
				}

				err = configManager.AddSkill(ctx, skill)
				if err != nil {
					t.Fatalf("AddSkill failed: %v", err)
				}

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
			},
		},
		{
			name: "TOML_Format_Error_Handling",
			setupFunc: func(t *testing.T) (context.Context, string, *domain.ConfigManager) {
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
				return ctx, configPath, configManager
			},
			testFunc: func(t *testing.T, ctx context.Context, configPath string, configManager *domain.ConfigManager) {
				_, err := configManager.Load(ctx)
				if err == nil {
					t.Fatal("Expected error when loading invalid TOML, got nil")
				}

				errMsg := err.Error()
				if errMsg == "" {
					t.Error("Expected non-empty error message")
				}
			},
		},
		{
			name: "Concurrent_Access_Safety",
			setupFunc: func(t *testing.T) (context.Context, string, *domain.ConfigManager) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)

				installDirs := []string{"~/.claude/skills"}
				err := configManager.Initialize(ctx, installDirs)
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}
				return ctx, configPath, configManager
			},
			testFunc: func(t *testing.T, ctx context.Context, configPath string, configManager *domain.ConfigManager) {
				const numReaders = 10
				errChan := make(chan error, numReaders)

				for range numReaders {
					go func() {
						_, err := configManager.Load(ctx)
						errChan <- err
					}()
				}

				for i := range numReaders {
					if err := <-errChan; err != nil {
						t.Errorf("Concurrent read %d failed: %v", i, err)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, configPath, configManager := tt.setupFunc(t)
			tt.testFunc(t, ctx, configPath, configManager)
		})
	}
}

// TestSkillOperationIntegration tests the integration of skill CRUD operations.
// Requirements: 2.2, 2.3, 2.4, 5.2, 8.1, 8.2
func TestSkillOperationIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc  func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager)
		setupFunc func(t *testing.T) (ctx context.Context, configManager *domain.ConfigManager)
		name      string
	}{
		{
			name: "Add_Update_Remove_Skill_Flow",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)

				err := configManager.Initialize(ctx, []string{"~/.claude/skills"})
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}
				return ctx, configManager
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager) {
				skills := []*domain.Skill{
					{
						Name:      "git-skill",
						Source:    "git",
						URL:       "https://github.com/example/git-skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "hash1",
					},
					{
						Name:      "npm-skill",
						Source:    "npm",
						URL:       "example-npm-skill",
						Version:   "1.0.0",
						HashAlgo:  "sha256",
						HashValue: "hash2",
					},
				}

				for _, skill := range skills {
					if addErr := configManager.AddSkill(ctx, skill); addErr != nil {
						t.Fatalf("AddSkill failed for %s: %v", skill.Name, addErr)
					}
				}

				listedSkills, err := configManager.ListSkills(ctx)
				if err != nil {
					t.Fatalf("ListSkills failed: %v", err)
				}

				if len(listedSkills) != len(skills) {
					t.Errorf("Expected %d skills, got %d", len(skills), len(listedSkills))
				}

				updatedSkill := &domain.Skill{
					Name:      "git-skill",
					Source:    "git",
					URL:       "https://github.com/example/git-skill.git",
					Version:   "v2.0.0",
					HashAlgo:  "sha256",
					HashValue: "new-hash1",
				}

				err = configManager.UpdateSkill(ctx, updatedSkill)
				if err != nil {
					t.Fatalf("UpdateSkill failed: %v", err)
				}

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

				err = configManager.RemoveSkill(ctx, "npm-skill")
				if err != nil {
					t.Fatalf("RemoveSkill failed: %v", err)
				}

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
			},
		},
		{
			name: "Duplicate_Skill_Error",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)

				err := configManager.Initialize(ctx, []string{"~/.claude/skills"})
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}
				return ctx, configManager
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager) {
				skill := &domain.Skill{
					Name:      "duplicate-skill",
					Source:    "git",
					URL:       "https://github.com/example/skill.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "hash",
				}

				err := configManager.AddSkill(ctx, skill)
				if err != nil {
					t.Fatalf("First AddSkill failed: %v", err)
				}

				err = configManager.AddSkill(ctx, skill)
				if err == nil {
					t.Fatal("Expected error when adding duplicate skill, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, configManager := tt.setupFunc(t)
			tt.testFunc(t, ctx, configManager)
		})
	}
}
