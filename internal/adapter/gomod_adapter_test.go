package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestGoModAdapter_SourceType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return go-module",
			want: "go-module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewGoModAdapter()
			if got := adapter.SourceType(); got != tt.want {
				t.Errorf("SourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoModAdapter_Download_InvalidSource(t *testing.T) {
	adapter := NewGoModAdapter()
	ctx := context.Background()

	tests := []struct {
		name    string
		source  *port.Source
		version string
		wantErr bool
	}{
		{
			name: "empty source type",
			source: &port.Source{
				Type: "",
				URL:  "golang.org/x/tools",
			},
			version: "v0.1.0",
			wantErr: true,
		},
		{
			name: "empty URL",
			source: &port.Source{
				Type: "go-module",
				URL:  "",
			},
			version: "v0.1.0",
			wantErr: true,
		},
		{
			name: "wrong source type",
			source: &port.Source{
				Type: "git",
				URL:  "golang.org/x/tools",
			},
			version: "v0.1.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.Download(ctx, tt.source, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoModAdapter_Download_ModuleErrors(t *testing.T) {
	tests := []struct {
		checkVersionType func(t *testing.T, version string)
		name             string
		url              string
		version          string
		checkVersion     string
		skipInShort      bool
		wantErr          bool
		checkNetworkErr  bool
		checkPath        bool
		checkGoMod       bool
	}{
		{
			name:            "module not found",
			url:             "github.com/this-module-absolutely-does-not-exist-12345/nonexistent",
			version:         "v1.0.0",
			wantErr:         true,
			checkNetworkErr: true,
		},
		{
			name:         "valid module with specific version",
			url:          "golang.org/x/exp",
			version:      "v0.0.0-20231110203233-9a3e6036ecaa",
			skipInShort:  true,
			wantErr:      false,
			checkPath:    true,
			checkVersion: "v0.0.0-20231110203233-9a3e6036ecaa",
			checkGoMod:   true,
		},
		{
			name:        "valid module with latest version",
			url:         "golang.org/x/exp",
			version:     "latest",
			skipInShort: true,
			wantErr:     false,
			checkPath:   true,
			checkVersionType: func(t *testing.T, version string) {
				if version == "" || version == "latest" {
					t.Errorf("Download() result.Version should be resolved, got %v", version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipInShort && testing.Short() {
				t.Skip("skipping integration test in short mode")
			}

			adapter := NewGoModAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "go-module",
				URL:  tt.url,
			}

			result, err := adapter.Download(ctx, source, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkNetworkErr && err != nil {
				if !domain.IsNetworkError(err) {
					t.Errorf("Download() error should be a network error, got %v", err)
				}
				return
			}

			if err != nil {
				return
			}

			// Clean up
			defer func() {
				_ = os.RemoveAll(result.Path)
			}()

			if tt.checkPath && result.Path == "" {
				t.Error("Download() result.Path is empty")
			}

			if tt.checkVersion != "" && result.Version != tt.checkVersion {
				t.Errorf("Download() result.Version = %v, want %v", result.Version, tt.checkVersion)
			}

			if tt.checkVersionType != nil {
				tt.checkVersionType(t, result.Version)
			}

			if tt.checkPath {
				if _, err := os.Stat(result.Path); os.IsNotExist(err) {
					t.Errorf("Download() directory does not exist: %s", result.Path)
				}
			}

			if tt.checkGoMod {
				goMod := filepath.Join(result.Path, "go.mod")
				if _, err := os.Stat(goMod); os.IsNotExist(err) {
					t.Error("Download() go.mod not found in downloaded module")
				}
			}
		})
	}
}

func TestGoModAdapter_GetLatestVersion_InvalidSource(t *testing.T) {
	adapter := NewGoModAdapter()
	ctx := context.Background()

	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			name: "empty source type",
			source: &port.Source{
				Type: "",
				URL:  "golang.org/x/tools",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			source: &port.Source{
				Type: "go-module",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "wrong source type",
			source: &port.Source{
				Type: "git",
				URL:  "golang.org/x/tools",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.GetLatestVersion(ctx, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoModAdapter_GetLatestVersion_ModuleErrors(t *testing.T) {
	tests := []struct {
		checkVersion    func(t *testing.T, version string)
		name            string
		url             string
		skipInShort     bool
		wantErr         bool
		checkNetworkErr bool
	}{
		{
			name:            "module not found",
			url:             "github.com/this-module-absolutely-does-not-exist-12345/nonexistent",
			wantErr:         true,
			checkNetworkErr: true,
		},
		{
			name:        "valid module",
			url:         "golang.org/x/exp",
			skipInShort: true,
			wantErr:     false,
			checkVersion: func(t *testing.T, version string) {
				if version == "" {
					t.Error("GetLatestVersion() returned empty version")
				}
				// Version should start with "v" (e.g., "v0.0.0-20231110203233-9a3e6036ecaa")
				if len(version) < 2 || version[0] != 'v' {
					t.Errorf("GetLatestVersion() version seems invalid: %s", version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipInShort && testing.Short() {
				t.Skip("skipping integration test in short mode")
			}

			adapter := NewGoModAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "go-module",
				URL:  tt.url,
			}

			version, err := adapter.GetLatestVersion(ctx, source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkNetworkErr && err != nil {
				if !domain.IsNetworkError(err) {
					t.Errorf("GetLatestVersion() error should be a network error, got %v", err)
				}
				return
			}

			if err != nil {
				return
			}

			if tt.checkVersion != nil {
				tt.checkVersion(t, version)
			}
		})
	}
}
