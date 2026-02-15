package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Factory provides directory resolution for the Factory agent.
// It returns the default installation directory for Factory agent when --agent flag is specified.
type Factory struct{}

// NewFactory creates a new Factory agent adapter instance.
func NewFactory() port.AgentProvider {
	return &Factory{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Factory agent, it returns ~/.factory/skills.
// Returns an error if the agent name is not "factory" or if the home directory cannot be determined.
func (a *Factory) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "factory" {
		return "", fmt.Errorf("unsupported agent: %s (only 'factory' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".factory", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Factory) AgentName() string {
	return "factory"
}
