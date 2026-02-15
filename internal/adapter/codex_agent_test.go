package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

// TestCodexAgentAdapter_ResolveAgentDir tests directory resolution for Codex agent.
func TestCodexAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewCodexAgentAdapter()

	t.Run("codex_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("codex")
		if err != nil {
			t.Fatalf("ResolveAgentDir(codex) error = %v, want nil", err)
		}

		// Verify it returns a valid path ending with .codex/skills
		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(codex) = %q, want absolute path", dir)
		}

		// Verify it ends with .codex/skills
		expectedSuffix := filepath.Join(".codex", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(codex) = %q, want path ending with %q", dir, expectedSuffix)
		}

		// Verify it starts with home directory
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".codex", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(codex) = %q, want %q", dir, expected)
		}
	})

	t.Run("unsupported_agent", func(t *testing.T) {
		_, err := provider.ResolveAgentDir("unsupported")
		if err == nil {
			t.Error("ResolveAgentDir(unsupported) error = nil, want error")
		}

		// Error message should mention the unsupported agent
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("ResolveAgentDir(unsupported) error message is empty")
		}
	})

	t.Run("empty_agent_name", func(t *testing.T) {
		_, err := provider.ResolveAgentDir("")
		if err == nil {
			t.Error("ResolveAgentDir(\"\") error = nil, want error")
		}
	})
}

// TestCodexAgentAdapter_AgentName tests agent name retrieval.
func TestCodexAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewCodexAgentAdapter()

	name := provider.AgentName()
	if name != "codex" {
		t.Errorf("AgentName() = %q, want \"codex\"", name)
	}
}
