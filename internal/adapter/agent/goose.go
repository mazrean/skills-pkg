package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Goose provides directory resolution for the Goose agent.
// It returns the default installation directory for Goose agent when --agent flag is specified.
type Goose struct{}

// NewGoose creates a new Goose agent adapter instance.
func NewGoose() port.AgentProvider {
	return &Goose{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Goose agent, it returns ~/.config/goose/skills.
// Returns an error if the agent name is not "goose" or if the home directory cannot be determined.
func (a *Goose) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "goose" {
		return "", fmt.Errorf("unsupported agent: %s (only 'goose' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "goose", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Goose) AgentName() string {
	return "goose"
}

// ProjectDir returns the project-level install directory for the agent.
func (a *Goose) ProjectDir() string {
	return ".goose/skills"
}
