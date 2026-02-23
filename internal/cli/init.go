package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/adapter/agent"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

const (
	// defaultConfigPath is the default path to the .skillspkg.toml configuration file
	defaultConfigPath = ".skillspkg.toml"
)

// InitCmd represents the init command
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 12.1, 12.2, 12.3, 12.4
type InitCmd struct {
	Agent      []string `help:"Agent name to use default directory (can be specified multiple times)" short:"a" enum:"claude,claude-code,codex,cursor,copilot,github-copilot,goose,opencode,gemini,gemini-cli,amp,kimi-cli,replit,universal,factory,droid,antigravity,augment,openclaw,cline,codebuddy,command-code,continue,cortex,crush,junie,iflow-cli,kilo,kiro-cli,kode,mcpjam,mistral-vibe,mux,openhands,pi,qoder,qwen-code,roo,trae,trae-cn,windsurf,zencoder,neovate,pochi,adal"`
	InstallDir []string `help:"Custom install directory (can be specified multiple times)" short:"d"`
	Global     bool     `help:"Use user-level directory instead of project-level directory (requires --agent)" short:"g" default:"false"`
}

// Run executes the init command
// This method initializes a new .skillspkg.toml configuration file with the specified install directories.
// It handles custom install directories (--install-dir) and agent-specific directories (--agent).
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 12.1, 12.2, 12.3, 12.4
func (c *InitCmd) Run(ctx *kong.Context) error {
	// Access verbose flag from the parsed CLI model using reflection
	verbose := false
	if model := ctx.Model; model != nil && model.Target.IsValid() {
		// Get the "Verbose" field from the CLI struct
		if verboseField := model.Target.FieldByName("Verbose"); verboseField.IsValid() && verboseField.Kind() == reflect.Bool {
			verbose = verboseField.Bool()
		}
	}

	return c.run(defaultConfigPath, verbose)
}

// run is the internal implementation that can be called from tests with custom parameters
func (c *InitCmd) run(configPath string, verbose bool) error {
	// Create logger with verbose setting (requirement 12.4)
	logger := NewLogger(verbose)

	// Display progress information (requirement 12.1)
	logger.Info("Initializing project with .skillspkg.toml")

	// Build install targets list (requirements 1.2, 1.3)
	installTargets, err := c.buildInstallTargets(logger)
	if err != nil {
		// Distinguish error types and provide cause and recommended action (requirements 12.2, 12.3)
		logger.Error("Failed to build install targets: %v", err)
		return err
	}

	logger.Verbose("Install targets: %v", installTargets)

	// Create ConfigManager
	configManager := domain.NewConfigManager(configPath)

	// Initialize configuration file (requirement 1.1, 1.5)
	if err := configManager.Initialize(context.Background(), installTargets); err != nil {
		// Handle different error types with appropriate messages (requirements 12.2, 12.3)
		if errors.Is(err, domain.ErrConfigExists) {
			// Configuration file already exists (requirement 1.4)
			logger.Error("Configuration file already exists at %s", configPath)
			logger.Error("Remove the existing file or use a different path")
			return err
		}

		// File system error - distinguish and report (requirements 12.2, 12.3)
		logger.Error("Failed to create configuration file: %v", err)
		logger.Error("Check file permissions and try again")
		return err
	}

	// Success message (requirement 12.1)
	logger.Info("Successfully initialized .skillspkg.toml")
	if len(installTargets) > 0 {
		logger.Info("Install targets:")
		for _, target := range installTargets {
			logger.Info("  - %s", target)
		}
	} else {
		logger.Info("No install targets configured. Add them later with 'skills-pkg add'")
	}

	return nil
}

// buildInstallTargets constructs the list of install target directories
// from custom directories (--install-dir) and agent-specific directories (--agent).
// Default behavior: Use project-level directory (./.skills or ./.{agent}/skills).
// With --global flag: Use user-level directory (e.g., ~/.claude/skills).
// Requirements: 1.2, 1.3, 10.3, 10.4
func (c *InitCmd) buildInstallTargets(logger *Logger) ([]string, error) {
	installTargets := make([]string, 0)

	// Add custom install directories (requirement 1.2)
	if len(c.InstallDir) > 0 {
		logger.Verbose("Adding custom install directories: %v", c.InstallDir)
		installTargets = append(installTargets, c.InstallDir...)
	}

	// Add agent-specific directories if --agent is specified (requirement 1.3)
	if len(c.Agent) > 0 {
		for _, agent := range c.Agent {
			logger.Verbose("Resolving agent directory for: %s (global=%v)", agent, c.Global)

			agentProvider, err := c.getAgentProvider(agent)
			if err != nil {
				return nil, fmt.Errorf("failed to get agent provider for %s: %w", agent, err)
			}

			var agentDir string
			if c.Global {
				// Use AgentProvider to resolve user-level directory (requirements 10.3, 10.4)
				agentDir, err = agentProvider.ResolveAgentDir(agent)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve agent directory for %s: %w", agent, err)
				}
				logger.Verbose("Resolved user-level agent directory: %s", agentDir)
			} else {
				// Use agent-specific project-level directory
				agentDir = agentProvider.ProjectDir()
				logger.Verbose("Using project-level agent directory: %s", agentDir)
			}

			installTargets = append(installTargets, agentDir)
		}
	}

	// If no install targets specified, use default project-level directory
	if len(installTargets) == 0 {
		defaultDir := "./.skills"
		logger.Verbose("Using default project-level directory: %s", defaultDir)
		installTargets = append(installTargets, defaultDir)
	}

	return installTargets, nil
}

// getAgentProvider returns the appropriate AgentProvider based on the agent name.
func (c *InitCmd) getAgentProvider(agentName string) (port.AgentProvider, error) {
	switch agentName {
	case "claude":
		return agent.NewClaude(), nil
	case "claude-code":
		return agent.NewClaudeCode(), nil
	case "codex":
		return agent.NewCodex(), nil
	case "cursor":
		return agent.NewCursor(), nil
	case "copilot":
		return agent.NewCopilot(), nil
	case "github-copilot":
		return agent.NewGithubCopilot(), nil
	case "goose":
		return agent.NewGoose(), nil
	case "opencode":
		return agent.NewOpencode(), nil
	case "gemini":
		return agent.NewGemini(), nil
	case "gemini-cli":
		return agent.NewGeminiCLI(), nil
	case "amp":
		return agent.NewAmp(), nil
	case "kimi-cli":
		return agent.NewKimiCLI(), nil
	case "replit":
		return agent.NewReplit(), nil
	case "universal":
		return agent.NewUniversal(), nil
	case "factory":
		return agent.NewFactory(), nil
	case "droid":
		return agent.NewDroid(), nil
	case "antigravity":
		return agent.NewAntigravity(), nil
	case "augment":
		return agent.NewAugment(), nil
	case "openclaw":
		return agent.NewOpenclaw(), nil
	case "cline":
		return agent.NewCline(), nil
	case "codebuddy":
		return agent.NewCodebuddy(), nil
	case "command-code":
		return agent.NewCommandCode(), nil
	case "continue":
		return agent.NewContinueAgent(), nil
	case "cortex":
		return agent.NewCortex(), nil
	case "crush":
		return agent.NewCrush(), nil
	case "junie":
		return agent.NewJunie(), nil
	case "iflow-cli":
		return agent.NewIflowCLI(), nil
	case "kilo":
		return agent.NewKilo(), nil
	case "kiro-cli":
		return agent.NewKiroCLI(), nil
	case "kode":
		return agent.NewKode(), nil
	case "mcpjam":
		return agent.NewMCPJam(), nil
	case "mistral-vibe":
		return agent.NewMistralVibe(), nil
	case "mux":
		return agent.NewMux(), nil
	case "openhands":
		return agent.NewOpenhands(), nil
	case "pi":
		return agent.NewPi(), nil
	case "qoder":
		return agent.NewQoder(), nil
	case "qwen-code":
		return agent.NewQwenCode(), nil
	case "roo":
		return agent.NewRoo(), nil
	case "trae":
		return agent.NewTrae(), nil
	case "trae-cn":
		return agent.NewTraeCN(), nil
	case "windsurf":
		return agent.NewWindsurf(), nil
	case "zencoder":
		return agent.NewZencoder(), nil
	case "neovate":
		return agent.NewNeovate(), nil
	case "pochi":
		return agent.NewPochi(), nil
	case "adal":
		return agent.NewAdal(), nil
	default:
		return nil, fmt.Errorf("unsupported agent: %s", agentName)
	}
}
