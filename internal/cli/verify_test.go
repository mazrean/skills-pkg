package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestVerifyCmd_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrType error
		setupFunc   func(t *testing.T) (configPath string, cleanup func())
		checkFunc   func(t *testing.T, output string)
		name        string
		wantErr     bool
	}{
		{
			name: "success: all skills verified with matching hashes",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				skillDir1 := filepath.Join(tmpDir, "skills", "skill1")
				skillDir2 := filepath.Join(tmpDir, "skills", "skill2")

				// Create skill directories with some content
				if err := os.MkdirAll(skillDir1, 0755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(skillDir1, "test.txt"), []byte("test content 1"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				if err := os.MkdirAll(skillDir2, 0755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(skillDir2, "test.txt"), []byte("test content 2"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				// Initialize config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{filepath.Join(tmpDir, "skills")}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Calculate actual hashes using the real hash service
				hashService := service.NewDirhash()
				hash1, _ := hashService.CalculateHash(context.Background(), skillDir1)
				hash2, _ := hashService.CalculateHash(context.Background(), skillDir2)

				// Add skills with correct hashes
				skill1 := &domain.Skill{
					Name:      "skill1",
					Source:    "git",
					URL:       "https://github.com/example/skill1.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: hash1.Value,
				}
				if err := cm.AddSkill(context.Background(), skill1); err != nil {
					t.Fatalf("failed to add skill1: %v", err)
				}

				skill2 := &domain.Skill{
					Name:      "skill2",
					Source:    "go-mod",
					URL:       "example-skill2",
					Version:   "2.0.0",
					HashAlgo:  "sha256",
					HashValue: hash2.Value,
				}
				if err := cm.AddSkill(context.Background(), skill2); err != nil {
					t.Fatalf("failed to add skill2: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check for verification summary (requirement 5.6)
				if !strings.Contains(output, "Verification complete") && !strings.Contains(output, "verification complete") {
					t.Errorf("output should contain verification summary, got: %s", output)
				}

				// Check for success count (requirement 5.6)
				if !strings.Contains(output, "2") {
					t.Errorf("output should show 2 successful verifications, got: %s", output)
				}

				// Check for zero failures (requirement 5.6)
				if !strings.Contains(output, "0") {
					t.Errorf("output should show 0 failures, got: %s", output)
				}

				// Should not contain warning messages (requirement 5.5)
				output = strings.ToLower(output)
				if strings.Contains(output, "warning") || strings.Contains(output, "mismatch") || strings.Contains(output, "tamper") {
					t.Errorf("output should not contain warnings for matching hashes, got: %s", output)
				}
			},
		},
		{
			name: "warning: hash mismatch detected",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")
				skillDir := filepath.Join(tmpDir, "skills", "skill1")

				// Create skill directory with content
				if err := os.MkdirAll(skillDir, 0755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(skillDir, "test.txt"), []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				// Initialize config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), []string{filepath.Join(tmpDir, "skills")}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add skill with incorrect hash
				skill := &domain.Skill{
					Name:      "skill1",
					Source:    "git",
					URL:       "https://github.com/example/skill1.git",
					Version:   "v1.0.0",
					HashAlgo:  "sha256",
					HashValue: "incorrect_hash_value",
				}
				if err := cm.AddSkill(context.Background(), skill); err != nil {
					t.Fatalf("failed to add skill: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false, // verify command should not return error, just warnings
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check for warning message (requirement 5.5)
				output = strings.ToLower(output)
				if !strings.Contains(output, "warning") && !strings.Contains(output, "mismatch") {
					t.Errorf("output should contain warning for hash mismatch, got: %s", output)
				}

				// Check for tamper detection message (requirement 5.5)
				if !strings.Contains(output, "tamper") && !strings.Contains(output, "modified") {
					t.Errorf("output should warn about possible tampering, got: %s", output)
				}

				// Check for failure count (requirement 5.6)
				if !strings.Contains(output, "1") {
					t.Errorf("output should show 1 failure, got: %s", output)
				}
			},
		},
		{
			name: "success: no skills to verify",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".skillspkg.toml")

				// Create empty config
				cm := domain.NewConfigManager(configPath)
				if err := cm.Initialize(context.Background(), nil); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				return configPath, func() {}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				t.Helper()

				// Check for appropriate message
				output = strings.ToLower(output)
				if !strings.Contains(output, "no skill") && !strings.Contains(output, "0 skill") {
					t.Errorf("output should indicate no skills to verify, got: %s", output)
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
			cmd := &VerifyCmd{}

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
