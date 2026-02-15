package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestUninstallCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setupFunc func(t *testing.T) (configPath string, cleanup func())
		checkFunc func(t *testing.T, configPath string)
		name      string
		skillName string
		wantErr   bool
	}{
		{
			name:      "success: basic uninstall functionality",
			skillName: "test-skill",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir := filepath.Join(tempDir, "skills")

				// Initialize config with a test skill
				configManager := domain.NewConfigManager(configPath)
				if err := configManager.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add a test skill
				skill := &domain.Skill{
					Name:      "test-skill",
					Source:    "git",
					URL:       "https://example.com/test.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "abc123",
				}
				if err := configManager.AddSkill(context.Background(), skill); err != nil {
					t.Fatalf("failed to add test skill: %v", err)
				}

				// Create skill directory to simulate installed skill
				skillDir := filepath.Join(installDir, "test-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				testFile := filepath.Join(skillDir, "test.txt")
				if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				tempDir := filepath.Dir(configPath)
				installDir := filepath.Join(tempDir, "skills")
				skillDir := filepath.Join(installDir, "test-skill")

				// Verify skill directory is removed
				if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
					t.Errorf("skill directory still exists after uninstall")
				}

				// Verify skill is removed from config
				configManager := domain.NewConfigManager(configPath)
				config, err := configManager.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load config: %v", err)
				}
				if config.FindSkillByName("test-skill") != nil {
					t.Errorf("skill still exists in configuration after uninstall")
				}
			},
		},
		{
			name:      "error: non-existent skill",
			skillName: "non-existent-skill",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir := filepath.Join(tempDir, "skills")

				// Initialize config without any skills
				configManager := domain.NewConfigManager(configPath)
				if err := configManager.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: true,
		},
		{
			name:      "error: config file not found",
			skillName: "test-skill",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "non-existent", ".skillspkg.toml")
				// Don't create config file
				return configPath, func() {}
			},
			wantErr: true,
		},
		{
			name:      "success: uninstall from multiple install directories",
			skillName: "test-skill",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir1 := filepath.Join(tempDir, "skills1")
				installDir2 := filepath.Join(tempDir, "skills2")

				// Initialize config with multiple install targets
				configManager := domain.NewConfigManager(configPath)
				if err := configManager.Initialize(context.Background(), []string{installDir1, installDir2}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add a test skill
				skill := &domain.Skill{
					Name:      "test-skill",
					Source:    "git",
					URL:       "https://example.com/test.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "abc123",
				}
				if err := configManager.AddSkill(context.Background(), skill); err != nil {
					t.Fatalf("failed to add test skill: %v", err)
				}

				// Create skill directories in both targets
				for _, installDir := range []string{installDir1, installDir2} {
					skillDir := filepath.Join(installDir, "test-skill")
					if err := os.MkdirAll(skillDir, 0o755); err != nil {
						t.Fatalf("failed to create skill directory in %s: %v", installDir, err)
					}
					testFile := filepath.Join(skillDir, "test.txt")
					if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
						t.Fatalf("failed to write test file in %s: %v", installDir, err)
					}
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, configPath string) {
				t.Helper()
				tempDir := filepath.Dir(configPath)
				installDir1 := filepath.Join(tempDir, "skills1")
				installDir2 := filepath.Join(tempDir, "skills2")

				// Verify skill directories are removed from both targets
				for _, installDir := range []string{installDir1, installDir2} {
					skillDir := filepath.Join(installDir, "test-skill")
					if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
						t.Errorf("skill directory still exists in %s after uninstall", installDir)
					}
				}

				// Verify skill is removed from config
				configManager := domain.NewConfigManager(configPath)
				config, err := configManager.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load config: %v", err)
				}
				if config.FindSkillByName("test-skill") != nil {
					t.Errorf("skill still exists in configuration after uninstall")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &UninstallCmd{
				SkillName: tt.skillName,
			}

			// Execute command directly using the internal run method for testing
			err := cmd.run(configPath, false) // non-verbose mode for testing

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
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
