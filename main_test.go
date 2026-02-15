package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestCLIExitCode tests that the CLI returns appropriate exit codes
func TestCLIExitCode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "help flag returns 0",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:     "invalid flag returns non-zero",
			args:     []string{"--invalid-flag"},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the binary for testing
			cmd := exec.Command("go", "build", "-o", "skills-pkg-test", ".")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build binary: %v", err)
			}
			defer os.Remove("skills-pkg-test")

			// Run with test args
			testCmd := exec.Command("./skills-pkg-test", tt.args...)
			err := testCmd.Run()

			var gotCode int
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					gotCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Unexpected error type: %v", err)
				}
			} else {
				gotCode = 0
			}

			if gotCode != tt.wantCode {
				t.Errorf("Exit code = %d, want %d", gotCode, tt.wantCode)
			}
		})
	}
}

// TestVerboseFlag tests that the --verbose flag is recognized
func TestVerboseFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "verbose flag is accepted",
			args:    []string{"--verbose", "--help"},
			wantErr: false,
		},
		{
			name:    "verbose short form is accepted",
			args:    []string{"-v", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the binary for testing
			cmd := exec.Command("go", "build", "-o", "skills-pkg-test", ".")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build binary: %v", err)
			}
			defer os.Remove("skills-pkg-test")

			// Run with test args
			testCmd := exec.Command("./skills-pkg-test", tt.args...)
			err := testCmd.Run()

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Exit code 0 is not an error for help
					if exitErr.ExitCode() != 0 {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			}
		})
	}
}

// TestSubcommandStructure tests that all expected subcommands are available
func TestSubcommandStructure(t *testing.T) {
	subcommands := []string{"init", "add", "install", "update", "list", "uninstall", "verify"}

	for _, subcmd := range subcommands {
		t.Run("subcommand_"+subcmd, func(t *testing.T) {
			// Build the binary for testing
			cmd := exec.Command("go", "build", "-o", "skills-pkg-test", ".")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build binary: %v", err)
			}
			defer os.Remove("skills-pkg-test")

			// Run with help to check if subcommand exists
			testCmd := exec.Command("./skills-pkg-test", subcmd, "--help")
			err := testCmd.Run()

			// Subcommand help should exit with code 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() != 0 {
						t.Errorf("Subcommand %s not properly defined, exit code: %d", subcmd, exitErr.ExitCode())
					}
				} else {
					t.Errorf("Unexpected error for subcommand %s: %v", subcmd, err)
				}
			}
		})
	}
}
