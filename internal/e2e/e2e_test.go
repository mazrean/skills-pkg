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
	testRepoURL := createTestGitRepo(t, testRepoDir, "test-skill")

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Setup: Define install targets
	installDir1 := filepath.Join(workspaceDir, "agent1", "skills")
	installDir2 := filepath.Join(workspaceDir, "agent2", "skills")

	tests := []struct {
		name           string
		validateOutput func(t *testing.T, output []byte, exitCode int)
		commandArgs    []string
	}{
		{
			name: "init",
			commandArgs: []string{"init",
				"--install-dir", installDir1,
				"--install-dir", installDir2,
			},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				configPath := filepath.Join(projectDir, ".skillspkg.toml")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Configuration file was not created at %s", configPath)
				}
			},
		},
		{
			name: "add",
			commandArgs: []string{"add",
				"test-skill",
				"--source", "git",
				"--url", testRepoURL,
				"--version", "v1.0.0",
			},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
			},
		},
		{
			name:        "install",
			commandArgs: []string{"install", "test-skill"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				skillPath1 := filepath.Join(installDir1, "test-skill")
				skillPath2 := filepath.Join(installDir2, "test-skill")

				if _, err := os.Stat(skillPath1); os.IsNotExist(err) {
					t.Errorf("Skill was not installed in %s", skillPath1)
				}
				if _, err := os.Stat(skillPath2); os.IsNotExist(err) {
					t.Errorf("Skill was not installed in %s", skillPath2)
				}
			},
		},
		{
			name:        "verify",
			commandArgs: []string{"verify"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				outputStr := string(output)
				hasSuccessIndicator := strings.Contains(outputStr, "Successful:") ||
					strings.Contains(outputStr, "成功") ||
					strings.Contains(strings.ToLower(outputStr), "success")
				if !hasSuccessIndicator {
					t.Errorf("Expected success message in verify output, got: %s", outputStr)
				}
				if strings.Contains(outputStr, "Failed: 0") == false &&
					strings.Contains(outputStr, "失敗: 0") == false {
					t.Errorf("Expected no failures in verify output, got: %s", outputStr)
				}
			},
		},
		{
			name:        "list",
			commandArgs: []string{"list"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				outputStr := string(output)
				if !strings.Contains(outputStr, "test-skill") {
					t.Errorf("Expected skill name in list output, got: %s", outputStr)
				}
			},
		},
		{
			name:        "uninstall",
			commandArgs: []string{"uninstall", "test-skill"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				skillPath1 := filepath.Join(installDir1, "test-skill")
				skillPath2 := filepath.Join(installDir2, "test-skill")

				if _, err := os.Stat(skillPath1); !os.IsNotExist(err) {
					t.Errorf("Skill was not removed from %s", skillPath1)
				}
				if _, err := os.Stat(skillPath2); !os.IsNotExist(err) {
					t.Errorf("Skill was not removed from %s", skillPath2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, binaryPath, tt.commandArgs...)
			cmd.Dir = projectDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s command failed: %v\nOutput: %s", tt.name, err, output)
			}

			tt.validateOutput(t, output, cmd.ProcessState.ExitCode())
		})
	}
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
	testRepoURL := createTestGitRepo(t, testRepoDir, "multi-agent-skill")

	// Setup: Build the CLI binary
	binaryPath := buildCLIBinary(t, workspaceDir)
	defer func() { _ = os.Remove(binaryPath) }()

	// Setup: Define install targets for multiple agents
	agentDirs := []string{
		filepath.Join(workspaceDir, "claude", "skills"),
		filepath.Join(workspaceDir, "codex", "skills"),
		filepath.Join(workspaceDir, "custom", "skills"),
	}

	tests := []struct {
		setupFunc      func(t *testing.T)
		validateOutput func(t *testing.T, output []byte, exitCode int)
		commandFunc    func(ctx context.Context) *exec.Cmd
		name           string
	}{
		{
			name:      "init_multiple_agents",
			setupFunc: func(t *testing.T) {},
			commandFunc: func(ctx context.Context) *exec.Cmd {
				args := []string{"init"}
				for _, dir := range agentDirs {
					args = append(args, "--install-dir", dir)
				}
				cmd := exec.CommandContext(ctx, binaryPath, args...)
				cmd.Dir = projectDir
				return cmd
			},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
			},
		},
		{
			name: "add_and_install",
			setupFunc: func(t *testing.T) {
				ctx := context.Background()
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
			},
			commandFunc: func(ctx context.Context) *exec.Cmd {
				cmd := exec.CommandContext(ctx, binaryPath, "install")
				cmd.Dir = projectDir
				return cmd
			},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				for _, agentDir := range agentDirs {
					skillPath := filepath.Join(agentDir, "multi-agent-skill")
					if _, err := os.Stat(skillPath); os.IsNotExist(err) {
						t.Errorf("Skill was not installed in %s", agentDir)
					}

					skillMdPath := filepath.Join(skillPath, "SKILL.md")
					if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
						t.Errorf("SKILL.md was not found in %s", skillPath)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc(t)

			ctx := context.Background()
			cmd := tt.commandFunc(ctx)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s command failed: %v\nOutput: %s", tt.name, err, output)
			}

			tt.validateOutput(t, output, cmd.ProcessState.ExitCode())
		})
	}
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
	testRepoURL := createTestGitRepo(t, testRepoDir, "tamper-test-skill")

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

	tests := []struct {
		name            string
		validateOutput  func(t *testing.T, output string)
		commandArgs     []string
		requireInitFile bool
	}{
		{
			name:            "verbose_init",
			commandArgs:     []string{"--verbose", "init"},
			requireInitFile: false,
			validateOutput: func(t *testing.T, output string) {
				if len(output) == 0 {
					t.Error("Expected verbose output, got none")
				}
				if !strings.Contains(strings.ToLower(output), ".skillspkg.toml") {
					t.Errorf("Expected verbose output to contain configuration details, got: %s", output)
				}
			},
		},
		{
			name:            "verbose_list",
			commandArgs:     []string{"--verbose", "list"},
			requireInitFile: true,
			validateOutput: func(t *testing.T, output string) {
				if len(output) == 0 {
					t.Error("Expected verbose output, got none")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, binaryPath, tt.commandArgs...)
			cmd.Dir = projectDir
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				t.Fatalf("%s command failed: %v\nStderr: %s", tt.name, err, stderr.String())
			}

			output := stdout.String() + stderr.String()
			tt.validateOutput(t, output)
		})
	}
}

// TestE2ERealRepository tests installing a skill from the actual vercel-labs/agent-skills repository
// This test verifies that the SubDir functionality works with real-world repositories
func TestE2ERealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Skip if network is not available
	if os.Getenv("SKIP_NETWORK_TESTS") == "true" {
		t.Skip("Skipping network-dependent test")
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

	// Setup: Define install target
	installDir := filepath.Join(workspaceDir, "skills")

	// Initialize project
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binaryPath, "init", "--install-dir", installDir)
	cmd.Dir = projectDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init command failed: %v\nOutput: %s", err, output)
	}

	tests := []struct {
		name           string
		validateOutput func(t *testing.T, output []byte, exitCode int)
		commandArgs    []string
	}{
		{
			name: "add_real_skill",
			commandArgs: []string{"add",
				"vercel-deploy",
				"--source", "git",
				"--url", "https://github.com/vercel-labs/agent-skills.git",
				"--version", "main",
				"--sub-dir", "skills/claude.ai/vercel-deploy-claimable",
			},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
			},
		},
		{
			name:        "install_real_skill",
			commandArgs: []string{"install", "vercel-deploy"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				skillPath := filepath.Join(installDir, "vercel-deploy")
				if _, err := os.Stat(skillPath); os.IsNotExist(err) {
					t.Errorf("Skill was not installed at %s", skillPath)
				}

				skillMdPath := filepath.Join(skillPath, "SKILL.md")
				if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
					t.Errorf("SKILL.md was not found in installed skill at %s", skillMdPath)
				}

				topReadme := filepath.Join(skillPath, "README.md")
				if _, err := os.Stat(topReadme); !os.IsNotExist(err) {
					t.Errorf("Top-level README.md should not exist in skill directory (found at %s), only the subdirectory should be installed", topReadme)
				}
			},
		},
		{
			name:        "verify_real_skill",
			commandArgs: []string{"verify"},
			validateOutput: func(t *testing.T, output []byte, exitCode int) {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
				outputStr := string(output)
				if strings.Contains(outputStr, "⚠ WARNING") || strings.Contains(outputStr, "Failed") {
					t.Logf("Verification output: %s", outputStr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.commandArgs...)
			cmd.Dir = projectDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s command failed: %v\nOutput: %s", tt.name, err, output)
			}

			tt.validateOutput(t, output, cmd.ProcessState.ExitCode())
		})
	}
}

// Helper: createTestGitRepo creates a test Git repository with a skill
// The skillName parameter specifies which skill directory to create under skills/
func createTestGitRepo(t *testing.T, repoPath string, skillName string) string {
	t.Helper()

	// Initialize Git repository
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Create skills directory structure
	skillDir := filepath.Join(repoPath, "skills", skillName)
	if mkdirErr := os.MkdirAll(skillDir, 0755); mkdirErr != nil {
		t.Fatalf("Failed to create skill directory: %v", mkdirErr)
	}

	// Create a SKILL.md file in the skill directory
	skillMdPath := filepath.Join(skillDir, "SKILL.md")
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

	if _, addErr := w.Add("."); addErr != nil {
		t.Fatalf("Failed to add files to Git: %v", addErr)
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
