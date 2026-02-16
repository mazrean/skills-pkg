package pkgmanager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

func TestGoMod_SourceType(t *testing.T) {
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
			adapter := NewGoMod()
			if got := adapter.SourceType(); got != tt.want {
				t.Errorf("SourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoMod_Download_InvalidSource(t *testing.T) {
	adapter := NewGoMod()
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

func TestGoMod_Download_ModuleErrors(t *testing.T) {
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

			adapter := NewGoMod()
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

func TestGoMod_GetLatestVersion_InvalidSource(t *testing.T) {
	adapter := NewGoMod()
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

func TestGoMod_GetLatestVersion_ModuleErrors(t *testing.T) {
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

			adapter := NewGoMod()
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
func TestParseGOPROXY(t *testing.T) {
	tests := []struct {
		name     string
		goproxy  string
		expected []proxyEntry
	}{
		{
			name:    "empty string defaults to proxy.golang.org,direct",
			goproxy: "",
			expected: []proxyEntry{
				{url: "https://proxy.golang.org", fallback: true},
				{url: "direct", fallback: true},
			},
		},
		{
			name:    "single proxy",
			goproxy: "https://goproxy.io",
			expected: []proxyEntry{
				{url: "https://goproxy.io", fallback: true},
			},
		},
		{
			name:    "direct only",
			goproxy: "direct",
			expected: []proxyEntry{
				{url: "direct", fallback: true},
			},
		},
		{
			name:    "off only",
			goproxy: "off",
			expected: []proxyEntry{
				{url: "off", fallback: true},
			},
		},
		{
			name:    "comma-separated proxies (fallback)",
			goproxy: "https://goproxy.io,https://proxy.golang.org,direct",
			expected: []proxyEntry{
				{url: "https://goproxy.io", fallback: true},
				{url: "https://proxy.golang.org", fallback: true},
				{url: "direct", fallback: true},
			},
		},
		{
			name:    "pipe-separated proxies (always try)",
			goproxy: "https://goproxy.io|https://proxy.golang.org",
			expected: []proxyEntry{
				{url: "https://goproxy.io", fallback: true},
				{url: "https://proxy.golang.org", fallback: false},
			},
		},
		{
			name:    "mixed comma and pipe",
			goproxy: "https://goproxy.io|https://mirror.io,https://proxy.golang.org,direct",
			expected: []proxyEntry{
				{url: "https://goproxy.io", fallback: true},
				{url: "https://mirror.io", fallback: false},
				{url: "https://proxy.golang.org", fallback: true},
				{url: "direct", fallback: true},
			},
		},
		{
			name:    "with whitespace",
			goproxy: " https://goproxy.io , https://proxy.golang.org , direct ",
			expected: []proxyEntry{
				{url: "https://goproxy.io", fallback: true},
				{url: "https://proxy.golang.org", fallback: true},
				{url: "direct", fallback: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGOPROXY(tt.goproxy)

			if len(got) != len(tt.expected) {
				t.Errorf("parseGOPROXY() returned %d entries, want %d", len(got), len(tt.expected))
				return
			}

			for i, entry := range got {
				if entry.url != tt.expected[i].url {
					t.Errorf("parseGOPROXY()[%d].url = %v, want %v", i, entry.url, tt.expected[i].url)
				}
				if entry.fallback != tt.expected[i].fallback {
					t.Errorf("parseGOPROXY()[%d].fallback = %v, want %v", i, entry.fallback, tt.expected[i].fallback)
				}
			}
		})
	}
}

func TestFindGoMod(t *testing.T) {
	tests := []struct {
		setupDir  func(t *testing.T) string
		name      string
		checkPath bool
		wantErr   bool
	}{
		{
			name: "go.mod in current directory",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tempDir := t.TempDir()
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
					t.Fatal(err)
				}
				return tempDir
			},
			wantErr:   false,
			checkPath: true,
		},
		{
			name: "go.mod in parent directory",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tempDir := t.TempDir()
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
					t.Fatal(err)
				}
				subDir := filepath.Join(tempDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				return subDir
			},
			wantErr:   false,
			checkPath: true,
		},
		{
			name: "no go.mod found",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tempDir := t.TempDir()
				return tempDir
			},
			wantErr:   true,
			checkPath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original working directory
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if chdirErr := os.Chdir(origDir); chdirErr != nil {
					t.Error(chdirErr)
				}
			}()

			// Change to test directory
			testDir := tt.setupDir(t)
			if err = os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			goModPath, err := findGoMod()
			if (err != nil) != tt.wantErr {
				t.Errorf("findGoMod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkPath && goModPath == "" {
				t.Error("findGoMod() returned empty path")
			}

			if tt.checkPath {
				if _, err := os.Stat(goModPath); os.IsNotExist(err) {
					t.Errorf("findGoMod() returned non-existent path: %s", goModPath)
				}
			}
		})
	}
}

func TestGetVersionFromGoMod(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		modulePath   string
		wantVersion  string
		wantErr      bool
	}{
		{
			name: "module found with version",
			goModContent: `module test

require (
	github.com/example/skill v1.2.3
	golang.org/x/tools v0.1.0
)
`,
			modulePath:  "github.com/example/skill",
			wantVersion: "v1.2.3",
			wantErr:     false,
		},
		{
			name: "module not found",
			goModContent: `module test

require (
	golang.org/x/tools v0.1.0
)
`,
			modulePath:  "github.com/example/skill",
			wantVersion: "",
			wantErr:     false,
		},
		{
			name: "indirect dependency",
			goModContent: `module test

require (
	github.com/example/skill v1.2.3 // indirect
	golang.org/x/tools v0.1.0
)
`,
			modulePath:  "github.com/example/skill",
			wantVersion: "v1.2.3",
			wantErr:     false,
		},
		{
			name: "pseudo-version",
			goModContent: `module test

require (
	github.com/example/skill v0.0.0-20231110203233-9a3e6036ecaa
)
`,
			modulePath:  "github.com/example/skill",
			wantVersion: "v0.0.0-20231110203233-9a3e6036ecaa",
			wantErr:     false,
		},
		{
			name:         "invalid go.mod",
			goModContent: "invalid content",
			modulePath:   "github.com/example/skill",
			wantVersion:  "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			goModPath := filepath.Join(tempDir, "go.mod")
			if err := os.WriteFile(goModPath, []byte(tt.goModContent), 0644); err != nil {
				t.Fatal(err)
			}

			version, err := getVersionFromGoMod(goModPath, tt.modulePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVersionFromGoMod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if version != tt.wantVersion {
				t.Errorf("getVersionFromGoMod() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestGoMod_Download_WithGoModVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chdirErr := os.Chdir(origDir); chdirErr != nil {
			t.Error(chdirErr)
		}
	}()

	// Create temporary directory with go.mod
	tempDir := t.TempDir()
	goModContent := `module test

require (
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
)
`
	goModPath := filepath.Join(tempDir, "go.mod")
	if err = os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	if err = os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	adapter := NewGoMod()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "golang.org/x/exp",
	}

	// Download without specifying version (should use go.mod version)
	result, err := adapter.Download(ctx, source, "")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if result.Version != "v0.0.0-20231110203233-9a3e6036ecaa" {
		t.Errorf("Download() version = %v, want v0.0.0-20231110203233-9a3e6036ecaa", result.Version)
	}

	// Verify FromGoMod flag is set
	if !result.FromGoMod {
		t.Errorf("Download() FromGoMod = %v, want true (version was resolved from go.mod)", result.FromGoMod)
	}

	// Clean up
	if err := os.RemoveAll(result.Path); err != nil {
		t.Error(err)
	}
}

func TestGoMod_Download_WithLatestExplicit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chdirErr := os.Chdir(origDir); chdirErr != nil {
			t.Error(chdirErr)
		}
	}()

	// Create temporary directory with go.mod
	tempDir := t.TempDir()
	goModContent := `module test

require (
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
)
`
	goModPath := filepath.Join(tempDir, "go.mod")
	if err = os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	if err = os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	adapter := NewGoMod()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "golang.org/x/exp",
	}

	// Download with explicit "latest" (should NOT use go.mod version)
	result, err := adapter.Download(ctx, source, "latest")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Should fetch latest version from proxy, not the go.mod version
	// We can't assert the exact version, but it should not be the go.mod version
	// if a newer version exists
	t.Logf("Downloaded version: %s (go.mod has v0.0.0-20231110203233-9a3e6036ecaa)", result.Version)

	// Verify FromGoMod flag is NOT set (explicit "latest" bypasses go.mod)
	if result.FromGoMod {
		t.Errorf("Download() FromGoMod = %v, want false (explicit 'latest' should not use go.mod)", result.FromGoMod)
	}

	// Clean up
	if err := os.RemoveAll(result.Path); err != nil {
		t.Error(err)
	}
}
