package domain_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
	"github.com/mazrean/skills-pkg/internal/domain"
)

func TestHashVerifier_Verify(t *testing.T) {
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
	hashService := adapter.NewDirhashService()
	expectedHash, err := hashService.CalculateHash(ctx, skillDir)
	if err != nil {
		t.Fatalf("failed to calculate expected hash: %v", err)
	}

	// Create a config manager and initialize it
	configManager := domain.NewConfigManager(configPath)
	if err := configManager.Initialize(ctx, []string{tmpDir}); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
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
	if err := configManager.AddSkill(ctx, testSkill); err != nil {
		t.Fatalf("failed to add test skill: %v", err)
	}

	// Create the hash verifier
	verifier := domain.NewHashVerifier(configManager, hashService)

	t.Run("verify matching hash", func(t *testing.T) {
		result, err := verifier.Verify(ctx, "test-skill", skillDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.SkillName != "test-skill" {
			t.Errorf("expected skill name 'test-skill', got: %s", result.SkillName)
		}

		if result.InstallDir != skillDir {
			t.Errorf("expected install dir %s, got: %s", skillDir, result.InstallDir)
		}

		if result.Expected != expectedHash.Value {
			t.Errorf("expected hash %s, got: %s", expectedHash.Value, result.Expected)
		}

		if result.Actual != expectedHash.Value {
			t.Errorf("expected actual hash %s, got: %s", expectedHash.Value, result.Actual)
		}

		if !result.Match {
			t.Error("expected match to be true")
		}
	})

	t.Run("verify mismatching hash", func(t *testing.T) {
		// Modify the file to change the hash
		if err := os.WriteFile(testFile, []byte("modified content"), 0o644); err != nil {
			t.Fatalf("failed to modify test file: %v", err)
		}

		result, err := verifier.Verify(ctx, "test-skill", skillDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.Match {
			t.Error("expected match to be false")
		}

		if result.Expected == result.Actual {
			t.Error("expected different hash values")
		}
	})

	t.Run("verify non-existent skill", func(t *testing.T) {
		_, err := verifier.Verify(ctx, "non-existent-skill", skillDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("verify non-existent directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tmpDir, "non-existent")
		_, err := verifier.Verify(ctx, "test-skill", nonExistentDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestHashVerifier_VerifyAll(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".skillspkg.toml")

	// Create test skill directories
	skill1Dir := filepath.Join(tmpDir, "skill1")
	skill2Dir := filepath.Join(tmpDir, "skill2")
	skill3Dir := filepath.Join(tmpDir, "skill3")

	for _, dir := range []string{skill1Dir, skill2Dir, skill3Dir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create skill directory: %v", err)
		}
		// Create a test file in each skill directory
		testFile := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Calculate expected hashes
	hashService := adapter.NewDirhashService()
	hash1, err := hashService.CalculateHash(ctx, skill1Dir)
	if err != nil {
		t.Fatalf("failed to calculate hash1: %v", err)
	}
	hash2, err := hashService.CalculateHash(ctx, skill2Dir)
	if err != nil {
		t.Fatalf("failed to calculate hash2: %v", err)
	}
	hash3, err := hashService.CalculateHash(ctx, skill3Dir)
	if err != nil {
		t.Fatalf("failed to calculate hash3: %v", err)
	}

	// Create a config manager and initialize it
	configManager := domain.NewConfigManager(configPath)
	if err := configManager.Initialize(ctx, []string{tmpDir}); err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	// Add test skills
	testSkills := []*domain.Skill{
		{
			Name:      "skill1",
			Source:    "git",
			URL:       "https://github.com/test/skill1.git",
			Version:   "v1.0.0",
			HashAlgo:  hash1.Algorithm,
			HashValue: hash1.Value,
		},
		{
			Name:      "skill2",
			Source:    "npm",
			URL:       "skill2",
			Version:   "1.0.0",
			HashAlgo:  hash2.Algorithm,
			HashValue: hash2.Value,
		},
		{
			Name:      "skill3",
			Source:    "go-module",
			URL:       "github.com/test/skill3",
			Version:   "v1.0.0",
			HashAlgo:  hash3.Algorithm,
			HashValue: hash3.Value,
		},
	}

	for _, skill := range testSkills {
		if err := configManager.AddSkill(ctx, skill); err != nil {
			t.Fatalf("failed to add test skill: %v", err)
		}
	}

	// Create the hash verifier
	verifier := domain.NewHashVerifier(configManager, hashService)

	t.Run("verify all matching hashes", func(t *testing.T) {
		summary, err := verifier.VerifyAll(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if summary == nil {
			t.Fatal("expected summary, got nil")
		}

		if summary.TotalSkills != 3 {
			t.Errorf("expected 3 total skills, got: %d", summary.TotalSkills)
		}

		if summary.SuccessCount != 3 {
			t.Errorf("expected 3 successful verifications, got: %d", summary.SuccessCount)
		}

		if summary.FailureCount != 0 {
			t.Errorf("expected 0 failed verifications, got: %d", summary.FailureCount)
		}

		if len(summary.Results) != 3 {
			t.Errorf("expected 3 results, got: %d", len(summary.Results))
		}

		// Verify all results are successful
		for _, result := range summary.Results {
			if !result.Match {
				t.Errorf("expected skill %s to match, but it didn't", result.SkillName)
			}
		}
	})

	t.Run("verify with one mismatching hash", func(t *testing.T) {
		// Modify skill2 to create a hash mismatch
		skill2File := filepath.Join(skill2Dir, "test.txt")
		if err := os.WriteFile(skill2File, []byte("modified content"), 0o644); err != nil {
			t.Fatalf("failed to modify skill2 file: %v", err)
		}

		summary, err := verifier.VerifyAll(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if summary == nil {
			t.Fatal("expected summary, got nil")
		}

		if summary.TotalSkills != 3 {
			t.Errorf("expected 3 total skills, got: %d", summary.TotalSkills)
		}

		if summary.SuccessCount != 2 {
			t.Errorf("expected 2 successful verifications, got: %d", summary.SuccessCount)
		}

		if summary.FailureCount != 1 {
			t.Errorf("expected 1 failed verification, got: %d", summary.FailureCount)
		}

		// Verify that skill2 is the one that failed
		foundFailedSkill := false
		for _, result := range summary.Results {
			if result.SkillName == "skill2" && !result.Match {
				foundFailedSkill = true
			}
		}
		if !foundFailedSkill {
			t.Error("expected skill2 to have failed verification")
		}
	})

	t.Run("verify with no skills", func(t *testing.T) {
		// Create a new config manager with no skills
		emptyConfigPath := filepath.Join(tmpDir, ".skillspkg-empty.toml")
		emptyConfigManager := domain.NewConfigManager(emptyConfigPath)
		if err := emptyConfigManager.Initialize(ctx, []string{tmpDir}); err != nil {
			t.Fatalf("failed to initialize empty config: %v", err)
		}

		emptyVerifier := domain.NewHashVerifier(emptyConfigManager, hashService)
		summary, err := emptyVerifier.VerifyAll(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if summary == nil {
			t.Fatal("expected summary, got nil")
		}

		if summary.TotalSkills != 0 {
			t.Errorf("expected 0 total skills, got: %d", summary.TotalSkills)
		}

		if summary.SuccessCount != 0 {
			t.Errorf("expected 0 successful verifications, got: %d", summary.SuccessCount)
		}

		if summary.FailureCount != 0 {
			t.Errorf("expected 0 failed verifications, got: %d", summary.FailureCount)
		}

		if len(summary.Results) != 0 {
			t.Errorf("expected 0 results, got: %d", len(summary.Results))
		}
	})
}
