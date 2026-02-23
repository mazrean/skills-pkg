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
}

// HashResult represents the result of a hash calculation.
// The Value field contains the hash with algorithm prefix (e.g., "h1:<base64>" for sha256).
// Requirements: 5.2
type HashResult struct {
	Value string // Hash value with algorithm prefix (e.g., "h1:<base64>")
}
