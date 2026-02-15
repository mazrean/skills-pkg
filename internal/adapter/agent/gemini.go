package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Gemini provides directory resolution for the Gemini CLI agent.
// It returns the default installation directory for Gemini agent when --agent flag is specified.
type Gemini struct{}

// NewGemini creates a new Gemini CLI agent adapter instance.
func NewGemini() port.AgentProvider {
	return &Gemini{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Gemini CLI agent, it returns ~/.gemini/skills.
// Returns an error if the agent name is not "gemini" or if the home directory cannot be determined.
func (a *Gemini) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "gemini" {
		return "", fmt.Errorf("unsupported agent: %s (only 'gemini' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".gemini", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
func (a *Gemini) AgentName() string {
	return "gemini"
}
