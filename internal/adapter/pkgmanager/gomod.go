// Package pkgmanager provides implementations of port interfaces for package manager integrations.
// It includes adapters for Go Module proxy and Git repositories.
package pkgmanager

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
)

const (
	// dirPerms is the default permission for created directories
	dirPerms = 0755

	// minGitLsRemoteFields is the minimum number of fields in git ls-remote output
	minGitLsRemoteFields = 2
)

// proxyEntry represents a single entry in GOPROXY.
type proxyEntry struct {
	url      string // "direct", "off", or an actual proxy URL
	fallback bool   // true if this is a fallback entry (comma-separated)
}

// GoMod implements the PackageManager interface for Go Module proxy.
// It handles downloading modules from Go Module proxy, extracting zip files,
// and retrieving the latest version.
// Requirements: 4.2, 4.5, 4.6, 7.4, 11.2
type GoMod struct {
	httpClient *http.Client
	proxies    []proxyEntry
}

// parseGOPROXY parses the GOPROXY environment variable.
// It supports comma-separated (fallback) and pipe-separated (always try) proxies.
// The special values "direct" and "off" are also supported.
func parseGOPROXY(goproxy string) []proxyEntry {
	if goproxy == "" {
		// Default to https://proxy.golang.org,direct
		return []proxyEntry{
			{url: "https://proxy.golang.org", fallback: true},
			{url: "direct", fallback: true},
		}
	}

	var entries []proxyEntry
	// Split by comma first (fallback mode)
	for part := range strings.SplitSeq(goproxy, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Each comma-separated part may contain pipe-separated entries
		i := 0
		for pipePart := range strings.SplitSeq(part, "|") {
			pipePart = strings.TrimSpace(pipePart)
			if pipePart == "" {
				continue
			}
			// Only the first pipe-separated entry is a fallback
			entries = append(entries, proxyEntry{
				url:      pipePart,
				fallback: i == 0,
			})
			i++
		}
	}

	if len(entries) == 0 {
		// Fallback to default
		return []proxyEntry{
			{url: "https://proxy.golang.org", fallback: true},
			{url: "direct", fallback: true},
		}
	}

	return entries
}

// NewGoMod creates a new Go Module adapter instance.
// It uses the default Go Module proxy (https://proxy.golang.org) unless
// overridden by the source options or GOPROXY environment variable.
func NewGoMod() *GoMod {
	goproxy := os.Getenv("GOPROXY")
	proxies := parseGOPROXY(goproxy)

	return &GoMod{
		proxies:    proxies,
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

	// Get proxies from source options if provided, otherwise use configured proxies
	proxies := a.proxies
	if url, ok := source.Options["proxy"]; ok && url != "" {
		proxies = parseGOPROXY(url)
	}

	// Resolve version
	resolvedVersion := version
	if version == "" || version == "latest" {
		latestVersion, err := a.fetchLatestVersionWithProxies(ctx, proxies, source.URL)
		if err != nil {
			return nil, err
		}
		resolvedVersion = latestVersion
	}

	// Create temp directory
	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Try downloading with each proxy
	err = a.downloadWithProxies(ctx, proxies, source.URL, resolvedVersion, tempDir)
	if err != nil {
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

	// Get proxies from source options if provided, otherwise use configured proxies
	proxies := a.proxies
	if url, ok := source.Options["proxy"]; ok && url != "" {
		proxies = parseGOPROXY(url)
	}

	return a.fetchLatestVersionWithProxies(ctx, proxies, source.URL)
}

// goModuleLatestInfo represents the response from the @latest endpoint.
type goModuleLatestInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

// fetchLatestVersionWithProxies tries to fetch the latest version using the configured proxies.
// It tries each proxy in order until one succeeds or all fail.
func (a *GoMod) fetchLatestVersionWithProxies(ctx context.Context, proxies []proxyEntry, modulePath string) (string, error) {
	var lastErr error

	for _, proxy := range proxies {
		if proxy.url == "off" {
			return "", fmt.Errorf("%w: GOPROXY is set to 'off', downloads are disabled", domain.ErrNetworkFailure)
		}

		if proxy.url == "direct" {
			version, err := a.fetchLatestVersionDirect(ctx, modulePath)
			if err == nil {
				return version, nil
			}
			lastErr = err
			if !proxy.fallback {
				continue
			}
			// If this is a fallback entry, try the next proxy
			continue
		}

		// Try proxy
		version, err := a.fetchLatestVersion(ctx, proxy.url, modulePath)
		if err == nil {
			return version, nil
		}

		lastErr = err
		// If this is not a fallback entry (pipe-separated), continue to the next one
		if !proxy.fallback {
			continue
		}
		// Otherwise, this is a fallback entry (comma-separated), so try the next one
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("%w: failed to fetch latest version for %s from any proxy", domain.ErrNetworkFailure, modulePath)
}

// downloadWithProxies tries to download the module using the configured proxies.
// It tries each proxy in order until one succeeds or all fail.
func (a *GoMod) downloadWithProxies(ctx context.Context, proxies []proxyEntry, modulePath, version, targetDir string) error {
	var lastErr error

	for _, proxy := range proxies {
		if proxy.url == "off" {
			return fmt.Errorf("%w: GOPROXY is set to 'off', downloads are disabled", domain.ErrNetworkFailure)
		}

		if proxy.url == "direct" {
			err := a.downloadDirect(ctx, modulePath, version, targetDir)
			if err == nil {
				return nil
			}
			lastErr = err
			if !proxy.fallback {
				continue
			}
			// If this is a fallback entry, try the next proxy
			continue
		}

		// Try proxy
		zipURL := fmt.Sprintf("%s/%s/@v/%s.zip", strings.TrimSuffix(proxy.url, "/"), modulePath, version)
		err := a.downloadAndExtractZip(ctx, zipURL, targetDir, modulePath, version)
		if err == nil {
			return nil
		}

		lastErr = err
		// If this is not a fallback entry (pipe-separated), continue to the next one
		if !proxy.fallback {
			continue
		}
		// Otherwise, this is a fallback entry (comma-separated), so try the next one
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("%w: failed to download module %s from any proxy", domain.ErrNetworkFailure, modulePath)
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

// fetchLatestVersionDirect fetches the latest version directly from the version control system.
// It uses git to query the repository for the latest tag.
func (a *GoMod) fetchLatestVersionDirect(ctx context.Context, modulePath string) (string, error) {
	// Convert module path to repository URL
	// For simplicity, we assume the module path is a valid git repository URL
	// In a full implementation, this would need to handle various VCS systems and URL schemes
	repoURL := "https://" + modulePath

	// Use git ls-remote to get the latest tag
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", repoURL)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("%w: failed to fetch tags from %s: %s", domain.ErrNetworkFailure, repoURL, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("%w: failed to fetch tags from %s: %w", domain.ErrNetworkFailure, repoURL, err)
	}

	// Parse the output to find the latest version tag
	lines := strings.Split(string(output), "\n")
	var latestVersion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < minGitLsRemoteFields {
			continue
		}

		ref := parts[1]
		// Extract version from refs/tags/vX.Y.Z
		if tag, found := strings.CutPrefix(ref, "refs/tags/"); found {
			// Simple heuristic: use the last tag as the latest
			// In a full implementation, this would need semantic version comparison
			latestVersion = tag
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("%w: no version tags found for module %s", domain.ErrNetworkFailure, modulePath)
	}

	return latestVersion, nil
}

// downloadDirect downloads a module directly from the version control system.
// It uses git to clone the repository at the specified version.
func (a *GoMod) downloadDirect(ctx context.Context, modulePath, version, targetDir string) error {
	// Convert module path to repository URL
	repoURL := "https://" + modulePath

	// Create a temporary directory for the clone
	cloneDir, err := os.MkdirTemp("", "skills-pkg-git-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(cloneDir)
	}()

	// Clone the repository with the specified tag/branch
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", version, repoURL, cloneDir)
	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("%w: failed to clone repository %s at version %s: %s", domain.ErrNetworkFailure, repoURL, version, string(exitErr.Stderr))
		}
		return fmt.Errorf("%w: failed to clone repository %s at version %s: %w", domain.ErrNetworkFailure, repoURL, version, err)
	}

	// Copy files from clone directory to target directory, excluding .git
	entries, err := os.ReadDir(cloneDir)
	if err != nil {
		return fmt.Errorf("failed to read clone directory: %w", err)
	}

	for _, entry := range entries {
		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}

		src := filepath.Join(cloneDir, entry.Name())
		dst := filepath.Join(targetDir, entry.Name())

		if entry.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", entry.Name(), err)
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	// Create destination directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
