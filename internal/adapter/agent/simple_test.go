package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter/agent"
	"github.com/mazrean/skills-pkg/internal/port"
)

type simpleAgentTestCase struct {
	agentName    string
	constructor  func() port.AgentProvider
	wantSuffix   string // expected path suffix under home dir
}

var simpleAgentTestCases = []simpleAgentTestCase{
	{
		agentName:   "claude-code",
		constructor: agent.NewClaudeCode,
		wantSuffix:  filepath.Join(".claude", "skills"),
	},
	{
		agentName:   "kimi-cli",
		constructor: agent.NewKimiCLI,
		wantSuffix:  filepath.Join(".config", "agents", "skills"),
	},
	{
		agentName:   "replit",
		constructor: agent.NewReplit,
		wantSuffix:  filepath.Join(".config", "agents", "skills"),
	},
	{
		agentName:   "universal",
		constructor: agent.NewUniversal,
		wantSuffix:  filepath.Join(".config", "agents", "skills"),
	},
	{
		agentName:   "antigravity",
		constructor: agent.NewAntigravity,
		wantSuffix:  filepath.Join(".gemini", "antigravity", "skills"),
	},
	{
		agentName:   "augment",
		constructor: agent.NewAugment,
		wantSuffix:  filepath.Join(".augment", "skills"),
	},
	{
		agentName:   "openclaw",
		constructor: agent.NewOpenclaw,
		wantSuffix:  filepath.Join(".openclaw", "skills"),
	},
	{
		agentName:   "cline",
		constructor: agent.NewCline,
		wantSuffix:  filepath.Join(".cline", "skills"),
	},
	{
		agentName:   "codebuddy",
		constructor: agent.NewCodebuddy,
		wantSuffix:  filepath.Join(".codebuddy", "skills"),
	},
	{
		agentName:   "command-code",
		constructor: agent.NewCommandCode,
		wantSuffix:  filepath.Join(".commandcode", "skills"),
	},
	{
		agentName:   "continue",
		constructor: agent.NewContinueAgent,
		wantSuffix:  filepath.Join(".continue", "skills"),
	},
	{
		agentName:   "cortex",
		constructor: agent.NewCortex,
		wantSuffix:  filepath.Join(".snowflake", "cortex", "skills"),
	},
	{
		agentName:   "crush",
		constructor: agent.NewCrush,
		wantSuffix:  filepath.Join(".config", "crush", "skills"),
	},
	{
		agentName:   "droid",
		constructor: agent.NewDroid,
		wantSuffix:  filepath.Join(".factory", "skills"),
	},
	{
		agentName:   "gemini-cli",
		constructor: agent.NewGeminiCLI,
		wantSuffix:  filepath.Join(".gemini", "skills"),
	},
	{
		agentName:   "github-copilot",
		constructor: agent.NewGithubCopilot,
		wantSuffix:  filepath.Join(".copilot", "skills"),
	},
	{
		agentName:   "junie",
		constructor: agent.NewJunie,
		wantSuffix:  filepath.Join(".junie", "skills"),
	},
	{
		agentName:   "iflow-cli",
		constructor: agent.NewIflowCLI,
		wantSuffix:  filepath.Join(".iflow", "skills"),
	},
	{
		agentName:   "kilo",
		constructor: agent.NewKilo,
		wantSuffix:  filepath.Join(".kilocode", "skills"),
	},
	{
		agentName:   "kiro-cli",
		constructor: agent.NewKiroCLI,
		wantSuffix:  filepath.Join(".kiro", "skills"),
	},
	{
		agentName:   "kode",
		constructor: agent.NewKode,
		wantSuffix:  filepath.Join(".kode", "skills"),
	},
	{
		agentName:   "mcpjam",
		constructor: agent.NewMCPJam,
		wantSuffix:  filepath.Join(".mcpjam", "skills"),
	},
	{
		agentName:   "mistral-vibe",
		constructor: agent.NewMistralVibe,
		wantSuffix:  filepath.Join(".vibe", "skills"),
	},
	{
		agentName:   "mux",
		constructor: agent.NewMux,
		wantSuffix:  filepath.Join(".mux", "skills"),
	},
	{
		agentName:   "openhands",
		constructor: agent.NewOpenhands,
		wantSuffix:  filepath.Join(".openhands", "skills"),
	},
	{
		agentName:   "pi",
		constructor: agent.NewPi,
		wantSuffix:  filepath.Join(".pi", "agent", "skills"),
	},
	{
		agentName:   "qoder",
		constructor: agent.NewQoder,
		wantSuffix:  filepath.Join(".qoder", "skills"),
	},
	{
		agentName:   "qwen-code",
		constructor: agent.NewQwenCode,
		wantSuffix:  filepath.Join(".qwen", "skills"),
	},
	{
		agentName:   "roo",
		constructor: agent.NewRoo,
		wantSuffix:  filepath.Join(".roo", "skills"),
	},
	{
		agentName:   "trae",
		constructor: agent.NewTrae,
		wantSuffix:  filepath.Join(".trae", "skills"),
	},
	{
		agentName:   "trae-cn",
		constructor: agent.NewTraeCN,
		wantSuffix:  filepath.Join(".trae-cn", "skills"),
	},
	{
		agentName:   "windsurf",
		constructor: agent.NewWindsurf,
		wantSuffix:  filepath.Join(".codeium", "windsurf", "skills"),
	},
	{
		agentName:   "zencoder",
		constructor: agent.NewZencoder,
		wantSuffix:  filepath.Join(".zencoder", "skills"),
	},
	{
		agentName:   "neovate",
		constructor: agent.NewNeovate,
		wantSuffix:  filepath.Join(".neovate", "skills"),
	},
	{
		agentName:   "pochi",
		constructor: agent.NewPochi,
		wantSuffix:  filepath.Join(".pochi", "skills"),
	},
	{
		agentName:   "adal",
		constructor: agent.NewAdal,
		wantSuffix:  filepath.Join(".adal", "skills"),
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
