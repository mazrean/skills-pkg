package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mazrean/skills-pkg/internal/port"
)

// simpleAgent is a generic agent implementation for agents that use a fixed path
// under the user's home directory.
type simpleAgent struct {
	name       string
	projectDir string
	pathParts  []string
}

func newSimpleAgent(name, projectDir string, pathParts ...string) port.AgentProvider {
	return &simpleAgent{name: name, projectDir: projectDir, pathParts: pathParts}
}

func (a *simpleAgent) ResolveAgentDir(agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name cannot be empty")
	}

	if agentName != a.name {
		return "", fmt.Errorf("unsupported agent: %s (only '%s' is supported by this adapter)", agentName, a.name)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	parts := append([]string{home}, a.pathParts...)
	return filepath.Join(parts...), nil
}

func (a *simpleAgent) AgentName() string {
	return a.name
}

func (a *simpleAgent) ProjectDir() string {
	return a.projectDir
}

// NewClaudeCode creates a new Claude Code agent adapter instance (global: ~/.claude/skills).
func NewClaudeCode() port.AgentProvider {
	return newSimpleAgent("claude-code", ".claude/skills", ".claude", "skills")
}

// NewKimiCLI creates a new Kimi Code CLI agent adapter instance (global: ~/.config/agents/skills).
func NewKimiCLI() port.AgentProvider {
	return newSimpleAgent("kimi-cli", ".agents/skills", ".config", "agents", "skills")
}

// NewReplit creates a new Replit agent adapter instance (global: ~/.config/agents/skills).
func NewReplit() port.AgentProvider {
	return newSimpleAgent("replit", ".agents/skills", ".config", "agents", "skills")
}

// NewUniversal creates a new Universal agent adapter instance (global: ~/.config/agents/skills).
func NewUniversal() port.AgentProvider {
	return newSimpleAgent("universal", ".agents/skills", ".config", "agents", "skills")
}

// NewAntigravity creates a new Antigravity agent adapter instance (global: ~/.gemini/antigravity/skills).
func NewAntigravity() port.AgentProvider {
	return newSimpleAgent("antigravity", ".agent/skills", ".gemini", "antigravity", "skills")
}

// NewAugment creates a new Augment agent adapter instance (global: ~/.augment/skills).
func NewAugment() port.AgentProvider {
	return newSimpleAgent("augment", ".augment/skills", ".augment", "skills")
}

// NewOpenclaw creates a new OpenClaw agent adapter instance (global: ~/.openclaw/skills).
func NewOpenclaw() port.AgentProvider {
	return newSimpleAgent("openclaw", "skills", ".openclaw", "skills")
}

// NewCline creates a new Cline agent adapter instance (global: ~/.cline/skills).
func NewCline() port.AgentProvider {
	return newSimpleAgent("cline", ".cline/skills", ".cline", "skills")
}

// NewCodebuddy creates a new CodeBuddy agent adapter instance (global: ~/.codebuddy/skills).
func NewCodebuddy() port.AgentProvider {
	return newSimpleAgent("codebuddy", ".codebuddy/skills", ".codebuddy", "skills")
}

// NewCommandCode creates a new Command Code agent adapter instance (global: ~/.commandcode/skills).
func NewCommandCode() port.AgentProvider {
	return newSimpleAgent("command-code", ".commandcode/skills", ".commandcode", "skills")
}

// NewContinueAgent creates a new Continue agent adapter instance (global: ~/.continue/skills).
func NewContinueAgent() port.AgentProvider {
	return newSimpleAgent("continue", ".continue/skills", ".continue", "skills")
}

// NewCortex creates a new Cortex Code agent adapter instance (global: ~/.snowflake/cortex/skills).
func NewCortex() port.AgentProvider {
	return newSimpleAgent("cortex", ".cortex/skills", ".snowflake", "cortex", "skills")
}

// NewCrush creates a new Crush agent adapter instance (global: ~/.config/crush/skills).
func NewCrush() port.AgentProvider {
	return newSimpleAgent("crush", ".crush/skills", ".config", "crush", "skills")
}

// NewDroid creates a new Droid agent adapter instance (global: ~/.factory/skills).
func NewDroid() port.AgentProvider {
	return newSimpleAgent("droid", ".factory/skills", ".factory", "skills")
}

// NewGeminiCLI creates a new Gemini CLI agent adapter instance (global: ~/.gemini/skills).
func NewGeminiCLI() port.AgentProvider {
	return newSimpleAgent("gemini-cli", ".agents/skills", ".gemini", "skills")
}

// NewGithubCopilot creates a new GitHub Copilot agent adapter instance (global: ~/.copilot/skills).
func NewGithubCopilot() port.AgentProvider {
	return newSimpleAgent("github-copilot", ".agents/skills", ".copilot", "skills")
}

// NewJunie creates a new Junie agent adapter instance (global: ~/.junie/skills).
func NewJunie() port.AgentProvider {
	return newSimpleAgent("junie", ".junie/skills", ".junie", "skills")
}

// NewIflowCLI creates a new iFlow CLI agent adapter instance (global: ~/.iflow/skills).
func NewIflowCLI() port.AgentProvider {
	return newSimpleAgent("iflow-cli", ".iflow/skills", ".iflow", "skills")
}

// NewKilo creates a new Kilo Code agent adapter instance (global: ~/.kilocode/skills).
func NewKilo() port.AgentProvider {
	return newSimpleAgent("kilo", ".kilocode/skills", ".kilocode", "skills")
}

// NewKiroCLI creates a new Kiro CLI agent adapter instance (global: ~/.kiro/skills).
func NewKiroCLI() port.AgentProvider {
	return newSimpleAgent("kiro-cli", ".kiro/skills", ".kiro", "skills")
}

// NewKode creates a new Kode agent adapter instance (global: ~/.kode/skills).
func NewKode() port.AgentProvider {
	return newSimpleAgent("kode", ".kode/skills", ".kode", "skills")
}

// NewMCPJam creates a new MCPJam agent adapter instance (global: ~/.mcpjam/skills).
func NewMCPJam() port.AgentProvider {
	return newSimpleAgent("mcpjam", ".mcpjam/skills", ".mcpjam", "skills")
}

// NewMistralVibe creates a new Mistral Vibe agent adapter instance (global: ~/.vibe/skills).
func NewMistralVibe() port.AgentProvider {
	return newSimpleAgent("mistral-vibe", ".vibe/skills", ".vibe", "skills")
}

// NewMux creates a new Mux agent adapter instance (global: ~/.mux/skills).
func NewMux() port.AgentProvider {
	return newSimpleAgent("mux", ".mux/skills", ".mux", "skills")
}

// NewOpenhands creates a new OpenHands agent adapter instance (global: ~/.openhands/skills).
func NewOpenhands() port.AgentProvider {
	return newSimpleAgent("openhands", ".openhands/skills", ".openhands", "skills")
}

// NewPi creates a new Pi agent adapter instance (global: ~/.pi/agent/skills).
func NewPi() port.AgentProvider {
	return newSimpleAgent("pi", ".pi/skills", ".pi", "agent", "skills")
}

// NewQoder creates a new Qoder agent adapter instance (global: ~/.qoder/skills).
func NewQoder() port.AgentProvider {
	return newSimpleAgent("qoder", ".qoder/skills", ".qoder", "skills")
}

// NewQwenCode creates a new Qwen Code agent adapter instance (global: ~/.qwen/skills).
func NewQwenCode() port.AgentProvider {
	return newSimpleAgent("qwen-code", ".qwen/skills", ".qwen", "skills")
}

// NewRoo creates a new Roo Code agent adapter instance (global: ~/.roo/skills).
func NewRoo() port.AgentProvider {
	return newSimpleAgent("roo", ".roo/skills", ".roo", "skills")
}

// NewTrae creates a new Trae agent adapter instance (global: ~/.trae/skills).
func NewTrae() port.AgentProvider {
	return newSimpleAgent("trae", ".trae/skills", ".trae", "skills")
}

// NewTraeCN creates a new Trae CN agent adapter instance (global: ~/.trae-cn/skills).
func NewTraeCN() port.AgentProvider {
	return newSimpleAgent("trae-cn", ".trae/skills", ".trae-cn", "skills")
}

// NewWindsurf creates a new Windsurf agent adapter instance (global: ~/.codeium/windsurf/skills).
func NewWindsurf() port.AgentProvider {
	return newSimpleAgent("windsurf", ".windsurf/skills", ".codeium", "windsurf", "skills")
}

// NewZencoder creates a new Zencoder agent adapter instance (global: ~/.zencoder/skills).
func NewZencoder() port.AgentProvider {
	return newSimpleAgent("zencoder", ".zencoder/skills", ".zencoder", "skills")
}

// NewNeovate creates a new Neovate agent adapter instance (global: ~/.neovate/skills).
func NewNeovate() port.AgentProvider {
	return newSimpleAgent("neovate", ".neovate/skills", ".neovate", "skills")
}

// NewPochi creates a new Pochi agent adapter instance (global: ~/.pochi/skills).
func NewPochi() port.AgentProvider {
	return newSimpleAgent("pochi", ".pochi/skills", ".pochi", "skills")
}

// NewAdal creates a new AdaL agent adapter instance (global: ~/.adal/skills).
func NewAdal() port.AgentProvider {
	return newSimpleAgent("adal", ".adal/skills", ".adal", "skills")
}
