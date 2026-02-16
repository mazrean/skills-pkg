// Package port defines interfaces for external system integrations.
// It provides abstractions for package managers, agent providers, and hash services.
package port

import (
	"context"
	"errors"
)

// PackageManager is the abstraction interface for downloading skills from various sources.
// It supports Git repositories and Go Module proxy.
// Requirements: 11.1, 11.3
type PackageManager interface {
	// Download downloads the skill from the source.
	// Returns the local directory path and actual version downloaded.
	Download(ctx context.Context, source *Source, version string) (*DownloadResult, error)

	// GetLatestVersion retrieves the latest version of the skill.
	GetLatestVersion(ctx context.Context, source *Source) (string, error)

	// SourceType returns the type of the source (git, go-mod).
	SourceType() string
}

// Source represents the source location for a skill.
// It contains the type, URL, and optional parameters.
// Requirements: 2.3, 2.4, 11.4
type Source struct {
	Options map[string]string // Optional parameters (e.g., registry URL)
	Type    string            // "git", "go-mod"
	URL     string            // Git URL, Go module path
}

// Validate validates the source configuration.
// It checks that required fields are present and that the source type is valid.
// Requirements: 11.4, 12.2, 12.3
func (s *Source) Validate() error {
	if s.Type == "" {
		return errors.New("source type is required")
	}
	if s.URL == "" {
		return errors.New("source URL is required")
	}

	// Validate source type
	validTypes := map[string]bool{
		"git":    true,
		"go-mod": true,
	}
	if !validTypes[s.Type] {
		return errors.New("invalid source type: must be git or go-mod")
	}

	return nil
}

// DownloadResult represents the result of a skill download operation.
// It contains the local directory path and the actual version downloaded.
// Requirements: 3.1, 4.1, 4.2
type DownloadResult struct {
	Path       string // Local directory path
	Version    string // Actual version downloaded
	FromGoMod  bool   // Whether the version was resolved from go.mod
}
