package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter/agent"
	"github.com/mazrean/skills-pkg/internal/port"
)

type simpleAgentTestCase struct {
	agentName      string
	constructor    func() port.AgentProvider
	wantSuffix     string // expected path suffix under home dir
	wantProjectDir string // expected project-level directory
}

var simpleAgentTestCases = []simpleAgentTestCase{
	{
		agentName:      "claude-code",
		constructor:    agent.NewClaudeCode,
		wantSuffix:     filepath.Join(".claude", "skills"),
		wantProjectDir: ".claude/skills",
	},
	{
		agentName:      "kimi-cli",
		constructor:    agent.NewKimiCLI,
		wantSuffix:     filepath.Join(".config", "agents", "skills"),
		wantProjectDir: ".agents/skills",
	},
	{
		agentName:      "replit",
		constructor:    agent.NewReplit,
		wantSuffix:     filepath.Join(".config", "agents", "skills"),
		wantProjectDir: ".agents/skills",
	},
	{
		agentName:      "universal",
		constructor:    agent.NewUniversal,
		wantSuffix:     filepath.Join(".config", "agents", "skills"),
		wantProjectDir: ".agents/skills",
	},
	{
		agentName:      "antigravity",
		constructor:    agent.NewAntigravity,
		wantSuffix:     filepath.Join(".gemini", "antigravity", "skills"),
		wantProjectDir: ".agent/skills",
	},
	{
		agentName:      "augment",
		constructor:    agent.NewAugment,
		wantSuffix:     filepath.Join(".augment", "skills"),
		wantProjectDir: ".augment/skills",
	},
	{
		agentName:      "openclaw",
		constructor:    agent.NewOpenclaw,
		wantSuffix:     filepath.Join(".openclaw", "skills"),
		wantProjectDir: "skills",
	},
	{
		agentName:      "cline",
		constructor:    agent.NewCline,
		wantSuffix:     filepath.Join(".cline", "skills"),
		wantProjectDir: ".cline/skills",
	},
	{
		agentName:      "codebuddy",
		constructor:    agent.NewCodebuddy,
		wantSuffix:     filepath.Join(".codebuddy", "skills"),
		wantProjectDir: ".codebuddy/skills",
	},
	{
		agentName:      "command-code",
		constructor:    agent.NewCommandCode,
		wantSuffix:     filepath.Join(".commandcode", "skills"),
		wantProjectDir: ".commandcode/skills",
	},
	{
		agentName:      "continue",
		constructor:    agent.NewContinueAgent,
		wantSuffix:     filepath.Join(".continue", "skills"),
		wantProjectDir: ".continue/skills",
	},
	{
		agentName:      "cortex",
		constructor:    agent.NewCortex,
		wantSuffix:     filepath.Join(".snowflake", "cortex", "skills"),
		wantProjectDir: ".cortex/skills",
	},
	{
		agentName:      "crush",
		constructor:    agent.NewCrush,
		wantSuffix:     filepath.Join(".config", "crush", "skills"),
		wantProjectDir: ".crush/skills",
	},
	{
		agentName:      "droid",
		constructor:    agent.NewDroid,
		wantSuffix:     filepath.Join(".factory", "skills"),
		wantProjectDir: ".factory/skills",
	},
	{
		agentName:      "gemini-cli",
		constructor:    agent.NewGeminiCLI,
		wantSuffix:     filepath.Join(".gemini", "skills"),
		wantProjectDir: ".agents/skills",
	},
	{
		agentName:      "github-copilot",
		constructor:    agent.NewGithubCopilot,
		wantSuffix:     filepath.Join(".copilot", "skills"),
		wantProjectDir: ".agents/skills",
	},
	{
		agentName:      "junie",
		constructor:    agent.NewJunie,
		wantSuffix:     filepath.Join(".junie", "skills"),
		wantProjectDir: ".junie/skills",
	},
	{
		agentName:      "iflow-cli",
		constructor:    agent.NewIflowCLI,
		wantSuffix:     filepath.Join(".iflow", "skills"),
		wantProjectDir: ".iflow/skills",
	},
	{
		agentName:      "kilo",
		constructor:    agent.NewKilo,
		wantSuffix:     filepath.Join(".kilocode", "skills"),
		wantProjectDir: ".kilocode/skills",
	},
	{
		agentName:      "kiro-cli",
		constructor:    agent.NewKiroCLI,
		wantSuffix:     filepath.Join(".kiro", "skills"),
		wantProjectDir: ".kiro/skills",
	},
	{
		agentName:      "kode",
		constructor:    agent.NewKode,
		wantSuffix:     filepath.Join(".kode", "skills"),
		wantProjectDir: ".kode/skills",
	},
	{
		agentName:      "mcpjam",
		constructor:    agent.NewMCPJam,
		wantSuffix:     filepath.Join(".mcpjam", "skills"),
		wantProjectDir: ".mcpjam/skills",
	},
	{
		agentName:      "mistral-vibe",
		constructor:    agent.NewMistralVibe,
		wantSuffix:     filepath.Join(".vibe", "skills"),
		wantProjectDir: ".vibe/skills",
	},
	{
		agentName:      "mux",
		constructor:    agent.NewMux,
		wantSuffix:     filepath.Join(".mux", "skills"),
		wantProjectDir: ".mux/skills",
	},
	{
		agentName:      "openhands",
		constructor:    agent.NewOpenhands,
		wantSuffix:     filepath.Join(".openhands", "skills"),
		wantProjectDir: ".openhands/skills",
	},
	{
		agentName:      "pi",
		constructor:    agent.NewPi,
		wantSuffix:     filepath.Join(".pi", "agent", "skills"),
		wantProjectDir: ".pi/skills",
	},
	{
		agentName:      "qoder",
		constructor:    agent.NewQoder,
		wantSuffix:     filepath.Join(".qoder", "skills"),
		wantProjectDir: ".qoder/skills",
	},
	{
		agentName:      "qwen-code",
		constructor:    agent.NewQwenCode,
		wantSuffix:     filepath.Join(".qwen", "skills"),
		wantProjectDir: ".qwen/skills",
	},
	{
		agentName:      "roo",
		constructor:    agent.NewRoo,
		wantSuffix:     filepath.Join(".roo", "skills"),
		wantProjectDir: ".roo/skills",
	},
	{
		agentName:      "trae",
		constructor:    agent.NewTrae,
		wantSuffix:     filepath.Join(".trae", "skills"),
		wantProjectDir: ".trae/skills",
	},
	{
		agentName:      "trae-cn",
		constructor:    agent.NewTraeCN,
		wantSuffix:     filepath.Join(".trae-cn", "skills"),
		wantProjectDir: ".trae/skills",
	},
	{
		agentName:      "windsurf",
		constructor:    agent.NewWindsurf,
		wantSuffix:     filepath.Join(".codeium", "windsurf", "skills"),
		wantProjectDir: ".windsurf/skills",
	},
	{
		agentName:      "zencoder",
		constructor:    agent.NewZencoder,
		wantSuffix:     filepath.Join(".zencoder", "skills"),
		wantProjectDir: ".zencoder/skills",
	},
	{
		agentName:      "neovate",
		constructor:    agent.NewNeovate,
		wantSuffix:     filepath.Join(".neovate", "skills"),
		wantProjectDir: ".neovate/skills",
	},
	{
		agentName:      "pochi",
		constructor:    agent.NewPochi,
		wantSuffix:     filepath.Join(".pochi", "skills"),
		wantProjectDir: ".pochi/skills",
	},
	{
		agentName:      "adal",
		constructor:    agent.NewAdal,
		wantSuffix:     filepath.Join(".adal", "skills"),
		wantProjectDir: ".adal/skills",
	},
}

func TestSimpleAgents_ResolveAgentDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	for _, tc := range simpleAgentTestCases {
		t.Run(tc.agentName, func(t *testing.T) {
			provider := tc.constructor()

			// Success case
			dir, err := provider.ResolveAgentDir(tc.agentName)
			if err != nil {
				t.Fatalf("ResolveAgentDir(%q) unexpected error = %v", tc.agentName, err)
			}
			if !filepath.IsAbs(dir) {
				t.Errorf("ResolveAgentDir(%q) = %q, want absolute path", tc.agentName, dir)
			}
			expected := filepath.Join(home, tc.wantSuffix)
			if dir != expected {
				t.Errorf("ResolveAgentDir(%q) = %q, want %q", tc.agentName, dir, expected)
			}

			// Unsupported agent name
			_, err = provider.ResolveAgentDir("unsupported")
			if err == nil {
				t.Errorf("ResolveAgentDir(%q) with unsupported name: want error, got nil", tc.agentName)
			}

			// Empty agent name
			_, err = provider.ResolveAgentDir("")
			if err == nil {
				t.Errorf("ResolveAgentDir(%q) with empty name: want error, got nil", tc.agentName)
			}
		})
	}
}

func TestSimpleAgents_ProjectDir(t *testing.T) {
	for _, tc := range simpleAgentTestCases {
		t.Run(tc.agentName, func(t *testing.T) {
			provider := tc.constructor()
			if got := provider.ProjectDir(); got != tc.wantProjectDir {
				t.Errorf("ProjectDir() = %q, want %q", got, tc.wantProjectDir)
			}
		})
	}
}

func TestSimpleAgents_AgentName(t *testing.T) {
	for _, tc := range simpleAgentTestCases {
		t.Run(tc.agentName, func(t *testing.T) {
			provider := tc.constructor()
			if got := provider.AgentName(); got != tc.agentName {
				t.Errorf("AgentName() = %q, want %q", got, tc.agentName)
			}
		})
	}
}
