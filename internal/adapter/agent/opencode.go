package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Opencode provides directory resolution for the OpenCode agent.
// It returns the default installation directory for OpenCode agent when --agent flag is specified.
type Opencode struct{}

// NewOpencode creates a new OpenCode agent adapter instance.
func NewOpencode() port.AgentProvider {
	return &Opencode{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For OpenCode agent, it returns ~/.config/opencode/skill.
// Returns an error if the agent name is not "opencode" or if the home directory cannot be determined.
func (a *Opencode) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "opencode" {
		return "", fmt.Errorf("unsupported agent: %s (only 'opencode' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "opencode", "skill"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Opencode) AgentName() string {
	return "opencode"
}

// ProjectDir returns the project-level install directory for the agent.
func (a *Opencode) ProjectDir() string {
	return ".agents/skills"
}
