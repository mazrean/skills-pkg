package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// TestHashVerificationIntegration tests the integration of hash verification with file system.
// This validates hash calculation, verification, and tampering detection.
// Requirements: 5.1-5.6, 12.1, 12.2, 12.3
func TestHashVerificationIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		testFunc  func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, skillDir string)
		setupFunc func(t *testing.T) (ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, skillDir string)
		name      string
	}{
		{
			name: "Calculate_And_Verify_Hash_Integration",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager, *service.Dirhash, string) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir := filepath.Join(tempDir, "skills")
				skillDir := filepath.Join(installDir, "test-skill")

				err := os.MkdirAll(skillDir, 0o755)
				if err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}

				testFiles := map[string]string{
					"SKILL.md":        "# Test Skill\nDescription of the skill",
					"skill.go":        "package skill\n\nfunc Execute() {}",
					"config.toml":     "name = \"test-skill\"",
					"subdir/test.txt": "test content",
				}

				for path, content := range testFiles {
					fullPath := filepath.Join(skillDir, path)
					dir := filepath.Dir(fullPath)
					if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
						t.Fatalf("Failed to create directory %s: %v", dir, mkdirErr)
					}
					if writeErr := os.WriteFile(fullPath, []byte(content), 0o644); writeErr != nil {
						t.Fatalf("Failed to create file %s: %v", path, writeErr)
					}
				}

				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)
				err = configManager.Initialize(ctx, []string{installDir})
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}

				hashService := service.NewDirhash()
				return ctx, configManager, hashService, skillDir
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, skillDir string) {
				skill := &domain.Skill{
					Name:      "test-skill",
					Source:    "git",
					URL:       skillDir,
					Version:   "1.0.0",
					HashValue: "placeholder",
				}

				err := configManager.AddSkill(ctx, skill)
				if err != nil {
					t.Fatalf("AddSkill failed: %v", err)
				}

				hashVerifier := domain.NewHashVerifier(configManager, hashService)
				hashResult, err := hashService.CalculateHash(ctx, skillDir)
				if err != nil {
					t.Fatalf("CalculateHash failed: %v", err)
				}

				if hashResult.Value == "" {
					t.Error("Hash value is empty")
				}

				skill.HashValue = hashResult.Value
				err = configManager.UpdateSkill(ctx, skill)
				if err != nil {
					t.Fatalf("UpdateSkill failed: %v", err)
				}

				verifyResult, err := hashVerifier.Verify(ctx, "test-skill", skillDir)
				if err != nil {
					t.Fatalf("Verify failed: %v", err)
				}

				if !verifyResult.Match {
					t.Errorf("Hash verification failed: expected %s, got %s\nSkill dir: %s", verifyResult.Expected, verifyResult.Actual, skillDir)
				}
			},
		},
		{
			name: "Detect_Tampering_Integration",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager, *service.Dirhash, string) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir := filepath.Join(tempDir, "skills")
				skillDir := filepath.Join(installDir, "tamper-skill")

				err := os.MkdirAll(skillDir, 0o755)
				if err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}

				testFile := filepath.Join(skillDir, "skill.go")
				originalContent := "package skill\n\nfunc Execute() {}"
				err = os.WriteFile(testFile, []byte(originalContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				ctx := context.Background()
				hashService := service.NewDirhash()
				originalHash, err := hashService.CalculateHash(ctx, skillDir)
				if err != nil {
					t.Fatalf("CalculateHash failed: %v", err)
				}

				configManager := domain.NewConfigManager(configPath)
				err = configManager.Initialize(ctx, []string{installDir})
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}

				skill := &domain.Skill{
					Name:      "tamper-skill",
					Source:    "git",
					URL:       skillDir,
					Version:   "1.0.0",
					HashValue: originalHash.Value,
				}

				err = configManager.AddSkill(ctx, skill)
				if err != nil {
					t.Fatalf("AddSkill failed: %v", err)
				}

				return ctx, configManager, hashService, skillDir
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, skillDir string) {
				testFile := filepath.Join(skillDir, "skill.go")
				tamperedContent := "package skill\n\n// MALICIOUS CODE\nfunc Execute() {}"
				err := os.WriteFile(testFile, []byte(tamperedContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to tamper file: %v", err)
				}

				hashVerifier := domain.NewHashVerifier(configManager, hashService)
				verifyResult, err := hashVerifier.Verify(ctx, "tamper-skill", skillDir)
				if err != nil {
					t.Fatalf("Verify failed: %v", err)
				}

				if verifyResult.Match {
					t.Error("Hash verification passed for tampered file - tampering was not detected")
				}

				if verifyResult.Expected == verifyResult.Actual {
					t.Error("Expected and actual hash should be different for tampered file")
				}
			},
		},
		{
			name: "Verify_All_Skills_Integration",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager, *service.Dirhash, string) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skillspkg.toml")
				installDir := filepath.Join(tempDir, "skills")

				ctx := context.Background()
				configManager := domain.NewConfigManager(configPath)
				err := configManager.Initialize(ctx, []string{installDir})
				if err != nil {
					t.Fatalf("Initialize failed: %v", err)
				}

				hashService := service.NewDirhash()
				numSkills := 3

				for i := range numSkills {
					skillName := filepath.Join(installDir, filepath.Base(tempDir)+"-skill-"+string(rune('A'+i)))
					skillDir := filepath.Join(installDir, filepath.Base(skillName))

					if mkdirErr := os.MkdirAll(skillDir, 0o755); mkdirErr != nil {
						t.Fatalf("Failed to create skill directory %s: %v", skillDir, mkdirErr)
					}

					testFile := filepath.Join(skillDir, "skill.go")
					content := "package skill" + string(rune('A'+i))
					if writeErr := os.WriteFile(testFile, []byte(content), 0o644); writeErr != nil {
						t.Fatalf("Failed to create test file: %v", writeErr)
					}

					hashResult, hashErr := hashService.CalculateHash(ctx, skillDir)
					if hashErr != nil {
						t.Fatalf("CalculateHash failed: %v", hashErr)
					}

					skill := &domain.Skill{
						Name:      filepath.Base(skillName),
						Source:    "git",
						URL:       skillDir,
						Version:   "1.0.0",
						HashValue: hashResult.Value,
					}

					err = configManager.AddSkill(ctx, skill)
					if err != nil {
						t.Fatalf("AddSkill failed: %v", err)
					}
				}

				return ctx, configManager, hashService, installDir
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, installDir string) {
				const numSkills = 3

				hashVerifier := domain.NewHashVerifier(configManager, hashService)
				summary, err := hashVerifier.VerifyAll(ctx)
				if err != nil {
					t.Fatalf("VerifyAll failed: %v", err)
				}

				if summary.TotalSkills != numSkills {
					t.Errorf("Expected %d total skills, got %d", numSkills, summary.TotalSkills)
				}

				if summary.SuccessCount != numSkills {
					t.Errorf("Expected %d successful verifications, got %d", numSkills, summary.SuccessCount)
				}

				if summary.FailureCount != 0 {
					t.Errorf("Expected 0 failures, got %d", summary.FailureCount)
				}

				config, err := configManager.Load(ctx)
				if err != nil {
					t.Fatalf("Load config failed: %v", err)
				}

				if len(config.Skills) > 1 {
					tamperedSkillDir := filepath.Join(installDir, config.Skills[1].Name)
					tamperedFile := filepath.Join(tamperedSkillDir, "skill.go")
					err = os.WriteFile(tamperedFile, []byte("tampered content"), 0o644)
					if err != nil {
						t.Fatalf("Failed to tamper file: %v", err)
					}

					summary, err = hashVerifier.VerifyAll(ctx)
					if err != nil {
						t.Fatalf("VerifyAll failed: %v", err)
					}

					if summary.SuccessCount != numSkills-1 {
						t.Errorf("Expected %d successful verifications, got %d", numSkills-1, summary.SuccessCount)
					}

					if summary.FailureCount != 1 {
						t.Errorf("Expected 1 failure, got %d", summary.FailureCount)
					}
				}
			},
		},
		{
			name: "Hash_Algorithm_Consistency",
			setupFunc: func(t *testing.T) (context.Context, *domain.ConfigManager, *service.Dirhash, string) {
				tempDir := t.TempDir()
				skillDir := filepath.Join(tempDir, "test-skill")

				err := os.MkdirAll(skillDir, 0o755)
				if err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}

				testFile := filepath.Join(skillDir, "test.txt")
				content := "test content"
				err = os.WriteFile(testFile, []byte(content), 0o644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				ctx := context.Background()
				hashService := service.NewDirhash()
				return ctx, nil, hashService, skillDir
			},
			testFunc: func(t *testing.T, ctx context.Context, configManager *domain.ConfigManager, hashService *service.Dirhash, skillDir string) {
				hash1, err := hashService.CalculateHash(ctx, skillDir)
				if err != nil {
					t.Fatalf("First CalculateHash failed: %v", err)
				}

				hash2, err := hashService.CalculateHash(ctx, skillDir)
				if err != nil {
					t.Fatalf("Second CalculateHash failed: %v", err)
				}

				if hash1.Value != hash2.Value {
					t.Errorf("Hash values are inconsistent: %s != %s", hash1.Value, hash2.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, configManager, hashService, skillDir := tt.setupFunc(t)
			tt.testFunc(t, ctx, configManager, hashService, skillDir)
		})
	}
}
