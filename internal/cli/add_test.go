package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestAddCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		skillName       string
		source          string
		url             string
		version         string
		packageManager  string
		setupFunc       func(t *testing.T) (configPath string, cleanup func())
		wantErr         bool
		wantErrType     error
		checkFunc       func(t *testing.T, configPath string)
	}{
		{
			name:           "success: add git skill",
			skillName:      "example-skill",
			source:         "git",
			url:            "https://github.com/example/skill.git",
			version:        "v1.0.0",
			packageManager: "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load config: %v", err)
				}

				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill, got %d", len(config.Skills))
				}

				skill := config.Skills[0]
				if skill.Name != "example-skill" {
					t.Errorf("expected name 'example-skill', got %s", skill.Name)
				}
				if skill.Source != "git" {
					t.Errorf("expected source 'git', got %s", skill.Source)
				}
				if skill.URL != "https://github.com/example/skill.git" {
					t.Errorf("expected URL 'https://github.com/example/skill.git', got %s", skill.URL)
				}
				if skill.Version != "v1.0.0" {
					t.Errorf("expected version 'v1.0.0', got %s", skill.Version)
				}
				if skill.PackageManager != "" {
					t.Errorf("expected empty package manager for git, got %s", skill.PackageManager)
				}
			},
		},
		{
			name:           "success: add npm skill",
			skillName:      "npm-skill",
			source:         "npm",
			url:            "example-skill",
			version:        "1.0.0",
			packageManager: "npm",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load config: %v", err)
				}

				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill, got %d", len(config.Skills))
				}

				skill := config.Skills[0]
				if skill.Name != "npm-skill" {
					t.Errorf("expected name 'npm-skill', got %s", skill.Name)
				}
				if skill.Source != "npm" {
					t.Errorf("expected source 'npm', got %s", skill.Source)
				}
				if skill.PackageManager != "npm" {
					t.Errorf("expected package manager 'npm', got %s", skill.PackageManager)
				}
			},
		},
		{
			name:           "success: add go-module skill",
			skillName:      "go-skill",
			source:         "go-module",
			url:            "github.com/example/skill",
			version:        "v1.0.0",
			packageManager: "go-module",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load config: %v", err)
				}

				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill, got %d", len(config.Skills))
				}

				skill := config.Skills[0]
				if skill.Name != "go-skill" {
					t.Errorf("expected name 'go-skill', got %s", skill.Name)
				}
				if skill.Source != "go-module" {
					t.Errorf("expected source 'go-module', got %s", skill.Source)
				}
				if skill.PackageManager != "go-module" {
					t.Errorf("expected package manager 'go-module', got %s", skill.PackageManager)
				}
			},
		},
		{
			name:           "error: config file not found",
			skillName:      "test-skill",
			source:         "git",
			url:            "https://github.com/example/skill.git",
			version:        "v1.0.0",
			packageManager: "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				// Don't create config file
				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrConfigNotFound,
		},
		{
			name:           "error: invalid source type",
			skillName:      "test-skill",
			source:         "invalid-source",
			url:            "https://example.com",
			version:        "v1.0.0",
			packageManager: "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrInvalidSource,
		},
		{
			name:           "error: duplicate skill name",
			skillName:      "existing-skill",
			source:         "git",
			url:            "https://github.com/example/skill.git",
			version:        "v1.0.0",
			packageManager: "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config with existing skill
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add a skill with the same name
				skill := &domain.Skill{
					Name:    "existing-skill",
					Source:  "git",
					URL:     "https://github.com/other/skill.git",
					Version: "v0.1.0",
				}
				if err := cm.AddSkill(context.Background(), skill); err != nil {
					t.Fatalf("failed to add existing skill: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrSkillExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &AddCmd{
				Name:           tt.skillName,
				Source:         tt.source,
				URL:            tt.url,
				Version:        tt.version,
				PackageManager: tt.packageManager,
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
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Run additional checks
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, configPath)
			}
		})
	}
}
