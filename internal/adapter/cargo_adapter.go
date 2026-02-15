// Package adapter provides implementations of port interfaces for external system integrations.
// It includes adapters for cargo (crates.io).
package adapter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

const (
	// cargoMaxPathSplitParts is the maximum number of parts to split a path into
	// when extracting the prefix directory (prefix + remaining path)
	cargoMaxPathSplitParts = 2
)

// CargoAdapter implements the PackageManager interface for crates.io (cargo).
// It handles downloading crates from crates.io, extracting them,
// and retrieving the latest version.
type CargoAdapter struct {
	httpClient  *http.Client
	registryURL string
}

// NewCargoAdapter creates a new cargo adapter instance.
// It uses the default crates.io registry (https://crates.io) unless
// overridden by the source options.
func NewCargoAdapter() *CargoAdapter {
	return &CargoAdapter{
		registryURL: "https://crates.io",
		httpClient:  &http.Client{},
	}
}

// SourceType returns "cargo" to identify this adapter as a cargo package manager.
func (a *CargoAdapter) SourceType() string {
	return "cargo"
}

// Download downloads a skill from crates.io.
// It fetches the crate metadata, downloads the crate file, and extracts it to a temporary directory.
// If version is "latest" or empty, it uses the latest version from the crate info.
func (a *CargoAdapter) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "cargo" {
		return nil, fmt.Errorf("source type must be 'cargo', got '%s'", source.Type)
	}

	// Get registry URL from source options if provided
	registryURL := a.registryURL
	if url, ok := source.Options["registry"]; ok && url != "" {
		registryURL = url
	}

	// Fetch crate metadata
	metadata, err := a.fetchCrateMetadata(ctx, registryURL, source.URL)
	if err != nil {
		return nil, err
	}

	// Resolve version
	resolvedVersion := version
	if version == "" || version == "latest" {
		resolvedVersion = metadata.Crate.MaxVersion
		if resolvedVersion == "" {
			return nil, fmt.Errorf("no latest version found for crate %s", source.URL)
		}
	}

	// Find the matching version in versions list
	var downloadPath string
	for _, v := range metadata.Versions {
		if v.Num == resolvedVersion {
			downloadPath = v.DLPath
			break
		}
	}

	if downloadPath == "" {
		return nil, fmt.Errorf("version %s not found for crate %s. Please verify the version is correct", resolvedVersion, source.URL)
	}

	// Construct full download URL
	downloadURL := fmt.Sprintf("%s%s", strings.TrimSuffix(registryURL, "/"), downloadPath)

	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := a.downloadAndExtractCrate(ctx, downloadURL, tempDir); err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &port.DownloadResult{
		Path:    tempDir,
		Version: resolvedVersion,
	}, nil
}

// GetLatestVersion retrieves the latest version from crates.io.
// It returns the version specified in the crate's max_version field.
func (a *CargoAdapter) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if err := source.Validate(); err != nil {
		return "", fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "cargo" {
		return "", fmt.Errorf("source type must be 'cargo', got '%s'", source.Type)
	}

	// Get registry URL from source options if provided
	registryURL := a.registryURL
	if url, ok := source.Options["registry"]; ok && url != "" {
		registryURL = url
	}

	// Fetch crate metadata
	metadata, err := a.fetchCrateMetadata(ctx, registryURL, source.URL)
	if err != nil {
		return "", err
	}

	// Return latest version
	if metadata.Crate.MaxVersion == "" {
		return "", fmt.Errorf("no latest version found for crate %s", source.URL)
	}

	return metadata.Crate.MaxVersion, nil
}

// crateMetadata represents the structure of crates.io crate metadata.
type crateMetadata struct {
	Crate    crateInfo       `json:"crate"`
	Versions []crateVersion  `json:"versions"`
}

// crateInfo represents the crate field in crates.io metadata.
type crateInfo struct {
	Name       string `json:"name"`
	MaxVersion string `json:"max_version"`
}

// crateVersion represents a single version in crates.io metadata.
type crateVersion struct {
	Num    string `json:"num"`
	DLPath string `json:"dl_path"`
}

// fetchCrateMetadata fetches crate metadata from crates.io.
func (a *CargoAdapter) fetchCrateMetadata(ctx context.Context, registryURL, crateName string) (*crateMetadata, error) {
	url := fmt.Sprintf("%s/api/v1/crates/%s", strings.TrimSuffix(registryURL, "/"), crateName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// crates.io requires a User-Agent header
	req.Header.Set("User-Agent", "skills-pkg (https://github.com/mazrean/skills-pkg)")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch crate metadata for %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, crateName)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: crate %s not found. Please verify the crate name is correct", domain.ErrNetworkFailure, crateName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to fetch crate metadata for %s: HTTP status %d", domain.ErrNetworkFailure, crateName, resp.StatusCode)
	}

	var metadata crateMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to parse crate metadata for %s: %w", crateName, err)
	}

	return &metadata, nil
}

// downloadAndExtractCrate downloads a crate file and extracts it to the target directory.
func (a *CargoAdapter) downloadAndExtractCrate(ctx context.Context, downloadURL, targetDir string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// crates.io requires a User-Agent header
	req.Header.Set("User-Agent", "skills-pkg (https://github.com/mazrean/skills-pkg)")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to download crate from %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, downloadURL)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to download crate from %s: HTTP status %d", domain.ErrNetworkFailure, downloadURL, resp.StatusCode)
	}

	// .crate files are tar.gz files
	if err := a.extractCrate(resp.Body, targetDir); err != nil {
		return fmt.Errorf("failed to extract crate: %w", err)
	}

	return nil
}

// extractCrate extracts a crate file (.crate is actually a tar.gz) to the target directory.
// Crate files typically have a directory prefix (crate-name-version/), which is stripped during extraction.
func (a *CargoAdapter) extractCrate(r io.Reader, targetDir string) error {
	// Create gzip reader
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Track the first directory to strip it
	var prefixDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Determine prefix directory from the first entry
		if prefixDir == "" {
			parts := strings.SplitN(header.Name, "/", cargoMaxPathSplitParts)
			if len(parts) > 0 {
				prefixDir = parts[0] + "/"
			}
		}

		// Strip the prefix directory from the path
		name, _ := strings.CutPrefix(header.Name, prefixDir)

		// Skip if the path is empty after stripping
		if name == "" {
			continue
		}

		target := filepath.Join(targetDir, name)

		// Ensure the target is within the target directory (security check)
		if !strings.HasPrefix(target, filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			target != filepath.Clean(targetDir) {
			return fmt.Errorf("invalid file path in crate: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, dirPerms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create file
			if err := os.MkdirAll(filepath.Dir(target), dirPerms); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", target, err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}

			_ = f.Close()
		}
	}

	return nil
}

// createTempDir creates a temporary directory for cargo crates.
// It uses the SKILLSPKG_TEMP_DIR environment variable if set, otherwise uses os.TempDir().
func (a *CargoAdapter) createTempDir() (string, error) {
	baseDir := os.Getenv("SKILLSPKG_TEMP_DIR")
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Generate a unique directory name using hash
	hash := sha256.New()
	pidBytes := fmt.Appendf(nil, "%d", os.Getpid())
	_, _ = hash.Write(pidBytes)
	dirName := fmt.Sprintf("skills-pkg-cargo-%x", hash.Sum(nil)[:8])

	tempDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(tempDir, dirPerms); err != nil {
		return "", err
	}

	return tempDir, nil
}
