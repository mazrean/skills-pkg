package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

// mockPackageManagerWithOptions extends mockPackageManager with error injection and FromGoMod support
type mockPackageManagerWithOptions struct {
	downloadErr error
	sourceType  string
	tmpDir      string
	fromGoMod   bool
}

func (m *mockPackageManagerWithOptions) SourceType() string { return m.sourceType }

func (m *mockPackageManagerWithOptions) Download(_ context.Context, _ *port.Source, version string) (*port.DownloadResult, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return &port.DownloadResult{
		Path:      m.tmpDir,
		Version:   version,
		FromGoMod: m.fromGoMod,
	}, nil
}

func (m *mockPackageManagerWithOptions) GetLatestVersion(_ context.Context, _ *port.Source) (string, error) {
	return "v0.1.0", nil
}

func TestInitCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrCheck     func(error) bool
		setupFunc        func(t *testing.T) (configPath string, cleanup func())
		setupInstallDirs func(t *testing.T) []string
		checkFunc        func(t *testing.T, configPath string)
		name             string
		agent            []string
		installDirs      []string
		global           bool
		wantErr          bool
	}{
		{
			name:        "success: initialize with default settings",
			installDirs: nil,
			agent:       nil,
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

				// managing-skills should be installed automatically
				if len(config.Skills) != 1 {
					t.Errorf("expected 1 skill (managing-skills), got %d skills", len(config.Skills))
				}
				if len(config.Skills) > 0 && config.Skills[0].Name != managingSkillsName {
					t.Errorf("expected skill name %q, got %q", managingSkillsName, config.Skills[0].Name)
				}

				// Default should now be ./.skills
				if len(config.InstallTargets) != 1 {
					t.Errorf("expected 1 install target (default ./.skills), got %d", len(config.InstallTargets))
				}

				if len(config.InstallTargets) > 0 && config.InstallTargets[0] != "./.skills" {
					t.Errorf("expected default install target ./.skills, got %s", config.InstallTargets[0])
				}
			},
		},
		{
			name:        "success: initialize with custom install directories",
			installDirs: nil, // set dynamically via setupInstallDirs
			agent:       nil,
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			setupInstallDirs: func(t *testing.T) []string {
				t.Helper()
				dir1 := filepath.Join(t.TempDir(), "custom-skills-1")
				dir2 := filepath.Join(t.TempDir(), "custom-skills-2")
				return []string{dir1, dir2}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				cm := domain.NewConfigManager(configPath)
				config, err := cm.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load created config: %v", err)
				}

				if len(config.InstallTargets) != 2 {
					t.Errorf("expected 2 install targets, got %d", len(config.InstallTargets))
				}
			},
		},
		{
			name:        "success: initialize with agent flag (project-level)",
			installDirs: nil,
			agent:       []string{"claude"},
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

				// Should contain the project-level claude agent directory
				expectedDir := ".claude/skills"
				if len(config.InstallTargets) > 0 && config.InstallTargets[0] != expectedDir {
					t.Errorf("expected install target %s, got %s", expectedDir, config.InstallTargets[0])
				}
			},
		},
		{
			name:        "success: initialize with both custom dirs and agent",
			installDirs: nil, // set dynamically via setupInstallDirs
			agent:       []string{"claude"},
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			setupInstallDirs: func(t *testing.T) []string {
				t.Helper()
				return []string{filepath.Join(t.TempDir(), "custom-skills")}
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
				if len(config.InstallTargets) < 2 {
					t.Errorf("expected at least 2 install targets, got %d", len(config.InstallTargets))
				}
			},
		},
		{
			name:        "success: initialize with multiple agents (project-level)",
			installDirs: nil,
			agent:       []string{"claude", "codex"},
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

				if len(config.InstallTargets) != 2 {
					t.Errorf("expected 2 install targets, got %d", len(config.InstallTargets))
				}

				// Should contain both project-level agent directories
				expectedDirs := []string{".claude/skills", ".agents/skills"}
				for i, expectedDir := range expectedDirs {
					if i >= len(config.InstallTargets) {
						t.Errorf("missing install target at index %d", i)
						continue
					}
					if config.InstallTargets[i] != expectedDir {
						t.Errorf("install target[%d]: expected %s, got %s", i, expectedDir, config.InstallTargets[i])
					}
				}
			},
		},
		{
			name:        "success: initialize with codex agent flag (project-level)",
			installDirs: nil,
			agent:       []string{"codex"},
			global:      false,
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

				// Should contain the project-level codex agent directory
				expectedDir := ".agents/skills"
				if len(config.InstallTargets) > 0 && config.InstallTargets[0] != expectedDir {
					t.Errorf("expected install target %s, got %s", expectedDir, config.InstallTargets[0])
				}
			},
		},
		{
			name:        "success: initialize with agent flag and global (user-level)",
			installDirs: nil,
			agent:       []string{"claude"},
			global:      true,
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

				// Should contain the user-level claude agent directory
				// The exact path will be resolved by ClaudeAgentAdapter (~/.claude/skills)
				if len(config.InstallTargets) > 0 && config.InstallTargets[0] == "" {
					t.Error("install target should not be empty")
				}

				// User-level directory should be an absolute path
				if len(config.InstallTargets) > 0 && !filepath.IsAbs(config.InstallTargets[0]) {
					t.Errorf("user-level directory should be absolute path, got %s", config.InstallTargets[0])
				}
			},
		},
		{
			name:        "error: config file already exists",
			installDirs: nil,
			agent:       nil,
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
			wantErr: true,
			wantErrCheck: func(err error) bool {
				_, ok := errors.AsType[*domain.ErrorConfigExists](err)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			installDirs := tt.installDirs
			if tt.setupInstallDirs != nil {
				installDirs = tt.setupInstallDirs(t)
			}

			cmd := &InitCmd{
				InstallDir: installDirs,
				Agent:      tt.agent,
				Global:     tt.global,
			}

			// Create mock download directory with managing-skills subdirectory
			mockDownloadDir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(mockDownloadDir, managingSkillsSubDir), 0o755); err != nil {
				t.Fatalf("failed to create mock managing-skills directory: %v", err)
			}

			hashService := &mockHashService{}
			packageManagers := []port.PackageManager{
				&mockPackageManager{sourceType: "git", tmpDir: mockDownloadDir},
				&mockPackageManager{sourceType: "go-mod", tmpDir: mockDownloadDir},
			}

			// Execute command using runWithDeps with mock dependencies for testing
			err := cmd.runWithDeps(configPath, false, hashService, packageManagers)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.wantErrCheck != nil && !tt.wantErrCheck(err) {
					t.Errorf("expected error check failed, got %v", err)
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

func TestInitCmd_RollbackOnInstallFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".skillspkg.toml")

	cmd := &InitCmd{}

	hashService := &mockHashService{}
	packageManagers := []port.PackageManager{
		&mockPackageManagerWithOptions{
			sourceType:  "go-mod",
			downloadErr: fmt.Errorf("network error"),
		},
	}

	err := cmd.runWithDeps(configPath, false, hashService, packageManagers)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Config file must not exist after rollback so that re-running init is possible
	if _, statErr := os.Stat(configPath); !os.IsNotExist(statErr) {
		t.Errorf("config file should have been removed on install failure, but it exists at %s", configPath)
	}
}

func TestInitCmd_ManagingSkillsFromGoMod(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".skillspkg.toml")
	installDir := filepath.Join(tmpDir, "install")

	// Prepare mock download directory
	mockDownloadDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(mockDownloadDir, managingSkillsSubDir), 0o755); err != nil {
		t.Fatalf("failed to create mock managing-skills directory: %v", err)
	}

	cmd := &InitCmd{InstallDir: []string{installDir}}

	hashService := &mockHashService{}
	// Simulate version resolved from go.mod (FromGoMod=true)
	packageManagers := []port.PackageManager{
		&mockPackageManagerWithOptions{
			sourceType: "go-mod",
			tmpDir:     mockDownloadDir,
			fromGoMod:  true,
		},
	}

	if err := cmd.runWithDeps(configPath, false, hashService, packageManagers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm := domain.NewConfigManager(configPath)
	config, err := cm.Load(context.Background())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(config.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(config.Skills))
	}

	skill := config.Skills[0]
	// When version is resolved from go.mod, Version and HashValue must be empty
	// so that go.mod/go.sum remains the source of truth
	if skill.Version != "" {
		t.Errorf("expected empty version (go.mod is source of truth), got %q", skill.Version)
	}
	if skill.HashValue != "" {
		t.Errorf("expected empty hash value (go.mod is source of truth), got %q", skill.HashValue)
	}
}
