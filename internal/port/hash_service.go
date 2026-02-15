package port

import "context"

// HashService is the abstraction interface for calculating directory hashes.
// It provides hash calculation for skill integrity verification.
// Requirements: 5.1
type HashService interface {
	// CalculateHash calculates the hash of a directory.
	// The hash includes both file names and file contents recursively.
	// Returns an error if the directory does not exist or cannot be read.
	CalculateHash(ctx context.Context, dirPath string) (*HashResult, error)

	// HashAlgorithm returns the hash algorithm name (e.g., "sha256").
	HashAlgorithm() string
}

// HashResult represents the result of a hash calculation.
// It contains the hash algorithm and the hex-encoded hash value.
// Requirements: 5.2
type HashResult struct {
	Algorithm string // Hash algorithm (e.g., "sha256")
	Value     string // Hex-encoded hash value
}
