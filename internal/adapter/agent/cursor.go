package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Cursor provides directory resolution for the Cursor editor agent.
// It returns the default installation directory for Cursor agent when --agent flag is specified.
type Cursor struct{}

// NewCursor creates a new Cursor agent adapter instance.
func NewCursor() port.AgentProvider {
	return &Cursor{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Cursor agent, it returns ~/.cursor/rules.
// Returns an error if the agent name is not "cursor" or if the home directory cannot be determined.
func (a *Cursor) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "cursor" {
		return "", fmt.Errorf("unsupported agent: %s (only 'cursor' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".cursor", "rules"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Cursor) AgentName() string {
	return "cursor"
}
