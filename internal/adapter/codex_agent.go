package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// CodexAgentAdapter provides directory resolution for the Codex CLI agent.
// It returns the default installation directory for Codex agent when --agent flag is specified.
type CodexAgentAdapter struct{}

// NewCodexAgentAdapter creates a new Codex agent adapter instance.
func NewCodexAgentAdapter() port.AgentProvider {
	return &CodexAgentAdapter{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Codex agent, it returns ~/.codex/skills.
// Returns an error if the agent name is not "codex" or if the home directory cannot be determined.
func (a *CodexAgentAdapter) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "codex" {
		return "", fmt.Errorf("unsupported agent: %s (only 'codex' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".codex", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *CodexAgentAdapter) AgentName() string {
	return "codex"
}
