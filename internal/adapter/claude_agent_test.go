package adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/adapter"
)

// TestClaudeAgentAdapter_ResolveAgentDir tests directory resolution for Claude agent.
// Requirements: 10.3, 10.4
func TestClaudeAgentAdapter_ResolveAgentDir(t *testing.T) {
	tests := []struct {
		name            string
		agentName       string
		checkSuffix     string
		wantErr         bool
		checkAbsolute   bool
		checkHomePrefix bool
	}{
		{
			name:            "claude agent",
			agentName:       "claude",
			wantErr:         false,
			checkAbsolute:   true,
			checkSuffix:     filepath.Join(".claude", "skills"),
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
			provider := adapter.NewClaudeAgentAdapter()

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
				expected := filepath.Join(home, ".claude", "skills")
				if dir != expected {
					t.Errorf("ResolveAgentDir(%q) = %q, want %q", tt.agentName, dir, expected)
				}
			}
		})
	}
}

// TestClaudeAgentAdapter_AgentName tests agent name retrieval.
// Requirements: 10.4
func TestClaudeAgentAdapter_AgentName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return claude",
			want: "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := adapter.NewClaudeAgentAdapter()
			if got := provider.AgentName(); got != tt.want {
				t.Errorf("AgentName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// hasPathSuffix checks if the path ends with the specified suffix.
func hasPathSuffix(path, suffix string) bool {
	cleanPath := filepath.Clean(path)
	cleanSuffix := filepath.Clean(suffix)

	// Split path into components
	pathParts := splitPath(cleanPath)
	suffixParts := splitPath(cleanSuffix)

	if len(pathParts) < len(suffixParts) {
		return false
	}

	// Compare suffix parts from the end
	for i := range len(suffixParts) {
		if pathParts[len(pathParts)-len(suffixParts)+i] != suffixParts[i] {
			return false
		}
	}

	return true
}

// splitPath splits a file path into components.
func splitPath(path string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		if dir == "" || dir == "/" {
			if dir == "/" {
				parts = append([]string{"/"}, parts...)
			}
			break
		}
		path = filepath.Clean(dir)
	}
	return parts
}
