package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

// TestCodexAgentAdapter_ResolveAgentDir tests directory resolution for Codex agent.
func TestCodexAgentAdapter_ResolveAgentDir(t *testing.T) {
	tests := []struct {
		name            string
		agentName       string
		wantErr         bool
		checkAbsolute   bool
		checkSuffix     string
		checkHomePrefix bool
	}{
		{
			name:            "codex agent",
			agentName:       "codex",
			wantErr:         false,
			checkAbsolute:   true,
			checkSuffix:     filepath.Join(".codex", "skills"),
			checkHomePrefix: true,
		},
		{
			name:      "unsupported agent",
			agentName: "unsupported",
			wantErr:   true,
		},
		{
			name:      "empty agent name",
			agentName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := adapter.NewCodexAgentAdapter()

			dir, err := provider.ResolveAgentDir(tt.agentName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveAgentDir(%q) error = %v, wantErr %v", tt.agentName, err, tt.wantErr)
			}

			if err != nil {
				// Error message should not be empty for errors
				if err.Error() == "" {
					t.Error("ResolveAgentDir error message is empty")
				}
				return
			}

			if tt.checkAbsolute && !filepath.IsAbs(dir) {
				t.Errorf("ResolveAgentDir(%q) = %q, want absolute path", tt.agentName, dir)
			}

			if tt.checkSuffix != "" && !hasPathSuffix(dir, tt.checkSuffix) {
				t.Errorf("ResolveAgentDir(%q) = %q, want path ending with %q", tt.agentName, dir, tt.checkSuffix)
			}

			if tt.checkHomePrefix {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("os.UserHomeDir() error = %v", err)
				}
				expected := filepath.Join(home, ".codex", "skills")
				if dir != expected {
					t.Errorf("ResolveAgentDir(%q) = %q, want %q", tt.agentName, dir, expected)
				}
			}
		})
	}
}

// TestCodexAgentAdapter_AgentName tests agent name retrieval.
func TestCodexAgentAdapter_AgentName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return codex",
			want: "codex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := adapter.NewCodexAgentAdapter()
			if got := provider.AgentName(); got != tt.want {
				t.Errorf("AgentName() = %q, want %q", got, tt.want)
			}
		})
	}
}
