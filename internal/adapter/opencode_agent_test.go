package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestOpenCodeAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewOpenCodeAgentAdapter()

	t.Run("opencode_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("opencode")
		if err != nil {
			t.Fatalf("ResolveAgentDir(opencode) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(opencode) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".config", "opencode", "skill")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(opencode) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".config", "opencode", "skill")
		if dir != expected {
			t.Errorf("ResolveAgentDir(opencode) = %q, want %q", dir, expected)
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

func TestOpenCodeAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewOpenCodeAgentAdapter()

	name := provider.AgentName()
	if name != "opencode" {
		t.Errorf("AgentName() = %q, want \"opencode\"", name)
	}
}
