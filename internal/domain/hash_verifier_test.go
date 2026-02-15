package domain_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter/service"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestHashVerifier_Verify(t *testing.T) {
	tests := []struct {
		validate   func(t *testing.T, result *domain.VerifyResult, expectedHash string, skillDir string)
		name       string
		skillName  string
		setupFile  bool
		modifyFile bool
		wantErr    bool
	}{
		{
			name:       "verify matching hash",
			setupFile:  true,
			modifyFile: false,
			skillName:  "test-skill",
			wantErr:    false,
			validate: func(t *testing.T, result *domain.VerifyResult, expectedHash string, skillDir string) {
				if result == nil {
					t.Fatal("expected result, got nil")
				}

				if result.SkillName != "test-skill" {
					t.Errorf("expected skill name 'test-skill', got: %s", result.SkillName)
				}

				if result.InstallDir != skillDir {
					t.Errorf("expected install dir %s, got: %s", skillDir, result.InstallDir)
				}

				if result.Expected != expectedHash {
					t.Errorf("expected hash %s, got: %s", expectedHash, result.Expected)
				}

				if result.Actual != expectedHash {
					t.Errorf("expected actual hash %s, got: %s", expectedHash, result.Actual)
				}

				if !result.Match {
					t.Error("expected match to be true")
				}
			},
		},
		{
			name:       "verify mismatching hash",
			setupFile:  true,
			modifyFile: true,
			skillName:  "test-skill",
			wantErr:    false,
			validate: func(t *testing.T, result *domain.VerifyResult, expectedHash string, skillDir string) {
				if result == nil {
					t.Fatal("expected result, got nil")
				}

				if result.Match {
					t.Error("expected match to be false")
				}

				if result.Expected == result.Actual {
					t.Error("expected different hash values")
				}
			},
		},
		{
			name:       "verify non-existent skill",
			setupFile:  true,
			modifyFile: false,
			skillName:  "non-existent-skill",
			wantErr:    true,
			validate:   nil,
		},
		{
			name:       "verify non-existent directory",
			setupFile:  false,
			modifyFile: false,
			skillName:  "test-skill",
			wantErr:    true,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create a test skill directory
			skillDir := filepath.Join(tmpDir, "test-skill")
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				t.Fatalf("failed to create skill directory: %v", err)
			}

			// Create a test file in the skill directory
			testFile := filepath.Join(skillDir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Calculate the expected hash
			hashService := service.NewDirhash()
			expectedHash, err := hashService.CalculateHash(ctx, skillDir)
			if err != nil {
				t.Fatalf("failed to calculate expected hash: %v", err)
			}

			// Create a config manager and initialize it
			configManager := domain.NewConfigManager(configPath)
			if initErr := configManager.Initialize(ctx, []string{tmpDir}); initErr != nil {
				t.Fatalf("failed to initialize config: %v", initErr)
			}

			// Add a test skill with the calculated hash
			testSkill := &domain.Skill{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/test/test-skill.git",
				Version:   "v1.0.0",
				HashAlgo:  expectedHash.Algorithm,
				HashValue: expectedHash.Value,
			}
			if addErr := configManager.AddSkill(ctx, testSkill); addErr != nil {
				t.Fatalf("failed to add test skill: %v", addErr)
			}

			// Create the hash verifier
			verifier := domain.NewHashVerifier(configManager, hashService)

			// Modify the file if requested
			if tt.modifyFile {
				if writeErr := os.WriteFile(testFile, []byte("modified content"), 0o644); writeErr != nil {
					t.Fatalf("failed to modify test file: %v", writeErr)
				}
			}

			// Execute Verify
			targetDir := skillDir
			if !tt.setupFile {
				targetDir = filepath.Join(tmpDir, "non-existent")
			}

			result, err := verifier.Verify(ctx, tt.skillName, targetDir)

			// Verify error
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Validate the result
			if tt.validate != nil {
				tt.validate(t, result, expectedHash.Value, skillDir)
			}
		})
	}
}

func TestHashVerifier_VerifyAll(t *testing.T) {
	tests := []struct {
		name                string
		wantFailedSkillName string
		skillCount          int
		modifySkillIndex    int
		wantTotalSkills     int
		wantSuccessCount    int
		wantFailureCount    int
		expectEmptyConfig   bool
		wantErr             bool
	}{
		{
			name:              "verify all matching hashes",
			skillCount:        3,
			modifySkillIndex:  -1,
			expectEmptyConfig: false,
			wantErr:           false,
			wantTotalSkills:   3,
			wantSuccessCount:  3,
			wantFailureCount:  0,
		},
		{
			name:                "verify with one mismatching hash",
			skillCount:          3,
			modifySkillIndex:    1,
			expectEmptyConfig:   false,
			wantErr:             false,
			wantTotalSkills:     3,
			wantSuccessCount:    2,
			wantFailureCount:    1,
			wantFailedSkillName: "skill2",
		},
		{
			name:              "verify with no skills",
			skillCount:        0,
			modifySkillIndex:  -1,
			expectEmptyConfig: true,
			wantErr:           false,
			wantTotalSkills:   0,
			wantSuccessCount:  0,
			wantFailureCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".skillspkg.toml")

			// Create test skill directories
			skillDirs := make([]string, tt.skillCount)
			hashService := service.NewDirhash()
			hashes := make([]*port.HashResult, tt.skillCount)

			for i := 0; i < tt.skillCount; i++ {
				skillDir := filepath.Join(tmpDir, "skill"+string(rune('1'+i)))
				skillDirs[i] = skillDir

				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				// Create a test file in each skill directory
				testFile := filepath.Join(skillDir, "test.txt")
				if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				// Calculate hash
				hash, err := hashService.CalculateHash(ctx, skillDir)
				if err != nil {
					t.Fatalf("failed to calculate hash: %v", err)
				}
				hashes[i] = hash
			}

			// Create a config manager and initialize it
			var configManager *domain.ConfigManager
			var verifier *domain.HashVerifier

			if tt.expectEmptyConfig {
				emptyConfigPath := filepath.Join(tmpDir, ".skillspkg-empty.toml")
				configManager = domain.NewConfigManager(emptyConfigPath)
				if err := configManager.Initialize(ctx, []string{tmpDir}); err != nil {
					t.Fatalf("failed to initialize empty config: %v", err)
				}
				verifier = domain.NewHashVerifier(configManager, hashService)
			} else {
				configManager = domain.NewConfigManager(configPath)
				if err := configManager.Initialize(ctx, []string{tmpDir}); err != nil {
					t.Fatalf("failed to initialize config: %v", err)
				}

				// Add test skills
				sources := []string{"git", "go-module", "go-module"}
				urlPrefixes := []string{"https://github.com/test/skill", "", "github.com/test/skill"}

				for i := 0; i < tt.skillCount; i++ {
					skillNum := i + 1
					skillName := "skill" + string(rune('0'+skillNum))
					sourceIdx := i % len(sources)

					url := urlPrefixes[sourceIdx]
					if url != "" {
						url += string(rune('0' + skillNum))
						if sourceIdx == 0 {
							url += ".git"
						}
					} else {
						url = skillName
					}

					testSkill := &domain.Skill{
						Name:      skillName,
						Source:    sources[sourceIdx],
						URL:       url,
						Version:   "v1.0.0",
						HashAlgo:  hashes[i].Algorithm,
						HashValue: hashes[i].Value,
					}

					if sourceIdx == 1 {
						testSkill.Version = "1.0.0"
					}

					if err := configManager.AddSkill(ctx, testSkill); err != nil {
						t.Fatalf("failed to add test skill: %v", err)
					}
				}

				// Create the hash verifier
				verifier = domain.NewHashVerifier(configManager, hashService)
			}

			// Modify a skill if requested
			if tt.modifySkillIndex >= 0 && tt.modifySkillIndex < len(skillDirs) {
				skillFile := filepath.Join(skillDirs[tt.modifySkillIndex], "test.txt")
				if err := os.WriteFile(skillFile, []byte("modified content"), 0o644); err != nil {
					t.Fatalf("failed to modify skill file: %v", err)
				}
			}

			// Execute VerifyAll
			summary, err := verifier.VerifyAll(ctx)

			// Verify error
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if summary == nil {
				t.Fatal("expected summary, got nil")
			}

			// Verify summary
			if summary.TotalSkills != tt.wantTotalSkills {
				t.Errorf("expected %d total skills, got: %d", tt.wantTotalSkills, summary.TotalSkills)
			}

			if summary.SuccessCount != tt.wantSuccessCount {
				t.Errorf("expected %d successful verifications, got: %d", tt.wantSuccessCount, summary.SuccessCount)
			}

			if summary.FailureCount != tt.wantFailureCount {
				t.Errorf("expected %d failed verifications, got: %d", tt.wantFailureCount, summary.FailureCount)
			}

			if len(summary.Results) != tt.wantTotalSkills {
				t.Errorf("expected %d results, got: %d", tt.wantTotalSkills, len(summary.Results))
			}

			// Verify specific failed skill if expected
			if tt.wantFailedSkillName != "" {
				foundFailedSkill := false
				for _, result := range summary.Results {
					if result.SkillName == tt.wantFailedSkillName && !result.Match {
						foundFailedSkill = true
					}
				}
				if !foundFailedSkill {
					t.Errorf("expected %s to have failed verification", tt.wantFailedSkillName)
				}
			}

			// Verify all results are successful if no failures expected
			if tt.wantFailureCount == 0 {
				for _, result := range summary.Results {
					if !result.Match {
						t.Errorf("expected skill %s to match, but it didn't", result.SkillName)
					}
				}
			}
		})
	}
}
