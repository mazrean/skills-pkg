package port

import "context"

// HashService is the abstraction interface for calculating directory hashes.
// It provides hash calculation for skill integrity verification.
// Requirements: 5.1
type HashService interface {
	// CalculateHash calculates the hash of a directory.
	// It recursively hashes all files including file names and contents.
	// Requirements: 5.1, 5.3
	CalculateHash(ctx context.Context, dirPath string) (*HashResult, error)

	// HashAlgorithm returns the hash algorithm name (e.g., "sha256").
	// Requirements: 5.2
	HashAlgorithm() string
}

// HashResult represents the result of a hash calculation.
// It contains the hash algorithm and the hex-encoded hash value.
// Requirements: 5.2
type HashResult struct {
	Algorithm string // Hash algorithm (e.g., "sha256")
	Value     string // Hex-encoded hash value
}
