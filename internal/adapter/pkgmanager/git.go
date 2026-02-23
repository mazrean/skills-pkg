package pkgmanager

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mazrean/skills-pkg/internal/domain"
	"github.com/mazrean/skills-pkg/internal/port"
	"golang.org/x/mod/semver"
)

const (
	// defaultDirPerm is the default permission for created directories
	defaultDirPerm = 0755
)

// Git implements the PackageManager interface for Git repositories.
// It handles cloning repositories, checking out specific versions (tags or commits),
// and retrieving the latest version.
// Requirements: 3.1, 3.2, 3.5, 3.6, 7.3, 11.2
type Git struct{}

// NewGit creates a new Git adapter instance.
func NewGit() *Git {
	return &Git{}
}

// SourceType returns "git" to identify this adapter as a Git package manager.
// Requirements: 11.2
func (a *Git) SourceType() string {
	return "git"
}

// Download downloads a skill from a Git repository.
// It clones the repository to a temporary directory and checks out the specified version.
// If version is "latest" or empty, it uses the default branch's latest commit.
// Requirements: 3.1, 3.2, 3.5, 3.6, 12.2, 12.3
func (a *Git) Download(ctx context.Context, source *port.Source, version string) (*port.DownloadResult, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "git" {
		return nil, fmt.Errorf("source type must be 'git', got '%s'", source.Type)
	}

	// Create temporary directory for cloning
	tempDir, err := a.createTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Clone the repository
	repo, err := a.cloneRepository(ctx, source.URL, tempDir)
	if err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	// Determine and checkout the target version
	actualVersion, err := a.checkoutVersion(repo, version)
	if err != nil {
		// Clean up on error
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &port.DownloadResult{
		Path:      tempDir,
		Version:   actualVersion,
		FromGoMod: false,
	}, nil
}

// GetLatestVersion retrieves the latest version from a Git repository.
// It returns the latest tag if available, otherwise the latest commit hash.
// Requirements: 7.3, 12.2, 12.3
func (a *Git) GetLatestVersion(ctx context.Context, source *port.Source) (string, error) {
	if err := source.Validate(); err != nil {
		return "", fmt.Errorf("invalid source configuration: %w", err)
	}

	if source.Type != "git" {
		return "", fmt.Errorf("source type must be 'git', got '%s'", source.Type)
	}

	// Create temporary directory for cloning
	tempDir, err := a.createTempDir()
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Clone the repository
	repo, err := a.cloneRepository(ctx, source.URL, tempDir)
	if err != nil {
		return "", err
	}

	// Try to get the latest tag first
	latestTag, err := a.getLatestTag(repo)
	if err == nil && latestTag != "" {
		return latestTag, nil
	}

	// Fall back to HEAD commit hash
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return head.Hash().String(), nil
}

// createTempDir creates a temporary directory for cloning Git repositories.
// It uses the SKILLSPKG_TEMP_DIR environment variable if set, otherwise uses os.TempDir().
func (a *Git) createTempDir() (string, error) {
	baseDir := os.Getenv("SKILLSPKG_TEMP_DIR")
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Generate a unique directory name using hash
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%d", os.Getpid())
	dirName := fmt.Sprintf("skills-pkg-%x", hash.Sum(nil)[:8])

	tempDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(tempDir, defaultDirPerm); err != nil {
		return "", err
	}

	return tempDir, nil
}

// cloneRepository clones a Git repository from the given URL to the target directory.
// Requirements: 3.1, 3.5, 12.2, 12.3
func (a *Git) cloneRepository(ctx context.Context, url, targetDir string) (*git.Repository, error) {
	auth, err := buildAuthMethod(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkFailure, err)
	}

	repo, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
		URL:      url,
		Auth:     auth,
		Progress: nil,
	})
	if err != nil {
		// Classify the error for better user feedback
		if strings.Contains(err.Error(), "authentication required") {
			return nil, fmt.Errorf("%w: failed to clone repository %s: authentication required. Set GIT_TOKEN, GITHUB_TOKEN, or GIT_USERNAME/GIT_PASSWORD environment variables for HTTPS, or ensure SSH credentials are configured", domain.ErrNetworkFailure, url)
		}
		if strings.Contains(err.Error(), "repository not found") {
			return nil, fmt.Errorf("%w: failed to clone repository %s: repository not found. Please verify the URL is correct", domain.ErrNetworkFailure, url)
		}
		if strings.Contains(err.Error(), "network") || strings.Contains(err.Error(), "connection") {
			return nil, fmt.Errorf("%w: failed to clone repository %s: network error. Please check your internet connection and try again", domain.ErrNetworkFailure, url)
		}
		return nil, fmt.Errorf("%w: failed to clone repository %s: %v", domain.ErrNetworkFailure, url, err)
	}

	return repo, nil
}

// checkoutVersion checks out the specified version in the repository.
// If version is "latest" or empty, it uses the HEAD of the default branch.
// Requirements: 3.1, 3.2, 3.6, 12.2, 12.3
func (a *Git) checkoutVersion(repo *git.Repository, version string) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Handle "latest" or empty version - use current HEAD
	if version == "" || version == "latest" {
		head, err := repo.Head()
		if err != nil {
			return "", fmt.Errorf("failed to get HEAD reference: %w", err)
		}
		return head.Hash().String(), nil
	}

	// Try to resolve as a tag first
	tagRef := plumbing.NewTagReferenceName(version)
	if _, err := repo.Reference(tagRef, true); err == nil {
		// Tag exists, checkout the tag
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: tagRef,
		}); err != nil {
			return "", fmt.Errorf("failed to checkout tag %s: %w", version, err)
		}
		return version, nil
	}

	// Try to resolve as a commit hash
	hash := plumbing.NewHash(version)
	if _, err := repo.CommitObject(hash); err == nil {
		// Commit exists, checkout the commit
		if err := worktree.Checkout(&git.CheckoutOptions{
			Hash: hash,
		}); err != nil {
			return "", fmt.Errorf("failed to checkout commit %s: %w", version, err)
		}
		return version, nil
	}

	// Try to resolve as a branch
	branchRef := plumbing.NewBranchReferenceName(version)
	if _, err := repo.Reference(branchRef, true); err == nil {
		// Branch exists, checkout the branch
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		}); err != nil {
			return "", fmt.Errorf("failed to checkout branch %s: %w", version, err)
		}

		// Return the actual commit hash
		head, err := repo.Head()
		if err != nil {
			return "", fmt.Errorf("failed to get HEAD after checkout: %w", err)
		}
		return head.Hash().String(), nil
	}

	// Version not found
	return "", fmt.Errorf("version %s not found: tag, commit, or branch does not exist. Please verify the version is correct", version)
}

// getLatestTag returns the latest tag in the repository.
// It returns an empty string if no tags are found.
// Requirements: 7.3
func (a *Git) getLatestTag(repo *git.Repository) (string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	var latestRelease, latestPre string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		if !semver.IsValid(tagName) {
			return nil
		}
		if semver.Prerelease(tagName) == "" {
			if semver.Compare(tagName, latestRelease) > 0 {
				latestRelease = tagName
			}
		} else {
			if semver.Compare(tagName, latestPre) > 0 {
				latestPre = tagName
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate tags: %w", err)
	}

	if latestRelease != "" {
		return latestRelease, nil
	}
	return latestPre, nil
}
