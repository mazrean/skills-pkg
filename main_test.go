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
