// Package adapter provides implementations of port interfaces for external system integrations.
// It includes adapters for Git repositories, npm registry, and directory hash calculation.
package adapter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

const (
	// dirPerms is the file permission mode for directories.
	dirPerms fs.FileMode = 0o755
)

// NpmAdapter implements the PackageManager interface for npm registry.
// It handles downloading packages from npm registry, extracting tarballs,
// and retrieving the latest version.
// Requirements: 4.1, 4.5, 4.6, 7.4, 11.2
type NpmAdapter struct {
	httpClient  *http.Client
	registryURL string
}

// NewNpmAdapter creates a new npm adapter instance.
// It uses the default npm registry (https://registry.npmjs.org) unless
// overridden by the source options.
func NewNpmAdapter() *NpmAdapter {
	return &NpmAdapter{
		registryURL: "https://registry.npmjs.org",
		httpClient:  &http.Client{},
	}
}

// SourceType returns "npm" to identify this adapter as an npm package manager.
// Requirements: 11.2
func (a *NpmAdapter) SourceType() string {
	return "npm"
}

// Download downloads a skill from the npm registry.
// It fetches the package metadata, downloads the tarball, and extracts it to a temporary directory.
// If version is "latest" or empty, it uses the latest version from dist-tags.
// Requirements: 4.1, 4.5, 4.6, 12.2, 12.3
func (a *NpmAdapter) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "npm" {
		return nil, fmt.Errorf("source type must be 'npm', got '%s'", source.Type)
	}

	// Get registry URL from source options if provided
	registryURL := a.registryURL
	if url, ok := source.Options["registry"]; ok && url != "" {
		registryURL = url
	}

	// Fetch package metadata
	metadata, err := a.fetchPackageMetadata(ctx, registryURL, source.URL)
	if err != nil {
		return nil, err
	}

	// Resolve version
	resolvedVersion := version
	if version == "" || version == "latest" {
		resolvedVersion = metadata.DistTags.Latest
		if resolvedVersion == "" {
			return nil, fmt.Errorf("no latest version found for package %s", source.URL)
		}
	}

	// Get version-specific metadata
	versionMetadata, ok := metadata.Versions[resolvedVersion]
	if !ok {
		return nil, fmt.Errorf("version %s not found for package %s. Please verify the version is correct", resolvedVersion, source.URL)
	}

	// Download tarball
	tarballURL := versionMetadata.Dist.Tarball
	if tarballURL == "" {
		return nil, fmt.Errorf("no tarball URL found for package %s version %s", source.URL, resolvedVersion)
	}

	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := a.downloadAndExtractTarball(ctx, tarballURL, tempDir); err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &port.DownloadResult{
		Path:    tempDir,
		Version: resolvedVersion,
	}, nil
}

// GetLatestVersion retrieves the latest version from the npm registry.
// It returns the version specified in the "latest" dist-tag.
// Requirements: 7.4, 12.2, 12.3
func (a *NpmAdapter) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if err := source.Validate(); err != nil {
		return "", fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "npm" {
		return "", fmt.Errorf("source type must be 'npm', got '%s'", source.Type)
	}

	// Get registry URL from source options if provided
	registryURL := a.registryURL
	if url, ok := source.Options["registry"]; ok && url != "" {
		registryURL = url
	}

	// Fetch package metadata
	metadata, err := a.fetchPackageMetadata(ctx, registryURL, source.URL)
	if err != nil {
		return "", err
	}

	// Return latest version
	if metadata.DistTags.Latest == "" {
		return "", fmt.Errorf("no latest version found for package %s", source.URL)
	}

	return metadata.DistTags.Latest, nil
}

// npmPackageMetadata represents the structure of npm package metadata.
type npmPackageMetadata struct {
	Versions map[string]npmVersionMetadata `json:"versions"`
	DistTags npmDistTags                   `json:"dist-tags"`
	Name     string                        `json:"name"`
}

// npmDistTags represents the dist-tags field in npm package metadata.
type npmDistTags struct {
	Latest string `json:"latest"`
}

// npmVersionMetadata represents version-specific metadata in npm package metadata.
type npmVersionMetadata struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Dist    npmDist `json:"dist"`
}

// npmDist represents the dist field in npm version metadata.
type npmDist struct {
	Tarball string `json:"tarball"`
}

// fetchPackageMetadata fetches package metadata from the npm registry.
// Requirements: 4.1, 4.5, 4.6, 12.2, 12.3
func (a *NpmAdapter) fetchPackageMetadata(ctx context.Context, registryURL, packageName string) (*npmPackageMetadata, error) {
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(registryURL, "/"), packageName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch package metadata for %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, packageName)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: package %s not found. Please verify the package name is correct", domain.ErrNetworkFailure, packageName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to fetch package metadata for %s: HTTP status %d", domain.ErrNetworkFailure, packageName, resp.StatusCode)
	}

	var metadata npmPackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to parse package metadata for %s: %w", packageName, err)
	}

	return &metadata, nil
}

// downloadAndExtractTarball downloads a tarball and extracts it to the target directory.
// Requirements: 4.1, 4.5, 12.2, 12.3
func (a *NpmAdapter) downloadAndExtractTarball(ctx context.Context, tarballURL, targetDir string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tarballURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to download tarball from %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, tarballURL)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to download tarball from %s: HTTP status %d", domain.ErrNetworkFailure, tarballURL, resp.StatusCode)
	}

	// Extract tarball
	if err := a.extractTarball(resp.Body, targetDir); err != nil {
		return fmt.Errorf("failed to extract tarball: %w", err)
	}

	return nil
}

// extractTarball extracts a gzipped tarball to the target directory.
// npm packages are typically wrapped in a "package/" directory, which is stripped during extraction.
// Requirements: 4.1
func (a *NpmAdapter) extractTarball(r io.Reader, targetDir string) error {
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

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Strip the "package/" prefix from the path
		name, _ := strings.CutPrefix(header.Name, "package/")

		// Skip if the path is empty after stripping
		if name == "" {
			continue
		}

		target := filepath.Join(targetDir, name)

		// Ensure the target is within the target directory (security check)
		if !strings.HasPrefix(target, filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			target != filepath.Clean(targetDir) {
			return fmt.Errorf("invalid file path in tarball: %s", header.Name)
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

// createTempDir creates a temporary directory for npm packages.
// It uses the SKILLSPKG_TEMP_DIR environment variable if set, otherwise uses os.TempDir().
func (a *NpmAdapter) createTempDir() (string, error) {
	baseDir := os.Getenv("SKILLSPKG_TEMP_DIR")
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Generate a unique directory name using hash
	hash := sha256.New()
	pidBytes := fmt.Appendf(nil, "%d", os.Getpid())
	_, _ = hash.Write(pidBytes)
	dirName := fmt.Sprintf("skills-pkg-npm-%x", hash.Sum(nil)[:8])

	tempDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(tempDir, dirPerms); err != nil {
		return "", err
	}

	return tempDir, nil
}
