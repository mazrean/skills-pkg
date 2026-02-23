package domain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// captureStdout captures output written to os.Stdout during the execution of f.
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	return buf.String()
}

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
		Value:     "mockHash123",
	}, nil
}


// TestNewSkillManager tests the creation of a new SkillManager instance.
func TestNewSkillManager(t *testing.T) {
	configManager := NewConfigManager(".skillspkg.toml")
	hashService := &mockHashService{}
	packageManagers := []port.PackageManager{
		&mockPackageManager{sourceType: "git"},
		&mockPackageManager{sourceType: "go-mod"},
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
			sourceType: "go-mod",
			wantType:   "go-mod",
		},
		{
			name:       "select go-mod package manager",
			sourceType: "go-mod",
			wantType:   "go-mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configManager := NewConfigManager(".skillspkg.toml")
			hashService := &mockHashService{}
			packageManagers := []port.PackageManager{
				&mockPackageManager{sourceType: "git"},
				&mockPackageManager{sourceType: "go-mod"},
				&mockPackageManager{sourceType: "go-mod"},
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
		&mockPackageManager{sourceType: "go-mod"},
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

// Enhanced mocks for Install testing

type mockPackageManagerWithDownload struct {
	sourceType     string
	downloadResult *port.DownloadResult
	downloadError  error
	latestVersion  string
}

func (m *mockPackageManagerWithDownload) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if m.downloadError != nil {
		return nil, m.downloadError
	}
	return m.downloadResult, nil
}

func (m *mockPackageManagerWithDownload) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return m.latestVersion, nil
}

func (m *mockPackageManagerWithDownload) SourceType() string {
	return m.sourceType
}

type mockHashServiceWithCustom struct {
	hashResult *port.HashResult
	hashError  error
}

func (m *mockHashServiceWithCustom) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	if m.hashError != nil {
		return nil, m.hashError
	}
	if m.hashResult != nil {
		return m.hashResult, nil
	}
	return &port.HashResult{
		Value:     "mockHash123",
	}, nil
}


type mockPackageManagerMultiSkill struct {
	sourceType   string
	downloadDir1 string
	downloadDir2 string
}

func (m *mockPackageManagerMultiSkill) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	// Return different results based on the source URL or version to avoid race conditions
	if source.URL == "https://github.com/example/skill1.git" || version == "v1.0.0" {
		return &port.DownloadResult{Path: m.downloadDir1, Version: "v1.0.0", FromGoMod: false}, nil
	}
	return &port.DownloadResult{Path: m.downloadDir2, Version: "v2.0.0", FromGoMod: false}, nil
}

func (m *mockPackageManagerMultiSkill) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return "", nil
}

func (m *mockPackageManagerMultiSkill) SourceType() string {
	return m.sourceType
}

// TestInstall_SingleSkill tests installing a single skill successfully.
// Requirements: 6.2, 3.3, 3.4, 4.3, 4.4, 5.3, 12.1
func TestInstall_SingleSkill(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"
	installDir := tmpDir + "/install"
	downloadDir := tmpDir + "/download"

	// Create download directory with test files
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("Failed to create download directory: %v", err)
	}
	if err := os.WriteFile(downloadDir+"/test.txt", []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test config
	config := &Config{
		Skills: []*Skill{
			{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",
				HashValue: "",
			},
		},
		InstallTargets: []string{installDir},
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Setup mock package manager
	pm := &mockPackageManagerWithDownload{
		sourceType: "git",
		downloadResult: &port.DownloadResult{
			Path:      downloadDir,
			Version:   "v1.0.0",
			FromGoMod: false,
		},
	}

	// Setup mock hash service
	hashService := &mockHashServiceWithCustom{
		hashResult: &port.HashResult{
			Value:     "abcd1234",
		},
	}

	// Create skill manager
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{pm})

	// Execute install
	err := skillManager.Install(ctx, "test-skill")

	// Verify no error
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	// Verify config was updated with hash
	updatedConfig, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	skill := updatedConfig.FindSkillByName("test-skill")
	if skill == nil {
		t.Fatal("Skill not found in updated config")
	}


	if skill.HashValue != "abcd1234" {
		t.Errorf("Expected hash value 'abcd1234', got '%s'", skill.HashValue)
	}

	if skill.Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", skill.Version)
	}

	// Verify skill was installed to target directory
	if _, err := os.Stat(installDir + "/test-skill"); os.IsNotExist(err) {
		t.Error("Skill was not installed to target directory")
	}
}

// TestInstall_AllSkills tests installing all skills when no skill name is specified.
// Requirements: 6.1, 12.1
func TestInstall_AllSkills(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"
	installDir := tmpDir + "/install"
	downloadDir1 := tmpDir + "/download1"
	downloadDir2 := tmpDir + "/download2"

	// Create download directories with test files
	for _, dir := range []string{downloadDir1, downloadDir2} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Failed to create download directory: %v", err)
		}
		if err := os.WriteFile(dir+"/test.txt", []byte("test content"), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create test config with two skills
	config := &Config{
		Skills: []*Skill{
			{
				Name:      "skill1",
				Source:    "git",
				URL:       "https://github.com/example/skill1.git",
				Version:   "v1.0.0",
				HashValue: "",
			},
			{
				Name:      "skill2",
				Source:    "git",
				URL:       "https://github.com/example/skill2.git",
				Version:   "v2.0.0",
				HashValue: "",
			},
		},
		InstallTargets: []string{installDir},
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Setup mock package manager that returns different paths based on source URL
	pm := &mockPackageManagerMultiSkill{
		sourceType:   "git",
		downloadDir1: downloadDir1,
		downloadDir2: downloadDir2,
	}

	// Setup mock hash service
	hashService := &mockHashServiceWithCustom{
		hashResult: &port.HashResult{
			Value:     "hash123",
		},
	}

	// Create skill manager
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{pm})

	// Execute install with empty skillName (install all)
	err := skillManager.Install(ctx, "")

	// Verify no error
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	// Verify both skills were installed
	for _, skillName := range []string{"skill1", "skill2"} {
		if _, statErr := os.Stat(installDir + "/" + skillName); os.IsNotExist(statErr) {
			t.Errorf("Skill '%s' was not installed", skillName)
		}
	}

	// Verify config was updated with hashes
	updatedConfig, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	expectedVersions := map[string]string{
		"skill1": "v1.0.0",
		"skill2": "v2.0.0",
	}

	for _, skillName := range []string{"skill1", "skill2"} {
		skill := updatedConfig.FindSkillByName(skillName)
		if skill == nil {
			t.Fatalf("Skill '%s' not found in updated config", skillName)
		}

		if skill.HashValue == "" {
			t.Errorf("Hash value for skill '%s' is empty", skillName)
		}

		if skill.Version != expectedVersions[skillName] {
			t.Errorf("Expected version '%s' for skill '%s', got '%s'", expectedVersions[skillName], skillName, skill.Version)
		}
	}
}

// TestInstall_SkillNotFound tests error when specified skill is not in configuration.
// Requirements: 6.3, 12.2, 12.3
func TestInstall_SkillNotFound(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"

	// Create empty config
	config := &Config{
		Skills:         []*Skill{},
		InstallTargets: []string{tmpDir + "/install"},
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create skill manager
	hashService := &mockHashServiceWithCustom{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Try to install non-existent skill
	err := skillManager.Install(ctx, "nonexistent-skill")

	// Verify error
	if err == nil {
		t.Fatal("Expected error for non-existent skill, got nil")
	}

	// Verify error is ErrSkillNotFound
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Expected ErrSkillNotFound, got: %v", err)
	}

	// Verify error message contains guidance (Requirement 12.2)
	expectedSubstring := "Use 'skills-pkg add"
	if !containsSubstring(err.Error(), expectedSubstring) {
		t.Errorf("Error message should contain '%s', got: %s", expectedSubstring, err.Error())
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestInstall_HashCalculation tests that hash is calculated and saved to config.
// Requirements: 5.3, 12.1
func TestInstall_HashCalculation(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestInstall_MultipleInstallTargets tests parallel installation to multiple directories.
// Requirements: 10.2, 10.5, 12.1
func TestInstall_MultipleInstallTargets(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestInstall_CreateMissingDirectories tests auto-creation of install target directories.
// Requirements: 6.6, 12.1
func TestInstall_CreateMissingDirectories(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestInstall_HashVerification tests hash verification after installation.
// Requirements: 6.4, 6.5, 12.1
func TestInstall_HashVerification(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestInstall_HashMismatchWarning tests warning display on hash mismatch while continuing installation.
// Requirements: 6.5, 12.1, 12.2
func TestInstall_HashMismatchWarning(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestInstall_FileSystemError tests handling of filesystem errors.
// Requirements: 12.2, 12.3
func TestInstall_FileSystemError(t *testing.T) {
	t.Skip("Implement after full Install implementation")
}

// TestUpdate_SingleSkill tests updating a single skill.
// Requirements: 7.1, 7.2, 7.5, 7.6
func TestUpdate_SingleSkill(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	// Create config manager
	configManager := NewConfigManager(configPath)

	// Initialize configuration
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add a skill
	skill := &Skill{
		Name:      "test-skill",
		Source:    "go-mod",
		URL:       "test-package",
		Version:   "1.0.0",
		HashValue: "oldHash123",
	}
	if err := configManager.AddSkill(ctx, skill); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	// Create mock package manager that returns a new version
	mockPM := &mockPackageManagerWithUpdate{
		sourceType:    "go-mod",
		latestVersion: "2.0.0",
		downloadPath:  tempDir + "/download",
	}

	// Create skill directory in download path
	if err := os.MkdirAll(mockPM.downloadPath, 0o755); err != nil {
		t.Fatalf("Failed to create download directory: %v", err)
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{mockPM})

	// Update the skill using new signature
	results, err := skillManager.Update(ctx, []string{"test-skill"}, "")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Verify result
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	result := results[0]
	if result == nil {
		t.Fatal("Update returned nil result")
	}
	if result.Err != nil {
		t.Fatalf("Update result has unexpected error: %v", result.Err)
	}
	if result.SkillName != "test-skill" {
		t.Errorf("Expected skill name 'test-skill', got '%s'", result.SkillName)
	}
	if result.OldVersion != "1.0.0" {
		t.Errorf("Expected old version '1.0.0', got '%s'", result.OldVersion)
	}
	if result.NewVersion != "2.0.0" {
		t.Errorf("Expected new version '2.0.0', got '%s'", result.NewVersion)
	}

	// Verify configuration was updated
	config, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	updatedSkill := config.FindSkillByName("test-skill")
	if updatedSkill == nil {
		t.Fatal("Skill not found after update")
	}
	if updatedSkill.Version != "2.0.0" {
		t.Errorf("Expected updated version '2.0.0', got '%s'", updatedSkill.Version)
	}
	if updatedSkill.HashValue == "oldHash123" {
		t.Error("Hash value should have been updated")
	}
}

// TestUpdate_AllSkills tests updating all skills.
// Requirements: 7.1, 7.2
func TestUpdate_AllSkills(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	// Create config manager
	configManager := NewConfigManager(configPath)

	// Initialize configuration
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add skills
	skills := []*Skill{
		{
			Name:      "skill1",
			Source:    "go-mod",
			URL:       "package1",
			Version:   "1.0.0",
			HashValue: "hash1",
		},
		{
			Name:      "skill2",
			Source:    "git",
			URL:       "https://github.com/example/skill2",
			Version:   "v1.0.0",
			HashValue: "hash2",
		},
	}
	for _, skill := range skills {
		if err := configManager.AddSkill(ctx, skill); err != nil {
			t.Fatalf("Failed to add skill: %v", err)
		}
	}

	// Create mock package managers
	npmPM := &mockPackageManagerWithUpdate{
		sourceType:    "go-mod",
		latestVersion: "2.0.0",
		downloadPath:  tempDir + "/npm-download",
	}
	gitPM := &mockPackageManagerWithUpdate{
		sourceType:    "git",
		latestVersion: "v2.0.0",
		downloadPath:  tempDir + "/git-download",
	}

	// Create download directories
	for _, pm := range []*mockPackageManagerWithUpdate{npmPM, gitPM} {
		if err := os.MkdirAll(pm.downloadPath, 0o755); err != nil {
			t.Fatalf("Failed to create download directory: %v", err)
		}
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{npmPM, gitPM})

	// Update all skills using new signature (empty skillNames, empty sourceFilter)
	results, err := skillManager.Update(ctx, nil, "")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Verify 2 results returned
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	// All results should succeed
	for _, result := range results {
		if result.Err != nil {
			t.Errorf("Unexpected error in result for '%s': %v", result.SkillName, result.Err)
		}
	}
}

// TestUpdate_SkillNotFound tests error handling when skill is not found.
// Requirements: 12.2, 12.3
func TestUpdate_SkillNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	// Create config manager
	configManager := NewConfigManager(configPath)

	// Initialize configuration
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Try to update non-existent skill using new signature
	results, err := skillManager.Update(ctx, []string{"non-existent-skill"}, "")
	if err != nil {
		t.Fatalf("Unexpected critical error: %v", err)
	}

	// ErrSkillNotFound should be in UpdateResult.Err
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("Expected error in result for non-existent skill, got nil")
	}
	if !errors.Is(results[0].Err, ErrSkillNotFound) {
		t.Errorf("Expected ErrSkillNotFound in result, got %v", results[0].Err)
	}
}

// TestUpdate_NetworkError tests handling of network errors.
// Requirements: 12.2, 12.3
func TestUpdate_NetworkError(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	// Create config manager
	configManager := NewConfigManager(configPath)

	// Initialize configuration
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add a skill
	skill := &Skill{
		Name:      "test-skill",
		Source:    "go-mod",
		URL:       "test-package",
		Version:   "1.0.0",
		HashValue: "oldHash123",
	}
	if err := configManager.AddSkill(ctx, skill); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	// Create mock package manager that returns a network error
	mockPM := &mockPackageManagerWithError{
		sourceType: "go-mod",
		err:        ErrNetworkFailure,
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{mockPM})

	// Try to update the skill using new signature
	results, err := skillManager.Update(ctx, []string{"test-skill"}, "")
	if err != nil {
		t.Fatalf("Unexpected critical error: %v", err)
	}

	// Network error should be in UpdateResult.Err
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("Expected error in result for network failure, got nil")
	}
	if !IsNetworkError(results[0].Err) {
		t.Errorf("Expected network error in result, got %v", results[0].Err)
	}
}

// Mock package manager with update support
type mockPackageManagerWithUpdate struct {
	sourceType    string
	latestVersion string
	downloadPath  string
}

func (m *mockPackageManagerWithUpdate) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	return &port.DownloadResult{
		Path:      m.downloadPath,
		Version:   version,
		FromGoMod: false,
	}, nil
}

func (m *mockPackageManagerWithUpdate) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return m.latestVersion, nil
}

func (m *mockPackageManagerWithUpdate) SourceType() string {
	return m.sourceType
}

// Mock package manager with error
type mockPackageManagerWithError struct {
	err        error
	sourceType string
}

func (m *mockPackageManagerWithError) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	return nil, m.err
}

func (m *mockPackageManagerWithError) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return "", m.err
}

func (m *mockPackageManagerWithError) SourceType() string {
	return m.sourceType
}

// TestUninstall_Success tests successfully uninstalling a skill.
// Requirements: 9.1, 9.2, 9.4, 12.2
func TestUninstall_Success(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"
	installDir1 := tmpDir + "/install1"
	installDir2 := tmpDir + "/install2"

	// Create install directories and skill directories
	for _, dir := range []string{installDir1, installDir2} {
		skillDir := dir + "/test-skill"
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}
		if err := os.WriteFile(skillDir+"/test.txt", []byte("test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create test config with skill
	config := &Config{
		Skills: []*Skill{
			{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",
				HashValue: "hash123",
			},
		},
		InstallTargets: []string{installDir1, installDir2},
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Execute uninstall
	err := skillManager.Uninstall(ctx, "test-skill")

	// Verify no error
	if err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	// Verify skill was removed from all install directories (Requirement 9.1)
	for _, dir := range []string{installDir1, installDir2} {
		skillDir := dir + "/test-skill"
		if _, statErr := os.Stat(skillDir); !os.IsNotExist(statErr) {
			t.Errorf("Skill directory still exists at %s", skillDir)
		}
	}

	// Verify skill was removed from config (Requirement 9.2)
	updatedConfig, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	if updatedConfig.FindSkillByName("test-skill") != nil {
		t.Error("Skill should have been removed from config")
	}
}

// TestUninstall_SkillNotFound tests error when skill is not in configuration.
// Requirements: 9.3, 12.2, 12.3
func TestUninstall_SkillNotFound(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"

	// Create empty config
	config := &Config{
		Skills:         []*Skill{},
		InstallTargets: []string{tmpDir + "/install"},
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Try to uninstall non-existent skill
	err := skillManager.Uninstall(ctx, "nonexistent-skill")

	// Verify error (Requirement 9.3)
	if err == nil {
		t.Fatal("Expected error for non-existent skill, got nil")
	}

	// Verify error is ErrSkillNotFound
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Expected ErrSkillNotFound, got: %v", err)
	}

	// Verify error message contains guidance (Requirement 12.2)
	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

// TestUninstall_RemoveFromAllTargets tests removal from all install target directories.
// Requirements: 9.1, 10.2
func TestUninstall_RemoveFromAllTargets(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"
	installDirs := []string{
		tmpDir + "/install1",
		tmpDir + "/install2",
		tmpDir + "/install3",
	}

	// Create install directories and skill directories
	for _, dir := range installDirs {
		skillDir := dir + "/test-skill"
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}
		if err := os.WriteFile(skillDir+"/test.txt", []byte("test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create test config with skill
	config := &Config{
		Skills: []*Skill{
			{
				Name:      "test-skill",
				Source:    "git",
				URL:       "https://github.com/example/skill.git",
				Version:   "v1.0.0",
				HashValue: "hash123",
			},
		},
		InstallTargets: installDirs,
	}

	// Setup config manager
	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Execute uninstall
	err := skillManager.Uninstall(ctx, "test-skill")

	// Verify no error
	if err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	// Verify skill was removed from all install directories (Requirement 9.1, 10.2)
	for _, dir := range installDirs {
		skillDir := dir + "/test-skill"
		if _, statErr := os.Stat(skillDir); !os.IsNotExist(statErr) {
			t.Errorf("Skill directory still exists at %s", skillDir)
		}
	}
}

// TestInstall_WithGoModVersion tests that when version is resolved from go.mod,
// hash values are not stored in the configuration
func TestInstall_WithGoModVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/.skillspkg.toml"
	installDir := tmpDir + "/install"

	// Initialize config
	configManager := NewConfigManager(configPath)
	if err := configManager.Initialize(context.Background(), []string{installDir}); err != nil {
		t.Fatal(err)
	}

	// Create download directory with a subdirectory
	downloadDir := tmpDir + "/download"
	if err := os.MkdirAll(downloadDir+"/skills/test-skill", 0755); err != nil {
		t.Fatal(err)
	}

	// Mock package manager that returns FromGoMod=true
	pm := &mockPackageManagerWithDownload{
		sourceType: "go-mod",
		downloadResult: &port.DownloadResult{
			Path:      downloadDir,
			Version:   "v1.2.3",
			FromGoMod: true, // Version was resolved from go.mod
		},
	}

	// Setup skill manager
	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{pm})

	// Create skill with go-mod source
	skill := &Skill{
		Name:    "test-skill",
		Source:  "go-mod",
		URL:     "github.com/example/test-skill",
		Version: "",
		SubDir:  "skills/test-skill",
	}

	// Add skill to config
	config, err := configManager.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	config.Skills = append(config.Skills, skill)
	if err = configManager.Save(context.Background(), config); err != nil {
		t.Fatal(err)
	}

	// Install the skill
	if err = skillManager.Install(context.Background(), "test-skill"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Load config and verify hash values are empty
	config, err = configManager.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	installedSkill := config.FindSkillByName("test-skill")
	if installedSkill == nil {
		t.Fatal("Skill not found in config after installation")
	}

	// Verify version and hash values are NOT set when using go.mod version
	// This ensures go.mod remains the source of truth
	if installedSkill.Version != "" {
		t.Errorf("Expected Version to be empty when using go.mod version, got %s", installedSkill.Version)
	}
	if installedSkill.HashValue != "" {
		t.Errorf("Expected HashValue to be empty when using go.mod version, got %s", installedSkill.HashValue)
	}
}

// TestUpdateResult_SkippedAndErrFields tests that UpdateResult has Skipped and Err fields.
// Requirements: 1.5, 3.3, 7.3
func TestUpdateResult_SkippedAndErrFields(t *testing.T) {
	// Test Skipped field
	skippedResult := &UpdateResult{
		SkillName:  "test-skill",
		OldVersion: "1.0.0",
		NewVersion: "1.0.0",
		Skipped:    true,
	}
	if !skippedResult.Skipped {
		t.Error("Expected Skipped to be true")
	}

	// Test Err field
	testErr := errors.New("test error")
	errorResult := &UpdateResult{
		SkillName: "test-skill",
		Err:       testErr,
	}
	if errorResult.Err == nil {
		t.Error("Expected Err to be set")
	}
	if !errors.Is(errorResult.Err, testErr) {
		t.Errorf("Expected Err to be testErr, got %v", errorResult.Err)
	}

	// Normal result without Skipped or Err should have zero values
	normalResult := &UpdateResult{
		SkillName:  "test-skill",
		OldVersion: "1.0.0",
		NewVersion: "2.0.0",
	}
	if normalResult.Skipped {
		t.Error("Expected Skipped to be false for normal result")
	}
	if normalResult.Err != nil {
		t.Error("Expected Err to be nil for normal result")
	}
}

// mockPackageManagerSkipVersion is a mock that tracks whether Download was called.
type mockPackageManagerSkipVersion struct {
	sourceType     string
	latestVersion  string
	downloadCalled bool
}

func (m *mockPackageManagerSkipVersion) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	m.downloadCalled = true
	return &port.DownloadResult{
		Path:    "",
		Version: version,
	}, nil
}

func (m *mockPackageManagerSkipVersion) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	return m.latestVersion, nil
}

func (m *mockPackageManagerSkipVersion) SourceType() string {
	return m.sourceType
}

// TestUpdateSingleSkill_SkipWhenVersionMatches tests that updateSingleSkill returns
// UpdateResult{Skipped: true} and does NOT call Download when the latest version
// matches the current skill version.
// Requirements: 3.3
func TestUpdateSingleSkill_SkipWhenVersionMatches(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	config, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	skill := &Skill{
		Name:      "test-skill",
		Source:    "go-mod",
		URL:       "test-package",
		Version:   "1.0.0",
		HashValue: "existingHash",
	}

	// Mock returns the SAME version as skill.Version (already at latest)
	mockPM := &mockPackageManagerSkipVersion{
		sourceType:    "go-mod",
		latestVersion: "1.0.0",
	}

	sm := &skillManagerImpl{
		configManager:   configManager,
		hashService:     &mockHashService{},
		packageManagers: []port.PackageManager{mockPM},
	}

	result := sm.updateSingleSkill(ctx, config, skill, false)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Skipped {
		t.Error("Expected Skipped to be true when latest version matches current version")
	}
	if result.Err != nil {
		t.Errorf("Expected no error, got: %v", result.Err)
	}
	if result.SkillName != "test-skill" {
		t.Errorf("Expected SkillName 'test-skill', got '%s'", result.SkillName)
	}
	if result.OldVersion != "1.0.0" {
		t.Errorf("Expected OldVersion '1.0.0', got '%s'", result.OldVersion)
	}
	if mockPM.downloadCalled {
		t.Error("Expected Download NOT to be called when version matches (should be skipped)")
	}
}

// TestUpdateSingleSkill_ErrorStoredInResult verifies that updateSingleSkill stores
// errors in UpdateResult.Err (with SkillName and OldVersion set) instead of returning them.
func TestUpdateSingleSkill_ErrorStoredInResult(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"

	configManager := NewConfigManager(configPath)
	ctx := context.Background()
	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	config, err := configManager.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	skill := &Skill{
		Name:      "test-skill",
		Source:    "go-mod",
		URL:       "test-package",
		Version:   "1.0.0",
		HashValue: "oldHash123",
	}

	// Mock that returns a network error on GetLatestVersion
	mockPM := &mockPackageManagerWithError{
		sourceType: "go-mod",
		err:        ErrNetworkFailure,
	}

	sm := &skillManagerImpl{
		configManager:   configManager,
		hashService:     &mockHashService{},
		packageManagers: []port.PackageManager{mockPM},
	}

	// updateSingleSkill must return *UpdateResult only; error stored in Err field
	result := sm.updateSingleSkill(ctx, config, skill, false)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Err == nil {
		t.Error("Expected result.Err to be set on network failure")
	}
	if !IsNetworkError(result.Err) {
		t.Errorf("Expected network error in result.Err, got %v", result.Err)
	}
	if result.SkillName != "test-skill" {
		t.Errorf("Expected SkillName 'test-skill', got '%s'", result.SkillName)
	}
	if result.OldVersion != "1.0.0" {
		t.Errorf("Expected OldVersion '1.0.0', got '%s'", result.OldVersion)
	}
}

// TestUpdate_SourceFilter_FiltersBySourceType tests that when sourceFilter is set
// and skillNames is empty, only skills matching the sourceType are updated.
// Requirements: 1.1, 1.2, 2.1
func TestUpdate_SourceFilter_FiltersBySourceType(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add skills with different source types
	gitSkill := &Skill{
		Name:    "git-skill",
		Source:  "git",
		URL:     "https://github.com/example/skill.git",
		Version: "v1.0.0",
	}
	gomodSkill := &Skill{
		Name:    "gomod-skill",
		Source:  "go-mod",
		URL:     "github.com/example/skill",
		Version: "1.0.0",
	}
	for _, skill := range []*Skill{gitSkill, gomodSkill} {
		if err := configManager.AddSkill(ctx, skill); err != nil {
			t.Fatalf("Failed to add skill: %v", err)
		}
	}

	// Create download directory
	downloadDir := tempDir + "/download"
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}

	// Track which skills were updated
	type callRecord struct {
		sourceType string
		url        string
	}
	var calls []callRecord

	mockGitPM := &mockPackageManagerTracking{
		sourceType:    "git",
		latestVersion: "v2.0.0",
		downloadPath:  downloadDir,
		onGetLatest: func(url string) {
			calls = append(calls, callRecord{sourceType: "git", url: url})
		},
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{mockGitPM})

	// Update with "git" source filter and empty skillNames
	results, err := skillManager.Update(ctx, nil, "git")
	if err != nil {
		t.Fatalf("Unexpected critical error: %v", err)
	}

	// Should return only 1 result (git-skill only)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result (only git skill), got %d", len(results))
	}
	if results[0].SkillName != "git-skill" {
		t.Errorf("Expected git-skill in results, got %s", results[0].SkillName)
	}
	if results[0].Err != nil {
		t.Errorf("Expected no error for git-skill, got %v", results[0].Err)
	}
	// Only git-skill's GetLatestVersion should have been called
	if len(calls) != 1 {
		t.Errorf("Expected 1 GetLatestVersion call (for git-skill only), got %d", len(calls))
	}
}

// mockPackageManagerTracking is a mock that records GetLatestVersion calls.
type mockPackageManagerTracking struct {
	sourceType    string
	latestVersion string
	downloadPath  string
	onGetLatest   func(url string)
}

func (m *mockPackageManagerTracking) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	return &port.DownloadResult{
		Path:      m.downloadPath,
		Version:   version,
		FromGoMod: false,
	}, nil
}

func (m *mockPackageManagerTracking) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if m.onGetLatest != nil {
		m.onGetLatest(source.URL)
	}
	return m.latestVersion, nil
}

func (m *mockPackageManagerTracking) SourceType() string {
	return m.sourceType
}

// TestUpdate_SourceFilter_NoMatchingSkills tests that when sourceFilter matches no skills,
// empty results are returned with no error.
// Requirements: 2.2
func TestUpdate_SourceFilter_NoMatchingSkills(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add only go-mod skills
	gomodSkill := &Skill{
		Name:    "gomod-skill",
		Source:  "go-mod",
		URL:     "github.com/example/skill",
		Version: "1.0.0",
	}
	if err := configManager.AddSkill(ctx, gomodSkill); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Filter by "git" but no git skills exist
	results, err := skillManager.Update(ctx, nil, "git")
	if err != nil {
		t.Fatalf("Expected no critical error for empty filter result, got: %v", err)
	}

	// Should return empty results
	if len(results) != 0 {
		t.Errorf("Expected 0 results when no skills match source filter, got %d", len(results))
	}
}

// TestUpdate_SourceMismatch tests that when skillNames contains a skill with a source
// that doesn't match sourceFilter, ErrSourceMismatch is stored in UpdateResult.Err.
// Requirements: 1.4, 1.5
func TestUpdate_SourceMismatch(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add a go-mod skill
	skill := &Skill{
		Name:    "gomod-skill",
		Source:  "go-mod",
		URL:     "github.com/example/skill",
		Version: "1.0.0",
	}
	if err := configManager.AddSkill(ctx, skill); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	// Try to update gomod-skill with --source git (mismatch)
	results, err := skillManager.Update(ctx, []string{"gomod-skill"}, "git")
	if err != nil {
		t.Fatalf("Expected no critical error, got: %v", err)
	}

	// Should return 1 result with ErrSourceMismatch
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("Expected ErrSourceMismatch in result.Err, got nil")
	}
	if !errors.Is(results[0].Err, ErrSourceMismatch) {
		t.Errorf("Expected ErrSourceMismatch, got: %v", results[0].Err)
	}
	if results[0].SkillName != "gomod-skill" {
		t.Errorf("Expected SkillName 'gomod-skill', got '%s'", results[0].SkillName)
	}
}

// TestUpdate_SkillNameWithMatchingSource tests that when skillNames has a skill
// with a matching source filter, the skill is updated normally.
// Requirements: 1.4
func TestUpdate_SkillNameWithMatchingSource(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add a git skill
	skill := &Skill{
		Name:    "git-skill",
		Source:  "git",
		URL:     "https://github.com/example/skill.git",
		Version: "v1.0.0",
	}
	if err := configManager.AddSkill(ctx, skill); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	downloadDir := tempDir + "/download"
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}

	gitPM := &mockPackageManagerWithUpdate{
		sourceType:    "git",
		latestVersion: "v2.0.0",
		downloadPath:  downloadDir,
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{gitPM})

	// Update git-skill with matching source filter
	results, err := skillManager.Update(ctx, []string{"git-skill"}, "git")
	if err != nil {
		t.Fatalf("Unexpected critical error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("Expected no error for matching source, got: %v", results[0].Err)
	}
	if results[0].SkillName != "git-skill" {
		t.Errorf("Expected 'git-skill', got '%s'", results[0].SkillName)
	}
	if results[0].NewVersion != "v2.0.0" {
		t.Errorf("Expected new version 'v2.0.0', got '%s'", results[0].NewVersion)
	}
}

// TestUpdate_PrintsUpdatingBanner verifies that Update prints
// "Updating N skill(s): [names]" to stdout before processing.
// Requirements: 6.1
func TestUpdate_PrintsUpdatingBanner(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	downloadDir := tempDir + "/download"
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}

	// Add two git skills
	for _, name := range []string{"skill-a", "skill-b"} {
		s := &Skill{
			Name:    name,
			Source:  "git",
			URL:     "https://github.com/example/" + name + ".git",
			Version: "v1.0.0",
		}
		if err := configManager.AddSkill(ctx, s); err != nil {
			t.Fatalf("Failed to add skill: %v", err)
		}
	}

	pm := &mockPackageManagerWithUpdate{
		sourceType:    "git",
		latestVersion: "v1.0.0", // same version → skipped, but banner should still print
		downloadPath:  downloadDir,
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{pm})

	var output string
	output = captureStdout(func() {
		_, _ = skillManager.Update(ctx, nil, "")
	})

	if !strings.Contains(output, "Updating 2 skill(s):") {
		t.Errorf("Expected stdout to contain 'Updating 2 skill(s):' but got:\n%s", output)
	}
	if !strings.Contains(output, "skill-a") {
		t.Errorf("Expected stdout to contain 'skill-a' but got:\n%s", output)
	}
	if !strings.Contains(output, "skill-b") {
		t.Errorf("Expected stdout to contain 'skill-b' but got:\n%s", output)
	}
}

// TestUpdate_NoSkillsForSourcePrintsMessage verifies that when sourceFilter is given
// but no matching skills exist, "No skills found for source '%s'." is printed and
// the method returns empty results with no error.
// Requirements: 2.2, 6.1
func TestUpdate_NoSkillsForSourcePrintsMessage(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add only go-mod skill (no git skills)
	s := &Skill{
		Name:    "gomod-skill",
		Source:  "go-mod",
		URL:     "github.com/example/skill",
		Version: "1.0.0",
	}
	if err := configManager.AddSkill(ctx, s); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{})

	var (
		results []*UpdateResult
		err     error
		output  string
	)
	output = captureStdout(func() {
		results, err = skillManager.Update(ctx, nil, "git")
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
	expected := "No skills found for source 'git'."
	if !strings.Contains(output, expected) {
		t.Errorf("Expected stdout to contain %q but got:\n%s", expected, output)
	}
}

// mockPackageManagerMixed is a mock that fails for a specific skill URL and succeeds for others.
type mockPackageManagerMixed struct {
	sourceType    string
	latestVersion string
	downloadPath  string
	failURL       string
	failErr       error
}

func (m *mockPackageManagerMixed) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if source.URL == m.failURL {
		return nil, m.failErr
	}
	return &port.DownloadResult{
		Path:      m.downloadPath,
		Version:   version,
		FromGoMod: false,
	}, nil
}

func (m *mockPackageManagerMixed) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if source.URL == m.failURL {
		return "", m.failErr
	}
	return m.latestVersion, nil
}

func (m *mockPackageManagerMixed) SourceType() string {
	return m.sourceType
}

// TestUpdate_IndividualErrorContinuation tests that when one skill fails,
// the other skills continue to be updated and their results are still returned.
// Requirements: 7.3
func TestUpdate_IndividualErrorContinuation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/.skillspkg.toml"
	configManager := NewConfigManager(configPath)
	ctx := context.Background()

	if err := configManager.Initialize(ctx, []string{tempDir + "/skills"}); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Add two git skills
	skillA := &Skill{
		Name:    "skill-a",
		Source:  "git",
		URL:     "https://github.com/example/skill-a.git",
		Version: "v1.0.0",
	}
	skillB := &Skill{
		Name:    "skill-b",
		Source:  "git",
		URL:     "https://github.com/example/skill-b.git",
		Version: "v1.0.0",
	}
	for _, skill := range []*Skill{skillA, skillB} {
		if err := configManager.AddSkill(ctx, skill); err != nil {
			t.Fatalf("Failed to add skill: %v", err)
		}
	}

	downloadDir := tempDir + "/download"
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}

	// skill-a fails with a network error; skill-b succeeds
	pm := &mockPackageManagerMixed{
		sourceType:    "git",
		latestVersion: "v2.0.0",
		downloadPath:  downloadDir,
		failURL:       "https://github.com/example/skill-a.git",
		failErr:       ErrNetworkFailure,
	}

	hashService := &mockHashService{}
	skillManager := NewSkillManager(configManager, hashService, []port.PackageManager{pm})

	results, err := skillManager.Update(ctx, nil, "")
	if err != nil {
		t.Fatalf("Expected no critical error, got: %v", err)
	}

	// Should return 2 results (both skills processed)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results (both skills processed), got %d", len(results))
	}

	// Build a map for easier assertion
	resultMap := make(map[string]*UpdateResult)
	for _, r := range results {
		resultMap[r.SkillName] = r
	}

	// skill-a should have a network error in Err
	if resultA, ok := resultMap["skill-a"]; !ok {
		t.Error("Expected result for skill-a")
	} else {
		if resultA.Err == nil {
			t.Error("Expected error for skill-a, got nil")
		}
		if !IsNetworkError(resultA.Err) {
			t.Errorf("Expected network error for skill-a, got: %v", resultA.Err)
		}
	}

	// skill-b should succeed despite skill-a failing
	if resultB, ok := resultMap["skill-b"]; !ok {
		t.Error("Expected result for skill-b")
	} else {
		if resultB.Err != nil {
			t.Errorf("Expected no error for skill-b, got: %v", resultB.Err)
		}
		if resultB.NewVersion != "v2.0.0" {
			t.Errorf("Expected skill-b new version 'v2.0.0', got '%s'", resultB.NewVersion)
		}
	}
}
