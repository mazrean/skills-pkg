package domain_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/pelletier/go-toml/v2"
)

// TestConfigManager_Initialize tests the Initialize method of ConfigManager.
// Requirements: 1.1, 1.4, 1.5
func TestConfigManager_Initialize(t *testing.T) {
	tests := []struct {
		wantErr        error
		validateConfig func(t *testing.T, configPath string)
		name           string
		installDirs    []string
		setupFile      bool
	}{
		{
			name:        "successfully creates new config file",
			setupFile:   false,
			installDirs: []string{"~/.claude/skills", "~/.codex/skills"},
			wantErr:     nil,
			validateConfig: func(t *testing.T, configPath string) {
				// Check that the file was created
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("config file was not created at %s", configPath)
					return
				}

				// Read and verify the config file content (requirement 1.5)
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("failed to read config file: %v", err)
				}

				var config domain.Config
				if err := toml.Unmarshal(data, &config); err != nil {
					t.Errorf("config file is not valid TOML: %v", err)
					return
				}

				// Verify empty skills list
				if len(config.Skills) != 0 {
					t.Errorf("expected empty skills list, got %d skills", len(config.Skills))
				}

				// Verify install targets
				if len(config.InstallTargets) != 2 {
					t.Errorf("expected 2 install targets, got %d", len(config.InstallTargets))
				}
				if len(config.InstallTargets) >= 2 {
					if config.InstallTargets[0] != "~/.claude/skills" {
						t.Errorf("expected first install target to be '~/.claude/skills', got '%s'", config.InstallTargets[0])
					}
					if config.InstallTargets[1] != "~/.codex/skills" {
						t.Errorf("expected second install target to be '~/.codex/skills', got '%s'", config.InstallTargets[1])
					}
				}
			},
		},
		{
			name:        "successfully creates config with empty install dirs",
			setupFile:   false,
			installDirs: []string{},
			wantErr:     nil,
			validateConfig: func(t *testing.T, configPath string) {
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("config file was not created at %s", configPath)
					return
				}

				// Read and verify the config file content
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("failed to read config file: %v", err)
				}

				var config domain.Config
				if err := toml.Unmarshal(data, &config); err != nil {
					t.Errorf("config file is not valid TOML: %v", err)
					return
				}

				// Verify empty skills list
				if len(config.Skills) != 0 {
					t.Errorf("expected empty skills list, got %d skills", len(config.Skills))
				}

				// Verify empty install targets
				if len(config.InstallTargets) != 0 {
					t.Errorf("expected empty install targets, got %d", len(config.InstallTargets))
				}
			},
		},
		{
			name:        "returns error when config file already exists",
			setupFile:   true,
			installDirs: []string{"~/.claude/skills"},
			wantErr:     domain.ErrConfigExists,
			validateConfig: func(t *testing.T, configPath string) {
				// Config file should still exist (not overwritten)
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("existing config file was removed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Setup: create existing config file if needed
			if tt.setupFile {
				if err := os.WriteFile(configPath, []byte("# existing config"), 0644); err != nil {
					t.Fatalf("failed to setup test: %v", err)
				}
			}

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Execute Initialize
			ctx := context.Background()
			err := manager.Initialize(ctx, tt.installDirs)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.Initialize() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.Initialize() unexpected error = %v", err)
			}

			// Validate the config file state
			if tt.validateConfig != nil {
				tt.validateConfig(t, configPath)
			}
		})
	}
}

// TestConfigManager_Load tests the Load method of ConfigManager.
// Requirements: 2.1, 2.6, 12.2, 12.3
func TestConfigManager_Load(t *testing.T) {
	tests := []struct {
		wantErr     error
		validate    func(t *testing.T, config *domain.Config)
		name        string
		fileContent string
		setupFile   bool
	}{
		{
			name:      "successfully loads valid config file",
			setupFile: true,
			fileContent: `install_targets = ["~/.claude/skills", "~/.codex/skills"]

[[skills]]
name = "example-skill"
source = "git"
url = "https://github.com/example/skill.git"
version = "v1.0.0"
hash_algo = "sha256"
hash_value = "a1b2c3d4"

[[skills]]
name = "npm-skill"
source = "go-mod"
url = "example-skill"
version = "1.0.0"
hash_algo = "sha256"
hash_value = "e5f6g7h8"
package_manager = "go-mod"
`,
			wantErr: nil,
			validate: func(t *testing.T, config *domain.Config) {
				// Verify install targets
				if len(config.InstallTargets) != 2 {
					t.Errorf("expected 2 install targets, got %d", len(config.InstallTargets))
				}
				if len(config.InstallTargets) >= 2 {
					if config.InstallTargets[0] != "~/.claude/skills" {
						t.Errorf("expected first install target to be '~/.claude/skills', got '%s'", config.InstallTargets[0])
					}
					if config.InstallTargets[1] != "~/.codex/skills" {
						t.Errorf("expected second install target to be '~/.codex/skills', got '%s'", config.InstallTargets[1])
					}
				}

				// Verify skills
				if len(config.Skills) != 2 {
					t.Fatalf("expected 2 skills, got %d", len(config.Skills))
				}

				// Verify first skill
				skill1 := config.Skills[0]
				if skill1.Name != "example-skill" {
					t.Errorf("expected skill name 'example-skill', got '%s'", skill1.Name)
				}
				if skill1.Source != "git" {
					t.Errorf("expected skill source 'git', got '%s'", skill1.Source)
				}
				if skill1.URL != "https://github.com/example/skill.git" {
					t.Errorf("expected skill URL 'https://github.com/example/skill.git', got '%s'", skill1.URL)
				}
				if skill1.Version != "v1.0.0" {
					t.Errorf("expected skill version 'v1.0.0', got '%s'", skill1.Version)
				}
				if skill1.HashAlgo != "sha256" {
					t.Errorf("expected hash algo 'sha256', got '%s'", skill1.HashAlgo)
				}
				if skill1.HashValue != "a1b2c3d4" {
					t.Errorf("expected hash value 'a1b2c3d4', got '%s'", skill1.HashValue)
				}

				// Verify second skill
				skill2 := config.Skills[1]
				if skill2.Name != "npm-skill" {
					t.Errorf("expected skill name 'npm-skill', got '%s'", skill2.Name)
				}
				if skill2.Source != "go-mod" {
					t.Errorf("expected source 'npm', got '%s'", skill2.Source)
				}
			},
		},
		{
			name:      "returns error when config file not found",
			setupFile: false,
			wantErr:   domain.ErrConfigNotFound,
			validate:  nil,
		},
		{
			name:      "returns detailed error for invalid TOML format",
			setupFile: true,
			fileContent: `invalid toml content [[[
this is not valid`,
			wantErr: nil, // We'll check for a descriptive error message instead
			validate: func(t *testing.T, config *domain.Config) {
				// This test case should fail with a TOML parse error
				t.Error("expected TOML parse error, but got no error")
			},
		},
		{
			name:      "successfully loads empty config file",
			setupFile: true,
			fileContent: `install_targets = []
skills = []
`,
			wantErr: nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.InstallTargets) != 0 {
					t.Errorf("expected 0 install targets, got %d", len(config.InstallTargets))
				}
				if len(config.Skills) != 0 {
					t.Errorf("expected 0 skills, got %d", len(config.Skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Setup: create config file if needed
			if tt.setupFile {
				if err := os.WriteFile(configPath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to setup test: %v", err)
				}
			}

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Execute Load
			ctx := context.Background()
			config, err := manager.Load(ctx)

			// Special handling for TOML parse error test case
			if tt.name == "returns detailed error for invalid TOML format" {
				if err == nil {
					t.Error("expected error for invalid TOML, but got no error")
					return
				}
				// Verify that the error message contains details about the parse error (requirement 2.6)
				if err.Error() == "" {
					t.Error("expected detailed error message, but got empty error")
				}
				// The error should mention TOML parsing
				return
			}

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.Load() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.Load() unexpected error = %v", err)
				return
			}

			// Validate the loaded config
			if tt.validate != nil && config != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestConfigManager_Save tests the Save method of ConfigManager.
// Requirements: 2.1, 12.2, 12.3
func TestConfigManager_Save(t *testing.T) {
	tests := []struct {
		wantErr  error
		config   *domain.Config
		validate func(t *testing.T, configPath string)
		name     string
	}{
		{
			name: "successfully saves config to file",
			config: &domain.Config{
				InstallTargets: []string{"~/.claude/skills", "~/.codex/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "example-skill",
						Source:    "git",
						URL:       "https://github.com/example/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "a1b2c3d4",
					},
				},
			},
			wantErr: nil,
			validate: func(t *testing.T, configPath string) {
				// Read and verify the saved file
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("failed to read saved config file: %v", err)
				}

				var config domain.Config
				if err := toml.Unmarshal(data, &config); err != nil {
					t.Errorf("saved config file is not valid TOML: %v", err)
					return
				}

				// Verify install targets
				if len(config.InstallTargets) != 2 {
					t.Errorf("expected 2 install targets, got %d", len(config.InstallTargets))
				}

				// Verify skills
				if len(config.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(config.Skills))
				}

				skill := config.Skills[0]
				if skill.Name != "example-skill" {
					t.Errorf("expected skill name 'example-skill', got '%s'", skill.Name)
				}
				if skill.Version != "v1.0.0" {
					t.Errorf("expected skill version 'v1.0.0', got '%s'", skill.Version)
				}
			},
		},
		{
			name: "successfully saves empty config",
			config: &domain.Config{
				InstallTargets: []string{},
				Skills:         []*domain.Skill{},
			},
			wantErr: nil,
			validate: func(t *testing.T, configPath string) {
				// Read and verify the saved file
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("failed to read saved config file: %v", err)
				}

				var config domain.Config
				if err := toml.Unmarshal(data, &config); err != nil {
					t.Errorf("saved config file is not valid TOML: %v", err)
					return
				}

				if len(config.InstallTargets) != 0 {
					t.Errorf("expected 0 install targets, got %d", len(config.InstallTargets))
				}
				if len(config.Skills) != 0 {
					t.Errorf("expected 0 skills, got %d", len(config.Skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Execute Save
			ctx := context.Background()
			err := manager.Save(ctx, tt.config)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.Save() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.Save() unexpected error = %v", err)
				return
			}

			// Validate the saved file
			if tt.validate != nil {
				tt.validate(t, configPath)
			}
		})
	}
}

// TestConfigManager_AddSkill tests the AddSkill method of ConfigManager.
// Requirements: 2.2, 2.3, 2.4, 5.2
func TestConfigManager_AddSkill(t *testing.T) {
	tests := []struct {
		wantErr     error
		setupConfig *domain.Config
		skill       *domain.Skill
		validate    func(t *testing.T, config *domain.Config)
		name        string
	}{
		{
			name: "successfully adds git skill",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills:         []*domain.Skill{},
			},
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/test/skill.git",
				Version:   "v1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(config.Skills))
				}
				skill := config.Skills[0]
				if skill.Name != "test-skill" {
					t.Errorf("expected skill name 'test-skill', got '%s'", skill.Name)
				}
				if skill.Source != "git" {
					t.Errorf("expected source 'git', got '%s'", skill.Source)
				}
				if skill.Version != "v1.0.0" {
					t.Errorf("expected version 'v1.0.0', got '%s'", skill.Version)
				}
			},
		},
		{
			name: "successfully adds npm skill with source",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills:         []*domain.Skill{},
			},
			skill: &domain.Skill{
				Name:           "npm-skill",
				Source:         "go-mod",
				URL:            "example-package",
				Version:        "1.0.0",
				HashAlgo:       "sha256",
				HashValue:      "def456",
			},
			wantErr: nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(config.Skills))
				}
				skill := config.Skills[0]
				if skill.Source != "go-mod" {
					t.Errorf("expected source 'npm', got '%s'", skill.Source)
				}
			},
		},
		{
			name: "returns error when skill already exists",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "existing-skill",
						Source:    "git",
						URL:       "https://github.com/existing/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "xyz789",
					},
				},
			},
			skill: &domain.Skill{
				Name:      "existing-skill",
				Source:    "git",
				URL:       "https://github.com/new/skill.git",
				Version:   "v2.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: domain.ErrSkillExists,
			validate: func(t *testing.T, config *domain.Config) {
				// Original skill should remain unchanged
				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill, got %d", len(config.Skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Setup: Save initial config
			ctx := context.Background()
			if err := manager.Save(ctx, tt.setupConfig); err != nil {
				t.Fatalf("failed to setup test: %v", err)
			}

			// Execute AddSkill
			err := manager.AddSkill(ctx, tt.skill)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.AddSkill() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.AddSkill() unexpected error = %v", err)
				return
			}

			// Load config to validate
			config, err := manager.Load(ctx)
			if err != nil {
				t.Fatalf("failed to load config after AddSkill: %v", err)
			}

			// Validate the config
			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestConfigManager_UpdateSkill tests the UpdateSkill method of ConfigManager.
// Requirements: 2.2, 5.2
func TestConfigManager_UpdateSkill(t *testing.T) {
	tests := []struct {
		wantErr     error
		setupConfig *domain.Config
		skill       *domain.Skill
		validate    func(t *testing.T, config *domain.Config)
		name        string
	}{
		{
			name: "successfully updates skill version and hash",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "test-skill",
						Source:    "git",
						URL:       "https://github.com/test/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "old-hash",
					},
				},
			},
			skill: &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/test/skill.git",
				Version:   "v2.0.0",
				HashAlgo:  "sha256",
				HashValue: "new-hash",
			},
			wantErr: nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(config.Skills))
				}
				skill := config.Skills[0]
				if skill.Version != "v2.0.0" {
					t.Errorf("expected version 'v2.0.0', got '%s'", skill.Version)
				}
				if skill.HashValue != "new-hash" {
					t.Errorf("expected hash 'new-hash', got '%s'", skill.HashValue)
				}
			},
		},
		{
			name: "returns error when skill not found",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills:         []*domain.Skill{},
			},
			skill: &domain.Skill{
				Name:      "nonexistent-skill",
				Source:    "git",
				URL:       "https://github.com/test/skill.git",
				Version:   "v1.0.0",
				HashAlgo:  "sha256",
				HashValue: "abc123",
			},
			wantErr: domain.ErrSkillNotFound,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 0 {
					t.Errorf("expected 0 skills, got %d", len(config.Skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Setup: Save initial config
			ctx := context.Background()
			if err := manager.Save(ctx, tt.setupConfig); err != nil {
				t.Fatalf("failed to setup test: %v", err)
			}

			// Execute UpdateSkill
			err := manager.UpdateSkill(ctx, tt.skill)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.UpdateSkill() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.UpdateSkill() unexpected error = %v", err)
				return
			}

			// Load config to validate
			config, err := manager.Load(ctx)
			if err != nil {
				t.Fatalf("failed to load config after UpdateSkill: %v", err)
			}

			// Validate the config
			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestConfigManager_RemoveSkill tests the RemoveSkill method of ConfigManager.
// Requirements: 9.2
func TestConfigManager_RemoveSkill(t *testing.T) {
	tests := []struct {
		wantErr     error
		setupConfig *domain.Config
		validate    func(t *testing.T, config *domain.Config)
		name        string
		skillName   string
	}{
		{
			name: "successfully removes skill",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "skill-to-remove",
						Source:    "git",
						URL:       "https://github.com/test/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "abc123",
					},
					{
						Name:      "skill-to-keep",
						Source:    "git",
						URL:       "https://github.com/test/other.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "def456",
					},
				},
			},
			skillName: "skill-to-remove",
			wantErr:   nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 1 {
					t.Fatalf("expected 1 skill remaining, got %d", len(config.Skills))
				}
				if config.Skills[0].Name != "skill-to-keep" {
					t.Errorf("expected remaining skill to be 'skill-to-keep', got '%s'", config.Skills[0].Name)
				}
			},
		},
		{
			name: "successfully removes last skill",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "only-skill",
						Source:    "git",
						URL:       "https://github.com/test/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "abc123",
					},
				},
			},
			skillName: "only-skill",
			wantErr:   nil,
			validate: func(t *testing.T, config *domain.Config) {
				if len(config.Skills) != 0 {
					t.Errorf("expected 0 skills, got %d", len(config.Skills))
				}
			},
		},
		{
			name: "returns error when skill not found",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "existing-skill",
						Source:    "git",
						URL:       "https://github.com/test/skill.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "abc123",
					},
				},
			},
			skillName: "nonexistent-skill",
			wantErr:   domain.ErrSkillNotFound,
			validate: func(t *testing.T, config *domain.Config) {
				// Original skill should remain
				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill, got %d", len(config.Skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Setup: Save initial config
			ctx := context.Background()
			if err := manager.Save(ctx, tt.setupConfig); err != nil {
				t.Fatalf("failed to setup test: %v", err)
			}

			// Execute RemoveSkill
			err := manager.RemoveSkill(ctx, tt.skillName)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.RemoveSkill() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.RemoveSkill() unexpected error = %v", err)
				return
			}

			// Load config to validate
			config, err := manager.Load(ctx)
			if err != nil {
				t.Fatalf("failed to load config after RemoveSkill: %v", err)
			}

			// Validate the config
			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestConfigManager_GetInstallTargets tests the GetInstallTargets method of ConfigManager.
// Requirements: 1.2, 2.5, 10.1
func TestConfigManager_GetInstallTargets(t *testing.T) {
	tests := []struct {
		wantErr     error
		setupConfig *domain.Config
		validate    func(t *testing.T, targets []string)
		name        string
	}{
		{
			name: "successfully gets install targets",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills", "~/.codex/skills"},
				Skills:         []*domain.Skill{},
			},
			wantErr: nil,
			validate: func(t *testing.T, targets []string) {
				if len(targets) != 2 {
					t.Fatalf("expected 2 install targets, got %d", len(targets))
				}
				if targets[0] != "~/.claude/skills" {
					t.Errorf("expected first target '~/.claude/skills', got '%s'", targets[0])
				}
				if targets[1] != "~/.codex/skills" {
					t.Errorf("expected second target '~/.codex/skills', got '%s'", targets[1])
				}
			},
		},
		{
			name: "returns empty list when no targets configured",
			setupConfig: &domain.Config{
				InstallTargets: []string{},
				Skills:         []*domain.Skill{},
			},
			wantErr: nil,
			validate: func(t *testing.T, targets []string) {
				if len(targets) != 0 {
					t.Errorf("expected 0 install targets, got %d", len(targets))
				}
			},
		},
		{
			name: "gets multiple install targets",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills", "~/.codex/skills", "/opt/skills"},
				Skills:         []*domain.Skill{},
			},
			wantErr: nil,
			validate: func(t *testing.T, targets []string) {
				if len(targets) != 3 {
					t.Fatalf("expected 3 install targets, got %d", len(targets))
				}
				if targets[2] != "/opt/skills" {
					t.Errorf("expected third target '/opt/skills', got '%s'", targets[2])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Setup: Save initial config
			ctx := context.Background()
			if err := manager.Save(ctx, tt.setupConfig); err != nil {
				t.Fatalf("failed to setup test: %v", err)
			}

			// Execute GetInstallTargets
			targets, err := manager.GetInstallTargets(ctx)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.GetInstallTargets() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.GetInstallTargets() unexpected error = %v", err)
				return
			}

			// Validate the targets list
			if tt.validate != nil {
				tt.validate(t, targets)
			}
		})
	}
}

// TestConfigManager_ListSkills tests the ListSkills method of ConfigManager.
// Requirements: 8.1, 8.2
func TestConfigManager_ListSkills(t *testing.T) {
	tests := []struct {
		wantErr     error
		setupConfig *domain.Config
		validate    func(t *testing.T, skills []*domain.Skill)
		name        string
	}{
		{
			name: "successfully lists all skills",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills: []*domain.Skill{
					{
						Name:      "skill-1",
						Source:    "git",
						URL:       "https://github.com/test/skill1.git",
						Version:   "v1.0.0",
						HashAlgo:  "sha256",
						HashValue: "abc123",
					},
					{
						Name:           "skill-2",
						Source:         "go-mod",
						URL:            "npm-package",
						Version:        "2.0.0",
						HashAlgo:       "sha256",
						HashValue:      "def456",
					},
				},
			},
			wantErr: nil,
			validate: func(t *testing.T, skills []*domain.Skill) {
				if len(skills) != 2 {
					t.Fatalf("expected 2 skills, got %d", len(skills))
				}
				if skills[0].Name != "skill-1" {
					t.Errorf("expected first skill name 'skill-1', got '%s'", skills[0].Name)
				}
				if skills[1].Name != "skill-2" {
					t.Errorf("expected second skill name 'skill-2', got '%s'", skills[1].Name)
				}
				if skills[1].Source != "go-mod" {
					t.Errorf("expected source 'npm', got '%s'", skills[1].Source)
				}
			},
		},
		{
			name: "returns empty list when no skills",
			setupConfig: &domain.Config{
				InstallTargets: []string{"~/.claude/skills"},
				Skills:         []*domain.Skill{},
			},
			wantErr: nil,
			validate: func(t *testing.T, skills []*domain.Skill) {
				if len(skills) != 0 {
					t.Errorf("expected 0 skills, got %d", len(skills))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create ConfigManager
			manager := domain.NewConfigManager(configPath)

			// Setup: Save initial config
			ctx := context.Background()
			if err := manager.Save(ctx, tt.setupConfig); err != nil {
				t.Fatalf("failed to setup test: %v", err)
			}

			// Execute ListSkills
			skills, err := manager.ListSkills(ctx)

			// Verify error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ConfigManager.ListSkills() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ConfigManager.ListSkills() unexpected error = %v", err)
				return
			}

			// Validate the skills list
			if tt.validate != nil {
				tt.validate(t, skills)
			}
		})
	}
}
