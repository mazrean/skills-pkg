package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// Mock PackageManager for testing
type mockPackageManager struct {
	sourceType string
}

func (m *mockPackageManager) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	return nil, nil
}

func (m *mockPackageManager) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return "", nil
}

func (m *mockPackageManager) SourceType() string {
	return m.sourceType
}

// Mock HashService for testing
type mockHashService struct{}

func (m *mockHashService) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	return &port.HashResult{
		Algorithm: "sha256",
		Value:     "mockHash123",
	}, nil
}

func (m *mockHashService) HashAlgorithm() string {
	return "sha256"
}

// TestNewSkillManager tests the creation of a new SkillManager instance.
func TestNewSkillManager(t *testing.T) {
	configManager := NewConfigManager(".skillspkg.toml")
	hashService := &mockHashService{}
	packageManagers := []port.PackageManager{
		&mockPackageManager{sourceType: "git"},
		&mockPackageManager{sourceType: "npm"},
	}

	skillManager := NewSkillManager(configManager, hashService, packageManagers)

	if skillManager == nil {
		t.Fatal("NewSkillManager returned nil")
	}
}

// TestSelectPackageManager_ValidSourceType tests selecting a package manager with a valid source type.
// Requirements: 11.4
func TestSelectPackageManager_ValidSourceType(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantType   string
	}{
		{
			name:       "select git package manager",
			sourceType: "git",
			wantType:   "git",
		},
		{
			name:       "select npm package manager",
			sourceType: "npm",
			wantType:   "npm",
		},
		{
			name:       "select go-module package manager",
			sourceType: "go-module",
			wantType:   "go-module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configManager := NewConfigManager(".skillspkg.toml")
			hashService := &mockHashService{}
			packageManagers := []port.PackageManager{
				&mockPackageManager{sourceType: "git"},
				&mockPackageManager{sourceType: "npm"},
				&mockPackageManager{sourceType: "go-module"},
			}

			skillManager := NewSkillManager(configManager, hashService, packageManagers).(*skillManagerImpl)

			pm, err := skillManager.selectPackageManager(tt.sourceType)
			if err != nil {
				t.Fatalf("selectPackageManager returned error: %v", err)
			}

			if pm.SourceType() != tt.wantType {
				t.Errorf("selectPackageManager returned wrong type: got %s, want %s", pm.SourceType(), tt.wantType)
			}
		})
	}
}

// TestSelectPackageManager_UnsupportedSourceType tests selecting a package manager with an unsupported source type.
// Requirements: 11.5, 12.2, 12.3
func TestSelectPackageManager_UnsupportedSourceType(t *testing.T) {
	configManager := NewConfigManager(".skillspkg.toml")
	hashService := &mockHashService{}
	packageManagers := []port.PackageManager{
		&mockPackageManager{sourceType: "git"},
		&mockPackageManager{sourceType: "npm"},
	}

	skillManager := NewSkillManager(configManager, hashService, packageManagers).(*skillManagerImpl)

	pm, err := skillManager.selectPackageManager("unsupported")
	if err == nil {
		t.Fatal("selectPackageManager should return error for unsupported source type")
	}

	if pm != nil {
		t.Error("selectPackageManager should return nil for unsupported source type")
	}

	// Verify error is wrapped with ErrInvalidSource
	if !errors.Is(err, ErrInvalidSource) {
		t.Errorf("selectPackageManager should return ErrInvalidSource, got: %v", err)
	}
}

// TestSelectPackageManager_EmptySourceType tests selecting a package manager with an empty source type.
// Requirements: 11.5, 12.2, 12.3
func TestSelectPackageManager_EmptySourceType(t *testing.T) {
	configManager := NewConfigManager(".skillspkg.toml")
	hashService := &mockHashService{}
	packageManagers := []port.PackageManager{
		&mockPackageManager{sourceType: "git"},
	}

	skillManager := NewSkillManager(configManager, hashService, packageManagers).(*skillManagerImpl)

	pm, err := skillManager.selectPackageManager("")
	if err == nil {
		t.Fatal("selectPackageManager should return error for empty source type")
	}

	if pm != nil {
		t.Error("selectPackageManager should return nil for empty source type")
	}

	// Verify error is wrapped with ErrInvalidSource
	if !errors.Is(err, ErrInvalidSource) {
		t.Errorf("selectPackageManager should return ErrInvalidSource, got: %v", err)
	}
}
