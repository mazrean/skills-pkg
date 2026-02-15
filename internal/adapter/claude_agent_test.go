package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

// TestClaudeAgentAdapter_ResolveAgentDir tests directory resolution for Claude agent.
// Requirements: 10.3, 10.4
func TestClaudeAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewClaudeAgentAdapter()

	t.Run("claude_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("claude")
		if err != nil {
			t.Fatalf("ResolveAgentDir(claude) error = %v, want nil", err)
		}

		// Verify it returns a valid path ending with .claude/skills
		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(claude) = %q, want absolute path", dir)
		}

		// Verify it ends with .claude/skills
		expectedSuffix := filepath.Join(".claude", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(claude) = %q, want path ending with %q", dir, expectedSuffix)
		}

		// Verify it starts with home directory
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".claude", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(claude) = %q, want %q", dir, expected)
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

// TestClaudeAgentAdapter_AgentName tests agent name retrieval.
// Requirements: 10.4
func TestClaudeAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewClaudeAgentAdapter()

	name := provider.AgentName()
	if name != "claude" {
		t.Errorf("AgentName() = %q, want \"claude\"", name)
	}
}

// hasPathSuffix checks if the path ends with the specified suffix.
func hasPathSuffix(path, suffix string) bool {
	cleanPath := filepath.Clean(path)
	cleanSuffix := filepath.Clean(suffix)

	// Split path into components
	pathParts := splitPath(cleanPath)
	suffixParts := splitPath(cleanSuffix)

	if len(pathParts) < len(suffixParts) {
		return false
	}

	// Compare suffix parts from the end
	for i := 0; i < len(suffixParts); i++ {
		if pathParts[len(pathParts)-len(suffixParts)+i] != suffixParts[i] {
			return false
		}
	}

	return true
}

// splitPath splits a file path into components.
func splitPath(path string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		if dir == "" || dir == "/" {
			if dir == "/" {
				parts = append([]string{"/"}, parts...)
			}
			break
		}
		path = filepath.Clean(dir)
	}
	return parts
}
