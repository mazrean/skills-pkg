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

// AddInstallTargetCmd represents the add-install-target command
type AddInstallTargetCmd struct {
	Target []string `arg:"" optional:"" help:"Install target directory path (can be specified multiple times)"`
	Agent  []string `help:"Agent name to use default directory (can be specified multiple times)" short:"a" enum:"claude,codex,cursor,copilot,goose,opencode,gemini,amp,factory"`
	Global bool     `help:"Use user-level directory instead of project-level directory (requires --agent)" short:"g" default:"false"`
}

// Run executes the add-install-target command
func (c *AddInstallTargetCmd) Run(ctx *kong.Context) error {
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

func (c *AddInstallTargetCmd) run(configPath string, verbose bool) error {
	logger := NewLogger(verbose)

	logger.Info("Adding install target '%s' to configuration", c.Target)
	logger.Verbose("Config path: %s", configPath)

	// Build the list of targets to add
	targets, err := c.buildTargets(logger)
	if err != nil {
		logger.Error("Failed to build install targets: %v", err)
		return err
	}

	logger.Verbose("Install targets to add: %v", targets)

	configManager := domain.NewConfigManager(configPath)

	for _, target := range targets {
		logger.Info("Adding install target '%s' to configuration", target)

		if err := configManager.AddInstallTarget(context.Background(), target); err != nil {
			if err, ok := errors.AsType[*domain.ErrorConfigNotFound](err); ok {
				logger.Error("Configuration file not found at %s", err.Path)
				logger.Error("Run 'skills-pkg init' to create a configuration file")
				return err
			}

			if err, ok := errors.AsType[*domain.ErrorInstallTargetExists](err); ok {
				logger.Error("Install target '%s' already exists in configuration", err.Target)
				return err
			}

			logger.Error("Failed to add install target '%s': %v", target, err)
			return err
		}

		logger.Info("Successfully added install target '%s'", target)
	}

	return nil
}

// buildTargets constructs the list of install target directories
// from positional arguments and/or agent-specific directories.
func (c *AddInstallTargetCmd) buildTargets(logger *Logger) ([]string, error) {
	targets := make([]string, 0)

	// Add directly specified targets
	if len(c.Target) > 0 {
		logger.Verbose("Adding specified targets: %v", c.Target)
		targets = append(targets, c.Target...)
	}

	// Add agent-specific directories if --agent is specified
	if len(c.Agent) > 0 {
		for _, agentName := range c.Agent {
			logger.Verbose("Resolving agent directory for: %s (global=%v)", agentName, c.Global)

			if c.Global {
				agentProvider, err := c.getAgentProvider(agentName)
				if err != nil {
					return nil, fmt.Errorf("failed to get agent provider for %s: %w. Supported agents: claude, codex, cursor, copilot, goose, opencode, gemini, amp, factory", agentName, err)
				}

				agentDir, err := agentProvider.ResolveAgentDir(agentName)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve agent directory for %s: %w", agentName, err)
				}

				logger.Verbose("Resolved user-level agent directory: %s", agentDir)
				targets = append(targets, agentDir)
			} else {
				agentDir := fmt.Sprintf("./.%s/skills", agentName)
				logger.Verbose("Using project-level agent directory: %s", agentDir)
				targets = append(targets, agentDir)
			}
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no install targets specified: provide a path argument or use --agent")
	}

	return targets, nil
}

// getAgentProvider returns the appropriate AgentProvider based on the agent name.
func (c *AddInstallTargetCmd) getAgentProvider(agentName string) (port.AgentProvider, error) {
	switch agentName {
	case "claude":
		return agent.NewClaude(), nil
	case "codex":
		return agent.NewCodex(), nil
	case "cursor":
		return agent.NewCursor(), nil
	case "copilot":
		return agent.NewCopilot(), nil
	case "goose":
		return agent.NewGoose(), nil
	case "opencode":
		return agent.NewOpencode(), nil
	case "gemini":
		return agent.NewGemini(), nil
	case "amp":
		return agent.NewAmp(), nil
	case "factory":
		return agent.NewFactory(), nil
	default:
		return nil, fmt.Errorf("unsupported agent: %s", agentName)
	}
}
