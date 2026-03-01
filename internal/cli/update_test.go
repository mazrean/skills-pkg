package cli

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

// newTestLogger returns a Logger that writes to an in-memory buffer for inspection.
func newTestLogger() (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.out = &buf
	return logger, &buf
}

func TestUpdateCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrType error
		setupFunc   func(t *testing.T) (configPath string, cleanup func())
		name        string
		skills      []string
		wantErr     bool
	}{
		{
			name:   "error: config file not found (update all)",
			skills: []string{},
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
			wantErr:     true,
			wantErrType: domain.ErrSkillNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			cmd := &UpdateCmd{
				Skills: tt.skills,
			}

			// Execute command directly using the internal run method for testing
			err := cmd.run(configPath, false, false, "text") // non-verbose, non-dry-run, text output

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

func TestUpdateCmd_DryRun_ConfigNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".skillspkg.toml")

	cmd := &UpdateCmd{Skills: []string{}}
	err := cmd.run(configPath, false, true, "text")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestUpdateCmd_DryRun_SkillNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".skillspkg.toml")

	cm := domain.NewConfigManager(configPath)
	if err := cm.Initialize(context.Background(), []string{filepath.Join(tmpDir, "skills")}); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	cmd := &UpdateCmd{Skills: []string{"no-such-skill"}}
	err := cmd.run(configPath, false, true, "text")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrSkillNotFound) {
		t.Errorf("expected ErrSkillNotFound, got %v", err)
	}
}

func TestUpdateCmd_DryRun_TextOutput(t *testing.T) {
	t.Parallel()

	// printDryRunText のロジックを直接テスト
	logger, buf := newTestLogger()

	cmd := &UpdateCmd{}
	results := []*domain.UpdateResult{
		{SkillName: "skill-a", OldVersion: "1.0.0", NewVersion: "2.0.0"},
		{SkillName: "skill-b", OldVersion: "3.0.0", NewVersion: "3.0.0"},
	}

	if err := cmd.printDryRunText(logger, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "skill-a") || !strings.Contains(out, "1.0.0") || !strings.Contains(out, "2.0.0") {
		t.Errorf("expected update line for skill-a in output:\n%s", out)
	}
	if !strings.Contains(out, "skill-b") || !strings.Contains(out, "up to date") {
		t.Errorf("expected 'up to date' line for skill-b in output:\n%s", out)
	}
	if !strings.Contains(out, "1 update") {
		t.Errorf("expected update count summary in output:\n%s", out)
	}
}

func TestUpdateCmd_DryRun_JSONOutput(t *testing.T) {
	t.Parallel()

	logger, buf := newTestLogger()

	cmd := &UpdateCmd{}
	results := []*domain.UpdateResult{
		{SkillName: "skill-a", OldVersion: "1.0.0", NewVersion: "2.0.0"},
		{SkillName: "skill-b", OldVersion: "3.0.0", NewVersion: "3.0.0"},
	}

	if err := cmd.printDryRunJSON(logger, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"skill_name": "skill-a"`) {
		t.Errorf("expected skill-a in JSON output:\n%s", out)
	}
	if !strings.Contains(out, `"has_update": true`) {
		t.Errorf("expected has_update:true in JSON output:\n%s", out)
	}
	if !strings.Contains(out, `"has_update": false`) {
		t.Errorf("expected has_update:false in JSON output:\n%s", out)
	}
}
