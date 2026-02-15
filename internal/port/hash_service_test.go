package port_test

import (
	"context"
	"testing"

	"github.com/mazrean/skills-pkg/internal/port"
)

// TestHashServiceInterface verifies that the HashService interface contract
// can be satisfied by a mock implementation.
// Requirements: 5.1
func TestHashServiceInterface(t *testing.T) {
	t.Run("interface_contract", func(t *testing.T) {
		// Verify that a mock implementation satisfies the interface
		var _ port.HashService = &mockHashService{}
	})
}

// TestHashResultStructure tests HashResult struct fields.
// Requirements: 5.2
func TestHashResultStructure(t *testing.T) {
	result := &port.HashResult{
		Algorithm: "sha256",
		Value:     "a1b2c3d4e5f6",
	}

	if result.Algorithm != "sha256" {
		t.Errorf("HashResult.Algorithm = %v, want %v", result.Algorithm, "sha256")
	}
	if result.Value != "a1b2c3d4e5f6" {
		t.Errorf("HashResult.Value = %v, want %v", result.Value, "a1b2c3d4e5f6")
	}
}

// TestHashServiceAlgorithm tests that hash algorithm is returned correctly.
// Requirements: 5.1, 5.2
func TestHashServiceAlgorithm(t *testing.T) {
	service := &mockHashService{}

	algo := service.HashAlgorithm()
	if algo == "" {
		t.Error("HashAlgorithm() returned empty string")
	}
}

// mockHashService is a mock implementation of HashService for testing.
type mockHashService struct{}

func (m *mockHashService) CalculateHash(ctx context.Context, dirPath string) (*port.HashResult, error) {
	return &port.HashResult{
		Algorithm: "sha256",
		Value:     "mockhash",
	}, nil
}

func (m *mockHashService) HashAlgorithm() string {
	return "sha256"
}
