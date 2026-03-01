package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestInstallCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrCheck func(error) bool
		setupFunc    func(t *testing.T) (configPath string, cleanup func())
		name         string
		skills       []string
		wantErr      bool
	}{
		{
			name:   "error: config file not found (install all)",
			skills: []string{},
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				// Don't create config file
				return configPath, func() {}
			},
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorConfigNotFound](err)
				return ok
			},
		},
		{
			name:   "error: config file not found (install specific)",
			skills: []string{"test-skill"},
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				// Don't create config file
				return configPath, func() {}
			},
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorConfigNotFound](err)
				return ok
			},
		},
		{
			name:   "error: skill not found in configuration",
			skills: []string{"nonexistent-skill"},
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "skills")

				// Create initial config with no skills
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorSkillsNotFound](err)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &InstallCmd{
				Skills: tt.skills,
			}

			// Execute command directly using the internal run method for testing
			err := cmd.run(configPath, false) // non-verbose mode for testing

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.wantErrCheck != nil && !tt.wantErrCheck(err) {
					t.Errorf("expected error check to pass, got %v", err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
