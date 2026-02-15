package port

// AgentProvider is the abstraction interface for resolving agent-specific directories.
// It provides default installation directory paths for different coding agents.
// Requirements: 10.3, 10.4
type AgentProvider interface {
	// ResolveAgentDir returns the default install directory for the agent.
	// Used only when --agent flag is specified during init.
	// Returns an error if the agent name is not supported.
	// Requirements: 1.3, 10.3, 10.4
	ResolveAgentDir(agentName string) (string, error)

	// AgentName returns the name of the agent (e.g., "claude", "codex").
	// Requirements: 10.4
	AgentName() string
}
