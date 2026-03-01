package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// mockHashService is a mock implementation of port.HashService for testing
type mockHashService struct{}

func (m *mockHashService) CalculateHash(ctx context.Context, path string) (*port.HashResult, error) {
	return &port.HashResult{
		Value:     "mock-hash-value",
	}, nil
}


// mockPackageManager is a mock implementation of port.PackageManager for testing
type mockPackageManager struct {
	sourceType string
	tmpDir     string
}

func (m *mockPackageManager) SourceType() string {
	return m.sourceType
}

func (m *mockPackageManager) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	// Use the temporary directory provided by the test
	return &port.DownloadResult{
		Path:      m.tmpDir,
		Version:   version,
		FromGoMod: false,
	}, nil
}

func (m *mockPackageManager) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return "latest", nil
}

// setupTestConfig creates a test configuration with install targets
func setupTestConfig(t *testing.T) (configPath string, cleanup func()) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, ".skillspkg.toml")
	installDir := filepath.Join(tmpDir, "install")

	// Create initial config with install target
	cm := domain.NewConfigManager(configPath)
	if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	return configPath, func() {}
}

func TestAddCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrCheck func(error) bool
		setupFunc    func(t *testing.T) (configPath string, cleanup func())
		checkFunc    func(t *testing.T, configPath string)
		name         string
		skillName    string
		source       string
		url          string
		version      string
		subDir       string
		wantErr      bool
	}{
		{
			name:      "success: add git skill",
			skillName: "example-skill",
			source:    "git",
			url:       "https://github.com/example/skill.git",
			version:   "v1.0.0",
			setupFunc: setupTestConfig,
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
				if skill.SubDir != "skills/example-skill" {
					t.Errorf("expected SubDir 'skills/example-skill', got %s", skill.SubDir)
				}
			},
		},
		{
			name:      "success: add go-mod skill",
			skillName: "go-skill",
			source:    "go-mod",
			url:       "github.com/example/skill",
			version:   "v1.0.0",
			setupFunc: setupTestConfig,
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
				if skill.Source != "go-mod" {
					t.Errorf("expected source 'go-mod', got %s", skill.Source)
				}
			},
		},
		{
			name:      "error: config file not found",
			skillName: "test-skill",
			source:    "git",
			url:       "https://github.com/example/skill.git",
			version:   "v1.0.0",
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
			name:      "error: invalid source type",
			skillName: "test-skill",
			source:    "invalid-source",
			url:       "https://example.com",
			version:   "v1.0.0",
			setupFunc: setupTestConfig,
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorInvalidSource](err)
				return ok
			},
		},
		{
			name:      "error: duplicate skill name",
			skillName: "existing-skill",
			source:    "git",
			url:       "https://github.com/example/skill.git",
			version:   "v1.0.0",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "install")

				// Create initial config with existing skill
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
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
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorSkillExists](err)
				return ok
			},
		},
		{
			name:      "success: add skill with default subdirectory",
			skillName: "default-skill",
			source:    "go-mod",
			url:       "github.com/example/skills",
			version:   "v1.0.0",
			subDir:    "", // Empty means use default
			setupFunc: setupTestConfig,
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
				if skill.Name != "default-skill" {
					t.Errorf("expected name 'default-skill', got %s", skill.Name)
				}
				if skill.SubDir != "skills/default-skill" {
					t.Errorf("expected default SubDir 'skills/default-skill', got %s", skill.SubDir)
				}
			},
		},
		{
			name:      "success: add skill with custom subdirectory",
			skillName: "custom-skill",
			source:    "go-mod",
			url:       "github.com/example/monorepo",
			version:   "v1.0.0",
			subDir:    "packages/agents/custom-skill",
			setupFunc: setupTestConfig,
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
				if skill.Name != "custom-skill" {
					t.Errorf("expected name 'custom-skill', got %s", skill.Name)
				}
				if skill.SubDir != "packages/agents/custom-skill" {
					t.Errorf("expected custom SubDir 'packages/agents/custom-skill', got %s", skill.SubDir)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &AddCmd{
				Name:    tt.skillName,
				Source:  tt.source,
				URL:     tt.url,
				Version: tt.version,
				SubDir:  tt.subDir,
			}

			// Execute command directly using the internal run method for testing
			// Note: This will attempt to install the skill, which may fail without network access
			// Use mock dependencies to avoid network access
			tmpDir := t.TempDir() // Create a temporary directory for mock downloads

			// Create the expected SubDir structure
			subDir := tt.subDir
			if subDir == "" {
				subDir = "skills/" + tt.skillName
			}
			if err := os.MkdirAll(filepath.Join(tmpDir, subDir), 0o755); err != nil {
				t.Fatalf("failed to create subdirectory: %v", err)
			}

			hashService := &mockHashService{}
			packageManagers := []port.PackageManager{
				&mockPackageManager{sourceType: "git", tmpDir: tmpDir},
				&mockPackageManager{sourceType: "go-mod", tmpDir: tmpDir},
			}
			err := cmd.runWithDeps(configPath, false, hashService, packageManagers) // non-verbose mode

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

			// Run additional checks
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, configPath)
			}
		})
	}
}
