// Package agent provides implementations of AgentProvider interface for various coding agents.
// It supports multiple coding agents including Claude, Codex, Cursor, Copilot, Goose, Opencode, Gemini, Amp, and Factory.
package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Amp provides directory resolution for the AMP agent.
// It returns the default installation directory for AMP agent when --agent flag is specified.
type Amp struct{}

// NewAmp creates a new AMP agent adapter instance.
func NewAmp() port.AgentProvider {
	return &Amp{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For AMP agent, it returns ~/.config/agents/skills.
// Returns an error if the agent name is not "amp" or if the home directory cannot be determined.
func (a *Amp) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "amp" {
		return "", fmt.Errorf("unsupported agent: %s (only 'amp' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "agents", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Amp) AgentName() string {
	return "amp"
}

// ProjectDir returns the project-level install directory for the agent.
func (a *Amp) ProjectDir() string {
	return ".agents/skills"
}
