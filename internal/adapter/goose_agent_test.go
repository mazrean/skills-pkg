package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestGooseAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewGooseAgentAdapter()

	t.Run("goose_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("goose")
		if err != nil {
			t.Fatalf("ResolveAgentDir(goose) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(goose) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".config", "goose", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(goose) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".config", "goose", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(goose) = %q, want %q", dir, expected)
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

func TestGooseAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewGooseAgentAdapter()

	name := provider.AgentName()
	if name != "goose" {
		t.Errorf("AgentName() = %q, want \"goose\"", name)
	}
}
