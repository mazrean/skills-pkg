// Package port defines interfaces for external system integrations.
// It provides abstractions for package managers, agent providers, and hash services.
package port

import (
	"errors"
)

// PackageManager is the abstraction interface for downloading skills from various sources.
// It supports Git repositories, npm registry, and Go Module proxy.
// Requirements: 11.1, 11.3

// Source represents the source location for a skill.
// It contains the type, URL, and optional parameters.
// Requirements: 2.3, 2.4, 11.4
type Source struct {
	Options map[string]string // Optional parameters (e.g., registry URL)
	Type    string            // "git", "npm", "go-module"
	URL     string            // Git URL, npm package name, Go module path
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
		"git":       true,
		"npm":       true,
		"go-module": true,
	}
	if !validTypes[s.Type] {
		return errors.New("invalid source type: must be git, npm, or go-module")
	}

	return nil
}

// DownloadResult represents the result of a skill download operation.
// It contains the local directory path and the actual version downloaded.
// Requirements: 3.1, 4.1, 4.2
type DownloadResult struct {
	Path    string // Local directory path
	Version string // Actual version downloaded
}
