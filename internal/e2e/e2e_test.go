package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestE2ECompleteFlow tests the complete workflow: init -> add -> install -> verify -> uninstall
// Requirements: 12.4, 12.5, 12.6
func TestE2ECompleteFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Create temporary workspace
	workspaceDir := t.TempDir()
	projectDir := filepath.Join(workspaceDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Setup: Create test Git repository
	testRepoDir := filepath.Join(workspaceDir, "test-skill-repo")
	testRepoURL := createTestGitRepo(t, testRepoDir)

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Setup: Define install targets
	installDir1 := filepath.Join(workspaceDir, "agent1", "skills")
	installDir2 := filepath.Join(workspaceDir, "agent2", "skills")

	// Test: Step 1 - Initialize project
	t.Run("init", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "init",
			"--install-dir", installDir1,
			"--install-dir", installDir2,
		)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify .skillspkg.toml was created
		configPath := filepath.Join(projectDir, ".skillspkg.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Configuration file was not created at %s", configPath)
		}
	})

	// Test: Step 2 - Add a skill
	t.Run("add", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "add",
			"test-skill",
			"--source", "git",
			"--url", testRepoURL,
			"--version", "v1.0.0",
		)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("add command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}
	})

	// Test: Step 3 - Install the skill
	t.Run("install", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "install", "test-skill")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("install command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify skill was installed in both directories
		skillPath1 := filepath.Join(installDir1, "test-skill")
		skillPath2 := filepath.Join(installDir2, "test-skill")

		if _, err := os.Stat(skillPath1); os.IsNotExist(err) {
			t.Errorf("Skill was not installed in %s", skillPath1)
		}
		if _, err := os.Stat(skillPath2); os.IsNotExist(err) {
			t.Errorf("Skill was not installed in %s", skillPath2)
		}
	})

	// Test: Step 4 - Verify skills
	t.Run("verify", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "verify")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("verify command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify output contains success message
		outputStr := string(output)
		// Check for either "Successful" count or "成功" (Japanese)
		hasSuccessIndicator := strings.Contains(outputStr, "Successful:") ||
			strings.Contains(outputStr, "成功") ||
			strings.Contains(strings.ToLower(outputStr), "success")
		if !hasSuccessIndicator {
			t.Errorf("Expected success message in verify output, got: %s", outputStr)
		}
		// Verify no failures
		if strings.Contains(outputStr, "Failed: 0") == false &&
			strings.Contains(outputStr, "失敗: 0") == false {
			t.Errorf("Expected no failures in verify output, got: %s", outputStr)
		}
	})

	// Test: Step 5 - List skills
	t.Run("list", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "list")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("list command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify output contains skill name
		outputStr := string(output)
		if !strings.Contains(outputStr, "test-skill") {
			t.Errorf("Expected skill name in list output, got: %s", outputStr)
		}
	})

	// Test: Step 6 - Uninstall the skill
	t.Run("uninstall", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "uninstall", "test-skill")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("uninstall command failed: %v\nOutput: %s", err, output)
		}

		// Verify exit code 0
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify skill was removed from both directories
		skillPath1 := filepath.Join(installDir1, "test-skill")
		skillPath2 := filepath.Join(installDir2, "test-skill")

		if _, err := os.Stat(skillPath1); !os.IsNotExist(err) {
			t.Errorf("Skill was not removed from %s", skillPath1)
		}
		if _, err := os.Stat(skillPath2); !os.IsNotExist(err) {
			t.Errorf("Skill was not removed from %s", skillPath2)
		}
	})
}

// TestE2EMultipleAgentInstallation tests installing skills to multiple agent directories
// Requirements: 10.1, 10.2, 10.5
func TestE2EMultipleAgentInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Create temporary workspace
	workspaceDir := t.TempDir()
	projectDir := filepath.Join(workspaceDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Setup: Create test Git repository
	testRepoDir := filepath.Join(workspaceDir, "test-skill-repo")
	testRepoURL := createTestGitRepo(t, testRepoDir)

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Setup: Define install targets for multiple agents
	agentDirs := []string{
		filepath.Join(workspaceDir, "claude", "skills"),
		filepath.Join(workspaceDir, "codex", "skills"),
		filepath.Join(workspaceDir, "custom", "skills"),
	}

	// Test: Initialize with multiple agent directories
	t.Run("init_multiple_agents", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"init"}
		for _, dir := range agentDirs {
			args = append(args, "--install-dir", dir)
		}

		cmd := exec.CommandContext(ctx, binaryPath, args...)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init command failed: %v\nOutput: %s", err, output)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
		}
	})

	// Test: Add and install skill
	t.Run("add_and_install", func(t *testing.T) {
		ctx := context.Background()
		// Add skill
		cmd := exec.CommandContext(ctx, binaryPath, "add",
			"multi-agent-skill",
			"--source", "git",
			"--url", testRepoURL,
			"--version", "v1.0.0",
		)
		cmd.Dir = projectDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("add command failed: %v\nOutput: %s", err, output)
		}

		// Install skill
		cmd = exec.CommandContext(ctx, binaryPath, "install")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("install command failed: %v\nOutput: %s", err, output)
		}

		// Verify skill is installed in all agent directories
		for _, agentDir := range agentDirs {
			skillPath := filepath.Join(agentDir, "multi-agent-skill")
			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				t.Errorf("Skill was not installed in %s", agentDir)
			}

			// Verify SKILL.md exists
			skillMdPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
				t.Errorf("SKILL.md was not found in %s", skillPath)
			}
		}
	})
}

// TestE2EHashMismatchWarning tests that hash mismatches are properly detected and warned
// Requirements: 5.4, 5.5, 5.6
func TestE2EHashMismatchWarning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Create temporary workspace
	workspaceDir := t.TempDir()
	projectDir := filepath.Join(workspaceDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Setup: Create test Git repository
	testRepoDir := filepath.Join(workspaceDir, "test-skill-repo")
	testRepoURL := createTestGitRepo(t, testRepoDir)

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Setup: Define install target
	installDir := filepath.Join(workspaceDir, "skills")

	// Test: Initialize and install skill
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binaryPath, "init", "--install-dir", installDir)
	cmd.Dir = projectDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init command failed: %v\nOutput: %s", err, output)
	}

	cmd = exec.CommandContext(ctx, binaryPath, "add",
		"tamper-test-skill",
		"--source", "git",
		"--url", testRepoURL,
		"--version", "v1.0.0",
	)
	cmd.Dir = projectDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("add command failed: %v\nOutput: %s", err, output)
	}

	cmd = exec.CommandContext(ctx, binaryPath, "install")
	cmd.Dir = projectDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("install command failed: %v\nOutput: %s", err, output)
	}

	// Test: Tamper with installed skill
	t.Run("tamper_and_verify", func(t *testing.T) {
		ctx := context.Background()
		// Modify the installed skill to trigger hash mismatch
		skillPath := filepath.Join(installDir, "tamper-test-skill")
		tamperFile := filepath.Join(skillPath, "TAMPER.txt")
		if err := os.WriteFile(tamperFile, []byte("tampered content"), 0644); err != nil {
			t.Fatalf("Failed to create tamper file: %v", err)
		}

		// Run verify command
		cmd := exec.CommandContext(ctx, binaryPath, "verify")
		cmd.Dir = projectDir
		output, _ := cmd.CombinedOutput()

		// Verify command should still succeed (exit code 0) but show warning
		if cmd.ProcessState.ExitCode() != 0 {
			t.Errorf("Expected exit code 0 even with hash mismatch, got %d", cmd.ProcessState.ExitCode())
		}

		// Verify output contains hash mismatch warning
		outputStr := string(output)
		if !strings.Contains(strings.ToLower(outputStr), "mismatch") &&
			!strings.Contains(strings.ToLower(outputStr), "改ざん") &&
			!strings.Contains(strings.ToLower(outputStr), "failed") {
			t.Errorf("Expected hash mismatch warning in output, got: %s", outputStr)
		}
	})
}

// TestE2EErrorHandlingAndExitCodes tests error scenarios and exit codes
// Requirements: 12.2, 12.3, 12.5, 12.6
func TestE2EErrorHandlingAndExitCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Build the CLI binary
	workspaceDir := t.TempDir()
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	testCases := []struct {
		name         string
		wantInOutput string
		setupFunc    func(t *testing.T) string // Returns project directory
		command      []string
		wantExitCode int
	}{
		{
			name: "init_on_existing_config_should_fail",
			setupFunc: func(t *testing.T) string {
				ctx := context.Background()
				projectDir := filepath.Join(workspaceDir, "existing-config")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("Failed to create project directory: %v", err)
				}
				// Create existing config
				cmd := exec.CommandContext(ctx, binaryPath, "init")
				cmd.Dir = projectDir
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("First init failed: %v\nOutput: %s", err, output)
				}
				return projectDir
			},
			command:      []string{"init"},
			wantExitCode: 1,
			wantInOutput: "already exists",
		},
		{
			name: "install_without_config_should_fail",
			setupFunc: func(t *testing.T) string {
				projectDir := filepath.Join(workspaceDir, "no-config")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("Failed to create project directory: %v", err)
				}
				return projectDir
			},
			command:      []string{"install"},
			wantExitCode: 1,
			wantInOutput: "not found",
		},
		{
			name: "add_nonexistent_skill_source_should_fail",
			setupFunc: func(t *testing.T) string {
				ctx := context.Background()
				projectDir := filepath.Join(workspaceDir, "invalid-source")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("Failed to create project directory: %v", err)
				}
				cmd := exec.CommandContext(ctx, binaryPath, "init")
				cmd.Dir = projectDir
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("init failed: %v\nOutput: %s", err, output)
				}
				return projectDir
			},
			command: []string{"add", "invalid-skill",
				"--source", "invalid-source-type",
				"--url", "https://example.com",
			},
			wantExitCode: 1,
			wantInOutput: "invalid",
		},
		{
			name: "uninstall_nonexistent_skill_should_fail",
			setupFunc: func(t *testing.T) string {
				ctx := context.Background()
				projectDir := filepath.Join(workspaceDir, "no-skill")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("Failed to create project directory: %v", err)
				}
				cmd := exec.CommandContext(ctx, binaryPath, "init")
				cmd.Dir = projectDir
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("init failed: %v\nOutput: %s", err, output)
				}
				return projectDir
			},
			command:      []string{"uninstall", "nonexistent-skill"},
			wantExitCode: 1,
			wantInOutput: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			projectDir := tc.setupFunc(t)

			cmd := exec.CommandContext(ctx, binaryPath, tc.command...)
			cmd.Dir = projectDir
			output, _ := cmd.CombinedOutput()

			gotExitCode := cmd.ProcessState.ExitCode()
			if gotExitCode != tc.wantExitCode {
				t.Errorf("Expected exit code %d, got %d\nOutput: %s",
					tc.wantExitCode, gotExitCode, output)
			}

			outputStr := strings.ToLower(string(output))
			wantLower := strings.ToLower(tc.wantInOutput)
			if !strings.Contains(outputStr, wantLower) {
				t.Errorf("Expected output to contain %q, got: %s",
					tc.wantInOutput, output)
			}
		})
	}
}

// TestE2EVerboseMode tests that verbose flag produces detailed output
// Requirements: 12.4
func TestE2EVerboseMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup: Create temporary workspace
	workspaceDir := t.TempDir()
	projectDir := filepath.Join(workspaceDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Test: Run init with verbose flag
	t.Run("verbose_init", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "--verbose", "init")
		cmd.Dir = projectDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Fatalf("init command failed: %v\nStderr: %s", err, stderr.String())
		}

		// Verbose output should contain debug information
		output := stdout.String() + stderr.String()
		if len(output) == 0 {
			t.Error("Expected verbose output, got none")
		}

		// Check for verbose indicators (e.g., file paths, detailed messages)
		if !strings.Contains(strings.ToLower(output), ".skillspkg.toml") {
			t.Errorf("Expected verbose output to contain configuration details, got: %s", output)
		}
	})

	// Test: Run list with verbose flag
	t.Run("verbose_list", func(t *testing.T) {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, binaryPath, "--verbose", "list")
		cmd.Dir = projectDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Fatalf("list command failed: %v\nStderr: %s", err, stderr.String())
		}

		// Verbose output should be present
		output := stdout.String() + stderr.String()
		if len(output) == 0 {
			t.Error("Expected verbose output, got none")
		}
	})
}

// Helper: createTestGitRepo creates a test Git repository with a skill
func createTestGitRepo(t *testing.T, repoPath string) string {
	t.Helper()

	// Initialize Git repository
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Create a SKILL.md file
	skillMdPath := filepath.Join(repoPath, "SKILL.md")
	skillContent := `# Test Skill

This is a test skill for E2E testing.

## Usage

Test usage instructions.
`
	if writeErr := os.WriteFile(skillMdPath, []byte(skillContent), 0644); writeErr != nil {
		t.Fatalf("Failed to create SKILL.md: %v", writeErr)
	}

	// Add file to Git
	w, wErr := repo.Worktree()
	if wErr != nil {
		t.Fatalf("Failed to get worktree: %v", wErr)
	}

	if _, addErr := w.Add("SKILL.md"); addErr != nil {
		t.Fatalf("Failed to add file to Git: %v", addErr)
	}

	// Commit
	signature := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	commitHash, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: signature,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create tag v1.0.0
	tagName := "v1.0.0"
	_, err = repo.CreateTag(tagName, commitHash, &git.CreateTagOptions{
		Message: "Version 1.0.0",
		Tagger:  signature,
	})
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Configure remote (for display purposes)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoPath},
	})
	if err != nil && err != git.ErrRemoteExists {
		t.Fatalf("Failed to create remote: %v", err)
	}

	// Checkout the tag to ensure it's available
	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", tagName)),
	}); err != nil {
		t.Logf("Warning: Failed to checkout tag: %v", err)
	}

	// Return file:// URL
	return "file://" + repoPath
}

// Helper: buildCLIBinary builds the CLI binary for testing
func buildCLIBinary(t *testing.T, workspaceDir string) string {
	t.Helper()

	binaryPath := filepath.Join(workspaceDir, "skills-pkg-test")

	// Get the project root (3 levels up from internal/e2e)
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	return binaryPath
}
