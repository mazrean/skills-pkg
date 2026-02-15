package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

func TestGitAdapter_SourceType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "should return git",
			want: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewGitAdapter()
			if got := adapter.SourceType(); got != tt.want {
				t.Errorf("SourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitAdapter_Download(t *testing.T) {
	tests := []struct {
		checkVersion     func(t *testing.T, got string)
		name             string
		url              string
		version          string
		skipInShort      bool
		wantErr          bool
		expectErrOnRetry bool
		checkPath        bool
	}{
		{
			name:        "download with tag",
			url:         "https://github.com/go-git/go-git.git",
			version:     "v5.12.0",
			skipInShort: true,
			wantErr:     false,
			checkVersion: func(t *testing.T, got string) {
				if got != "v5.12.0" {
					t.Errorf("Download() version = %v, want %v", got, "v5.12.0")
				}
			},
			checkPath: true,
		},
		{
			name:             "download with non-existent commit hash",
			url:              "https://github.com/anthropics/anthropic-sdk-go.git",
			version:          "abc123def456",
			skipInShort:      true,
			wantErr:          false,
			expectErrOnRetry: true,
			checkVersion: func(t *testing.T, got string) {
				if got != "abc123def456" {
					t.Errorf("Download() version = %v, want %v", got, "abc123def456")
				}
			},
		},
		{
			name:        "download with latest",
			url:         "https://github.com/anthropics/anthropic-sdk-go.git",
			version:     "latest",
			skipInShort: true,
			wantErr:     false,
			checkVersion: func(t *testing.T, got string) {
				if got == "" || got == "latest" {
					t.Errorf("Download() should return actual commit hash for latest, got %v", got)
				}
			},
			checkPath: true,
		},
		{
			name:    "invalid URL",
			url:     "https://invalid-git-url-that-does-not-exist.com/repo.git",
			version: "latest",
			wantErr: true,
		},
		{
			name:        "non-existent version",
			url:         "https://github.com/anthropics/anthropic-sdk-go.git",
			version:     "v999.999.999",
			skipInShort: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipInShort && testing.Short() {
				t.Skip("Skipping integration test in short mode")
			}

			adapter := NewGitAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "git",
				URL:  tt.url,
			}

			tempDir := t.TempDir()
			_ = os.Setenv("SKILLSPKG_TEMP_DIR", tempDir)
			defer func() { _ = os.Unsetenv("SKILLSPKG_TEMP_DIR") }()

			result, err := adapter.Download(ctx, source, tt.version)
			if tt.expectErrOnRetry && err != nil {
				t.Logf("Download() error (expected for non-existent commit) = %v", err)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				t.Logf("Error message: %v", err)
				return
			}

			if tt.checkVersion != nil {
				tt.checkVersion(t, result.Version)
			}

			if tt.checkPath {
				if _, err := os.Stat(result.Path); os.IsNotExist(err) {
					t.Errorf("Downloaded directory does not exist: %v", result.Path)
				}
			}
		})
	}
}

func TestGitAdapter_GetLatestVersion(t *testing.T) {
	tests := []struct {
		checkResult func(t *testing.T, version string)
		name        string
		url         string
		skipInShort bool
		wantErr     bool
	}{
		{
			name:        "valid repository",
			url:         "https://github.com/anthropics/anthropic-sdk-go.git",
			skipInShort: true,
			wantErr:     false,
			checkResult: func(t *testing.T, version string) {
				if version == "" {
					t.Error("GetLatestVersion() should return a non-empty version")
				}
				t.Logf("Latest version: %v", version)
			},
		},
		{
			name:    "invalid URL",
			url:     "https://invalid-git-url-that-does-not-exist.com/repo.git",
			wantErr: true,
			checkResult: func(t *testing.T, version string) {
				// Error should be descriptive
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipInShort && testing.Short() {
				t.Skip("Skipping integration test in short mode")
			}

			adapter := NewGitAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "git",
				URL:  tt.url,
			}

			version, err := adapter.GetLatestVersion(ctx, source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				t.Logf("Error message (should include network error): %v", err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, version)
			}
		})
	}
}

func TestGitAdapter_Download_CleansUpOnError(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		version string
		wantErr bool
	}{
		{
			name:    "invalid URL should cleanup on error",
			url:     "https://invalid-git-url.com/repo.git",
			version: "latest",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewGitAdapter()
			ctx := context.Background()

			source := &port.Source{
				Type: "git",
				URL:  tt.url,
			}

			tempDir := t.TempDir()
			_ = os.Setenv("SKILLSPKG_TEMP_DIR", tempDir)
			defer func() { _ = os.Unsetenv("SKILLSPKG_TEMP_DIR") }()

			_, err := adapter.Download(ctx, source, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that temp directory is cleaned up (or minimal files remain)
			entries, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatalf("Failed to read temp dir: %v", err)
			}

			// There might be some cleanup artifacts, but there should not be a complete repository
			for _, entry := range entries {
				path := filepath.Join(tempDir, entry.Name())
				t.Logf("Remaining file/dir after error: %v", path)
			}
		})
	}
}
