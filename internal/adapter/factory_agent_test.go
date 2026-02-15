package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestFactoryAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewFactoryAgentAdapter()

	t.Run("factory_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("factory")
		if err != nil {
			t.Fatalf("ResolveAgentDir(factory) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(factory) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".factory", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(factory) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".factory", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(factory) = %q, want %q", dir, expected)
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

func TestFactoryAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewFactoryAgentAdapter()

	name := provider.AgentName()
	if name != "factory" {
		t.Errorf("AgentName() = %q, want \"factory\"", name)
	}
}
