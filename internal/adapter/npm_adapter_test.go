package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestNpmAdapter_SourceType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return npm",
			want: "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewNpmAdapter()
			if got := adapter.SourceType(); got != tt.want {
				t.Errorf("SourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNpmAdapter_Download_InvalidSource(t *testing.T) {
	adapter := NewNpmAdapter()
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
				URL:  "test-package",
			},
			version: "1.0.0",
			wantErr: true,
		},
		{
			name: "empty URL",
			source: &port.Source{
				Type: "npm",
				URL:  "",
			},
			version: "1.0.0",
			wantErr: true,
		},
		{
			name: "wrong source type",
			source: &port.Source{
				Type: "git",
				URL:  "test-package",
			},
			version: "1.0.0",
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

func TestNpmAdapter_Download_PackageErrors(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		version          string
		skipInShort      bool
		wantErr          bool
		checkNetworkErr  bool
		checkPath        bool
		checkVersion     string
		checkPackageJSON bool
		checkVersionType func(t *testing.T, version string)
	}{
		{
			name:            "package not found",
			url:             "this-package-absolutely-does-not-exist-12345",
			version:         "1.0.0",
			wantErr:         true,
			checkNetworkErr: true,
		},
		{
			name:             "valid package with specific version",
			url:              "lodash",
			version:          "4.17.21",
			skipInShort:      true,
			wantErr:          false,
			checkPath:        true,
			checkVersion:     "4.17.21",
			checkPackageJSON: true,
		},
		{
			name:        "valid package with latest version",
			url:         "lodash",
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

			adapter := NewNpmAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "npm",
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

			if tt.checkPackageJSON {
				packageJSON := filepath.Join(result.Path, "package.json")
				if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
					t.Error("Download() package.json not found in downloaded package")
				}
			}
		})
	}
}

func TestNpmAdapter_GetLatestVersion_InvalidSource(t *testing.T) {
	adapter := NewNpmAdapter()
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
				URL:  "test-package",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			source: &port.Source{
				Type: "npm",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "wrong source type",
			source: &port.Source{
				Type: "git",
				URL:  "test-package",
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

func TestNpmAdapter_GetLatestVersion_PackageErrors(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		skipInShort     bool
		wantErr         bool
		checkNetworkErr bool
		checkVersion    func(t *testing.T, version string)
	}{
		{
			name:            "package not found",
			url:             "this-package-absolutely-does-not-exist-12345",
			wantErr:         true,
			checkNetworkErr: true,
		},
		{
			name:        "valid package",
			url:         "lodash",
			skipInShort: true,
			wantErr:     false,
			checkVersion: func(t *testing.T, version string) {
				if version == "" {
					t.Error("GetLatestVersion() returned empty version")
				}
				// Version should be in semantic version format (e.g., "4.17.21")
				if len(version) < 5 {
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

			adapter := NewNpmAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "npm",
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
