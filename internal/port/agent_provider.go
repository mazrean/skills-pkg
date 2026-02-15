package port

// AgentProvider is the abstraction interface for resolving agent-specific directories.
// It provides default installation directory paths for different coding agents.
// This interface is used only when --agent flag is specified during init.
// The actual install targets are managed in the Config's install_targets field,
// which is the single source of truth for installation directories.
// Requirements: 10.3, 10.4
type AgentProvider interface {
	// ResolveAgentDir returns the default install directory for the agent.
	// Used only when --agent flag is specified during init.
	// Returns an error if the agent is not supported.
	ResolveAgentDir(agentName string) (string, error)

	// AgentName returns the name of the agent (e.g., "claude", "codex").
	AgentName() string
}
