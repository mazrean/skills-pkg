package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// CopilotAgentAdapter provides directory resolution for the GitHub Copilot agent.
// It returns the default installation directory for Copilot agent when --agent flag is specified.
type CopilotAgentAdapter struct{}

// NewCopilotAgentAdapter creates a new GitHub Copilot agent adapter instance.
func NewCopilotAgentAdapter() port.AgentProvider {
	return &CopilotAgentAdapter{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For GitHub Copilot agent, it returns ~/.github/skills.
// Returns an error if the agent name is not "copilot" or if the home directory cannot be determined.
func (a *CopilotAgentAdapter) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "copilot" {
		return "", fmt.Errorf("unsupported agent: %s (only 'copilot' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".github", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *CopilotAgentAdapter) AgentName() string {
	return "copilot"
}
