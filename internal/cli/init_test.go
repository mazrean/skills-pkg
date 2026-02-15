package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestInitCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrType error
		setupFunc   func(t *testing.T) (configPath string, cleanup func())
		checkFunc   func(t *testing.T, configPath string)
		name        string
		agent       string
		installDirs []string
		wantErr     bool
	}{
		{
			name:        "success: initialize with default settings",
			installDirs: nil,
			agent:       "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				// Verify config file was created
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("config file was not created at %s", configPath)
				}

				// Verify config contents
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load created config: %v", err)
				}

				if len(config.Skills) != 0 {
					t.Errorf("expected empty skills, got %d skills", len(config.Skills))
				}

				if len(config.InstallTargets) != 0 {
					t.Errorf("expected empty install targets, got %v", config.InstallTargets)
				}
			},
		},
		{
			name:        "success: initialize with custom install directories",
			installDirs: []string{"~/.custom/skills", "/opt/skills"},
			agent:       "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load created config: %v", err)
				}

				expectedDirs := []string{"~/.custom/skills", "/opt/skills"}
				if len(config.InstallTargets) != len(expectedDirs) {
					t.Errorf("expected %d install targets, got %d", len(expectedDirs), len(config.InstallTargets))
				}

				for i, expected := range expectedDirs {
					if config.InstallTargets[i] != expected {
						t.Errorf("install target[%d]: expected %s, got %s", i, expected, config.InstallTargets[i])
					}
				}
			},
		},
		{
			name:        "success: initialize with agent flag",
			installDirs: nil,
			agent:       "claude",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load created config: %v", err)
				}

				if len(config.InstallTargets) != 1 {
					t.Errorf("expected 1 install target, got %d", len(config.InstallTargets))
				}

				// Should contain the claude agent directory
				// The exact path will be resolved by ClaudeAgentAdapter
				if len(config.InstallTargets) > 0 && config.InstallTargets[0] == "" {
					t.Error("install target should not be empty")
				}
			},
		},
		{
			name:        "success: initialize with both custom dirs and agent",
			installDirs: []string{"/custom/path"},
			agent:       "claude",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load created config: %v", err)
				}

				// Should have both custom dir and agent dir
				if len(config.InstallTargets) < 1 {
					t.Errorf("expected at least 1 install target, got %d", len(config.InstallTargets))
				}

				// First should be the custom dir
				if config.InstallTargets[0] != "/custom/path" {
					t.Errorf("first install target: expected /custom/path, got %s", config.InstallTargets[0])
				}
			},
		},
		{
			name:        "error: config file already exists",
			installDirs: nil,
			agent:       "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create existing config file
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to create existing config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrConfigExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &InitCmd{
				InstallDir: tt.installDirs,
				Agent:      tt.agent,
			}

			// Execute command directly using the internal run method for testing
			err := cmd.run(configPath, false) // non-verbose mode for testing

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
					t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Run additional checks
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, configPath)
			}
		})
	}
}
