package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestCursorAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewCursorAgentAdapter()

	t.Run("cursor_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("cursor")
		if err != nil {
			t.Fatalf("ResolveAgentDir(cursor) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(cursor) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".cursor", "rules")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(cursor) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".cursor", "rules")
		if dir != expected {
			t.Errorf("ResolveAgentDir(cursor) = %q, want %q", dir, expected)
		}
	})

	t.Run("unsupported_agent", func(t *testing.T) {
		_, err := provider.ResolveAgentDir("unsupported")
		if err == nil {
			t.Error("ResolveAgentDir(unsupported) error = nil, want error")
		}
	})

	t.Run("empty_agent_name", func(t *testing.T) {
		_, err := provider.ResolveAgentDir("")
		if err == nil {
			t.Error("ResolveAgentDir(\"\") error = nil, want error")
		}
	})
}

func TestCursorAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewCursorAgentAdapter()

	name := provider.AgentName()
	if name != "cursor" {
		t.Errorf("AgentName() = %q, want \"cursor\"", name)
	}
}
