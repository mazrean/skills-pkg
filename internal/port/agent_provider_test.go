package port_test

import (
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestAgentProviderInterface verifies that the AgentProvider interface contract
// can be satisfied by a mock implementation.
// Requirements: 10.4
func TestAgentProviderInterface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "interface_contract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that a mock implementation satisfies the interface
			var _ port.AgentProvider = &mockAgentProvider{}
		})
	}
}

// TestAgentProviderResolveAgentDir tests directory resolution contract.
// Requirements: 10.3, 10.4
func TestAgentProviderResolveAgentDir(t *testing.T) {
	tests := []struct {
		name           string
		provider       port.AgentProvider
		agentName      string
		wantErr        bool
		wantNonEmptyDir bool
	}{
		{
			name:           "valid_agent",
			provider:       &mockAgentProvider{},
			agentName:      "claude",
			wantErr:        false,
			wantNonEmptyDir: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := tt.provider.ResolveAgentDir(tt.agentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAgentDir() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNonEmptyDir && dir == "" {
				t.Error("ResolveAgentDir() returned empty directory")
			}
		})
	}
}

// TestAgentProviderAgentName tests AgentName method.
// Requirements: 10.3, 10.4
func TestAgentProviderAgentName(t *testing.T) {
	tests := []struct {
		name         string
		provider     port.AgentProvider
		wantNonEmpty bool
	}{
		{
			name:         "agent_name",
			provider:     &mockAgentProvider{},
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.provider.AgentName()
			if tt.wantNonEmpty && name == "" {
				t.Error("AgentName() returned empty string")
			}
		})
	}
}

// mockAgentProvider is a mock implementation of AgentProvider for testing.
type mockAgentProvider struct{}

func (m *mockAgentProvider) ResolveAgentDir(agentName string) (string, error) {
	return "/mock/.claude/skills", nil
}

func (m *mockAgentProvider) AgentName() string {
	return "mock"
}
