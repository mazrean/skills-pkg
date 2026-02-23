package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Claude provides directory resolution for the Claude Code agent.
// It returns the default installation directory for Claude agent when --agent flag is specified.
// Requirements: 10.3, 10.4
type Claude struct{}

// NewClaude creates a new Claude agent adapter instance.
func NewClaude() port.AgentProvider {
	return &Claude{}
}

// ResolveAgentDir returns the default install directory for the agent.
// For Claude agent, it returns ~/.claude/skills.
// Returns an error if the agent name is not "claude" or if the home directory cannot be determined.
// Requirements: 10.3, 12.2, 12.3
func (a *Claude) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != "claude" {
		return "", fmt.Errorf("unsupported agent: %s (only 'claude' is supported by this adapter)", agentName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".claude", "skills"), nil
}

// AgentName returns the name of the agent this adapter supports.
// Requirements: 10.4
func (a *Claude) AgentName() string {
	return "claude"
}

// ProjectDir returns the project-level install directory for the agent.
func (a *Claude) ProjectDir() string {
	return ".claude/skills"
}
