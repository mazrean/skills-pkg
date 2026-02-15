// Package pkgmanager provides implementations of port interfaces for package manager integrations.
// It includes adapters for Go Module proxy and Git repositories.
package pkgmanager

import (
	"archive/zip"
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
	// dirPerms is the default permission for created directories
	dirPerms = 0755
)

// GoMod implements the PackageManager interface for Go Module proxy.
// It handles downloading modules from Go Module proxy, extracting zip files,
// and retrieving the latest version.
// Requirements: 4.2, 4.5, 4.6, 7.4, 11.2
type GoMod struct {
	httpClient *http.Client
	proxyURL   string
}

// NewGoMod creates a new Go Module adapter instance.
// It uses the default Go Module proxy (https://proxy.golang.org) unless
// overridden by the source options or GOPROXY environment variable.
func NewGoMod() *GoMod {
	proxyURL := os.Getenv("GOPROXY")
	if proxyURL == "" || proxyURL == "direct" || strings.Contains(proxyURL, ",") {
		// Use default proxy if GOPROXY is not set, is "direct", or contains multiple proxies
		proxyURL = "https://proxy.golang.org"
	}

	return &GoMod{
		proxyURL:   proxyURL,
		httpClient: &http.Client{},
	}
}

// SourceType returns "go-module" to identify this adapter as a Go Module package manager.
// Requirements: 11.2
func (a *GoMod) SourceType() string {
	return "go-module"
}

// Download downloads a skill from the Go Module proxy.
// It fetches the module metadata, downloads the zip file, and extracts it to a temporary directory.
// If version is "latest" or empty, it uses the latest version from the proxy.
// Requirements: 4.2, 4.5, 4.6, 12.2, 12.3
func (a *GoMod) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "go-module" {
		return nil, fmt.Errorf("source type must be 'go-module', got '%s'", source.Type)
	}

	// Get proxy URL from source options if provided
	proxyURL := a.proxyURL
	if url, ok := source.Options["proxy"]; ok && url != "" {
		proxyURL = url
	}

	// Resolve version
	resolvedVersion := version
	if version == "" || version == "latest" {
		latestVersion, err := a.fetchLatestVersion(ctx, proxyURL, source.URL)
		if err != nil {
			return nil, err
		}
		resolvedVersion = latestVersion
	}

	// Download zip file
	zipURL := fmt.Sprintf("%s/%s/@v/%s.zip", strings.TrimSuffix(proxyURL, "/"), source.URL, resolvedVersion)

	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := a.downloadAndExtractZip(ctx, zipURL, tempDir, source.URL, resolvedVersion); err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &port.DownloadResult{
		Path:    tempDir,
		Version: resolvedVersion,
	}, nil
}

// GetLatestVersion retrieves the latest version from the Go Module proxy.
// It returns the version specified by the @latest endpoint.
// Requirements: 7.4, 12.2, 12.3
func (a *GoMod) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if err := source.Validate(); err != nil {
		return "", fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "go-module" {
		return "", fmt.Errorf("source type must be 'go-module', got '%s'", source.Type)
	}

	// Get proxy URL from source options if provided
	proxyURL := a.proxyURL
	if url, ok := source.Options["proxy"]; ok && url != "" {
		proxyURL = url
	}

	return a.fetchLatestVersion(ctx, proxyURL, source.URL)
}

// goModuleLatestInfo represents the response from the @latest endpoint.
type goModuleLatestInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

// fetchLatestVersion fetches the latest version from the Go Module proxy.
// Requirements: 7.4, 12.2, 12.3
func (a *GoMod) fetchLatestVersion(ctx context.Context, proxyURL, modulePath string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", strings.TrimSuffix(proxyURL, "/"), modulePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: failed to fetch latest version for %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, modulePath)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("%w: module %s not found. Please verify the module path is correct", domain.ErrNetworkFailure, modulePath)
	}

	if resp.StatusCode == http.StatusGone {
		return "", fmt.Errorf("%w: module %s has been removed from the proxy", domain.ErrNetworkFailure, modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: failed to fetch latest version for %s: HTTP status %d", domain.ErrNetworkFailure, modulePath, resp.StatusCode)
	}

	var info goModuleLatestInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to parse latest version info for %s: %w", modulePath, err)
	}

	if info.Version == "" {
		return "", fmt.Errorf("no version found in latest version info for module %s", modulePath)
	}

	return info.Version, nil
}

// downloadAndExtractZip downloads a zip file and extracts it to the target directory.
// Requirements: 4.2, 4.5, 12.2, 12.3
func (a *GoMod) downloadAndExtractZip(ctx context.Context, zipURL, targetDir, modulePath, version string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zipURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to download module from %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, zipURL)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: version %s not found for module %s. Please verify the version is correct", domain.ErrNetworkFailure, version, modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to download module from %s: HTTP status %d", domain.ErrNetworkFailure, zipURL, resp.StatusCode)
	}

	// Create a temporary file to store the zip
	tmpFile, err := os.CreateTemp("", "skills-pkg-gomod-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	// Download to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to download zip file: %w", err)
	}

	// Extract zip file
	if err := a.extractZip(tmpFile.Name(), targetDir, modulePath, version); err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}

	return nil
}

// extractZip extracts a zip file to the target directory.
// Go Module zip files have a prefix directory with the module path and version,
// which is stripped during extraction.
// Requirements: 4.2
func (a *GoMod) extractZip(zipPath, targetDir, modulePath, version string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	// Go Module zip files have a prefix directory: <module>@<version>/
	prefix := fmt.Sprintf("%s@%s/", modulePath, version)

	for _, f := range r.File {
		// Strip the prefix directory from the path
		name, found := strings.CutPrefix(f.Name, prefix)
		if !found {
			// Skip files that don't have the expected prefix
			continue
		}

		// Skip if the path is empty after stripping
		if name == "" {
			continue
		}

		target := filepath.Join(targetDir, name)

		// Ensure the target is within the target directory (security check)
		if !strings.HasPrefix(target, filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			target != filepath.Clean(targetDir) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(target, dirPerms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		} else {
			// Create file
			if err := os.MkdirAll(filepath.Dir(target), dirPerms); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", target, err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}

			rc, err := f.Open()
			if err != nil {
				_ = outFile.Close()
				return fmt.Errorf("failed to open file in zip: %w", err)
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				_ = rc.Close()
				_ = outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}

			_ = rc.Close()
			_ = outFile.Close()
		}
	}

	return nil
}

// createTempDir creates a temporary directory for Go modules.
// It uses the SKILLSPKG_TEMP_DIR environment variable if set, otherwise uses os.TempDir().
func (a *GoMod) createTempDir() (string, error) {
	baseDir := os.Getenv("SKILLSPKG_TEMP_DIR")
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Generate a unique directory name using hash
	hash := sha256.New()
	pidBytes := fmt.Appendf(nil, "%d", os.Getpid())
	_, _ = hash.Write(pidBytes)
	dirName := fmt.Sprintf("skills-pkg-gomod-%x", hash.Sum(nil)[:8])

	tempDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(tempDir, dirPerms); err != nil {
		return "", err
	}

	return tempDir, nil
}
