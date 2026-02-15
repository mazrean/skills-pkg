package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
)

// TestUninstallCmd_Run tests the basic uninstall functionality
func TestUninstallCmd_Run(t *testing.T) {
	// Create a temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".skillspkg.toml")
	installDir := filepath.Join(tempDir, "skills")

	// Initialize config with a test skill
	configManager := domain.NewConfigManager(configPath)
	err := configManager.Initialize(context.Background(), []string{installDir})
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
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
	err = configManager.AddSkill(context.Background(), skill)
	if err != nil {
		t.Fatalf("Failed to add test skill: %v", err)
	}

	// Create skill directory to simulate installed skill
	skillDir := filepath.Join(installDir, "test-skill")
	err = os.MkdirAll(skillDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	testFile := filepath.Join(skillDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run uninstall command
	cmd := &UninstallCmd{
		SkillName: "test-skill",
	}

	cmdErr := cmd.run(configPath, false)
	if cmdErr != nil {
		t.Fatalf("Uninstall command failed: %v", cmdErr)
	}

	// Verify skill directory is removed
	if _, statErr := os.Stat(skillDir); !os.IsNotExist(statErr) {
		t.Errorf("Skill directory still exists after uninstall")
	}

	// Verify skill is removed from config
	config, err := configManager.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.FindSkillByName("test-skill") != nil {
		t.Errorf("Skill still exists in configuration after uninstall")
	}
}

// TestUninstallCmd_Run_NonExistentSkill tests error handling for non-existent skill
func TestUninstallCmd_Run_NonExistentSkill(t *testing.T) {
	// Create a temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".skillspkg.toml")
	installDir := filepath.Join(tempDir, "skills")

	// Initialize config without any skills
	configManager := domain.NewConfigManager(configPath)
	err := configManager.Initialize(context.Background(), []string{installDir})
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Run uninstall command for non-existent skill
	cmd := &UninstallCmd{
		SkillName: "non-existent-skill",
	}

	err = cmd.run(configPath, false)
	if err == nil {
		t.Fatal("Expected error for non-existent skill, got nil")
	}

	// Verify error message contains helpful information
	expectedMsg := "skill 'non-existent-skill' not found"
	if err.Error() == "" || len(err.Error()) < len(expectedMsg) {
		t.Errorf("Error message too short or empty: %v", err)
	}
}

// TestUninstallCmd_Run_ConfigNotFound tests error handling when config file doesn't exist
func TestUninstallCmd_Run_ConfigNotFound(t *testing.T) {
	// Use a non-existent config path
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent", ".skillspkg.toml")

	// Run uninstall command
	cmd := &UninstallCmd{
		SkillName: "test-skill",
	}

	err := cmd.run(configPath, false)
	if err == nil {
		t.Fatal("Expected error for non-existent config, got nil")
	}
}

// TestUninstallCmd_Run_MultipleTargets tests uninstall from multiple install directories
func TestUninstallCmd_Run_MultipleTargets(t *testing.T) {
	// Create a temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".skillspkg.toml")
	installDir1 := filepath.Join(tempDir, "skills1")
	installDir2 := filepath.Join(tempDir, "skills2")

	// Initialize config with multiple install targets
	configManager := domain.NewConfigManager(configPath)
	err := configManager.Initialize(context.Background(), []string{installDir1, installDir2})
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
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
	err = configManager.AddSkill(context.Background(), skill)
	if err != nil {
		t.Fatalf("Failed to add test skill: %v", err)
	}

	// Create skill directories in both targets
	for _, installDir := range []string{installDir1, installDir2} {
		skillDir := filepath.Join(installDir, "test-skill")
		err = os.MkdirAll(skillDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create skill directory in %s: %v", installDir, err)
		}
		testFile := filepath.Join(skillDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0o644)
		if err != nil {
			t.Fatalf("Failed to write test file in %s: %v", installDir, err)
		}
	}

	// Run uninstall command
	cmd := &UninstallCmd{
		SkillName: "test-skill",
	}

	cmdErr := cmd.run(configPath, false)
	if cmdErr != nil {
		t.Fatalf("Uninstall command failed: %v", cmdErr)
	}

	// Verify skill directories are removed from both targets
	for _, installDir := range []string{installDir1, installDir2} {
		skillDir := filepath.Join(installDir, "test-skill")
		if _, statErr := os.Stat(skillDir); !os.IsNotExist(statErr) {
			t.Errorf("Skill directory still exists in %s after uninstall", installDir)
		}
	}

	// Verify skill is removed from config
	config, err := configManager.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.FindSkillByName("test-skill") != nil {
		t.Errorf("Skill still exists in configuration after uninstall")
	}
}
