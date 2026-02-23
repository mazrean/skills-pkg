package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestUpdateCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrType error
		setupFunc   func(t *testing.T) (configPath string, cleanup func())
		name        string
		skills      []string
		source      string
		wantErr     bool
	}{
		{
			name:   "error: config file not found (update all)",
			skills: []string{},
			source: "",
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
			name:   "error: config file not found (update specific)",
			skills: []string{"test-skill"},
			source: "",
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
			// After task 5.3: run() collects all errors and returns an aggregate error
			// (no longer returns ErrSkillNotFound directly)
			name:   "error: skill not found in configuration returns aggregate error",
			skills: []string{"nonexistent-skill"},
			source: "",
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
			wantErr:     true,
			wantErrType: nil, // aggregate error, not ErrSkillNotFound directly
		},
		{
			name:   "source field: config not found when source is git",
			skills: []string{},
			source: "git",
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
			name:   "error: invalid source type",
			skills: []string{},
			source: "invalid-source",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrInvalidSource,
		},
		{
			// source="go-mod" と config なし → ErrConfigNotFound (git と対称)
			name:   "source field: config not found when source is go-mod",
			skills: []string{},
			source: "go-mod",
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
			// スキルが0件の設定でソースフィルタなし → 0件更新で正常終了
			name:   "success: empty config (update all with no skills)",
			skills: []string{},
			source: "",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "skills")

				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			// スキルが0件 かつ --source git → "No skills found for source 'git'." を出力し正常終了 (req 2.2)
			name:   "success: no skills found for source git (empty config)",
			skills: []string{},
			source: "git",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "skills")

				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			// go-mod スキルのみの設定で --source git → マッチなし、正常終了 (req 2.2)
			name:   "success: no skills found for source git (only go-mod skills in config)",
			skills: []string{},
			source: "git",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "skills")

				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}
				if err := cm.AddSkill(context.Background(), &domain.Skill{
					Name:    "gomod-skill",
					Source:  "go-mod",
					URL:     "github.com/example/gomod-skill",
					Version: "1.0.0",
				}); err != nil {
					t.Fatalf("failed to add skill: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			// skillNames 指定 かつ skill のソースがフィルタと不一致 → ErrSourceMismatch が UpdateResult.Err に格納され
			// run() はエラー集計として "update completed with 1 error(s)" を返す (req 1.5, 7.3)
			name:   "aggregate error: source mismatch when skill source does not match --source filter",
			skills: []string{"gomod-skill"},
			source: "git",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				installDir := filepath.Join(tmpDir, "skills")

				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{installDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}
				if err := cm.AddSkill(context.Background(), &domain.Skill{
					Name:    "gomod-skill",
					Source:  "go-mod",
					URL:     "github.com/example/gomod-skill",
					Version: "1.0.0",
				}); err != nil {
					t.Fatalf("failed to add skill: %v", err)
				}

				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: nil, // aggregate error "update completed with 1 error(s)", not ErrSourceMismatch directly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &UpdateCmd{
				Skills: tt.skills,
				Source: tt.source,
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
		})
	}
}
