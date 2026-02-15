package port

// HashService is the abstraction interface for calculating directory hashes.
// It provides hash calculation for skill integrity verification.
// Requirements: 5.1

// HashResult represents the result of a hash calculation.
// It contains the hash algorithm and the hex-encoded hash value.
// Requirements: 5.2
type HashResult struct {
	Algorithm string // Hash algorithm (e.g., "sha256")
	Value     string // Hex-encoded hash value
}
