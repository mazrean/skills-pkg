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
		name           string
		setupFile      bool // whether to create an existing config file
		installDirs    []string
		wantErr        error
		validateConfig func(t *testing.T, configPath string)
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
