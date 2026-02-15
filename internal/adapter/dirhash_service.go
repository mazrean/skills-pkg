package adapter

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/mazrean/skills-pkg/internal/port"
)

// DirhashService is an implementation of HashService using golang.org/x/mod/sumdb/dirhash.
// It calculates directory hashes using SHA-256 algorithm.
// Requirements: 5.1
type DirhashService struct{}

// NewDirhashService creates a new DirhashService instance.
func NewDirhashService() *DirhashService {
	return &DirhashService{}
}

// CalculateHash calculates the hash of a directory recursively.
// It includes both file names and file contents in the hash calculation.
// The hash is calculated using the SHA-256 algorithm via golang.org/x/mod/sumdb/dirhash.HashDir.
// Requirements: 5.1, 12.2, 12.3
func (s *DirhashService) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	// Verify that the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s: %w", dirPath, err)
		}
		return nil, fmt.Errorf("failed to access directory %s: %w", dirPath, err)
	}

	// Verify that the path is a directory, not a file
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Calculate hash using dirhash.HashDir (SHA-256 based)
	// HashDir returns format "h1:<base64-encoded-sha256>" which is the standard Go module hash format
	hashValue, err := dirhash.HashDir(dirPath, "", dirhash.Hash1)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash for directory %s: %w", dirPath, err)
	}

	// HashDir returns format "h1:<base64>" - we use this as-is for consistency with Go module ecosystem
	return &port.HashResult{
		Algorithm: "sha256",
		Value:     hashValue,
	}, nil
}

// HashAlgorithm returns the hash algorithm name used by this service.
// Requirements: 5.1, 5.2
func (s *DirhashService) HashAlgorithm() string {
	return "sha256"
}
