package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestGeminiAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewGeminiAgentAdapter()

	t.Run("gemini_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("gemini")
		if err != nil {
			t.Fatalf("ResolveAgentDir(gemini) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(gemini) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".gemini", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(gemini) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".gemini", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(gemini) = %q, want %q", dir, expected)
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

func TestGeminiAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewGeminiAgentAdapter()

	name := provider.AgentName()
	if name != "gemini" {
		t.Errorf("AgentName() = %q, want \"gemini\"", name)
	}
}
