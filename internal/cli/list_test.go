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

func TestListCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrType error
		setupFunc   func(t *testing.T) (configPath string, cleanup func())
		checkFunc   func(t *testing.T, output string)
		name        string
		wantErr     bool
	}{
		{
			name: "success: list skills with multiple entries",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config with multiple skills
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add git skill
				gitSkill := &domain.Skill{
					Name:      "git-skill",
					Source:    "git",
					URL:       "https://github.com/example/skill.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "abc123",
				}
				if err := cm.AddSkill(context.Background(), gitSkill); err != nil {
					t.Fatalf("failed to add git skill: %v", err)
				}

				// Add go-module skill
				goModSkill := &domain.Skill{
					Name:      "go-module-skill",
					Source:    "go-module",
					URL:       "github.com/example/skill",
					Version:   "v2.0.0",
					HashAlgo:  "sha256",
					HashValue: "def456",
				}
				if err := cm.AddSkill(context.Background(), goModSkill); err != nil {
					t.Fatalf("failed to add go-module skill: %v", err)
				}

				// Add go-module skill
				goSkill := &domain.Skill{
					Name:      "go-skill",
					Source:    "go-module",
					URL:       "github.com/example/go-skill",
					Version:   "v0.5.0",
					HashAlgo:  "sha256",
					HashValue: "ghi789",
				}
				if err := cm.AddSkill(context.Background(), goSkill); err != nil {
					t.Fatalf("failed to add go skill: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check that all skill names are present (requirement 8.2)
				if !strings.Contains(output, "git-skill") {
					t.Errorf("output should contain 'git-skill', got: %s", output)
				}
				if !strings.Contains(output, "go-module-skill") {
					t.Errorf("output should contain 'go-module-skill', got: %s", output)
				}
				if !strings.Contains(output, "go-skill") {
					t.Errorf("output should contain 'go-skill', got: %s", output)
				}

				// Check that source types are present (requirement 8.2)
				if !strings.Contains(output, "git") {
					t.Errorf("output should contain source 'git', got: %s", output)
				}
				if !strings.Contains(output, "go-module") {
					t.Errorf("output should contain source 'go-module', got: %s", output)
				}
				if !strings.Contains(output, "go-module") {
					t.Errorf("output should contain source 'go-module', got: %s", output)
				}

				// Check that versions are present (requirement 8.2)
				if !strings.Contains(output, "v1.0.0") {
					t.Errorf("output should contain version 'v1.0.0', got: %s", output)
				}
				if !strings.Contains(output, "2.0.0") {
					t.Errorf("output should contain version '2.0.0', got: %s", output)
				}
				if !strings.Contains(output, "v0.5.0") {
					t.Errorf("output should contain version 'v0.5.0', got: %s", output)
				}

				// Check for table-like formatting (requirement 8.3)
				// Should have headers or structured format
				lines := strings.Split(output, "\n")
				if len(lines) < 3 { // At least header + 3 skills
					t.Errorf("output should have at least 4 lines (header + 3 skills), got %d lines", len(lines))
				}
			},
		},
		{
			name: "success: empty skill list shows message",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create initial config with no skills
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check for message indicating no skills (requirement 8.4)
				output = strings.ToLower(output)
				if !strings.Contains(output, "no skill") && !strings.Contains(output, "empty") {
					t.Errorf("output should indicate no skills are installed, got: %s", output)
				}
			},
		},
		{
			name: "error: configuration file not found",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				// Don't create config file

				return configPath, func() {}
			},
			wantErr:     true,
			wantErrType: domain.ErrConfigNotFound,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check for appropriate error message (requirements 12.2, 12.3)
				output = strings.ToLower(output)
				if !strings.Contains(output, "not found") {
					t.Errorf("output should indicate config not found, got: %s", output)
				}
				if !strings.Contains(output, "init") {
					t.Errorf("output should suggest running 'init', got: %s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Create command
			cmd := &ListCmd{}

			// Capture output
			var outBuf, errBuf bytes.Buffer
			logger := &Logger{
				out:     &outBuf,
				errOut:  &errBuf,
				verbose: false,
			}

			// Run command with test logger
			err := cmd.runWithLogger(configPath, logger)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check error type if specified
			if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
				t.Errorf("run() error type = %v, want %v", err, tt.wantErrType)
				return
			}

			// Check output if check function is provided
			if tt.checkFunc != nil {
				output := outBuf.String() + errBuf.String()
				tt.checkFunc(t, output)
			}
		})
	}
}
