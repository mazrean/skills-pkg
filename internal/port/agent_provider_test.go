package port_test

import (
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestAgentProviderInterface verifies that the AgentProvider interface contract
// can be satisfied by a mock implementation.
// Requirements: 10.4
func TestAgentProviderInterface(t *testing.T) {
	t.Run("interface_contract", func(t *testing.T) {
		// Verify that a mock implementation satisfies the interface
		var _ port.AgentProvider = &mockAgentProvider{}
	})
}

// TestAgentProviderResolveAgentDir tests directory resolution contract.
// Requirements: 10.3, 10.4
func TestAgentProviderResolveAgentDir(t *testing.T) {
	provider := &mockAgentProvider{}

	t.Run("valid_agent", func(t *testing.T) {
		dir, err := provider.ResolveAgentDir("claude")
		if err != nil {
			t.Errorf("ResolveAgentDir() error = %v, want nil", err)
		}
		if dir == "" {
			t.Error("ResolveAgentDir() returned empty directory")
		}
	})

	t.Run("agent_name", func(t *testing.T) {
		name := provider.AgentName()
		if name == "" {
			t.Error("AgentName() returned empty string")
		}
	})
}

// mockAgentProvider is a mock implementation of AgentProvider for testing.
type mockAgentProvider struct{}

func (m *mockAgentProvider) ResolveAgentDir(agentName string) (string, error) {
	return "/mock/.claude/skills", nil
}

func (m *mockAgentProvider) AgentName() string {
	return "mock"
}
