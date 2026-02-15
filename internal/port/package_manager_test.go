package port_test

import (
	"context"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestPackageManagerInterface verifies that the PackageManager interface contract
// can be satisfied by a mock implementation.
// Requirements: 11.1, 11.3
func TestPackageManagerInterface(t *testing.T) {
	t.Run("interface_contract", func(t *testing.T) {
		// Verify that a mock implementation satisfies the interface
		var _ port.PackageManager = &mockPackageManager{}
	})
}

// TestSourceValidation tests Source struct validation.
// Requirements: 11.4
func TestSourceValidation(t *testing.T) {
	tests := []struct {
		source  *port.Source
		name    string
		wantErr bool
	}{
		{
			name: "valid_git_source",
			source: &port.Source{
				Type: "git",
				URL:  "https://github.com/example/skill.git",
			},
			wantErr: false,
		},
		{
			name: "valid_npm_source",
			source: &port.Source{
				Type: "npm",
				URL:  "example-skill",
			},
			wantErr: false,
		},
		{
			name: "valid_go_module_source",
			source: &port.Source{
				Type: "go-module",
				URL:  "github.com/example/skill",
			},
			wantErr: false,
		},
		{
			name: "empty_type",
			source: &port.Source{
				Type: "",
				URL:  "https://github.com/example/skill.git",
			},
			wantErr: true,
		},
		{
			name: "empty_url",
			source: &port.Source{
				Type: "git",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "invalid_type",
			source: &port.Source{
				Type: "invalid",
				URL:  "https://example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Source.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDownloadResultStructure tests DownloadResult struct fields.
// Requirements: 3.1, 4.1, 4.2
func TestDownloadResultStructure(t *testing.T) {
	result := &port.DownloadResult{
		Path:    "/tmp/skill",
		Version: "v1.0.0",
	}

	if result.Path != "/tmp/skill" {
		t.Errorf("DownloadResult.Path = %v, want %v", result.Path, "/tmp/skill")
	}
	if result.Version != "v1.0.0" {
		t.Errorf("DownloadResult.Version = %v, want %v", result.Version, "v1.0.0")
	}
}

// mockPackageManager is a mock implementation of PackageManager for testing.
type mockPackageManager struct{}

func (m *mockPackageManager) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	return &port.DownloadResult{
		Path:    "/tmp/mock",
		Version: version,
	}, nil
}

func (m *mockPackageManager) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return "latest", nil
}

func (m *mockPackageManager) SourceType() string {
	return "mock"
}
