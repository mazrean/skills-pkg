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
	adapter := NewGoModAdapter()
	if got := adapter.SourceType(); got != "go-module" {
		t.Errorf("SourceType() = %v, want %v", got, "go-module")
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
				Type: "npm",
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

func TestGoModAdapter_Download_ModuleNotFound(t *testing.T) {
	adapter := NewGoModAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "github.com/this-module-absolutely-does-not-exist-12345/nonexistent",
	}

	_, err := adapter.Download(ctx, source, "v1.0.0")
	if err == nil {
		t.Error("Download() expected error for non-existent module, got nil")
	}

	// Should be a network failure error
	if !domain.IsNetworkError(err) {
		t.Errorf("Download() error should be a network error, got %v", err)
	}
}

func TestGoModAdapter_Download_ValidModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewGoModAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "golang.org/x/exp",
	}

	result, err := adapter.Download(ctx, source, "v0.0.0-20231110203233-9a3e6036ecaa")
	if err != nil {
		t.Fatalf("Download() unexpected error: %v", err)
	}

	// Clean up
	defer func() {
		_ = os.RemoveAll(result.Path)
	}()

	// Verify download result
	if result.Path == "" {
		t.Error("Download() result.Path is empty")
	}
	if result.Version != "v0.0.0-20231110203233-9a3e6036ecaa" {
		t.Errorf("Download() result.Version = %v, want %v", result.Version, "v0.0.0-20231110203233-9a3e6036ecaa")
	}

	// Verify directory exists
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("Download() directory does not exist: %s", result.Path)
	}

	// Verify go.mod exists
	goMod := filepath.Join(result.Path, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		t.Error("Download() go.mod not found in downloaded module")
	}
}

func TestGoModAdapter_Download_LatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewGoModAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "golang.org/x/exp",
	}

	result, err := adapter.Download(ctx, source, "latest")
	if err != nil {
		t.Fatalf("Download() unexpected error: %v", err)
	}

	// Clean up
	defer func() {
		_ = os.RemoveAll(result.Path)
	}()

	// Verify download result
	if result.Path == "" {
		t.Error("Download() result.Path is empty")
	}
	if result.Version == "" || result.Version == "latest" {
		t.Errorf("Download() result.Version should be resolved, got %v", result.Version)
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
				Type: "npm",
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

func TestGoModAdapter_GetLatestVersion_ModuleNotFound(t *testing.T) {
	adapter := NewGoModAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "github.com/this-module-absolutely-does-not-exist-12345/nonexistent",
	}

	_, err := adapter.GetLatestVersion(ctx, source)
	if err == nil {
		t.Error("GetLatestVersion() expected error for non-existent module, got nil")
	}

	// Should be a network failure error
	if !domain.IsNetworkError(err) {
		t.Errorf("GetLatestVersion() error should be a network error, got %v", err)
	}
}

func TestGoModAdapter_GetLatestVersion_ValidModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewGoModAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "go-module",
		URL:  "golang.org/x/exp",
	}

	version, err := adapter.GetLatestVersion(ctx, source)
	if err != nil {
		t.Fatalf("GetLatestVersion() unexpected error: %v", err)
	}

	if version == "" {
		t.Error("GetLatestVersion() returned empty version")
	}

	// Version should start with "v" (e.g., "v0.0.0-20231110203233-9a3e6036ecaa")
	if len(version) < 2 || version[0] != 'v' {
		t.Errorf("GetLatestVersion() version seems invalid: %s", version)
	}
}
