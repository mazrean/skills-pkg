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
	adapter := NewNpmAdapter()
	if got := adapter.SourceType(); got != "npm" {
		t.Errorf("SourceType() = %v, want %v", got, "npm")
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

func TestNpmAdapter_Download_PackageNotFound(t *testing.T) {
	adapter := NewNpmAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "npm",
		URL:  "this-package-absolutely-does-not-exist-12345",
	}

	_, err := adapter.Download(ctx, source, "1.0.0")
	if err == nil {
		t.Error("Download() expected error for non-existent package, got nil")
	}

	// Should be a network failure error
	if !domain.IsNetworkError(err) {
		t.Errorf("Download() error should be a network error, got %v", err)
	}
}

func TestNpmAdapter_Download_ValidPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewNpmAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "npm",
		URL:  "lodash",
	}

	result, err := adapter.Download(ctx, source, "4.17.21")
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
	if result.Version != "4.17.21" {
		t.Errorf("Download() result.Version = %v, want %v", result.Version, "4.17.21")
	}

	// Verify directory exists
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("Download() directory does not exist: %s", result.Path)
	}

	// Verify package.json exists
	packageJSON := filepath.Join(result.Path, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		t.Error("Download() package.json not found in downloaded package")
	}
}

func TestNpmAdapter_Download_LatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewNpmAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "npm",
		URL:  "lodash",
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

func TestNpmAdapter_GetLatestVersion_PackageNotFound(t *testing.T) {
	adapter := NewNpmAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "npm",
		URL:  "this-package-absolutely-does-not-exist-12345",
	}

	_, err := adapter.GetLatestVersion(ctx, source)
	if err == nil {
		t.Error("GetLatestVersion() expected error for non-existent package, got nil")
	}

	// Should be a network failure error
	if !domain.IsNetworkError(err) {
		t.Errorf("GetLatestVersion() error should be a network error, got %v", err)
	}
}

func TestNpmAdapter_GetLatestVersion_ValidPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := NewNpmAdapter()
	ctx := context.Background()

	source := &port.Source{
		Type: "npm",
		URL:  "lodash",
	}

	version, err := adapter.GetLatestVersion(ctx, source)
	if err != nil {
		t.Fatalf("GetLatestVersion() unexpected error: %v", err)
	}

	if version == "" {
		t.Error("GetLatestVersion() returned empty version")
	}

	// Version should be in semantic version format (e.g., "4.17.21")
	if len(version) < 5 {
		t.Errorf("GetLatestVersion() version seems invalid: %s", version)
	}
}
