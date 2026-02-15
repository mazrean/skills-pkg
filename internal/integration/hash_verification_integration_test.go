package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/domain"
)

// TestHashVerificationIntegration tests the integration of hash verification with file system.
// This validates hash calculation, verification, and tampering detection.
// Requirements: 5.1-5.6, 12.1, 12.2, 12.3
func TestHashVerificationIntegration(t *testing.T) {
	t.Parallel()

	t.Run("Calculate_And_Verify_Hash_Integration", func(t *testing.T) {
		t.Parallel()

		// Setup: Create test skill directory
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")
		installDir := filepath.Join(tempDir, "skills")
		skillDir := filepath.Join(installDir, "test-skill")

		err := os.MkdirAll(skillDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}

		// Create test files in skill directory
		testFiles := map[string]string{
			"SKILL.md":      "# Test Skill\nDescription of the skill",
			"skill.go":      "package skill\n\nfunc Execute() {}",
			"config.toml":   "name = \"test-skill\"",
			"subdir/test.txt": "test content",
		}

		for path, content := range testFiles {
			fullPath := filepath.Join(skillDir, path)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
				t.Fatalf("Failed to create file %s: %v", path, err)
			}
		}

		ctx := context.Background()

		// Initialize configuration
		configManager := domain.NewConfigManager(configPath)
		err = configManager.Initialize(ctx, []string{installDir})
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// Add skill with placeholder hash
		skill := &domain.Skill{
			Name:      "test-skill",
			Source:    "git",
			URL:       skillDir,
			Version:   "1.0.0",
			HashAlgo:  "sha256",
			HashValue: "placeholder",
		}

		err = configManager.AddSkill(ctx, skill)
		if err != nil {
			t.Fatalf("AddSkill failed: %v", err)
		}

		// Setup hash service and verifier
		hashService := adapter.NewDirhashService()
		hashVerifier := domain.NewHashVerifier(configManager, hashService)

		// Test: Calculate hash
		hashResult, err := hashService.CalculateHash(ctx, skillDir)
		if err != nil {
			t.Fatalf("CalculateHash failed: %v", err)
		}

		if hashResult.Algorithm != "sha256" {
			t.Errorf("Expected algorithm sha256, got %s", hashResult.Algorithm)
		}

		if hashResult.Value == "" {
			t.Error("Hash value is empty")
		}

		// Update skill with calculated hash
		skill.HashValue = hashResult.Value
		err = configManager.UpdateSkill(ctx, skill)
		if err != nil {
			t.Fatalf("UpdateSkill failed: %v", err)
		}

		// Test: Verify hash directly on the skill directory (should pass)
		// Verify expects the skill directory path, not the parent directory
		verifyResult, err := hashVerifier.Verify(ctx, "test-skill", skillDir)
		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		}

		if !verifyResult.Match {
			t.Errorf("Hash verification failed: expected %s, got %s\nSkill dir: %s", verifyResult.Expected, verifyResult.Actual, skillDir)
		}
	})

	t.Run("Detect_Tampering_Integration", func(t *testing.T) {
		t.Parallel()

		// Setup: Create test skill directory
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")
		installDir := filepath.Join(tempDir, "skills")
		skillDir := filepath.Join(installDir, "tamper-skill")

		err := os.MkdirAll(skillDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}

		// Create test file
		testFile := filepath.Join(skillDir, "skill.go")
		originalContent := "package skill\n\nfunc Execute() {}"
		err = os.WriteFile(testFile, []byte(originalContent), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		ctx := context.Background()

		// Calculate hash of original content
		hashService := adapter.NewDirhashService()
		originalHash, err := hashService.CalculateHash(ctx, skillDir)
		if err != nil {
			t.Fatalf("CalculateHash failed: %v", err)
		}

		// Initialize configuration with original hash
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
			HashAlgo:  "sha256",
			HashValue: originalHash.Value,
		}

		err = configManager.AddSkill(ctx, skill)
		if err != nil {
			t.Fatalf("AddSkill failed: %v", err)
		}

		// Test: Tamper with file content
		tamperedContent := "package skill\n\n// MALICIOUS CODE\nfunc Execute() {}"
		err = os.WriteFile(testFile, []byte(tamperedContent), 0o644)
		if err != nil {
			t.Fatalf("Failed to tamper file: %v", err)
		}

		// Test: Verify hash (should fail)
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
	})

	t.Run("Verify_All_Skills_Integration", func(t *testing.T) {
		t.Parallel()

		// Setup: Create multiple test skill directories
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".skillspkg.toml")
		installDir := filepath.Join(tempDir, "skills")

		ctx := context.Background()

		configManager := domain.NewConfigManager(configPath)
		err := configManager.Initialize(ctx, []string{installDir})
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		hashService := adapter.NewDirhashService()

		// Create multiple skills
		numSkills := 3
		skills := make([]*domain.Skill, numSkills)

		for i := 0; i < numSkills; i++ {
			skillName := filepath.Join(installDir, filepath.Base(tempDir)+"-skill-"+string(rune('A'+i)))
			skillDir := filepath.Join(installDir, filepath.Base(skillName))

			err := os.MkdirAll(skillDir, 0o755)
			if err != nil {
				t.Fatalf("Failed to create skill directory %s: %v", skillDir, err)
			}

			testFile := filepath.Join(skillDir, "skill.go")
			content := "package skill" + string(rune('A'+i))
			err = os.WriteFile(testFile, []byte(content), 0o644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			hashResult, err := hashService.CalculateHash(ctx, skillDir)
			if err != nil {
				t.Fatalf("CalculateHash failed: %v", err)
			}

			skills[i] = &domain.Skill{
				Name:      filepath.Base(skillName),
				Source:    "git",
				URL:       skillDir,
				Version:   "1.0.0",
				HashAlgo:  "sha256",
				HashValue: hashResult.Value,
			}

			err = configManager.AddSkill(ctx, skills[i])
			if err != nil {
				t.Fatalf("AddSkill failed: %v", err)
			}
		}

		// Test: Verify all skills
		hashVerifier := domain.NewHashVerifier(configManager, hashService)
		summary, err := hashVerifier.VerifyAll(ctx)
		if err != nil {
			t.Fatalf("VerifyAll failed: %v", err)
		}

		// Verify: All skills passed verification
		if summary.TotalSkills != numSkills {
			t.Errorf("Expected %d total skills, got %d", numSkills, summary.TotalSkills)
		}

		if summary.SuccessCount != numSkills {
			t.Errorf("Expected %d successful verifications, got %d", numSkills, summary.SuccessCount)
		}

		if summary.FailureCount != 0 {
			t.Errorf("Expected 0 failures, got %d", summary.FailureCount)
		}

		// Test: Tamper with one skill
		tamperedSkillDir := filepath.Join(installDir, skills[1].Name)
		tamperedFile := filepath.Join(tamperedSkillDir, "skill.go")
		err = os.WriteFile(tamperedFile, []byte("tampered content"), 0o644)
		if err != nil {
			t.Fatalf("Failed to tamper file: %v", err)
		}

		// Test: Verify all skills again
		summary, err = hashVerifier.VerifyAll(ctx)
		if err != nil {
			t.Fatalf("VerifyAll failed: %v", err)
		}

		// Verify: One skill failed verification
		if summary.SuccessCount != numSkills-1 {
			t.Errorf("Expected %d successful verifications, got %d", numSkills-1, summary.SuccessCount)
		}

		if summary.FailureCount != 1 {
			t.Errorf("Expected 1 failure, got %d", summary.FailureCount)
		}
	})

	t.Run("Hash_Algorithm_Consistency", func(t *testing.T) {
		t.Parallel()

		// Setup
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
		hashService := adapter.NewDirhashService()

		// Test: Calculate hash multiple times
		hash1, err := hashService.CalculateHash(ctx, skillDir)
		if err != nil {
			t.Fatalf("First CalculateHash failed: %v", err)
		}

		hash2, err := hashService.CalculateHash(ctx, skillDir)
		if err != nil {
			t.Fatalf("Second CalculateHash failed: %v", err)
		}

		// Verify: Hash values are consistent
		if hash1.Value != hash2.Value {
			t.Errorf("Hash values are inconsistent: %s != %s", hash1.Value, hash2.Value)
		}

		if hash1.Algorithm != hash2.Algorithm {
			t.Errorf("Hash algorithms are inconsistent: %s != %s", hash1.Algorithm, hash2.Algorithm)
		}

		if hash1.Algorithm != "sha256" {
			t.Errorf("Expected sha256 algorithm, got %s", hash1.Algorithm)
		}
	})
}
