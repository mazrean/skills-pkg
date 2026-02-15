// Package adapter provides implementations of port interfaces for external system integrations.
// It includes adapters for pip (PyPI).
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
	// pipPathSplitLimit is the limit for splitting paths to separate the first directory from the rest
	pipPathSplitLimit = 2
)

// PipAdapter implements the PackageManager interface for PyPI (pip).
// It handles downloading packages from PyPI, extracting distributions,
// and retrieving the latest version.
type PipAdapter struct {
	httpClient *http.Client
	indexURL   string
}

// NewPipAdapter creates a new pip adapter instance.
// It uses the default PyPI index (https://pypi.org) unless
// overridden by the source options.
func NewPipAdapter() *PipAdapter {
	return &PipAdapter{
		indexURL:   "https://pypi.org",
		httpClient: &http.Client{},
	}
}

// SourceType returns "pip" to identify this adapter as a pip package manager.
func (a *PipAdapter) SourceType() string {
	return "pip"
}

// Download downloads a skill from PyPI.
// It fetches the package metadata, downloads the distribution, and extracts it to a temporary directory.
// If version is "latest" or empty, it uses the latest version from the package info.
func (a *PipAdapter) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "pip" {
		return nil, fmt.Errorf("source type must be 'pip', got '%s'", source.Type)
	}

	// Get index URL from source options if provided
	indexURL := a.indexURL
	if url, ok := source.Options["index"]; ok && url != "" {
		indexURL = url
	}

	// Fetch package metadata
	metadata, err := a.fetchPackageMetadata(ctx, indexURL, source.URL)
	if err != nil {
		return nil, err
	}

	// Resolve version
	resolvedVersion := version
	if version == "" || version == "latest" {
		resolvedVersion = metadata.Info.Version
		if resolvedVersion == "" {
			return nil, fmt.Errorf("no latest version found for package %s", source.URL)
		}
	}

	// Get version-specific releases
	releases, ok := metadata.Releases[resolvedVersion]
	if !ok || len(releases) == 0 {
		return nil, fmt.Errorf("version %s not found for package %s. Please verify the version is correct", resolvedVersion, source.URL)
	}

	// Find source distribution (sdist) - prefer .tar.gz
	var downloadURL string
	for _, release := range releases {
		if release.PackageType == "sdist" && strings.HasSuffix(release.Filename, ".tar.gz") {
			downloadURL = release.URL
			break
		}
	}

	// If no sdist found, try wheel
	if downloadURL == "" {
		for _, release := range releases {
			if release.PackageType == "bdist_wheel" {
				downloadURL = release.URL
				break
			}
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no suitable distribution found for package %s version %s", source.URL, resolvedVersion)
	}

	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := a.downloadAndExtract(ctx, downloadURL, tempDir); err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &port.DownloadResult{
		Path:    tempDir,
		Version: resolvedVersion,
	}, nil
}

// GetLatestVersion retrieves the latest version from PyPI.
// It returns the version specified in the package info.
func (a *PipAdapter) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if err := source.Validate(); err != nil {
		return "", fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "pip" {
		return "", fmt.Errorf("source type must be 'pip', got '%s'", source.Type)
	}

	// Get index URL from source options if provided
	indexURL := a.indexURL
	if url, ok := source.Options["index"]; ok && url != "" {
		indexURL = url
	}

	// Fetch package metadata
	metadata, err := a.fetchPackageMetadata(ctx, indexURL, source.URL)
	if err != nil {
		return "", err
	}

	// Return latest version
	if metadata.Info.Version == "" {
		return "", fmt.Errorf("no latest version found for package %s", source.URL)
	}

	return metadata.Info.Version, nil
}

// pypiPackageMetadata represents the structure of PyPI package metadata.
type pypiPackageMetadata struct {
	Releases map[string][]pypiReleaseFile `json:"releases"`
	Info     pypiInfo                     `json:"info"`
}

// pypiInfo represents the info field in PyPI package metadata.
type pypiInfo struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

// pypiReleaseFile represents a single release file in PyPI package metadata.
type pypiReleaseFile struct {
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	PackageType string `json:"packagetype"` // "sdist" or "bdist_wheel"
}

// fetchPackageMetadata fetches package metadata from PyPI.
func (a *PipAdapter) fetchPackageMetadata(ctx context.Context, indexURL, packageName string) (*pypiPackageMetadata, error) {
	url := fmt.Sprintf("%s/pypi/%s/json", strings.TrimSuffix(indexURL, "/"), packageName)

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

	var metadata pypiPackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to parse package metadata for %s: %w", packageName, err)
	}

	return &metadata, nil
}

// downloadAndExtract downloads a distribution and extracts it to the target directory.
func (a *PipAdapter) downloadAndExtract(ctx context.Context, downloadURL, targetDir string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to download distribution from %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, downloadURL)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to download distribution from %s: HTTP status %d", domain.ErrNetworkFailure, downloadURL, resp.StatusCode)
	}

	// Extract based on file type
	switch {
	case strings.HasSuffix(downloadURL, ".tar.gz"):
		if err := a.extractTarGz(resp.Body, targetDir); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	case strings.HasSuffix(downloadURL, ".whl"):
		// For wheel files, we need to download first then extract as zip
		tmpFile, err := os.CreateTemp("", "skills-pkg-pip-*.whl")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer func() {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
		}()

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return fmt.Errorf("failed to download wheel file: %w", err)
		}

		// Wheel files are zip files, but we'll treat them simply by extracting
		// For simplicity, we could use the zip extraction logic
		return fmt.Errorf("wheel file extraction not yet implemented")
	default:
		return fmt.Errorf("unsupported distribution format: %s", downloadURL)
	}

	return nil
}

// extractTarGz extracts a gzipped tarball to the target directory.
// Python packages typically have a directory prefix, which is stripped during extraction.
func (a *PipAdapter) extractTarGz(r io.Reader, targetDir string) error {
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
			parts := strings.SplitN(header.Name, "/", pipPathSplitLimit)
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

// createTempDir creates a temporary directory for pip packages.
// It uses the SKILLSPKG_TEMP_DIR environment variable if set, otherwise uses os.TempDir().
func (a *PipAdapter) createTempDir() (string, error) {
	baseDir := os.Getenv("SKILLSPKG_TEMP_DIR")
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Generate a unique directory name using hash
	hash := sha256.New()
	pidBytes := fmt.Appendf(nil, "%d", os.Getpid())
	_, _ = hash.Write(pidBytes)
	dirName := fmt.Sprintf("skills-pkg-pip-%x", hash.Sum(nil)[:8])

	tempDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(tempDir, dirPerms); err != nil {
		return "", err
	}

	return tempDir, nil
}
