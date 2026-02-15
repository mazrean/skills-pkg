package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

func TestCopilotAgentAdapter_ResolveAgentDir(t *testing.T) {
	provider := adapter.NewCopilotAgentAdapter()

	t.Run("copilot_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("copilot")
		if err != nil {
			t.Fatalf("ResolveAgentDir(copilot) error = %v, want nil", err)
		}

		if !filepath.IsAbs(dir) {
			t.Errorf("ResolveAgentDir(copilot) = %q, want absolute path", dir)
		}

		expectedSuffix := filepath.Join(".github", "skills")
		if !hasPathSuffix(dir, expectedSuffix) {
			t.Errorf("ResolveAgentDir(copilot) = %q, want path ending with %q", dir, expectedSuffix)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}
		expected := filepath.Join(home, ".github", "skills")
		if dir != expected {
			t.Errorf("ResolveAgentDir(copilot) = %q, want %q", dir, expected)
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

func TestCopilotAgentAdapter_AgentName(t *testing.T) {
	provider := adapter.NewCopilotAgentAdapter()

	name := provider.AgentName()
	if name != "copilot" {
		t.Errorf("AgentName() = %q, want \"copilot\"", name)
	}
}
