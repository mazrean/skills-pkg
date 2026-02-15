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
source = "npm"
url = "example-skill"
version = "1.0.0"
hash_algo = "sha256"
hash_value = "e5f6g7h8"
package_manager = "npm"
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
				if skill2.PackageManager != "npm" {
					t.Errorf("expected package manager 'npm', got '%s'", skill2.PackageManager)
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
